import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { authStorage, isTokenExpired } from './auth'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
})

// 요청 인터셉터: 액세스 토큰 자동 주입
apiClient.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    let token = authStorage.getAccessToken()

    // 토큰 만료 시 갱신
    if (token && isTokenExpired(token)) {
      const refreshToken = authStorage.getRefreshToken()
      if (refreshToken && !isTokenExpired(refreshToken)) {
        try {
          const { data } = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshToken })
          authStorage.setTokens(data.data.access_token, data.data.refresh_token)
          token = data.data.access_token
        } catch {
          authStorage.clearTokens()
          window.location.href = '/login'
          return Promise.reject(new Error('Token refresh failed'))
        }
      } else {
        authStorage.clearTokens()
        window.location.href = '/login'
        return Promise.reject(new Error('Session expired'))
      }
    }

    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }

    // Tenant ID 헤더 (Super Admin용)
    const tenantId = sessionStorage.getItem('selected_tenant_id')
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId
    }

    return config
  },
  (error) => Promise.reject(error)
)

// 응답 인터셉터: 401 처리
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      authStorage.clearTokens()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API 함수들
export const authApi = {
  login: (email: string, password: string) =>
    apiClient.post('/auth/login', { email, password }),
  refresh: (refreshToken: string) =>
    apiClient.post('/auth/refresh', { refresh_token: refreshToken }),
  logout: () => apiClient.post('/auth/logout'),
  me: () => apiClient.get('/auth/me'),
}

export const clusterApi = {
  list: (params?: Record<string, unknown>) => apiClient.get('/clusters', { params }),
  get: (id: string) => apiClient.get(`/clusters/${id}`),
  create: (data: unknown) => apiClient.post('/clusters', data),
  delete: (id: string) => apiClient.delete(`/clusters/${id}`),
  scale: (id: string, data: unknown) => apiClient.post(`/clusters/${id}/scale`, data),
  upgrade: (id: string, data: unknown) => apiClient.post(`/clusters/${id}/upgrade`, data),
  getKubeconfig: (id: string) =>
    apiClient.get(`/clusters/${id}/kubeconfig`, { responseType: 'blob' }),
  getEvents: (id: string, params?: Record<string, unknown>) =>
    apiClient.get(`/clusters/${id}/events`, { params }),
  getNodes: (id: string) => apiClient.get(`/clusters/${id}/nodes`),
}

export const tenantApi = {
  list: (params?: Record<string, unknown>) => apiClient.get('/tenants', { params }),
  get: (id: string) => apiClient.get(`/tenants/${id}`),
  create: (data: unknown) => apiClient.post('/tenants', data),
  update: (id: string, data: unknown) => apiClient.put(`/tenants/${id}`, data),
  delete: (id: string) => apiClient.delete(`/tenants/${id}`),
}

export const openstackApi = {
  getFlavors: () => apiClient.get('/openstack/flavors'),
  getImages: () => apiClient.get('/openstack/images'),
  getNetworks: () => apiClient.get('/openstack/networks'),
  getSubnets: (networkId: string) => apiClient.get(`/openstack/networks/${networkId}/subnets`),
  getSecurityGroups: () => apiClient.get('/openstack/security-groups'),
  getKeypairs: () => apiClient.get('/openstack/keypairs'),
  getExternalNetworks: () => apiClient.get('/openstack/external-networks'),
  getQuotas: () => apiClient.get('/openstack/quotas'),
}

export const userApi = {
  list: (params?: Record<string, unknown>) => apiClient.get('/users', { params }),
  get: (id: string) => apiClient.get(`/users/${id}`),
  create: (data: unknown) => apiClient.post('/users', data),
  update: (id: string, data: unknown) => apiClient.put(`/users/${id}`, data),
  delete: (id: string) => apiClient.delete(`/users/${id}`),
  changePassword: (id: string, data: unknown) => apiClient.put(`/users/${id}/password`, data),
  changeRole: (id: string, data: unknown) => apiClient.put(`/users/${id}/role`, data),
}

export const auditApi = {
  list: (params?: Record<string, unknown>) => apiClient.get('/audit-logs', { params }),
  get: (id: string) => apiClient.get(`/audit-logs/${id}`),
}
