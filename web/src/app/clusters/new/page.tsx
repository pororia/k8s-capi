'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { ChevronRight, ChevronLeft } from 'lucide-react'
import { MainLayout } from '@/components/layout/MainLayout'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  useFlavors,
  useImages,
  useNetworks,
  useSubnets,
  useKeypairs,
  useExternalNetworks,
} from '@/hooks/useOpenStackResources'
import { useCreateCluster } from '@/hooks/useClusters'
import { formatBytes } from '@/lib/utils'

const schema = z.object({
  cluster_name: z.string().min(3).max(63).regex(/^[a-z][a-z0-9-]{2,62}$/, '소문자, 숫자, 하이픈만 사용. 소문자로 시작'),
  kubernetes_version: z.string().regex(/^v\d+\.\d+\.\d+$/, 'v1.28.0 형식'),
  control_plane_flavor: z.string().min(1),
  worker_flavor: z.string().min(1),
  os_image: z.string().min(1),
  control_plane_count: z.coerce.number().refine((n) => [1, 3, 5].includes(n), '1, 3, 5 중 선택'),
  worker_count: z.coerce.number().min(1).max(100),
  network_id: z.string().min(1),
  subnet_id: z.string().min(1),
  external_network_id: z.string().min(1),
  ssh_key_name: z.string().min(1),
  pod_cidr: z.string().default('192.168.0.0/16'),
  service_cidr: z.string().default('10.96.0.0/12'),
  dns_nameservers: z.string().default('8.8.8.8'),
})

type FormData = z.infer<typeof schema>

const STEPS = ['기본 정보', '컨트롤 플레인', 'Worker 노드', '네트워크 설정']

const K8S_VERSIONS = ['v1.30.0', 'v1.29.4', 'v1.28.9', 'v1.27.13']

