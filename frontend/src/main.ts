import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { install as MonacoVueEditor } from '@guolao/vue-monaco-editor'
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

// Register all Element Plus icons globally
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(ElementPlus)
app.use(MonacoVueEditor)
app.use(i18n)
app.use(router)
app.mount('#app')
