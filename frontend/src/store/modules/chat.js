import { defineStore } from 'pinia';

export const useChatStore = defineStore('chat', {
  state: () => ({
    isStreamingChat: false,
    chatComponent: null, // 存储 chat 组件实例
  }),
  
  getters: {
    // 获取流式输出状态
    getIsStreamLoad: (state) => state.isStreamingChat,
    
    // 获取 chat 组件实例
    getChatComponent: (state) => state.chatComponent,
  },
  
  actions: {
    // 设置流式输出状态
    setStreamLoad(status) {
      this.isStreamingChat = status;
    },
    
    // 设置 chat 组件实例
    setChatComponent(component) {
      this.chatComponent = component;
    },
    
    // 停止流式输出
    async stopStream() {
      if (this.chatComponent && this.isStreamingChat) {
        try {
          // 调用组件的停止方法，它会返回 Promise
          await this.chatComponent.onStop();
          
          // 等待一个微任务确保所有异步操作完成
          await new Promise(resolve => setTimeout(resolve, 0));
        } catch (error) {
          throw error;
        }
      }
    },
  },
});
