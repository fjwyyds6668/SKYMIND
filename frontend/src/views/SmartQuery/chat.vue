<template>
  <div class="chat-area">
    <!-- 对话列表 -->
    <div v-if="conversationList.length > 0" class="conversation-list">
      <!-- 助手提示词显示 -->
      <div v-if="selectedAssistantData && selectedAssistantData.prompt" class="assistant-prompt" @click="openAssistantSettings">
        <div class="prompt-content">{{ selectedAssistantData.prompt }}</div>
      </div>
      <div v-for="(conversation, index) in conversationList" :key="index" class="conversation-wrapper">
        <!-- 对话标题 -->
        <div class="conversation-header">
          <span class="conversation-title">{{ conversation.title }}</span>
          <span class="conversation-time">{{ formatTime(conversation.updated_at) }}</span>
        </div>

        <!-- 对话内容 -->
        <div class="conversation-content">
          <Conversation
            ref="conversationItemRef"
            :conversation="conversation"
            :selectedAssistant="selectedAssistantData"
            :is-streaming-chat="isStreamingChat"
            @regenerate-message="handleRegenerateMessage"
            @conversation-deleted="handleConversationDeleted"
            @config-change="handleConfigChange"
            @messages-change="handleMessagesChange"
          />
        </div>
      </div>

      <!-- 回到顶部按钮 -->
      <div v-if="isShowToTop" class="back-to-top" @click="scrollToConversationTop">
        <IconArrowUp />
      </div>

      <!-- 回到底部按钮 -->
      <div v-if="isShowToBottom" class="back-to-bottom" @click="scrollToConversationBottom">
        <IconArrowDown />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-conversation">
      <!-- 助手提示词显示 -->
      <div v-if="selectedAssistantData && selectedAssistantData.prompt" class="assistant-prompt" @click="openAssistantSettings">
        <div class="prompt-content">{{ selectedAssistantData.prompt }}</div>
      </div>
      <t-empty description="请发送消息开始对话" />
    </div>

    <!-- 输入区域 -->
    <div class="input-section">
      <t-chat-sender
        v-model="inputText"
        :loading="isStreamingChat || isOptimizingPrompt"
        :disabled="isStreamingChat || !selectedAssistantData || !selectedTopicData"
        :placeholder="selectedAssistantData && selectedTopicData ? '请输入您的问题...' : '请先选择一个助手和话题'"
        :textarea-props="{
          placeholder: selectedAssistantData && selectedTopicData ? '请输入您的问题...' : '请先选择一个助手和话题',
        }"
        :attachments-props="{
          items: filesList,
          overflow: 'scrollX',
        }"
        @send="handleSendMessage"
        @stop="handleStopChat"
        @file-select="handleUploadFile"
        @file-click="handleFileClick"
        @remove="handleRemoveFile"
      >
        <template #footer-prefix>
          <div class="model-select">
            <t-button class="thinking-btn" :class="{ 'is-active': deepThinkingEnabled }" variant="text" @click="toggleDeepThinking">
              <IconBrain />
              <span>深度思考</span>
            </t-button>
            <t-button
              theme="primary"
              shape="round"
              @click="handleOptimizePrompt"
              :loading="isOptimizingPrompt"
              :disabled="!selectedAssistantData || isOptimizingPrompt || !inputText.trim()"
            >
              <div class="button-content">
                <IconSparkles v-if="!isOptimizingPrompt" :size="16" />
                <span>优化提示词</span>
              </div>
            </t-button>
          </div>
        </template>
        <template #suffix="{ renderPresets }">
          <component :is="renderPresets([{ name: 'uploadAttachment' }])" />
        </template>
      </t-chat-sender>
    </div>

    <!-- 助手设置对话框 -->
    <t-dialog
      v-model:visible="showAssistantSettings"
      :header="dialogTitle"
      :width="'66%'"
      :close-on-overlay-click="false"
      class="assistant-settings-dialog"
      confirm-btn="保存设置"
      cancel-btn="取消"
      :on-confirm="handleAssistantSettingsSave"
      :on-close="handleAssistantSettingsCancel"
      :placement="'top'"
      :top="'10vh'"
    >
      <AssistantSettings
        v-if="showAssistantSettings"
        ref="assistantSettingsRef"
        :assistant="selectedAssistantData"
        @save="handleAssistantSettingsUpdate"
      />
    </t-dialog>
  </div>
