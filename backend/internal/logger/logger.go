package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"imageshare/config"
)

var (
	logDir     string
	currentFile *os.File
	currentDate string // 当前日志文件对应的日期
	mu         sync.Mutex
)

// Init 初始化日志系统
func Init(baseDir string) error {
	logDir = filepath.Join(baseDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	if err := openNewLogFile(); err != nil {
		return fmt.Errorf("创建日志文件失败: %v", err)
	}

	// 启动定时器：每分钟检查是否需要分割日志
	go func() {
		for {
			time.Sleep(time.Minute)
			checkRotation()
		}
	}()

	// 启动定时器：每天清理旧日志
	go func() {
		for {
			cleanupOldLogs()
			time.Sleep(time.Hour)
		}
	}()

	return nil
}

// openNewLogFile 创建新的日志文件
func openNewLogFile() error {
	now := time.Now()
	fileName := now.Format("20060102-15-04") + ".log"
	filePath := filepath.Join(logDir, fileName)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if currentFile != nil {
		currentFile.Close()
	}
	currentFile = f
	currentDate = now.Format("20060102")

	return nil
}

// checkRotation 检查是否需要分割日志（跨天）
func checkRotation() {
	mu.Lock()
	defer mu.Unlock()

	today := time.Now().Format("20060102")
	if today != currentDate {
		openNewLogFile()
	}
}

// cleanupOldLogs 清理超出保留数量的旧日志
func cleanupOldLogs() {
	mu.Lock()
	defer mu.Unlock()

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry.Name())
		}
	}

	retention := config.AppConfig.LogRetention
	if retention <= 0 {
		retention = 7
	}

	if len(logFiles) <= retention {
		return
	}

	// 按文件名排序（文件名包含日期时间，排序后旧文件在前）
	sort.Strings(logFiles)

	// 删除超出保留数量的旧文件
	for i := 0; i < len(logFiles)-retention; i++ {
		os.Remove(filepath.Join(logDir, logFiles[i]))
	}
}

// WriteLog 写入日志（同时输出到控制台和文件）
func WriteLog(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// 输出到控制台（带颜色）
	fmt.Print(msg)

	// 写入文件（去除ANSI颜色码）
	mu.Lock()
	defer mu.Unlock()

	if currentFile != nil {
		cleanMsg := stripANSI(msg)
		currentFile.WriteString(cleanMsg)
		currentFile.Sync()
	}
}

// WriteLogConsole 仅输出到控制台
func WriteLogConsole(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// stripANSI 去除ANSI转义序列
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// GetCurrentLogPath 获取当前日志文件路径
func GetCurrentLogPath() string {
	mu.Lock()
	defer mu.Unlock()
	if currentFile != nil {
		return currentFile.Name()
	}
	return ""
}

// ReadCurrentLog 读取当前日志文件内容
func ReadCurrentLog(maxLines int) string {
	mu.Lock()
	path := ""
	if currentFile != nil {
		path = currentFile.Name()
	}
	mu.Unlock()

	if path == "" {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, "\n")
}

// ReadCurrentLogOps 读取当前日志文件中的操作日志（过滤网络请求）
func ReadCurrentLogOps(maxLines int) string {
	mu.Lock()
	path := ""
	if currentFile != nil {
		path = currentFile.Name()
	}
	mu.Unlock()

	if path == "" {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// 操作日志关键词
	opKeywords := []string{"[登录]", "[上传]", "[创建]", "[删除]", "[修改]", "[系统]"}
	var result []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		matched := false
		for _, kw := range opKeywords {
			if strings.Contains(line, kw) {
				matched = true
				break
			}
		}
		if matched {
			result = append(result, line)
		}
	}

	if len(result) > maxLines {
		result = result[len(result)-maxLines:]
	}
	return strings.Join(result, "\n")
}

// Writer 实现io.Writer接口，用于Gin日志输出
type Writer struct{}

func (w *Writer) Write(p []byte) (n int, err error) {
	msg := string(p)
	// 输出到控制台（msg已包含ANSI颜色）
	fmt.Print(msg)

	// 写入文件（去除颜色码）
	mu.Lock()
	defer mu.Unlock()
	if currentFile != nil {
		cleanMsg := stripANSI(msg)
		currentFile.WriteString(cleanMsg)
		currentFile.Sync()
	}

	return len(p), nil
}

// Ensure Writer implements io.Writer
var _ io.Writer = (*Writer)(nil)
