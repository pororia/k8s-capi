import { create } from 'zustand'
import { User } from '@/types/user'
import { authStorage, parseJwt } from '@/lib/auth'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  setUser: (user: User | null) => void
  setLoading: (loading: boolean) => void
  initFromToken: () => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true,

  setUser: (user) => set({ user, isAuthenticated: !!user }),

  setLoading: (isLoading) => set({ isLoading }),

  initFromToken: () => {
    const token = authStorage.getAccessToken()
    if (!token) {
      set({ user: null, isAuthenticated: false, isLoading: false })
      return
    }
    const claims = parseJwt(token)
    if (!claims) {
      authStorage.clearTokens()
      set({ user: null, isAuthenticated: false, isLoading: false })
      return
    }
    // Basic user from claims — will be replaced by /auth/me call
    set({
      user: {
        id: claims.user_id as string,
        tenant_id: claims.tenant_id as string,
        email: '',
        name: '',
        role: claims.role as User['role'],
        status: 'active' as User['status'],
        created_at: '',
        updated_at: '',
      },
      isAuthenticated: true,
      isLoading: false,
    })
  },

  logout: () => {
    authStorage.clearTokens()
    set({ user: null, isAuthenticated: false })
  },
}))
