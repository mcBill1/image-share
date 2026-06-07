package models

import (
	"errors"
	"math/rand"
	"regexp"
	"time"

	"gorm.io/gorm"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenerateCode 生成游客链接识别码（6位，自动扩位）
func GenerateCode(minLength int) string {
	length := minLength
	for {
		code := generateRandomString(length)
		if isCodeUnique(code) {
			return code
		}
		length++
		if length > 12 {
			panic("cannot generate unique code after multiple attempts")
		}
	}
}

// GenerateFileCode 生成6位文件识别码（自动扩位，逻辑同游客链接）
func GenerateFileCode() string {
	length := 6
	for {
		code := generateRandomString(length)
		if isFileCodeUnique(code) {
			return code
		}
		length++
		if length > 12 {
			panic("cannot generate unique file code after multiple attempts")
		}
	}
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func isCodeUnique(code string) bool {
	var task UploadTask
	err := DB.Where("code = ?", code).First(&task).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func isFileCodeUnique(code string) bool {
	var image Image
	err := DB.Where("file_code = ?", code).First(&image).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// ValidateUsername 验证用户名：2-10位，仅允许字母数字下划线
func ValidateUsername(username string) bool {
	if len(username) < 2 || len(username) > 10 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return matched
}

// IsUsernameExists 检查用户名是否已存在
func IsUsernameExists(username string) bool {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	return result.Error != gorm.ErrRecordNotFound
}
