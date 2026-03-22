export type AuditResult = 'success' | 'failure'

export interface AuditLog {
  id: string
  tenant_id?: string
  user_id?: string
  action: string
  resource_type: string
  resource_id?: string
  resource_name?: string
  request_body?: Record<string, unknown>
  response_status?: number
  result: AuditResult
  error_message?: string
  ip_address?: string
  user_agent?: string
  duration_ms?: number
  created_at: string
}
