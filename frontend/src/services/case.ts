import { get, post, put, del } from './http';

export interface Case {
  id?: number;
  title: string;
  caseNumber: string;
  clientId: number;
  clientName: string;
  lawyerId: number;
  lawyerName: string;
  type: string;
  status: '待处理' | '进行中' | '已完成' | '已关闭';
  priority: '低' | '中' | '高' | '紧急';
  description: string;
  startDate: string;
  expectedEndDate: string;
  actualEndDate?: string;
  createdAt?: string;
  updatedAt?: string;
  remark?: string;
}

export interface CaseListResponse {
  total: number;
  pageNum: number;
  pageSize: number;
  list: Case[];
}

export interface CaseStats {
  total: number;
  active: number;
  completed: number;
  typeStats: Record<string, number>;
  statusStats: Record<string, number>;
  priorityStats: Record<string, number>;
}

export interface CaseQueryParams {
  keyword?: string;
  clientId?: number;
  lawyerId?: number;
  type?: string;
  status?: string;
  priority?: string;
  startDate?: string;
  endDate?: string;
  pageNum?: number;
  pageSize?: number;
}

export const caseService = {
  // 获取案件列表
  getCaseList: (params: CaseQueryParams) =>
    get<CaseListResponse>('/cases', params),

  // 获取案件详情
  getCase: (id: number) => get<Case>(`/cases/${id}`),

  // 新增案件 - 修正数据格式
  addCase: (data: any) => {
    const submitData = {
      title: data.caseName || data.title,
      description: data.description || '',
      client_id: data.clientId,
      lawyer_id: data.lawyerId,
      case_type: data.caseType,
      priority: data.priority || 'medium',
      status: data.status || 'pending'
    };
    return post<Case>('/cases', submitData);
  },

  // 更新案件信息 - 修正数据格式
  updateCase: (id: number, data: any) => {
    const updateData = {
      title: data.caseName || data.title,
      description: data.description,
      lawyer_id: data.lawyerId,
      case_type: data.caseType,
      priority: data.priority,
      status: data.status
    };
    return put<Case>(`/cases/${id}`, updateData);
  },

  // 删除案件
  deleteCase: (id: number) => del(`/cases/${id}`),

  // 分配律师
  assignLawyer: (id: number, lawyerId: number) =>
    post(`/cases/${id}/assign-lawyer`, { lawyer_id: lawyerId }),

  // 更新案件状态
  updateStatus: (id: number, status: string) =>
    put(`/cases/${id}/status`, { status }),

  // 获取案件统计
  getCaseStats: () => get<CaseStats>('/cases/stats'),
};