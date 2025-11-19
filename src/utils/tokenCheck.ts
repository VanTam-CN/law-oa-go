/**
 * Token统一性检查工具
 * 验证所有模块使用相同的token存储key
 */

import { getToken as storageGetToken } from './storage'
import { getToken as authGetToken } from '../services/auth'
import { getAuthToken as utilGetToken } from './auth'

export interface TokenCheckResult {
  storage: string | null
  auth: string | null
  util: string | null
  consistent: boolean
  details: string[]
}

export function checkTokenConsistency(): TokenCheckResult {
  const storageToken = storageGetToken()
  const authToken = authGetToken()
  const utilToken = utilGetToken()

  const details: string[] = []
  details.push(`storage.ts getToken(): ${storageToken ? '✅ 已设置' : '❌ 未设置'}`)
  details.push(`services/auth.ts getToken(): ${authToken ? '✅ 已设置' : '❌ 未设置'}`)
  details.push(`utils/auth.ts getAuthToken(): ${utilToken ? '✅ 已设置' : '❌ 未设置'}`)

  const consistent = storageToken === authToken && authToken === utilToken
  details.push(`一致性检查: ${consistent ? '✅ 所有模块一致' : '❌ 存在不一致'}`)

  if (storageToken) {
    details.push(`当前token: ${storageToken.substring(0, 20)}...`)
  }

  return {
    storage: storageToken,
    auth: authToken,
    util: utilToken,
    consistent,
    details,
  }
}

export function checkUserInfoConsistency(): { userInfo: any; hasUserData: boolean } {
  const userInfo = localStorage.getItem('user_info')
  const hasUserData = !!userInfo

  return {
    userInfo: userInfo ? JSON.parse(userInfo) : null,
    hasUserData,
  }
}

// 开发环境下自动检查并输出结果
if (import.meta.env.DEV) {
  console.group('🔍 Token统一性检查')
  const tokenCheck = checkTokenConsistency()
  console.log('Token检查结果:', tokenCheck)

  const userInfoCheck = checkUserInfoConsistency()
  console.log('用户信息检查:', userInfoCheck.hasUserData ? '✅ 已设置' : '❌ 未设置')

  if (!tokenCheck.consistent) {
    console.warn('⚠️ Token存储不一致，可能影响认证功能')
    tokenCheck.details.forEach((detail) => console.warn(detail))
  } else {
    console.log('✅ Token存储一致性检查通过')
  }
  console.groupEnd()

  // 暴露到全局，便于调试
  ;(window as any).tokenCheck = checkTokenConsistency
  ;(window as any).userInfoCheck = checkUserInfoConsistency
}
