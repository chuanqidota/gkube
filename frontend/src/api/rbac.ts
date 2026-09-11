import request from '@/api/request'

// 获取所有角色列表
export const getRoles = () => request.get('/rbac/roles')

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
export const deleteBinding = (id: number) => request.delete(`/rbac/bindings/${id}`)

// 获取当前用户权限
export const getMyPermissions = () => request.get('/rbac/my-permissions')

// 获取集群成员列表
export const getClusterMembers = (clusterId: number) =>
  request.get('/rbac/cluster-members', { params: { clusterId } })

// 获取命名空间列表（K8s 路由，仍用 clusterName）
export const getNamespaceList = (clusterName: string) =>
  request.get('/k8s/namespace/list', { params: { clusterName } })

// 搜索用户（添加绑定时使用）
export const searchUsers = (params: { keyword?: string; page?: number; size?: number }) =>
  request.get('/users', { params })

// --- P1 集合级操作 ---

// 按组批量换角色：把某用户在某集群下 fromRoleId 的所有绑定换成 toRoleId
export const updateBindingBatch = (data: {
  userId: number
  clusterId: number
  fromRoleId: number
  toRoleId: number
}) => request.put('/rbac/bindings/batch', data)

// 按组批量删除：删除某用户在某集群下指定角色的全部绑定
export const deleteBindingBatch = (data: { userId: number; clusterId: number; roleId: number }) =>
  request.delete('/rbac/bindings/batch', { data })

// 移出集群：删除该用户在该集群下所有绑定
export const removeClusterMember = (userId: number, clusterId: number) =>
  request.delete(`/rbac/bindings/user/${userId}`, { params: { clusterId } })

// --- P2 角色管理 ---

// 获取资源组×动词字典（角色矩阵编辑器数据源）
export const getResourceDict = () => request.get('/rbac/resources')

// 创建自定义角色
export const createRole = (data: {
  name: string
  displayName: string
  scopeType: 'cluster' | 'namespace'
  description?: string
  permissions: Record<string, string[]>
}) => request.post('/rbac/roles', data)

// 更新自定义角色（预置角色不可改）
export const updateRole = (
  id: number,
  data: {
    displayName: string
    description?: string
    permissions: Record<string, string[]>
  },
) => request.put(`/rbac/roles/${id}`, data)

// 删除自定义角色
export const deleteRole = (id: number) => request.delete(`/rbac/roles/${id}`)

// 权限诊断
export const canI = (data: {
  userId: number
  clusterId: number
  namespace?: string
  resourceGroup: string
  verb: string
}) => request.post('/rbac/can-i', data)
