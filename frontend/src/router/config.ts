import type { RouteRecordRaw } from 'vue-router'

export const configRoutes: RouteRecordRaw[] = [
  // ConfigMap
  {
    path: 'config/configmaps',
    name: 'ConfigMapList',
    component: () => import('@/views/config/configmap/ConfigMapList.vue'),
    meta: { title: 'ConfigMap', icon: 'Tickets' },
  },
  {
    path: 'config/configmaps/create',
    name: 'ConfigMapCreate',
    component: () => import('@/views/config/configmap/ConfigMapCreate.vue'),
    meta: { title: '创建ConfigMap', parent: 'ConfigMapList' },
  },
  {
    path: 'config/configmaps/:namespace/:name',
    name: 'ConfigMapDetail',
    component: () => import('@/views/config/configmap/ConfigMapDetail.vue'),
    props: true,
    meta: { title: 'ConfigMap详情', parent: 'ConfigMapList' },
  },
  // Secret
  {
    path: 'config/secrets',
    name: 'SecretList',
    component: () => import('@/views/config/secret/SecretList.vue'),
    meta: { title: 'Secret', icon: 'Key' },
  },
  {
    path: 'config/secrets/create',
    name: 'SecretCreate',
    component: () => import('@/views/config/secret/SecretCreate.vue'),
    meta: { title: '创建Secret', parent: 'SecretList' },
  },
  {
    path: 'config/secrets/:namespace/:name',
    name: 'SecretDetail',
    component: () => import('@/views/config/secret/SecretDetail.vue'),
    props: true,
    meta: { title: 'Secret详情', parent: 'SecretList' },
  },
  // ResourceQuota
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
    component: () => import('@/views/config/resourcequota/ResourceQuotaDetail.vue'),
    props: true,
    meta: { title: 'ResourceQuota详情', parent: 'ResourceQuotaList' },
  },
  // LimitRange
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
    props: true,
    meta: { title: 'LimitRange详情', parent: 'LimitRangeList' },
  },
]
