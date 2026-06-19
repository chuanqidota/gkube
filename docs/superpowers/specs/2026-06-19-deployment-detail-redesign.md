# Deployment 详情页重新设计

**日期**: 2026-06-19  
**状态**: 草稿  
**作者**: Claude Code

## 1. 概述

### 1.1 目标

重新设计 Deployment 详情页，采用左右分栏布局，左侧显示 ReplicaSet 列表，右侧显示对应 Pod 列表，提供更直观的版本管理和 Pod 监控体验。

### 1.2 背景

当前 Deployment 详情页采用传统 Tab 布局（Info、Pods、Events、YAML），用户需要在不同 Tab 之间切换才能查看 ReplicaSet 和 Pod 信息。新设计将 ReplicaSet 和 Pod 信息整合到一个视图中，提升操作效率。

### 1.3 范围

- 重新设计 DeploymentDetail.vue 的「Pods」Tab，改为「Replicasets & Pods」左右分栏布局
- 新增后端 API：获取 Deployment 关联的 ReplicaSet 列表
- 保持其他 Tab（Info、Events、YAML）不变
- 保持列表页（DeploymentList.vue）和创建页（DeploymentCreate.vue）不变

## 2. 页面布局设计

### 2.1 整体结构

```
┌─────────────────────────────────────────────────────────────────────┐
│  [← Back]  Deployment: nginx-deployment    [Scale] [Restart] [⋮]  │
├─────────────────────────────────────────────────────────────────────┤
│  Info  │  Replicasets & Pods  │  Events  │  YAML                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  当前 Tab 内容                                                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 「Replicasets & Pods」Tab 布局

```
┌─────────────────────────────────────────────────────────────────────┐
│  Replicasets & Pods                                    [Expand All] │
├───────────────────────┬─────────────────────────────────────────────┤
│                       │                                             │
│  ▼ rs/nginx-xxx (v3) │  Pods (3/3 ready)                          │
│    Image: nginx:1.25  │  ┌─────────────┬─────────┬────────┬──────┐ │
│    Created: 2h ago    │  │ Name        │ Status  │ Restarts│ Age  │ │
│    Replicas: 3/3      │  ├─────────────┼─────────┼────────┼──────┤ │
│    [Current] [Rollback│  │ pod-abc-123 │ Running │ 0      │ 2h   │ │
│                       │  │ pod-abc-456 │ Running │ 0      │ 2h   │ │
│  ▶ rs/nginx-xxx (v2) │  │ pod-abc-789 │ Running │ 0      │ 2h   │ │
│    Image: nginx:1.24  │  └─────────────┴─────────┴────────┴──────┘ │
│    Created: 1d ago    │                                             │
│    Replicas: 0/0      │  [Logs] [Exec] [Delete]                    │
│    [Rollback]         │                                             │
│                       │                                             │
│  ▶ rs/nginx-xxx (v1) │                                             │
│    Image: nginx:1.23  │                                             │
│    Created: 3d ago    │                                             │
│    Replicas: 0/0      │                                             │
│    [Rollback]         │                                             │
│                       │                                             │
└───────────────────────┴─────────────────────────────────────────────┘
```

### 2.3 左侧 ReplicaSet 列表

**列信息**：
- **Revision**: 版本号（从 annotation `deployment.kubernetes.io/revision` 获取）
- **Image**: 容器镜像（显示第一个容器的镜像，截断显示）
- **Created**: 创建时间（相对时间，如 "2h ago"）
- **Replicas**: 副本数（ready/desired 格式）
- **Status**: 状态标签
  - `Current`: 当前活跃的 RS（绿色标签）
  - `Active`: 有运行中 Pod 的历史 RS（蓝色标签）
  - `Inactive`: 无运行中 Pod 的历史 RS（灰色标签）

**交互**：
- 点击 RS 行 → 右侧 Pod 列表切换为该 RS 的 Pods
- 选中的 RS 行左侧显示蓝色边框高亮
- 默认展开当前活跃的 RS
- 每个 RS 行右侧显示「Rollback」按钮（当前 RS 不显示）

### 2.4 右侧 Pod 列表

**列信息**：
- **Name**: Pod 名称（可点击跳转到 Pod 详情页）
- **Status**: 状态标签（颜色编码）
  - Running: 绿色
  - Pending: 黄色
  - Failed: 红色
  - Succeeded: 蓝色
  - Unknown: 灰色
- **Restarts**: 重启次数（超过 0 次显示警告色）
- **Age**: 创建时间（相对时间）
- **Node**: 所在节点

**行操作**：
- **Logs**: 跳转到日志查看页面（`/logs/:namespace/:podName`）
- **Exec**: 跳转到终端页面（`/terminal/:namespace/:podName`）
- **Delete**: 删除 Pod（二次确认）

**空状态**：
- 如果选中的 RS 没有 Pods，显示「No pods found」空状态提示

### 2.5 响应式设计

- **桌面端**（≥1200px）：左右分栏，左侧固定 320px，右侧自适应
- **平板端**（768px-1199px）：左右分栏，左侧固定 280px，右侧自适应
- **移动端**（<768px）：改为上下布局，RS 列表在上（可折叠），Pod 列表在下

## 3. 后端 API 设计

### 3.1 新增 API：获取 ReplicaSet 列表

**Endpoint**: `GET /k8s/deployment/replicasets`

**请求参数**:
```json
{
  "cluster_name": "my-cluster",
  "namespace": "default",
  "name": "nginx-deployment"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "items": [
      {
        "metadata": {
          "name": "nginx-deployment-7fb96c8cc6",
          "namespace": "default",
          "creationTimestamp": "2026-06-17T10:00:00Z",
          "annotations": {
            "deployment.kubernetes.io/revision": "3",
            "deployment.kubernetes.io/desired-replicas": "3"
          },
          "labels": {
            "app": "nginx",
            "pod-template-hash": "7fb96c8cc6"
          }
        },
        "spec": {
          "replicas": 3,
          "template": {
            "spec": {
              "containers": [
                {
                  "name": "nginx",
                  "image": "nginx:1.25"
                }
              ]
            }
          }
        },
        "status": {
          "readyReplicas": 3,
          "availableReplicas": 3
        }
      }
    ]
  }
}
```

**实现逻辑**:
1. 通过 Deployment 的 `spec.selector.matchLabels` 获取 label selector
2. 使用 client-go 的 `AppsV1().ReplicaSets(namespace).List()` 并传入 label selector
3. 按 revision annotation 降序排序（最新版本在前）
4. 标记当前活跃的 RS（`metadata.ownerReferences` 中包含当前 Deployment）

### 3.2 现有 API 复用

**Pod 列表**: 复用现有 `GET /k8s/deployment/pods` API，传入 Deployment 名称获取关联的 Pods。

**Rollback**: 复用现有 `POST /k8s/deployment/rollback` API。

## 4. 前端组件设计

### 4.1 新增组件

**ReplicaSetPanel.vue**
- 路径: `frontend/src/components/ReplicaSetPanel.vue`
- 职责: 左侧 ReplicaSet 列表展示
- Props:
  - `replicasets`: ReplicaSet 数组
  - `currentRevision`: 当前活跃的 revision
  - `loading`: 加载状态
- Events:
  - `@select`: 选中 RS 时触发，传递 RS 对象

**PodListPanel.vue**
- 路径: `frontend/src/components/PodListPanel.vue`
- 职责: 右侧 Pod 列表展示
- Props:
  - `pods`: Pod 数组
  - `loading`: 加载状态
  - `replicasetName`: 当前选中的 RS 名称
- Events:
  - `@logs`: 点击 Logs 按钮时触发
  - `@exec`: 点击 Exec 按钮时触发
  - `@delete`: 点击 Delete 按钮时触发

### 4.2 修改组件

**DeploymentDetail.vue**
- 路径: `frontend/src/views/workload/DeploymentDetail.vue`
- 修改内容:
  - 将现有「Pods」Tab 重命名为「Replicasets & Pods」
  - 替换原有 Pods Tab 内容为左右分栏布局
  - 新增 `fetchReplicaSets()` 方法
  - 新增 `selectedReplicaset` 状态
  - 保留现有 Info、Events、YAML Tab 不变

### 4.3 新增 API 函数

**resource.ts**
- 路径: `frontend/src/api/resource.ts`
- 新增函数:
  - `getDeploymentReplicaSets(params)`: 获取 Deployment 关联的 ReplicaSet 列表

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
│     - fetchPods(selectedReplicaset) → Pod 列表                 │
│                                                                 │
│  3. 用户点击 Rollback:                                          │
│     - 确认对话框 → rollbackDeployment() → 刷新数据              │
│                                                                 │
│  4. 用户点击 Pod 的 Logs/Exec:                                  │
│     - 跳转到 /logs 或 /terminal 路由                            │
└─────────────────────────────────────────────────────────────────┘
```

