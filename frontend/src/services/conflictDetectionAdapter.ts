/**
 * 冲突检测API服务适配器
 * 用于处理现有冲突检测系统的数据格式，并将其转换为豁免审批系统可用的格式
 */

import { apiClient } from './apiClient'
import type {
  ConflictCase,
  ConflictCheckRequest,
  ConflictCheckResponse,
  EnhancedConflictCase,
  ConflictRiskLevel
} from '@/types/waiverApproval'

// 现有冲突检测系统的数据格式
export interface LegacyConflictCase {
  CaseID: string
  CaseName: string
  ConflictType: string
  RiskLevel: string
  ConflictDetail: string
  CreatedTime: string
  Status: string
  LawyerName?: string
  ClientName?: string
  Description?: string
}

export interface LegacyConflictCheckResponse {
  hasConflict: boolean
  conflictCases: LegacyConflictCase[]
  checkTime: string
  riskAssessment?: {
    overallRiskLevel: string
    highRiskCount: number
    mediumRiskCount: number
    lowRiskCount: number
  }
  lawyerInfo?: {
    lawyerId: string
    lawyerName: string
    department: string
  }
}

export interface LegacyConflictCheckRequest {
  lawyerId: string
  lawyerName: string
  caseId?: string
  caseTitle?: string
  clientIds?: string[]
  caseType?: string
  checkType?: 'automatic' | 'manual'
}

/**
 * 冲突检测API服务适配器
 * 处理现有系统数据格式到新系统格式的转换
 */
export class ConflictDetectionAdapter {
  private static instance: ConflictDetectionAdapter

  static getInstance(): ConflictDetectionAdapter {
    if (!ConflictDetectionAdapter.instance) {
      ConflictDetectionAdapter.instance = new ConflictDetectionAdapter()
    }
    return ConflictDetectionAdapter.instance
  }

  /**
   * 执行冲突检测（使用现有API）
   */
  async checkConflicts(request: ConflictCheckRequest): Promise<ConflictCheckResponse> {
    try {
      // 转换为现有系统的请求格式
      const legacyRequest = this.convertToLegacyRequest(request)

      // 调用现有的冲突检测API
      const response = await apiClient.post<LegacyConflictCheckResponse>('/api/conflicts/check', legacyRequest)

      // 转换响应格式
      return this.convertFromLegacyResponse(response.data)
    } catch (error) {
      console.error('冲突检测失败:', error)
      throw new Error(`冲突检测失败: ${error instanceof Error ? error.message : '未知错误'}`)
    }
  }

  /**
   * 获取律师的冲突案例历史
   */
  async getLawyerConflicts(lawyerId: string): Promise<ConflictCase[]> {
    try {
      const response = await apiClient.get<LegacyConflictCase[]>(`/api/conflicts/lawyer/${lawyerId}`)

      // 转换每个冲突案例的格式
      return response.data.map(legacyCase => this.convertConflictCase(legacyCase))
    } catch (error) {
      console.error('获取律师冲突案例失败:', error)
      throw new Error(`获取律师冲突案例失败: ${error instanceof Error ? error.message : '未知错误'}`)
    }
  }

  /**
   * 获取冲突检测历史记录
   */
  async getConflictHistory(lawyerId: string, limit: number = 10): Promise<ConflictCheckResponse[]> {
    try {
      const response = await apiClient.get<LegacyConflictCheckResponse[]>(`/api/conflicts/history/${lawyerId}`, {
        params: { limit }
      })

      return response.data.map(legacyResponse => this.convertFromLegacyResponse(legacyResponse))
    } catch (error) {
      console.error('获取冲突历史失败:', error)
      throw new Error(`获取冲突历史失败: ${error instanceof Error ? error.message : '未知错误'}`)
    }
  }

