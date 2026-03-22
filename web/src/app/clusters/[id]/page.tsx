'use client'

import { useParams, useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Download, Scale, ArrowUp, Trash2, RefreshCw } from 'lucide-react'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ClusterStatusBadge } from '@/components/clusters/ClusterStatusBadge'
import { useCluster, useClusterNodes, useClusterEvents, useDeleteCluster } from '@/hooks/useClusters'
import { useClusterWebSocket } from '@/hooks/useWebSocket'
import { useAuthStore } from '@/stores/auth-store'
import { clusterApi } from '@/lib/api-client'
import { formatDate, formatRelativeTime } from '@/lib/utils'
import { Node, ClusterStreamEvent } from '@/types/cluster'

const NODE_STATUS_COLORS: Record<string, string> = {
  running: 'success',
  provisioning: 'info',
  failed: 'destructive',
  deleting: 'secondary',
}

export default function ClusterDetailPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const { user } = useAuthStore()
  const canWrite = user?.role !== 'viewer'

  const { data: cluster, isLoading, refetch } = useCluster(id)
  const { data: nodes = [] } = useClusterNodes(id)
  const { data: eventsData } = useClusterEvents(id, { page: 1, page_size: 50 })
  const events = eventsData?.data || []
  const { mutate: deleteCluster, isPending: isDeleting } = useDeleteCluster()

  useClusterWebSocket(id, (_event: ClusterStreamEvent) => {
    refetch()
  })

  const handleDownload = async () => {
    try {
      const res = await clusterApi.getKubeconfig(id)
      const url = window.URL.createObjectURL(res.data as Blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `kubeconfig-${cluster?.name}.yaml`
      a.click()
      window.URL.revokeObjectURL(url)
    } catch {
      toast.error('Kubeconfig 다운로드에 실패했습니다.')
    }
  }

  const handleDelete = () => {
    if (!cluster) return
    if (!confirm(`"${cluster.name}" 클러스터를 삭제하시겠습니까?`)) return
    deleteCluster(id, {
      onSuccess: () => {
        toast.success('삭제 요청이 완료되었습니다.')
        router.push('/clusters')
      },
      onError: () => toast.error('삭제에 실패했습니다.'),
    })
  }

  if (isLoading) {
    return (
      <AuthGuard>
        <MainLayout>
          <div className="space-y-4">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-32 w-full" />
          </div>
        </MainLayout>
      </AuthGuard>
    )
  }

  if (!cluster) {
    return (
      <AuthGuard>
        <MainLayout>
          <p>클러스터를 찾을 수 없습니다.</p>
        </MainLayout>
      </AuthGuard>
    )
  }

  return (
    <AuthGuard>
      <MainLayout>
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-3">
                <h1 className="text-2xl font-bold">{cluster.name}</h1>
                <ClusterStatusBadge status={cluster.status} />
              </div>
              <p className="text-sm text-muted-foreground">
                생성: {formatDate(cluster.created_at)} · 버전: {cluster.kubernetes_version}
              </p>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => refetch()}>
                <RefreshCw className="mr-2 h-4 w-4" />
                새로고침
              </Button>
              <Button variant="outline" size="sm" onClick={handleDownload}>
                <Download className="mr-2 h-4 w-4" />
                Kubeconfig
              </Button>
              {canWrite && (
                <>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => router.push(`/clusters/${id}/scale`)}
                    disabled={cluster.status !== 'running'}
                  >
                    <Scale className="mr-2 h-4 w-4" />
                    스케일
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => router.push(`/clusters/${id}/upgrade`)}
                    disabled={cluster.status !== 'running'}
                  >
                    <ArrowUp className="mr-2 h-4 w-4" />
                    업그레이드
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={handleDelete}
                    disabled={isDeleting || !['running', 'failed'].includes(cluster.status as string)}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    삭제
                  </Button>
                </>
              )}
            </div>
          </div>

          <Tabs defaultValue="overview">
            <TabsList>
              <TabsTrigger value="overview">개요</TabsTrigger>
              <TabsTrigger value="nodes">노드 ({nodes.length})</TabsTrigger>
              <TabsTrigger value="events">이벤트 ({events.length})</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Control Plane</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-bold">{cluster.cp_count}</p>
                    <p className="text-xs text-muted-foreground">Flavor: {cluster.cp_flavor}</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Worker 노드</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-bold">{cluster.worker_count}</p>
                    <p className="text-xs text-muted-foreground">Flavor: {cluster.worker_flavor}</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">API 엔드포인트</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm font-mono break-all">{cluster.api_endpoint || '-'}</p>
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">상세 정보</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                    <dt className="text-muted-foreground">네트워크 ID</dt>
                    <dd className="font-mono text-xs">{cluster.network_id || '-'}</dd>
                    <dt className="text-muted-foreground">서브넷 ID</dt>
                    <dd className="font-mono text-xs">{cluster.subnet_id || '-'}</dd>
                    <dt className="text-muted-foreground">외부 네트워크 ID</dt>
                    <dd className="font-mono text-xs">{cluster.external_network_id || '-'}</dd>
                    <dt className="text-muted-foreground">Pod CIDR</dt>
                    <dd>{cluster.pod_cidr || '-'}</dd>
                    <dt className="text-muted-foreground">Service CIDR</dt>
                    <dd>{cluster.service_cidr || '-'}</dd>
                    <dt className="text-muted-foreground">SSH 키페어</dt>
                    <dd>{cluster.ssh_key_name || '-'}</dd>
                    <dt className="text-muted-foreground">이미지</dt>
                    <dd>{cluster.os_image || '-'}</dd>
                    <dt className="text-muted-foreground">마지막 업데이트</dt>
                    <dd>{formatDate(cluster.updated_at)}</dd>
                  </dl>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="nodes">
              <Card>
                <CardContent className="p-0">
                  {nodes.length === 0 ? (
                    <div className="flex h-32 items-center justify-center text-muted-foreground">
                      노드 정보 없음
                    </div>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>이름</TableHead>
                          <TableHead>역할</TableHead>
                          <TableHead>상태</TableHead>
                          <TableHead>IP</TableHead>
                          <TableHead>생성일</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {nodes.map((node: Node) => (
                          <TableRow key={node.id}>
                            <TableCell className="font-mono text-sm">{node.name}</TableCell>
                            <TableCell>
                              <Badge variant="outline">{node.role}</Badge>
                            </TableCell>
                            <TableCell>
                              <Badge variant={(NODE_STATUS_COLORS[node.status] || 'secondary') as 'success' | 'destructive' | 'secondary'}>
                                {node.status}
                              </Badge>
                            </TableCell>
                            <TableCell className="font-mono text-sm">{node.private_ip || '-'}</TableCell>
                            <TableCell>{formatRelativeTime(node.created_at)}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="events">
              <Card>
                <CardContent className="p-0">
                  {events.length === 0 ? (
                    <div className="flex h-32 items-center justify-center text-muted-foreground">
                      이벤트 없음
                    </div>
                  ) : (
                    <div className="divide-y">
                      {events.map((event) => (
                        <div key={event.id} className="flex gap-3 p-4">
                          <div
                            className={`mt-1 h-2 w-2 flex-shrink-0 rounded-full ${
                              event.severity === 'error'
                                ? 'bg-destructive'
                                : event.severity === 'warning'
                                ? 'bg-yellow-500'
                                : 'bg-green-500'
                            }`}
                          />
                          <div className="flex-1 space-y-1">
                            <div className="flex items-center justify-between">
                              <p className="text-sm font-medium">{event.message}</p>
                              <p className="text-xs text-muted-foreground">
                                {formatRelativeTime(event.created_at)}
                              </p>
                            </div>
                            <p className="text-xs text-muted-foreground">{event.type}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </MainLayout>
    </AuthGuard>
  )
}
