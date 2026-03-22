import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clusterApi } from '@/lib/api-client'
import { Cluster, ClusterEvent, Node, CreateClusterRequest, ScaleRequest, UpgradeRequest } from '@/types/cluster'
import { ApiResponse, PaginatedResponse } from '@/types/api'

export function useClusters(params?: Record<string, unknown>) {
  return useQuery({
    queryKey: ['clusters', params],
    queryFn: async () => {
      const res = await clusterApi.list(params)
      return res.data as PaginatedResponse<Cluster>
    },
  })
}

export function useCluster(id: string) {
  return useQuery({
    queryKey: ['clusters', id],
    queryFn: async () => {
      const res = await clusterApi.get(id)
      return (res.data as ApiResponse<Cluster>).data
    },
    enabled: !!id,
    refetchInterval: 15000,
  })
}

export function useClusterNodes(id: string) {
  return useQuery({
    queryKey: ['clusters', id, 'nodes'],
    queryFn: async () => {
      const res = await clusterApi.getNodes(id)
      return (res.data as ApiResponse<Node[]>).data
    },
    enabled: !!id,
    refetchInterval: 30000,
  })
}

export function useClusterEvents(id: string, params?: Record<string, unknown>) {
  return useQuery({
    queryKey: ['clusters', id, 'events', params],
    queryFn: async () => {
      const res = await clusterApi.getEvents(id, params)
      return res.data as PaginatedResponse<ClusterEvent>
    },
    enabled: !!id,
  })
}

export function useCreateCluster() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateClusterRequest) => clusterApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['clusters'] }),
  })
}

export function useDeleteCluster() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => clusterApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['clusters'] }),
  })
}

export function useScaleCluster() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ScaleRequest }) => clusterApi.scale(id, data),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['clusters', id] })
      qc.invalidateQueries({ queryKey: ['clusters'] })
    },
  })
}

export function useUpgradeCluster() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpgradeRequest }) => clusterApi.upgrade(id, data),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['clusters', id] })
    },
  })
}
