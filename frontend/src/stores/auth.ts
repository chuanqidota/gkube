import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, getMe } from '@/api/auth'
import { getToken, setToken, removeToken, setRefreshToken } from '@/utils/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<any>(null)
  const roles = ref<string[]>([])

  async function login(form: { username: string; password: string }) {
    const res: any = await loginApi(form)
    token.value = res.data.accessToken
    setToken(res.data.accessToken)
    setRefreshToken(res.data.refreshToken)
    user.value = res.data.user
    return res
  }

  async function fetchUserInfo() {
    const res: any = await getMe()
    user.value = res.data
    roles.value = res.data.roles?.map((r: any) => r.name) || []
    return res
  }

  function logout() {
    token.value = null
    user.value = null
    roles.value = []
    removeToken()
  }

  const isSuperAdmin = () => roles.value.includes('super_admin')

  return { token, user, roles, login, fetchUserInfo, logout, isSuperAdmin }
})
