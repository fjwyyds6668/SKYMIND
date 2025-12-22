<template>
  <div class="smart-query-container">
    <!-- 左侧Tab区域 -->
    <div class="sidebar">
      <t-tabs v-model="activeTab" class="sidebar-tabs">
        <!-- 助手列表 -->
        <t-tab-panel value="assistants">
          <template #label> <IconUser class="tabs-icon-margin" :size="20" /> 助手 </template>
          <div class="assistants-panel">
            <div v-if="assistants.length === 0" class="empty-state">
              <t-empty description="暂无助手，请先创建一个助手" />
            </div>
            <div v-else class="assistants-list">
              <DraggableList
                :items="assistants"
                :selected-id="selectedAssistant?.id"
                :show-delete-button="false"
                :show-topic-count="true"
                :selected-assistant-topics="topics"
                @item-click="(value) => selectAssistant(value, false)"
                @drag-end="handleAssistantDragEnd"
              >
                <!-- 自定义助手列表项内容 -->
                <template #item="{ element, isSelected }">
                  <div class="assistant-item-content">
                    <div class="assistant-avatar-container">
                      <div class="assistant-avatar">{{ element.emoji }}</div>
                      <!-- 流式输出指示器光圈 -->
                      <div v-if="hasAssistantActiveTopics(element.id)" class="stream-indicator assistant-stream-ring"></div>
                    </div>
                    <div class="assistant-info">
                      <div class="assistant-name">{{ element.name }}</div>
                      <div class="assistant-desc">{{ element.description }}</div>
                    </div>
                  </div>
                </template>

                <!-- 自定义助手操作区域 -->
                <template #item-actions="{ element, isSelected }">
                  <div v-if="isSelected" class="topic-count-white">
                    {{ topics.length }}
                  </div>
                </template>
              </DraggableList>
            </div>
            <div class="add-assistant-section">
              <t-button variant="text" @click="showAssistantDialog">
                <div class="button-content">
                  <IconPlus class="add-icon-margin" :size="16" />
                  <span>新增助手</span>
                </div>
              </t-button>
            </div>
          </div>
        </t-tab-panel>
        <!-- 话题列表 -->
        <t-tab-panel value="topics">
          <template #label> <IconMessageCircle class="tabs-icon-margin" :size="20" /> 话题 </template>
          <div class="topics-panel">
            <div v-if="!selectedAssistant" class="empty-state">
              <t-empty description="请先选择一个助手" />
            </div>
            <div v-else-if="topics.length === 0" class="empty-state">
              <t-empty description="该助手暂无话题" />
            </div>
            <div v-else class="topics-list">
              <DraggableList 
                :items="topics" 
                :selected-id="selectedTopic?.id" 
                :show-delete-button="true"
                @item-click="selectTopic" 
                @drag-end="handleTopicDragEnd"
                @item-delete="handleTopicDelete"
              >
                <!-- 自定义话题列表项内容 -->
                <template #item="{ element, isSelected }">
                  <div class="topic-item-content">
                    <!-- 流式输出指示器 -->
                    <div v-if="hasActiveChatStream(element.id)" class="stream-indicator topic-stream-dot"></div>
                    <div class="topic-info">
                      <div class="topic-name">{{ element.name }}</div>
                      <div class="topic-time">{{ formatTopicTime(element.created_at) }}</div>
                    </div>
                  </div>
                </template>

                <!-- 自定义话题操作区域 -->
                <template #item-actions="{ element, isSelected }">
                  <div v-if="isSelected" class="delete-button">
                    <t-popconfirm
                      :content="getDeleteConfirmContent(element, 'topic')"
                      placement="right"
                      :overlay-style="{ width: '400px' }"
                      @confirm="handleTopicDelete(element)"
                    >
                      <IconX :size="16" />
                    </t-popconfirm>
                  </div>
                </template>
              </DraggableList>
            </div>
            <div class="add-topic-section">
              <t-button variant="text" @click="createTopic">
                <div class="button-content">
                  <IconPlus class="add-icon-margin" :size="16" />
                  <span>新增话题</span>
                </div>
              </t-button>
            </div>
          </div>
        </t-tab-panel>
        <!-- 设置面板 -->
        <t-tab-panel value="settings">
          <template #label> <IconSettings class="tabs-icon-margin" :size="20" /> 设置 </template>
          <div class="settings-panel">
            <div v-if="!selectedAssistant" class="empty-state">
              <t-empty description="请先选择一个助手" />
            </div>
            <div v-else class="settings-form">
              <div class="setting-item">
                <div class="setting-label">
                  <span>模型温度</span>
                  <t-tooltip>
                    <template #content> 控制回复的随机性，值越高回复越随机 </template>
                    <IconHelpCircle class="help-icon" :size="14" />
                  </t-tooltip>
                </div>
                <t-slider
                  v-model="assistantSettings.temperature"
                  :min="0"
                  :max="2"
                  :step="0.1"
                  :marks="{ 0: '0', 1: '1', 2: '2' }"
                  show-value
                  class="setting-slider"
                />
              </div>

              <div class="setting-item">
                <div class="setting-label">
                  <span>上下文数</span>
                  <t-tooltip>
                    <template #content> 保留的对话上下文数量 </template>
                    <IconHelpCircle class="help-icon" :size="14" />
                  </t-tooltip>
                </div>
                <t-slider
                  v-model="assistantSettings.contextCount"
                  :min="0"
                  :max="20"
                  :step="1"
                  :marks="{ 0: '0', 5: '5', 10: '10', 15: '15', 20: '20' }"
                  show-value
                  class="setting-slider"
                />
              </div>
            </div>
            <div class="save-settings-section">
              <div class="settings-buttons">
                <t-button variant="text" @click="saveSettings">
                  <div class="button-content">
                    <IconSave class="add-icon-margin" :size="16" />
                    <span>保存设置</span>
                  </div>
                </t-button>
                <t-popconfirm
                  v-if="selectedAssistant && !selectedAssistant.is_default"
                  :content="getDeleteConfirmContent(selectedAssistant, 'assistant')"
                  placement="top"
                  :confirm-btn="{ content: '确定', theme: 'danger' }"
                  :cancel-btn="{ content: '取消' }"
                  @confirm="handleDeleteAssistant"
                >
                  <t-button variant="text" theme="danger">
                    <div class="button-content delete-button">
                      <IconTrash2 class="add-icon-margin" :size="16" />
                      <span>删除助手</span>
                    </div>
                  </t-button>
                </t-popconfirm>
              </div>
            </div>
          </div>
        </t-tab-panel>
      </t-tabs>
    </div>

    <!-- 右侧聊天区域 -->
    <AIChat
      ref="chatRef"
      :topic-id="selectedTopic?.id"
      :selected-assistant="selectedAssistant"
      :selected-topic="selectedTopic"
      :assistant-settings="assistantSettings"
      @conversation-created="handleConversationCreated"
      @assistant-updated="handleAssistantUpdated"
    />

    <!-- 新增助手对话框 -->
    <t-dialog v-model:visible="assistantDialogVisible" header="选择助手类型" width="60%" :footer="false" :close-on-overlay-click="true">
      <AssistantsList @select-assistant="createAssistant" />
    </t-dialog>

    <!-- 确认对话框 -->
    <t-dialog
      v-model:visible="confirmDialogVisible"
      header="确认操作"
      width="400px"
      :confirm-btn="{ content: '确认', theme: 'primary' }"
      :cancel-btn="{ content: '取消' }"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    >
      <div class="confirm-content">
        <p>当前正在生成AI回复，确定要终止输出并跳转吗？</p>
        <p class="confirm-tip">已生成的内容将被保存</p>
      </div>
    </t-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed } from "vue";
