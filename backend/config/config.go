package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Config struct {
	Server                        ServerConfig  `json:"server"`
	JWTSecret                     string        `json:"jwt_secret"`
	TokenExpire                   time.Duration `json:"-"`
	UploadPath                    string        `json:"upload_path"`
	AdminMaxSize                  int64         `json:"-"`
	DefaultUserStorageLimitMB     int           `json:"default_user_storage_limit_mb"`
	DefaultUserImageLimit         int           `json:"default_user_image_limit"`
	DefaultUserSingleImageLimitMB int           `json:"default_user_single_image_limit_mb"`
	LogRetention                  int           `json:"log_retention"`
}

var AppConfig = Config{
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	},
	JWTSecret:                     "image_share_secret_key_2024",
	TokenExpire:                   7 * 24 * time.Hour,
	UploadPath:                    "./uploads",
	AdminMaxSize:                  20 * 1024 * 1024,
	DefaultUserStorageLimitMB:     100,
	DefaultUserImageLimit:         50,
	DefaultUserSingleImageLimitMB: 10,
	LogRetention:                  7,
}

var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

var AllowedMIMETypes = map[string]bool{
	"image/jpeg":  true,
	"image/png":   true,
	"image/gif":   true,
	"image/webp":  true,
	"image/pjpeg": true,
	"image/x-png": true,
}

// GetConfigPath 获取配置文件路径（与可执行文件同目录）
func GetConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exePath), "config.json")
}

// stripComments 剥离JSON中的 // 注释
func stripComments(data []byte) []byte {
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		// 跳过纯注释行
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		// 去除行内注释（不在引号内的 //）
		inQuote := false
		quoteChar := byte(0)
		result := make([]byte, 0, len(line))
		for i := 0; i < len(line); i++ {
			ch := line[i]
			if inQuote {
				result = append(result, ch)
				if ch == quoteChar && (i == 0 || line[i-1] != '\\') {
					inQuote = false
				}
			} else {
				if ch == '"' || ch == '\'' {
					inQuote = true
					quoteChar = ch
					result = append(result, ch)
				} else if i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
					break
				} else {
					result = append(result, ch)
				}
			}
		}
		lines = append(lines, string(result))
	}
	return []byte(strings.Join(lines, "\n"))
}

// LoadConfig 从配置文件加载配置
func LoadConfig() error {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return SaveConfig()
		}
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 剥离注释后解析JSON
	cleanData := stripComments(data)
	var cfg Config
	if err := json.Unmarshal(cleanData, &cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	if cfg.Server.Host != "" {
		AppConfig.Server.Host = cfg.Server.Host
	}
	if cfg.Server.Port > 0 {
		AppConfig.Server.Port = cfg.Server.Port
	}
	if cfg.JWTSecret != "" {
		AppConfig.JWTSecret = cfg.JWTSecret
	}
	if cfg.UploadPath != "" {
		AppConfig.UploadPath = cfg.UploadPath
	}
	if cfg.DefaultUserStorageLimitMB > 0 {
		AppConfig.DefaultUserStorageLimitMB = cfg.DefaultUserStorageLimitMB
	}
	if cfg.DefaultUserImageLimit > 0 {
		AppConfig.DefaultUserImageLimit = cfg.DefaultUserImageLimit
	}
	if cfg.DefaultUserSingleImageLimitMB > 0 {
		AppConfig.DefaultUserSingleImageLimitMB = cfg.DefaultUserSingleImageLimitMB
	}
	if cfg.LogRetention > 0 {
		AppConfig.LogRetention = cfg.LogRetention
	}

	return nil
}

// SaveConfig 保存当前配置到文件（带注释）
func SaveConfig() error {
	configPath := GetConfigPath()

	content := `{
    // 服务器配置
    // host: 监听地址 (默认: "0.0.0.0")
    //   "0.0.0.0"       监听所有IPv4地址，允许外部访问
    //   "127.0.0.1"     仅监听本地IPv4回环地址，仅本机可访问
    //   "::"            监听所有IPv6地址（兼容IPv4），允许外部访问
    //   "::1"           仅监听本地IPv6回环地址，仅本机可访问
    //   "192.168.x.x"   绑定到指定IPv4网卡地址
    //   "fe80::x"       绑定到指定IPv6网卡地址
    "server": {
        "host": "` + AppConfig.Server.Host + `",
        "port": ` + fmt.Sprintf("%d", AppConfig.Server.Port) + `
    },
    // JWT密钥 (默认: "image_share_secret_key_2024")
    // 请修改为随机字符串以增强安全性，修改后已登录用户需重新登录
    "jwt_secret": "` + AppConfig.JWTSecret + `",
    // 上传文件存储路径 (默认: "./uploads")
    // 相对于可执行文件目录，也可使用绝对路径如 "D:/imageshare/uploads"
    "upload_path": "` + AppConfig.UploadPath + `",
    // 新用户默认存储空间MB (默认: 100)
    "default_user_storage_limit_mb": ` + fmt.Sprintf("%d", AppConfig.DefaultUserStorageLimitMB) + `,
    // 新用户默认图片数量上限 (默认: 50)
    "default_user_image_limit": ` + fmt.Sprintf("%d", AppConfig.DefaultUserImageLimit) + `,
    // 新用户默认单张图片大小上限MB (默认: 10)
    "default_user_single_image_limit_mb": ` + fmt.Sprintf("%d", AppConfig.DefaultUserSingleImageLimitMB) + `,
    // 日志保留数量 (默认: 7)
    // 超过此数量的旧日志文件将被自动清理
    "log_retention": ` + fmt.Sprintf("%d", AppConfig.LogRetention) + `
}`

	return os.WriteFile(configPath, []byte(content), 0644)
}

// GetListenAddr 获取监听地址字符串
func GetListenAddr() string {
	host := AppConfig.Server.Host
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, AppConfig.Server.Port)
	}
	return fmt.Sprintf("%s:%d", host, AppConfig.Server.Port)
}

// GetUploadPath 获取上传目录的绝对路径
func GetUploadPath() string {
	if filepath.IsAbs(AppConfig.UploadPath) {
		return AppConfig.UploadPath
	}
	exePath, err := os.Executable()
	if err != nil {
		return AppConfig.UploadPath
	}
	return filepath.Join(filepath.Dir(exePath), AppConfig.UploadPath)
}
