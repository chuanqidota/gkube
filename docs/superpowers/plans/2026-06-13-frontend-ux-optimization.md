# Frontend UX Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Optimize frontend interactions to match Kuboard-level UX: shared layout, breadcrumbs, auto-refresh, and translation fixes.

**Architecture:** Vue 3 nested routes wrap all business pages in AppLayout (Sidebar + Header). A `useAutoRefresh` composable provides polling. Breadcrumbs build from `route.matched`. English translations are regenerated.

**Tech Stack:** Vue 3, Vue Router 4, Element Plus, vue-i18n, TypeScript

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `frontend/src/router/index.ts` | Modify | Restructure flat routes → nested under AppLayout parent |
| `frontend/src/components/Layout/Header.vue` | Modify | Breadcrumb from `route.matched` instead of just `route.meta.title` |
| `frontend/src/composables/useAutoRefresh.ts` | Create | Reusable polling composable with countdown + pause |
| `frontend/src/views/dashboard/DashboardView.vue` | Modify | Integrate `useAutoRefresh` |
| `frontend/src/views/workload/DeploymentList.vue` | Modify | Integrate `useAutoRefresh` |
| `frontend/src/views/workload/PodList.vue` | Modify | Integrate `useAutoRefresh` |
| `frontend/src/locales/en.ts` | Modify | Fix all Chinese-in-English translation errors |

---

### Task 1: Restructure Router to Nested Routes

**Files:**
- Modify: `frontend/src/router/index.ts`

**Context:** The router currently has ~80 flat top-level routes. `AppLayout` exists at `frontend/src/components/Layout/AppLayout.vue` with Sidebar + Header + `<router-view />` but is not used. We need to make all business pages children of a `/` parent route that renders `AppLayout`. Login and OIDC callback remain top-level (no sidebar/header).

- [ ] **Step 1: Read current router to capture all route paths**

Read `frontend/src/router/index.ts` and list every route path. These paths must NOT change — only the nesting structure changes.

- [ ] **Step 2: Rewrite router with nested structure**

Replace the entire `routes` array. The new structure:

