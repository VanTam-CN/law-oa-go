/**
 * ConflictCheckInline组件的类型定义
 * 支持内联冲突检测和快速结果显示
 */

import { ReactNode } from 'react'
import type { AlertProps } from 'antd/es/alert'

// 冲突检测状态
export type ConflictCheckStatus = 'idle' | 'checking' | 'success' | 'warning' | 'error'

// 冲突类型
export type ConflictType = 'direct' | 'indirect' | 'potential' | 'none'

// 冲突严重程度
export type ConflictSeverity = 'low' | 'medium' | 'high' | 'critical'

// 冲突项目信息
export interface ConflictCase {
  /** 案件ID */
  id: string
  /** 案件编号 */
  caseNumber: string
  /** 案件标题 */
  title: string
  /** 客户名称 */
  clientName: string
  /** 负责律师 */
  lawyerName: string
  /** 冲突类型 */
  conflictType: ConflictType
  /** 严重程度 */
  severity: ConflictSeverity
  /** 冲突描述 */
  description: string
  /** 建议措施 */
  recommendation?: string
  /** 创建时间 */
  createdAt: string
}

// 冲突检测结果
export interface ConflictCheckResult {
  /** 检测状态 */
  status: ConflictCheckStatus
  /** 是否存在冲突 */
  hasConflict: boolean
  /** 冲突项目列表 */
  conflicts: ConflictCase[]
  /** 检测时间 */
  checkedAt?: string
  /** 错误信息 */
  error?: string
  /** 统计信息 */
  stats: {
    total: number
    direct: number
    indirect: number
    potential: number
  }
}

// 内联显示配置
export interface InlineDisplayConfig {
  /** 是否显示详细信息 */
  showDetails: boolean
  /** 是否显示统计信息 */
  showStats: boolean
  /** 是否显示操作按钮 */
  showActions: boolean
  /** 最大显示数量 */
  maxDisplayCount: number
  /** 是否启用1080p紧凑模式 */
  isCompact: boolean
  /** 显示模式 */
  displayMode: 'card' | 'alert' | 'inline'
}

// 快速操作配置
export interface QuickActionConfig {
  /** 是否允许重新检测 */
  allowRecheck: boolean
  /** 是否允许查看详情 */
  allowViewDetails: boolean
  /** 是否允许标记为已解决 */
  allowMarkResolved: boolean
  /** 自定义操作按钮 */
  customActions?: Array<{
    key: string
    label: string
    icon?: ReactNode
    onClick: () => void
    danger?: boolean
  }>
}

// 冲突检测参数
export interface ConflictCheckParams {
  /** 客户名称 */
  clientName: string
  /** 对方当事人 */
  opposingParty?: string
  /** 案件类型 */
  caseType?: string
  /** 负责律师ID */
  lawyerId?: string
  /** 助理律师ID */
  assistingLawyerId?: string
  /** 检测范围 */
  scope?: 'all' | 'lawyer' | 'client' | 'case'
}

// 组件属性
export interface ConflictCheckInlineProps {
  /** 检测参数 */
  checkParams: ConflictCheckParams
  /** 当前检测结果 */
  result?: ConflictCheckResult
  /** 检测状态变化回调 */
  onStatusChange?: (status: ConflictCheckStatus) => void
  /** 检测完成回调 */
  onCheckComplete?: (result: ConflictCheckResult) => void
  /** 查看详情回调 */
  onViewDetails?: (conflict: ConflictCase) => void
  /** 重新检测回调 */
  onRecheck?: () => void
  /** 标记为已解决回调 */
  onMarkResolved?: (conflictIds: string[]) => void
  /** 内联显示配置 */
  displayConfig?: Partial<InlineDisplayConfig>
  /** 快速操作配置 */
  actionConfig?: Partial<QuickActionConfig>
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
  /** 自动检测（参数变化时） */
  autoCheck?: boolean
  /** 检测防抖延迟(ms) */
  debounceDelay?: number
}

// 统计信息卡片属性
export interface ConflictStatsCardProps {
  /** 统计数据 */
  stats: ConflictCheckResult['stats']
  /** 总数 */
  total: number
  /** 是否为紧凑模式 */
  isCompact: boolean
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
}

// 冲突列表项属性
export interface ConflictListItemProps {
  /** 冲突项目 */
  conflict: ConflictCase
  /** 索引 */
  index: number
  /** 是否显示详细信息 */
  showDetails: boolean
  /** 是否为紧凑模式 */
  isCompact: boolean
  /** 查看详情回调 */
  onViewDetails?: (conflict: ConflictCase) => void
  /** 标记为已解决回调 */
  onMarkResolved?: (conflictId: string) => void
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
}

// 快速操作按钮属性
export interface QuickActionsProps {
  /** 检测状态 */
  status: ConflictCheckStatus
  /** 是否有冲突 */
  hasConflict: boolean
  /** 操作配置 */
  config: QuickActionConfig
  /** 是否为紧凑模式 */
  isCompact: boolean
  /** 重新检测回调 */
  onRecheck?: () => void
  /** 查看所有冲突回调 */
  onViewAllConflicts?: () => void
  /** 标记为已解决回调 */
  onMarkAllResolved?: () => void
  /** 自定义操作回调 */
  onCustomAction?: (actionKey: string) => void
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
}

// 检测状态指示器属性
export interface CheckStatusIndicatorProps {
  /** 检测状态 */
  status: ConflictCheckStatus
  /** 进度百分比 */
  progress?: number
  /** 是否显示文本 */
  showText?: boolean
  /** 是否为紧凑模式 */
  isCompact: boolean
  /** 自定义状态文本 */
  statusText?: Record<ConflictCheckStatus, string>
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
}

// 冲突严重程度配置
export interface SeverityConfig {
  /** 颜色映射 */
  colors: Record<ConflictSeverity, string>
  /** 图标映射 */
  icons: Record<ConflictSeverity, ReactNode>
  /** 文本映射 */
  labels: Record<ConflictSeverity, string>
  /** 描述映射 */
  descriptions: Record<ConflictSeverity, string>
}

// 导出默认配置
export const DEFAULT_INLINE_DISPLAY_CONFIG: InlineDisplayConfig = {
  showDetails: true,
  showStats: true,
  showActions: true,
  maxDisplayCount: 3,
  isCompact: false,
  displayMode: 'card',
}

export const DEFAULT_QUICK_ACTION_CONFIG: QuickActionConfig = {
  allowRecheck: true,
  allowViewDetails: true,
  allowMarkResolved: true,
  customActions: [],
}

export const DEFAULT_SEVERITY_CONFIG: SeverityConfig = {
  colors: {
    low: '#52c41a',
    medium: '#faad14',
    high: '#ff7a45',
    critical: '#ff4d4f',
  },
  icons: {},
  labels: {
    low: '低风险',
    medium: '中风险',
    high: '高风险',
    critical: '严重冲突',
  },
  descriptions: {
    low: '存在轻微冲突，可以通过常规流程处理',
    medium: '存在中等风险冲突，需要谨慎处理',
    high: '存在高风险冲突，需要详细评估',
    critical: '存在严重冲突，需要立即处理',
  },
}
