/**
 * React Query 配置 - 基于TanStack Query v5
 * 提供统一的数据获取、缓存和错误处理策略
 */

import {
  QueryClient,
  QueryClientProvider,
  useQuery,
  useMutation,
  UseQueryOptions,
  UseMutationOptions,
  QueryKey,
  MutationKey,
} from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import React, { ReactNode } from 'react'
import { ApiError, ApiResponse, PaginatedResponse } from '../services/apiClient'

// 创建QueryClient实例
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // 默认缓存时间：5分钟
      staleTime: 5 * 60 * 1000,
      // 默认垃圾回收时间：10分钟
      gcTime: 10 * 60 * 1000,
      // 默认重试次数
      retry: (failureCount, error) => {
        // 4xx错误不重试，5xx错误重试2次
        if (error instanceof ApiError && error.status && error.status >= 400 && error.status < 500) {
          return false
        }
        return failureCount < 2
      },
      // 默认重试延迟
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      // 窗口聚焦时重新获取
      refetchOnWindowFocus: false,
      // 网络重新连接时重新获取
      refetchOnReconnect: true,
    },
    mutations: {
      // 默认不重试mutation
      retry: false,
    },
  },
})

// React Query Provider组件
interface QueryProviderProps {
  children: ReactNode
  client?: QueryClient
}

export const QueryProvider: React.FC<QueryProviderProps> = ({ children, client = queryClient }) => {
  return (
    <QueryClientProvider client={client}>
      {children}
      {process.env.NODE_ENV === 'development' && (
        <ReactQueryDevtools initialIsOpen={false} />
      )}
    </QueryClientProvider>
  );
};

// 自定义查询Hook
export function useApiQuery<T>(
  queryKey: QueryKey,
  queryFn: () => Promise<ApiResponse<T>>,
  options?: Omit<UseQueryOptions<ApiResponse<T>, ApiError, T>, 'queryKey' | 'queryFn'>
) {
  return useQuery<ApiResponse<T>, ApiError, T>({
    queryKey,
    queryFn,
    select: (data) => data.data,
    ...options,
  })
}

// 分页查询Hook
export function usePaginatedQuery<T>(
  queryKey: QueryKey,
  queryFn: () => Promise<ApiResponse<PaginatedResponse<T>>>,
  options?: Omit<UseQueryOptions<ApiResponse<PaginatedResponse<T>>, ApiError, PaginatedResponse<T>>, 'queryKey' | 'queryFn'>
) {
  return useQuery<ApiResponse<PaginatedResponse<T>>, ApiError, PaginatedResponse<T>>({
    queryKey,
    queryFn,
    select: (data) => data.data,
    ...options,
  })
}

// 无限查询Hook（用于无限滚动）
export function useInfiniteQuery<T>(
  queryKey: QueryKey,
  queryFn: ({ pageParam }: { pageParam: number }) => Promise<ApiResponse<PaginatedResponse<T>>>,
  options?: Omit<
    UseQueryOptions<ApiResponse<PaginatedResponse<T>>, ApiError, PaginatedResponse<T>>,
    'queryKey' | 'queryFn'
  >
) {
  return useQuery({
    queryKey,
    queryFn: () => queryFn({ pageParam: 1 }),
    select: (data) => data.data,
    ...options,
  })
}

// 自定义Mutation Hook
export function useApiMutation<TData, TVariables>(
  mutationFn: (variables: TVariables) => Promise<ApiResponse<TData>>,
  options?: UseMutationOptions<ApiResponse<TData>, ApiError, TVariables, TData>
) {
  return useMutation({
    mutationFn,
    onSuccess: (data) => {
      console.log('[Mutation Success]', data)
    },
    onError: (error) => {
      console.error('[Mutation Error]', error)
    },
    select: (data) => data.data,
    ...options,
  })
}

