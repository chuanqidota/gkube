import type { RouteRecordRaw } from 'vue-router'

export const workloadRoutes: RouteRecordRaw[] = [
  // Pod
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
    meta: { title: 'Pod详情', parent: 'PodList' },
  },
  // Deployment
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
  // StatefulSet
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
  // DaemonSet
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
  // Job
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
  // CronJob
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
  // ReplicaSet
  {
    path: 'workloads/replicasets',
    name: 'ReplicaSetList',
    component: () => import('@/views/workload/ReplicaSetList.vue'),
    meta: { title: 'ReplicaSet', icon: 'CopyDocument' },
  },
  {
    path: 'workloads/replicasets/:namespace/:name',
    name: 'ReplicaSetDetail',
    component: () => import('@/views/workload/ReplicaSetDetail.vue'),
    props: true,
    meta: { title: 'ReplicaSet 详情', parent: 'ReplicaSetList' },
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
    meta: { title: '弹性伸缩', icon: 'DataLine' },
    children: [
      {
        path: 'hpa',
        name: 'AutoscalingHPAList',
        component: () => import('@/views/workload/hpa/HPAList.vue'),
        meta: { title: 'HPA', parent: 'Autoscaling' },
      },
      {
        path: 'hpa/create',
        name: 'AutoscalingHPACreate',
        component: () => import('@/views/workload/hpa/HPACreate.vue'),
        meta: { title: '创建HPA', parent: 'Autoscaling' },
      },
      {
        path: 'hpa/:namespace/:name',
        name: 'AutoscalingHPADetail',
        component: () => import('@/views/workload/hpa/HPADetail.vue'),
        props: true,
        meta: { title: 'HPA详情', parent: 'Autoscaling' },
      },
    ],
  },
]
