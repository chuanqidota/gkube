import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { getToken, removeToken, getRefreshToken, setToken, setRefreshToken } from '@/utils/auth'
import { useClusterStore } from '@/stores/cluster'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// 后端业务成功状态码（与后端契约对齐）
const SUCCESS_CODE = 200

// 判断一个值是否为“普通对象”，避免对 FormData/Blob/ArrayBuffer/字符串等写入属性时抛错
function isPlainObject(val: unknown): val is Record<string, any> {
  if (val === null || typeof val !== 'object') return false
  if (val instanceof FormData || val instanceof Blob || val instanceof ArrayBuffer) return false
  const proto = Object.getPrototypeOf(val)
  return proto === Object.prototype || proto === null
}

// Request interceptor: attach Bearer token and cluster name
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // Inject cluster name from store for all K8s API requests
    try {
      const clusterStore = useClusterStore()
      // 用 store 的 clusterName computed(带 clusterName||cluster_name||name fallback),
      // 与各视图取集群名的逻辑一致,避免 localStorage 旧形态对象时静默不注入
      const clusterName = clusterStore.clusterName
      if (clusterName && config.url?.startsWith('/k8s/')) {
        if (!config.params) config.params = {}
        if (!config.params.clusterName && !(isPlainObject(config.data) && config.data.clusterName)) {
          config.params.clusterName = clusterName
        }
        // POST/PUT/DELETE requests: also inject into body so ShouldBindJSON can read it.
        // 仅当 body 是普通对象时写入，避免对 FormData/Blob/字符串等抛错。
        if (config.method !== 'get' && isPlainObject(config.data) && !config.data.clusterName) {
          config.data.clusterName = clusterName
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
// 排队请求：成功时回放，刷新失败时统一 reject，避免悬挂 Promise 导致内存泄漏
let pendingRequests: Array<{ resolve: (token: string) => void; reject: (err: Error) => void }> = []

request.interceptors.response.use(
  (response: AxiosResponse) => {
    const data = response.data
    // 正向判断业务成功：code === 200 视为成功，否则 reject（与后端契约对齐）
    if (data && typeof data === 'object' && 'code' in data && data.code !== SUCCESS_CODE) {
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
        return new Promise((resolve, reject) => {
          pendingRequests.push({
            resolve: (token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`
              resolve(request(originalRequest))
            },
            reject: (err: Error) => reject(err),
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
        const res = await refreshClient.post('/auth/refresh', { refreshToken })
        const data = res.data as any
        if (data?.code !== 200) throw new Error(data?.msg || '刷新Token失败')

        const newToken = data?.accessToken
        const newRefreshToken = data?.refreshToken

        if (newToken) {
          setToken(newToken)
          if (newRefreshToken) {
            setRefreshToken(newRefreshToken)
          }

          // Retry original request
          originalRequest.headers.Authorization = `Bearer ${newToken}`

          // Retry all pending requests
          const queue = pendingRequests
          pendingRequests = []
          queue.forEach((cb) => cb.resolve(newToken))

          return request(originalRequest)
        } else {
          throw new Error('刷新Token响应格式异常')
        }
      } catch {
        // Refresh failed: clear tokens, reject 排队中的请求，再由调用方/路由守卫处理跳转。
        // 不再返回永不 resolve 的 Promise（避免内存泄漏）。
        removeToken()
        const failure = new Error('登录已过期')
        const queue = pendingRequests
        pendingRequests = []
        queue.forEach((cb) => cb.reject(failure))

        const { pathname, search } = window.location
        const current = pathname + search
        const target = current && current !== '/login'
          ? `/login?redirect=${encodeURIComponent(current)}`
          : '/login'
        if (pathname !== '/login') {
          window.location.assign(target)
        }
        return Promise.reject(failure)
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
