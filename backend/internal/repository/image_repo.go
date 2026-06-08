package repository

import (
	"imageshare/internal/models"
)

func CreateImage(image *models.Image) error {
	return models.DB.Create(image).Error
}

func UpdateImage(image *models.Image) error {
	return models.DB.Save(image).Error
}

func DeleteImage(id uint) error {
	return models.DB.Delete(&models.Image{}, id).Error
}

func GetImageByID(id uint) (*models.Image, error) {
	var image models.Image
	err := models.DB.First(&image, id).Error
	return &image, err
}

func GetImageByFileCode(fileCode string) (*models.Image, error) {
	var image models.Image
	err := models.DB.Where("file_code = ?", fileCode).First(&image).Error
	return &image, err
}

func GetImagesByOwner(ownerType string, ownerID uint) ([]models.Image, error) {
	var images []models.Image
	err := models.DB.Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).Order("upload_time DESC").Find(&images).Error
	return images, err
}

func GetImagesByTaskCode(taskCode string) ([]models.Image, error) {
	var images []models.Image
	err := models.DB.Where("task_code = ?", taskCode).Order("upload_time DESC").Find(&images).Error
	return images, err
}

func GetAllImages() ([]models.Image, error) {
	var images []models.Image
	err := models.DB.Order("upload_time DESC").Find(&images).Error
	return images, err
}

// GetImagesPage 分页获取图片
func GetImagesPage(page, pageSize int) ([]models.Image, int64, error) {
	var images []models.Image
	var total int64
	models.DB.Model(&models.Image{}).Count(&total)
	offset := (page - 1) * pageSize
	err := models.DB.Order("upload_time DESC").Offset(offset).Limit(pageSize).Find(&images).Error
	return images, total, err
}

// GetImagesOffset 偏移量获取图片（用于滚动加载）
func GetImagesOffset(offset, limit int) ([]models.Image, int64, error) {
	var images []models.Image
	var total int64
	models.DB.Model(&models.Image{}).Count(&total)
	err := models.DB.Order("upload_time DESC").Offset(offset).Limit(limit).Find(&images).Error
	return images, total, err
}

type StatsResult struct {
	ImageCount int64
	TotalSize  int64
	UserCount  int64
}

func GetTotalStats() (int64, int64, int64, error) {
	var imageCount int64
	var totalSize int64
	var userCount int64

	err := models.DB.Model(&models.Image{}).Count(&imageCount).Error
	if err != nil {
		return 0, 0, 0, err
	}

	err = models.DB.Model(&models.Image{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize).Error
	if err != nil {
		return 0, 0, 0, err
	}

	// 包含管理员在内的用户总数
	err = models.DB.Model(&models.User{}).Count(&userCount).Error
	if err != nil {
		return 0, 0, 0, err
	}

	return imageCount, totalSize, userCount, nil
}

func GetTaskCount() (int64, error) {
	var count int64
	err := models.DB.Model(&models.UploadTask{}).Count(&count).Error
	return count, err
}
