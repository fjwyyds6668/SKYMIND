<template>
  <div class="message-send-container">
    <t-chat-message
      :avatar="currentMessage.avatar"
      :datetime="currentMessage.datetime"
      :name="currentMessage.name"
      role="user"
      placement="right"
      variant="base"
      :class="{
        'user-chat-item': true,
      }"
    >
      <!-- 用户头像插槽 -->
      <template #avatar v-if="currentMessage.emoji">
        <div class="user-avatar">
          {{ currentMessage.emoji }}
        </div>
      </template>

      <!-- 主要内容 -->
      <template v-if="currentMessage.content && currentMessage.content.length > 0" #content>
        <div v-if="isEditing" class="edit-content-container">
          <t-textarea
            v-model="editContent"
            :autosize="{ minRows: 1, maxRows: 8 }"
            placeholder="请输入内容..."
            class="edit-textarea"
            @blur="handleTextareaBlur"
          />
        </div>
        <div v-else class="message-content">
          {{ currentMessage.content }}
        </div>
      </template>

      <!-- 操作按钮 -->
      <template #actionbar>
        <div class="actions-container">
          <MessageAction
            v-if="messagesData.length > 0"
            :content="currentMessage.content || currentMessage.reasoning || ''"
            :message="currentMessage"
            :is-editing="isEditing"
            :visible-buttons="{
              copy: true,
              replay: !props.isStreamingChat,
              export: false,
              delete: !props.isStreamingChat,
              edit: !props.isStreamingChat,
            }"
            @replay="handleRegenerate"
            @export="handleExport"
            @delete="handleDelete"
            @edit="handleEdit"
            @save="handleSave"
          />
          <MessageSwitch
            v-if="messagesData.length > 1"
            :current-index="currentIndex + 1"
            :total-count="messagesData.length"
            :disabled="props.isStreamingChat"
            @prev="handlePrev"
            @next="handleNext"
          />
        </div>
      </template>
    </t-chat-message>
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import MessageAction from "./action.vue";
import MessageSwitch from "./switch.vue";
import { cloneDeep } from "lodash";

// Props
const props = defineProps({
  messages: {
    type: Array,
    required: true,
    default: () => [],
  },
  currentId: {
    type: String,
    default: "",
  },
  isStreamingChat: {
    type: Boolean,
    default: false,
  },
});

// Emits
const emit = defineEmits(["regenerate-message", "export-message", "delete-message", "index-change", "edit-message", "save-message"]);

// 响应式数据
const isEditing = ref(false);
const editContent = ref("");

// 定义中间值来接收props.messages
const messagesData = ref([]);

// 初始化方法
const init = () => {
  messagesData.value = cloneDeep(props.messages);
};

// 组件初始化时调用init
init();

// 计算当前显示的消息索引
const currentIndex = computed(() => {
  if (!props.currentId || messagesData.value.length === 0) return 0;
  const index = messagesData.value.findIndex((msg) => msg.id === props.currentId);
  return index !== -1 ? index : 0;
});

// 计算当前显示的消息
const currentMessage = computed(() => {
  return messagesData.value[currentIndex.value] || {};
});

// 监听props.messages变化，更新messagesData
watch(
  () => props.messages,
  (newMessages) => {
    messagesData.value = cloneDeep(newMessages);
  },
  { deep: true }
);

// 监听当前消息变化，重置编辑状态
watch(
  () => currentIndex.value,
  () => {
    isEditing.value = false;
    editContent.value = "";
  }
);

// 监听当前消息内容变化，更新编辑内容
watch(
  () => currentMessage.value.content,
  (newContent) => {
    if (!isEditing.value && newContent) {
      editContent.value = newContent;
    }
  },
  { immediate: true }
);

// 方法
const handleRegenerate = (message) => {
  emit("regenerate-message", message);
};

const handleExport = (message) => {
  emit("export-message", message);
};

const handleDelete = (message) => {
  emit("delete-message", message);
};

const handlePrev = (newIndex) => {
  emit("index-change", newIndex - 1); // MessageSwitch 使用 1-based 索引
};

const handleNext = (newIndex) => {
  emit("index-change", newIndex - 1); // MessageSwitch 使用 1-based 索引
};

// 处理编辑操作
const handleEdit = (message) => {
  isEditing.value = true;
  editContent.value = message.content || "";
  emit("edit-message", message);
};

// 处理保存操作
const handleSave = (message) => {
  if (editContent.value.trim() === "") {
    return; // 内容为空不保存
  }

  if (editContent.value === (message.content || "")) {
    // 内容未改变，直接退出编辑模式
    isEditing.value = false;
    return;
  }

  // 内容已改变，保存新消息（不包含原始消息的ID，让后端生成新ID）
  const newMessage = {
    // 不使用 ...message，避免保留原始消息ID
    conversation_id: message.conversationId || message.conversation_id,
    topic_id: message.topicId || message.topic_id,
    assistant_id: message.assistantId || message.assistant_id,
    role: "user", // 明确设置为用户消息
    content: editContent.value.trim(),
    datetime: new Date().toISOString(),
    // 不包含id字段，让后端自动生成
  };

  emit("save-message", newMessage);
  isEditing.value = false;
};

// 处理文本框失焦
const handleTextareaBlur = () => {
  // 延迟处理，避免与保存按钮点击冲突
  setTimeout(() => {
    if (isEditing.value) {
      handleSave(currentMessage.value);
    }
  }, 200);
};
</script>

<style lang="less" scoped>
.message-send-container {
  margin-bottom: 16px;
}

.actions-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-top: 10px;
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

// 消息内容样式
.message-content {
  padding: 16px 16px;
  background-color: var(--td-brand-color, #0052d9);
  color: white;
  border-radius: 8px 8px 8px 8px;
  font-size: 14px;
  line-height: 1.5;
  word-wrap: break-word;
  max-width: 100%;
}

// 编辑内容容器样式
.edit-content-container {
  width: 100%;

  .edit-textarea {
    width: 100%;
    min-height: 30px;
    padding: 8px;
    border: 2px solid var(--td-brand-color, #0052d9);
    border-radius: 8px;
    background-color: var(--td-bg-color-container, #fff);
    color: var(--td-text-color-primary, #333);
    font-size: 14px;
    line-height: 1;
    resize: vertical;
    transition: border-color 0.2s ease;

    &:focus {
      outline: none;
      border-color: var(--td-brand-color, #0052d9);
      box-shadow: 0 0 0 2px var(--td-brand-color-1, #e0e8ff);
    }

    &::placeholder {
      color: var(--td-text-color-placeholder, #999);
    }
  }
}

// 深色模式适配
@media (prefers-color-scheme: dark) {
  .edit-content-container {
    .edit-textarea {
      background-color: var(--td-bg-color-container, #1f1f1f);
      color: var(--td-text-color-primary, #e7e7e7);
      border-color: var(--td-brand-color, #0052d9);

      &::placeholder {
        color: var(--td-text-color-placeholder, #666);
      }
    }
  }
}
</style>