import { MessagePlugin } from "tdesign-vue-next";
import {
  GetAssistants,
  CreateAssistant,
  UpdateAssistant,
  DeleteAssistant,
  GetAssistantByID,
  GetTopics,
  CreateTopic,
  DeleteTopic,
  UpdateAssistantsSortOrder,
  UpdateTopicsSortOrder,
} from "../../../wailsjs/go/main/App";
import AssistantsList from "./assistants.vue";
import AIChat from "./chat.vue";
import DraggableList from "../../components/draggableList.vue";
import { useStreamStore, StreamType, StreamStatus } from "../../store/modules/stream.js";
import { getDeleteConfirmContent, formatTopicTime } from "./utils.js";

// 响应式数据
const loading = ref(false);

// 对话框相关
const assistantDialogVisible = ref(false);
const confirmDialogVisible = ref(false);
const pendingAction = ref(null);
const chatRef = ref(null);

// Tab相关
const activeTab = ref("assistants");
const assistants = ref([]);
const selectedAssistant = ref(null);
const topics = ref([]);
const selectedTopic = ref(null);

// 助手设置
const assistantSettings = reactive({
  temperature: 1.0,
  contextCount: 5,
});

// 流式输出 store
const streamStore = useStreamStore();

// 获取所有活跃CHAT流式输出的助手ID列表
const activeChatAssistantIds = computed(() => {
  return streamStore.getActiveChatAssistantIds();
});