```ts
import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Standalone pages (no AppLayout)
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/oidc/callback',
      name: 'OIDCCallback',
      component: () => import('@/views/login/OIDCCallback.vue'),
      meta: { public: true },
    },
    // AppLayout wrapper for all authenticated pages
    {
      path: '/',
      component: () => import('@/components/Layout/AppLayout.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
          meta: { title: '仪表盘', icon: 'Odometer' },
        },
        {
          path: 'search',
          name: 'ResourceSearch',
          component: () => import('@/views/search/ResourceSearch.vue'),
          meta: { title: '资源搜索', icon: 'Search' },
        },
        {
          path: 'system/overview',
          name: 'SystemOverview',
          component: () => import('@/views/dashboard/SystemOverview.vue'),
          meta: { title: '系统概览', icon: 'Monitor' },
        },
        // Clusters
        {
          path: 'clusters',
          name: 'ClusterList',
          component: () => import('@/views/ClusterList.vue'),
          meta: { title: '集群管理', icon: 'Connection' },
        },
        {
          path: 'clusters/create',
          name: 'ClusterCreate',
          component: () => import('@/views/ClusterCreate.vue'),
          meta: { title: '创建集群', parent: 'ClusterList' },
        },
        {
          path: 'clusters/:id',
          name: 'ClusterDetail',
          component: () => import('@/views/ClusterDetail.vue'),
          props: true,
          meta: { title: '集群详情', parent: 'ClusterList' },
        },
        // Workloads
        {
          path: 'workloads/pods',
          name: 'PodList',
          component: () => import('@/views/workload/PodList.vue'),
          meta: { title: 'Pod', icon: 'Coin' },
        },
        {
          path: 'workloads/pods/:namespace/:name',
          name: 'PodDetail',
          component: () => import('@/views/workload/PodDetail.vue'),
          props: true,
          meta: { title: 'Pod详情', parent: 'PodList' },
        },
        {
          path: 'workloads/deployments',
          name: 'DeploymentList',
          component: () => import('@/views/workload/DeploymentList.vue'),
          meta: { title: 'Deployment', icon: 'Files' },
        },
        {
          path: 'workloads/deployments/create',
          name: 'DeploymentCreate',
          component: () => import('@/views/workload/DeploymentCreate.vue'),
          meta: { title: '创建Deployment', parent: 'DeploymentList' },
        },
        {
          path: 'workloads/deployments/:namespace/:name',
          name: 'DeploymentDetail',
          component: () => import('@/views/workload/DeploymentDetail.vue'),
          props: true,
          meta: { title: 'Deployment详情', parent: 'DeploymentList' },
        },
        {
          path: 'workloads/statefulsets',
          name: 'StatefulSetList',
          component: () => import('@/views/workload/StatefulSetList.vue'),
          meta: { title: 'StatefulSet', icon: 'Files' },
        },
        {
          path: 'workloads/statefulsets/create',
          name: 'StatefulSetCreate',
          component: () => import('@/views/workload/StatefulSetCreate.vue'),
          meta: { title: '创建StatefulSet', parent: 'StatefulSetList' },
        },
        {
          path: 'workloads/statefulsets/:namespace/:name',
          name: 'StatefulSetDetail',
          component: () => import('@/views/workload/StatefulSetDetail.vue'),
          props: true,
          meta: { title: 'StatefulSet详情', parent: 'StatefulSetList' },
        },
        {
          path: 'workloads/daemonsets',
          name: 'DaemonSetList',
          component: () => import('@/views/workload/DaemonSetList.vue'),
          meta: { title: 'DaemonSet', icon: 'Files' },
        },
        {
          path: 'workloads/daemonsets/create',
          name: 'DaemonSetCreate',
          component: () => import('@/views/workload/DaemonSetCreate.vue'),
          meta: { title: '创建DaemonSet', parent: 'DaemonSetList' },
        },
        {
          path: 'workloads/daemonsets/:namespace/:name',
          name: 'DaemonSetDetail',
          component: () => import('@/views/workload/DaemonSetDetail.vue'),
          props: true,
          meta: { title: 'DaemonSet详情', parent: 'DaemonSetList' },
        },
        {
          path: 'workloads/jobs',
          name: 'JobList',
          component: () => import('@/views/workload/JobList.vue'),
          meta: { title: 'Job', icon: 'Files' },
        },
        {
          path: 'workloads/jobs/create',
          name: 'JobCreate',
          component: () => import('@/views/workload/JobCreate.vue'),
          meta: { title: '创建Job', parent: 'JobList' },
        },
        {
          path: 'workloads/jobs/:namespace/:name',
          name: 'JobDetail',
          component: () => import('@/views/workload/JobDetail.vue'),
          props: true,
          meta: { title: 'Job详情', parent: 'JobList' },
        },
        {
          path: 'workloads/cronjobs',
          name: 'CronJobList',
          component: () => import('@/views/workload/CronJobList.vue'),
          meta: { title: 'CronJob', icon: 'Files' },
        },
        {
          path: 'workloads/cronjobs/create',
          name: 'CronJobCreate',
          component: () => import('@/views/workload/CronJobCreate.vue'),
          meta: { title: '创建CronJob', parent: 'CronJobList' },
        },
        {
          path: 'workloads/cronjobs/:namespace/:name',
          name: 'CronJobDetail',
          component: () => import('@/views/workload/CronJobDetail.vue'),
          props: true,
          meta: { title: 'CronJob详情', parent: 'CronJobList' },
        },
        {
          path: 'workloads/hpa',
          name: 'HPAList',
          component: () => import('@/views/workload/hpa/HPAList.vue'),
          meta: { title: 'HPA', icon: 'DataLine' },
        },
        {
          path: 'workloads/hpa/create',
          name: 'HPACreate',
          component: () => import('@/views/workload/hpa/HPACreate.vue'),
          meta: { title: '创建HPA', parent: 'HPAList' },
        },
        {
          path: 'workloads/hpa/:namespace/:name',
          name: 'HPADetail',
          component: () => import('@/views/workload/hpa/HPADetail.vue'),
          props: true,
          meta: { title: 'HPA详情', parent: 'HPAList' },
        },
        {
          path: 'workloads/pdb',
          name: 'PDBList',
          component: () => import('@/views/workload/pdb/PDBList.vue'),
          meta: { title: 'PDB', icon: 'Warning' },
        },
        {
          path: 'workloads/pdb/create',
          name: 'PDBCreate',
          component: () => import('@/views/workload/pdb/PDBCreate.vue'),
          meta: { title: '创建PDB', parent: 'PDBList' },
        },
        {
          path: 'workloads/pdb/:namespace/:name',
          name: 'PDBDetail',
          component: () => import('@/views/workload/pdb/PDBDetail.vue'),
          meta: { title: 'PDB详情', parent: 'PDBList' },
        },
        // Config
        {
          path: 'config/configmaps',
          name: 'ConfigMapList',
          component: () => import('@/views/config/ConfigMapList.vue'),
          meta: { title: 'ConfigMap', icon: 'Tickets' },
        },
        {
          path: 'config/configmaps/create',
          name: 'ConfigMapCreate',
          component: () => import('@/views/config/ConfigMapCreate.vue'),
          meta: { title: '创建ConfigMap', parent: 'ConfigMapList' },
        },
        {
          path: 'config/configmaps/:namespace/:name',
          name: 'ConfigMapDetail',
          component: () => import('@/views/config/ConfigMapDetail.vue'),
          props: true,
          meta: { title: 'ConfigMap详情', parent: 'ConfigMapList' },
        },
        {
          path: 'config/secrets',
          name: 'SecretList',
          component: () => import('@/views/config/SecretList.vue'),
          meta: { title: 'Secret', icon: 'Key' },
        },
        {
          path: 'config/secrets/create',
          name: 'SecretCreate',
          component: () => import('@/views/config/SecretCreate.vue'),
          meta: { title: '创建Secret', parent: 'SecretList' },
        },
        {
          path: 'config/secrets/:namespace/:name',
          name: 'SecretDetail',
          component: () => import('@/views/config/SecretDetail.vue'),
          props: true,
          meta: { title: 'Secret详情', parent: 'SecretList' },
        },
        {
          path: 'config/resourcequotas',
          name: 'ResourceQuotaList',
          component: () => import('@/views/config/resourcequota/ResourceQuotaList.vue'),
          meta: { title: 'ResourceQuota', icon: 'Coin' },
        },
        {
          path: 'config/resourcequotas/create',
          name: 'ResourceQuotaCreate',
          component: () => import('@/views/config/resourcequota/ResourceQuotaCreate.vue'),
          meta: { title: '创建ResourceQuota', parent: 'ResourceQuotaList' },
        },
        {
          path: 'config/resourcequotas/:namespace/:name',
          name: 'ResourceQuotaDetail',
          component: () => import('@/views/config/ResourceQuotaDetail.vue'),
          props: true,
          meta: { title: 'ResourceQuota详情', parent: 'ResourceQuotaList' },
        },
        {
          path: 'config/limitranges',
          name: 'LimitRangeList',
          component: () => import('@/views/config/limitrange/LimitRangeList.vue'),
          meta: { title: 'LimitRange', icon: 'ScaleToOriginal' },
        },
        {
          path: 'config/limitranges/create',
          name: 'LimitRangeCreate',
          component: () => import('@/views/config/limitrange/LimitRangeCreate.vue'),
          meta: { title: '创建LimitRange', parent: 'LimitRangeList' },
        },
        {
          path: 'config/limitranges/:namespace/:name',
          name: 'LimitRangeDetail',
          component: () => import('@/views/config/limitrange/LimitRangeDetail.vue'),
          meta: { title: 'LimitRange详情', parent: 'LimitRangeList' },
        },
        // Storage
        {
          path: 'storage/pvs',
          name: 'PVList',
          component: () => import('@/views/storage/PVList.vue'),
          meta: { title: 'PersistentVolume', icon: 'Coin' },
        },
        {
          path: 'storage/pvs/create',
          name: 'PVCreate',
          component: () => import('@/views/storage/PVCreate.vue'),
          meta: { title: '创建PV', parent: 'PVList' },
        },
        {
          path: 'storage/pvs/:name',
          name: 'PVDetail',
          component: () => import('@/views/storage/PVDetail.vue'),
          props: true,
          meta: { title: 'PV详情', parent: 'PVList' },
        },
        {
          path: 'storage/pvcs',
          name: 'PVCList',
          component: () => import('@/views/storage/PVCList.vue'),
          meta: { title: 'PVC', icon: 'Box' },
        },
        {
          path: 'storage/pvcs/create',
          name: 'PVCCreate',
          component: () => import('@/views/storage/PVCCreate.vue'),
          meta: { title: '创建PVC', parent: 'PVCList' },
        },
        {
          path: 'storage/pvcs/:namespace/:name',
          name: 'PVCDetail',
          component: () => import('@/views/storage/PVCDetail.vue'),
          props: true,
          meta: { title: 'PVC详情', parent: 'PVCList' },
        },
        {
          path: 'storage/storageclasses',
          name: 'StorageClassList',
          component: () => import('@/views/storage/StorageClassList.vue'),
          meta: { title: 'StorageClass', icon: 'Files' },
        },
        {
          path: 'storage/storageclasses/create',
          name: 'StorageClassCreate',
          component: () => import('@/views/storage/StorageClassCreate.vue'),
          meta: { title: '创建StorageClass', parent: 'StorageClassList' },
        },
        {
          path: 'storage/storageclasses/:name',
          name: 'StorageClassDetail',
          component: () => import('@/views/storage/StorageClassDetail.vue'),
          props: true,
          meta: { title: 'StorageClass详情', parent: 'StorageClassList' },
        },
        // Network
        {
          path: 'services',
          name: 'ServiceList',
          component: () => import('@/views/network/ServiceList.vue'),
          meta: { title: 'Service', icon: 'Connection' },
        },
        {
          path: 'services/create',
          name: 'ServiceCreate',
          component: () => import('@/views/network/ServiceCreate.vue'),
          meta: { title: '创建Service', parent: 'ServiceList' },
        },
        {
          path: 'services/:namespace/:name',
          name: 'ServiceDetail',
          component: () => import('@/views/network/ServiceDetail.vue'),
          props: true,
          meta: { title: 'Service详情', parent: 'ServiceList' },
        },
        {
          path: 'ingresses',
          name: 'IngressList',
          component: () => import('@/views/network/IngressList.vue'),
          meta: { title: 'Ingress', icon: 'Link' },
        },
        {
          path: 'ingresses/create',
          name: 'IngressCreate',
          component: () => import('@/views/network/IngressCreate.vue'),
          meta: { title: '创建Ingress', parent: 'IngressList' },
        },
        {
          path: 'ingresses/:namespace/:name',
          name: 'IngressDetail',
          component: () => import('@/views/network/IngressDetail.vue'),
          props: true,
          meta: { title: 'Ingress详情', parent: 'IngressList' },
        },
        {
          path: 'network/networkpolicies',
          name: 'NetworkPolicyList',
          component: () => import('@/views/network/networkpolicy/NetworkPolicyList.vue'),
          meta: { title: 'NetworkPolicy', icon: 'Lock' },
        },
        {
          path: 'network/networkpolicies/create',
          name: 'NetworkPolicyCreate',
          component: () => import('@/views/network/networkpolicy/NetworkPolicyCreate.vue'),
          meta: { title: '创建NetworkPolicy', parent: 'NetworkPolicyList' },
        },
        {
          path: 'network/networkpolicies/:namespace/:name',
          name: 'NetworkPolicyDetail',
          component: () => import('@/views/network/networkpolicy/NetworkPolicyDetail.vue'),
          props: true,
          meta: { title: 'NetworkPolicy详情', parent: 'NetworkPolicyList' },
        },
        // Nodes
        {
          path: 'nodes',
          name: 'NodeList',
          component: () => import('@/views/node/NodeList.vue'),
          meta: { title: '节点', icon: 'Cpu' },
        },
        {
          path: 'nodes/:name',
          name: 'NodeDetail',
          component: () => import('@/views/node/NodeDetail.vue'),
          props: true,
          meta: { title: '节点详情', parent: 'NodeList' },
        },
        // Namespaces
        {
          path: 'namespaces',
          name: 'NamespaceList',
          component: () => import('@/views/namespace/NamespaceList.vue'),
          meta: { title: '命名空间', icon: 'FolderOpened' },
        },
        {
          path: 'namespaces/:name',
          name: 'NamespaceDetail',
          component: () => import('@/views/namespace/NamespaceDetail.vue'),
          props: true,
          meta: { title: '命名空间详情', parent: 'NamespaceList' },
        },
        {
          path: 'namespaces/manager',
          name: 'NamespaceManager',
          component: () => import('@/views/namespace/NamespaceManager.vue'),
          meta: { title: '命名空间管理', parent: 'NamespaceList' },
        },
        // Events
        {
          path: 'events',
          name: 'EventList',
          component: () => import('@/views/event/EventList.vue'),
          meta: { title: '事件', icon: 'Bell' },
        },
        {
          path: 'events/viewer',
          name: 'EventViewer',
          component: () => import('@/views/event/EventViewer.vue'),
          meta: { title: '事件查看器', parent: 'EventList' },
        },
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
        // CRD
        {
          path: 'crd',
          name: 'CRDList',
          component: () => import('@/views/crd/CRDList.vue'),
          meta: { title: 'CRD', icon: 'Grid' },
        },
        {
          path: 'crd/create',
          name: 'CRDCreate',
          component: () => import('@/views/crd/CRDCreate.vue'),
          meta: { title: '创建CRD', parent: 'CRDList' },
        },
        {
          path: 'crd/resources',
          name: 'CustomResourceList',
          component: () => import('@/views/crd/CustomResourceList.vue'),
          meta: { title: '自定义资源', parent: 'CRDList' },
        },
        // Catalog
        {
          path: 'catalog',
          name: 'AppCatalog',
          component: () => import('@/views/catalog/AppCatalog.vue'),
          meta: { title: '应用商店', icon: 'Grid' },
        },
        // GitOps
        {
          path: 'gitops',
          name: 'GitOpsView',
          component: () => import('@/views/gitops/GitOpsView.vue'),
          meta: { title: 'GitOps', icon: 'Connection' },
        },
        // Tenancy
        {
          path: 'tenancy',
          name: 'TenantList',
          component: () => import('@/views/tenancy/TenantList.vue'),
          meta: { title: '多租户', icon: 'UserFilled' },
        },
        // Approvals
        {
          path: 'approvals',
          name: 'ApprovalList',
          component: () => import('@/views/approval/ApprovalList.vue'),
          meta: { title: '审批', icon: 'CircleCheck' },
        },
        // Tools
        {
          path: 'terminal',
          name: 'Terminal',
          component: () => import('@/views/terminal/TerminalView.vue'),
          meta: { title: 'Web终端', icon: 'Promotion' },
        },
        {
          path: 'logs',
          name: 'Logs',
          component: () => import('@/views/logviewer/LogView.vue'),
          meta: { title: '日志查看', icon: 'Document' },
        },
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
        {
          path: 'tools/diff',
          name: 'ResourceDiff',
          component: () => import('@/views/tools/ResourceDiff.vue'),
          meta: { title: '资源Diff', icon: 'Document' },
        },
        {
          path: 'tools/yaml-editor',
          name: 'YAMLEditor',
          component: () => import('@/views/tools/YAMLEditor.vue'),
          meta: { title: 'YAML编辑器', icon: 'Document' },
        },
        {
          path: 'tools/batch',
          name: 'BatchOperations',
          component: () => import('@/views/tools/BatchOperations.vue'),
          meta: { title: '批量操作', icon: 'Document' },
        },
        // System
        {
          path: 'users',
          name: 'UserList',
          component: () => import('@/views/UserList.vue'),
          meta: { title: '用户管理', icon: 'User' },
        },
        {
          path: 'roles',
          name: 'RoleList',
          component: () => import('@/views/RoleList.vue'),
          meta: { title: '角色管理', icon: 'UserFilled' },
        },
        {
          path: 'settings/auth',
          name: 'OIDCSettings',
          component: () => import('@/views/settings/OIDCSettings.vue'),
          meta: { title: '认证设置', icon: 'Setting' },
        },
        {
          path: 'audit',
          name: 'AuditLog',
          component: () => import('@/views/audit/AuditLog.vue'),
          meta: { title: '审计日志', icon: 'Document' },
        },
        {
          path: 'notifications',
          name: 'NotificationCenter',
          component: () => import('@/views/notification/NotificationCenter.vue'),
          meta: { title: '通知中心', icon: 'Bell' },
        },
        {
          path: 'watcher',
          name: 'ResourceWatcher',
          component: () => import('@/views/watcher/ResourceWatcher.vue'),
          meta: { title: '资源监控', icon: 'Monitor' },
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = getToken()
  if (!to.meta.public && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
```

