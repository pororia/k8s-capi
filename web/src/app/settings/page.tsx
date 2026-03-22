'use client'

import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Moon, Sun, Monitor } from 'lucide-react'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { userApi } from '@/lib/api-client'
import { useAuthStore } from '@/stores/auth-store'
import { useUIStore } from '@/stores/ui-store'

const passwordSchema = z
  .object({
    current_password: z.string().min(1),
    new_password: z.string().min(8),
    confirm_password: z.string().min(8),
  })
  .refine((d) => d.new_password === d.confirm_password, {
    message: '새 비밀번호가 일치하지 않습니다',
    path: ['confirm_password'],
  })

type PasswordForm = z.infer<typeof passwordSchema>

export default function SettingsPage() {
  const { user } = useAuthStore()
  const { theme, setTheme } = useUIStore()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PasswordForm>({ resolver: zodResolver(passwordSchema) })

  const { mutate: changePassword, isPending } = useMutation({
    mutationFn: (data: PasswordForm) =>
      userApi.changePassword(user!.id, {
        current_password: data.current_password,
        new_password: data.new_password,
      }),
    onSuccess: () => {
      toast.success('비밀번호가 변경되었습니다.')
      reset()
    },
    onError: () => toast.error('비밀번호 변경에 실패했습니다.'),
  })

  const themeOptions = [
    { value: 'light', label: '라이트', icon: Sun },
    { value: 'dark', label: '다크', icon: Moon },
    { value: 'system', label: '시스템', icon: Monitor },
  ] as const

  return (
    <AuthGuard>
      <MainLayout>
        <div className="mx-auto max-w-2xl space-y-6">
          <div>
            <h1 className="text-2xl font-bold">설정</h1>
          </div>

          {/* Profile */}
          <Card>
            <CardHeader>
              <CardTitle>내 프로필</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <dl className="grid grid-cols-3 gap-2 text-sm">
                <dt className="text-muted-foreground">이름</dt>
                <dd className="col-span-2 font-medium">{user?.name}</dd>
                <dt className="text-muted-foreground">이메일</dt>
                <dd className="col-span-2">{user?.email}</dd>
                <dt className="text-muted-foreground">역할</dt>
                <dd className="col-span-2">{user?.role}</dd>
              </dl>
            </CardContent>
          </Card>

          {/* Theme */}
          <Card>
            <CardHeader>
              <CardTitle>테마</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex gap-3">
                {themeOptions.map(({ value, label, icon: Icon }) => (
                  <button
                    key={value}
                    onClick={() => setTheme(value)}
                    className={`flex flex-1 flex-col items-center gap-2 rounded-lg border p-4 transition-colors ${
                      theme === value
                        ? 'border-primary bg-primary/5'
                        : 'hover:border-primary/50 hover:bg-muted'
                    }`}
                  >
                    <Icon className="h-5 w-5" />
                    <span className="text-sm">{label}</span>
                  </button>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Change password */}
          <Card>
            <CardHeader>
              <CardTitle>비밀번호 변경</CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit((d) => changePassword(d))} className="space-y-4">
                <div className="space-y-2">
                  <Label>현재 비밀번호 *</Label>
                  <Input type="password" {...register('current_password')} />
                  {errors.current_password && (
                    <p className="text-xs text-destructive">{errors.current_password.message}</p>
                  )}
                </div>
                <Separator />
                <div className="space-y-2">
                  <Label>새 비밀번호 *</Label>
                  <Input type="password" {...register('new_password')} />
                  {errors.new_password && (
                    <p className="text-xs text-destructive">{errors.new_password.message}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label>새 비밀번호 확인 *</Label>
                  <Input type="password" {...register('confirm_password')} />
                  {errors.confirm_password && (
                    <p className="text-xs text-destructive">{errors.confirm_password.message}</p>
                  )}
                </div>
                <Button type="submit" disabled={isPending}>
                  {isPending ? '변경 중...' : '비밀번호 변경'}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      </MainLayout>
    </AuthGuard>
  )
}
