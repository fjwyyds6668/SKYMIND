package smart_query

import "skymind/service"

// ApiGroup 智能查询API组
type ApiGroup struct {
	AssistantAPI
	TopicAPI
	ConversationAPI
	MessageAPI
	GeneratorAPI
	FileAPI
}

var (
	assistantService    = service.ServiceGroupApp.SmartQueryServiceGroup.AssistantService
	topicService        = service.ServiceGroupApp.SmartQueryServiceGroup.TopicService
	conversationService = service.ServiceGroupApp.SmartQueryServiceGroup.ConversationService
	messageService      = service.ServiceGroupApp.SmartQueryServiceGroup.MessageService
	generatorService    = service.ServiceGroupApp.SmartQueryServiceGroup.GeneratorService
	fileService         = service.ServiceGroupApp.SmartQueryServiceGroup.FileService
)
