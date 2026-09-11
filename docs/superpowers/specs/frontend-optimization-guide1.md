# gkube 前端优化修复指南

> 本文档供 AI（Claude Code）和开发者使用，包含所有待修复问题的详细描述、修复方案和代码示例。
> 每个问题独立成节，可按优先级逐个执行。
> 经过三轮 code review（代码正确性 + AI 可执行性 + 交叉验证），所有已知问题已修正，代码示例已验证可执行性。
> 共 48 个优化项：#1-24（初版）+ #25-48（补充扫描）。其中 #26、#33 已确认为误报（代码已有正确实现）。

---

## Prerequisites

- **Node.js >= 18**（Vite 8 要求）。建议项目根目录加 `.nvmrc`。
- **TypeScript strict 模式**已开启（`tsconfig.app.json`）。
- 每个修复完成后必须执行 `npm run build` 确认无编译错误。
- 打包优化类修复（P1）有回滚风险：如果按需导入后组件样式丢失或图标不显示，在 `main.ts` 中临时恢复全量引入，定位问题组件后再按需修复。

---

## P0 — Bug / 内存泄漏（必须立即修）

### 1. ResourceQuotaDetail ECharts 实例泄漏

**文件:** `src/views/config/resourcequota/ResourceQuotaDetail.vue`

**问题:**
- `updateChart()` 每次调用都 `echarts.init()` 创建新实例，从不 dispose 旧实例
- 每次 `updateChart()` 都 `window.addEventListener('resize', ...)` 注册匿名回调，从不移除
- 组件没有 `onBeforeUnmount` 清理逻辑

**修复方案:**

复用已有实例，只更新配置，不要每次 dispose + reinit。以下是改造后的关键代码片段（需合并到现有组件中）：

```vue
<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'  // 此注释为后续步骤 #4 的提醒，当前步骤保持此 import 不变

const chartRef = ref<HTMLDivElement>()
let chartInstance: echarts.ECharts | null = null
let resizeHandler: (() => void) | null = null

function initChart() {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  resizeHandler = () => chartInstance?.resize()
  window.addEventListener('resize', resizeHandler)
}

function updateChart() {
  // 复用已有实例，只更新配置；首次调用时 init
  if (!chartInstance) initChart()
  if (!chartInstance) return
  chartInstance.setOption({ /* 保留原有的 option 配置 */ }, true)
}

// 在 onMounted 中调用 initChart（或让首次 updateChart 自动 init）
onMounted(() => initChart())

onBeforeUnmount(() => {
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
    resizeHandler = null
  }
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})
</script>
```

**AI 执行要点：** 这是代码片段，不是完整替换。需要在现有 `ResourceQuotaDetail.vue` 中：
1. 找到现有的 `updateChart` 函数，改为复用实例（去掉 `echarts.init`，改为 `setOption`）
2. 找到 `window.addEventListener('resize', ...)` 的位置，提取为具名函数并只注册一次
3. 添加 `onBeforeUnmount` 清理逻辑
4. 保留原有的 `chart.setOption(option)` 中的 option 配置

**验证方法:** 打开 ResourceQuotaDetail 页面，反复切换 tab，检查 DevTools Memory 面板中 ECharts 实例数量是否稳定。

---

### 2. UserList 搜索防抖定时器未清理

**文件:** `src/views/system/UserList.vue`

**问题:** `searchTimer` 在组件卸载时没有 `clearTimeout`，会导致在已卸载组件上执行 `fetchUsers()`。

**修复方案:**

```vue
<script setup lang="ts">
import { onUnmounted } from 'vue'

let searchTimer: ReturnType<typeof setTimeout> | null = null

function handleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchUsers(), 300)
}

// 添加这一行
onUnmounted(() => {
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
})
</script>
```

**验证方法:** 在 UserList 页面快速输入搜索内容，立即导航离开，检查控制台无 Vue 警告。

---

## P1 — 打包体积优化

### 3. Element Plus 按需导入

**文件:** `src/main.ts`, `vite.config.ts`, `src/App.vue`

**问题:** `app.use(ElementPlus)` 导入整个组件库，+200-300KB gzipped。

**修复方案:**

第一步，安装依赖：
```bash
cd frontend
npm i -D unplugin-vue-components unplugin-auto-import
```

第二步，修改 `vite.config.ts`：
```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
  // ⚠️ 保留现有的 resolve.alias、server.proxy 等配置，不要删除
  resolve: { /* 现有 alias 配置 */ },
  server: { /* 现有 proxy 配置 */ },
})
```

第三步，修改 `src/main.ts`，删除以下三行：
```diff
- import ElementPlus from 'element-plus'
- import 'element-plus/dist/index.css'
- app.use(ElementPlus)
```

注意保留：
```ts
import 'element-plus/theme-chalk/dark/css-vars.css' // 暗色主题变量需要保留
```

**⚠️ 注意 locale 配置：** 当前 `App.vue` 使用 `<el-config-provider :locale="...">` 做中英文切换。移除 `app.use(ElementPlus)` 后，`ElConfigProvider` 本身需要被按需导入。`unplugin-vue-components` 的 `ElementPlusResolver` 会自动处理 `ElConfigProvider` 的导入，但需要验证。如果 locale 切换失效，在 `App.vue` 中手动局部注册：
```ts
import { ElConfigProvider } from 'element-plus'
```

**⚠️ `unplugin-auto-import` 和 `unplugin-vue-components` 会在 `frontend/` 目录下生成 `auto-imports.d.ts` 和 `components.d.ts`。** 这两个文件必须提交到 git（它们是类型声明，不生成它们会导致 TypeScript 编译报错）。首次运行 `npm run dev` 或 `npm run build` 后会自动生成。

**⚠️ `imports: ['vue', 'vue-router', 'pinia']` 会自动导入这三个库的所有导出。** 现有的 `import { ref } from 'vue'` 等显式导入不会报错（冗余但无害），新代码可以省略。`auto-imports.d.ts` 提供 TypeScript 类型支持。如果团队偏好显式导入以提高可读性，可以去掉 `imports` 数组，只保留 `resolvers`。

**⚠️ 执行顺序：** 先修改 `vite.config.ts` 添加插件 → 运行 `npm run dev`（让它生成类型声明文件）→ 确认 `auto-imports.d.ts` 和 `components.d.ts` 已生成 → 然后再修改 `main.ts` 删除全量引入。如果顺序反了（先删 main.ts 再加插件），`npm run build` 会因为缺少类型声明而报错。

**⚠️ Element Plus 的样式由 `ElementPlusResolver` 自动按需导入。** 不再需要 `import 'element-plus/dist/index.css'` 全量样式。但 `element-plus/theme-chalk/dark/css-vars.css`（暗色主题变量）必须保留，它不是组件样式，而是 CSS 变量定义。

**回滚方案:** 如果按需导入后组件样式丢失，在 `main.ts` 中临时恢复 `import ElementPlus from 'element-plus'` + `app.use(ElementPlus)`，定位问题组件后再按需修复。

**验证方法:** `npm run build`，对比前后 `dist/assets/` 中 JS 文件总大小。同时验证 App.vue 的 locale 切换仍然正常。

---

### 4. ECharts 按需导入

**文件:**
- `src/views/dashboard/DashboardView.vue`
- `src/views/config/resourcequota/ResourceQuotaDetail.vue`

**问题:** `import * as echarts from 'echarts'` 导入整个 ECharts（~1MB），只用了 gauge 和 bar。

**⚠️ 此修复必须在 #7（Vite chunk 拆分）之前完成，否则 chunk 配置与实际导入不一致。**

**修复方案:**

创建共享的 echarts 初始化文件 `src/utils/echarts.ts`：
```ts
import * as echarts from 'echarts/core'
import { GaugeChart, BarChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  GraphicComponent, // DashboardView 使用 echarts.graphic.LinearGradient 需要
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([
  GaugeChart,
  BarChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  GraphicComponent,
  CanvasRenderer,
])

export { echarts }
```

然后在 DashboardView.vue 和 ResourceQuotaDetail.vue 中替换：
```diff
- import * as echarts from 'echarts'
+ import { echarts } from '@/utils/echarts'
```

**验证方法:** `npm run build`，检查 echarts chunk 大小从 ~1MB 降至 ~200KB。

---

### 5. 图标按需注册

**文件:** `src/main.ts`

**问题:** 遍历 `@element-plus/icons-vue` 全部注册为全局组件，+几十KB。

**修复方案:**

