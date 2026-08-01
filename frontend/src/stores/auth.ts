import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi } from '@/api/auth'
import { getToken, setToken, removeToken, setRefreshToken } from '@/utils/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<any>(null)

  async function login(form: { username: string; password: string }) {
    const res: any = await loginApi(form)
    token.value = res.data.accessToken
    setToken(res.data.accessToken)
    setRefreshToken(res.data.refreshToken)
    user.value = res.data.user
    return res
  }

  function logout() {
    token.value = null
    user.value = null
    removeToken()
  }

  return { token, user, login, logout }
})
