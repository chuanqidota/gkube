# Feature Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove 6 non-core feature domains (Catalog, GitOps, Tenancy, Approval, Monitoring, Developer Tools, K8s RBAC pages) and the Redis dependency to reduce infrastructure requirements and maintenance burden.

**Architecture:** Single-pass cleanup removing frontend views/routes/sidebar items, backend API handlers/routes/models, and Redis infrastructure. Retained stack: MySQL + Elasticsearch + S3/MinIO.

**Tech Stack:** Vue 3 + TypeScript (frontend), Go + Gin + GORM (backend)

---

## File Structure

### Files to DELETE

**Frontend views (10 directories, 15 files):**
- `frontend/src/views/catalog/AppCatalog.vue`
- `frontend/src/views/gitops/GitOpsView.vue`
- `frontend/src/views/notification/NotificationCenter.vue`
- `frontend/src/views/watcher/ResourceWatcher.vue`
- `frontend/src/views/approval/ApprovalList.vue`
- `frontend/src/views/tenancy/TenantList.vue`
- `frontend/src/views/monitoring/MonitoringView.vue`
- `frontend/src/views/monitoring/MonitoringDashboard.vue`
- `frontend/src/views/monitoring/ResourceDashboard.vue`
- `frontend/src/views/monitoring/PrometheusView.vue`
- `frontend/src/views/topology/TopologyView.vue`
- `frontend/src/views/topology/TopologyGraph.vue`
- `frontend/src/views/tools/BatchOperations.vue`
- `frontend/src/views/tools/ResourceDiff.vue`
- `frontend/src/views/tools/YAMLEditor.vue`
- `frontend/src/views/rbac/RBACView.vue`
- `frontend/src/views/rbac/RBACMatrix.vue`

**Backend API handlers (8 files):**
- `backend/app/k8s/api/catalog.go`
- `backend/app/k8s/api/gitops.go`
- `backend/app/k8s/api/approval.go`
- `backend/app/k8s/api/tenancy.go`
- `backend/app/k8s/api/metrics.go`
- `backend/app/k8s/api/prometheus.go`
- `backend/app/k8s/api/topology.go`
- `backend/app/k8s/api/rbac.go`

**Backend model (1 file):**
- `backend/app/k8s/model/approval.go`

**Backend packages (4 directories):**
- `backend/pkg/redis/` (api.go)
- `backend/pkg/k8s/serviceaccount/` (api.go)
- `backend/pkg/k8s/clusterrole/` (api.go)
- `backend/pkg/k8s/rolebinding/` (api.go)

### Files to MODIFY

- `frontend/src/router/index.ts` — remove 12 route blocks
- `frontend/src/components/Layout/Sidebar.vue` — remove 9 menu items
- `frontend/src/locales/zh-CN.ts` — remove 9 translation sections + sidebar keys
- `frontend/src/locales/en.ts` — same
- `backend/router/router.go` — remove 8 route groups (~100 lines)
- `backend/cmd/root.go` — remove redis import + Init()
- `backend/config/config.go` — remove Redis struct field
- `backend/config/config.yaml` — remove redis section
- `backend/docker-compose.yaml` — remove redis service
- `backend/go.mod` — run `go mod tidy` to remove go-redis

---

## Task 1: Delete Frontend View Directories

**Files:**
- Delete: `frontend/src/views/catalog/` (entire directory)
- Delete: `frontend/src/views/gitops/` (entire directory)
- Delete: `frontend/src/views/notification/` (entire directory)
- Delete: `frontend/src/views/watcher/` (entire directory)
- Delete: `frontend/src/views/approval/` (entire directory)
- Delete: `frontend/src/views/tenancy/` (entire directory)
- Delete: `frontend/src/views/monitoring/` (entire directory)
- Delete: `frontend/src/views/topology/` (entire directory)
- Delete: `frontend/src/views/rbac/` (entire directory)
- Delete: `frontend/src/views/tools/BatchOperations.vue`
- Delete: `frontend/src/views/tools/ResourceDiff.vue`
- Delete: `frontend/src/views/tools/YAMLEditor.vue`

