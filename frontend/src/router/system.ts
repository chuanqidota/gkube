import type { RouteRecordRaw } from 'vue-router'

export const systemRoutes: RouteRecordRaw[] = [
  // Namespaces
  {
    path: 'namespaces',
    name: 'NamespaceList',
    component: () => import('@/views/namespace/NamespaceList.vue'),
    meta: { titleKey: 'sidebar.namespaces', icon: 'FolderOpened' },
  },
  {
    path: 'namespaces/:name',
    name: 'NamespaceDetail',
    component: () => import('@/views/namespace/NamespaceDetail.vue'),
    props: true,
    meta: { titleKey: 'namespace.namespaceDetail', parent: 'NamespaceList' },
  },
  // CRD
  {
    path: 'crd',
    name: 'CRDList',
    component: () => import('@/views/crd/CRDList.vue'),
    meta: { titleKey: 'sidebar.crd', icon: 'Grid' },
  },
  {
    path: 'crd/create',
    name: 'CRDCreate',
    component: () => import('@/views/crd/CRDCreate.vue'),
    meta: { titleKey: 'crd.crdCreate', parent: 'CRDList' },
  },
  {
    path: 'crd/resources',
    name: 'CustomResourceList',
    component: () => import('@/views/crd/CustomResourceList.vue'),
    meta: { titleKey: 'crd.customResources', parent: 'CRDList' },
  },
  {
    path: 'crd/detail',
    name: 'CRDDetail',
    component: () => import('@/views/crd/CRDDetail.vue'),
    meta: { titleKey: 'crd.crdDetail', parent: 'CRDList' },
  },
  {
    path: 'crd/resources/create',
    name: 'CustomResourceCreate',
    component: () => import('@/views/crd/CustomResourceCreate.vue'),
    meta: { titleKey: 'crd.customResourceCreate', parent: 'CustomResourceList' },
  },
  {
    path: 'crd/resources/detail',
    name: 'CustomResourceDetail',
    component: () => import('@/views/crd/CustomResourceDetail.vue'),
    meta: { titleKey: 'crd.customResourceDetail', parent: 'CustomResourceList' },
  },
  // Users
  {
    path: 'users',
    name: 'UserList',
    component: () => import('@/views/system/UserList.vue'),
    meta: { titleKey: 'sidebar.users', icon: 'User', requireAdmin: true },
  },
  // RBAC
  {
    path: 'roles',
    name: 'RoleManagement',
    component: () => import('@/views/rbac/RoleManagement.vue'),
    meta: { titleKey: 'sidebar.roles', icon: 'Avatar', requireAdmin: true },
  },
  {
    path: 'my-permissions',
    name: 'MyPermissions',
    component: () => import('@/views/rbac/MyPermissions.vue'),
    meta: { titleKey: 'rbac.myPermissions', icon: 'Avatar' },
  },
  // Audit
  {
    path: 'audit',
    name: 'AuditLog',
    component: () => import('@/views/audit/AuditLog.vue'),
    meta: { titleKey: 'sidebar.audit', icon: 'Document', requireAdmin: true },
  },
]
