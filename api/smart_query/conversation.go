package smart_query

import (
	"skymind/models"
)

// ConversationAPI 对话API
type ConversationAPI struct {
}

// GetConversations 获取话题的对话列表
func (api *ConversationAPI) GetConversations(topicID string) ([]interface{}, error) {
	conversations, err := conversationService.GetConversations(topicID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(conversations))
	for i, conversation := range conversations {
		result[i] = conversation
	}
	return result, nil
}

// CreateConversation 创建新对话
func (api *ConversationAPI) CreateConversation(conversation map[string]interface{}) (string, error) {
	// 将map转换为模型结构
	conversationModel := &models.Conversation{
		Title:       getString(conversation, "title"),
		TopicID:     getString(conversation, "topic_id"),
		AssistantID: getString(conversation, "assistant_id"),
		Settings:    getString(conversation, "settings"),
		IsArchived:  getBool(conversation, "is_archived"),
	}

	return conversationService.CreateConversation(conversationModel)
}

// UpdateConversation 更新对话
func (api *ConversationAPI) UpdateConversation(conversation map[string]interface{}) error {
	// 将map转换为模型结构
	conversationModel := &models.Conversation{
		ID:          getString(conversation, "id"),
		Title:       getString(conversation, "title"),
		TopicID:     getString(conversation, "topic_id"),
		AssistantID: getString(conversation, "assistant_id"),
		Settings:    getString(conversation, "settings"),
		IsArchived:  getBool(conversation, "is_archived"),
	}

	return conversationService.UpdateConversation(conversationModel)
}

// DeleteConversation 删除对话
func (api *ConversationAPI) DeleteConversation(id string) error {
	return conversationService.DeleteConversation(id)
}

// GetConversationByID 根据ID获取对话
func (api *ConversationAPI) GetConversationByID(id string) (interface{}, error) {
	conversation, err := conversationService.GetConversationByID(id)
	if err != nil {
		return nil, err
	}
	return conversation, nil
}

// ArchiveConversation 归档对话
func (api *ConversationAPI) ArchiveConversation(id string) error {
	return conversationService.ArchiveConversation(id)
}

// UnarchiveConversation 取消归档对话
func (api *ConversationAPI) UnarchiveConversation(id string) error {
	return conversationService.UnarchiveConversation(id)
}

// GetArchivedConversations 获取已归档的对话列表
func (api *ConversationAPI) GetArchivedConversations(topicID string) ([]interface{}, error) {
	conversations, err := conversationService.GetArchivedConversations(topicID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(conversations))
	for i, conversation := range conversations {
		result[i] = conversation
	}
	return result, nil
}

// UpdateConversationTitle 更新对话标题
func (api *ConversationAPI) UpdateConversationTitle(id, title string) error {
	return conversationService.UpdateConversationTitle(id, title)
}

// UpdateConversationSettings 更新对话设置
func (api *ConversationAPI) UpdateConversationSettings(id, settings string) error {
	return conversationService.UpdateConversationSettings(id, settings)
}