- [ ] **Step 1: Delete all frontend view directories and files**

```bash
cd /Users/zqqzqq/05_github/gkube
rm -rf frontend/src/views/catalog
rm -rf frontend/src/views/gitops
rm -rf frontend/src/views/notification
rm -rf frontend/src/views/watcher
rm -rf frontend/src/views/approval
rm -rf frontend/src/views/tenancy
rm -rf frontend/src/views/monitoring
rm -rf frontend/src/views/topology
rm -rf frontend/src/views/rbac
rm -f frontend/src/views/tools/BatchOperations.vue
rm -f frontend/src/views/tools/ResourceDiff.vue
rm -f frontend/src/views/tools/YAMLEditor.vue
```

- [ ] **Step 2: Verify tools directory is now empty and remove it**

```bash
ls frontend/src/views/tools/
# Should show nothing - remove the empty directory
rmdir frontend/src/views/tools 2>/dev/null || echo "Directory not empty or doesn't exist"
```

- [ ] **Step 3: Commit**

```bash
git add -A frontend/src/views/catalog frontend/src/views/gitops frontend/src/views/notification frontend/src/views/watcher frontend/src/views/approval frontend/src/views/tenancy frontend/src/views/monitoring frontend/src/views/topology frontend/src/views/rbac frontend/src/views/tools
git commit -m "chore: remove unused frontend view directories (catalog, gitops, notification, watcher, approval, tenancy, monitoring, topology, rbac, tools)"
```

---

## Task 2: Update Frontend Router

**Files:**
- Modify: `frontend/src/router/index.ts`

Remove these route blocks from the `children` array (lines 464-637):

| Lines | Route |
|-------|-------|
| 464-475 | RBAC (`/rbac`, `/rbac/matrix`) |
| 496-501 | Catalog (`/catalog`) |
| 503-508 | GitOps (`/gitops`) |
| 510-515 | Tenancy (`/tenancy`) |
| 517-522 | Approvals (`/approvals`) |
| 538-561 | Monitoring (`/monitoring`, `/monitoring/dashboard`, `/monitoring/resources`, `/monitoring/prometheus`) |
| 563-574 | Topology (`/topology`, `/topology/graph`) |
| 576-581 | Tools - Diff (`/tools/diff`) |
| 583-588 | Tools - YAML Editor (`/tools/yaml-editor`) |
| 590-595 | Tools - Batch (`/tools/batch`) |
| 625-630 | Notifications (`/notifications`) |
| 632-637 | Watcher (`/watcher`) |

- [ ] **Step 1: Remove RBAC routes (lines 463-475)**

Remove from `// RBAC` through the closing `},` of rbac/matrix:

```typescript
        // RBAC
        {
          path: 'rbac',
          name: 'RBACView',
          component: () => import('@/views/rbac/RBACView.vue'),
          meta: { title: 'RBAC', icon: 'UserFilled' },
        },
        {
          path: 'rbac/matrix',
          name: 'RBACMatrix',
          component: () => import('@/views/rbac/RBACMatrix.vue'),
          meta: { title: '权限矩阵', parent: 'RBACView' },
        },
```

- [ ] **Step 2: Remove Catalog route (lines 495-501)**

```typescript
        // Catalog
        {
          path: 'catalog',
          name: 'AppCatalog',
          component: () => import('@/views/catalog/AppCatalog.vue'),
          meta: { title: '应用商店', icon: 'Grid' },
        },
```

- [ ] **Step 3: Remove GitOps route (lines 502-508)**

```typescript
        // GitOps
        {
          path: 'gitops',
          name: 'GitOpsView',
          component: () => import('@/views/gitops/GitOpsView.vue'),
          meta: { title: 'GitOps', icon: 'Connection' },
        },
```

- [ ] **Step 4: Remove Tenancy route (lines 509-515)**

```typescript
        // Tenancy
        {
          path: 'tenancy',
          name: 'TenantList',
          component: () => import('@/views/tenancy/TenantList.vue'),
          meta: { title: '多租户', icon: 'UserFilled' },
        },
```

- [ ] **Step 5: Remove Approvals route (lines 516-522)**

