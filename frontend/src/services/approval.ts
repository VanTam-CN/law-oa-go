import { get, post, put, del } from './http';
import { getUserInfo } from '@/utils/storage';

export interface ApprovalItem {
  id: string; // 后端使用UUID字符串
  type: string;
  title: string;
  content: string;
  applicant: string;
  applicantId: string; // 后端使用字符串ID
  department: string;
  createTime: string;
  status: 'draft' | 'submitted' | 'under_review' | 'approved' | 'rejected' | 'cancelled' | 'expired';
  urgency: 'normal' | 'urgent' | 'very_urgent';
  priority: 'low' | 'medium' | 'high' | 'critical';
  currentApprover?: string;
  currentApproverId?: string;
  requestNumber?: string;
  submissionDate?: string;
  currentStage?: string;
  workflowType?: string;
}

// 获取当前用户ID的函数
const getCurrentUserId = (): string => {
  const userInfo = getUserInfo();

  console.log('🔍 getCurrentUserId - 原始用户信息:', userInfo);
  console.log('🔍 getCurrentUserId - 用户ID:', userInfo?.id);
  console.log('🔍 getCurrentUserId - 用户ID类型:', typeof userInfo?.id);

  // 如果有真实用户信息，使用真实用户ID
  if (userInfo?.id !== undefined && userInfo?.id !== null) {
    const userIdStr = userInfo.id.toString();
    console.log('✅ getCurrentUserId - 使用真实用户ID:', userIdStr);
    return userIdStr;
  }

  // 如果没有用户信息，记录警告并使用默认值
  console.warn('⚠️ getCurrentUserId - 没有找到用户信息，使用默认值');

  // 开发环境默认用户ID
  const isDevMode = process.env.NODE_ENV === 'development';
  if (isDevMode) {
    console.log('🛠️ 开发模式：使用默认测试用户ID');
    return '1';
  }

  // 生产环境默认值
  return '1';
};

export interface ApprovalDetail extends ApprovalItem {
  records: ApprovalRecord[];
}

export interface ApprovalRecord {
  id: string;
  approvalRequestID: string;
  approver: string;
  approverId: string;
  decision: 'approve' | 'reject' | 'request_changes' | 'reassign';
  decisionReason: string;
  decisionComments: string;
  approvalDate: string;
  stage: string;
  stageOrder: number;
  status: string;
}

export interface ApprovalStats {
  totalRequests: number;
  pendingRequests: number;
  myPendingRequests: number;
  approvedRequests: number;
  rejectedRequests: number;
}

export interface CreateApprovalParams {
  type: string;
  title: string;
  content: string;
  category?: string;
  applicant: string;
  applicantId: string;
  department: string;
  departmentId?: string;
  urgency: 'normal' | 'urgent' | 'very_urgent';
  priority?: 'low' | 'medium' | 'high' | 'critical';
  workflowType?: string;
  expectedEffectiveDate?: string;
  expectedExpiryDate?: string;
  durationDays?: number;
  attachments?: any[];
  metadata?: any;
}

// 利益冲突审批相关接口
export interface ConflictApprovalParams {
  caseId: string;
  caseTitle: string;
  conflictReason: string;
  riskLevel: string;
  conflictCases: Array<{
    caseId: string;
    caseName: string;
    conflictType: string;
    riskLevel: string;
    description: string;
  }>;
  applicant: string;
  applicantId: string;
  department: string;
  departmentId?: string;
  urgency?: 'normal' | 'urgent' | 'very_urgent';
  priority?: 'low' | 'medium' | 'high' | 'critical';
  additionalNotes?: string;
}

export interface ConflictApprovalResult {
  approvalId: string;
  approvalNumber: string;
  status: 'draft' | 'submitted' | 'under_review' | 'approved' | 'rejected' | 'cancelled' | 'expired';
  submitTime: string;
  expectedProcessingTime: string;
}

/**
 * 获取审批列表
 * @param type 类型：pending-待我审批，my-我的申请
 * @returns 审批列表
 */