  /**
   * 创建豁免申请建议
   */
  createWaiverSuggestion(conflictResponse: ConflictCheckResponse): {
    recommendWaiver: boolean
    suggestedCases: ConflictCase[]
    reasoning: string
    riskLevel: ConflictRiskLevel
    waiverType: string
    mitigationMeasures: string
  } {
    if (!conflictResponse.hasConflict || conflictResponse.conflictCases.length === 0) {
      return {
        recommendWaiver: false,
        suggestedCases: [],
        reasoning: '未检测到需要豁免的冲突',
        riskLevel: 'LOW',
        waiverType: 'ORGANIZATIONAL',
        mitigationMeasures: '无需特殊措施'
      }
    }

    const riskLevel = this.determineRiskLevel(conflictResponse)
    const waiverType = this.determineWaiverType(conflictResponse.conflictCases)
    const mitigationMeasures = this.generateMitigationMeasures(conflictResponse.conflictCases)

    return {
      recommendWaiver: true,
      suggestedCases: conflictResponse.conflictCases,
      reasoning: '检测到潜在利益冲突，建议申请豁免审批',
      riskLevel,
      waiverType,
      mitigationMeasures
    }
  }

  /**
   * 转换请求格式到现有系统格式
   */
  private convertToLegacyRequest(request: ConflictCheckRequest): LegacyConflictCheckRequest {
    return {
      lawyerId: request.lawyerId,
      lawyerName: request.lawyerName,
      caseId: request.caseId,
      caseTitle: request.caseTitle,
      clientIds: request.clientIds,
      caseType: request.caseType,
      checkType: request.checkType || 'automatic'
    }
  }

  /**
   * 转换响应格式从现有系统格式
   */
  private convertFromLegacyResponse(legacyResponse: LegacyConflictCheckResponse): ConflictCheckResponse {
    return {
      checkId: `CHECK_${Date.now()}`,
      hasConflict: legacyResponse.hasConflict,
      conflictCases: legacyResponse.conflictCases.map(legacyCase => this.convertConflictCase(legacyCase)),
      checkTime: legacyResponse.checkTime,
      riskAssessment: legacyResponse.riskAssessment ? {
        overallRiskLevel: this.convertRiskLevel(legacyResponse.riskAssessment.overallRiskLevel),
        highRiskCount: legacyResponse.riskAssessment.highRiskCount,
        mediumRiskCount: legacyResponse.riskAssessment.mediumRiskCount,
        lowRiskCount: legacyResponse.riskAssessment.lowRiskCount,
        riskFactors: this.generateRiskFactors(legacyResponse.conflictCases)
      } : undefined,
      lawyerInfo: legacyResponse.lawyerInfo
    }
  }

  /**
   * 转换冲突案例格式
   */
  convertConflictCase(legacyCase: LegacyConflictCase): ConflictCase {
    return {
      id: legacyCase.CaseID,
      caseId: legacyCase.CaseID,
      caseName: legacyCase.CaseName,
      conflictType: this.mapConflictType(legacyCase.ConflictType),
      description: legacyCase.ConflictDetail || legacyCase.Description || '',
      riskLevel: this.convertRiskLevel(legacyCase.RiskLevel),
      status: this.mapStatus(legacyCase.Status),
      createdAt: legacyCase.CreatedTime,
      updatedAt: legacyCase.CreatedTime,
      parties: [
        {
          name: legacyCase.LawyerName || '未知律师',
          type: 'LAWYER',
          role: '代理律师'
        },
        {
          name: legacyCase.ClientName || '未知客户',
          type: 'CLIENT',
          role: '相关方'
        }
      ].filter(p => p.name !== '未知律师' && p.name !== '未知客户'),
      mitigationMeasures: this.getDefaultMitigationMeasures(legacyCase.ConflictType),
      resolutionNotes: '',
      assignedTo: legacyCase.LawyerName || '',
      resolvedAt: null
    }
  }

  /**
   * 转换风险等级
   */
  private convertRiskLevel(level: string): ConflictRiskLevel {
    switch (level.toUpperCase()) {
      case 'HIGH':
      case 'CRITICAL':
        return 'HIGH'
      case 'MEDIUM':
      case 'MED':
        return 'MEDIUM'
      case 'LOW':
      default:
        return 'LOW'
    }
  }

  /**
   * 映射冲突类型
   */
  private mapConflictType(type: string): string {
    const typeMap: { [key: string]: string } = {
      '案件冲突': 'REPRESENTATION_CONFLICT',
      '利益冲突': 'CONFLICT_OF_INTEREST',
      '业务关系': 'BUSINESS_RELATION',
      '代理冲突': 'REPRESENTATION_CONFLICT',
      '时间冲突': 'TIME_CONFLICT',
      '其他': 'OTHER'
    }

    return typeMap[type] || 'OTHER'
  }

