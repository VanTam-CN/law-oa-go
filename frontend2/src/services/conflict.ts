import { post } from './http';

export interface ConflictCheckParams {
  project_name: string;
  client_name: string;
  opposite_parties: string;
  project_type: string;
  team_members: string[];
  description?: string;
}

export interface ConflictResult {
  has_conflict: boolean;
  conflict_level: 'none' | 'low' | 'medium' | 'high';
  conflicts: {
    id: number;
    type: string;
    entity: string;
    project: string;
    level: 'low' | 'medium' | 'high';
    description: string;
  }[];
}

/**
 * 执行利益冲突检查
 * @param params 冲突检查参数
 * @returns 冲突检查结果
 */
export const performConflictCheck = (params: ConflictCheckParams): Promise<ConflictResult> => {
  return post<ConflictResult>('/conflict-check/check', params);
};

/**
 * 执行利益冲突预检
 * @param params 预检参数
 * @returns 预检结果
 */
export const performPreScreen = (params: {
  our_client_ids: number[];
  opponent_parties: string[];
  third_parties?: string[];
}): Promise<any> => {
  return post<any>('/conflict-check/pre-screen', params);
};