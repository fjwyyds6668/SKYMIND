# 并行流式输出实现文档

## 概述

本项目实现了一套完整的并行流式输出系统，支持同时进行多个不同类型的流式输出操作，包括聊天对话、标题生成、提示词优化等。该系统通过 StreamID 唯一标识每个流式输出，结合后端协程并发处理和前端全局状态管理，实现了无冲突的多流式输出并行处理。

## 核心架构

### 1. 整体架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端组件     │    │  Stream Store  │    │   WebSocket    │
│                │    │                │    │   事件监听     │
│ Chat.vue       │◄──►│                │◄──►│                │
│ Settings.vue   │    │                │    │                │
│ Index.vue      │    │                │    │                │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Wails API    │    │   后端 API     │    │   协程池       │
│                │    │                │    │                │
│ StreamChat...  │◄──►│                │◄──►│                │
│ StopStream...  │    │                │    │                │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   大模型 API    │
                       │                │
                       │ OpenAI/其他     │
                       └─────────────────┘
```

### 2. 核心组件

#### 2.1 StreamID 生成器
- **位置**: `global/snowflake.go`
- **功能**: 使用短雪花算法生成唯一的流式输出标识符
- **特点**: 保证分布式环境下的唯一性，支持高并发

#### 2.2 Stream 模型
- **位置**: `models/stream.go`
- **功能**: 定义流式输出的数据结构
- **字段**: ID、类型、状态、内容、元数据等

#### 2.3 Stream Store
- **位置**: `frontend/src/store/modules/stream.js`
- **功能**: 前端全局状态管理，管理所有活跃的流式输出
- **特点**: 基于 Pinia，支持响应式更新

## 前端实现详解

### 1. 流式输出不中断的核心原理

#### 1.1 全局状态管理的设计理念

本项目的流式输出不中断机制基于一个核心设计理念：**流式输出的生命周期与UI组件的生命周期解耦**。

传统做法的问题：
- 流式输出状态与特定组件绑定
- 组件切换时状态丢失
- 无法支持跨组件的并发流式输出

本项目的解决方案：
- 流式输出状态存储在全局 Store 中
- 组件仅作为流式输出的展示层
- 通过 StreamID 和 relatedId 建立关联关系

#### 1.2 关键设计决策

```javascript
// ❌ 传统做法：组件级状态管理
const ChatComponent = {
  data() {
    return {
      isStreaming: false,  // 组件级状态
      streamContent: '',
      streamId: null
    }
  }
}

// ✅ 本项目做法：全局状态管理
const streamStore = useStreamStore();
// 流式输出状态独立于任何组件存在
```

### 2. Stream Store 核心逻辑

#### 1.1 流式输出类型定义

```javascript
export const StreamType = {
  CHAT: 'chat',                           // 聊天对话
  CONVERSATION_TITLE_GENERATION: 'conversation_title_generation',  // 对话标题生成
  TOPIC_TITLE_GENERATION: 'topic_title_generation',  // 话题标题生成
  PROMPT_OPTIMIZATION: 'prompt_optimization',  // 提示词优化
  SYSTEM_PROMPT: 'system_prompt',  // 系统提示词生成
};
```

#### 1.2 流式输出状态管理

```javascript
export const StreamStatus = {
  IDLE: 'idle',           // 空闲
  CONNECTING: 'connecting', // 连接中
  STREAMING: 'streaming',  // 流式输出中
  COMPLETED: 'completed',  // 已完成
  ERROR: 'error',         // 错误
  STOPPED: 'stopped',     // 已停止
};
```

#### 1.3 核心数据结构

```javascript
// 活跃的流式输出列表 - 使用 Map 保证 O(1) 查找性能
const activeStreams = ref(new Map());

