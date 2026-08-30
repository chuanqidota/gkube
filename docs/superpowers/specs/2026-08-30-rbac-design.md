# gkube RBAC 权限管控系统 — 完整设计方案

> 版本: v2.7 | 日期: 2026-08-30 | 经过 19 轮 review

---

## 一、设计理念

gkube 是多集群 K8s 管理平台，RBAC 的核心目标：**让管理员能精确控制 "谁能对哪个集群的哪个命名空间做什么操作"**。

### 1.1 设计原则

| 原则 | 说明 |
|------|------|
| 平台级拦截 | 权限在 gkube 网关层判断，不直接委托 K8s 原生 RBAC（多集群需要统一管理） |
| 在上下文中管理 | 成员管理入口放在集群列表页（资源上下文），不做独立全局页面 |
| 最小表数 | 2 张新表 + 1 张现有表加字段，不引入用户组（Phase 2） |
| 后端为准 | 前端辅助（隐藏/灰化按钮），后端 403 兜底 |
| 一个作用域一个角色 | UNIQUE(user_id, cluster_name, namespace)，不允许同作用域多角色 |

### 1.1.1 安全边界说明

```
用户 → gkube 平台 (RBAC 过滤) → K8s API (cluster-admin kubeconfig)
              ↑ 唯一权限边界
用户 → kubectl (绕过平台RBAC) → K8s API ← 不受 gkube RBAC 管控
```

gkube 使用统一的 kubeconfig（通常是 cluster-admin）访问所有 K8s 集群。平台 RBAC 是**唯一的权限过滤层**，K8s 层面不做二次过滤。这与 [Rancher](https://docs.ranchermanager.rancher.io/) 的架构一致。

**运维要求：** 生产环境应限制 K8s API Server 的直接访问（网络层隔离），确保所有操作都经过 gkube 平台。

### 1.2 权限层级

```
平台管理员 (super_admin)
├── config.yaml admin_users 白名单 + user.is_super_admin = true
├── 跳过所有权限检查
└── 管理 RBAC 配置本身
│
├── 集群级角色 (cluster scope)
│   ├── cluster-admin   ─ 全权（含节点管理）
│   ├── cluster-editor  ─ 读写，不能删，不能管节点
│   └── cluster-viewer  ─ 只读 + 终端
│
└── 命名空间级角色 (namespace scope)
    ├── ns-admin        ─ 该 NS 内全权
    ├── ns-editor       ─ 该 NS 内读写，不能删
    └── ns-viewer       ─ 该 NS 内只读
```

### 1.3 权限继承规则

```
用户在 (集群X, 命名空间Y) 的有效权限
  = MAX( 集群X 的集群级角色, 命名空间Y 的NS级角色 )
  = 取并集（权限只增不减）

判定优先级（高→低）:
  1. super_admin → 直接放行
  2. 集群级绑定 → 该集群内全部放行（按角色权限）
  3. 命名空间级绑定 → 精确匹配 namespace
  4. 无任何绑定 → 403
```

---

## 二、数据模型

### 2.1 `user` 表 — 新增 1 个字段

```go
// internal/auth/model/user.go
// 现有 User struct 新增:
IsSuperAdmin bool `gorm:"column:is_super_admin;not null;default:false;comment:平台管理员" json:"isSuperAdmin"`
```

- `gkube seed` 时自动设 admin 为 `true`
- 后续 admin 可在前端对其他用户设置
- 与 config.yaml `admin_users` 白名单取 OR

### 2.2 `role` 表 — 新建

```go
// internal/rbac/model/role.go
type Role struct {
    ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name          string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
    DisplayName   string    `gorm:"type:varchar(100);not null" json:"displayName"`
    ScopeType     string    `gorm:"type:varchar(20);not null;comment:cluster|namespace" json:"scopeType"`
    IsSystem      bool      `gorm:"not null;default:true;comment:系统预置不可删除" json:"isSystem"`
    Permissions   string    `gorm:"type:text;not null;comment:JSON权限定义" json:"permissions"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    PermissionsMap map[string][]string `gorm:"-" json:"-"` // 运行时缓存，不入库
}

func (Role) TableName() string { return "role" }
```

**`Permissions` 字段 JSON 格式：**

```json
{
  "workload":     ["read","create","update","delete","terminal"],
  "network":      ["read","create","update","delete"],
  "storage":      ["read","create","update","delete"],
  "config":       ["read","create","update","delete"],
  "node":         ["read","cordon","taint","drain","delete"],
  "namespace":    ["read","create","update","delete"],
  "event":        ["read"],
  "audit":        ["read"],
  "crd":          ["read","create","update","delete"],
  "terminal":     ["terminal"],
  "cluster_mgmt": ["read"]
}
```

**`GetPermissionsMap()` 方法：** 解析 JSON 到 `PermissionsMap`，解析失败返回空 map 并记录日志（不 panic）。

### 2.3 `permission_binding` 表 — 新建

```go
// internal/rbac/model/binding.go
type PermissionBinding struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID      uint      `gorm:"not null;uniqueIndex:idx_user_cluster_ns" json:"userId"`
    RoleID      uint      `gorm:"not null" json:"roleId"`
    ClusterName string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_user_cluster_ns" json:"clusterName"`
    Namespace   string    `gorm:"type:varchar(100);not null;default:'';uniqueIndex:idx_user_cluster_ns;comment:空串=集群级" json:"namespace"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Role        Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (PermissionBinding) TableName() string { return "permission_binding" }
```

**约束说明：**

| 约束 | 说明 |
|------|------|
| `UNIQUE(user_id, cluster_name, namespace)` | 一个用户在一个作用域只能有一个角色 |
| `namespace NOT NULL DEFAULT ''` | 集群级绑定统一存空串，不允许 NULL |
| `namespace = ""` + `role.scopeType = "cluster"` | 集群级绑定 |
| `namespace = "xxx"` + `role.scopeType = "namespace"` | 命名空间级绑定 |

### 2.4 关系图

```
User (1) ──< (N) PermissionBinding (N) >── (1) Role
                        │
                        └── scope = (clusterName, namespace)
```

**总计：1 张现有表加字段 + 2 张新表 = 3 张表。**

---

## 三、6 个预置角色权限矩阵

### 3.1 集群级角色

| 权限组 | 含义 | cluster-admin | cluster-editor | cluster-viewer |
|--------|------|:---:|:---:|:---:|
| workload | Deployment/Pod/StatefulSet/DaemonSet/Job/CronJob/ReplicaSet | R C U D T | R C U T | R T |
| network | Service/Ingress/NetworkPolicy | R C U D | R C U | R |
| storage | PV/PVC/StorageClass/VolumeSnapshot/VolumeSnapshotClass | R C U D | R C U | R |
| config | ConfigMap/Secret/ResourceQuota/LimitRange | R C U D | R C U | R |
| node | 节点管理 (cordon/taint/drain) | R CORDON TAINT DRAIN D | R | R |
| namespace | 命名空间管理 | R C U D | R | R |
| event | 事件查看 | R | R | R |
| audit | 审计日志 | R | R | R |
| crd | CRD + 自定义资源 | R C U D | R C U | R |
| terminal | exec + log | T | T | T |
| cluster_mgmt | 集群管理 | R | — | — |

### 3.2 命名空间级角色

| 权限组 | ns-admin | ns-editor | ns-viewer |
|--------|:---:|:---:|:---:|
| workload | R C U D T | R C U T | R |
| network | R C U D | R C U | R |
| storage | R C U D | R C U | R |
| config | R C U D | R C U | R |
| event | R | R | R |
| crd | R C U D | R C U | R |
| terminal | T | T | — |
| node | — | — | — |
| namespace | — | — | — |
| audit | — | — | — |
| cluster_mgmt | — | — | — |

**图例：** R=read C=create U=update D=delete T=terminal(含exec和log) —=无权限

> **注意：** `audit`（审计日志）、`node`、`namespace`、`cluster_mgmt` 为集群级资源，仅集群级角色可访问。NS 级角色无权查看审计日志、管理节点或创建命名空间。

---

## 四、权限判定中间件

### 4.1 签名

```go
// pkg/middleware/permission.go
func RequirePermission() gin.HandlerFunc
```

无需传参，从请求路径自动推断 `resourceGroup` 和 `verb`。

### 4.2 判定流程

```
请求进入 RequirePermission
  │
  ├─ 1. 提取作用域
  │     ├─ clusterName = c.Query("clusterName")        ← GET/POST 统一从 query 读
  │     └─ namespace   = extractNamespace(c)            ← GET 从 query，POST 从 body
  │
  ├─ 2. 查询用户权限（permcache，5min TTL）
  │     ├─ 缓存 miss → 一次 DB 查询：user.is_super_admin + 所有 bindings
  │     └─ 缓存 hit  → 直接使用
  │
  ├─ 3. 超级管理员？（从缓存读取，不查 DB）
  │     ├─ auth.IsAdmin(username)           ← config.yaml 白名单（内存）
  │     ├─ cachedPermissions.IsSuperAdmin   ← DB 缓存值
  │     └─ 任一为 true → 放行
  │
  ├─ 4. 从路径推断 resourceGroup + verb
  │     ├─ resourceGroup = resolveResourceGroup(path)
  │     └─ verb = resolveVerb(resourceGroup, path, method)
  │
  ├─ 5. 检查任一绑定的角色 permissions 是否包含 resourceGroup + verb
  │     ├─ 命中 → 放行
  │     └─ 未命中 → 403 "权限不足"
  │
  └─ 6. 绑定变更时主动失效缓存
```

### 4.3 namespace 提取策略

```go
func extractNamespace(c *gin.Context) string {
    // 优先从 query 读（GET 请求及前端可能注入的场景）
    if ns := c.Query("namespace"); ns != "" {
        return ns
    }
    // POST/PUT/DELETE: 从 body 读（ShouldBindBodyWith 不消费 body）
    if c.Request.Method != "GET" {
        var body struct {
            Namespace string `json:"namespace"`
        }
        if err := c.ShouldBindBodyWith(binding.JSON, &body); err == nil {
            return body.Namespace
        }
    }
    return ""
}
```

