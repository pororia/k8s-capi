import { Badge } from '@/components/ui/badge'
import { ClusterStatus } from '@/types/cluster'

const statusConfig: Record<ClusterStatus, { label: string; variant: 'success' | 'warning' | 'destructive' | 'secondary' | 'info' }> = {
  pending: { label: '대기 중', variant: 'secondary' },
  provisioning: { label: '프로비저닝 중', variant: 'info' },
  running: { label: '실행 중', variant: 'success' },
  updating: { label: '업데이트 중', variant: 'warning' },
  deleting: { label: '삭제 중', variant: 'secondary' },
  deleted: { label: '삭제됨', variant: 'secondary' },
  failed: { label: '실패', variant: 'destructive' },
}

export function ClusterStatusBadge({ status }: { status: ClusterStatus }) {
  const config = statusConfig[status] || statusConfig.unknown
  return <Badge variant={config.variant}>{config.label}</Badge>
}
