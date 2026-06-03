/**
 * 利益冲突检查相关的类型定义
 * 与后端 API 完全对齐
 */

// ==================== 基础实体类型 ====================

/** 实体类型：自然人或法人实体 */
export interface Entity {
  id: number
  entity_type: 'NATURAL_PERSON' | 'LEGAL_ENTITY'
  name: string
  identity_type: string
  identity_number: string
  status: string
  created_at?: string
  updated_at?: string
}

/** 案件当事人角色 */
export type CasePartyRole = 'PLAINTIFF' | 'DEFENDANT' | 'THIRD_PARTY' | 'INTERESTED_PARTY'

/** 案件当事人 */
export interface CaseParty {
  entity_id: number
  role: CasePartyRole
  party_type: string
  entity_name?: string
}

// ==================== 冲突检查请求 ====================

/** 后端期望的请求格式 (基于 CheckConflictRequest) */
export interface ConflictCheckRequest {
  clientId: string // 必填 - 客户ID
  clientName: string // 必填 - 客户名称
  caseName: string // 必填 - 案件名称
  caseType: CaseType // 必填 - 案件类型
  clientType: ClientType // 必填 - 客户类型 (PERSON | COMPANY)
  otherParties: string[] // 对方当事人列表
  searchYears?: number // 搜索年限 (默认: 5)
  includeCorporateRelations?: boolean // 是否包含企业关系 (默认: true)
  searchDepth?: SearchDepth // 搜索深度 (默认: STANDARD)
  userId?: string // 用户ID (用于审计)
  requestTime?: string // 请求时间
}

/** 冲突检查创建请求 */
export interface ConflictCheckCreateRequest {
  case_id: number
  parties: CaseParty[]
}

// ==================== 冲突结果类型 ====================

/** 风险等级 */
export type RiskLevel = 'CRITICAL' | 'BLOCKED' | 'HIGH' | 'MEDIUM' | 'LOW' | 'MINIMAL'

/** 冲突结果详情 */
export interface ConflictResult {
  id: number
  matched_entity: Entity
  matched_case: {
    id: number
    case_no: string
    case_name: string
  }
  conflict_type: string
  risk_level: RiskLevel
  description: string
}

/** 冲突检查记录状态（用于审批流程） */
export type ConflictCheckRecordStatus = 'PENDING' | 'APPROVED' | 'WAIVED' | 'REJECTED'

/** 冲突检查记录（用于审批流程） */
export interface ConflictCheckRecord {
  id: number
  case_id: number
  status: ConflictCheckRecordStatus
  result_summary: string
  details: ConflictResult[]
  checked_at: string
  created_at?: string
  updated_at?: string
}

// ==================== 枚举类型 ====================

/** 案件类型枚举 (与后端验证规则保持一致) */
export enum CaseType {
  CIVIL = 'civil',
  COMMERCIAL = 'commercial',
  CRIMINAL = 'criminal',
  ADMINISTRATIVE = 'administrative',
  ARBITRATION = 'arbitration',
  CONSULTATION = 'consultation',
  OTHER = 'other',
}

// 客户类型枚举
export enum ClientType {
  PERSON = 'PERSON',
  COMPANY = 'COMPANY',
}

// 搜索深度枚举
export enum SearchDepth {
  BASIC = 'BASIC',
  STANDARD = 'STANDARD',
  DEEP = 'DEEP',
}

// 后端响应格式 (基于 CheckConflictResponse)
export interface ConflictCheckResponse {
  success: boolean
  message: string
  data?: ConflictCheckResult
  error?: string
  details?: Record<string, any>
  requestId?: string
  timestamp: string
}

// 冲突检查结果
export interface ConflictCheckResult {
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

// 冲突案件信息
export interface ConflictCase {
  caseId: string
  caseName: string
  clientName: string
  status: string
  conflictType: string
  riskLevel: 'MINIMAL' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
  description: string
}

// 检查统计信息
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

// 风险评估
export interface RiskAssessment {
  overallRisk: 'MINIMAL' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
  riskScore: number
  riskReason: string
  requiresApproval: boolean
  riskFactors: string[]
  mitigation: string[]
}

// 错误信息类型
export interface ConflictCheckError {
  code: string
  message: string
  field?: string
  details?: Record<string, any>
}

// 前端业务数据格式 (用于表单)
export interface ConflictCheckFormData {
  clientId?: string
  clientName?: string
  caseName: string
  caseType: string
  clientType?: ClientType
  opponentInfo?: string
  lawyerId?: string
  causeOfAction?: string
  searchYears?: number
  searchDepth?: SearchDepth
  includeCorporateRelations?: boolean
  description?: string
}

// 验证错误类型
export interface ValidationError {
  field: string
  message: string
  code: string
}

// API 响应包装器
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: ConflictCheckError
  message: string
  requestId?: string
  timestamp: string
}

// 冲突检查执行状态（用于前端UI状态）
export enum ConflictCheckExecutionStatus {
  IDLE = 'idle',
  CHECKING = 'checking',
  COMPLETED = 'completed',
  ERROR = 'error',
}

// 冲突检查结果状态
export enum ConflictCheckResultStatus {
  NO_CONFLICT = 'no_conflict',
  HAS_CONFLICT = 'has_conflict',
  ERROR = 'error',
}

// ==================== 列表查询参数 ====================

/** 冲突检查列表查询参数 */
export interface ConflictCheckListParams {
  page?: number
  page_size?: number
  status?: ConflictCheckRecordStatus
  case_id?: number
  start_date?: string
  end_date?: string
}

/** 冲突检查列表响应 */
export interface ConflictCheckListResponse {
  total: number
  page: number
  page_size: number
  items: ConflictCheckRecord[]
}

// ==================== 实体搜索 ====================

/** 实体搜索结果 */
export interface EntitySearchResult {
  id: number
  name: string
  entity_type: 'NATURAL_PERSON' | 'LEGAL_ENTITY' | 'INDIVIDUAL' | 'LEGAL_PERSON' | 'ORGANIZATION'
  identity_type: string
  identity_number: string
  matched_field: string
}

/** 实体搜索参数 */
export interface EntitySearchParams {
  keyword: string
  entity_type?: 'NATURAL_PERSON' | 'LEGAL_ENTITY' | 'INDIVIDUAL' | 'LEGAL_PERSON' | 'ORGANIZATION'
  entityType?: 'NATURAL_PERSON' | 'LEGAL_ENTITY' | 'INDIVIDUAL' | 'LEGAL_PERSON' | 'ORGANIZATION'
  page?: number
  page_size?: number
}
