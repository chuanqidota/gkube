# 节点管理列表：卡片视图（表格/卡片切换）

## 背景

节点列表页 `NodeList.vue` 目前是单一表格视图（刚加入 CPU/内存/Pods 进度条列）。用户希望节点列表支持类似集群管理 `ClusterList.vue` 的卡片形态，以便在节点不多时快速纵览每台机器的整体状态（状态 + 资源进度条 + 关键信息）。

与集群不同，节点往往数量较多且存在「快速对比/找异常」的表格场景，因此不彻底替换表格，而是提供**表格/卡片视图切换**。这也为后续集群页加入同样切换、形成一致交互铺路。

## 决策摘要

| 决策 | 选择 |
|------|------|
| 视图形态 | 工具栏 toggle，表格/卡片两种都保留，默认卡片 |
| 卡片操作布局 | 主次分离：YAML + 封锁按钮平铺，其余 4 项收进「更多」下拉 |
| 卡片信息密度 | 均衡：3 条进度条 + IP/角色/版本/年龄 4 行 label-value |
| 网格列数 | 响应式 1/2/3 列（`xs=24 sm=12 md=8`） |
| 视图切换器 | 抽共享 `ViewModeToggle` 组件，带可选 `localStorage` 持久化 |

## 整体结构

`NodeList.vue` 改动遵循「不动逻辑层、只换表现层」：

- **不动**：`fetchNodes`、`useAutoRefresh`、`ResourceListToolbar`、`AutoRefreshToolbar`、所有操作弹窗（YAML drawer / 污点 / 标签 / 驱逐）、所有操作函数（`handleViewYaml` / `handleCordon` / `handleTaints` / `handleLabels` / `handleDrain` / `handleDelete`）。
- **新增**：页面级 `viewMode` 状态 + 工具栏里的 `ViewModeToggle`。
- **主体二选一**：`viewMode==='table'` 渲染现有 `<el-table>`（含已加的 CPU/内存/Pods 进度条列，不改）；`viewMode==='card'` 渲染新的 `el-row`/`el-col` 卡片网格。两种视图共用同一份 `filteredList` 和操作函数。

`viewMode` 为 `ref<'card'|'table'>`，初始值从 `localStorage` 读 `gkube.node.viewMode`，默认 `'card'`，变化时写回。

## 卡片布局

单张卡片镜像集群卡片骨架（`el-card shadow="hover"` + `height:100%`），结构：

- **Header**：节点名（`el-button link type="primary"` → `/nodes/${name}`，与现表格一致）+ 状态 `el-tag`（复用 `statusType`：Ready=success / NotReady=danger / 其余=warning）+ `unschedulable` 时额外一个 `el-tag type="warning" size="small"`「已封锁」。
- **Body 上半**：4 行 `.node-detail`（同集群卡 `.cluster-detail`：`.label` 固定宽 70px 次色 + `.value` 主色）。字段：IP、角色、版本、年龄。空值显示 `-`。
- **Body 下半**：3 条进度条，**直接复用** `NodeList.vue` 现有 `usagePercent` + `progressColor` + `el-progress`（stroke-width 14, text-inside）+ `.res-caption` 小字绝对值。每条进度条上方一行小标题（「CPU（请求/容量）」等，与表格列头口径文案一致）。不新增任何进度条逻辑。
- **Footer**（主次分离，`border-top` + `padding-top` 分隔）：
  - 左：`YAML` + `封锁/解除封锁`（`unschedulable` 时变 `type="success"` 文案「解除封锁」）。
  - 右：`el-dropdown`「更多 ▾」，含 污点 / 标签 / 驱逐 / 删除（删除项 `divided` 分隔 + 红色 `type="danger"`）。

**网格**：`el-row :gutter="16"` + `el-col :xs="24" :sm="12" :md="8"`，每列 `margin-bottom:16px`。卡片等高靠 `.node-card { height:100% }`。