- [ ] **Step 3: Verify no duplicate route names**

The old router had duplicate `EventViewer` name (registered twice at `/events` and `/events/viewer`). The new structure only has one `EventViewer` at `events/viewer`.

- [ ] **Step 4: Test the app loads correctly**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run dev`
Expected: App starts without errors, sidebar and header visible on all pages except login.

- [ ] **Step 5: Commit**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/router/index.ts
git commit -m "refactor(router): restructure flat routes to nested under AppLayout"
```

---

### Task 2: Enhance Breadcrumb Navigation

**Files:**
- Modify: `frontend/src/components/Layout/Header.vue`

**Context:** The current breadcrumb only shows "Home > <title>". With nested routes and `meta.parent`, we can build a full breadcrumb chain: `首页 / 工作负载 / Deployment / my-app`.

- [ ] **Step 1: Update Header.vue breadcrumb logic**

Replace the breadcrumb template and add the computed logic. The breadcrumb should:
1. Always show "首页" linking to `/dashboard`
2. For each matched route with `meta.title`, show a breadcrumb item
3. If `meta.parent` exists, look up the parent route's title and insert it

Replace the `<el-breadcrumb>` section in the template (lines 7-10) with:

```html
<el-breadcrumb separator="/">
  <el-breadcrumb-item :to="{ path: '/dashboard' }">{{ t('common.home') }}</el-breadcrumb-item>
  <el-breadcrumb-item
    v-for="item in breadcrumbs"
    :key="item.path"
    :to="item.to"
  >
    {{ item.title }}
  </el-breadcrumb-item>
</el-breadcrumb>
```

