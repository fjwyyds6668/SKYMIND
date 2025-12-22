package smart_query

import (
	"context"
	"skymind/models"
)

// MessageAPI 消息API
type MessageAPI struct {
}

// GetMessages 获取对话的消息列表
func (api *MessageAPI) GetMessages(conversationID string) ([]interface{}, error) {
	messages, err := messageService.GetMessages(conversationID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(messages))
	for i, message := range messages {
		result[i] = message
	}
	return result, nil
}

// CreateMessage 创建新消息
func (api *MessageAPI) CreateMessage(message map[string]interface{}) (interface{}, error) {
	// 将map转换为模型结构
	messageModel := &models.Message{
		TopicID:        getString(message, "topic_id"),
		ConversationID: getString(message, "conversation_id"),
		Role:           getString(message, "role"),
		Content:        getString(message, "content"),
		Reasoning:      getString(message, "reasoning"), // 添加思考内容
		Metadata:       getString(message, "metadata"),
	}

	// 处理token_count
	if tokenCount, ok := message["token_count"].(float64); ok {
		messageModel.TokenCount = int(tokenCount)
	}

	err := messageService.CreateMessage(messageModel)
	if err != nil {
		return nil, err
	}

	return messageModel, nil
}

// UpdateMessage 更新消息
func (api *MessageAPI) UpdateMessage(message map[string]interface{}) error {
	// 将map转换为模型结构
	messageModel := &models.Message{
		ID:             getString(message, "id"),
		ConversationID: getString(message, "conversation_id"),
		Role:           getString(message, "role"),
		Content:        getString(message, "content"),
		Reasoning:      getString(message, "reasoning"), // 添加思考内容
		Metadata:       getString(message, "metadata"),
	}

	// 处理token_count
	if tokenCount, ok := message["token_count"].(float64); ok {
		messageModel.TokenCount = int(tokenCount)
	}

	return messageService.UpdateMessage(messageModel)
}

// DeleteMessage 删除消息
func (api *MessageAPI) DeleteMessage(id, conversationID string) error {
	return messageService.DeleteMessage(id)
}

// GetMessageByID 根据ID获取消息
func (api *MessageAPI) GetMessageByID(id string) (interface{}, error) {
	message, err := messageService.GetMessageByID(id)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// UpdateTokenCount 更新消息的token数量
func (api *MessageAPI) UpdateTokenCount(messageID string, tokenCount int) error {
	return messageService.UpdateTokenCount(messageID, tokenCount)
}

// GetMessagesByRole 根据角色获取消息列表
func (api *MessageAPI) GetMessagesByRole(conversationID, role string) ([]interface{}, error) {
	messages, err := messageService.GetMessagesByRole(conversationID, role)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(messages))
	for i, message := range messages {
		result[i] = message
	}
	return result, nil
}

// GetLastMessage 获取对话的最后一条消息
func (api *MessageAPI) GetLastMessage(conversationID string) (interface{}, error) {
	message, err := messageService.GetLastMessage(conversationID)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// StreamChatCompletion 流式调用大模型
func (api *MessageAPI) StreamChatCompletion(ctx context.Context, streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
	return generatorService.StreamChatCompletion(ctx, streamID, streamType, relatedID, messages, modelType)
}

// StopStreamChatCompletion 停止指定的流式聊天
func (api *MessageAPI) StopStreamChatCompletion(streamID string) {
	generatorService.StopStreamChatCompletion(streamID)
}

// GetConversationMessagesWithLimit 获取对话的消息列表（限制数量）
func (api *MessageAPI) GetConversationMessagesWithLimit(conversationID string, limit int) ([]interface{}, error) {
	messages, err := messageService.GetConversationMessagesWithLimit(conversationID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(messages))
	for i, message := range messages {
		result[i] = message
	}
	return result, nil
}

// CountMessages 统计对话的消息数量
func (api *MessageAPI) CountMessages(conversationID string) (int64, error) {
	return messageService.CountMessages(conversationID)
}

// DeleteConversationsAfter 删除指定对话之后的所有对话及其消息
func (api *MessageAPI) DeleteConversationsAfter(conversationID string) error {
	return messageService.DeleteConversationsAfter(conversationID)
}
