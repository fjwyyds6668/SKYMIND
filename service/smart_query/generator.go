package smart_query

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"skymind/database"
	"skymind/logger"
	"skymind/models"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GeneratorService AI生成服务
type GeneratorService struct {
	streams      map[string]*models.StreamInfo
	streamsMutex sync.RWMutex
}

// init 初始化方法
func (s *GeneratorService) init() {
	if s.streams == nil {
		s.streams = make(map[string]*models.StreamInfo)
	}
}

// GenerateSystemPrompt 生成系统提示词（使用InstructModel）
func (s *GeneratorService) GenerateSystemPrompt(ctx context.Context, name, description, userInput string) (string, error) {
	logger.LogInfo("Generating system prompt", map[string]interface{}{
		"name":        name,
		"description": description,
		"userInput":   userInput,
	})

	// 生成提示词
	promptGenerator := models.PromptGenerator{}
	generatedPrompt := promptGenerator.SystemPromptGenerator(name, description, userInput)

	return generatedPrompt, nil
}

// OptimizeUserPrompt 优化用户提示词（使用InstructModel）
func (s *GeneratorService) OptimizeUserPrompt(ctx context.Context, originalPrompt string) (string, error) {
	logger.LogInfo("Optimizing user prompt", map[string]interface{}{
		"originalLength": len(originalPrompt),
	})

	// 生成提示词
	promptGenerator := models.PromptGenerator{}
	generatedPrompt := promptGenerator.UserPromptOptimizer(originalPrompt)
	return generatedPrompt, nil
}

// GenerateConversationTitle 生成对话标题
func (s *GeneratorService) GenerateConversationTitle(ctx context.Context, userMessage, aiResponse string) (string, error) {
	logger.LogInfo("Generating conversation title", map[string]interface{}{
		"userMessageLength": len(userMessage),
		"responseLength":    len(aiResponse),
	})

	// 生成提示词
	promptGenerator := models.PromptGenerator{}
	generatedPrompt := promptGenerator.ConversationTitleGenerator(userMessage, aiResponse)

	return generatedPrompt, nil
}

// GenerateTopicTitle 生成话题标题
func (s *GeneratorService) GenerateTopicTitle(ctx context.Context, conversationTitles []string) (string, error) {
	logger.LogInfo("Generating topic title", map[string]interface{}{
		"conversationCount": len(conversationTitles),
	})

	// 生成提示词
	promptGenerator := models.PromptGenerator{}
	generatedPrompt := promptGenerator.TopicTitleGenerator(conversationTitles)

	return generatedPrompt, nil
}

// StreamChatCompletion 流式调用大模型
func (s *GeneratorService) StreamChatCompletion(ctx context.Context, streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
	logger.LogInfo("Starting stream chat completion", map[string]interface{}{
		"streamID":     streamID,
		"streamType":   streamType,
		"relatedID":    relatedID,
		"messageCount": len(messages),
		"modelType":    modelType,
	})

	// 确保streams map已初始化
	s.init()

	// 创建可取消的上下文
	streamCtx, cancel := context.WithCancel(ctx)

	// 创建流式信息
	streamInfo := &models.StreamInfo{
		ID:        streamID,
		Type:      streamType,
		RelatedID: relatedID,
		Ctx:       streamCtx,
		Cancel:    cancel,
		StartTime: time.Now(),
	}

	// 添加到流式输出映射
	s.streamsMutex.Lock()
	s.streams[streamID] = streamInfo
	s.streamsMutex.Unlock()

	// 根据模型类型选择模型配置
	var config models.ModelConfig

	switch modelType {
	case "thinking":
		config = database.GetThinkingModelConfig()
		logger.LogInfo("Using thinking model", map[string]interface{}{"model": config.ID})
	case "fast":
		config = database.GetFastModelConfig()
		logger.LogInfo("Using fast model", map[string]interface{}{"model": config.ID})
	case "instruct":
		fallthrough
	default:
		config = database.GetInstructModelConfig()
		logger.LogInfo("Using instruct model", map[string]interface{}{"model": config.ID})
	}
	apiBase := config.ApiBase
	apiKey := config.ApiKey
	model := config.ID

	// 构建请求数据
	requestData := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}

	// 根据模型类型添加特定的思考控制参数
	switch modelType {
	case "thinking":
		requestData["thinking"] = map[string]interface{}{
			"type": "enabled",
		}

		logger.LogInfo("Enabled thinking mode for model", map[string]interface{}{
			"model":     model,
			"modelType": modelType,
		})
	case "instruct":
		requestData["thinking"] = map[string]interface{}{
			"type": "disabled",
		}
		logger.LogInfo("Disabled thinking mode for model", map[string]interface{}{
			"model":     model,
			"modelType": modelType,
		})
	}

	// 发送HTTP请求
	resp, err := s.sendStreamRequestWithCancel(streamCtx, apiBase, apiKey, requestData)
	if err != nil {
		// 清理流式信息
		s.streamsMutex.Lock()
		delete(s.streams, streamID)
		s.streamsMutex.Unlock()
		logger.LogError("Failed to send stream request", err, map[string]interface{}{
			"streamID": streamID,
			"model":    model,
			"apiBase":  apiBase,
		})
		return err
	}
	defer resp.Body.Close()

	logger.LogInfo("Stream request sent successfully", map[string]interface{}{
		"streamID":   streamID,
		"statusCode": resp.StatusCode,
		"model":      model,
	})

	// 处理流式响应并通过 WebSocket 发送
	err = s.handleStreamResponseWithWebSocket(streamCtx, resp, streamID)

	// 清理流式信息
	s.streamsMutex.Lock()
	delete(s.streams, streamID)
	s.streamsMutex.Unlock()

	if err != nil {
		logger.LogError("Stream chat completion failed", err, map[string]interface{}{
			"streamID": streamID,
		})
	} else {
		logger.LogInfo("Stream chat completion completed successfully", map[string]interface{}{
			"streamID": streamID,
		})
	}

	return err
}

