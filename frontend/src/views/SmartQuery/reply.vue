<template>
  <div class="message-reply-container">
    <t-chat-message
      :avatar="currentMessage.avatar"
      :datetime="currentMessage.datetime"
      :name="currentMessage.name"
      role="assistant"
      placement="left"
      variant="base"
    >
      <!-- AI助手头像插槽 -->
      <template #avatar v-if="currentMessage.emoji">
        <div class="assistant-avatar">
          {{ currentMessage.emoji }}
        </div>
      </template>

      <!-- 自定义消息内容 -->
      <template #content>
        <!-- 思考过程 -->
        <t-chat-thinking
          v-if="currentMessage.reasoning && currentMessage.reasoning.length > 0"
          :content="{
            text: currentMessage.reasoning,
            title: isThinking ? '正在思考中...' : '已深度思考',
          }"
          :status="isThinking ? 'pending' : 'complete'"
          :max-height="100"
          :collapsed="thinkingCollapsed"
          animation="dots"
          @collapsed-change="handleCollapsedChange"
        />

        <!-- 回复内容 -->
        <div class="markdown-content-wrapper">
          <MarkdownRender v-if="currentMessage.content && currentMessage.content.length > 0" :content="currentMessage.content" />
        </div>
      </template>

      <!-- 操作按钮 -->
      <template #actionbar>
        <div class="actions-container">
          <t-chat-loading v-if="isStreamingChat && !isThinking" animation="dots" />
          <div v-if="!isStreamingChat && !isThinking" class="action-buttons">
            <MessageAction
              :content="currentMessage.content || currentMessage.reasoning || ''"
              :message="currentMessage"
              :visible-buttons="{
                copy: true,
                replay: true,
                export: true,
                delete: true,
              }"
              @replay="handleRegenerate"
              @export="handleExport"
              @delete="handleDelete"
            />
            <MessageSwitch
              v-if="messagesData.length > 1"
              :current-index="currentIndex + 1"
              :total-count="messagesData.length"
              :disabled="isStreamingChat"
              @prev="handlePrev"
              @next="handleNext"
            />
          </div>
        </div>
      </template>
    </t-chat-message>
  </div>
</template>

<script setup>
import { computed, watch, ref } from "vue";
import MessageAction from "./action.vue";
import MessageSwitch from "./switch.vue";
import { cloneDeep } from "lodash";
import MarkdownRender from "markstream-vue";
import "markstream-vue/index.css";

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
  isThinking: {
    type: Boolean,
    default: false,
  },
});

// 定义中间值来接收props
const messagesData = ref([]);

// 定义思考过程折叠状态
const thinkingCollapsed = ref(true);

// 初始化方法
const init = () => {
  messagesData.value = cloneDeep(props.messages);
};

// 组件初始化时调用init
init();

// 监听props.messages变化，更新messagesData
watch(
  () => props.messages,
  (newMessages) => {
    messagesData.value = cloneDeep(newMessages);
  },
  { deep: true }
);

// 监听isThinking状态变化，自动调整折叠状态
watch(
  () => props.isThinking,
  (newIsThinking) => {
    thinkingCollapsed.value = !newIsThinking;
  }
);

// Emits
const emit = defineEmits(["regenerate-message", "export-message", "delete-message", "index-change"]);

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

// 处理折叠状态变化事件
const handleCollapsedChange = (e) => {
  // t-chat-thinking 组件的折叠状态变化处理
  // e.detail 包含折叠状态信息
  if (e && e.detail !== undefined) {
    thinkingCollapsed.value = e.detail;
  }
};

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
</script>

<style lang="less" scoped>
.actions-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-top: 10px;
}

.action-buttons {
  display: flex;
  align-items: center;
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

// Markdown 内容容器样式
.markdown-content-wrapper {
  width: calc(100% - 24px);
  max-width: 100%;
  overflow-x: auto;
  word-wrap: break-word;
  word-break: break-word;
  min-width: 0; // 防止flex子项溢出
  display: block; // 确保块级显示
  background-color: #dffadf; // 低饱和度浅绿色背景，适合阅读
  border-radius: 6px;
  padding-left: 12px;
  padding-right: 12px;
}

// 深度选择器：针对 table-node-wrapper 类名修复表格样式
:deep(.table-node-wrapper) {
  width: 100%;
  max-width: 100%;
  overflow-x: auto;

  table {
    width: 100%;
    max-width: 100%;
    border-collapse: collapse;
    table-layout: fixed;
    word-wrap: break-word;
    word-break: break-word;

    th,
    td {
      word-wrap: break-word;
      word-break: break-word;
      white-space: normal;
      padding: 8px 12px;
      text-align: left;
      border: 1px solid #e0e0e0;
      vertical-align: top;
      min-width: 80px;
      max-width: 200px;
      overflow: hidden;
      text-overflow: ellipsis;
      line-height: 1.4;
    }

    th {
      background-color: #f5f5f5;
      font-weight: 600;
      color: #333;
    }

    td {
      background-color: #fff;
      color: #666;
    }

    // 响应式表格
    @media (max-width: 768px) {
      th,
      td {
        max-width: 150px;
        font-size: 14px;
        padding: 6px 8px;
      }
    }

    @media (max-width: 480px) {
      th,
      td {
        max-width: 100px;
        font-size: 12px;
        padding: 4px 6px;
      }
    }
  }
}

// 深度选择器：修复其他可能的表格容器
:deep(.markdown-render) {
  width: 100%;
  max-width: 100%;

  .table-node-wrapper {
    width: 100%;
    max-width: 100%;
    overflow-x: auto;
  }

  table {
    width: 100% !important;
    max-width: 100% !important;
    table-layout: fixed !important;
    word-wrap: break-word !important;
    word-break: break-word !important;

    th,
    td {
      word-wrap: break-word !important;
      word-break: break-word !important;
      white-space: normal !important;
      max-width: 200px !important;
      overflow: hidden !important;
      text-overflow: ellipsis !important;
    }
  }
}

// MarkdownRender 组件背景色
:deep(.markdown-content-wrapper .markdown-render) {
  background-color: #f5faf5 !important; // 低饱和度浅绿色背景，适合阅读
  border-radius: 6px;
  padding: 12px;
}

// 修复代码块和其他可能溢出的元素
:deep(pre) {
  max-width: 100%;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

:deep(code) {
  word-wrap: break-word;
  word-break: break-word;
}

:deep(img) {
  max-width: 100%;
  height: auto;
}
</style>
