# 列表页面工具栏按钮设计方案

## 概述

重新设计无状态负载（Deployment）列表页面的工具栏布局，解决当前筛选器和操作按钮混在一起、视觉层次不清晰的问题。采用单行左右分区布局，参考详情页的连接式按钮组设计语言。

## 背景

### 当前问题

当前列表页面工具栏将所有元素排列在同一行的 flex 容器中：

```
[搜索] [命名空间] [自动刷新] [创建] [删除] [总计]
```

- 筛选器和操作按钮混在一起，没有视觉分隔
- 按钮之间间距相同，主次不分明
- 不符合详情页已建立的设计语言

### 参考

Deployment 详情页（`DeploymentDetail.vue`）使用了成熟的工具栏设计：
- 左侧资源信息，右侧操作按钮
- 连接式按钮组（共享边框，无圆角间隔）
- 竖线分隔符区分主要操作和辅助工具

## 设计方案

### 布局结构

```
┌──────────────────────────────────────────────────────────────────────────┐
│  [🔍 搜索] [命名空间 ▾]  总计: 128   │   [创建] [批量删除]  │  [⟳ 自动刷新] │
└──────────────────────────────────────────────────────────────────────────┘
│← ── 左侧筛选区 ── →│← 分隔 →│← ── 主要操作 ── →│← 分隔 →│← 辅助工具 →│
```

### 区域定义

#### 左侧筛选区

- **搜索输入框**：`el-input`，宽度 220px，带 Search 图标前缀，可清空
- **命名空间选择**：`el-select`，宽度 180px，placeholder "所有命名空间"，可清空
- **总计数**：`span.total-count`，显示 "总计: N"，使用次要文字色（`var(--el-text-color-secondary)`），字号 13px
- 元素间距：12px

#### 竖线分隔符

- `div.action-divider`
- 宽度 1px，高度 20px
- 背景色 `var(--el-border-color-lighter)`
- 水平间距 4px

#### 主要操作组

采用**连接式按钮组**风格，按钮紧密排列，共享边框：

- **创建按钮**：`el-button type="success"`，Plus 图标，文字 "创建"
- **批量删除按钮**：`el-button type="danger"`，Delete 图标，文字 "删除 (N)"，未选中时 `disabled`

连接式按钮组 CSS：
```css
.action-group {
  display: inline-flex;
}

.action-group .el-button {
  border-radius: 0;
  margin-left: -1px;
}

.action-group .el-button:first-child {
  border-radius: 4px 0 0 4px;
  margin-left: 0;
}

.action-group .el-button:last-child {
  border-radius: 0 4px 4px 0;
}
```

#### 辅助工具区

- 保持现有 `AutoRefreshToolbar` 组件不变
- 包含：手动刷新按钮、自动刷新开关、刷新间隔选择器

### 容器样式

```css
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
```

### 变更范围

| 文件 | 变更内容 |
|------|---------|
| `frontend/src/views/workload/DeploymentList.vue` | 重构 filter-bar 布局，添加分隔符和 action-group |
| `frontend/src/views/workload/StatefulSetList.vue` | 同步应用新布局 |
| `frontend/src/views/workload/DaemonSetList.vue` | 同步应用新布局 |
| `frontend/src/views/workload/JobList.vue` | 同步应用新布局 |
| `frontend/src/views/workload/CronJobList.vue` | 同步应用新布局 |
| `frontend/src/views/workload/PodList.vue` | 同步应用新布局（无创建按钮） |
| `frontend/src/views/workload/ReplicaSetList.vue` | 同步应用新布局（无创建按钮） |

### 不变的部分

- `useResourceList` composable：不变
- `useAutoRefresh` composable：不变
- `AutoRefreshToolbar` 组件：不变
- 表格列定义：不变
- YAML 编辑器：不变

## 与其他页面的一致性

此设计应统一应用到所有资源列表页面，包括：
- 工作负载：Deployment、StatefulSet、DaemonSet、Job、CronJob、Pod、ReplicaSet
- 网络：Service、Ingress
- 存储：PersistentVolumeClaim
- 配置：ConfigMap、Secret

各页面的按钮数量可能不同（如 Pod 无创建按钮），但布局结构和视觉语言保持一致。
