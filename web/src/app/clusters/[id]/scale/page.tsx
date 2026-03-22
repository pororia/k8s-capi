'use client'

import { useParams, useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useCluster, useScaleCluster } from '@/hooks/useClusters'

const schema = z.object({
  worker_count: z.coerce.number().min(1).max(100, '최대 100개'),
})

type FormData = z.infer<typeof schema>

export default function ScalePage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const { data: cluster } = useCluster(id)
  const { mutate: scale, isPending } = useScaleCluster()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { worker_count: cluster?.worker_count || 3 },
  })

  const onSubmit = (data: FormData) => {
    scale(
      { id, data },
      {
        onSuccess: () => {
          toast.success('스케일 요청이 완료되었습니다.')
          router.push(`/clusters/${id}`)
        },
        onError: () => toast.error('스케일 요청에 실패했습니다.'),
      }
    )
  }

  return (
    <AuthGuard>
      <MainLayout>
        <div className="mx-auto max-w-md space-y-6">
          <div>
            <h1 className="text-2xl font-bold">클러스터 스케일</h1>
            <p className="text-muted-foreground">{cluster?.name}</p>
          </div>
          <Card>
            <CardHeader>
              <CardTitle>Worker 노드 수 조정</CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label>현재 Worker 수</Label>
                  <p className="text-2xl font-bold">{cluster?.worker_count || '-'}</p>
                </div>
                <div className="space-y-2">
                  <Label>변경할 Worker 수 *</Label>
                  <Input type="number" min={1} max={100} {...register('worker_count')} />
                  {errors.worker_count && (
                    <p className="text-xs text-destructive">{errors.worker_count.message}</p>
                  )}
                </div>
                <div className="flex gap-2">
                  <Button type="button" variant="outline" onClick={() => router.back()}>
                    취소
                  </Button>
                  <Button type="submit" disabled={isPending}>
                    {isPending ? '요청 중...' : '스케일 적용'}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>
      </MainLayout>
    </AuthGuard>
  )
}