export const getApprovals = async (type: 'pending' | 'my'): Promise<ApprovalItem[]> => {
  try {
    console.log('🚀 获取审批列表 - 类型:', type);

    if (type === 'pending') {
      // 待我审批
      console.log('📋 获取待我审批的列表...');
      const response = await get('/approvals/pending');
      return response.list || response;
    } else {
      // 我的申请
      const currentUserId = getCurrentUserId();
      const requestUrl = '/approvals?applicantId=' + currentUserId;
      console.log('📝 获取我的申请列表 - 用户ID:', currentUserId);
      console.log('📝 请求URL:', requestUrl);
      const response = await get(requestUrl);
      console.log('✅ 我的申请列表响应:', response);
      return response.list || response;
    }
  } catch (error) {
    console.error('❌ 获取审批列表失败:', error);
    throw error;
  }
};

/**
 * 获取审批详情
 * @param id 审批ID
 * @returns 审批详情
 */
export const getApprovalDetail = async (id: string): Promise<ApprovalDetail> => {
  try {
    console.log('🔍 获取审批详情 - ID:', id);
    console.log('🔍 ID类型:', typeof id);
    const response = await get(`/approvals/${id}`);
    console.log('✅ 审批详情响应:', response);
    return response; // HTTP拦截器已经提取了data字段
  } catch (error) {
    console.error('❌ 获取审批详情失败:', error);
    throw error;
  }
};

/**
 * 创建审批
 * @param params 审批参数
 * @returns 创建结果
 */
export const createApproval = async (params: CreateApprovalParams): Promise<ApprovalItem> => {
  try {
    const response = await post('/approvals', params);
    return response.data;
  } catch (error) {
    console.error('创建审批失败:', error);
    throw error;
  }
};

/**
 * 审批操作
 * @param id 审批ID
 * @param action 审批动作
 * @param comment 审批意见
 * @returns 操作结果
 */
export const handleApproval = (id: string, action: 'approve' | 'reject', comment: string): Promise<any> => {
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({ success: true, message: `审批${action === 'approve' ? '通过' : '拒绝'}成功` });
    }, 300);
  });
};

/**
 * 撤回审批
 * @param id 审批ID
 * @returns 操作结果
 */
export const cancelApproval = async (id: string): Promise<any> => {
  try {
    const response = await post(`/approvals/${id}/cancel`);
    return response.data;
  } catch (error) {
    console.error('撤回审批失败:', error);
    throw error;
  }
};

/**
 * 获取审批统计
 * @returns 统计数据
 */
export const getApprovalStats = async (): Promise<ApprovalStats> => {
  try {
    const response = await get('/approvals/stats');
    return response.data;
  } catch (error) {
    console.error('获取审批统计失败:', error);
    throw error;
  }
};

/**
 * 提交利益冲突审批申请
 * @param params 利益冲突审批参数
 * @returns 审批结果
 */
export const submitConflictApproval = async (params: ConflictApprovalParams): Promise<ConflictApprovalResult> => {
  try {
    const currentUserId = getCurrentUserId();
    console.log('提交利益冲突审批 - 当前用户ID:', currentUserId);
    console.log('提交参数:', params);

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
        additionalNotes: params.additionalNotes
      }
    };

    const response = await post('/approvals', approvalData);

    // 修复：HTTP拦截器已经返回了data对象，所以直接访问response而不是response.data
    const result: ConflictApprovalResult = {
      approvalId: response.id,
      approvalNumber: `CO${new Date().getFullYear()}${String(response.id).padStart(6, '0')}`,
      status: response.status,
      submitTime: response.created_at || response.submission_date,
      expectedProcessingTime: params.urgency === 'urgent' ? '1-2个工作日' : '2-3个工作日'
    };

    console.log('利益冲突审批申请已提交:', result);
    return result;
  } catch (error) {
    console.error('提交利益冲突审批失败:', error);
    throw error;
  }
};

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
        expectedProcessingTime: '2-3个工作日'
      };
      resolve(result);
    }, 200);
  });
};