// 乐观更新Mutation Hook
export function useOptimisticMutation<TData, TVariables>(
  mutationFn: (variables: TVariables) => Promise<ApiResponse<TData>>,
  updateQuery: (variables: TVariables, previous: TData | undefined) => TData | undefined,
  options?: UseMutationOptions<ApiResponse<TData>, ApiError, TVariables, TData>
) {
  return useMutation({
    mutationFn,
    onMutate: async (variables) => {
      // 取消正在进行的查询
      await queryClient.cancelQueries()

      // 保存之前的数据
      const previousData = queryClient.getQueryData(['api-data'])

      // 乐观更新
      if (updateQuery) {
        queryClient.setQueryData(['api-data'], (old: any) =>
          updateQuery(variables, old)
        )
      }

      return { previousData }
    },
    onError: (error, variables, context) => {
      // 回滚到之前的数据
      if (context?.previousData) {
        queryClient.setQueryData(['api-data'], context.previousData)
      }
    },
    onSettled: () => {
      // 重新获取数据
      queryClient.invalidateQueries({ queryKey: ['api-data'] })
    },
    ...options,
  })
}

// 批量操作Mutation Hook
export function useBatchMutation<TData, TVariables>(
  mutations: Array<{
    key: MutationKey
    fn: (variables: TVariables) => Promise<ApiResponse<TData>>
  }>,
  options?: UseMutationOptions<
    ApiResponse<TData>[],
    ApiError,
    TVariables,
    ApiResponse<TData>[]
  >
) {
  return useMutation({
    mutationFn: async (variables: TVariables) => {
      const results = await Promise.allSettled(
        mutations.map(({ fn }) => fn(variables))
      )

      return results.map((result, index) => {
        if (result.status === 'fulfilled') {
          return result.value
        } else {
          throw new Error(`批量操作失败 (${index + 1}/${mutations.length})`)
        }
      })
    },
    ...options,
  })
}

// 缓存操作工具函数
export const queryUtils = {
  // 设置缓存数据
  setData: <T>(queryKey: QueryKey, data: T) => {
    queryClient.setQueryData(queryKey, data)
  },

  // 获取缓存数据
  getData: <T>(queryKey: QueryKey): T | undefined => {
    return queryClient.getQueryData<T>(queryKey)
  },

  // 移除缓存数据
  removeData: (queryKey: QueryKey) => {
    queryClient.removeQueries({ queryKey })
  },

  // 使缓存失效
  invalidate: (queryKey: QueryKey) => {
    queryClient.invalidateQueries({ queryKey })
  },

  // 预取数据
  prefetch: async <T>(
    queryKey: QueryKey,
    queryFn: () => Promise<T>,
    options?: UseQueryOptions<T>
  ) => {
    return queryClient.prefetchQuery(queryKey, queryFn, options)
  },

  // 重置缓存
  reset: () => {
    queryClient.resetQueries()
  },

  // 清除缓存
  clear: () => {
    queryClient.clear()
  },
}

// 查询工厂函数
export const createQuery = <T>(
  key: string[],
  fetcher: (params?: any) => Promise<ApiResponse<T>>
) => ({
  useQuery: (params?: any, options?: UseQueryOptions<ApiResponse<T>, ApiError, T>) =>
    useApiQuery([key, params], () => fetcher(params), options),

  usePrefetch: (params?: any, options?: UseQueryOptions<ApiResponse<T>, ApiError, T>) =>
    queryUtils.prefetch([key, params], () => fetcher(params), options),

  invalidate: (params?: any) => queryUtils.invalidate([key, params]),

  remove: (params?: any) => queryUtils.removeData([key, params]),

  getData: (params?: any) => queryUtils.getData<T>([key, params]),

  setData: (params: any, data: T) => queryUtils.setData([key, params], data),
})

// Mutation工厂函数
export const createMutation = <TData, TVariables>(
  key: string[],
  mutator: (variables: TVariables) => Promise<ApiResponse<TData>>
) => ({
  useMutation: (options?: UseMutationOptions<ApiResponse<TData>, ApiError, TVariables, TData>) =>
    useApiMutation(mutator, options),

  useOptimisticMutation: (
    updateFn: (variables: TVariables, previous: TData | undefined) => TData | undefined,
    options?: UseMutationOptions<ApiResponse<TData>, ApiError, TVariables, TData>
  ) => useOptimisticMutation(mutator, updateFn, options),

  invalidateRelated: () => {
    // 自动使相关查询失效
    queryClient.invalidateQueries({ predicate: (query) => query.queryKey[0] === key[0] })
  },
})

// 导出类型
export type {
  QueryKey,
  MutationKey,
  UseQueryOptions,
  UseMutationOptions,
}