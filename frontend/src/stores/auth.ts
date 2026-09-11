import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { setToken, setRefreshToken, removeToken, getToken } from '@/utils/auth'
import { login as apiLogin, logout as apiLogout } from '@/api/auth'
import request from '@/api/request'
import { getRoles } from '@/api/rbac'

interface PermissionBinding {
  clusterId: number
  namespace: string // "" = 集群级
  roleName: string // "cluster-admin" | "ns-editor" | ...
}

interface UserInfo {
  id?: number | string
  username: string
  email?: string
  display_name?: string
  isAdmin?: boolean
  isSuperAdmin?: boolean
  permissions?: PermissionBinding[] | null
  [key: string]: unknown
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const token = ref<string | null>(getToken())

  const isLoggedIn = computed(() => !!token.value)

  // 从 localStorage 恢复用户信息
  const savedUser = localStorage.getItem('gkube_user')
  if (savedUser) {
    try {
      user.value = JSON.parse(savedUser)
    } catch {
      // ignore
    }
  }

  async function login(data: { username: string; password: string }) {
    const res: any = await apiLogin(data)
    const payload = res.accessToken ? res : res.data?.data || res.data || res
    token.value = payload.accessToken
    setToken(payload.accessToken)
    setRefreshToken(payload.refreshToken)
    const newUser: UserInfo = {
      ...payload.user,
      isAdmin: payload.isAdmin,
      isSuperAdmin: payload.isSuperAdmin,
    }
    user.value = newUser
    localStorage.setItem('gkube_user', JSON.stringify(newUser))
    // 登录后加载角色权限定义（不阻塞登录流程）
    loadRoles()
  }

  async function logout() {
    // 先调后端注销（需要 token 认证），再清本地状态
    try {
      await apiLogout()
    } catch {
      /* 后端可能尚未实现 /auth/logout */
    }
    user.value = null
    token.value = null
    removeToken()
    localStorage.removeItem('gkube_user')
  }

  function setUser(newUser: UserInfo) {
    user.value = newUser
    localStorage.setItem('gkube_user', JSON.stringify(newUser))
  }

  async function fetchPermissions(): Promise<void> {
    if (!user.value) return
    try {
      const res: any = await request.get('/rbac/my-permissions')
      const payload = res?.data ?? res
      if (payload) {
        user.value = {
          ...user.value,
          isSuperAdmin: payload.isSuperAdmin,
          permissions: payload.bindings || [],
        }
        localStorage.setItem('gkube_user', JSON.stringify(user.value))
      }
    } catch {
      user.value = { ...user.value, permissions: null }
    }
  }

  // 角色权限定义缓存（登录后加载一次，进程生命周期内有效）
  const rolePerms = ref<Map<string, Record<string, string[]>>>(new Map())

  /**
   * 加载角色权限定义。登录后调用一次，将 GET /rbac/roles 的结果
   * 缓存到 rolePerms 中，供 canDo() 判断使用。
   */
  async function loadRoles(): Promise<void> {
    try {
      const res: any = await getRoles()
      const data = res?.data ?? res
      const roles = Array.isArray(data) ? data : data?.items || []
      const map = new Map<string, Record<string, string[]>>()
      for (const r of roles) {
        map.set(r.name, r.permissions || {})
      }
      rolePerms.value = map
    } catch {
      // 加载失败不阻塞页面，canDo 会返回 false（安全降级）
    }
  }

  /**
   * 判断当前用户是否可以对指定资源执行指定操作。
   * 判断逻辑与后端 RequirePermission 中间件一致：
   *   1. 超级管理员直接放行
   *   2. 集群级绑定（namespace=""）覆盖该集群所有命名空间
   *   3. 命名空间级绑定精确匹配
   *   4. 查角色 permissions JSON 中 resourceGroup 是否包含 verb
   */
  function canDo(
    clusterId: number,
    resourceGroup: string,
    verb: string,
    namespace?: string,
  ): boolean {
    if (!user.value) return false
    if (user.value.isSuperAdmin) return true
    if (!user.value.permissions) return false

    return user.value.permissions.some((p) => {
      if (p.clusterId !== clusterId) return false
      // 集群级绑定覆盖所有 namespace
      if (p.namespace === '' || (namespace && p.namespace === namespace)) {
        const perms = rolePerms.value.get(p.roleName)
        return perms?.[resourceGroup]?.includes(verb) ?? false
      }
      return false
    })
  }

  // 判断当前用户在 (clusterId, namespace) 是否有写权限（editor/admin 级）。
  // 集群级绑定（namespace=""）覆盖该集群所有 ns；角色名含 admin/editor 视为可写。
  function canWrite(clusterId: number, namespace?: string): boolean {
    if (!user.value) return false
    if (user.value.isSuperAdmin) return true
    if (!user.value.permissions) return false
    const writable = (roleName: string) => /admin|editor/.test(roleName)
    return user.value.permissions.some((p) => {
      if (p.clusterId !== clusterId) return false
      if (p.namespace === '') return writable(p.roleName)
      if (namespace && p.namespace === namespace) return writable(p.roleName)
      return false
    })
  }

  function canAccess(clusterId: number, namespace?: string): boolean {
    if (!user.value) return false
    if (user.value.isSuperAdmin) return true
    if (!user.value.permissions) return false

    return user.value.permissions.some((p) => {
      if (p.clusterId !== clusterId) return false
      if (p.namespace === '') return true
      if (namespace && p.namespace === namespace) return true
      return false
    })
  }

  function hasRole(clusterId: number, namespace?: string, roles?: string[]): boolean {
    if (!user.value) return false
    if (user.value.isSuperAdmin) return true
    if (!user.value.permissions) return false

    return user.value.permissions.some((p) => {
      if (p.clusterId !== clusterId) return false
      if (p.namespace === '') {
        return !roles || roles.includes(p.roleName)
      }
      if (namespace && p.namespace === namespace) {
        return !roles || roles.includes(p.roleName)
      }
      return false
    })
  }

  return {
    user,
    token,
    isLoggedIn,
    login,
    logout,
    setUser,
    fetchPermissions,
    loadRoles,
    canAccess,
    hasRole,
    canDo,
    canWrite,
  }
})
