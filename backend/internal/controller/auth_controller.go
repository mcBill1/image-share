package controller

import (
	"net/http"
	"strings"
	"time"

	"imageshare/config"
	"imageshare/internal/logger"
	"imageshare/internal/middleware"
	"imageshare/internal/service"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, userID, role, err := service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 获取用户信息以返回 force_change_password
	user, _ := service.GetUser(userID)

	// 设置 HttpOnly cookie，7天有效期
	maxAge := int(config.AppConfig.TokenExpire.Seconds())
	c.SetCookie("auth_token", token, maxAge, "/", "", false, true)

	// 业务日志
	ip := c.ClientIP()
	roleText := "用户"
	if role == "admin" {
		roleText = "管理员"
	}
	logger.WriteLog("\033[90m[%s]\033[0m \033[32m[登录]\033[0m %s %s 登录成功 (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), roleText, req.Username, ip)

	c.JSON(http.StatusOK, gin.H{
		"token":                 token,
		"user_id":               userID,
		"role":                  role,
		"force_change_password": user.ForceChangePassword,
	})
}

func Logout(c *gin.Context) {
	// 从cookie或header获取token并加入黑名单
	var tokenString string

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}
	if tokenString == "" {
		if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
			tokenString = cookie
		}
	}

	if tokenString != "" {
		middleware.BlacklistToken(tokenString)
	}

	// 清除cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// VerifyToken 验证当前token是否有效，返回用户信息
func VerifyToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	user, err := service.GetUser(userID.(uint))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":               userID,
		"role":                  role,
		"force_change_password": user.ForceChangePassword,
	})
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := service.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := service.GetUser(userID.(uint))
	logger.WriteLog("\033[90m[%s]\033[0m \033[33m[修改]\033[0m 用户 %s 修改密码成功 (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), user.Username, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
