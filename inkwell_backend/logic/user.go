package logic

import (
	"inkwell_backend/dao/mysql"
	"inkwell_backend/models"
	"errors"
	"unicode/utf8"
)

// 用户名/密码长度限制, 用户名上限与 user.username 列长度保持一致
const (
	UserNameMinLen = 2
	UserNameMaxLen = 64
	PasswordMinLen = 6
	PasswordMaxLen = 64
)

// UserPasswordValid 校验用户名和密码是否合法
func UserPasswordValid(userName, password string) error {
	if l := utf8.RuneCountInString(userName); l < UserNameMinLen || l > UserNameMaxLen {
		return ErrorUserNameInvalid
	}
	// 密码按字节长度校验(密码是 ASCII 为主的不可见字符)
	if l := len(password); l < PasswordMinLen || l > PasswordMaxLen {
		return ErrorPasswordInvalid
	}
	return nil
}

// Register 校验参数并注册用户
func Register(user *models.User) error {
	if err := UserPasswordValid(user.UserName, user.Password); err != nil {
		return err
	}
	if err := mysql.Register(user); err != nil {
		if errors.Is(err, mysql.ErrorUserExit) {
			return ErrorUserExist
		}
		return err
	}
	return nil
}

// Login 校验用户名密码
func Login(user *models.User) error {
	return mysql.Login(user)
}

// GetUserProfile 查询用户主页信息(只有公开字段)
func GetUserProfile(userID uint64) (*models.UserProfile, error) {
	profile, err := mysql.GetUserProfileByID(userID)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExit) {
			return nil, ErrorUserNotExist
		}
		return nil, err
	}
	return profile, nil
}