```typescript
        // Approvals
        {
          path: 'approvals',
          name: 'ApprovalList',
          component: () => import('@/views/approval/ApprovalList.vue'),
          meta: { title: '审批', icon: 'CircleCheck' },
        },
```

- [ ] **Step 6: Remove Monitoring routes (lines 537-561)**

```typescript
        // Tools - Monitoring
        {
          path: 'monitoring',
          name: 'MonitoringView',
          component: () => import('@/views/monitoring/MonitoringView.vue'),
          meta: { title: '监控', icon: 'TrendCharts' },
        },
        {
          path: 'monitoring/dashboard',
          name: 'MonitoringDashboard',
          component: () => import('@/views/monitoring/MonitoringDashboard.vue'),
          meta: { title: '监控面板', parent: 'MonitoringView' },
        },
        {
          path: 'monitoring/resources',
          name: 'ResourceDashboard',
          component: () => import('@/views/monitoring/ResourceDashboard.vue'),
          meta: { title: '资源面板', parent: 'MonitoringView' },
        },
        {
          path: 'monitoring/prometheus',
          name: 'PrometheusView',
          component: () => import('@/views/monitoring/PrometheusView.vue'),
          meta: { title: 'Prometheus', parent: 'MonitoringView' },
        },
```

- [ ] **Step 7: Remove Topology routes (lines 562-574)**

```typescript
        // Tools - Topology
        {
          path: 'topology',
          name: 'TopologyView',
          component: () => import('@/views/topology/TopologyView.vue'),
          meta: { title: '拓扑', icon: 'Share' },
        },
        {
          path: 'topology/graph',
          name: 'TopologyGraph',
          component: () => import('@/views/topology/TopologyGraph.vue'),
          meta: { title: '拓扑图', parent: 'TopologyView' },
        },
```

- [ ] **Step 8: Remove Diff route (lines 575-581)**

```typescript
        // Tools - Diff
        {
          path: 'tools/diff',
          name: 'ResourceDiff',
          component: () => import('@/views/tools/ResourceDiff.vue'),
          meta: { title: '资源Diff', icon: 'Document' },
        },
```

- [ ] **Step 9: Remove YAML Editor route (lines 582-588)**

```typescript
        // Tools - YAML Editor
        {
          path: 'tools/yaml-editor',
          name: 'YAMLEditor',
          component: () => import('@/views/tools/YAMLEditor.vue'),
          meta: { title: 'YAML编辑器', icon: 'Document' },
        },
```

- [ ] **Step 10: Remove Batch Operations route (lines 589-595)**

```typescript
        // Tools - Batch
        {
          path: 'tools/batch',
          name: 'BatchOperations',
          component: () => import('@/views/tools/BatchOperations.vue'),
          meta: { title: '批量操作', icon: 'Document' },
        },
```

- [ ] **Step 11: Remove Notifications route (lines 624-630)**

```typescript
        // System - Notifications
        {
          path: 'notifications',
          name: 'NotificationCenter',
          component: () => import('@/views/notification/NotificationCenter.vue'),
          meta: { title: '通知中心', icon: 'Bell' },
        },
```

- [ ] **Step 12: Remove Watcher route (lines 631-637)**

```typescript
        // System - Watcher
        {
          path: 'watcher',
          name: 'ResourceWatcher',
          component: () => import('@/views/watcher/ResourceWatcher.vue'),
          meta: { title: '资源监控', icon: 'Monitor' },
        },
```

- [ ] **Step 13: Commit**

```bash
git add frontend/src/router/index.ts
git commit -m "chore: remove routes for catalog, gitops, tenancy, approval, monitoring, topology, tools, rbac, notification, watcher"
```

---

## Task 3: Update Frontend Sidebar

**Files:**
- Modify: `frontend/src/components/Layout/Sidebar.vue`

Remove these menu items from the template:

