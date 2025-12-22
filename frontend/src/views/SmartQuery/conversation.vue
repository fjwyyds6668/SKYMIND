<template>
  <div class="conversation-container">
    <!-- 聊天消息区域 -->
    <div class="messages-wrapper" ref="messagesContainer" @scroll="handleChatScroll">
      <!-- 用户消息组件 -->
      <MessageSend
        :messages="userMessages"
        :current-id="currentSendId"
        :is-streaming-chat="isStreamingChat && currentChatStream"
        @regenerate-message="handleRegenerate"
        @export-message="handleExport"
        @delete-message="handleDelete"
        @edit-message="handleEditMessage"
        @save-message="handleSaveMessage"
        @index-change="handleSendIndexChange"
      />

      <!-- AI回复消息组件 -->
      <MessageReply
        :messages="assistantMessages"
        :current-id="currentReplyId"
        :is-streaming-chat="isStreamingChat && currentChatStream"
        :is-thinking="currentChatStream?.thinkingPhase || false"
        @regenerate-message="handleRegenerate"
        @index-change="handleReplyIndexChange"
        @export-message="handleExport"
        @delete-message="handleDelete"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed, reactive, nextTick } from "vue";
import { MessagePlugin, DialogPlugin } from "tdesign-vue-next";
import MessageSend from "./send.vue";
import MessageReply from "./reply.vue";
import { DeleteMessage, CreateMessage, UpdateConversationSettings } from "../../../wailsjs/go/main/App";
import { cloneDeep } from "lodash";
import { formatDateTime } from "./utils.js";
import { useStreamStore } from "../../store/modules/stream.js";

// Props
const props = defineProps({
  conversation: {
    type: Object,
    required: true,
  },
  selectedAssistant: {
    type: Object,
    default: null,
  },
  isStreamingChat: {
    type: Boolean,
    default: false,
  },
  streamingContent: {
    type: String,
    default: "",
  },
});

// Store
const streamStore = useStreamStore();

// Emits
const emit = defineEmits(["regenerate-message", "delete-message", "conversation-deleted", "config-change", "messages-change"]);

// 响应式数据
const messagesContainer = ref(null);
const currentSendIndex = ref(0);
const currentReplyIndex = ref(0);

// 定义中间值来接收props
const conversationData = ref({});
const selectedAssistantData = ref({});

// 初始化方法
const init = () => {
  conversationData.value = cloneDeep(props.conversation);
  selectedAssistantData.value = cloneDeep(props.selectedAssistant);
};

// 组件初始化时调用init
init();

// 监听props.conversation变化，更新conversationPropsData
watch(
  () => props.conversation,
  (newConversation) => {
    conversationData.value = cloneDeep(newConversation);
    // 强制重新计算计算属性
    nextTick(() => {});
  },
  { deep: true, immediate: true }
);

// 监听props.selectedAssistant变化，更新selectedAssistantData
watch(
  () => props.selectedAssistant,
  (newAssistant) => {
    selectedAssistantData.value = cloneDeep(newAssistant);
  },
  { deep: true }
);

// 获取当前聊天的流式输出内容
const currentChatStream = computed(() => {
  const chatStreams = streamStore.getActiveStreamsByConversationId(conversationData.value.id);
  return chatStreams.length > 0 ? chatStreams[0] : null;
});

// 计算当前显示的消息ID
const currentSendId = computed(() => {
  if (!conversationData.value?.settings) return "";
  try {
    const settings = JSON.parse(conversationData.value.settings);
    return settings.currentSendId || "";
  } catch (error) {
    return "";
  }
});

const currentReplyId = computed(() => {
  if (!conversationData.value?.settings) return "";
  try {
    const settings = JSON.parse(conversationData.value.settings);
    return settings.currentReplyId || "";
  } catch (error) {
    return "";
  }
});

// 计算属性
const displayMessages = computed(() => {
  // 使用可选链操作符简化空值检查
  if (!conversationData.value?.messages?.length) {
    return [];
  }

  // 缓存助手信息，避免重复访问
  const assistantName = selectedAssistantData.value?.name || "AI助手";
  const assistantEmoji = selectedAssistantData.value?.emoji || "🤖";

  // 创建消息的深拷贝，确保每个组件都有独立的数据副本
  return conversationData.value.messages.map((msg) => {
    const messageCopy = { ...msg };
    return reactive({
      id: messageCopy.id,
      conversationId: messageCopy.conversation_id || conversationData.value?.id, // 确保有conversationId
      avatar: null, // 所有消息都不使用图片头像
      name: messageCopy.role === "user" ? "用户" : assistantName,
      datetime: formatDateTime(messageCopy.datetime || messageCopy.created_at),
      content: messageCopy.content,
      role: messageCopy.role,
      reasoning: messageCopy.reasoning || "", // 添加reasoning属性
      emoji: messageCopy.role === "user" ? "😊" : assistantEmoji, // 为所有消息添加emoji属性
    });
  });
});

