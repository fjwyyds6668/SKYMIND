// vite.config.js
import { defineConfig } from "file:///E:/33_Lab/skymind/frontend/node_modules/vite/dist/node/index.js";
import vue from "file:///E:/33_Lab/skymind/frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs";
import path from "path";
var __vite_injected_original_dirname = "E:\\33_Lab\\skymind\\frontend";
var vite_config_default = defineConfig({
  base: "./",
  // 使用相对路径
  resolve: {
    alias: {
      "@": path.resolve(__vite_injected_original_dirname, "src")
      // 设置别名 @ 指向 src 目录
    }
  },
  plugins: [
    // vue 3
    vue()
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
    assetsInlineLimit: 0,
    // 确保字体文件不被内联
    rollupOptions: {
      output: {
        assetFileNames: (assetInfo) => {
          if (assetInfo.name && assetInfo.name.includes("NotoSansSC")) {
            return "assets/fonts/[name][extname]";
          }
          return "assets/[name]-[hash][extname]";
        }
      }
    }
  }
});
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZS5jb25maWcuanMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCJFOlxcXFwzM19MYWJcXFxcc2t5bWluZFxcXFxmcm9udGVuZFwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiRTpcXFxcMzNfTGFiXFxcXHNreW1pbmRcXFxcZnJvbnRlbmRcXFxcdml0ZS5jb25maWcuanNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL0U6LzMzX0xhYi9za3ltaW5kL2Zyb250ZW5kL3ZpdGUuY29uZmlnLmpzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSAndml0ZSdcclxuaW1wb3J0IHZ1ZSBmcm9tICdAdml0ZWpzL3BsdWdpbi12dWUnXHJcbi8vIGltcG9ydCBBdXRvSW1wb3J0IGZyb20gJ3VucGx1Z2luLWF1dG8taW1wb3J0L3ZpdGUnO1xyXG4vLyBpbXBvcnQgQ29tcG9uZW50cyBmcm9tICd1bnBsdWdpbi12dWUtY29tcG9uZW50cy92aXRlJztcclxuLy8gaW1wb3J0IHsgVERlc2lnblJlc29sdmVyIH0gZnJvbSAndW5wbHVnaW4tdnVlLWNvbXBvbmVudHMvcmVzb2x2ZXJzJztcclxuaW1wb3J0IHBhdGggZnJvbSAncGF0aCdcclxuLy8gaHR0cHM6Ly92aXRlanMuZGV2L2NvbmZpZy9cclxuZXhwb3J0IGRlZmF1bHQgZGVmaW5lQ29uZmlnKHtcclxuICBiYXNlOiAnLi8nLCAvLyBcdTRGN0ZcdTc1MjhcdTc2RjhcdTVCRjlcdThERUZcdTVGODRcclxuICByZXNvbHZlOiB7XHJcbiAgICBhbGlhczoge1xyXG4gICAgICAnQCc6IHBhdGgucmVzb2x2ZShfX2Rpcm5hbWUsICdzcmMnKSAvLyBcdThCQkVcdTdGNkVcdTUyMkJcdTU0MEQgQCBcdTYzMDdcdTU0MTEgc3JjIFx1NzZFRVx1NUY1NVxyXG4gICAgfVxyXG4gIH0sXHJcbiAgcGx1Z2luczogW1xyXG4gICAgLy8gdnVlIDNcclxuICAgIHZ1ZSgpLFxyXG4gICAgLy8gXHU2MzA5XHU5NzAwXHU1RjE1XHU1MTY1XHU3RUM0XHU0RUY2XHJcbiAgICAvLyBBdXRvSW1wb3J0KHtcclxuICAgIC8vICAgcmVzb2x2ZXJzOiBbVERlc2lnblJlc29sdmVyKHtcclxuICAgIC8vICAgICBsaWJyYXJ5OiAndnVlLW5leHQnXHJcbiAgICAvLyAgIH0pXVxyXG4gICAgLy8gfSksXHJcbiAgICAvLyBDb21wb25lbnRzKHtcclxuICAgIC8vICAgcmVzb2x2ZXJzOiBbVERlc2lnblJlc29sdmVyKHtcclxuICAgIC8vICAgICBsaWJyYXJ5OiAndnVlLW5leHQnXHJcbiAgICAvLyAgIH0pXVxyXG4gICAgLy8gfSlcclxuICAgIC8vIG90aGVyIHBsdWdpbnNcclxuICBdLFxyXG4gIGJ1aWxkOiB7XHJcbiAgICBhc3NldHNJbmxpbmVMaW1pdDogMCwgLy8gXHU3ODZFXHU0RkREXHU1QjU3XHU0RjUzXHU2NTg3XHU0RUY2XHU0RTBEXHU4OEFCXHU1MTg1XHU4MDU0XHJcbiAgICByb2xsdXBPcHRpb25zOiB7XHJcbiAgICAgIG91dHB1dDoge1xyXG4gICAgICAgIGFzc2V0RmlsZU5hbWVzOiAoYXNzZXRJbmZvKSA9PiB7XHJcbiAgICAgICAgICAvLyBcdTc4NkVcdTRGRERcdTVCNTdcdTRGNTNcdTY1ODdcdTRFRjZcdTRGRERcdTYzMDFcdTUzOUZcdTU5Q0JcdTU0MERcdTc5RjBcdTU0OENcdTdFRDNcdTY3ODRcclxuICAgICAgICAgIGlmIChhc3NldEluZm8ubmFtZSAmJiBhc3NldEluZm8ubmFtZS5pbmNsdWRlcygnTm90b1NhbnNTQycpKSB7XHJcbiAgICAgICAgICAgIHJldHVybiAnYXNzZXRzL2ZvbnRzL1tuYW1lXVtleHRuYW1lXSc7XHJcbiAgICAgICAgICB9XHJcbiAgICAgICAgICByZXR1cm4gJ2Fzc2V0cy9bbmFtZV0tW2hhc2hdW2V4dG5hbWVdJztcclxuICAgICAgICB9XHJcbiAgICAgIH1cclxuICAgIH1cclxuICB9LFxyXG59KVxyXG4iXSwKICAibWFwcGluZ3MiOiAiO0FBQXdRLFNBQVMsb0JBQW9CO0FBQ3JTLE9BQU8sU0FBUztBQUloQixPQUFPLFVBQVU7QUFMakIsSUFBTSxtQ0FBbUM7QUFPekMsSUFBTyxzQkFBUSxhQUFhO0FBQUEsRUFDMUIsTUFBTTtBQUFBO0FBQUEsRUFDTixTQUFTO0FBQUEsSUFDUCxPQUFPO0FBQUEsTUFDTCxLQUFLLEtBQUssUUFBUSxrQ0FBVyxLQUFLO0FBQUE7QUFBQSxJQUNwQztBQUFBLEVBQ0Y7QUFBQSxFQUNBLFNBQVM7QUFBQTtBQUFBLElBRVAsSUFBSTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBLEVBYU47QUFBQSxFQUNBLE9BQU87QUFBQSxJQUNMLG1CQUFtQjtBQUFBO0FBQUEsSUFDbkIsZUFBZTtBQUFBLE1BQ2IsUUFBUTtBQUFBLFFBQ04sZ0JBQWdCLENBQUMsY0FBYztBQUU3QixjQUFJLFVBQVUsUUFBUSxVQUFVLEtBQUssU0FBUyxZQUFZLEdBQUc7QUFDM0QsbUJBQU87QUFBQSxVQUNUO0FBQ0EsaUJBQU87QUFBQSxRQUNUO0FBQUEsTUFDRjtBQUFBLElBQ0Y7QUFBQSxFQUNGO0FBQ0YsQ0FBQzsiLAogICJuYW1lcyI6IFtdCn0K