// 获取所有活跃CHAT流式输出的话题ID列表
const activeChatTopicIds = computed(() => {
  return streamStore.getActiveChatTopicIds();
});

// 计算属性：检查指定话题是否有活跃的CHAT流式输出
const hasActiveChatStream = (topicId) => {
  return activeChatTopicIds.value.includes(topicId);
};

// 计算属性：检查指定助手是否有活跃的话题（即该助手下有话题正在进行CHAT流式输出）
const hasAssistantActiveTopics = (assistantId) => {
  return activeChatAssistantIds.value.includes(assistantId);
};

// 初始化数据
const initializeData = async () => {
  try {
    // 加载助手列表
    const assistantsData = await GetAssistants();
    assistants.value = assistantsData || [];

    if (assistants.value.length === 0) {
      // 如果没有助手，创建默认助手
      await createDefaultAssistant();
      // 重新加载助手列表
      const newAssistantsData = await GetAssistants();
      assistants.value = newAssistantsData || [];
    }

    // 选择第一个助手
    if (assistants.value.length > 0) {
      await selectAssistant(assistants.value[0], true);
    }
  } catch (error) {
    MessagePlugin.error("初始化数据失败: " + error);
  }
};

// 创建默认助手
const createDefaultAssistant = async () => {
  try {
    const defaultAssistant = {
      name: "默认助手",
      description: "这是一个通用的AI助手，可以帮助您回答问题和完成任务。",
      prompt: "你是一个有用的AI助手。",
      emoji: "🤖",
    };

    await CreateAssistant(defaultAssistant);
    MessagePlugin.success("默认助手创建成功");
  } catch (error) {
    MessagePlugin.error("创建默认助手失败: " + error);
  }
};

// 选择助手
const selectAssistant = async (assistant, isInit) => {
  selectedAssistant.value = assistant;

  // 加载助手设置
  try {
    const assistantData = await GetAssistantByID(assistant.id);
    if (assistantData && assistantData.settings) {
      const settings = JSON.parse(assistantData.settings);
      Object.assign(assistantSettings, settings);
    }
  } catch (error) {
    MessagePlugin.error("加载助手设置失败:", error);
  }
  // 加载助手的话题
  await loadTopics(assistant.id);
  activeTab.value = isInit ? "assistants" : "topics";
};

// 加载话题列表
const loadTopics = async (assistantId) => {
  try {
    const topicsData = await GetTopics(assistantId);
    topics.value = topicsData || [];

    // 选择第一个话题
    if (topics.value.length > 0) {
      selectedTopic.value = topics.value[0];
    } else {
      // 如果没有话题，清空聊天记录
      selectedTopic.value = null;
    }
  } catch (error) {
    MessagePlugin.error("加载话题列表失败: " + error);
    topics.value = [];
    selectedTopic.value = null;
  }
};

// 选择话题
const selectTopic = async (topic) => {
  selectedTopic.value = topic;
};

// 处理对话创建事件
const handleConversationCreated = (conversation) => {
  // 可以在这里处理对话创建后的逻辑，比如更新UI或发送通知
};

// 处理助手更新事件
const handleAssistantUpdated = (updatedAssistant) => {
  // 更新本地助手列表中的对应项
  const index = assistants.value.findIndex(a => a.id === updatedAssistant.id);
  if (index !== -1) {
    assistants.value[index] = { ...assistants.value[index], ...updatedAssistant };
    
    // 如果更新的是当前选中的助手，也要更新 selectedAssistant
    if (selectedAssistant.value && selectedAssistant.value.id === updatedAssistant.id) {
      selectedAssistant.value = { ...selectedAssistant.value, ...updatedAssistant };
    }
  }
};

