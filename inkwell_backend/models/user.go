package models

import (
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	// UserID 是雪花算法生成的 64 位 ID, 超过了 JS Number 的安全范围(2^53-1),
	// 因此序列化成字符串, 避免前端 JSON.parse 后精度丢失。
	UserID   uint64 `json:"user_id,string" db:"user_id"`
	UserName string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

// UserProfile 用户主页对外展示的信息(只有公开字段, 不含密码等敏感数据)
type UserProfile struct {
	UserID     uint64    `json:"user_id,string" db:"user_id"`
	UserName   string    `json:"username" db:"username"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
}

func (u *User) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		UserName string `json:"username" db:"username"`
		Password string `json:"password" db:"password"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if len(required.UserName) == 0 {
		err = errors.New("缺少必填字段username")
	} else if len(required.Password) == 0 {
		err = errors.New("缺少必填字段password")
	} else {
		u.UserName = required.UserName
		u.Password = required.Password
	}
	return
}

// RegisterForm 注册请求参数
type RegisterForm struct {
	UserName        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	// RePassword 兼容旧版前端使用的 re_password 字段
	RePassword string `json:"re_password"`
}

func (r *RegisterForm) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		UserName        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		RePassword      string `json:"re_password"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	}
	// 兼容 confirm_password / re_password 两种字段名
	if required.ConfirmPassword == "" {
		required.ConfirmPassword = required.RePassword
	}
	switch {
	case len(required.UserName) == 0:
		err = errors.New("缺少必填字段username")
	case len(required.Password) == 0:
		err = errors.New("缺少必填字段password")
	case required.Password != required.ConfirmPassword:
		err = errors.New("两次密码不一致")
	default:
		r.UserName = required.UserName
		r.Password = required.Password
		r.ConfirmPassword = required.ConfirmPassword
	}
	return
}
