import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { install as MonacoVueEditor } from '@guolao/vue-monaco-editor'
import router from './router'
import i18n from './locales'
import App from './App.vue'
import './style.css'

const app = createApp(App)

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