// 显示助手列表，用于创建助手
const showAssistantDialog = () => {
  assistantDialogVisible.value = true;
};

// 选择预设助手，创建助手
const createAssistant = async (assistantTemplate) => {
  try {
    const newAssistant = {
      name: assistantTemplate.name,
      description: assistantTemplate.description,
      prompt: assistantTemplate.prompt,
      emoji: assistantTemplate.emoji,
    };

    const createdAssistant = await CreateAssistant(newAssistant);
    assistants.value.push(createdAssistant);
    assistantDialogVisible.value = false;
    selectAssistant(createdAssistant, true);
  } catch (error) {
    MessagePlugin.error("助手创建失败: " + error);
  }
};

// 创建话题
const createTopic = async () => {
  if (!selectedAssistant.value) {
    MessagePlugin.error("请先选择一个助手");
    return;
  }

  try {
    // 创建新话题
    const newTopic = {
      assistant_id: selectedAssistant.value.id,
      name: "默认话题",
      is_name_manually_edited: false,
    };

    const createdTopic = await CreateTopic(newTopic);
    // 重新加载话题列表
    await loadTopics(selectedAssistant.value.id);
    selectTopic(createdTopic);
  } catch (error) {
    MessagePlugin.error("创建话题失败: " + error);
  }
};

// 保存设置
const saveSettings = async () => {
  if (!selectedAssistant.value) return;

  try {
    const settingsString = JSON.stringify(assistantSettings);
    const updatedAssistant = {
      id: selectedAssistant.value.id,
      name: selectedAssistant.value.name,
      description: selectedAssistant.value.description,
      prompt: selectedAssistant.value.prompt,
      emoji: selectedAssistant.value.emoji,
      settings: settingsString,
    };

    await UpdateAssistant(updatedAssistant);

    // 更新本地 selectedAssistant 的 settings 字段，确保数据一致性
    selectedAssistant.value.settings = settingsString;

    MessagePlugin.success("设置保存成功");
  } catch (error) {
    MessagePlugin.error("设置保存失败: " + error);
  }
};

// 处理助手拖拽结束
const handleAssistantDragEnd = async (event) => {
  const { oldIndex, newIndex } = event;

  // 没有移动，直接返回
  if (oldIndex === newIndex) return;

  const currentList = [...assistants.value];
  const changedItems = [];

  // 确定受影响的范围
  const minIndex = Math.min(oldIndex, newIndex);
  const maxIndex = Math.max(oldIndex, newIndex);

  // 获取被移动的项目ID
  const movedItemId = currentList[oldIndex].id;

  // 只更新受影响范围内的项目
  for (let i = minIndex; i <= maxIndex; i++) {
    const currentItem = currentList[i];
    let newSortOrder;

    if (currentItem.id === movedItemId) {
      // 被移动的项目：设置为新位置
      newSortOrder = newIndex;
    } else if (newIndex > oldIndex) {
      // 向后拖拽：中间的项目前移一位
      if (i > oldIndex && i <= newIndex) {
        newSortOrder = i - 1;
      } else {
        newSortOrder = i; // 保持不变
      }
    } else {
      // 向前拖拽：中间的项目后移一位
      if (i >= newIndex && i < oldIndex) {
        newSortOrder = i + 1;
      } else {
        newSortOrder = i; // 保持不变
      }
    }

    changedItems.push({
      id: currentItem.id,
      sort_order: newSortOrder,
    });
  }

  try {
    // 调用后端更新
    await UpdateAssistantsSortOrder(changedItems);

    // 更新本地数据
    const [movedItem] = currentList.splice(oldIndex, 1);
    currentList.splice(newIndex, 0, movedItem);
    assistants.value = currentList;
  } catch (error) {
    MessagePlugin.error("更新助手排序失败: " + error);
  }
};

