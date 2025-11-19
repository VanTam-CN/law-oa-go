/**
 * 豁免审批相关类型定义
 */

// 豁免申请状态
export enum WaiverStatus {
  DRAFT = 'DRAFT',                    // 草稿
  SUBMITTED = 'SUBMITTED',            // 已提交
  UNDER_REVIEW = 'UNDER_REVIEW',      // 审查中
  APPROVED = 'APPROVED',              // 已批准
  REJECTED = 'REJECTED',              // 已拒绝
  ESCALATED = 'ESCALATED',            // 已上报
  EXPIRED = 'EXPIRED',                // 已过期
  CANCELLED = 'CANCELLED'             // 已取消
}

// 豁免类型
export enum WaiverType {
  CONFLICT_OF_INTEREST = 'CONFLICT_OF_INTEREST',  // 利益冲突
  BUSINESS_RELATIONSHIP = 'BUSINESS_RELATIONSHIP',  // 业务关系
  REPRESENTATION_CONFLICT = 'REPRESENTATION_CONFLICT',  // 代理冲突
  FINANCIAL_INTEREST = 'FINANCIAL_INTEREST',  // 财务利益
  PERSONAL_RELATIONSHIP = 'PERSONAL_RELATIONSHIP',  // 个人关系
  ORGANIZATIONAL = 'ORGANIZATIONAL'  // 组织冲突
}

// 风险等级
export enum RiskLevel {
  LOW = 'LOW',          // 低风险
  MEDIUM = 'MEDIUM',    // 中风险
  HIGH = 'HIGH',        // 高风险
  CRITICAL = 'CRITICAL' // 关键风险
}

// 审批决策
export enum ApprovalDecision {
  APPROVE = 'APPROVE',      // 批准
  REJECT = 'REJECT',        // 拒绝
  ESCALATE = 'ESCALATE',    // 上报
  REQUEST_INFO = 'REQUEST_INFO'  // 要求更多信息
}

// 利益相关方类型
export enum StakeholderType {
  CLIENT = 'CLIENT',              // 客户
  LAWYER = 'LAWYER',              // 律师
  PARTY = 'PARTY',                // 当事人
  WITNESS = 'WITNESS',            // 证人
  EXPERT = 'EXPERT',              // 专家
  ORGANIZATION = 'ORGANIZATION'  // 组织
}

// 客户信息接口
export interface WaiverClient {
  id: string;
  name: string;
  type: 'INDIVIDUAL' | 'COMPANY';
  registrationNumber?: string;
  contactInfo?: string;
  businessNature?: string;
}

// 利益相关方接口
export interface Stakeholder {
  id: string;
  name: string;
  type: StakeholderType;
  organization?: string;
  relationshipDescription: string;
  contactInfo?: string;
  conflictDetails: string;
}

// 冲突检测详情
export interface ConflictDetectionDetail {
  type: string;
  description: string;
  riskLevel: RiskLevel;
  detectedDate: string;
  detectedBy: string;
  relatedParties: string[];
}

// 豁免申请接口
export interface WaiverApplication {
  id: string;
  caseId: string;
  caseTitle: string;
  waiverType: WaiverType;
  riskLevel: RiskLevel;
  applicantId: string;
  applicantName: string;
  description: string;
  justification: string;
  mitigationMeasures: string;
  affectedParties: string[];
  stakeholders: Stakeholder[];
  conflictDetectionDetails: ConflictDetectionDetail[];
  status: WaiverStatus;
  submittedAt: string;
  reviewedAt?: string;
  approvedAt?: string;
  expiresAt?: string;
  attachments: WaiverAttachment[];
  approvalRecords: WaiverApprovalRecord[];
  monitoringTasks: WaiverMonitoringTask[];
  metadata: {
    createdAt: string;
    updatedAt: string;
    version: number;
    createdBy: string;
    lastModifiedBy: string;
  };
}

// 豁免附件接口
export interface WaiverAttachment {
  id: string;
  fileName: string;
  originalName: string;
  fileSize: number;
  fileType: string;
  filePath: string;
  uploadedBy: string;
  uploadedAt: string;
  description?: string;
  isPublic: boolean;
}

// 审批记录接口
export interface WaiverApprovalRecord {
  id: string;
  applicationId: string;
  reviewerId: string;
  reviewerName: string;
  reviewerRole: string;
  decision: ApprovalDecision;
  comments: string;
  conditions?: string;
  reviewedAt: string;
  ipAddress: string;
  userAgent: string;
  isElectronicSignature: boolean;
  signatureData?: {
    signatureBase64: string;
    certificateId: string;
    timestamp: string;
    biometricData?: {
      fingerprintHash: string;
      facialRecognitionHash: string;
    };
  };
  escalatedTo?: {
    userId: string;
    userName: string;
    role: string;
  };
  nextReviewDate?: string;
}

// 监控任务接口
export interface WaiverMonitoringTask {
  id: string;
  applicationId: string;
  taskTitle: string;
  taskDescription: string;
  assignedTo: string;
  assignedToName: string;
  dueDate: string;
  frequency: 'DAILY' | 'WEEKLY' | 'MONTHLY' | 'QUARTERLY';
  isActive: boolean;
  lastCompletedAt?: string;
  nextDueDate: string;
  completionHistory: MonitoringTaskCompletion[];
}

