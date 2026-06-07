package repository

import (
	"imageshare/internal/models"

	"gorm.io/gorm"
)

func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := models.DB.First(&user, id).Error
	return &user, err
}

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := models.DB.Where("username = ?", username).First(&user).Error
	return &user, err
}

func CreateUser(user *models.User) error {
	return models.DB.Create(user).Error
}

func UpdateUser(user *models.User) error {
	return models.DB.Save(user).Error
}

func DeleteUser(id uint) error {
	return models.DB.Delete(&models.User{}, id).Error
}

// GetAllUsers 获取所有用户（包含管理员）
func GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := models.DB.Order("role ASC, created_at ASC").Find(&users).Error
	return users, err
}

// GetUsersPage 分页获取用户
func GetUsersPage(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	models.DB.Model(&models.User{}).Count(&total)
	offset := (page - 1) * pageSize
	err := models.DB.Order("role ASC, created_at ASC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// GetUsersOffset 偏移量获取用户（用于滚动加载）
func GetUsersOffset(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	models.DB.Model(&models.User{}).Count(&total)
	err := models.DB.Order("role ASC, created_at ASC").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

type UserStatsResult struct {
	TotalSize int64
	Count     int64
}

func GetUserStats(userID uint) (int64, int64, error) {
	var result UserStatsResult
	err := models.DB.Model(&models.Image{}).
		Where("owner_type = ? AND owner_id = ?", "user", userID).
		Select("COALESCE(SUM(file_size), 0) as total_size, COUNT(*) as count").
		Scan(&result).Error
	return result.TotalSize, result.Count, err
}

func IsUsernameExists(username string) bool {
	var user models.User
	result := models.DB.Where("username = ?", username).First(&user)
	return result.Error != gorm.ErrRecordNotFound
}
