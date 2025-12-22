package models

import (
	"time"

	"gorm.io/gorm"
)

// File 文件模型
type File struct {
	ID           string         `gorm:"type:varchar(36);primaryKey;comment:文件唯一标识" json:"id"`
	OriginalPath string         `gorm:"type:varchar(500);not null;comment:文件原始路径" json:"originalPath"`
	OriginalMD5  string         `gorm:"type:varchar(32);not null;comment:文件原始MD5值" json:"originalMd5"`
	OriginalName string         `gorm:"type:varchar(255);not null;comment:原始文件名(不含后缀)" json:"originalName"`
	FileSuffix   string         `gorm:"type:varchar(10);not null;comment:文件后缀" json:"fileSuffix"`
	FileSize     int64          `gorm:"not null;comment:文件大小(字节)" json:"fileSize"`
	RelatedID    string         `gorm:"type:varchar(36);not null;comment:关联ID(如消息ID)" json:"relatedId"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index" db:"-"` // 软删除支持
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}
