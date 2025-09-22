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
  remark?: string;
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
  getClientList: (params: ClientQueryParams) => 
    get<ClientListResponse>('/clients', params),
  
  // 获取客户详情
  getClient: (id: number) => get<Client>(`/clients/${id}`),
  
  // 新增客户
  addClient: (data: Client) => post<Client>('/clients', data),
  
  // 更新客户信息
  updateClient: (id: number, data: Client) => put<Client>(`/clients/${id}`, data),
  
  // 删除客户
  deleteClient: (id: number) => del(`/clients/${id}`),
  
  // 获取客户统计
  getClientStats: () => get<ClientStats>('/clients/stats'),
};