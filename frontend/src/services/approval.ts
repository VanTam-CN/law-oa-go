import { get, post, put, del } from './http'
import { getUserInfo } from '@/utils/storage'

export interface ApprovalItem {
  id: string // 后端使用UUID字符串
  type: string
  title: string
  content: string
  applicant: string
  applicantId: string // 后端使用字符串ID
  department: string
  createTime: string
  status:
    | 'draft'
    | 'submitted'
    | 'under_review'
    | 'approved'
    | 'rejected'
    | 'cancelled'
    | 'expired'
    | 'pending'
    | 'needs_revision' // 新增：需要修改
    | 'resubmitted' // 新增：重新提交
  urgency: 'normal' | 'urgent' | 'very_urgent'
  priority: 'low' | 'medium' | 'high' | 'critical'
  currentApprover?: string
  currentApproverId?: string
  requestNumber?: string
  submissionDate?: string
  currentStage?: string
  workflowType?: string
  metadata?: string
}

// 获取当前用户ID的函数
const getCurrentUserId = (): string => {
  const userInfo = getUserInfo()

  console.log('🔍 getCurrentUserId - 原始用户信息:', userInfo)
  console.log('🔍 getCurrentUserId - 用户ID:', userInfo?.id)
  console.log('🔍 getCurrentUserId - 用户ID类型:', typeof userInfo?.id)

  // 如果有真实用户信息，使用真实用户ID
  if (userInfo?.id !== undefined && userInfo?.id !== null) {
    const userIdStr = userInfo.id.toString()
    console.log('✅ getCurrentUserId - 使用真实用户ID:', userIdStr)
    return userIdStr
  }

  // 如果没有用户信息，记录警告并使用默认值
  console.warn('⚠️ getCurrentUserId - 没有找到用户信息，使用默认值')

  // 开发环境默认用户ID
  const isDevMode = process.env.NODE_ENV === 'development'
  if (isDevMode) {
    console.log('🛠️ 开发模式：使用默认测试用户ID')
    return '1'
  }

  // 生产环境默认值
  return '1'
}

export interface ApprovalDetail extends ApprovalItem {
  records: ApprovalRecord[]
}

export interface ApprovalRecord {
  id: string
  approvalRequestID: string
  approver: string
  approverId: string
  decision: 'approve' | 'reject' | 'request_changes' | 'reassign'
  decisionReason: string
  decisionComments: string
  approvalDate: string
  stage: string
  stageOrder: number
  status: string
}

export interface ApprovalStats {
  totalRequests: number
  pendingRequests: number
  myPendingRequests: number
  approvedRequests: number
  rejectedRequests: number
}

export interface CreateApprovalParams {
  type: string
  title: string
  content: string
  category?: string
  applicant: string
  applicantId: string
  department: string
  departmentId?: string
  urgency: 'normal' | 'urgent' | 'very_urgent'
  priority?: 'low' | 'medium' | 'high' | 'critical'
  workflowType?: string
  expectedEffectiveDate?: string
  expectedExpiryDate?: string
  durationDays?: number
  attachments?: any[]
  metadata?: any
}

// 利益冲突审批相关接口
export interface ConflictApprovalParams {
  caseId: string
  caseTitle: string
  conflictReason: string
  riskLevel: string
  conflictCases: Array<{
    caseId: string
    caseName: string
    conflictType: string
    riskLevel: string
    description: string
  }>
  applicant: string
  applicantId: string
  department: string
  departmentId?: string
  urgency?: 'normal' | 'urgent' | 'very_urgent'
  priority?: 'low' | 'medium' | 'high' | 'critical'
  additionalNotes?: string
}

export interface ConflictApprovalResult {
  approvalId: string
  approvalNumber: string
  status:
    | 'draft'
    | 'submitted'
    | 'under_review'
    | 'approved'
    | 'rejected'
    | 'cancelled'
    | 'expired'
    | 'pending'
  submitTime: string
  expectedProcessingTime: string
}

/**
 * 获取审批列表
 * @param type 类型：pending-待我审批，my-我的申请
 * @returns 审批列表
 */
