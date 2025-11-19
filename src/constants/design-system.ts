/**
 * 律所OA系统统一设计系统规范
 * 基于现有design-tokens.css，提供TypeScript类型安全的设计系统
 * 确保所有组件遵循统一的设计规范
 */

// =============================================================================
// 1. 设计令牌接口定义
// =============================================================================

export interface DesignTokens {
  // 颜色系统
  colors: {
    primary: string
    primaryHover: string
    primaryLight: string
    accent: string
    accentHover: string
    success: string
    successHover: string
    successLight: string
    warning: string
    warningHover: string
    warningLight: string
    error: string
    errorHover: string
    errorLight: string
    info: string
    infoHover: string
    // 功能色
    bgPage: string
    bgContainer: string
    bgCard: string
    bgHover: string
    bgActive: string
    // 文字色
    textPrimary: string
    textSecondary: string
    textTertiary: string
    textDisabled: string
    textLink: string
    textInverse: string
    // 边框色
    borderBase: string
    borderSplit: string
    borderHover: string
    borderActive: string
    // 状态色
    priorityHigh: string
    priorityMedium: string
    priorityLow: string
    statusActive: string
    statusInactive: string
    statusPending: string
    statusProcessing: string
  }

  // 间距系统
  spacing: {
    xs: string // 4px
    sm: string // 8px
    md: string // 12px
    lg: string // 16px
    xl: string // 20px
    xxl: string // 24px
    xxxl: string // 32px
    xxxxl: string // 40px
    xxxxxl: string // 48px
    xxxxxxl: string // 64px
  }

  // 字体系统
  typography: {
    xs: { fontSize: string; lineHeight: string }
    sm: { fontSize: string; lineHeight: string }
    base: { fontSize: string; lineHeight: string }
    lg: { fontSize: string; lineHeight: string }
    xl: { fontSize: string; lineHeight: string }
    xxl: { fontSize: string; lineHeight: string }
    xxxl: { fontSize: string; lineHeight: string }
    xxxx: { fontSize: string; lineHeight: string }
    xxxxx: { fontSize: string; lineHeight: string }
  }

  // 圆角系统
  radius: {
    none: string
    sm: string
    md: string
    lg: string
    xl: string
    xxl: string
    full: string
  }

  // 阴影系统
  shadows: {
    sm: string
    md: string
    lg: string
    xl: string
    xxl: string
    inner: string
    colored: string
  }

  // 动画系统
  animation: {
    durationFast: string
    durationNormal: string
    durationSlow: string
    easeOut: string
    easeIn: string
    easeInOut: string
  }

  // 尺寸系统
  sizing: {
    xs: string
    sm: string
    md: string
    lg: string
    xl: string
  }

  // 图标尺寸
  icons: {
    xs: string
    sm: string
    md: string
    lg: string
    xl: string
  }
}

// =============================================================================
// 2. 统一设计令牌配置
// =============================================================================

