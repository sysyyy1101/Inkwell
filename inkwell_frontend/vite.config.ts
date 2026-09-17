import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// 开发环境后端地址, 需要时可用环境变量覆盖: VITE_DEV_PROXY=http://127.0.0.1:8081 npm run dev
const backendTarget = process.env.VITE_DEV_PROXY || 'http://127.0.0.1:8081'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  // 打包产物保持 dist/ + static/ 的目录结构, 和老版本一致
  build: {
    outDir: 'dist',
    assetsDir: 'static',
    chunkSizeWarningLimit: 800,
  },
  server: {
    host: '127.0.0.1',
    port: 8080,
    // 端口被占用时直接报错, 避免前端悄悄换到别的端口导致代理地址对不上
    strictPort: true,
    // 前端只请求 /api/v1/... 和 /swagger/..., 由开发服务器转发到后端, 因此不存在跨域问题。
    // /swagger 一起代理, 页脚里的"接口文档"链接就能用相对路径(换后端端口不用改代码)。
    proxy: {
      '/api/v1': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/swagger': {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
  preview: {
    host: '127.0.0.1',
    port: 8080,
    strictPort: true,
    // 预览打包产物时也把接口转发到后端, 方便本地验收
    proxy: {
      '/api/v1': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/swagger': {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
})
