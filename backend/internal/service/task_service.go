package service

import (
	"imageshare/internal/models"
	"imageshare/internal/repository"
	"time"
)

func CreateTask(maxCount int, expireDays int) (*models.UploadTask, error) {
	code := models.GenerateCode(6)
	expireTime := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)

	task := models.UploadTask{
		Code:       code,
		MaxCount:   maxCount,
		ExpireTime: expireTime,
		Status:     1,
	}

	err := repository.CreateTask(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func GetTaskByCode(code string) (*models.UploadTask, error) {
	return repository.GetTaskByCode(code)
}

func GetTaskByID(id uint) (*models.UploadTask, error) {
	return repository.GetTaskByID(id)
}

func GetAllTasks() ([]models.UploadTask, error) {
	return repository.GetAllTasks()
}

func GetTasksPage(page, pageSize int) ([]models.UploadTask, int64, error) {
	return repository.GetTasksPage(page, pageSize)
}

func GetTasksOffset(offset, limit int) ([]models.UploadTask, int64, error) {
	return repository.GetTasksOffset(offset, limit)
}

func UpdateTask(id uint, maxCount int, expireDays int) error {
	task, err := repository.GetTaskByID(id)
	if err != nil {
		return err
	}

	if maxCount > 0 {
		task.MaxCount = maxCount
	}
	if expireDays > 0 {
		task.ExpireTime = time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)
	}

	return repository.UpdateTask(task)
}

func DeleteTask(id uint, deleteFiles bool) error {
	if deleteFiles {
		task, err := repository.GetTaskByID(id)
		if err == nil {
			DeleteImagesByTaskCode(task.Code)
		}
	}
	return repository.DeleteTask(id)
}

func IncrementTaskUploadCount(taskID uint) error {
	task, err := repository.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	task.UploadedCount++
	return repository.UpdateTask(task)
}

// IsTaskValid 检查游客链接是否可上传（未过期且未满）
func IsTaskValid(code string) (bool, *models.UploadTask) {
	task, err := repository.GetTaskByCode(code)
	if err != nil {
		return false, nil
	}

	// 链接被管理员禁用
	if task.Status != 1 {
		return false, task
	}

	// 已过期
	if task.ExpireTime.Before(time.Now()) {
		return false, task
	}

	// 上传次数已满 - 不删除链接，但不可上传
	if task.UploadedCount >= task.MaxCount {
		return false, task
	}

	return true, task
}

// CanViewTask 检查游客链接是否可查看（只要未被管理员禁用且未过期）
func CanViewTask(code string) (bool, *models.UploadTask) {
	task, err := repository.GetTaskByCode(code)
	if err != nil {
		return false, nil
	}

	if task.Status != 1 {
		return false, task
	}

	if task.ExpireTime.Before(time.Now()) {
		return false, task
	}

	return true, task
}