// 处理话题拖拽结束
const handleTopicDragEnd = async (event) => {
  const { oldIndex, newIndex } = event;

  // 没有移动，直接返回
  if (oldIndex === newIndex) return;

  const currentList = [...topics.value];
  const changedItems = [];

  // 确定受影响的范围
  const minIndex = Math.min(oldIndex, newIndex);
  const maxIndex = Math.max(oldIndex, newIndex);

  // 获取被移动的项目ID
  const movedItemId = currentList[oldIndex].id;

  // 只更新受影响范围内的项目
  for (let i = minIndex; i <= maxIndex; i++) {
    const currentItem = currentList[i];
    let newSortOrder;

    if (currentItem.id === movedItemId) {
      // 被移动的项目：设置为新位置
      newSortOrder = newIndex;
    } else if (newIndex > oldIndex) {
      // 向后拖拽：中间的项目前移一位
      if (i > oldIndex && i <= newIndex) {
        newSortOrder = i - 1;
      } else {
        newSortOrder = i; // 保持不变
      }
    } else {
      // 向前拖拽：中间的项目后移一位
      if (i >= newIndex && i < oldIndex) {
        newSortOrder = i + 1;
      } else {
        newSortOrder = i; // 保持不变
      }
    }

    changedItems.push({
      id: currentItem.id,
      sort_order: newSortOrder,
    });
  }

  try {
    // 调用后端更新
    await UpdateTopicsSortOrder(changedItems);

    // 更新本地数据
    const [movedItem] = currentList.splice(oldIndex, 1);
    currentList.splice(newIndex, 0, movedItem);
    topics.value = currentList;
  } catch (error) {
    MessagePlugin.error("更新话题排序失败: " + error);
  }
};

// 确认对话框处理
const handleConfirm = async () => {
  if (!pendingAction.value) return;

  try {
    // 执行待处理的操作
    if (pendingAction.value.type === "assistant") {
      await selectAssistant(pendingAction.value.data, pendingAction.value.isInit);
    } else if (pendingAction.value.type === "topic") {
      selectedTopic.value = pendingAction.value.data;
    }
  } catch (error) {
    MessagePlugin.error("处理确认操作失败:");
  } finally {
    // 清理状态
    confirmDialogVisible.value = false;
    pendingAction.value = null;
  }
};

// 取消对话框处理
const handleCancel = () => {
  confirmDialogVisible.value = false;
  pendingAction.value = null;
};

// 处理话题删除
const handleTopicDelete = async (topic) => {
  try {
    // 判断是否是最后一个话题
    const isLastTopic = topics.value.length === 1;
    const isSelectedTopic = selectedTopic.value?.id === topic.id;
    
    // 找到当前删除话题在列表中的索引
    const currentIndex = topics.value.findIndex(t => t.id === topic.id);
    
    // 如果是最后一个话题，则不删除话题本身，只删除对话和消息
    await DeleteTopic(topic.id, !isLastTopic);
    
    if (isLastTopic) {
      MessagePlugin.success("话题内容已清空");
    } else {
      MessagePlugin.success("话题删除成功");
    }
    
    // 重新加载话题列表
    await loadTopics(selectedAssistant.value.id);
    
    // 处理删除后的选中逻辑
    if (isSelectedTopic) {
      if (isLastTopic) {
        // 如果是最后一个话题且是当前选中的话题，保持选中状态
        if (topics.value.length > 0) {
          selectedTopic.value = topics.value[0];
          // 通知聊天组件刷新数据（清空内容）
          if (chatRef.value) {
            chatRef.value.refreshChat();
          }
        }
      } else {
        // 如果不是最后一个话题且删除的是当前选中的话题，自动选择下一个话题
        if (topics.value.length > 0) {
          // 优先选择下一个话题，如果下一个不存在则选择上一个话题
          if (currentIndex < topics.value.length) {
            // 下一个话题存在
            selectedTopic.value = topics.value[currentIndex];
          } else {
            // 下一个话题不存在，选择上一个话题
            selectedTopic.value = topics.value[currentIndex - 1];
          }
        } else {
          // 没有话题了，清空选中状态
          selectedTopic.value = null;
        }
      }
    }
  } catch (error) {
    MessagePlugin.error("删除话题失败: " + error);
  }
};