export const DESIGN_TOKENS: DesignTokens = {
  colors: {
    primary: '#1677FF',
    primaryHover: '#0958D9',
    primaryLight: '#E6F7FF',
    accent: '#FA8C16',
    accentHover: '#D46B08',
    success: '#52C41A',
    successHover: '#389E0D',
    successLight: '#F6FFED',
    warning: '#FAAD14',
    warningHover: '#D48806',
    warningLight: '#FFF7E6',
    error: '#F5222D',
    errorHover: '#CF1322',
    errorLight: '#FFF1F0',
    info: '#40A9FF',
    infoHover: '#1677FF',
    // 功能色
    bgPage: '#F5F5F5',
    bgContainer: '#FAFAFA',
    bgCard: '#FFFFFF',
    bgHover: '#F0F0F0',
    bgActive: '#D9D9D9',
    // 文字色
    textPrimary: '#262626',
    textSecondary: '#8C8C8C',
    textTertiary: '#BFBFBF',
    textDisabled: '#BFBFBF',
    textLink: '#1677FF',
    textInverse: '#FFFFFF',
    // 边框色
    borderBase: '#F0F0F0',
    borderSplit: '#D9D9D9',
    borderHover: '#69C0FF',
    borderActive: '#1677FF',
    // 状态色
    priorityHigh: '#F5222D',
    priorityMedium: '#FAAD14',
    priorityLow: '#52C41A',
    statusActive: '#52C41A',
    statusInactive: '#BFBFBF',
    statusPending: '#FAAD14',
    statusProcessing: '#1677FF',
  },

  spacing: {
    xs: '4px',
    sm: '8px',
    md: '12px',
    lg: '16px',
    xl: '20px',
    xxl: '24px',
    xxxl: '32px',
    xxxxl: '40px',
    xxxxxl: '48px',
    xxxxxxl: '64px',
  },

  typography: {
    xs: { fontSize: '11px', lineHeight: '16px' },
    sm: { fontSize: '12px', lineHeight: '18px' },
    base: { fontSize: '14px', lineHeight: '20px' },
    lg: { fontSize: '16px', lineHeight: '24px' },
    xl: { fontSize: '18px', lineHeight: '28px' },
    xxl: { fontSize: '20px', lineHeight: '30px' },
    xxxl: { fontSize: '24px', lineHeight: '36px' },
    xxxx: { fontSize: '32px', lineHeight: '48px' },
    xxxxx: { fontSize: '48px', lineHeight: '72px' },
  },

  radius: {
    none: '0px',
    sm: '4px',
    md: '6px',
    lg: '8px',
    xl: '12px',
    xxl: '16px',
    full: '9999px',
  },

  shadows: {
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px 0 rgba(0, 0, 0, 0.01)',
    md: '0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.10), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
    xl: '0 20px 25px -5px rgba(0, 0, 0, 0.10), 0 8px 10px -6px rgba(0, 0, 0, 0.05)',
    xxl: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
    inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
    colored: '0 10px 15px -3px rgba(22, 119, 255, 0.10), 0 4px 6px -2px rgba(22, 119, 255, 0.05)',
  },

  animation: {
    durationFast: '150ms',
    durationNormal: '300ms',
    durationSlow: '500ms',
    easeOut: 'cubic-bezier(0.4, 0, 0.2, 1)',
    easeIn: 'cubic-bezier(0.4, 0, 1, 1)',
    easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)',
  },

  sizing: {
    xs: '24px',
    sm: '32px',
    md: '40px',
    lg: '48px',
    xl: '56px',
  },

  icons: {
    xs: '14px',
    sm: '16px',
    md: '20px',
    lg: '24px',
    xl: '32px',
  },
}

// =============================================================================
// 3. 组件样式生成器
// =============================================================================