## 6. 错误处理

### 6.1 API 错误

- ReplicaSet 列表加载失败: 显示错误提示，提供重试按钮
- Pod 列表加载失败: 显示错误提示，提供重试按钮
- Rollback 操作失败: 显示错误消息，不关闭对话框

### 6.2 空状态

- 无 ReplicaSet: 显示「No ReplicaSets found」
- 选中 RS 无 Pod: 显示「No pods found」
- 数据加载中: 显示骨架屏或加载动画

## 7. 测试策略

### 7.1 单元测试

- ReplicaSetPanel 组件: 测试渲染、选中状态、Rollback 按钮显示逻辑
- PodListPanel 组件: 测试渲染、状态颜色、操作按钮事件

### 7.2 集成测试

- DeploymentDetail 页面: 测试 Tab 切换、RS 选择联动、数据刷新

### 7.3 E2E 测试

- 完整流程: 创建 Deployment → 查看 RS 列表 → 选择 RS → 查看 Pod → Rollback → 验证更新

## 8. 实施计划

### 8.1 第一阶段：后端 API

1. 新增 `GetDeploymentReplicaSets` 函数
2. 新增 `ReplicaSetListParams` 参数结构
3. 注册路由 `GET /k8s/deployment/replicasets`

### 8.2 第二阶段：前端组件

1. 创建 `ReplicaSetPanel.vue` 组件
2. 创建 `PodListPanel.vue` 组件
3. 在 `resource.ts` 中新增 API 函数

### 8.3 第三阶段：页面集成

1. 修改 `DeploymentDetail.vue`，集成新组件
2. 实现 RS 选择联动逻辑
3. 实现 Rollback 操作

### 8.4 第四阶段：测试与优化

1. 编写单元测试
2. 响应式适配
3. 性能优化（懒加载、虚拟滚动）

## 9. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| ReplicaSet 数量过多 | 左侧列表过长 | 支持分页或虚拟滚动 |
| Pod 频繁变化 | 列表频繁刷新 | 使用 WebSocket 或轮询（15秒间隔） |
| Rollback 操作失败 | 用户困惑 | 提供详细错误信息和操作建议 |

## 10. 未来扩展

- **镜像更新历史**: 显示每次更新的镜像变更记录
- **Pod 日志对比**: 支持对比不同 RS 版本的 Pod 日志
- **自动刷新**: 支持 Pod 状态自动刷新（复用现有 useAutoRefresh）
- **批量操作**: 支持批量删除 Pod、批量重启
