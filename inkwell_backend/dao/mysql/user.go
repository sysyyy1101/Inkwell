package mysql

import (
	"inkwell_backend/models"
	"inkwell_backend/pkg/snowflake"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost bcrypt 的计算强度, DefaultCost(10) 在安全性和登录耗时之间比较平衡
const bcryptCost = bcrypt.DefaultCost

// encryptPassword 生成 bcrypt 哈希。
// bcrypt 会自动生成随机盐并把盐写进结果(60 字节, user.password 列是 varchar(64)),
// 同一个密码每次生成的哈希都不同, 校验时用 bcrypt.CompareHashAndPassword。
func encryptPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// isDuplicateEntry 判断错误是否为唯一索引冲突(MySQL 错误码 1062)
func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

// Register 注册新用户, 用户名已存在时返回 ErrorUserExit
func Register(user *models.User) (err error) {
	sqlStr := "select count(user_id) from user where username = ?"
	var count int64
	err = db.Get(&count, sqlStr, user.UserName)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if count > 0 {
		// 用户已存在
		return ErrorUserExit
	}
	// 生成user_id
	userID, err := snowflake.GetID()
	if err != nil {
		return ErrorGenIDFailed
	}
	// 生成加密密码
	password, err := encryptPassword(user.Password)
	if err != nil {
		return err
	}
	// 把用户插入数据库
	sqlStr = "insert into user(user_id, username, password) values (?,?,?)"
	_, err = db.Exec(sqlStr, userID, user.UserName, password)
	if err != nil {
		// 上面的 count 查询和 insert 之间存在并发竞态, 这里靠唯一索引兜底
		if isDuplicateEntry(err) {
			return ErrorUserExit
		}
		return err
	}
	user.UserID = userID
	return
}

// Login 校验用户名和密码, 校验通过后会把数据库里的用户信息写回 user
func Login(user *models.User) (err error) {
	originPassword := user.Password // 记录一下原始密码
	sqlStr := "select user_id, username, password from user where username = ?"
	err = db.Get(user, sqlStr, user.UserName)
	if err != nil && err != sql.ErrNoRows {
		// 查询数据库出错
		return
	}
	if err == sql.ErrNoRows {
		// 用户不存在
		return ErrorUserNotExit
	}
	// 生成加密密码与查询到的密码比较
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(originPassword)); err != nil {
		return ErrorPasswordWrong
	}
	return
}

// GetUserByID 根据用户 ID 查询用户
func GetUserByID(userID uint64) (user *models.User, err error) {
	user = new(models.User)
	sqlStr := `select user_id, username from user where user_id = ?`
	err = db.Get(user, sqlStr, userID)
	if err == sql.ErrNoRows {
		return nil, ErrorUserNotExit
	}
	if err != nil {
		return nil, ErrorQueryFailed
	}
	return
}

// GetUserProfileByID 查询用户主页需要的公开信息
func GetUserProfileByID(userID uint64) (profile *models.UserProfile, err error) {
	profile = new(models.UserProfile)
	sqlStr := `select user_id, username, create_time from user where user_id = ?`
	err = db.Get(profile, sqlStr, userID)
	if err == sql.ErrNoRows {
		return nil, ErrorUserNotExit
	}
	if err != nil {
		zap.L().Error("query user profile failed", zap.Uint64("user_id", userID), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return
}

// GetAllUserIDs 查询全部用户 ID(按注册时间正序)。
// 目前给 cmd/seed 造测试投票数据用: 投票记录在 Redis 里, 需要一份"可投票的人"清单。
func GetAllUserIDs() (userIDs []uint64, err error) {
	sqlStr := `select user_id from user order by create_time asc, user_id asc`
	userIDs = make([]uint64, 0, 64)
	if err = db.Select(&userIDs, sqlStr); err != nil {
		zap.L().Error("query all user ids failed", zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return userIDs, nil
}
