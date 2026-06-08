package controller

import (
	"net/http"
	"strconv"
	"time"

	"imageshare/internal/logger"
	"imageshare/internal/service"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Username           string `json:"username" binding:"required"`
	Password           string `json:"password" binding:"required"`
	StorageLimitMB     int    `json:"storage_limit_mb"`
	ImageLimit         int    `json:"image_limit"`
	SingleImageLimitMB int    `json:"single_image_limit_mb"`
}

type UpdateUserRequest struct {
	Username           string `json:"username"`
	StorageLimitMB     int    `json:"storage_limit_mb"`
	ImageLimit         int    `json:"image_limit"`
	SingleImageLimitMB int    `json:"single_image_limit_mb"`
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := service.CreateUser(req.Username, req.Password, req.StorageLimitMB, req.ImageLimit, req.SingleImageLimitMB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[33m[创建]\033[0m 管理员创建用户 %s (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), req.Username, c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func GetUsers(c *gin.Context) {
	// 支持分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "0"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	if page > 0 && pageSize > 0 {
		users, total, err := service.GetUsersPage(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": users, "total": total})
	} else if limit > 0 {
		users, total, err := service.GetUsersOffset(offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": users, "total": total})
	} else {
		users, err := service.GetAllUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := service.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.UpdateUser(uint(id), req.Username, req.StorageLimitMB, req.ImageLimit, req.SingleImageLimitMB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	deleteFiles := c.Query("delete_files") == "true"

	err = service.DeleteUser(uint(id), deleteFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[31m[删除]\033[0m 管理员删除用户 ID:%d (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), id, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func GetUserStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	usedSize, usedCount, err := service.GetUserStats(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取用户限制信息
	user, err := service.GetUser(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"used_size":          usedSize,
		"used_count":         usedCount,
		"storage_limit":      int64(user.StorageLimitMB) * 1024 * 1024,
		"image_limit":        user.ImageLimit,
		"single_image_limit": int64(user.SingleImageLimitMB) * 1024 * 1024,
	})
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

func ResetUserPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.ResetUserPassword(uint(id), req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[33m[修改]\033[0m 管理员重置用户 ID:%d 密码 (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), id, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
