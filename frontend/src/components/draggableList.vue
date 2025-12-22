<template>
  <div class="draggable-list">
    <draggable
      :model-value="localItems"
      :animation="200"
      :disabled="disabled"
      ghost-class="ghost"
      chosen-class="chosen"
      drag-class="drag"
      @end="onDragEnd"
      item-key="id"
    >
      <template #item="{ element, index }">
        <div 
          :class="getItemClass(element)" 
          @click="handleClick(element)"
        >
          <!-- 拖拽手柄插槽 -->
          <slot name="drag-handle" :element="element" :index="index">
            <div class="drag-handle">
              <IconGripVertical class="drag-icon" :size="16" />
            </div>
          </slot>

          <!-- 主要内容插槽 -->
          <slot name="item" :element="element" :index="index" :isSelected="selectedId === element.id">
            <!-- 默认内容：保持向后兼容 -->
            <div class="item-content">
              <div v-if="element.emoji" class="item-emoji">{{ element.emoji }}</div>
              <div class="item-info">
                <div class="item-name">{{ element.name }}</div>
                <div v-if="element.description" class="item-desc">{{ element.description }}</div>
                <div v-else-if="element.created_at" class="item-desc">{{ formatTime(element.created_at) }}</div>
              </div>
            </div>
          </slot>

          <!-- 操作区域插槽 -->
          <slot name="item-actions" :element="element" :index="index" :isSelected="selectedId === element.id">
            <!-- 默认操作：保持向后兼容 -->
            <div v-if="showDeleteButton && selectedId === element.id" class="delete-button">
              <t-popconfirm
                :content="getDeleteConfirmContent(element)"
                placement="right"
                :overlay-style="{ width: '400px' }"
                @confirm="handleDelete(element)"
              >
                <IconX :size="16" />
              </t-popconfirm>
            </div>
            <div v-else-if="showTopicCount && selectedId === element.id" class="topic-count">
              {{ selectedAssistantTopicsData.length }}
            </div>
          </slot>
        </div>
      </template>
    </draggable>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from "vue";
import draggable from "vuedraggable";
import { cloneDeep } from "lodash";

// Props
const props = defineProps({
  items: {
    type: Array,
    default: () => [],
  },
  selectedId: {
    type: String,
    default: "",
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  showDeleteButton: {
    type: Boolean,
    default: false,
  },
  showTopicCount: {
    type: Boolean,
    default: false,
  },
  selectedAssistantTopics: {
    type: Array,
    default: () => [],
  },
});

// Emits
const emit = defineEmits(["item-click", "drag-end", "item-delete"]);

// 定义中间值来接收props
const itemsData = ref([]);
const selectedAssistantTopicsData = ref([]);

// 初始化方法
const init = () => {
  itemsData.value = cloneDeep(props.items);
  selectedAssistantTopicsData.value = cloneDeep(props.selectedAssistantTopics);
};

// 组件初始化时调用init
init();

// 本地数据
const localItems = ref([...itemsData.value]);

// 标记是否正在处理内部更新，防止循环
let isInternalUpdate = false;

// 监听props.items变化，更新itemsData
watch(
  () => props.items,
  (newItems) => {
    itemsData.value = cloneDeep(newItems);
    // 只有在非内部更新时才同步外部数据
    if (!isInternalUpdate) {
      localItems.value = [...itemsData.value];
    }
  },
  { deep: true }
);

// 监听props.selectedAssistantTopics变化，更新selectedAssistantTopicsData
watch(
  () => props.selectedAssistantTopics,
  (newTopics) => {
    selectedAssistantTopicsData.value = cloneDeep(newTopics);
  },
  { deep: true }
);

// 处理点击事件
const handleClick = (item) => {
  emit("item-click", item);
};

// 处理拖拽结束事件
const onDragEnd = (event) => {
  emit("drag-end", {
    oldIndex: event.oldIndex,
    newIndex: event.newIndex,
    item: event.item,
  });
};

// 处理更新事件
const handleUpdate = (newItems) => {
  // 标记为内部更新
  isInternalUpdate = true;
  localItems.value = newItems;
  // 在下一个事件循环中重置标记
  nextTick(() => {
    isInternalUpdate = false;
  });
};

// 处理删除事件
const handleDelete = (item) => {
  emit("item-delete", item);
};

// 获取删除确认内容
const getDeleteConfirmContent = (element) => {
  return `确定要删除话题"${element.name}"吗？删除后将无法恢复，同时会删除该话题下的所有对话和消息。`;
};

// 获取项目样式类名
const getItemClass = (element) => {
  const baseClass = 'draggable-item';
  const isSelected = props.selectedId === element.id;
  return {
    [baseClass]: true,
    selected: isSelected
  };
};

// 格式化时间
const formatTime = (timeString) => {
  if (!timeString) return "";
  const date = new Date(timeString);
  return date.toLocaleString();
};
</script>

<style lang="less" scoped>
.draggable-list {
  width: 100%;
}

.draggable-item {
  display: flex;
  align-items: center;
  padding: 12px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background-color: var(--td-bg-color-container, #fff);
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);

  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
  }

  &.selected {
    background-color: #e0e0e0;
    color: var(--td-text-color-primary, #333);
    border-color: #e0e0e0;
  }

  &.ghost {
    opacity: 0.5;
    background-color: var(--td-bg-color-component, #f5f5f5);
  }

  &.chosen {
    background-color: var(--td-bg-color-component, #f5f5f5);
    border-color: var(--td-brand-color, #0052d9);
  }

  &.drag {
    opacity: 0.8;
    transform: rotate(5deg);
  }
}

.drag-handle {
  margin-right: 8px;
  cursor: grab;
  display: none; /* 隐藏拖拽手柄 */

  &:active {
    cursor: grabbing;
  }
}

.drag-icon {
  font-size: 16px;
  color: var(--td-text-color-secondary, #666);
  transition: color 0.2s ease;

  .draggable-item:hover & {
    color: var(--td-text-color-primary, #333);
  }
}

.item-content {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.item-emoji {
  font-size: 24px;
  margin-right: 12px;
  flex-shrink: 0;
}

.item-info {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.item-name {
  font-weight: 600;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-desc {
  font-size: 12px;
  color: var(--td-text-color-secondary, #666);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.delete-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  margin-left: 8px;
  border-radius: 4px;
  cursor: pointer;
  color: var(--td-text-color-secondary, #666);
  transition: all 0.2s ease;

  &:hover {
    background-color: var(--td-error-color, #e34d59);
    color: white;
  }
}

.topic-count {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 24px;
  margin-left: 8px;
  border-radius: 12px;
  background-color: var(--td-brand-color, #0052d9);
  color: white;
  font-size: 12px;
  font-weight: 500;
  padding: 0 6px;
  box-sizing: border-box;
}

// 滚动条样式
.draggable-list {
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
</style>
