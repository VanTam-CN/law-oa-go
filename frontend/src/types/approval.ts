/**
 * 审批流程相关类型定义
 * 与后端 API 完全对齐
 */

// 审批状态枚举
export enum ApprovalStatus {
  DRAFT = 'draft', // 草稿
  SUBMITTED = 'submitted', // 已提交
  UNDER_REVIEW = 'under_review', // 审核中
  APPROVED = 'approved', // 已通过
  REJECTED = 'rejected', // 已拒绝
  CANCELLED = 'cancelled', // 已撤回
  EXPIRED = 'expired', // 已过期
  PENDING = 'pending', // 待审批
  NEEDS_REVISION = 'needs_revision', // 需要修改
  RESUBMITTED = 'resubmitted', // 重新提交
}

// 紧急程度枚举
export enum ApprovalUrgency {
  NORMAL = 'normal', // 普通
  URGENT = 'urgent', // 紧急
  VERY_URGENT = 'very_urgent', // 特急
}

// 优先级枚举
export enum ApprovalPriority {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  CRITICAL = 'critical',
}

// 审批类型枚举
export enum ApprovalType {
  SEAL = 'seal', // 用章申请
  CASE_REGISTRATION = 'case_registration', // 立案申请
  CONFLICT = 'conflict', // 利益冲突审批
  INVOICE = 'invoice', // 开票申请
  CONTRACT = 'contract', // 合同变更
  BID = 'bid', // 投标申请
  LEAVE = 'leave', // 请假申请
  REIMBURSEMENT = 'reimbursement', // 报销申请
  PROCUREMENT = 'procurement', // 采购申请
}

// 审批决定类型
export enum ApprovalDecision {
  APPROVE = 'approve', // 通过
  REJECT = 'reject', // 拒绝
  REQUEST_CHANGES = 'request_changes', // 要求修改
  DEFER = 'defer', // 延期
  ESCALATE = 'escalate', // 上报
  REASSIGN = 'reassign', // 转派
}

// 审批工作流类型
export enum WorkflowType {
  CONFLICT_APPROVAL = 'CONFLICT_APPROVAL', // 利益冲突审批
  SEAL_APPROVAL = 'SEAL_APPROVAL', // 用章审批
  CASE_APPROVAL = 'CASE_APPROVAL', // 立案审批
  GENERAL_APPROVAL = 'GENERAL_APPROVAL', // 通用审批
}

// 审批项目（列表项）
export interface ApprovalItem {
  id: string
  type: string
  title: string
  content: string
  applicant: string
  applicantId: string
  department: string
  createTime: string
  submissionDate?: string
  status: ApprovalStatus
  urgency: ApprovalUrgency
  priority?: ApprovalPriority
  currentApprover?: string
  currentApproverId?: string
  requestNumber?: string
  currentStage?: string
  workflowType?: string
  metadata?: string
}

// 审批详情
export interface ApprovalDetail extends ApprovalItem {
  records: ApprovalRecord[]
  workflow?: WorkflowInfo
}

// 审批记录
export interface ApprovalRecord {
  id: string
  approvalRequestID: string
  approver: string
  approverId: string
  decision: ApprovalDecision
  decisionReason: string
  decisionComments?: string
  approvalDate: string
  stage: string
  stageOrder: number
  status: string
}

// 工作流信息
export interface WorkflowInfo {
  id: string
  name: string
  type: string
  stages: WorkflowStage[]
  currentStage?: string
}

// 工作流阶段
export interface WorkflowStage {
  id: string
  name: string
  order: number
  approvers: Approver[]
  status: string
}

// 审批人信息
export interface Approver {
  id: string
  name: string
  department: string
  position?: string
  avatar?: string
}

// 审批统计
export interface ApprovalStats {
  totalRequests: number
  pendingRequests: number
  myPendingRequests: number
  approvedRequests: number
  rejectedRequests: number
}

// 创建审批参数
export interface CreateApprovalParams {
  type: string
  title: string
  content: string
  category?: string
  applicant: string
  applicantId: string
  department: string
  departmentId?: string
  urgency: ApprovalUrgency
  priority?: ApprovalPriority
  workflowType?: WorkflowType
  expectedEffectiveDate?: string
  expectedExpiryDate?: string
  durationDays?: number
  attachments?: Attachment[]
  metadata?: Record<string, any>
}

// 附件信息
export interface Attachment {
  id: string
  name: string
  url: string
  type: string
  size: number
  uploadTime: string
}

// 审批决定参数
export interface ApprovalDecisionParams {
  decision: ApprovalDecision
  decisionReason: string
  decisionComments?: string
  approvedConditions?: Record<string, any>
  imposedRequirements?: Record<string, any>
  followUpActions?: Record<string, any>[]
  supportingDocuments?: Attachment[]
  evidenceReferences?: string[]
  nextApproverId?: string
}

// 利益冲突审批参数
export interface ConflictApprovalParams {
  caseId: string
  caseTitle: string
  conflictReason: string
  riskLevel: string
  conflictCases: ConflictCaseInfo[]
  applicant: string
  applicantId: string
  department: string
  departmentId?: string
  urgency?: ApprovalUrgency
  priority?: ApprovalPriority
  additionalNotes?: string
}

// 冲突案件信息
export interface ConflictCaseInfo {
  caseId: string
  caseName: string
  conflictType: string
  riskLevel: string
  description: string
}

// 利益冲突审批结果
export interface ConflictApprovalResult {
  approvalId: string
  approvalNumber: string
  status: ApprovalStatus
  submitTime: string
  expectedProcessingTime: string
}

// API 响应包装器
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: ApiError
  message: string
  requestId?: string
  timestamp: string
}

// API 错误类型
export interface ApiError {
  code: string
  message: string
  field?: string
  details?: Record<string, any>
}

// 审批列表查询参数
export interface ApprovalListQuery {
  type?: 'pending' | 'my' | 'approved' | 'rejected'
  status?: ApprovalStatus
  page?: number
  pageSize?: number
  applicantId?: string
  startDate?: string
  endDate?: string
}

// 用章申请专用参数
export interface SealApprovalParams extends CreateApprovalParams {
  sealType: 'official' | 'contract' | 'financial' | 'other'
  sealCount: number
  documentName: string
  documentUrl?: string
  usageReason: string
}

// 立案申请专用参数
export interface CaseApprovalParams extends CreateApprovalParams {
  caseName: string
  caseType: string
  clientName: string
  estimatedFee?: number
  leadLawyer?: string
  caseDescription: string
  conflictCheckPassed?: boolean
  conflictCheckId?: string
}
