import type { RouteRecordRaw } from 'vue-router'

export const nodeRoutes: RouteRecordRaw[] = [
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
]
