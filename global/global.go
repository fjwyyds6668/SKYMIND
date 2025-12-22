package global

import (
	"gorm.io/gorm"
)

// 全局变量
var (
	// SLDB 全局数据库连接
	SLDB *gorm.DB
)

// GetDB 获取全局数据库连接
func GetDB() *gorm.DB {
	return SLDB
}

// SetDB 设置全局数据库连接
func SetDB(db *gorm.DB) {
	SLDB = db
}
