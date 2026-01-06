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
	"skymind/global"
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

// GenerateSystemPrompt 生成系统提示词（通过 Dify API）
func (s *GeneratorService) GenerateSystemPrompt(ctx context.Context, name, description, userInput string) (string, error) {
	logger.LogInfo("Generating system prompt via Dify API", map[string]interface{}{
		"name":        name,
		"description": description,
		"userInput":   userInput,
	})

	// 构建提示词请求内容
	var promptRequest string
	if userInput == "" {
		promptRequest = fmt.Sprintf("请根据以下信息生成一个专业的AI助手系统提示词：\n助手名称：%s\n助手描述：%s\n要求：提示词要以第二人称定义助手的角色和专业领域，明确助手的专业技能、回答风格和行为准则。", name, description)
	} else {
		promptRequest = fmt.Sprintf("请根据以下信息生成一个专业的AI助手系统提示词：\n助手名称：%s\n助手描述：%s\n用户需求：%s\n要求：提示词要以第二人称定义助手的角色和专业领域，明确助手的专业技能、回答风格和行为准则。", name, description, userInput)
	}

	// 调用 Dify API 生成提示词
	return s.callDifyAPI(ctx, promptRequest, "instruct")
}

// OptimizeUserPrompt 优化用户提示词（通过 Dify API）
func (s *GeneratorService) OptimizeUserPrompt(ctx context.Context, originalPrompt string) (string, error) {
	logger.LogInfo("Optimizing user prompt via Dify API", map[string]interface{}{
		"originalLength": len(originalPrompt),
	})

	// 构建优化请求
	optimizeRequest := fmt.Sprintf("请优化以下用户提示词，使其更加清晰、准确和有效。保持用户原意，提升表达清晰度，添加必要的上下文信息，使用更精确的词汇，确保问题结构合理。\n\n原始提示词：%s\n\n请直接返回优化后的提示词，不要包含其他解释。", originalPrompt)

	// 调用 Dify API 优化提示词
	return s.callDifyAPI(ctx, optimizeRequest, "instruct")
}

// GenerateConversationTitle 生成对话标题（通过 Dify API）
func (s *GeneratorService) GenerateConversationTitle(ctx context.Context, userMessage, aiResponse string) (string, error) {
	logger.LogInfo("Generating conversation title via Dify API", map[string]interface{}{
		"userMessageLength": len(userMessage),
		"responseLength":    len(aiResponse),
	})

	// 构建标题生成请求
	titleRequest := fmt.Sprintf("请根据以下对话内容生成一个简洁、准确的对话标题（10-20个字）：\n\n用户输入：%s\nAI回复：%s\n\n请直接返回标题，不要包含其他内容。", userMessage, aiResponse)

	// 调用 Dify API 生成标题
	return s.callDifyAPI(ctx, titleRequest, "fast")
}

