export default [
    {
      path: '/',
      redirect: '/smart-query'
    },
    {
      path: '/smart-query',
      name: 'smart-query',
      component: () => import('@/views/SmartQuery/index.vue'),
    },
    {
      path: '/file-interpreter',
      name: 'file-interpreter',
      component: () => import('@/views/FileInterpreter/index.vue'),
    },
    {
      path: '/full-search',
      name: 'full-search',
      component: () => import('@/views/FullSearch/index.vue'),
    },
    {
      path: '/task-assistant',
      name: 'task-assistant',
      component: () => import('@/views/TaskAssistant/index.vue'),
    },
    {
      path: '/system-advisor',
      name: 'system-advisor',
      component: () => import('@/views/SystemAdvisor/index.vue'),
    },
  ];
