import { get, post, put } from './http'

// 集成工作流服务

export interface IntegrationRequest {
  type: string
  title: string
  content: string
  applicant_name: string
  department_name: string
  urgency?: string
  priority?: string
  workflow_type: string
  expected_duration?: number
  category?: string
  metadata?: Record<string, any>
  conflict_check_config?: ConflictCheckConfig
  case_creation_config?: Record<string, any>
}

export interface ConflictCheckConfig {
  user_id: string
  client_ids: string[]
  search_scope: string
  check_type: string
  include_potential?: boolean
  mitigation_required?: boolean
}

export interface ConflictCheckResult {
  check_id: string
  status: string
  has_conflict: boolean
  conflict_count: number
  risk_level: string
  risk_score: number
  check_time: string
  duration: number
  conflict_cases?: any[]
  recommendations?: string[]
}

export interface CaseCreationResult {
  case_id: string
  case_number: string
  status: string
  message: string
  created_at: string
  case_data?: Record<string, any>
}

export interface IntegrationResult {
  approval_id: string
  status: string
  message: string
  created_at: string
  conflict_check?: ConflictCheckResult
  case_creation?: CaseCreationResult
  integration_details?: any
}

export interface IntegrationStatus {
  approval_id: string
  overall_status: string
  conflict_check?: {
    status: string
    check_id?: string
    has_conflict?: boolean
    risk_level?: string
    last_checked?: string
    duration?: number
  }
  case_creation?: {
    status: string
    case_id?: string
    case_number?: string
    last_error?: string
    retry_count?: number
    last_attempted?: string
  }
  next_actions?: string[]
  timeline?: TimelineEvent[]
}

export interface TimelineEvent {
  timestamp: string
  event_type: string
  description: string
  actor: string
  details?: Record<string, any>
}

export interface IntegrationStatistics {
  total_integrations: number
  completed_integrations: number
  pending_integrations: number
  conflict_detection_rate: number
  auto_case_creation_rate: number
  average_processing_time: number
  success_rate: number
}

/**
 * 创建带冲突检测的集成审批申请
 * @param request 集成请求参数
 * @returns 集成结果
 */
export const createApprovalWithConflict = async (
  request: IntegrationRequest,
): Promise<IntegrationResult> => {
  try {
    const response = await post('/integration/approvals/with-conflict', request)
    return response.data
  } catch (error) {
    console.error('创建带冲突检测的审批申请失败:', error)
    throw error
  }
}

/**
 * 创建集成审批申请
 * @param request 集成请求参数
 * @returns 集成结果
 */
export const createIntegratedApproval = async (
  request: IntegrationRequest,
): Promise<IntegrationResult> => {
  try {
    const response = await post('/integration/approvals', request)
    return response.data
  } catch (error) {
    console.error('创建集成审批申请失败:', error)
    throw error
  }
}

/**
 * 手动触发冲突检测
 * @param config 冲突检测配置
 * @returns 冲突检测结果
 */
export const triggerConflictCheck = async (
  config: ConflictCheckConfig,
): Promise<ConflictCheckResult> => {
  try {
    const response = await post('/integration/conflict-check', config)
    return response.data
  } catch (error) {
    console.error('触发冲突检测失败:', error)
    throw error
  }
}

/**
 * 查询集成状态
 * @param approvalId 审批ID
 * @returns 集成状态
 */
export const getIntegrationStatus = async (approvalId: string): Promise<IntegrationStatus> => {
  try {
    const response = await get(`/integration/approvals/${approvalId}/status`)
    return response.data
  } catch (error) {
    console.error('查询集成状态失败:', error)
    throw error
  }
}

/**
 * 为审批申请创建案件
 * @param approvalId 审批ID
 * @param caseData 案件数据
 * @returns 案件创建结果
 */
export const createCaseFromApproval = async (
  approvalId: string,
  caseData: Record<string, any>,
): Promise<CaseCreationResult> => {
  try {
    const response = await post(`/integration/approvals/${approvalId}/case`, {
      case_data: caseData,
    })
    return response.data
  } catch (error) {
    console.error('从审批申请创建案件失败:', error)
    throw error
  }
}

/**
 * 获取集成统计信息
 * @returns 统计数据
 */
export const getIntegrationStatistics = async (): Promise<IntegrationStatistics> => {
  try {
    const response = await get('/integration/statistics')
    return response.data
  } catch (error) {
    console.error('获取集成统计失败:', error)
    throw error
  }
}

