import { ReactNode, CSSProperties } from 'react'

export interface ResponsiveFormLayoutProps {
  /** 子组件内容 */
  children: ReactNode

  /** 列数配置，支持1-4列 */
  columns?: 1 | 2 | 3 | 4

  /** 间距配置 */
  spacing?: 'small' | 'medium' | 'large'

  /** 响应式断点 */
  breakpoint?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'xxl'

  /** 自定义类名 */
  className?: string

  /** 自定义样式 */
  style?: CSSProperties

  /** 是否启用自适应间距 */
  adaptiveSpacing?: boolean

  /** 最大宽度限制 */
  maxWidth?: number

  /** 是否启用垂直滚动 */
  enableVerticalScroll?: boolean

  /** 容器内边距 */
  padding?: {
    top?: number
    right?: number
    bottom?: number
    left?: number
    horizontal?: number
    vertical?: number
  }

  /** 是否显示响应式提示 */
  showResponsiveHint?: boolean
}

export interface BreakpointConfig {
  xs?: {
    span: number
    offset?: number
    order?: number
    push?: number
    pull?: number
  }

  sm?: {
    span: number
    offset?: number
    order?: number
    push?: number
    pull?: number
  }

  md?: {
    span: number
    offset?: number
    order?: number
    push?: number
    pull?: number
  }

  lg?: {
    span: number
    offset?: number
    order?: number
    push?: number
    pull?: number
  }

  xl?: {
    span: number
    offset?: number
    order?: number
    push?: number
    pull?: number
  }

  xxl?: {
    span: number
    offset?: number
    order?: number
    push?: number
    pull?: number
  }
}

export interface SpacingConfig {
  horizontal: number
  vertical: number
}

export interface ResponsiveConfig {
  columns: number
  spacing: number
  isCompact: boolean
  maxWidth: number
  padding: number
  margin: number
}

export interface BreakpointInfo {
  name: string
  minWidth: number
  maxWidth?: number
  columns: number
  isCompact: boolean
}

// 1080p显示器标准配置
export const DISPLAY_1080P_CONFIG = {
  width: 1920,
  height: 1080,
  dpi: 96,
  aspectRatio: 16 / 9,
  recommendedColumns: {
    form: 3,
    table: 4,
    card: 4,
  },
  optimalSpacing: {
    horizontal: 12,
    vertical: 12,
  },
  compactSpacing: {
    horizontal: 8,
    vertical: 8,
  },
}

// 响应式断点配置
export const RESPONSIVE_BREAKPOINTS = {
  xs: 480, // 超小屏幕
  sm: 576, // 小屏幕
  md: 768, // 中等屏幕
  lg: 992, // 大屏幕 (包含1080p)
  xl: 1200, // 超大屏幕
  xxl: 1600, // 超超大屏幕
} as const

export type BreakpointType = keyof typeof RESPONSIVE_BREAKPOINTS
