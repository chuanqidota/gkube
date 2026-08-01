import request from '@/api/request'

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data)
}

export function refreshToken(data: { refreshToken: string }) {
  return request.post('/auth/refresh', data)
}