**⚠️ `ElementPlusResolver` 只解析 Element Plus 组件（`el-button` 等），不解析 `@element-plus/icons-vue` 的图标。** 删除全局注册后，每个使用图标的组件需要手动局部注册。

首先，删除 `main.ts` 中的全局注册：
```diff
- import * as ElementPlusIconsVue from '@element-plus/icons-vue'
- for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
-   app.component(key, component)
- }
```

然后在每个使用图标的组件中局部注册：
```vue
<script setup lang="ts">
import { Delete, Edit, Search } from '@element-plus/icons-vue'
</script>
```

**⚠️ 图标使用方式有两种：** 组件标签 `<Delete />` 和 prop 传入 `:icon="Delete"`。两种都需要局部 import。全面搜索所有使用点：
```bash
grep -rn -E '<[A-Z][a-zA-Z]+ />' src/ --include='*.vue' | grep -v 'el-' | head -20
grep -rn ':icon=' src/ --include='*.vue'
```

**⚠️ 图标使用方式不仅限于 `<el-icon>`。** 很多 Element Plus 组件通过 `icon` prop 传入图标（如 `<el-button :icon="Delete">`）。全面搜索所有使用点：
```bash
grep -rn -E '@element-plus/icons-vue|el-icon|:icon=' src/ --include='*.vue' --include='*.ts'
```

**验证方法:** `npm run build`，检查图标相关 chunk 大小。逐一检查使用图标的页面确认图标正常显示。

---

### 6. 删除 vue-echarts 无用依赖

**文件:** `package.json`

**问题:** `vue-echarts` 列为依赖但从未被 import，+50KB。

**修复方案:**
```bash
cd frontend
npm uninstall vue-echarts
```

**验证方法:** `npm run build` 成功，无编译错误。

---

### 7. Vite Chunk 拆分

**文件:** `vite.config.ts`

**问题:** echarts、monaco-editor、element-plus、xterm、vis-network 全部打进一个 bundle。

**⚠️ 此修复必须在 #3（Element Plus 按需导入）和 #4（ECharts 按需导入）之后执行。**

在 `vite.config.ts` 的 `defineConfig` 中添加：
```ts
build: {
  rollupOptions: {
    output: {
      manualChunks: {
        'echarts': ['echarts/core', 'echarts/charts', 'echarts/components', 'echarts/renderers'],
        'element-plus': ['element-plus', '@element-plus/icons-vue'],
        'monaco': ['@guolao/vue-monaco-editor', 'monaco-editor'],
        'xterm': ['@xterm/xterm', '@xterm/addon-fit', '@xterm/addon-web-links'],
        'vis': ['vis-network', 'vis-data'],
        'vendor': ['vue', 'vue-router', 'pinia', 'axios'],
      },
    },
  },
  chunkSizeWarningLimit: 600,
},
```

**验证方法:** `npm run build`，检查 `dist/assets/` 中生成了多个独立 chunk 文件。

---

### 8. Monaco Editor 按需加载

**文件:** `src/components/YamlEditor.vue`, `src/components/YamlDrawer.vue`

**问题:** Monaco Editor 体积大（~1MB），但只在用户打开 YAML 编辑器时才需要。

**修复方案:**

将 Monaco Editor 改为动态导入。实际代码中 YamlEditor.vue 使用的是 `import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'`，命名导出为 `Editor`：

```vue
<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const MonacoEditor = defineAsyncComponent(() =>
  import('@guolao/vue-monaco-editor').then(mod => mod.Editor)
)
</script>
```

同时在 `main.ts` 中移除全局注册：
```diff
- import { install as MonacoVueEditor } from '@guolao/vue-monaco-editor'
- app.use(MonacoVueEditor)
```

**⚠️ 该库的命名导出是 `Editor`（不是 `MonacoEditor`）。** 已确认源码中使用 `import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'`，`Editor` 是正确的导出名。

**验证方法:** 首屏加载时不包含 monaco chunk，打开 YAML 编辑器时才加载且功能正常。

---

## P2 — 用户体验修复

### 9. 表单"未保存离开"保护

**文件:** 新增 `src/composables/useUnsavedGuard.ts`，所有 Create/Edit 视图

**问题:** 用户填了一半表单后导航离开，数据静默丢失。

**修复方案:**

创建独立的 composable（不要改 `useEditDrawer`，职责不同）：

```ts
// src/composables/useUnsavedGuard.ts
import { ref, onUnmounted } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { ElMessageBox } from 'element-plus'

export function useUnsavedGuard() {
  const isDirty = ref(false)

  function handleBeforeUnload(e: BeforeUnloadEvent) {
    e.preventDefault()
    e.returnValue = ''
  }

  function markDirty() {
    if (!isDirty.value) {
      isDirty.value = true
      window.addEventListener('beforeunload', handleBeforeUnload)
    }
  }

  function markClean() {
    isDirty.value = false
    window.removeEventListener('beforeunload', handleBeforeUnload)
  }

  onBeforeRouteLeave((to, from) => {
    if (!isDirty.value) {
      return true
    }
    return ElMessageBox.confirm(
      '有未保存的更改，确定离开吗？',
      '提示',
      { type: 'warning', confirmButtonText: '离开', cancelButtonText: '取消' }
    ).then(() => true).catch(() => false)
  })

  onUnmounted(() => {
    window.removeEventListener('beforeunload', handleBeforeUnload)
  })

  return { isDirty, markDirty, markClean }
}
```

在表单组件中使用：
```vue
<script setup>
const { isDirty, markDirty, markClean } = useUnsavedGuard()

// 表单变更时调用 markDirty()
// 提交成功后调用 markClean()
</script>
```

**⚠️ 不要在 `onUnmounted` 里注册 `beforeunload`，那时组件已销毁，监听器来不及生效。** 必须在 `markDirty()` 时注册，`markClean()` 时移除。

**验证方法:** 在 Create 页面填写部分字段，点击侧边栏导航，应弹出确认对话框。刷新浏览器应弹出系统级确认。

---

### 10. 详情页 404 / 错误状态处理

**文件:** `src/composables/useDetailPage.ts`

**问题:** `fetchDetail` 失败时只弹 toast，页面显示空白。

**修复方案:**

在 `useDetailPage.ts` 中添加错误状态：
```ts
import { ref } from 'vue'
import { useRoute } from 'vue-router'  // 如果现有代码中未引入

export interface DetailError {
  type: 'not-found' | 'network' | 'forbidden' | 'unknown'
  message: string
}

export function useDetailPage(options: DetailPageOptions) {
  const route = useRoute()
  const detail = ref<any>(null)
  const loading = ref(false)
  const error = ref<DetailError | null>(null)  // 新增

  async function fetchDetail() {
    loading.value = true
    error.value = null
    try {
      const res: any = await options.fetchDetail(options.buildParams())
      detail.value = res?.data ?? res
    } catch (e: any) {
      const status = e?.response?.status
      if (status === 404) {
        error.value = { type: 'not-found', message: '该资源不存在或已被删除' }
      } else if (status === 403) {
        error.value = { type: 'forbidden', message: '没有权限访问该资源' }
      } else if (!e?.response) {
        error.value = { type: 'network', message: '网络错误，请检查连接后重试' }
      } else {
        error.value = { type: 'unknown', message: e?.message || '加载失败' }
      }
    } finally {
      loading.value = false
    }
  }

  return { detail, loading, error, fetchDetail }
}
```

在详情页视图中使用（`listRoute` 需要每个视图自己定义）：
```vue
<template>
  <div v-if="loading" v-loading="true" style="min-height: 200px"></div>
  <div v-else-if="error" class="error-state">
    <el-empty :description="error.message">
      <el-button v-if="error.type === 'network'" @click="fetchDetail">重试</el-button>
      <el-button @click="$router.push(listRoute)">返回列表</el-button>
    </el-empty>
  </div>
  <template v-else-if="detail">
    <!-- 正常内容 -->
  </template>
</template>

<script setup>
// 每个详情页需要定义 listRoute
const listRoute = '/workloads/pods'
</script>
```

**验证方法:** 访问一个不存在的资源详情页，应显示"该资源不存在或已被删除"+ 返回按钮，而非空白页。

---

### 11. 骨架屏

**文件:** Dashboard、列表页、详情页

**问题:** 所有加载状态都是 v-loading 旋转覆盖空白区域。

**修复方案:**

使用 Element Plus 自带的 `el-skeleton`：

