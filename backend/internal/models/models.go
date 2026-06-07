package models

import "time"

type User struct {
	ID                  uint      `gorm:"primary_key;auto_increment" json:"id"`
	Username            string    `gorm:"unique;not null" json:"username"`
	PasswordHash        string    `gorm:"not null" json:"-"`
	Role                string    `gorm:"not null" json:"role"`
	StorageLimitMB      int       `gorm:"default:100" json:"storage_limit_mb"`
	ImageLimit          int       `gorm:"default:50" json:"image_limit"`
	SingleImageLimitMB  int       `gorm:"default:10" json:"single_image_limit_mb"`
	ForceChangePassword int       `gorm:"default:1" json:"force_change_password"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type UploadTask struct {
	ID            uint      `gorm:"primary_key;auto_increment" json:"id"`
	Code          string    `gorm:"unique;not null" json:"code"`
	MaxCount      int       `gorm:"default:5" json:"max_count"`
	UploadedCount int       `gorm:"default:0" json:"uploaded_count"`
	ExpireTime    time.Time `json:"expire_time"`
	Status        int       `gorm:"default:1" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type Image struct {
	ID           uint      `gorm:"primary_key;auto_increment" json:"id"`
	OwnerType    string    `gorm:"not null" json:"owner_type"`
	OwnerID      uint      `json:"owner_id"`
	TaskCode     string    `gorm:"default:''" json:"task_code"`
	FileCode     string    `gorm:"default:''" json:"file_code"`
	OriginalName string    `json:"original_name"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	StoragePath  string    `json:"storage_path"`
	PublicURL    string    `json:"public_url"`
	UploadTime   time.Time `json:"upload_time"`
	OwnerName    string    `gorm:"-" json:"owner_name,omitempty"`
}

func (User) TableName() string {
	return "users"
}

func (UploadTask) TableName() string {
	return "upload_tasks"
}

func (Image) TableName() string {
	return "images"
}
