package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"imageshare/internal/logger"
	"imageshare/internal/models"
	"imageshare/internal/service"

	"github.com/gin-gonic/gin"
)

func AdminUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	valid, msg := service.ValidateImage(file)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	valid, msg = service.CheckAdminQuota(file.Size)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	image, err := service.SaveImage(fileContent, file, "admin", 0, "admin", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[36m[上传]\033[0m 管理员上传图片 %s (%s) (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), image.OriginalName, formatSize(image.FileSize), c.ClientIP())

	c.JSON(http.StatusCreated, image)
}

func UserUpload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uid := userID.(uint)

	// 获取用户名用于存储路径
	user, err := service.GetUser(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	valid, msg := service.ValidateImage(file)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// 加锁：配额检查和保存必须原子操作，防止并发上传绕过
	lock := service.GetUserUploadLock(uid)
	lock.Lock()
	defer lock.Unlock()

	// 在锁内检查配额，确保读取最新数据
	valid, msg = service.CheckUserQuota(uid, file.Size)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	image, err := service.SaveImage(fileContent, file, "user", uid, user.Username, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[36m[上传]\033[0m 用户 %s 上传图片 %s (%s) (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), user.Username, image.OriginalName, formatSize(image.FileSize), c.ClientIP())

	c.JSON(http.StatusCreated, image)
}

func GuestUpload(c *gin.Context) {
	code := c.Param("code")

	// 先获取task以拿到ID用于加锁
	task, err := service.GetTaskByCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "链接已失效"})
		return
	}

	// 加锁：防止并发上传绕过游客链接次数限制
	lock := service.GetTaskUploadLock(task.ID)
	lock.Lock()
	defer lock.Unlock()

	// 在锁内重新从数据库获取最新task状态，防止并发读取旧数据
	task, err = service.GetTaskByCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "链接已失效"})
		return
	}

	if task.Status != 1 || task.ExpireTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "链接已失效"})
		return
	}
	if task.UploadedCount >= task.MaxCount {
		c.JSON(http.StatusBadRequest, gin.H{"message": "上传次数已满"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	valid, msg := service.ValidateImage(file)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	image, err := service.SaveImage(fileContent, file, "guest", 0, "", task.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = service.IncrementTaskUploadCount(task.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[36m[上传]\033[0m 游客 %s 上传图片 %s (%s) (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), task.Code, image.OriginalName, formatSize(image.FileSize), c.ClientIP())

	c.JSON(http.StatusCreated, image)
}

// GetGuestTaskInfo 获取游客链接信息（可查看，即使上传已满）
func GetGuestTaskInfo(c *gin.Context) {
	code := c.Param("code")
	canView, task := service.CanViewTask(code)
	if !canView {
		c.JSON(http.StatusBadRequest, gin.H{"message": "链接已失效"})
		return
	}

	// 获取该链接下的图片
	images, err := service.GetImagesByTaskCode(task.Code)
	var imageList interface{}
	if err != nil {
		imageList = []models.Image{}
	} else {
		imageList = images
	}

	c.JSON(http.StatusOK, gin.H{
		"task":   task,
		"images": imageList,
	})
}

func GetImages(c *gin.Context) {
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")

	if role == "admin" {
		// 支持分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "0"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

		if page > 0 && pageSize > 0 {
			// 分页模式
			images, total, err := service.GetImagesPage(page, pageSize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": images, "total": total})
		} else if limit > 0 {
			// 滚动加载模式
			images, total, err := service.GetImagesOffset(offset, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": images, "total": total})
		} else {
			// 兼容旧模式：一次性加载全部
			imgs, err := service.GetAllImages()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			service.FillOwnerNames(imgs)
			c.JSON(http.StatusOK, imgs)
		}
	} else if role == "user" {
		imgs, err := service.GetImagesByOwner("user", userID.(uint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, imgs)
	} else {
		c.JSON(http.StatusOK, []interface{}{})
	}
}

func DeleteImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	err = service.DeleteImage(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	role, _ := c.Get("role")
	roleText := "管理员"
	if role == "user" {
		roleText = "用户"
	}
	logger.WriteLog("\033[90m[%s]\033[0m \033[31m[删除]\033[0m %s 删除图片 ID:%d (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), roleText, id, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

func ServeImage(c *gin.Context) {
	code := c.Param("code")

	image, err := service.GetImageByFileCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	file, err := os.Open(image.StoragePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image file not found"})
		return
	}
	defer file.Close()

	c.Header("Content-Type", "image/jpeg")
	c.File(image.StoragePath)
}

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const k = 1024
	sizes := []string{"B", "KB", "MB", "GB"}
	i := 0
	fb := float64(bytes)
	for fb >= k && i < len(sizes)-1 {
		fb /= k
		i++
	}
	return fmt.Sprintf("%.1f %s", fb, sizes[i])
}

func GetDashboardStats(c *gin.Context) {
	stats, err := service.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetLogs 获取当前操作日志（仅操作日志，过滤网络请求）
func GetLogs(c *gin.Context) {
	lines := 200
	if l, err := strconv.Atoi(c.DefaultQuery("lines", "200")); err == nil && l > 0 && l <= 1000 {
		lines = l
	}

	content := logger.ReadCurrentLogOps(lines)
	logPath := logger.GetCurrentLogPath()
	fileName := ""
	if logPath != "" {
		fileName = filepath.Base(logPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"content":   content,
		"file_name": fileName,
		"lines":     lines,
	})
}
