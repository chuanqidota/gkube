# 工作负载列表页设计文档

> 日期：2026-05-31
> 状态：已批准

## 概述

为 StatefulSet、DaemonSet、Job、CronJob 四种 K8s 资源创建前端列表页，风格与现有 Pod/Deployment 列表页一致。

## 设计

### 页面结构

每种资源一个列表页，包含：
- 命名空间筛选下拉框
- 集群名称输入（默认从 store 读取）
- 数据表格：名称、命名空间、状态相关列、创建时间
- 操作列：查看 YAML（弹窗）、删除（确认弹窗）
- 分页

### 各资源表格列

**StatefulSetList.vue**：名称、命名空间、Ready、年龄
**DaemonSetList.vue**：名称、命名空间、Desired、Current、Ready、年龄
**JobList.vue**：名称、命名空间、Completions、Duration、年龄
**CronJobList.vue**：名称、命名空间、Schedule、Suspend、Active、年龄

### API 路由

复用现有后端 API：
- GET /v1/k8s/statefulset/list
- GET /v1/k8s/statefulset/get-yaml
- DELETE /v1/k8s/statefulset/delete
- GET /v1/k8s/daemonset/list
- GET /v1/k8s/daemonset/get-yaml
- DELETE /v1/k8s/daemonset/delete
- GET /v1/k8s/job/list
- GET /v1/k8s/job/get-yaml
- DELETE /v1/k8s/job/delete
- GET /v1/k8s/cronjob/list
- GET /v1/k8s/cronjob/get-yaml
- DELETE /v1/k8s/cronjob/delete

### 文件

- `frontend/src/views/workload/StatefulSetList.vue`
- `frontend/src/views/workload/DaemonSetList.vue`
- `frontend/src/views/workload/JobList.vue`
- `frontend/src/views/workload/CronJobList.vue`
- `frontend/src/api/resource.ts`（添加 API 函数）
- `frontend/src/router/index.ts`（添加路由）
