package smart_query

import (
	"fmt"
	"skymind/global"
	"skymind/logger"
	"time"

	"skymind/database"
	"skymind/models"

	"gorm.io/gorm"
)

// TopicService 话题服务
type TopicService struct{}

// GetTopics 获取助手的话题列表
func (s *TopicService) GetTopics(assistantID string) ([]models.Topic, error) {
	var topics []models.Topic
	err := global.SLDB.Where("assistant_id = ?", assistantID).
		Order("sort_order ASC, created_at ASC").
		Find(&topics).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query topics: %w", err)
	}
	return topics, nil
}

// CreateTopic 创建新话题
func (s *TopicService) CreateTopic(topic *models.Topic) (*models.Topic, error) {
	logger.LogInfo("Creating new topic", map[string]interface{}{
		"assistantId": topic.AssistantID,
		"name":        topic.Name,
	})
	
	// 生成话题ID
	if topic.ID == "" {
		id, err := database.GenerateIDString()
		if err != nil {
			logger.LogError("Failed to generate topic ID", err)
			return nil, fmt.Errorf("failed to generate topic ID: %w", err)
		}
		topic.ID = id
	}

	// 获取当前最大的sort_order值，新话题排在最后（使用0-based索引）
	var maxSortOrder int
	global.SLDB.Model(&models.Topic{}).Where("assistant_id = ?", topic.AssistantID).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSortOrder)
	topic.SortOrder = maxSortOrder + 1

	// 设置默认值
	if topic.CreatedAt.IsZero() {
		topic.CreatedAt = time.Now()
	}
	if topic.UpdatedAt.IsZero() {
		topic.UpdatedAt = time.Now()
	}

	if err := global.SLDB.Create(topic).Error; err != nil {
		logger.LogError("Failed to create topic", err, map[string]interface{}{
			"assistantId": topic.AssistantID,
			"name":        topic.Name,
		})
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}
	
	logger.LogDatabaseOperation("create", "topics", topic.ID, nil)
	return topic, nil
}

// UpdateTopic 更新话题
func (s *TopicService) UpdateTopic(topic *models.Topic) error {
	// 检查话题是否存在
	var existing models.Topic
	if err := global.SLDB.Where("id = ?", topic.ID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("topic not found")
		}
		return fmt.Errorf("failed to check topic: %w", err)
	}

	topic.UpdatedAt = time.Now()

	if err := global.SLDB.Model(topic).Updates(topic).Error; err != nil {
		return fmt.Errorf("failed to update topic: %w", err)
	}

	return nil
}

// UpdateTopicTitle 更新话题标题
func (s *TopicService) UpdateTopicTitle(id, title string) error {
	logger.LogInfo("Updating topic title", map[string]interface{}{
		"topicId": id,
		"title":   title,
	})
	
	if err := global.SLDB.Model(&models.Topic{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        title,
			"updated_at":  time.Now(),
		}).Error; err != nil {
		logger.LogError("Failed to update topic title", err, map[string]interface{}{
			"topicId": id,
		})
		return fmt.Errorf("failed to update topic title: %w", err)
	}
	
	logger.LogDatabaseOperation("update", "topics", id, nil)
	return nil
}

// UpdateTopicsSortOrder 批量更新话题排序
func (s *TopicService) UpdateTopicsSortOrder(sortOrders []map[string]interface{}) error {
	// 使用事务确保数据一致性
	return global.SLDB.Transaction(func(tx *gorm.DB) error {
		for _, item := range sortOrders {
			id, ok := item["id"].(string)
			if !ok {
				return fmt.Errorf("invalid topic id")
			}
			sortOrder, ok := item["sort_order"].(float64)
			if !ok {
				return fmt.Errorf("invalid sort order for topic %s", id)
			}

			if err := tx.Model(&models.Topic{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"sort_order": int(sortOrder),
					"updated_at": time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("failed to update topic sort order: %w", err)
			}
		}
		return nil
	})
}

// DeleteTopic 删除话题（级联删除相关的对话和消息）
// deleteTopic: 是否删除话题本身，false表示只删除对话和消息
func (s *TopicService) DeleteTopic(id string, deleteTopic bool) error {
	logger.LogInfo("Deleting topic", map[string]interface{}{
		"topicId":    id,
		"deleteTopic": deleteTopic,
	})
	
	// 检查话题是否存在
	var topic models.Topic
	if err := global.SLDB.Where("id = ?", id).First(&topic).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.LogError("Topic not found for deletion", fmt.Errorf("topic not found"), map[string]interface{}{
				"topicId": id,
			})
			return fmt.Errorf("topic not found")
		}
		logger.LogError("Failed to check topic for deletion", err, map[string]interface{}{
			"topicId": id,
		})
		return fmt.Errorf("failed to check topic: %w", err)
	}

	// 使用事务确保数据一致性
	return global.SLDB.Transaction(func(tx *gorm.DB) error {
		// 1. 删除该话题下的所有消息
		if err := tx.Exec(`
			DELETE FROM messages 
			WHERE conversation_id IN (
				SELECT id FROM conversations WHERE topic_id = ?
			)
		`, id).Error; err != nil {
			return fmt.Errorf("failed to delete messages: %w", err)
		}

		// 2. 删除该话题下的所有对话
		if err := tx.Where("topic_id = ?", id).Delete(&models.Conversation{}).Error; err != nil {
			return fmt.Errorf("failed to delete conversations: %w", err)
		}

		// 3. 根据参数决定是否删除话题本身
		if deleteTopic {
			if err := tx.Delete(&topic).Error; err != nil {
				return fmt.Errorf("failed to delete topic: %w", err)
			}
		}

		return nil
	})
}

// GetTopicByID 根据ID获取话题
func (s *TopicService) GetTopicByID(id string) (*models.Topic, error) {
	var topic models.Topic
	err := global.SLDB.Where("id = ?", id).First(&topic).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("topic not found")
		}
		return nil, fmt.Errorf("failed to query topic: %w", err)
	}
	return &topic, nil
}

// createDefaultTopicForAssistant 为助手创建默认话题（内部方法）
func (s *TopicService) createDefaultTopicForAssistant(assistantID string) error {
	logger.LogInfo("Creating default topic for assistant", map[string]interface{}{
		"assistantId": assistantID,
	})
	
	// 生成话题ID
	topicID, err := database.GenerateIDString()
	if err != nil {
		logger.LogError("Failed to generate default topic ID", err)
		return fmt.Errorf("failed to generate topic ID: %w", err)
	}

	// 创建默认话题
	topic := &models.Topic{
		ID:          topicID,
		AssistantID: assistantID,
		Name:        "默认话题",
		SortOrder:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := global.SLDB.Create(topic).Error; err != nil {
		logger.LogError("Failed to create default topic", err, map[string]interface{}{
			"assistantId": assistantID,
			"topicId": topicID,
		})
		return fmt.Errorf("failed to create default topic: %w", err)
	}
	
	logger.LogDatabaseOperation("create", "topics", topicID, nil)
	return nil
}
