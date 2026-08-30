import request from '@/api/request'

// 获取所有角色列表
export const getRoles = () =>
  request.get('/rbac/roles')

// 查询权限绑定列表（管理员）
export const getBindings = (params: {
  clusterId?: number
  userId?: number
  page?: number
  size?: number
}) => request.get('/rbac/bindings', { params })

// 创建权限绑定（支持多命名空间）
export const createBinding = (data: {
  userId: number
  roleId: number
  clusterId: number
  namespace?: string
  namespaces?: string[]
}) => request.post('/rbac/bindings', data)

// 修改权限绑定（只能改角色）
export const updateBinding = (id: number, data: { roleId: number }) =>
  request.put(`/rbac/bindings/${id}`, data)

// 删除权限绑定
export const deleteBinding = (id: number) =>
  request.delete(`/rbac/bindings/${id}`)

// 获取当前用户权限
export const getMyPermissions = () =>
  request.get('/rbac/my-permissions')

// 获取集群成员列表
export const getClusterMembers = (clusterId: number) =>
  request.get('/rbac/cluster-members', { params: { clusterId } })

// 获取命名空间列表（K8s 路由，仍用 clusterName）
export const getNamespaceList = (clusterName: string) =>
  request.get('/k8s/namespace/list', { params: { clusterName } })

// 搜索用户（添加绑定时使用）
export const searchUsers = (params: { keyword?: string; page?: number; size?: number }) =>
  request.get('/users', { params })
