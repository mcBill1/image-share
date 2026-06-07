package controller

import (
	"net/http"
	"strconv"
	"time"

	"imageshare/internal/logger"
	"imageshare/internal/service"

	"github.com/gin-gonic/gin"
)

type CreateTaskRequest struct {
	MaxCount   int `json:"max_count"`
	ExpireDays int `json:"expire_days"`
}

type UpdateTaskRequest struct {
	MaxCount   int `json:"max_count"`
	ExpireDays int `json:"expire_days"`
}

type DeleteTaskRequest struct {
	DeleteFiles bool `json:"delete_files"`
}

func CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxCount := req.MaxCount
	if maxCount <= 0 {
		maxCount = 5
	}

	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 7
	}

	task, err := service.CreateTask(maxCount, expireDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[33m[创建]\033[0m 管理员创建游客链接 %s (最多%d张, %d天有效) (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), task.Code, maxCount, expireDays, c.ClientIP())

	c.JSON(http.StatusCreated, task)
}

func GetTasks(c *gin.Context) {
	// 支持分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "0"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	if page > 0 && pageSize > 0 {
		tasks, total, err := service.GetTasksPage(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": tasks, "total": total})
	} else if limit > 0 {
		tasks, total, err := service.GetTasksOffset(offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": tasks, "total": total})
	} else {
		tasks, err := service.GetAllTasks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tasks)
	}
}

func GetTask(c *gin.Context) {
	code := c.Param("code")
	task, err := service.GetTaskByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.UpdateTask(uint(id), req.MaxCount, req.ExpireDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// 检查是否需要同时删除文件
	deleteFiles := c.Query("delete_files") == "true"

	err = service.DeleteTask(uint(id), deleteFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WriteLog("\033[90m[%s]\033[0m \033[31m[删除]\033[0m 管理员删除游客链接 ID:%d (IP: %s)\n",
		time.Now().Format("2006/01/02 15:04:05"), id, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func CheckTask(c *gin.Context) {
	code := c.Param("code")
	valid, _ := service.IsTaskValid(code)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"message": "链接已失效"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "链接有效"})
}
