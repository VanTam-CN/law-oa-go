import { get, post } from './http';

export interface ProjectType {
  id: number;
  name: string;
  code: string;
  description?: string;
}

export interface ProjectInfo {
  id: number;
  project_code: string;
  name: string;
  client_id: number;
  opposite_party: string;
  lawyer_id: number;
  team_members: number[];
  status: string;
  project_type: string;
  contract_amount: number;
  start_date: string;
  end_date: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export const getProjectTypes = (): Promise<ProjectType[]> => {
  return get<ProjectType[]>('/projects/types');
};

/**
 * 获取项目列表
 * @param params 查询参数
 * @returns 项目列表
 */
export const getProjectList = (params?: any): Promise<{ list: ProjectInfo[]; total: number }> => {
  return get<{ list: ProjectInfo[]; total: number }>('/projects', params);
};

/**
 * 获取项目详情
 * @param id 项目ID
 * @returns 项目详情
 */
export const getProjectDetail = (id: number): Promise<ProjectInfo> => {
  return get<ProjectInfo>(`/projects/${id}`);
};

/**
 * 新增项目
 * @param data 项目数据
 * @returns 新增的项目
 */
export const addProject = (data: ProjectInfo): Promise<ProjectInfo> => {
  return post<ProjectInfo>('/projects', data);
};

/**
 * 更新项目信息
 * @param id 项目ID
 * @param data 项目数据
 * @returns 更新后的项目
 */
export const updateProject = (id: number, data: ProjectInfo): Promise<ProjectInfo> => {
  return post<ProjectInfo>(`/projects/${id}`, data);
};