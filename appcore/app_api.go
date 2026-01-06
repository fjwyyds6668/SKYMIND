package appcore

import "skymind/models"

// GetAssistants 获取所有助手 - Wails API方法
func GetAssistants(a *App) ([]interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.GetAssistants()
}

// GetAssistantByID 根据ID获取助手 - Wails API方法
func GetAssistantByID(a *App, id string) (interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.GetAssistantByID(id)
}

// CreateAssistant 创建新助手 - Wails API方法
func CreateAssistant(a *App, assistant map[string]interface{}) (interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.CreateAssistant(assistant)
}

// UpdateAssistant 更新助手 - Wails API方法
func UpdateAssistant(a *App, assistant map[string]interface{}) error {
	return a.SmartQueryAPI.AssistantAPI.UpdateAssistant(assistant)
}

// DeleteAssistant 删除助手 - Wails API方法
func DeleteAssistant(a *App, id string) error {
	return a.SmartQueryAPI.AssistantAPI.DeleteAssistant(id)
}

// GetDefaultAssistant 获取默认助手 - Wails API方法
func GetDefaultAssistant(a *App) (interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.GetDefaultAssistant()
}

// GetTopics 获取助手的话题列表 - Wails API方法
func GetTopics(a *App, assistantID string) ([]interface{}, error) {
	return a.SmartQueryAPI.TopicAPI.GetTopics(assistantID)
}

// CreateTopic 创建新话题 - Wails API方法
func CreateTopic(a *App, topic map[string]interface{}) (interface{}, error) {
	return a.SmartQueryAPI.TopicAPI.CreateTopic(topic)
}

// UpdateTopic 更新话题 - Wails API方法
func UpdateTopic(a *App, topic map[string]interface{}) error {
	return a.SmartQueryAPI.TopicAPI.UpdateTopic(topic)
}

// UpdateTopicTitle 更新话题标题 - Wails API方法
func UpdateTopicTitle(a *App, id, title string) error {
	return a.SmartQueryAPI.TopicAPI.UpdateTopicTitle(id, title)
}

// DeleteTopic 删除话题 - Wails API方法
func DeleteTopic(a *App, id string, deleteTopic bool) error {
	return a.SmartQueryAPI.TopicAPI.DeleteTopic(id, deleteTopic)
}

// GetMessages 获取对话的消息列表 - Wails API方法
func GetMessages(a *App, conversationID string) ([]interface{}, error) {
	return a.SmartQueryAPI.MessageAPI.GetMessages(conversationID)
}

// GetConversations 获取话题的对话列表 - Wails API方法
func GetConversations(a *App, topicID string) ([]interface{}, error) {
	return a.SmartQueryAPI.ConversationAPI.GetConversations(topicID)
}

// CreateConversation 创建新对话 - Wails API方法
func CreateConversation(a *App, conversation map[string]interface{}) (string, error) {
	return a.SmartQueryAPI.ConversationAPI.CreateConversation(conversation)
}

// UpdateConversationSettings 更新对话设置 - Wails API方法
func UpdateConversationSettings(a *App, id, settings string) error {
	return a.SmartQueryAPI.ConversationAPI.UpdateConversationSettings(id, settings)
}

// UpdateConversationTitle 更新对话标题 - Wails API方法
func UpdateConversationTitle(a *App, id, title string) error {
	return a.SmartQueryAPI.ConversationAPI.UpdateConversationTitle(id, title)
}

// GenerateConversationTitle 生成对话标题 - Wails API方法
func GenerateConversationTitle(a *App, userMessage, aiResponse string) (string, error) {
	return a.SmartQueryAPI.GeneratorAPI.GenerateConversationTitle(userMessage, aiResponse)
}

// CreateMessage 创建新消息 - Wails API方法
func CreateMessage(a *App, message map[string]interface{}) (interface{}, error) {
	return a.SmartQueryAPI.MessageAPI.CreateMessage(message)
}

// DeleteMessage 删除消息 - Wails API方法
func DeleteMessage(a *App, id, conversationID string) error {
	return a.SmartQueryAPI.MessageAPI.DeleteMessage(id, conversationID)
}

// UpdateMessage 更新消息 - Wails API方法
func UpdateMessage(a *App, id, content, reasoning string) error {
	message := map[string]interface{}{
		"id":        id,
		"content":   content,
		"reasoning": reasoning,
	}
	return a.SmartQueryAPI.MessageAPI.UpdateMessage(message)
}

// StreamChatCompletion 流式调用大模型 - Wails API方法
func StreamChatCompletion(a *App, streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
	return a.SmartQueryAPI.GeneratorAPI.StreamChatCompletion(a.Ctx, streamID, streamType, relatedID, messages, modelType)
}

