import type { RouteRecordRaw } from 'vue-router'

export const systemRoutes: RouteRecordRaw[] = [
  // Cluster management
  {
    path: 'clusters',
    name: 'ClusterList',
    component: () => import('@/views/cluster/ClusterList.vue'),
    meta: { title: '集群管理', icon: 'Connection' },
  },
  {
    path: 'clusters/create',
    name: 'ClusterCreate',
    component: () => import('@/views/cluster/ClusterCreate.vue'),
    meta: { title: '创建集群', parent: 'ClusterList', requireAdmin: true },
  },
  // Namespaces
  {
    path: 'namespaces',
    name: 'NamespaceList',
    component: () => import('@/views/namespace/NamespaceList.vue'),
    meta: { title: '命名空间', icon: 'FolderOpened' },
  },
  {
    path: 'namespaces/:name',
    name: 'NamespaceDetail',
    component: () => import('@/views/namespace/NamespaceDetail.vue'),
    props: true,
    meta: { title: '命名空间详情', parent: 'NamespaceList' },
  },
  // Events
  {
    path: 'events',
    name: 'EventList',
    component: () => import('@/views/event/EventList.vue'),
    meta: { title: '事件', icon: 'Bell' },
  },
  // CRD
  {
    path: 'crd',
    name: 'CRDList',
    component: () => import('@/views/crd/CRDList.vue'),
    meta: { title: 'CRD', icon: 'Grid' },
  },
  {
    path: 'crd/create',
    name: 'CRDCreate',
    component: () => import('@/views/crd/CRDCreate.vue'),
    meta: { title: '创建CRD', parent: 'CRDList' },
  },
  {
    path: 'crd/resources',
    name: 'CustomResourceList',
    component: () => import('@/views/crd/CustomResourceList.vue'),
    meta: { title: '自定义资源', parent: 'CRDList' },
  },
  {
    path: 'crd/detail',
    name: 'CRDDetail',
    component: () => import('@/views/crd/CRDDetail.vue'),
    meta: { title: 'CRD详情', parent: 'CRDList' },
  },
  {
    path: 'crd/resources/create',
    name: 'CustomResourceCreate',
    component: () => import('@/views/crd/CustomResourceCreate.vue'),
    meta: { title: '创建自定义资源', parent: 'CustomResourceList' },
  },
  {
    path: 'crd/resources/detail',
    name: 'CustomResourceDetail',
    component: () => import('@/views/crd/CustomResourceDetail.vue'),
    meta: { title: '自定义资源详情', parent: 'CustomResourceList' },
  },
  // Users
  {
    path: 'users',
    name: 'UserList',
    component: () => import('@/views/system/UserList.vue'),
    meta: { title: '用户管理', icon: 'User', requireAdmin: true },
  },
  // RBAC
  {
    path: 'roles',
    name: 'RoleManagement',
    component: () => import('@/views/rbac/RoleManagement.vue'),
    meta: { title: '角色管理', icon: 'Avatar', requireAdmin: true },
  },
  {
    path: 'my-permissions',
    name: 'MyPermissions',
    component: () => import('@/views/rbac/MyPermissions.vue'),
    meta: { title: '我的权限', icon: 'Avatar' },
  },
  // Audit
  {
    path: 'audit',
    name: 'AuditLog',
    component: () => import('@/views/audit/AuditLog.vue'),
    meta: { title: '审计日志', icon: 'Document', requireAdmin: true },
  },
]
