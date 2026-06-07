package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"imageshare/config"
	"imageshare/internal/controller"
	"imageshare/internal/logger"
	"imageshare/internal/middleware"
	"imageshare/internal/models"
	"imageshare/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed all:frontend
var embeddedFrontend embed.FS

// ANSI颜色常量
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// getBaseDir 获取可执行文件所在目录
func getBaseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

// customLogger 自定义GIN日志中间件，过滤静态资源，只显示API请求
func customLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// 过滤静态资源请求
		if strings.HasPrefix(path, "/assets/") || path == "/favicon.ico" || path == "/vite.svg" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		statusCode := c.Writer.Status()

		// 状态码颜色
		statusColor := colorGreen
		if statusCode >= 500 {
			statusColor = colorRed
		} else if statusCode >= 400 {
			statusColor = colorYellow
		} else if statusCode >= 300 {
			statusColor = colorCyan
		}

		// 方法颜色
		methodColor := colorCyan
		switch c.Request.Method {
		case "GET":
			methodColor = colorBlue
		case "POST":
			methodColor = colorGreen
		case "PUT":
			methodColor = colorYellow
		case "DELETE":
			methodColor = colorRed
		}

		logger.WriteLog("%s[%s]%s %s%3d%s %s%7v%s %s%-7s%s %s\n",
			colorGray, start.Format("2006/01/02 15:04:05"), colorReset,
			statusColor, statusCode, colorReset,
			colorGray, latency.Round(time.Microsecond), colorReset,
			methodColor, c.Request.Method, colorReset,
			path,
		)
	}
}

// bizLog 业务日志
func bizLog(category, message string) {
	var catColor string
	switch category {
	case "登录":
		catColor = colorGreen
	case "上传":
		catColor = colorCyan
	case "创建":
		catColor = colorYellow
	case "删除":
		catColor = colorRed
	case "修改":
		catColor = colorYellow
	case "系统":
		catColor = colorBlue
	default:
		catColor = colorCyan
	}
	logger.WriteLog("%s[%s]%s %s[%s]%s %s\n",
		colorGray, time.Now().Format("2006/01/02 15:04:05"), colorReset,
		catColor, category, colorReset,
		message,
	)
}

// printBanner 打印启动Logo
func printBanner() {
	separator := strings.Repeat("=", 78)
	fmt.Println()
	fmt.Println(separator)
	fmt.Println()
	fmt.Println(`/$$$$$$                                                      /$$$$$$  /$$                                    `)
	fmt.Println(`|_  $$_/                                                     /$$__  $$| $$                                    `)
	fmt.Println(`  | $$   /$$$$$$/$$$$   /$$$$$$   /$$$$$$   /$$$$$$         | $$  \__/| $$$$$$$   /$$$$$$   /$$$$$$   /$$$$$$ `)
	fmt.Println(`  | $$  | $$_  $$_  $$ |____  $$ /$$__  $$ /$$__  $$ /$$$$$$|  $$$$$$ | $$__  $$ |____  $$ /$$__  $$ /$$__  $$`)
	fmt.Println(`  | $$  | $$ \ $$ \ $$  /$$$$$$$| $$  \ $$| $$$$$$$$|______/ \____  $$| $$  \ $$  /$$$$$$$| $$  \__/| $$$$$$$$`)
	fmt.Println(`  | $$  | $$ | $$ | $$ /$$__  $$| $$  | $$| $$_____/         /$$  \ $$| $$  | $$ /$$__  $$| $$      | $$_____/`)
	fmt.Println(` /$$$$$$| $$ | $$ | $$|  $$$$$$$|  $$$$$$$|  $$$$$$$        |  $$$$$$/| $$  | $$|  $$$$$$$| $$      |  $$$$$$$`)
	fmt.Println(`|______/|__/ |__/ |__/ \_______/ \____  $$ \_______/         \______/ |__/  |__/ \_______/|__/       \_______/`)
	fmt.Println(`                                 /$$  \ $$                                                                    `)
	fmt.Println(`                                |  $$$$$$/                                                                    `)
	fmt.Println(`                                 \______/`)
	fmt.Println()
	fmt.Println("  Made By mcBill")
	fmt.Println("  Github : https://github.com/mcbill1")
	fmt.Println("  Site   : https://mcbill.top")
	fmt.Println()
	fmt.Println("  Image-Share v1.0 release 2026/06/07")
	fmt.Println()
	fmt.Println(separator)
}

// checkPort 检查端口是否被占用
func checkPort(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	listener.Close()
	return nil
}

