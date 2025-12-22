<template>
  <div class="home-container">
    <t-card title="欢迎使用天灵AI工作台" :bordered="false">
      <t-space direction="vertical" size="large">
        <div class="welcome-section">
          <h2>🤖 天灵AI工作台</h2>
          <p class="subtitle">您的智能工作伙伴，让AI赋能每一个工作场景</p>
        </div>
        
        <t-alert theme="success" message="系统状态" :closeBtn="true">
          <template #message>
            <div class="status-info">
              <p><IconCheckCircle :size="16" /> 系统运行正常</p>
              <p><IconWifi :size="16" /> 网络连接正常</p>
              <p><IconServer :size="16" /> AI服务正常</p>
            </div>
          </template>
        </t-alert>
        
        <t-divider />
        
        <h3>🚀 功能模块</h3>
        <t-row :gutter="[16, 16]">
          <t-col :span="6">
            <t-card title="AI助手中心" :bordered="true" hover-shadow @click="navigateToModule('smart-query')">
              <template #avatar>
                <IconMessageCircle :size="24" />
              </template>
              <p>AI驱动的AI助手中心系统，快速解答您的疑问</p>
              <template #footer>
                <t-button theme="primary" variant="text">开始问答</t-button>
              </template>
            </t-card>
          </t-col>
          <t-col :span="6">
            <t-card title="系统智能查询" :bordered="true" hover-shadow @click="navigateToModule('file-interpreter')">
              <template #avatar>
                <IconFileText :size="24" />
              </template>
              <p>智能解析各应用系统数据，提取关键信息</p>
              <template #footer>
                <t-button theme="primary" variant="text">解读文件</t-button>
              </template>
            </t-card>
          </t-col>
          <t-col :span="6">
            <t-card title="全盘智能搜索" :bordered="true" hover-shadow @click="navigateToModule('full-search')">
              <template #avatar>
                <IconSearch :size="24" />
              </template>
              <p>强大的全文检索功能，快速定位所需内容</p>
              <template #footer>
                <t-button theme="primary" variant="text">开始搜索</t-button>
              </template>
            </t-card>
          </t-col>
          <t-col :span="6">
            <t-card title="个人计划助理" :bordered="true" hover-shadow @click="navigateToModule('task-assistant')">
              <template #avatar>
                <IconCheckSquare :size="24" />
              </template>
              <p>智能任务管理，提升工作效率</p>
              <template #footer>
                <t-button theme="primary" variant="text">管理任务</t-button>
              </template>
            </t-card>
          </t-col>
        </t-row>
        
        <t-row :gutter="[16, 16]" style="margin-top: 16px;">
          <t-col :span="6">
            <t-card title="系统使用顾问" :bordered="true" hover-shadow @click="navigateToModule('system-advisor')">
              <template #avatar>
                <IconSettings :size="24" />
              </template>
              <p>系统级建议和优化，保障最佳性能</p>
              <template #footer>
                <t-button theme="primary" variant="text">获取建议</t-button>
              </template>
            </t-card>
          </t-col>
        </t-row>
        
        <t-divider />
        
        <h3>📊 使用统计</h3>
        <t-row :gutter="[16, 16]">
          <t-col :span="4">
            <t-statistic title="今日问答" :value="128" suffix="次" />
          </t-col>
          <t-col :span="4">
            <t-statistic title="系统智能查询" :value="45" suffix="个" />
          </t-col>
          <t-col :span="4">
            <t-statistic title="搜索次数" :value="256" suffix="次" />
          </t-col>
          <t-col :span="4">
            <t-statistic title="任务完成" :value="89" suffix="%" :decimalPlaces="1" />
          </t-col>
        </t-row>
      </t-space>
    </t-card>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';

const router = useRouter();

// 导航到指定模块
const navigateToModule = (moduleName) => {
  router.push({ name: moduleName });
};

const handleCreateProject = () => {
  
};

const handleViewData = () => {
  
};

const handleSystemSettings = () => {
  
};

const handleHelp = () => {
  
};
</script>

<style scoped>
.home-container {
  max-width: 1200px;
  margin: 0 auto;
}

.welcome-section {
  text-align: center;
  padding: 20px 0;
}

.welcome-section h2 {
  margin: 0 0 10px 0;
  font-size: 2.5rem;
  font-weight: 600;
  color: #0052d9;
}

.subtitle {
  font-size: 1.1rem;
  color: #666;
  margin: 0;
}

.status-info p {
  margin: 5px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.t-divider {
  margin: 20px 0;
}

/* 卡片悬停效果 */
.t-card {
  cursor: pointer;
  transition: all 0.3s ease;
}

.t-card:hover {
  transform: translateY(-2px);
}

/* 统计数据样式 */
.t-statistic {
  text-align: center;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .welcome-section h2 {
    font-size: 2rem;
  }
  
  .subtitle {
    font-size: 1rem;
  }
}
</style>