export const createComponentStyles = {
  // 卡片样式
  card: (variant: 'default' | 'hoverable' | 'bordered' = 'default') => ({
    background: DESIGN_TOKENS.colors.bgCard,
    border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
    borderRadius: DESIGN_TOKENS.radius.lg,
    boxShadow: variant === 'hoverable' ? DESIGN_TOKENS.shadows.md : DESIGN_TOKENS.shadows.sm,
    transition: `all ${DESIGN_TOKENS.animation.durationNormal} ${DESIGN_TOKENS.animation.easeOut}`,
    ...(variant === 'hoverable' && {
      '&:hover': {
        boxShadow: DESIGN_TOKENS.shadows.lg,
        transform: 'translateY(-2px)',
      },
    }),
  }),

  // 统计卡片样式
  statisticCard: (color: string = DESIGN_TOKENS.colors.primary) => ({
    background: DESIGN_TOKENS.colors.bgCard,
    border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
    borderLeft: `4px solid ${color}`,
    borderRadius: DESIGN_TOKENS.radius.lg,
    boxShadow: DESIGN_TOKENS.shadows.sm,
    transition: `all ${DESIGN_TOKENS.animation.durationNormal} ${DESIGN_TOKENS.animation.easeOut}`,
    '&:hover': {
      boxShadow: DESIGN_TOKENS.shadows.md,
      transform: 'translateY(-1px)',
    },
  }),

  // 表格样式
  table: {
    background: DESIGN_TOKENS.colors.bgCard,
    borderRadius: DESIGN_TOKENS.radius.lg,
    border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
    overflow: 'hidden',
  },

  // 按钮样式
  button: {
    primary: {
      background: `linear-gradient(135deg, ${DESIGN_TOKENS.colors.primary} 0%, ${DESIGN_TOKENS.colors.primaryHover} 100%)`,
      border: 'none',
      borderRadius: DESIGN_TOKENS.radius.md,
      boxShadow: DESIGN_TOKENS.shadows.sm,
      transition: `all ${DESIGN_TOKENS.animation.durationNormal} ${DESIGN_TOKENS.animation.easeOut}`,
      '&:hover': {
        boxShadow: DESIGN_TOKENS.shadows.md,
        transform: 'translateY(-1px)',
      },
    },
    default: {
      background: DESIGN_TOKENS.colors.bgCard,
      border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
      borderRadius: DESIGN_TOKENS.radius.md,
      transition: `all ${DESIGN_TOKENS.animation.durationNormal} ${DESIGN_TOKENS.animation.easeOut}`,
      '&:hover': {
        borderColor: DESIGN_TOKENS.colors.borderHover,
        boxShadow: DESIGN_TOKENS.shadows.sm,
      },
    },
  },

  // 输入框样式
  input: {
    border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
    borderRadius: DESIGN_TOKENS.radius.sm,
    transition: `all ${DESIGN_TOKENS.animation.durationFast} ${DESIGN_TOKENS.animation.easeOut}`,
    '&:focus': {
      borderColor: DESIGN_TOKENS.colors.borderActive,
      boxShadow: `0 0 0 2px ${DESIGN_TOKENS.colors.primaryLight}`,
    },
  },

  // 标签样式
  tag: (color: string) => ({
    background: color,
    color: DESIGN_TOKENS.colors.textInverse,
    borderRadius: DESIGN_TOKENS.radius.sm,
    fontSize: DESIGN_TOKENS.typography.xs.fontSize,
    fontWeight: '500',
    padding: `${DESIGN_TOKENS.spacing.xs} ${DESIGN_TOKENS.spacing.sm}`,
  }),

  // 页面容器样式
  pageContainer: {
    background: DESIGN_TOKENS.colors.bgPage,
    minHeight: '100vh',
    padding: DESIGN_TOKENS.spacing.lg,
  },

  // 内容区域样式
  contentArea: {
    background: DESIGN_TOKENS.colors.bgContainer,
    borderRadius: DESIGN_TOKENS.radius.xl,
    padding: DESIGN_TOKENS.spacing.xxl,
    boxShadow: DESIGN_TOKENS.shadows.md,
  },

  // 状态指示器样式
  statusIndicator: (status: string) => {
    const statusColors = {
      active: DESIGN_TOKENS.colors.statusActive,
      inactive: DESIGN_TOKENS.colors.statusInactive,
      pending: DESIGN_TOKENS.colors.statusPending,
      processing: DESIGN_TOKENS.colors.statusProcessing,
      error: DESIGN_TOKENS.colors.error,
      success: DESIGN_TOKENS.colors.success,
      warning: DESIGN_TOKENS.colors.warning,
    }

    return {
      display: 'inline-flex',
      alignItems: 'center',
      gap: DESIGN_TOKENS.spacing.xs,
      padding: `${DESIGN_TOKENS.spacing.xs} ${DESIGN_TOKENS.spacing.sm}`,
      borderRadius: DESIGN_TOKENS.radius.sm,
      fontSize: DESIGN_TOKENS.typography.xs.fontSize,
      fontWeight: '500',
      background:
        statusColors[status as keyof typeof statusColors] || DESIGN_TOKENS.colors.textTertiary,
      color: DESIGN_TOKENS.colors.textInverse,
    }
  },
}

// =============================================================================
// 4. 业务状态映射
// =============================================================================

