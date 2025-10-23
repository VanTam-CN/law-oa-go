/**
 * 案件管理状态 - 基于Zustand v5
 * 管理案件相关的复杂状态和业务逻辑
 */

import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

// 案件状态枚举
export enum CaseStatus {
  DRAFT = 'draft',
  ACTIVE = 'active',
  IN_PROGRESS = 'in_progress',
  COMPLETED = 'completed',
  CLOSED = 'closed',
  ARCHIVED = 'archived'
}

export enum CasePriority {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  URGENT = 'urgent'
}

// 案件类型
export interface CaseType {
  id: string
  name: string
  code: string
  description?: string
  color?: string
}

// 客户信息
export interface Client {
  id: string
  name: string
  email?: string
  phone?: string
  address?: string
  type: 'individual' | 'corporate'
  contactPerson?: string
}

// 案件实体
export interface Case {
  id: string
  caseNumber: string
  title: string
  description: string
  caseType: CaseType
  client: Client
  status: CaseStatus
  priority: CasePriority
  assignedLawyer: string
  assistant?: string
  startDate: string
  expectedEndDate?: string
  actualEndDate?: string
  tags: string[]
  documents: Document[]
  conflicts: Conflict[]
  amount?: number
  currency?: string
  progress: number // 0-100
  notes: CaseNote[]
  createdAt: string
  updatedAt: string
  createdBy: string
  updatedBy: string
}

// 文档信息
export interface Document {
  id: string
  name: string
  type: string
  size: number
  url: string
  uploadedAt: string
  uploadedBy: string
  category: string
  isPublic: boolean
}

// 冲突记录
export interface Conflict {
  id: string
  title: string
  description: string
  severity: 'low' | 'medium' | 'high'
  status: 'detected' | 'resolved' | 'ignored'
  detectedAt: string
  resolvedAt?: string
  relatedDocuments: string[]
  aiConfidence: number
}

// 案件备注
export interface CaseNote {
  id: string
  content: string
  type: 'note' | 'reminder' | 'milestone'
  isPrivate: boolean
  createdAt: string
  createdBy: string
}

// 搜索过滤器
export interface CaseFilter {
  keyword?: string
  status?: CaseStatus[]
  priority?: CasePriority[]
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

// 排序配置
export interface CaseSort {
  field: keyof Case
  direction: 'asc' | 'desc'
}

// 案件管理状态接口
export interface CaseStoreState {
  // 数据状态
  cases: Case[]
  currentCase: Case | null
  caseTypes: CaseType[]
  clients: Client[]

  // UI状态
  isLoading: boolean
  isCreating: boolean
  isUpdating: boolean
  isDeleting: boolean

  // 分页和过滤
  pagination: {
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
  filter: CaseFilter
  sort: CaseSort

  // 临时状态
  selectedCaseIds: string[]
  editingCase: Case | null

  // Actions
  setCases: (cases: Case[]) => void
  addCase: (case: Case) => void
  updateCase: (id: string, updates: Partial<Case>) => void
  deleteCase: (id: string) => void
  setCurrentCase: (case: Case | null) => void
  updateCaseProgress: (id: string, progress: number) => void

  // 批量操作
  selectAllCases: () => void
  deselectAllCases: () => void
  toggleCaseSelection: (id: string) => void
  deleteSelectedCases: () => void

  // 过滤和排序
  setFilter: (filter: Partial<CaseFilter>) => void
  setSort: (sort: CaseSort) => void
  applyFilter: () => Case[]
  clearFilter: () => void

  // 分页
  setPagination: (pagination: Partial<CaseStoreState['pagination']>) => void
  nextPage: () => void
  prevPage: () => void
  setPageSize: (pageSize: number) => void

  // 加载状态
  setLoading: (loading: boolean) => void
  setCreating: (creating: boolean) => void
  setUpdating: (updating: boolean) => void
  setDeleting: (deleting: boolean) => void

  // 编辑模式
  startEditing: (case: Case) => void
  cancelEditing: () => void
  saveEditing: () => void

  // 业务操作
  addDocument: (caseId: string, document: Document) => void
  removeDocument: (caseId: string, documentId: string) => void
  addNote: (caseId: string, note: CaseNote) => void
  updateConflictStatus: (caseId: string, conflictId: string, status: Conflict['status']) => void

  // 统计数据
  getStatistics: () => {
    total: number
    byStatus: Record<CaseStatus, number>
    byPriority: Record<CasePriority, number>
    byType: Record<string, number>
    avgProgress: number
  }

  // 重置状态
  reset: () => void
}

// 创建案件管理Store
export const useCaseStore = create<CaseStoreState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      cases: [],
      currentCase: null,
      caseTypes: [],
      clients: [],
      isLoading: false,
      isCreating: false,
      isUpdating: false,
      isDeleting: false,
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0,
        totalPages: 0,
      },
      filter: {},
      sort: { field: 'createdAt', direction: 'desc' },
      selectedCaseIds: [],
      editingCase: null,

