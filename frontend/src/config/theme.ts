/**
 * Ant Design 主题配置 - 律所OA系统
 *
 * 设计理念：
 * - 专业深蓝色系，传达信任与权威
 * - 金色点缀，体现高端品质
 * - 柔和的语义色，减少视觉疲劳
 *
 * 更新日期：2025-12-29
 */

import { ThemeConfig } from 'antd'

/**
 * 律所OA系统统一主题配置
 *
 * 使用方式：
 * ```tsx
 * import { ConfigProvider } from 'antd'
 * import { antdTheme } from '@/config/theme'
 *
 * <ConfigProvider theme={antdTheme}>
 *   <App />
 * </ConfigProvider>
 * ```
 */
export const antdTheme: ThemeConfig = {
  // ========================================
  // 全局令牌配置 (Global Tokens)
  // ========================================
  token: {
    // 主色调 - 律所专业深蓝
    colorPrimary: '#1E5A8D', // 主深蓝 - 专业、信任、权威

    // 语义化颜色 - 柔和色调
    colorSuccess: '#3FAF56', // 成功绿 - 更柔和
    colorWarning: '#F5A623', // 警告橙 - 醒目但不刺眼
    colorError: '#E8484C', // 错误红 - 清晰明确
    colorInfo: '#1E5A8D', // 信息蓝 - 与主色统一

    // 中性色 - 页面背景
    colorBgBase: '#FFFFFF',
    colorBgContainer: '#FFFFFF',
    colorBgElevated: '#FFFFFF',
    colorBgLayout: '#F4F6F8', // 页面背景灰

    // 文字色系
    colorText: '#374151', // 主要文字
    colorTextSecondary: '#6B7280', // 次要文字
    colorTextTertiary: '#9CA3AF', // 辅助文字
    colorTextQuaternary: '#D1D5DB', // 禁用文字

    // 边框色系
    colorBorder: '#E8EBF0', // 基础边框
    colorBorderSecondary: '#F4F6F8', // 次要边框

    // 圆角系统
    borderRadius: 6, // 默认圆角 - 按钮
    borderRadiusLG: 8, // 大圆角 - 卡片
    borderRadiusSM: 4, // 小圆角 - 输入框
    borderRadiusXS: 2,

    // 间距系统 (8px 基准)
    marginXS: 4, // 0.5x
    marginSM: 8, // 1x
    margin: 12, // 1.5x
    marginMD: 16, // 2x
    marginLG: 24, // 3x
    marginXL: 32, // 4x
    marginXXL: 48, // 6x

    // 字体系统
    fontSize: 14, // 基础字号
    fontSizeSM: 12, // 小字号
    fontSizeLG: 16, // 大字号
    fontSizeXL: 18, // 特大字号
    fontSizeXXL: 20, // 超大字号

    // 字重系统
    fontWeightStrong: 600, // 半粗体
    fontWeightNormal: 400, // 常规
    fontWeightMedium: 500, // 中等

    // 行高系统
    lineHeight: 1.5, // 基础行高
    lineHeightLG: 1.75, // 宽松行高
    lineHeightSM: 1.25, // 紧凑行高

    // 阴影系统
    boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px 0 rgba(0, 0, 0, 0.06)', // 小阴影
    boxShadowSecondary: '0 4px 6px -1px rgba(0, 0, 0, 0.10)', // 标准阴影

    // 动画系统
    motionDurationFast: '0.1s', // 快速动画
    motionDurationMid: '0.2s', // 中速动画
    motionDurationSlow: '0.3s', // 慢速动画

    motionEaseInOut: 'cubic-bezier(0.4, 0, 0.2, 1)', // 缓入缓出
    motionEaseOut: 'cubic-bezier(0, 0, 0.2, 1)', // 缓出
    motionEaseIn: 'cubic-bezier(0.4, 0, 1, 1)', // 缓入
  } as NonNullable<ThemeConfig['token']> & Record<string, unknown>,

  // ========================================
  // 组件级配置 (Component Tokens)
  // ========================================
  components: {
    // 按钮组件
    Button: {
      primaryShadow: '0 2px 4px rgba(30, 90, 141, 0.2)', // 主按钮阴影
      defaultShadow: '0 1px 2px rgba(0, 0, 0, 0.05)', // 默认按钮阴影
      paddingContentHorizontal: 20, // 水平内边距
      paddingInline: 20,
      paddingBlock: 8, // 垂直内边距
      fontWeight: 500, // 字重
      borderRadius: 6, // 圆角

      // 状态颜色
      colorPrimaryHover: '#1A4B75', // 悬停 - 深蓝
      colorPrimaryActive: '#163C5D', // 激活 - 更深蓝

      // 尺寸
      controlHeight: 40, // 默认高度
      controlHeightLG: 48, // 大按钮
      controlHeightSM: 32, // 小按钮
    },

    // 卡片组件
    Card: {
      borderRadiusLG: 8, // 圆角
      boxShadowTertiary: '0 1px 2px rgba(0, 0, 0, 0.05)', // 卡片阴影
      paddingLG: 24, // 内边距
      headerHeight: 56, // 头部高度
    },

    // 输入框组件
    Input: {
      borderRadius: 4, // 圆角
      paddingInline: 12, // 水平内边距
      paddingBlock: 8, // 垂直内边距
      activeBorderColor: '#1E5A8D', // 激活边框 - 主深蓝
      hoverBorderColor: '#2D6FA8', // 悬停边框 - 浅蓝
      colorBorder: '#E8EBF0', // 默认边框
      colorTextPlaceholder: '#9CA3AF', // 占位符文字
      controlHeight: 40, // 输入框高度
    },

    // 密码输入框
    Password: {
      borderRadius: 4,
      paddingInline: 12,
      paddingBlock: 8,
      activeBorderColor: '#1E5A8D',
      hoverBorderColor: '#2D6FA8',
      controlHeight: 40,
    },

    // 表单组件
    Form: {
      itemMarginBottom: 20, // 表单项底部间距
      verticalLabelPadding: '0 0 8px', // 垂直标签内边距
      labelRequiredMarkColor: '#E8484C', // 必填星号颜色 - 错误红
      labelColor: '#374151', // 标签文字颜色
      fontSize: 14, // 字号
    },

    // 表格组件
    Table: {
      headerBg: '#F4F6F8', // 表头背景
      headerSplitColor: '#E8EBF0', // 表头分割线
      borderColor: '#E8EBF0', // 边框颜色
      rowHoverBg: '#F0F7FF', // 行悬停背景 - 最浅蓝
      fontSize: 14, // 字号
      padding: 16, // 内边距
      paddingContentVertical: 16,
      paddingContentHorizontal: 16,

      // 固定列阴影
      cellPaddingInline: 16,
      cellPaddingBlock: 16,
    },

    // 菜单组件
    Menu: {
      itemBorderRadius: 6, // 菜单项圆角
      itemMarginInline: 4, // 水平间距
      itemMarginBlock: 4, // 垂直间距
      itemHeight: 40, // 菜单项高度
      itemPaddingInlineDirection: 'horizontal', // 内边距方向
      itemPaddingInline: 16, // 水平内边距
      itemPaddingBlock: 8, // 垂直内边距

      // 颜色
      itemActiveBg: 'rgba(30, 90, 141, 0.1)', // 激活背景
      itemSelectedBg: '#C5A572', // 选中背景 - 金色点缀
      itemSelectedColor: '#FFFFFF', // 选中文字 - 白色
    },

    // 模态框组件
    Modal: {
      borderRadiusLG: 12, // 圆角
      paddingLG: 24, // 内边距
      padding: 24,
      headerBg: '#FFFFFF', // 头部背景
      headerBorderRadius: 12,
      contentBg: '#FFFFFF', // 内容背景
    },

    // 下拉菜单
    Dropdown: {
      borderRadius: 8, // 圆角
      boxShadowSecondary: '0 10px 15px -3px rgba(0, 0, 0, 0.10)', // 阴影
    },

    // 通知组件
    Notification: {
      borderRadiusLG: 8,
      paddingContent: 16,
    },

    // 消息组件
    Message: {
      borderRadiusLG: 8,
      paddingContent: 16,
    },

    // 标签组件
    Tag: {
      borderRadiusSM: 4,
      defaultBg: '#F4F6F8',
      defaultColor: '#6B7280',
    },

    // 分页组件
    Pagination: {
      borderRadius: 6,
      itemSize: 40,
    },

    // 选择器
    Select: {
      borderRadius: 4,
      optionSelectedBg: '#F0F7FF',
      optionActiveBg: '#E0EFFD',
    },

    // 复选框
    Checkbox: {
      borderRadius: 4,
      colorPrimary: '#1E5A8D',
    },

    // 单选框
    Radio: {
      borderRadius: 4,
      colorPrimary: '#1E5A8D',
    },

    // 开关
    Switch: {
      colorPrimary: '#1E5A8D',
      colorPrimaryHover: '#2D6FA8',
    },

    // 滑块
    Slider: {
      colorPrimary: '#1E5A8D',
      colorPrimaryBorderHover: '#1A4B75',
      trackBg: '#E8EBF0',
      trackHoverBg: '#D1D5DB',
    },

    // 进度条
    Progress: {
      colorSuccess: '#3FAF56',
      colorInfo: '#1E5A8D',
      colorWarning: '#F5A623',
      colorError: '#E8484C',
    },

    // 时间选择器
    TimePicker: {
      borderRadius: 4,
      cellHoverBg: '#F0F7FF',
      cellActiveBg: '#1E5A8D',
    },

    // 日期选择器
    DatePicker: {
      borderRadius: 4,
      activeBarWidth: 40,
      cellHoverBg: '#F0F7FF',
      cellActiveBg: '#1E5A8D',
    },

    // 上传组件
    Upload: {
      colorPrimary: '#1E5A8D',
    },

    // 树形控件
    Tree: {
      nodeSelectedBg: '#F0F7FF',
      nodeHoverBg: '#E0EFFD',
    },

    // 标签页
    Tabs: {
      itemActiveColor: '#1E5A8D',
      itemSelectedColor: '#1E5A8D',
      inkBarColor: '#1E5A8D',
    },

    // 步骤条
    Steps: {
      colorPrimary: '#1E5A8D',
    },

    // 骨架屏
    Skeleton: {
      colorBase: '#E8EBF0',
    },

    // 空状态
    Empty: {
      colorText: '#6B7280',
    },

    // 警告框
    Alert: {
      borderRadius: 8,
      defaultPadding: 16,
    },

    // 抽屉
    Drawer: {
      borderRadiusLG: 12,
      paddingLG: 24,
    },

    // 弹出框
    Popover: {
      borderRadius: 8,
      boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.10)',
    },

    // 工具提示
    Tooltip: {
      borderRadius: 6,
      colorBgDefault: 'rgba(0, 0, 0, 0.85)',
    },
  } as ThemeConfig['components'],

  // ========================================
  // 算法配置 (Algorithm)
  // ========================================
  algorithm: [
    // 使用默认算法，确保颜色系统的正确性
    // Ant Design 会根据 colorPrimary 自动生成派生色
  ],
}

/**
 * 获取金色主题变体（用于特殊场景）
 *
 * 注意：金色仅用于强调和点缀，不超过5%面积
 */
export const goldenAccentTheme: ThemeConfig = {
  ...antdTheme,
  token: {
    ...antdTheme.token,
    colorPrimary: '#C5A572', // 品质金色
  },
}

/**
 * 暗色模式主题（预留）
 */
export const darkTheme: ThemeConfig = {
  ...antdTheme,
  token: {
    ...antdTheme.token,
    colorBgBase: '#0F172A',
    colorBgContainer: '#1E293B',
    colorBgLayout: '#0F172A',
    colorText: '#E5E7EB',
    colorTextSecondary: '#9CA3AF',
  },
}
