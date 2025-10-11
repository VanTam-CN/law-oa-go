/**
 * 本地存储工具函数
 */
import type { Role, Permission } from '@/services/role';

// Token相关操作
const TOKEN_KEY = 'law_oa_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

// 用户信息相关操作
const USER_INFO_KEY = 'law_oa_user_info';

export function getUserInfo(): any | null {
  const userInfo = localStorage.getItem(USER_INFO_KEY);
  return userInfo ? JSON.parse(userInfo) : null;
}

export function setUserInfo(userInfo: any): void {
  localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo));
}

export function removeUserInfo(): void {
  localStorage.removeItem(USER_INFO_KEY);
}

// 角色信息相关操作
const ROLES_KEY = 'law_oa_roles';

export function getRoles(): Role[] | null {
  const roles = localStorage.getItem(ROLES_KEY);
  return roles ? JSON.parse(roles) : null;
}

export function setRoles(roles: Role[]): void {
  localStorage.setItem(ROLES_KEY, JSON.stringify(roles));
}

export function removeRoles(): void {
  localStorage.removeItem(ROLES_KEY);
}

// 权限信息相关操作
const PERMISSIONS_KEY = 'law_oa_permissions';

export function getPermissions(): Permission[] | null {
  const permissions = localStorage.getItem(PERMISSIONS_KEY);
  return permissions ? JSON.parse(permissions) : null;
}

export function setPermissions(permissions: Permission[]): void {
  localStorage.setItem(PERMISSIONS_KEY, JSON.stringify(permissions));
}

export function removePermissions(): void {
  localStorage.removeItem(PERMISSIONS_KEY);
}

// 清除所有存储
export function clearStorage(): void {
  removeToken();
  removeUserInfo();
  removeRoles();
  removePermissions();
}