      // 基础操作
      setCases: (cases) => set({ cases }),

      addCase: (newCase) =>
        set((state) => ({
          cases: [newCase, ...state.cases],
          pagination: {
            ...state.pagination,
            total: state.pagination.total + 1,
          },
        })),

      updateCase: (id, updates) =>
        set((state) => ({
          cases: state.cases.map((case_) =>
            case_.id === id ? { ...case_, ...updates, updatedAt: new Date().toISOString() } : case_
          ),
          currentCase:
            state.currentCase?.id === id
              ? { ...state.currentCase, ...updates, updatedAt: new Date().toISOString() }
              : state.currentCase,
        })),

      deleteCase: (id) =>
        set((state) => ({
          cases: state.cases.filter((case_) => case_.id !== id),
          currentCase: state.currentCase?.id === id ? null : state.currentCase,
          selectedCaseIds: state.selectedCaseIds.filter((id_) => id_ !== id),
          pagination: {
            ...state.pagination,
            total: state.pagination.total - 1,
          },
        })),

      setCurrentCase: (case_) => set({ currentCase: case_ }),

      updateCaseProgress: (id, progress) =>
        set((state) => ({
          cases: state.cases.map((case_) =>
            case_.id === id ? { ...case_, progress, updatedAt: new Date().toISOString() } : case_
          ),
          currentCase:
            state.currentCase?.id === id
              ? { ...state.currentCase, progress, updatedAt: new Date().toISOString() }
              : state.currentCase,
        })),

      // 批量操作
      selectAllCases: () =>
        set((state) => ({
          selectedCaseIds: state.cases.map((case_) => case_.id),
        })),

      deselectAllCases: () => set({ selectedCaseIds: [] }),

      toggleCaseSelection: (id) =>
        set((state) => ({
          selectedCaseIds: state.selectedCaseIds.includes(id)
            ? state.selectedCaseIds.filter((id_) => id_ !== id)
            : [...state.selectedCaseIds, id],
        })),

      deleteSelectedCases: () =>
        set((state) => {
          const remainingCases = state.cases.filter(
            (case_) => !state.selectedCaseIds.includes(case_.id)
          )
          const deletedCount = state.cases.length - remainingCases.length
          return {
            ...state,
            cases: remainingCases,
            selectedCaseIds: [],
            currentCase: state.selectedCaseIds.includes(state.currentCase?.id || '')
              ? null
              : state.currentCase,
            pagination: {
              ...state.pagination,
              total: state.pagination.total - deletedCount,
            },
          }
        }),

      // 过滤和排序
      setFilter: (newFilter) =>
        set((state) => ({
          filter: { ...state.filter, ...newFilter },
        })),

      setSort: (sort) => set({ sort }),

