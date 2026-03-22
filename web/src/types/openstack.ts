export interface Flavor {
  id: string
  name: string
  vcpus: number
  ram: number   // MB
  disk: number  // GB
}

export interface OSImage {
  id: string
  name: string
  status: string
  k8s_version?: string
}

export interface Network {
  id: string
  name: string
  external?: boolean
  status: string
}

export interface Subnet {
  id: string
  name: string
  network_id: string
  cidr: string
  ip_version: number
}

export interface SecurityGroup {
  id: string
  name: string
  description?: string
}

export interface Keypair {
  name: string
  fingerprint: string
}

export interface QuotaDetail {
  limit: number
  used: number
}

export interface Quota {
  compute: {
    instances: QuotaDetail
    cores: QuotaDetail
    ram: QuotaDetail
  }
}
