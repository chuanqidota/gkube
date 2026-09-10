import type { RouteRecordRaw } from 'vue-router'

export const clusterRoutes: RouteRecordRaw[] = [
  {
    path: 'clusters',
    name: 'ClusterList',
    component: () => import('@/views/cluster/ClusterList.vue'),
    meta: { titleKey: 'sidebar.clusters', icon: 'Connection' },
  },
  {
    path: 'clusters/create',
    name: 'ClusterCreate',
    component: () => import('@/views/cluster/ClusterCreate.vue'),
    meta: { titleKey: 'cluster.createCluster', parent: 'ClusterList', requireAdmin: true },
  },
]