| Lines | Item |
|-------|------|
| 141-144 | RBAC (`/rbac`) |
| 149-152 | Catalog (`/catalog`) |
| 153-156 | GitOps (`/gitops`) |
| 157-160 | Tenancy (`/tenancy`) |
| 161-164 | Approvals (`/approvals`) |
| 178-181 | Monitoring (`/monitoring`) — inside tools sub-menu |
| 182-185 | Topology (`/topology`) — inside tools sub-menu |
| 186-189 | Diff (`/tools/diff`) — inside tools sub-menu |
| 190-193 | YAML Editor (`/tools/yaml-editor`) — inside tools sub-menu |

Also remove unused icon imports from the `<script>` section:
- `TrendCharts` (used only by monitoring)
- `ScaleToOriginal` (used only by limitranges — KEEP, actually used)
- `CircleCheck` (used only by approvals)

Wait — `ScaleToOriginal` is used by limitranges menu item (line 89), so keep it. Only remove `CircleCheck` and `TrendCharts` if they're not used elsewhere.

Let me check: `TrendCharts` is used only in the monitoring menu item. `CircleCheck` is used only in approvals. After removing those items, these icons are unused.

- [ ] **Step 1: Remove top-level menu items (RBAC, Catalog, GitOps, Tenancy, Approvals)**

Remove lines 141-164 (between events menu item and tools sub-menu):

```html
      <el-menu-item index="/rbac">
        <el-icon><UserFilled /></el-icon>
        <template #title>{{ t('sidebar.rbac') }}</template>
      </el-menu-item>
      <el-menu-item index="/crd">
        <el-icon><Grid /></el-icon>
        <template #title>{{ t('sidebar.crd') }}</template>
      </el-menu-item>
      <el-menu-item index="/catalog">
        <el-icon><Grid /></el-icon>
        <template #title>{{ t('sidebar.catalog') }}</template>
      </el-menu-item>
      <el-menu-item index="/gitops">
        <el-icon><Connection /></el-icon>
        <template #title>{{ t('sidebar.gitops') }}</template>
      </el-menu-item>
      <el-menu-item index="/tenancy">
        <el-icon><UserFilled /></el-icon>
        <template #title>{{ t('sidebar.tenancy') }}</template>
      </el-menu-item>
      <el-menu-item index="/approvals">
        <el-icon><CircleCheck /></el-icon>
        <template #title>{{ t('sidebar.approvals') }}</template>
      </el-menu-item>
```

After removal, the CRD menu item (lines 145-148) should remain. The sequence should go: events → crd → tools sub-menu.

- [ ] **Step 2: Remove items from inside tools sub-menu (Monitoring, Topology, Diff, YAML Editor)**

Remove lines 178-193 from inside the tools sub-menu. After removal, the tools sub-menu should only contain Terminal and Logs:

```html
        <el-menu-item index="/monitoring">
          <el-icon><TrendCharts /></el-icon>
          <template #title>{{ t('sidebar.monitoring') }}</template>
        </el-menu-item>
        <el-menu-item index="/topology">
          <el-icon><Share /></el-icon>
          <template #title>{{ t('sidebar.topology') }}</template>
        </el-menu-item>
        <el-menu-item index="/tools/diff">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.diff') }}</template>
        </el-menu-item>
        <el-menu-item index="/tools/yaml-editor">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.yamlEditor') }}</template>
        </el-menu-item>
```

- [ ] **Step 3: Remove unused icon imports**

