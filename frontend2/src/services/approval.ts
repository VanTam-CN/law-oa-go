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

// 模拟数据
const mockApprovals: ApprovalItem[] = [
  {
    id: 1,
    type: 'leave',
    title: '请假申请',
    content: '因个人原因需请假3天',
    applicant: '张三',
    applicantId: 1,
    department: '技术部',
    createTime: '2024-01-15T09:00:00Z',
    status: 'pending',
    urgency: 'normal',
    currentApprover: '李四',
    currentApproverId: 2
  },
  {
    id: 2,
    type: 'expense',
    title: '费用报销',
    content: '出差费用报销，共计2800元',
    applicant: '王五',
    applicantId: 3,
    department: '市场部',
    createTime: '2024-01-14T14:30:00Z',
    status: 'pending',
    urgency: 'urgent',
    currentApprover: '赵六',
    currentApproverId: 4
  }
];

const mockApprovalStats = {
  pendingCount: 5,
  myPendingCount: 2,
  myTotalCount: 12
};

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
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      if (type === 'pending') {
        resolve(mockApprovals.filter(item => item.status === 'pending'));
      } else {
        resolve(mockApprovals);
      }
    }, 300);
  });
};

/**
 * 获取审批详情
 * @param id 审批ID
 * @returns 审批详情
 */
export const getApprovalDetail = (id: number): Promise<ApprovalDetail> => {
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      const approval = mockApprovals.find(item => item.id === id);
      if (approval) {
        resolve({
          ...approval,
          records: [
            {
              id: 1,
              approvalId: id,
              approver: '李四',
              approverId: 2,
              action: 'approve' as const,
              comment: '同意申请',
              createTime: '2024-01-15T10:00:00Z'
            }
          ]
        });
      } else {
        throw new Error('审批不存在');
      }
    }, 300);
  });
};

/**
 * 创建审批
 * @param params 审批参数
 * @returns 创建结果
 */
export const createApproval = (params: CreateApprovalParams): Promise<ApprovalItem> => {
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      const newApproval: ApprovalItem = {
        id: Date.now(),
        ...params,
        createTime: new Date().toISOString(),
        status: 'pending'
      };
      resolve(newApproval);
    }, 300);
  });
};

/**
 * 审批操作
 * @param id 审批ID
 * @param action 审批动作
 * @param comment 审批意见
 * @returns 操作结果
 */
export const handleApproval = (id: number, action: 'approve' | 'reject', comment: string): Promise<any> => {
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
export const cancelApproval = (id: number): Promise<any> => {
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({ success: true, message: '撤回成功' });
    }, 300);
  });
};

/**
 * 获取审批统计
 * @returns 统计数据
 */
export const getApprovalStats = (): Promise<ApprovalStats> => {
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(mockApprovalStats);
    }, 300);
  });
};