Dashboard 骨架屏示例：
```vue
<template>
  <div v-if="loading">
    <!-- 统计卡片骨架 -->
    <el-row :gutter="16">
      <el-col :span="6" v-for="i in 4" :key="i">
        <el-skeleton :rows="2" animated />
      </el-col>
    </el-row>
    <!-- 图表骨架 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-skeleton-item variant="circle" style="width: 200px; height: 200px; margin: 0 auto" />
      </el-col>
      <el-col :span="12">
        <el-skeleton :rows="6" animated />
      </el-col>
    </el-row>
  </div>
  <template v-else>
    <!-- 实际内容 -->
  </template>
</template>
```

列表页骨架屏示例：
```vue
<template>
  <el-skeleton v-if="loading" :rows="10" animated />
  <el-table v-else :data="list">
    <!-- 列定义 -->
  </el-table>
</template>
```

**建议优先添加骨架屏的页面:**
1. `DashboardView.vue` — 6 个数据源，体验最差
2. 所有 List 视图（PodList、DeploymentList 等）
3. 所有 Detail 视图

**验证方法:** 在 DevTools Network 面板中将网速调为 Slow 3G，对比骨架屏 vs v-loading 的体验差异。

---

### 12. 命名空间/搜索同步到 URL

**文件:** `src/composables/useResourceList.ts`

**问题:** 标签过滤已同步到 URL `ls` 参数，但命名空间和搜索文本没有。

**修复方案:**

合并为一个 watcher，避免多个 watcher 竞争写入 `router.replace`：

```ts
import { useRoute, useRouter } from 'vue-router'

// 在 composable 内部
const route = useRoute()
const router = useRouter()

// 初始化时从 URL 读取
const selectedNamespace = ref(route.query.ns as string || '')
const searchName = ref(route.query.q as string || '')

// 合并为一个 watcher，避免竞争
watch([selectedNamespace, searchName], ([ns, q]) => {
  router.replace({
    query: {
      ...route.query,
      ns: ns || undefined,  // 空值时移除参数
      q: q || undefined,
    },
  })
})
```

**⚠️ 不要拆成两个独立 watcher。** 快速切换命名空间 + 输入搜索时，两个 watcher 可能基于同一个旧 `route.query` 写入，后一个覆盖前一个导致丢失参数。

**⚠️ 现有 `labelConditions` 已有独立 watcher 写入 `route.query.ls`。** 新增的 watcher 需要与它共存。建议将三个过滤条件（namespace、search、labelConditions）合并到同一个 watcher 中，避免互相覆盖。或者保留 `labelConditions` 的独立 watcher，但确保新 watcher 的 `...route.query` 展开能保留 `ls` 参数。

**验证方法:** 选择命名空间 "production"，搜索 "api"，复制 URL，在新标签页打开，应保持相同的筛选状态。快速切换命名空间和搜索，URL 参数不应丢失。

---

### 13. 关键错误使用 ElNotification

**文件:** 所有使用 `ElMessage.error` 的关键操作

**问题:** `ElMessage.error` 3 秒自动消失，用户可能错过。

**修复方案:**

将以下场景的 `ElMessage.error` 改为 `ElNotification.error`：
- 资源删除失败
- YAML 保存失败
- 资源创建失败
- 集群连接失败

```ts
import { ElNotification } from 'element-plus'

// 之前
ElMessage.error('删除失败: ' + e.message)

// 之后
ElNotification.error({
  title: '删除失败',
  message: e.message || '操作失败，请重试',
  duration: 0,  // 不自动关闭
})
```

保持 `ElMessage.error` 的场景（非关键，自动消失即可）：
- 列表加载失败（用户可以刷新重试）
- 搜索失败

**验证方法:** 触发一个删除失败，通知应停留在页面右上角直到手动关闭。

---

### 14. 空状态统一

**文件:** `src/composables/useResourceList.ts`, 各详情页

**问题:** 空状态有的用 `el-empty`，有的用纯文本，且不区分"无数据"和"搜索无结果"。

**修复方案:**

在 `useResourceList.ts` 中添加空状态计算属性：
```ts
import { computed } from 'vue'  // 如果文件中已有则无需重复导入

const emptyText = computed(() => {
  if (searchName.value) {
    return `没有匹配 "${searchName.value}" 的 ${options.resourceName}`
  }
  if (selectedNamespace.value) {
    return `命名空间 "${selectedNamespace.value}" 下没有 ${options.resourceName}`
  }
  return `暂无 ${options.resourceName}`
})

const showEmpty = computed(() => !loading.value && filteredList.value.length === 0)
```

在列表页模板中使用：
```vue
<el-empty v-if="showEmpty" :description="emptyText">
  <el-button type="primary" @click="$router.push(createRoute)">创建</el-button>
</el-empty>
```

统一所有详情页的空状态也使用 `el-empty`，替换 `NodeDetail.vue` 中的 `<div class="empty-hint">` 模式。

**验证方法:** 在列表页搜索一个不存在的名称，应显示"没有匹配 xxx 的 Pod"而非"暂无 Pod"。

---

### 15. Dashboard 缓存 + keep-alive

**文件:** `src/components/Layout/AppLayout.vue`, `src/views/dashboard/DashboardView.vue`

**问题:** Dashboard 每次访问都发 6 个 API 请求，无缓存。

**修复方案 A — keep-alive:**

在 `AppLayout.vue` 中：
```vue
<template>
  <el-container>
    <el-aside><Sidebar /></el-aside>
    <el-container>
      <el-header><Header /></el-header>
      <el-main>
        <!-- 将现有的 <router-view :key="reloadKey" /> 替换为以下内容 -->
        <router-view v-slot="{ Component, route }">
          <keep-alive :include="keepAlivePages">
            <component :is="Component" :key="clusterStore.currentCluster?.id || route.fullPath" />
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
// 保留现有的 isCollapse、sidebar 等逻辑，只新增 keepAlivePages
const keepAlivePages = ['DashboardView', 'PodList', 'DeploymentList', 'ServiceList']
// 可删除现有的 reloadKey computed（已由 :key 替代）
</script>
```

**⚠️ AppLayout 现有 `:key="clusterStore.currentCluster?.id"` 用于集群切换时 remount 视图。** 使用 keep-alive 时，`:key` 需要保留在 `<component>` 上（如上方模板所示），确保集群切换时缓存失效并重新加载。`route.fullPath` 作为 fallback 确保非集群切换的路由变化也能正确渲染。

**⚠️ `<script setup>` 不会自动设置组件 `name`，而 `keep-alive :include` 依赖组件 name 匹配。** 每个被缓存的组件必须显式声明 name：
```vue
<!-- DashboardView.vue -->
<script setup lang="ts">
defineOptions({ name: 'DashboardView' })
// ...
</script>
```

同样在 PodList、DeploymentList、ServiceList 中添加对应的 `defineOptions`。

**修复方案 B — 内存缓存（可与 keep-alive 配合）:**

在 DashboardView.vue 中添加简单缓存：
```ts
const dataCache = new Map<string, { data: any; timestamp: number }>()
const CACHE_TTL = 30_000 // 30秒

async function cachedFetch(key: string, fetcher: () => Promise<any>) {
  const cached = dataCache.get(key)
  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return cached.data
  }
  const data = await fetcher()
  dataCache.set(key, { data, timestamp: Date.now() })
  return data
}

// 使用
const overview = ref(null)
async function fetchOverview() {
  overview.value = await cachedFetch('overview', () => getDashboardOverview(clusterId.value))
}
```

**验证方法:** 访问 Dashboard，切到 PodList，再切回 Dashboard，第二次不应有 loading 状态（30秒内）。

---

### 16. 键盘快捷键

**文件:** `src/composables/useResourceList.ts`（已有 R 快捷键），新增全局快捷键

**修复方案:**

在 `src/composables/useKeyboard.ts` 中创建全局快捷键管理：
```ts
import { ref, type Ref, onMounted, onUnmounted } from 'vue'

type ShortcutHandler = (e: KeyboardEvent) => void

interface Shortcut {
  key: string
  handler: ShortcutHandler
  description: string
}

export function useKeyboard(shortcuts: Shortcut[], enabled?: Ref<boolean>) {
  const isEnabled = enabled ?? ref(true) // 每次调用创建独立 ref，避免共享
  function handleKeydown(e: KeyboardEvent) {
    if (!isEnabled.value) return

    // 输入框内不触发快捷键
    const target = e.target as HTMLElement
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
      return
    }

    for (const shortcut of shortcuts) {
      if (e.key === shortcut.key || e.key === shortcut.key.toLowerCase()) {
        e.preventDefault()
        shortcut.handler(e)
        return
      }
    }
  }

  onMounted(() => window.addEventListener('keydown', handleKeydown))
  onUnmounted(() => window.removeEventListener('keydown', handleKeydown))
}
```

