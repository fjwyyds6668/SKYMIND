package models

import (
	"time"
	"gorm.io/gorm"
)

// Topic 话题模型（类似CherryStudio的topics）
type Topic struct {
	ID                 string `json:"id" gorm:"primaryKey;type:TEXT" db:"id"`
	AssistantID        string `json:"assistant_id" gorm:"type:TEXT;not null;index" db:"assistant_id"`
	Name               string `json:"name" gorm:"type:TEXT;not null" db:"name"`
	IsNameManuallyEdited bool `json:"is_name_manually_edited" gorm:"type:BOOLEAN;default:false" db:"is_name_manually_edited"`
	SortOrder          int    `json:"sort_order" gorm:"type:INTEGER;default:0" db:"sort_order"`     // 排序字段
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index" db:"-"` // 软删除支持
	
}

// TableName 指定表名
func (Topic) TableName() string {
	return "topics"
}
