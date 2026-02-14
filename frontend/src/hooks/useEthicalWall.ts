/**
 * 隔离墙管理 API Hooks
 * 基于 TanStack Query v5
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '../services/apiClient'
import {
  EthicalWall,
  WhitelistResponse,
  WhitelistUser,
  AddWhitelistRequest,
  UserOption,
} from '../types/ethicalWall'

// 查询键
const queryKeys = {
  ethicalWall: (caseId: string) => ['ethical-wall', caseId] as const,
  whitelist: (caseId: string) => ['ethical-wall', caseId, 'whitelist'] as const,
  users: ['users'] as const,
}

/**
 * 获取案件隔离墙状态
 */
export const useEthicalWall = (caseId: string) => {
  return useQuery({
    queryKey: queryKeys.ethicalWall(caseId),
    queryFn: async () => {
      const response = await apiClient.get<EthicalWall>(`/cases/${caseId}/ethical-wall`)

      if (response.error) {
        // 如果隔离墙未启用，返回默认状态
        if (response.error.code === 'ETHICAL_WALL_NOT_ENABLED') {
          return {
            caseId,
            enabled: false,
          } as EthicalWall
        }
        throw new Error(response.error.message)
      }

      return response.data
    },
    enabled: !!caseId,
    staleTime: 1 * 60 * 1000, // 1分钟
    gcTime: 3 * 60 * 1000, // 3分钟
    retry: false,
  })
}

/**
 * 获取白名单列表
 */
export const useWhitelist = (caseId: string, enabled = true) => {
  return useQuery({
    queryKey: queryKeys.whitelist(caseId),
    queryFn: async () => {
      const response = await apiClient.get<WhitelistResponse>(
        `/cases/${caseId}/ethical-wall/whitelist`,
      )

      if (response.error) {
        // 如果隔离墙未启用，返回空列表
        if (response.error.code === 'ETHICAL_WALL_NOT_ENABLED') {
          return {
            caseId,
            users: [],
            total: 0,
          } as WhitelistResponse
        }
        throw new Error(response.error.message)
      }

      return response.data
    },
    enabled: !!caseId && enabled,
    staleTime: 1 * 60 * 1000,
    gcTime: 3 * 60 * 1000,
    retry: false,
  })
}

/**
 * 启用隔离墙
 */
export const useEnableEthicalWall = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (caseId: string) => {
      const response = await apiClient.post<EthicalWall>(
        `/cases/${caseId}/ethical-wall`,
        {},
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (data, caseId) => {
      // 更新隔离墙状态缓存
      queryClient.setQueryData(queryKeys.ethicalWall(caseId), data)

      // 使相关查询失效
      queryClient.invalidateQueries({ queryKey: queryKeys.ethicalWall(caseId) })
    },
  })
}

/**
 * 禁用隔离墙
 */
export const useDisableEthicalWall = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (caseId: string) => {
      const response = await apiClient.delete<{ message: string }>(
        `/cases/${caseId}/ethical-wall`,
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (_, caseId) => {
      // 更新隔离墙状态为禁用
      queryClient.setQueryData(queryKeys.ethicalWall(caseId), {
        caseId,
        enabled: false,
      })

      // 清空白名单缓存
      queryClient.setQueryData(queryKeys.whitelist(caseId), {
        caseId,
        users: [],
        total: 0,
      })

      // 使相关查询失效
      queryClient.invalidateQueries({ queryKey: queryKeys.ethicalWall(caseId) })
      queryClient.invalidateQueries({ queryKey: queryKeys.whitelist(caseId) })
    },
  })
}

/**
 * 添加用户到白名单
 */
export const useAddWhitelist = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ caseId, userId, reason }: AddWhitelistRequest & { caseId: string }) => {
      const response = await apiClient.post<WhitelistUser>(
        `/cases/${caseId}/ethical-wall/whitelist`,
        { userId, reason },
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (newUser, variables) => {
      const { caseId } = variables

      // 更新白名单缓存
      queryClient.setQueryData(
        queryKeys.whitelist(caseId),
        (old: WhitelistResponse | undefined) => {
          if (!old) {
            return {
              caseId,
              users: [newUser],
              total: 1,
            }
          }
          return {
            ...old,
            users: [...old.users, newUser],
            total: old.total + 1,
          }
        },
      )

      // 使查询失效以获取最新数据
      queryClient.invalidateQueries({ queryKey: queryKeys.whitelist(caseId) })
    },
  })
}

/**
 * 从白名单移除用户
 */
export const useRemoveWhitelist = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ caseId, userId }: { caseId: string; userId: string }) => {
      const response = await apiClient.delete<{ message: string }>(
        `/cases/${caseId}/ethical-wall/whitelist/${userId}`,
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (_, variables) => {
      const { caseId, userId } = variables

      // 更新白名单缓存
      queryClient.setQueryData(
        queryKeys.whitelist(caseId),
        (old: WhitelistResponse | undefined) => {
          if (!old) {
            return {
              caseId,
              users: [],
              total: 0,
            }
          }
          return {
            ...old,
            users: old.users.filter((u) => u.userId !== userId),
            total: Math.max(0, old.total - 1),
          }
        },
      )

      // 使查询失效以获取最新数据
      queryClient.invalidateQueries({ queryKey: queryKeys.whitelist(caseId) })
    },
  })
}

/**
 * 搜索用户（用于添加白名单）
 */
export const useUserSearch = (keyword: string, enabled = true) => {
  return useQuery({
    queryKey: ['users', 'search', keyword],
    queryFn: async () => {
      const response = await apiClient.get<UserOption[]>('/users/search', {
        params: { keyword },
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    enabled: enabled && keyword.length >= 2,
    staleTime: 5 * 60 * 1000, // 5分钟
    gcTime: 10 * 60 * 1000, // 10分钟
  })
}

export { queryKeys }
