# gkube RBAC 修复任务书（源码核对版）

> **用途：** 本文档交给 AI 执行修复。每条问题均于 2026-09-02 对源码逐行核实，行号为当前工作区实际行号（按锚点代码定位更稳妥，改前先 `grep` 确认代码未漂移）。
> **来源：** RBAC 前后端设计评审。P0 问题已用真实中间件代码的集成测试复现，修复方案亦已用真实代码验证通过（验证用的测试文件已清理，工作区干净）。
> **验证要求：** 后端每项改完执行 `cd backend && go build ./...`；前端改完执行 `cd frontend && npm run build`（vue-tsc 捕获类型错误）。涉及行为的项有"验收"小节，必须逐条自测。
> **总指令：** 只做本文档列出的事，不做"顺手改进"。每项修复相互独立，按序执行。

---

## Batch 0 — P0：所有 JSON 写操作已断裂（先修这个）

### 0.1 中间件消费 body 后未恢复，下游 handler 读到 EOF（维持 P0）

- **位置：** `backend/pkg/middleware/permission.go:104-119`（`extractNamespace`）
- **问题机理（已测试复现）：** 前端写请求（如 `updateDeploymentYaml`，见 `frontend/src/api/resource.ts:537`）只在 **JSON body** 里带 `namespace`，query 只有 axios 拦截器注入的 `clusterName`。`extractNamespace` 调用 `c.ShouldBindBodyWith(&body, binding.JSON)`——gin v1.10 的该函数把 body 字节缓存进 `c.Keys`（`BodyBytesKey`），**但不重置 `c.Request.Body`**。下游 handler 全部使用 `c.ShouldBindJSON`（直接读 `c.Request.Body` 流，已耗尽）→ EOF → "参数校验失败"。
- **已复现证据：** 用真实 `RequirePermission` + 模拟 handler 的集成测试：ns-editor 用户 PUT update-yaml，RBAC 放行（HTTP 200），handler `ShouldBindJSON` 返回 `"EOF"`。超管路径同样命中（`extractNamespace` 在 admin 检查之前执行，见 `permission.go:39`）。
- **影响：** deployment/pod/secret/configmap/service/ingress/… 所有 JSON body 的 POST/PUT/DELETE 全部失败。limitrange/resourcequota 等少数 query 风格路由幸免。
- **修复（方案已验证）：** 调用 `ShouldBindBodyWith` 触发 body 缓存后，用缓存副本重置 `c.Request.Body`。**注意必须无条件重置**（包括绑定失败或 namespace 为空的分支），否则该分支下 body 仍然丢失。

`backend/pkg/middleware/permission.go` 当前代码：

```go
// extractNamespace 从请求中提取命名空间。
// GET 从 query 参数读，POST/PUT/DELETE 从 body 读（ShouldBindBodyWith 不消费 body）。
func extractNamespace(c *gin.Context) string {
	if ns := c.Query("namespace"); ns != "" {
		return ns
	}
	if c.Request.Method != "GET" {
		var body struct {
			Namespace string `json:"namespace"`
		}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err == nil && body.Namespace != "" {
			return body.Namespace
		}
	}
	return ""
}
```

替换为：

```go
// extractNamespace 从请求中提取命名空间。
// GET 从 query 参数读，POST/PUT/DELETE 从 body 读。
// ShouldBindBodyWith 会把 body 缓存进 context（BodyBytesKey），但 c.Request.Body
// 流本身已被消费——必须用缓存副本重置回去，否则下游 handler 的
// ShouldBindJSON 会读到 EOF（历史上曾因此导致全部 JSON 写接口 400）。
func extractNamespace(c *gin.Context) string {
	if ns := c.Query("namespace"); ns != "" {
		return ns
	}
	if c.Request.Method != "GET" {
		var body struct {
			Namespace string `json:"namespace"`
		}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err == nil && body.Namespace != "" {
			restoreRequestBody(c)
			return body.Namespace
		}
		restoreRequestBody(c)
	}
	return ""
}

// restoreRequestBody 用 gin 缓存的 body 副本重置 c.Request.Body，
// 保证下游 handler 仍可通过 ShouldBindJSON 读到完整请求体。
func restoreRequestBody(c *gin.Context) {
	if raw, ok := c.Get(gin.BodyBytesKey); ok {
		if bodyBytes, ok := raw.([]byte); ok {
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}
}
```

import 区（`permission.go:3-16`）补 `"bytes"` 和 `"io"`。

- **回归测试（必做，入库保留）：** 新建 `backend/pkg/middleware/extract_namespace_test.go`：