**`enabled` 参数用于避免快捷键冲突。** 比如列表页和详情页都注册了 `e` 键，通过 `enabled` 控制只有当前活跃的 composable 响应。

建议的快捷键方案：
```
全局:
  /        → 聚焦搜索框
  Esc      → 关闭当前 drawer/dialog

列表页:
  R        → 刷新列表（已有）
  n        → 跳转新建页面
  Enter    → 打开选中行的详情

详情页:
  Backspace / ← → 返回列表
  e        → 打开 YAML 编辑器
```

**验证方法:** 在列表页按 `/`，搜索框应获得焦点。

---

## P3 — 架构 / 工程规范改进

### 17. useResourceList 拆分

**文件:** `src/composables/useResourceList.ts`（497行）

**问题:** 职责过多：数据获取、分页、搜索、过滤、YAML 编辑、删除、pending-delete 清理、键盘快捷键。

**修复方案:** 分步拆分，不要一次性重构。每拆一个模块，运行所有列表页确认功能正常后再继续下一步。

**AI 执行要点：** 先读取 `useResourceList.ts` 全文，理解各函数的依赖关系后再拆分。拆分后保持 `useResourceList` 的导出接口不变（返回值类型和字段名），这样所有消费它的视图文件不需要改动。

**第一步** — 抽取删除逻辑到 `src/composables/useResourceDelete.ts`：

**⚠️ 当前代码使用 `ref<Record<string, boolean>>({})` 存储 pending-delete 状态。** 下方示例改为 `shallowRef(new Set<string>())` 以解决响应性问题，但需要同步更新所有调用点：`pendingDeleteIds.value[id]`（Record 访问）改为 `pendingDeleteIds.value.has(id)`（Set 访问）。如果不想改调用点，保持 `Record<string, boolean>` 即可。

```ts
// src/composables/useResourceDelete.ts
import { shallowRef, onUnmounted } from 'vue'

export function useResourceDelete() {
  // ⚠️ 用 shallowRef 而非 ref，因为 ref(new Set()) 的 .add()/.delete() 不会自动触发 watcher
  const pendingDeleteIds = shallowRef(new Set<string>())
  const cleanupTimers = new Map<string, ReturnType<typeof setTimeout>>()

  function markPendingDelete(id: string) {
    const next = new Set(pendingDeleteIds.value)
    next.add(id)
    pendingDeleteIds.value = next // 重新赋值触发响应性
    scheduleCleanup(id)
  }

  function clearPendingDelete(id: string) {
    const next = new Set(pendingDeleteIds.value)
    next.delete(id)
    pendingDeleteIds.value = next
  }

  function scheduleCleanup(id: string) {
    const existing = cleanupTimers.get(id)
    if (existing) clearTimeout(existing)
    const timer = setTimeout(() => {
      clearPendingDelete(id)
      cleanupTimers.delete(id)
    }, 5000)
    cleanupTimers.set(id, timer)
  }

  onUnmounted(() => {
    cleanupTimers.forEach((timer) => clearTimeout(timer))
    cleanupTimers.clear()
  })

  return { pendingDeleteIds, markPendingDelete, clearPendingDelete }
}
```

**第二步** — 抽取 YAML 操作到 `src/composables/useResourceYaml.ts`

**第三步** — 抽取过滤逻辑到 `src/composables/useResourceFilter.ts`

**第四步** — `useResourceList` 变为组合层：
```ts
export function useResourceList(options) {
  const { data, loading, fetch } = useResourceFetch(options)
  const { filters, searchName, selectedNamespace } = useResourceFilter(options)
  const { pendingDeleteIds, markPendingDelete } = useResourceDelete()
  const { yamlDrawer, openYaml } = useResourceYaml(options)

  // 组装返回值
  return { data, loading, fetch, filters, searchName, selectedNamespace, ... }
}
```

**验证方法:** 每拆一个模块，运行所有列表页确认功能正常。

---

### 18. useClusterGraph 移动位置

**文件:** `src/composables/useClusterGraph.ts`（575行）

**问题:** 纯视觉/动画代码，只服务于登录页，放在 composables/ 里不合适。

**修复方案:**
```bash
mv src/composables/useClusterGraph.ts src/views/login/graph.ts
```

更新 LoginView.vue 中的 import：
```diff
- import { useClusterGraph, graphState } from '@/composables/useClusterGraph'
+ import { useClusterGraph, graphState } from './graph'
```

**验证方法:** 登录页动画正常工作。

---

### 19. 路由文件拆分

**文件:** `src/router/index.ts`（630行）

**问题:** 所有 ~80 条路由定义在一个文件里。

**修复方案:**

创建 `src/router/modules/` 目录，按领域拆分：

```bash
mkdir src/router/modules
```

`src/router/modules/workload.ts`:
```ts
import type { RouteRecordRaw } from 'vue-router'

const workloadRoutes: RouteRecordRaw[] = [
  {
    path: '/workloads/pods',
    name: 'PodList',
    component: () => import('@/views/workload/PodList.vue'),
    meta: { title: 'Pod', parent: 'Dashboard' },
  },
  // ... 其他 workload 路由
]

export default workloadRoutes
```

`src/router/index.ts` 改为：
```ts
import { createRouter, createWebHistory } from 'vue-router'
import workloadRoutes from './modules/workload'
import networkRoutes from './modules/network'
import storageRoutes from './modules/storage'
// ...

const routes = [
  { path: '/login', ... },
  ...workloadRoutes,
  ...networkRoutes,
  ...storageRoutes,
  // ...
]

export default createRouter({ history: createWebHistory(), routes })
```

**验证方法:** 所有路由正常工作，`npm run build` 无错误。

---

### 20. 全局错误边界组件

**文件:** 新增 `src/components/ErrorBoundary.vue`

**问题:** 渲染层报错用户只看到白屏。

**修复方案:**

```vue
<!-- src/components/ErrorBoundary.vue -->
<template>
  <div v-if="error" class="error-boundary">
    <el-result icon="error" title="页面渲染出错" :sub-title="error.message">
      <template #extra>
        <el-button type="primary" @click="retry">重试</el-button>
        <el-button @click="$router.push('/')">返回首页</el-button>
      </template>
    </el-result>
  </div>
  <div v-else :key="retryCount">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, onErrorCaptured, provide } from 'vue'

const error = ref<Error | null>(null)
const retryCount = ref(0)

onErrorCaptured((err) => {
  error.value = err
  console.error('[ErrorBoundary]', err)
  return false // 阻止错误继续向上传播
})

function retry() {
  error.value = null
  retryCount.value++ // 强制 slot 内容 remount，避免子组件内部状态损坏
}

// 提供给子组件手动重置错误的能力
provide('resetError', () => { error.value = null })
</script>

<style scoped>
.error-boundary {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}
</style>
```

**⚠️ Vue 3 的 `<slot>` 不支持 `:key` 属性。** 用 `<div :key="retryCount"><slot /></div>` 包裹，通过 key 变化强制整个 div 及其子内容 remount。

**⚠️ `retry` 必须通过 `retryCount` 强制 remount 子组件。** 仅清空 `error.value` 不够——子组件内部状态可能已经损坏，需要完全重建。

在 AppLayout.vue 中使用：
```vue
<template>
  <el-container>
    <el-aside><Sidebar /></el-aside>
    <el-container>
      <el-header><Header /></el-header>
      <el-main>
        <ErrorBoundary>
          <router-view />
        </ErrorBoundary>
      </el-main>
    </el-container>
  </el-container>
</template>
```

**验证方法:** 在某个组件中临时抛出错误 `throw new Error('test')`，应显示错误页面。点击"重试"后应恢复正常。

---

### 21. i18n 决策

**文件:** `src/locales/`, 所有 .vue 文件

**问题:** vue-i18n 已引入，但大量硬编码中文字符串没有走 i18n。

**建议：保持现状，不投入精力补全，也不移除已有的。**

原因：
- 已有 853 行英文翻译和完整的 `zh-CN.ts`，移除是破坏性的
- 假装支持 i18n 但实际不完整确实不好，但移除已有的翻译更浪费
- 新代码可以选择性地写 i18n，旧代码渐进迁移

**如果确实要移除（只服务中文用户）：**
```bash
npm uninstall vue-i18n
```
删除 `src/locales/` 目录，删除 `main.ts` 中的 `app.use(i18n)`，所有 `t('key')` 替换为直接中文字符串。

**如果要补全（需要英文支持）：**

