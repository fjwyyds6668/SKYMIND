<template>
  <div class="assistants-list">
    <div 
      v-for="assistant in assistantTemplates" 
      :key="assistant.id"
      class="assistant-template-item"
      @click="selectAssistant(assistant)"
    >
      <div class="assistant-name-section">
        <span class="assistant-emoji">{{ assistant.emoji }}</span>
        <span class="assistant-name">{{ assistant.name }}</span>
      </div>
      <div class="assistant-description">
        <div class="description-text" :class="{ 'has-overflow': isTextOverflow(assistant.description) }">
          {{ assistant.description }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

// 助手模板数据
const assistantTemplates = ref([
  {
    id: 'programming-assistant',
    name: '编程助手',
    description: '专业的编程助手，精通多种编程语言，可以帮助您解决代码问题、优化算法、调试程序等',
    prompt: '你是一个专业的编程助手，精通JavaScript、Python、Java、Go、C++等多种编程语言。你可以帮助用户解决编程问题、代码调试、性能优化、架构设计等问题。请提供清晰、准确、实用的代码示例和解决方案。',
    emoji: '💻'
  },
  {
    id: 'product-manager',
    name: '产品经理',
    description: '经验丰富的产品经理，擅长需求分析、产品设计、用户研究和项目管理',
    prompt: '你是一位经验丰富的产品经理，擅长产品规划、需求分析、用户研究、原型设计和项目管理。你可以帮助用户进行产品定位、功能设计、用户体验优化、市场分析等工作。请提供专业的产品建议和解决方案。',
    emoji: '📊'
  },
  {
    id: 'data-analyst',
    name: '数据分析师',
    description: '专业的数据分析师，精通数据挖掘、统计分析、数据可视化和机器学习',
    prompt: '你是一位专业的数据分析师，精通Python、R、SQL等数据分析工具，擅长数据挖掘、统计分析、机器学习和数据可视化。你可以帮助用户进行数据清洗、探索性分析、建模预测等工作。请提供详细的数据分析报告和建议。',
    emoji: '📈'
  },
  {
    id: 'translator',
    name: '翻译',
    description: '多语言翻译专家，精通中英日韩等多种语言，提供准确流畅的翻译服务',
    prompt: '你是一位专业的翻译专家，精通中文、英文、日文、韩文、法文、德文等多种语言。你可以提供准确、流畅、符合语言习惯的翻译服务，包括文档翻译、口语翻译、本地化等。请确保翻译的准确性和文化适应性。',
    emoji: '🌐'
  },
  {
    id: 'writing-assistant',
    name: '写作助手',
    description: '专业的写作助手，擅长各类文案创作、文章撰写、内容优化和语言润色',
    prompt: '你是一位专业的写作助手，擅长各种文体写作，包括商业文案、技术文档、创意写作、学术论文等。你可以帮助用户进行内容创作、语言润色、结构优化、风格调整等工作。请提供高质量的写作建议和修改意见。',
    emoji: '✍️'
  },
  {
    id: 'designer',
    name: '设计师',
    description: '创意设计师，精通UI/UX设计、平面设计、品牌设计和用户体验优化',
    prompt: '你是一位专业的设计师，精通UI/UX设计、平面设计、品牌设计、交互设计等。你可以帮助用户进行界面设计、用户体验优化、品牌策划、视觉设计等工作。请提供专业的设计建议和创意方案。',
    emoji: '🎨'
  },
  {
    id: 'marketing-specialist',
    name: '营销专家',
    description: '资深营销专家，擅长市场策略、品牌推广、内容营销和用户增长',
    prompt: '你是一位资深的营销专家，精通市场分析、品牌策略、数字营销、内容营销、用户增长等。你可以帮助用户制定营销策略、优化推广方案、分析市场趋势等工作。请提供实用的营销建议和执行方案。',
    emoji: '📢'
  },
  {
    id: 'financial-advisor',
    name: '财务顾问',
    description: '专业财务顾问，擅长财务分析、投资理财、风险管理和财务规划',
    prompt: '你是一位专业的财务顾问，精通财务分析、投资理财、风险管理、税务规划等。你可以帮助用户进行财务状况分析、投资组合优化、风险评估等工作。请提供专业的财务建议和规划方案。',
    emoji: '💰'
  },
  {
    id: 'legal-expert',
    name: '法律专家',
    description: '资深法律专家，擅长合同审查、法律咨询、合规管理和风险防控',
    prompt: '你是一位资深的法律专家，精通合同法、公司法、知识产权法等多个法律领域。你可以帮助用户进行合同审查、法律风险评估、合规建议等工作。请注意：我的建议仅供参考，重要法律事务请咨询专业律师。',
    emoji: '⚖️'
  },
  {
    id: 'education-consultant',
    name: '教育顾问',
    description: '专业教育顾问，擅长学习规划、课程设计、教育方法和职业发展指导',
    prompt: '你是一位专业的教育顾问，精通教育学理论、课程设计、学习方法和职业规划。你可以帮助用户制定学习计划、选择学习资源、优化学习方法等工作。请提供个性化的教育建议和发展规划。',
    emoji: '🎓'
  },
  {
    id: 'health-coach',
    name: '健康教练',
    description: '专业健康教练，擅长健康管理、运动指导、营养建议和生活方式优化',
    prompt: '你是一位专业的健康教练，精通健康管理、运动科学、营养学等知识。你可以帮助用户制定健康计划、提供运动指导、营养建议等工作。请注意：我的建议仅供参考，重要健康问题请咨询专业医生。',
    emoji: '🏃‍♂️'
  },
  {
    id: 'project-manager',
    name: '项目经理',
    description: '资深项目经理，擅长项目管理、团队协作、进度控制和风险管理',
    prompt: '你是一位资深的项目经理，精通敏捷开发、项目管理方法论、团队管理等。你可以帮助用户进行项目规划、进度管理、风险控制、团队协作等工作。请提供实用的项目管理建议和最佳实践。',
    emoji: '📋'
  }
]);

// 定义事件
const emit = defineEmits(['select-assistant']);

// 选择助手
const selectAssistant = (assistant) => {
  emit('select-assistant', assistant);
};

// 检查文字是否溢出
const isTextOverflow = (text) => {
  // 简单估算：如果字符数超过60个字符，可能就会溢出
  // 这是一个粗略的估算，实际项目中可能需要更精确的计算
  return text.length > 60;
};
</script>

<style lang="less" scoped>
.assistants-list {
  max-height: 50vh;
  overflow-y: auto;
}

.assistant-template-item {
  display: flex;
  align-items: flex-start;
  padding: 8px 16px 8px 16px;
  margin-bottom: 12px;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background-color: var(--td-bg-color-container, #fff);
  height: 56px; // 固定高度，约等于三行文字的高度

  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
    border-color: var(--td-brand-color, #0052d9);
    box-shadow: 0 2px 8px rgba(0, 82, 217, 0.1);
  }

  &:active {
    transform: translateY(1px);
  }
}

.assistant-name-section {
  display: flex;
  align-items: center;
  width: 25%;
  min-width: 120px;
  margin-right: 16px;
}

.assistant-emoji {
  font-size: 20px;
  margin-right: 8px;
}

.assistant-name {
  font-weight: 600;
  font-size: 14px;
  color: var(--td-text-color-primary, #333);
}

.assistant-description {
  flex: 1;
  font-size: 13px;
  color: var(--td-text-color-secondary, #666);
  line-height: 1.4;
  height: 56px; // 固定高度，约等于三行文字 (13px * 1.4 * 3 ≈ 54px，留一点余量)
  overflow: hidden;
  position: relative;
}

.description-text {
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 3; // 限制显示3行
  -webkit-box-orient: vertical;
  text-overflow: ellipsis;
  
  // 有溢出的文字在悬停时滚动
  &.has-overflow {
    .assistant-template-item:hover & {
      -webkit-line-clamp: unset; // 移除行数限制
      -webkit-box-orient: unset;
      display: block;
      animation: scrollText 10s linear infinite;
    }
  }
}

// 文字滚动动画
@keyframes scrollText {
  0% {
    transform: translateY(0);
  }
  20% {
    transform: translateY(0);
  }
  80% {
    transform: translateY(-100%); // 滚动到显示完整内容
  }
  100% {
    transform: translateY(0); // 回到起始位置
  }
}

// 滚动条样式
.assistants-list::-webkit-scrollbar {
  width: 6px;
}

.assistants-list::-webkit-scrollbar-track {
  background: var(--td-scroll-track-color, #f1f1f1);
  border-radius: 3px;
}

.assistants-list::-webkit-scrollbar-thumb {
  background: var(--td-scrollbar-color, #c1c1c1);
  border-radius: 3px;
}

.assistants-list::-webkit-scrollbar-thumb:hover {
  background: var(--td-scrollbar-hover-color, #a8a8a8);
}
</style>
