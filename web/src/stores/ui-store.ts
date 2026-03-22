import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface UIState {
  sidebarOpen: boolean
  theme: 'light' | 'dark' | 'system'
  selectedTenantId: string | null
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setTheme: (theme: 'light' | 'dark' | 'system') => void
  setSelectedTenantId: (id: string | null) => void
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarOpen: true,
      theme: 'system',
      selectedTenantId: null,

      toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
      setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),
      setTheme: (theme) => set({ theme }),
      setSelectedTenantId: (selectedTenantId) => {
        if (selectedTenantId) {
          sessionStorage.setItem('selected_tenant_id', selectedTenantId)
        } else {
          sessionStorage.removeItem('selected_tenant_id')
        }
        set({ selectedTenantId })
      },
    }),
    {
      name: 'cmp-ui',
      partialize: (state) => ({ theme: state.theme, sidebarOpen: state.sidebarOpen }),
    }
  )
)