export const BUSINESS_STATUS = {
  // 案件状态
  CASE_STATUS: {
    PENDING: { text: '待处理', color: DESIGN_TOKENS.colors.warning },
    ACTIVE: { text: '进行中', color: DESIGN_TOKENS.colors.primary },
    COMPLETED: { text: '已完成', color: DESIGN_TOKENS.colors.success },
    SUSPENDED: { text: '已暂停', color: DESIGN_TOKENS.colors.error },
    CANCELLED: { text: '已取消', color: DESIGN_TOKENS.colors.textTertiary },
  },

  // 案件类型
  CASE_TYPE: {
    CIVIL: { text: '民事案件', color: DESIGN_TOKENS.colors.primary },
    COMMERCIAL: { text: '商事案件', color: DESIGN_TOKENS.colors.accent },
    CRIMINAL: { text: '刑事案件', color: DESIGN_TOKENS.colors.error },
    ADMINISTRATIVE: { text: '行政案件', color: DESIGN_TOKENS.colors.info },
  },

  // 用户类型
  USER_TYPE: {
    SYSTEM: { text: '系统用户', color: DESIGN_TOKENS.colors.error },
    LAWYER: { text: '律师', color: DESIGN_TOKENS.colors.primary },
    ASSISTANT: { text: '助理', color: DESIGN_TOKENS.colors.success },
    ADMIN: { text: '行政', color: DESIGN_TOKENS.colors.accent },
  },

  // 优先级
  PRIORITY: {
    HIGH: { text: '高', color: DESIGN_TOKENS.colors.priorityHigh },
    MEDIUM: { text: '中', color: DESIGN_TOKENS.colors.priorityMedium },
    LOW: { text: '低', color: DESIGN_TOKENS.colors.priorityLow },
  },

  // 审批状态
  APPROVAL_STATUS: {
    PENDING: { text: '待审批', color: DESIGN_TOKENS.colors.warning },
    APPROVED: { text: '已通过', color: DESIGN_TOKENS.colors.success },
    REJECTED: { text: '已拒绝', color: DESIGN_TOKENS.colors.error },
    CANCELLED: { text: '已撤销', color: DESIGN_TOKENS.colors.textTertiary },
  },
} as const

// =============================================================================
// 5. 响应式断点
// =============================================================================

export const BREAKPOINTS = {
  xs: '480px',
  sm: '576px',
  md: '768px',
  lg: '992px',
  xl: '1200px',
  xxl: '1600px',
} as const

// =============================================================================
// 6. 工具函数
// =============================================================================

export const designUtils = {
  // 获取状态颜色
  getStatusColor: (status: string, type: 'case' | 'user' | 'priority' | 'approval' = 'case') => {
    const statusMaps = {
      case: BUSINESS_STATUS.CASE_STATUS,
      user: BUSINESS_STATUS.USER_TYPE,
      priority: BUSINESS_STATUS.PRIORITY,
      approval: BUSINESS_STATUS.APPROVAL_STATUS,
    }

    const statusMap = statusMaps[type]
    const statusKey = Object.keys(statusMap).find(
      (key) =>
        key === status.toUpperCase() || statusMap[key as keyof typeof statusMap].text === status,
    )

    return statusKey
      ? statusMap[statusKey as keyof typeof statusMap].color
      : DESIGN_TOKENS.colors.textTertiary
  },

  // 格式化尺寸
  formatSize: (size: number | string, unit: string = 'px') => {
    return typeof size === 'number' ? `${size}${unit}` : size
  },

  // 创建渐变
  createGradient: (from: string, to: string, direction: string = '135deg') => {
    return `linear-gradient(${direction}, ${from} 0%, ${to} 100%)`
  },

  // 响应式样式生成
  responsive: (styles: Record<string, any>) => {
    return Object.entries(styles).reduce((acc, [breakpoint, style]) => {
      if (breakpoint === 'base') {
        return { ...acc, ...style }
      }
      return {
        ...acc,
        [`@media (min-width: ${BREAKPOINTS[breakpoint as keyof typeof BREAKPOINTS]})`]: style,
      }
    }, {})
  },

  // 深度合并对象
  deepMerge: (target: any, source: any) => {
    const output = { ...target }
    if (isObject(target) && isObject(source)) {
      Object.keys(source).forEach((key) => {
        if (isObject(source[key])) {
          if (!(key in target)) {
            Object.assign(output, { [key]: source[key] })
          } else {
            output[key] = designUtils.deepMerge(target[key], source[key])
          }
        } else {
          Object.assign(output, { [key]: source[key] })
        }
      })
    }
    return output
  },
}

// 辅助函数
function isObject(item: any) {
  return item && typeof item === 'object' && !Array.isArray(item)
}

// =============================================================================
// 7. 导出默认值
// =============================================================================

export default {
  DESIGN_TOKENS,
  createComponentStyles,
  BUSINESS_STATUS,
  BREAKPOINTS,
  designUtils,
}