/**
 * 获取集成工作流列表
 * @returns 工作流列表
 */
export const getIntegrationWorkflows = async (): Promise<any[]> => {
  try {
    const response = await get('/integration/workflows')
    return response.data
  } catch (error) {
    console.error('获取集成工作流失败:', error)
    throw error
  }
}

/**
 * 验证集成配置
 * @param config 集成配置
 * @returns 验证结果
 */
export const validateIntegrationConfig = (
  config: IntegrationRequest,
): { isValid: boolean; errors: string[] } => {
  const errors: string[] = []

  if (!config.type || config.type.trim().length === 0) {
    errors.push('集成类型不能为空')
  }

  if (!config.title || config.title.trim().length === 0) {
    errors.push('标题不能为空')
  }

  if (!config.content || config.content.trim().length === 0) {
    errors.push('内容不能为空')
  }

  if (!config.applicant_name || config.applicant_name.trim().length === 0) {
    errors.push('申请人姓名不能为空')
  }

  if (!config.department_name || config.department_name.trim().length === 0) {
    errors.push('部门名称不能为空')
  }

  if (!config.workflow_type || config.workflow_type.trim().length === 0) {
    errors.push('工作流类型不能为空')
  }

  // 如果包含冲突检测配置，验证相关字段
  if (config.conflict_check_config) {
    if (!config.conflict_check_config.user_id) {
      errors.push('用户ID不能为空')
    }

    if (
      !config.conflict_config.client_ids ||
      config.conflict_check_config.client_ids.length === 0
    ) {
      errors.push('客户ID列表不能为空')
    }
  }

  return {
    isValid: errors.length === 0,
    errors,
  }
}

/**
 * 根据风险级别获取建议的审批级别
 * @param riskLevel 风险级别
 * @returns 审批级别
 */
export const getApprovalLevelByRisk = (riskLevel: string): string => {
  const riskMapping: Record<string, string> = {
    LOW: '普通',
    MEDIUM: '中级',
    HIGH: '高级',
    CRITICAL: '紧急',
  }

  return riskMapping[riskLevel.toUpperCase()] || '普通'
}

/**
 * 生成集成ID
 * @returns 集成ID
 */
export const generateIntegrationId = (): string => {
  const timestamp = new Date().getTime().toString(36)
  const random = Math.random().toString(36).substr(2, 5)
  return `INT_${timestamp}_${random}`.toUpperCase()
}

/**
 * 格式化集成状态显示
 * @param status 集成状态
 * @returns 格式化的状态文本
 */
export const formatIntegrationStatus = (status: string): string => {
  const statusMapping: Record<string, string> = {
    pending: '等待处理',
    in_progress: '进行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
  }

  return statusMapping[status] || status
}

/**
 * 格式化风险级别显示
 * @param riskLevel 风险级别
 * @returns 格式化的风险级别文本
 */
export const formatRiskLevel = (riskLevel: string): string => {
  const riskMapping: Record<string, string> = {
    LOW: '低风险',
    MEDIUM: '中风险',
    HIGH: '高风险',
    CRITICAL: '紧急风险',
  }

  return riskMapping[riskLevel.toUpperCase()] || riskLevel
}

/**
 * 计算集成处理时间
 * @param startTime 开始时间
 * @param endTime 结束时间
 * @returns 处理时间（毫秒）
 */
export const calculateProcessingTime = (startTime: string, endTime: string): number => {
  const start = new Date(startTime).getTime()
  const end = new Date(endTime).getTime()
  return end - start
}

/**
 * 格式化处理时间为可读文本
 * @param timeMs 毫秒
 * @returns 格式化时间文本
 */
export const formatProcessingTime = (timeMs: number): string => {
  if (timeMs < 1000) {
    return `${timeMs}毫秒`
  } else if (timeMs < 60000) {
    return `${Math.round(timeMs / 1000)}秒`
  } else if (timeMs < 3600000) {
    return `${Math.round(timeMs / 60000)}分钟`
  } else {
    return `${Math.round(timeMs / 3600000)}小时`
  }
}

export default {
  createApprovalWithConflict,
  createIntegratedApproval,
  triggerConflictCheck,
  getIntegrationStatus,
  createCaseFromApproval,
  getIntegrationStatistics,
  getIntegrationWorkflows,
  validateIntegrationConfig,
  getApprovalLevelByRisk,
  generateIntegrationId,
  formatIntegrationStatus,
  formatRiskLevel,
  calculateProcessingTime,
  formatProcessingTime,
}
