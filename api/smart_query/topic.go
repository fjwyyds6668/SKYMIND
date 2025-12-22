package smart_query

import (
	"skymind/models"
)

// TopicAPI 话题API
type TopicAPI struct {
}

// GetTopics 获取助手的话题列表
func (api *TopicAPI) GetTopics(assistantID string) ([]interface{}, error) {
	topics, err := topicService.GetTopics(assistantID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(topics))
	for i, topic := range topics {
		result[i] = topic
	}
	return result, nil
}

// CreateTopic 创建新话题
func (api *TopicAPI) CreateTopic(topic map[string]interface{}) (interface{}, error) {
	// 将map转换为模型结构
	topicModel := &models.Topic{
		AssistantID:          getString(topic, "assistant_id"),
		Name:                 getString(topic, "name"),
		IsNameManuallyEdited: getBool(topic, "is_name_manually_edited"),
	}

	// 处理sort_order
	if sortOrder, ok := topic["sort_order"].(float64); ok {
		topicModel.SortOrder = int(sortOrder)
	}

	result, err := topicService.CreateTopic(topicModel)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateTopic 更新话题
func (api *TopicAPI) UpdateTopic(topic map[string]interface{}) error {
	// 将map转换为模型结构
	topicModel := &models.Topic{
		ID:                   getString(topic, "id"),
		AssistantID:          getString(topic, "assistant_id"),
		Name:                 getString(topic, "name"),
		IsNameManuallyEdited: getBool(topic, "is_name_manually_edited"),
	}

	// 处理sort_order
	if sortOrder, ok := topic["sort_order"].(float64); ok {
		topicModel.SortOrder = int(sortOrder)
	}

	return topicService.UpdateTopic(topicModel)
}

// UpdateTopicTitle 更新话题标题
func (api *TopicAPI) UpdateTopicTitle(id, title string) error {
	return topicService.UpdateTopicTitle(id, title)
}

// DeleteTopic 删除话题
func (api *TopicAPI) DeleteTopic(id string, deleteTopic bool) error {
	return topicService.DeleteTopic(id, deleteTopic)
}

// GetTopicByID 根据ID获取话题
func (api *TopicAPI) GetTopicByID(id string) (interface{}, error) {
	topic, err := topicService.GetTopicByID(id)
	if err != nil {
		return nil, err
	}
	return topic, nil
}

// UpdateTopicsSortOrder 批量更新话题排序
func (api *TopicAPI) UpdateTopicsSortOrder(sortOrders []map[string]interface{}) error {
	return topicService.UpdateTopicsSortOrder(sortOrders)
}
