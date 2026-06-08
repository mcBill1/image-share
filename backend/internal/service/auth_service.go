package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"imageshare/internal/middleware"
	"imageshare/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// md5Hash 计算字符串的MD5哈希
func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func Login(username, password string) (string, uint, string, error) {
	user, err := repository.GetUserByUsername(username)
	if err != nil {
		return "", 0, "", err
	}

	// 前端已MD5，直接bcrypt验证
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", 0, "", fmt.Errorf("invalid credentials")
	}

	token, err := middleware.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", 0, "", err
	}

	return token, user.ID, user.Role, nil
}

func ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码（前端已MD5，直接bcrypt验证）
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("旧密码不正确")
	}

	// 检查新密码不能与旧密码相同
	if oldPassword == newPassword {
		return fmt.Errorf("新密码不能与旧密码相同")
	}

	// 验证新密码格式
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	// 前端已MD5，直接bcrypt存储
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.ForceChangePassword = 0

	return repository.UpdateUser(user)
}

// ForceChangeAdminPassword 强制修改管理员密码（命令行使用，明文输入）
func ForceChangeAdminPassword(newPassword string) error {
	user, err := repository.GetUserByUsername("admin")
	if err != nil {
		return fmt.Errorf("管理员账户不存在")
	}

	// 命令行输入的是明文，需要先MD5再bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(md5Hash(newPassword)), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.ForceChangePassword = 0

	return repository.UpdateUser(user)
}
