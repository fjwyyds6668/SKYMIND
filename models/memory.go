package models

import (
	"time"
	"gorm.io/gorm"
)

// Memory 记忆模型（用于向量搜索和智能召回）
type Memory struct {
	ID        string `json:"id" gorm:"primaryKey;type:TEXT" db:"id"`
	Memory    string `json:"memory" gorm:"type:TEXT;not null" db:"memory"`
	Hash      string `json:"hash" gorm:"type:TEXT;uniqueIndex" db:"hash"`
	Embedding []byte `json:"embedding" gorm:"type:BLOB" db:"embedding"`
	Metadata  string `json:"metadata" gorm:"type:TEXT;default:'{}'" db:"metadata"`
	UserID    string `json:"user_id" gorm:"type:TEXT;default:'';index" db:"user_id"`
	AgentID   string `json:"agent_id" gorm:"type:TEXT;default:'';index" db:"agent_id"`
	RunID     string `json:"run_id" gorm:"type:TEXT;default:''" db:"run_id"`
	IsDeleted bool   `json:"is_deleted" gorm:"type:BOOLEAN;default:false" db:"is_deleted"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index" db:"-"` // 软删除支持
}

// TableName 指定表名
func (Memory) TableName() string {
	return "memories"
}

// MemoryHistory 记忆历史表
type MemoryHistory struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	MemoryID     string    `json:"memory_id" gorm:"type:TEXT;not null;index" db:"memory_id"`
	PreviousValue string    `json:"previous_value" gorm:"type:TEXT" db:"previous_value"`
	NewValue     string    `json:"new_value" gorm:"type:TEXT" db:"new_value"`
	Action       string    `json:"action" gorm:"type:TEXT;not null;check:action IN ('ADD', 'UPDATE', 'DELETE')" db:"action"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	
	// 关联关系
	Memory *Memory `json:"memory,omitempty" gorm:"foreignKey:MemoryID"`
}

// TableName 指定表名
func (MemoryHistory) TableName() string {
	return "memory_history"
}