**关键：** `ShouldBindBodyWith` 在 Gin 内部缓存 body，后续 handler 的 `ShouldBindJSON` 仍能正常读取。不需要前端改动。

### 4.4 路径 → resourceGroup 推断

```go
func resolveResourceGroup(path string) string {
    // 长前缀优先匹配
    switch {
    case contains(path, "/volumesnapshotclass/"):
        return "storage"
    case contains(path, "/volumesnapshot/"):
        return "storage"
    case contains(path, "/networkpolicy/"):
        return "network"
    case contains(path, "/resourcequota/"):
        return "config"
    case contains(path, "/limitrange/"):
        return "config"
    case contains(path, "/storageclass/"):
        return "storage"
    case contains(path, "/configmap/"):
        return "config"
    case contains(path, "/replicaset/"):
        return "workload"
    case contains(path, "/statefulset/"):
        return "workload"
    case contains(path, "/daemonset/"):
        return "workload"
    case contains(path, "/cronjob/"):
        return "workload"
    case contains(path, "/deployment/"):
        return "workload"
    case contains(path, "/namespace/"):
        return "namespace"
    case contains(path, "/ingress/"):
        return "network"
    case contains(path, "/service/"):
        return "network"
    case contains(path, "/secret/"):
        return "config"
    case contains(path, "/container/exec"):
        return "terminal"
    case contains(path, "/pod/"):
        return "workload"
    case contains(path, "/event/"):
        return "event"
    case contains(path, "/node/"):
        return "node"
    case contains(path, "/job/"):
        return "workload"
    case contains(path, "/pvc/"):
        return "storage"
    case contains(path, "/pv/"):
        return "storage"
    case contains(path, "/hpa/"):
        return "workload"
    case contains(path, "/crd/"):
        return "crd"
    case contains(path, "/log"):
        return "terminal"
    default:
        // ⚠️ 未注册的路由会走到这里，记录警告日志便于排查
        logger.Warn(fmt.Sprintf("RequirePermission: 未识别的资源路径 %s，拒绝访问", path))
        return ""
    }
}
```

> **运维提示：** 新增 K8s 资源路由时，必须同步在 `resolveResourceGroup` 的 switch 中注册对应条目，否则该路由会静默 403。

### 4.5 HTTP Method → verb 默认映射

| Method | verb |
|:---:|:---:|
| GET | read |
| POST | create |
| PUT | update |
| DELETE | delete |

### 4.6 特殊路由覆写表

```go
var specialVerbOverrides = []struct {
    Suffix string
    Method string
    Verb   string
}{
    // 节点管理
    {"/node/cordon",      "PUT",    "cordon"},
    {"/node/taints",      "PUT",    "taint"},
    {"/node/drain",       "PUT",    "drain"},
    // 工作负载操作
    {"/deployment/scale",      "PUT",  "update"},
    {"/deployment/restart",    "POST", "update"},
    {"/deployment/rollback",   "POST", "update"},
    {"/statefulset/scale",     "PUT",  "update"},
    {"/statefulset/restart",   "POST", "update"},
    {"/statefulset/rollback",  "POST", "update"},
    {"/daemonset/restart",     "POST", "update"},
    {"/daemonset/rollback",    "POST", "update"},
    {"/cronjob/suspend",       "PUT",  "update"},
    {"/cronjob/resume",        "PUT",  "update"},
    {"/cronjob/trigger",       "POST", "update"},
    {"/hpa/pause",             "POST", "update"},
    {"/hpa/resume",            "POST", "update"},
    // 命名空间特殊
    {"/namespace/create",  "POST", "create"},
    {"/namespace/list",    "GET",  "read"},
    {"/namespace/detail",  "GET",  "read"},
    // 终端/日志
    {"/container/exec",    "GET",  "terminal"},
    {"/log",               "GET",  "terminal"},
    {"/log/stream",        "GET",  "terminal"},
}
```

### 4.7 与现有 RequireAdmin() 的关系

| 路由类别 | 当前中间件 | 改造后 | 说明 |
|---|---|---|---|
| K8s 写路由 (POST/PUT/DELETE) | `RequireAdmin()` | `RequirePermission()` | **替换** |
| K8s 读路由 (GET) | 无 | `RequirePermission()` | **新增**（约 80 条） |
| RBAC 读 (roles/my-permissions/cluster-members) | — | 仅 JWTAuth | 所有用户可访问 |
| RBAC 写 (bindings CRUD) | — | `RequireAdmin()` | 仅管理员 |
| 用户管理 (`/v1/users/`) | `RequireAdmin()` | 不变 | 仅管理员 |
| 集群管理 (`/v1/clusters/`) | `RequireAdmin()` | 不变 | 仅管理员 |
| 仪表盘 (`/v1/dashboard/`) | 无 | **保持无权限检查** | 任何登录用户可查看仪表盘概览 |

---

## 五、后端 API 设计

### 5.1 路由注册 (`internal/router/rbac.go`)

```go
func registerRbacRoutes(rg *gin.RouterGroup) {
    // 所有已认证用户可访问（普通用户查自己权限、查看角色列表、查看集群成员）
    rg.GET("rbac/roles", rbacHandler.ListRoles)
    rg.GET("rbac/my-permissions", rbacHandler.MyPermissions)
    rg.GET("rbac/cluster-members", rbacHandler.ClusterMembers)

    // 仅管理员可访问（写操作）
    admin := rg.Group("rbac", middleware.RequireAdmin())
    {
        admin.GET("bindings", rbacHandler.ListBindings)
        admin.POST("bindings", rbacHandler.CreateBinding)
        admin.PUT("bindings/:id", rbacHandler.UpdateBinding)
        admin.DELETE("bindings/:id", rbacHandler.DeleteBinding)
    }
}
```

**路由权限分层：**

| 路由 | 中间件 | 说明 |
|------|--------|------|
| `GET /rbac/roles` | JWTAuth | 所有用户可查看角色列表 |
| `GET /rbac/my-permissions` | JWTAuth | 普通用户获取自己的权限 |
| `GET /rbac/cluster-members` | JWTAuth | 查看集群成员（只读，所有用户可见） |
| `GET /rbac/bindings` | RequireAdmin | 管理员查询绑定（支持筛选） |
| `POST /rbac/bindings` | RequireAdmin | 管理员创建绑定 |
| `PUT /rbac/bindings/:id` | RequireAdmin | 管理员修改绑定 |
| `DELETE /rbac/bindings/:id` | RequireAdmin | 管理员删除绑定 |

### 5.2 Handler 参数结构体

```go
// internal/rbac/handler.go

package rbac

import (
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    authmodel "gkube/internal/auth/model"
    "gkube/internal/rbac/model"
    "gkube/pkg/auth"
    "gkube/pkg/database"
    "gkube/pkg/logger"
    "gkube/pkg/response"
)

type rbacHandler struct{}
var RbacHandler = new(rbacHandler)

// --- 参数结构体 ---

type CreateBindingParams struct {
    UserID      uint   `json:"userId" binding:"required" label:"用户ID"`
    RoleID      uint   `json:"roleId" binding:"required" label:"角色ID"`
    ClusterName string `json:"clusterName" binding:"required" label:"集群名称"`
    Namespace   string `json:"namespace"` // 可选，缺省=""（集群级）
}

type UpdateBindingParams struct {
    RoleID uint `json:"roleId" binding:"required" label:"角色ID"`
}

type ListBindingsQuery struct {
    ClusterName string `form:"clusterName" json:"clusterName"`
    UserID      *uint  `form:"userId" json:"userId"`
    Page        int    `form:"page" json:"page"`
    Size        int    `form:"size" json:"size"`
}

type ClusterMembersQuery struct {
    ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
}

// parseID 从 URL path 参数解析 uint ID，失败返回 0 + false
func parseID(c *gin.Context) (uint, bool) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Fail(c, "无效的ID")
        return 0, false
    }
    return uint(id), true
}

// isDuplicateErr 检查 MySQL 唯一约束冲突
func isDuplicateErr(err error) bool {
    return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}
```

### 5.3 API 详情

#### `GET /v1/rbac/roles`

返回所有角色列表。