// 处理助手删除
const handleDeleteAssistant = async () => {
  if (!selectedAssistant.value) return;
  
  try {
    // 判断是否是最后一个助手
    const isLastAssistant = assistants.value.length === 1;
    const isSelectedAssistant = selectedAssistant.value;
    
    // 找到当前删除助手在列表中的索引
    const currentIndex = assistants.value.findIndex(a => a.id === selectedAssistant.value.id);
    
    // 删除助手（会级联删除该助手下的所有话题、对话和消息）
    await DeleteAssistant(selectedAssistant.value.id);
    
    if (isLastAssistant) {
      MessagePlugin.success("助手已删除，正在创建默认助手...");
    } else {
      MessagePlugin.success("助手删除成功");
    }
    
    // 重新加载助手列表
    const assistantsData = await GetAssistants();
    assistants.value = assistantsData || [];
    
    // 处理删除后的选中逻辑
    if (isLastAssistant) {
      // 如果是最后一个助手，创建默认助手和默认话题
      await createDefaultAssistant();
      // 重新加载助手列表
      const newAssistantsData = await GetAssistants();
      assistants.value = newAssistantsData || [];
      
      // 选择新创建的默认助手
      if (assistants.value.length > 0) {
        await selectAssistant(assistants.value[0], true);
      }
    } else {
      // 如果不是最后一个助手且删除的是当前选中的助手，自动选择下一个助手
      if (assistants.value.length > 0) {
        // 优先选择下一个助手，如果下一个不存在则选择上一个助手
        if (currentIndex < assistants.value.length) {
          // 下一个助手存在
          await selectAssistant(assistants.value[currentIndex], false);
        } else {
          // 下一个助手不存在，选择上一个助手
          await selectAssistant(assistants.value[currentIndex - 1], false);
        }
      } else {
        // 没有助手了，清空选中状态
        selectedAssistant.value = null;
        selectedTopic.value = null;
        topics.value = [];
      }
    }
    
    // 删除助手后切换到助手Tab
    activeTab.value = "assistants";
  } catch (error) {
    MessagePlugin.error("删除助手失败: " + error);
  }
};

// 监听流式输出状态，更新话题标题
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
      const topicIndex = topics.value.findIndex(topic => topic.id === stream.metadata.topicId);
      if (topicIndex !== -1 && stream.content) {
        topics.value[topicIndex].name = stream.content;
      }
    });
  },
  { deep: true }
);

// 组件挂载时初始化数据
onMounted(async () => {
  await initializeData();
});
</script>

