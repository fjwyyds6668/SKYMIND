import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { 
  StreamChatCompletion, 
  GenerateStreamID,
  UpdateMessage,
  UpdateTopicTitle,
  UpdateConversationTitle,
  GenerateConversationTitle,
  GenerateTopicTitle,
  GetConversations
} from '../../../wailsjs/go/main/App';

// 流式输出类型枚举
export const StreamType = {
  CHAT: 'chat',           // 聊天对话
  CONVERSATION_TITLE_GENERATION: 'conversation_title_generation',  // 对话标题生成
  TOPIC_TITLE_GENERATION: 'topic_title_generation',  // 话题标题生成
  PROMPT_OPTIMIZATION: 'prompt_optimization',  // 提示词优化
  SYSTEM_PROMPT: 'system_prompt',  // 系统提示词生成
};

// 流式输出状态枚举
export const StreamStatus = {
  IDLE: 'idle',           // 空闲
  CONNECTING: 'connecting', // 连接中
  STREAMING: 'streaming',  // 流式输出中
  COMPLETED: 'completed',  // 已完成
  ERROR: 'error',         // 错误
  STOPPED: 'stopped',     // 已停止
};

export const useStreamStore = defineStore('stream', () => {
  // 活跃的流式输出列表
  const activeStreams = ref(new Map());

  // 从后端生成短雪花ID
  const generateStreamId = async () => {
    try {
      return await GenerateStreamID();
    } catch (error) {
      console.error('Failed to generate stream ID:', error);
      // 降级到前端生成
      return Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
    }
  };

  // 初始化WebSocket事件监听
  const initWebSocketListeners = () => {
    if (window.runtime) {
      try {
        window.runtime.EventsOn("websocket-stream-data", handleWebSocketStreamData);
        window.runtime.EventsOn("websocket-stream-end", handleWebSocketStreamEnd);
        window.runtime.EventsOn("websocket-stream-error", handleWebSocketStreamError);
      } catch (error) {
        console.error("Failed to add WebSocket event listeners:", error);
      }
    }
  };

  // 清理WebSocket事件监听
  const cleanupWebSocketListeners = () => {
    if (window.runtime) {
      try {
        window.runtime.EventsOff("websocket-stream-data");
        window.runtime.EventsOff("websocket-stream-end");
        window.runtime.EventsOff("websocket-stream-error");
      } catch (error) {
        console.error("Failed to remove WebSocket event listeners:", error);
      }
    }
  };

  // 处理WebSocket流式数据事件
  const handleWebSocketStreamData = (event) => {
    const { streamID, data } = event;
    
    if (!data || !data.choices || data.choices.length === 0) return;

    const choice = data.choices[0];
    const delta = choice.delta;
    const stream = activeStreams.value.get(streamID);
    
    if (!stream) return;

    // 更新流式输出内容
    const content = delta.content || "";
    const reasoningContent = delta.reasoning_content || "";
    
    // 只有当有内容或思考内容时才更新
    if (content !== "" || reasoningContent !== "") {
      updateStreamContent(streamID, content, reasoningContent);
    }
  };

  // 处理WebSocket流式结束事件
  const handleWebSocketStreamEnd = (event) => {
    const { streamID } = event;
    completeStream(streamID);
    
    // 处理流式输出完成后的数据库保存
    handleStreamCompletion(streamID);
  };

  // 处理WebSocket流式错误事件
  const handleWebSocketStreamError = (event) => {
    const { streamID, error } = event;
    setStreamError(streamID, error);
  };

  // 初始化事件监听
  initWebSocketListeners();

  // 创建新的流式输出
  const createStream = async (streamType, metadata = {}) => {
    const streamId = await generateStreamId();
    const stream = {
      id: streamId,
      type: streamType,
      status: StreamStatus.IDLE,
      content: '',
      reasoning: '',
      metadata: metadata,
      startTime: null,
      endTime: null,
      error: null,
      // 添加思考状态跟踪
      thinkingPhase: true, // 是否处于思考阶段
      hasStartedContent: false, // 是否已经开始生成内容
    };
    
    activeStreams.value.set(streamId, stream);
    return streamId;
  };

  // 开始流式输出
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

  // 更新流式输出内容
  const updateStreamContent = (streamId, content, reasoning = '') => {
    const stream = activeStreams.value.get(streamId);
    if (stream) {
      if (content) {
        stream.content += content;
        // 一旦开始接收内容，标记思考阶段结束
        if (!stream.hasStartedContent) {
          stream.hasStartedContent = true;
          stream.thinkingPhase = false;
        }
      }
      if (reasoning) {
        stream.reasoning += reasoning;
      }
    }
  };

  // 完成流式输出
  const completeStream = (streamId) => {
    const stream = activeStreams.value.get(streamId);
    if (stream) {
      stream.status = StreamStatus.COMPLETED;
      stream.endTime = new Date();
      
      // 从活跃列表中移除（延迟移除，给UI一些时间显示完成状态）
      setTimeout(() => {
        activeStreams.value.delete(streamId);
      }, 500);
    }
  };

  // 停止流式输出
  const stopStream = (streamId) => {
    const stream = activeStreams.value.get(streamId);
    if (stream) {
      stream.status = StreamStatus.STOPPED;
      stream.endTime = new Date();
      
      // 从活跃列表中移除
      activeStreams.value.delete(streamId);
    }
  };

  // 设置流式输出错误
  const setStreamError = (streamId, error) => {
    const stream = activeStreams.value.get(streamId);
    if (stream) {
      stream.status = StreamStatus.ERROR;
      stream.error = error;
      stream.endTime = new Date();
      
      // 从活跃列表中移除
      activeStreams.value.delete(streamId);
    }
  };

  // 获取指定类型的活跃流式输出
  const getActiveStreamsByType = (streamType) => {
    const result = [];
    for (const [streamId, stream] of activeStreams.value) {
      if (stream.type === streamType) {
        result.push(stream);
      }
    }
    return result;
  };

  // 获取指定话题ID的活跃流式输出
  const getActiveStreamsByTopicId = (topicId) => {
    const result = [];
    for (const [streamId, stream] of activeStreams.value) {
      if (stream.metadata.topicId === topicId) {
        result.push(stream);
      }
    }
    return result;
  };

  // 获取指定对话ID的活跃流式输出
  const getActiveStreamsByConversationId = (conversationId) => {
    const result = [];
    for (const [streamId, stream] of activeStreams.value) {
      if (stream.metadata.conversationId === conversationId) {
        result.push(stream);
      }
    }
    return result;
  };

  // 获取指定助手ID的活跃流式输出
  const getActiveStreamsByAssistantId = (assistantId) => {
    const result = [];
    for (const [streamId, stream] of activeStreams.value) {
      if (stream.metadata.assistantId === assistantId) {
        result.push(stream);
      }
    }
    return result;
  };

  // 检查指定类型是否有活跃流式输出
  const hasActiveStreamByType = (streamType) => {
    for (const stream of activeStreams.value.values()) {
      if (stream.type === streamType && stream.status === StreamStatus.STREAMING) {
        return true;
      }
    }
    return false;
  };

  // 检查指定话题ID是否有活跃流式输出
  const hasActiveStreamByTopicId = (topicId) => {
    for (const stream of activeStreams.value.values()) {
      if (stream.metadata.topicId === topicId && stream.status === StreamStatus.STREAMING) {
        return true;
      }
    }
    return false;
  };

  // 检查指定对话ID是否有活跃流式输出
  const hasActiveStreamByConversationId = (conversationId) => {
    for (const stream of activeStreams.value.values()) {
      if (stream.metadata.conversationId === conversationId && stream.status === StreamStatus.STREAMING) {
        return true;
      }
    }
    return false;
  };

  // 检查指定助手ID是否有活跃流式输出
  const hasActiveStreamByAssistantId = (assistantId) => {
    for (const stream of activeStreams.value.values()) {
      if (stream.metadata.assistantId === assistantId && stream.status === StreamStatus.STREAMING) {
        return true;
      }
    }
    return false;
  };

  // 获取所有活跃CHAT流式输出的助手ID列表
  const getActiveChatAssistantIds = () => {
    const assistantIds = new Set();
    for (const stream of activeStreams.value.values()) {
      if (stream.type === StreamType.CHAT && stream.status === StreamStatus.STREAMING && stream.metadata.assistantId) {
        assistantIds.add(stream.metadata.assistantId);
      }
    }
    return Array.from(assistantIds);
  };

  // 获取所有活跃CHAT流式输出的话题ID列表
  const getActiveChatTopicIds = () => {
    const topicIds = new Set();
    for (const stream of activeStreams.value.values()) {
      if (stream.type === StreamType.CHAT && stream.status === StreamStatus.STREAMING && stream.metadata.topicId) {
        topicIds.add(stream.metadata.topicId);
      }
    }
    return Array.from(topicIds);
  };

  // 停止所有流式输出
  const stopAllStreams = () => {
    for (const [streamId, stream] of activeStreams.value) {
      if (stream.status === StreamStatus.STREAMING) {
        stopStream(streamId);
      }
    }
  };

  // 计算属性
  const activeStreamsCount = computed(() => activeStreams.value.size);
  
  const streamingStreamsCount = computed(() => {
    let count = 0;
    for (const stream of activeStreams.value.values()) {
      if (stream.status === StreamStatus.STREAMING) {
        count++;
      }
    }
    return count;
  });

  const activeStreamsList = computed(() => {
    return Array.from(activeStreams.value.values());
  });

  // 获取指定流式输出的信息
  const getStreamInfo = (streamId) => {
    return activeStreams.value.get(streamId) || null;
  };

  // 处理流式输出完成后的数据库保存
  const handleStreamCompletion = async (streamId) => {
    const stream = activeStreams.value.get(streamId);
    if (!stream) return;

    try {
      switch (stream.type) {
        case StreamType.CHAT:
          // 聊天完成：更新AI回复内容到数据库
          if (stream.metadata.aiMessageId && stream.content && stream.content.trim()) {
            await UpdateMessage(stream.metadata.aiMessageId, stream.content, stream.reasoning || "");
            console.log("AI消息内容已保存到数据库:", stream.metadata.aiMessageId);
            
            // 触发对话标题生成
            await handleConversationTitleGeneration(stream);
          }
          break;

        case StreamType.CONVERSATION_TITLE_GENERATION:
          // 对话标题生成完成：保存标题到数据库
          if (stream.metadata.conversationId && stream.content && stream.content.trim() && stream.content !== "新对话") {
            try {
              await UpdateConversationTitle(stream.metadata.conversationId, stream.content.trim());
              console.log("对话标题已保存到数据库:", stream.content.trim());
              
              // 触发话题标题生成
              await handleTopicTitleGeneration(stream);
            } catch (titleError) {
              console.error("保存对话标题失败:", titleError);
            }
          } else {
            console.log("对话标题生成完成，但内容为空或为默认标题，跳过保存:", {
              conversationId: stream.metadata.conversationId,
              content: stream.content,
              contentTrim: stream.content?.trim(),
              isNewConversation: stream.content === "新对话"
            });
          }
          break;

        case StreamType.TOPIC_TITLE_GENERATION:
          // 话题标题生成完成：保存标题到数据库
          if (stream.metadata.topicId && stream.content && stream.content.trim()) {
            await UpdateTopicTitle(stream.metadata.topicId, stream.content.trim());
            console.log("话题标题已保存到数据库:", stream.content.trim());
          }
          break;

        case StreamType.PROMPT_OPTIMIZATION:
        case StreamType.SYSTEM_PROMPT:
          // 提示词优化和系统提示词生成完成
          // 这些类型的流式输出通常不需要额外的数据库保存
          // 因为内容已经通过其他方式处理
          console.log(`${stream.type} 流式输出完成`);
          break;

        default:
          console.warn("未知的流式输出类型:", stream.type);
      }
    } catch (error) {
      console.error(`处理流式输出完成时发生错误 (${stream.type}):`, error);
    }
  };

  // 处理对话标题生成
  const handleConversationTitleGeneration = async (stream) => {
    try {
      // 添加延迟避免API频率限制
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // 获取对话信息来生成标题
      const conversationsData = await GetConversations(stream.metadata.topicId);
      const conversations = conversationsData || [];
      
      const targetConversation = conversations.find(conv => conv.id === stream.metadata.conversationId);
      if (!targetConversation) {
        console.warn("找不到目标对话:", stream.metadata.conversationId);
        return;
      }

      // 从对话设置中获取用户消息
      let userMessage = "";
      try {
        const settings = targetConversation.settings ? JSON.parse(targetConversation.settings) : {};
        const currentSendId = settings.currentSendId;
        
        if (currentSendId && targetConversation.messages) {
          const currentUserMessage = targetConversation.messages.find(msg => msg.role === "user" && msg.id === currentSendId);
          if (currentUserMessage) {
            userMessage = currentUserMessage.content;
          }
        }
      } catch (error) {
        console.error("解析对话设置失败:", error);
      }
      
      const generatedPrompt = await GenerateConversationTitle(userMessage, stream.content);
      
      // 再次延迟确保不会与聊天请求冲突
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // 创建对话标题生成的流式输出
      const titleStreamId = await generateStreamId();
      const titleStream = {
        id: titleStreamId,
        type: StreamType.CONVERSATION_TITLE_GENERATION,
        status: StreamStatus.IDLE,
        content: '',
        reasoning: '',
        metadata: {
          conversationId: stream.metadata.conversationId,
          topicId: stream.metadata.topicId, // 添加topicId以确保独立性
        },
        startTime: null,
        endTime: null,
        error: null,
      };
      
      activeStreams.value.set(titleStreamId, titleStream);
      startStream(titleStreamId);
      const messages = [
        {
          role: "user",
          content: generatedPrompt,
        },
      ];

      await StreamChatCompletion(titleStreamId, StreamType.CONVERSATION_TITLE_GENERATION, stream.metadata.conversationId, messages, "fast");
    } catch (error) {
      console.error("生成对话标题失败:", error);
      // 如果是频率限制错误，不显示给用户，静默处理
      if (error.toString().includes("429") || error.toString().includes("rate limit")) {
        console.log("标题生成遇到频率限制，跳过本次生成");
        return;
      }
    }
  };

  // 处理话题标题生成
  const handleTopicTitleGeneration = async (stream) => {
    try {
      // 添加延迟避免API频率限制
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // 获取该话题下的所有对话标题
      const conversationsData = await GetConversations(stream.metadata.topicId);
      const conversations = conversationsData || [];
      
      if (conversations.length === 0) {
        console.log("话题下没有对话，跳过话题标题生成");
        return;
      }
      
      // 提取所有非默认标题的对话标题
      const conversationTitles = conversations
        .filter(conv => conv.title && conv.title !== "新对话" && conv.title.trim())
        .map(conv => conv.title);
      
      if (conversationTitles.length === 0) {
        console.log("话题下没有有效的对话标题，跳过话题标题生成");
        return;
      }
      
      const generatedPrompt = await GenerateTopicTitle(conversationTitles);
      
      // 再次延迟确保不会与其他请求冲突
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // 创建话题标题生成的流式输出
      const topicStreamId = await generateStreamId();
      const topicStream = {
        id: topicStreamId,
        type: StreamType.TOPIC_TITLE_GENERATION,
        status: StreamStatus.IDLE,
        content: '',
        reasoning: '',
        metadata: {
          topicId: stream.metadata.topicId,
        },
        startTime: null,
        endTime: null,
        error: null,
      };
      
      activeStreams.value.set(topicStreamId, topicStream);
      startStream(topicStreamId);

      // 调用后端API生成话题标题
      const messages = [
        {
          role: "user",
          content: generatedPrompt,
        },
      ];

      await StreamChatCompletion(topicStreamId, StreamType.TOPIC_TITLE_GENERATION, stream.metadata.topicId, messages, "fast");
    } catch (error) {
      console.error("生成话题标题失败:", error);
      // 如果是频率限制错误，不显示给用户，静默处理
      if (error.toString().includes("429") || error.toString().includes("rate limit")) {
        console.log("话题标题生成遇到频率限制，跳过本次生成");
        return;
      }
    }
  };

  return {
    // 状态
    activeStreams,
    activeStreamsCount,
    streamingStreamsCount,
    activeStreamsList,
    
    // 枚举
    StreamType,
    StreamStatus,
    
    // 方法
    generateStreamId,
    createStream,
    startStream,
    updateStreamContent,
    completeStream,
    stopStream,
    setStreamError,
    getActiveStreamsByType,
    getActiveStreamsByTopicId,
    getActiveStreamsByConversationId,
    getActiveStreamsByAssistantId,
    hasActiveStreamByType,
    hasActiveStreamByTopicId,
    hasActiveStreamByConversationId,
    hasActiveStreamByAssistantId,
    getActiveChatAssistantIds,
    getActiveChatTopicIds,
    stopAllStreams,
    getStreamInfo,
    handleStreamCompletion,
  };
});