// GenerateTopicTitle 生成话题标题（通过 Dify API）
func (s *GeneratorService) GenerateTopicTitle(ctx context.Context, conversationTitles []string) (string, error) {
	logger.LogInfo("Generating topic title via Dify API", map[string]interface{}{
		"conversationCount": len(conversationTitles),
	})

	// 构建话题标题列表文本
	var titlesText string
	for i, title := range conversationTitles {
		if i > 0 {
			titlesText += "\n"
		}
		titlesText += fmt.Sprintf("%d. %s", i+1, title)
	}

	// 构建标题生成请求
	titleRequest := fmt.Sprintf("请根据以下对话标题列表生成一个概括性和吸引力的话题标题（8-15个字）：\n\n对话标题列表：\n%s\n\n请直接返回话题标题，不要包含其他内容。", titlesText)

	// 调用 Dify API 生成标题
	return s.callDifyAPI(ctx, titleRequest, "fast")
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

	// 构建 Dify API 请求数据
	// 从 messages 中提取最后一条用户消息作为 query
	var query string
	var conversationID string
	fileIDs := make(map[string]struct{})

	// 检查 relatedID 是否是有效的 UUID 格式
	// Dify API 要求 conversation_id 必须是有效的 UUID，如果不是则留空让 Dify 自动创建
	if relatedID != "" && s.isValidUUID(relatedID) {
		conversationID = relatedID
	} else {
		// relatedID 不是有效的 UUID，不传 conversation_id，让 Dify 自动创建新会话
		conversationID = ""
		logger.LogInfo("relatedID is not a valid UUID, Dify will create a new conversation", map[string]interface{}{
			"relatedID": relatedID,
		})
	}

	// 提取最后一条用户消息
	for i := len(messages) - 1; i >= 0; i-- {
		if role, ok := messages[i]["role"].(string); ok && role == "user" {
			if content, ok := messages[i]["content"].(string); ok {
				query = content
			}
			// 收集用户消息携带的文件ID（可选）
			if files, ok := messages[i]["files"].([]interface{}); ok {
				logger.LogInfo("找到用户消息中的files字段", map[string]interface{}{
					"files":     files,
					"filesType": fmt.Sprintf("%T", files),
				})
				for _, f := range files {
					if id, ok := f.(string); ok && strings.TrimSpace(id) != "" {
						fileIDs[strings.TrimSpace(id)] = struct{}{}
						logger.LogInfo("收集到文件ID", map[string]interface{}{
							"fileID": strings.TrimSpace(id),
						})
					} else {
						logger.LogInfo("文件ID格式不正确", map[string]interface{}{
							"fileID":     f,
							"fileIDType": fmt.Sprintf("%T", f),
						})
					}
				}
			} else {
				logger.LogInfo("用户消息中没有files字段或格式不正确", map[string]interface{}{
					"hasFiles":     ok,
					"messageIndex": i,
					"messageKeys": func() []string {
						keys := make([]string, 0, len(messages[i]))
						for k := range messages[i] {
							keys = append(keys, k)
						}
						return keys
					}(),
				})
			}
			break
		}
	}

	// 如果找不到用户消息，使用空字符串
	if query == "" {
		logger.LogInfo("No user message found in messages, using empty query", nil)
		query = ""
	}

	// 构建 Dify API 请求数据
	requestData := map[string]interface{}{
		"inputs":        map[string]interface{}{}, // Dify 的输入变量，通常为空
		"query":         query,
		"response_mode": "streaming",
		"user":          streamID, // 使用 streamID 作为用户标识
	}

	// 如果有文件ID，按 Dify 规范附加 files 数组
	// Dify API 要求 files 是对象数组，格式：[{ "type": "image", "transfer_method": "local_file", "upload_file_id": "uuid" }]
	if len(fileIDs) > 0 {
		filesArray := make([]map[string]interface{}, 0, len(fileIDs))

		for difyFileID := range fileIDs {
			// 根据 Dify 文件ID查询数据库（original_path 字段存储的是 Dify 文件ID）
			var file models.File
			err := global.SLDB.Where("original_path = ?", difyFileID).First(&file).Error

			fileType := "document" // 默认类型
			if err == nil {
				// 根据文件扩展名判断类型
				fileType = s.getDifyFileType(file.FileSuffix)
			} else {
				logger.LogInfo("未找到文件信息，使用默认类型", map[string]interface{}{
					"difyFileID": difyFileID,
					"error":      err.Error(),
				})
			}

			fileObj := map[string]interface{}{
				"type":            fileType,
				"transfer_method": "local_file",
				"upload_file_id":  difyFileID,
			}
			filesArray = append(filesArray, fileObj)
		}

		requestData["files"] = filesArray
		logger.LogInfo("附加文件ID到Dify请求", map[string]interface{}{
			"fileCount": len(filesArray),
			"files":     filesArray,
		})
	}

	// 只有当 conversationID 是有效的 UUID 时才添加到请求中
	if conversationID != "" {
		requestData["conversation_id"] = conversationID
	}

	// 注意：Dify API 通过 API Key 自动识别应用，通常不需要 app_id
	// 如果您的 Dify 配置需要 app_id，可以在配置文件的 ID 字段中存储 Dify 应用 ID
	// 然后取消下面的注释：
	// if config.ID != "" {
	// 	requestData["app_id"] = config.ID
	// }

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

	// 记录实际发送的请求体（用于调试）
	logger.LogInfo("发送给 Dify 的完整请求体", map[string]interface{}{
		"requestBody": string(jsonData),
		"apiBase":     apiBase,
	})

	// Dify API 端点
	// 如果 apiBase 已经包含 /v1，则不重复添加
	apiEndpoint := "/chat-messages"
	if strings.HasSuffix(apiBase, "/v1") || strings.HasSuffix(apiBase, "/v1/") {
		apiEndpoint = "/chat-messages"
	} else {
		apiEndpoint = "/v1/chat-messages"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiBase+apiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 对于流式请求，使用更长的超时时间（5分钟）
	// 因为流式响应可能需要较长时间
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	logger.LogInfo("Sending request to Dify API", map[string]interface{}{
		"url":     apiBase + apiEndpoint,
		"timeout": "5 minutes",
	})

	resp, err := client.Do(req)
	if err != nil {
		logger.LogError("Failed to send request to Dify API", err, map[string]interface{}{
			"url":     apiBase + apiEndpoint,
			"timeout": "5 minutes",
		})
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
	// 增加缓冲区大小以支持更长的 SSE 事件行（默认64KB，增加到1MB）
	maxCapacity := 1024 * 1024 // 1MB
	buf := make([]byte, 0, bufio.MaxScanTokenSize)
	scanner.Buffer(buf, maxCapacity)

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

			// 检查是否是结束事件（在转换之前检查）
			if event, ok := data["event"].(string); ok {
				if event == "message_end" || event == "workflow_finished" {
					// 发送结束信号
					runtime.EventsEmit(ctx, "websocket-stream-end", map[string]interface{}{
						"streamID": streamID,
					})
					return nil
				}
			}

			// Dify API 响应格式处理
			// Dify 的响应格式：{"event": "message", "answer": "...", "conversation_id": "...", ...}
			// 需要转换为前端期望的格式
			convertedData := s.convertDifyResponseToOpenAIFormat(data)

			// 处理内容
			if convertedData != nil {
				// 过滤第一个内容块中的开头的空行和空白字符
				if choices, ok := convertedData["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if delta, ok := choice["delta"].(map[string]interface{}); ok {
							if content, ok := delta["content"].(string); ok && isFirstContent && content != "" {
								// 去掉开头的所有空白字符
								delta["content"] = strings.TrimLeftFunc(content, func(r rune) bool {
									return r == ' ' || r == '\t' || r == '\n' || r == '\r'
								})
								isFirstContent = false
							}
						}
					}
				}

				// 标记已接收到数据
				hasReceivedData = true

				// 通过 WebSocket 事件发送流式数据，包含streamID
				runtime.EventsEmit(ctx, "websocket-stream-data", map[string]interface{}{
					"streamID": streamID,
					"data":     convertedData,
				})
			}
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

// convertDifyResponseToOpenAIFormat 将 Dify API 响应转换为 OpenAI 兼容格式
func (s *GeneratorService) convertDifyResponseToOpenAIFormat(difyData map[string]interface{}) map[string]interface{} {
	// Dify 响应格式示例：
	// {"event": "message", "answer": "部分内容", "conversation_id": "...", "message_id": "..."}
	// 需要转换为 OpenAI 格式：
	// {"choices": [{"delta": {"content": "部分内容"}}]}

	converted := map[string]interface{}{
		"choices": []interface{}{
			map[string]interface{}{
				"delta": map[string]interface{}{},
			},
		},
	}

	// 提取 answer 字段作为 content
	if answer, ok := difyData["answer"].(string); ok && answer != "" {
		if choices, ok := converted["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					delta["content"] = answer
				}
			}
		}
	}

	// 提取 metadata 中的 reasoning（如果有）
	if metadata, ok := difyData["metadata"].(map[string]interface{}); ok {
		if reasoning, ok := metadata["reasoning"].(string); ok && reasoning != "" {
			if choices, ok := converted["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						delta["reasoning_content"] = reasoning
					}
				}
			}
		}
	}

	// 如果没有任何内容，返回 nil
	if choices, ok := converted["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if len(delta) == 0 {
					return nil // 没有内容，不发送
				}
			}
		}
	}

	return converted
}

