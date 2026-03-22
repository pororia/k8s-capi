export type ClusterStatus =
  | 'pending'
  | 'provisioning'
  | 'running'
  | 'updating'
  | 'failed'
  | 'deleting'
  | 'deleted'

export type NodeRole = 'control_plane' | 'worker'
export type NodeStatus = 'provisioning' | 'running' | 'failed' | 'deleting'
export type EventSeverity = 'info' | 'warning' | 'error'

export interface Cluster {
  id: string
  tenant_id: string
  name: string
  kubernetes_version: string
  status: ClusterStatus
  cp_count: number
  worker_count: number
  cp_flavor: string
  worker_flavor: string
  os_image: string
  network_id: string
  subnet_id: string
  external_network_id: string
  ssh_key_name: string
  pod_cidr: string
  service_cidr: string
  cni_plugin: string
  dns_nameservers: string[]
  api_endpoint?: string
  capi_namespace: string
  error_message?: string
  created_by?: string
  created_at: string
  updated_at: string
  nodes?: Node[]
}

export interface Node {
  id: string
  cluster_id: string
  name: string
  role: NodeRole
  status: NodeStatus
  machine_id?: string
  openstack_server_id?: string
  private_ip?: string
  floating_ip?: string
  flavor: string
  ready: boolean
  kubernetes_version?: string
  created_at: string
  updated_at: string
}

export interface ClusterEvent {
  id: string
  cluster_id: string
  type: string
  severity: EventSeverity
  message: string
  source?: string
  metadata?: Record<string, unknown>
  created_at: string
}

export interface CreateClusterRequest {
  cluster_name: string
  kubernetes_version: string
  control_plane_count: number
  worker_count: number
  control_plane_flavor: string
  worker_flavor: string
  os_image: string
  network_id: string
  subnet_id: string
  external_network_id: string
  ssh_key_name: string
  pod_cidr?: string
  service_cidr?: string
  cni_plugin?: string
  dns_nameservers?: string[]
}

export interface ScaleRequest {
  control_plane_count?: number
  worker_count?: number
}

export interface UpgradeRequest {
  target_version: string
}

// WebSocket 이벤트
export interface ClusterStreamEvent {
  type: string
  cluster_id: string
  data: {
    phase?: string
    ready?: boolean
    control_plane_ready?: boolean
    nodes_ready?: number
    nodes_total?: number
    conditions?: Array<{
      type: string
      status: string
      last_transition_time: string
    }>
  }
  timestamp: string
}
