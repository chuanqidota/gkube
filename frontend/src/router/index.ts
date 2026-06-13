import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
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
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/dashboard/DashboardView.vue'),
    },
    {
      path: '/clusters',
      name: 'ClusterList',
      component: () => import('@/views/ClusterList.vue'),
    },
    {
      path: '/clusters/create',
      name: 'ClusterCreate',
      component: () => import('@/views/ClusterCreate.vue'),
    },
    {
      path: '/clusters/:id',
      name: 'ClusterDetail',
      component: () => import('@/views/ClusterDetail.vue'),
      props: true,
    },
    {
      path: '/users',
      name: 'UserList',
      component: () => import('@/views/UserList.vue'),
    },
    {
      path: '/roles',
      name: 'RoleList',
      component: () => import('@/views/RoleList.vue'),
    },
    // Settings routes
    {
      path: '/settings/auth',
      name: 'OIDCSettings',
      component: () => import('@/views/settings/OIDCSettings.vue'),
    },
    {
      path: '/terminal',
      name: 'Terminal',
      component: () => import('@/views/terminal/TerminalView.vue'),
      meta: { title: 'Web 终端' },
    },
    {
      path: '/logs',
      name: 'Logs',
      component: () => import('@/views/logviewer/LogView.vue'),
      meta: { title: '日志查看' },
    },
    // Workload routes
    {
      path: '/workloads/pods',
      name: 'PodList',
      component: () => import('@/views/workload/PodList.vue'),
    },
    {
      path: '/workloads/pods/:namespace/:name',
      name: 'PodDetail',
      component: () => import('@/views/workload/PodDetail.vue'),
      props: true,
    },
    {
      path: '/workloads/deployments',
      name: 'DeploymentList',
      component: () => import('@/views/workload/DeploymentList.vue'),
    },
    {
      path: '/workloads/deployments/create',
      name: 'DeploymentCreate',
      component: () => import('@/views/workload/DeploymentCreate.vue'),
    },
    {
      path: '/workloads/deployments/:namespace/:name',
      name: 'DeploymentDetail',
      component: () => import('@/views/workload/DeploymentDetail.vue'),
      props: true,
    },
    {
      path: '/workloads/statefulsets',
      name: 'StatefulSetList',
      component: () => import('@/views/workload/StatefulSetList.vue'),
    },
    {
      path: '/workloads/statefulsets/create',
      name: 'StatefulSetCreate',
      component: () => import('@/views/workload/StatefulSetCreate.vue'),
    },
    {
      path: '/workloads/statefulsets/:namespace/:name',
      name: 'StatefulSetDetail',
      component: () => import('@/views/workload/StatefulSetDetail.vue'),
      props: true,
    },
    {
      path: '/workloads/daemonsets',
      name: 'DaemonSetList',
      component: () => import('@/views/workload/DaemonSetList.vue'),
    },
    {
      path: '/workloads/daemonsets/create',
      name: 'DaemonSetCreate',
      component: () => import('@/views/workload/DaemonSetCreate.vue'),
    },
    {
      path: '/workloads/daemonsets/:namespace/:name',
      name: 'DaemonSetDetail',
      component: () => import('@/views/workload/DaemonSetDetail.vue'),
      props: true,
    },
    {
      path: '/workloads/jobs',
      name: 'JobList',
      component: () => import('@/views/workload/JobList.vue'),
    },
    {
      path: '/workloads/jobs/create',
      name: 'JobCreate',
      component: () => import('@/views/workload/JobCreate.vue'),
    },
    {
      path: '/workloads/jobs/:namespace/:name',
      name: 'JobDetail',
      component: () => import('@/views/workload/JobDetail.vue'),
      props: true,
    },
    {
      path: '/workloads/cronjobs',
      name: 'CronJobList',
      component: () => import('@/views/workload/CronJobList.vue'),
    },
    {
      path: '/workloads/cronjobs/create',
      name: 'CronJobCreate',
      component: () => import('@/views/workload/CronJobCreate.vue'),
    },
    {
      path: '/workloads/cronjobs/:namespace/:name',
      name: 'CronJobDetail',
      component: () => import('@/views/workload/CronJobDetail.vue'),
      props: true,
    },
    // HPA routes
    {
      path: '/workloads/hpa',
      name: 'HPAList',
      component: () => import('@/views/workload/hpa/HPAList.vue'),
    },
    {
      path: '/workloads/hpa/create',
      name: 'HPACreate',
      component: () => import('@/views/workload/hpa/HPACreate.vue'),
    },
    {
      path: '/workloads/hpa/:namespace/:name',
      name: 'HPADetail',
      component: () => import('@/views/workload/hpa/HPADetail.vue'),
      props: true,
    },
    // PDB routes
    {
      path: '/workloads/pdb',
      name: 'PDBList',
      component: () => import('@/views/workload/pdb/PDBList.vue'),
    },
    {
      path: '/workloads/pdb/create',
      name: 'PDBCreate',
      component: () => import('@/views/workload/pdb/PDBCreate.vue'),
    },
    // CRD routes
    {
      path: '/crd',
      name: 'CRDList',
      component: () => import('@/views/crd/CRDList.vue'),
    },
    {
      path: '/crd/create',
      name: 'CRDCreate',
      component: () => import('@/views/crd/CRDCreate.vue'),
    },
    {
      path: '/crd/resources',
      name: 'CustomResourceList',
      component: () => import('@/views/crd/CustomResourceList.vue'),
    },
    // Network routes
    {
      path: '/services',
      name: 'ServiceList',
      component: () => import('@/views/network/ServiceList.vue'),
    },
    {
      path: '/services/create',
      name: 'ServiceCreate',
      component: () => import('@/views/network/ServiceCreate.vue'),
    },
    {
      path: '/services/:namespace/:name',
      name: 'ServiceDetail',
      component: () => import('@/views/network/ServiceDetail.vue'),
      props: true,
    },
    {
      path: '/ingresses',
      name: 'IngressList',
      component: () => import('@/views/network/IngressList.vue'),
    },
    {
      path: '/ingresses/create',
      name: 'IngressCreate',
      component: () => import('@/views/network/IngressCreate.vue'),
    },
    {
      path: '/ingresses/:namespace/:name',
      name: 'IngressDetail',
      component: () => import('@/views/network/IngressDetail.vue'),
      props: true,
    },
    // NetworkPolicy routes
    {
      path: '/network/networkpolicies',
      name: 'NetworkPolicyList',
      component: () => import('@/views/network/networkpolicy/NetworkPolicyList.vue'),
    },
    {
      path: '/network/networkpolicies/create',
      name: 'NetworkPolicyCreate',
      component: () => import('@/views/network/networkpolicy/NetworkPolicyCreate.vue'),
    },
    {
      path: '/network/networkpolicies/:namespace/:name',
      name: 'NetworkPolicyDetail',
      component: () => import('@/views/network/networkpolicy/NetworkPolicyDetail.vue'),
      props: true,
    },
    // Node routes
    {
      path: '/nodes',
      name: 'NodeList',
      component: () => import('@/views/node/NodeList.vue'),
    },
    {
      path: '/nodes/:name',
      name: 'NodeDetail',
      component: () => import('@/views/node/NodeDetail.vue'),
      props: true,
    },
    // Namespace routes
    {
      path: '/namespaces',
      name: 'NamespaceList',
      component: () => import('@/views/namespace/NamespaceList.vue'),
    },
    {
      path: '/namespaces/:name',
      name: 'NamespaceDetail',
      component: () => import('@/views/namespace/NamespaceDetail.vue'),
      props: true,
    },
    // Events
    {
      path: '/events',
      name: 'EventList',
      component: () => import('@/views/event/EventList.vue'),
    },
    // RBAC routes
    {
      path: '/rbac',
      name: 'RBACView',
      component: () => import('@/views/rbac/RBACView.vue'),
    },
    // Monitoring routes
    {
      path: '/monitoring',
      name: 'MonitoringView',
      component: () => import('@/views/monitoring/MonitoringView.vue'),
    },
    {
      path: '/monitoring/dashboard',
      name: 'MonitoringDashboard',
      component: () => import('@/views/monitoring/MonitoringDashboard.vue'),
    },
    {
      path: '/monitoring/resources',
      name: 'ResourceDashboard',
      component: () => import('@/views/monitoring/ResourceDashboard.vue'),
    },
    {
      path: '/monitoring/prometheus',
      name: 'PrometheusView',
      component: () => import('@/views/monitoring/PrometheusView.vue'),
    },
    // Topology routes
    {
      path: '/topology',
      name: 'TopologyView',
      component: () => import('@/views/topology/TopologyView.vue'),
    },
    {
      path: '/topology/graph',
      name: 'TopologyGraph',
      component: () => import('@/views/topology/TopologyGraph.vue'),
    },
    // Config routes
    {
      path: '/config/configmaps',
      name: 'ConfigMapList',
      component: () => import('@/views/config/ConfigMapList.vue'),
    },
    {
      path: '/config/configmaps/create',
      name: 'ConfigMapCreate',
      component: () => import('@/views/config/ConfigMapCreate.vue'),
    },
    {
      path: '/config/configmaps/:namespace/:name',
      name: 'ConfigMapDetail',
      component: () => import('@/views/config/ConfigMapDetail.vue'),
      props: true,
    },
    {
      path: '/config/secrets',
      name: 'SecretList',
      component: () => import('@/views/config/SecretList.vue'),
    },
    {
      path: '/config/secrets/create',
      name: 'SecretCreate',
      component: () => import('@/views/config/SecretCreate.vue'),
    },
    {
      path: '/config/secrets/:namespace/:name',
      name: 'SecretDetail',
      component: () => import('@/views/config/SecretDetail.vue'),
      props: true,
    },
    // ResourceQuota routes
    {
      path: '/config/resourcequotas',
      name: 'ResourceQuotaList',
      component: () => import('@/views/config/resourcequota/ResourceQuotaList.vue'),
    },
    {
      path: '/config/resourcequotas/create',
      name: 'ResourceQuotaCreate',
      component: () => import('@/views/config/resourcequota/ResourceQuotaCreate.vue'),
    },
    // LimitRange routes
    {
      path: '/config/limitranges',
      name: 'LimitRangeList',
      component: () => import('@/views/config/limitrange/LimitRangeList.vue'),
    },
    {
      path: '/config/limitranges/create',
      name: 'LimitRangeCreate',
      component: () => import('@/views/config/limitrange/LimitRangeCreate.vue'),
    },
    // Storage routes
    {
      path: '/storage/pvs',
      name: 'PVList',
      component: () => import('@/views/storage/PVList.vue'),
    },
    {
      path: '/storage/pvs/create',
      name: 'PVCreate',
      component: () => import('@/views/storage/PVCreate.vue'),
    },
    {
      path: '/storage/pvs/:name',
      name: 'PVDetail',
      component: () => import('@/views/storage/PVDetail.vue'),
      props: true,
    },
    {
      path: '/storage/pvcs',
      name: 'PVCList',
      component: () => import('@/views/storage/PVCList.vue'),
    },
    {
      path: '/storage/pvcs/create',
      name: 'PVCCreate',
      component: () => import('@/views/storage/PVCCreate.vue'),
    },
    {
      path: '/storage/pvcs/:namespace/:name',
      name: 'PVCDetail',
      component: () => import('@/views/storage/PVCDetail.vue'),
      props: true,
    },
    {
      path: '/storage/storageclasses',
      name: 'StorageClassList',
      component: () => import('@/views/storage/StorageClassList.vue'),
    },
    {
      path: '/storage/storageclasses/:name',
      name: 'StorageClassDetail',
      component: () => import('@/views/storage/StorageClassDetail.vue'),
      props: true,
    },

    // Tools routes
    {
      path: '/tools/diff',
      name: 'ResourceDiff',
      component: () => import('@/views/tools/ResourceDiff.vue'),
    },

    // Catalog routes
    {
      path: '/catalog',
      name: 'AppCatalog',
      component: () => import('@/views/catalog/AppCatalog.vue'),
    },

    // GitOps routes
    {
      path: '/gitops',
      name: 'GitOpsView',
      component: () => import('@/views/gitops/GitOpsView.vue'),
    },

    // Tenancy routes
    {
      path: '/tenancy',
      name: 'TenantList',
      component: () => import('@/views/tenancy/TenantList.vue'),
    },

    // Event routes
    {
      path: '/events',
      name: 'EventViewer',
      component: () => import('@/views/event/EventViewer.vue'),
    },

    // Approval routes
    {
      path: '/approvals',
      name: 'ApprovalList',
      component: () => import('@/views/approval/ApprovalList.vue'),
    },

    // RBAC Matrix route
    {
      path: '/rbac/matrix',
      name: 'RBACMatrix',
      component: () => import('@/views/rbac/RBACMatrix.vue'),
    },

    // Event routes
    {
      path: '/events/viewer',
      name: 'EventViewer',
      component: () => import('@/views/event/EventViewer.vue'),
    },

    // Watcher routes
    {
      path: '/watcher',
      name: 'ResourceWatcher',
      component: () => import('@/views/watcher/ResourceWatcher.vue'),
    },

    // Namespace Manager route
    {
      path: '/namespaces/manager',
      name: 'NamespaceManager',
      component: () => import('@/views/namespace/NamespaceManager.vue'),
    },

    // Batch Operations route
    {
      path: '/tools/batch',
      name: 'BatchOperations',
      component: () => import('@/views/tools/BatchOperations.vue'),
    },

    // Resource Quota Detail route
    {
      path: '/config/resourcequotas/:namespace/:name',
      name: 'ResourceQuotaDetail',
      component: () => import('@/views/config/ResourceQuotaDetail.vue'),
      props: true,
    },

    // YAML Editor route
    {
      path: '/tools/yaml-editor',
      name: 'YAMLEditor',
      component: () => import('@/views/tools/YAMLEditor.vue'),
    },

    // System Overview route
    {
      path: '/system/overview',
      name: 'SystemOverview',
      component: () => import('@/views/dashboard/SystemOverview.vue'),
    },

    // Audit Log route
    {
      path: '/audit',
      name: 'AuditLog',
      component: () => import('@/views/audit/AuditLog.vue'),
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