      applyFilter: () => {
        const state = get()
        let filteredCases = [...state.cases]

        // 关键词搜索
        if (state.filter.keyword) {
          const keyword = state.filter.keyword.toLowerCase()
          filteredCases = filteredCases.filter(
            (case_) =>
              case_.title.toLowerCase().includes(keyword) ||
              case_.description.toLowerCase().includes(keyword) ||
              case_.caseNumber.toLowerCase().includes(keyword) ||
              case_.client.name.toLowerCase().includes(keyword)
          )
        }

        // 状态过滤
        if (state.filter.status?.length) {
          filteredCases = filteredCases.filter((case_) =>
            state.filter.status!.includes(case_.status)
          )
        }

        // 优先级过滤
        if (state.filter.priority?.length) {
          filteredCases = filteredCases.filter((case_) =>
            state.filter.priority!.includes(case_.priority)
          )
        }

        // 案件类型过滤
        if (state.filter.caseType?.length) {
          filteredCases = filteredCases.filter((case_) =>
            state.filter.caseType!.includes(case_.caseType.id)
          )
        }

        // 律师过滤
        if (state.filter.assignedLawyer?.length) {
          filteredCases = filteredCases.filter((case_) =>
            state.filter.assignedLawyer!.includes(case_.assignedLawyer)
          )
        }

        // 客户过滤
        if (state.filter.client) {
          filteredCases = filteredCases.filter(
            (case_) => case_.client.id === state.filter.client
          )
        }

        // 日期范围过滤
        if (state.filter.dateRange) {
          const { start, end } = state.filter.dateRange
          filteredCases = filteredCases.filter((case_) => {
            const caseDate = new Date(case_.startDate)
            const startDate = new Date(start)
            const endDate = new Date(end)
            return caseDate >= startDate && caseDate <= endDate
          })
        }

        // 标签过滤
        if (state.filter.tags?.length) {
          filteredCases = filteredCases.filter((case_) =>
            state.filter.tags!.some((tag) => case_.tags.includes(tag))
          )
        }

        // 排序
        filteredCases.sort((a, b) => {
          const { field, direction } = state.sort
          let aValue = a[field]
          let bValue = b[field]

          if (typeof aValue === 'string') {
            aValue = aValue.toLowerCase()
            bValue = (bValue as string).toLowerCase()
          }

          if (direction === 'asc') {
            return aValue > bValue ? 1 : -1
          } else {
            return aValue < bValue ? 1 : -1
          }
        })

        return filteredCases
      },

      clearFilter: () => set({ filter: {}, sort: { field: 'createdAt', direction: 'desc' } }),

      // 分页
      setPagination: (pagination) =>
        set((state) => ({
          pagination: { ...state.pagination, ...pagination },
        })),

      nextPage: () =>
        set((state) => ({
          pagination: {
            ...state.pagination,
            page: Math.min(state.pagination.page + 1, state.pagination.totalPages),
          },
        })),

      prevPage: () =>
        set((state) => ({
          pagination: {
            ...state.pagination,
            page: Math.max(state.pagination.page - 1, 1),
          },
        })),

      setPageSize: (pageSize) =>
        set((state) => ({
          pagination: {
            ...state.pagination,
            pageSize,
            page: 1,
            totalPages: Math.ceil(state.pagination.total / pageSize),
          },
        })),

      // 加载状态
      setLoading: (loading) => set({ isLoading: loading }),
      setCreating: (creating) => set({ isCreating: creating }),
      setUpdating: (updating) => set({ isUpdating: updating }),
      setDeleting: (deleting) => set({ isDeleting: deleting }),

      // 编辑模式
      startEditing: (case_) => set({ editingCase: { ...case_ } }),
      cancelEditing: () => set({ editingCase: null }),
      saveEditing: () => {
        const { editingCase, updateCase } = get()
        if (editingCase) {
          updateCase(editingCase.id, editingCase)
          set({ editingCase: null })
        }
      },

      // 业务操作
      addDocument: (caseId, document) =>
        set((state) => ({
          cases: state.cases.map((case_) =>
            case_.id === caseId
              ? { ...case_, documents: [...case_.documents, document] }
              : case_
          ),
          currentCase:
            state.currentCase?.id === caseId
              ? { ...state.currentCase, documents: [...state.currentCase.documents, document] }
              : state.currentCase,
        })),

