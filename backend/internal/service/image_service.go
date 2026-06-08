package service

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "golang.org/x/image/webp"

	"imageshare/config"
	"imageshare/internal/models"
	"imageshare/internal/repository"
)

// 上传互斥锁，防止并发上传绕过配额检查
var (
	userUploadLocks = make(map[uint]*sync.Mutex)
	userLockMapLock sync.Mutex
	taskUploadLocks = make(map[uint]*sync.Mutex)
	taskLockMapLock sync.Mutex
)

func GetUserUploadLock(userID uint) *sync.Mutex {
	userLockMapLock.Lock()
	defer userLockMapLock.Unlock()
	if lock, ok := userUploadLocks[userID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	userUploadLocks[userID] = lock
	return lock
}

func GetTaskUploadLock(taskID uint) *sync.Mutex {
	taskLockMapLock.Lock()
	defer taskLockMapLock.Unlock()
	if lock, ok := taskUploadLocks[taskID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	taskUploadLocks[taskID] = lock
	return lock
}

func ValidateImage(file *multipart.FileHeader) (bool, string) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !config.AllowedExtensions[ext] {
		return false, "File type not allowed"
	}

	if file.Size == 0 {
		return false, "File is empty"
	}

	return true, ""
}

func GetImageDimensions(file multipart.File) (int, int, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, err
	}
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}

func SaveImage(file multipart.File, fileHeader *multipart.FileHeader, ownerType string, ownerID uint, ownerName string, taskCode string) (*models.Image, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileCode := models.GenerateFileCode()
	fileName := fileCode + ext

	var storagePath string
	switch ownerType {
	case "admin":
		storagePath = filepath.Join(config.GetUploadPath(), "admin", fileName)
	case "user":
		storagePath = filepath.Join(config.GetUploadPath(), "user", ownerName, fileName)
	case "guest":
		storagePath = filepath.Join(config.GetUploadPath(), "guest", taskCode, fileName)
	default:
		return nil, fmt.Errorf("invalid owner type")
	}

	err := os.MkdirAll(filepath.Dir(storagePath), 0755)
	if err != nil {
		return nil, err
	}

	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(storagePath)
		return nil, err
	}
	dst.Close()

	// 从已保存的文件读取尺寸，确保可靠
	width, height := 0, 0
	f, err := os.Open(storagePath)
	if err == nil {
		img, _, imgErr := image.Decode(f)
		f.Close()
		if imgErr == nil {
			bounds := img.Bounds()
			width = bounds.Dx()
			height = bounds.Dy()
		}
	}

	image := models.Image{
		OwnerType:    ownerType,
		OwnerID:      ownerID,
		TaskCode:     taskCode,
		OriginalName: fileHeader.Filename,
		FileName:     fileName,
		FileCode:     fileCode,
		FileSize:     fileHeader.Size,
		Width:        width,
		Height:       height,
		StoragePath:  storagePath,
		PublicURL:    "",
		UploadTime:   time.Now(),
	}

	err = repository.CreateImage(&image)
	if err != nil {
		return nil, err
	}

	image.PublicURL = fmt.Sprintf("/i/%s", image.FileCode)
	err = repository.UpdateImage(&image)
	if err != nil {
		return nil, err
	}

	return &image, nil
}

func DeleteImage(id uint) error {
	image, err := repository.GetImageByID(id)
	if err != nil {
		return err
	}

	os.Remove(image.StoragePath)

	return repository.DeleteImage(id)
}

// DeleteImagesByTaskCode 删除指定游客链接的所有图片文件
func DeleteImagesByTaskCode(taskCode string) (int, error) {
	images, err := repository.GetImagesByTaskCode(taskCode)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, img := range images {
		os.Remove(img.StoragePath)
		repository.DeleteImage(img.ID)
		count++
	}

	// 尝试删除目录
	dir := filepath.Join(config.GetUploadPath(), "guest", taskCode)
	os.Remove(dir)

	return count, nil
}

