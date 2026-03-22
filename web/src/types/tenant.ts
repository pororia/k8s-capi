export type TenantStatus = 'active' | 'suspended' | 'deleted'

export interface Tenant {
  id: string
  name: string
  description?: string
  os_auth_url: string
  os_username?: string
  os_project_id: string
  os_project_name: string
  os_domain_name: string
  max_clusters: number
  max_nodes_per_cluster: number
  status: TenantStatus
  created_at: string
  updated_at: string
}

export interface CreateTenantRequest {
  name: string
  description?: string
  os_auth_url: string
  os_project_id: string
  os_project_name: string
  os_domain_name?: string
  credentials: {
    username: string
    password: string
    user_domain_name: string
  }
  max_clusters?: number
  max_nodes_per_cluster?: number
}
