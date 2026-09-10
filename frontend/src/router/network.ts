import type { RouteRecordRaw } from 'vue-router'

export const networkRoutes: RouteRecordRaw[] = [
  // Service
  {
    path: 'network/services',
    name: 'ServiceList',
    component: () => import('@/views/network/ServiceList.vue'),
    meta: { title: 'Service', icon: 'Connection' },
  },
  {
    path: 'network/services/create',
    name: 'ServiceCreate',
    component: () => import('@/views/network/ServiceCreate.vue'),
    meta: { title: '创建Service', parent: 'ServiceList' },
  },
  {
    path: 'network/services/:namespace/:name',
    name: 'ServiceDetail',
    component: () => import('@/views/network/ServiceDetail.vue'),
    props: true,
    meta: { title: 'Service详情', parent: 'ServiceList' },
  },
  // Ingress
  {
    path: 'network/ingresses',
    name: 'IngressList',
    component: () => import('@/views/network/IngressList.vue'),
    meta: { title: 'Ingress', icon: 'Link' },
  },
  {
    path: 'network/ingresses/create',
    name: 'IngressCreate',
    component: () => import('@/views/network/IngressCreate.vue'),
    meta: { title: '创建Ingress', parent: 'IngressList' },
  },
  {
    path: 'network/ingresses/:namespace/:name',
    name: 'IngressDetail',
    component: () => import('@/views/network/IngressDetail.vue'),
    props: true,
    meta: { title: 'Ingress详情', parent: 'IngressList' },
  },
  // NetworkPolicy
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
]
