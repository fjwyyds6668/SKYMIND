import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
// import AutoImport from 'unplugin-auto-import/vite';
// import Components from 'unplugin-vue-components/vite';
// import { TDesignResolver } from 'unplugin-vue-components/resolvers';
import path from 'path'
// https://vitejs.dev/config/
export default defineConfig({
  base: './', // 使用相对路径
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src') // 设置别名 @ 指向 src 目录
    }
  },
  plugins: [
    // vue 3
    vue(),
    // 按需引入组件
    // AutoImport({
    //   resolvers: [TDesignResolver({
    //     library: 'vue-next'
    //   })]
    // }),
    // Components({
    //   resolvers: [TDesignResolver({
    //     library: 'vue-next'
    //   })]
    // })
    // other plugins
  ],
  build: {
    assetsInlineLimit: 0, // 确保字体文件不被内联
    rollupOptions: {
      output: {
        assetFileNames: (assetInfo) => {
          // 确保字体文件保持原始名称和结构
          if (assetInfo.name && assetInfo.name.includes('NotoSansSC')) {
            return 'assets/fonts/[name][extname]';
          }
          return 'assets/[name]-[hash][extname]';
        }
      }
    }
  },
})