Remove `CircleCheck` and `TrendCharts` from the import block (lines 225-256). Keep all other icons.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/Layout/Sidebar.vue
git commit -m "chore: remove sidebar menu items for catalog, gitops, tenancy, approvals, monitoring, topology, diff, yaml-editor, rbac"
```

---

## Task 4: Update Frontend i18n Files

**Files:**
- Modify: `frontend/src/locales/zh-CN.ts`
- Modify: `frontend/src/locales/en.ts`

Remove these top-level translation sections from both files:
- `rbac` (lines 604-632 in zh-CN, 604-632 in en)
- `monitoring` (lines 633-672 in zh-CN, 633-672 in en)
- `topology` (lines 673-687 in zh-CN, 673-687 in en)
- `catalog` (lines 688-719 in zh-CN, 688-719 in en)
- `gitops` (lines 720-754 in zh-CN, 720-754 in en)
- `tenancy` (lines 755-787 in zh-CN, 755-787 in en)
- `approval` (lines 788-834 in zh-CN, 788-834 in en)
- `notification` (lines 862-887 in zh-CN, 862-887 in en)
- `watcher` (lines 963-972 in zh-CN, 963-972 in en)

Remove these keys from the `sidebar` section in both files:
- `rbac: 'RBAC'`
- `catalog: '应用目录'` / `'Apply目录'`
- `gitops: 'GitOps'`
- `tenancy: '多租户'` / `'Tenancy'`
- `approvals: '审批流程'` / `'Approvals'`
- `monitoring: '资源监控'` / `'Resource监控'`
- `topology: '资源拓扑'` / `'Resource拓扑'`
- `diff: '资源对比'` / `'Resource对比'`
- `yamlEditor: 'YAML 编辑器'` / `'YAML Edit器'`

Also remove the entire `tools` section (lines 905-962 in zh-CN, 905-962 in en) since it contains only YAML Editor, Resource Diff, and Batch Operations translations.

- [ ] **Step 1: Edit zh-CN.ts — remove feature sections**

Remove the following sections entirely from `frontend/src/locales/zh-CN.ts`:
- Lines 604-632: `rbac: { ... }`
- Lines 633-672: `monitoring: { ... }`
- Lines 673-687: `topology: { ... }`
- Lines 688-719: `catalog: { ... }`
- Lines 720-754: `gitops: { ... }`
- Lines 755-787: `tenancy: { ... }`
- Lines 788-834: `approval: { ... }`
- Lines 862-887: `notification: { ... }`
- Lines 905-962: `tools: { ... }`
- Lines 963-972: `watcher: { ... }`

- [ ] **Step 2: Edit zh-CN.ts — remove sidebar keys**

Remove from the `sidebar` section:
```
    rbac: 'RBAC',
    catalog: '应用目录',
    gitops: 'GitOps',
    tenancy: '多租户',
    approvals: '审批流程',
    monitoring: '资源监控',
    topology: '资源拓扑',
    diff: '资源对比',
    yamlEditor: 'YAML 编辑器',
