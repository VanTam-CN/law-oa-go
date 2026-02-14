/**
 * 隔离墙 (Ethical Wall) 相关类型定义
 * 与后端 API 完全对齐
 */

// 隔离墙状态
export enum EthicalWallStatus {
  ENABLED = 'ENABLED',
  DISABLED = 'DISABLED',
}

// 隔离墙信息
export interface EthicalWall {
  caseId: string
  enabled: boolean
  createdAt?: string
  createdBy?: string
  description?: string
}

// 白名单用户信息
export interface WhitelistUser {
  userId: string
  userName: string
  userEmail?: string
  department?: string
  position?: string
  addedAt: string
  addedBy: string
  reason?: string
}

// 白名单列表响应
export interface WhitelistResponse {
  caseId: string
  users: WhitelistUser[]
  total: number
}

// 添加白名单请求
export interface AddWhitelistRequest {
  userId: string
  reason?: string
}

// 用户选择项（用于搜索选择）
export interface UserOption {
  id: string
  name: string
  email: string
  department?: string
  position?: string
  avatar?: string
}

// 隔离墙统计数据
export interface EthicalWallStats {
  totalCasesWithWall: number
  totalWhitelistEntries: number
  recentActivities: EthicalWallActivity[]
}

// 隔离墙活动记录
export interface EthicalWallActivity {
  id: string
  caseId: string
  caseName: string
  action: 'enabled' | 'disabled' | 'whitelist_added' | 'whitelist_removed'
  performedBy: string
  performedAt: string
  details?: string
}

// API 错误类型
export interface EthicalWallError {
  code: string
  message: string
  field?: string
  details?: Record<string, any>
}
