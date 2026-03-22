'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { Toaster } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { useUIStore } from '@/stores/ui-store'

function ThemeProvider({ children }: { children: React.ReactNode }) {
  const { theme } = useUIStore()

  useEffect(() => {
    const root = document.documentElement
    root.classList.remove('light', 'dark')
    if (theme === 'system') {
      const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      root.classList.add(systemDark ? 'dark' : 'light')
    } else {
      root.classList.add(theme)
    }
  }, [theme])

  return <>{children}</>
}

function AuthInitializer({ children }: { children: React.ReactNode }) {
  const { initFromToken } = useAuthStore()

  useEffect(() => {
    initFromToken()
  }, [initFromToken])

  return <>{children}</>
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            retry: 1,
          },
        },
      })
  )

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <AuthInitializer>
          {children}
          <Toaster richColors position="top-right" />
        </AuthInitializer>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
