import { post, get } from '@/services/http'
// 冲突检索请求接口
export interface ConflictCheckRequest {
  clientId: string
  clientName: string
  clientType: 'PERSON' | 'COMPANY' | 'ANY'
  otherParties: string[]
  caseName: string
  caseType: string
  searchYears?: number
  includeCorporateRelations?: boolean
  searchDepth?: 'BASIC' | 'STANDARD' | 'DEEP'
  userId?: string // 用户ID，后端需要
}

// 冲突检索响应接口
export interface ConflictCheckResponse {
  checkId: string
  hasConflict: boolean
  conflictCases: ConflictCase[]
  checkStatistics: CheckStatistics
  riskAssessment: RiskAssessment
  recommendations: string[]
  checkTime: string
  duration: number
}

export interface ConflictCase {
  caseId: string
  caseName: string
  caseNo: string
  conflictType: string
  riskLevel: string
  description: string
  caseStatus: string
  createTime: string
  conflictDetails: string
}

export interface CheckStatistics {
  totalCasesChecked: number
  clientHistoryCases: number
  relatedPartiesChecked: number
  corporateRelationsChecked: number
  timeRange: string
  searchScope: string
}

export interface RiskAssessment {
  overallRisk: string
  riskReason: string
  requiresApproval: boolean
  approvalLevel?: string
  riskFactors: string[]
}

// 冲突检索服务
export class ConflictCheckService {
  /**
   * 执行完整的利益冲突检索
   */
  static async performConflictCheck(request: ConflictCheckRequest): Promise<ConflictCheckResponse> {
    try {
      // 验证必要参数
      this.validateRequest(request)

      const response = await post<ConflictCheckResponse>('/conflict/check', request)
      if (!response) {
        throw new Error('后端服务无响应')
      }

      const result = response
      if (typeof result.hasConflict === 'undefined') {
        result.hasConflict = result.conflictCases && result.conflictCases.length > 0
      }

      return result
    } catch (error) {
      console.error('冲突检索服务错误:', error)

      // 确保错误信息清晰
      if (error instanceof Error) {
        throw new Error(`冲突检索失败: ${error.message}`)
      } else {
        throw new Error('冲突检索服务异常')
      }
    }
  }

  /**
   * 快速预检
   */
  static async performPreCheck(clientName: string, otherParties: string[]): Promise<any> {
    try {
      const response = await post('/conflict/pre-check', { clientName, otherParties })
      return response
    } catch (error) {
      console.error('预检服务错误:', error)
      throw error
    }
  }

  /**
   * 获取检索历史
   */
  static async getCheckHistory(clientId?: string): Promise<any[]> {
    try {
      const endpoint = clientId ? `/conflict/history?clientId=${clientId}` : '/conflict/history'

      const response = await get(endpoint)
      return response || []
    } catch (error) {
      console.error('获取检索历史错误:', error)
      return []
    }
  }

  /**
   * 验证请求参数
   */
  private static validateRequest(request: ConflictCheckRequest): void {
    if (!request.clientId || !request.clientName) {
      throw new Error('委托人信息不能为空')
    }

    if (!request.caseName || !request.caseType) {
      throw new Error('案件名称和类型不能为空')
    }

    if (!['PERSON', 'COMPANY', 'ANY'].includes(request.clientType)) {
      throw new Error('委托人类型无效')
    }
  }

  /**
   * 格式化冲突类型标签
   */
  static formatConflictType(type: string): string {
    const labels: { [key: string]: string } = {
      SAME_PARTY: '同一当事人',
      RELATED_PARTY: '关联方',
      OPPOSING_PARTY: '对方当事人',
    }
    return labels[type] || type
  }

  /**
   * 格式化风险等级
   */
  static formatRiskLevel(level: string): { text: string; color: string } {
    const formats: { [key: string]: { text: string; color: string } } = {
      LOW: { text: '低风险', color: 'green' },
      MEDIUM: { text: '中等风险', color: 'orange' },
      HIGH: { text: '高风险', color: 'red' },
      CRITICAL: { text: '严重风险', color: 'red' },
    }
    return formats[level] || { text: level, color: 'gray' }
  }
}

// 冲突检索结果处理工具
export class ConflictCheckResultProcessor {
  /**
   * 转换后端数据为前端展示格式
   */
  static processResult(backendResult: any): any {
    return {
      hasConflict: backendResult.hasConflict,
      conflictCases: (backendResult.conflictCases || []).map((conflict: any) => ({
        id: conflict.caseId,
        caseName: conflict.caseName,
        caseNo: conflict.caseNo,
        conflictType: ConflictCheckService.formatConflictType(conflict.conflictType),
        riskLevel: conflict.riskLevel,
        description: conflict.description,
        caseStatus: conflict.caseStatus,
        createTime: conflict.createTime,
      })),
      checkTime: backendResult.checkTime,
      checkDetails: {
        totalCasesChecked: backendResult.checkStatistics?.totalCasesChecked || 0,
        clientHistoryCases: backendResult.checkStatistics?.clientHistoryCases || 0,
        relatedPartiesChecked: backendResult.checkStatistics?.relatedPartiesChecked || 0,
        corporateRelationsChecked: backendResult.checkStatistics?.corporateRelationsChecked || 0,
        timeRange: backendResult.checkStatistics?.timeRange || '未知',
        riskAssessment: backendResult.riskAssessment?.overallRisk || 'UNKNOWN',
      },
      recommendations: backendResult.recommendations || [],
    }
  }

  /**
   * 生成检索统计摘要
   */
  static generateSummary(checkDetails: any): string {
    const { totalCasesChecked, clientHistoryCases, relatedPartiesChecked } = checkDetails
    return `检索了${totalCasesChecked}个案件，发现${clientHistoryCases}个历史案件，检查了${relatedPartiesChecked}个当事人`
  }
}
