import { createApp } from 'vue';
import { createPinia } from 'pinia';

import App from './App.vue';
import router from './router';
import store from './store';
import { setupRouterGuards } from './permission.js';

import './style.css';

import TDesign from 'tdesign-vue-next';
import TDesignChat from '@tdesign-vue-next/chat';

// 引入组件库全局样式资源
import 'tdesign-vue-next/es/style/index.css'; // 引入少量全局样式变量
// 引入 TDesign Chat 组件样式
import '@tdesign-vue-next/chat/es/style/index.css';

// 全局注册 lucide-vue-next 图标
import * as Icons from 'lucide-vue-next';

var app = createApp(App)
const pinia = createPinia();

// 全局注册所有图标组件，添加 Icon 前缀防止冲突
Object.entries(Icons).forEach(([name, component]) => {
  const iconName = `Icon${name}`;
  app.component(iconName, component);
});

app.use(TDesign)
app.use(TDesignChat)
app.use(router)
app.use(store)
app.use(pinia)

// 设置路由守卫
setupRouterGuards(router);

app.mount('#app');