// 监控任务完成记录
export interface MonitoringTaskCompletion {
  id: string;
  taskId: string;
  completedBy: string;
  completedByName: string;
  completedAt: string;
  notes: string;
  evidenceAttachments: string[];
  status: 'COMPLETED' | 'PARTIALLY_COMPLETED' | 'MISSED';
}

// 豁免模板接口
export interface WaiverTemplate {
  id: string;
  name: string;
  description: string;
  waiverType: WaiverType;
  riskLevel: RiskLevel;
  template: string;
  variables: TemplateVariable[];
  isActive: boolean;
  version: number;
  createdBy: string;
  createdAt: string;
  updatedBy: string;
  updatedAt: string;
}

// 模板变量接口
export interface TemplateVariable {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'date' | 'select' | 'textarea';
  label: string;
  required: boolean;
  defaultValue?: string;
  options?: string[];
  validation?: {
    pattern?: string;
    minLength?: number;
    maxLength?: number;
    min?: number;
    max?: number;
  };
}

// 统计数据接口
export interface WaiverStatistics {
  totalApplications: number;
  pendingApplications: number;
  approvedApplications: number;
  rejectedApplications: number;
  escalatedApplications: number;
  expiredApplications: number;
  averageReviewTime: number;
  approvalRate: number;
  rejectionRate: number;
  escalationRate: number;
  riskLevelDistribution: {
    [key in RiskLevel]: number;
  };
  waiverTypeDistribution: {
    [key in WaiverType]: number;
  };
  monthlyTrends: MonthlyTrend[];
}

// 月度趋势接口
export interface MonthlyTrend {
  month: string;
  year: number;
  applications: number;
  approvals: number;
  rejections: number;
  escalations: number;
}

// 创建豁免申请请求
export interface CreateWaiverApplicationRequest {
  caseId: string;
  waiverType: WaiverType;
  description: string;
  justification: string;
  mitigationMeasures: string;
  affectedParties: string[];
  stakeholders: Stakeholder[];
  attachments?: File[];
  submitImmediately?: boolean;
}

// 更新豁免申请请求
export interface UpdateWaiverApplicationRequest {
  id: string;
  description?: string;
  justification?: string;
  mitigationMeasures?: string;
  affectedParties?: string[];
  stakeholders?: Stakeholder[];
  attachments?: File[];
}

// 提交豁免申请请求
export interface SubmitWaiverApplicationRequest {
  id: string;
  submissionNotes?: string;
}

// 审批豁免申请请求
export interface ApproveWaiverApplicationRequest {
  applicationId: string;
  comments: string;
  conditions?: string;
  nextReviewDate?: string;
  isElectronicSignature: boolean;
  signatureData?: {
    signatureBase64: string;
    certificateId: string;
  };
}

// 拒绝豁免申请请求
export interface RejectWaiverApplicationRequest {
  applicationId: string;
  comments: string;
  reason: string;
  isElectronicSignature: boolean;
  signatureData?: {
    signatureBase64: string;
    certificateId: string;
  };
}

// 上报豁免申请请求
export interface EscalateWaiverApplicationRequest {
  applicationId: string;
  escalationReason: string;
  escalatedToUserId: string;
  comments: string;
  isElectronicSignature: boolean;
  signatureData?: {
    signatureBase64: string;
    certificateId: string;
  };
}

// 创建监控任务请求
export interface CreateMonitoringTaskRequest {
  applicationId: string;
  taskTitle: string;
  taskDescription: string;
  assignedTo: string;
  dueDate: string;
  frequency: 'DAILY' | 'WEEKLY' | 'MONTHLY' | 'QUARTERLY';
}

// 查询参数接口
export interface WaiverApplicationQueryParams {
  page?: number;
  pageSize?: number;
  status?: WaiverStatus[];
  waiverType?: WaiverType[];
  riskLevel?: RiskLevel[];
  applicantId?: string;
  caseId?: string;
  dateFrom?: string;
  dateTo?: string;
  keyword?: string;
  sortBy?: 'submittedAt' | 'reviewedAt' | 'approvedAt' | 'riskLevel';
  sortOrder?: 'asc' | 'desc';
}

// API响应接口
export interface WaiverApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
  error?: string;
}

// 分页响应接口
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 审批历史记录接口
export interface ApprovalHistory {
  applicationId: string;
  applicationTitle: string;
  currentStatus: WaiverStatus;
  submissionDate: string;
  lastReviewedDate?: string;
  approvalDate?: string;
  reviewerName?: string;
  daysInReview: number;
  currentAssignee?: string;
}

// 工作流状态接口
export interface WorkflowStatus {
  stage: string;
  status: 'PENDING' | 'IN_PROGRESS' | 'COMPLETED' | 'FAILED';
  assignedTo?: string;
  dueDate?: string;
  completedAt?: string;
  actions: WorkflowAction[];
}

// 工作流动作接口
export interface WorkflowAction {
  id: string;
  name: string;
  description: string;
  isAvailable: boolean;
  requiresApproval: boolean;
  permissions: string[];
}