```

- [ ] **Step 3: Edit en.ts — remove feature sections**

Same removals as zh-CN.ts:
- Lines 604-632: `rbac: { ... }`
- Lines 633-672: `monitoring: { ... }`
- Lines 673-687: `topology: { ... }`
- Lines 688-719: `catalog: { ... }`
- Lines 720-754: `gitops: { ... }`
- Lines 755-787: `tenancy: { ... }`
- Lines 788-834: `approval: { ... }`
- Lines 862-887: `notification: { ... }`
- Lines 905-962: `tools: { ... }`
- Lines 963-972: `watcher: { ... }`

- [ ] **Step 4: Edit en.ts — remove sidebar keys**

Remove from the `sidebar` section:
```
    rbac: 'RBAC',
    catalog: 'Apply目录',
    gitops: 'GitOps',
    tenancy: 'Tenancy',
    approvals: 'Approvals',
    monitoring: 'Resource监控',
    topology: 'Resource拓扑',
    diff: 'Resource对比',
    yamlEditor: 'YAML Edit器',
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/locales/zh-CN.ts frontend/src/locales/en.ts
git commit -m "chore: remove i18n translations for removed features"
```

---

## Task 5: Delete Backend API Handlers

**Files:**
- Delete: `backend/app/k8s/api/catalog.go`
- Delete: `backend/app/k8s/api/gitops.go`
- Delete: `backend/app/k8s/api/approval.go`
- Delete: `backend/app/k8s/api/tenancy.go`
- Delete: `backend/app/k8s/api/metrics.go`
- Delete: `backend/app/k8s/api/prometheus.go`
- Delete: `backend/app/k8s/api/topology.go`
- Delete: `backend/app/k8s/api/rbac.go`
- Delete: `backend/app/k8s/model/approval.go`

- [ ] **Step 1: Delete all backend API handler files**

```bash
cd /Users/zqqzqq/05_github/gkube
rm -f backend/app/k8s/api/catalog.go
rm -f backend/app/k8s/api/gitops.go
rm -f backend/app/k8s/api/approval.go
rm -f backend/app/k8s/api/tenancy.go
rm -f backend/app/k8s/api/metrics.go
rm -f backend/app/k8s/api/prometheus.go
rm -f backend/app/k8s/api/topology.go
rm -f backend/app/k8s/api/rbac.go
```

- [ ] **Step 2: Delete approval model**

```bash
rm -f backend/app/k8s/model/approval.go
```

- [ ] **Step 3: Commit**

```bash
git add backend/app/k8s/api/catalog.go backend/app/k8s/api/gitops.go backend/app/k8s/api/approval.go backend/app/k8s/api/tenancy.go backend/app/k8s/api/metrics.go backend/app/k8s/api/prometheus.go backend/app/k8s/api/topology.go backend/app/k8s/api/rbac.go backend/app/k8s/model/approval.go
git commit -m "chore: remove backend API handlers for catalog, gitops, approval, tenancy, metrics, prometheus, topology, rbac"
```

---

## Task 6: Delete Backend K8s RBAC Packages

**Files:**
- Delete: `backend/pkg/k8s/serviceaccount/` (entire directory)
- Delete: `backend/pkg/k8s/clusterrole/` (entire directory)
- Delete: `backend/pkg/k8s/rolebinding/` (entire directory)

These packages are only used by `backend/app/k8s/api/rbac.go` (already deleted in Task 5).

- [ ] **Step 1: Delete the three RBAC packages**

```bash
rm -rf backend/pkg/k8s/serviceaccount
rm -rf backend/pkg/k8s/clusterrole
rm -rf backend/pkg/k8s/rolebinding
```

- [ ] **Step 2: Commit**

```bash
git add backend/pkg/k8s/serviceaccount backend/pkg/k8s/clusterrole backend/pkg/k8s/rolebinding
git commit -m "chore: remove K8s RBAC packages (serviceaccount, clusterrole, rolebinding)"
```

---

## Task 7: Update Backend Router

**Files:**
- Modify: `backend/router/router.go`

Remove these route groups from the `k8s` group (lines 246-361):

| Lines | Routes | Handler |
|-------|--------|---------|
| 246-270 | K8s RBAC (serviceaccount, clusterrole, role, clusterrolebinding, rolebinding — 20 routes) | `k8sApi.Rbac` |
| 312-314 | Metrics (2 routes) | `k8sApi.Metrics` |
| 316-322 | Prometheus (6 routes) | `k8sApi.Prometheus` |
| 324-327 | Topology (3 routes) | `k8sApi.Topology` |
| 329-334 | Catalog (5 routes) | `k8sApi.Catalog` |
| 336-343 | GitOps (7 routes) | `k8sApi.GitOps` |
| 345-351 | Tenancy (6 routes) | `k8sApi.Tenancy` |
| 353-360 | Approval (7 routes) | `k8sApi.Approval` |

- [ ] **Step 1: Remove K8s RBAC routes (lines 246-270)**

Remove from `// RBAC` comment through the `rolebinding/delete` route:

```go
			// RBAC
			k8s.GET("serviceaccount/list", k8sApi.Rbac.GetServiceAccountList)
			k8s.GET("serviceaccount/yaml", k8sApi.Rbac.GetServiceAccountYaml)
			k8s.GET("serviceaccount/get-yaml", k8sApi.Rbac.GetServiceAccountYaml) // alias
			k8s.DELETE("serviceaccount/delete", k8sApi.Rbac.DeleteServiceAccount)

			k8s.GET("clusterrole/list", k8sApi.Rbac.GetClusterRoleList)
			k8s.GET("clusterrole/yaml", k8sApi.Rbac.GetClusterRoleYaml)
			k8s.GET("clusterrole/get-yaml", k8sApi.Rbac.GetClusterRoleYaml) // alias
			k8s.DELETE("clusterrole/delete", k8sApi.Rbac.DeleteClusterRole)

			k8s.GET("role/list", k8sApi.Rbac.GetRoleList)
			k8s.GET("role/yaml", k8sApi.Rbac.GetRoleYaml)
			k8s.GET("role/get-yaml", k8sApi.Rbac.GetRoleYaml) // alias
			k8s.DELETE("role/delete", k8sApi.Rbac.DeleteRole)

			k8s.GET("clusterrolebinding/list", k8sApi.Rbac.GetClusterRoleBindingList)
			k8s.GET("clusterrolebinding/yaml", k8sApi.Rbac.GetClusterRoleBindingYaml)
			k8s.GET("clusterrolebinding/get-yaml", k8sApi.Rbac.GetClusterRoleBindingYaml) // alias
			k8s.DELETE("clusterrolebinding/delete", k8sApi.Rbac.DeleteClusterRoleBinding)

			k8s.GET("rolebinding/list", k8sApi.Rbac.GetRoleBindingList)
			k8s.GET("rolebinding/yaml", k8sApi.Rbac.GetRoleBindingYaml)
			k8s.GET("rolebinding/get-yaml", k8sApi.Rbac.GetRoleBindingYaml) // alias
			k8s.DELETE("rolebinding/delete", k8sApi.Rbac.DeleteRoleBinding)
```