</template>

<script setup>
import { ref, nextTick, reactive, watch, onMounted, onUnmounted, computed } from "vue";
import { MessagePlugin } from "tdesign-vue-next";
import {
  GetConversations,
  CreateConversation,
  CreateMessage,
  StreamChatCompletion,
  StopStreamChatCompletion,
  UpdateConversationSettings,
  DeleteConversationsAfter,
  OptimizeUserPrompt,
  SaveFile,
  ProcessFileContent,
} from "../../../wailsjs/go/main/App";
import Conversation from "./conversation.vue";
import AssistantSettings from "./settings.vue";
import { useStreamStore, StreamType, StreamStatus } from "../../store/modules/stream.js";
import { cloneDeep } from "lodash";
import { processFile, formatFileSize } from "./utils.js";

// Props
const props = defineProps({
  topicId: {
    type: [String, Number],
    default: null,
  },
  selectedAssistant: {
    type: Object,
    default: null,
  },
  selectedTopic: {
    type: Object,
    default: null,
  },
  assistantSettings: {
    type: Object,
    default: () => ({
      temperature: 1.0,
      contextCount: 5,
    }),
  },
});

// Emits
const emit = defineEmits(["conversation-created", "assistant-updated"]);

// Store
const streamStore = useStreamStore();

// 响应式数据
const conversationItemRef = ref([]);
const isShowToBottom = ref(false);
const isShowToTop = ref(false);
const inputText = ref("");
const conversationList = ref([]);
const newConversationId = ref(0);
const deepThinkingEnabled = ref(false);
const showAssistantSettings = ref(false);
const assistantSettingsRef = ref(null);
const isOptimizingPrompt = ref(false);
const filesList = ref([]);

// 定义中间值来接收props
const selectedAssistantData = ref({});
const selectedTopicData = ref({});
const assistantSettingsData = ref({});

// 计算属性
const dialogTitle = computed(() => {
  return selectedAssistantData.value?.name ? `【${selectedAssistantData.value.name}】设置` : "助手设置";
});

// 检查是否有聊天流式输出在进行
const isStreamingChat = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.CHAT) && streamStore.hasActiveStreamByTopicId(selectedTopicData.value.id) ;
});

// 初始化方法
const init = () => {
  selectedAssistantData.value = cloneDeep(props.selectedAssistant);
  selectedTopicData.value = cloneDeep(props.selectedTopic);
  assistantSettingsData.value = cloneDeep(props.assistantSettings);
};

// 组件初始化时调用init
init();

// 监听props变化
watch(
  () => props.selectedAssistant,
  (newAssistant) => {
    selectedAssistantData.value = cloneDeep(newAssistant);
  },
  { deep: true }
);

watch(
  () => props.selectedTopic,
  (newTopic) => {
    selectedTopicData.value = cloneDeep(newTopic);
  },
  { deep: true }
);

watch(
  () => props.assistantSettings,
  (newSettings) => {
    assistantSettingsData.value = cloneDeep(newSettings);
  },
  { deep: true }
);

// 加载对话列表
const loadConversations = async (topic) => {
  try {
    const conversationsData = await GetConversations(topic.id);
    conversationList.value = conversationsData || [];
  } catch (error) {
    MessagePlugin.error("加载对话列表失败: " + error);
    conversationList.value = [];
  }
};

// 监听 TopicId 变化，加载对应的对话数据
watch(
  () => props.topicId,
  async (newTopicId) => {
    if (newTopicId) {
      await loadConversations({ id: newTopicId });
    } else {
      conversationList.value = [];
    }

    // 重新设置滚动监听器
    await nextTick();
    setupScrollListener();
  },
  { immediate: true }
);

// 格式化时间
const formatTime = (timeString) => {
  if (!timeString) return "";
  const date = new Date(timeString);
  return date.toLocaleString();
};

