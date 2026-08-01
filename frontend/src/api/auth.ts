import request from '@/api/request'

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data)
}

/**
 * 获取一次性 WebSocket/SSE 鉴权 ticket（30s 有效）。
 * 连接 WS/SSE 前先调用，URL 使用 ?ticket=<ticket> 而非直接携带 access token。
 */
export function getWsTicket() {
  return request.post('/auth/ws-ticket')
}

// TODO: 后端注销接口就绪后，logout 应调用此接口使当前 token 失效。
// 目前仅在前端清除 token（见 stores/auth.ts），待后端补充 /auth/logout。
export function logout() {
  return request.post('/auth/logout').catch(() => {
    // 后端尚未实现注销接口时静默失败，前端仍会清除本地凭证
  })
}
