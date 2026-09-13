import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'element-plus/theme-chalk/dark/css-vars.css'

// Element Plus 按需注册（见 vite.config.ts 的 resolver）。resolver 只覆盖模板标签，
// 程序式调用的 ElMessage / ElMessageBox 需显式引入样式。
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'

// Self-hosted fonts (font-display: swap). Makes --gk-font-sans / --gk-font-mono real.
import '@fontsource/inter/400.css'
import '@fontsource/inter/500.css'
import '@fontsource/inter/600.css'
import '@fontsource/inter/700.css'
import '@fontsource/jetbrains-mono/400.css'
import '@fontsource/jetbrains-mono/500.css'

import router from './router'
import i18n from './locales'
import App from './App.vue'
import { initTheme } from './styles/theme-switcher'
import './styles/index.css'
import './style.css'

// Initialize theme before mounting
initTheme()

const app = createApp(App)

// Global error handler: surface uncaught render/runtime errors instead of a blank subtree
app.config.errorHandler = (err, _instance, info) => {
  console.error('[gkube] Uncaught error:', err, info)
}

// Global unhandled promise rejection handler
window.addEventListener('unhandledrejection', (event) => {
  // Ignore cancelled requests (AbortController)
  if (event.reason?.name === 'AbortError' || event.reason?.code === 'ERR_CANCELED') {
    event.preventDefault()
    return
  }
  console.error('[gkube] Unhandled promise rejection:', event.reason)
})

app.use(createPinia())
app.use(i18n)
app.use(router)
app.mount('#app')
