# 节点列表卡片视图 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在节点列表页 `NodeList.vue` 增加卡片视图，工具栏提供表格/卡片切换（默认卡片、localStorage 持久化），卡片复用现有进度条逻辑与集群卡片骨架。

**Architecture:** 新增一个无业务依赖的共享组件 `ViewModeToggle.vue`（v-model + 可选 localStorage 持久化）。`NodeList.vue` 增加 `viewMode` 页面级 ref，主体用 `v-if` 在现有 `<el-table>`（不改）与新 `el-row`/`el-col` 卡片网格间二选一，共用同一份 `filteredList` 与操作函数。

**Tech Stack:** Vue 3 + TypeScript + Element Plus 2.14 + @element-plus/icons-vue 2.3 + Vite。CSS 变量沿用项目现有 `--gk-color-*`。

## Global Constraints

- 项目无自动化测试框架（无 `*.test.ts`/`*.spec.ts`，CLAUDE.md 明确）。每个任务的「测试周期」= `npx vue-tsc --noEmit` 类型检查 + `npm run dev` 浏览器手测，不伪造测试代码。
- TypeScript 开启 `noUnusedLocals` / `noUnusedParameters` —— 引入的 import 必须用到。
- 复用现有 `usagePercent` / `progressColor`（`frontend/src/utils/helpers.ts`），不新增进度条逻辑。
- 卡片 CSS 类与口径镜像 `frontend/src/views/cluster/ClusterList.vue` 的 `.cluster-card` / `.cluster-detail` / `.cluster-footer` 约定。
- 颜色用 CSS 变量：`--gk-color-text-secondary` / `--gk-color-text-primary` / `--gk-color-border-light` / `--gk-color-danger`。
- 口径文案：CPU/内存进度条标题为「CPU（请求/容量）」「内存（请求/容量）」「Pods（请求/容量）」——非真实负载。

---

### Task 1: ViewModeToggle 共享组件

**Files:**
- Create: `frontend/src/components/ViewModeToggle.vue`

**Interfaces:**
- Consumes: 无（纯组件）。
- Produces: `ViewModeToggle` 组件，props `modelValue: 'card'|'table'`、`storageKey?: string`；emit `update:modelValue`。供 Task 2 引入。

- [ ] **Step 1: 创建组件文件**

写入 `frontend/src/components/ViewModeToggle.vue`：

```vue
<script setup lang="ts">
import { Grid, List } from '@element-plus/icons-vue'

type ViewMode = 'card' | 'table'

const props = defineProps<{
  modelValue: ViewMode
  storageKey?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [mode: ViewMode] }>()

function set(mode: ViewMode) {
  emit('update:modelValue', mode)
  if (props.storageKey) {
    localStorage.setItem(props.storageKey, mode)
  }
}
</script>

<template>
  <el-button-group>
    <el-button
      :type="modelValue === 'card' ? 'primary' : ''"
      size="small"
      @click="set('card')"
    >
      <el-icon><Grid /></el-icon> 卡片
    </el-button>
    <el-button
      :type="modelValue === 'table' ? 'primary' : ''"
      size="small"
      @click="set('table')"
    >
      <el-icon><List /></el-icon> 表格
    </el-button>
  </el-button-group>
</template>
```

- [ ] **Step 2: 类型检查**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: 无新增错误（`Grid`、`List` 均在 @element-plus/icons-vue 2.3 中存在，已确认 `dist/types/components/grid.vue.d.ts` 与 `list.vue.d.ts`）。`props`、`emit` 均被使用，不触发 `noUnusedLocals`。

- [ ] **Step 3: 浏览器冒烟验证（可选，本任务无页面接入，可跳到 Task 2 一起验）**

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ViewModeToggle.vue
git commit -m "feat(node): add ViewModeToggle shared component"
```

---

### Task 2: NodeList 接入视图切换 + 卡片网格

**Files:**
- Modify: `frontend/src/views/node/NodeList.vue`（script 区 1-11 行 import、55 行附近新增 `viewMode`；template 工具栏 `#extra`、主体表格包 `v-if` + 新增卡片 `v-else`；style 新增卡片 CSS）
- Read-only 参考: `frontend/src/views/cluster/ClusterList.vue:276-286`（卡片 CSS 约定）

