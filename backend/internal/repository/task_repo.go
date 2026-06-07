package repository

import (
	"time"

	"imageshare/internal/models"
)

func GetTaskByCode(code string) (*models.UploadTask, error) {
	var task models.UploadTask
	err := models.DB.Where("code = ?", code).First(&task).Error
	return &task, err
}

func GetTaskByID(id uint) (*models.UploadTask, error) {
	var task models.UploadTask
	err := models.DB.First(&task, id).Error
	return &task, err
}

func CreateTask(task *models.UploadTask) error {
	return models.DB.Create(task).Error
}

func UpdateTask(task *models.UploadTask) error {
	return models.DB.Save(task).Error
}

func DeleteTask(id uint) error {
	return models.DB.Delete(&models.UploadTask{}, id).Error
}

func GetAllTasks() ([]models.UploadTask, error) {
	var tasks []models.UploadTask
	err := models.DB.Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetTasksPage 分页获取游客任务
func GetTasksPage(page, pageSize int) ([]models.UploadTask, int64, error) {
	var tasks []models.UploadTask
	var total int64
	models.DB.Model(&models.UploadTask{}).Count(&total)
	offset := (page - 1) * pageSize
	err := models.DB.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

// GetTasksOffset 偏移量获取游客任务（用于滚动加载）
func GetTasksOffset(offset, limit int) ([]models.UploadTask, int64, error) {
	var tasks []models.UploadTask
	var total int64
	models.DB.Model(&models.UploadTask{}).Count(&total)
	err := models.DB.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error
	return tasks, total, err
}

// IsTaskValid 检查链接是否有效（仅检查状态和过期时间，不检查上传次数）
func IsTaskValid(task *models.UploadTask) bool {
	if task.Status != 1 {
		return false
	}
	return task.ExpireTime.After(time.Now())
}
