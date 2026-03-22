'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { auditApi } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import { AuditLog } from '@/types/audit'
import { PaginatedResponse } from '@/types/api'

export default function AuditLogsPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', page, search],
    queryFn: async () => {
      const res = await auditApi.list({ page, page_size: 50, ...(search ? { action: search } : {}) })
      return res.data as PaginatedResponse<AuditLog>
    },
  })

  const logs = data?.data || []
  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <AuthGuard>
      <MainLayout>
        <div className="space-y-4">
          <div>
            <h1 className="text-2xl font-bold">감사 로그</h1>
            <p className="text-muted-foreground">전체 {data?.total || 0}건</p>
          </div>

          <div className="flex gap-2">
            <Input
              placeholder="액션 검색..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="max-w-xs"
            />
          </div>

          <Card>
            <CardContent className="p-0">
              {isLoading ? (
                <div className="flex h-40 items-center justify-center">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              ) : logs.length === 0 ? (
                <div className="flex h-40 items-center justify-center text-muted-foreground">
                  감사 로그 없음
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>시간</TableHead>
                      <TableHead>액션</TableHead>
                      <TableHead>리소스</TableHead>
                      <TableHead>결과</TableHead>
                      <TableHead>IP</TableHead>
                      <TableHead>소요시간</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.map((log: AuditLog) => (
                      <TableRow key={log.id}>
                        <TableCell className="whitespace-nowrap text-xs">
                          {formatDate(log.created_at)}
                        </TableCell>
                        <TableCell>
                          <code className="text-xs bg-muted px-1 py-0.5 rounded">{log.action}</code>
                        </TableCell>
                        <TableCell>
                          <span className="text-sm">{log.resource_type}</span>
                          {log.resource_name && (
                            <span className="ml-1 text-xs text-muted-foreground">({log.resource_name})</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <Badge variant={log.result === 'success' ? 'success' : 'destructive'}>
                            {log.result === 'success' ? '성공' : '실패'}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">{log.ip_address || '-'}</TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {log.duration_ms != null ? `${log.duration_ms}ms` : '-'}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>

          {totalPages > 1 && (
            <div className="flex justify-center gap-2">
              <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage((p) => p - 1)}>이전</Button>
              <span className="flex items-center text-sm text-muted-foreground">{page} / {totalPages}</span>
              <Button variant="outline" size="sm" disabled={page === totalPages} onClick={() => setPage((p) => p + 1)}>다음</Button>
            </div>
          )}
        </div>
      </MainLayout>
    </AuthGuard>
  )
}
