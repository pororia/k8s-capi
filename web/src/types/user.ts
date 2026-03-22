export type UserRole = 'super_admin' | 'tenant_admin' | 'operator' | 'viewer'
export type UserStatus = 'active' | 'inactive' | 'suspended'

export interface User {
  id: string
  tenant_id?: string
  email: string
  name: string
  role: UserRole
  status: UserStatus
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: User
}

export interface Claims {
  sub: string
  tenant_id?: string
  role: UserRole
  exp: number
  iat: number
}
