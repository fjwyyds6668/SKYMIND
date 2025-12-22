<template>
  <div class="message-switch">
    <t-tooltip content="上一个回答">
      <t-button
        theme="default"
        size="small"
        class="switch-btn prev-btn"
        :disabled="currentIndex <= 1 || disabled"
        @click="handlePrev"
      >
        <IconChevronLeft :size="14" />
      </t-button>
    </t-tooltip>
    
    <span class="switch-counter">{{ currentIndex }}/{{ totalCount }}</span>
    
    <t-tooltip content="下一个回答">
      <t-button
        theme="default"
        size="small"
        class="switch-btn next-btn"
        :disabled="currentIndex >= totalCount || disabled"
        @click="handleNext"
      >
        <IconChevronRight :size="14" />
      </t-button>
    </t-tooltip>
  </div>
</template>

<script setup>

// Props
const props = defineProps({
  currentIndex: {
    type: Number,
    default: 1
  },
  totalCount: {
    type: Number,
    default: 1
  },
  disabled: {
    type: Boolean,
    default: false
  }
});

// Emits
const emit = defineEmits(['prev', 'next']);

// 处理上一个回答
const handlePrev = () => {
  if (props.currentIndex > 1) {
    emit('prev', props.currentIndex - 1);
  }
};

// 处理下一个回答
const handleNext = () => {
  if (props.currentIndex < props.totalCount) {
    emit('next', props.currentIndex + 1);
  }
};
</script>

<style lang="less" scoped>
.message-switch {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: 8px;
}

.switch-btn {
  min-width: 24px;
  height: 24px;
  padding: 0;
  border-radius: 4px;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
  background-color: var(--td-bg-color-container, #fff);
  color: var(--td-text-color-primary, #333);
  transition: all 0.2s ease;
  
  &:hover:not(:disabled) {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
    border-color: var(--td-border-level-2-color, #d9d9d9);
  }
  
  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  
  svg {
    font-size: 14px;
  }
}

.switch-counter {
  font-size: 12px;
  color: var(--td-text-color-secondary, #666);
  font-weight: 500;
  min-width: 40px;
  text-align: center;
  font-family: "NotoSans SC", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto",
    "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
    sans-serif !important;
}

.prev-btn {
  margin-right: 4px;
}

.next-btn {
  margin-left: 4px;
}

// 深色模式适配
@media (prefers-color-scheme: dark) {
  .switch-btn {
    background-color: var(--td-bg-color-container, #1f1f1f);
    color: var(--td-text-color-primary, #e7e7e7);
    border-color: var(--td-border-level-1-color, #393939);
    
    &:hover:not(:disabled) {
      background-color: var(--td-bg-color-container-hover, #2a2a2a);
      border-color: var(--td-border-level-2-color, #4a4a4a);
    }
  }
  
  .switch-counter {
    color: var(--td-text-color-secondary, #999);
  }
}
</style>