// 流式输出对象结构
const stream = {
  id: streamId,                    // 唯一标识符
  type: streamType,                // 流式输出类型
  relatedId: relatedId,            // 关联的实体ID（对话、话题、助手等）
  status: StreamStatus.IDLE,        // 当前状态
  content: '',                     // 流式内容
  reasoning: '',                   // 思考过程（深度思考模式）
  metadata: {},                    // 元数据
  startTime: null,                 // 开始时间
  endTime: null,                   // 结束时间
  error: null,                     // 错误信息
};
```

### 2. WebSocket 事件处理

#### 2.1 事件监听器初始化

```javascript
const initWebSocketListeners = () => {
  if (window.runtime) {
    window.runtime.EventsOn("websocket-stream-data", handleWebSocketStreamData);
    window.runtime.EventsOn("websocket-stream-end", handleWebSocketStreamEnd);
    window.runtime.EventsOn("websocket-stream-error", handleWebSocketStreamError);
  }
};
```

#### 2.2 流式数据处理

```javascript
const handleWebSocketStreamData = (event) => {
  const { streamID, data } = event;
  
  if (!data || !data.choices || data.choices.length === 0) return;

  const choice = data.choices[0];
  const delta = choice.delta;
  const stream = activeStreams.value.get(streamID);
  
  if (!stream) return;

  // 根据流式输出类型更新内容
  if (delta.content && delta.content !== "") {
    updateStreamContent(streamID, delta.content, delta.reasoning_content || "");
  }
};
```

### 3. 流式输出生命周期管理

#### 3.1 创建流式输出

```javascript
const createStream = async (streamType, relatedId, metadata = {}) => {
  // 生成唯一 StreamID
  const streamId = await generateStreamId();
  
  // 创建流式输出对象
  const stream = {
    id: streamId,
    type: streamType,
    relatedId: relatedId,
    status: StreamStatus.IDLE,
    content: '',
    reasoning: '',
    metadata: metadata,
    startTime: null,
    endTime: null,
    error: null,
  };
  
  // 添加到活跃列表
  activeStreams.value.set(streamId, stream);
  return streamId;
};
```

#### 3.2 开始流式输出

```javascript
const startStream = (streamId) => {
  const stream = activeStreams.value.get(streamId);
  if (stream) {
    stream.status = StreamStatus.STREAMING;
    stream.startTime = new Date();
    stream.content = '';
    stream.reasoning = '';
    stream.error = null;
  }
};
```

#### 3.3 更新流式内容

```javascript
const updateStreamContent = (streamId, content, reasoning = '') => {
  const stream = activeStreams.value.get(streamId);
  if (stream) {
    if (content) {
      stream.content += content;
    }
    if (reasoning) {
      stream.reasoning += reasoning;
    }
  }
};
```

#### 3.4 完成流式输出

```javascript
const completeStream = (streamId) => {
  const stream = activeStreams.value.get(streamId);
  if (stream) {
    stream.status = StreamStatus.COMPLETED;
    stream.endTime = new Date();
    
    // 延迟移除，给UI时间显示完成状态
    setTimeout(() => {
      activeStreams.value.delete(streamId);
    }, 1000);
  }
};
```

### 4. 前端组件流式输出绑定机制详解

#### 4.1 核心绑定原理

本项目的流式输出绑定机制基于**观察者模式 + 响应式数据流**，实现了流式输出状态与UI元素的精确绑定。

#### 4.2 组件与Stream Store的交互模式

```javascript
// 组件通过计算属性获取特定类型的流式输出
const isStreamingChat = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.CHAT);
});

const currentSystemPromptStream = computed(() => {
  const streams = streamStore.getActiveStreamsByType(StreamType.SYSTEM_PROMPT);
  return streams.length > 0 ? streams[0] : null;
});

// 组件通过监听器响应流式输出变化
watch(
  () => streamStore.activeStreams,
  (newStreams) => {
    // 根据流式输出类型和状态更新对应的UI元素
  },
  { deep: true }
);
```

#### 4.3 AI回复内容绑定 - reply.vue

**组件职责**: 专门负责显示AI回复内容，包括思考过程和最终回复。

**绑定机制**:
```javascript
// reply.vue 中的流式输出绑定
const currentMessage = computed(() => {
  return messagesData.value[currentIndex.value] || {};
});

// 通过props接收流式输出状态
const props = defineProps({
  isStreamingChat: {
    type: Boolean,
    default: false,
  },
  isActive: {
    type: Boolean,
    default: false,
  },
});

// 流式输出状态控制UI显示
<t-chat-item
  :loading="isStreamingChat && !currentMessage.content"
  :class="{ 'assistant-chat-item': true }"
