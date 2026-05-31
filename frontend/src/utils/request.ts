import axios from 'axios'
import { getToken, getRefreshToken, setToken, removeToken } from '@/utils/auth'

const request = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

request.interceptors.response.use(
  (response) => response.data,
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve) => {
          pendingRequests.push((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(request(originalRequest))
          })
        })
      }
      originalRequest._retry = true
      isRefreshing = true
      try {
        const refreshToken = getRefreshToken()
        if (!refreshToken) {
          throw new Error('No refresh token')
        }
        const res: any = await axios.post('/api/auth/refresh', { refreshToken })
        const newToken = res.data.accessToken
        setToken(newToken)
        pendingRequests.forEach((cb) => cb(newToken))
        pendingRequests = []
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return request(originalRequest)
      } catch {
        removeToken()
        window.location.href = '/login'
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }
    return Promise.reject(error)
  }
)

export default request
