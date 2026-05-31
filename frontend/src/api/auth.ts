import request from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  access_token: string
  refresh_token: string
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  roles: string[]
}

export function login(data: LoginParams) {
  return request.post<LoginResult>('/auth/login', data)
}

export function refreshToken(data: { refresh_token: string }) {
  return request.post<LoginResult>('/auth/refresh', data)
}

export function getMe() {
  return request.get<UserInfo>('/auth/me')
}
