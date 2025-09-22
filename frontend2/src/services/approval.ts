import { get, post, put, del } from './http';

export interface ApprovalItem {
  id: number;
  type: string;
  title: string;
  content: string;
  applicant: string;
  applicantId: number;
  department: string;
  createTime: string;
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  urgency: 'normal' | 'urgent' | 'very_urgent';
  currentApprover?: string;
  currentApproverId?: number;
}

export interface ApprovalDetail extends ApprovalItem {
  records: ApprovalRecord[];
}

export interface ApprovalRecord {
  id: number;
  approvalId: number;
  approver: string;
  approverId: number;
  action: 'approve' | 'reject';
  comment: string;
  createTime: string;
}

export interface ApprovalStats {
  pendingCount: number;
  myPendingCount: number;
  myTotalCount: number;
}

export interface CreateApprovalParams {
  type: string;
  title: string;
  content: string;
  applicant: string;
  applicantId: number;
  department: string;
  urgency: 'normal' | 'urgent' | 'very_urgent';
}

/**
 * 获取审批列表
 * @param type 类型：pending-待我审批，my-我的申请
 * @returns 审批列表
 */
export const getApprovals = (type: 'pending' | 'my'): Promise<ApprovalItem[]> => {
  return get<ApprovalItem[]>(`/approvals/${type}`);
};

/**
 * 获取审批详情
 * @param id 审批ID
 * @returns 审批详情
 */
export const getApprovalDetail = (id: number): Promise<ApprovalDetail> => {
  return get<ApprovalDetail>(`/approvals/detail/${id}`);
};

/**
 * 创建审批
 * @param params 审批参数
 * @returns 创建结果
 */
export const createApproval = (params: CreateApprovalParams): Promise<ApprovalItem> => {
  return post<ApprovalItem>('/approvals', params);
};

/**
 * 审批操作
 * @param id 审批ID
 * @param action 审批动作
 * @param comment 审批意见
 * @returns 操作结果
 */
export const handleApproval = (id: number, action: 'approve' | 'reject', comment: string): Promise<any> => {
  return post<any>(`/approvals/${id}/${action}`, { comment });
};

/**
 * 撤回审批
 * @param id 审批ID
 * @returns 操作结果
 */
export const cancelApproval = (id: number): Promise<any> => {
  return post<any>(`/approvals/${id}/cancel`);
};

/**
 * 获取审批统计
 * @returns 统计数据
 */
export const getApprovalStats = (): Promise<ApprovalStats> => {
  return get<ApprovalStats>('/approvals/stats');
};