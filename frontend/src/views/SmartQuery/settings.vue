<template>
  <t-layout class="settings-layout">
    <!-- 左侧导航 -->
    <t-aside class="settings-sidebar" :width="'200px'">
      <t-tabs v-model="activeTab" placement="left" class="settings-tabs">
        <t-tab-panel value="prompt" label="提示词设置">
          <template #label>
            <div class="tab-label">
              <IconEdit :size="16" />
              <span>提示词设置</span>
            </div>
          </template>
        </t-tab-panel>
        <t-tab-panel value="model" label="模型设置">
          <template #label>
            <div class="tab-label">
              <IconSettings :size="16" />
              <span>模型设置</span>
            </div>
          </template>
        </t-tab-panel>
        <t-tab-panel value="knowledge" label="知识库设置">
          <template #label>
            <div class="tab-label">
              <IconBook :size="16" />
              <span>知识库设置</span>
            </div>
          </template>
        </t-tab-panel>
        <t-tab-panel value="phrases" label="常用短语">
          <template #label>
            <div class="tab-label">
              <IconMessageSquare :size="16" />
              <span>常用短语</span>
            </div>
          </template>
        </t-tab-panel>
        <t-tab-panel value="memory" label="全局记忆">
          <template #label>
            <div class="tab-label">
              <IconDatabase :size="16" />
              <span>全局记忆</span>
            </div>
          </template>
        </t-tab-panel>
      </t-tabs>
    </t-aside>

    <!-- 右侧内容 -->
    <t-content class="settings-content">
      <!-- 提示词设置 -->
      <div v-if="activeTab === 'prompt'" class="prompt-settings">
        <div class="form-row">
          <label class="form-label">助手名称</label>
          <t-input v-model="formData.name" placeholder="请输入助手名称" :clearable="true" class="form-input" />
        </div>

        <div class="form-row">
          <label class="form-label">助手描述</label>
          <t-input v-model="formData.description" placeholder="请输入助手描述" :clearable="true" class="form-input" />
        </div>

        <div class="form-row">
          <label class="form-label">系统提示词</label>
          <div class="prompt-input-container">
            <t-button
              theme="primary"
              @click="handleGeneratePrompt"
              :loading="isGeneratingPrompt"
              :disabled="!formData.name.trim() || !formData.description.trim() || isGeneratingPrompt"
            >
              <div class="button-content">
                <IconSparkles v-if="!isGeneratingPrompt" :size="16" />
                <span>生成/优化提示词</span>
              </div>
            </t-button>
            <t-button v-if="isGeneratingPrompt" theme="default" @click="handleStopGeneration" class="stop-button">
              <div class="button-content">
                <IconX :size="16" />
                <span>停止</span>
              </div>
            </t-button>
          </div>
        </div>
        <t-textarea
          v-model="formData.prompt"
          placeholder="请输入系统提示词，用于定义助手的角色和行为"
          :autosize="{ minRows: 10, maxRows: 15 }"
          class="form-textarea"
        />
      </div>

      <!-- 模型设置（预留） -->
      <div v-else-if="activeTab === 'model'" class="model-settings">
        <div class="placeholder-content">
          <IconSettings :size="48" />
          <p>模型设置功能正在开发中...</p>
        </div>
      </div>

      <!-- 知识库设置（预留） -->
      <div v-else-if="activeTab === 'knowledge'" class="knowledge-settings">
        <div class="placeholder-content">
          <IconBook :size="48" />
          <p>知识库设置功能正在开发中...</p>
        </div>
      </div>

      <!-- 常用短语（预留） -->
      <div v-else-if="activeTab === 'phrases'" class="phrases-settings">
        <div class="placeholder-content">
          <IconMessageSquare :size="48" />
          <p>常用短语功能正在开发中...</p>
        </div>
      </div>

      <!-- 全局记忆（预留） -->
      <div v-else-if="activeTab === 'memory'" class="memory-settings">
        <div class="placeholder-content">
          <IconDatabase :size="48" />
          <p>全局记忆功能正在开发中...</p>
        </div>
      </div>
    </t-content>
  </t-layout>
</template>

