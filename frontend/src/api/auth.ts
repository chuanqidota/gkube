import request from '@/api/request'

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data)
}

export function getMe() {
  return request.get('/auth/me')
}

export function refreshToken(data: { refreshToken: string }) {
  return request.post('/auth/refresh', data)
}

export function getOidcLoginUrl() {
  return request.get('/auth/oidc/login')
}

export function handleOidcCallback(code: string, state: string) {
  return request.get('/auth/oidc/callback', { params: { code, state } })
}