用脚本扫描硬编码中文：
```bash
# Linux (GNU grep)
grep -rn '[一-鿿]' src/ --include='*.vue' --include='*.ts' | grep -v 'locales/' | grep -v 'node_modules/'

# macOS (BSD grep 需要 -E)
grep -rn -E '[一-鿿]' src/ --include='*.vue' --include='*.ts' | grep -v 'locales/' | grep -v 'node_modules/'
```

逐个文件替换为 `t()` 调用，将中文文本移入 `zh-CN.ts`，英文翻译写入 `en.ts`。

---

### 22. 触摸设备支持

**文件:** `src/composables/useResizable.ts`

**问题:** 只监听 mouse 事件，触摸设备无法拖拽调整面板。

**修复方案:**
```ts
function startResize(e: MouseEvent | TouchEvent) {
  e.preventDefault()
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX

  const onMove = (e: Event) => {
    const x = 'touches' in e ? (e as TouchEvent).touches[0].clientX : (e as MouseEvent).clientX
    // ... 计算新宽度
  }

  const onEnd = () => {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onEnd)
    document.removeEventListener('touchmove', onMove)
    document.removeEventListener('touchend', onEnd)
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onEnd)
  document.addEventListener('touchmove', onMove, { passive: false })
  document.addEventListener('touchend', onEnd)
}

// 绑定事件时同时绑定 mouse 和 touch
handle.addEventListener('mousedown', startResize)
handle.addEventListener('touchstart', startResize, { passive: false })
```

---

### 23. Monaco Editor 生命周期验证

**文件:** `src/components/YamlEditor.vue`

**问题:** 依赖 `@guolao/vue-monaco-editor` 内部处理 dispose，但该库有已知的不 dispose 问题。

**修复方案:**
```vue
<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'

const editorRef = ref<any>(null)

function handleEditorMount(editor: any) {
  editorRef.value = editor
}

onBeforeUnmount(() => {
  if (editorRef.value) {
    editorRef.value.dispose()
    editorRef.value = null
  }
})
</script>

<template>
  <!-- @guolao/vue-monaco-editor 的 editor 挂载事件为 @mount -->
  <MonacoEditor @mount="handleEditorMount" ... />
</template>
```

**⚠️ `@mount` 事件名已确认正确。** 当前 `YamlEditor.vue`（line 115）已使用 `@mount="handleEditorMount"` 且工作正常。

---

### 24. YAML 提交基础校验

**文件:** `src/views/workload/WorkloadCreate.vue`

**问题:** YAML 模式只做非空检查 + js-yaml 解析，不校验 K8s 清单结构。

**修复方案:**
```ts
import yaml from 'js-yaml'
import { ElMessage } from 'element-plus'

function validateK8sManifest(content: string): string | null {
  let parsed: any
  try {
    parsed = yaml.load(content)
  } catch (e: any) {
    return `YAML 解析错误: ${e.message}`
  }

  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    return 'YAML 内容不是有效的对象'
  }
  if (!parsed.apiVersion) {
    return '缺少 apiVersion 字段'
  }
  if (!parsed.kind) {
    return '缺少 kind 字段'
  }
  if (!parsed.metadata?.name) {
    return '缺少 metadata.name 字段'
  }
  return null // 校验通过
}

function handleYamlSubmit() {
  const error = validateK8sManifest(yamlContent.value)
  if (error) {
    ElMessage.error(error)
    return
  }
  // 继续提交
}
```

---

## P0 — Bug / 内存泄漏（补充）

### 25. PVDetail / HPADetail / NetworkPolicyDetail resize 监听器泄漏

**文件:**
- `src/views/storage/PVDetail.vue` (lines 29-51)
- `src/views/workload/hpa/HPADetail.vue` (lines 270-311)
- `src/views/network/networkpolicy/NetworkPolicyDetail.vue` (line 218)

**问题:** 三处内联的 `onVResizeStart` 函数向 `document` 注册 `mousemove` / `mouseup` 监听器，但没有 `onBeforeUnmount` 清理。如果用户拖拽面板期间导航离开，监听器仍在 document 上，回调操作已卸载的组件。

**修复方案:**

统一使用已有的 `useResizable` composable（`src/composables/useResizable.ts`），它已处理 `onBeforeUnmount` 清理。HPADetail 和 NetworkPolicyDetail 为单区域调整，可直接替换。PVDetail 需要先扩展 composable（见下方警告）：

```diff
- // PVDetail.vue 中的内联 resize 逻辑
- function onVResizeStart(e: MouseEvent) {
-   // ... 20 行内联代码
- }
+ import { useResizable } from '@/composables/useResizable'
+ const { onHResizeStart, onVResizeStart } = useResizable({ initialWidth: 320, minWidth: 200, maxWidth: 600 })
```

**⚠️ PVDetail 的 `onVResizeStart` 接受 `target` 参数（`'section1' | 'section2'`）来控制调整哪个区域的高度，而 `useResizable` 的 `onVResizeStart` 不支持此参数。** 需要先扩展 `useResizable` 支持多区域，或为 PVDetail 创建专用的 `useSectionResizable` composable。HPADetail 和 NetworkPolicyDetail 为单区域调整，可直接使用 `useResizable`。

**验证方法:** 在 PVDetail 拖拽面板，拖拽过程中点击侧边栏导航，控制台无报错。

---

### 26. ~~v-for 缺少 :key~~（已确认无需修改）

**已验证:** 对文档列出的所有 20+ 处逐一检查，均已存在 `:key` 绑定（如 `:key="sec"`、`:key="key"`、`:key="rs.metadata.name"` 等）。此条目为误报，跳过。

**注意:** `grep -rn 'v-for=' src/ --include='*.vue' | grep -v ':key'` 会产生误报（`:key` 常在下一行）。如需严格检查，使用 ESLint 的 `vue/no-v-for-without-key` 规则。

---

### 27. YamlEditor 全屏模式暗色主题背景硬编码白色

**文件:** `src/components/YamlEditor.vue`, line 387

**问题:** `.yaml-editor.is-fullscreen` 使用 `background: #fff` 硬编码白色。暗色模式下全屏编辑器白底+暗色文字，不可读。

**修复方案:**

```diff
  .yaml-editor.is-fullscreen {
-   background: #fff;
+   background: var(--el-bg-color);
  }
```

**验证方法:** 切换到暗色主题，打开 YAML 编辑器全屏模式，背景应跟随主题色。

---

## P1 — 打包体积优化（补充）

### 28. 删除 vis-network / vis-data 无用依赖

**文件:** `package.json` (lines 25-26)

**问题:** `vis-network` 和 `vis-data` 列为依赖但源码中从未 import。已被 `src/composables/useClusterGraph.ts`（基于 Canvas API）替代。两个包合计 ~200KB+。

**修复方案:**

```bash
cd frontend
npm uninstall vis-network vis-data
```

**⚠️ 先确认确实无引用：**
```bash
grep -rn "from 'vis-network'\|from 'vis-data'" frontend/src/ --include='*.vue' --include='*.ts'
```

**验证方法:** `npm run build` 成功，无编译错误。

---

## P2 — 用户体验修复（补充）

### 29. Router 缺少 scrollBehavior 配置

**文件:** `src/router/index.ts`

**问题:** `createRouter()` 没有 `scrollBehavior`。从长列表页导航到详情页再返回，滚动位置不重置，用户落在列表底部。

**修复方案:**

```diff
  const router = createRouter({
    history: createWebHistory(),
+   scrollBehavior(to, from, savedPosition) {
+     if (savedPosition) return savedPosition
+     return { top: 0 }
+   },
    routes: [...],
  })
```

**验证方法:** 在 PodList 滚动到底部，点击某个 Pod 进入详情页，按返回，列表应回到顶部。

---

### 30. TerminalView / LogView 选择器代码重复 ~200 行

**文件:**
- `src/views/terminal/TerminalView.vue`
- `src/views/logviewer/LogView.vue`

**问题:** 两个文件各有 ~200 行几乎相同的代码：`fetchClusters()`、`fetchNamespaces()`、`fetchPods()`、`fetchContainers()`，以及 `selectedCluster → selectedNamespace → selectedPod → selectedContainer` 的 watcher 链。两处已有微小分歧（错误处理不一致）。

**修复方案:**

提取共享 composable `src/composables/useTerminalSelectors.ts`：

