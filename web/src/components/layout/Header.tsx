'use client'

import { useEffect } from 'react'
import { Moon, Sun, LogOut, User, ChevronDown, Building2 } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useAuthStore } from '@/stores/auth-store'
import { useUIStore } from '@/stores/ui-store'
import { useAuth } from '@/hooks/useAuth'
import { tenantApi } from '@/lib/api-client'
import { Tenant } from '@/types/tenant'
import { PaginatedResponse } from '@/types/api'

export function Header() {
  const { user } = useAuthStore()
  const { theme, setTheme, selectedTenantId, setSelectedTenantId } = useUIStore()
  const { logout } = useAuth()
  const isSuperAdmin = user?.role === 'super_admin'

  // Super Admin인 경우 테넌트 목록 조회
  const { data: tenantsData } = useQuery({
    queryKey: ['tenants-selector'],
    queryFn: async () => {
      const res = await tenantApi.list({ page: 1, page_size: 100 })
      return res.data as PaginatedResponse<Tenant>
    },
    enabled: isSuperAdmin,
  })
  const tenants = tenantsData?.data || []

  // 테넌트가 하나뿐이면 자동 선택
  useEffect(() => {
    if (isSuperAdmin && tenants.length === 1 && !selectedTenantId) {
      setSelectedTenantId(tenants[0].id)
    }
  }, [tenants, isSuperAdmin, selectedTenantId, setSelectedTenantId])

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }

  const roleLabel: Record<string, string> = {
    super_admin: 'Super Admin',
    tenant_admin: 'Tenant Admin',
    operator: 'Operator',
    viewer: 'Viewer',
  }

  const selectedTenantName = tenants.find((t) => t.id === selectedTenantId)?.name

  return (
    <header className="flex h-14 items-center justify-between border-b bg-card px-4">
      <div className="flex items-center gap-3">
        <span className="text-sm text-muted-foreground hidden md:block">
          Kubernetes Cluster Management Platform
        </span>

        {/* Super Admin 전용 테넌트 선택기 */}
        {isSuperAdmin && (
          <div className="flex items-center gap-2">
            <Building2 className="h-4 w-4 text-muted-foreground" />
            <Select
              value={selectedTenantId || ''}
              onValueChange={(val) => setSelectedTenantId(val || null)}
            >
              <SelectTrigger className="h-8 w-48 text-xs">
                <SelectValue placeholder="테넌트 선택 (필수)" />
              </SelectTrigger>
              <SelectContent>
                {tenants.map((t) => (
                  <SelectItem key={t.id} value={t.id}>
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {!selectedTenantId && (
              <span className="text-xs text-destructive font-medium">← 테넌트를 선택하세요</span>
            )}
          </div>
        )}
      </div>

      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon" onClick={toggleTheme}>
          {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="gap-2">
              <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-primary-foreground text-xs font-bold">
                {user?.name?.[0]?.toUpperCase() || user?.email?.[0]?.toUpperCase() || 'U'}
              </div>
              <div className="hidden text-left md:block">
                <p className="text-sm font-medium leading-none">{user?.name || user?.email}</p>
                <p className="text-xs text-muted-foreground">{roleLabel[user?.role || ''] || user?.role}</p>
              </div>
              <ChevronDown className="h-4 w-4 opacity-50" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuLabel>내 계정</DropdownMenuLabel>
            {isSuperAdmin && selectedTenantName && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuLabel className="text-xs text-muted-foreground font-normal">
                  선택된 테넌트: {selectedTenantName}
                </DropdownMenuLabel>
              </>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <User className="mr-2 h-4 w-4" />
              프로필
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={logout} className="text-destructive">
              <LogOut className="mr-2 h-4 w-4" />
              로그아웃
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
