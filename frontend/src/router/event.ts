import type { RouteRecordRaw } from 'vue-router'

export const eventRoutes: RouteRecordRaw[] = [
  {
    path: 'events',
    name: 'EventList',
    component: () => import('@/views/event/EventList.vue'),
    meta: { titleKey: 'sidebar.events', icon: 'Bell' },
  },
]
