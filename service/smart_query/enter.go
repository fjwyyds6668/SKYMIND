package smart_query

// ServiceGroup 智能查询服务组
type ServiceGroup struct {
	AssistantService
	TopicService
	ConversationService
	MessageService
	GeneratorService
	FileService
}