Response:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "cluster-admin",
      "displayName": "集群管理员",
      "scopeType": "cluster",
      "isSystem": true,
      "permissions": { "workload": ["read","create","update","delete","terminal"], ... }
    }
  ]
}
```

#### `GET /v1/rbac/bindings`

查询参数（均为可选，组合筛选）：

| 参数 | 类型 | 说明 |
|------|------|------|
| clusterName | string | 按集群筛选 |
| userId | uint | 按用户筛选 |
| page | int | 页码，默认 1 |
| size | int | 每页数量，默认 20 |

Response:
```json
{
  "code": 200,
  "data": {
    "items": [
      {
        "id": 1,
        "userId": 2,
        "roleId": 4,
        "clusterName": "prod-cluster",
        "namespace": "",
        "user": { "id": 2, "username": "zhangsan", "displayName": "张三" },
        "role": { "id": 4, "name": "cluster-editor", "displayName": "集群编辑者", "scopeType": "cluster" },
        "createdAt": "2026-08-30T10:00:00Z"
      }
    ],
    "total": 6
  }
}
```

#### `POST /v1/rbac/bindings`

Body:
```json
{
  "userId": 2,
  "roleId": 4,
  "clusterName": "prod-cluster",
  "namespace": ""
}
```

校验规则：
- `userId`、`roleId`、`clusterName` 必填
- `namespace` 可选，缺省等同于空串（集群级绑定）
- `roleId` 对应的 `role.scopeType` 必须与 `namespace` 是否为空匹配
- `UNIQUE(user_id, cluster_name, namespace)` 冲突时返回 409

Response:
| 场景 | HTTP Status | Body |
|------|:-----------:|------|
| 创建成功 | 200 | `{code:20, msg:"创建成功", data: {id:1, ...}}` |
| 参数校验失败 | 400 | `{code:0, msg:"参数校验失败", data:null}` |
| 角色类型与作用域不匹配 | 400 | `{code:0, msg:"角色类型与绑定作用域不匹配", data:null}` |
| 重复绑定 | 409 | `{code:0, msg:"该用户在此作用域已有绑定", data:null}` |
| 服务器错误 | 500 | `{code:0, msg:"创建绑定失败", data:null}` |

#### `PUT /v1/rbac/bindings/:id`

Body（只能改角色）:
```json
{
  "roleId": 1
}
```

校验规则：
- `roleId` 必填
- 新角色的 `scopeType` 必须与原绑定一致：原绑定 `namespace=""` → 只能换集群级角色；原绑定 `namespace="xxx"` → 只能换 NS 级角色
- 不匹配时返回 400 "角色类型与绑定作用域不匹配"

Response:
| 场景 | HTTP Status | Body |
|------|:-----------:|------|
| 修改成功 | 200 | `{code:200, msg:"更新成功", data:{id:1, ...}}` |
| binding 不存在 | 404 | `{code:0, msg:"绑定不存在", data:null}` |
| 角色类型不匹配 | 400 | `{code:0, msg:"角色类型与绑定作用域不匹配", data:null}` |
| 服务器错误 | 500 | `{code:0, msg:"更新绑定失败", data:null}` |

#### `DELETE /v1/rbac/bindings/:id`

Response:
| 场景 | HTTP Status | Body |
|------|:-----------:|------|
| 删除成功 | 200 | `{code:200, msg:"删除成功", data:null}` |
| binding 不存在 | 404 | `{code:0, msg:"绑定不存在", data:null}` |
| 服务器错误 | 500 | `{code:0, msg:"删除绑定失败", data:null}` |

#### `GET /v1/rbac/my-permissions`

当前用户调用，返回所有集群的权限。

Response:
```json
{
  "code": 200,
  "data": {
    "isSuperAdmin": false,
    "bindings": [
      { "clusterName": "prod", "namespace": "", "roleName": "cluster-editor" },
      { "clusterName": "prod", "namespace": "default", "roleName": "ns-admin" }
    ]
  }
}
```

#### `GET /v1/rbac/cluster-members?clusterName=xxx`

返回某集群的全部绑定（分组展示）。

Response:
```json
{
  "code": 200,
  "data": {
    "items": [
      {
        "id": 1,
        "userId": 2,
        "username": "zhangsan",
        "displayName": "张三",
        "scopeType": "cluster",
        "namespace": "",
        "roleName": "cluster-editor",
        "roleDisplayName": "集群编辑者"
      },
      {
        "id": 2,
        "userId": 3,
        "username": "lisi",
        "displayName": "李四",
        "scopeType": "namespace",
        "namespace": "default",
        "roleName": "ns-editor",
        "roleDisplayName": "空间编辑者"
      }
    ],
    "total": 6
  }
}
```

### 5.4 Handler 实现示例

以 `CreateBinding` 为参考，其余 handler 遵循相同模式：

```go
// CreateBinding 创建权限绑定
func (h *rbacHandler) CreateBinding(c *gin.Context) {
    var p CreateBindingParams
    if err := c.ShouldBindJSON(&p); err != nil {
        response.Fail(c, "参数校验失败")
        return
    }

    // 校验用户存在
    var user model.User
    if err := database.DB.First(&user, p.UserID).Error; err != nil {
        response.Fail(c, "用户不存在")
        return
    }

    // 校验角色存在
    var role model.Role
    if err := database.DB.First(&role, p.RoleID).Error; err != nil {
        response.Fail(c, "角色不存在")
        return
    }

    // 校验 scopeType 匹配
    if role.ScopeType == "cluster" && p.Namespace != "" {
        response.Fail(c, "集群级角色不能绑定到命名空间")
        return
    }
    if role.ScopeType == "namespace" && p.Namespace == "" {
        response.Fail(c, "命名空间级角色必须指定命名空间")
        return
    }

    // 创建绑定
    binding := model.PermissionBinding{
        UserID:      p.UserID,
        RoleID:      p.RoleID,
        ClusterName: p.ClusterName,
        Namespace:   p.Namespace,
    }
    if err := database.DB.Create(&binding).Error; err != nil {
        // UNIQUE 约束冲突
        if isDuplicateErr(err) {
            response.FailWithStatus(c, http.StatusConflict, "该用户在此作用域已有绑定")
            return
        }
        logger.Error(fmt.Sprintf("创建绑定失败: %v", err))
        response.FailWithStatus(c, http.StatusInternalServerError, "创建绑定失败")
        return
    }

    // 失效缓存
    auth.InvalidateUserPermissions(p.UserID)

    response.Success(c, "创建成功", binding)
}

// isDuplicateErr 检查是否为 MySQL 唯一约束冲突错误
func isDuplicateErr(err error) bool {
    return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}
```

> **其余 handler 参考此模式：** `ShouldBindJSON/Query` → 校验 → DB 操作 → 失效缓存 → `response.Success/Fail`。错误码使用 `response.Fail`（400）或 `response.FailWithStatus`（404/409/500）。

---

## 六、权限缓存设计

```go
// pkg/auth/permcache.go

package auth

import (
    "gkube/internal/rbac/model"
)

type CachedPermissions struct {
    IsSuperAdmin bool
    Bindings     []model.PermissionBinding  // 含 Preload("Role")
    ExpireAt     time.Time
}

var permCache sync.Map // key: userID (uint), value: *CachedPermissions

const permCacheTTL = 5 * time.Minute

func GetUserPermissions(userID uint) *CachedPermissions
func InvalidateUserPermissions(userID uint)
func InvalidateAllPermissions()  // 仅管理员批量操作时使用
```

**`GetUserPermissions` 内部逻辑：**

```go
package auth

import (
    "fmt"
    "time"

    "gkube/internal/rbac/model"
    "gkube/pkg/database"
    "gkube/pkg/logger"
)

func GetUserPermissions(userID uint) *CachedPermissions {
    // 1. 查缓存
    if v, ok := permCache.Load(userID); ok {
        cp := v.(*CachedPermissions)
        if time.Now().Before(cp.ExpireAt) {
            return cp
        }
        permCache.Delete(userID) // 过期删除
    }

    // 2. 查 DB：一次查询获取 user.is_super_admin + 所有 bindings
    var user struct{ IsSuperAdmin bool }
    database.DB.Model(&model.User{}).Select("is_super_admin").Where("id = ?", userID).First(&user)

    var bindings []model.PermissionBinding
    database.DB.Preload("Role").Where("user_id = ?", userID).Find(&bindings)

    // 3. 写入缓存
    cp := &CachedPermissions{
        IsSuperAdmin: user.IsSuperAdmin,
        Bindings:     bindings,
        ExpireAt:     time.Now().Add(permCacheTTL),
    }
    permCache.Store(userID, cp)
    logger.Info(fmt.Sprintf("权限缓存已加载: userID=%d, isSuperAdmin=%v, bindings=%d",
        userID, user.IsSuperAdmin, len(bindings)))
    return cp
}
```

**关键：** `isSuperAdmin` 从 DB 读取后与 bindings 一起缓存。中间件不再每次请求查 DB。

**缓存策略：**

| 操作 | 行为 |
|------|------|
| 权限查询 | 先查缓存，miss 则查 DB 并写入缓存 |
| 创建/修改/删除绑定 | 调用 `InvalidateUserPermissions(userID)` |
| 角色变更（Phase 2 自定义角色） | 调用 `InvalidateAllPermissions()` |

---

## 七、RBAC 审计日志

### 7.1 现有审计中间件适配

现有 `AuditLog()` 中间件通过 `parseK8sPath()` 解析 `/v1/k8s/` 前缀路径。RBAC 路由 `/v1/rbac/` 不会被捕获。

**修改 `parseK8sPath()` 支持双前缀：**

```go
func parseK8sPath(path string) []string {
    prefixes := []string{"/v1/k8s/", "/v1/rbac/"}
    for _, prefix := range prefixes {
        if strings.HasPrefix(path, prefix) {
            rest := strings.TrimPrefix(path, prefix)
            parts := strings.SplitN(rest, "/", 2)
            if len(parts) >= 2 {
                return parts
            }
        }
    }
    return nil
}
```

**RBAC 路由审计映射：**

| 路由 | Resource | Action | 说明 |
|------|----------|--------|------|
| POST /rbac/bindings | rbac | create-binding | 创建权限绑定 |
| PUT /rbac/bindings/:id | rbac | update-binding | 修改权限绑定 |
| DELETE /rbac/bindings/:id | rbac | delete-binding | 删除权限绑定 |

### 7.2 RBAC 路由挂载审计中间件

```go
// internal/router/rbac.go
admin := rg.Group("rbac", middleware.RequireAdmin(), middleware.AuditLog())
```

### 7.3 审计日志中的 clusterName 提取

RBAC 路由的 `clusterName` 在请求 body（POST/PUT）或 query（DELETE）中，不在路径中。需扩展 AuditLog 中间件的 cluster 提取逻辑：

```go
// 现有逻辑只读 query 参数
cluster := c.Query("clusterName")

// 扩展：RBAC 路由从 body 读取（POST/PUT）
if cluster == "" && (method == "POST" || method == "PUT") {
    var body struct{ ClusterName string }
    if err := c.ShouldBindBodyWith(binding.JSON, &body); err == nil {
        cluster = body.ClusterName
    }
}
```

### 7.4 Secret 值安全

**现状：** `SecretDetail.vue` 直接显示 base64 解码后的 Secret 值，无 masking。

**风险：** ns-viewer 可查看 Secret 原始值（base64 编码非加密）。

**Phase 1 处理：** 在 `SecretDetail.vue` 中增加 masking：

```vue
<!-- 默认 masked，点击 reveal -->
<template #default="{ row }">
  <span v-if="!row.revealed">•••••••</span>
  <span v-else>{{ row.decodedValue }}</span>
  <el-button link @click="row.revealed = !row.revealed">
    {{ row.revealed ? '隐藏' : '显示' }}
  </el-button>
</template>
```

对齐 Rancher / Kuboard 的 Secret 安全实践。

---

## 八、后端路由迁移

### 8.1 RequireAdmin → RequirePermission 替换 + GET 路由新增

`internal/router/k8s.go` 中：

**写路由（约 20 条）：** 现有 `RequireAdmin()` 替换为 `RequirePermission()`

```go
// 替换前
rg.POST("deployment/create", middleware.RequireAdmin(), k8s.Deployment.CreateDeployment)

