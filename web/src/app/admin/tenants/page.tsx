'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Plus, Trash2, Pencil } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { tenantApi } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import { Tenant } from '@/types/tenant'
import { PaginatedResponse } from '@/types/api'

const tenantSchema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
  os_auth_url: z.string().url(),
  os_username: z.string().min(1),
  os_password: z.string().min(1),
  os_project_name: z.string().min(1),
  os_domain_name: z.string().default('Default'),
})

const editTenantSchema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
  os_auth_url: z.string().url(),
  os_username: z.string().min(1),
  os_password: z.string().optional(), // 비워두면 기존 비밀번호 유지
  os_project_name: z.string().min(1),
  os_domain_name: z.string().default('Default'),
})

type TenantForm = z.infer<typeof tenantSchema>
type EditTenantForm = z.infer<typeof editTenantSchema>

export default function TenantsPage() {
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [showCreate, setShowCreate] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Tenant | null>(null)
  const [editTarget, setEditTarget] = useState<Tenant | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['tenants', page],
    queryFn: async () => {
      const res = await tenantApi.list({ page, page_size: 20 })
      return res.data as PaginatedResponse<Tenant>
    },
  })

  const { mutate: createTenant, isPending: isCreating } = useMutation({
    mutationFn: (data: TenantForm) => tenantApi.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tenants'] })
      setShowCreate(false)
      toast.success('테넌트가 생성되었습니다.')
    },
    onError: () => toast.error('테넌트 생성에 실패했습니다.'),
  })

  const { mutate: updateTenant, isPending: isUpdating } = useMutation({
    mutationFn: ({ id, data }: { id: string; data: EditTenantForm }) => {
      const payload: Record<string, unknown> = { ...data }
      if (!payload.os_password) delete payload.os_password
      return tenantApi.update(id, payload)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tenants'] })
      setEditTarget(null)
      toast.success('테넌트가 수정되었습니다.')
    },
    onError: () => toast.error('테넌트 수정에 실패했습니다.'),
  })

  const { mutate: deleteTenant, isPending: isDeleting } = useMutation({
    mutationFn: (id: string) => tenantApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tenants'] })
      setDeleteTarget(null)
      toast.success('테넌트가 삭제되었습니다.')
    },
    onError: () => toast.error('테넌트 삭제에 실패했습니다.'),
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<TenantForm>({ resolver: zodResolver(tenantSchema) })

  const {
    register: registerEdit,
    handleSubmit: handleEditSubmit,
    reset: resetEdit,
    formState: { errors: editErrors },
  } = useForm<EditTenantForm>({ resolver: zodResolver(editTenantSchema) })

  const openEdit = (tenant: Tenant) => {
    resetEdit({
      name: tenant.name,
      description: tenant.description || '',
      os_auth_url: tenant.os_auth_url,
      os_username: tenant.os_username,
      os_password: '',
      os_project_name: tenant.os_project_name,
      os_domain_name: tenant.os_domain_name || 'Default',
    })
    setEditTarget(tenant)
  }

  const tenants = data?.data || []
  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <AuthGuard>
      <MainLayout>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">테넌트 관리</h1>
              <p className="text-muted-foreground">전체 {data?.total || 0}개</p>
            </div>
            <Button onClick={() => { reset(); setShowCreate(true) }}>
              <Plus className="mr-2 h-4 w-4" />
              새 테넌트
            </Button>
          </div>

          <Card>
            <CardContent className="p-0">
              {isLoading ? (
                <div className="flex h-40 items-center justify-center">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              ) : tenants.length === 0 ? (
                <div className="flex h-40 items-center justify-center text-muted-foreground">
                  테넌트 없음
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>이름</TableHead>
                      <TableHead>설명</TableHead>
                      <TableHead>상태</TableHead>
                      <TableHead>생성일</TableHead>
                      <TableHead className="text-right">작업</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tenants.map((tenant: Tenant) => (
                      <TableRow key={tenant.id}>
                        <TableCell className="font-medium">{tenant.name}</TableCell>
                        <TableCell className="text-muted-foreground">{tenant.description || '-'}</TableCell>
                        <TableCell>
                          <Badge variant={tenant.status === 'active' ? 'success' : 'secondary'}>
                            {tenant.status === 'active' ? '활성' : tenant.status}
                          </Badge>
                        </TableCell>
                        <TableCell>{formatDate(tenant.created_at)}</TableCell>
                        <TableCell>
                          <div className="flex justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => openEdit(tenant)}
                            >
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="text-destructive"
                              onClick={() => setDeleteTarget(tenant)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
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

        {/* Create dialog */}
        <Dialog open={showCreate} onOpenChange={setShowCreate}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>새 테넌트 생성</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleSubmit((d) => createTenant(d))} className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>이름 *</Label>
                  <Input {...register('name')} placeholder="my-tenant" />
                  {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>도메인 *</Label>
                  <Input {...register('os_domain_name')} placeholder="Default" />
                </div>
              </div>
              <div className="space-y-2">
                <Label>설명</Label>
                <Input {...register('description')} />
              </div>
              <div className="space-y-2">
                <Label>OpenStack Auth URL *</Label>
                <Input {...register('os_auth_url')} placeholder="https://openstack.example.com:5000/v3" />
                {errors.os_auth_url && <p className="text-xs text-destructive">{errors.os_auth_url.message}</p>}
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>OS 사용자명 *</Label>
                  <Input {...register('os_username')} />
                </div>
                <div className="space-y-2">
                  <Label>OS 비밀번호 *</Label>
                  <Input type="password" {...register('os_password')} />
                </div>
              </div>
              <div className="space-y-2">
                <Label>프로젝트명 *</Label>
                <Input {...register('os_project_name')} />
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => setShowCreate(false)}>취소</Button>
                <Button type="submit" disabled={isCreating}>
                  {isCreating ? '생성 중...' : '생성'}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        {/* Edit dialog */}
        <Dialog open={!!editTarget} onOpenChange={() => setEditTarget(null)}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>테넌트 수정</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleEditSubmit((d) => editTarget && updateTenant({ id: editTarget.id, data: d }))} className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>이름 *</Label>
                  <Input {...registerEdit('name')} placeholder="my-tenant" />
                  {editErrors.name && <p className="text-xs text-destructive">{editErrors.name.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>도메인 *</Label>
                  <Input {...registerEdit('os_domain_name')} placeholder="Default" />
                </div>
              </div>
              <div className="space-y-2">
                <Label>설명</Label>
                <Input {...registerEdit('description')} />
              </div>
              <div className="space-y-2">
                <Label>OpenStack Auth URL *</Label>
                <Input {...registerEdit('os_auth_url')} placeholder="https://openstack.example.com:5000/v3" />
                {editErrors.os_auth_url && <p className="text-xs text-destructive">{editErrors.os_auth_url.message}</p>}
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>OS 사용자명 *</Label>
                  <Input {...registerEdit('os_username')} />
                </div>
                <div className="space-y-2">
                  <Label>OS 비밀번호</Label>
                  <Input type="password" {...registerEdit('os_password')} placeholder="변경 시에만 입력" />
                  <p className="text-xs text-muted-foreground">비워두면 기존 비밀번호 유지</p>
                </div>
              </div>
              <div className="space-y-2">
                <Label>프로젝트명 *</Label>
                <Input {...registerEdit('os_project_name')} />
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => setEditTarget(null)}>취소</Button>
                <Button type="submit" disabled={isUpdating}>
                  {isUpdating ? '저장 중...' : '저장'}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        {/* Delete dialog */}
        <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>테넌트 삭제</DialogTitle>
            </DialogHeader>
            <p className="text-sm text-muted-foreground">
              <strong>{deleteTarget?.name}</strong> 테넌트를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.
            </p>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDeleteTarget(null)}>취소</Button>
              <Button variant="destructive" onClick={() => deleteTarget && deleteTenant(deleteTarget.id)} disabled={isDeleting}>
                {isDeleting ? '삭제 중...' : '삭제'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </MainLayout>
    </AuthGuard>
  )
}
