package models

import (
	"time"
	"gorm.io/gorm"
)

// Assistant 助手模型
type Assistant struct {
	ID          string `json:"id" gorm:"primaryKey;type:TEXT" db:"id"`
	Name        string `json:"name" gorm:"type:TEXT;not null" db:"name"`
	Emoji       string `json:"emoji" gorm:"type:TEXT;default:''" db:"emoji"`
	Prompt      string `json:"prompt" gorm:"type:TEXT;default:''" db:"prompt"`
	Description string `json:"description" gorm:"type:TEXT;default:''" db:"description"`
	Settings    string `json:"settings" gorm:"type:TEXT;default:'{}'" db:"settings"`       // JSON格式的设置
	SortOrder   int    `json:"sort_order" gorm:"type:INTEGER;default:0" db:"sort_order"`     // 排序字段
	IsDefault   bool   `json:"is_default" gorm:"type:BOOLEAN;default:false" db:"is_default"`
	IsActive    bool   `json:"is_active" gorm:"type:BOOLEAN;default:true" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index" db:"-"` // 软删除支持
}

// TableName 指定表名
func (Assistant) TableName() string {
	return "assistants"
}

// Settings 助手设置
type AssistantSettings struct {
	Temperature      float64 `json:"temperature"`
	ContextCount     int     `json:"contextCount"`
	EnableMaxTokens  bool    `json:"enableMaxTokens"`
	MaxTokens        int     `json:"maxTokens"`
	StreamOutput     bool    `json:"streamOutput"`
	TopP             float64 `json:"topP"`
	ToolUseMode      string  `json:"toolUseMode"`
	ReasoningEffort  string  `json:"reasoning_effort"`
	QwenThinkMode    bool    `json:"qwenThinkMode"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	ID       string `json:"id"`
	ApiBase  string `json:"apiBase"`
	ApiKey   string `json:"apiKey"`
	Name     string `json:"name"`
	Group    string `json:"group"`
}