// 滚动到对话列表底部
const scrollToConversationBottom = () => {
  const conversationListElement = document.querySelector(".conversation-list");
  if (conversationListElement) {
    conversationListElement.scrollTop = conversationListElement.scrollHeight;
  }

  if (conversationItemRef.value && conversationItemRef.value.length > 0) {
    const lastConversationRef = conversationItemRef.value[conversationItemRef.value.length - 1];
    if (lastConversationRef && lastConversationRef.messagesContainer) {
      lastConversationRef.messagesContainer.scrollTop = lastConversationRef.messagesContainer.scrollHeight;
    }
  }
  isShowToBottom.value = false;
};

// 滚动到对话列表顶部
const scrollToConversationTop = () => {
  const conversationListElement = document.querySelector(".conversation-list");
  if (conversationListElement) {
    conversationListElement.scrollTop = 0;
  }

  if (conversationItemRef.value && conversationItemRef.value.length > 0) {
    const firstConversationRef = conversationItemRef.value[0];
    if (firstConversationRef && firstConversationRef.messagesContainer) {
      firstConversationRef.messagesContainer.scrollTop = 0;
    }
  }
  isShowToTop.value = false;
};

// 处理重新生成请求
const handleRegenerateMessage = async (data) => {
  const { userMessage, conversationId } = data;
  await nextTick();
  scrollToConversationBottom();

  try {
    // 1. 删除指定对话之后的所有对话及其消息
    await DeleteConversationsAfter(conversationId);

    // 2. 重新加载对话列表
    await loadConversations({ id: props.topicId });

    // 3. 确保使用正确的对话ID
    newConversationId.value = conversationId;

    // 4. 找到对应的对话和用户消息
    const targetConversation = conversationList.value.find((conv) => conv.id === conversationId);
    if (!targetConversation) {
      MessagePlugin.error("找不到对应的对话");
      return;
    }

    const targetUserMessage = targetConversation.messages.find((msg) => msg.role === "user" && msg.content === userMessage.content);
    if (!targetUserMessage) {
      MessagePlugin.error("找不到对应的用户消息");
      return;
    }

    // 5. 立即创建空的AI回复Message并保存到数据库
    let newAiMessage = null;
    try {
      const savedAiMessage = await CreateMessage({
        topic_id: selectedTopicData.value.id,
        conversation_id: conversationId,
        role: "assistant",
        content: "",
        reasoning: "",
        token_count: 0,
        metadata: "{}",
      });

      if (savedAiMessage && savedAiMessage.id) {
        const aiCurrentTime = new Date().toLocaleString();
        newAiMessage = reactive({
          id: savedAiMessage.id,
          avatar: selectedAssistantData.value?.emoji || "🤖",
          name: selectedAssistantData.value?.name || "AI助手",
          datetime: aiCurrentTime,
          content: "",
          reasoning: "",
          role: "assistant",
        });

        targetConversation.messages.push(newAiMessage);

        // 立即更新对话设置中的CurrentReplyID
        if (targetConversation) {
          let settings = {};
          try {
            settings = targetConversation.settings ? JSON.parse(targetConversation.settings) : {};
          } catch (error) {
            settings = {};
          }
          settings.currentReplyId = savedAiMessage.id;

          await UpdateConversationSettings(conversationId, JSON.stringify(settings));
          targetConversation.settings = JSON.stringify(settings);
        }

        // 启动聊天流式输出
        await startChatStream(targetUserMessage.content, newAiMessage, conversationId);
      }
    } catch (error) {
      MessagePlugin.error("创建AI回复消息失败:", error);
      const aiCurrentTime = new Date().toLocaleString();
      newAiMessage = reactive({
        avatar: selectedAssistantData.value?.emoji || "🤖",
        name: selectedAssistantData.value?.name || "AI助手",
        datetime: aiCurrentTime,
        content: "",
        reasoning: "",
        role: "assistant",
      });
      targetConversation.messages.push(newAiMessage);
      await startChatStream(targetUserMessage.content, newAiMessage, conversationId);
    }
  } catch (error) {
    MessagePlugin.error("重新生成失败: " + (error.message || error));
  }
};

