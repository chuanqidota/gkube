import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { setToken, setRefreshToken, removeToken, getToken } from '@/utils/auth'
import { login as apiLogin } from '@/api/auth'
import request from '@/api/request'

interface PermissionBinding {
  clusterId: number
  namespace: string        // "" = 集群级
  roleName: string         // "cluster-admin" | "ns-editor" | ...
}

interface UserInfo {
  id?: number | string
  username: string
  email?: string
  display_name?: string
  isAdmin?: boolean
  isSuperAdmin?: boolean
  permissions?: PermissionBinding[]
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
    const payload = res.accessToken ? res : (res.data?.data || res.data || res)
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
  }

  function logout() {
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
      user.value = { ...user.value, permissions: [] }
    }
  }

  function canAccess(clusterId: number, namespace?: string): boolean {
    if (!user.value) return false
    if (user.value.isSuperAdmin) return true
    if (!user.value.permissions) return false

    return user.value.permissions.some(p => {
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

    return user.value.permissions.some(p => {
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

  return { user, token, isLoggedIn, login, logout, setUser, fetchPermissions, canAccess, hasRole }
})
