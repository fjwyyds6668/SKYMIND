<template>
  <div class="layout-container">
    <t-layout class="full-layout">
      <!-- 侧边栏 -->
      <t-aside class="sidebar" :width="collapsed ? '60px' : '200px'">
        <div class="sidebar-content">
          <t-menu
            :collapsed="collapsed"
            :default-expanded="['menu']"
            :default-active="activeMenu"
            :theme="'light'"
            :value="activeMenu"
            :onChange="changeMenu"
            width="200px"
          >
            <template #logo>
              <div class="logo-container">
                <img
                  v-if="!collapsed"
                  src="@/assets/images/logo-rectangle.png"
                  class="logo"
                  alt="logo"
                />
                <img
                  v-else
                  src="@/assets/images/logo-round.png"
                  class="logo-collapsed"
                  alt="logo"
                />
              </div>
            </template>

            <t-menu-item value="smart-query">
              <template #icon>
                <IconMessageCircle :size="20" />
              </template>
              <div class="menu-text">AI助手中心</div>
            </t-menu-item>

            <t-menu-item value="file-interpreter">
              <template #icon>
                <IconFileText :size="20" />
              </template>
              <div class="menu-text">系统智能查询</div>
            </t-menu-item>

            <t-menu-item value="full-search">
              <template #icon>
                <IconSearch :size="20" />
              </template>
              <div class="menu-text">全盘智能搜索</div>
            </t-menu-item>

            <t-menu-item value="task-assistant">
              <template #icon>
                <IconCheckSquare :size="20" />
              </template>
              <div class="menu-text">个人计划助理</div>
            </t-menu-item>

            <t-menu-item value="system-advisor">
              <template #icon>
                <IconSettings :size="20" />
              </template>
              <div class="menu-text">系统使用顾问</div>
            </t-menu-item>
          </t-menu>
        </div>
      </t-aside>

      <!-- 主体内容 -->
      <t-layout class="main-layout">
        <t-header class="header">
          <div class="header-left">
            <t-button
              theme="default"
              shape="square"
              variant="text"
              @click="toggleCollapse"
            >
              <IconMenu :size="20" />
            </t-button>
            <h3 class="page-title">{{ pageTitle }}</h3>
          </div>
          <div class="header-center">
            <span class="system-time">{{ systemTime }}</span>
          </div>
          <div class="header-right">
            <t-button theme="default" variant="text" @click="handleLogout">
              <IconPower :size="20" />
              <div class="logout-text">退出</div>
            </t-button>
          </div>
        </t-header>

        <t-content class="content">
          <div class="content-wrapper">
            <router-view />
          </div>
        </t-content>
      </t-layout>
    </t-layout>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { useRouter, useRoute } from "vue-router";
import { HideWindow } from '../../wailsjs/go/main/App';
import { formatSystemTime } from '../views/SmartQuery/utils.js';

const router = useRouter();
const route = useRoute();

// 侧边栏折叠状态
const collapsed = ref(false);

// 当前激活的菜单项
const activeMenu = ref(route.name || "smart-query");

// 页面标题
const pageTitle = computed(() => {
  const titles = {
    "smart-query": "AI助手中心",
    "file-interpreter": "系统智能查询",
    "full-search": "全盘智能搜索",
    "task-assistant": "个人计划助理",
    "system-advisor": "系统使用顾问",
  };
  return titles[activeMenu.value] || "天灵AI工作台";
});

// 系统时间
const systemTime = ref("");

// 更新系统时间
const updateSystemTime = () => {
  systemTime.value = formatSystemTime();
};

// 定时器
let timer = null;

// 切换侧边栏折叠状态
const toggleCollapse = () => {
  collapsed.value = !collapsed.value;
};

// 菜单切换处理
const changeMenu = (value) => {
  activeMenu.value = value;
  router.push({ name: value });
};

// 登出处理
const handleLogout = async () => {
  try {
    await HideWindow();
  } catch (error) {
    MessagePlugin.error("退出至托盘失败:");
  }
};

// 监听路由变化，同步激活菜单项
watch(
  () => route.name,
  (newName) => {
    if (newName) {
      activeMenu.value = newName;
    }
  }
);

// 组件挂载时启动定时器
onMounted(() => {
  updateSystemTime();
  timer = setInterval(updateSystemTime, 1000);
});

// 组件卸载前清除定时器
onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer);
  }
});
</script>

<style scoped>
.layout-container {
  height: 100vh;
  padding: 10px;
  background-color: #f0f2f5;
  overflow: hidden;
  box-sizing: border-box;
}

.full-layout {
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 100%;
}

.sidebar {
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 2px 0 8px 0 rgba(0, 0, 0, 0.1);
  transition: width 0.3s ease;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
  flex-shrink: 0;
}

.sidebar-content {
  flex: 1;
  overflow-x: hidden;
  overflow-y: auto;
  height: calc(100% - 32px);
  box-sizing: border-box;
}

.logo-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  border-bottom: 1px solid #eee;
  box-sizing: border-box;
}

.logo {
  width: 200px;
  height: auto;
  max-width: 100%;
  margin-left: -24px;
}

.logo-collapsed {
  width: 30px;
  height: auto;
  max-width: 100%;
}

.main-layout {
  padding-left: 10px;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: calc(100% - 10px);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 10px;
  height: 64px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px 0 rgba(0, 0, 0, 0.1);
  z-index: 100;
  margin-bottom: 10px;
  flex-shrink: 0;
  overflow: hidden;
  box-sizing: border-box;
  width: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

.header-center {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
}

.system-time {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.header-right {
  display: flex;
  align-items: center;
}

.content {
  padding: 10px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px 0 rgba(0, 0, 0, 0.1);
  flex: 1;
  overflow-x: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  width: 100%;
}

.content-wrapper {
  flex: 1;
  overflow-x: hidden;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.menu-text {
  margin-left: 8px;
}

.logout-text {
  margin-left: 6px;
  font-size: 16px;
  line-height: 20px;
}
</style>
