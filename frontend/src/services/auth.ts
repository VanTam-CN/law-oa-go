import { post, get } from './http'
import {
  getToken as getStoredToken,
  removeToken as removeStoredToken,
  setToken as setStoredToken,
} from '@/utils/storage'

// 用户登录接口参数类型
interface LoginParams {
  email: string
  password: string
  remember?: boolean
}

// 登录响应类型
interface LoginResponse {
  token: string
  user: {
    id: number
    username: string
    real_name: string
    email: string
    role: string
    department: string
    [key: string]: any
  }
  data?: {
    token?: string
    user?: LoginResponse['user']
  }
}

// 用户信息类型
interface UserInfo {
  id: number
  username: string
  real_name: string
  email: string
  role: string
  department: string
  [key: string]: any
}

/**
 * 用户登录
 * @param data 登录参数
 * @returns 登录响应
 */
export const login = (data: LoginParams): Promise<LoginResponse> => {
  return post<LoginResponse>('/auth/login', data)
}

/**
 * 用户登出
 * @param token 当前访问令牌
 * @returns 登出响应
 */
export const logout = (token: string): Promise<any> => {
  return post('/auth/logout', { token })
}

/**
 * 获取当前用户信息
 * @returns 用户信息
 */
export const getCurrentUser = (): Promise<UserInfo> => {
  return get<UserInfo>('/users/me')
}

/**
 * 修改密码
 * @param data 修改密码参数
 * @returns 修改密码响应
 */
export const changePassword = (data: {
  old_password: string
  new_password: string
}): Promise<any> => {
  return post('/users/change-password', data)
}

/**
 * 获取存储的token
 * @returns token字符串或null
 */
export const getToken = (): string | null => {
  return getStoredToken()
}

/**
 * 存储token
 * @param token token字符串
 */
export const setToken = (token: string): void => {
  setStoredToken(token)
}

/**
 * 清除token
 */
export const clearToken = (): void => {
  removeStoredToken()
}
