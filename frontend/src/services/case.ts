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
  // 获取案件列表 - 修正参数格式
  getCaseList: (params: any) => {
    // 转换参数格式以匹配后端API
    const apiParams: any = {
      page: params.page || 1,
      page_size: params.page_size || 10,
      search: params.search,
      status: params.status,
      case_type: params.case_type,
      priority: params.priority,
      lawyer_id: params.lawyer_id,
      client_id: params.client_id
    };

    // 清理和验证参数
    Object.keys(apiParams).forEach(key => {
      const value = apiParams[key];
      if (value === undefined || value === null || value === '') {
        delete apiParams[key];
      } else if ((key === 'lawyer_id' || key === 'client_id') && typeof value === 'string') {
        // 确保ID字段为数字类型
        const numValue = Number(value);
        if (!isNaN(numValue) && numValue > 0) {
          apiParams[key] = numValue;
        } else {
          delete apiParams[key];
        }
      }
    });

    console.log('案件筛选API参数:', apiParams);
    return get<any>('/cases', apiParams);
  },

  // 获取案件详情
  getCase: (id: number) => get<Case>(`/cases/${id}`),

  // 新增案件 - 修正数据格式
  addCase: (data: any) => {
    const submitData = {
      title: data.caseName || data.title,
      description: data.description || '',
      client_id: Number(data.clientId),
      lawyer_id: Number(data.lawyerId),
      case_type: data.caseType,
      priority: data.priority || 'medium',
      status: data.status || 'pending'
    };

    // 验证必填字段
    if (!submitData.title || submitData.title.trim() === '') {
      throw new Error('案件名称不能为空');
    }
    if (!submitData.client_id || isNaN(submitData.client_id)) {
      throw new Error('请选择有效的委托客户');
    }
    if (!submitData.lawyer_id || isNaN(submitData.lawyer_id)) {
      throw new Error('请选择有效的负责律师');
    }
    if (!submitData.case_type) {
      throw new Error('请选择案件类型');
    }

    console.log('提交案件数据:', submitData);
    return post<Case>('/cases', submitData);
  },

  // 更新案件信息 - 修正数据格式
  updateCase: (id: number, data: any) => {
    const updateData: any = {
      title: data.caseName || data.title,
      description: data.description,
      case_type: data.caseType,
      priority: data.priority,
      status: data.status
    };

    // 只有当lawyerId不为空且是有效数字时才包含它
    if (data.lawyerId !== null && data.lawyerId !== undefined && !isNaN(Number(data.lawyerId))) {
      updateData.lawyer_id = Number(data.lawyerId);
    }

    console.log('更新案件数据:', updateData);
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