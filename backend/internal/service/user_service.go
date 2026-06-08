package service

import (
	"fmt"
	"imageshare/config"
	"imageshare/internal/models"
	"imageshare/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(username, password string, storageLimitMB, imageLimit, singleImageLimitMB int) error {
	// 验证用户名
	if !models.ValidateUsername(username) {
		return fmt.Errorf("用户名必须为2-10位，仅允许字母、数字和下划线")
	}
	if models.IsUsernameExists(username) {
		return fmt.Errorf("用户名已存在")
	}

	// 验证密码
	if err := ValidatePassword(password); err != nil {
		return err
	}

	// 前端已MD5，直接bcrypt存储
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Username:            username,
		PasswordHash:        string(hash),
		Role:                "user",
		StorageLimitMB:      storageLimitMB,
		ImageLimit:          imageLimit,
		SingleImageLimitMB:  singleImageLimitMB,
		ForceChangePassword: 1,
	}

	return repository.CreateUser(&user)
}

func GetUser(id uint) (*models.User, error) {
	return repository.GetUserByID(id)
}

func GetAllUsers() ([]models.User, error) {
	return repository.GetAllUsers()
}

func GetUsersPage(page, pageSize int) ([]models.User, int64, error) {
	return repository.GetUsersPage(page, pageSize)
}

func GetUsersOffset(offset, limit int) ([]models.User, int64, error) {
	return repository.GetUsersOffset(offset, limit)
}

func UpdateUser(id uint, username string, storageLimitMB, imageLimit, singleImageLimitMB int) error {
	user, err := repository.GetUserByID(id)
	if err != nil {
		return err
	}

	if username != "" && username != user.Username {
		if !models.ValidateUsername(username) {
			return fmt.Errorf("用户名必须为2-10位，仅允许字母、数字和下划线")
		}
		if models.IsUsernameExists(username) {
			return fmt.Errorf("用户名已存在")
		}
		user.Username = username
	}

	// 始终更新限制值（0表示无限制）
	user.StorageLimitMB = storageLimitMB
	user.ImageLimit = imageLimit
	user.SingleImageLimitMB = singleImageLimitMB

	return repository.UpdateUser(user)
}

func DeleteUser(id uint, deleteFiles bool) error {
	if deleteFiles {
		DeleteImagesByUser(id)
	}
	return repository.DeleteUser(id)
}

func GetUserStats(userID uint) (int64, int64, error) {
	return repository.GetUserStats(userID)
}

func CheckUserQuota(userID uint, fileSize int64) (bool, string) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return false, "User not found"
	}

	usedSize, usedCount, err := repository.GetUserStats(userID)
	if err != nil {
		return false, "Failed to get user stats"
	}

	maxSize := int64(user.StorageLimitMB) * 1024 * 1024
	if usedSize+fileSize > maxSize {
		return false, "Storage limit exceeded"
	}

	if usedCount >= int64(user.ImageLimit) {
		return false, "Image limit exceeded"
	}

	maxSingleSize := int64(user.SingleImageLimitMB) * 1024 * 1024
	if fileSize > maxSingleSize {
		return false, "Single image size exceeded"
	}

	return true, ""
}

func CheckAdminQuota(fileSize int64) (bool, string) {
	if fileSize > config.AppConfig.AdminMaxSize {
		return false, "File size exceeds admin limit (20MB)"
	}
	return true, ""
}

// ValidatePassword 验证密码（前端已MD5，后端只验证非空）
func ValidatePassword(password string) error {
	if len(password) == 0 {
		return fmt.Errorf("密码不能为空")
	}
	return nil
}

// ResetUserPassword 管理员重置用户密码
func ResetUserPassword(userID uint, newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 前端已MD5，直接bcrypt存储
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.ForceChangePassword = 1

	return repository.UpdateUser(user)
}
