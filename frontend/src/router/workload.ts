import type { RouteRecordRaw } from 'vue-router'

export const workloadRoutes: RouteRecordRaw[] = [
  // Pod
  {
    path: 'workloads/pods',
    name: 'PodList',
    component: () => import('@/views/workload/PodList.vue'),
    meta: { titleKey: 'sidebar.pods', icon: 'Coin' },
  },
  {
    path: 'workloads/pods/:namespace/:name',
    name: 'PodDetail',
    component: () => import('@/views/workload/PodDetail.vue'),
    meta: { titleKey: 'workload.podDetail', parent: 'PodList' },
  },
  // Deployment
  {
    path: 'workloads/deployments',
    name: 'DeploymentList',
    component: () => import('@/views/workload/DeploymentList.vue'),
    meta: { titleKey: 'sidebar.deployments', icon: 'Files' },
  },
  {
    path: 'workloads/deployments/create',
    name: 'DeploymentCreate',
    component: () => import('@/views/workload/DeploymentCreate.vue'),
    meta: { titleKey: 'workload.deploymentCreate', parent: 'DeploymentList' },
  },
  {
    path: 'workloads/deployments/:namespace/:name',
    name: 'DeploymentDetail',
    component: () => import('@/views/workload/DeploymentDetail.vue'),
    props: true,
    meta: { titleKey: 'workload.deploymentDetail', parent: 'DeploymentList' },
  },
  // StatefulSet
  {
    path: 'workloads/statefulsets',
    name: 'StatefulSetList',
    component: () => import('@/views/workload/StatefulSetList.vue'),
    meta: { titleKey: 'sidebar.statefulsets', icon: 'Files' },
  },
  {
    path: 'workloads/statefulsets/create',
    name: 'StatefulSetCreate',
    component: () => import('@/views/workload/StatefulSetCreate.vue'),
    meta: { titleKey: 'workload.statefulsetCreate', parent: 'StatefulSetList' },
  },
  {
    path: 'workloads/statefulsets/:namespace/:name',
    name: 'StatefulSetDetail',
    component: () => import('@/views/workload/StatefulSetDetail.vue'),
    props: true,
    meta: { titleKey: 'workload.statefulsetDetail', parent: 'StatefulSetList' },
  },
  // DaemonSet
  {
    path: 'workloads/daemonsets',
    name: 'DaemonSetList',
    component: () => import('@/views/workload/DaemonSetList.vue'),
    meta: { titleKey: 'sidebar.daemonsets', icon: 'Files' },
  },
  {
    path: 'workloads/daemonsets/create',
    name: 'DaemonSetCreate',
    component: () => import('@/views/workload/DaemonSetCreate.vue'),
    meta: { titleKey: 'workload.daemonsetCreate', parent: 'DaemonSetList' },
  },
  {
    path: 'workloads/daemonsets/:namespace/:name',
    name: 'DaemonSetDetail',
    component: () => import('@/views/workload/DaemonSetDetail.vue'),
    props: true,
    meta: { titleKey: 'workload.daemonsetDetail', parent: 'DaemonSetList' },
  },
  // Job
  {
    path: 'workloads/jobs',
    name: 'JobList',
    component: () => import('@/views/workload/JobList.vue'),
    meta: { titleKey: 'sidebar.jobs', icon: 'Files' },
  },
  {
    path: 'workloads/jobs/create',
    name: 'JobCreate',
    component: () => import('@/views/workload/JobCreate.vue'),
    meta: { titleKey: 'workload.jobCreate', parent: 'JobList' },
  },
  {
    path: 'workloads/jobs/:namespace/:name',
    name: 'JobDetail',
    component: () => import('@/views/workload/JobDetail.vue'),
    props: true,
    meta: { titleKey: 'workload.jobDetail', parent: 'JobList' },
  },
  // CronJob
  {
    path: 'workloads/cronjobs',
    name: 'CronJobList',
    component: () => import('@/views/workload/CronJobList.vue'),
    meta: { titleKey: 'sidebar.cronjobs', icon: 'Files' },
  },
  {
    path: 'workloads/cronjobs/create',
    name: 'CronJobCreate',
    component: () => import('@/views/workload/CronJobCreate.vue'),
    meta: { titleKey: 'workload.cronJobCreate', parent: 'CronJobList' },
  },
  {
    path: 'workloads/cronjobs/:namespace/:name',
    name: 'CronJobDetail',
    component: () => import('@/views/workload/CronJobDetail.vue'),
    props: true,
    meta: { titleKey: 'workload.cronJobDetail', parent: 'CronJobList' },
  },
  // ReplicaSet
  {
    path: 'workloads/replicasets',
    name: 'ReplicaSetList',
    component: () => import('@/views/workload/ReplicaSetList.vue'),
    meta: { titleKey: 'sidebar.replicasets', icon: 'CopyDocument' },
  },
  {
    path: 'workloads/replicasets/:namespace/:name',
    name: 'ReplicaSetDetail',
    component: () => import('@/views/workload/ReplicaSetDetail.vue'),
    props: true,
    meta: { titleKey: 'workload.replicasetDetail', parent: 'ReplicaSetList' },
  },
  // HPA legacy redirects
  {
    path: 'workloads/hpa',
    redirect: '/autoscaling/hpa',
  },
  {
    path: 'workloads/hpa/create',
    redirect: '/autoscaling/hpa/create',
  },
  {
    path: 'workloads/hpa/:namespace/:name',
    redirect: to => `/autoscaling/hpa/${to.params.namespace}/${to.params.name}`,
  },
  // Autoscaling
  {
    path: 'autoscaling',
    name: 'Autoscaling',
    component: () => import('@/views/autoscaling/AutoscalingView.vue'),
    redirect: '/autoscaling/hpa',
    meta: { titleKey: 'sidebar.autoscaling', icon: 'DataLine' },
    children: [
      {
        path: 'hpa',
        name: 'AutoscalingHPAList',
        component: () => import('@/views/workload/hpa/HPAList.vue'),
        meta: { titleKey: 'sidebar.hpa', parent: 'Autoscaling' },
      },
      {
        path: 'hpa/create',
        name: 'AutoscalingHPACreate',
        component: () => import('@/views/workload/hpa/HPACreate.vue'),
        meta: { titleKey: 'workload.hpaCreate', parent: 'Autoscaling' },
      },
      {
        path: 'hpa/:namespace/:name',
        name: 'AutoscalingHPADetail',
        component: () => import('@/views/workload/hpa/HPADetail.vue'),
        props: true,
        meta: { titleKey: 'workload.hpaDetail', parent: 'Autoscaling' },
      },
    ],
  },
]