// 用户消息计算属性
const userMessages = computed(() => {
  return displayMessages.value.filter((message) => message.role === "user");
});

// AI消息计算属性
const assistantMessages = computed(() => {
  return displayMessages.value.filter((message) => message.role === "assistant");
});

// 监听对话变化，重置索引
watch(
  () => conversationData.value?.id,
  (newId, oldId) => {
    if (newId !== oldId) {
      currentSendIndex.value = 0;
      currentReplyIndex.value = 0;
    }
  }
);

// 方法
const handleChatScroll = (e) => {
  // 滚动事件处理，现在不需要跟踪滚动状态
  // 相关逻辑已移至 index.vue 中处理
};

const handleRegenerate = async (message) => {
  let userMessage = null;

  if (message.role === "user") {
    // 如果是用户消息，直接使用该用户消息
    userMessage = message;
  } else if (message.role === "assistant") {
    // 如果是AI回复，根据ConversationSettings中的currentSendId找到对应的用户消息
    const currentSendIdValue = currentSendId.value;
    if (currentSendIdValue) {
      // 在所有消息中查找对应的用户消息
      userMessage = displayMessages.value.find((m) => m.id === currentSendIdValue && m.role === "user");
    }

    // 如果通过currentSendId找不到，则使用备用逻辑：查找该AI消息的上一条消息
    if (!userMessage) {
      const messageIndex = displayMessages.value.findIndex((m) => m.id === message.id);
      if (messageIndex > 0) {
        const prevMessage = displayMessages.value[messageIndex - 1];
        if (prevMessage.role === "user") {
          userMessage = prevMessage;
        }
      }
    }
  }

  if (userMessage) {
    // 通知父组件重新生成，包含当前对话ID以确保消息保存到正确的对话
    emit("regenerate-message", {
      userMessage: userMessage,
      conversationId: conversationData.value?.id, // 添加当前对话ID
    });
  } else {
    MessagePlugin.error("找不到对应的用户消息");
  }
};

// 处理导出操作
const handleExport = (message) => {
  // 导出功能已在 MessageAction 组件中实现
  // 这里可以添加额外的导出逻辑，比如记录日志等
};

// 处理删除操作
const handleDelete = async (message) => {
  const dialog = DialogPlugin.confirm({
    header: "确认删除",
    body: `确定要删除这条${message.role === "assistant" ? "AI助手" : "用户"}消息吗？此操作不可撤销。`,
    confirmBtn: "确定删除",
    cancelBtn: "取消",
    onConfirm: async () => {
      try {
        await DeleteMessage(message.id, message.conversationId);

        // 向父组件发送messages-change事件，让父组件重新加载数据
        emit("messages-change");

        MessagePlugin.success("消息已删除");
      } catch (error) {
        MessagePlugin.error("删除消息失败: " + (error.message || "未知错误"));
      }
      dialog.destroy();
    },
    onCancel: () => {
      dialog.destroy();
    },
  });
};

// 处理用户消息索引变化
const handleSendIndexChange = (newIndex) => {
  currentSendIndex.value = newIndex;
  updateSendConversationConfig();
};

// 处理AI回复消息索引变化
const handleReplyIndexChange = (newIndex) => {
  currentReplyIndex.value = newIndex;
  updateReplyConversationConfig();
};

// 处理编辑消息事件
const handleEditMessage = (message) => {
  // 编辑功能已在 MessageSend 组件中实现
  // 这里可以添加额外的编辑逻辑，比如记录日志等
};

