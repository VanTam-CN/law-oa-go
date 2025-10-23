// 认证工具，统一使用storage.ts中的函数
import { getToken, setToken, removeToken } from './storage';

// 为了向后兼容，保持原有API
export const setAuthToken = setToken;
export const getAuthToken = getToken;
export const removeAuthToken = removeToken;

// 临时设置一个测试token
export const setTestToken = () => {
  // 使用最新的有效token
  const testToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYwMjUwODQ2LCJpYXQiOjE3NjAxNjQ0NDZ9.4N-Gj2OCUQQRb_sAh1lxxdGyROfn591sFCQ_kNRSOtc';

  // 统一设置到auth_token
  setAuthToken(testToken);

  console.log('测试token已设置到 auth_token');
};