package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbDir string) error {
	err := os.MkdirAll(dbDir, 0755)
	if err != nil {
		return err
	}

	dbPath := filepath.Join(dbDir, "sqlite.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	DB = db

	err = db.AutoMigrate(&User{}, &UploadTask{}, &Image{})
	if err != nil {
		return err
	}

	err = createDefaultAdmin()
	if err != nil {
		return err
	}

	return nil
}

func createDefaultAdmin() error {
	var admin User
	result := DB.Where("username = ?", "admin").First(&admin)
	if result.Error == gorm.ErrRecordNotFound {
		// 前端登录发送MD5(原始密码)，后端存储bcrypt(MD5(原始密码))
		defaultPasswordMD5 := md5Hash("image123456")
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(defaultPasswordMD5), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		admin := User{
			Username:            "admin",
			PasswordHash:        string(passwordHash),
			Role:                "admin",
			StorageLimitMB:      10240,
			ImageLimit:          1000,
			SingleImageLimitMB:  20,
			ForceChangePassword: 1,
		}

		err = DB.Create(&admin).Error
		if err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}
	}
	return nil
}

func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