Replace the script section with:

```vue
<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useClusterStore } from '@/stores/cluster'

defineEmits(['toggleCollapse'])
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const clusterStore = useClusterStore()
const { locale, t } = useI18n()

const currentLang = computed(() => locale.value === 'zh-CN' ? '中文' : 'English')

const breadcrumbs = computed(() => {
  const items: Array<{ title: string; to?: { path: string } }> = []
  const matched = route.matched

  // Find parent breadcrumb from meta.parent
  if (route.meta?.parent) {
    const parentRoute = router.getRoutes().find(r => r.name === route.meta.parent)
    if (parentRoute?.meta?.title) {
      items.push({ title: parentRoute.meta.title as string, to: { path: parentRoute.path } })
    }
  }

  // Current page title
  if (route.meta?.title) {
    items.push({ title: route.meta.title as string })
  }

  return items
})

onMounted(() => {
  clusterStore.fetchClusters()
})

function handleLangChange(lang: string) {
  locale.value = lang
  localStorage.setItem('gkube_locale', lang)
}

function handleClusterChange(val: string) {
  clusterStore.setCurrentCluster(val || null)
}

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  }
}
</script>
```

- [ ] **Step 2: Test breadcrumb navigation**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run dev`
Navigate to:
- `/dashboard` → should show "首页" only
- `/workloads/deployments` → should show "首页 / Deployment"
- `/workloads/deployments/default/my-app` → should show "首页 / Deployment / Deployment详情"

- [ ] **Step 3: Commit**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/components/Layout/Header.vue
git commit -m "feat(ui): enhance breadcrumb with route.matched and meta.parent"
```

---

### Task 3: Create useAutoRefresh Composable

**Files:**
- Create: `frontend/src/composables/useAutoRefresh.ts`

**Context:** Dashboard and list pages need auto-refresh via polling. The composable provides a countdown timer, pause/resume, and auto-cleanup on unmount.

- [ ] **Step 1: Create composables directory**

```bash
mkdir -p /Users/zqqzqq/05_github/gkube/frontend/src/composables
```

- [ ] **Step 2: Write useAutoRefresh.ts**

Create `frontend/src/composables/useAutoRefresh.ts`:

```ts
import { ref, onUnmounted } from 'vue'

export function useAutoRefresh(fetchFn: () => Promise<void>, interval = 15000) {
  const isRunning = ref(true)
  const countdown = ref(Math.floor(interval / 1000))
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let countdownTimer: ReturnType<typeof setInterval> | null = null

  function startCountdown() {
    countdown.value = Math.floor(interval / 1000)
    countdownTimer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        countdown.value = Math.floor(interval / 1000)
      }
    }, 1000)
  }

  function startPolling() {
    pollTimer = setInterval(() => {
      fetchFn()
    }, interval)
  }

  function start() {
    isRunning.value = true
    startCountdown()
    startPolling()
  }

  function stop() {
    isRunning.value = false
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    if (countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = null
    }
  }

  function toggle() {
    if (isRunning.value) {
      stop()
    } else {
      start()
    }
  }

  function refresh() {
    fetchFn()
    if (isRunning.value) {
      // Reset countdown on manual refresh
      stop()
      start()
    }
  }

  // Auto-start polling
  start()

  // Cleanup on unmount
  onUnmounted(() => {
    stop()
  })

  return {
    isRunning,
    countdown,
    toggle,
    refresh,
    start,
    stop,
  }
}
```

- [ ] **Step 3: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npx vue-tsc --noEmit 2>&1 | head -20`
Expected: No errors related to the new composable file.

- [ ] **Step 4: Commit**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/composables/useAutoRefresh.ts
git commit -m "feat(composable): add useAutoRefresh for polling with countdown"
```

---

### Task 4: Integrate Auto-Refresh into Dashboard

**Files:**
- Modify: `frontend/src/views/dashboard/DashboardView.vue`

**Context:** The dashboard loads data once on mount. We need to add `useAutoRefresh` and update the UI to show countdown and pause button.

- [ ] **Step 1: Read current DashboardView.vue**

Read the file to understand the current fetch function and template structure.

- [ ] **Step 2: Add useAutoRefresh import and integration**

In the `<script setup>` section, add the import:

```ts
import { useAutoRefresh } from '@/composables/useAutoRefresh'
```

After the existing fetch function, add:

```ts
const { isRunning, countdown, toggle, refresh: autoRefresh } = useAutoRefresh(fetchAll, 15000)
```

