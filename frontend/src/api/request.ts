import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { getToken, removeToken, getRefreshToken, setToken, setRefreshToken } from '@/utils/auth'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// Request interceptor: attach Bearer token
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor: handle 401 with silent refresh, then redirect on failure
let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

request.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  async (error) => {
    const originalRequest = error.config

    // If 401 and we haven't already retried this request
    if (error.response?.status === 401 && !originalRequest._retry) {
      // If already refreshing, queue this request
      if (isRefreshing) {
        return new Promise((resolve) => {
          pendingRequests.push((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(request(originalRequest))
          })
        })
      }

      isRefreshing = true
      originalRequest._retry = true

      try {
        const refreshToken = getRefreshToken()
        if (!refreshToken) {
          throw new Error('No refresh token')
        }

        const { data } = await axios.post('/api/v1/auth/refresh', {
          refresh_token: refreshToken,
        })

        const newToken = data.data?.access_token
        const newRefreshToken = data.data?.refresh_token

        if (newToken) {
          setToken(newToken)
          if (newRefreshToken) {
            setRefreshToken(newRefreshToken)
          }

          // Retry original request
          originalRequest.headers.Authorization = `Bearer ${newToken}`

          // Retry all pending requests
          pendingRequests.forEach((cb) => cb(newToken))
          pendingRequests = []

          return request(originalRequest)
        }
      } catch {
        // Refresh failed: clear tokens and redirect to login
        removeToken()
        window.location.href = '/login'
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

export default request
