import { get, post, put, del } from './http';

export interface Lawyer {
  lawyerId?: number;
  lawyerName: string;
  phone: string;
  email: string;
  licenseNo: string;
  specialty: string;
  department: string;
  position: string;
  delFlag: string;
  // 前端需要的额外字段
  id?: number;
  name?: string;
  licenseNumber?: string;
  gender?: 'male' | 'female';
  experience?: number;
  status?: 'active' | 'inactive' | 'on_leave';
  joinDate?: string;
  profile?: string;
  avatar?: string;
}

export interface LawyerStats {
  total: number;
  active: number;
  inactive: number;
  onLeave: number;
  departmentStats: Record<string, number>;
  specialtyStats: Record<string, number>;
}

/**
 * 获取律师列表
 */
export const getLawyerList = (params?: any): Promise<{ list: Lawyer[]; total: number }> => {
  return get<{ list: Lawyer[]; total: number }>('/lawfirm/lawyers', params);
};

/**
 * 获取律师详情
 */
export const getLawyerDetail = (id: number): Promise<Lawyer> => {
  return get<Lawyer>(`/lawfirm/lawyers/${id}`);
};

/**
 * 新增律师
 */
export const addLawyer = (data: Lawyer): Promise<Lawyer> => {
  return post<Lawyer>('/lawfirm/lawyers', data);
};

/**
 * 更新律师信息
 */
export const updateLawyer = (id: number, data: Lawyer): Promise<Lawyer> => {
  return put<Lawyer>(`/lawfirm/lawyers/${id}`, data);
};

/**
 * 删除律师
 */
export const deleteLawyer = (id: number): Promise<void> => {
  return del<void>(`/lawfirm/lawyers/${id}`);
};

/**
 * 获取律师统计信息
 */
export const getLawyerStats = (): Promise<LawyerStats> => {
  return get<LawyerStats>('/lawfirm/lawyers/stats');
};

/**
 * 更新律师状态
 */
export const updateLawyerStatus = (id: number, status: string): Promise<void> => {
  return put<void>(`/lawfirm/lawyers/${id}/status`, { status });
};

// 导出所有服务函数作为lawyerService对象
export const lawyerService = {
  getLawyerList,
  getLawyerDetail,
  addLawyer,
  updateLawyer,
  deleteLawyer,
  getLawyerStats,
  updateLawyerStatus
};