In the template, find the refresh button and replace it with:

```html
<el-button @click="autoRefresh()">
  <el-icon><Refresh /></el-icon>
  {{ t('common.refresh') }} ({{ countdown }}s)
</el-button>
<el-button @click="toggle()" :type="isRunning ? 'warning' : 'success'" size="small">
  {{ isRunning ? '暂停' : '恢复' }}
</el-button>
```

- [ ] **Step 3: Test dashboard auto-refresh**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run dev`
Expected: Dashboard shows countdown timer, data refreshes every 15 seconds, pause button works.

- [ ] **Step 4: Commit**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/views/dashboard/DashboardView.vue
git commit -m "feat(dashboard): integrate auto-refresh with 15s polling"
```

---

### Task 5: Integrate Auto-Refresh into List Pages

**Files:**
- Modify: `frontend/src/views/workload/DeploymentList.vue`
- Modify: `frontend/src/views/workload/PodList.vue`

**Context:** Same pattern as Dashboard. Add `useAutoRefresh` to the list fetch function.

- [ ] **Step 1: Update DeploymentList.vue**

Add import:

```ts
import { useAutoRefresh } from '@/composables/useAutoRefresh'
```

After the `fetchDeployments` function, add:

```ts
const { isRunning, countdown, toggle, refresh: autoRefresh } = useAutoRefresh(fetchDeployments, 15000)
```

In the template, replace the refresh button with:

```html
<el-button @click="autoRefresh()">
  <el-icon><Refresh /></el-icon>
  {{ t('common.refresh') }} ({{ countdown }}s)
</el-button>
<el-button @click="toggle()" :type="isRunning ? 'warning' : 'success'" size="small">
  {{ isRunning ? '暂停' : '恢复' }}
</el-button>
```

- [ ] **Step 2: Update PodList.vue**

Same pattern as DeploymentList — import `useAutoRefresh`, wrap `fetchPods`, update refresh button.

- [ ] **Step 3: Test both pages**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run dev`
Expected: Both pages auto-refresh, countdown visible, pause works.

- [ ] **Step 4: Commit**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/views/workload/DeploymentList.vue frontend/src/views/workload/PodList.vue
git commit -m "feat(workload): integrate auto-refresh into DeploymentList and PodList"
```

---

### Task 6: Fix English Translations

**Files:**
- Modify: `frontend/src/locales/en.ts`

**Context:** The current `en.ts` has many Chinese strings mixed in (e.g., `"deployment": "无Status负载"`). We need to regenerate it with proper English values while keeping the same key structure as `zh-CN.ts`.

- [ ] **Step 1: Read zh-CN.ts to get the full key structure**

Read `frontend/src/locales/zh-CN.ts` completely. This is the source of truth for all keys.

- [ ] **Step 2: Rewrite en.ts with proper English translations**

Replace the entire `en.ts` file. Every value must be English. Use the zh-CN keys as reference. Here is the complete file:

