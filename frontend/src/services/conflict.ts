import { post } from './http'
import {
  ConflictCheckRequest,
  ConflictCheckResponse,
  ConflictCheckFormData,
  SearchDepth,
} from '@/types/conflict'
import {
  transformToConflictCheckRequest,
  debugConflictCheckRequest,
} from '@/utils/conflictTransform'
import {
  shouldRetry,
  getUserFriendlyMessage,
  ErrorHandler,
} from '@/utils/errorHandler'

export interface ConflictCheckParams {
  project_name: string
  client_name: string
  opposite_parties: string
  project_type: string
  team_members: string[]
  description?: string
}

export interface ConflictResult {
  has_conflict: boolean
  conflict_level: 'none' | 'low' | 'medium' | 'high'
  conflicts: {
    id: number
    type: string
    entity: string
    project: string
    level: 'low' | 'medium' | 'high'
    description: string
  }[]
}

// 创建专用的冲突检查错误处理器
const conflictErrorHandler = new ErrorHandler({
  enableLogging: true,
  enableRetry: true,
  maxRetries: 2,
  retryDelay: 1000,
  logLevel: 'ERROR',
})

/**
 * 转换为增强冲突检测API格式
 */
const transformToEnhancedConflictCheckRequest = (request: any) => {
  return {
    clientId: request.clientId,
    clientName: request.clientName,
    caseName: request.caseName,
    caseType: request.caseType.toLowerCase(),
    clientType: request.clientType || 'PERSON',
    otherParties: request.otherParties || [],
    searchYears: request.searchYears || 5,
    includeCorporateRelations: request.includeCorporateRelations !== false,
    searchDepth: request.searchDepth || SearchDepth.STANDARD,
    userId: request.userId?.toString(),
    requestTime: new Date().toISOString(),
  }
}

/**
 * 转换增强API响应为前端格式
 */
const transformEnhancedApiResponse = (apiResponse: any): ConflictCheckResponse => {
  // 🔧 修复：HTTP拦截器已经处理了响应格式，apiResponse就是data
  // 不再需要检查apiResponse.data，因为拦截器已经提取了data字段
  const data = apiResponse

  return {
    success: true, // 如果能到这里说明请求成功
    message: '冲突检测完成',
    error: undefined,
      data: {
      checkId: data.checkId || `REQ_${Date.now()}`,
      hasConflict: data.hasConflict || false,
      riskAssessment: {
        overallRisk: data.riskAssessment?.overallRisk || 'MINIMAL',
        riskScore: data.riskAssessment?.riskScore || 0,
        riskReason: data.riskAssessment?.riskReason || '',
        requiresApproval: data.riskAssessment?.requiresApproval || false,
        riskFactors: data.riskAssessment?.riskFactors || [],
        mitigation: data.riskAssessment?.mitigation || [],
      },
      conflictCases: (data.conflictCases || []).map((item: any) => ({
        caseId: item.caseId || '-1',
        clientName: data.checkStatistics?.clientHistoryCases > 0 ? '历史客户' : '未知客户',
        caseName: item.caseName || '未知案件',
        conflictType: item.conflictType || '代理冲突',
        riskLevel: item.riskLevel || 'MEDIUM',
        riskScore: data.riskAssessment?.riskScore || 50,
        description: item.description || '检测到利益冲突',
      })),
      checkStatistics: data.checkStatistics || {
        totalCasesChecked: 0,
        clientHistoryCases: 0,
        relatedPartiesChecked: 0,
        corporateRelationsChecked: 0,
        timeRange: '',
        searchScope: '',
        startTime: new Date().toISOString(),
        endTime: new Date().toISOString(),
      },
      recommendations: data.recommendations || [],
      checkTime: data.checkTime || new Date().toISOString(),
      duration: data.duration || 0,
    },
    requestId: `REQ_${Date.now()}`,
    timestamp: new Date().toISOString(),
  }
}

/**
 * 执行利益冲突检查 (新版本 - 使用增强API)
 * @param formData 表单数据
 * @param clientInfo 客户信息 (可选)
 * @param userInfo 用户信息 (可选)
 * @param attempt 当前重试次数 (内部使用)
 * @returns 冲突检查结果
 */
export const performConflictCheck = async (
  formData: ConflictCheckFormData,
  clientInfo?: any,
  userInfo?: any,
  attempt: number = 1,
): Promise<ConflictCheckResponse> => {
  try {
    // 1. 验证和转换参数
    const { request, errors } = transformToConflictCheckRequest(formData, clientInfo, userInfo)

    if (errors.length > 0) {
      const errorMessage = errors.map((e) => `${e.field}: ${e.message}`).join('; ')
      console.error('🚫 冲突检查请求验证失败:', errorMessage)

      return {
        success: false,
        message: '请求参数验证失败',
        error: errorMessage,
        data: undefined,
        requestId: `VALIDATION_${Date.now()}`,
        timestamp: new Date().toISOString(),
      } as ConflictCheckResponse
    }

    debugConflictCheckRequest(request)

    // 2. 转换为增强API格式
    const enhancedRequest = transformToEnhancedConflictCheckRequest(request)

    console.group('🚀 发送到增强API的请求')
    console.log('转换后请求体:', JSON.stringify(enhancedRequest, null, 2))
    console.groupEnd()

    // 3. 发送增强API请求
    const response = await post<any>('/conflict/check', enhancedRequest)

    // 4. 转换响应格式
    const transformedResponse = transformEnhancedApiResponse(response)

    console.log('✅ 增强冲突检查API调用成功:', transformedResponse.requestId)
    return transformedResponse
  } catch (error: any) {
    console.error(`💥 增强冲突检查API调用失败 (尝试 ${attempt}):`, error)

    // 标准化错误处理
    const standardError = conflictErrorHandler.handleError(error, 'ConflictCheck')

    // 检查是否应该重试
    if (shouldRetry(standardError, attempt)) {
      console.log(`🔄 准备重试 (${attempt}/${conflictErrorHandler['config'].maxRetries})...`)

      const delay = conflictErrorHandler.getRetryDelay(attempt)
      await new Promise((resolve) => setTimeout(resolve, delay))

      return performConflictCheck(formData, clientInfo, userInfo, attempt + 1)
    }

    // 返回用户友好的错误响应
    return {
      success: false,
      message: getUserFriendlyMessage(standardError),
      error: standardError.message,
      data: undefined,
      requestId: standardError.requestId,
      timestamp: standardError.timestamp,
    } as ConflictCheckResponse
  }
}

