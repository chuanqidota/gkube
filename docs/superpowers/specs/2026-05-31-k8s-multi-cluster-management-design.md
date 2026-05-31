# gkube K8s 多集群管理功能设计文档

> 日期：2026-05-31
> 状态：已批准
> 定位：轻量级 K8s 多集群管理面板（类 Kuboard）

---

## 1. 概述

### 1.1 项目定位

gkube 是一个轻量级的 Kubernetes 多集群管理面板，采用中心化直连模式，通过 kubeconfig 连接和管理多个 K8s 集群。技术栈为 Go + Vue 3 + Element Plus，前后端分离架构。

### 1.2 核心功能

- **集群管理**：集群注册（kubeconfig 导入）、编辑、删除、健康检测
- **全局 Dashboard**：跨集群状态总览、资源使用率、工作负载统计、事件聚合
- **认证鉴权**：本地账号系统 + JWT + 细粒度 RBAC（集群/命名空间/资源级别）
- **K8s 资源管理**：Pod、Deployment、Service 等资源的 CRUD 操作
- **Web 终端**：容器 exec 终端 + 日志流

### 1.3 设计原则

- **渐进增强**：基于现有代码结构扩展，复用已有模块
- **轻量优先**：不引入不必要的复杂度，保持系统简洁
- **安全第一**：kubeconfig 加密存储、密码哈希、JWT 认证、RBAC 权限控制

---

## 2. 项目结构

```
gkube/
├── backend/                        # Go 后端
│   ├── main.go                     # 入口
│   ├── cmd/                        # Cobra 命令
│   ├── config/                     # 配置
│   ├── router/                     # 路由注册
│   ├── middleware/                  # 中间件（JWT、RBAC、CORS）
│   ├── app/                        # 业务模块
│   │   ├── auth/                   # 认证模块
│   │   │   ├── api/                # Handler
│   │   │   ├── model/              # 数据模型
│   │   │   ├── params/             # 请求参数
│   │   │   └── service/            # 业务逻辑
│   │   ├── cluster/                # 集群管理模块
│   │   │   ├── api/
│   │   │   ├── model/
│   │   │   ├── params/
│   │   │   └── service/
│   │   ├── dashboard/              # Dashboard 模块
│   │   │   ├── api/
│   │   │   └── service/
│   │   └── k8sresource/            # K8s 资源管理（重构现有）
│   │       ├── api/
│   │       ├── params/
│   │       └── service/
│   ├── pkg/                        # 公共包
│   │   ├── k8s/                    # K8s 客户端工厂
│   │   ├── middleware/             # JWT、RBAC、CORS 中间件
│   │   ├── response/               # 统一响应
│   │   └── ...
│   └── docs/                       # Swagger 文档
│
├── frontend/                       # Vue 3 前端
│   ├── src/
│   │   ├── api/                    # API 请求封装
│   │   ├── views/                  # 页面
│   │   │   ├── login/              # 登录页
│   │   │   ├── dashboard/          # 全局 Dashboard
│   │   │   ├── cluster/            # 集群管理
│   │   │   ├── resource/           # K8s 资源浏览
│   │   │   ├── user/               # 用户管理
│   │   │   └── terminal/           # Web 终端
│   │   ├── components/             # 通用组件
│   │   ├── router/                 # 路由配置
│   │   ├── stores/                 # Pinia 状态管理
│   │   └── utils/                  # 工具函数
│   ├── package.json
│   └── vite.config.ts
│
├── docker-compose.yaml             # 开发环境
├── Makefile                        # 构建命令
└── README.md
```

---

## 3. 后端架构

### 3.1 模块职责

| 模块 | 职责 | 关键能力 |
|------|------|----------|
| **auth** | 认证鉴权 | 登录/注册、JWT Token、用户 CRUD、角色 RBAC |
| **cluster** | 集群管理 | 集群 CRUD、kubeconfig 加密存储、健康检测、状态缓存 |
| **dashboard** | 数据聚合 | 跨集群资源统计、工作负载聚合、事件汇总 |
| **k8sresource** | K8s 资源操作 | Pod/Deployment/Service 等 CRUD（重构现有代码） |

### 3.2 API 路由设计