export const getApprovals = async (type: 'pending' | 'my'): Promise<ApprovalItem[]> => {
  try {
    console.log('🚀 获取审批列表 - 类型:', type)

    if (type === 'pending') {
      // 待我审批
      console.log('📋 获取待我审批的列表...')
      const response = (await get('/approvals/pending')) as any
      return response.list || response
    } else {
      // 我的申请
      const currentUserId = getCurrentUserId()
      const requestUrl = `/approvals?applicantId=${currentUserId}`
      console.log('📝 获取我的申请列表 - 用户ID:', currentUserId)
      console.log('📝 请求URL:', requestUrl)
      const response = (await get(requestUrl)) as any
      console.log('✅ 我的申请列表响应:', response)
      return response.list || response
    }
  } catch (error) {
    console.error('❌ 获取审批列表失败:', error)
    throw error
  }
}

/**
 * 获取审批详情
 * @param id 审批ID
 * @returns 审批详情
 */
export const getApprovalDetail = async (id: string): Promise<ApprovalDetail> => {
  try {
    console.log('🔍 获取审批详情 - ID:', id)
    console.log('🔍 ID类型:', typeof id)
    const response = (await get(`/approvals/${id}`)) as any
    console.log('✅ 审批详情响应:', response)
    return response // HTTP拦截器已经提取了data字段
  } catch (error) {
    console.error('❌ 获取审批详情失败:', error)
    throw error
  }
}

/**
 * 创建审批
 * @param params 审批参数
 * @returns 创建结果
 */
export const createApproval = async (params: CreateApprovalParams): Promise<ApprovalItem> => {
  try {
    const response = (await post('/approvals', params)) as any
    return response.data
  } catch (error) {
    console.error('创建审批失败:', error)
    throw error
  }
}

/**
 * 审批操作
 * @param id 审批ID
 * @param action 审批动作
 * @param comment 审批意见
 * @returns 操作结果
 */
export const handleApproval = (
  id: string,
  action: 'approve' | 'reject',
  comment: string,
): Promise<any> => {
  return processApprovalDecision(id, {
    decision: action,
    decisionReason: comment,
  })
}

/**
 * 审批决定请求参数
 */
export interface ApprovalDecisionParams {
  decision: 'approve' | 'reject' | 'request_changes' | 'defer' | 'escalate' | 'reassign'
  decisionReason: string
  decisionComments?: string
  approvedConditions?: Record<string, any>
  imposedRequirements?: Record<string, any>
  followUpActions?: Array<Record<string, any>>
  supportingDocuments?: Array<Record<string, any>>
  evidenceReferences?: Array<Record<string, any>>
  nextApproverId?: string // 用于转派
}

/**
 * 处理审批决定（真实的API调用）
 * @param id 审批ID
 * @param params 审批决定参数
 * @returns 操作结果
 */
export const processApprovalDecision = async (
  id: string,
  params: ApprovalDecisionParams,
): Promise<ApprovalItem> => {
  try {
    console.log('🔍 处理审批决定 - ID:', id, '决定:', params.decision)
    const response = (await post(`/approvals/${id}/decision`, params)) as any
    console.log('✅ 审批决定响应:', response)
    return response
  } catch (error) {
    console.error('❌ 处理审批决定失败:', error)
    throw error
  }
}

/**
 * 提交审批申请（从草稿状态提交）
 * @param id 审批ID
 * @returns 提交结果
 */
export const submitApproval = async (id: string): Promise<ApprovalItem> => {
  try {
    console.log('📤 提交审批申请 - ID:', id)
    const response = (await post(`/approvals/${id}/submit`)) as any
    console.log('✅ 提交审批响应:', response)
    return response
  } catch (error) {
    console.error('❌ 提交审批失败:', error)
    throw error
  }
}

/**
 * 重新提交被驳回的审批申请
 * @param id 审批ID
 * @param revisionNote 修改说明
 * @returns 重新提交结果
 */
export const resubmitApproval = async (
  id: string,
  revisionNote?: string,
): Promise<ApprovalItem> => {
  try {
    console.log('🔄 重新提交审批 - ID:', id, '说明:', revisionNote)
    const response = (await post(`/approvals/${id}/resubmit`, {
      revision_note: revisionNote || '',
    })) as any
    console.log('✅ 重新提交响应:', response)
    return response
  } catch (error) {
    console.error('❌ 重新提交审批失败:', error)
    throw error
  }
}

