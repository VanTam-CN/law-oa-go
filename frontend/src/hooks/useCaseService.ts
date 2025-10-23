/**
 * 案件管理服务钩子 - 基于TanStack Query v5
 * 提供案件CRUD、搜索、过滤等完整的业务逻辑
 */

import { useMutation, useQuery, useQueryClient, useInfiniteQuery } from '@tanstack/react-query'
import { apiClient } from '../services/apiClient'
import { useCaseStore, Case, CaseFilter, CaseSort } from '../stores/useCaseStore'

// 案件相关请求接口
export interface CreateCaseRequest {
  title: string
  description: string
  caseTypeId: string
  clientId: string
  priority: string
  assignedLawyer: string
  assistant?: string
  expectedEndDate?: string
  amount?: number
  currency?: string
  tags?: string[]
}

export interface UpdateCaseRequest extends Partial<CreateCaseRequest> {
  id: string
  status?: string
  progress?: number
  actualEndDate?: string
}

export interface CaseListRequest {
  page?: number
  pageSize?: number
  filter?: CaseFilter
  sort?: CaseSort
}

export interface CaseListResponse {
  cases: Case[]
  pagination: {
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
  statistics: {
    total: number
    byStatus: Record<string, number>
    byPriority: Record<string, number>
    avgProgress: number
  }
}

export interface CaseSearchParams {
  keyword?: string
  status?: string[]
  priority?: string[]
  caseType?: string[]
  assignedLawyer?: string[]
  client?: string
  dateRange?: {
    start: string
    end: string
  }
  tags?: string[]
  amountRange?: {
    min?: number
    max?: number
  }
}

// 查询键
const queryKeys = {
  cases: ['cases'] as const,
  caseList: ['cases', 'list'] as const,
  caseDetail: (id: string) => ['cases', 'detail', id] as const,
  caseStatistics: ['cases', 'statistics'] as const,
  caseTypes: ['cases', 'types'] as const,
  clients: ['cases', 'clients'] as const,
  caseSearch: (params: CaseSearchParams) => ['cases', 'search', params] as const,
}

// 获取案件列表Hook
export const useCaseList = (params: CaseListRequest = {}) => {
  const {
    page = 1,
    pageSize = 20,
    filter = {},
    sort = { field: 'createdAt', direction: 'desc' }
  } = params

  return useQuery({
    queryKey: [...queryKeys.caseList, { page, pageSize, filter, sort }],
    queryFn: async () => {
      const response = await apiClient.get<CaseListResponse>('/api/v1/cases', {
        params: {
          page,
          pageSize,
          ...filter,
          sortBy: sort.field,
          sortOrder: sort.direction,
        },
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    staleTime: 2 * 60 * 1000, // 2分钟
    gcTime: 5 * 60 * 1000,    // 5分钟
    keepPreviousData: true,    // 保持之前的数据，提供更好的用户体验
  })
}

// 无限滚动案件列表Hook
export const useInfiniteCaseList = (params: Omit<CaseListRequest, 'page'> = {}) => {
  const { pageSize = 20, filter = {}, sort = { field: 'createdAt', direction: 'desc' } } = params

  return useInfiniteQuery({
    queryKey: [...queryKeys.caseList, 'infinite', { pageSize, filter, sort }],
    queryFn: async ({ pageParam = 1 }) => {
      const response = await apiClient.get<CaseListResponse>('/api/v1/cases', {
        params: {
          page: pageParam,
          pageSize,
          ...filter,
          sortBy: sort.field,
          sortOrder: sort.direction,
        },
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    getNextPageParam: (lastPage) => {
      if (lastPage.pagination.page < lastPage.pagination.totalPages) {
        return lastPage.pagination.page + 1
      }
      return undefined
    },
    staleTime: 2 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  })
}

// 获取案件详情Hook
export const useCaseDetail = (caseId: string) => {
  const queryClient = useQueryClient()

  return useQuery({
    queryKey: queryKeys.caseDetail(caseId),
    queryFn: async () => {
      const response = await apiClient.get<Case>(`/api/v1/cases/${caseId}`)

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    enabled: !!caseId,
    staleTime: 1 * 60 * 1000, // 1分钟
    gcTime: 3 * 60 * 1000,    // 3分钟
    onSuccess: (data) => {
      // 预加载相关数据
      queryClient.prefetchQuery({
        queryKey: queryKeys.caseDetail(data.id),
        queryFn: () => apiClient.get<Case>(`/api/v1/cases/${data.id}`),
        staleTime: 1 * 60 * 1000,
      })
    },
  })
}

// 创建案件Hook
export const useCreateCase = () => {
  const queryClient = useQueryClient()
  const { addCase } = useCaseStore()

  return useMutation({
    mutationFn: async (caseData: CreateCaseRequest) => {
      const response = await apiClient.post<Case>('/api/v1/cases', caseData)

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (newCase) => {
      // 更新本地状态
      addCase(newCase)

      // 更新React Query缓存
      queryClient.setQueryData(queryKeys.caseDetail(newCase.id), newCase)

      // 使列表查询失效以获取最新数据
      queryClient.invalidateQueries({ queryKey: queryKeys.caseList })
      queryClient.invalidateQueries({ queryKey: queryKeys.caseStatistics })

      console.log('案件创建成功:', newCase.caseNumber)
    },
    onError: (error) => {
      console.error('案件创建失败:', error.message)
    },
  })
}

// 更新案件Hook
export const useUpdateCase = () => {
  const queryClient = useQueryClient()
  const { updateCase } = useCaseStore()

  return useMutation({
    mutationFn: async ({ id, ...updates }: UpdateCaseRequest) => {
      const response = await apiClient.put<Case>(`/api/v1/cases/${id}`, updates)

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onMutate: async ({ id, ...updates }) => {
      // 取消任何正在进行的查询
      await queryClient.cancelQueries({ queryKey: queryKeys.caseDetail(id) })

      // 快照当前值
      const previousCase = queryClient.getQueryData(queryKeys.caseDetail(id))

      // 乐观更新
      queryClient.setQueryData(queryKeys.caseDetail(id), (old: Case) =>
        old ? { ...old, ...updates, updatedAt: new Date().toISOString() } : { id, ...updates }
      )

      // 更新本地状态
      updateCase(id, updates)

      return { previousCase }
    },
    onError: (error, variables, context) => {
      // 恢复之前的值
      if (context?.previousCase) {
        queryClient.setQueryData(queryKeys.caseDetail(variables.id), context.previousCase)
      }
      console.error('案件更新失败:', error.message)
    },
    onSettled: (data, error, variables) => {
      // 重新获取数据
      queryClient.invalidateQueries({ queryKey: queryKeys.caseDetail(variables.id) })
      queryClient.invalidateQueries({ queryKey: queryKeys.caseList })
      queryClient.invalidateQueries({ queryKey: queryKeys.caseStatistics })
    },
  })
}

// 删除案件Hook
export const useDeleteCase = () => {
  const queryClient = useQueryClient()
  const { deleteCase } = useCaseStore()

  return useMutation({
    mutationFn: async (caseId: string) => {
      const response = await apiClient.delete<{ message: string }>(`/api/v1/cases/${caseId}`)

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (_, caseId) => {
      // 更新本地状态
      deleteCase(caseId)

      // 从缓存中移除
      queryClient.removeQueries({ queryKey: queryKeys.caseDetail(caseId) })

      // 使列表查询失效
      queryClient.invalidateQueries({ queryKey: queryKeys.caseList })
      queryClient.invalidateQueries({ queryKey: queryKeys.caseStatistics })

      console.log('案件删除成功:', caseId)
    },
    onError: (error) => {
      console.error('案件删除失败:', error.message)
    },
  })
}

// 批量删除案件Hook
export const useBulkDeleteCases = () => {
  const queryClient = useQueryClient()
  const { deleteSelectedCases } = useCaseStore()

  return useMutation({
    mutationFn: async (caseIds: string[]) => {
      const response = await apiClient.post<{ message: string }>('/api/v1/cases/bulk-delete', {
        caseIds,
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: () => {
      // 更新本地状态
      deleteSelectedCases()

      // 使所有相关查询失效
      queryClient.invalidateQueries({ queryKey: queryKeys.cases })

      console.log('批量删除案件成功')
    },
    onError: (error) => {
      console.error('批量删除案件失败:', error.message)
    },
  })
}

// 案件搜索Hook
export const useCaseSearch = (searchParams: CaseSearchParams) => {
  return useQuery({
    queryKey: queryKeys.caseSearch(searchParams),
    queryFn: async () => {
      const response = await apiClient.get<Case[]>('/api/v1/cases/search', {
        params: searchParams,
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    enabled: Object.keys(searchParams).length > 0,
    staleTime: 1 * 60 * 1000, // 1分钟
    gcTime: 3 * 60 * 1000,    // 3分钟
  })
}

// 获取案件统计信息Hook
export const useCaseStatistics = () => {
  return useQuery({
    queryKey: queryKeys.caseStatistics,
    queryFn: async () => {
      const response = await apiClient.get<{
        total: number
        byStatus: Record<string, number>
        byPriority: Record<string, number>
        byType: Record<string, number>
        avgProgress: number
        recentActivity: Array<{
          caseId: string
          action: string
          timestamp: string
        }>
      }>('/api/v1/cases/statistics')

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    staleTime: 5 * 60 * 1000, // 5分钟
    gcTime: 10 * 60 * 1000,   // 10分钟
  })
}

// 获取案件类型Hook
export const useCaseTypes = () => {
  return useQuery({
    queryKey: queryKeys.caseTypes,
    queryFn: async () => {
      const response = await apiClient.get<Array<{
        id: string
        name: string
        code: string
        description?: string
        color?: string
      }>>('/api/v1/cases/types')

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    staleTime: 30 * 60 * 1000, // 30分钟
    gcTime: 60 * 60 * 1000,    // 1小时
  })
}

// 获取客户列表Hook
export const useClients = () => {
  return useQuery({
    queryKey: queryKeys.clients,
    queryFn: async () => {
      const response = await apiClient.get<Array<{
        id: string
        name: string
        email?: string
        phone?: string
        type: 'individual' | 'corporate'
        contactPerson?: string
      }>>('/api/v1/cases/clients')

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    staleTime: 15 * 60 * 1000, // 15分钟
    gcTime: 30 * 60 * 1000,    // 30分钟
  })
}

// 更新案件进度Hook
export const useUpdateCaseProgress = () => {
  const queryClient = useQueryClient()
  const { updateCaseProgress } = useCaseStore()

  return useMutation({
    mutationFn: async ({ caseId, progress }: { caseId: string; progress: number }) => {
      const response = await apiClient.patch<Case>(`/api/v1/cases/${caseId}/progress`, {
        progress,
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onMutate: async ({ caseId, progress }) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.caseDetail(caseId) })

      const previousCase = queryClient.getQueryData(queryKeys.caseDetail(caseId))

      queryClient.setQueryData(queryKeys.caseDetail(caseId), (old: Case) =>
        old ? { ...old, progress, updatedAt: new Date().toISOString() } : old
      )

      updateCaseProgress(caseId, progress)

      return { previousCase }
    },
    onError: (error, variables, context) => {
      if (context?.previousCase) {
        queryClient.setQueryData(queryKeys.caseDetail(variables.caseId), context.previousCase)
      }
      console.error('更新案件进度失败:', error.message)
    },
    onSettled: (data, error, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.caseDetail(variables.caseId) })
      queryClient.invalidateQueries({ queryKey: queryKeys.caseStatistics })
    },
  })
}

// 添加案件备注Hook
export const useAddCaseNote = () => {
  const queryClient = useQueryClient()
  const { addNote } = useCaseStore()

  return useMutation({
    mutationFn: async ({
      caseId,
      content,
      type = 'note',
      isPrivate = false,
    }: {
      caseId: string
      content: string
      type?: 'note' | 'reminder' | 'milestone'
      isPrivate?: boolean
    }) => {
      const response = await apiClient.post<{
        id: string
        content: string
        type: 'note' | 'reminder' | 'milestone'
        isPrivate: boolean
        createdAt: string
        createdBy: string
      }>(`/api/v1/cases/${caseId}/notes`, {
        content,
        type,
        isPrivate,
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (newNote, variables) => {
      addNote(variables.caseId, newNote)
      queryClient.invalidateQueries({ queryKey: queryKeys.caseDetail(variables.caseId) })
    },
    onError: (error) => {
      console.error('添加案件备注失败:', error.message)
    },
  })
}

// 上传案件文档Hook
export const useUploadCaseDocument = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      caseId,
      file,
      category = 'general',
      isPublic = false,
    }: {
      caseId: string
      file: File
      category?: string
      isPublic?: boolean
    }) => {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('category', category)
      formData.append('isPublic', isPublic.toString())

      const response = await apiClient.post<{
        id: string
        name: string
        type: string
        size: number
        url: string
        uploadedAt: string
        uploadedBy: string
        category: string
        isPublic: boolean
      }>(`/api/v1/cases/${caseId}/documents`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })

      if (response.error) {
        throw new Error(response.error.message)
      }

      return response.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.caseDetail(variables.caseId) })
    },
    onError: (error) => {
      console.error('上传案件文档失败:', error.message)
    },
  })
}

export {
  queryKeys,
}