// StopStreamChatCompletion 停止指定的流式聊天 - Wails API方法
func StopStreamChatCompletion(a *App, streamID string) {
	a.SmartQueryAPI.GeneratorAPI.StopStreamChatCompletion(streamID)
}

// StopAllStreams 停止所有流式聊天 - Wails API方法
func StopAllStreams(a *App) {
	a.SmartQueryAPI.GeneratorAPI.StopAllStreams()
}

// GetActiveStreams 获取活跃的流式输出列表 - Wails API方法
func GetActiveStreams(a *App) map[string]*models.StreamInfo {
	return a.SmartQueryAPI.GeneratorAPI.GetActiveStreams()
}

// IsStreamActive 检查指定流式输出是否活跃 - Wails API方法
func IsStreamActive(a *App, streamID string) bool {
	return a.SmartQueryAPI.GeneratorAPI.IsStreamActive(streamID)
}

// UpdateAssistantsSortOrder 批量更新助手排序 - Wails API方法
func UpdateAssistantsSortOrder(a *App, sortOrders []map[string]interface{}) error {
	return a.SmartQueryAPI.AssistantAPI.UpdateAssistantsSortOrder(sortOrders)
}

// UpdateTopicsSortOrder 批量更新话题排序 - Wails API方法
func UpdateTopicsSortOrder(a *App, sortOrders []map[string]interface{}) error {
	return a.SmartQueryAPI.TopicAPI.UpdateTopicsSortOrder(sortOrders)
}

// DeleteConversationsAfter 删除指定对话之后的所有对话及其消息 - Wails API方法
func DeleteConversationsAfter(a *App, conversationID string) error {
	return a.SmartQueryAPI.MessageAPI.DeleteConversationsAfter(conversationID)
}

// GenerateSystemPrompt 生成系统提示词 - Wails API方法
func GenerateSystemPrompt(a *App, name, description, userInput string) (string, error) {
	return a.SmartQueryAPI.GeneratorAPI.GenerateSystemPrompt(name, description, userInput)
}

// OptimizeUserPrompt 优化用户提示词 - Wails API方法
func OptimizeUserPrompt(a *App, originalPrompt string) (string, error) {
	return a.SmartQueryAPI.GeneratorAPI.OptimizeUserPrompt(originalPrompt)
}

// GenerateTopicTitle 生成话题标题 - Wails API方法
func GenerateTopicTitle(a *App, conversationTitles []string) (string, error) {
	return a.SmartQueryAPI.GeneratorAPI.GenerateTopicTitle(conversationTitles)
}

// GenerateStreamID 生成流式输出ID - Wails API方法
func GenerateStreamID(a *App) (string, error) {
	return a.SmartQueryAPI.GeneratorAPI.GenerateStreamID()
}

// SaveFile 保存文件 - Wails API方法
func SaveFile(a *App, fileName, originalName, fileSuffix, md5, localPath string, fileSize int64, relatedID string, fileContentBase64 string) (interface{}, error) {
	return a.SmartQueryAPI.FileAPI.SaveFile(fileName, originalName, fileSuffix, md5, localPath, fileSize, relatedID, fileContentBase64)
}

// GetFileByID 根据ID获取文件信息 - Wails API方法
func GetFileByID(a *App, id string) (interface{}, error) {
	return a.SmartQueryAPI.FileAPI.GetFileByID(id)
}

// GetFilesByRelatedID 根据关联ID获取文件列表 - Wails API方法
func GetFilesByRelatedID(a *App, relatedID string) ([]interface{}, error) {
	files, err := a.SmartQueryAPI.FileAPI.GetFilesByRelatedID(relatedID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(files))
	for i, file := range files {
		result[i] = file
	}
	return result, nil
}

// DeleteFile 删除文件 - Wails API方法
func DeleteFile(a *App, id string) error {
	return a.SmartQueryAPI.FileAPI.DeleteFile(id)
}

// DeleteFilesByRelatedID 根据关联ID删除文件 - Wails API方法
func DeleteFilesByRelatedID(a *App, relatedID string) error {
	return a.SmartQueryAPI.FileAPI.DeleteFilesByRelatedID(relatedID)
}

// GetFilePath 获取文件物理路径 - Wails API方法
func GetFilePath(a *App, fileID string) (string, error) {
	return a.SmartQueryAPI.FileAPI.GetFilePath(fileID)
}

// ProcessFileContent 处理文件内容并转换为markdown格式 - Wails API方法
func ProcessFileContent(a *App, fileID string) error {
	return a.SmartQueryAPI.FileAPI.ProcessFileContent(fileID)
}
