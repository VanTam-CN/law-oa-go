import { get, post, put, del } from './http';

export interface Client {
  id?: number;
  name: string;
  type: '个人' | '企业';
  phone: string;
  email: string;
  idCard?: string;
  address: string;
  company?: string;
  industry?: string;
  contactPerson?: string;
  contactPhone?: string;
  status: 'active' | 'inactive';
  source?: string;
  createdAt?: string;
  updatedAt?: string;
  notes?: string;
}

export interface ClientListResponse {
  total: number;
  pageNum: number;
  pageSize: number;
  list: Client[];
}

export interface ClientStats {
  total: number;
  monthlyNew: number;
  typeStats: Record<string, number>;
  statusStats: Record<string, number>;
  sourceStats: Record<string, number>;
}

export interface ClientQueryParams {
  name?: string;
  type?: string;
  status?: string;
  pageNum?: number;
  pageSize?: number;
}

export const clientService = {
  // 获取客户列表
  getClientList: (params: ClientQueryParams) => {
    // 参数映射：前端字段名 -> 后端期望的字段名
    const mappedParams = {
      page: params.pageNum || 1,
      page_size: Math.min(params.pageSize || 10, 100), // 🔧 修复：限制最大值为100，符合后端验证规则
      search: params.name, // 🔧 修复：将前端的name映射为后端的search
      type: params.type,
      status: params.status
    };
    // 移除空值参数，但保留空字符串以支持搜索所有客户
    const filteredParams = Object.fromEntries(
      Object.entries(mappedParams).filter(([_, value]) => value !== undefined)
    );
    console.log('客户搜索参数:', filteredParams); // 调试日志
    return get<ClientListResponse>('/clients', filteredParams);
  },

  // 获取客户详情
  getClient: (id: number) => get<Client>(`/clients/${id}`),

  // 新增客户
  addClient: (data: Client) => post<Client>('/clients', data),

  // 兼容前端页面调用的方法名
  createClient: (data: Client) => post<Client>('/clients', data),

  // 更新客户信息
  updateClient: (id: number, data: Client) => put<Client>(`/clients/${id}`, data),

  // 删除客户
  deleteClient: (id: number) => del(`/clients/${id}`),

  // 获取客户统计
  getClientStats: () => get<ClientStats>('/clients/stats'),
};