```ts
export default {
  login: {
    subtitle: 'Kubernetes Cluster Management Platform',
    usernamePlaceholder: 'Enter username',
    passwordPlaceholder: 'Enter password',
    usernameRequired: 'Username is required',
    passwordRequired: 'Password is required',
    rememberMe: 'Remember me',
    loginButton: 'Login',
    or: 'or',
    oidcLogin: 'Login with OIDC',
    loginFailed: 'Login failed',
    oidcUrlFailed: 'Failed to get OIDC login URL',
    oidcFailed: 'OIDC login failed',
  },
  common: {
    home: 'Home',
    dashboard: 'Dashboard',
    cluster: 'Cluster Management',
    workload: 'Workloads',
    config: 'Configuration',
    storage: 'Storage',
    network: 'Service Discovery',
    node: 'Node Management',
    namespace: 'Namespaces',
    event: 'Events',
    tools: 'Tools',
    system: 'System Management',
    rbac: 'RBAC',
    crd: 'CRD',
    refresh: 'Refresh',
    create: 'Create',
    delete: 'Delete',
    edit: 'Edit',
    save: 'Save',
    cancel: 'Cancel',
    confirm: 'Confirm',
    back: 'Back',
    search: 'Search',
    name: 'Name',
    namespace_label: 'Namespace',
    status: 'Status',
    age: 'Age',
    actions: 'Actions',
    yaml: 'YAML',
    detail: 'Detail',
    list: 'List',
    loading: 'Loading...',
    noData: 'No Data',
    success: 'Success',
    error: 'Error',
    warning: 'Warning',
    info: 'Info',
    selectCluster: 'Select Cluster',
    logout: 'Logout',
    yes: 'Yes',
    no: 'No',
    ok: 'OK',
    close: 'Close',
    reset: 'Reset',
    submit: 'Submit',
    backToList: 'Back to List',
    operation: 'Operation',
    description: 'Description',
    labels: 'Labels',
    annotations: 'Annotations',
    createdAt: 'Created At',
    replicas: 'Replicas',
    ready: 'Ready',
    pending: 'Pending',
    running: 'Running',
    succeeded: 'Succeeded',
    failed: 'Failed',
    unknown: 'Unknown',
    cpu: 'CPU',
    memory: 'Memory',
    storage_label: 'Storage',
    total: 'Total',
    used: 'Used',
    available: 'Available',
    enabled: 'Enabled',
    disabled: 'Disabled',
    paused: 'Paused',
    pause: 'Pause',
    resume: 'Resume',
    scale: 'Scale',
    restart: 'Restart',
    rollback: 'Rollback',
    update: 'Update',
    viewYaml: 'View YAML',
    editYaml: 'Edit YAML',
    downloadYaml: 'Download YAML',
    copyYaml: 'Copy YAML',
    formatYaml: 'Format YAML',
    copySuccess: 'Copied to clipboard',
    copyFailed: 'Copy failed',
    confirmDelete: 'Confirm Delete',
    confirmDeleteMsg: 'Are you sure you want to delete this resource?',
    batchDelete: 'Batch Delete',
    selectedItems: 'selected',
    clearSelection: 'Clear Selection',
    refreshPaused: 'Paused',
    refreshRunning: 'Running',
    autoRefresh: 'Auto Refresh',
    countdown: 'Countdown',
    noEvents: 'No Events',
    noPods: 'No Pods',
    noContainers: 'No Containers',
    noConditions: 'No Conditions',
    noLabels: 'No Labels',
    noAnnotations: 'No Annotations',
    noTolerations: 'No Tolerations',
    noNodeSelector: 'No Node Selector',
    noImagePullSecrets: 'No Image Pull Secrets',
  },
  sidebar: {
    dashboard: 'Dashboard',
    search: 'Search',
    systemOverview: 'System Overview',
    clusters: 'Clusters',
    workloads: 'Workloads',
    pods: 'Pods',
    deployments: 'Deployments',
    statefulsets: 'StatefulSets',
    daemonsets: 'DaemonSets',
    jobs: 'Jobs',
    cronjobs: 'CronJobs',
    hpa: 'HPA',
    pdb: 'PDB',
    config: 'Configuration',
    configmaps: 'ConfigMaps',
    secrets: 'Secrets',
    resourcequotas: 'ResourceQuotas',
    limitranges: 'LimitRanges',
    storage: 'Storage',
    pvs: 'PersistentVolumes',
    pvcs: 'PVCs',
    storageclasses: 'StorageClasses',
    network: 'Network',
    services: 'Services',
    ingresses: 'Ingresses',
    networkpolicies: 'NetworkPolicies',
    nodes: 'Nodes',
    namespaces: 'Namespaces',
    events: 'Events',
    rbac: 'RBAC',
    crd: 'CRD',
    catalog: 'App Catalog',
    gitops: 'GitOps',
    tenancy: 'Tenancy',
    approvals: 'Approvals',
    tools: 'Tools',
    terminal: 'Terminal',
    logs: 'Logs',
    monitoring: 'Monitoring',
    topology: 'Topology',
    diff: 'Resource Diff',
    yamlEditor: 'YAML Editor',
    system: 'System',
    users: 'Users',
    roles: 'Roles',
    authSettings: 'Auth Settings',
    audit: 'Audit Log',
  },
  cluster: {
    title: 'Cluster Management',
    create: 'Create Cluster',
    name: 'Cluster Name',
    apiServer: 'API Server',
    kubeconfig: 'Kubeconfig',
    status: 'Status',
    version: 'Version',
    nodeCount: 'Node Count',
    connected: 'Connected',
    disconnected: 'Disconnected',
    unknown_status: 'Unknown',
    testConnection: 'Test Connection',
    deleteConfirm: 'Are you sure you want to delete this cluster?',
    createSuccess: 'Cluster created successfully',
    createFailed: 'Failed to create cluster',
    deleteSuccess: 'Cluster deleted successfully',
    deleteFailed: 'Failed to delete cluster',
    connectionSuccess: 'Connection successful',
    connectionFailed: 'Connection failed',
    pasteKubeconfig: 'Paste your kubeconfig here',
    clusterNamePlaceholder: 'Enter cluster name',
    apiServerPlaceholder: 'Enter API server URL',
  },
  workload: {
    deployment: 'Deployment',
    statefulset: 'StatefulSet',
    daemonset: 'DaemonSet',
    job: 'Job',
    cronjob: 'CronJob',
    pod: 'Pod',
    container: 'Container',
    image: 'Image',
    imagePullPolicy: 'Image Pull Policy',
    ports: 'Ports',
    envVars: 'Environment Variables',
    resources: 'Resources',
    requests: 'Requests',
    limits: 'Limits',
    volumeMounts: 'Volume Mounts',
    volumes: 'Volumes',
    probes: 'Probes',
    livenessProbe: 'Liveness Probe',
    readinessProbe: 'Readiness Probe',
    startupProbe: 'Startup Probe',
    securityContext: 'Security Context',
    strategy: 'Strategy',
    nodeSelector: 'Node Selector',
    tolerations: 'Tolerations',
    imagePullSecrets: 'Image Pull Secrets',
    serviceAccount: 'Service Account',
    selector: 'Selector',
    matchLabels: 'Match Labels',
    minReadySeconds: 'Min Ready Seconds',
    revisionHistoryLimit: 'Revision History Limit',
    progressDeadlineSeconds: 'Progress Deadline Seconds',
    rollingUpdate: 'Rolling Update',
    recreate: 'Recreate',
    maxSurge: 'Max Surge',
    maxUnavailable: 'Max Unavailable',
    nodeName: 'Node Name',
    hostIP: 'Host IP',
    podIP: 'Pod IP',
    restartCount: 'Restart Count',
    containerID: 'Container ID',
    terminated: 'Terminated',
    waiting: 'Waiting',
    startedAt: 'Started At',
    finishedAt: 'Finished At',
    exitCode: 'Exit Code',
    reason: 'Reason',
    message: 'Message',
    conditions: 'Conditions',
    lastUpdateTime: 'Last Update Time',
    lastTransitionTime: 'Last Transition Time',
    targetReplicas: 'Target Replicas',
    currentReplicas: 'Current Replicas',
    readyReplicas: 'Ready Replicas',
    availableReplicas: 'Available Replicas',
    unavailableReplicas: 'Unavailable Replicas',
    updatedReplicas: 'Updated Replicas',
    collisionCount: 'Collision Count',
    cronExpression: 'Cron Expression',
    schedule: 'Schedule',
    suspend: 'Suspend',
    concurrencyPolicy: 'Concurrency Policy',
    failedJobsHistoryLimit: 'Failed Jobs History Limit',
    successfulJobsHistoryLimit: 'Successful Jobs History Limit',
    startingDeadlineSeconds: 'Starting Deadline Seconds',
    parallelism: 'Parallelism',
    completions: 'Completions',
    backoffLimit: 'Backoff Limit',
    activeDeadlineSeconds: 'Active Deadline Seconds',
    ttlSecondsAfterFinished: 'TTL Seconds After Finished',
    maxPods: 'Max Pods',
    minPods: 'Min Pods',
    targetCPU: 'Target CPU',
    targetMemory: 'Target Memory',
    metrics: 'Metrics',
    currentMetrics: 'Current Metrics',
    behavior: 'Behavior',
    scaleUp: 'Scale Up',
    scaleDown: 'Scale Down',
    stabilizationWindowSeconds: 'Stabilization Window Seconds',
    selectPolicy: 'Select Policy',
    maxReplicas: 'Max Replicas',
    minReplicas: 'Min Replicas',
    podSelector: 'Pod Selector',
    maxUnavailablePods: 'Max Unavailable Pods',
    minAvailablePods: 'Min Available Pods',
    unhealthyPodEvictionPolicy: 'Unhealthy Pod Eviction Policy',
  },
  network: {
    service: 'Service',
    ingress: 'Ingress',
    networkPolicy: 'NetworkPolicy',
    clusterIP: 'ClusterIP',
    nodePort: 'NodePort',
    loadBalancer: 'LoadBalancer',
    externalName: 'ExternalName',
    sessionAffinity: 'Session Affinity',
    externalTrafficPolicy: 'External Traffic Policy',
    loadBalancerIP: 'Load Balancer IP',
    targetPort: 'Target Port',
    protocol: 'Protocol',
    port: 'Port',
    nodePort_label: 'Node Port',
    endpoints: 'Endpoints',
    ingressClass: 'Ingress Class',
    tls: 'TLS',
    hosts: 'Hosts',
    paths: 'Paths',
    pathType: 'Path Type',
    backend: 'Backend',
    rules: 'Rules',
    defaultBackend: 'Default Backend',
    ingressClassName: 'Ingress Class Name',
    secretName: 'Secret Name',
    allowAll: 'Allow All',
    denyAll: 'Deny All',
    ingressRules: 'Ingress Rules',
    egressRules: 'Egress Rules',
    podSelector_label: 'Pod Selector',
    policyTypes: 'Policy Types',
    ipBlock: 'IP Block',
    cidr: 'CIDR',
    except: 'Except',
    namespaceSelector: 'Namespace Selector',
  },
  storage: {
    persistentVolume: 'PersistentVolume',
    persistentVolumeClaim: 'PersistentVolumeClaim',
    storageClass: 'StorageClass',
    capacity: 'Capacity',
    accessModes: 'Access Modes',
    reclaimPolicy: 'Reclaim Policy',
    volumeMode: 'Volume Mode',
    storageClassName: 'Storage Class Name',
    provisioner: 'Provisioner',
    mountOptions: 'Mount Options',
    volumeBindingMode: 'Volume Binding Mode',
    allowVolumeExpansion: 'Allow Volume Expansion',
    parameters: 'Parameters',
    claimRef: 'Claim Ref',
    phase: 'Phase',
    reason_label: 'Reason',
    available: 'Available',
    bound: 'Bound',
    released: 'Released',
    failed: 'Failed',
    retain: 'Retain',
    recycle: 'Recycle',
    deletePolicy: 'Delete',
    fileSystem: 'Filesystem',
    block: 'Block',
    immediate: 'Immediate',
    waitForFirstConsumer: 'WaitForFirstConsumer',
    readWriteOnce: 'ReadWriteOnce',
    readOnlyMany: 'ReadOnlyMany',
    readWriteMany: 'ReadWriteMany',
    readWriteOncePod: 'ReadWriteOncePod',
  },
  config: {
    configmap: 'ConfigMap',
    secret: 'Secret',
    resourceQuota: 'ResourceQuota',
    limitRange: 'LimitRange',
    data: 'Data',
    immutable: 'Immutable',
    secretType: 'Secret Type',
    opaque: 'Opaque',
    dockerConfigJson: 'Docker Config JSON',
    tlsSecret: 'TLS Secret',
    basicAuth: 'Basic Auth',
    sshAuth: 'SSH Auth',
    serviceAccountToken: 'Service Account Token',
    hardLimits: 'Hard Limits',
    usedLimits: 'Used Limits',
    max: 'Max',
    min: 'Min',
    default: 'Default',
    defaultRequest: 'Default Request',
    maxLimitRequestRatio: 'Max Limit Request Ratio',
    scopeSelector: 'Scope Selector',
    scopes: 'Scopes',
    operator: 'Operator',
    values: 'Values',
  },
  node: {
    node: 'Node',
    roles: 'Roles',
    os: 'OS',
    kernelVersion: 'Kernel Version',
    containerRuntime: 'Container Runtime',
    kubeletVersion: 'Kubelet Version',
    architecture: 'Architecture',
    addresses: 'Addresses',
    conditions: 'Conditions',
    capacity: 'Capacity',
    allocatable: 'Allocatable',
    daemonEndpoints: 'Daemon Endpoints',
    nodeInfo: 'Node Info',
    unschedulable: 'Unschedulable',
    taints: 'Taints',
    cordon: 'Cordon',
    uncordon: 'Uncordon',
    drain: 'Drain',
    taintKey: 'Key',
    taintValue: 'Value',
    taintEffect: 'Effect',
    noSchedule: 'NoSchedule',
    noExecute: 'NoExecute',
    preferNoSchedule: 'PreferNoSchedule',
  },
  namespace: {
    namespace: 'Namespace',
    active: 'Active',
    terminating: 'Terminating',
    manager: 'Namespace Manager',
  },
  event: {
    event: 'Event',
    type: 'Type',
    reason_label: 'Reason',
    message: 'Message',
    lastTimestamp: 'Last Timestamp',
    firstTimestamp: 'First Timestamp',
    count: 'Count',
    involvedObject: 'Involved Object',
    normal: 'Normal',
    warning: 'Warning',
    viewer: 'Event Viewer',
  },
  rbac: {
    rbac: 'RBAC',
    matrix: 'Permission Matrix',
    role: 'Role',
    clusterRole: 'ClusterRole',
    roleBinding: 'RoleBinding',
    clusterRoleBinding: 'ClusterRoleBinding',
    subjects: 'Subjects',
    subjectKind: 'Subject Kind',
    user: 'User',
    group: 'Group',
    serviceAccount: 'Service Account',
    roleName: 'Role Name',
    apiGroups: 'API Groups',
    resources: 'Resources',
    verbs: 'Verbs',
    nonResourceURLs: 'Non-Resource URLs',
  },
  terminal: {
    terminal: 'Terminal',
    connect: 'Connect',
    disconnect: 'Disconnect',
    connected: 'Connected',
    disconnected: 'Disconnected',
    connecting: 'Connecting...',
    selectPod: 'Select Pod',
    selectContainer: 'Select Container',
    selectNamespace: 'Select Namespace',
    terminalSettings: 'Terminal Settings',
    fontSize: 'Font Size',
    fontFamily: 'Font Family',
    cursorBlink: 'Cursor Blink',
    scrollback: 'Scrollback',
  },
  logviewer: {
    logViewer: 'Log Viewer',
    selectPod: 'Select Pod',
    selectContainer: 'Select Container',
    tailLines: 'Tail Lines',
    sinceSeconds: 'Since Seconds',
    follow: 'Follow',
    previous: 'Previous',
    timestamps: 'Timestamps',
    autoScroll: 'Auto Scroll',
    clear: 'Clear',
    download: 'Download',
    searchInLogs: 'Search in logs',
  },
  monitoring: {
    monitoring: 'Monitoring',
    dashboard: 'Dashboard',
    resources: 'Resources',
    prometheus: 'Prometheus',
    cpuUsage: 'CPU Usage',
    memoryUsage: 'Memory Usage',
    networkIO: 'Network I/O',
    diskIO: 'Disk I/O',
    podCount: 'Pod Count',
    nodeCount: 'Node Count',
    timeRange: 'Time Range',
    refreshInterval: 'Refresh Interval',
    last5m: 'Last 5 minutes',
    last15m: 'Last 15 minutes',
    last30m: 'Last 30 minutes',
    last1h: 'Last 1 hour',
    last3h: 'Last 3 hours',
    last6h: 'Last 6 hours',
    last12h: 'Last 12 hours',
    last24h: 'Last 24 hours',
    custom: 'Custom',
    query: 'Query',
    execute: 'Execute',
    addChart: 'Add Chart',
    removeChart: 'Remove Chart',
  },
  topology: {
    topology: 'Topology',
    graph: 'Graph',
    refresh: 'Refresh',
    zoomIn: 'Zoom In',
    zoomOut: 'Zoom Out',
    fitView: 'Fit View',
    showLabels: 'Show Labels',
    showRelations: 'Show Relations',
    podTopology: 'Pod Topology',
    serviceTopology: 'Service Topology',
    nodeTopology: 'Node Topology',
    namespaceTopology: 'Namespace Topology',
  },
  catalog: {
    catalog: 'App Catalog',
    install: 'Install',
    uninstall: 'Uninstall',
    repository: 'Repository',
    version: 'Version',
    description: 'Description',
    maintainers: 'Maintainers',
    keywords: 'Keywords',
    home: 'Home',
    sources: 'Sources',
    chartName: 'Chart Name',
    chartVersion: 'Chart Version',
    appVersion: 'App Version',
    values: 'Values',
    releaseName: 'Release Name',
    releaseNamespace: 'Release Namespace',
  },
  gitops: {
    gitops: 'GitOps',
    application: 'Application',
    repository: 'Repository',
    syncPolicy: 'Sync Policy',
    autoSync: 'Auto Sync',
    selfHeal: 'Self Heal',
    prune: 'Prune',
    lastSync: 'Last Sync',
    syncStatus: 'Sync Status',
    healthStatus: 'Health Status',
    synced: 'Synced',
    outOfSync: 'OutOfSync',
    healthy: 'Healthy',
    degraded: 'Degraded',
    progressing: 'Progressing',
    suspended: 'Suspended',
    missing: 'Missing',
    unknown_status: 'Unknown',
    sourceRepo: 'Source Repository',
    sourcePath: 'Source Path',
    sourceTargetRevision: 'Target Revision',
    destinationServer: 'Destination Server',
    destinationNamespace: 'Destination Namespace',
  },
  tenancy: {
    tenancy: 'Tenancy',
    tenant: 'Tenant',
    tenantName: 'Tenant Name',
    tenantDesc: 'Tenant Description',
    namespacePrefix: 'Namespace Prefix',
    resourceQuotas: 'Resource Quotas',
    assignedNamespaces: 'Assigned Namespaces',
    createTenant: 'Create Tenant',
    deleteTenant: 'Delete Tenant',
    manageNamespaces: 'Manage Namespaces',
  },
  approval: {
    approval: 'Approval',
    pendingApproval: 'Pending Approval',
    approved: 'Approved',
    rejected: 'Rejected',
    approve: 'Approve',
    reject: 'Reject',
    approver: 'Approver',
    requestor: 'Requestor',
    resourceType: 'Resource Type',
    resourceName: 'Resource Name',
    action: 'Action',
    requestTime: 'Request Time',
    approveTime: 'Approve Time',
    comment: 'Comment',
  },
  audit: {
    audit: 'Audit Log',
    operator: 'Operator',
    resource: 'Resource',
    action: 'Action',
    timestamp: 'Timestamp',
    ipAddress: 'IP Address',
    userAgent: 'User Agent',
    requestBody: 'Request Body',
    statusCode: 'Status Code',
    clearAll: 'Clear All',
    export: 'Export',
    filterByUser: 'Filter by User',
    filterByResource: 'Filter by Resource',
    filterByAction: 'Filter by Action',
    filterByTime: 'Filter by Time',
  },
  notification: {
    notification: 'Notification',
    center: 'Notification Center',
    markAllRead: 'Mark All Read',
    clearAll: 'Clear All',
    read: 'Read',
    unread: 'Unread',
    type: 'Type',
    message: 'Message',
    timestamp: 'Timestamp',
    noNotifications: 'No Notifications',
  },
  search: {
    search: 'Resource Search',
    searchPlaceholder: 'Search by name, label, annotation...',
    resourceType: 'Resource Type',
    allResources: 'All Resources',
    searchResults: 'Search Results',
    noResults: 'No results found',
    nameFilter: 'Name Filter',
    labelFilter: 'Label Filter',
    namespaceFilter: 'Namespace Filter',
    statusFilter: 'Status Filter',
  },
  watcher: {
    watcher: 'Resource Watcher',
    watching: 'Watching',
    stopped: 'Stopped',
    startWatch: 'Start Watch',
    stopWatch: 'Stop Watch',
    events: 'Events',
    added: 'Added',
    modified: 'Modified',
    deleted: 'Deleted',
  },
  crd: {
    crd: 'Custom Resource Definition',
    group: 'Group',
    version: 'Version',
    kind: 'Kind',
    plural: 'Plural',
    singular: 'Singular',
    scope: 'Scope',
    namespaced: 'Namespaced',
    cluster: 'Cluster',
    established: 'Established',
    storedVersions: 'Stored Versions',
    additionalPrinterColumns: 'Additional Printer Columns',
    customResource: 'Custom Resource',
  },
  tools: {
    diff: 'Resource Diff',
    yamlEditor: 'YAML Editor',
    batch: 'Batch Operations',
    originalYaml: 'Original YAML',
    modifiedYaml: 'Modified YAML',
    compare: 'Compare',
    apply: 'Apply',
    differences: 'Differences',
    noDifferences: 'No Differences',
    selectResource: 'Select Resource',
    executeBatch: 'Execute Batch',
    batchResult: 'Batch Result',
  },
}
```