>
  <!-- 思考过程显示 -->
  <ChatReasoning
    v-if="currentMessage.reasoning && currentMessage.reasoning.length > 0"
    :content="renderReasoningContent(currentMessage.reasoning)"
  />
  
  <!-- 回复内容显示 -->
  <t-chat-content
    v-if="currentMessage.content && currentMessage.content.length > 0"
    :content="currentMessage.content"
    role="assistant"
  />
</t-chat-item>
```

**数据流向**:
```
Stream Store 流式数据
    ↓
Chat.vue 监听器接收
    ↓
更新 conversationList.value
    ↓
传递给 reply.vue (props)
    ↓
currentMessage 计算属性更新
    ↓
UI 实时显示流式内容
```

#### 4.4 系统提示词绑定 - settings.vue

**组件职责**: 助手设置对话框，包含系统提示词的流式生成和显示。

**绑定机制**:
```javascript
// settings.vue 中的流式输出绑定
const isGeneratingPrompt = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.SYSTEM_PROMPT);
});

const currentSystemPromptStream = computed(() => {
  const streams = streamStore.getActiveStreamsByType(StreamType.SYSTEM_PROMPT);
  return streams.length > 0 ? streams[0] : null;
});

// 监听系统提示词流式输出内容变化
watch(
  () => currentSystemPromptStream.value,
  (stream) => {
    if (stream && stream.content !== undefined) {
      formData.prompt = stream.content;
    }
  },
  { deep: true }
);

// UI绑定
<t-textarea
  v-model="formData.prompt"
  placeholder="请输入系统提示词，用于定义助手的角色和行为"
  :autosize="{ minRows: 10, maxRows: 15 }"
  class="form-textarea"
/>

<t-button
  @click="handleGeneratePrompt"
  :loading="isGeneratingPrompt"
  :disabled="!formData.name.trim() || !formData.description.trim() || isGeneratingPrompt"
>
  <div class="button-content">
    <IconSparkles v-if="!isGeneratingPrompt" :size="16" />
    <span>优化提示词</span>
  </div>
</t-button>
```

**数据流向**:
```
用户点击"优化提示词"
    ↓
创建 SYSTEM_PROMPT 流式输出
    ↓
Stream Store 处理流式数据
    ↓
settings.vue 监听器接收
    ↓
更新 formData.prompt
    ↓
textarea 实时显示流式内容
```

#### 4.5 用户提示词优化绑定 - chat.vue

**组件职责**: 聊天界面，包含用户输入框和提示词优化功能。

**绑定机制**:
```javascript
// chat.vue 中的提示词优化流式输出绑定
const isOptimizingPrompt = ref(false);

// 监听流式输出内容变化
watch(
  () => streamStore.activeStreams,
  (newStreams) => {
    for (const [streamId, stream] of newStreams) {
      if (stream.type === StreamType.PROMPT_OPTIMIZATION) {
        // 优化提示词模式：将内容显示在输入框
        if (stream.content !== undefined) {
          inputText.value = stream.content;
        }
        
        // 流式输出完成时保持内容不被清空
        if (stream.status === StreamStatus.COMPLETED) {
          const finalContent = stream.content || inputText.value;
          if (finalContent && finalContent.trim()) {
            nextTick(() => {
              inputText.value = finalContent;
            });
          }
        }
      }
    }
  },
  { deep: true }
);

// UI绑定
<t-chat-sender
  ref="chatSenderRef"
  v-model="inputText"
  :loading="isStreamingChat || isOptimizingPrompt"
  :disabled="isStreamingChat || !selectedAssistantData || !selectedTopicData"
  @send="handleSendMessage"
  @stop="handleStopChat"
>
  <template #prefix>
    <t-button
      @click="handleOptimizePrompt"
      :loading="isOptimizingPrompt"
      :disabled="!selectedAssistantData || isOptimizingPrompt || !inputText.trim()"
    >
      <div class="button-content">
        <IconSparkles v-if="!isOptimizingPrompt" :size="16" />
        <span>优化提示词</span>
      </div>
    </t-button>
  </template>
</t-chat-sender>
```

**数据流向**:
```
用户点击"优化提示词"
    ↓
创建 PROMPT_OPTIMIZATION 流式输出
    ↓
