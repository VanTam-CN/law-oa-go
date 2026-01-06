/**
 * ProgressIndicator组件的类型定义
 * 支持表单完成进度显示和步骤导航
 */

import { ReactNode } from 'react'
import type { StepsProps } from 'antd/es/steps'

// 步骤状态
export type StepStatus = 'wait' | 'process' | 'finish' | 'error'

// 步骤配置
export interface StepConfig {
  /** 步骤唯一标识 */
  key: string
  /** 步骤标题 */
  title: string
  /** 步骤描述 */
  description?: string
  /** 步骤图标 */
  icon?: ReactNode
  /** 步骤状态 */
  status?: StepStatus
  /** 是否禁用点击导航 */
  disabled?: boolean
  /** 自定义点击处理函数 */
  onClick?: () => void
  /** 步骤是否完成 */
  completed?: boolean
  /** 步骤验证规则 */
  validation?: {
    required: boolean
    fields: string[]
    validator?: () => boolean
  }
  /** 步骤权重（用于进度计算） */
  weight?: number
}

// 进度统计信息
export interface ProgressStats {
  /** 总步骤数 */
  totalSteps: number
  /** 已完成步骤数 */
  completedSteps: number
  /** 当前步骤索引 */
  currentStep: number
  /** 完成百分比 */
  percentage: number
  /** 错误步骤数 */
  errorSteps: number
  /** 警告步骤数 */
  warningSteps: number
}

// 进度指示器属性
export interface ProgressIndicatorProps {
  /** 步骤配置数组 */
  steps: StepConfig[]
  /** 当前步骤key */
  currentStepKey?: string
  /** 步骤变化回调 */
  onStepChange?: (stepKey: string, stepIndex: number) => void
  /** 是否显示进度条 */
  showProgressBar?: boolean
  /** 是否显示统计信息 */
  showStats?: boolean
  /** 是否启用步骤点击导航 */
  enableStepNavigation?: boolean
  /** 进度条样式 */
  progressType?: 'line' | 'circle' | 'dashboard'
  /** 步骤方向 */
  direction?: 'horizontal' | 'vertical'
  /** 步骤大小 */
  size?: 'default' | 'small'
  /** 是否为1080p紧凑模式 */
  isCompact?: boolean
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
  /** Steps组件的其他属性 */
  stepsProps?: Omit<StepsProps, 'items' | 'current' | 'onChange'>
}

// 进度条组件属性
export interface ProgressBarProps {
  /** 进度统计信息 */
  stats: ProgressStats
  /** 进度条类型 */
  type?: 'line' | 'circle' | 'dashboard'
  /** 是否显示百分比文本 */
  showText?: boolean
  /** 是否为1080p紧凑模式 */
  isCompact?: boolean
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
}

// 统计信息组件属性
export interface ProgressStatsProps {
  /** 进度统计信息 */
  stats: ProgressStats
  /** 是否显示详细统计 */
  showDetails?: boolean
  /** 是否为1080p紧凑模式 */
  isCompact?: boolean
  /** 自定义格式化函数 */
  formatter?: (stats: ProgressStats) => {
    percentage: string
    completed: string
    total: string
    errors: string
  }
  /** 自定义样式 */
  className?: string
  style?: React.CSSProperties
}

// 步骤导航事件
export interface StepNavigationEvent {
  /** 步骤key */
  stepKey: string
  /** 步骤索引 */
  stepIndex: number
  /** 步骤配置 */
  step: StepConfig
  /** 事件类型 */
  type: 'click' | 'validate' | 'error'
  /** 额外数据 */
  data?: any
}

// 验证结果
export interface ValidationResult {
  /** 是否有效 */
  valid: boolean
  /** 错误字段 */
  errors: string[]
  /** 警告字段 */
  warnings: string[]
  /** 错误步骤 */
  errorSteps: string[]
  /** 警告步骤 */
  warningSteps: string[]
}

// 进度计算配置
export interface ProgressCalculationConfig {
  /** 是否包含禁用步骤 */
  includeDisabled?: boolean
  /** 是否加权计算 */
  useWeighting?: boolean
  /** 自定义权重映射 */
  weights?: Record<string, number>
  /** 最小进度值（防止0%显示） */
  minProgress?: number
}

// 导航配置
export interface NavigationConfig {
  /** 是否允许跳转到未完成步骤 */
  allowSkipIncomplete?: boolean
  /** 是否在点击时验证当前步骤 */
  validateOnClick?: boolean
  /** 自定义导航确认对话框 */
  confirmDialog?: {
    title?: string
    content?: string
    okText?: string
    cancelText?: string
  }
}
