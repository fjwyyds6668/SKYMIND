package smart_query

import (
	"encoding/json"
	"fmt"
	"time"

	"skymind/database"
	"skymind/global"
	"skymind/logger"
	"skymind/models"

	"gorm.io/gorm"
)

// AssistantService 助手服务
type AssistantService struct {
	topicService TopicService
}

// GetAssistants 获取所有助手
func (s *AssistantService) GetAssistants() ([]models.Assistant, error) {
	var assistants []models.Assistant
	err := global.SLDB.Where("is_active = ?", true).
		Order("is_default DESC, sort_order ASC, created_at ASC").
		Find(&assistants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query assistants: %w", err)
	}
	return assistants, nil
}

// GetAssistantByID 根据ID获取助手
func (s *AssistantService) GetAssistantByID(id string) (*models.Assistant, error) {
	var assistant models.Assistant
	err := global.SLDB.Where("id = ? AND is_active = ?", id, true).First(&assistant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("assistant not found")
		}
		return nil, fmt.Errorf("failed to query assistant: %w", err)
	}
	return &assistant, nil
}

// CreateAssistant 创建新助手
func (s *AssistantService) CreateAssistant(assistant *models.Assistant) (*models.Assistant, error) {
	logger.LogInfo("Creating new assistant", map[string]interface{}{"name": assistant.Name})
	
	// 使用雪花算法生成ID
	if assistant.ID == "" {
		id, err := database.GenerateIDString()
		if err != nil {
			logger.LogError("Failed to generate assistant ID", err)
			return nil, fmt.Errorf("failed to generate assistant ID: %w", err)
		}
		assistant.ID = id
	}

	// 获取当前最大的sort_order值，新助手排在最后（使用0-based索引）
	var maxSortOrder int
	global.SLDB.Model(&models.Assistant{}).Where("is_active = ?", true).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSortOrder)
	assistant.SortOrder = maxSortOrder + 1

	// 设置默认值
	if assistant.Emoji == "" {
		assistant.Emoji = "🤖"
	}
	if assistant.CreatedAt.IsZero() {
		assistant.CreatedAt = time.Now()
	}
	if assistant.UpdatedAt.IsZero() {
		assistant.UpdatedAt = time.Now()
	}

	// 如果没有提供设置，使用默认设置
	if assistant.Settings == "" {
		defaultSettings := database.GetDefaultAssistantSettings()
		settingsJSON, _ := json.Marshal(defaultSettings)
		assistant.Settings = string(settingsJSON)
	}

	// 移除模型配置设置，现在统一从用户配置文件读取

	// 如果要设置为默认，需要先取消其他默认助手
	if assistant.IsDefault {
		global.SLDB.Model(&models.Assistant{}).Where("id != ?", assistant.ID).Update("is_default", false)
	}

	if err := global.SLDB.Create(assistant).Error; err != nil {
		logger.LogError("Failed to create assistant", err, map[string]interface{}{"id": assistant.ID, "name": assistant.Name})
		return nil, fmt.Errorf("failed to create assistant: %w", err)
	}
	
	logger.LogDatabaseOperation("create", "assistants", assistant.ID, nil)
	
	err := s.topicService.createDefaultTopicForAssistant(assistant.ID)
	if err != nil {
		// 记录错误但不阻止助手创建
		logger.LogError("Failed to create default topic for assistant", err, map[string]interface{}{"assistantId": assistant.ID})
	}

	return assistant, nil
}

// UpdateAssistant 更新助手
func (s *AssistantService) UpdateAssistant(assistant *models.Assistant) error {
	logger.LogInfo("Updating assistant", map[string]interface{}{
		"id":   assistant.ID,
		"name": assistant.Name,
		"isDefault": assistant.IsDefault,
	})
	
	// 检查助手是否存在
	existing, err := s.GetAssistantByID(assistant.ID)
	if err != nil {
		logger.LogError("Failed to check assistant for update", err, map[string]interface{}{
			"id": assistant.ID,
		})
		return fmt.Errorf("assistant not found: %w", err)
	}

	// 如果是默认助手，不能设置为非默认
	if existing.IsDefault && !assistant.IsDefault {
		logger.LogError("Cannot unset default assistant", fmt.Errorf("cannot unset default assistant"), map[string]interface{}{
			"id": assistant.ID,
		})
		return fmt.Errorf("cannot unset default assistant")
	}

	// 如果要设置为默认，需要先取消其他默认助手
	if assistant.IsDefault && !existing.IsDefault {
		global.SLDB.Model(&models.Assistant{}).Where("id != ?", assistant.ID).Update("is_default", false)
	}

	assistant.UpdatedAt = time.Now()

	if err := global.SLDB.Model(assistant).Updates(assistant).Error; err != nil {
		logger.LogError("Failed to update assistant", err, map[string]interface{}{
			"id":   assistant.ID,
			"name": assistant.Name,
		})
		return fmt.Errorf("failed to update assistant: %w", err)
	}
	
	logger.LogDatabaseOperation("update", "assistants", assistant.ID, nil)
	return nil
}

