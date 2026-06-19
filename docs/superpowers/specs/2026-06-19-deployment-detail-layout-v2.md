# Deployment 详情页布局设计 v2

**日期**: 2026-06-19  
**状态**: 草稿

## 1. 概述

### 1.1 目标

重新设计 Deployment 详情页，移除所有 Tab，采用单页面布局：左侧显示 ReplicaSet 列表，右侧上方显示 Events，右侧下方显示 Pods。YAML 编辑通过弹窗实现。

### 1.2 设计原则

- **无 Tab 设计**: 所有信息在一个页面内展示，无需切换
- **信息密度**: 重要信息（Events、Pods）直接可见
- **操作便捷**: YAML 编辑、终端跳转等操作通过按钮触发

## 2. 页面布局

### 2.1 整体结构

```
┌─────────────────────────────────────────────────────────────────────┐
│  Deployment: nginx-deployment                                       │
│  [Scale] [Restart] [Rollback] [YAML] [Back to List]                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┬──────────────────────────────────────────────────┐│
│  │              │  Events                                          ││
│  │  ReplicaSet  │  ┌────────────────────────────────────────────┐ ││
│  │  List        │  │ Type │ Reason │ Message │ Last Seen        │ ││
│  │              │  ├──────┼────────┼─────────┼──────────────────┤ ││
│  │  ▼ rs-xxx(v3)│  │ Info │ Create │ Created │ 2h ago           │ ││
│  │    nginx:1.25│  │ Info │ Scale  │ Scaled  │ 1h ago           │ ││
│  │    2h ago    │  └──────┴────────┴─────────┴──────────────────┘ ││
│  │    3/3       │                                                  ││
│  │    [Current] │  Pods for rs-xxx-v3 (3)                         ││
│  │              │  ┌────────────────────────────────────────────┐ ││
│  │  ▶ rs-xxx(v2)│  │ Name │ Status │ Restarts │ Age │ Actions │ ││
│  │    nginx:1.24│  ├──────┼────────┼──────────┼─────┼─────────┤ ││
│  │    1d ago    │  │ pod1 │ Running│ 0        │ 2h  │ Logs Exec│ ││
│  │    0/0       │  │ pod2 │ Running│ 0        │ 2h  │ Logs Exec│ ││
│  │    [Rollback]│  │ pod3 │ Running│ 0        │ 2h  │ Logs Exec│ ││
│  │              │  └──────┴────────┴──────────┴─────┴─────────┘ ││
│  └──────────────┴──────────────────────────────────────────────────┘│
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 头部区域

**左侧**:
- Deployment 名称（大字体显示）

**右侧按钮**:
- Scale（主色调）- 打开扩缩容对话框
- Restart（警告色）- 二次确认后重启
- Rollback（危险色）- 打开回滚对话框
- YAML（默认色）- 打开 YAML 编辑弹窗
- Back to List - 返回列表页

### 2.3 左侧面板 - ReplicaSet 列表

**复用组件**: `ReplicaSetPanel.vue`

**内容**:
- 每个 RS 显示：名称、镜像、创建时间、副本数、状态标签
- 当前活跃 RS 高亮显示
- 历史 RS 显示 Rollback 按钮
- 点击 RS 时，右侧面板切换为该 RS 的 Events 和 Pods

### 2.4 右侧面板 - Events 区域

**位置**: 右侧面板上方

**内容**:
- 表格显示：Type（颜色标签）、Reason、Message、Last Seen
- 最大高度 250px，超出滚动
- 无事件时显示空状态提示

### 2.5 右侧面板 - Pods 区域

**位置**: 右侧面板下方

**复用组件**: `PodListPanel.vue`

**内容**:
- 表格显示：Name（可点击跳转详情）、Status（颜色标签）、Restarts、Age、Node、Actions
- Actions 列：
  - Logs 按钮 - 跳转到日志页面（新窗口）
  - Exec 按钮 - 跳转到终端页面（新窗口）
  - Delete 按钮 - 二次确认后删除

## 3. YAML 编辑弹窗

### 3.1 触发方式

点击头部的 "YAML" 按钮打开弹窗

### 3.2 弹窗设计

```
┌─────────────────────────────────────────────────────────────────────┐
│  YAML Editor                                    [Format] [Copy]     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  apiVersion: apps/v1                                                │
│  kind: Deployment                                                   │
│  metadata:                                                          │
│    name: nginx-deployment                                           │
│    namespace: default                                               │
│  spec:                                                              │
│    replicas: 3                                                      │
│    ...                                                              │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                              [Cancel] [Save]        │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.3 功能

