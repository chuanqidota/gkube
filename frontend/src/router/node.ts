import type { RouteRecordRaw } from 'vue-router'

export const nodeRoutes: RouteRecordRaw[] = [
  {
    path: 'nodes',
    name: 'NodeList',
    component: () => import('@/views/node/NodeList.vue'),
    meta: { titleKey: 'sidebar.nodes', icon: 'Cpu' },
  },
  {
    path: 'nodes/:name',
    name: 'NodeDetail',
    component: () => import('@/views/node/NodeDetail.vue'),
    props: true,
    meta: { titleKey: 'node.nodeDetail', parent: 'NodeList' },
  },
]