/**
 * 更新审批申请（仅草稿或需要修改状态）
 * @param id 审批ID
 * @param params 更新参数
 * @returns 更新结果
 */
export const updateApproval = async (
  id: string,
  params: {
    title?: string
    content?: string
    metadata?: Record<string, any>
    attachments?: Array<Record<string, any>>
  },
): Promise<ApprovalItem> => {
  try {
    console.log('✏️ 更新审批申请 - ID:', id)
    const response = (await put(`/approvals/${id}`, params)) as any
    console.log('✅ 更新审批响应:', response)
    return response
  } catch (error) {
    console.error('❌ 更新审批失败:', error)
    throw error
  }
}

/**
 * 撤回审批
 * @param id 审批ID
 * @returns 操作结果
 */
export const cancelApproval = async (id: string): Promise<any> => {
  try {
    const response = (await post(`/approvals/${id}/cancel`)) as any
    return response.data
  } catch (error) {
    console.error('撤回审批失败:', error)
    throw error
  }
}

/**
 * 获取审批统计
 * @returns 统计数据
 */
export const getApprovalStats = async (): Promise<ApprovalStats> => {
  try {
    const response = (await get('/approvals/stats')) as any
    return response.data
  } catch (error) {
    console.error('获取审批统计失败:', error)
    throw error
  }
}

/**
 * 提交利益冲突审批申请
 * @param params 利益冲突审批参数
 * @returns 审批结果
 */
export const submitConflictApproval = async (
  params: ConflictApprovalParams,
): Promise<ConflictApprovalResult> => {
  try {
    const currentUserId = getCurrentUserId()
    console.log('提交利益冲突审批 - 当前用户ID:', currentUserId)
    console.log('提交参数:', params)

    // 转换为通用审批格式
    const approvalData: CreateApprovalParams = {
      type: 'conflict',
      title: params.caseTitle,
      content: params.conflictReason,
      category: 'conflict',
      applicant: params.applicant,
      applicantId: params.applicantId || currentUserId,
      department: params.department,
      departmentId: params.departmentId,
      urgency: params.urgency || 'normal',
      priority: params.priority || 'medium',
      workflowType: 'CONFLICT_APPROVAL',
      metadata: {
        conflictCases: params.conflictCases,
        riskLevel: params.riskLevel,
        additionalNotes: params.additionalNotes,
      },
    }

    const response = (await post('/approvals', approvalData)) as any

    // 修复：HTTP拦截器已经返回了data对象，所以直接访问response而不是response.data
    const result: ConflictApprovalResult = {
      approvalId: response.id,
      approvalNumber: `CO${new Date().getFullYear()}${String(response.id).padStart(6, '0')}`,
      status: response.status,
      submitTime: response.created_at || response.submission_date,
      expectedProcessingTime: params.urgency === 'urgent' ? '1-2个工作日' : '2-3个工作日',
    }

    console.log('利益冲突审批申请已提交:', result)
    return result
  } catch (error) {
    console.error('提交利益冲突审批失败:', error)
    throw error
  }
}

/**
 * 查询利益冲突审批状态
 * @param approvalId 审批ID
 * @returns 审批状态
 */
export const getConflictApprovalStatus = (approvalId: string): Promise<ConflictApprovalResult> => {
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      const result: ConflictApprovalResult = {
        approvalId,
        approvalNumber: `CO${new Date().getFullYear()}${String(approvalId).slice(-6)}`,
        status: 'pending',
        submitTime: new Date().toISOString(),
        expectedProcessingTime: '2-3个工作日',
      }
      resolve(result)
    }, 200)
  })
}

/**
 * 创建集成审批（包含冲突检测和案件创建配置）
 * @param data 集成审批请求数据
 * @returns 创建结果
 */
export const createIntegratedApproval = async (data: any): Promise<any> => {
  try {
    const response = (await post('/integration/approvals/with-conflict', data)) as any
    return response
  } catch (error) {
    console.error('创建集成审批失败:', error)
    throw error
  }
}
