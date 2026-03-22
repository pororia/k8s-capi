'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import { Server, Activity, AlertTriangle, CheckCircle2 } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ClusterStatusBadge } from '@/components/clusters/ClusterStatusBadge'
import { useClusters } from '@/hooks/useClusters'
import { formatRelativeTime } from '@/lib/utils'
import { Cluster } from '@/types/cluster'

const STATUS_COLORS: Record<string, string> = {
  running: '#22c55e',
  provisioning: '#3b82f6',
  scaling: '#f59e0b',
  upgrading: '#f59e0b',
  failed: '#ef4444',
  deleting: '#6b7280',
  unknown: '#9ca3af',
}

export default function DashboardPage() {
  const { data: clustersData, isLoading } = useClusters({ page: 1, page_size: 100 })
  const clusters = clustersData?.data || []

  const stats = useMemo(() => {
    const total = clusters.length
    const running = clusters.filter((c: Cluster) => c.status === 'running').length
    const failed = clusters.filter((c: Cluster) => c.status === 'failed').length
    const provisioning = clusters.filter((c: Cluster) =>
      ['provisioning', 'updating', 'pending'].includes(c.status)
    ).length

    return { total, running, failed, provisioning }
  }, [clusters])

  const statusChartData = useMemo(() => {
    const counts: Record<string, number> = {}
    clusters.forEach((c: Cluster) => {
      counts[c.status] = (counts[c.status] || 0) + 1
    })
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [clusters])

  const versionChartData = useMemo(() => {
    const counts: Record<string, number> = {}
    clusters.forEach((c: Cluster) => {
      if (c.kubernetes_version) {
        counts[c.kubernetes_version] = (counts[c.kubernetes_version] || 0) + 1
      }
    })
    return Object.entries(counts)
      .map(([version, count]) => ({ version, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 8)
  }, [clusters])

  const recentClusters = clusters.slice(0, 5)

  return (
    <AuthGuard>
      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold">대시보드</h1>
            <p className="text-muted-foreground">클러스터 현황을 한눈에 확인하세요</p>
          </div>

          {/* Summary cards */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">전체 클러스터</CardTitle>
                <Server className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{isLoading ? '...' : stats.total}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">실행 중</CardTitle>
                <CheckCircle2 className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-green-600">{isLoading ? '...' : stats.running}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">작업 중</CardTitle>
                <Activity className="h-4 w-4 text-yellow-500" />
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-yellow-600">{isLoading ? '...' : stats.provisioning}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">실패</CardTitle>
                <AlertTriangle className="h-4 w-4 text-destructive" />
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-destructive">{isLoading ? '...' : stats.failed}</p>
              </CardContent>
            </Card>
          </div>

          {/* Charts */}
          <div className="grid gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">상태별 클러스터</CardTitle>
              </CardHeader>
              <CardContent>
                {statusChartData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={200}>
                    <PieChart>
                      <Pie
                        data={statusChartData}
                        cx="50%"
                        cy="50%"
                        innerRadius={50}
                        outerRadius={80}
                        dataKey="value"
                        label={({ name, value }) => `${name}: ${value}`}
                      >
                        {statusChartData.map((entry, index) => (
                          <Cell key={index} fill={STATUS_COLORS[entry.name] || '#9ca3af'} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="flex h-[200px] items-center justify-center text-muted-foreground">
                    클러스터 없음
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">버전별 클러스터</CardTitle>
              </CardHeader>
              <CardContent>
                {versionChartData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={200}>
                    <BarChart data={versionChartData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="version" tick={{ fontSize: 11 }} />
                      <YAxis allowDecimals={false} />
                      <Tooltip />
                      <Bar dataKey="count" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="flex h-[200px] items-center justify-center text-muted-foreground">
                    클러스터 없음
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Recent clusters */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-base">최근 클러스터</CardTitle>
              <Link href="/clusters" className="text-sm text-primary hover:underline">
                전체 보기
              </Link>
            </CardHeader>
            <CardContent>
              {recentClusters.length === 0 ? (
                <p className="text-center text-muted-foreground py-8">클러스터가 없습니다</p>
              ) : (
                <div className="space-y-3">
                  {recentClusters.map((cluster: Cluster) => (
                    <div
                      key={cluster.id}
                      className="flex items-center justify-between rounded-lg border p-3"
                    >
                      <div className="space-y-1">
                        <Link
                          href={`/clusters/${cluster.id}`}
                          className="font-medium hover:underline"
                        >
                          {cluster.name}
                        </Link>
                        <p className="text-xs text-muted-foreground">
                          {cluster.kubernetes_version} · CP {cluster.cp_count}대 · Worker{' '}
                          {cluster.worker_count}대 · {formatRelativeTime(cluster.created_at)}
                        </p>
                      </div>
                      <ClusterStatusBadge status={cluster.status} />
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </MainLayout>
    </AuthGuard>
  )
}