- [ ] **Step 2: Remove Metrics routes (lines 312-314)**

```go
			// Metrics
			k8s.GET("metrics/nodes", k8sApi.Metrics.GetNodeMetrics)
			k8s.GET("metrics/pods", k8sApi.Metrics.GetPodMetrics)
```

- [ ] **Step 3: Remove Prometheus routes (lines 316-322)**

```go
			// Prometheus代理
			k8s.GET("prometheus/query", k8sApi.Prometheus.QueryPrometheus)
			k8s.GET("prometheus/query_range", k8sApi.Prometheus.QueryPrometheusRange)
			k8s.GET("prometheus/targets", k8sApi.Prometheus.GetPrometheusTargets)
			k8s.GET("prometheus/alerts", k8sApi.Prometheus.GetPrometheusAlerts)
			k8s.GET("prometheus/config", k8sApi.Prometheus.GetPrometheusConfig)
			k8s.PUT("prometheus/config", k8sApi.Prometheus.UpdatePrometheusConfig)
```

- [ ] **Step 4: Remove Topology routes (lines 324-327)**

```go
			// Topology
			k8s.GET("topology/deployment", k8sApi.Topology.GetDeploymentTopology)
			k8s.GET("topology/statefulset", k8sApi.Topology.GetStatefulSetTopology)
			k8s.GET("topology/daemonset", k8sApi.Topology.GetDaemonSetTopology)
```

- [ ] **Step 5: Remove Catalog routes (lines 329-334)**

```go
			// Catalog
			k8s.GET("catalog/charts", k8sApi.Catalog.ListCharts)
			k8s.GET("catalog/chart", k8sApi.Catalog.GetChartDetails)
			k8s.POST("catalog/install", k8sApi.Catalog.InstallChart)
			k8s.GET("catalog/releases", k8sApi.Catalog.ListReleases)
			k8s.DELETE("catalog/release", k8sApi.Catalog.UninstallRelease)
```

- [ ] **Step 6: Remove GitOps routes (lines 336-343)**

```go
			// GitOps
			k8s.GET("gitops/applications", k8sApi.GitOps.ListApplications)
			k8s.GET("gitops/application", k8sApi.GitOps.GetApplication)
			k8s.POST("gitops/sync", k8sApi.GitOps.SyncApplication)
			k8s.GET("gitops/history", k8sApi.GitOps.GetApplicationHistory)
			k8s.POST("gitops/rollback", k8sApi.GitOps.RollbackApplication)
			k8s.POST("gitops/create", k8sApi.GitOps.CreateApplication)
			k8s.DELETE("gitops/delete", k8sApi.GitOps.DeleteApplication)
```

- [ ] **Step 7: Remove Tenancy routes (lines 345-351)**

```go
			// Tenancy
			k8s.GET("tenancy/tenants", k8sApi.Tenancy.ListTenants)
			k8s.GET("tenancy/tenant", k8sApi.Tenancy.GetTenant)
			k8s.POST("tenancy/create", k8sApi.Tenancy.CreateTenant)
			k8s.DELETE("tenancy/delete", k8sApi.Tenancy.DeleteTenant)
			k8s.POST("tenancy/namespace/add", k8sApi.Tenancy.AddNamespaceToTenant)
			k8s.POST("tenancy/namespace/remove", k8sApi.Tenancy.RemoveNamespaceFromTenant)
```

