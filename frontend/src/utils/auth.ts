// 认证工具，统一使用storage.ts中的函数
import { getToken, setToken, removeToken } from './storage'

// 为了向后兼容，保持原有API
export const setAuthToken = setToken
export const getAuthToken = getToken
export const removeAuthToken = removeToken