```ts
// src/composables/useTerminalSelectors.ts
import { ref, watch } from 'vue'
import { useClusterStore } from '@/stores/cluster'

export function useTerminalSelectors() {
  const clusterStore = useClusterStore()

  const clusters = ref<any[]>([])
  const namespaces = ref<string[]>([])
  const pods = ref<string[]>([])
  const containers = ref<string[]>([])

  const selectedCluster = ref<any>(null)
  const selectedNamespace = ref('')
  const selectedPod = ref('')
  const selectedContainer = ref('')

  async function fetchClusters() { /* ... */ }
  async function fetchNamespaces() { /* ... */ }
  async function fetchPods() { /* ... */ }
  async function fetchContainers() { /* ... */ }

  watch(selectedCluster, () => { fetchNamespaces(); pods.value = []; containers.value = [] })
  watch(selectedNamespace, () => { fetchPods(); containers.value = [] })
  watch(selectedPod, () => { fetchContainers() })

  return {
    clusters, namespaces, pods, containers,
    selectedCluster, selectedNamespace, selectedPod, selectedContainer,
    fetchClusters,
  }
}
```

然后在 TerminalView 和 LogView 中：
```diff
- const clusters = ref([])
- const namespaces = ref([])
- // ... ~200 行重复代码
+ import { useTerminalSelectors } from '@/composables/useTerminalSelectors'
+ const { clusters, namespaces, pods, containers, selectedCluster, ... } = useTerminalSelectors()
```

**验证方法:** Terminal 和 Log 页面的集群/命名空间/Pod/容器选择链仍然正常工作。

---

### 31. AuditLog 无分页，一次性加载全部日志

**文件:** `src/views/audit/AuditLog.vue` (lines 44-55)

**问题:** `fetchAuditLogs` 调用 `request.get('/k8s/audit/list')` 无分页参数，`filteredLogs` 在客户端做全量过滤。生产环境日志量大时内存暴涨、过滤卡顿。

**修复方案:**

实现服务端分页：

```diff
+ const pagination = ref({ page: 1, size: 20, total: 0 })

  async function fetchAuditLogs() {
    loading.value = true
    try {
-     const res = await request.get('/k8s/audit/list')
-     auditLogs.value = res.data || []
+     const res = await request.get('/k8s/audit/list', {
+       params: {
+         page: pagination.value.page,
+         size: pagination.value.size,
+         // 服务端过滤参数
+         username: filterForm.username || undefined,
+         action: filterForm.action || undefined,
+         resource: filterForm.resource || undefined,
+       },
+     })
+     auditLogs.value = res.data?.items || []
+     pagination.value.total = res.data?.total || 0
    } finally {
      loading.value = false
    }
  }
```

模板中添加分页组件：
```vue
<el-pagination
  v-model:current-page="pagination.page"
  v-model:page-size="pagination.size"
  :total="pagination.total"
  @current-change="fetchAuditLogs"
  @size-change="fetchAuditLogs"
/>
```

**⚠️ 需要后端配合修改 `/k8s/audit/list` 接口支持 `page`、`size`、`username`、`action`、`resource` 参数。** 如果后端暂不支持，可先保留客户端过滤，但添加虚拟滚动作为临时方案。

**验证方法:** 审计日志页只显示第一页数据，翻页正常。

---

### 32. WorkloadForm generatedYaml computed 每次按键都执行

**文件:** `src/views/workload/components/WorkloadForm.vue`, line 303

**问题:** `const generatedYaml = computed(() => yaml.dump(buildK8sResource(), ...))` 中 `buildK8sResource()` 是 ~170 行的函数，`yaml.dump()` 是 YAML 序列化。每次表单字段变化（每个按键）都重新执行，造成输入延迟。

**修复方案:**

改为只在提交时计算（`generatedYaml` 仅在 `handleSubmit` 中使用，不需要实时预览）：

```diff
- const generatedYaml = computed(() => yaml.dump(buildK8sResource(), { indent: 2, lineWidth: -1, noRefs: true }))
+ function getGeneratedYaml() {
+   return yaml.dump(buildK8sResource(), { indent: 2, lineWidth: -1, noRefs: true })
+ }
```

然后在 `handleSubmit` 中将 `generatedYaml.value` 替换为 `getGeneratedYaml()`（约 3 处，lines 537, 548, 553）。如果 `computed` 没有其他用途，可移除 import。

**验证方法:** 在 WorkloadCreate 表单中快速打字，输入不应有延迟。

---

### 33. ~~ClusterList searchDebounce 未在组件卸载时清理~~（已确认无需修改）

**文件:** `src/views/cluster/ClusterList.vue`

**已验证:** 该文件 line 208 已有 `onUnmounted(() => clearTimeout(searchDebounce))`，定时器清理已存在。此条目为误报，跳过。

---

### 34. TerminalView / LogView 合成 cluster 对象缺少 id 字段

**文件:**
- `src/views/terminal/TerminalView.vue`, lines 54-58
- `src/views/logviewer/LogView.vue`, lines 276-280

**问题:** 当 URL 带 `cluster` 查询参数时，两个视图构造合成对象 `{ clusterName: name, cluster_name: name, name }` 并传入 `clusterStore.setCurrentCluster()`。该对象缺少 `id` 字段，导致 `clusterStore.clusterId` 返回 `0`，后续权限检查（`canDo`、`canWrite`）失败。

**修复方案:**

从 `clusterStore.clusterList` 中查找真实集群对象：

```diff
- const syntheticCluster = { clusterName: name, cluster_name: name, name }
- clusterStore.setCurrentCluster(syntheticCluster as any)
+ // 如果 clusterList 尚未加载（用户直接从 URL 进入），先加载
+ if (clusterStore.clusterList.length === 0) {
+   await clusterStore.fetchClusters()
+ }
+ const realCluster = clusterStore.clusterList.find(c => c.clusterName === name || c.name === name)
+ if (realCluster) {
+   clusterStore.setCurrentCluster(realCluster)
+ } else {
+   ElMessage.error(`集群 "${name}" 不存在`)
+   router.push('/login')
+ }
```

**验证方法:** 带 `?cluster=xxx` 查询参数直接访问 Terminal 页面，检查权限相关功能正常。

---

### 35. useAutoRefresh 轮询无错误退避

**文件:** `src/composables/useAutoRefresh.ts` (lines 47-51)

**问题:** `startPolling` 用 `setInterval` 固定间隔调用 `fetchFn`。网络错误或集群宕机时，每 N 秒持续失败请求，无退避策略。

**修复方案:**

```ts
let consecutiveFailures = 0
const MAX_BACKOFF = 60_000 // 最大退避 1 分钟

async function pollingTick() {
  try {
    await fetchFn()
    consecutiveFailures = 0 // 成功后重置
  } catch (e) {
    consecutiveFailures++
    console.warn(`[useAutoRefresh] Poll failed (${consecutiveFailures}x):`, e)
  } finally {
    // 连续失败时指数退避
    const delay = consecutiveFailures >= 3
      ? Math.min(baseInterval * Math.pow(2, consecutiveFailures - 3), MAX_BACKOFF)
      : baseInterval
    pollTimer = setTimeout(pollingTick, delay)
  }
}
```

**验证方法:** 模拟网络断开（DevTools Offline），观察请求间隔是否逐渐增大。恢复网络后自动回到正常间隔。

---

### 36. 缺少全局 unhandledrejection 处理

**文件:** `src/main.ts`

**问题:** 有 `app.config.errorHandler` 捕获 Vue 渲染错误，但没有 `window.addEventListener('unhandledrejection', ...)` 处理非 Vue 上下文的未捕获 Promise 错误（WebSocket 回调、SSE 流、轮询等）。

**修复方案:**

在 `main.ts` 中添加：

```ts
window.addEventListener('unhandledrejection', (event) => {
  // 忽略被取消的请求（AbortController）
  if (event.reason?.name === 'AbortError' || event.reason?.code === 'ERR_CANCELED') {
    event.preventDefault()
    return
  }
  console.error('[gkube] Unhandled promise rejection:', event.reason)
})
```

**⚠️ 不要无条件 `event.preventDefault()`，否则所有未处理的 Promise 错误都被静默吞掉。只过滤已知的无害错误（如 AbortError）。**

**验证方法:** 在终端 WebSocket 回调中临时抛出错误，检查控制台有 `[gkube]` 前缀的错误日志。

---

### 37. ContainerConfigForm 直接修改 prop

**文件:** `src/views/workload/components/form/ContainerConfigForm.vue` (lines 19-24)

**问题:** `addContainer()`、`removeContainer()`、`addPort()` 等函数通过 `props.containers.push(...)` 和 `props.containers.splice(...)` 直接修改 prop 数组。Vue 3 对 reactive 对象内的数组允许此操作，但违反单向数据流原则，调试困难，且未来如果 prop 改为 readonly 会静默失败。

