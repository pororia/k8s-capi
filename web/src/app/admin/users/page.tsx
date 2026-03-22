'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Plus, Trash2 } from 'lucide-react'
import { useForm, Controller } from 'react-hook-form'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { userApi } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import { User } from '@/types/user'
import { PaginatedResponse } from '@/types/api'

const userSchema = z.object({
  email: z.string().email(),
  name: z.string().min(2),
  password: z.string().min(8),
  role: z.enum(['super_admin', 'tenant_admin', 'operator', 'viewer']),
  tenant_id: z.string().optional(),
})

type UserForm = z.infer<typeof userSchema>

const roleLabels: Record<string, string> = {
  super_admin: 'Super Admin',
  tenant_admin: 'Tenant Admin',
  operator: 'Operator',
  viewer: 'Viewer',
}

const roleVariants: Record<string, 'default' | 'secondary' | 'outline'> = {
  super_admin: 'default',
  tenant_admin: 'secondary',
  operator: 'outline',
  viewer: 'outline',
}

export default function UsersPage() {
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [showCreate, setShowCreate] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['users', page],
    queryFn: async () => {
      const res = await userApi.list({ page, page_size: 20 })
      return res.data as PaginatedResponse<User>
    },
  })

  const { mutate: createUser, isPending: isCreating } = useMutation({
    mutationFn: (data: UserForm) => userApi.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['users'] })
      setShowCreate(false)
      toast.success('사용자가 생성되었습니다.')
    },
    onError: () => toast.error('사용자 생성에 실패했습니다.'),
  })

  const { mutate: deleteUser, isPending: isDeleting } = useMutation({
    mutationFn: (id: string) => userApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['users'] })
      setDeleteTarget(null)
      toast.success('사용자가 삭제되었습니다.')
    },
    onError: () => toast.error('사용자 삭제에 실패했습니다.'),
  })

  const { register, handleSubmit, control, reset, formState: { errors } } = useForm<UserForm>({
    resolver: zodResolver(userSchema),
    defaultValues: { role: 'viewer' },
  })

  const users = data?.data || []
  const totalPages = data ? Math.ceil(data.total / data.page_size) : 1

  return (
    <AuthGuard>
      <MainLayout>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">사용자 관리</h1>
              <p className="text-muted-foreground">전체 {data?.total || 0}명</p>
            </div>
            <Button onClick={() => { reset(); setShowCreate(true) }}>
              <Plus className="mr-2 h-4 w-4" />
              새 사용자
            </Button>
          </div>

          <Card>
            <CardContent className="p-0">
              {isLoading ? (
                <div className="flex h-40 items-center justify-center">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>이름</TableHead>
                      <TableHead>이메일</TableHead>
                      <TableHead>역할</TableHead>
                      <TableHead>상태</TableHead>
                      <TableHead>생성일</TableHead>
                      <TableHead className="text-right">작업</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {users.map((user: User) => (
                      <TableRow key={user.id}>
                        <TableCell className="font-medium">{user.name}</TableCell>
                        <TableCell>{user.email}</TableCell>
                        <TableCell>
                          <Badge variant={roleVariants[user.role] || 'outline'}>
                            {roleLabels[user.role] || user.role}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={user.status === 'active' ? 'success' : 'secondary'}>
                            {user.status === 'active' ? '활성' : user.status}
                          </Badge>
                        </TableCell>
                        <TableCell>{formatDate(user.created_at)}</TableCell>
                        <TableCell>
                          <div className="flex justify-end">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="text-destructive"
                              onClick={() => setDeleteTarget(user)}
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
          <DialogContent>
            <DialogHeader>
              <DialogTitle>새 사용자 생성</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleSubmit((d) => createUser(d))} className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>이름 *</Label>
                  <Input {...register('name')} />
                  {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label>이메일 *</Label>
                  <Input type="email" {...register('email')} />
                  {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
                </div>
              </div>
              <div className="space-y-2">
                <Label>비밀번호 *</Label>
                <Input type="password" {...register('password')} />
                {errors.password && <p className="text-xs text-destructive">{errors.password.message}</p>}
              </div>
              <div className="space-y-2">
                <Label>역할 *</Label>
                <Controller
                  control={control}
                  name="role"
                  render={({ field }) => (
                    <Select onValueChange={field.onChange} value={field.value}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="super_admin">Super Admin</SelectItem>
                        <SelectItem value="tenant_admin">Tenant Admin</SelectItem>
                        <SelectItem value="operator">Operator</SelectItem>
                        <SelectItem value="viewer">Viewer</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                />
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

        {/* Delete dialog */}
        <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>사용자 삭제</DialogTitle>
            </DialogHeader>
            <p className="text-sm text-muted-foreground">
              <strong>{deleteTarget?.name}</strong> ({deleteTarget?.email}) 사용자를 삭제하시겠습니까?
            </p>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDeleteTarget(null)}>취소</Button>
              <Button variant="destructive" onClick={() => deleteTarget && deleteUser(deleteTarget.id)} disabled={isDeleting}>
                {isDeleting ? '삭제 중...' : '삭제'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </MainLayout>
    </AuthGuard>
  )
}