// UpdateAssistantsSortOrder 批量更新助手排序
func (s *AssistantService) UpdateAssistantsSortOrder(sortOrders []map[string]interface{}) error {
	// 使用事务确保数据一致性
	return global.SLDB.Transaction(func(tx *gorm.DB) error {
		for _, item := range sortOrders {
			id, ok := item["id"].(string)
			if !ok {
				return fmt.Errorf("invalid assistant id")
			}
			sortOrder, ok := item["sort_order"].(float64)
			if !ok {
				return fmt.Errorf("invalid sort order for assistant %s", id)
			}

			if err := tx.Model(&models.Assistant{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"sort_order": int(sortOrder),
					"updated_at": time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("failed to update assistant sort order: %w", err)
			}
		}
		return nil
	})
}

// DeleteAssistant 删除助手（级联删除相关的话题、对话和消息）
func (s *AssistantService) DeleteAssistant(id string) error {
	logger.LogInfo("Deleting assistant", map[string]interface{}{
		"assistantId": id,
	})
	
	// 检查是否为默认助手
	var assistant models.Assistant
	if err := global.SLDB.Where("id = ?", id).First(&assistant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.LogError("Assistant not found for deletion", fmt.Errorf("assistant not found"), map[string]interface{}{
				"assistantId": id,
			})
			return fmt.Errorf("assistant not found")
		}
		logger.LogError("Failed to check assistant for deletion", err, map[string]interface{}{
			"assistantId": id,
		})
		return fmt.Errorf("failed to check assistant: %w", err)
	}

	if assistant.IsDefault {
		logger.LogError("Cannot delete default assistant", fmt.Errorf("cannot delete default assistant"), map[string]interface{}{
			"assistantId": id,
			"name": assistant.Name,
		})
		return fmt.Errorf("cannot delete default assistant")
	}

	// 使用事务确保数据一致性
	return global.SLDB.Transaction(func(tx *gorm.DB) error {
		// 1. 删除该助手下所有话题的所有消息
		if err := tx.Exec(`
			DELETE FROM messages 
			WHERE conversation_id IN (
				SELECT id FROM conversations WHERE topic_id IN (
					SELECT id FROM topics WHERE assistant_id = ?
				)
			)
		`, id).Error; err != nil {
			logger.LogError("Failed to delete messages for assistant", err, map[string]interface{}{
				"assistantId": id,
			})
			return fmt.Errorf("failed to delete messages: %w", err)
		}

		// 2. 删除该助手下所有话题的所有对话
		if err := tx.Exec(`
			DELETE FROM conversations 
			WHERE topic_id IN (
				SELECT id FROM topics WHERE assistant_id = ?
			)
		`, id).Error; err != nil {
			logger.LogError("Failed to delete conversations for assistant", err, map[string]interface{}{
				"assistantId": id,
			})
			return fmt.Errorf("failed to delete conversations: %w", err)
		}

		// 3. 删除该助手下的所有话题
		if err := tx.Where("assistant_id = ?", id).Delete(&models.Topic{}).Error; err != nil {
			logger.LogError("Failed to delete topics for assistant", err, map[string]interface{}{
				"assistantId": id,
			})
			return fmt.Errorf("failed to delete topics: %w", err)
		}

		// 4. 软删除助手
		if err := tx.Delete(&assistant).Error; err != nil {
			logger.LogError("Failed to delete assistant", err, map[string]interface{}{
				"assistantId": id,
				"name": assistant.Name,
			})
			return fmt.Errorf("failed to delete assistant: %w", err)
		}
		
		logger.LogDatabaseOperation("delete", "assistants", id, nil)
		return nil
	})
}

// GetAssistantSettings 获取助手设置
func (s *AssistantService) GetAssistantSettings(id string) (*models.AssistantSettings, error) {
	assistant, err := s.GetAssistantByID(id)
	if err != nil {
		return nil, err
	}

	var settings models.AssistantSettings
	if assistant.Settings != "" {
		err = json.Unmarshal([]byte(assistant.Settings), &settings)
		if err != nil {
			return nil, fmt.Errorf("failed to parse settings: %w", err)
		}
	}

	return &settings, nil
}

// GetAssistantModelConfig 获取助手模型配置
// 现在统一从用户配置文件读取，不再从助手记录中获取
func (s *AssistantService) GetAssistantModelConfig(id string) (*models.ModelConfig, error) {
	// 直接返回默认的指示模型配置
	// 后续可以根据需要扩展为根据不同助手返回不同配置
	config := database.GetInstructModelConfig()
	return &config, nil
}

// GetDefaultAssistant 获取默认助手
func (s *AssistantService) GetDefaultAssistant() (*models.Assistant, error) {
	var assistant models.Assistant
	err := global.SLDB.Where("is_default = ? AND is_active = ?", true, true).First(&assistant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no default assistant found")
		}
		return nil, fmt.Errorf("failed to query default assistant: %w", err)
	}
	return &assistant, nil
}