export default function NewClusterPage() {
  const router = useRouter()
  const [step, setStep] = useState(0)
  const { mutate: createCluster, isPending } = useCreateCluster()

  const {
    register,
    handleSubmit,
    control,
    watch,
    trigger,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      control_plane_count: 3,
      worker_count: 3,
      pod_cidr: '192.168.0.0/16',
      service_cidr: '10.96.0.0/12',
      dns_nameservers: '8.8.8.8',
    },
  })

  const selectedNetworkId = watch('network_id')

  const { data: flavors = [] } = useFlavors()
  const { data: images = [] } = useImages()
  const { data: networks = [] } = useNetworks()
  const { data: subnets = [] } = useSubnets(selectedNetworkId)
  const { data: keypairs = [] } = useKeypairs()
  const { data: externalNetworks = [] } = useExternalNetworks()

  const stepFields: Record<number, (keyof FormData)[]> = {
    0: ['cluster_name', 'kubernetes_version'],
    1: ['control_plane_flavor', 'os_image', 'control_plane_count'],
    2: ['worker_flavor', 'worker_count'],
    3: ['network_id', 'subnet_id', 'external_network_id', 'ssh_key_name'],
  }

  const handleNext = async () => {
    const valid = await trigger(stepFields[step])
    if (valid) setStep((s) => s + 1)
  }

  const onSubmit = (data: FormData) => {
    const payload = {
      ...data,
      dns_nameservers: data.dns_nameservers.split(',').map((s) => s.trim()),
    }
    createCluster(payload, {
      onSuccess: () => {
        toast.success('클러스터 생성 요청이 완료되었습니다.')
        router.push('/clusters')
      },
      onError: () => {
        toast.error('클러스터 생성에 실패했습니다.')
      },
    })
  }

  return (
    <AuthGuard>
      <MainLayout>
        <div className="mx-auto max-w-2xl space-y-6">
          <div>
            <h1 className="text-2xl font-bold">새 클러스터 생성</h1>
          </div>

          {/* Step indicator */}
          <div className="flex items-center gap-2">
            {STEPS.map((s, i) => (
              <div key={i} className="flex items-center gap-2">
                <div
                  className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${
                    i === step
                      ? 'bg-primary text-primary-foreground'
                      : i < step
                      ? 'bg-green-500 text-white'
                      : 'bg-muted text-muted-foreground'
                  }`}
                >
                  {i + 1}
                </div>
                <span className={`text-sm ${i === step ? 'font-medium' : 'text-muted-foreground'}`}>
                  {s}
                </span>
                {i < STEPS.length - 1 && <ChevronRight className="h-4 w-4 text-muted-foreground" />}
              </div>
            ))}
          </div>

          <form onSubmit={handleSubmit(onSubmit)}>
            {/* Step 0: Basic info */}
            {step === 0 && (
              <Card>
                <CardHeader>
                  <CardTitle>기본 정보</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label>클러스터 이름 *</Label>
                    <Input placeholder="my-cluster" {...register('cluster_name')} />
                    {errors.cluster_name && <p className="text-xs text-destructive">{errors.cluster_name.message}</p>}
                    <p className="text-xs text-muted-foreground">소문자, 숫자, 하이픈만 사용 (3-63자)</p>
                  </div>
                  <div className="space-y-2">
                    <Label>Kubernetes 버전 *</Label>
                    <Controller
                      control={control}
                      name="kubernetes_version"
                      render={({ field }) => (
                        <Select onValueChange={field.onChange} value={field.value}>
                          <SelectTrigger>
                            <SelectValue placeholder="버전 선택" />
                          </SelectTrigger>
                          <SelectContent>
                            {K8S_VERSIONS.map((v) => (
                              <SelectItem key={v} value={v}>{v}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    />
                    {errors.kubernetes_version && <p className="text-xs text-destructive">{errors.kubernetes_version.message}</p>}
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Step 1: Control Plane */}
            {step === 1 && (
              <Card>
                <CardHeader>
                  <CardTitle>컨트롤 플레인 설정</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label>Flavor *</Label>
                    <Controller
                      control={control}
                      name="control_plane_flavor"
                      render={({ field }) => (
                        <Select onValueChange={field.onChange} value={field.value}>
                          <SelectTrigger>
                            <SelectValue placeholder="Flavor 선택" />
                          </SelectTrigger>
                          <SelectContent>
                            {flavors.map((f) => (
                              <SelectItem key={f.id} value={f.id}>
                                {f.name} ({f.vcpus}vCPU / {formatBytes(f.ram)} / {f.disk}GB)
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    />
                    {errors.control_plane_flavor && <p className="text-xs text-destructive">필수 항목입니다</p>}
                  </div>
                  <div className="space-y-2">
                    <Label>OS 이미지 *</Label>
                    <Controller
                      control={control}
                      name="os_image"
                      render={({ field }) => (
                        <Select onValueChange={field.onChange} value={field.value}>
                          <SelectTrigger>
                            <SelectValue placeholder="이미지 선택" />
                          </SelectTrigger>
                          <SelectContent>
                            {images.map((img) => (
                              <SelectItem key={img.id} value={img.name}>
                                {img.name} {img.k8s_version && `(k8s ${img.k8s_version})`}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    />
                    {errors.os_image && <p className="text-xs text-destructive">필수 항목입니다</p>}
                  </div>
                  <div className="space-y-2">
                    <Label>Control Plane 수 *</Label>
                    <Controller
                      control={control}
                      name="control_plane_count"
                      render={({ field }) => (
                        <Select onValueChange={(v) => field.onChange(parseInt(v))} value={String(field.value)}>
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="1">1 (개발용)</SelectItem>
                            <SelectItem value="3">3 (권장)</SelectItem>
                            <SelectItem value="5">5 (고가용성)</SelectItem>
                          </SelectContent>
                        </Select>
                      )}
                    />
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Step 2: Worker */}
            {step === 2 && (
              <Card>
                <CardHeader>
                  <CardTitle>Worker 노드 설정</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label>Worker Flavor *</Label>
                    <Controller
                      control={control}
                      name="worker_flavor"
                      render={({ field }) => (
                        <Select onValueChange={field.onChange} value={field.value}>
                          <SelectTrigger>
                            <SelectValue placeholder="Flavor 선택" />
                          </SelectTrigger>
                          <SelectContent>
                            {flavors.map((f) => (
                              <SelectItem key={f.id} value={f.id}>
                                {f.name} ({f.vcpus}vCPU / {formatBytes(f.ram)} / {f.disk}GB)
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    />
                    {errors.worker_flavor && <p className="text-xs text-destructive">필수 항목입니다</p>}
                  </div>
                  <div className="space-y-2">
                    <Label>Worker 수 *</Label>
                    <Input type="number" min={1} max={100} {...register('worker_count')} />
                    {errors.worker_count && <p className="text-xs text-destructive">{errors.worker_count.message}</p>}
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Step 3: Network */}
            {step === 3 && (
              <Card>
                <CardHeader>
                  <CardTitle>네트워크 설정</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>네트워크 *</Label>
                      <Controller
                        control={control}
                        name="network_id"
                        render={({ field }) => (
                          <Select onValueChange={field.onChange} value={field.value}>
                            <SelectTrigger>
                              <SelectValue placeholder="네트워크 선택" />
                            </SelectTrigger>
                            <SelectContent>
                              {networks.map((n) => (
                                <SelectItem key={n.id} value={n.id}>{n.name}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>서브넷 *</Label>
                      <Controller
                        control={control}
                        name="subnet_id"
                        render={({ field }) => (
                          <Select onValueChange={field.onChange} value={field.value} disabled={!selectedNetworkId}>
                            <SelectTrigger>
                              <SelectValue placeholder="서브넷 선택" />
                            </SelectTrigger>
                            <SelectContent>
                              {subnets.map((s) => (
                                <SelectItem key={s.id} value={s.id}>{s.name} ({s.cidr})</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label>외부 네트워크 *</Label>
                    <Controller
                      control={control}
                      name="external_network_id"
                      render={({ field }) => (
                        <Select onValueChange={field.onChange} value={field.value}>
                          <SelectTrigger>
                            <SelectValue placeholder="외부 네트워크 선택" />
                          </SelectTrigger>
                          <SelectContent>
                            {externalNetworks.map((n) => (
                              <SelectItem key={n.id} value={n.id}>{n.name}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>SSH 키페어 *</Label>
                      <Controller
                        control={control}
                        name="ssh_key_name"
                        render={({ field }) => (
                          <Select onValueChange={field.onChange} value={field.value}>
                            <SelectTrigger>
                              <SelectValue placeholder="키페어 선택" />
                            </SelectTrigger>
                            <SelectContent>
                              {keypairs.map((kp) => (
                                <SelectItem key={kp.name} value={kp.name}>{kp.name}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>Pod CIDR</Label>
                      <Input {...register('pod_cidr')} />
                    </div>
                    <div className="space-y-2">
                      <Label>Service CIDR</Label>
                      <Input {...register('service_cidr')} />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label>DNS 서버</Label>
                    <Input {...register('dns_nameservers')} placeholder="8.8.8.8,8.8.4.4" />
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Navigation */}
            <div className="flex justify-between pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => (step === 0 ? router.back() : setStep((s) => s - 1))}
              >
                <ChevronLeft className="mr-2 h-4 w-4" />
                {step === 0 ? '취소' : '이전'}
              </Button>
              {step < STEPS.length - 1 ? (
                <Button type="button" onClick={handleNext}>
                  다음
                  <ChevronRight className="ml-2 h-4 w-4" />
                </Button>
              ) : (
                <Button type="submit" disabled={isPending}>
                  {isPending ? '생성 중...' : '클러스터 생성'}
                </Button>
              )}
            </div>
          </form>
        </div>
      </MainLayout>
    </AuthGuard>
  )
}
