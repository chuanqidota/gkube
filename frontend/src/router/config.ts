import type { RouteRecordRaw } from 'vue-router'

export const configRoutes: RouteRecordRaw[] = [
  // ConfigMap
  {
    path: 'config/configmaps',
    name: 'ConfigMapList',
    component: () => import('@/views/config/configmap/ConfigMapList.vue'),
    meta: { titleKey: 'sidebar.configmaps', icon: 'Tickets' },
  },
  {
    path: 'config/configmaps/create',
    name: 'ConfigMapCreate',
    component: () => import('@/views/config/configmap/ConfigMapCreate.vue'),
    meta: { titleKey: 'config.configMapCreate', parent: 'ConfigMapList' },
  },
  {
    path: 'config/configmaps/:namespace/:name',
    name: 'ConfigMapDetail',
    component: () => import('@/views/config/configmap/ConfigMapDetail.vue'),
    props: true,
    meta: { titleKey: 'config.configMapDetail', parent: 'ConfigMapList' },
  },
  // Secret
  {
    path: 'config/secrets',
    name: 'SecretList',
    component: () => import('@/views/config/secret/SecretList.vue'),
    meta: { titleKey: 'sidebar.secrets', icon: 'Key' },
  },
  {
    path: 'config/secrets/create',
    name: 'SecretCreate',
    component: () => import('@/views/config/secret/SecretCreate.vue'),
    meta: { titleKey: 'config.secretCreate', parent: 'SecretList' },
  },
  {
    path: 'config/secrets/:namespace/:name',
    name: 'SecretDetail',
    component: () => import('@/views/config/secret/SecretDetail.vue'),
    props: true,
    meta: { titleKey: 'config.secretDetail', parent: 'SecretList' },
  },
  // ResourceQuota
  {
    path: 'config/resourcequotas',
    name: 'ResourceQuotaList',
    component: () => import('@/views/config/resourcequota/ResourceQuotaList.vue'),
    meta: { titleKey: 'config.resourcequota', icon: 'Coin' },
  },
  {
    path: 'config/resourcequotas/create',
    name: 'ResourceQuotaCreate',
    component: () => import('@/views/config/resourcequota/ResourceQuotaCreate.vue'),
    meta: { titleKey: 'config.resourceQuotaCreate', parent: 'ResourceQuotaList' },
  },
  {
    path: 'config/resourcequotas/:namespace/:name',
    name: 'ResourceQuotaDetail',
    component: () => import('@/views/config/resourcequota/ResourceQuotaDetail.vue'),
    props: true,
    meta: { titleKey: 'config.resourceQuotaDetail', parent: 'ResourceQuotaList' },
  },
  // LimitRange
  {
    path: 'config/limitranges',
    name: 'LimitRangeList',
    component: () => import('@/views/config/limitrange/LimitRangeList.vue'),
    meta: { titleKey: 'config.limitrange', icon: 'ScaleToOriginal' },
  },
  {
    path: 'config/limitranges/create',
    name: 'LimitRangeCreate',
    component: () => import('@/views/config/limitrange/LimitRangeCreate.vue'),
    meta: { titleKey: 'config.limitRangeCreate', parent: 'LimitRangeList' },
  },
  {
    path: 'config/limitranges/:namespace/:name',
    name: 'LimitRangeDetail',
    component: () => import('@/views/config/limitrange/LimitRangeDetail.vue'),
    props: true,
    meta: { titleKey: 'config.limitRangeDetail', parent: 'LimitRangeList' },
  },
]