// DeleteImagesByUser 删除指定用户的所有图片文件和数据库记录
func DeleteImagesByUser(userID uint) (int, error) {
	images, err := repository.GetImagesByOwner("user", userID)
	if err != nil {
		return 0, err
	}

	count := 0
	var dirPath string
	for _, img := range images {
		os.Remove(img.StoragePath)
		repository.DeleteImage(img.ID)
		count++
		// 从第一个图片路径提取用户目录
		if dirPath == "" {
			dirPath = filepath.Dir(img.StoragePath)
		}
	}

	// 尝试删除用户目录
	if dirPath != "" {
		os.RemoveAll(dirPath)
	}

	return count, nil
}

func GetImage(id uint) (*models.Image, error) {
	return repository.GetImageByID(id)
}

func GetImageByFileCode(fileCode string) (*models.Image, error) {
	return repository.GetImageByFileCode(fileCode)
}

func GetImagesByOwner(ownerType string, ownerID uint) ([]models.Image, error) {
	return repository.GetImagesByOwner(ownerType, ownerID)
}

func GetImagesByTaskCode(taskCode string) ([]models.Image, error) {
	return repository.GetImagesByTaskCode(taskCode)
}

func GetAllImages() ([]models.Image, error) {
	return repository.GetAllImages()
}

func GetImagesPage(page, pageSize int) ([]models.Image, int64, error) {
	images, total, err := repository.GetImagesPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	FillOwnerNames(images)
	return images, total, nil
}

func GetImagesOffset(offset, limit int) ([]models.Image, int64, error) {
	images, total, err := repository.GetImagesOffset(offset, limit)
	if err != nil {
		return nil, 0, err
	}
	FillOwnerNames(images)
	return images, total, nil
}

// FillOwnerNames 为图片列表填充所有者用户名
func FillOwnerNames(images []models.Image) {
	// 收集需要查询的owner_id
	userIDs := make(map[uint]bool)
	for _, img := range images {
		if img.OwnerType == "user" && img.OwnerID > 0 {
			userIDs[img.OwnerID] = true
		}
	}

	// 批量查询用户名
	userNames := make(map[uint]string)
	for uid := range userIDs {
		user, err := repository.GetUserByID(uid)
		if err == nil {
			userNames[uid] = user.Username
		}
	}

	// 填充
	for i := range images {
		if images[i].OwnerType == "user" {
			if name, ok := userNames[images[i].OwnerID]; ok {
				images[i].OwnerName = name
			}
		}
	}
}

func GetDashboardStats() (map[string]interface{}, error) {
	imageCount, totalSize, userCount, err := repository.GetTotalStats()
	if err != nil {
		return nil, err
	}

	taskCount, err := repository.GetTaskCount()
	if err != nil {
		return nil, err
	}

	// 获取磁盘空间信息
	var diskTotal, diskFree uint64
	// 获取可执行文件所在目录的磁盘信息
	exePath, err := os.Executable()
	if err == nil {
		var rootPath string
		// Windows: 获取盘符根目录
		if len(exePath) >= 2 && exePath[1] == ':' {
			rootPath = exePath[:3] // e.g. "C:\"
		} else {
			rootPath = "/"
		}
		type StatFS struct {
			Total uint64
			Free  uint64
		}
		var stat StatFS
		stat.Total, stat.Free, _ = getDiskUsage(rootPath)
		diskTotal = stat.Total
		diskFree = stat.Free
	}

	return map[string]interface{}{
		"image_count":  imageCount,
		"user_count":   userCount,
		"task_count":   taskCount,
		"storage_used": totalSize,
		"disk_total":   diskTotal,
		"disk_free":    diskFree,
	}, nil
}

// FixMissingDimensions 修复数据库中缺少宽高的图片记录
func FixMissingDimensions() int {
	var images []models.Image
	models.DB.Where("width = 0 OR height = 0").Find(&images)

	fixed := 0
	for _, img := range images {
		f, err := os.Open(img.StoragePath)
		if err != nil {
			continue
		}
		decoded, _, imgErr := image.Decode(f)
		f.Close()
		if imgErr != nil {
			continue
		}
		bounds := decoded.Bounds()
		img.Width = bounds.Dx()
		img.Height = bounds.Dy()
		if img.Width > 0 && img.Height > 0 {
			models.DB.Model(&img).Updates(map[string]interface{}{"width": img.Width, "height": img.Height})
			fixed++
		}
	}
	return fixed
}