Stream Store 处理流式数据
    ↓
chat.vue 监听器接收
    ↓
更新 inputText.value
    ↓
t-chat-sender 输入框实时显示流式内容
```

#### 4.6 聊天对话流式输出绑定 - chat.vue

**组件职责**: 处理聊天对话的流式输出，包括AI回复和思考过程。

**绑定机制**:
```javascript
// chat.vue 中的聊天流式输出绑定
const isStreamingChat = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.CHAT);
});

const currentChatStream = computed(() => {
  const chatStreams = streamStore.getActiveStreamsByType(StreamType.CHAT);
  return chatStreams.length > 0 ? chatStreams[0] : null;
});

// 监听流式输出内容变化
watch(
  () => streamStore.activeStreams,
  (newStreams) => {
    for (const [streamId, stream] of newStreams) {
      if (stream.type === StreamType.CHAT) {
        // 聊天模式：更新AI消息内容
        const currentConversation = conversationList.value[conversationList.value.length - 1];
        if (currentConversation && currentConversation.messages && currentConversation.messages.length > 0) {
          const lastMessage = currentConversation.messages[currentConversation.messages.length - 1];
          if (lastMessage && lastMessage.role === "assistant" && lastMessage.id === stream.metadata.aiMessageId) {
            // 更新内容和思考过程
            if (stream.reasoning !== undefined) {
              lastMessage.reasoning = stream.reasoning;
            }
            if (stream.content !== undefined) {
              lastMessage.content = stream.content;
            }
            
            // 在流式输出期间，如果用户在底部，立即滚动到底部
            if (stream.status === StreamStatus.STREAMING && !isShowToBottom.value) {
              nextTick(() => {
                scrollToConversationBottom();
              });
            }
          }
        }
      }
    }
  },
  { deep: true }
);
```

**数据流向**:
```
用户发送消息
    ↓
创建 CHAT 流式输出
    ↓
Stream Store 处理流式数据
    ↓
chat.vue 监听器接收
    ↓
更新 conversationList.value[].messages
    ↓
传递给 reply.vue (props)
    ↓
reply.vue 显示流式AI回复
```

#### 4.7 话题标题生成绑定 - index.vue

**组件职责**: 侧边栏，管理助手和话题列表，处理话题标题的流式生成。

**绑定机制**:
```javascript
// index.vue 中的话题标题生成绑定
watch(
  () => streamStore.activeStreamsList,
  (streams) => {
    // 查找话题标题生成的流式输出
    const topicTitleStreams = streams.filter(stream => 
      stream.type === StreamType.TOPIC_TITLE_GENERATION && 
      stream.status === StreamStatus.STREAMING
    );

    topicTitleStreams.forEach(stream => {
      // 找到对应的话题并更新标题
      const topicIndex = topics.value.findIndex(topic => topic.id === stream.relatedId);
      if (topicIndex !== -1 && stream.content) {
        topics.value[topicIndex].name = stream.content;
      }
    });
  },
  { deep: true }
);
```

**数据流向**:
```
对话标题生成完成
    ↓
触发话题标题生成
    ↓
创建 TOPIC_TITLE_GENERATION 流式输出
    ↓
Stream Store 处理流式数据
    ↓
index.vue 监听器接收
    ↓
更新 topics.value[].name
    ↓
话题列表实时显示新标题
```

### 5. 切换助手和话题时流式输出不中断的实现

#### 5.1 关键设计原则

**原则1**: 流式输出状态与UI组件状态完全分离
- 流式输出状态存储在全局 Store 中
- UI组件仅作为展示层，不持有流式输出状态

**原则2**: 通过关联ID建立绑定关系
- StreamID: 唯一标识流式输出实例
- relatedId: 关联到具体的业务实体（对话、话题、助手）
- metadata: 包含额外的绑定信息

#### 5.2 具体实现机制

**机制1**: 全局状态持久化
```javascript
// Stream Store 中的流式输出对象
const stream = {
  id: streamId,                    // 全局唯一标识
  type: streamType,                // 流式输出类型
  relatedId: relatedId,            // 关联的实体ID
  status: StreamStatus.STREAMING,   // 独立的状态管理
  content: '',                     // 流式内容
  metadata: {},                    // 绑定元数据
  startTime: new Date(),           // 开始时间
  endTime: null,                   // 结束时间
};