// 替换后
rg.POST("deployment/create", middleware.RequirePermission(), k8s.Deployment.CreateDeployment)
```

**读路由（约 80 条）：** 现无中间件，新增 `RequirePermission()`

```go
// 改造前
rg.GET("deployment/list", k8s.Deployment.GetDeploymentList)

// 改造后
rg.GET("deployment/list", middleware.RequirePermission(), k8s.Deployment.GetDeploymentList)
```

**影响路由总数：** 约 100 条 K8s 路由（20 条替换 + 80 条新增）。

### 8.2 特殊操作的权限映射

以下路由的 verb 被覆写（通过 specialVerbOverrides 表）：

| 路由 | HTTP Method | 实际 verb | resourceGroup |
|------|:---:|:---:|:---:|
| /node/cordon | PUT | cordon | node |
| /node/taints | PUT | taint | node |
| /node/drain | PUT | drain | node |
| /deployment/scale | PUT | update | workload |
| /deployment/restart | POST | update | workload |
| /deployment/rollback | POST | update | workload |
| /statefulset/scale | PUT | update | workload |
| /statefulset/restart | POST | update | workload |
| /statefulset/rollback | POST | update | workload |
| /daemonset/restart | POST | update | workload |
| /daemonset/rollback | POST | update | workload |
| /cronjob/suspend | PUT | update | workload |
| /cronjob/resume | PUT | update | workload |
| /cronjob/trigger | POST | update | workload |
| /hpa/pause | POST | update | workload |
| /hpa/resume | POST | update | workload |
| /namespace/create | POST | create | namespace |
| /container/exec | GET | terminal | terminal |
| /log | GET | terminal | terminal |
| /log/stream | GET | terminal | terminal |

---

## 九、前端设计

### 9.1 架构变更

| 变更点 | 说明 |
|--------|------|
| `stores/auth.ts` | user 增加 `isSuperAdmin`、`permissions` 字段 |
| `stores/auth.ts` | 新增 `fetchPermissions()`、`canAccess()`、`hasRole()` |
| `api/request.ts` | 403 统一拦截（见下方代码） |
| `api/rbac.ts` | 新建，RBAC API 客户端 |
| `views/cluster/ClusterList.vue` | 每行新增 "成员" 按钮 + "成员数" 列 |
| `views/cluster/ClusterMembersDialog.vue` | 新建，集群成员管理弹窗 |
| `views/cluster/AddBindingDialog.vue` | 新建，添加/编辑绑定弹窗 |
| `locales/zh-CN.ts` + `en.ts` | RBAC 相关翻译 |

### 9.2 Auth Store 扩展

```typescript
// src/stores/auth.ts

interface PermissionBinding {
  clusterName: string
  namespace: string        // "" = 集群级
  roleName: string         // "cluster-admin" | "ns-editor" | ...
}

interface UserInfo {
  id?: number | string
  username: string
  email?: string
  display_name?: string
  isSuperAdmin?: boolean   // 替代原 isAdmin
  permissions?: PermissionBinding[]
  [key: string]: unknown
}

// 新增方法
async function fetchPermissions(): Promise<void>
// 调用 GET /rbac/my-permissions，结果存入 user.permissions

function canAccess(clusterName: string, namespace?: string): boolean
// 检查 user 是否有指定作用域的权限（任意角色即可）

function hasRole(clusterName: string, namespace?: string, roles?: string[]): boolean
// 检查 user 是否有指定角色
```

**`fetchPermissions` 完整实现：**

```typescript
import { getMyPermissions } from '@/api/rbac'

async function fetchPermissions(): Promise<void> {
  if (!user.value) return
  try {
    const res = await getMyPermissions()
    // res.data 已被 axios 拦截器解包为 { isSuperAdmin, bindings }
    user.value = {
      ...user.value,
      isSuperAdmin: res.data.isSuperAdmin,
      permissions: res.data.bindings || [],
    }
  } catch {
    // 网络失败时设空权限，页面照常渲染（后端 403 兜底）
    user.value = { ...user.value, permissions: [] }
  }
}
```

### 9.2.1 登录流程

```
用户提交登录表单
  → authStore.login()
     → POST /auth/login → 拿到 accessToken + refreshToken + isSuperAdmin + user
     → 存入 token + user（含 isSuperAdmin）
  → router.push(redirect)
  → 路由守卫 beforeEach:
     if (user && !user.permissions) {
       try {
         await authStore.fetchPermissions()  // ← 拉取权限，存入 user.permissions
       } catch {
         // 网络失败时设置空权限，页面照常渲染（后端 403 兜底）
         user.permissions = []
       }
     }
  → 菜单/按钮根据 permissions 渲染
```

**关键：** `fetchPermissions()` 在路由守卫中按需调用（仅首次），不是每次路由跳转都调用。`permissions` 持久化到 localStorage（和 user 一起），刷新后无需重新拉取。管理员修改权限后，被修改用户下次操作会因 403 或手动刷新触发重新拉取。

**`request.ts` 403 拦截器新增代码：**

在现有 response interceptor 的 `error` 处理中，401 逻辑之后增加 403 处理：

```typescript
// api/request.ts response interceptor error 分支中，401 处理之后：

if (error.response?.status === 403) {
  ElMessage.error(error.response?.data?.msg || '权限不足')
  return Promise.reject(new Error('权限不足'))
}
```

### 9.2.2 `canAccess()` 逻辑

```typescript
function canAccess(clusterName: string, namespace?: string): boolean {
  if (!user.value) return false
  if (user.value.isSuperAdmin) return true
  if (!user.value.permissions) return false

  return user.value.permissions.some(p => {
    if (p.clusterName !== clusterName) return false
    // 集群级绑定 (namespace="") 覆盖该集群所有 NS
    if (p.namespace === '') return true
    // NS 级绑定精确匹配
    if (namespace && p.namespace === namespace) return true
    return false
  })
}
```

**核心逻辑：** 集群级绑定（`namespace=""`）覆盖该集群下所有命名空间。NS 级绑定精确匹配。

### 9.2.3 `hasRole()` 逻辑

```typescript
function hasRole(clusterName: string, namespace?: string, roles?: string[]): boolean {
  if (!user.value) return false
  if (user.value.isSuperAdmin) return true
  if (!user.value.permissions) return false

  return user.value.permissions.some(p => {
    if (p.clusterName !== clusterName) return false
    // 集群级绑定覆盖所有 NS
    if (p.namespace === '') {
      return !roles || roles.includes(p.roleName)
    }
    // NS 级精确匹配
    if (namespace && p.namespace === namespace) {
      return !roles || roles.includes(p.roleName)
    }
    return false
  })
}
```

**`canAccess` vs `hasRole` 区别：**
- `canAccess(cluster, ns)` → 该用户在此作用域是否有**任意**权限（用于控制按钮可用性）
- `hasRole(cluster, ns, ['cluster-admin'])` → 该用户在此作用域是否有**指定**角色（用于控制高危操作可见性）

### 9.3 前端权限控制策略

| 层面 | 策略 |
|------|------|
| 菜单可见性 | "系统管理" 菜单仅 `isSuperAdmin` 可见（不变） |
| K8s 操作按钮 | 根据 `canAccess()` 控制 disabled + tooltip "权限不足" |
| 403 处理 | axios 拦截器统一弹 `ElMessage.error` |
| 路由守卫 | 前端不做路由拦截写权限，后端 403 兜底 |

### 9.4 RBAC API 客户端

```typescript
// src/api/rbac.ts
import request from './request'

export const getRoles = () =>
  request.get('/rbac/roles')

export const getBindings = (params: {
  clusterName?: string
  userId?: number
  page?: number
  size?: number
}) => request.get('/rbac/bindings', { params })

export const createBinding = (data: {
  userId: number
  roleId: number
  clusterName: string
  namespace?: string
}) => request.post('/rbac/bindings', data)

export const updateBinding = (id: number, data: { roleId: number }) =>
  request.put(`/rbac/bindings/${id}`, data)

export const deleteBinding = (id: number) =>
  request.delete(`/rbac/bindings/${id}`)

export const getMyPermissions = () =>
  request.get('/rbac/my-permissions')

export const getClusterMembers = (clusterName: string) =>
  request.get('/rbac/cluster-members', { params: { clusterName } })
