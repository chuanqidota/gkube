import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { getToken, removeToken, getRefreshToken, setToken, setRefreshToken } from '@/utils/auth'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// Request interceptor: attach Bearer token and cluster name
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // Inject cluster name from localStorage for all K8s API requests
    try {
      const saved = localStorage.getItem('gkube_cluster')
      if (saved) {
        const cluster = JSON.parse(saved)
        const clusterName = cluster?.clusterName
        if (clusterName && config.url?.startsWith('/k8s/')) {
          if (!config.params) config.params = {}
          if (!config.params.clusterName && !config.data?.clusterName) {
            config.params.clusterName = clusterName
          }
        }
      }
    } catch {
      // ignore
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
    // 后端返回 code=0 表示业务失败，需要 reject
    const data = response.data
    if (data && data.code === 0) {
      return Promise.reject(new Error(data.msg || '请求失败'))
    }
    // 解包后端响应：将 { code, msg, data } 中的 data 提升到 response.data
    // 使得调用方可以直接通过 res.data 访问业务数据
    if (data && data.data !== undefined) {
      response.data = data.data
    }
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

        // 使用独立的 axios 实例发送刷新请求，避免携带过期的 Authorization header
        const refreshClient = axios.create({ baseURL: '/api/v1', timeout: 15000 })
        const { data } = await refreshClient.post('/auth/refresh', {
          refreshToken: refreshToken,
        })

        const newToken = data.data?.accessToken
        const newRefreshToken = data.data?.refreshToken

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
        } else {
          throw new Error('刷新Token响应格式异常')
        }
      } catch {
        // Refresh failed: clear tokens and redirect to login
        removeToken()
        pendingRequests = []
        window.location.href = '/login'
        // 返回一个永不 resolve 的 Promise，阻止原始请求继续抛出错误
        return new Promise(() => {})
      } finally {
        isRefreshing = false
      }
    }

    // 对于非 401 的 HTTP 错误，提取后端返回的错误信息
    if (error.response?.data?.msg) {
      return Promise.reject(new Error(error.response.data.msg))
    }
    return Promise.reject(error)
  }
)

export default request