// 存储在全局 Map 中，不依赖任何组件
const activeStreams = ref(new Map());
```

**机制2**: 组件切换时的状态保持
```javascript
// 当用户切换助手或话题时
const selectAssistant = async (assistant, isInit) => {
  // 1. 更新UI状态
  selectedAssistant.value = assistant;
  selectedTopic.value = null;
  
  // 2. 流式输出状态不受影响
  // Stream Store 中的流式输出继续运行
  // 不会因为组件状态变化而中断
  
  // 3. 新组件可以访问到正在进行的流式输出
  const activeStreams = streamStore.activeStreamsList;
  for (const stream of activeStreams) {
    if (stream.type === StreamType.CHAT) {
      // 可以继续显示正在进行的聊天流式输出
      console.log('Chat stream continues:', stream.id);
    }
  }
};
```

**机制3**: 跨组件状态同步
```javascript
// 不同组件可以访问相同的流式输出状态
// Chat.vue 中
const isStreamingChat = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.CHAT);
});

// Index.vue 中
const hasActiveChat = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.CHAT);
});

// Settings.vue 中
const isGeneratingPrompt = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.PROMPT_OPTIMIZATION);
});
```

#### 5.3 实际运行流程示例

**场景**: 用户在聊天流式输出过程中切换到其他助手

```
1. 用户在助手A的话题1中开始聊天
   ↓
2. 创建 CHAT 流式输出 (streamId: "abc123", relatedId: "topic1")
   ↓
3. 流式输出开始，AI回复逐步显示
   ↓
4. 用户切换到助手B
   ↓
5. selectAssistant() 更新UI状态
   ↓
6. 流式输出 "abc123" 继续在 Stream Store 中运行
   ↓
7. 用户可以在助手B中开始新的聊天，创建新的流式输出 "def456"
   ↓
8. 两个流式输出并行运行，互不干扰
   ↓
9. 用户可以随时切换回助手A，看到流式输出 "abc123" 的进度
```

#### 5.4 技术优势

**优势1**: 状态隔离
- 流式输出状态不依赖任何特定组件
- 组件切换不会影响流式输出的生命周期
- 支持任意数量的并发流式输出

**优势2**: 灵活绑定
- 通过 relatedId 可以绑定到任何业务实体
- 通过 metadata 可以传递任意绑定信息
- 支持动态的UI更新策略

**优势3**: 用户体验
- 用户可以在流式输出过程中自由导航
- 不会因为页面切换而丢失进度
- 支持多任务并行处理

**优势4**: 开发效率
- 组件职责清晰，只负责展示
- 流式输出逻辑集中在 Store 中
- 易于维护和扩展

## 后端实现详解

### 1. 数据模型定义

#### 1.1 Stream 结构体

```go
// Stream 流式输出模型
type Stream struct {
    ID         string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
    Type       string    `gorm:"type:varchar(50);not null;index" json:"type"`
    RelatedID  string    `gorm:"type:varchar(64);not null;index" json:"relatedId"`
    Status     string    `gorm:"type:varchar(20);not null" json:"status"`
    Content    string    `gorm:"type:text" json:"content"`
    Reasoning  string    `gorm:"type:text" json:"reasoning"`
    Metadata   string    `gorm:"type:text" json:"metadata"`
    StartTime  *time.Time `json:"startTime"`
    EndTime    *time.Time `json:"endTime"`
    Error      string    `gorm:"type:text" json:"error"`
    CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
```

### 2. Service 层实现

#### 2.1 流式输出服务

```go
type StreamService struct{}

// StreamChatCompletion 流式聊天完成
func (s *StreamService) StreamChatCompletion(streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
    logger.LogInfo("Starting stream chat completion", map[string]interface{}{
        "streamID":   streamID,
        "streamType":  streamType,
        "relatedID":   relatedID,
        "messageCount": len(messages),
        "modelType":   modelType,
    })

    // 在协程中处理流式输出，支持并发
    go func() {
        defer func() {
            if r := recover(); r != nil {
                logger.LogError("Stream chat completion panic", fmt.Errorf("%v", r), map[string]interface{}{
                    "streamID": streamID,
                })
                s.emitStreamError(streamID, fmt.Sprintf("Internal error: %v", r))
            }
        }()

        // 调用大模型API
        err := s.callModelAPI(streamID, messages, modelType)
        if err != nil {
            logger.LogError("Stream chat completion failed", err, map[string]interface{}{
                "streamID": streamID,
            })
            s.emitStreamError(streamID, err.Error())
            return
        }

        // 发送流式输出完成事件
        s.emitStreamEnd(streamID)
    }()

    return nil
}
```

#### 2.2 并发处理机制

```go
// 使用协程池管理并发流式输出
func (s *StreamService) callModelAPI(streamID string, messages []map[string]interface{}, modelType string) error {
    // 根据模型类型选择合适的API
    var apiURL string
    switch modelType {
    case "thinking":
        apiURL = s.config.ThinkingModelURL
    case "fast":
        apiURL = s.config.FastModelURL
    default:
        apiURL = s.config.InstructModelURL
    }

    // 创建HTTP请求
    req, err := http.NewRequest("POST", apiURL, bytes.NewReader(requestBody))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    // 发送请求并处理流式响应
    resp, err := httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    // 处理流式响应
    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                break
            }

            // 解析并发送流式数据
            var streamResponse map[string]interface{}
            if err := json.Unmarshal([]byte(data), &streamResponse); err == nil {
                s.emitStreamData(streamID, streamResponse)
            }
        }
    }

    return nil
}
```

### 3. API 层实现

#### 3.1 流式输出API

```go
type StreamAPI struct {
    streamService *service.StreamService
}