**Interfaces:**
- Consumes: Task 1 的 `ViewModeToggle` 组件；现有 `usagePercent`/`progressColor`/`statusType`/`fmtCpu`/`fmtMem`；现有操作函数 `handleViewYaml`/`handleCordon`/`handleTaints`/`handleLabels`/`handleDrain`/`handleDelete`/`handleDetail`。
- Produces: 节点列表页支持卡片/表格切换，默认卡片，`localStorage` key `gkube.node.viewMode` 持久化。

- [ ] **Step 1: script 区新增 import 与 viewMode 状态**

在 `NodeList.vue` 第 5 行的 icons import 后追加 `ArrowDown`：

```ts
import { Delete, Plus, ArrowDown } from '@element-plus/icons-vue'
```

在第 11 行 `ResourceListToolbar` import 后追加：

```ts
import ViewModeToggle from '@/components/ViewModeToggle.vue'
```

在 `const router = useRouter()`（第 13 行）之前或之后的状态声明区，新增 `viewMode`：

```ts
const viewMode = ref<'card' | 'table'>(
  (localStorage.getItem('gkube.node.viewMode') as 'card' | 'table') || 'card'
)
```

- [ ] **Step 2: 工具栏 #extra 内放入 ViewModeToggle**

找到 template 中 `AutoRefreshToolbar` 所在的 `<template #extra>` 块（约 196-207 行），在 `AutoRefreshToolbar` 后面追加 `ViewModeToggle`：

```vue
<template #extra>
  <AutoRefreshToolbar
    :is-running="isRunning"
    :countdown="countdown"
    :current-interval="currentInterval"
    :available-intervals="availableIntervals"
    :loading="loading"
    @refresh="manualRefresh()"
    @toggle="toggle()"
    @interval-change="setIntervalOption"
  />
  <ViewModeToggle v-model="viewMode" storage-key="gkube.node.viewMode" />
</template>
```

- [ ] **Step 3: 给现有 el-table 包 v-if，新增卡片网格 v-else**

找到 `<el-card shadow="never" class="table-card">` 内的 `<el-table ...>`（约 214 行起），给 `<el-table>` 标签加 `v-if="viewMode === 'table'"`。**表格内部列定义一字不改。**

在 `</el-table>` 之后、`</el-card>` 之前，插入卡片网格：

```vue
<el-row v-else :gutter="16">
  <el-col v-for="node in filteredList" :key="node.name" :xs="24" :sm="12" :md="8" style="margin-bottom: 16px;">
    <el-card shadow="hover" class="node-card">
      <template #header>
        <div class="node-header">
          <el-button link type="primary" @click="handleDetail(node)">{{ node.name }}</el-button>
          <div class="node-header-tags">
            <el-tag :type="statusType(node)" size="small" effect="dark">{{ node.status || 'Unknown' }}</el-tag>
            <el-tag v-if="node.unschedulable" type="warning" size="small">已封锁</el-tag>
          </div>
        </div>
      </template>

      <div class="node-body">
        <div class="node-detail"><span class="label">IP</span><span class="value">{{ node.internal_ip || '-' }}</span></div>
        <div class="node-detail"><span class="label">角色</span><span class="value">{{ node.roles || '-' }}</span></div>
        <div class="node-detail"><span class="label">版本</span><span class="value">{{ node.version || '-' }}</span></div>
        <div class="node-detail"><span class="label">年龄</span><span class="value">{{ node.age || '-' }}</span></div>
      </div>

      <div class="node-usage">
        <div class="usage-item">
          <div class="usage-title">CPU（请求/容量）</div>
          <el-progress :percentage="usagePercent(node.cpu_used, node.cpu_total)" :color="progressColor(usagePercent(node.cpu_used, node.cpu_total))" :stroke-width="14" :text-inside="true" />
          <div class="res-caption">{{ fmtCpu(node.cpu_used) }} / {{ fmtCpu(node.cpu_total) }} 核</div>
        </div>
        <div class="usage-item">
          <div class="usage-title">内存（请求/容量）</div>
          <el-progress :percentage="usagePercent(node.mem_used, node.mem_total)" :color="progressColor(usagePercent(node.mem_used, node.mem_total))" :stroke-width="14" :text-inside="true" />
          <div class="res-caption">{{ fmtMem(node.mem_used) }} / {{ fmtMem(node.mem_total) }} GiB</div>
        </div>
        <div class="usage-item">
          <div class="usage-title">Pods（请求/容量）</div>
          <el-progress :percentage="usagePercent(node.pod_count, node.pod_total)" :color="progressColor(usagePercent(node.pod_count, node.pod_total))" :stroke-width="14" :text-inside="true" />
          <div class="res-caption">{{ node.pod_count || 0 }} / {{ node.pod_total || 0 }}</div>
        </div>
      </div>

      <div class="node-footer">
        <div class="node-footer-main">
          <el-button size="small" @click="handleViewYaml(node)">YAML</el-button>
          <el-button size="small" :type="node.unschedulable ? 'success' : 'warning'" @click="handleCordon(node)">
            {{ node.unschedulable ? '解除封锁' : '封锁' }}
          </el-button>
        </div>
        <el-dropdown trigger="click" @command="(cmd: string) => handleMoreCommand(cmd, node)">
          <el-button size="small">更多<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="taints">污点</el-dropdown-item>
              <el-dropdown-item command="labels">标签</el-dropdown-item>
              <el-dropdown-item command="drain">驱逐</el-dropdown-item>
              <el-dropdown-item command="delete" divided class="danger-item">删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-card>
  </el-col>
</el-row>
```