CSS 类沿用集群卡约定：`.node-card` / `.node-header` / `.node-detail` / `.node-footer`，颜色用现有 CSS 变量（`--gk-color-text-secondary` / `--gk-color-text-primary` / `--gk-color-border-light`）。

## ViewModeToggle 共享组件

**新文件** `frontend/src/components/ViewModeToggle.vue`（约 30 行）：

- Props：`modelValue: 'card'|'table'`、可选 `storageKey?: string`。
- Emits：`update:modelValue`。
- 实现：`el-button-group` 两个按钮（卡片 / 表格），各带 `el-icon`（`Grid` / `List` from `@element-plus/icons-vue`）。当前模式按钮 `type="primary"`。
- 选中时若传了 `storageKey` 则自动写 `localStorage`；不传则只 emit，不持久化（便于无持久化需求的复用场景）。
- 纯展示 + v-model，无业务依赖，任何资源列表页可复用。

在 `NodeList.vue` 中：
- `const viewMode = ref<'card'|'table'>((localStorage.getItem('gkube.node.viewMode') as 'card'|'table') || 'card')`
- 工具栏 `#extra` 内、`AutoRefreshToolbar` 旁放 `<ViewModeToggle v-model="viewMode" storage-key="gkube.node.viewMode" />`。
- 主体：`<el-table v-if="viewMode==='table'">…</el-table>`，`<el-row v-else>…卡片网格…</el-row>`。

**设计要点**：视图模式是页面级状态（驱动主体 `v-if`），必须留在页面 ref 里；组件只负责渲染与通知，职责单一可复用。

## 复用清单

- 卡片骨架与 CSS 模式：`frontend/src/views/cluster/ClusterList.vue`（`.cluster-card` / `.cluster-detail` / `.cluster-footer`，191-286 行）
- 响应式断点风格：`frontend/src/views/dashboard/DashboardView.vue`（`xs/sm/md` 用法）
- 进度条逻辑（不新增）：`NodeList.vue` 现有 `usagePercent` + `progressColor` + `.res-caption`，源自 `frontend/src/utils/helpers.ts`
- 工具栏：`ResourceListToolbar`、`AutoRefreshToolbar`、`useAutoRefresh` 全部不动
- 状态 tag：复用 `NodeList.vue` 内 `statusType`（注：与 `utils/helpers.ts` 的通用 `statusType` 不同，本次不统一，沿用页面内现有函数）

## 不做（YAGNI）

- 不彻底替换表格，不引入视图自动降级逻辑。
- 不抽通用 `ResourceCard` 组件——当前只有节点卡，集群卡是现成的内联实现，YAGNI。
- 不统一 `statusType`（页面内与 helpers 各有一份）——超出本次范围。
- 不给集群页本次加 toggle——本 spec 只做节点页；集群页留作后续独立小改动，复用同一 `ViewModeToggle`。
- 卡片不做分页——节点列表现有表格也无分页（全量渲染），卡片同样全量；若未来节点数极大需虚拟化，另立 spec。

## 验证

1. 类型检查：`cd frontend && npx vue-tsc --noEmit` 通过，无 `noUnusedLocals` 报错。
2. `npm run dev` 打开节点列表页：
   - 默认进卡片视图，切换到表格视图，两种都能正确渲染同一份数据。
   - 刷新页面后视图模式保持上次选择（`localStorage` 持久化生效）。
   - 卡片：状态 tag、4 行信息、3 条进度条数值与表格列一致；封锁节点显示「已封锁」tag。
   - 卡片操作：YAML、封锁切换直接可用；「更多」下拉里 污点/标签/驱逐/删除 均触发与表格一致的弹窗。
   - 搜索过滤、自动刷新在两种视图下都生效。
   - 响应式：窗口缩放时卡片在 1/2/3 列间正确切换，不破版。
3. `npm run build`（vue-tsc + 生产构建）通过。
