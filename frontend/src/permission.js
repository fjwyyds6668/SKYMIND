import { DialogPlugin } from 'tdesign-vue-next';
import { useChatStore } from './store/modules/chat.js';

// 全局路由拦截器
export function setupRouterGuards(router) {
  router.beforeEach((to, from, next) => {
    // 检查是否在 SmartQuery 页面内跳转到其他页面
    if (from.name === 'smart-query' && to.name !== 'smart-query') {
      // 获取聊天 store
      const chatStore = useChatStore();
      
      if (chatStore.getIsStreamLoad) {
        // 创建确认对话框
        DialogPlugin.confirm({
          header: '确认操作',
          body: '当前正在生成AI回复，确定要终止输出并跳转吗？<br><small style="color: var(--td-text-color-secondary, #666);">已生成的内容将被保存</small>',
          confirmBtn: {
            content: '确认',
            theme: 'primary',
          },
          cancelBtn: {
            content: '取消',
          },
          onConfirm: async () => {
            try {
              // 停止流式输出
              await chatStore.stopStream();
              // 继续路由跳转
              next();
            } catch (error) {
              // 即使停止失败，也允许跳转
              next();
            }
          },
          onCancel: () => {
            // 取消跳转
            next(false);
          },
        });
        return false; // 阻止当前跳转
      }
    }
    
    next();
  });
}
