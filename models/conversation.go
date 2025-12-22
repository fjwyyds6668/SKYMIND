package models

import (
	"time"

	"gorm.io/gorm"
)

// Conversation 对话模型
type Conversation struct {
	ID          string         `json:"id" gorm:"primaryKey;type:TEXT" db:"id"`
	Title       string         `json:"title" gorm:"type:TEXT;not null" db:"title"`
	TopicID     string         `json:"topic_id" gorm:"type:TEXT;index" db:"topic_id"` // 关联到话题
	AssistantID string         `json:"assistant_id" gorm:"type:TEXT;index" db:"assistant_id"`
	Settings    string         `json:"settings" gorm:"type:TEXT;default:'{}'" db:"settings"` // JSON格式的设置
	IsArchived  bool           `json:"is_archived" gorm:"type:BOOLEAN;default:false" db:"is_archived"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index" db:"-"` // 软删除支持

	// 预加载的消息列表（不使用外键约束）
	Messages []Message `json:"messages,omitempty" gorm:"-"`
}

// TableName 指定表名
func (Conversation) TableName() string {
	return "conversations"
}

// ConversationSettings 对话设置
type ConversationSettings struct {
	Temperature     float64 `json:"temperature"`
	MaxTokens       int     `json:"maxTokens"`
	StreamOutput    bool    `json:"streamOutput"`
	TopP            float64 `json:"topP"`
	FrequencyPenalty float64 `json:"frequencyPenalty"`
	PresencePenalty float64 `json:"presencePenalty"`
	CurrentSendID   string `json:"currentSendId"`   // 当前显示的用户消息ID
	CurrentReplyID  string `json:"currentReplyId"`  // 当前显示的AI回复消息ID
}
