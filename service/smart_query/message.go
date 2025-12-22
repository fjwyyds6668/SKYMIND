package smart_query

import (
	"encoding/json"
	"fmt"
	"skymind/global"
	"time"

	"skymind/database"
	"skymind/logger"
	"skymind/models"

	"gorm.io/gorm"
)

// MessageService 消息服务
type MessageService struct{}

// GetMessages 获取对话的消息列表
func (s *MessageService) GetMessages(conversationID string) ([]models.Message, error) {
	var messages []models.Message
	err := global.SLDB.Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	return messages, nil
}

// CreateMessage 创建新消息
func (s *MessageService) CreateMessage(message *models.Message) error {
	logger.LogInfo("Creating new message", map[string]interface{}{
		"conversationId": message.ConversationID,
		"role":           message.Role,
		"contentLength":  len(message.Content),
	})

	// 生成消息ID
	if message.ID == "" {
		id, err := database.GenerateIDString()
		if err != nil {
			logger.LogError("Failed to generate message ID", err)
			return fmt.Errorf("failed to generate message ID: %w", err)
		}
		message.ID = id
	}

	// 设置默认值
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	if err := global.SLDB.Create(message).Error; err != nil {
		logger.LogError("Failed to create message", err, map[string]interface{}{
			"conversationId": message.ConversationID,
			"role":           message.Role,
		})
		return fmt.Errorf("failed to create message: %w", err)
	}

	logger.LogDatabaseOperation("create", "messages", message.ID, nil)
	return nil
}

// UpdateMessage 更新消息
func (s *MessageService) UpdateMessage(message *models.Message) error {
	// 检查消息是否存在
	var existing models.Message
	if err := global.SLDB.Where("id = ?", message.ID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to check message: %w", err)
	}

	if err := global.SLDB.Model(message).Updates(message).Error; err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

// DeleteMessage 删除消息（软删除）并更新对话设置
func (s *MessageService) DeleteMessage(id string) error {
	logger.LogInfo("Deleting message", map[string]interface{}{
		"messageId": id,
	})
	
	// 检查消息是否存在
	var message models.Message
	if err := global.SLDB.Where("id = ?", id).First(&message).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.LogError("Message not found for deletion", fmt.Errorf("message not found"), map[string]interface{}{
				"messageId": id,
			})
			return fmt.Errorf("message not found")
		}
		logger.LogError("Failed to check message for deletion", err, map[string]interface{}{
			"messageId": id,
		})
		return fmt.Errorf("failed to check message: %w", err)
	}

	conversationID := message.ConversationID
	role := message.Role

	// 根据conversation_id和role，找到属于同一Conversation，同一类型的message，并根据创建时间排序
	var sameRoleMessages []models.Message
	if err := global.SLDB.Where("conversation_id = ? AND role = ?", conversationID, role).
		Order("created_at ASC").
		Find(&sameRoleMessages).Error; err != nil {
		return fmt.Errorf("failed to query same role messages: %w", err)
	}

	// 找到待删除消息在列表中的位置
	deleteIndex := -1
	for i, msg := range sameRoleMessages {
		if msg.ID == id {
			deleteIndex = i
			break
		}
	}

	if deleteIndex == -1 {
		return fmt.Errorf("message not found in same role messages list")
	}

	// 软删除待删除的消息
	if err := global.SLDB.Delete(&message).Error; err != nil {
		logger.LogError("Failed to delete message", err, map[string]interface{}{
			"messageId": id,
			"role":      role,
		})
		return fmt.Errorf("failed to delete message: %w", err)
	}
	
	logger.LogDatabaseOperation("delete", "messages", id, nil)

	// 获取对话的当前设置
	var conversation models.Conversation
	if err := global.SLDB.Where("id = ?", conversationID).First(&conversation).Error; err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// 解析设置
	var settings map[string]interface{}
	if conversation.Settings != "" {
		if err := json.Unmarshal([]byte(conversation.Settings), &settings); err != nil {
			// 如果解析失败，使用空的设置
			settings = make(map[string]interface{})
		}
	} else {
		settings = make(map[string]interface{})
	}

	// 确定下一个消息的ID
	var nextMessageID string
	if deleteIndex+1 < len(sameRoleMessages) {
		// 如果有下一个消息，使用下一个消息的ID
		nextMessageID = sameRoleMessages[deleteIndex+1].ID
	} else if deleteIndex-1 >= 0 {
		// 如果没有下一个消息但有上一个消息，使用上一个消息的ID
		nextMessageID = sameRoleMessages[deleteIndex-1].ID
	}

	// 根据角色更新相应的设置字段
	if role == "user" {
		if nextMessageID != "" {
			settings["currentSendId"] = nextMessageID
		} else {
			// 如果没有其他消息，删除该字段
			delete(settings, "currentSendId")
		}
	} else if role == "assistant" {
		if nextMessageID != "" {
			settings["currentReplyId"] = nextMessageID
		} else {
			// 如果没有其他消息，删除该字段
			delete(settings, "currentReplyId")
		}
	}

	// 更新对话设置
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := global.SLDB.Model(&conversation).
		Update("settings", string(settingsJSON)).Error; err != nil {
		return fmt.Errorf("failed to update conversation settings: %w", err)
	}

	// 检查该对话是否还有其他消息
	var messageCount int64
	if err := global.SLDB.Model(&models.Message{}).
		Where("conversation_id = ?", conversationID).
		Count(&messageCount).Error; err != nil {
		return fmt.Errorf("failed to count remaining messages: %w", err)
	}

	// 如果没有其他消息，则软删除该对话
	if messageCount == 0 {
		// 软删除对话
		if err := global.SLDB.Delete(&conversation).Error; err != nil {
			return fmt.Errorf("failed to delete conversation: %w", err)
		}
	}

	return nil
}

