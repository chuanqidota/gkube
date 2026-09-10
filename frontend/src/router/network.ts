import type { RouteRecordRaw } from 'vue-router'

export const networkRoutes: RouteRecordRaw[] = [
  // Service
  {
    path: 'network/services',
    name: 'ServiceList',
    component: () => import('@/views/network/ServiceList.vue'),
    meta: { titleKey: 'sidebar.services', icon: 'Connection' },
  },
  {
    path: 'network/services/create',
    name: 'ServiceCreate',
    component: () => import('@/views/network/ServiceCreate.vue'),
    meta: { titleKey: 'network.serviceCreate', parent: 'ServiceList' },
  },
  {
    path: 'network/services/:namespace/:name',
    name: 'ServiceDetail',
    component: () => import('@/views/network/ServiceDetail.vue'),
    props: true,
    meta: { titleKey: 'network.serviceDetail', parent: 'ServiceList' },
  },
  // Ingress
  {
    path: 'network/ingresses',
    name: 'IngressList',
    component: () => import('@/views/network/IngressList.vue'),
    meta: { titleKey: 'sidebar.ingresses', icon: 'Link' },
  },
  {
    path: 'network/ingresses/create',
    name: 'IngressCreate',
    component: () => import('@/views/network/IngressCreate.vue'),
    meta: { titleKey: 'network.ingressCreate', parent: 'IngressList' },
  },
  {
    path: 'network/ingresses/:namespace/:name',
    name: 'IngressDetail',
    component: () => import('@/views/network/IngressDetail.vue'),
    props: true,
    meta: { titleKey: 'network.ingressDetail', parent: 'IngressList' },
  },
  // NetworkPolicy
  {
    path: 'network/networkpolicies',
    name: 'NetworkPolicyList',
    component: () => import('@/views/network/networkpolicy/NetworkPolicyList.vue'),
    meta: { titleKey: 'sidebar.networkpolicies', icon: 'Lock' },
  },
  {
    path: 'network/networkpolicies/create',
    name: 'NetworkPolicyCreate',
    component: () => import('@/views/network/networkpolicy/NetworkPolicyCreate.vue'),
    meta: { titleKey: 'network.networkPolicyCreate', parent: 'NetworkPolicyList' },
  },
  {
    path: 'network/networkpolicies/:namespace/:name',
    name: 'NetworkPolicyDetail',
    component: () => import('@/views/network/networkpolicy/NetworkPolicyDetail.vue'),
    props: true,
    meta: { titleKey: 'network.networkPolicyDetail', parent: 'NetworkPolicyList' },
  },
]