  /**
   * 映射状态
   */
  private mapStatus(status: string): string {
    const statusMap: { [key: string]: string } = {
      '进行中': 'ACTIVE',
      '已解决': 'RESOLVED',
      '待处理': 'PENDING',
      '监控中': 'MONITORING',
      '已关闭': 'CLOSED'
    }

    return statusMap[status] || 'ACTIVE'
  }

  /**
   * 确定风险等级
   */
  private determineRiskLevel(response: ConflictCheckResponse): ConflictRiskLevel {
    if (response.riskAssessment) {
      return response.riskAssessment.overallRiskLevel
    }

    // 根据冲突案例数量和严重性确定风险等级
    if (response.conflictCases.length > 5) {
      return 'HIGH'
    } else if (response.conflictCases.length > 2) {
      return 'MEDIUM'
    }

    return 'LOW'
  }

  /**
   * 确定豁免类型
   */
  private determineWaiverType(conflictCases: ConflictCase[]): string {
    const conflictTypes = conflictCases.reduce((acc, conflict) => {
      acc[conflict.conflictType] = (acc[conflict.conflictType] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    // 根据冲突类型优先级确定豁免类型
    if (conflictTypes['REPRESENTATION_CONFLICT'] > 0) {
      return 'REPRESENTATION_CONFLICT'
    }
    if (conflictTypes['CONFLICT_OF_INTEREST'] > 0) {
      return 'CONFLICT_OF_INTEREST'
    }
    if (conflictTypes['BUSINESS_RELATION'] > 0) {
      return 'BUSINESS_RELATION'
    }

    return 'ORGANIZATIONAL'
  }

  /**
   * 生成风险控制措施
   */
  private generateMitigationMeasures(conflictCases: ConflictCase[]): string {
    if (conflictCases.length === 0) {
      return '无需特殊措施'
    }

    const measures = [
      '风险原因：记录冲突情况',
      '定期复查：保持警惕',
      '建立预防措施：针对代理冲突类型案件制定专门处理流程'
    ]

    return measures.join('；')
  }

  /**
   * 获取默认风险控制措施
   */
  private getDefaultMitigationMeasures(conflictType: string): string {
    const measuresMap: { [key: string]: string } = {
      'REPRESENTATION_CONFLICT': '信息隔离：确保不同案件的信息严格隔离；内部沟通：建立定期沟通机制',
      'CONFLICT_OF_INTEREST': '客户告知：如有必要，向客户充分披露潜在冲突；持续监控：建立定期冲突检查机制',
      'BUSINESS_RELATION': '文档记录：详细记录所有检测和处理过程；独立审查：安排独立律师审查相关文件',
      'TIME_CONFLICT': '时间管理：合理安排工作时间；助理支持：必要时增加助理支持',
      'OTHER': '综合措施：根据具体情况制定相应的风险控制措施'
    }

    return measuresMap[conflictType] || measuresMap['OTHER']
  }

  /**
   * 生成风险因素
   */
  private generateRiskFactors(conflictCases: LegacyConflictCase[]): string[] {
    const factors = new Set<string>()

    conflictCases.forEach(conflictCase => {
      if (conflictCase.RiskLevel === 'HIGH') {
        factors.add('高风险冲突')
      }
      if (conflictCase.ConflictType.includes('代理')) {
        factors.add('代理冲突')
      }
      if (conflictCase.ConflictType.includes('利益')) {
        factors.add('利益冲突')
      }
      if (conflictCase.Status === '进行中') {
        factors.add('未解决冲突')
      }
    })

    return Array.from(factors)
  }
}

// 导出单例实例
export const conflictDetectionAdapter = ConflictDetectionAdapter.getInstance()

// 导出便捷方法
export const checkConflicts = (request: ConflictCheckRequest) =>
  conflictDetectionAdapter.checkConflicts(request)

export const getLawyerConflicts = (lawyerId: string) =>
  conflictDetectionAdapter.getLawyerConflicts(lawyerId)

export const createWaiverSuggestion = (conflictResponse: ConflictCheckResponse) =>
  conflictDetectionAdapter.createWaiverSuggestion(conflictResponse)