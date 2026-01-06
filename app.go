package main

import (
	"context"

	"skymind/appcore"
	"skymind/models"
)

// App struct - 主应用结构体，嵌入appcore.App
type App struct {
	*appcore.App
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		App: appcore.NewApp(),
	}
}

// Wails lifecycle methods - 委托给appcore包
func (a *App) startup(ctx context.Context) {
	appcore.Startup(a.App, ctx)
}

func (a *App) domReady(ctx context.Context) {
	appcore.DomReady(a.App, ctx)
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return appcore.BeforeClose(a.App, ctx)
}

func (a *App) shutdown(ctx context.Context) {
	appcore.Shutdown(a.App, ctx)
}

// System tray methods - 委托给appcore包
func (a *App) InitTray() {
	appcore.InitTray(a.App)
}

// Window management methods - 委托给appcore包
func (a *App) ShowWindow() {
	appcore.ShowWindow(a.App)
}

func (a *App) HideWindow() {
	appcore.HideWindow(a.App)
}

func (a *App) QuitApp() {
	appcore.QuitApp(a.App)
}

func (a *App) TerminateProcess() {
	appcore.TerminateProcess(a.App)
}

// Windows API methods - 委托给appcore包
func (a *App) CheckSingleInstance() bool {
	return appcore.CheckSingleInstance(a.App)
}

func (a *App) ReleaseMutex() {
	appcore.ReleaseMutex(a.App)
}

// Utility methods - 委托给appcore包
func (a *App) Greet(name string) string {
	return appcore.Greet(a.App, name)
}

// Database initialization - 委托给appcore包
func (a *App) initDatabase() error {
	return appcore.InitDatabase(a.App)
}

// API methods - 委托给appcore包
func (a *App) GetAssistants() ([]interface{}, error) {
	return appcore.GetAssistants(a.App)
}

func (a *App) GetAssistantByID(id string) (interface{}, error) {
	return appcore.GetAssistantByID(a.App, id)
}

func (a *App) CreateAssistant(assistant map[string]interface{}) (interface{}, error) {
	return appcore.CreateAssistant(a.App, assistant)
}

func (a *App) UpdateAssistant(assistant map[string]interface{}) error {
	return appcore.UpdateAssistant(a.App, assistant)
}

func (a *App) DeleteAssistant(id string) error {
	return appcore.DeleteAssistant(a.App, id)
}

func (a *App) GetDefaultAssistant() (interface{}, error) {
	return appcore.GetDefaultAssistant(a.App)
}

func (a *App) GetTopics(assistantID string) ([]interface{}, error) {
	return appcore.GetTopics(a.App, assistantID)
}

func (a *App) CreateTopic(topic map[string]interface{}) (interface{}, error) {
	return appcore.CreateTopic(a.App, topic)
}

func (a *App) UpdateTopic(topic map[string]interface{}) error {
	return appcore.UpdateTopic(a.App, topic)
}

func (a *App) UpdateTopicTitle(id, title string) error {
	return appcore.UpdateTopicTitle(a.App, id, title)
}

func (a *App) DeleteTopic(id string, deleteTopic bool) error {
	return appcore.DeleteTopic(a.App, id, deleteTopic)
}

func (a *App) GetMessages(topicID string) ([]interface{}, error) {
	return appcore.GetMessages(a.App, topicID)
}

func (a *App) GetConversations(topicID string) ([]interface{}, error) {
	return appcore.GetConversations(a.App, topicID)
}

func (a *App) CreateConversation(conversation map[string]interface{}) (string, error) {
	return appcore.CreateConversation(a.App, conversation)
}

func (a *App) UpdateConversationSettings(id, settings string) error {
	return appcore.UpdateConversationSettings(a.App, id, settings)
}

func (a *App) UpdateConversationTitle(id, title string) error {
	return appcore.UpdateConversationTitle(a.App, id, title)
}

func (a *App) GenerateConversationTitle(userMessage, aiResponse string) (string, error) {
	return appcore.GenerateConversationTitle(a.App, userMessage, aiResponse)
}

func (a *App) CreateMessage(message map[string]interface{}) (interface{}, error) {
	return appcore.CreateMessage(a.App, message)
}

func (a *App) DeleteMessage(id, conversationID string) error {
	return appcore.DeleteMessage(a.App, id, conversationID)
}

func (a *App) UpdateMessage(id, content, reasoning string) error {
	return appcore.UpdateMessage(a.App, id, content, reasoning)
}

func (a *App) StreamChatCompletion(streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
	return appcore.StreamChatCompletion(a.App, streamID, streamType, relatedID, messages, modelType)
}

func (a *App) StopStreamChatCompletion(streamID string) {
	appcore.StopStreamChatCompletion(a.App, streamID)
}

func (a *App) StopAllStreams() {
	appcore.StopAllStreams(a.App)
}

func (a *App) GetActiveStreams() map[string]*models.StreamInfo {
	return appcore.GetActiveStreams(a.App)
}

func (a *App) IsStreamActive(streamID string) bool {
	return appcore.IsStreamActive(a.App, streamID)
}

func (a *App) UpdateAssistantsSortOrder(sortOrders []map[string]interface{}) error {
	return appcore.UpdateAssistantsSortOrder(a.App, sortOrders)
}

func (a *App) UpdateTopicsSortOrder(sortOrders []map[string]interface{}) error {
	return appcore.UpdateTopicsSortOrder(a.App, sortOrders)
}

func (a *App) DeleteConversationsAfter(conversationID string) error {
	return appcore.DeleteConversationsAfter(a.App, conversationID)
}

// Generator API methods - 委托给appcore包
func (a *App) GenerateSystemPrompt(name, description, userInput string) (string, error) {
	return appcore.GenerateSystemPrompt(a.App, name, description, userInput)
}

func (a *App) OptimizeUserPrompt(originalPrompt string) (string, error) {
	return appcore.OptimizeUserPrompt(a.App, originalPrompt)
}

func (a *App) GenerateTopicTitle(conversationTitles []string) (string, error) {
	return appcore.GenerateTopicTitle(a.App, conversationTitles)
}

func (a *App) GenerateStreamID() (string, error) {
	return appcore.GenerateStreamID(a.App)
}

// File API methods - 委托给appcore包
func (a *App) SaveFile(fileName, originalName, fileSuffix, md5, localPath string, fileSize int64, relatedID string, fileContentBase64 string) (interface{}, error) {
	return appcore.SaveFile(a.App, fileName, originalName, fileSuffix, md5, localPath, fileSize, relatedID, fileContentBase64)
}

func (a *App) GetFileByID(id string) (interface{}, error) {
	return appcore.GetFileByID(a.App, id)
}

func (a *App) GetFilesByRelatedID(relatedID string) ([]interface{}, error) {
	return appcore.GetFilesByRelatedID(a.App, relatedID)
}

func (a *App) DeleteFile(id string) error {
	return appcore.DeleteFile(a.App, id)
}

func (a *App) DeleteFilesByRelatedID(relatedID string) error {
	return appcore.DeleteFilesByRelatedID(a.App, relatedID)
}

func (a *App) GetFilePath(fileID string) (string, error) {
	return appcore.GetFilePath(a.App, fileID)
}

func (a *App) ProcessFileContent(fileID string) error {
	return appcore.ProcessFileContent(a.App, fileID)
}
