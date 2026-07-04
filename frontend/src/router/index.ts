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
    // Fullscreen layout for terminal and logs (no sidebar/header, opens in new tab)
    {
      path: '/',
      component: () => import('@/components/Layout/FullscreenLayout.vue'),
      children: [
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
      ],
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
          path: 'system/overview',
          redirect: '/dashboard',
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
        // Workloads - Pod
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
        // Workloads - Deployment
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
        // Workloads - StatefulSet
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
        // Workloads - DaemonSet
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
        // Workloads - Job
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
        // Workloads - CronJob
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
        // Workloads - ReplicaSet
        {
          path: 'workloads/replicasets',
          name: 'ReplicaSetList',
          component: () => import('@/views/workload/ReplicaSetList.vue'),
          meta: { title: 'ReplicaSet', icon: 'CopyDocument' },
        },
        // Workloads - HPA
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
        // Workloads - PDB
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
        // Config - ConfigMap
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
        // Config - Secret
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
        // Config - ResourceQuota
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
        // Config - LimitRange
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
        // Storage - PV
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
        // Storage - PVC
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
        // Storage - StorageClass
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
        // Storage - VolumeSnapshot
        {
          path: 'storage/volumesnapshots',
          name: 'VolumeSnapshotList',
          component: () => import('@/views/storage/VolumeSnapshotList.vue'),
          meta: { title: 'VolumeSnapshot', icon: 'Camera' },
        },
        {
          path: 'storage/volumesnapshots/create',
          name: 'VolumeSnapshotCreate',
          component: () => import('@/views/storage/VolumeSnapshotCreate.vue'),
          meta: { title: '创建VolumeSnapshot', parent: 'VolumeSnapshotList' },
        },
        {
          path: 'storage/volumesnapshots/:namespace/:name',
          name: 'VolumeSnapshotDetail',
          component: () => import('@/views/storage/VolumeSnapshotDetail.vue'),
          props: true,
          meta: { title: 'VolumeSnapshot详情', parent: 'VolumeSnapshotList' },
        },
        // Storage - VolumeSnapshotClass
        {
          path: 'storage/volumesnapshotclasses',
          name: 'VolumeSnapshotClassList',
          component: () => import('@/views/storage/VolumeSnapshotClassList.vue'),
          meta: { title: 'VolumeSnapshotClass', icon: 'CameraFilled' },
        },
        {
          path: 'storage/volumesnapshotclasses/create',
          name: 'VolumeSnapshotClassCreate',
          component: () => import('@/views/storage/VolumeSnapshotClassCreate.vue'),
          meta: { title: '创建VolumeSnapshotClass', parent: 'VolumeSnapshotClassList' },
        },
        {
          path: 'storage/volumesnapshotclasses/:name',
          name: 'VolumeSnapshotClassDetail',
          component: () => import('@/views/storage/VolumeSnapshotClassDetail.vue'),
          props: true,
          meta: { title: 'VolumeSnapshotClass详情', parent: 'VolumeSnapshotClassList' },
        },
        // Network - Service
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
        // Network - Ingress
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
        // Network - NetworkPolicy
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
        // Events
        {
          path: 'events',
          name: 'EventList',
          component: () => import('@/views/event/EventList.vue'),
          meta: { title: '事件', icon: 'Bell' },
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
        // RBAC - ServiceAccount
        {
          path: 'rbac/serviceaccounts',
          name: 'ServiceAccountList',
          component: () => import('@/views/rbac/ServiceAccountList.vue'),
          meta: { title: 'ServiceAccount', icon: 'User' },
        },
        // RBAC - Role
        {
          path: 'rbac/roles',
          name: 'RBACRoleList',
          component: () => import('@/views/rbac/RoleList.vue'),
          meta: { title: 'Role', icon: 'UserFilled' },
        },
        // RBAC - ClusterRole
        {
          path: 'rbac/clusterroles',
          name: 'ClusterRoleList',
          component: () => import('@/views/rbac/ClusterRoleList.vue'),
          meta: { title: 'ClusterRole', icon: 'Stamp' },
        },
        // RBAC - RoleBinding
        {
          path: 'rbac/rolebindings',
          name: 'RoleBindingList',
          component: () => import('@/views/rbac/RoleBindingList.vue'),
          meta: { title: 'RoleBinding', icon: 'Link' },
        },
        // RBAC - ClusterRoleBinding
        {
          path: 'rbac/clusterrolebindings',
          name: 'ClusterRoleBindingList',
          component: () => import('@/views/rbac/ClusterRoleBindingList.vue'),
          meta: { title: 'ClusterRoleBinding', icon: 'CircleCheck' },
        },
        // System - Users
        {
          path: 'users',
          name: 'UserList',
          component: () => import('@/views/UserList.vue'),
          meta: { title: '用户管理', icon: 'User' },
        },
        // System - Roles
        {
          path: 'roles',
          name: 'RoleList',
          component: () => import('@/views/RoleList.vue'),
          meta: { title: '角色管理', icon: 'UserFilled' },
        },
        // System - Auth Settings
        {
          path: 'settings/auth',
          name: 'OIDCSettings',
          component: () => import('@/views/settings/OIDCSettings.vue'),
          meta: { title: '认证设置', icon: 'Setting' },
        },
        // System - Audit
        {
          path: 'audit',
          name: 'AuditLog',
          component: () => import('@/views/audit/AuditLog.vue'),
          meta: { title: '审计日志', icon: 'Document' },
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
