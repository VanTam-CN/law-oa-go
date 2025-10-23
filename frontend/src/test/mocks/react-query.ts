/**
 * React Query Mock - 现代化服务器状态管理测试
 * 基于TanStack Query v5最佳实践
 */

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactNode } from 'react'

// Mock QueryClient工厂
export class MockQueryClientFactory {
  static create(override: any = {}): QueryClient {
    return new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          gcTime: 0,
          staleTime: 0,
          refetchOnWindowFocus: false,
          refetchOnReconnect: false,
          ...override.queries
        },
        mutations: {
          retry: false,
          ...override.mutations
        }
      },
      ...override
    })
  }

  static createWithSuccessData<T>(queryKey: string[], data: T): QueryClient {
    const queryClient = this.create()
    queryClient.setQueryData(queryKey, data)
    return queryClient
  }

  static createWithError(queryKey: string[], error: Error): QueryClient {
    const queryClient = this.create()
    queryClient.setQueryData(queryKey, undefined)
    queryClient.setQueryDefaults(queryKey, {
      queryFn: () => Promise.reject(error)
    })
    return queryClient
  }
}

// Mock QueryClientProvider组件
interface MockQueryProviderProps {
  children: ReactNode
  client?: QueryClient
}

export const MockQueryProvider: React.FC<MockQueryProviderProps> = ({
  children,
  client = MockQueryClientFactory.create()
}) => {
  return (
    <QueryClientProvider client={client}>
      {children}
    </QueryClientProvider>
  );
}

// 预设的查询状态Mock
export const mockQueryStates = {
  loading: {
    status: 'pending',
    fetchStatus: 'fetching',
    isLoading: true,
    isSuccess: false,
    isError: false,
    isPending: true
  },
  success: {
    status: 'success',
    fetchStatus: 'idle',
    isLoading: false,
    isSuccess: true,
    isError: false,
    isPending: false,
    data: undefined
  },
  error: {
    status: 'error',
    fetchStatus: 'idle',
    isLoading: false,
    isSuccess: false,
    isError: true,
    isPending: false,
    error: new Error('Query failed')
  }
}

// Hook Mock工厂
export const createMockQueryHook = <T>(
  state: 'loading' | 'success' | 'error',
  data?: T,
  error?: Error
) => {
  const baseState = mockQueryStates[state]

  return {
    ...baseState,
    data: state === 'success' ? data : undefined,
    error: state === 'error' ? error || new Error('Query failed') : undefined,
    refetch: jest.fn(),
    invalidateQueries: jest.fn(),
    prefetchQuery: jest.fn(),
    setQueryData: jest.fn(),
    getQueryData: jest.fn(),
    removeQueries: jest.fn(),
    resetQueries: jest.fn(),
    isFetching: false,
    isFetched: state === 'success',
    isFetchedAfterMount: state === 'success',
    fetchStatus: state === 'loading' ? 'fetching' : 'idle'
  }
}

// Mutation Mock工厂
export const createMockMutation = <T, V>(
  state: 'idle' | 'pending' | 'success' | 'error',
  data?: T,
  error?: Error
) => {
  return {
    mutate: jest.fn(),
    mutateAsync: jest.fn(),
    reset: jest.fn(),
    isIdle: state === 'idle',
    isPending: state === 'pending',
    isSuccess: state === 'success',
    isError: state === 'error',
    data: state === 'success' ? data : undefined,
    error: state === 'error' ? error : undefined,
    variables: undefined as V | undefined,
    submittedAt: undefined as number | undefined
  }
}

// 常用Hook Mocks
export const mockUseQuery = createMockQueryHook('success')
export const mockUseQueryLoading = createMockQueryHook('loading')
export const mockUseQueryError = createMockQueryHook('error', undefined, new Error('Query failed'))

export const mockUseMutation = createMockMutation('idle')
export const mockUseMutationLoading = createMockMutation('pending')
export const mockUseMutationSuccess = createMockMutation('success')
export const mockUseMutationError = createMockMutation('error', undefined, new Error('Mutation failed'))

// React Query DevTools Mock
export const MockReactQueryDevtools = () => null

export default {
  MockQueryClientFactory,
  MockQueryProvider,
  mockQueryStates,
  createMockQueryHook,
  createMockMutation,
  mockUseQuery,
  mockUseQueryLoading,
  mockUseQueryError,
  mockUseMutation,
  mockUseMutationLoading,
  mockUseMutationSuccess,
  mockUseMutationError,
  MockReactQueryDevtools
}