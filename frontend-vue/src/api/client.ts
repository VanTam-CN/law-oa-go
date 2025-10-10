import { get, post, put, del } from '../services/http';
import { clientService } from '../services/client';

export { clientService };

// 客户信息类型
interface Client {
  id: number;
  name: string;
  contact_person: string;
  phone: string;
  email: string;
  address: string;
  company_type: string;
  industry: string;
  status: string;
  created_at: string;
  updated_at: string;
  [key: string]: any;
}

// 客户列表响应类型
interface ClientListResponse {
  clients: Client[];
  total: number;
  page: number;
  page_size: number;
}

// 创建客户参数
interface CreateClientParams {
  name: string;
  contact_person: string;
  phone: string;
  email: string;
  address: string;
  company_type: string;
  industry: string;
}

// 更新客户参数
interface UpdateClientParams {
  id: number;
  name?: string;
  contact_person?: string;
  phone?: string;
  email?: string;
  address?: string;
  company_type?: string;
  industry?: string;
  status?: string;
}

/**
 * 获取客户列表
 * @param params 查询参数
 * @returns 客户列表
 */
export const getClients = (params?: { page?: number; page_size?: number; keyword?: string }): Promise<ClientListResponse> => {
  return get<ClientListResponse>('/clients', params);
};

/**
 * 获取单个客户信息
 * @param id 客户ID
 * @returns 客户信息
 */
export const getClient = (id: number): Promise<Client> => {
  return get<Client>(`/clients/${id}`);
};

/**
 * 创建客户
 * @param data 创建客户参数
 * @returns 创建的客户信息
 */
export const createClient = (data: CreateClientParams): Promise<Client> => {
  return post<Client>('/clients', data);
};

/**
 * 更新客户
 * @param id 客户ID
 * @param data 更新参数
 * @returns 更新后的客户信息
 */
export const updateClient = (id: number, data: UpdateClientParams): Promise<Client> => {
  return put<Client>(`/clients/${id}`, data);
};

/**
 * 删除客户
 * @param id 客户ID
 * @returns 删除结果
 */
export const deleteClient = (id: number): Promise<any> => {
  return del(`/clients/${id}`);
};

/**
 * 获取客户统计信息
 * @returns 统计信息
 */
export const getClientStats = (): Promise<any> => {
  return get('/clients/stats');
};