/**
 * @deprecated 请使用新的 performConflictCheck 方法，此方法为向后兼容保留
 * 执行利益冲突检查 (旧版本)
 * @param params 旧的参数格式
 * @returns 冲突检查结果
 */
export const performConflictCheckLegacy = (
  params: ConflictCheckParams,
): Promise<ConflictResult> => {
  console.warn(
    '⚠️  使用了已弃用的performConflictCheckLegacy方法，建议使用新的performConflictCheck方法',
  )

  // 转换为新的表单数据格式
  const formData: ConflictCheckFormData = {
    caseName: params.project_name,
    clientName: params.client_name,
    caseType: params.project_type,
    opponentInfo: params.opposite_parties,
    // 旧版接口未提供的字段使用保守配置
    searchYears: 5,
    searchDepth: SearchDepth.DEEP,
    includeCorporateRelations: true,
  }

  // 调用新方法
  return performConflictCheck(formData)
    .then((response): ConflictResult => {
      // 转换为旧格式
      return {
        has_conflict: response.data?.hasConflict || false,
        conflict_level:
          response.data?.riskAssessment?.overallRisk === 'CRITICAL' ||
          response.data?.riskAssessment?.overallRisk === 'HIGH'
            ? 'high'
            : response.data?.riskAssessment?.overallRisk === 'MEDIUM'
              ? 'medium'
              : response.data?.riskAssessment?.overallRisk === 'LOW'
                ? 'low'
                : 'none',
        conflicts:
          response.data?.conflictCases?.map((c) => ({
            id: parseInt(c.caseId) || 0,
            type: c.conflictType,
            entity: c.clientName,
            project: c.caseName,
            level: (['critical', 'high'].includes(c.riskLevel.toLowerCase()) ? 'high' : c.riskLevel.toLowerCase() === 'medium' ? 'medium' : 'low') as 'low' | 'medium' | 'high',
            description: c.description,
          })) || [],
      }
    })
    .catch((error): never => {
      throw error
    })
}

/**
 * 执行利益冲突预检
 * @param params 预检参数
 * @returns 预检结果
 */
export const performPreScreen = (params: {
  our_client_ids: number[]
  opponent_parties: string[]
  third_parties?: string[]
}): Promise<any> => {
  return post<any>('/conflict-check/pre-screen', params)
}

// ==================== 新增 API 方法 ====================

import { get } from './http'
import type {
  ConflictCheckCreateRequest,
  ConflictCheckRecord,
  ConflictCheckListParams,
  ConflictCheckListResponse,
  EntitySearchParams,
  EntitySearchResult,
} from '@/types/conflict'

/**
 * 创建冲突检查记录
 * @param data 冲突检查创建请求
 * @returns 创建的冲突检查记录
 */
export const createConflictCheck = async (
  data: ConflictCheckCreateRequest,
): Promise<ConflictCheckRecord> => {
  const response = await post<ConflictCheckRecord>('/conflict/checks', data)
  return response
}

/**
 * 获取冲突检查详情
 * @param id 冲突检查ID
 * @returns 冲突检查详情
 */
export const getConflictCheck = async (id: number): Promise<ConflictCheckRecord> => {
  const response = await get<ConflictCheckRecord>(`/conflict/checks/${id}`)
  return response
}

/**
 * 获取冲突检查列表
 * @param params 查询参数
 * @returns 冲突检查列表
 */
export const listConflictChecks = async (
  params?: ConflictCheckListParams,
): Promise<ConflictCheckListResponse> => {
  const queryParams = new URLSearchParams()
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, String(value))
      }
    })
  }
  const url = `/conflict/checks${queryParams.toString() ? `?${queryParams.toString()}` : ''}`
  const response = await get<ConflictCheckListResponse>(url)
  return response
}

/**
 * 搜索实体（支持模糊搜索）
 * @param keyword 搜索关键词
 * @param params 额外搜索参数
 * @returns 实体搜索结果
 */
export const searchEntities = async (
  keyword: string,
  params?: Omit<EntitySearchParams, 'keyword'>,
): Promise<EntitySearchResult[]> => {
  const queryParams = new URLSearchParams({ query: keyword })
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, String(value))
      }
    })
  }
  const url = `/conflict-v2/entities/search${queryParams.toString() ? `?${queryParams.toString()}` : ''}`
  const response = await get<EntitySearchResult[] | { results?: EntitySearchResult[] }>(url)
  return Array.isArray(response) ? response : response.results || []
}
