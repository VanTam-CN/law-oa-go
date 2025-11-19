/**
 * 统一UI组件导出
 * 基于设计系统，提供一致的UI组件
 */
export { default as StandardCard } from './StandardCard'
export { default as StandardPage } from './StandardPage'
export { default as StandardTable, createStandardColumns } from './StandardTable'

// 重新导出设计系统
export {
  DESIGN_TOKENS,
  createComponentStyles,
  BUSINESS_STATUS,
  BREAKPOINTS,
  designUtils,
} from '@/constants/design-system'
