'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Plus, Download, Trash2, Scale, ArrowUp } from 'lucide-react'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ClusterStatusBadge } from '@/components/clusters/ClusterStatusBadge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useClusters, useDeleteCluster } from '@/hooks/useClusters'
import { useAuthStore } from '@/stores/auth-store'
import { clusterApi } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import { Cluster } from '@/types/cluster'

export default function ClustersPage() {
  const router = useRouter()
  const { user } = useAuthStore()
  const [page, setPage] = useState(1)
  const [deleteTarget, setDeleteTarget] = useState<Cluster | null>(null)

  const { data, isLoading } = useClusters({ page, page_size: 20 })
  const { mutate: deleteCluster, isPending: isDeleting } = useDeleteCluster()

  const clusters = data?.data || []
  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  const canWrite = user?.role !== 'viewer'

  const handleDelete = () => {
    if (!deleteTarget) return
    deleteCluster(deleteTarget.id, {
      onSuccess: () => {
        toast.success(`클러스터 "${deleteTarget.name}" 삭제 요청이 완료되었습니다.`)
        setDeleteTarget(null)
      },
      onError: () => {
        toast.error('클러스터 삭제에 실패했습니다.')
      },
    })
  }

  const handleDownloadKubeconfig = async (cluster: Cluster) => {
    try {
      const res = await clusterApi.getKubeconfig(cluster.id)
      const url = window.URL.createObjectURL(res.data as Blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `kubeconfig-${cluster.name}.yaml`
      a.click()
      window.URL.revokeObjectURL(url)
    } catch {
      toast.error('Kubeconfig 다운로드에 실패했습니다.')
    }
  }

  return (
    <AuthGuard>
      <MainLayout>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">클러스터</h1>
              <p className="text-muted-foreground">전체 {data?.total || 0}개</p>
            </div>
            {canWrite && (
              <Button asChild>
                <Link href="/clusters/new">
                  <Plus className="mr-2 h-4 w-4" />
                  새 클러스터
                </Link>
              </Button>
            )}
          </div>

          <Card>
            <CardContent className="p-0">
              {isLoading ? (
                <div className="flex h-40 items-center justify-center">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              ) : clusters.length === 0 ? (
                <div className="flex h-40 flex-col items-center justify-center gap-2 text-muted-foreground">
                  <p>클러스터가 없습니다</p>
                  {canWrite && (
                    <Button variant="outline" asChild>
                      <Link href="/clusters/new">첫 클러스터 만들기</Link>
                    </Button>
                  )}
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>이름</TableHead>
                      <TableHead>상태</TableHead>
                      <TableHead>버전</TableHead>
                      <TableHead>Control Plane</TableHead>
                      <TableHead>Worker</TableHead>
                      <TableHead>생성일</TableHead>
                      <TableHead className="text-right">작업</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {clusters.map((cluster: Cluster) => (
                      <TableRow
                        key={cluster.id}
                        className="cursor-pointer"
                        onClick={() => router.push(`/clusters/${cluster.id}`)}
                      >
                        <TableCell className="font-medium">{cluster.name}</TableCell>
                        <TableCell>
                          <ClusterStatusBadge status={cluster.status} />
                        </TableCell>
                        <TableCell>{cluster.kubernetes_version}</TableCell>
                        <TableCell>{cluster.cp_count}</TableCell>
                        <TableCell>{cluster.worker_count}</TableCell>
                        <TableCell>{formatDate(cluster.created_at)}</TableCell>
                        <TableCell>
                          <div
                            className="flex justify-end gap-1"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <Button
                              variant="ghost"
                              size="icon"
                              title="Kubeconfig 다운로드"
                              onClick={() => handleDownloadKubeconfig(cluster)}
                            >
                              <Download className="h-4 w-4" />
                            </Button>
                            {canWrite && (
                              <>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  title="스케일"
                                  onClick={() => router.push(`/clusters/${cluster.id}/scale`)}
                                  disabled={cluster.status !== 'running'}
                                >
                                  <Scale className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  title="업그레이드"
                                  onClick={() => router.push(`/clusters/${cluster.id}/upgrade`)}
                                  disabled={cluster.status !== 'running'}
                                >
                                  <ArrowUp className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  title="삭제"
                                  className="text-destructive hover:text-destructive"
                                  onClick={() => setDeleteTarget(cluster)}
                                  disabled={!['running', 'failed'].includes(cluster.status as string)}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page === 1}
                onClick={() => setPage((p) => p - 1)}
              >
                이전
              </Button>
              <span className="flex items-center text-sm text-muted-foreground">
                {page} / {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page === totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                다음
              </Button>
            </div>
          )}
        </div>

        {/* Delete dialog */}
        <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>클러스터 삭제</DialogTitle>
              <DialogDescription>
                <strong>{deleteTarget?.name}</strong> 클러스터를 삭제하시겠습니까?
                <br />이 작업은 되돌릴 수 없습니다.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDeleteTarget(null)}>
                취소
              </Button>
              <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
                {isDeleting ? '삭제 중...' : '삭제'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </MainLayout>
    </AuthGuard>
  )
}
