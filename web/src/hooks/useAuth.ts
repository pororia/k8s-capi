import { useCallback } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { authApi } from '@/lib/api-client'
import { authStorage } from '@/lib/auth'
import { useAuthStore } from '@/stores/auth-store'
import { User } from '@/types/user'
import { ApiResponse } from '@/types/api'

export function useAuth() {
  const { user, isAuthenticated, isLoading, setUser, logout: storeLogout } = useAuthStore()
  const router = useRouter()
  const queryClient = useQueryClient()

  const { refetch: refetchMe } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      const res = await authApi.me()
      const data = res.data as ApiResponse<User>
      setUser(data.data)
      return data.data
    },
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
  })

  const loginMutation = useMutation({
    mutationFn: async ({ email, password }: { email: string; password: string }) => {
      const res = await authApi.login(email, password)
      return res.data
    },
    onSuccess: (data) => {
      const { access_token, refresh_token, user } = data.data
      authStorage.setTokens(access_token, refresh_token)
      if (user) setUser(user)
      else refetchMe()
      router.push('/dashboard')
    },
  })

  const logoutMutation = useMutation({
    mutationFn: () => authApi.logout(),
    onSettled: () => {
      storeLogout()
      queryClient.clear()
      router.push('/login')
    },
  })

  const logout = useCallback(() => {
    logoutMutation.mutate()
  }, [logoutMutation])

  return {
    user,
    isAuthenticated,
    isLoading,
    login: loginMutation.mutate,
    isLoginLoading: loginMutation.isPending,
    loginError: loginMutation.error,
    logout,
  }
}
