import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL || 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
      '/v1': {
        target: process.env.VITE_API_URL || 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
      },
    },
  },
  build: {
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (id.includes('node_modules/element-plus')) {
            return 'element-plus';
          }
          if (id.includes('node_modules/vue-i18n')) {
            return 'vue-i18n';
          }
          if (id.includes('node_modules/echarts')) {
            return 'echarts';
          }
          if (id.includes('node_modules/monaco-editor') || id.includes('node_modules/@guolao/vue-monaco-editor')) {
            return 'monaco-editor';
          }
          if (id.includes('node_modules/vue/') || id.includes('node_modules/vue-router') || id.includes('node_modules/pinia') || id.includes('node_modules/axios')) {
            return 'vendor';
          }
        },
      },
    },
  },
  optimizeDeps: {
    include: ['element-plus'],
  },
})
