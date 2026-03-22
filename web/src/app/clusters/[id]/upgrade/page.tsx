'use client'

import { useParams, useRouter } from 'next/navigation'
import { Controller, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useCluster, useUpgradeCluster } from '@/hooks/useClusters'

const K8S_VERSIONS = ['v1.30.0', 'v1.29.4', 'v1.28.9', 'v1.27.13']

const schema = z.object({
  target_version: z.string().regex(/^v\d+\.\d+\.\d+$/),
})

type FormData = z.infer<typeof schema>

export default function UpgradePage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const { data: cluster } = useCluster(id)
  const { mutate: upgrade, isPending } = useUpgradeCluster()

  const { control, handleSubmit } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const onSubmit = (data: FormData) => {
    upgrade(
      { id, data },
      {
        onSuccess: () => {
          toast.success('업그레이드 요청이 완료되었습니다.')
          router.push(`/clusters/${id}`)
        },
        onError: () => toast.error('업그레이드 요청에 실패했습니다.'),
      }
    )
  }

  const availableVersions = K8S_VERSIONS.filter((v) => v !== cluster?.kubernetes_version)

  return (
    <AuthGuard>
      <MainLayout>
        <div className="mx-auto max-w-md space-y-6">
          <div>
            <h1 className="text-2xl font-bold">클러스터 업그레이드</h1>
            <p className="text-muted-foreground">{cluster?.name}</p>
          </div>
          <Card>
            <CardHeader>
              <CardTitle>Kubernetes 버전 업그레이드</CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <div className="space-y-2">
                  <Label>현재 버전</Label>
                  <p className="text-2xl font-bold">{cluster?.kubernetes_version || '-'}</p>
                </div>
                <div className="space-y-2">
                  <Label>업그레이드 버전 *</Label>
                  <Controller
                    control={control}
                    name="target_version"
                    render={({ field }) => (
                      <Select onValueChange={field.onChange} value={field.value}>
                        <SelectTrigger>
                          <SelectValue placeholder="버전 선택" />
                        </SelectTrigger>
                        <SelectContent>
                          {availableVersions.map((v) => (
                            <SelectItem key={v} value={v}>{v}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  />
                </div>
                <div className="flex gap-2">
                  <Button type="button" variant="outline" onClick={() => router.back()}>
                    취소
                  </Button>
                  <Button type="submit" disabled={isPending}>
                    {isPending ? '요청 중...' : '업그레이드 시작'}
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
