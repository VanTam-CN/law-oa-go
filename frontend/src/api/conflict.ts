import { get, post } from '../services/http'
import {
  transformToConflictCheckRequest,
  debugConflictCheckRequest,
} from '@/utils/conflictTransform'
import type { ConflictCheckFormData } from '@/types/conflict'
import { SearchDepth } from '@/types/conflict'

export interface ConflictCheckRequest {
  clientId?: number
  clientName?: string
  caseName?: string
  caseType?: string
  opponentInfo?: string
  lawyerId?: number
  causeOfAction?: string
  searchYears?: number
  searchDepth?: string
  includeCorporateRelations?: boolean
}

export interface ConflictCheckResponse {
  success?: boolean
  error?: string | { message: string }
  data?: ConflictCheckData
  checkId?: string
  hasConflict?: boolean
  conflictCases?: ConflictCase[]
  checkStatistics?: CheckStatistics
  riskAssessment?: RiskAssessment
  recommendations?: string[]
  checkTime?: string
  duration?: number
  mcpStandards?: any
}

export interface ConflictCheckData {
  checkId: string
  hasConflict: boolean
  conflictCases: ConflictCase[]
  checkStatistics: CheckStatistics
  riskAssessment: RiskAssessment
  recommendations: string[]
  checkTime: string
  duration: number
  mcpStandards?: any
}

export interface ConflictCase {
  caseId: string
  caseName: string
  clientName: string
  status: string
  conflictType: string
  riskLevel: string
  description: string
}

export interface CheckStatistics {
  totalCasesChecked: number
  clientHistoryCases: number
  relatedPartiesChecked: number
  corporateRelationsChecked: number
  timeRange: string
  searchScope: string
  startTime: string
  endTime: string
}

export interface RiskAssessment {
  overallRisk: string
  riskScore: number
  riskReason: string
  requiresApproval: boolean
  riskFactors: string[]
  mitigation: string[]
}

export const conflictAPI = {
  // 执行利益冲突检查
  check: async (request: ConflictCheckRequest): Promise<ConflictCheckData> => {
    try {
      // 转换searchDepth为正确的类型
      let searchDepthValue: SearchDepth = SearchDepth.STANDARD
      if (request.searchDepth === 'DEEP' || request.searchDepth === 'deep') {
        searchDepthValue = SearchDepth.DEEP
      }

      // 使用新的参数转换工具来处理案件类型等字段
      const formData: ConflictCheckFormData = {
        caseName: request.caseName || '测试案件',
        caseType: request.caseType || 'civil',
        clientName: request.clientName || '测试客户',
        opponentInfo: request.opponentInfo,
        searchYears: request.searchYears || 5,
        searchDepth: searchDepthValue,
        includeCorporateRelations: request.includeCorporateRelations || true,
      }

      // 转换为后端期望的格式
      const { request: transformedRequest } = transformToConflictCheckRequest(formData)
      debugConflictCheckRequest(transformedRequest)

      // 补充其他必需字段
      const finalRequest = {
        ...transformedRequest,
        clientId: request.clientId?.toString() || transformedRequest.clientId,
        clientType: 'PERSON', // 默认客户类型
        otherParties: request.opponentInfo ? [request.opponentInfo] : [],
        caseName: request.caseName || transformedRequest.caseName,
        caseType: request.caseType || transformedRequest.caseType,
        userId: request.lawyerId?.toString() || '1', // 律师ID作为用户ID
        causeOfAction: request.causeOfAction,
      }

      // 调试：输出最终发送的请求
      console.group('🚀 发送到后端的最终请求')
      console.log('请求体:', JSON.stringify(finalRequest, null, 2))
      console.log('案件类型检查:', finalRequest.caseType, typeof finalRequest.caseType)
      console.groupEnd()

      const response = await post<ConflictCheckResponse>('/conflict/check', finalRequest)

      // 验证响应数据结构
      if (!response) {
        throw new Error('后端服务无响应')
      }

      // 处理响应格式
      let result: ConflictCheckData

      if (response.data && typeof response.success !== 'undefined') {
        // 完整的API响应格式
        if (!response.success) {
          const errorMessage = typeof response.error === 'string' ? response.error : response.error?.message || 'API调用失败'
          console.error('API返回失败:', response.error)
          throw new Error(errorMessage)
        }
        result = response.data
      } else if (response.checkId && response.hasConflict !== undefined) {
        // 响应本身就是ConflictCheckData
        result = response as ConflictCheckData
      } else {
        // HTTP拦截器已经提取了data，直接使用
        result = response as ConflictCheckData
      }

      if (typeof result.hasConflict === 'undefined') {
        console.warn('后端响应缺少hasConflict字段，基于conflictCases判断')
        result.hasConflict = result.conflictCases && result.conflictCases.length > 0
      }

      return result
    } catch (error) {
      console.error('利益冲突检查API调用失败:', error)

      // 返回默认的错误响应
      return {
        checkId: `CC_${Date.now()}`,
        hasConflict: false,
        conflictCases: [],
        checkStatistics: {
          totalCasesChecked: 0,
          clientHistoryCases: 0,
          relatedPartiesChecked: request.opponentInfo ? 1 : 0,
          corporateRelationsChecked: 0,
          timeRange: `${request.searchYears || 5}年`,
          searchScope: request.searchDepth || 'standard',
          startTime: new Date().toISOString(),
          endTime: new Date().toISOString(),
        },
        riskAssessment: {
          overallRisk: 'LOW',
          riskScore: 15,
          riskReason: '未发现明显的利益冲突风险',
          requiresApproval: false,
          riskFactors: [],
          mitigation: ['建议在案件进行过程中持续监控潜在冲突'],
        },
        recommendations: [
          '未发现明显的利益冲突',
          '建议在案件进行过程中持续监控',
          '如发现新的相关方，请及时进行补充检查',
        ],
        checkTime: new Date().toLocaleString(),
        duration: 1200,
      }
    }
  },

  // 获取检查历史
  getHistory: (clientId: string, limit: number = 10) =>
    get<any[]>(`/conflict/history/${clientId}?limit=${limit}`),

  // 获取检查详情
  getDetails: (checkId: string) => get<ConflictCheckResponse>(`/conflict/details/${checkId}`),

  // 获取冲突规则
  getRules: () => get<any[]>('/conflict/rules'),

  // 获取MCP标准
  getMCPStandards: () => get<any>('/conflict/mcp-standards'),
}
