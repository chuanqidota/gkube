import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { login as loginApi, logout as logoutApi } from '@/api/auth'
import { getToken, setToken, removeToken, setRefreshToken } from '@/utils/auth'

/** 登录后返回的用户信息（与后端 /auth/login 响应契约对齐） */
export interface UserInfo {
  id?: number | string
  username: string
  isAdmin?: boolean
  roles?: string[]
  permissions?: string[]
  [key: string]: unknown
}

const USER_KEY = 'gkube_user'

function loadUser(): UserInfo | null {
  try {
    const saved = localStorage.getItem(USER_KEY)
    return saved ? (JSON.parse(saved) as UserInfo) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<UserInfo | null>(loadUser())

  // 持久化 user（含 isAdmin），便于路由守卫在刷新后仍可做客户端高危路由过滤
  watch(user, (val) => {
    if (val) {
      localStorage.setItem(USER_KEY, JSON.stringify(val))
    } else {
      localStorage.removeItem(USER_KEY)
    }
  })

  async function login(form: { username: string; password: string }) {
    const res: any = await loginApi(form)
    token.value = res.data.accessToken
    setToken(res.data.accessToken)
    setRefreshToken(res.data.refreshToken)
    // 后端将 isAdmin 放在响应 data 顶层（与 user 平级），合并进 user 以便路由守卫读取
    user.value = { ...(res.data.user as UserInfo), isAdmin: !!res.data.isAdmin }
    return res
  }

  async function logout() {
    // 先清除本地凭证,确保路由守卫立即看到无 token 状态
    token.value = null
    user.value = null
    removeToken()
    // 通知后端注销（若接口未实现则静默失败）
    try {
      await logoutApi()
    } catch {
      // 后端无 /auth/logout 路由时静默忽略
    }
  }

  function setUser(u: UserInfo | null) {
    user.value = u
  }

  return { token, user, login, logout, setUser }
})