// isValidUUID 检查字符串是否是有效的 UUID 格式
func (s *GeneratorService) isValidUUID(str string) bool {
	// UUID 格式：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (32个十六进制字符，4个连字符)
	if len(str) != 36 {
		return false
	}

	// 检查格式：8-4-4-4-12
	parts := strings.Split(str, "-")
	if len(parts) != 5 {
		return false
	}

	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}

	// 检查每个部分是否都是十六进制字符
	hexChars := "0123456789abcdefABCDEF"
	for _, part := range parts {
		for _, char := range part {
			if !strings.ContainsRune(hexChars, char) {
				return false
			}
		}
	}

	return true
}

// callDifyAPI 非流式调用 Dify API（用于生成标题、优化提示词等）
func (s *GeneratorService) callDifyAPI(ctx context.Context, query, modelType string) (string, error) {
	// 根据模型类型选择模型配置
	var config models.ModelConfig
	switch modelType {
	case "thinking":
		config = database.GetThinkingModelConfig()
	case "fast":
		config = database.GetFastModelConfig()
	case "instruct":
		fallthrough
	default:
		config = database.GetInstructModelConfig()
	}
	apiBase := config.ApiBase
	apiKey := config.ApiKey

	// 构建 Dify API 请求数据（非流式）
	requestData := map[string]interface{}{
		"inputs":        map[string]interface{}{},
		"query":         query,
		"response_mode": "blocking", // 使用阻塞模式，等待完整响应
		"user":          "system",   // 系统调用
	}

	// 构建 API 端点
	apiEndpoint := "/chat-messages"
	if strings.HasSuffix(apiBase, "/v1") || strings.HasSuffix(apiBase, "/v1/") {
		apiEndpoint = "/chat-messages"
	} else {
		apiEndpoint = "/v1/chat-messages"
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiBase+apiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Dify API 返回错误: status=%d, message=%s", resp.StatusCode, string(responseBody))
	}

	// 解析响应
	var difyResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &difyResponse); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取答案
	answer, ok := difyResponse["answer"].(string)
	if !ok {
		// 尝试从其他可能的字段提取
		if choices, ok := difyResponse["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						answer = content
					}
				}
			}
		}
		if answer == "" {
			return "", fmt.Errorf("Dify 响应中未找到答案: %v", difyResponse)
		}
	}

	return strings.TrimSpace(answer), nil
}

// getDifyFileType 根据文件扩展名判断 Dify 文件类型
func (s *GeneratorService) getDifyFileType(fileSuffix string) string {
	fileSuffix = strings.ToLower(strings.TrimSpace(fileSuffix))

	// 图片类型
	imageExtensions := []string{"png", "jpg", "jpeg", "gif", "webp", "bmp", "svg", "ico"}
	for _, ext := range imageExtensions {
		if fileSuffix == ext {
			return "image"
		}
	}

	// 默认返回 document（文档类型）
	// Dify 支持的其他类型可能包括：audio, video 等
	return "document"
}