func main() {
	// 检查命令行参数：-changepasswd <密码>
	if len(os.Args) >= 3 && os.Args[1] == "-changepasswd" {
		newPassword := os.Args[2]
		enableANSI()

		// 加载配置
		if err := config.LoadConfig(); err != nil {
			fmt.Printf("%s[错误]%s 加载配置文件失败: %v\n", colorRed, colorReset, err)
			os.Exit(1)
		}

		// 初始化数据库
		baseDir := getBaseDir()
		dbPath := filepath.Join(baseDir, "database")
		if err := models.InitDB(dbPath); err != nil {
			fmt.Printf("%s[错误]%s 数据库初始化失败: %v\n", colorRed, colorReset, err)
			os.Exit(1)
		}

		// 验证密码格式
		if len(newPassword) < 6 {
			fmt.Printf("%s[错误]%s 密码至少需要6位\n", colorRed, colorReset)
			os.Exit(1)
		}

		// 强制修改admin密码
		if err := service.ForceChangeAdminPassword(newPassword); err != nil {
			fmt.Printf("%s[错误]%s 修改管理员密码失败: %v\n", colorRed, colorReset, err)
			os.Exit(1)
		}

		fmt.Printf("%s[成功]%s 管理员密码已修改\n", colorGreen, colorReset)
		os.Exit(0)
	}

	// 0. Windows环境检测
	checkTerminal()

	// 启用ANSI颜色支持
	enableANSI()

	baseDir := getBaseDir()

	// 1. 加载配置文件（优先创建配置文件，防止默认端口被占用时无法修改）
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("%s[错误]%s 加载配置文件失败: %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}
	fmt.Printf("%s[配置]%s 配置文件已加载: %s\n", colorGreen, colorReset, config.GetConfigPath())

	// 2. 打印启动Logo
	printBanner()

	// 3. 检查端口是否被占用
	listenAddr := config.GetListenAddr()
	if err := checkPort(listenAddr); err != nil {
		fmt.Printf("%s[错误]%s 端口已被占用，无法启动服务: %s\n", colorRed, colorReset, listenAddr)
		fmt.Printf("%s[错误]%s 请修改配置文件中的端口: %s\n", colorRed, colorReset, config.GetConfigPath())
		fmt.Printf("%s[错误]%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}

	// 4. 初始化数据库
	dbPath := filepath.Join(baseDir, "database")
	if err := models.InitDB(dbPath); err != nil {
		fmt.Printf("%s[错误]%s 数据库初始化失败: %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}
	fmt.Printf("%s[数据库]%s 初始化完成\n", colorGreen, colorReset)

	// 5. 修复缺少宽高的图片记录
	fixed := service.FixMissingDimensions()
	if fixed > 0 {
		fmt.Printf("%s[修复]%s 已修复 %d 张图片的分辨率信息\n", colorGreen, colorReset, fixed)
	}

	// 6. 初始化日志系统
	if err := logger.Init(baseDir); err != nil {
		fmt.Printf("%s[错误]%s 日志系统初始化失败: %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}
	bizLog("系统", fmt.Sprintf("日志系统已启动，日志文件: %s", logger.GetCurrentLogPath()))

	// 7. 创建上传目录
	uploadPath := filepath.Join(baseDir, "uploads")
	os.MkdirAll(uploadPath, 0755)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(customLogger(), gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// API 路由
	r.POST("/api/auth/login", controller.Login)
	r.POST("/api/auth/logout", controller.Logout)
	r.GET("/api/auth/verify", middleware.JWTMiddleware(), controller.VerifyToken)
	r.PUT("/api/profile/password", middleware.JWTMiddleware(), controller.ChangePassword)

	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTMiddleware(), middleware.AdminMiddleware(), middleware.CheckForceChangePassword(), middleware.AdminRateLimit())
	{
		admin.POST("/users", controller.CreateUser)
		admin.GET("/users", controller.GetUsers)
		admin.GET("/users/:id", controller.GetUser)
		admin.PUT("/users/:id", controller.UpdateUser)
		admin.PUT("/users/:id/password", controller.ResetUserPassword)
		admin.DELETE("/users/:id", controller.DeleteUser)

		admin.POST("/tasks", controller.CreateTask)
		admin.GET("/tasks", controller.GetTasks)
		admin.GET("/tasks/:id", controller.GetTask)
		admin.PUT("/tasks/:id", controller.UpdateTask)
		admin.DELETE("/tasks/:id", controller.DeleteTask)

		admin.POST("/upload", controller.AdminUpload)
		admin.GET("/images", controller.GetImages)
		admin.DELETE("/images/:id", controller.DeleteImage)

		admin.GET("/stats", controller.GetDashboardStats)
		admin.GET("/logs", controller.GetLogs)
	}

	user := r.Group("/api/user")
	user.Use(middleware.JWTMiddleware(), middleware.UserMiddleware(), middleware.CheckForceChangePassword())
	{
		user.POST("/upload", controller.UserUpload)
		user.GET("/images", controller.GetImages)
		user.DELETE("/images/:id", controller.DeleteImage)
		user.GET("/stats", controller.GetUserStats)
	}

	guest := r.Group("/api/upload")
	guest.Use(middleware.GuestRateLimit())
	{
		guest.POST("/:code", controller.GuestUpload)
		guest.GET("/:code", controller.CheckTask)
	}

	// 游客链接信息查看（不需要限流）
	r.GET("/api/guest/:code", controller.GetGuestTaskInfo)

	// 图片服务（通过file_code访问）
	r.GET("/i/:code", controller.ServeImage)

	// 前端静态文件
	// 优先使用磁盘上的 frontend 目录（方便更新），否则使用内嵌的前端文件
	frontendDir := filepath.Join(baseDir, "frontend")
	assetsDir := filepath.Join(frontendDir, "assets")

	useDisk := false
	if _, err := os.Stat(filepath.Join(frontendDir, "index.html")); err == nil {
		useDisk = true
	}

	if useDisk {
		r.Static("/assets", assetsDir)
		fmt.Printf("%s[前端]%s 从磁盘加载静态文件\n", colorGreen, colorReset)
	} else {
		// 使用内嵌的前端文件
		subFS, err := fs.Sub(embeddedFrontend, "frontend")
		if err != nil {
			fmt.Printf("%s[警告]%s 内嵌前端文件读取失败\n", colorYellow, colorReset)
		} else {
			assetsSubFS, err := fs.Sub(subFS, "assets")
			if err != nil {
				fmt.Printf("%s[警告]%s 内嵌前端assets目录读取失败\n", colorYellow, colorReset)
			} else {
				r.StaticFS("/assets", http.FS(assetsSubFS))
				fmt.Printf("%s[前端]%s 从内嵌文件加载静态文件\n", colorGreen, colorReset)
			}
		}
	}

	// SPA 路由回退
	r.NoRoute(func(c *gin.Context) {
		// 如果请求的是API路径，返回JSON 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API Not Found"})
			return
		}

		if useDisk {
			path := c.Request.URL.Path
			// 如果请求的是文件（有扩展名），尝试直接提供
			if len(filepath.Ext(path)) > 0 {
				filePath := filepath.Join(frontendDir, path)
				if _, err := os.Stat(filePath); err == nil {
					c.File(filePath)
					return
				}
			}
			// SPA 路由返回 index.html
			c.File(filepath.Join(frontendDir, "index.html"))
		} else {
			// 从内嵌文件提供
			subFS, err := fs.Sub(embeddedFrontend, "frontend")
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Frontend Not Found"})
				return
			}

			path := c.Request.URL.Path
			// 去掉前导 /
			if len(path) > 0 && path[0] == '/' {
				path = path[1:]
			}
			if path == "" {
				path = "index.html"
			}

			// 尝试读取请求的文件
			if data, err := fs.ReadFile(subFS, path); err == nil {
				contentType := "application/octet-stream"
				switch filepath.Ext(path) {
				case ".html":
					contentType = "text/html; charset=utf-8"
				case ".js":
					contentType = "application/javascript; charset=utf-8"
				case ".css":
					contentType = "text/css; charset=utf-8"
				case ".svg":
					contentType = "image/svg+xml"
				case ".png":
					contentType = "image/png"
				case ".ico":
					contentType = "image/x-icon"
				}
				c.Data(http.StatusOK, contentType, data)
				return
			}

			// SPA 路由回退到 index.html
			if data, err := fs.ReadFile(subFS, "index.html"); err == nil {
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}

			c.JSON(http.StatusNotFound, gin.H{"error": "Frontend Not Found"})
		}
	})

	// 输出监听信息
	host := config.AppConfig.Server.Host
	port := config.AppConfig.Server.Port
	fmt.Println()
	fmt.Printf("%s[服务]%s 监听链接: http://%s\n", colorGreen, colorReset, listenAddr)
	if host == "0.0.0.0" || host == "::" {
		fmt.Printf("%s[服务]%s 本地访问: http://127.0.0.1:%d\n", colorGreen, colorReset, port)
		fmt.Printf("%s[服务]%s 局域网访问: http://<本机IP>:%d\n", colorGreen, colorReset, port)
	} else if host == "127.0.0.1" || host == "::1" {
		fmt.Printf("%s[服务]%s 本地访问: http://localhost:%d\n", colorGreen, colorReset, port)
	} else {
		fmt.Printf("%s[服务]%s 访问地址: http://%s:%d\n", colorGreen, colorReset, host, port)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("=", 78))
	fmt.Println()

	// 定期清理token黑名单
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			middleware.CleanupBlacklist()
		}
	}()

	if err := r.Run(listenAddr); err != nil {
		fmt.Printf("%s[错误]%s 服务启动失败: %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}
}