// GetMessageByID 根据ID获取消息
func (s *MessageService) GetMessageByID(id string) (*models.Message, error) {
	var message models.Message
	err := global.SLDB.Where("id = ?", id).First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to query message: %w", err)
	}
	return &message, nil
}

// BatchCreateMessages 批量创建消息
func (s *MessageService) BatchCreateMessages(messages []models.Message) error {
	// 为每个消息生成ID和设置默认值
	for i := range messages {
		if messages[i].ID == "" {
			id, err := database.GenerateIDString()
			if err != nil {
				return fmt.Errorf("failed to generate message ID: %w", err)
			}
			messages[i].ID = id
		}
		if messages[i].CreatedAt.IsZero() {
			messages[i].CreatedAt = time.Now()
		}
	}

	// 批量插入
	if err := global.SLDB.CreateInBatches(messages, 100).Error; err != nil {
		return fmt.Errorf("failed to batch create messages: %w", err)
	}

	return nil
}

// UpdateTokenCount 更新消息的token数量
func (s *MessageService) UpdateTokenCount(messageID string, tokenCount int) error {
	if err := global.SLDB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Update("token_count", tokenCount).Error; err != nil {
		return fmt.Errorf("failed to update token count: %w", err)
	}
	return nil
}

// GetMessagesByRole 根据角色获取消息列表
func (s *MessageService) GetMessagesByRole(conversationID, role string) ([]models.Message, error) {
	var messages []models.Message
	err := global.SLDB.Where("conversation_id = ? AND role = ?", conversationID, role).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by role: %w", err)
	}
	return messages, nil
}

// GetLastMessage 获取对话的最后一条消息
func (s *MessageService) GetLastMessage(conversationID string) (*models.Message, error) {
	var message models.Message
	err := global.SLDB.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no messages found")
		}
		return nil, fmt.Errorf("failed to query last message: %w", err)
	}
	return &message, nil
}


// GetConversationMessagesWithLimit 获取对话的消息列表（限制数量）
func (s *MessageService) GetConversationMessagesWithLimit(conversationID string, limit int) ([]models.Message, error) {
	var messages []models.Message
	query := global.SLDB.Where("conversation_id = ?", conversationID).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query messages with limit: %w", err)
	}
	return messages, nil
}

// GetMessagesByTimeRange 根据时间范围获取消息
func (s *MessageService) GetMessagesByTimeRange(conversationID string, startTime, endTime time.Time) ([]models.Message, error) {
	var messages []models.Message
	err := global.SLDB.Where("conversation_id = ? AND created_at BETWEEN ? AND ?",
		conversationID, startTime, endTime).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by time range: %w", err)
	}
	return messages, nil
}

// CountMessages 统计对话的消息数量
func (s *MessageService) CountMessages(conversationID string) (int64, error) {
	var count int64
	err := global.SLDB.Model(&models.Message{}).
		Where("conversation_id = ?", conversationID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}
	return count, nil
}

// DeleteConversationsAfter 删除指定对话之后的所有对话及其消息
func (s *MessageService) DeleteConversationsAfter(conversationID string) error {
	// 获取指定对话的信息
	var targetConversation models.Conversation
	if err := global.SLDB.Where("id = ?", conversationID).First(&targetConversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("conversation not found")
		}
		return fmt.Errorf("failed to query target conversation: %w", err)
	}

	// 开始事务
	tx := global.SLDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询同一topic下创建时间晚于目标对话的所有对话
	var laterConversations []models.Conversation
	if err := tx.Where("topic_id = ? AND created_at > ?", targetConversation.TopicID, targetConversation.CreatedAt).
		Find(&laterConversations).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to query later conversations: %w", err)
	}

	// 删除这些对话及其所有消息
	for _, conv := range laterConversations {
		// 软删除该对话的所有消息
		if err := tx.Where("conversation_id = ?", conv.ID).Delete(&models.Message{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete messages of later conversation %s: %w", conv.ID, err)
		}

		// 软删除对话
		if err := tx.Delete(&conv).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete later conversation %s: %w", conv.ID, err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
