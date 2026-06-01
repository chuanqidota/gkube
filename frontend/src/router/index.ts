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
      path: '/ingresses',
      name: 'IngressList',
      component: () => import('@/views/network/IngressList.vue'),
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
      path: '/config/secrets',
      name: 'SecretList',
      component: () => import('@/views/config/SecretList.vue'),
    },
    // Storage routes
    {
      path: '/storage/pvs',
      name: 'PVList',
      component: () => import('@/views/storage/PVList.vue'),
    },
    {
      path: '/storage/pvcs',
      name: 'PVCList',
      component: () => import('@/views/storage/PVCList.vue'),
    },
    {
      path: '/storage/storageclasses',
      name: 'StorageClassList',
      component: () => import('@/views/storage/StorageClassList.vue'),
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
