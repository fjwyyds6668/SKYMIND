package smart_query

import (
	"skymind/models"
)

// AssistantAPI 助手API
type AssistantAPI struct {
}

// GetAssistants 获取所有助手
func (api *AssistantAPI) GetAssistants() ([]interface{}, error) {
	assistants, err := assistantService.GetAssistants()
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(assistants))
	for i, assistant := range assistants {
		result[i] = assistant
	}
	return result, nil
}

// GetAssistantByID 根据ID获取助手
func (api *AssistantAPI) GetAssistantByID(id string) (interface{}, error) {
	assistant, err := assistantService.GetAssistantByID(id)
	if err != nil {
		return nil, err
	}
	return assistant, nil
}

// CreateAssistant 创建新助手
func (api *AssistantAPI) CreateAssistant(assistant map[string]interface{}) (interface{}, error) {
	// 将map转换为模型结构
	assistantModel := &models.Assistant{
		Name:        getString(assistant, "name"),
		Emoji:       getString(assistant, "emoji"),
		Prompt:      getString(assistant, "prompt"),
		Description: getString(assistant, "description"),
		Settings:    getString(assistant, "settings"),
		IsDefault:   getBool(assistant, "is_default"),
		IsActive:    getBool(assistant, "is_active"),
	}

	// 处理sort_order
	if sortOrder, ok := assistant["sort_order"].(float64); ok {
		assistantModel.SortOrder = int(sortOrder)
	}

	result, err := assistantService.CreateAssistant(assistantModel)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateAssistant 更新助手
func (api *AssistantAPI) UpdateAssistant(assistant map[string]interface{}) error {
	// 将map转换为模型结构
	assistantModel := &models.Assistant{
		ID:          getString(assistant, "id"),
		Name:        getString(assistant, "name"),
		Emoji:       getString(assistant, "emoji"),
		Prompt:      getString(assistant, "prompt"),
		Description: getString(assistant, "description"),
		Settings:    getString(assistant, "settings"),
		IsDefault:   getBool(assistant, "is_default"),
		IsActive:    getBool(assistant, "is_active"),
	}

	// 处理sort_order
	if sortOrder, ok := assistant["sort_order"].(float64); ok {
		assistantModel.SortOrder = int(sortOrder)
	}

	return assistantService.UpdateAssistant(assistantModel)
}

// DeleteAssistant 删除助手
func (api *AssistantAPI) DeleteAssistant(id string) error {
	return assistantService.DeleteAssistant(id)
}

// GetDefaultAssistant 获取默认助手
func (api *AssistantAPI) GetDefaultAssistant() (interface{}, error) {
	assistant, err := assistantService.GetDefaultAssistant()
	if err != nil {
		return nil, err
	}
	return assistant, nil
}

// UpdateAssistantsSortOrder 批量更新助手排序
func (api *AssistantAPI) UpdateAssistantsSortOrder(sortOrders []map[string]interface{}) error {
	return assistantService.UpdateAssistantsSortOrder(sortOrders)
}