- 默认只读模式，显示当前 YAML
- 点击 "Edit" 按钮切换到编辑模式
- 编辑模式下显示 Save 和 Cancel 按钮
- 保存后刷新页面数据

## 4. 终端/日志跳转

### 4.1 跳转方式

- Logs 按钮：新窗口打开 `/logs?namespace=xxx&pod=xxx`
- Exec 按钮：新窗口打开 `/terminal?namespace=xxx&pod=xxx`

### 4.2 目标页面适配

TerminalView 和 LogView 页面需要：
- 读取 `route.query` 中的 `namespace` 和 `pod` 参数
- 自动填充对应的下拉框

## 5. 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                        DeploymentDetail                         │
├─────────────────────────────────────────────────────────────────┤
│  1. onMounted:                                                  │
│     - fetchDeploymentDetail() → 基本信息                        │
│     - fetchReplicaSets() → ReplicaSet 列表                     │
│     - fetchEvents() → 事件列表                                  │
│                                                                 │
│  2. 用户点击左侧 RS:                                            │
│     - selectedReplicaset = clickedRS                            │
│     - fetchReplicasetPods(selectedReplicaset) → Pod 列表       │
│                                                                 │
│  3. 用户点击 YAML 按钮:                                         │
│     - fetchYaml() → 加载 YAML 内容                              │
│     - yamlDialogVisible = true                                  │
│                                                                 │
│  4. 用户点击 Pod 的 Logs/Exec:                                  │
│     - 新窗口打开 /logs 或 /terminal 路由                        │
└─────────────────────────────────────────────────────────────────┘
```

## 6. 组件变更

### 6.1 DeploymentDetail.vue

**主要变更**:
- 移除 `el-tabs` 组件
- 移除 Info、Events、YAML Tab 内容
- 新增 YAML 编辑弹窗
- 重组布局：左侧面板 + 右侧面板（Events + Pods）
- 数据加载改为 onMounted 时全部加载

### 6.2 TerminalView.vue

**变更**:
- 读取 `route.query` 中的 `namespace`、`pod` 参数
- 自动填充对应的下拉框

### 6.3 LogView.vue

**变更**:
- 读取 `route.query` 中的 `namespace`、`pod` 参数
- 自动填充对应的下拉框

## 7. 响应式设计

- **桌面端**（≥1200px）：左右分栏，左侧 320px，右侧自适应
- **平板端**（768px-1199px）：左右分栏，左侧 280px，右侧自适应
- **移动端**（<768px）：上下布局，RS 列表在上，Events + Pods 在下

## 8. 实施计划

### 8.1 第一阶段：重构 DeploymentDetail.vue

1. 移除 `el-tabs` 组件
2. 移除 Info、Events、YAML Tab 内容
3. 新增 YAML 编辑弹窗
4. 重组布局
5. 修改数据加载逻辑

### 8.2 第二阶段：更新终端/日志页面

1. 更新 TerminalView.vue 读取 query 参数
2. 更新 LogView.vue 读取 query 参数

### 8.3 第三阶段：测试验证

1. 验证页面布局正确
2. 验证 YAML 编辑功能
3. 验证终端/日志跳转功能