      removeDocument: (caseId, documentId) =>
        set((state) => ({
          cases: state.cases.map((case_) =>
            case_.id === caseId
              ? { ...case_, documents: case_.documents.filter((doc) => doc.id !== documentId) }
              : case_
          ),
          currentCase:
            state.currentCase?.id === caseId
              ? {
                  ...state.currentCase,
                  documents: state.currentCase.documents.filter((doc) => doc.id !== documentId),
                }
              : state.currentCase,
        })),

      addNote: (caseId, note) =>
        set((state) => ({
          cases: state.cases.map((case_) =>
            case_.id === caseId
              ? { ...case_, notes: [...case_.notes, note] }
              : case_
          ),
          currentCase:
            state.currentCase?.id === caseId
              ? { ...state.currentCase, notes: [...state.currentCase.notes, note] }
              : state.currentCase,
        })),

      updateConflictStatus: (caseId, conflictId, status) =>
        set((state) => ({
          cases: state.cases.map((case_) =>
            case_.id === caseId
              ? {
                  ...case_,
                  conflicts: case_.conflicts.map((conflict) =>
                    conflict.id === conflictId ? { ...conflict, status } : conflict
                  ),
                }
              : case_
          ),
        })),

      // 统计数据
      getStatistics: () => {
        const state = get()
        const stats = {
          total: state.cases.length,
          byStatus: {} as Record<CaseStatus, number>,
          byPriority: {} as Record<CasePriority, number>,
          byType: {} as Record<string, number>,
          avgProgress: 0,
        }

        state.cases.forEach((case_) => {
          // 按状态统计
          stats.byStatus[case_.status] = (stats.byStatus[case_.status] || 0) + 1
          // 按优先级统计
          stats.byPriority[case_.priority] = (stats.byPriority[case_.priority] || 0) + 1
          // 按类型统计
          stats.byType[case_.caseType.id] = (stats.byType[case_.caseType.id] || 0) + 1
          // 累计进度
          stats.avgProgress += case_.progress
        })

        stats.avgProgress = state.cases.length > 0 ? stats.avgProgress / state.cases.length : 0

        return stats
      },

      // 重置状态
      reset: () =>
        set({
          cases: [],
          currentCase: null,
          selectedCaseIds: [],
          editingCase: null,
          isLoading: false,
          isCreating: false,
          isUpdating: false,
          isDeleting: false,
          pagination: {
            page: 1,
            pageSize: 20,
            total: 0,
            totalPages: 0,
          },
          filter: {},
          sort: { field: 'createdAt', direction: 'desc' },
        }),
    }),
    {
      name: 'Case Management Store',
    }
  )
)

// 选择器Hook
export const useCases = () => useCaseStore((state) => state.cases)
export const useCurrentCase = () => useCaseStore((state) => state.currentCase)
export const useCasePagination = () => useCaseStore((state) => state.pagination)
export const useCaseFilter = () => useCaseStore((state) => state.filter)
export const useCaseSort = () => useCaseStore((state) => state.sort)
export const useSelectedCases = () => useCaseStore((state) => state.selectedCaseIds)

// 计算属性Hook
export const useFilteredCases = () => {
  const { applyFilter, pagination } = useCaseStore()
  const filteredCases = applyFilter()

  const startIndex = (pagination.page - 1) * pagination.pageSize
  const endIndex = startIndex + pagination.pageSize
  const paginatedCases = filteredCases.slice(startIndex, endIndex)

  return {
    filteredCases,
    paginatedCases,
    hasData: filteredCases.length > 0,
    isEmpty: filteredCases.length === 0,
  }
}

export const useCaseStatistics = () => {
  const getStatistics = useCaseStore((state) => state.getStatistics)
  return getStatistics()
}

// 导出类型
export type { Case, CaseFilter, CaseSort, CaseStoreState }