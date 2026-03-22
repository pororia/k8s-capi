import { useEffect, useRef } from 'react'
import { wsManager } from '@/lib/websocket'
import { ClusterStreamEvent } from '@/types/cluster'

export function useClusterWebSocket(
  clusterId: string | null,
  onEvent: (event: ClusterStreamEvent) => void
) {
  const handlerRef = useRef(onEvent)
  handlerRef.current = onEvent

  useEffect(() => {
    if (!clusterId) return

    const handler = (data: unknown) => handlerRef.current(data as ClusterStreamEvent)
    wsManager.connect(clusterId)
    wsManager.on('*', handler)

    return () => {
      wsManager.off('*', handler)
      wsManager.disconnect()
    }
  }, [clusterId])
}

export function useAllClustersWebSocket(onEvent: (event: ClusterStreamEvent) => void) {
  const handlerRef = useRef(onEvent)
  handlerRef.current = onEvent

  useEffect(() => {
    const handler = (data: unknown) => handlerRef.current(data as ClusterStreamEvent)
    wsManager.connectAll()
    wsManager.on('*', handler)

    return () => {
      wsManager.off('*', handler)
      wsManager.disconnect()
    }
  }, [])
}
