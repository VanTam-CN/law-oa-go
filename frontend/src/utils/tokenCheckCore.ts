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
    details.push('当前token: 已设置')
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
  const userInfo = localStorage.getItem('law_oa_user_info')

  if (!userInfo) {
    return { userInfo: null, hasUserData: false }
  }

  try {
    const parsedUserInfo = JSON.parse(userInfo)

    if (
      typeof parsedUserInfo === 'object' &&
      parsedUserInfo !== null &&
      !Array.isArray(parsedUserInfo)
    ) {
      return { userInfo: parsedUserInfo, hasUserData: true }
    }
  } catch {
    // Treat malformed local data as absent without modifying the user's browser storage.
  }

  return { userInfo: null, hasUserData: false }
}
