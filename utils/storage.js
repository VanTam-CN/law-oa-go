/**
 * 本地存储工具函数
 */

// Token相关操作
const TOKEN_KEY = 'law_oa_token';

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token) {
  return localStorage.setItem(TOKEN_KEY, token);
}

export function removeToken() {
  return localStorage.removeItem(TOKEN_KEY);
}

// 用户信息相关操作
const USER_INFO_KEY = 'law_oa_user_info';

export function getUserInfo() {
  const userInfo = localStorage.getItem(USER_INFO_KEY);
  return userInfo ? JSON.parse(userInfo) : null;
}

export function setUserInfo(userInfo) {
  return localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo));
}

export function removeUserInfo() {
  return localStorage.removeItem(USER_INFO_KEY);
}

// 清除所有存储
export function clearStorage() {
  removeToken();
  removeUserInfo();
}