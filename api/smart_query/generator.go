package smart_query

import (
	"context"
	"skymind/models"
)

// GeneratorAPI AI生成API
type GeneratorAPI struct {
}

// GenerateSystemPrompt 生成系统提示词
func (api *GeneratorAPI) GenerateSystemPrompt(name, description, userInput string) (string, error) {
	return generatorService.GenerateSystemPrompt(context.Background(), name, description, userInput)
}

// OptimizeUserPrompt 优化用户提示词
func (api *GeneratorAPI) OptimizeUserPrompt(originalPrompt string) (string, error) {
	return generatorService.OptimizeUserPrompt(context.Background(), originalPrompt)
}

// GenerateConversationTitle 生成对话标题
func (api *GeneratorAPI) GenerateConversationTitle(userMessage, aiResponse string) (string, error) {
	return generatorService.GenerateConversationTitle(context.Background(), userMessage, aiResponse)
}

// GenerateTopicTitle 生成话题标题
func (api *GeneratorAPI) GenerateTopicTitle(conversationTitles []string) (string, error) {
	return generatorService.GenerateTopicTitle(context.Background(), conversationTitles)
}

// StreamChatCompletion 流式调用大模型
func (api *GeneratorAPI) StreamChatCompletion(ctx context.Context, streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
	return generatorService.StreamChatCompletion(ctx, streamID, streamType, relatedID, messages, modelType)
}

// StopStreamChatCompletion 停止指定的流式聊天
func (api *GeneratorAPI) StopStreamChatCompletion(streamID string) {
	generatorService.StopStreamChatCompletion(streamID)
}

// StopAllStreams 停止所有流式聊天
func (api *GeneratorAPI) StopAllStreams() {
	generatorService.StopAllStreams()
}

// GetActiveStreams 获取活跃的流式输出列表
func (api *GeneratorAPI) GetActiveStreams() map[string]*models.StreamInfo {
	return generatorService.GetActiveStreams()
}

// IsStreamActive 检查指定流式输出是否活跃
func (api *GeneratorAPI) IsStreamActive(streamID string) bool {
	return generatorService.IsStreamActive(streamID)
}

// GenerateStreamID 生成流式输出ID
func (api *GeneratorAPI) GenerateStreamID() (string, error) {
	return generatorService.GenerateStreamID()
}