```
# 认证
POST   /v1/auth/login              # 登录
POST   /v1/auth/refresh            # 刷新 Token
GET    /v1/auth/me                  # 当前用户信息

# 用户管理
GET    /v1/users                    # 用户列表
POST   /v1/users                    # 创建用户
PUT    /v1/users/:id                # 更新用户
DELETE /v1/users/:id                # 删除用户

# 角色管理
GET    /v1/roles                    # 角色列表
POST   /v1/roles                    # 创建角色
PUT    /v1/roles/:id                # 更新角色
DELETE /v1/roles/:id                # 删除角色

# 集群管理
GET    /v1/clusters                 # 集群列表（含健康状态）
POST   /v1/clusters                 # 注册集群（导入 kubeconfig）
GET    /v1/clusters/:id             # 集群详情
PUT    /v1/clusters/:id             # 更新集群
DELETE /v1/clusters/:id             # 删除集群
POST   /v1/clusters/:id/check      # 手动连通性检测

# Dashboard
GET    /v1/dashboard/overview       # 全局概览
GET    /v1/dashboard/resources      # 资源使用率统计
GET    /v1/dashboard/workloads      # 工作负载统计
GET    /v1/dashboard/events         # 跨集群事件聚合

# K8s 资源（挂在集群路由下）
GET    /v1/clusters/:cid/pods
GET    /v1/clusters/:cid/deployments
GET    /v1/clusters/:cid/services
GET    /v1/clusters/:cid/namespaces
...（其他资源类型）
```

### 3.3 中间件链

```
请求 → CORS → JWT 认证 → RBAC 权限校验 → Handler
```

---

## 4. 数据模型

### 4.1 用户模型

```go
type User struct {
    ID           uint      `gorm:"primaryKey"`
    Username     string    `gorm:"unique;size:50"`
    PasswordHash string    `gorm:"size:255"`    // bcrypt 加密
    Email        string    `gorm:"size:100"`
    DisplayName  string    `gorm:"size:100"`
    Status       int       `gorm:"default:1"`   // 1=启用 0=禁用
    Roles        []Role    `gorm:"many2many:user_roles;"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 4.2 角色模型

```go
type Role struct {
    ID          uint         `gorm:"primaryKey"`
    Name        string       `gorm:"unique;size:50"`
    Description string       `gorm:"size:200"`
    Permissions []Permission `gorm:"many2many:role_permissions;"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 4.3 权限模型

```go
type Permission struct {
    ID          uint   `gorm:"primaryKey"`
    Resource    string `gorm:"size:100"`  // cluster, pod, deployment, ...
    Action      string `gorm:"size:50"`   // create, read, update, delete, exec
    ClusterID   *uint                     // 可选：限定集群
    Namespace   string `gorm:"size:100"`  // 可选：限定命名空间
    Description string `gorm:"size:200"`
}
```

### 4.4 集群模型