```

### 9.5 用户搜索（添加绑定弹窗用）

添加绑定弹窗中的 "用户" 下拉需要远程搜索。**复用现有 API，无需新建：**

```
GET /v1/users?keyword=zhang&page=1&size=20
```

前端 `el-select` 的 `remote-method` 调用此接口，搜索结果格式化为 `username (displayName)` 显示。仅 admin 可调用（已有 `RequireAdmin()` 保护）。

---

## 十、前端页面设计图

### 10.1 集群列表页改造（ClusterList.vue）

每行新增 "成员数" 列和 "成员" 操作按钮。

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  集群管理                                                                   │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ ResourceListToolbar                                                   │  │
│  │                                                                       │  │
│  │ 搜索集群...                                         共 3 个集群        │  │
│  │                                                                       │  │
│  │  [+ 添加集群]  [删除 (0)]                          [⟳ 刷新控件]        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ el-card "table-card"                                                  │  │
│  │                                                                       │  │
│  │ ┌────┬───────────┬────────┬────────┬────────┬────────┬──────────────┐│  │
│  │ │ ☐  │ 集群名称   │ 状态    │ 版本    │ 节点数  │ 成员数  │    操作     ││  │
│  │ ├────┼───────────┼────────┼────────┼────────┼────────┼──────────────┤│  │
│  │ │ ☐  │ prod-clus │ 🟢在线 │ v1.28  │   5    │   6    │ 👥 成员     ││  │
│  │ │    │ ter       │        │        │        │        │ ✏️ 编辑 🗑 删除││  │
│  │ ├────┼───────────┼────────┼────────┼────────┼────────┼──────────────┤│  │
│  │ │ ☐  │ staging-  │ 🟢在线 │ v1.27  │   3    │   4    │ 👥 成员     ││  │
│  │ │    │ cluster   │        │        │        │        │ ✏️ 编辑 🗑 删除││  │
│  │ ├────┼───────────┼────────┼────────┼────────┼────────┼──────────────┤│  │
│  │ │ ☐  │ dev-clus  │ 🔴离线 │ —      │   —    │   2    │ 👥 成员     ││  │
│  │ │    │ ter       │        │        │        │        │ ✏️ 编辑 🗑 删除││  │
│  │ └────┴───────────┴────────┴────────┴────────┴────────┴──────────────┘│  │
│  │                                                                       │  │
│  │                                        ◀ 1 ▶                         │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.2 集群成员管理弹窗（ClusterMembersDialog.vue）

宽度 960px，集群级和 NS 级在同一张表中，用 "作用域" + "命名空间" 列区分。

**权限控制：** 所有登录用户可打开弹窗查看成员列表；"添加成员"、"编辑"、"删除" 按钮仅 `isSuperAdmin` 可见。

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  成员管理: prod-cluster                                                [✕]  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ [+ 添加成员]  [批量删除 (0)]   搜索用户...       共 6 条    [⟳ 刷新]  │  │
│  │                                                                       │  │
│  │ 作用域筛选: ( ● 全部  ○ 集群级  ○ 命名空间级 )                        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ el-table stripe                                                       │  │
│  │                                                                       │  │
│  │ ┌────┬──────────┬──────────┬────────┬──────────┬─────────────┬──────┐│  │
│  │ │ ☐  │ 用户      │ 显示名称  │ 作用域  │ 命名空间  │ 角色         │ 操作 ││  │
│  │ ├────┼──────────┼──────────┼────────┼──────────┼─────────────┼──────┤│  │
│  │ │ ☐  │ admin    │ 系统管理 │ 集群级  │    —     │ ┌─────────┐│ 编辑 ││  │
│  │ │    │          │ 员       │        │          │ │平台管理员││ 删除 ││  │
│  │ │    │          │          │        │          │ └─────────┘│      ││  │
│  │ │    │          │          │        │          │   (红色)    │      ││  │
│  │ ├────┼──────────┼──────────┼────────┼──────────┼─────────────┼──────┤│  │
│  │ │ ☐  │ zhangsan │ 张三     │ 集群级  │    —     │ ┌─────────┐│ 编辑 ││  │
│  │ │    │          │          │        │          │ │集群编辑者││ 删除 ││  │
│  │ │    │          │          │        │          │ └─────────┘│      ││  │
│  │ │    │          │          │        │          │   (蓝色)    │      ││  │
│  │ ├────┼──────────┼──────────┼────────┼──────────┼─────────────┼──────┤│  │
│  │ │ ☐  │ zhaoliu  │ 赵六     │ 集群级  │    —     │ ┌─────────┐│ 编辑 ││  │
│  │ │    │          │          │        │          │ │集群观察者││ 删除 ││  │
│  │ │    │          │          │        │          │ └─────────┘│      ││  │
│  │ │    │          │          │        │          │   (灰色)    │      ││  │
│  │ ├────┼──────────┼──────────┼────────┼──────────┼─────────────┼──────┤│  │
│  │ │ ☐  │ lisi     │ 李四     │ NS级   │ default  │ ┌─────────┐│ 编辑 ││  │
│  │ │    │          │          │        │          │ │空间管理员││ 删除 ││  │
│  │ │    │          │          │        │          │ └─────────┘│      ││  │
│  │ │    │          │          │        │          │  (浅橙色)   │      ││  │
│  │ ├────┼──────────┼──────────┼────────┼──────────┼─────────────┼──────┤│  │
│  │ │ ☐  │ lisi     │ 李四     │ NS级   │monitoring│ ┌─────────┐│ 编辑 ││  │
│  │ │    │          │          │        │          │ │空间编辑者││ 删除 ││  │
│  │ │    │          │          │        │          │ └─────────┘│      ││  │
│  │ │    │          │          │        │          │  (浅蓝色)   │      ││  │
│  │ ├────┼──────────┼──────────┼────────┼──────────┼─────────────┼──────┤│  │
│  │ │ ☐  │ wangwu   │ 王五     │ NS级   │ staging  │ ┌─────────┐│ 编辑 ││  │
│  │ │    │          │          │        │          │ │空间观察者││ 删除 ││  │
│  │ │    │          │          │        │          │ └─────────┘│      ││  │
│  │ │    │          │          │        │          │  (浅灰色)   │      ││  │
│  │ └────┴──────────┴──────────┴────────┴──────────┴─────────────┴──────┘│  │
│  │                                                                       │  │
│  │                                              共 6 条    ◀ 1 ▶        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.3 空状态

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  成员管理: dev-cluster                                                 [✕]  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ [+ 添加成员]  [批量删除 (0)]   搜索用户...       共 0 条    [⟳ 刷新]  │  │
│  │                                                                       │  │
│  │ 作用域筛选: ( ● 全部  ○ 集群级  ○ 命名空间级 )                        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                                                                       │  │
│  │                                                                       │  │
│  │                          ┌──────────────────┐                         │  │
│  │                          │   🔒 暂无成员     │                         │  │
│  │                          │                  │                         │  │
│  │                          │  点击上方「添加   │                         │  │
│  │                          │  成员」开始配置   │                         │  │
│  │                          │  集群权限         │                         │  │
│  │                          └──────────────────┘                         │  │
│  │                                                                       │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.4 添加绑定弹窗（AddBindingDialog.vue）

宽度 640px。集群详情页进入时集群字段自动锁定。

```
┌────────────────────────────────────────────────────────────────┐
│  添加成员                                                 [✕]  │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  用户        ┌────────────────────────────────────────────┐    │
│              │ 🔍 zhangsan (张三)                    [▼] │    │
│              └────────────────────────────────────────────┘    │
│              el-select filterable remote                        │
│              输入关键字远程搜索用户列表                           │
│              选项格式: username (displayName)                   │
│                                                                │
│  集群        ┌────────────────────────────────────────────┐    │
│              │ prod-cluster                           [▼] │    │
│              └────────────────────────────────────────────┘    │
│              集群列表页进入时 → disabled + 自动填充              │
│              全局页面进入时 → el-select 可选                    │
│                                                                │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─   │
│                                                                │
│  作用域      ( ● 集群级    ○ 命名空间级 )                      │
│              el-radio-group                                    │
│              切换时清空已选角色                                  │
│                                                                │
│  命名空间    ┌────────────────────────────────────────────┐    │
│  (条件显示)  │ 🔍 default                             [▼] │    │
│              └────────────────────────────────────────────┘    │
│              仅 "命名空间级" 时显示（v-show）                   │
│              el-select filterable                               │
│              选项: 该集群的 namespace 列表（从 K8s API 获取）    │
│              加载失败时允许手动输入:                              │
│              ┌────────────────────────────────────────────┐    │
│              │ ⚠️ 集群离线，无法获取命名空间列表              │    │
│              │ 请输入命名空间名称:                           │    │
│              │ ┌────────────────────────────────────────┐ │    │
│              │ │ 请输入...                               │ │    │
│              │ └────────────────────────────────────────┘ │    │
│              └────────────────────────────────────────────┘    │
│                                                                │
│  角色        ┌────────────────────────────────────────────┐    │
│              │ 集群编辑者 (cluster-editor)           [▼] │    │
│              └────────────────────────────────────────────┘    │
│              按作用域自动过滤角色列表:                           │
│              集群级 → cluster-admin / editor / viewer          │
│              NS级   → ns-admin / editor / viewer               │
│              选项格式: displayName (name)                      │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ 💡 集群编辑者: 集群内读写资源，不能删除，不能管理节点。   │ │
│  │    可使用终端和日志查看。                                 │ │
│  └──────────────────────────────────────────────────────────┘ │
│  el-alert type="info"                                          │
│  选中角色后自动显示该角色的权限摘要                              │
│                                                                │
│                                        [ 取消 ]  [ 确定 ]     │
└────────────────────────────────────────────────────────────────┘
```

### 10.5 编辑绑定弹窗（EditBindingDialog.vue）

宽度 520px。只有 "角色" 字段可编辑。

```
┌────────────────────────────────────────────────────────────────┐
│  编辑绑定                                                 [✕]  │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  用户        zhangsan (张三)              [el-input disabled]  │
│                                                                │
│  集群        prod-cluster                 [el-input disabled]  │
│                                                                │
│  作用域      集群级                       [el-input disabled]  │
│                                                                │
│  命名空间    —                            [el-input disabled]  │
│              (集群级时显示 —，NS级时显示NS名)                    │
│                                                                │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─   │
│                                                                │
│  角色        ┌────────────────────────────────────────────┐    │
│              │ 集群管理员 (cluster-admin)            [▼] │    │
│              └────────────────────────────────────────────┘    │
│              el-select，按当前作用域过滤可选角色                 │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ 💡 集群管理员: 集群内所有操作权限，包含节点管理、资源删除 │ │
│  │    等高危操作。                                           │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ ⚠️ 如需更改作用域（集群级 ↔ 命名空间级），请删除后重新添加。│ │
│  └──────────────────────────────────────────────────────────┘ │
│  el-alert type="warning"                                       │
│                                                                │
│                                        [ 取消 ]  [ 保存 ]     │
└────────────────────────────────────────────────────────────────┘
```

### 10.6 批量删除确认

```
┌──────────────────────────────────────────────────────────────┐
│  确认批量删除                                           [✕]  │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ⚠️  即将删除以下 3 条权限绑定：                              │
│                                                              │
│  ┌──────────┬──────────┬──────────┬────────────────────┐     │
│  │ 用户      │ 作用域    │ 命名空间  │ 角色                │     │
│  ├──────────┼──────────┼──────────┼────────────────────┤     │
│  │ zhangsan │ 集群级    │    —     │ ⚠️ 集群管理员       │     │
│  │ lisi     │ NS级     │ default  │ 空间编辑者          │     │
│  │ wangwu   │ NS级     │ staging  │ 空间观察者          │     │
│  └──────────┴──────────┴──────────┴────────────────────┘     │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ ⚠️ 其中包含 1 个管理员级别角色。删除后该用户将失去     │   │
│  │    对应集群的管理权限。请确认操作。                     │   │
│  └──────────────────────────────────────────────────────┘   │
│  el-alert type="warning" (仅当包含 admin 角色时显示)         │
│                                                              │
│                                [ 取消 ]  [ 确认删除 ]        │
└──────────────────────────────────────────────────────────────┘
```

### 10.7 角色标签色板

```
集群级角色（实色）:
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │ 平台管理员    │  │ 集群管理员    │  │ 集群编辑者    │  │ 集群观察者    │
  │ el-tag        │  │ el-tag        │  │ el-tag        │  │ el-tag        │
  │ type="danger" │  │ type="warning"│  │ type="primary"│  │ type="info"   │
  │ 红色 #F56C6C  │  │ 橙色 #E6A23C  │  │ 蓝色 #409EFF  │  │ 灰色 #909399  │
  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘

命名空间级角色（plain 镂空）:
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │ 空间管理员    │  │ 空间编辑者    │  │ 空间观察者    │
  │ el-tag plain  │  │ el-tag plain  │  │ el-tag plain  │
  │ type="warning"│  │ type="primary"│  │ type="info"   │
  │ 橙色镂空      │  │ 蓝色镂空      │  │ 灰色镂空      │
  └──────────────┘  └──────────────┘  └──────────────┘

规则: 集群级 = 实色，NS级 = plain 镂空，一眼区分层级。
```

---

## 十一、前端样式规范

### 11.1 设计体系基础

所有 RBAC 组件必须使用 gkube 设计 Token（`src/styles/tokens.css`），禁止硬编码颜色/间距值。

**引用方式：**

```css
/* ✅ 正确 */
color: var(--gk-color-text-primary);
padding: var(--gk-space-4);

/* ❌ 错误 */
color: #0f172a;
padding: 16px;
```

### 11.2 组件样式规范

#### 11.2.1 页面容器（ClusterMembersDialog 内部）

```css
/* 对齐现有 page-container 模式 */
.dialog-body {
  padding: var(--gk-space-5);          /* 20px */
  background: var(--gk-color-bg-page); /* 主题自适应 */
}
```

#### 11.2.2 工具栏区域

```css
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--gk-space-3);              /* 12px */
  padding-bottom: var(--gk-space-4);   /* 16px */
  border-bottom: 1px solid var(--gk-color-border-light);
  margin-bottom: var(--gk-space-4);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);              /* 8px */
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--gk-space-3);              /* 12px */
}

/* 搜索框 */
.toolbar .el-input {
  width: 240px;
}

/* 作用域筛选 Radio */
.toolbar .el-radio-group {
  margin-left: var(--gk-space-4);
}
```

#### 11.2.3 表格

```css
/* 对齐现有 table-card 模式 */
.table-card {
  border-radius: var(--gk-radius-md);  /* 8px */
}

/* el-table 统一配置 */
/* stripe 属性开启斑马纹 */
/* v-loading="loading" 加载时显示遮罩 */
/* :header-cell-style 统一表头样式 */
:deep(.el-table th.el-table__cell) {
  background: var(--gk-color-bg-page);
  color: var(--gk-color-text-secondary);
  font-weight: 600;
  font-size: var(--gk-font-size-sm);   /* 13px */
}

:deep(.el-table td.el-table__cell) {
  font-size: var(--gk-font-size-base); /* 14px */
  color: var(--gk-color-text-primary);
}
```

#### 11.2.4 弹窗（Dialog）

```css
/* ClusterMembersDialog — 成员管理 */
/* width="960px" */
/* :close-on-click-modal="false" */
/* destroy-on-close */

/* AddBindingDialog — 添加成员 */
/* width="640px" */
/* :close-on-click-modal="false" */
/* destroy-on-close */

/* EditBindingDialog — 编辑绑定 */
/* width="520px" */
/* :close-on-click-modal="false" */
/* destroy-on-close */

/* 批量删除确认 */
/* width="560px" */
/* :close-on-click-modal="false" */

/* 弹窗内部间距 */
:deep(.el-dialog__body) {
  padding: var(--gk-space-5);          /* 20px */
}

/* 弹窗底部按钮 */
:deep(.el-dialog__footer) {
  padding: var(--gk-space-4) var(--gk-space-5);
  border-top: 1px solid var(--gk-color-border-light);
}

/* 提交按钮 loading 状态：
   <el-button type="primary" :loading="saving" @click="handleSubmit">
     {{ bindingId ? '保存' : '确定' }}
   </el-button>
   saving = ref(false)，API 调用期间设为 true，完成后 false
*/
```

#### 11.2.5 表单（Form）

```css
/* 对齐现有 UserList.vue 表单模式 */
/* label-width="100px" */
/* label-position="right" */

/* 表单项间距 */
:deep(.el-form-item) {
  margin-bottom: var(--gk-space-5);    /* 20px */
}

/* 标签文字 */
:deep(.el-form-item__label) {
  color: var(--gk-color-text-secondary);
  font-size: var(--gk-font-size-base); /* 14px */
}
```

**AddBindingDialog 表单校验规则：**

```typescript
const rules: FormRules = {
  userId: [
    { required: true, message: '请选择用户', trigger: 'change' }
  ],
  roleId: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ],
  clusterName: [
    { required: true, message: '请选择集群', trigger: 'change' }
  ],
  // namespace: 仅 scopeType=namespace 时动态添加 required 规则
}

// 动态添加 namespace 校验
watch(() => form.scopeType, (val) => {
  if (val === 'namespace') {
    rules.namespace = [{ required: true, message: '请选择或输入命名空间', trigger: 'change' }]
  } else {
    rules.namespace = []
    form.namespace = ''
  }
})
```

**EditBindingDialog 表单校验规则：**

```typescript
const rules: FormRules = {
  roleId: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}
```

/* 分隔线（作用域选择区域前） */
.form-divider {
  border: none;
  border-top: 1px dashed var(--gk-color-border);
  margin: var(--gk-space-5) 0;
}

/* 用户远程搜索下拉：
   <el-select
     v-model="form.userId"
     filterable
     remote
     :remote-method="searchUsers"
     :loading="userSearchLoading"
     placeholder="输入用户名搜索"
   >
     <el-option
       v-for="u in userOptions"
       :key="u.id"
       :label="`${u.username} (${u.display_name})`"
       :value="u.id"
     />
   </el-select>

   userSearchLoading = ref(false)
   searchUsers(keyword) → 调用 GET /v1/users?keyword=xxx
   300ms debounce
*/

/* 集群下拉（添加弹窗中，集群列表页进入时 disabled）：
   数据来源：useClusterStore().clusters（Pinia store，已有）
   <el-select v-model="form.clusterName" :disabled="!!presetClusterName">
     <el-option
       v-for="c in clusterStore.clusters"
       :key="c.name"
       :label="c.name"
       :value="c.name"
     />
   </el-select>
   ClusterMembersDialog 打开 AddBindingDialog 时传入 presetClusterName
*/

/* 命名空间下拉（scopeType=namespace 时显示）：
   数据来源：GET /v1/k8s/namespace/list?clusterName=xxx
   API 客户端：复用现有 api/resource.ts 中的通用请求方法
   <el-select v-model="form.namespace" filterable :loading="nsLoading">
     <el-option v-for="ns in nsList" :key="ns" :label="ns" :value="ns" />
   </el-select>
   nsLoading = ref(false)
   watch(() => form.clusterName, async (name) => {
     if (!name) return
     nsLoading.value = true
     try {
       const res = await getNamespaceList(name)  // GET /k8s/namespace/list?clusterName=xxx
       nsList.value = res.data.items.map(ns => ns.metadata.name)
     } catch {
       nsLoadError.value = true  // 降级为手动输入
     } finally {
       nsLoading.value = false
     }
   })
*/

/* 命名空间下拉加载失败降级：
   集群离线时 NS 下拉加载失败，显示手动输入框替代：
   <template v-if="nsLoadError">
     <el-alert type="warning" :closable="false">
       集群离线，无法获取命名空间列表
     </el-alert>
     <el-input v-model="form.namespace" placeholder="请输入命名空间名称" />
   </template>
   <el-select v-else v-model="form.namespace" filterable>
     ...
   </el-select>
*/

/* 角色说明区域（el-alert），仅选中角色后显示：
   <el-alert
     v-if="selectedRole"
     type="info"
     :closable="false"
     class="role-description"
   >
     {{ selectedRole.displayName }}: {{ roleDescriptionMap[selectedRole.name] }}
   </el-alert>
*/

/* roleDescriptionMap 定义（与 i18n 翻译 key 对齐）：
   const roleDescriptionMap: Record<string, string> = {
     'super-admin':    '平台管理员，跳过所有权限检查。',
     'cluster-admin':  '集群内所有操作权限，包含节点管理、资源删除等高危操作。',
     'cluster-editor': '集群内读写资源，不能删除，不能管理节点。可使用终端和日志。',
     'cluster-viewer': '集群内只读。可使用终端和日志。',
     'ns-admin':       '该命名空间内所有操作权限。',
     'ns-editor':      '该命名空间内读写资源，不能删除。可使用终端和日志。',
     'ns-viewer':      '该命名空间内只读，不能使用终端。',
   }
*/
```

#### 11.2.6 操作列按钮

```css
/* 对齐现有 action-buttons 模式 */
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: var(--gk-space-1);              /* 4px */
}

.action-buttons .el-button + .el-button {
  margin-left: 0;
}

/* 按钮规格：
   - 主操作（添加成员）: type="success", size="default", <el-icon><Plus /></el-icon>
   - 行操作（编辑）    : size="small"
   - 行操作（删除）    : type="danger", plain, size="small"
   - 批量删除          : type="danger", size="default", :disabled="!selectedRows.length"
   - 确认弹窗（确认删除）: type="danger"
   - 确认弹窗（取消）  : type="default"
   - 弹窗底部（取消）  : type="default"
   - 弹窗底部（确定）  : type="primary", :loading="saving"
*/
```

### 11.3 角色标签样式