// 处理对话删除事件
const handleConversationDeleted = (conversationId) => {
  const index = conversationList.value.findIndex((conv) => conv.id === conversationId);
  if (index !== -1) {
    conversationList.value.splice(index, 1);
  }
};

// 处理配置更改事件
const handleConfigChange = async (data) => {
  try {
    await UpdateConversationSettings(data.conversationId, data.settings);
    await loadConversations({ id: props.topicId });
  } catch (error) {
    MessagePlugin.error("更新对话设置失败: " + (error.message || "未知错误"));
  }
};

// 处理消息变化事件
const handleMessagesChange = async () => {
  try {
    await loadConversations({ id: props.topicId });
  } catch (error) {
    MessagePlugin.error("刷新对话数据失败: " + (error.message || "未知错误"));
  }
};

// 停止聊天流式输出
const handleStopChat = async () => {
  const chatStreams = streamStore.getActiveStreamsByType(StreamType.CHAT);
  for (const stream of chatStreams) {
    try {
      await StopStreamChatCompletion(stream.id);
      streamStore.stopStream(stream.id);
    } catch (error) {
      console.error("停止流式输出失败:", error);
    }
  }
};

// 发送消息
const handleSendMessage = async (content) => {
  const inputValue = content.trim();
  if (isStreamingChat.value || !inputValue || !selectedAssistantData.value || !selectedTopicData.value) return;

  await nextTick();
  scrollToConversationBottom();

  // 创建新对话
  try {
    const newConversation = {
      topic_id: selectedTopicData.value.id,
      assistant_id: selectedAssistantData.value.id,
      title: "新对话",
      user_id: "",
      model_id: "",
      settings: JSON.stringify(assistantSettingsData.value),
      is_archived: false,
      messages: [],
    };

    newConversationId.value = await CreateConversation(newConversation);
    newConversation.id = newConversationId.value;
    conversationList.value.push(newConversation);

    emit("conversation-created", newConversation);
  } catch (error) {
    MessagePlugin.error("创建对话失败，无法发送消息");
    return;
  }

  inputText.value = "";

  // 构建消息内容，包含附件信息
  let messageContent = inputValue;
  if (filesList.value.length > 0) {
    const attachmentInfo = filesList.value.map(file => {
      return `[附件: ${file.name}${file.localPath ? ` (路径: ${file.localPath})` : ''}]`;
    }).join('\n');
    messageContent = `${inputValue}\n\n${attachmentInfo}`;
  }

  // 添加用户消息
  const currentTime = new Date().toLocaleString();
  const userMessage = reactive({
    avatar: "/images/avatar.jpg",
    name: "用户",
    datetime: currentTime,
    content: messageContent,
    role: "user",
    attachments: [...filesList.value], // 保存附件信息
  });
  conversationList.value[conversationList.value.length - 1].messages.push(userMessage);

  // 保存用户消息到数据库
  try {
    const messageMetadata = {
      attachments: filesList.value.map(file => ({
        name: file.name,
        size: file.size,
        localPath: file.localPath,
        key: file.key
      }))
    };

    const savedUserMessage = await CreateMessage({
      topic_id: selectedTopicData.value.id,
      conversation_id: newConversationId.value,
      role: "user",
      content: messageContent,
      token_count: 0,
      metadata: JSON.stringify(messageMetadata),
    });

    if (savedUserMessage && savedUserMessage.id) {
      userMessage.id = savedUserMessage.id;

      const currentConversation = conversationList.value[conversationList.value.length - 1];
      if (currentConversation) {
        let settings = {};
        try {
          settings = currentConversation.settings ? JSON.parse(currentConversation.settings) : {};
        } catch (error) {
          settings = {};
        }
        settings.currentSendId = savedUserMessage.id;

        await UpdateConversationSettings(newConversationId.value, JSON.stringify(settings));
        currentConversation.settings = JSON.stringify(settings);
      }
    }
  } catch (error) {
    console.error("保存用户消息失败:", error);
  }

  // 清空附件列表
  filesList.value = [];

  // 创建AI回复消息
  let aiMessage = null;
  try {
    const savedAiMessage = await CreateMessage({
      topic_id: selectedTopicData.value.id,
      conversation_id: newConversationId.value,
      role: "assistant",
      content: "",
      reasoning: "",
      token_count: 0,
      metadata: "{}",
    });

    if (savedAiMessage && savedAiMessage.id) {
      const aiCurrentTime = new Date().toLocaleString();
      aiMessage = reactive({
        id: savedAiMessage.id,
        avatar: selectedAssistantData.value?.emoji || "🤖",
        name: selectedAssistantData.value?.name || "AI助手",
        datetime: aiCurrentTime,
        content: "",
        reasoning: "",
        role: "assistant",
      });

      conversationList.value[conversationList.value.length - 1].messages.push(aiMessage);

      const currentConversation = conversationList.value[conversationList.value.length - 1];
      if (currentConversation) {
        let settings = {};
        try {
          settings = currentConversation.settings ? JSON.parse(currentConversation.settings) : {};
        } catch (error) {
          settings = {};
        }
        settings.currentReplyId = savedAiMessage.id;

        await UpdateConversationSettings(newConversationId.value, JSON.stringify(settings));
        currentConversation.settings = JSON.stringify(settings);
      }

      await startChatStream(inputValue, aiMessage, newConversationId.value);
    }
  } catch (error) {
    MessagePlugin.error("创建AI回复消息失败:", error);
    const aiCurrentTime = new Date().toLocaleString();
    aiMessage = reactive({
      avatar: selectedAssistantData.value?.emoji || "🤖",
      name: selectedAssistantData.value?.name || "AI助手",
      datetime: aiCurrentTime,
      content: "",
      reasoning: "",
      role: "assistant",
    });
    conversationList.value[conversationList.value.length - 1].messages.push(aiMessage);
    await startChatStream(inputValue, aiMessage, newConversationId.value);
  }
};