// StreamChatCompletion 流式聊天完成API
func (api *StreamAPI) StreamChatCompletion(streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
    return api.streamService.StreamChatCompletion(streamID, streamType, relatedID, messages, modelType)
}

// StopStreamChatCompletion 停止流式聊天完成API
func (api *StreamAPI) StopStreamChatCompletion(streamID string) error {
    return api.streamService.StopStreamChatCompletion(streamID)
}
```

### 4. Wails 集成

#### 4.1 App Core 层

```go
// StreamChatCompletion 流式聊天完成 - Wails API方法
func StreamChatCompletion(streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
    return a.SmartQueryAPI.StreamAPI.StreamChatCompletion(streamID, streamType, relatedID, messages, modelType)
}

// StopStreamChatCompletion 停止流式聊天完成 - Wails API方法
func StopStreamChatCompletion(streamID string) error {
    return a.SmartQueryAPI.StreamAPI.StopStreamChatCompletion(streamID)
}
```

## 数据流向详解

### 1. 聊天对话流程

```
用户输入消息
    ↓
Chat.vue.handleSendMessage()
    ↓
创建新对话和消息
    ↓
streamStore.createStream(StreamType.CHAT, conversationId, metadata)
    ↓
streamStore.startStream(streamId)
    ↓
StreamChatCompletion(streamId, StreamType.CHAT, conversationId, messages, modelType)
    ↓
后端协程处理
    ↓
调用大模型API
    ↓
流式响应处理
    ↓
WebSocket事件发送
    ↓
前端WebSocket监听器接收
    ↓
streamStore.updateStreamContent()
    ↓
Vue响应式更新UI
    ↓
用户看到流式输出
```

### 2. 提示词优化流程

```
用户点击"优化提示词"
    ↓
Chat.vue.handleOptimizePrompt()
    ↓
streamStore.createStream(StreamType.PROMPT_OPTIMIZATION, assistantId, metadata)
    ↓
streamStore.startStream(streamId)
    ↓
StreamChatCompletion(streamId, StreamType.PROMPT_OPTIMIZATION, assistantId, messages, "instruct")
    ↓
后端协程处理
    ↓
调用大模型API
    ↓
流式响应处理
    ↓
WebSocket事件发送
    ↓
前端WebSocket监听器接收
    ↓
streamStore.updateStreamContent()
    ↓
Vue响应式更新inputText
    ↓
用户看到优化后的提示词
```

### 3. 并发处理机制

#### 3.1 后端并发

```go
// 每个流式输出都在独立的协程中处理
go func() {
    // 处理流式输出逻辑
    err := s.callModelAPI(streamID, messages, modelType)
    // 处理结果
}()
```

#### 3.2 前端并发

```javascript
// 使用 Map 存储多个活跃流式输出
const activeStreams = ref(new Map());