// sendStreamRequestWithCancel 发送支持取消的流式HTTP请求
func (s *GeneratorService) sendStreamRequestWithCancel(ctx context.Context, apiBase, apiKey string, requestData map[string]interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiBase+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error! status: %d, message: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// handleStreamResponseWithWebSocket 使用 WebSocket 处理流式响应
func (s *GeneratorService) handleStreamResponseWithWebSocket(ctx context.Context, resp *http.Response, streamID string) error {
	scanner := bufio.NewScanner(resp.Body)
	hasReceivedData := false
	isFirstContent := true // 标记是否是第一个内容块

	for scanner.Scan() {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			// 上下文被取消，发送停止信号并返回特殊错误表示用户主动停止
			runtime.EventsEmit(ctx, "websocket-stream-end", map[string]interface{}{
				"streamID": streamID,
			})
			return context.Canceled
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// 处理 SSE 格式
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")

			if dataStr == "[DONE]" {
				// 使用 Wails 事件发送结束信号
				runtime.EventsEmit(ctx, "websocket-stream-end", map[string]interface{}{
					"streamID": streamID,
				})
				return nil
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
				continue // 忽略解析错误，继续处理下一行
			}

			// 处理选择项
			if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						// 过滤第一个内容块中的开头的空行和空白字符
						if content, ok := delta["content"].(string); ok && isFirstContent {
							// 去掉开头的所有空白字符（包括空格、制表符、换行符等）
							delta["content"] = strings.TrimLeftFunc(content, func(r rune) bool {
								return r == ' ' || r == '\t' || r == '\n' || r == '\r'
							})
							isFirstContent = false
						}
						
						// 同样处理 reasoning_content
						if reasoningContent, ok := delta["reasoning_content"].(string); ok && isFirstContent {
							delta["reasoning_content"] = strings.TrimLeftFunc(reasoningContent, func(r rune) bool {
								return r == ' ' || r == '\t' || r == '\n' || r == '\r'
							})
						}
					}
				}
			}

			// 标记已接收到数据
			hasReceivedData = true

			// 通过 WebSocket 事件发送流式数据，包含streamID
			runtime.EventsEmit(ctx, "websocket-stream-data", map[string]interface{}{
				"streamID": streamID,
				"data":     data,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		// 如果已经接收到数据，发送结束信号让前端保存
		if hasReceivedData {
			runtime.EventsEmit(ctx, "websocket-stream-end", map[string]interface{}{
				"streamID": streamID,
			})
		} else {
			runtime.EventsEmit(ctx, "websocket-stream-error", map[string]interface{}{
				"streamID": streamID,
				"error":    fmt.Sprintf("stream reading error: %v", err),
			})
		}
		return fmt.Errorf("stream reading error: %v", err)
	}

	// 正常结束，发送结束信号
	runtime.EventsEmit(ctx, "websocket-stream-end", map[string]interface{}{
		"streamID": streamID,
	})
	return nil
}

// StopStreamChatCompletion 停止指定的流式聊天
func (s *GeneratorService) StopStreamChatCompletion(streamID string) {
	s.init()
	s.streamsMutex.Lock()
	defer s.streamsMutex.Unlock()

	if streamInfo, exists := s.streams[streamID]; exists {
		if cancel, ok := streamInfo.Cancel.(context.CancelFunc); ok {
			cancel()
		}
		delete(s.streams, streamID)
		logger.LogInfo("Stream stopped", map[string]interface{}{
			"streamID": streamID,
		})
	}
}

// StopAllStreams 停止所有流式聊天
func (s *GeneratorService) StopAllStreams() {
	s.init()
	s.streamsMutex.Lock()
	defer s.streamsMutex.Unlock()

	for streamID, streamInfo := range s.streams {
		if cancel, ok := streamInfo.Cancel.(context.CancelFunc); ok {
			cancel()
		}
		logger.LogInfo("Stream stopped", map[string]interface{}{
			"streamID": streamID,
		})
	}

	// 清空所有流式输出
	s.streams = make(map[string]*models.StreamInfo)
}

// GetActiveStreams 获取活跃的流式输出列表
func (s *GeneratorService) GetActiveStreams() map[string]*models.StreamInfo {
	s.init()
	s.streamsMutex.RLock()
	defer s.streamsMutex.RUnlock()

	// 返回副本避免并发问题
	result := make(map[string]*models.StreamInfo)
	for id, stream := range s.streams {
		result[id] = stream
	}
	return result
}

// IsStreamActive 检查指定流式输出是否活跃
func (s *GeneratorService) IsStreamActive(streamID string) bool {
	s.init()
	s.streamsMutex.RLock()
	defer s.streamsMutex.RUnlock()

	_, exists := s.streams[streamID]
	return exists
}

// GenerateStreamID 生成流式输出ID
func (s *GeneratorService) GenerateStreamID() (string, error) {
	return database.GenerateIDString()
}
