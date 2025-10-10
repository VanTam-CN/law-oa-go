import { post, get, put } from '../services/http';

// 用户登录接口参数类型
interface LoginParams {
  email: string;
  password: string;
  remember?: boolean;
}

// 登录响应类型
interface LoginResponse {
  token: string;
  user: {
    id: number;
    username: string;
    real_name: string;
    email: string;
    role: string;
    department: string;
    [key: string]: any;
  };
}

// 用户信息类型
interface UserInfo {
  id: number;
  username: string;
  real_name: string;
  email: string;
  role: string;
  department: string;
  [key: string]: any;
}

/**
 * 用户登录
 * @param data 登录参数
 * @returns 登录响应
 */
export const login = (data: LoginParams): Promise<LoginResponse> => {
  return post<LoginResponse>('/auth/login', data);
};

/**
 * 用户登出
 * @returns 登出响应
 */
export const logout = (): Promise<any> => {
  return post('/auth/logout');
};

/**
 * 获取当前用户信息
 * @returns 用户信息
 */
export const getCurrentUser = (): Promise<UserInfo> => {
  return get<UserInfo>('/auth/me');
};

/**
 * 修改密码
 * @param data 修改密码参数
 * @returns 修改密码响应
 */
export const changePassword = (data: { old_password: string; new_password: string }): Promise<any> => {
  return post('/auth/change-password', data);
};

/**
 * 更新用户资料
 * @param data 用户资料数据
 * @returns 更新后的用户信息
 */
export const updateProfile = (data: Partial<UserInfo>): Promise<UserInfo> => {
  return put<UserInfo>('/auth/profile', data);
};

/**
 * 上传用户头像
 * @param file 头像文件
 * @returns 上传结果
 */
export const uploadAvatar = (file: File): Promise<any> => {
  const formData = new FormData();
  formData.append('avatar', file);
  return post('/auth/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
};