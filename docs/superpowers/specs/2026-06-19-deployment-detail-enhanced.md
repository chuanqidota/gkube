# Deployment 详情页增强设计

**日期**: 2026-06-19  
**状态**: 草稿

## 1. 概述

### 1.1 目标

在当前分栏布局基础上，增加 Kuboard 风格的功能元素：概览区、容器详情、标签/注解显示。

### 1.2 设计原则

- **保持分栏布局**: 左侧 RS 列表 + 右侧 Events/Pods
- **增加信息密度**: 在头部区域增加概览信息
- **增强 Pod 信息**: 显示容器详情和标签/注解

## 2. 页面布局

### 2.1 整体结构

```
┌─────────────────────────────────────────────────────────────────────┐
│  Deployment: nginx-deployment                                       │
│  [Scale] [Restart] [Rollback] [YAML] [Back to List]                │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │ 概览区                                                          ││
│  │ 副本: 3/3 | 可用: 3 | 更新: 3 | 策略: RollingUpdate            ││
│  │ Labels: app=nginx, env=prod                                     ││
│  │ Selector: app=nginx                                             ││
│  └─────────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┬──────────────────────────────────────────────────┐│
│  │              │  Events                                          ││
│  │  ReplicaSet  │  [Events table]                                  ││
│  │  List        │                                                  ││
│  │              │  Pods for rs-xxx (3)                             ││
│  │              │  [Pod table with container details]              ││
│  └──────────────┴──────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 概览区设计（紧凑型）

**位置**: 页面头部下方，主内容区上方

**内容**:
- **副本信息**: Ready/Desired、Available、Updated
- **更新策略**: RollingUpdate 或 Recreate
- **Labels**: 显示 Deployment 的 Labels（使用 el-tag）
- **Selector**: 显示 Pod Selector（使用 el-tag）

**样式**:
- 使用 `el-descriptions` 组件，单行显示
- Labels 和 Selector 使用 `el-tag` 显示
- 背景色：`var(--el-fill-color-lighter)`
- 高度紧凑，不占用过多空间

### 2.3 容器详情设计

**位置**: Pod 列表的展开行或详情列

**方案 A: 展开行（推荐）**
- 点击 Pod 行展开显示容器详情
- 显示：容器名称、镜像、端口、资源限制、环境变量

**方案 B: 详情列**
- 在 Pod 表格中增加 "Containers" 列
- 显示容器数量和主要镜像

### 2.4 标签/注解设计

**位置**: 概览区内

**内容**:
- **Labels**: 显示 Deployment 的 Labels
- **Selector**: 显示 Pod Selector
- **Annotations**: 可选，显示重要注解（如 revision）

## 3. 组件变更

### 3.1 DeploymentDetail.vue

**变更**:
- 新增概览区组件
- 修改 Pod 列表支持容器详情展开
- 调整布局结构

### 3.2 PodListPanel.vue

**变更**:
- 新增展开行显示容器详情
- 显示容器名称、镜像、端口、资源限制

## 4. 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                        DeploymentDetail                         │
├─────────────────────────────────────────────────────────────────┤
│  1. onMounted:                                                  │
│     - fetchDeploymentDetail() → 基本信息 + Labels + Selector   │
│     - fetchReplicaSets() → ReplicaSet 列表                     │
│     - fetchEvents() → 事件列表                                  │
│                                                                 │
│  2. 概览区显示:                                                  │
│     - deployment.replicas, ready, available, updated            │
│     - deployment.strategy                                       │
│     - deployment.labels, deployment.selector                    │
│                                                                 │
│  3. Pod 容器详情:                                                │
│     - 从 Pod 的 spec.containers 获取                            │
│     - 显示容器名称、镜像、端口、资源限制                          │
└─────────────────────────────────────────────────────────────────┘
```

## 5. 实施计划

### 5.1 第一阶段：增加概览区

1. 在 DeploymentDetail.vue 中新增概览区
2. 显示副本信息、更新策略、Labels、Selector

### 5.2 第二阶段：增加容器详情

1. 修改 PodListPanel.vue 支持展开行
2. 显示容器名称、镜像、端口、资源限制

### 5.3 第三阶段：测试验证

1. 验证概览区显示正确
2. 验证容器详情展开功能
3. 验证响应式布局