```vue
<!-- 集群级角色：实色 -->
<el-tag :type="roleTagType(roleName)" size="small">
  {{ roleDisplayName }}
</el-tag>

<!-- 命名空间级角色：plain 镂空 -->
<el-tag :type="roleTagType(roleName)" size="small" effect="plain">
  {{ roleDisplayName }}
</el-tag>
```

```typescript
function roleTagType(roleName: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    'super-admin':    'danger',
    'cluster-admin':  'warning',
    'cluster-editor': '',        // primary (Element Plus 默认 primary 用 '')
    'cluster-viewer': 'info',
    'ns-admin':       'warning',
    'ns-editor':      '',
    'ns-viewer':      'info',
  }
  return map[roleName] ?? 'info'
}

function isClusterRole(scopeType: string): boolean {
  return scopeType === 'cluster'
}
```

```vue
<!-- 使用示例 -->
<el-tag
  :type="roleTagType(row.roleName)"
  size="small"
  :effect="isClusterRole(row.scopeType) ? 'dark' : 'plain'"
>
  {{ row.roleDisplayName }}
</el-tag>
```

### 11.4 空状态

```vue
<el-empty description="暂无成员">
  <template #description>
    <p>暂无成员</p>
    <p style="color: var(--gk-color-text-secondary); font-size: var(--gk-font-size-sm);">
      点击上方「添加成员」开始配置集群权限
    </p>
  </template>
  <el-button type="primary" @click="openAdd">
    <el-icon><Plus /></el-icon> 添加成员
  </el-button>
</el-empty>
```

### 11.5 分页

```vue
<!-- 对齐现有 pagination 模式 -->
<div v-if="total > pageSize" class="pagination">
  <el-pagination
    :current-page="page"
    :page-size="pageSize"
    :total="total"
    layout="prev, pager, next"
    @current-change="handlePageChange"
  />
</div>
```

```css
.pagination {
  display: flex;
  justify-content: flex-end;
  padding: var(--gk-space-3) 0;        /* 12px */
  border-top: 1px solid var(--gk-color-border-light);
}
```

### 11.6 主题兼容

所有样式使用 `var(--gk-*)` 变量，自动适配 light/dark 主题。**不需要编写任何主题条件样式。**

| 元素 | Token | Light 值 | Dark 值 |
|------|-------|---------|---------|
| 弹窗背景 | `--gk-color-bg-card` | `#ffffff` | `#1e293b` |
| 表格表头 | `--gk-color-bg-page` | `#f8fafc` | `#0f172a` |
| 主文字 | `--gk-color-text-primary` | `#0f172a` | `#f1f5f9` |
| 次文字 | `--gk-color-text-secondary` | `#64748b` | `#94a3b8` |
| 边框 | `--gk-color-border` | `#e2e8f0` | `#334155` |

> **暗色主题额外处理：** 表格 `stripe` 的斑马纹行在暗色主题下使用 `--gk-neutral-100`（暗色中为 `#1e293b`），与 `--gk-color-bg-card` 相同，导致斑马纹不可见。
>
> **修复：** 在 `src/styles/themes/dark.css` 中增加（全局生效，所有 el-table 受益）：
> ```css
> [data-theme="dark"] .el-table--striped .el-table__body tr.el-table__row--striped td.el-table__cell {
>   background: var(--gk-neutral-50); /* 比 card bg 略深 */
> }
> ```

### 11.7 响应式

| 断点 | 行为 |
|------|------|
| ≥1200px | 弹窗使用设计宽度（960/640/520px） |
| 768px~1199px | 弹窗宽度自适应 `width="90%"`，表格横向滚动 |
| <768px | 弹窗全屏，表格隐藏次要列（命名空间、显示名称），筛选折叠 |

```typescript
// ClusterMembersDialog.vue
import { useMediaQuery } from '@vueuse/core'

const isMobile = useMediaQuery('(max-width: 767px)')
const isTablet = useMediaQuery('(max-width: 1199px)')

const dialogWidth = computed(() => {
  if (isMobile.value) return '100%'
  if (isTablet.value) return '90%'
  return '960px'
})
```

> **备选方案：** 如不引入 `@vueuse/core`，可用 `window.matchMedia` + `ref`：
> ```typescript
> const isMobile = ref(window.matchMedia('(max-width: 767px)').matches)
> window.matchMedia('(max-width: 767px)').addEventListener('change', e => { isMobile.value = e.matches })
> ```

### 11.8 图标规范

使用 Element Plus 内置图标，与现有页面一致：

| 场景 | 图标 | 引用 |
|------|------|------|
| 添加成员 | `Plus` | `@element-plus/icons-vue` |
| 删除 | `Delete` | `@element-plus/icons-vue` |
| 编辑 | `Edit` | `@element-plus/icons-vue` |
| 成员管理入口 | `User` | `@element-plus/icons-vue` |
| 刷新 | `Refresh` | `@element-plus/icons-vue` |
| 搜索 | `Search` | `@element-plus/icons-vue`（el-input 内置） |

### 11.9 交互状态

| 元素 | 状态 | 样式 |
|------|------|------|
| 表格行 | hover | Element Plus 默认 hover 高亮 |
| 表格行 | selected | 复选框选中，行保持默认色 |
| 编辑按钮 | hover | Element Plus 默认 primary hover |
| 删除按钮 | hover | Element Plus 默认 danger hover |
| 角色下拉 | disabled 选项 | 已选中的角色灰色不可选 |
| 添加成员按钮 | loading | `el-button :loading="saving"` |
| 表单必填项 | 校验失败 | Element Plus 默认红色边框 + 提示文字 |

### 11.10 CSS 文件组织

```css
/* ClusterMembersDialog.vue <style scoped> */
<style scoped>
.toolbar { ... }
.toolbar-left { ... }
.toolbar-right { ... }
.empty-state { ... }
.pagination { ... }
</style>

/* AddBindingDialog.vue <style scoped> */
<style scoped>
.form-divider { ... }
.role-description { ... }
.ns-fallback { ... }
</style>
```

**所有样式使用 `<style scoped>`**，避免全局污染。使用 Element Plus 的 `:deep()` 穿透修改子组件样式。

### 11.11 i18n 翻译 Key 清单

在 `src/locales/zh-CN.ts` 和 `en.ts` 中新增：

```typescript
// zh-CN.ts
sidebar: {
  // 现有 key 之后新增：
  rbac: '权限管理',
},
rbac: {
  // 成员管理弹窗
  memberManagement: '成员管理',
  addMember: '添加成员',
  batchDelete: '批量删除',
  scopeFilter: '作用域筛选',
  all: '全部',
  clusterScope: '集群级',
  namespaceScope: '命名空间级',
  username: '用户',
  displayName: '显示名称',
  scope: '作用域',
  namespace: '命名空间',
  role: '角色',
  noMembers: '暂无成员',
  noMembersDesc: '点击上方「添加成员」开始配置集群权限',

  // 添加/编辑弹窗
  addUser: '添加成员',
  editBinding: '编辑绑定',
  selectUser: '请选择用户',
  selectCluster: '请选择集群',
  selectNamespace: '请选择或输入命名空间',
  selectRole: '请选择角色',
  scopeChangeWarning: '如需更改作用域（集群级 ↔ 命名空间级），请删除后重新添加。',
  clusterOfflineNsHint: '集群离线，无法获取命名空间列表',
  inputNamespace: '请输入命名空间名称',

  // 批量删除确认
  confirmBatchDelete: '确认批量删除',
  batchDeleteWarning: '即将删除以下 {count} 条权限绑定：',
  adminRoleWarning: '其中包含 {count} 个管理员级别角色。删除后该用户将失去对应集群的管理权限。请确认操作。',
  bindingExists: '该用户在此作用域已有绑定，请编辑现有绑定',

  // 角色名
  superAdmin: '平台管理员',
  clusterAdmin: '集群管理员',
  clusterEditor: '集群编辑者',
  clusterViewer: '集群观察者',
  nsAdmin: '空间管理员',
  nsEditor: '空间编辑者',
  nsViewer: '空间观察者',

  // 集群列表
  members: '成员',
  memberCount: '成员数',
}
```

```typescript
// en.ts
sidebar: {
  rbac: 'Permissions',
},
rbac: {
  memberManagement: 'Member Management',
  addMember: 'Add Member',
  batchDelete: 'Batch Delete',
  scopeFilter: 'Scope Filter',
  all: 'All',
  clusterScope: 'Cluster',
  namespaceScope: 'Namespace',
  username: 'User',
  displayName: 'Display Name',
  scope: 'Scope',
  namespace: 'Namespace',
  role: 'Role',
  noMembers: 'No Members',
  noMembersDesc: 'Click "Add Member" above to configure cluster permissions',
  addUser: 'Add Member',
  editBinding: 'Edit Binding',
  selectUser: 'Select user',
  selectCluster: 'Select cluster',
  selectNamespace: 'Select or enter namespace',
  selectRole: 'Select role',
  scopeChangeWarning: 'To change scope (cluster ↔ namespace), please delete and re-add.',
  clusterOfflineNsHint: 'Cluster offline, unable to fetch namespace list',
  inputNamespace: 'Enter namespace name',
  confirmBatchDelete: 'Confirm Batch Delete',
  batchDeleteWarning: 'Will delete {count} permission binding(s):',
  adminRoleWarning: 'Includes {count} admin-level role(s). The user will lose management permissions for the corresponding cluster. Please confirm.',
  bindingExists: 'This user already has a binding in this scope. Please edit the existing binding.',
  superAdmin: 'Platform Admin',
  clusterAdmin: 'Cluster Admin',
  clusterEditor: 'Cluster Editor',
  clusterViewer: 'Cluster Viewer',
  nsAdmin: 'NS Admin',
  nsEditor: 'NS Editor',
  nsViewer: 'NS Viewer',
  members: 'Members',
  memberCount: 'Members',
}
```

---

## 十二、向后兼容策略