**修复方案:**

改为 emit 事件：

```diff
+ const emit = defineEmits<{
+   'update:containers': [value: any[]]
+ }>()

  function addContainer() {
-   props.containers.push({ name: '', image: '', ports: [] })
+   emit('update:containers', [...props.containers, { name: '', image: '', ports: [] }])
  }

  function removeContainer(index: number) {
-   props.containers.splice(index, 1)
+   const next = [...props.containers]
+   next.splice(index, 1)
+   emit('update:containers', next)
  }
```

父组件使用 `v-model:containers`：
```vue
<ContainerConfigForm v-model:containers="formData.containers" />
```

**验证方法:** 添加/删除容器操作仍然正常，Vue DevTools 中不再出现 prop 直接修改的警告。

---

### 38. 搜索框清空时 debouncedSearch 有 200ms 延迟

**文件:** `src/composables/useResourceList.ts` (lines 75-82)

**问题:** 用户清空搜索框后，`debouncedSearch` 仍需等 200ms 才更新为空，列表在 200ms 内仍显示过滤后的结果。

**修复方案:**

清空时立即生效，不清空时才 debounce：

```diff
  function onSearchInput(value: string) {
    searchName.value = value
+   if (!value) {
+     if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
+     debouncedSearch.value = ''
+     return
+   }
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchDebounceTimer = setTimeout(() => { debouncedSearch.value = value }, 200)
  }
```

**验证方法:** 搜索 "api" 后清空搜索框，列表应立即恢复全部结果，无 200ms 闪烁。

---

## P3 — 架构 / 工程规范改进（补充）

### 39. 响应解包破坏类型安全

**文件:** `src/api/request.ts` (lines 72-79)

**问题:** 响应拦截器执行 `response.data = data.data`，覆盖了原始的 `AxiosResponse` 类型。所有 API 调用中 `request.get<T>` 的泛型 `T` 实际指向的是内层 data，但类型系统认为是外层 `ApiResponse<T>`。导致全代码库 119+ 处 `any` 类型，`createResourceApi` 的泛型形同虚设。

**修复方案:**

创建类型安全的解包函数，而不是修改 `response.data`：

```diff
+ // src/api/types.ts
+ export interface ApiResponse<T = any> {
+   code: number
+   msg: string
+   data: T
+ }
+
+ export function unwrap<T>(response: AxiosResponse<ApiResponse<T>>): T {
+   return response.data.data
+ }

  // src/api/resource.ts 中使用
  async function getDetail(namespace: string, name: string): Promise<TDetail> {
-   return request.get(`/k8s/${resource}/detail/${namespace}/${name}`)
+   return unwrap<TDetail>(await request.get(`/k8s/${resource}/detail/${namespace}/${name}`))
  }
```

同时删除响应拦截器中的 `response.data = data.data`，让它返回原始的 `AxiosResponse`。

**⚠️ 这是一个大范围重构，需要渐进式迁移。** 具体步骤：
1. 创建 `src/api/types.ts`，添加 `ApiResponse` 和 `unwrap` 函数
2. 逐个文件替换：先在 `resource.ts` 的 factory 函数中使用 `unwrap()`，确认类型正确
3. 继续替换其他 API 文件（`auth.ts`、`cluster.ts`、`dashboard.ts` 等）
4. **只有所有 API 调用都迁移完成后**，才删除响应拦截器中的 `response.data = data.data`
5. 如需并行迁移，可创建 `requestRaw` 实例（不 unwrap），逐步切换

**⚠️ 不要在迁移完成前删除 `response.data = data.data`，否则所有现有 API 调用会同时崩溃。**

**验证方法:** `npm run build` 无错误，逐步删除 `any` 类型注解，TypeScript 仍能正确推断返回类型。

---

### 40. cluster store deep watch 多余

**文件:** `src/stores/cluster.ts`, line 44

**问题:** `watch(currentCluster, ..., { deep: true })` 深度遍历整个集群对象来同步 localStorage。但 `setCurrentCluster()` 通过重新赋值触发浅层 watch 即可，`deep: true` 增加了不必要的性能开销（Cluster 接口有 `[key: string]: unknown`）。

**修复方案:**

```diff
- watch(currentCluster, (val) => {
-   localStorage.setItem('selectedCluster', JSON.stringify(val))
- }, { deep: true })
+ watch(currentCluster, (val) => {
+   localStorage.setItem('selectedCluster', JSON.stringify(val))
+ })  // 移除 { deep: true }，setCurrentCluster 已通过重新赋值触发
```

**验证方法:** 切换集群后刷新页面，选中的集群仍正确恢复。

---

### 41. localStorage 反序列化无版本控制

**文件:**
- `src/stores/cluster.ts` (lines 18-25)
- `src/stores/auth.ts` (line 33)

**问题:** `JSON.parse(localStorage.getItem(...))` 在模块顶层执行。如果存储的 JSON 是旧 schema（应用升级后），解析可能静默失败返回 `null`，用户丢失登录状态和集群选择，无任何提示。

**修复方案:**

添加版本标识和失败清理：

```ts
const STORAGE_VERSION = 2 // 递增此值以触发旧数据清理
// 实际 key 名称以代码为准：cluster store 用 'gkube_cluster'，auth store 用 'gkube_user'
const STORAGE_KEY = 'gkube_cluster'

function loadFromStorage<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return fallback
    const parsed = JSON.parse(raw)
    // 兼容旧格式（没有 _version 包裹的裸数据）
    if (parsed._version === undefined) {
      // 迁移：将旧数据包裹为新格式
      localStorage.setItem(key, JSON.stringify({ _version: STORAGE_VERSION, value: parsed }))
      return parsed as T
    }
    // 版本不匹配时清理
    if (parsed._version !== STORAGE_VERSION) {
      localStorage.removeItem(key)
      return fallback
    }
    return parsed.value
  } catch {
    localStorage.removeItem(key) // 清理损坏的数据
    return fallback
  }
}

function saveToStorage(key: string, value: unknown) {
  localStorage.setItem(key, JSON.stringify({ _version: STORAGE_VERSION, value }))
}
```

**⚠️ 首次部署此改动时，旧格式数据会自动迁移为新格式，不会丢失用户状态。**

**验证方法:** 手动在 localStorage 中修改存储的 JSON 为非法值，刷新页面，应用应正常降级而非白屏。

---

### 42. logout() 不调用后端 API

**文件:** `src/stores/auth.ts`

**问题:** `authStore.logout()` 只清除本地 token 和用户信息，不调用后端 `/auth/logout`。JWT 在过期前仍有效。`src/api/auth.ts` 有 `logout()` 函数（line 15）但从未被 store 调用。

**修复方案:**

```diff
  async function logout() {
+   // 先清除本地状态，确保用户立即退出（不被 API 超时阻塞）
    token.value = ''
    userInfo.value = null
    permissions.value = undefined
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
+   // 再通知后端注销（fire-and-forget）
+   try {
+     await authApi.logout()
+   } catch {
+     // 后端失败不影响本地退出
+   }
  }
```

**⚠️ 后端可能尚未实现 `/auth/logout`。** `api/auth.ts` 的 `logout()` 已有 try/catch 包裹，即使 404 也不会阻断流程。

**验证方法:** 点击退出登录后，用旧 token 调用 API 应返回 401。

---

### 43. permissions 一次性获取无重试

**文件:** `src/router/index.ts` (lines 607-621)

**问题:** 路由守卫中 `authStore.user.permissions === undefined` 时获取权限（router/index.ts:609）。失败时 catch 块将其设为 `null`（line 617），此后 `permissions !== undefined` 永远为真，不再重试。用户被永久锁定在权限功能之外，直到重新登录。

**修复方案:**

用独立的 `permissionsFetched` 标志替代 `undefined` vs `null` 的语义：

```diff
+ let permissionsFetchAttempted = false
+ let permissionsFetchError = false

  // 路由守卫中
- if (permissions.value === undefined) {
+ if (!permissionsFetchAttempted || permissionsFetchError) {
    try {
+     permissionsFetchAttempted = true
+     permissionsFetchError = false
      await authStore.fetchPermissions()
    } catch {
+     permissionsFetchError = true
      // 允许导航继续，但权限功能受限
    }
  }
```

集群切换时重置（新增代码，需要在路由守卫附近添加 store import）：
```ts
import { useClusterStore } from '@/stores/cluster'

const clusterStore = useClusterStore()
watch(() => clusterStore.clusterId, () => {
  permissionsFetchAttempted = false
  permissionsFetchError = false
})
```

