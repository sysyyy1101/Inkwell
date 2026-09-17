package mysql

import (
	"inkwell_backend/settings"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

// connMaxLifetime 连接最长存活时间。
// MySQL 默认 8 小时会断开空闲连接, 不设置的话客户端可能拿到已经失效的连接。
const connMaxLifetime = time.Hour

// Init 初始化MySQL连接
func Init(cfg *settings.MySQLConfig) (err error) {
	// "user:password@tcp(host:port)/dbname"
	// charset=utf8mb4 保证中文不会乱码
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	return
}

// Close 关闭MySQL连接
func Close() {
	_ = db.Close()
}