| 点 | 策略 |
|---|---|
| `RequireAdmin()` | 保留不动，用于用户管理、集群管理、RBAC 写操作 |
| `auth.IsAdmin()` | 保留，扩展为同时检查 config 白名单 + DB is_super_admin |
| 现有用户权限 | 上线后默认无 K8s 权限，admin 手动分配 |
| 前端 `isAdmin` | 改为 `isSuperAdmin`，同步更新前端 store |
| Dashboard 路由 | 保持无权限检查，任何登录用户可查看 |
| K8s 读路由 | 从无中间件 → 新增 `RequirePermission()`，**未分配权限的用户将无法查看任何 K8s 资源** |
| 集群离线 | 查看/编辑/删除绑定正常（数据在 MySQL）；NS 下拉降级为手动输入 |
| 删除用户/集群 | 前提示关联绑定数量，确认后级联删除 |

> **⚠️ 升级注意：** K8s GET 路由从无权限检查变为需要 `RequirePermission()`。升级后，除 super_admin 外的所有用户将无法查看 K8s 资源，直到 admin 在成员管理中为其分配角色。建议升级前先为常用用户预设好绑定数据。

---

## 十三、边界场景处理

| 场景 | 处理方式 |
|------|---------|
| 集群离线时添加 NS 级绑定 | NS 下拉降级为手动输入模式 |
| 删除用户时有关联绑定 | 前端 ElMessageBox 提示绑定数，确认后**硬删除关联 bindings**（`db.Where("user_id = ?", id).Delete(&PermissionBinding{})`），再软删除用户 |
| 删除集群时有关联绑定 | 同上：硬删除关联 bindings（`db.Where("cluster_name = ?", name).Delete(&PermissionBinding{})`），再删除集群 |
| 添加重复绑定 (user+cluster+ns) | 后端返回 409，前端提示 "已有绑定，请编辑" |
| Role.Permissions JSON 损坏 | GetPermissionsMap() 返回空 map + 日志，不 panic |
| 权限缓存一致性 | 绑定变更时主动失效该用户缓存；TTL 5 分钟兜底 |
| POST body 被中间件消费 | 用 ShouldBindBodyWith（Gin 内部缓存 body），不影响 handler |

---

## 十四、实施分期

### Phase 1 — 核心 RBAC 引擎（本次实现）

```
后端:
├── internal/rbac/model/role.go         ← Role 模型 + 预置角色
├── internal/rbac/model/binding.go      ← PermissionBinding 模型
├── internal/rbac/handler.go            ← RBAC handler (6 个 API)
├── internal/router/rbac.go             ← RBAC 路由注册 + AuditLog
├── pkg/middleware/permission.go         ← RequirePermission 中间件
├── pkg/middleware/audit.go              ← parseK8sPath 支持 /v1/rbac/
├── pkg/auth/permcache.go               ← 权限缓存
├── cmd/migrate.go                      ← +Role +PermissionBinding
├── cmd/seed.go                         ← +预置角色 +is_super_admin
├── internal/auth/model/user.go         ← +IsSuperAdmin 字段
├── internal/router/k8s.go              ← RequireAdmin → RequirePermission
├── internal/router/router.go           ← +registerRbacRoutes
└── pkg/auth/keys.go                    ← IsAdmin 扩展

前端:
├── api/rbac.ts                         ← RBAC API 客户端
├── views/cluster/ClusterMembersDialog  ← 成员管理弹窗
├── views/cluster/AddBindingDialog      ← 添加/编辑绑定弹窗
├── stores/auth.ts                      ← +isSuperAdmin +permissions
├── views/cluster/ClusterList.vue       ← +成员按钮 +成员数列
├── views/config/secret/SecretDetail.vue ← Secret 值 masking
└── locales/zh-CN.ts + en.ts            ← +RBAC 翻译
```

**总计：13 个后端文件 + 7 个前端文件 = 20 个文件。**

### Phase 1.5 — 完善（后续）

- 用户列表页权限 tag + 权限详情弹窗
- 全局 `/rbac` 页面（角色管理 + 权限矩阵）
- NS 详情页成员面板
- 权限调试端点 `GET /v1/rbac/check`（等效 `kubectl auth can-i`）
- Force Delete 独立 verb（区分普通删除和强制删除）

### Phase 2 — 扩展

- 自定义角色 CRUD + 权限矩阵编辑器（参考 Rancher RoleTemplate）
- 用户组支持（参考 Kuboard 权限组 / Portainer Team）
  - 需新增 `user_group` + `user_group_member` 表
  - `permission_binding` 表需扩展：`user_id` → `subject_type` + `subject_id`（破坏性迁移）
- Project 抽象层（NS 组，参考 Rancher Project）
- LDAP/OIDC 集成
- 权限变更通知（邮件/webhook）

---

## 十五、检查清单

| # | 检查项 | 状态 |
|---|--------|:----:|
| 1 | Role 表 UNIQUE(name) | ✅ |
| 2 | PermissionBinding UNIQUE(user_id, cluster_name, namespace) | ✅ |
| 3 | namespace 字段 NOT NULL DEFAULT '' | ✅ |
| 4 | 中间件用 ShouldBindBodyWith 读 POST body namespace | ✅ |
| 5 | 不需要前端 namespace 注入到 query params | ✅ |
| 6 | 添加弹窗切换作用域时清空已选角色 | ✅ |
| 7 | 编辑弹窗只允许改角色 | ✅ |
| 8 | 集群离线时 NS 下拉降级为手动输入 | ✅ |
| 9 | 批量删除列出受影响绑定 + 高危角色警告 | ✅ |
| 10 | 删除用户/集群前检查关联绑定 | ✅ |
| 11 | 角色 JSON 损坏时返回空 map 不 panic | ✅ |
| 12 | 权限缓存绑定变更时主动失效 | ✅ |
| 13 | 空状态引导提示 | ✅ |
| 14 | 角色色板: 集群级实色，NS级镂空 | ✅ |
| 15 | RBAC 写路由 RequireAdmin，读路由 JWTAuth | ✅ 修订 |
| 16 | terminal 权限组 verb 统一为 ["terminal"] | ✅ 修订 |
| 17 | GET K8s 路由新增 RequirePermission（非替换） | ✅ 修订 |
| 18 | 登录流程 fetchPermissions() 在路由守卫中调用 | ✅ 修订 |
| 19 | canAccess() 集群级绑定覆盖所有 NS | ✅ 修订 |
| 20 | 用户搜索复用 GET /v1/users?keyword= | ✅ 修订 |
| 21 | Dashboard 路由保持无权限检查 | ✅ 修订 |
| 22 | PUT binding 校验新角色 scopeType 与原绑定一致 | ✅ 修订 |
| 23 | POST binding 响应用 200 非 201（对齐 gkube 约定） | ✅ 修订 |
| 24 | RBAC 写路由挂 AuditLog 中间件 | ✅ 新增 |
| 25 | audit 中间件 parseK8sPath 支持 /v1/rbac/ 前缀 | ✅ 新增 |
| 26 | SecretDetail.vue 值 masking（默认隐藏，点击显示） | ✅ 新增 |
| 27 | 审计日志作用域注释（集群级资源说明） | ✅ 新增 |
| 28 | Phase 1.5 增加权限调试端点 + Force Delete verb | ✅ 新增 |
| 29 | Phase 2 增加 Project 抽象层 + 权限变更通知 | ✅ 新增 |
| 30 | isSuperAdmin 与 bindings 一起缓存，不每次查 DB | ✅ 修订 |
| 31 | 前端章节子编号修正（8.x→9.x, 9.x→10.x） | ✅ 修订 |
| 32 | cluster-members 路由安全说明（所有用户可见） | ✅ 修订 |
| 33 | Phase 2 用户组需 binding 表破坏性迁移说明 | ✅ 新增 |
| 34 | 前端样式规范章节（Token引用/组件样式/主题兼容/响应式） | ✅ 新增 |
| 35 | 响应式用 useMediaQuery 替代 window.innerWidth | ✅ 修订 |
| 36 | 暗色主题斑马纹修复放 dark.css 全局 | ✅ 修订 |
| 37 | 弹窗提交按钮 loading 状态规范 | ✅ 新增 |
| 38 | 用户远程搜索下拉 loading + 降级方案 | ✅ 新增 |
| 39 | 操作列按钮完整规格（含图标/loading/disabled） | ✅ 修订 |
| 40 | 表单校验规则完整定义（AddBinding + EditBinding） | ✅ 新增 |
| 41 | 集群下拉数据来源（useClusterStore） | ✅ 新增 |
| 42 | 角色说明区域未选中时隐藏逻辑 | ✅ 新增 |
| 43 | 删除用户/集群时硬删除关联 bindings | ✅ 修订 |
| 44 | fetchPermissions() 失败时降级为空权限 | ✅ 新增 |
| 45 | i18n 翻译 Key 完整清单（zh-CN + en） | ✅ 新增 |
| 46 | resolveResourceGroup 未知路径增加 Warn 日志 | ✅ K8s 架构师终审 |
| 47 | Handler 代码片段完整 imports（含 fmt/strings/net/http） | ✅ 资深开发审查 |
| 48 | failNotFound 替换为 response.FailWithStatus(404) | ✅ 资深开发审查 |
| 49 | API 契约完整错误响应表（POST/PUT/DELETE 全场景） | ✅ 资深开发审查 |
| 50 | POST body namespace 缺省语义说明（缺省="" 集群级） | ✅ 资深开发审查 |
| 51 | Handler Params 结构体完整定义 | ✅ 资深开发审查 |
| 52 | CreateBinding 完整实现示例（含 isDuplicateErr） | ✅ 资深开发审查 |
| 53 | Handler imports 去除未使用的 strconv/gorm，补全 strings | ✅ 编译验证 |
| 54 | CachedPermissions.Bindings 类型修正为 model.PermissionBinding | ✅ 编译验证 |
| 55 | fetchPermissions() 前端完整实现代码 | ✅ 实现补全 |
| 56 | request.ts 403 拦截器代码 | ✅ 实现补全 |
| 57 | parseID 工具函数（UpdateBinding/DeleteBinding 共用） | ✅ 实现补全 |
| 58 | hasRole() 完整实现 + 与 canAccess 区别说明 | ✅ 实现补全 |
| 59 | roleDescriptionMap 完整定义（7 个角色描述） | ✅ 实现补全 |
| 60 | 命名空间列表 API 端点明确（GET /k8s/namespace/list） | ✅ 实现补全 |