// 构建聊天消息历史
const buildChatMessages = (inputValue) => {
  const assistantSettings = JSON.parse(selectedAssistantData.value.settings);
  const messages = [];

  // 添加系统提示词
  if (selectedAssistantData.value && selectedAssistantData.value.prompt) {
    messages.push({
      role: "system",
      content: selectedAssistantData.value.prompt,
    });
  }

  // 计算起始索引，确保只遍历最近contextCount个对话
  const startIndex = Math.max(0, conversationList.value.length - assistantSettings.contextCount - 1);

  for (let i = startIndex; i < conversationList.value.length; i++) {
    const conversation = conversationList.value[i];
    const conversationSettings = JSON.parse(conversation.settings);
    if (conversation.messages) {
      for (let j = 0; j < conversation.messages.length; j++) {
        const msg = conversation.messages[j];
        if ((msg.role === "user" || msg.role === "assistant") && msg.content) {
          if (msg.role === "user" && conversationSettings.currentSendId && msg.id !== conversationSettings.currentSendId) {
            continue;
          }
          if (msg.role === "assistant" && conversationSettings.currentReplyId && msg.id !== conversationSettings.currentReplyId) {
            continue;
          }
          messages.push({
            role: msg.role,
            content: msg.content,
          });
        }
      }
    }
  }

  // 添加当前用户输入
  const lastMessage = messages[messages.length - 1];
  if (!lastMessage || lastMessage.role !== "user" || lastMessage.content !== inputValue) {
    messages.push({
      role: "user",
      content: inputValue,
    });
  }

  return messages;
};

// 启动聊天流式输出
const startChatStream = async (inputValue, aiMessage, conversationId) => {
  try {
    // 构建消息历史
    const messages = buildChatMessages(inputValue);

    // 创建流式输出
    const streamId = await streamStore.createStream(StreamType.CHAT, {
      aiMessageId: aiMessage.id,
      conversationId: conversationId,
      topicId: selectedTopicData.value.id,
      assistantId: selectedAssistantData.value.id,
    });

    // 开始流式输出
    streamStore.startStream(streamId);

    // 调用后端流式API
    const modelType = deepThinkingEnabled.value ? "thinking" : "instruct";
    await StreamChatCompletion(streamId, StreamType.CHAT, conversationId, messages, modelType);
  } catch (error) {
    const errorStr = error.toString().toLowerCase();
    if (errorStr.includes("context canceled") || errorStr.includes("canceled")) {
      // 用户主动停止
      return;
    } else {
      // 真正的错误
      aiMessage.role = "error";
      aiMessage.content = `抱歉，连接AI服务时出现错误：${error}。请检查网络连接或稍后重试。`;
    }
  }
};

