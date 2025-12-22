<template>
  <div class="message-actions">
    <!-- 复制按钮 -->
    <t-tooltip v-if="visibleButtons.copy" content="复制">
      <t-button
        theme="default"
        size="small"
        class="copy-btn custom-action-btn"
        :disabled="disabled"
        @click="handleCopy"
      >
        <IconCopy :size="16" />
      </t-button>
    </t-tooltip>
    
    <!-- 重新生成按钮 -->
    <div v-if="visibleButtons.copy && visibleButtons.replay" class="action-divider"></div>
    <t-tooltip v-if="visibleButtons.replay" content="重新生成">
      <t-button
        theme="default"
        size="small"
        class="custom-action-btn"
        :disabled="disabled"
        @click="handleReplay"
      >
        <IconRefreshCw :size="16" />
      </t-button>
    </t-tooltip>
    
    <!-- 编辑按钮 -->
    <div v-if="(visibleButtons.copy || visibleButtons.replay) && visibleButtons.edit" class="action-divider"></div>
    <t-tooltip v-if="visibleButtons.edit" :content="isEditing ? '保存' : '编辑'">
      <t-button
        theme="default"
        size="small"
        class="custom-action-btn edit-btn"
        :disabled="disabled"
        @click="handleEdit"
      >
        <IconCheck v-if="isEditing" :size="16" />
        <IconEdit v-else :size="16" />
      </t-button>
    </t-tooltip>
    
    <!-- 导出按钮 -->
    <div v-if="(visibleButtons.copy || visibleButtons.replay || visibleButtons.edit) && visibleButtons.export" class="action-divider"></div>
    <t-tooltip v-if="visibleButtons.export" content="导出消息">
      <t-button
        theme="default"
        size="small"
        class="custom-action-btn"
        :disabled="disabled"
        @click="handleExport"
      >
        <IconDownload :size="16" />
      </t-button>
    </t-tooltip>
    
    <!-- 删除按钮 -->
    <div v-if="(visibleButtons.copy || visibleButtons.replay || visibleButtons.export) && visibleButtons.delete" class="action-divider"></div>
    <t-tooltip v-if="visibleButtons.delete" content="删除消息">
      <t-button
        theme="default"
        size="small"
        class="custom-action-btn delete-btn"
        :disabled="disabled"
        @click="handleDelete"
      >
        <IconTrash2 :size="16" />
      </t-button>
    </t-tooltip>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { cloneDeep } from "lodash";

// Props
const props = defineProps({
  // 消息内容，用于复制等功能
  content: {
    type: String,
    default: ''
  },
  // 消息对象
  message: {
    type: Object,
    required: true
  },
  // 是否禁用所有按钮
  disabled: {
    type: Boolean,
    default: false
  },
  // 控制按钮显示的对象
  visibleButtons: {
    type: Object,
    default: () => ({
      copy: false,   // 复制按钮
      replay: false, // 重新生成按钮
      export: false, // 导出按钮
      delete: false, // 删除按钮
      edit: false    // 编辑按钮
    })
  },
  // 是否处于编辑状态
  isEditing: {
    type: Boolean,
    default: false
  }
});

// 定义中间值来接收props
const messageData = ref({});

// 初始化方法
const init = () => {
  messageData.value = cloneDeep(props.message);
};

// 组件初始化时调用init
init();

// 监听props.message变化，更新messageData
watch(
  () => props.message,
  (newMessage) => {
    messageData.value = cloneDeep(newMessage);
  },
  { deep: true }
);

// Emits
const emit = defineEmits(['copy', 'replay', 'export', 'delete', 'edit', 'save']);

// 处理复制操作
const handleCopy = async () => {
  try {
    const textToCopy = props.content || messageData.value.content || messageData.value.reasoning || '';
    
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(textToCopy);
      MessagePlugin.success('复制成功');
    } else {
      //降级方案
      const textArea = document.createElement('textarea');
      textArea.value = textToCopy;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      textArea.style.top = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      MessagePlugin.success('复制成功');
    }
    
    emit('copy', messageData.value);
  } catch (err) {
    MessagePlugin.error('复制失败');
  }
};

// 处理重新生成操作
const handleReplay = () => {
  emit('replay', messageData.value);
};

// 处理导出操作
const handleExport = () => {
  try {
    // 创建导出内容
    const exportData = {
      role: messageData.value.role,
      name: messageData.value.name,
      content: messageData.value.content || '',
      reasoning: messageData.value.reasoning || '',
      datetime: messageData.value.datetime,
      timestamp: new Date().toISOString()
    };

    // 转换为 JSON 字符串
    const jsonString = JSON.stringify(exportData, null, 2);
    
    // 创建 Blob 对象
    const blob = new Blob([jsonString], { type: 'application/json' });
    
    // 创建下载链接
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `message_${messageData.value.role}_${Date.now()}.json`;
    
    // 触发下载
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    // 清理 URL 对象
    URL.revokeObjectURL(url);
    
    MessagePlugin.success('消息导出成功');
    emit('export', messageData.value);
  } catch (error) {
    MessagePlugin.error('导出失败');
  }
};

// 处理编辑操作
const handleEdit = () => {
  if (props.isEditing) {
    // 保存操作
    emit('save', messageData.value);
  } else {
    // 开始编辑
    emit('edit', messageData.value);
  }
};

// 处理删除操作
const handleDelete = () => {
  emit('delete', messageData.value);
};
</script>

<style lang="less" scoped>
.message-actions {
  display: flex;
  align-items: center;
  gap: 0; // 移除默认间距，使用分隔线控制间距
}

.action-divider {
  width: 1px;
  height: 16px;
  background-color: var(--td-border-level-1-color, #e7e7e7);
  margin: 0 4px; // 分隔线左右的间距
}

.custom-action-btn {
  // 仿照 TDesign Chat 的按钮样式
  min-width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 6px;
  border: none;
  background-color: transparent;
  color: var(--td-text-color-primary, #333);
  transition: all 0.2s ease;
  
  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
  }
  
  &:active {
    background-color: var(--td-bg-color-container-active, #e8e8e8);
  }
  
  // 删除按钮特殊样式
  &.delete-btn {
    color: var(--td-error-color, #e34d59);
    
    &:hover {
      background-color: var(--td-error-color-1, #fcebeb);
    }
  }
  
  // 编辑按钮特殊样式
  &.edit-btn {
    color: var(--td-brand-color, #0052d9);
    
    &:hover {
      background-color: var(--td-brand-color-1, #e0e8ff);
    }
  }
  
  // 图标样式
  svg {
    font-size: 16px;
  }
}

// 响应式设计
@media (max-width: 768px) {
  .message-actions {
    flex-wrap: wrap;
    gap: 4px;
  }
  
  .action-divider {
    display: none; // 在小屏幕上隐藏分隔线
  }
  
  .custom-action-btn {
    min-width: 28px;
    height: 28px;
    
    svg {
      font-size: 14px;
    }
  }
}

// 深色模式适配
@media (prefers-color-scheme: dark) {
  .custom-action-btn {
    background-color: transparent;
    color: var(--td-text-color-primary, #e7e7e7);
    
    &:hover {
      background-color: var(--td-bg-color-container-hover, #2a2a2a);
    }
    
    &:active {
      background-color: var(--td-bg-color-container-active, #333333);
    }
  }
  
  .action-divider {
    background-color: var(--td-border-level-1-color, #393939);
  }
}
</style>
