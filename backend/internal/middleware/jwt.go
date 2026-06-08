package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"imageshare/config"
	"imageshare/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Token黑名单
var (
	tokenBlacklist     = make(map[string]time.Time)
	tokenBlacklistLock sync.RWMutex
)

// BlacklistToken 将token加入黑名单
func BlacklistToken(tokenString string) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return
	}
	tokenBlacklistLock.Lock()
	defer tokenBlacklistLock.Unlock()
	tokenBlacklist[tokenString] = claims.ExpiresAt.Time
}

// isTokenBlacklisted 检查token是否在黑名单中
func isTokenBlacklisted(tokenString string) bool {
	tokenBlacklistLock.RLock()
	defer tokenBlacklistLock.RUnlock()
	_, exists := tokenBlacklist[tokenString]
	return exists
}

// CleanupBlacklist 定期清理过期的黑名单条目
func CleanupBlacklist() {
	tokenBlacklistLock.Lock()
	defer tokenBlacklistLock.Unlock()
	now := time.Now()
	for token, expiry := range tokenBlacklist {
		if now.After(expiry) {
			delete(tokenBlacklist, token)
		}
	}
}

func GenerateToken(userID uint, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AppConfig.TokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 优先从 Authorization header 获取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 如果 header 没有，从 cookie 获取
		if tokenString == "" {
			if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		// 检查token是否在黑名单中
		if isTokenBlacklisted(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("token_string", tokenString)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || (role != "admin" && role != "user") {
			c.JSON(http.StatusForbidden, gin.H{"error": "User access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckForceChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		var user models.User
		if err := models.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			c.Abort()
			return
		}

		if user.ForceChangePassword == 1 && c.Request.URL.Path != "/api/profile/password" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Please change your password first", "redirect": "/profile/change-password"})
			c.Abort()
			return
		}
		c.Next()
	}
}
