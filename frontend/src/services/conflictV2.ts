import { get, post } from './http'

// v2 冲突检测相关 API 接口

export interface ConflictCheckRequestV2 {
  lawyerId: number
  clientName: string
  clientTaxId?: string
  caseId?: number
  opposingParties?: string[]
  searchDepth?: 'basic' | 'standard' | 'deep'
  includeRelated?: boolean
}

export interface ConflictMatchV2 {
  matchId: string
  matchType: 'direct' | 'indirect' | 'related' | 'opposing' | 'api'
  lawyerId: number
  lawyerName: string
  caseId: number
  caseTitle: string
  caseType: string
  relationship: string
  matchReason: string
  entityInfo: {
    name: string
    standardName: string
    taxId: string
    type: string
  }
  riskLevel: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'PASS'
  riskFactors: string[]
}

export interface ConflictCheckResultV2 {
  checkId: string
  riskLevel: string
  riskScore: number
  matchCount: number
  matches: ConflictMatchV2[]
  checkTime: string
  durationMs: number
  searchScope: string
  recommendations: string[]
}

export interface ReportGenerationRequest {
  checkedBy: number
  checkTime: string
  checkDurationMs?: number
  clientName: string
  clientTaxId?: string
  opposingParty?: string
  riskLevel: string
  matchedCases: any
  relatedCompanies: any
  conflictDetails: any
  templateType?: string
}

export interface ConflictReport {
  id: number
  reportNumber: string
  clientName: string
  riskLevel: string
  reportUrl: string
  createdAt: string
}

export interface ConflictScanJob {
  id: number
  scanType: string
  scanScope: string
  status: string
  scannedCases: number
  scannedLawyers: number
  foundConflicts: number
  createdAt: string
  startedAt?: string
  completedAt?: string
}

/**
 * v2 冲突检测 API
 */
export const conflictV2Api = {
  /**
   * 快速冲突检测
   */
  quickCheck: (request: ConflictCheckRequestV2) => {
    return post<ConflictCheckResultV2>('/api/v2/conflict/quick-check', request)
  },

  /**
   * 详细冲突检测
   */
  detailedCheck: (request: ConflictCheckRequestV2) => {
    return post<ConflictCheckResultV2>('/api/v2/conflict/detailed-check', request)
  },

  /**
   * 生成 PDF 报告
   */
  generateReport: (request: ReportGenerationRequest) => {
    return post<ConflictReport>('/api/conflict/report', request)
  },

  /**
   * 获取报告列表
   */
  listReports: (params?: {
    checkedBy?: number
    riskLevel?: string
    limit?: number
    offset?: number
  }) => {
    return get<{ list: ConflictReport[]; total: number }>(`/api/conflict/reports?${new URLSearchParams(params as any).toString()}`)
  },

  /**
   * 下载报告
   */
  downloadReport: (reportId: number) => {
    return get<Blob>(`/api/conflict/reports/${reportId}/download`)
  },

  /**
   * 获取扫描任务列表
   */
  listScanJobs: (params?: {
    scanType?: string
    status?: string
    limit?: number
    offset?: number
  }) => {
    return get<{ list: ConflictScanJob[]; total: number }>(`/api/conflict/scan-jobs?${new URLSearchParams(params as any).toString()}`)
  },

  /**
   * 手动触发扫描
   */
  triggerScan: (request: {
    triggeredBy: number
    triggerReason?: string
    scanScope?: string
  }) => {
    return post<ConflictScanJob>('/api/conflict/scan/trigger', request)
  },

  /**
   * 获取扫描统计
   */
  getScanStats: () => {
    return get<{
      totalJobs: number
      runningJobs: number
      completedJobs: number
      totalConflicts: number
      lastScanTime?: string
    }>('/api/conflict/scan/stats')
  },

  /**
   * 验证报告签名
   */
  verifySignature: (reportId: number) => {
    return get<{
      valid: boolean
      signedBy: string
      signedAt: string
      hash: string
    }>(`/api/conflict/reports/${reportId}/verify`)
  },
}

export default conflictV2Api