<style lang="less" scoped>
.smart-query-container {
  display: flex;
  height: 100%;
  background-color: var(--td-bg-color-page, #f5f5f5);
}

.sidebar {
  width: 253px;
  min-width: 253px;
  border-right: 1px solid var(--td-border-level-1-color, #e7e7e7);
  background-color: var(--td-bg-color-container, #fff);
  display: flex;
  flex-direction: column;
}

.sidebar-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;

  :deep(.t-tabs__nav) {
    flex-shrink: 0;
  }

  :deep(.t-tabs__content) {
    flex: 1;
    padding: 0;
    display: flex;
    flex-direction: column;
  }

  :deep(.t-tab-panel) {
    height: 100%;
    padding: 16px;
    display: flex;
    flex-direction: column;
  }
}

.tabs-icon-margin {
  margin-right: 4px;
}

.assistants-panel,
.topics-panel,
.settings-panel {
  height: 96%;
  display: flex;
  flex-direction: column;
  overflow: hidden; // 防止panel本身出现滚动条
}

.assistants-header,
.topics-header,
.settings-header {
  margin-bottom: 16px;

  h3 {
    margin: 0 0 8px 0;
    font-size: 16px;
    font-weight: 600;
  }
}

.assistants-list {
  height: calc(100% - 60px); // 减去add-assistant-section的高度
  overflow-y: auto;

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

.assistant-item {
  display: flex;
  align-items: center;
  padding: 12px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
  }

  &.selected {
    background-color: #e0e0e0; // 比悬停色更深的灰色
    color: var(--td-text-color-primary, #333);
  }
}

.add-assistant-section,
.add-topic-section,
.save-settings-section {
  padding: 12px 0;
  border-top: 1px solid var(--td-border-level-1-color, #e7e7e7);
  margin-top: 8px;
}

.add-icon-margin {
  margin-right: 4px;
  vertical-align: middle;
}

.add-assistant-section .t-button,
.add-topic-section .t-button,
.save-settings-section .t-button {
  display: flex;
  align-items: center;
  justify-content: center;
}

.assistant-avatar {
  font-size: 24px;
  margin-right: 12px;
}

.assistant-info {
  flex: 1;

  .assistant-name {
    font-weight: 600;
    margin-bottom: 4px;
  }

  .assistant-desc {
    font-size: 12px;
    color: var(--td-text-color-secondary, #666);
  }
}

.topics-list {
  height: calc(100% - 60px); // 减去add-topic-section的高度
  overflow-y: auto;

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

.topic-item {
  padding: 12px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);

  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
  }

  &.selected {
    background-color: #e0e0e0; // 与助手选中相同的深灰色
    color: var(--td-text-color-primary, #333);
    border-color: #e0e0e0;
  }
}

.topic-name {
  font-weight: 600;
  margin-bottom: 4px;
}

.topic-time {
  font-size: 12px;
  color: var(--td-text-color-secondary, #666);
}

.settings-form {
  flex: 1;
  overflow-y: auto;
}

.setting-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  width: 100%;
}

.setting-slider {
  width: 100%;
  margin-bottom: 16px;
  padding: 0 12px; // 为滑块左右添加内边距，确保滑块完整显示
  box-sizing: border-box; // 确保padding不会增加总宽度
}

.help-icon {
  font-size: 14px;
  color: var(--td-text-color-secondary, #666);
  cursor: pointer;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
  border-radius: 50%;
  padding: 2px;
  transition: all 0.2s ease;

  &:hover {
    color: var(--td-brand-color, #0052d9);
    border-color: var(--td-brand-color, #0052d9);
  }
}

.save-icon-margin {
  margin-right: 4px;
}

.setting-item {
  margin-bottom: 20px;
}

.save-button-container {
  text-align: center;
  margin-top: 16px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-content {
  text-align: center;
  padding: 20px 0;

  p {
    margin: 0 0 8px 0;
    font-size: 14px;
    color: var(--td-text-color-primary, #333);
  }

  .confirm-tip {
    font-size: 12px;
    color: var(--td-text-color-secondary, #666);
    font-style: italic;
  }
}

.settings-buttons {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.delete-button {
  color: var(--td-error-color, #e34d59);
  
  &:hover {
    color: var(--td-error-color-hover, #c53030);
  }
}

// 自定义助手列表项样式
.assistant-item-content {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;

  .assistant-info {
    flex: 1;
    min-width: 0;
    overflow: hidden;

    .assistant-name {
      font-weight: 600;
      margin-bottom: 4px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .assistant-desc {
      font-size: 12px;
      color: var(--td-text-color-secondary, #666);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}

// 自定义话题列表项样式
.topic-item-content {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;

  .topic-info {
    flex: 1;
    min-width: 0;
    overflow: hidden;

    .topic-name {
      font-weight: 600;
      margin-bottom: 4px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .topic-time {
      font-size: 12px;
      color: var(--td-text-color-secondary, #666);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}

// 白色背景的话题数量样式
.topic-count-white {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 24px;
  margin-left: 8px;
  border-radius: 12px;
  background-color: #ffffff;
  color: #000000;
  font-size: 12px;
  font-weight: 500;
  padding: 0 6px;
  box-sizing: border-box;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
}

// 流式输出指示器样式
.stream-indicator {
  animation: pulse-blue 1.5s ease-in-out infinite;
}

// 话题流式输出小光点
.topic-stream-dot {
  width: 8px;
  height: 8px;
  background-color: #87CEEB; // 浅蓝色
  border-radius: 50%;
  margin-right: 8px;
  flex-shrink: 0;
  box-shadow: 0 0 4px rgba(135, 206, 235, 0.6);
}

// 助手流式输出光圈容器
.assistant-avatar-container {
  position: relative;
  display: inline-block;
  margin-right: 12px;
  flex-shrink: 0;
}

// 助手流式输出光圈
.assistant-stream-ring {
  position: absolute;
  top: -7px;
  left: -5px;
  width: 38px;
  height: 38px;
  border: 2px solid #87CEEB; // 浅蓝色
  border-radius: 50%;
  box-shadow: 0 0 8px rgba(135, 206, 235, 0.8);
  pointer-events: none; // 确保不影响点击事件
}

// 浅蓝色闪烁动画
@keyframes pulse-blue {
  0% {
    opacity: 0.3;
    transform: scale(0.95);
  }
  50% {
    opacity: 1;
    transform: scale(1.05);
  }
  100% {
    opacity: 0.3;
    transform: scale(0.95);
  }
}
</style>
