package models

import "time"

// StreamInfo 流式输出信息
type StreamInfo struct {
	ID           string
	Type         string // 'chat', 'title', 'prompt', 'system_prompt'
	RelatedID    string // 关联的conversationId, topicId, assistantId
	Ctx          interface{} // context.Context，避免导入循环
	Cancel       interface{} // context.CancelFunc，避免导入循环
	StartTime    time.Time
}