// 处理保存消息事件
const handleSaveMessage = async (newMessage) => {
  try {
    // 调用后端API保存新消息
    const savedMessage = await CreateMessage({
      conversation_id: conversationData.value?.id,
      topic_id: conversationData.value?.topic_id,
      assistant_id: conversationData.value?.assistant_id,
      role: "user",
      content: newMessage.content,
      model: "",
      token_count: 0,
      metadata: "{}",
    });

    if (savedMessage && savedMessage.id) {
      // 将新消息添加到本地数组
      conversationData.value.messages.push({
        ...savedMessage,
        datetime: newMessage.datetime,
      });

      // 强制触发响应式更新
      conversationData.value = { ...conversationData.value };

      MessagePlugin.success("消息已保存");

      // 立即更新对话设置中的CurrentSendID（仿照AI回复消息的逻辑）
      let settings = {};
      if (conversationData.value?.settings) {
        try {
          settings = JSON.parse(conversationData.value.settings);
        } catch (error) {
          settings = {};
        }
      }
      settings.currentSendId = savedMessage.id;

      // 调用后端API更新对话设置
      await UpdateConversationSettings(conversationData.value?.id, JSON.stringify(settings));
      // 更新本地设置
      conversationData.value.settings = JSON.stringify(settings);

      // 等待DOM更新
      await nextTick();

      // 通知父组件配置已更改（关键步骤！）
      emit("config-change", {
        conversationId: conversationData.value?.id,
        settings: JSON.stringify(settings),
      });

      // 等待父组件更新props
      await nextTick();

      // 直接更新本地索引，确保send.vue中的currentIndex计算属性能正确计算
      // 新消息应该是最后一个用户消息，所以索引应该是userMessages.length - 1
      const newUserMessages = conversationData.value.messages.filter((msg) => msg.role === "user");
      const newIndex = newUserMessages.length - 1;
      if (newIndex >= 0) {
        // 直接更新本地索引，避免循环调用
        currentSendIndex.value = newIndex;
      }
    }
  } catch (error) {
    MessagePlugin.error("保存消息失败: " + (error.message || "未知错误"));
  }
};

// 更新用户消息对话配置
const updateSendConversationConfig = () => {
  const currentSendMessage = userMessages.value[currentSendIndex.value];

  let settings = {};
  if (conversationData.value?.settings) {
    try {
      settings = JSON.parse(conversationData.value.settings);
    } catch (error) {
      MessagePlugin.warn("解析现有设置失败，使用默认设置:");
    }
  }

  // 只更新用户消息ID
  if (currentSendMessage) {
    settings.currentSendId = currentSendMessage.id;
  }

  // 通知父组件配置已更改
  emit("config-change", {
    conversationId: conversationData.value?.id,
    settings: JSON.stringify(settings),
  });
};

// 更新AI回复消息对话配置
const updateReplyConversationConfig = () => {
  const currentReplyMessage = assistantMessages.value[currentReplyIndex.value];

  let settings = {};
  if (conversationData.value?.settings) {
    try {
      settings = JSON.parse(conversationData.value.settings);
    } catch (error) {
      MessagePlugin.warn("解析现有设置失败，使用默认设置:");
    }
  }

  // 只更新AI回复消息ID
  if (currentReplyMessage) {
    settings.currentReplyId = currentReplyMessage.id;
  }

  // 通知父组件配置已更改
  emit("config-change", {
    conversationId: conversationData.value?.id,
    settings: JSON.stringify(settings),
  });
};

// 监听 streamingContent 变化，更新最后一条AI消息
watch(
  () => props.streamingContent,
  (newContent) => {
    // 只有在当前对话是活跃状态时才更新流式内容
    if (props.isActive && newContent && newContent.trim()) {
      const lastMessage = displayMessages.value[displayMessages.value.length - 1];

      if (lastMessage && lastMessage.role === "assistant") {
        lastMessage.content = newContent;
      }
    }
  },
  { immediate: true }
); // 添加 immediate 选项以确保初始值也被处理

defineExpose({
  messagesContainer,
});
</script>

<style lang="less" scoped>
.conversation-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;
}

.messages-wrapper {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background-color: var(--td-bg-color-container, #fff);
  margin: 8px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

  // 滚动条样式
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

.back-to-bottom {
  position: absolute;
  bottom: 80px;
  right: 24px;
  width: 40px;
  height: 40px;
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

  &:hover {
    background-color: var(--td-brand-color-hover, #003cab);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  svg {
    font-size: 20px;
  }
}

// 回答切换器样式
.answer-switcher {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 8px 12px;
  background-color: rgba(0, 82, 204, 0.1);
  border-radius: 6px;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
}

.answer-counter {
  font-size: 12px;
  color: var(--td-text-color-secondary, #666);
  font-weight: 500;
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
}

.answer-buttons {
  display: flex;
  gap: 4px;
}

// AI助手头像样式
.assistant-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #333;
  flex-shrink: 0;
}

// 用户头像样式
.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #81c784;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: white;
  flex-shrink: 0;
}
</style>