<script setup>
import { ref, reactive, watch, computed, onMounted, onUnmounted } from "vue";
import { MessagePlugin } from "tdesign-vue-next";
import { UpdateAssistant, GenerateSystemPrompt, StreamChatCompletion, StopStreamChatCompletion } from "../../../wailsjs/go/main/App";
import { useStreamStore, StreamType, StreamStatus } from "../../store/modules/stream.js";

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  assistant: {
    type: Object,
    default: null,
  },
});

// Emits
const emit = defineEmits(["update:visible", "save"]);

// Store
const streamStore = useStreamStore();

// 响应式数据
const activeTab = ref("prompt");

// 表单数据
const formData = reactive({
  name: "",
  description: "",
  prompt: "",
});

// 计算属性
const isGeneratingPrompt = computed(() => {
  return streamStore.hasActiveStreamByType(StreamType.SYSTEM_PROMPT);
});

// 获取当前系统提示词生成的流式输出
const currentSystemPromptStream = computed(() => {
  const streams = streamStore.getActiveStreamsByType(StreamType.SYSTEM_PROMPT);
  return streams.length > 0 ? streams[0] : null;
});

// 初始化表单数据
const initFormData = () => {
  if (props.assistant) {
    formData.name = props.assistant.name || "";
    formData.description = props.assistant.description || "";
    formData.prompt = props.assistant.prompt || "";
  } else {
    formData.name = "";
    formData.description = "";
    formData.prompt = "";
  }
};

// 监听助手数据变化
watch(
  () => props.assistant,
  () => {
    initFormData();
  },
  { immediate: true, deep: true }
);

// 监听对话框显示状态
watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      initFormData();
      activeTab.value = "prompt";
    }
  }
);

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

// 监听系统提示词流式输出状态变化
watch(
  () => streamStore.activeStreamsList,
  (streams) => {
    for (const stream of streams) {
      if (stream.type === StreamType.SYSTEM_PROMPT) {
        if (stream.status === StreamStatus.ERROR) {
          MessagePlugin.error("生成提示词失败：" + (stream.error || "未知错误"));
        }
      }
    }
  },
  { deep: true }
);

// 处理生成提示词
const handleGeneratePrompt = async () => {
  try {
    // 调用后端API生成系统提示词（后端已通过 Dify API 流式获取结果）
    const generatedPrompt = await GenerateSystemPrompt(formData.name.trim(), formData.description.trim(), formData.prompt.trim());
    
    // 直接将生成的提示词设置到表单
    if (generatedPrompt && generatedPrompt.trim()) {
      formData.prompt = generatedPrompt.trim();
      MessagePlugin.success('提示词生成完成');
    } else {
      MessagePlugin.warning('未获取到生成的提示词');
    }
  } catch (error) {
    console.error("生成提示词失败：", error);
    const errorMessage = error?.message || error?.toString() || '未知错误';
    MessagePlugin.error(`生成提示词失败: ${errorMessage}`);
  }
};

// 处理停止生成
const handleStopGeneration = async () => {
  const systemPromptStreams = streamStore.getActiveStreamsByType(StreamType.SYSTEM_PROMPT);
  for (const stream of systemPromptStreams) {
    try {
      await StopStreamChatCompletion(stream.id);
      streamStore.stopStream(stream.id);
    } catch (error) {
      console.error("停止生成失败:", error);
    }
  }
};

// 处理保存
const handleSave = async () => {
  try {
    // 验证表单
    if (!formData.name.trim()) {
      MessagePlugin.error("请输入助手名称");
      return;
    }

    if (!formData.prompt.trim()) {
      MessagePlugin.error("请输入系统提示词");
      return;
    }

    // 构建保存数据
    const saveData = {
      ...props.assistant,
      name: formData.name.trim(),
      description: formData.description.trim(),
      prompt: formData.prompt.trim(),
    };

    // 调用后端API更新助手
    await UpdateAssistant(saveData);

    // 发送保存事件给父组件
    emit("save", saveData);

    MessagePlugin.success("助手设置保存成功");
  } catch (error) {
    MessagePlugin.error("保存失败：" + (error.message || error));
  }
};

// 暴露方法给父组件
defineExpose({
  initFormData,
  handleSave,
});
</script>