// 处理优化提示词
const handleOptimizePrompt = async () => {
  if (!selectedAssistantData.value) {
    MessagePlugin.error("请先选择一个助手");
    return;
  }

  if (!inputText.value || !inputText.value.trim()) {
    MessagePlugin.error("请输入要优化的提示词内容");
    return;
  }

  try {
    const originalPrompt = inputText.value.trim();
    const generatedPrompt = await OptimizeUserPrompt(originalPrompt);

    inputText.value = "";

    // 创建流式输出
    const streamId = await streamStore.createStream(StreamType.PROMPT_OPTIMIZATION, {
      originalPrompt: originalPrompt,
      topicId: selectedTopicData.value.id,
    });

    // 开始流式输出
    streamStore.startStream(streamId);

    // 构建消息数组用于流式调用
    const messages = [
      {
        role: "user",
        content: generatedPrompt,
      },
    ];

    // 设置优化提示词模式标志
    isOptimizingPrompt.value = true;

    // 调用后端API流式生成优化后的提示词
    await StreamChatCompletion(streamId, StreamType.PROMPT_OPTIMIZATION, selectedAssistantData.value.id, messages, "instruct");
  } catch (error) {
    const errorStr = error.toString().toLowerCase();
    if (errorStr.includes("context canceled") || errorStr.includes("canceled")) {
      isOptimizingPrompt.value = false;
      return;
    }

    console.error("优化提示词失败：", error);
    isOptimizingPrompt.value = false;
  } finally {
    isOptimizingPrompt.value = false;
  }
};

// 监听流式输出内容变化
watch(
  () => streamStore.activeStreams,
  (newStreams) => {
    // 处理所有活跃的流式输出
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
      } else if (stream.type === StreamType.PROMPT_OPTIMIZATION) {
        // 优化提示词模式：将内容显示在输入框
        if (stream.content !== undefined) {
          inputText.value = stream.content;
        }

        // 如果流式输出完成，保持输入框内容不被清空
        if (stream.status === StreamStatus.COMPLETED) {
          // 确保输入框保持最终内容
          const finalContent = stream.content || inputText.value;
          if (finalContent && finalContent.trim()) {
            // 使用 nextTick 确保在 DOM 更新后设置
            nextTick(() => {
              inputText.value = finalContent;
            });
          }
        }
      } else if (stream.type === StreamType.CONVERSATION_TITLE_GENERATION) {
        // 对话标题生成模式：更新对话标题
        // 根据conversationId找到对应的对话，而不是只更新最后一个对话
        const targetConversation = conversationList.value.find((conv) => conv.id === stream.metadata.conversationId);
        if (targetConversation && stream.content !== undefined) {
          targetConversation.title = stream.content;
          console.log("对话标题已更新到UI:", stream.content);
        }
      }
    }
  },
  { deep: true }
);

// 监听流式输出状态变化
watch(
  () => streamStore.activeStreamsList,
  (streams) => {
    for (const stream of streams) {
      if (stream.type === StreamType.PROMPT_OPTIMIZATION) {
        // 优化提示词状态变化
        if (stream.status === StreamStatus.COMPLETED || stream.status === StreamStatus.ERROR || stream.status === StreamStatus.STOPPED) {
          isOptimizingPrompt.value = false;
        }
      }
    }
  },
  { deep: true }
);

// 处理对话列表滚动事件
const handleConversationListScroll = (e) => {
  const scrollTop = e.target.scrollTop;
  const scrollHeight = e.target.scrollHeight;
  const clientHeight = e.target.clientHeight;

  isShowToBottom.value = scrollHeight - scrollTop - clientHeight > 20;
  isShowToTop.value = scrollTop > 20;
};

