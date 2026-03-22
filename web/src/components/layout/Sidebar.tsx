'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  Server,
  Building2,
  Users,
  ClipboardList,
  Settings,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useUIStore } from '@/stores/ui-store'
import { useAuthStore } from '@/stores/auth-store'

const navItems = [
  { href: '/dashboard', label: '대시보드', icon: LayoutDashboard, minRole: 'viewer' },
  { href: '/clusters', label: '클러스터', icon: Server, minRole: 'viewer' },
]

const adminItems = [
  { href: '/admin/tenants', label: '테넌트 관리', icon: Building2 },
  { href: '/admin/users', label: '사용자 관리', icon: Users },
  { href: '/admin/audit-logs', label: '감사 로그', icon: ClipboardList },
]

export function Sidebar() {
  const pathname = usePathname()
  const { sidebarOpen, toggleSidebar } = useUIStore()
  const { user } = useAuthStore()
  const isSuperAdmin = user?.role === 'super_admin'

  return (
    <aside
      className={cn(
        'relative flex h-full flex-col border-r bg-card transition-all duration-200',
        sidebarOpen ? 'w-56' : 'w-14'
      )}
    >
      {/* Logo */}
      <div className="flex h-14 items-center border-b px-3">
        {sidebarOpen ? (
          <span className="font-bold text-lg text-primary">K8s CMP</span>
        ) : (
          <span className="font-bold text-lg text-primary">K</span>
        )}
      </div>

      {/* Nav */}
      <nav className="flex-1 space-y-1 p-2">
        {navItems.map((item) => {
          const Icon = item.icon
          const active = pathname === item.href || pathname.startsWith(item.href + '/')
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-md px-2 py-2 text-sm font-medium transition-colors',
                active
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {sidebarOpen && <span>{item.label}</span>}
            </Link>
          )
        })}

        {isSuperAdmin && (
          <>
            {sidebarOpen && (
              <p className="px-2 pt-4 pb-1 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                관리자
              </p>
            )}
            {!sidebarOpen && <div className="my-2 border-t" />}
            {adminItems.map((item) => {
              const Icon = item.icon
              const active = pathname === item.href
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    'flex items-center gap-3 rounded-md px-2 py-2 text-sm font-medium transition-colors',
                    active
                      ? 'bg-primary/10 text-primary'
                      : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                  )}
                >
                  <Icon className="h-4 w-4 shrink-0" />
                  {sidebarOpen && <span>{item.label}</span>}
                </Link>
              )
            })}
          </>
        )}
      </nav>

      {/* Settings */}
      <div className="border-t p-2">
        <Link
          href="/settings"
          className={cn(
            'flex items-center gap-3 rounded-md px-2 py-2 text-sm font-medium transition-colors',
            pathname === '/settings'
              ? 'bg-primary/10 text-primary'
              : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
          )}
        >
          <Settings className="h-4 w-4 shrink-0" />
          {sidebarOpen && <span>설정</span>}
        </Link>
      </div>

      {/* Toggle */}
      <button
        onClick={toggleSidebar}
        className="absolute -right-3 top-16 flex h-6 w-6 items-center justify-center rounded-full border bg-background shadow-sm hover:bg-accent"
      >
        {sidebarOpen ? (
          <ChevronLeft className="h-3 w-3" />
        ) : (
          <ChevronRight className="h-3 w-3" />
        )}
      </button>
    </aside>
  )
}