<style lang="less" scoped>
.assistant-settings-dialog {
  :deep(.t-dialog) {
    border-radius: 12px;
    overflow: hidden;
  }

  :deep(.t-dialog__header) {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 20px 24px;
    border-bottom: none;
  }

  :deep(.t-dialog__header__title) {
    font-size: 18px;
    font-weight: 600;
  }

  :deep(.t-dialog__body) {
    padding: 0;
    max-height: 70vh;
    overflow: hidden;
  }
}

.settings-layout {
  height: 50vh;
  min-height: 350px;
  max-height: 600px;
  background-color: var(--td-bg-color-container, #fff);
}

.settings-sidebar {
  background-color: var(--td-bg-color-container-hover, #f8f9fa);
  border-right: 1px solid var(--td-border-level-1-color, #e7e7e7);
}

.settings-tabs {
  height: 100%;

  :deep(.t-tabs__nav-wrap) {
    width: 200px;
  }

  :deep(.t-tabs__nav) {
    background-color: transparent;
    width: 100%;
    height: 100%;
  }

  :deep(.t-tabs__nav-item) {
    margin: 0;
    border-radius: 0;
    border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
    width: 100%;

    &.t-is-active {
      background-color: var(--td-brand-color-light, #e6f7ff);
      color: var(--td-brand-color, #0052d9);
    }
  }

  :deep(.t-tabs__content) {
    display: none;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 500;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
}

.settings-content {
  padding: 20px;
  overflow-y: auto;

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

.settings-section {
  margin-bottom: 20px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary, #333);
  margin-bottom: 12px;
  padding-bottom: 6px;
  border-bottom: 2px solid var(--td-brand-color, #0052d9);
}

.form-item {
  margin-bottom: 16px;
}

.form-row {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  gap: 16px;
}

.form-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--td-text-color-primary, #333);
  margin-bottom: 8px;
  min-width: 80px;
  flex-shrink: 0;
}

.form-row .form-label {
  margin-bottom: 0;
  margin-right: 0;
}

.form-row .form-input {
  width: 1;
  flex: 1;
}

.form-input {
  width: 100%;

  :deep(.t-input) {
    border-radius: 8px;
    border: 1px solid var(--td-border-level-1-color, #e7e7e7);
    transition: all 0.3s ease;

    &:hover {
      border-color: var(--td-brand-color, #0052d9);
    }

    &:focus-within {
      border-color: var(--td-brand-color, #0052d9);
      box-shadow: 0 0 0 2px rgba(0, 82, 217, 0.1);
    }
  }
}

.prompt-input-container {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;

  .stop-button {
    background-color: var(--td-error-color, #f5222d);
    border-color: var(--td-error-color, #f5222d);
    color: white;

    &:hover {
      background-color: var(--td-error-color-hover, #ff4d4f);
      border-color: var(--td-error-color-hover, #ff4d4f);
    }
  }
}

.form-textarea {
  width: 100%;

  :deep(.t-textarea) {
    border-radius: 8px;
    border: 1px solid var(--td-border-level-1-color, #e7e7e7);
    transition: all 0.3s ease;
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
      sans-serif;

    &:hover {
      border-color: var(--td-brand-color, #0052d9);
    }

    &:focus-within {
      border-color: var(--td-brand-color, #0052d9);
      box-shadow: 0 0 0 2px rgba(0, 82, 217, 0.1);
    }
  }
}

.settings-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid var(--td-border-level-1-color, #e7e7e7);

  .t-button {
    border-radius: 8px;
    padding: 8px 24px;
    font-weight: 500;
    transition: all 0.3s ease;

    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }
  }
}

.placeholder-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--td-text-color-secondary, #666);

  svg {
    margin-bottom: 16px;
    opacity: 0.5;
  }

  p {
    font-size: 16px;
    margin: 0;
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .settings-layout {
    flex-direction: column;
    height: auto;
    max-height: 65vh;
  }

  .settings-content {
    max-height: 100%;
  }
}

@media (max-height: 800px) {
  .settings-layout {
    height: 55vh;
    min-height: 300px;
  }

  .settings-content {
    max-height: 100%;
  }
}

@media (max-height: 600px) {
  .settings-layout {
    height: 50vh;
    min-height: 250px;
  }

  .settings-content {
    max-height: 100%;
  }
}
</style>