```go
type K8SCluster struct {
    ID              uint      `gorm:"primaryKey"`
    ClusterName     string    `gorm:"unique;size:100"`
    DisplayName     string    `gorm:"size:100"`
    Description     string    `gorm:"size:500"`
    KubeConfig      string    `gorm:"size:12800"` // AES 加密存储
    Status          string    `gorm:"size:20"`    // online/offline/unknown
    ClusterVersion  string    `gorm:"size:100"`
    NodeCount       int
    LastHealthCheck time.Time
    Labels          string    `gorm:"size:500"`   // JSON: {"env":"prod","region":"cn-east"}
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

---

## 5. 认证与 RBAC

### 5.1 认证流程

1. 用户提交用户名密码 → `POST /v1/auth/login`
2. 后端 bcrypt 验证密码
3. 生成 JWT access_token（2 小时）+ refresh_token（7 天）
4. 后续请求在 Header 中携带 `Authorization: Bearer <token>`
5. JWT 中间件验证 token，提取用户信息
6. RBAC 中间件检查用户是否有该资源的操作权限

### 5.2 权限控制

三级权限控制：
- **集群级别**：指定用户可访问哪些集群
- **命名空间级别**：指定用户可访问哪些命名空间
- **资源级别**：指定用户可操作哪些资源（create/read/update/delete/exec）

预设角色：
- `super_admin`：全部权限
- `cluster_admin`：指定集群全部权限
- `developer`：指定命名空间读写权限
- `viewer`：只读权限

---

## 6. 集群管理

### 6.1 集群注册流程

1. 用户填写集群名称、描述、标签
2. 粘贴 kubeconfig 内容
3. 后端解析 kubeconfig，验证连通性（调用 `/healthz`）
4. 验证通过后，AES-256 加密 kubeconfig 存储到 MySQL
5. 首次采集集群信息（版本、节点数）
6. 返回注册成功

### 6.2 健康检测机制

- **定时探测**：后台 goroutine 每 30 秒检测所有集群
- **检测内容**：调用 K8s API `/healthz` 和 `/version`
- **状态标记**：连续 3 次失败标记为 offline
- **数据采集**：节点数量、CPU/内存容量与使用率、Pod 数量
- **状态缓存**：健康数据缓存到 Redis，TTL 60 秒
- **手动刷新**：`POST /v1/clusters/:id/check` 触发即时检测

---

## 7. Dashboard

### 7.1 全局概览

- 集群总数、在线/离线数量
- 节点总数、就绪节点数
- Pod 总数、运行中 Pod 数

### 7.2 资源使用率

- 各集群 CPU 总量与使用率
- 各集群内存总量与使用率
- 各集群 Pod 总量与运行数

### 7.3 工作负载统计

- 跨集群 Deployment/StatefulSet/DaemonSet/Job/CronJob 数量

### 7.4 事件聚合

- 跨集群 Warning 事件汇总
- 支持按集群、类型筛选

---

## 8. 前端架构

### 8.1 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 框架 | Vue 3 + TypeScript | Composition API |
| UI 库 | Element Plus | 与 Kuboard 一致 |
| 构建 | Vite | 快速开发体验 |
| 状态管理 | Pinia | Vue 3 官方推荐 |
| 路由 | Vue Router 4 | 动态路由 + 路由守卫 |
| HTTP | Axios | 拦截器统一处理 |
| 终端 | Xterm.js | Web 终端 |

### 8.2 核心设计

**请求拦截器**：自动附加 JWT Token，401 自动跳转登录

**路由守卫**：未登录跳转登录页，已登录跳转 Dashboard

**状态管理**：
- `auth store`：用户认证状态、Token 管理
- `cluster store`：当前选中集群、集群列表

**页面布局**：
- 顶部导航栏：Logo + 集群选择器 + 用户信息
- 侧边栏菜单：Dashboard、集群管理、资源浏览、用户管理
- 主内容区：根据路由渲染

### 8.3 页面结构

- **登录页**：用户名密码登录
- **Dashboard**：集群状态卡片 + 资源使用率图表 + 最新事件
- **集群管理**：集群列表 + 添加/编辑集群 + 集群详情
- **资源浏览**：Pod/Deployment/Service 等列表 + YAML 编辑 + 日志查看
- **Web 终端**：Xterm.js + 容器选择 + 会话录制
- **用户管理**：用户 CRUD + 角色分配
- **角色管理**：角色 CRUD + 权限配置

---

## 9. 部署方案

### 9.1 独立部署

```yaml
# docker-compose.yaml
services:
  backend:
    build: ./backend
    ports: ["8080:8080"]
    depends_on: [mysql, redis]

  frontend:
    build: ./frontend
    ports: ["80:80"]
    depends_on: [backend]

  mysql:
    image: mysql:8.0
    volumes: [mysql-data:/var/lib/mysql]

  redis:
    image: redis:7-alpine
```

### 9.2 集群内部署

使用 Helm Chart 部署到 K8s 集群：

```bash
helm install gkube ./deploy/helm/gkube \
  --namespace gkube-system \
  --create-namespace
```

---

## 10. 安全设计

### 10.1 后端安全

- kubeconfig AES-256 加密存储
- 密码 bcrypt 加盐哈希
- JWT Token 过期时间 2 小时
- Refresh Token 过期时间 7 天
- RBAC 权限校验中间件
- 请求频率限制（防暴力破解）
- 操作审计日志

### 10.2 前端安全

- Token 存储在 localStorage
- Axios 拦截器自动附加 Token
- 401 自动跳转登录页
- 路由守卫防止未授权访问
- 敏感操作二次确认弹窗
- XSS 防护（Vue 模板转义）

---

## 11. 技术选型汇总

| 层级 | 技术 | 说明 |
|------|------|------|
| 后端框架 | Go + Gin | 保留现有技术栈 |
| 前端框架 | Vue 3 + TypeScript | Composition API |
| UI 组件库 | Element Plus | 与 Kuboard 一致 |
| 构建工具 | Vite | 快速开发体验 |
| 状态管理 | Pinia | Vue 3 官方推荐 |
| 路由 | Vue Router 4 | 动态路由 + 路由守卫 |
| HTTP 客户端 | Axios | 拦截器统一处理 |
| K8s 客户端 | client-go | 保留现有 |
| 数据库 | MySQL + GORM | 保留现有 |
| 缓存 | Redis | Token + 集群状态缓存 |
| 终端 | Xterm.js + WebSocket | Web 终端 |
