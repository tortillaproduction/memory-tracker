import { useQuery } from '@tanstack/react-query'
import { fetchCurrentUser, logout as logoutRequest } from '../api/auth'
import { ApiError } from '../api/client'

export function useAuth() {
    const { data: user, isLoading, isError, error } = useQuery({
        queryKey: ['currentUser'],
        queryFn: fetchCurrentUser,
        retry: false, // 401は再試行しても意味がないので即座に諦める
  })

  // 未ログイン(401)は「エラー」ではなく「未認証状態」として扱う。
  const isUnauthenticated = isError && error instanceof ApiError && error.status === 401

  async function logout() {
    await logoutRequest()
    window.location.href = '/'
  }

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    isUnauthenticated,
    logout,
  }
}