- [ ] **Step 3: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npx vue-tsc --noEmit 2>&1 | head -20`
Expected: No type errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/locales/en.ts
git commit -m "fix(i18n): regenerate English translations with proper values"
```

---

### Task 7: Final Verification

**Files:** None (verification only)

- [ ] **Step 1: Build the frontend**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: Build succeeds with no errors.

- [ ] **Step 2: Verify all routes work**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run dev`
Navigate to these URLs and verify they load with sidebar + header:
- `/dashboard`
- `/workloads/deployments`
- `/workloads/pods`
- `/config/configmaps`
- `/storage/pvs`
- `/services`
- `/nodes`
- `/namespaces`
- `/events`
- `/terminal`
- `/logs`
- `/login` (should NOT have sidebar/header)

- [ ] **Step 3: Verify breadcrumbs show correctly**

Navigate to:
- `/workloads/deployments` → "首页 / Deployment"
- `/workloads/deployments/default/my-app` → "首页 / Deployment / Deployment详情"

- [ ] **Step 4: Verify auto-refresh works**

On `/dashboard` and `/workloads/deployments`:
- Countdown timer visible
- Data refreshes every 15 seconds
- Pause/Resume button works

- [ ] **Step 5: Verify English translations**

Switch language to English. Check:
- Sidebar labels are all English
- Dashboard labels are all English
- No Chinese characters visible in English mode

- [ ] **Step 6: Final commit**

```bash
cd /Users/zqqzqq/05_github/gkube
git add -A
git commit -m "feat(ui): complete frontend UX optimization - nested routes, breadcrumbs, auto-refresh, i18n fix"
```