**⚠️ 两个变量声明在 `src/router/index.ts` 模块顶层。** JavaScript 单线程事件循环保证路由守卫和 watcher 不会并发执行，无需加锁。

**验证方法:** 模拟 `fetchPermissions` 失败（DevTools 拦截），刷新页面后应自动重试。

---

### 44. Route props: true 使用不一致

**文件:** `src/router/index.ts`

**问题:** 部分详情路由有 `props: true`（DeploymentDetail line 94, StatefulSetDetail line 113），但其他没有（PodDetail line 73）。混合使用 `route.params` 和 props 传参方式不一致。

**修复方案:**

统一为一种方式。推荐移除 `props: true`，统一使用 `route.params`（因为大部分已经这样做了）：

```diff
  {
    path: '/workloads/deployments/:namespace/:name',
    name: 'DeploymentDetail',
    component: () => import('@/views/workload/DeploymentDetail.vue'),
-   props: true,
    meta: { title: 'Deployment 详情', parent: 'DeploymentList' },
  },
```

对应组件中将 `defineProps` 改为从 `useRoute()` 读取 params。

**验证方法:** 所有详情页正常加载，参数传递正确。

---

### 45. 生产环境缺少 source map

**文件:** `vite.config.ts`

**问题:** Vite 生产构建默认不生成 source map。线上错误无法定位到源码行号。

**修复方案:**

```diff
  export default defineConfig({
+   build: {
+     sourcemap: 'hidden', // 生成 source map 但不内联引用，防止 casual 访问
+   },
  })
```

**⚠️ `hidden` 模式不会在 JS 文件末尾添加 `//# sourceMappingURL=` 注释。** 需要配合错误监控工具手动上传 `.map` 文件。如果不需要隐藏，用 `sourcemap: true`。

**⚠️ 与 #46 共用 `build` 配置块，合并时不要创建重复的 `build` 对象。** 合并后的结果：
```ts
build: {
  sourcemap: 'hidden',
  target: 'es2020',
},
```

**验证方法:** `npm run build` 后 `dist/assets/` 中有 `.map` 文件。

---

### 46. 未指定 build.target

**文件:** `vite.config.ts`

**问题:** Vite 默认 `build.target: 'modules'`（ES2015+），但代码使用了可选链（ES2020）和指数运算符（ES2016）。当前默认值够用，但没有显式声明，Vite 主版本升级可能改变默认值。

**修复方案:**

```diff
  export default defineConfig({
+   build: {
+     target: 'es2020',
+   },
  })
```

**⚠️ 与 #45 共用 `build` 配置块，合并到已有 `build` 对象中。**

**验证方法:** `npm run build` 无错误，产物中包含可选链语法。

---

### 47. 可测试性架构问题

**文件:** 全局

**问题:** 当前架构难以编写单元测试：
1. composable 内部调用 `useRouter()` / `useNamespaceStore()`，无法脱离 Vue 上下文测试
2. API 层使用单例 axios 实例，无法 mock 单个接口
3. 组件直接 import API 函数，无依赖注入点
4. store action 有跨 store 耦合（`clusterStore.fetchClusters()` 调用 `namespaceStore.clearCache()`）

**修复方案:**

渐进式改进，不需要一次性重构：

**第一步 — composable 接受依赖注入：**
```ts
// 现在
export function useResourceList(options) {
  const router = useRouter() // 硬编码
}

// 改为
export function useResourceList(options, { router = useRouter() } = {}) {
  // 可测试时传入 mock router
}
```

**第二步 — API 工厂函数：**
```ts
// src/api/createClient.ts
export function createApiClient(baseRequest = request) {
  return {
    getPods: (ns: string) => baseRequest.get(`/k8s/pod/list/${ns}`),
    // ...
  }
}

// 生产使用
export const api = createApiClient()

// 测试使用
const mockRequest = { get: vi.fn().mockResolvedValue(...) }
const testApi = createApiClient(mockRequest)
```

**第三步 — store 测试用 `createTestingPinia`：**
```ts
import { createTestingPinia } from '@pinia/testing'

const pinia = createTestingPinia({ createSpy: vi.fn })
```

**验证方法:** 为 `useResourceDelete` 编写一个测试用例，验证 `markPendingDelete` 和自动清理逻辑。

---

### 48. DashboardView getComputedStyle 每次图表更新都调用

**文件:** `src/views/dashboard/DashboardView.vue`

**问题:** `tk()`、`threshColor()`、`readyColor()` 等函数在每次图表更新时调用 `getComputedStyle(document.documentElement)` 读取 CSS 变量。`getComputedStyle` 强制浏览器进行样式重计算。每次数据刷新涉及 5 个图表 × 4+ 次调用 = 20+ 次样式重计算。

**修复方案:**

缓存 CSS 变量值，仅在主题变化时刷新：

```ts
let cachedTokens: Record<string, string> | null = null

function getTokens() {
  if (!cachedTokens) {
    const style = getComputedStyle(document.documentElement)
    cachedTokens = {
      textColor: style.getPropertyValue('--el-text-color-primary'),
      bgColor: style.getPropertyValue('--el-bg-color'),
      // ... 其他需要的变量
    }
  }
  return cachedTokens
}

// MutationObserver 检测主题变化时清空缓存
const observer = new MutationObserver(() => {
  cachedTokens = null // 下次 getTokens() 时重新读取
  updateAllCharts()
})
// Element Plus 暗色主题通过 <html> 上的 class="dark" 切换。
// 现有代码只监听 data-theme（theme-switcher.ts 同时设置两者，所以目前能工作）。
// 加入 'class' 到 attributeFilter 使检测更健壮，防止直接操作 class 时遗漏。
observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class', 'data-theme'] })

// onBeforeUnmount 中清理
onBeforeUnmount(() => observer.disconnect())
```

**验证方法:** 在 Performance 面板中录制 Dashboard 数据刷新，对比 `getComputedStyle` 调用次数从 20+ 降至 0。

---

## 执行顺序建议（更新）

```
第一阶段（2-3天）：止血
  → #1  ResourceQuotaDetail ECharts 泄漏（复用实例，不要每次 dispose）
  → #2  UserList 定时器清理
  → #25 PVDetail/HPADetail/NetworkPolicyDetail resize 泄漏（HPADetail/NetworkPolicyDetail 直接用 useResizable；PVDetail 需扩展 composable）
  → #9  表单离开保护（beforeunload 在 markDirty 时注册）
  → #10 详情页 404 处理
  → #27 YamlEditor 全屏暗色背景

第二阶段（1周）：体验提升
  → #29 Router scrollBehavior（5 行代码）
  → #38 搜索清空立即生效
  → #12 URL 同步筛选（合并 watcher）
  → #15 keep-alive（每个缓存组件加 defineOptions）
  → #11 骨架屏
  → #13 关键错误通知升级
  → #14 空状态统一
  → #34 Terminal/Log cluster 合成对象修复

第三阶段（2-3天）：打包优化（严格按顺序）
  → #4  ECharts 按需导入（必须先于 #7）
  → #3  Element Plus 按需导入（注意 locale 配置）
  → #5  图标按需注册（unplugin 自动处理）
  → #6  删除 vue-echarts
  → #28 删除 vis-network / vis-data
  → #7  Vite chunk 拆分（依赖 #3、#4 完成）
  → #8  Monaco 按需加载（验证导出方式）
  → #45 生产 source map
  → #46 显式 build.target

第四阶段（2-3天）：架构改进
  → #20 错误边界组件（retryCount 强制 remount）
  → #23 Monaco Editor 清理
  → #24 YAML 校验
  → #36 全局 unhandledrejection 处理
  → #37 ContainerConfigForm emit 替代直接修改 prop
  → #42 logout 调用后端 API
  → #43 permissions 获取失败重试
  → #44 Route props: true 统一

第五阶段（1周）：性能与架构
  → #30 Terminal/Log 选择器去重（提取 composable）
  → #31 AuditLog 分页（需后端配合）
  → #32 WorkloadForm YAML 计算优化
  → #35 useAutoRefresh 错误退避
  → #39 API 类型安全解包（渐进迁移）
  → #40 cluster store deep watch 移除
  → #41 localStorage 版本控制
  → #48 Dashboard getComputedStyle 缓存

第六阶段（持续）：长期投资
  → #17 useResourceList 拆分（Set 用 shallowRef）
  → #18 useClusterGraph 移动
  → #19 路由拆分
  → #21 i18n 保持现状
  → #16 键盘快捷键（加 enabled 控制）
  → #22 触摸支持
  → #47 可测试性改进（依赖注入）
```
