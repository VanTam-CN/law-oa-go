/**
 * 认证服务钩子 - 基于TanStack Query v5
 * 提供完整的认证状态管理和API调用
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo } from 'react'
import { apiClient } from '../services/apiClient'
import { useAppStore } from '../stores/useAppStore'
import { User } from '../stores/useAppStore'
import { clearStorage, getToken, getUserInfo } from '@/utils/storage'

// 认证相关接口
export interface LoginRequest {
  account: string
  password: string
  captcha?: string
}

export interface LoginResponse {
  user: User
  token: string
  refreshToken: string
  expiresIn: number
}

export interface RefreshTokenRequest {
  refreshToken: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  realName: string
  phone?: string
  department?: string
  position?: string
}

export interface ResetPasswordRequest {
  email: string
  code: string
  newPassword: string
}

// 查询键
const queryKeys = {
  auth: ['auth'] as const,
  currentUser: ['auth', 'currentUser'] as const,
  permissions: ['auth', 'permissions'] as const,
  profile: ['auth', 'profile'] as const,
}

// 认证状态查询Hook（优化版本，使用useMemo缓存计算结果）
export const useAuthState = () => {
  const { user, isAuthenticated, isLoading } = useAppStore()
  const queryClient = useQueryClient()

  // 使用useMemo缓存计算结果，避免重复计算
  const authData = useMemo(
    () => ({ user, isAuthenticated, isLoading }),
    [user, isAuthenticated, isLoading],
  )

  return useQuery({
    queryKey: queryKeys.auth,
    queryFn: async () => {
      // 从localStorage恢复认证状态
      const token = getToken()
      const userInfo = getUserInfo()

      if (token && userInfo) {
        try {
          return { user: userInfo, isAuthenticated: true }
        } catch (error) {
          console.error('Failed to parse user info:', error)
          clearStorage()
          return { user: null, isAuthenticated: false }
        }
      }

      return { user: null, isAuthenticated: false }
    },
    initialData: authData,
    staleTime: 0, // 总是检查最新状态
    gcTime: 0, // 不缓存认证状态
  })
}

// 登录Hook
export const useLogin = () => {
  const queryClient = useQueryClient()
  const { login } = useAppStore()

  return useMutation({
    mutationFn: async (loginData: LoginRequest) => {
      const response = await apiClient.post<LoginResponse>('/api/v1/auth/login', loginData)

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (data) => {
      // 更新应用状态
      login(data.user, data.token)

      // 更新React Query缓存
      queryClient.setQueryData(queryKeys.auth, {
        user: data.user,
        isAuthenticated: true,
      })

      // 预加载用户权限
      queryClient.prefetchQuery({
        queryKey: queryKeys.permissions,
        queryFn: () => apiClient.get('/api/v1/auth/permissions'),
        staleTime: 10 * 60 * 1000, // 10分钟
      })

      console.log('登录成功:', data.user.username)
    },
    onError: (error) => {
      console.error('登录失败:', error.message)
    },
    onSettled: () => {
      // 清除可能存在的过时数据
      queryClient.removeQueries({ queryKey: queryKeys.profile })
    },
  })
}

// 注册Hook
export const useRegister = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (registerData: RegisterRequest) => {
      const response = await apiClient.post<{ message: string }>(
        '/api/v1/auth/register',
        registerData,
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: () => {
      console.log('注册成功')
      // 可以在这里添加自动登录逻辑或显示成功消息
    },
    onError: (error) => {
      console.error('注册失败:', error.message)
    },
  })
}

// 登出Hook
export const useLogout = () => {
  const queryClient = useQueryClient()
  const { logout } = useAppStore()

  return useMutation({
    mutationFn: async () => {
      // 调用服务器登出接口
      const response = await apiClient.post('/api/v1/auth/logout')

      if (response.error) {
        console.warn('服务器登出失败:', response.error.message)
      }

      return response
    },
    onSuccess: () => {
      // 清除本地状态
      logout()

      // 清除React Query缓存
      queryClient.clear()

      console.log('登出成功')
    },
    onError: (error) => {
      console.error('登出过程中出现错误:', error.message)
      // 即使服务器登出失败，也要清除本地状态
      logout()
      queryClient.clear()
    },
  })
}

// 刷新Token Hook
export const useRefreshToken = () => {
  const queryClient = useQueryClient()
  const { login, logout } = useAppStore()

  return useMutation({
    mutationFn: async (refreshToken: string) => {
      const response = await apiClient.post<LoginResponse>('/api/v1/auth/refresh', {
        refreshToken,
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (data) => {
      // 更新token和用户信息
      login(data.user, data.token)

      queryClient.setQueryData(queryKeys.auth, {
        user: data.user,
        isAuthenticated: true,
      })

      console.log('Token刷新成功')
    },
    onError: (error) => {
      console.error('Token刷新失败:', error.message)
      // 刷新失败，执行登出
      logout()
      queryClient.clear()
    },
  })
}

// 获取用户详细资料Hook
export const useUserProfile = (userId?: string) => {
  const currentUser = useAppStore((state) => state.user)
  const targetUserId = userId || currentUser?.id

  return useQuery({
    queryKey: queryKeys.profile,
    queryFn: async () => {
      if (!targetUserId) {
        throw new Error('用户ID不存在')
      }

      const response = await apiClient.get<User>(`/api/v1/users/${targetUserId}`)

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    enabled: !!targetUserId,
    staleTime: 5 * 60 * 1000, // 5分钟
    gcTime: 10 * 60 * 1000, // 10分钟
  })
}

// 更新用户资料Hook
export const useUpdateProfile = () => {
  const queryClient = useQueryClient()
  const { updateUser } = useAppStore()

  return useMutation({
    mutationFn: async (profileData: Partial<User>) => {
      const response = await apiClient.put<User>('/api/v1/auth/profile', profileData)

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onMutate: async (newProfile) => {
      // 取消任何正在进行的查询
      await queryClient.cancelQueries({ queryKey: queryKeys.profile })

      // 快照当前值
      const previousProfile = queryClient.getQueryData(queryKeys.profile)

      // 乐观更新
      queryClient.setQueryData(queryKeys.profile, (old: User) =>
        old ? { ...old, ...newProfile } : newProfile,
      )

      // 更新应用状态
      updateUser(newProfile)

      return { previousProfile }
    },
    onError: (error, variables, context) => {
      // 恢复之前的值
      if (context?.previousProfile) {
        queryClient.setQueryData(queryKeys.profile, context.previousProfile)
      }
      console.error('更新用户资料失败:', error.message)
    },
    onSettled: () => {
      // 重新获取数据
      queryClient.invalidateQueries({ queryKey: queryKeys.profile })
    },
  })
}

// 修改密码Hook
export const useChangePassword = () => {
  return useMutation({
    mutationFn: async (passwordData: {
      currentPassword: string
      newPassword: string
      confirmPassword: string
    }) => {
      const response = await apiClient.post<{ message: string }>(
        '/api/v1/auth/change-password',
        passwordData,
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: () => {
      console.log('密码修改成功')
    },
    onError: (error) => {
      console.error('密码修改失败:', error.message)
    },
  })
}

// 重置密码Hook
export const useResetPassword = () => {
  return useMutation({
    mutationFn: async (resetData: ResetPasswordRequest) => {
      const response = await apiClient.post<{ message: string }>(
        '/api/v1/auth/reset-password',
        resetData,
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: () => {
      console.log('密码重置成功')
    },
    onError: (error) => {
      console.error('密码重置失败:', error.message)
    },
  })
}

// 检查用户权限Hook
export const usePermissions = () => {
  const { user } = useAppStore()

  return useQuery({
    queryKey: queryKeys.permissions,
    queryFn: async () => {
      const response = await apiClient.get<{
        permissions: string[]
        roles: string[]
      }>('/api/v1/auth/permissions')

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    enabled: !!user,
    staleTime: 15 * 60 * 1000, // 15分钟
    gcTime: 30 * 60 * 1000, // 30分钟
  })
}

// 工具Hook：检查是否有特定权限（优化版本，使用useMemo缓存结果）
export const useHasPermission = (permission: string) => {
  const { user } = useAppStore()
  const { data: permissions, isLoading } = usePermissions()

  const hasPermission = useMemo(() => {
    // 首先检查用户状态中的权限
    if (user?.permissions && user.permissions.length > 0) {
      return user.permissions.includes(permission)
    }

    // 然后检查API权限
    if (permissions?.permissions) {
      return permissions.permissions.includes(permission)
    }

    // 如果都没有，根据用户角色进行基本权限判断
    if (user?.roles) {
      // 管理员拥有所有权限
      if (user.roles.includes('admin')) {
        return true
      }

      // 律师的基本权限
      if (user.roles.includes('lawyer')) {
        const lawyerPermissions = [
          'dashboard:view',
          'client.view',
          'client.manage',
          'case.view',
          'case.manage',
          'file.view',
          'file.manage',
        ]
        return lawyerPermissions.includes(permission)
      }
    }

    return false
  }, [user, permissions?.permissions, permission])

  return {
    hasPermission,
    isLoading,
  }
}

// 工具Hook：检查是否有特定角色（优化版本，使用useMemo缓存结果）
export const useHasRole = (role: string) => {
  const { user } = useAppStore()
  const { data: permissions, isLoading } = usePermissions()

  const hasRole = useMemo(() => {
    // 首先检查用户状态中的角色
    if (user?.roles && user.roles.length > 0) {
      return user.roles.includes(role)
    }

    // 然后检查API角色
    if (permissions?.roles) {
      return permissions.roles.includes(role)
    }

    return false
  }, [user, permissions?.roles, role])

  return {
    hasRole,
    isLoading,
  }
}

// 自动登录Hook（用于页面刷新时恢复状态）
export const useAutoLogin = () => {
  const { login } = useAppStore()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      const token = getToken()
      const userInfo = getUserInfo()

      if (!token || !userInfo) {
        throw new Error('没有找到认证信息')
      }

      // 验证token是否仍然有效
      const response = await apiClient.get<User>('/api/v1/auth/verify-token')

      if (response.error) {
        throw new Error(response.error.message)
      }

      return { user: response.data, token }
    },
    onSuccess: ({ user, token }) => {
      login(user, token)
      queryClient.setQueryData(queryKeys.auth, {
        user,
        isAuthenticated: true,
      })
    },
    onError: () => {
      // Token无效，清除本地存储
      clearStorage()
    },
  })
}

export { queryKeys }
