/**
 * Token统一性检查工具
 * 验证所有模块使用相同的token存储key
 */

import {
  checkTokenConsistency,
  checkUserInfoConsistency,
  type TokenCheckResult,
} from './tokenCheckCore'

export type { TokenCheckResult }
export { checkTokenConsistency, checkUserInfoConsistency }

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