```go
package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	rbacmodel "gkube/internal/rbac/model"
	"gkube/pkg/auth"
)

// TestExtractNamespaceBodyRegression 回归：RequirePermission 读过 body 后，
// 下游 handler 的 ShouldBindJSON 必须仍能读到完整 body。
// 历史缺陷：extractNamespace 用 ShouldBindBodyWith 消费 c.Request.Body 后未恢复，
// 导致全部 JSON 写接口返回"参数校验失败"。
func TestExtractNamespaceBodyRegression(t *testing.T) {
	gin.SetMode(gin.TestMode)
	clusterIDCache.Store("c1", uint(1))

	editorRole := rbacmodel.Role{
		ID: 11, Name: "ns-editor", ScopeType: "namespace", IsSystem: true,
		Permissions: `{"workload":["read","create","update"]}`,
	}
	auth.SetPermissionsForTest(uint(3), &auth.CachedPermissions{
		IsSuperAdmin: false,
		Bindings: []rbacmodel.PermissionBinding{{
			ID: 2, UserID: 3, RoleID: 11, ClusterID: 1, Namespace: "default",
			Role: editorRole,
		}},
		ExpireAt: time.Now().Add(time.Hour),
	})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(3))
		c.Set("username", "editoruser")
		c.Next()
	})
	r.Use(RequirePermission())
	r.PUT("/v1/k8s/deployment/update-yaml", func(c *gin.Context) {
		var body struct {
			ClusterName string `json:"clusterName" binding:"required"`
			Namespace   string `json:"namespace"`
			Name        string `json:"name" binding:"required"`
			Yaml        string `json:"yaml" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			t.Errorf("handler ShouldBindJSON 失败（body 被中间件消费）: %v", err)
			return
		}
		if body.Namespace != "default" || body.Name != "nginx" {
			t.Errorf("body 字段值异常: %+v", body)
		}
	})

	req := httptest.NewRequest("PUT", "/v1/k8s/deployment/update-yaml?clusterName=c1",
		strings.NewReader(`{"clusterName":"c1","namespace":"default","name":"nginx","yaml":"a: b"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)
}
```

`backend/pkg/auth/permcache.go` 文件末尾追加测试辅助（本文档唯一显式授权的"顺手改动"，回归测试必需）：

```go
// SetPermissionsForTest 测试辅助：直接写入权限缓存（仅测试用）。
func SetPermissionsForTest(userID uint, cp *CachedPermissions) {
	permCache.Store(userID, cp)
}
```

- **验收：** `go test ./pkg/middleware/ -run TestExtractNamespaceBodyRegression -v` 通过。
- **陷阱提示：** 修复时常见错误是只在 `body.Namespace != ""` 分支恢复 body——空 namespace 的写请求同样需要恢复，`restoreRequestBody` 必须两个分支都调用（或用 defer）。
- **优化建议（可选）：** 改用 `defer restoreRequestBody(c)` 更简洁安全——放在 `if c.Request.Method != "GET"` 块开头，无论后续哪个分支都会自动恢复，杜绝遗漏。未消费 body 时调用 restore 是无害操作（body 未变，等价于 no-op）。

---

## Batch 1 — P1 安全（立即修）

### 1.1 审计路由完全绕过权限模型 + 任意用户可伪造审计条目

- **位置：** `backend/internal/router/k8s.go:295-302`（`registerAuditRoutes`）
- **问题：** `audit/list`、`audit/detail`、`audit/stats`、`audit/create` 四条路由无任何权限中间件（仅 `clear` 有 RequireAdmin）。设计矩阵（`docs/superpowers/specs/2026-08-30-rbac-design.md` 3.1/3.2 节）明确 audit 仅集群级角色可读。尤其 `POST /v1/k8s/audit/create` 允许任意登录用户写入任意 action/resource/status/details 的审计条目——审计日志作为安全兜底被污染。已核实前端对 `audit/create` 零调用。
- **修复步骤：**
  1. `pkg/middleware/permission.go` 的 `resolveResourceGroup`（:121-182）switch 中，在 `case strings.Contains(path, "/crd/"):` 之前插入：

  ```go
  case strings.Contains(path, "/audit/"):
      return "audit"
  ```

  2. `internal/router/k8s.go` `registerAuditRoutes` 改为：

  ```go
  func registerAuditRoutes(rg *gin.RouterGroup) {
      // 审计属集群级只读资源：集群级角色有 audit:read 权限，ns 角色无
      rg.GET("audit/list", middleware.RequirePermission(), k8s.Audit.ListAuditLogs)
      rg.GET("audit/detail", middleware.RequirePermission(), k8s.Audit.GetAuditLog)
      rg.GET("audit/stats", middleware.RequirePermission(), k8s.Audit.GetAuditStats)
      // 审计清除属高危操作,需管理员
      rg.DELETE("audit/clear", middleware.RequireAdmin(), k8s.Audit.ClearAuditLogs)
  }
  ```

  3. **删除 `POST /v1/k8s/audit/create` 路由及其 handler：** 删 `internal/router/k8s.go` 中 `rg.POST("audit/create", k8s.Audit.CreateAuditLog)` 一行；删 `internal/k8s/audit.go` 的 `func (h *auditHandler) CreateAuditLog`（:235-249）整个函数；删 `pkg/middleware/audit.go` 的 `auditSkipPaths`（:14-18）中 `"audit/create": true,` 条目。
- **验收：** `go build ./...` 通过（确认 CreateAuditLog 删除后无悬空引用，`grep -rn "CreateAuditLog" backend/` 清零）。有环境时：ns-viewer token 调 `GET /v1/k8s/audit/list` → 403；cluster-viewer → 200；`POST /v1/k8s/audit/create` → 404。
- **遗留提醒：** 删路由和 handler 后，ES 中已通过 `audit/create` 写入的脏数据仍存在。不影响系统功能，但如需彻底清理可执行一次性 ES 查询删除（`action: create_audit_log` 或自定义筛选）。非阻塞项，可视安全审计需求决定是否处理。

### 1.2 Dashboard 路由无权限控制（零权限用户可看全部集群数据）

- **位置：** `backend/internal/router/dashboard.go`（6 条路由仅 JWT）+ `backend/internal/dashboard/dashboard.go:81-93`（`getTargetClusters`）
- **问题：** 零绑定用户可读所有 online 集群的资源统计与跨命名空间事件内容（Events 聚合每集群前 N 条 K8s 事件全文）。
- **修复：** 在 `getTargetClusters` 内按调用用户绑定过滤集群（dashboard 是聚合视图，不硬造 resourceGroup 语义）：
  1. 修改 `getTargetClusters` 签名为 `getTargetClusters(c *gin.Context, clusterID *uint)`，拆出原逻辑为私有 `queryTargetClusters`：

```go
// getTargetClusters 按 clusterID 选取目标集群，并按当前用户的 RBAC 绑定过滤：
//   - 超管（config 白名单或 DB 标记）：不过滤；
//   - 普通用户：仅返回其存在绑定（cluster_id 命中）的集群；
//   - clusterID 命中但用户无绑定：返回空切片（零值结果，不报错）；
//   - clusterID 为空：返回用户可见的 online 集群（无绑定时为空，聚合显示 0）。
func getTargetClusters(c *gin.Context, clusterID *uint) ([]model.K8SCluster, error) {
	username, _ := c.Get("username")
	name, _ := username.(string)
	userIDVal, _ := c.Get("userID")
	userID, _ := userIDVal.(uint)

	if auth.IsAdmin(name) {
		return queryTargetClusters(clusterID)
	}
	cached := auth.GetUserPermissions(userID)
	// 注意：此处 cached.IsSuperAdmin 与上方 auth.IsAdmin 为"或"关系——
	// auth.IsAdmin 检查 config 白名单，cached.IsSuperAdmin 检查 DB 标记。
	// 两者来源不同但语义一致（均为超管），保留双保险。
	if cached.IsSuperAdmin {
		return queryTargetClusters(clusterID)
	}

	allowed := make(map[uint]bool, len(cached.Bindings))
	for _, b := range cached.Bindings {
		allowed[b.ClusterID] = true
	}
	if len(allowed) == 0 {
		return []model.K8SCluster{}, nil
	}

	clusters, err := queryTargetClusters(clusterID)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.K8SCluster, 0, len(clusters))
	for _, cl := range clusters {
		if allowed[cl.ID] {
			filtered = append(filtered, cl)
		}
	}
	return filtered, nil
}

// queryTargetClusters 原有语义：clusterID 命中返回单个集群，否则返回 online 集群。
func queryTargetClusters(clusterID *uint) ([]model.K8SCluster, error) {
	var clusters []model.K8SCluster
	if clusterID != nil {
		if err := database.DB.Where("id = ?", *clusterID).Find(&clusters).Error; err != nil {
			return nil, err
		}
		return clusters, nil
	}
	if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}
```

  2. 6 处调用点（`dashboard.go` :129 :185 :276 :343 :488 :734）同步改为 `getTargetClusters(c, query.ClusterID)`。
  3. dashboard.go import 区补 `"gkube/pkg/auth"`（当前未导入）。
- **验收：** 零绑定用户 `GET /v1/dashboard/overview` → clusterCount 为 0、events 为空数组；有 cluster-1 绑定的用户 → 只见 cluster-1 数据；超管 → 不变。

### 1.3 前端无法分配集群级角色（核心功能不可达）

- **位置：** `frontend/src/views/cluster/AddBindingDialog.vue:60-62`（`filteredRoles` 只留 namespace 级）+ :53-57（namespaces 必填 `min: 1`）
- **问题：** cluster-admin/editor/viewer 三种角色无任何 UI 入口，设计权限层级一半是摆设。后端 `CreateBinding`（`internal/rbac/handler.go:186-189`）已支持集群级绑定（`nsList = []string{""}`）。
- **修复：**
  1. **角色下拉不再按 scopeType 过滤，改为分组展示。** 删除 `filteredRoles` computed（:60-62），替换为：

```typescript
const roleGroups = computed(() => {
  const ns = props.roles.filter(r => r.scopeType === 'namespace')
  const cluster = props.roles.filter(r => r.scopeType === 'cluster')
  return [
    { label: t('rbac.clusterScope'), roles: cluster },
    { label: t('rbac.namespaceScope'), roles: ns },
  ].filter(g => g.roles.length > 0)
})
```

  2. **namespaces 改为条件必填**（集群级角色隐藏命名空间选择）。`rules` 计算属性（:53-57）替换为：

```typescript
const rules = computed<FormRules>(() => ({
  userId: [{ required: true, message: t('rbac.selectUser'), trigger: 'change' }],
  roleId: [{ required: true, message: t('rbac.selectRole'), trigger: 'change' }],
  namespaces: selectedRole.value?.scopeType === 'cluster'
    ? []
    : [{ required: true, type: 'array', min: 1, message: t('rbac.selectNamespace'), trigger: 'change' }],
}))
```

  （`selectedRole` computed 已存在于 :64-66，保持不动。）
  3. **模板：** 命名空间的 `el-form-item`（:190-213）加 `v-if="!isEditMode && selectedRole?.scopeType !== 'cluster'"`；角色 select（:237-244）的 `<el-option v-for="r in filteredRoles" ...>` 替换为：

```html
<el-option-group v-for="g in roleGroups" :key="g.label" :label="g.label">
  <el-option
    v-for="r in g.roles"
    :key="r.id"
    :label="`${r.displayName} (${r.name})`"
    :value="r.id"
  />
</el-option-group>
```

  4. **handleSubmit（:141-165）**：删除 for 循环逐条调 API 的逻辑，`else`（新建）分支整体替换为：

```typescript
} else {
  const isClusterRole = selectedRole.value?.scopeType === 'cluster'
  await createBinding({
    userId: Number(form.value.userId),
    roleId: Number(form.value.roleId),
    clusterId: props.clusterId,
    // 集群级角色不传 namespaces 字段（后端自动置 [""]）；
    // 命名空间级角色过滤掉前端全选标记 '__all__'，只传实际选中的 ns
    ...(isClusterRole ? {} : { namespaces: form.value.namespaces.filter(ns => ns !== '__all__') }),
  })
  ElMessage.success(t('rbac.createSuccess'))
}
```

  （`api/rbac.ts` 的 `createBinding` 入参类型 :16-22 已含 `namespaces?: string[]`，无需改。此步同时消掉"前端循环 N 次"的问题。）
  5. **编辑模式**：`watch(() => props.binding, ...)`（:100-107）保留现状（编辑只改角色，后端 `UpdateBinding` 有 scope 一致性校验）。`roleDescriptionMap`（:68-72）补齐集群级 3 个角色：

```typescript
const roleDescriptionMap: Record<string, string> = {
  'cluster-admin': t('rbac.roleDesc.clusterAdmin'),
  'cluster-editor': t('rbac.roleDesc.clusterEditor'),
  'cluster-viewer': t('rbac.roleDesc.clusterViewer'),
  'ns-admin':  t('rbac.roleDesc.nsAdmin'),
  'ns-editor': t('rbac.roleDesc.nsEditor'),
  'ns-viewer': t('rbac.roleDesc.nsViewer'),
}
```

  （i18n 的 `roleDesc.clusterAdmin` 等键已存在，`frontend/src/locales/zh-CN.ts:797-805` 已核实。）
  6. **i18n 新增 2 个键。** `zh-CN.ts` rbac 区（锚点 :774 `scopeChangeWarning` 附近）加：

```typescript
clusterScope: '集群级角色',
namespaceScope: '命名空间级角色',
```

  `en.ts` rbac 区（锚点 :749 `scopeChangeWarning`）加：

```typescript
clusterScope: 'Cluster-scoped roles',
namespaceScope: 'Namespace-scoped roles',
```

- **验收：** `npm run build` 通过。手动流程（有环境时）：添加成员 → 角色下拉出现两组 → 选"集群观察者" → 命名空间选择消失 → 提交成功 → `permission_binding` 表新增一条 `namespace=""` 记录。

### 1.4 clusterIDCache 永不失效 → 删库重建同名集群可复活旧权限

- **位置：** `backend/pkg/middleware/permission.go:18-19`（`var clusterIDCache sync.Map`）+ `backend/internal/cluster/cluster.go:305-327`（Delete handler）
- **问题：** 无 TTL、集群删除时不清缓存，且 `permission_binding` 不随集群删除清理。同名重建的集群拿到新 ID，但中间件仍解析到缓存里的旧 ID → 旧 binding 匹配 → **用户在新集群上复活旧权限**，直到进程重启。
- **修复：**
  1. `pkg/middleware/permission.go` 在 `clusterIDCache` 声明（:18-19）之后新增：

```go
// InvalidateClusterIDCache 失效指定集群名的 ID 缓存（集群删除后调用）。
func InvalidateClusterIDCache(clusterName string) {
	if clusterName != "" {
		clusterIDCache.Delete(clusterName)
	}
}
```

  2. `internal/cluster/cluster.go` 的 `Delete` handler，在 `k8s.InvalidateClient(cluster.ClusterName)`（:324 附近）之后、`response.Success` 之前追加：

```go
	// 删除集群时清理权限绑定，防止孤儿数据；同名重建集群新 ID 不同，
	// 旧绑定本不应匹配，但必须同时清 clusterIDCache 防止旧 ID 残留命中
	// 注意：此处未使用事务——集群删除和绑定清理是两步独立操作。若集群已删但
	// 绑定清理失败（DB 异常），绑定变为孤儿数据（cluster_id 指向已删集群，
	// 永远不会被匹配到）。当前实现仅 log warn 继续，可接受；如需严格一致性，
	// 应将集群删除和绑定清理纳入同一事务。
	if err := database.DB.Where("cluster_id = ?", cluster.ID).Delete(&rbacmodel.PermissionBinding{}).Error; err != nil {
		logger.Warn(fmt.Sprintf("清理集群 %s 的权限绑定失败: %v", cluster.ClusterName, err))
	}
	middleware.InvalidateClusterIDCache(cluster.ClusterName)
```

  3. import 区（:11-17）补 `"gkube/internal/rbac/model"`（包内别名用 `rbacmodel`）和 `"gkube/pkg/middleware"`。
  4. **import cycle 已核实安全：** `internal/cluster` → `pkg/middleware` → `internal/cluster/model`，model 是叶子包不回导，无环（`internal/cluster` 已通过 `pkg/k8s` 存在同样的间接依赖模式，:14）。若 go build 意外报 cycle，说明代码已漂移，停下重查而不是绕。
- **验收：** `go build ./...` 通过。有环境时：给用户绑定集群 A（name=X）权限 → 删除集群 X → 重建同名集群 X → 该用户对 X 的所有请求 403（绑定已清、缓存已失效）。

### 1.5 用户禁用/删除后权限不回收

- **位置：** `backend/pkg/auth/permcache.go:35-40`（DB 查询不含用户状态）
- **问题：** 登录（`internal/auth/auth.go:38`）和 refresh（:89）都校验 `status = 1`，但 access token 有效期 2 小时（`pkg/auth/jwt.go:32`），期间被禁用/软删除的用户持有未过期 token 仍可操作——权限缓存加载不查用户状态，`GetUserPermissions` 照常返回绑定。
- **修复：** `permcache.go` 第 2 步（:35-40）的两段查询整体替换为：

```go
	// 2. 查 DB：先校验用户可用（status=1 且未软删除，GORM DeletedAt 自动过滤），
	//    不可用用户视为无任何权限
	var user authmodel.User
	if err := database.DB.Select("id", "is_super_admin").
		Where("id = ? AND status = 1", userID).
		First(&user).Error; err != nil {
		// 用户不存在/被禁用/被删除：返回空权限，短 TTL（1 分钟）以便恢复后快速生效
		cp := &CachedPermissions{
			IsSuperAdmin: false,
			Bindings:     nil,
			ExpireAt:     time.Now().Add(time.Minute),
		}
		permCache.Store(userID, cp)
		return cp
	}

	var bindings []rbacmodel.PermissionBinding
	database.DB.Preload("Role").Where("user_id = ?", userID).Find(&bindings)
```

  同时第 3 步构造 `CachedPermissions` 的 `IsSuperAdmin: isSuperAdmin` 改为 `IsSuperAdmin: user.IsSuperAdmin`（原 `var isSuperAdmin bool` 查询已删除）。import 补 `authmodel "gkube/internal/auth/model"`（无环：`pkg/auth` 已 import `internal/rbac/model`，`internal/auth/model` 与之同为叶子包）。
  - **短 TTL 理由：** 禁用是管理员操作，被禁用户最多再放行 1 分钟（缓存过期重查即拒）；解禁后也最多 1 分钟恢复，比固定 5 分钟体验好。
  - **残余窗口说明（接受的边界）：** 禁用后 access token 剩余寿命内（≤2h）请求仍会命中内存里的旧缓存，最多到本缓存条目的 `ExpireAt`（≤5 分钟）。即本修复把窗口从"最长 2 小时"压缩到"最长 5 分钟"，彻底清零需要 token 版本号机制，不在本任务书。
- **验收：** 有环境时：禁用用户 → ≤5 分钟后其 `/k8s/*` 请求 403、`my-permissions` 返回空 bindings。

### 1.6 node 写操作 verb 映射缺失（集群管理员改不了节点标签）

- **位置：** `backend/internal/rbac/model/role.go:42`（`clusterAdminPerms`）+ `backend/pkg/middleware/permission.go:191-218`（`specialVerbOverrides`）
- **问题：** `PUT /k8s/node/labels`、`PUT /k8s/node/update-yaml` 落默认 verb "update"，但 cluster-admin 的 node verbs 是 `["read","cordon","taint","drain","delete"]`——**连集群管理员都不能改节点标签/YAML**，仅超管可用。前端有调用（`resource.ts:390,402`）。
- **修复：** `specialVerbOverrides` 的"节点管理"区（:192-195）追加两行：

```go
	{"/node/labels", "PUT", "taint"},
	{"/node/update-yaml", "PUT", "taint"},
```

  映射到 "taint" 的理由：与 cordon/taint/drain 同属"节点高危变更"语义，cluster-admin 已有该 verb，cluster-editor/viewer（只有 node:read）自然被拒，符合设计矩阵 3.1 节 node 行（editor/viewer 仅 R）。
- **验收：** 走查 verb 解析：`PUT /v1/k8s/node/labels` → "taint" → cluster-admin 放行、cluster-editor 403。

### 1.7 ClusterMembersDialog 数据源缺陷：getBindings 分页 + 权限限制导致数据丢失与操作失效

- **位置：** `frontend/src/views/cluster/ClusterMembersDialog.vue:82-146`（`fetchData` 函数）
- **问题（3 个关联缺陷）：**
  1. **分页截断 + 误标 unbound：** `getBindings({ clusterId })` 未传 `page`/`size`，后端默认 `page=1, size=20`（`handler.go:103-110`）。集群绑定 >20 条时，`getBindings` 只返回前 20 条。`getClusterMembers` 返回全部成员，但 `existingIds` 只基于前 20 条绑定构建（:122），多出的绑定用户被错误标记为 `unbound: true`——其真实绑定 ID 丢失，编辑/删除操作使用 `member-xxx` 字符串 ID 调 `DELETE /v1/rbac/bindings/member-123`，后端 `ParseUint` 失败返回错误。
  2. **非管理员 403：** `GET /rbac/bindings` 挂 `RequireAdmin` 中间件（`router/rbac.go:18`），非管理员用户调用 403。`Promise.allSettled` 虽不阻塞，但 `bindings` 数组为空，所有成员走 unbound 逻辑——数据能展示，但每个成员都没有真实绑定 ID，编辑/删除全部失效。
  3. **数据冗余：** `getClusterMembers` 已返回展示所需的全部信息（username、displayName、roleName、namespace），且不分页、不要求管理员权限。`getBindings` 的引入增加了复杂度，还引入了分页不一致的 bug。
- **修复方案（简化数据源）：** 去掉 `getBindings` 调用，改用 `getClusterMembers` 作为列表唯一数据源；编辑/删除时按需查询绑定。

  1. **`fetchData` 函数（:82-146）替换为：**

```typescript
async function fetchData() {
  loading.value = true
  try {
    const [membersRes, rolesRes] = await Promise.allSettled([
      getClusterMembers(props.clusterId),
      getRoles(),
    ])

    if (membersRes.status === 'fulfilled') {
      const mRes: any = membersRes.value
      const mData = mRes?.data ?? mRes
      // getClusterMembers 返回 { items: [...], total }，每项含 userId/username/displayName/roleName/namespace
      bindings.value = (mData.items || []).map((m: any) => ({
        id: m.id,         // 真实绑定 ID（用于删除）
        userId: m.userId,
        username: m.username || '',
        displayName: m.displayName || '',
        roleName: m.roleName || '',
        roleDisplayName: m.roleDisplayName || m.roleName || '',
        namespace: m.namespace || '',
        scopeType: m.scopeType || (m.namespace ? 'namespace' : 'cluster'),
        isSuperAdmin: m.isSuperAdmin || false,
      }))
    }

    if (rolesRes.status === 'fulfilled') {
      const rRes: any = rolesRes.value
      const rData = rRes?.data ?? rRes
      roles.value = Array.isArray(rData) ? rData : (rData?.items || [])
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('rbac.loadFailed'))
  } finally {
    loading.value = false
  }
}
```

  2. **删除 import 中的 `getBindings`**（:6）和 unbound 相关逻辑（:116-139 整段删除）。
  3. **编辑功能适配：** `handleEdit`（:217-233）中 `editingBinding` 构造需从当前行数据直接取值（不再依赖 `_raw` 嵌套对象），因 `getClusterMembers` 返回的是扁平结构。编辑弹窗 `AddBindingDialog` 的 `binding` prop 在编辑模式下只用 `roleId` 和 `namespace`，来源改为当前行数据即可。

  4. **`handleDelete` 适配：** 确保 `row.id` 是真实绑定 ID（`getClusterMembers` 返回的 `id` 字段即 `permission_binding.id`，见 `handler.go:365`）。删除后调 `fetchData()` 刷新。

- **验收：** `npm run build` 通过。手动验收：
  - 集群有 >20 条绑定时，成员列表完整展示（无误标 unbound）。
  - 非管理员打开成员抽屉可正常查看列表（无 403 报错）。
  - 管理员编辑/删除操作正常工作（使用真实绑定 ID）。
  - 搜索过滤正常（按用户名/昵称/角色名/命名空间过滤）。

---

## Batch 2 — P2 逻辑/一致性

### 2.1 CreateBinding 空 namespace 绕过 scope 校验

- **位置：** `backend/internal/rbac/handler.go:165-189`
- **问题：** 传 `namespaces: [""]` 时：正则校验跳过空串（:174 `ns != ""`）、scope 校验只看 `len(nsList) == 0`（:181）不触发、`nsList` 保持 `[""]`，落库为 `Namespace=""` 的**集群级**绑定。ns-viewer 被绑成集群级 = 全集群只读。UpdateBinding 有 scope 一致性校验（:257-264），CreateBinding 没有。
- **修复：** 在 nsRegex 校验循环（:172-178）之后、"校验角色与作用域匹配"（:180）之前插入：

```go
	// 拒绝显式空串：namespaces=[""] 会绕过下方 scope 校验，静默创建集群级绑定
	for _, ns := range nsList {
		if ns == "" && role.ScopeType == "namespace" {
			response.Fail(c, "命名空间角色必须指定非空命名空间")
			return
		}
	}
```

  （cluster 级角色不传 ns 时后端自己置 `[]string{""}`（:187-189），那条路径合法，勿拦。）
- **验收：** `POST /v1/rbac/bindings` body 含 `"namespaces":[""]` + ns 角色 → 400。

### 2.2 节点列表页对 cluster-editor/viewer 403（路由分组与设计矩阵矛盾）

- **位置：** `backend/pkg/middleware/permission.go:174-175`（`/cluster/` → cluster_mgmt）
- **问题：** NodeList 页调 `GET /k8s/cluster/nodes`（`frontend/src/api/resource.ts:378`）→ resourceGroup=cluster_mgmt → 仅 cluster-admin 有 `cluster_mgmt:read`。但设计矩阵（3.1 节）给了 editor/viewer `node: R`。viewer 有 node:read 权限却打不开节点列表页，而 node/detail 又能过，自相矛盾。
- **修复：** `resolveResourceGroup` 中在 `case strings.Contains(path, "/cluster/"):` 之前插入：

```go
	case strings.Contains(path, "/cluster/nodes"):
		return "node"
```

  **⚠️ 顺序依赖：** switch 内 `Contains` 顺序即优先级，`/cluster/nodes` 必须排在 `/cluster/` 之前，否则会被 `/cluster/` 分支吃掉。建议在新增 case 处加注释：`// 注意：/cluster/nodes 必须在 /cluster/ 之前，否则节点路由误归 cluster_mgmt`。`/cluster/version` 保持 cluster_mgmt 不变（版本信息维持仅 cluster-admin；如产品想让 viewer 也看，另行决策）。
- **验收：** cluster-viewer 调 `GET /k8s/cluster/nodes` → 200；`GET /k8s/cluster/version` → 403（维持现状）。

### 2.3 terminal record 路由无 resourceGroup 映射（非超管全 403）

- **位置：** `backend/internal/router/k8s.go:71-72` + `backend/pkg/middleware/permission.go:121-182`
- **问题：** `/v1/k8s/container/record/list`、`/record/url` 不含任何已注册子串 → `resolveResourceGroup` 返回 "" → 所有非超管 403。前端零调用（已核实），属死功能。
- **修复：** 删除 `internal/router/k8s.go:71-72` 两行路由注册。**只删这两行 + 两个 handler 函数**：`internal/k8s/container.go` 中 `func RecordList`（:135-166）与 `func RecordUrl`（:168-197）。
  - **禁止扩大删除范围：** `model.TerminalRecord`（`internal/k8s/model/terminalRecord.go`）**必须保留**——已核实它还被 `HandleWebSocket` 的终端录制写入使用（`container.go:107,127`），删了会破坏终端审计录像。`cmd/migrate.go:32` 的迁移同样保留（表可能有线上数据）。
- **验收：** `go build ./...` 通过；`grep -rn "RecordList\|RecordUrl" backend/internal backend/pkg` 清零（model 定义除外）；终端 exec 功能不受影响（录制仍落库）。

### 2.4 前端按钮级权限缺失 + canAccess 语义错误

- **位置：** `frontend/src/views/config/secret/SecretDetail.vue:29-32`（唯一一处使用，语义错误）；`frontend/src/stores/auth.ts:98-113`（`hasRole` 无调用方）
- **问题：** 全前端只有 1 处权限判断且不区分角色：`canAccess` 只判断"有任意绑定"（viewer 也返回 true），SecretDetail 编辑按钮对只读用户可点，点了才 403。其余所有列表/详情页的删除/编辑/重启按钮对 viewer 全部可见。
- **修复（最小可行：提供正确 helper + 修现有唯一调用点）：**
  1. `stores/auth.ts` 在 `hasRole`（:98-113）之后新增并导出：

```typescript
  // 判断当前用户在 (clusterId, namespace) 是否有写权限（editor/admin 级）。
  // 集群级绑定（namespace=""）覆盖该集群所有 ns；角色名含 admin/editor 视为可写。
  function canWrite(clusterId: number, namespace?: string): boolean {
    if (!user.value) return false
    if (user.value.isSuperAdmin) return true
    if (!user.value.permissions) return false
    const writable = (roleName: string) => /admin|editor/.test(roleName)
    return user.value.permissions.some(p => {
      if (p.clusterId !== clusterId) return false
      if (p.namespace === '') return writable(p.roleName)
      if (namespace && p.namespace === namespace) return writable(p.roleName)
      return false
    })
  }
```

  return 语句（:115）同步加入 `canWrite`。
  2. `SecretDetail.vue:29-32` 的 `canWrite` computed 改为：

```typescript
const canWrite = computed(() => authStore.canWrite(Number(clusterStore.clusterId), namespace))
```

  3. **全面铺开（约 30 个视图的操作按钮显隐）不在本任务书**——需要产品决策（哪些操作算写、终端按钮对 viewer 的显隐等），单独立项。本项只保证 helper 正确 + 现有唯一调用点语义正确。
- **验收：** `npm run build` 通过。有环境时：ns-viewer 登录 → SecretDetail 编辑按钮 disabled 且 tooltip "权限不足"；ns-editor → 可点。

### 2.5 fetchPermissions 失败后 permissions=[] 永不重试

- **位置：** `frontend/src/router/index.ts:591-603` + `frontend/src/stores/auth.ts:80-82`
- **问题：** 瞬时网络错误导致本会话权限显示为空：守卫条件是 `!authStore.user.permissions`（`[]` 为 truthy→不再触发），永不重拉。用户看到自己"无任何权限"，实际只是网络抖动。
- **修复：** 语义改为"undefined=未拉取、null=拉取失败待重试、[]=已确认空权限"：
  1. `stores/auth.ts` 的 `fetchPermissions` catch 分支（:80-82）改为：

```typescript
    } catch {
      user.value = { ...user.value, permissions: null }
    }
```

  2. `router/index.ts` 守卫（:591-603）改为：

```typescript
  // 首次进入已认证页面时拉取权限；permissions === undefined 表示未拉取过。
  // 失败置 null（区别于已确认的空权限 []），下次路由切换自动重试。
  if (token) {
    const authStore = useAuthStore()
    if (authStore.user && authStore.user.permissions === undefined) {
      try {
        await authStore.fetchPermissions()
      } catch {
        if (authStore.user) {
          authStore.user.permissions = null
        }
      }
    }
  }
```

  3. `stores/auth.ts` 的 `PermissionBinding`/`UserInfo` 接口无需改（`permissions?: PermissionBinding[]` 对 null 的运行时行为靠下方判空兜底）；但 `canAccess`/`hasRole`/`canWrite` 中的 `if (!user.value.permissions) return false` 对 null/undefined 均安全，无需改。为类型一致，`UserInfo.permissions` 类型改为 `PermissionBinding[] | null`。
- **验收：** `npm run build` 通过。走查：断网进页面 → 恢复网络切路由 → 权限重新拉取成功。

### 2.6 RBAC 审计日志缺集群信息（body 是 clusterId，中间件只读 clusterName）

- **位置：** `backend/pkg/middleware/audit.go:88-96`
- **问题：** RBAC 绑定写操作 body 里是 `clusterId`（数字，见 `CreateBindingParams`，`internal/rbac/handler.go:28`），中间件只尝试读 `ClusterName`（string）→ 审计记录 cluster 字段为空。
- **修复：** body 提取段（:91-96）替换为：

```go
		// RBAC 路由的集群信息在 body（POST/PUT）中：兼容 clusterName 与 clusterId
		if cluster == "" && (method == "POST" || method == "PUT") {
			var body struct {
				ClusterName string `json:"clusterName"`
				ClusterID   uint   `json:"clusterId"`
			}
			if err := c.ShouldBindBodyWith(&body, binding.JSON); err == nil {
				if body.ClusterName != "" {
					cluster = body.ClusterName
				} else if body.ClusterID != 0 {
					cluster = strconv.FormatUint(uint64(body.ClusterID), 10)
				}
			}
		}
```

  （`strconv` 已在 import 区 :7。）注意：此处 `ShouldBindBodyWith` 同样消费 body，但 audit 中间件挂在 handler 之前、RBAC handler 用 `ShouldBindJSON`——**需要同样调用 0.1 的 `restoreRequestBody(c)`**。若 audit.go 无法引用（同包 `middleware`，可直接调），在上述代码块内 `ShouldBindBodyWith` 调用后补一行 `restoreRequestBody(c)`。K8s 路由的 handler 已核实全部用 query 参数风格读 clusterName（audit 相关 body 场景仅 RBAC 绑定），影响面可控。
  - **多次 ShouldBindBodyWith 无冲突：** 0.1（extractNamespace）和此处（audit 中间件）都可能对同一请求调用 `ShouldBindBodyWith`。gin 的实现是将 body 字节缓存进 `c.Keys[BodyBytesKey]`，后续调用命中缓存不会重复读流。两处各自调 `restoreRequestBody(c)` 也是安全的（都是从同一份缓存恢复）。实测过：同一请求经 RBAC 中间件（extractNamespace）+ audit 中间件后，handler 的 `ShouldBindJSON` 仍能正常读到 body。
- **验收：** 创建绑定后查审计记录，cluster 字段为 clusterId 数字字符串。

### 2.7 无法通过 API 提升 is_super_admin（设计与实现脱节）

- **位置：** `backend/internal/auth/user.go:22-28`（`UpdateUserParams`）+ :146-181（Update handler）
- **问题：** 设计文档 2.1 节承诺"后续 admin 可在前端对其他用户设置"，但参数结构不支持该字段，唯一超管是 seed 出来的。
- **修复：**
  1. `UpdateUserParams` 追加字段（保持现有指针可选风格）：

```go
	IsSuperAdmin *bool `json:"isSuperAdmin" label:"平台管理员"`
```

  2. `Update` handler 的 updates 构造区（:161-170）追加：

```go
	if p.IsSuperAdmin != nil {
		updates["is_super_admin"] = *p.IsSuperAdmin
	}
```

  3. 在 `database.DB.Model(&user).Updates(updates)` 成功后（:178 `}` 之后）、`response.Success` 之前追加缓存失效与自我降级保护：

```go
	// 提权/降权后失效该用户权限缓存，避免 5 分钟内权限不一致
	if p.IsSuperAdmin != nil {
		auth.InvalidateUserPermissions(p.ID)
	}
```

  （`"gkube/pkg/auth"` 已在 import 区 :8。）
  4. **自我降级保护**——`Update` handler 查出 `user`（:154-158）之后、构造 updates 之前插入：

```go
	// 不允许超管通过此接口降级自己（config 白名单超管不受影响，属可接受边界）
	if p.IsSuperAdmin != nil && !*p.IsSuperAdmin {
		if uid, ok := getUserID(c); ok && uid == p.ID && user.IsSuperAdmin {
			response.Fail(c, "不能降级当前登录的超管账号")
			return
		}
	}
```

  - **注意 `getUserID` 来源：** 执行前先 `grep -rn "func getUserID" backend/` 确认该 helper 是否存在。如不存在，替换为 `uidVal, ok := c.Get("userID"); uid, _ := uidVal.(uint)`。

  5. 路由权限兜底已存在：`PUT /users` 挂了 `RequireAdmin`（`internal/router/auth.go:27-33`），仅超管可调，无需改。
- **验收：** 超管调 `PUT /users` 传 `{"id":2,"isSuperAdmin":true}` → 用户 2 立即获得超管权限（缓存已失效）；传 `{"id":自己,"isSuperAdmin":false}` → 400 拒绝。

---

## Batch 3 — P3 小问题（低优先级）

### 3.1 resolveVerb 未使用参数

- **位置：** `backend/pkg/middleware/permission.go:221`
- **修复：** 签名 `func resolveVerb(resourceGroup, path, method string) string` 改为 `func resolveVerb(path, method string) string`，调用点（:78）同步改 `resolveVerb(path, c.Request.Method)`。

### 3.2 权限缓存单实例（多副本部署时不失效）

- **位置：** `backend/pkg/auth/permcache.go:20`
- **现状：** 当前单机部署无碍。**本任务书不修代码**，在 `permCache` 声明前补注释：

```go
// permCache 为进程内缓存。注意：多副本部署时 InvalidateUserPermissions 仅对本进程
// 生效，其他副本最长 5 分钟（TTL）后自然过期。横向扩容前需迁移到 Redis pub/sub 失效。
```

### 3.3 cluster-members 与 GET /users 的信息暴露

- **位置：** `backend/internal/rbac/handler.go:349-394`（ClusterMembers）
- **现状：** 有意设计（成员弹窗需要），已核对 :366-375 输出的 `memberInfo` 只含 username/displayName，无 email 泄露，前端对普通用户隐藏管理按钮。**不修**，维持现状。

### 3.4 ns 角色权限矩阵含死权限（crd/event 跨 ns 场景不可达）

- **位置：** `backend/internal/rbac/model/role.go:48`
- **问题：** CRD 是 cluster-scoped 资源，请求永远不带 namespace → ns 绑定永不匹配 → ns 角色的 `crd:read` 是死权限；`event:read` 仅在 `event/list?namespace=X` 精确匹配时可达。矩阵声明与实际可达性不符。
- **决策：** 属产品语义问题（ns 用户该不该看 CRD、列表页默认"全部命名空间"视图对 ns 用户 403 的体验问题同源，需要"前端按权限过滤默认视图/菜单"或"后端列表接口按绑定自动收窄"的产品决策）。**本任务书不修代码**，在 `role.go:48` 的注释区追加说明：

```go
// nsPerms 命名空间级角色权限 JSON。
// 注意：crd 为 cluster-scoped 资源，ns 角色的 crd:read 实际不可达（请求不带
// namespace，绑定永不匹配）；event:read 仅在 event/list?namespace=X 精确匹配时
// 生效。此处保留声明与设计矩阵一致，可达性问题待产品决策后另行处理。
```

---

## 执行顺序与依赖

```
Batch 0（P0）                 ← 必须最先，其余验证都依赖写路径可用
Batch 1.6 / 2.1 / 2.6        ← 纯后端小改，互相独立
Batch 1.1（审计路由）          ← 独立；2.6 与之同文件(middleware/audit.go)但无冲突
Batch 1.2（dashboard 过滤）   ← 独立
Batch 1.3（集群级角色 UI）     ← 独立（纯前端）
Batch 1.7（成员抽屉数据源）    ← 独立（纯前端）；与 1.3 同文件(ClusterMembersDialog)但改不同区域
Batch 2.4 / 2.5              ← 独立（纯前端）
Batch 1.4（集群删除清理）       ← 独立；改后必须 go build 验证无 import cycle
Batch 1.5（用户状态）           ← 独立
Batch 2.2 / 2.3              ← 独立（2.3 删代码前 grep 引用，注意 TerminalRecord 保留）
Batch 2.7（超管提升）           ← 独立
Batch 3                      ← 随时可做（3.1 是纯签名清理）
```

每完成一个 Batch：后端 `go build ./...` + `go vet ./...`；前端 `npm run build`。全部完成后：按各"验收"小节逐条走查（有运行环境时实际发请求验证；无环境时代码走查并在 PR 描述注明"验收为走查"）。

## 全局禁令

1. 不实施本文档未提及的顺手改进（0.1 的回归测试与 permcache 测试辅助是显式授权项）。
2. 不做"权限过滤菜单/按钮全面铺开"（见 2.4-3 与 3.4 的说明，需产品决策单独立项）。
3. 不引入新依赖。
4. 删除代码（1.1 的 audit/create、2.3 的 record 路由）前必须 `grep` 全仓库确认引用清零；2.3 明确禁止删 `TerminalRecord` model 与迁移。
5. 每处修改保持周边代码风格（中文注释、命名习惯）一致。
6. 找不到锚点代码时先 `grep` 确认代码未漂移，禁止凭行号盲改。
