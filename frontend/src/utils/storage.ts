/**
 * 本地存储工具函数
 */
import type { Role, Permission } from '@/services/role'

// Token相关操作
const TOKEN_KEY = 'auth_token'
const USER_INFO_KEY = 'law_oa_user_info'
const ROLES_KEY = 'law_oa_roles'
const PERMISSIONS_KEY = 'law_oa_permissions'
const SESSION_STORAGE_MODE_KEY = 'law_oa_session_only'
const AUTH_STORAGE_KEYS = [TOKEN_KEY, USER_INFO_KEY, ROLES_KEY, PERMISSIONS_KEY]

function authStorage(): Storage {
  return sessionStorage.getItem(SESSION_STORAGE_MODE_KEY) === '1' ? sessionStorage : localStorage
}

export function setStoragePersistence(persistent: boolean): void {
  AUTH_STORAGE_KEYS.forEach((key) => {
    localStorage.removeItem(key)
    sessionStorage.removeItem(key)
  })
  if (persistent) {
    sessionStorage.removeItem(SESSION_STORAGE_MODE_KEY)
  } else {
    sessionStorage.setItem(SESSION_STORAGE_MODE_KEY, '1')
  }
}

function parseJwtPayload(token: string): Record<string, any> | null {
  try {
    const [, payload] = token.split('.')
    if (!payload) {
      return null
    }

    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), '=')
    const decoded = atob(padded)
    return JSON.parse(decoded)
  } catch (error) {
    return null
  }
}

export function isTokenExpired(token: string | null, skewSeconds = 30): boolean {
  if (!token) {
    return true
  }

  const payload = parseJwtPayload(token)
  if (!payload || typeof payload.exp !== 'number') {
    return true
  }

  return payload.exp * 1000 <= Date.now() + skewSeconds * 1000
}

export function getToken(): string | null {
  const token = authStorage().getItem(TOKEN_KEY)
  if (!token) {
    return null
  }

  if (isTokenExpired(token)) {
    return null
  }

  return token
}

export function setToken(token: string): void {
  authStorage().setItem(TOKEN_KEY, token)
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  sessionStorage.removeItem(TOKEN_KEY)
}

// 用户信息相关操作
export function getUserInfo(): any | null {
  const userInfo = authStorage().getItem(USER_INFO_KEY)

  if (!userInfo) {
    return null
  }

  try {
    const parsedUserInfo = JSON.parse(userInfo)
    return typeof parsedUserInfo === 'object' &&
      parsedUserInfo !== null &&
      !Array.isArray(parsedUserInfo)
      ? parsedUserInfo
      : null
  } catch {
    return null
  }
}

export function setUserInfo(userInfo: any): void {
  authStorage().setItem(USER_INFO_KEY, JSON.stringify(userInfo))
}

export function removeUserInfo(): void {
  localStorage.removeItem(USER_INFO_KEY)
  sessionStorage.removeItem(USER_INFO_KEY)
}

// 角色信息相关操作
export function getRoles(): Role[] | null {
  const roles = authStorage().getItem(ROLES_KEY)
  return roles ? JSON.parse(roles) : null
}

export function setRoles(roles: Role[]): void {
  authStorage().setItem(ROLES_KEY, JSON.stringify(roles))
}

export function removeRoles(): void {
  localStorage.removeItem(ROLES_KEY)
  sessionStorage.removeItem(ROLES_KEY)
}

// 权限信息相关操作
export function getPermissions(): Permission[] | null {
  const permissions = authStorage().getItem(PERMISSIONS_KEY)
  return permissions ? JSON.parse(permissions) : null
}

export function setPermissions(permissions: Permission[]): void {
  authStorage().setItem(PERMISSIONS_KEY, JSON.stringify(permissions))
}

export function removePermissions(): void {
  localStorage.removeItem(PERMISSIONS_KEY)
  sessionStorage.removeItem(PERMISSIONS_KEY)
}

// 清除所有存储
export function clearStorage(): void {
  removeToken()
  removeUserInfo()
  removeRoles()
  removePermissions()
  sessionStorage.removeItem(SESSION_STORAGE_MODE_KEY)
}
