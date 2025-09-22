import { get } from './http';

export interface UserInfo {
  id: number;
  username: string;
  real_name: string;
  email: string;
  phone: string;
  status: string;
  created_at: string;
  updated_at: string;
}

/**
 * 获取所有用户列表
 * @returns 用户列表
 */
export const getUserList = (): Promise<UserInfo[]> => {
  return get<UserInfo[]>('/admin/employees');
};

/**
 * 获取当前用户信息
 * @returns 当前用户信息
 */
export const getCurrentUser = (): Promise<UserInfo> => {
  return get<UserInfo>('/users/me');
};