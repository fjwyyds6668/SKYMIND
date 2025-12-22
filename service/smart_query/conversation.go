package smart_query

import (
	"fmt"
	"time"

	"skymind/database"
	"skymind/global"
	"skymind/logger"
	"skymind/models"

	"gorm.io/gorm"
)

// ConversationService 对话服务
type ConversationService struct{}

// GetConversations 获取话题的对话列表（手动加载消息）
func (s *ConversationService) GetConversations(topicID string) ([]models.Conversation, error) {
	var conversations []models.Conversation

	// 查询对话列表，按创建时间排序保持对话顺序稳定
	err := global.SLDB.Where("topic_id = ?", topicID).
		Order("created_at ASC").
		Find(&conversations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}

	// 手动加载每个对话的消息
	for i := range conversations {
		var messages []models.Message
		err := global.SLDB.Where("conversation_id = ?", conversations[i].ID).
			Order("created_at ASC").
			Find(&messages).Error
		if err != nil {
			return nil, fmt.Errorf("failed to query messages for conversation %s: %w", conversations[i].ID, err)
		}
		conversations[i].Messages = messages
	}

	return conversations, nil
}

// CreateConversation 创建新对话
func (s *ConversationService) CreateConversation(conversation *models.Conversation) (string, error) {
	logger.LogInfo("Creating new conversation", map[string]interface{}{
		"topicId": conversation.TopicID,
		"title":   conversation.Title,
	})
	
	// 生成对话ID
	if conversation.ID == "" {
		id, err := database.GenerateIDString()
		if err != nil {
			logger.LogError("Failed to generate conversation ID", err)
			return "", fmt.Errorf("failed to generate conversation ID: %w", err)
		}
		conversation.ID = id
	}

	// 设置默认值
	if conversation.CreatedAt.IsZero() {
		conversation.CreatedAt = time.Now()
	}
	if conversation.UpdatedAt.IsZero() {
		conversation.UpdatedAt = time.Now()
	}

	// 如果没有标题，生成默认标题
	if conversation.Title == "" {
		conversation.Title = "新对话"
	}

	if err := global.SLDB.Create(conversation).Error; err != nil {
		logger.LogError("Failed to create conversation", err, map[string]interface{}{
			"topicId": conversation.TopicID,
			"title":   conversation.Title,
		})
		return "", fmt.Errorf("failed to create conversation: %w", err)
	}
	
	logger.LogDatabaseOperation("create", "conversations", conversation.ID, nil)
	return conversation.ID, nil
}

// UpdateConversation 更新对话
func (s *ConversationService) UpdateConversation(conversation *models.Conversation) error {
	// 检查对话是否存在
	var existing models.Conversation
	if err := global.SLDB.Where("id = ?", conversation.ID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("conversation not found")
		}
		return fmt.Errorf("failed to check conversation: %w", err)
	}

	conversation.UpdatedAt = time.Now()

	if err := global.SLDB.Model(conversation).Updates(conversation).Error; err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// DeleteConversation 删除对话（软删除）
func (s *ConversationService) DeleteConversation(id string) error {
	logger.LogInfo("Deleting conversation", map[string]interface{}{
		"conversationId": id,
	})
	
	// 检查对话是否存在
	var conversation models.Conversation
	if err := global.SLDB.Where("id = ?", id).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.LogError("Conversation not found for deletion", fmt.Errorf("conversation not found"), map[string]interface{}{
				"conversationId": id,
			})
			return fmt.Errorf("conversation not found")
		}
		logger.LogError("Failed to check conversation for deletion", err, map[string]interface{}{
			"conversationId": id,
		})
		return fmt.Errorf("failed to check conversation: %w", err)
	}

	// 软删除
	if err := global.SLDB.Delete(&conversation).Error; err != nil {
		logger.LogError("Failed to delete conversation", err, map[string]interface{}{
			"conversationId": id,
		})
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	
	logger.LogDatabaseOperation("delete", "conversations", id, nil)
	return nil
}

// GetConversationByID 根据ID获取对话
func (s *ConversationService) GetConversationByID(id string) (*models.Conversation, error) {
	var conversation models.Conversation
	err := global.SLDB.Where("id = ?", id).First(&conversation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to query conversation: %w", err)
	}

	// 手动加载消息
	var messages []models.Message
	err = global.SLDB.Where("conversation_id = ?", conversation.ID).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query messages for conversation %s: %w", conversation.ID, err)
	}
	conversation.Messages = messages

	return &conversation, nil
}

// ArchiveConversation 归档对话
func (s *ConversationService) ArchiveConversation(id string) error {
	if err := global.SLDB.Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("is_archived", true).Error; err != nil {
		return fmt.Errorf("failed to archive conversation: %w", err)
	}
	return nil
}

// UnarchiveConversation 取消归档对话
func (s *ConversationService) UnarchiveConversation(id string) error {
	if err := global.SLDB.Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("is_archived", false).Error; err != nil {
		return fmt.Errorf("failed to unarchive conversation: %w", err)
	}
	return nil
}

// GetArchivedConversations 获取已归档的对话列表
func (s *ConversationService) GetArchivedConversations(topicID string) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := global.SLDB.Where("topic_id = ? AND is_archived = ?", topicID, true).
		Order("updated_at DESC").
		Find(&conversations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query archived conversations: %w", err)
	}
	return conversations, nil
}

// UpdateConversationTitle 更新对话标题
func (s *ConversationService) UpdateConversationTitle(id, title string) error {
	if err := global.SLDB.Model(&models.Conversation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"title":      title,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update conversation title: %w", err)
	}
	return nil
}

// UpdateConversationSettings 更新对话设置
func (s *ConversationService) UpdateConversationSettings(id, settings string) error {
	logger.LogInfo("Updating conversation settings", map[string]interface{}{
		"conversationId": id,
		"settingsLength": len(settings),
	})
	
	if err := global.SLDB.Model(&models.Conversation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"settings":   settings,
			"updated_at": time.Now(),
		}).Error; err != nil {
		logger.LogError("Failed to update conversation settings", err, map[string]interface{}{
			"conversationId": id,
		})
		return fmt.Errorf("failed to update conversation settings: %w", err)
	}
	
	logger.LogDatabaseOperation("update", "conversations", id, nil)
	return nil
}