// 支持同时存在多个不同类型的流式输出
const chatStream = activeStreams.value.get(chatStreamId);
const promptStream = activeStreams.value.get(promptStreamId);
const titleStream = activeStreams.value.get(titleStreamId);
```

## 关键特性

### 1. 无冲突设计

#### 1.1 StreamID 隔离
- 每个流式输出都有唯一的 StreamID
- 前后端通过 StreamID 精确匹配
- 避免不同类型流式输出的内容混淆

#### 1.2 类型隔离
- 不同类型的流式输出有不同的处理逻辑
- 通过 StreamType 枚举明确区分
- 组件根据类型选择合适的UI更新策略

#### 1.3 状态隔离
- 每个流式输出独立管理状态
- 互不影响的生命周期
- 独立的错误处理机制

### 2. 实时状态同步

#### 2.1 WebSocket 事件
- 实时推送流式数据
- 低延迟的状态更新
- 支持双向通信

#### 2.2 响应式更新
- 基于 Vue 3 的响应式系统
- 自动UI更新
- 计算属性缓存优化

### 3. 错误处理和降级

#### 3.1 前端降级
```javascript
const generateStreamId = async () => {
  try {
    return await GenerateStreamID();
  } catch (error) {
    // 降级到前端生成
    return Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
  }
};
```

#### 3.2 后端错误处理
```go
defer func() {
    if r := recover(); r != nil {
        logger.LogError("Stream panic", fmt.Errorf("%v", r), map[string]interface{}{
            "streamID": streamID,
        })
        s.emitStreamError(streamID, fmt.Sprintf("Internal error: %v", r))
    }
}()
```

### 4. 性能优化

#### 4.1 内存管理
- 流式输出完成后自动清理
- 延迟删除机制避免UI闪烁
- Map 数据结构保证 O(1) 查找性能

#### 4.2 网络优化
- 流式传输减少延迟
- 连接池复用
- 错误重试机制

## 使用示例

### 1. 创建聊天流式输出

```javascript
// 创建流式输出
const streamId = await streamStore.createStream(StreamType.CHAT, conversationId, {
  aiMessageId: aiMessage.id,
  conversationId: conversationId,
});

// 开始流式输出
streamStore.startStream(streamId);

// 调用后端API
await StreamChatCompletion(streamId, StreamType.CHAT, conversationId, messages, modelType);
```

### 2. 创建提示词优化流式输出

```javascript
// 创建流式输出
const streamId = await streamStore.createStream(StreamType.PROMPT_OPTIMIZATION, assistantId, {
  originalPrompt: originalPrompt,
});

// 开始流式输出
streamStore.startStream(streamId);

// 调用后端API
await StreamChatCompletion(streamId, StreamType.PROMPT_OPTIMIZATION, assistantId, messages, "instruct");
```

### 3. 监听流式输出更新

```javascript
watch(
  () => streamStore.activeStreams,
  (newStreams) => {
    for (const [streamId, stream] of newStreams) {
      if (stream.type === StreamType.CHAT) {
        // 处理聊天流式输出
        updateChatMessage(stream);
      } else if (stream.type === StreamType.PROMPT_OPTIMIZATION) {
        // 处理提示词优化流式输出
        updateInputText(stream);
      }
    }
  },
  { deep: true }
);
```

## 总结

本项目的并行流式输出系统通过以下关键技术实现了高效的多流式输出并发处理：

1. **唯一标识**: 使用 StreamID 精确标识每个流式输出
2. **类型隔离**: 通过 StreamType 枚举区分不同类型的流式输出
3. **状态管理**: 基于 Pinia 的全局状态管理，支持响应式更新
4. **并发处理**: 后端协程池 + 前端异步处理
5. **实时通信**: WebSocket 事件驱动的实时数据同步
6. **错误处理**: 完善的错误处理和降级机制
7. **性能优化**: 内存管理、连接复用、延迟清理等优化策略

该系统已经成功应用于聊天对话、提示词优化、标题生成等多个场景，为用户提供了流畅的多任务并行处理体验。
