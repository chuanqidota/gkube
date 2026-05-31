const TOKEN_KEY = 'gkube_access_token'
const REFRESH_KEY = 'gkube_refresh_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

export function setRefreshToken(token: string): void {
  localStorage.setItem(REFRESH_KEY, token)
}