// 设置滚动监听器
const setupScrollListener = () => {
  const conversationListElement = document.querySelector(".conversation-list");

  if (conversationListElement) {
    conversationListElement.removeEventListener("scroll", handleConversationListScroll);
    conversationListElement.addEventListener("scroll", handleConversationListScroll);
    handleConversationListScroll({ target: conversationListElement });
  } else {
    setTimeout(setupScrollListener, 100);
  }
};

// 组件挂载时初始化
onMounted(async () => {
  nextTick(() => {
    setupScrollListener();
  });
});

// 组件卸载时清理
onUnmounted(() => {
  const conversationListElement = document.querySelector(".conversation-list");
  if (conversationListElement) {
    conversationListElement.removeEventListener("scroll", handleConversationListScroll);
  }
});

// 切换深度思考模式
const toggleDeepThinking = () => {
  deepThinkingEnabled.value = !deepThinkingEnabled.value;
};

// 打开助手设置对话框
const openAssistantSettings = () => {
  if (selectedAssistantData.value) {
    showAssistantSettings.value = true;
  }
};

// 处理助手设置保存
const handleAssistantSettingsSave = async () => {
  try {
    if (assistantSettingsRef.value && assistantSettingsRef.value.handleSave) {
      await assistantSettingsRef.value.handleSave();
    }

    showAssistantSettings.value = false;
  } catch (error) {
    console.error("助手设置保存失败:", error);
  }
};

// 处理助手设置更新事件
const handleAssistantSettingsUpdate = (updatedAssistant) => {
  selectedAssistantData.value = { ...selectedAssistantData.value, ...updatedAssistant };
  emit("assistant-updated", updatedAssistant);
};

// 处理助手设置取消
const handleAssistantSettingsCancel = () => {
  showAssistantSettings.value = false;
};

// 刷新聊天数据
const refreshChat = async () => {
  if (props.topicId) {
    await loadConversations({ id: props.topicId });
  } else {
    conversationList.value = [];
  }
};

// 处理文件上传
const handleUploadFile = async ({ files, name, e }) => {
  console.log('🚀 ~ handleUploadFile ~ e:', e, files, name);
  
  try {
    // 处理文件（压缩图片等）
    const processedFile = await processFile(files[0]);
    
    // 添加新文件并模拟上传进度
    const newFile = {
      key: processedFile.uuid, // 使用UUID作为唯一key
      name: files[0].name,
      originalName: processedFile.fileName,
      size: processedFile.size,
      status: 'progress', // 上传中状态
      description: '上传中',
      localPath: processedFile.originalPath, // 记录本地路径
      fileSuffix: processedFile.fileSuffix,
      md5: processedFile.md5,
      processedFile: processedFile.processedFile, // 保存处理后的文件对象
    };

    filesList.value = [newFile, ...filesList.value];
    console.log('🚀 ~ handleUploadFile ~ filesList:', filesList);
    
    // 保存文件到后端并处理内容
    try {
      // 调用后端API保存文件
      const savedFile = await SaveFile(files[0].name, processedFile.size, newConversationId.value || 'temp');
      
      if (savedFile && savedFile.id) {
        // 更新文件的key为后端返回的ID
        newFile.key = savedFile.id;
        
        // 调用后端API处理文件内容
        await ProcessFileContent(savedFile.id);
        
        // 更新文件状态为成功
        filesList.value = filesList.value.map((file) =>
          file.key === newFile.key
            ? {
                ...file,
                status: 'success',
                description: formatFileSize(processedFile.size),
              }
            : file,
        );
        
        MessagePlugin.success('文件上传并处理完成');
      }
    } catch (saveError) {
      console.error('保存文件失败:', saveError);
      MessagePlugin.error(`文件保存失败: ${saveError.message}`);
      
      // 移除失败的文件
      filesList.value = filesList.value.filter((item) => item.key !== newFile.key);
    }
  } catch (error) {
    console.error('文件处理失败:', error);
    MessagePlugin.error(`文件处理失败: ${error.message}`);
  }
};

// 处理文件移除
const handleRemoveFile = (e) => {
  const fileToRemove = e.detail;
  filesList.value = filesList.value.filter((item) => item.key !== fileToRemove.key);
};

