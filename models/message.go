package models

import (
	"time"
	"gorm.io/gorm"
)

// Message 消息模型
type Message struct {
	ID         string         `json:"id" gorm:"primaryKey;type:TEXT" db:"id"`
	TopicID    string         `json:"topic_id" gorm:"type:TEXT;index" db:"topic_id"`           // 关联到主题
	ConversationID string         `json:"conversation_id" gorm:"type:TEXT;index" db:"conversation_id"` // 关联到对话
	Role       string         `json:"role" gorm:"type:TEXT;not null" db:"role"`           // 角色：user, assistant, system
	Content    string         `json:"content" gorm:"type:TEXT;not null" db:"content"`
	Reasoning  string         `json:"reasoning" gorm:"type:TEXT" db:"reasoning"` // 思考过程内容
	TokenCount int            `json:"token_count" gorm:"type:INTEGER;default:0" db:"token_count"` // token数量
	Metadata   string         `json:"metadata" gorm:"type:TEXT;default:'{}'" db:"metadata"` // JSON格式的元数据
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index" db:"-"` // 软删除支持
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}