- [ ] **Step 8: Remove Approval routes (lines 353-360)**

```go
			// Approval Workflows
			k8s.GET("approval/list", k8sApi.Approval.ListApprovals)
			k8s.GET("approval/detail", k8sApi.Approval.GetApproval)
			k8s.POST("approval/create", k8sApi.Approval.CreateApproval)
			k8s.POST("approval/approve", k8sApi.Approval.ApproveRequest)
			k8s.POST("approval/reject", k8sApi.Approval.RejectRequest)
			k8s.DELETE("approval/delete", k8sApi.Approval.DeleteApproval)
			k8s.GET("approval/stats", k8sApi.Approval.GetApprovalStats)
```

- [ ] **Step 9: Commit**

```bash
git add backend/router/router.go
git commit -m "chore: remove API routes for catalog, gitops, approval, tenancy, metrics, prometheus, topology, rbac"
```

---

## Task 8: Remove Redis Dependency

**Files:**
- Delete: `backend/pkg/redis/` (entire directory)
- Modify: `backend/cmd/root.go`
- Modify: `backend/config/config.go`
- Modify: `backend/config/config.yaml`
- Modify: `backend/docker-compose.yaml`
- Modify: `backend/go.mod` (via `go mod tidy`)

- [ ] **Step 1: Delete Redis package**

```bash
rm -rf backend/pkg/redis
```

- [ ] **Step 2: Update cmd/root.go — remove redis import and Init()**

In `backend/cmd/root.go`, remove line 13:
```go
	"gkube/pkg/redis"
```

Remove line 64:
```go
	redis.Init()    // 初始化redis
```

The final `init()` function should be:
```go
func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	config.Init()   // 初始化配置文件
	logger.Init()   // 初始化日志
	database.Init() // 初始化数据库
	es.Init()       // 初始化es
}
```

- [ ] **Step 3: Update config/config.go — remove Redis struct**

In `backend/config/config.go`, remove lines 24-28:
```go
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
```

- [ ] **Step 4: Update config/config.yaml — remove redis section**

In `backend/config/config.yaml`, remove lines 10-13:
```yaml
redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
```

- [ ] **Step 5: Update docker-compose.yaml — remove redis service**

In `backend/docker-compose.yaml`, remove lines 15-21:
```yaml
  # reids
  redis:
    image: redis:latest
    restart: always
    ports:
      - "6379:6379"
```

- [ ] **Step 6: Run go mod tidy**

```bash
cd backend && go mod tidy
```

- [ ] **Step 7: Commit**

```bash
git add backend/pkg/redis backend/cmd/root.go backend/config/config.go backend/config/config.yaml backend/docker-compose.yaml backend/go.mod backend/go.sum
git commit -m "chore: remove Redis dependency"
```

---

## Task 9: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

Update the project documentation to reflect the simplified feature set.

- [ ] **Step 1: Update Infrastructure section**

In the `docker-compose` command section, remove the note about Redis. The command stays the same but Redis is no longer a service.

- [ ] **Step 2: Update Key Technology Stack table**

Remove from the table:
- `Cache` row (go-redis)

- [ ] **Step 3: Update Backend Architecture section**

In the startup chain, remove `redis.Init()`. The chain becomes:
`config.Init() → logger.Init() → database.Init() → es.Init() → HTTP server on :8080 → cluster health checker goroutine`

- [ ] **Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md to reflect simplified feature set"
```

---

## Task 10: Verify Build

- [ ] **Step 1: Build frontend**

```bash
cd frontend && npm run build
```

Expected: Clean build with no errors.

- [ ] **Step 2: Build backend**

```bash
cd backend && go build -o gkube .
```

Expected: Clean build with no errors.

- [ ] **Step 3: Final commit if any fixes needed**

If any compilation errors were found and fixed during build verification:

```bash
git add -A
git commit -m "fix: resolve build errors from feature removal"
```