- [ ] **Step 4: 新增 handleMoreCommand 派发函数**

在 script 区操作函数区（`handleDelete` 之后，约 180 行）新增：

```ts
function handleMoreCommand(cmd: string, row: any) {
  switch (cmd) {
    case 'taints': handleTaints(row); break
    case 'labels': handleLabels(row); break
    case 'drain': handleDrain(row); break
    case 'delete': handleDelete(row); break
  }
}
```

- [ ] **Step 5: 新增卡片 CSS**

在 `<style scoped>` 块（含 `.page-container`、`.table-card`、`.res-caption`）内追加：

```css
.node-card { height: 100%; }
.node-header { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.node-header-tags { display: flex; align-items: center; gap: 6px; }
.node-body { margin-bottom: 12px; }
.node-detail { display: flex; margin-bottom: 8px; }
.node-detail .label { color: var(--gk-color-text-secondary); width: 70px; flex-shrink: 0; }
.node-detail .value { color: var(--gk-color-text-primary); }
.node-usage { margin-bottom: 12px; }
.usage-item { margin-bottom: 12px; }
.usage-title { font-size: 12px; color: var(--gk-color-text-secondary); margin-bottom: 4px; }
.node-footer { display: flex; justify-content: space-between; align-items: center; border-top: 1px solid var(--gk-color-border-light); padding-top: 12px; }
.node-footer-main { display: flex; gap: 8px; }
.danger-item { color: var(--gk-color-danger); }
```

- [ ] **Step 6: 类型检查**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: 通过。确认 `ArrowDown`、`ViewModeToggle`、`viewMode`、`handleMoreCommand` 均被使用，无 `noUnusedLocals` 报错。`@command` 回调参数标注 `string` 避免隐式 any。

- [ ] **Step 7: 浏览器端到端验证**

Run: `cd frontend && npm run dev`（后端 `cd backend && go run main.go` 起服务，提供 `/v1/k8s/cluster/nodes`）

打开节点列表页，核对：
- 默认进卡片视图；点「表格」切到现有表格（含进度条列），点「卡片」切回；刷新页面后视图模式保持。
- 卡片：状态 tag 颜色正确；封锁节点显示「已封锁」tag；4 行信息 + 3 条进度条数值与表格列一致；颜色阈值 <70 主色 / 70-90 警告 / ≥90 危险。
- 卡片操作：YAML、封锁/解除封锁 直接可用；「更多」下拉 污点/标签/驱逐/删除 触发与表格一致的弹窗。
- 搜索过滤、自动刷新在两种视图下都生效。
- 缩放窗口：卡片在 1/2/3 列间切换不破版。

- [ ] **Step 8: 生产构建**

Run: `cd frontend && npm run build`
Expected: vue-tsc + Vite 构建通过。

- [ ] **Step 9: Commit**

```bash
git add frontend/src/views/node/NodeList.vue
git commit -m "feat(node): add card view with table/card toggle on node list"
```

---

## 验证汇总

- 类型检查：`cd frontend && npx vue-tsc --noEmit` 通过。
- 生产构建：`cd frontend && npm run build` 通过。
- 端到端：`npm run dev` + `go run main.go`，按 Task 2 Step 7 清单逐项核对。
- 无后端改动（节点卡片用到的 `cpu_used/cpu_total/mem_used/mem_total/pod_total/pod_count` 字段上一轮已加好）。
