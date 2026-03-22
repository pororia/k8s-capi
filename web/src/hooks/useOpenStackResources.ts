import { useQuery } from '@tanstack/react-query'
import { openstackApi } from '@/lib/api-client'
import { Flavor, OSImage, Network, Subnet, SecurityGroup, Keypair, Quota } from '@/types/openstack'
import { ApiResponse } from '@/types/api'

const staleTime = 5 * 60 * 1000

export function useFlavors() {
  return useQuery({
    queryKey: ['openstack', 'flavors'],
    queryFn: async () => {
      const res = await openstackApi.getFlavors()
      return (res.data as ApiResponse<Flavor[]>).data
    },
    staleTime,
  })
}

export function useImages() {
  return useQuery({
    queryKey: ['openstack', 'images'],
    queryFn: async () => {
      const res = await openstackApi.getImages()
      return (res.data as ApiResponse<OSImage[]>).data
    },
    staleTime,
  })
}

export function useNetworks() {
  return useQuery({
    queryKey: ['openstack', 'networks'],
    queryFn: async () => {
      const res = await openstackApi.getNetworks()
      return (res.data as ApiResponse<Network[]>).data
    },
    staleTime,
  })
}

export function useSubnets(networkId: string) {
  return useQuery({
    queryKey: ['openstack', 'subnets', networkId],
    queryFn: async () => {
      const res = await openstackApi.getSubnets(networkId)
      return (res.data as ApiResponse<Subnet[]>).data
    },
    enabled: !!networkId,
    staleTime,
  })
}

export function useSecurityGroups() {
  return useQuery({
    queryKey: ['openstack', 'security-groups'],
    queryFn: async () => {
      const res = await openstackApi.getSecurityGroups()
      return (res.data as ApiResponse<SecurityGroup[]>).data
    },
    staleTime,
  })
}

export function useKeypairs() {
  return useQuery({
    queryKey: ['openstack', 'keypairs'],
    queryFn: async () => {
      const res = await openstackApi.getKeypairs()
      return (res.data as ApiResponse<{ name: string; fingerprint: string }[]>).data
    },
    staleTime,
  })
}

export function useExternalNetworks() {
  return useQuery({
    queryKey: ['openstack', 'external-networks'],
    queryFn: async () => {
      const res = await openstackApi.getExternalNetworks()
      return (res.data as ApiResponse<Network[]>).data
    },
    staleTime,
  })
}

export function useQuotas() {
  return useQuery({
    queryKey: ['openstack', 'quotas'],
    queryFn: async () => {
      const res = await openstackApi.getQuotas()
      return (res.data as ApiResponse<Quota>).data
    },
    staleTime,
  })
}
