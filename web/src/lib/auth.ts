const ACCESS_TOKEN_KEY = 'cmp_access_token'
const REFRESH_TOKEN_KEY = 'cmp_refresh_token'

export const authStorage = {
  getAccessToken: (): string | null => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(ACCESS_TOKEN_KEY)
  },
  getRefreshToken: (): string | null => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  },
  setTokens: (accessToken: string, refreshToken: string) => {
    localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  },
  clearTokens: () => {
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  },
}

export function parseJwt(token: string): Record<string, unknown> | null {
  try {
    const base64 = token.split('.')[1]
    const json = atob(base64.replace(/-/g, '+').replace(/_/g, '/'))
    return JSON.parse(json)
  } catch {
    return null
  }
}

export function isTokenExpired(token: string): boolean {
  const payload = parseJwt(token)
  if (!payload?.exp) return true
  return Date.now() >= (payload.exp as number) * 1000
}