// 处理文件点击
const handleFileClick = (e) => {
  const clickedFile = e.detail;
  console.log('fileClick', clickedFile);
  
  // 如果是本地文件，可以尝试打开文件所在目录
  if (clickedFile.localPath) {
    // 这里可以添加打开文件或文件所在目录的逻辑
    MessagePlugin.info(`文件路径: ${clickedFile.localPath}`);
  }
};

// 暴露方法给父组件
defineExpose({
  isStreamingChat,
  handleStopChat,
  refreshChat,
});
</script>

<style lang="less" scoped>
.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.assistant-prompt {
  padding: 12px 16px;
  cursor: pointer;
  background-color: var(--td-bg-color-container, #fff);
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
  border-radius: 5px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);

  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
  }
}

.prompt-content {
  font-size: 13px;
  line-height: 1.5;
  color: var(--td-text-color-secondary, #666);
  word-wrap: break-word;
  white-space: pre-wrap;
  font-family:
    "NotoSans SC",
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    "Roboto",
    "Oxygen",
    "Ubuntu",
    "Cantarell",
    "Fira Sans",
    "Droid Sans",
    "Helvetica Neue",
    sans-serif !important;
  height: 40px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  text-overflow: ellipsis;
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
  padding-top: 1px;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-track {
    background: var(--td-scroll-track-color, #f1f1f1);
    border-radius: 3px;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--td-scrollbar-color, #c1c1c1);
    border-radius: 3px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: var(--td-scrollbar-hover-color, #a8a8a8);
  }
}

.conversation-wrapper {
  margin-bottom: 2px;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
  border-radius: 5px;
  background-color: var(--td-bg-color-container, #fff);

  &:hover {
    border-color: var(--td-border-level-2-color, #d9d9d9);
  }

  &.active {
    border-color: var(--td-brand-color, #0052d9);
    box-shadow: 0 0 0 2px rgba(0, 82, 217, 0.1);
  }
}

.conversation-header {
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
  transition: background-color 0.2s ease;
}

.conversation-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--td-text-color-primary, #333);
  flex: 1;
  margin-right: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-time {
  font-size: 12px;
  color: var(--td-text-color-secondary, #666);
  white-space: nowrap;
}

.conversation-content {
  border-top: 1px solid var(--td-border-level-1-color, #e7e7e7);
  overflow-x: auto;
}

.empty-conversation {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
}

.empty-conversation .assistant-prompt {
  margin-bottom: 16px;
  align-self: stretch;
}

.empty-conversation :deep(.t-empty) {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  min-height: 200px;
}

.input-section {
  padding: 16px;
  background-color: var(--td-bg-color-container, #fff);
  border-top: 1px solid var(--td-border-level-1-color, #e7e7e7);
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.06);
}

.back-to-top {
  position: absolute;
  bottom: 320px;
  right: 50px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: var(--td-brand-color, #0052d9);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
  z-index: 100;
  line-height: 1;
  text-align: center;

  &:hover {
    background-color: var(--td-brand-color-hover, #003cab);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  svg {
    font-size: 16px;
    display: block;
    margin: 0 auto;
  }
}

.back-to-bottom {
  position: absolute;
  bottom: 280px;
  right: 50px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: var(--td-brand-color, #0052d9);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
  z-index: 100;
  line-height: 1;
  text-align: center;

  &:hover {
    background-color: var(--td-brand-color-hover, #003cab);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  svg {
    font-size: 16px;
    display: block;
    margin: 0 auto;
  }
}

.model-select {
  display: flex;
  align-items: center;
  gap: 10px;

  .thinking-btn {
    width: 112px;
    height: var(--td-comp-size-m);
    border-radius: 32px;
    border: 0;
    background: var(--td-bg-color-component);
    color: var(--td-text-color-primary);
    box-sizing: border-box;
    flex: 0 0 auto;

    .t-button__text {
      display: flex;
      align-items: center;
      justify-content: center;

      span {
        margin-left: var(--td-comp-margin-xs);
      }
    }

    &.is-active {
      border: 1px solid var(--td-brand-color-focus);
      background: var(--td-brand-color-light);
      color: var(--td-text-color-brand);
    }
  }
}
</style>
