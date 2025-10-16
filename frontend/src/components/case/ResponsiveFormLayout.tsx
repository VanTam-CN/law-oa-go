import React, { useState, useEffect, useMemo } from 'react';
import {
  Row,
  Col,
  ConfigProvider,
  theme,
  Grid
} from 'antd';
import {
  ResponsiveFormLayoutProps,
  BreakpointConfig,
  SpacingConfig
} from './types/ResponsiveFormLayout.types';
import './ResponsiveFormLayout.less';

const { useToken } = theme;

/**
 * 1080p优化的响应式表单布局组件
 * 支持动态列数调整和紧凑间距设置
 * 专为案件创建表单设计，充分利用屏幕空间
 */
const ResponsiveFormLayout: React.FC<ResponsiveFormLayoutProps> = ({
  children,
  columns = 2,
  spacing = 'medium',
  breakpoint = 'lg',
  className = '',
  style,
  ...restProps
}) => {
  const { token } = useToken();
  const [currentBreakpoint, setCurrentBreakpoint] = useState<string>();
  const [isCompact, setIsCompact] = useState(false);

  // 获取当前响应式断点
  const screens = Grid.useBreakpoint();

  // 根据断点计算实际列数
  const actualColumns = useMemo(() => {
    const {
      xs = 24,
      sm = 24,
      md = 12,
      lg = 8,
      xl = 6,
      xxl = 4
    } = screens;

    // 1080p (通常对应lg断点) 的优化逻辑
    if (currentBreakpoint === 'lg') {
      // 1080p显示器上的紧凑布局
      return columns === 3 ? 3 : columns;
    } else if (currentBreakpoint === 'xl' || currentBreakpoint === 'xxl') {
      // 大屏幕可以显示更多列
      return columns === 3 ? 4 : Math.min(columns + 1, 4);
    } else {
      // 小屏幕使用单列或双列
      return currentBreakpoint === 'md' ? 2 : 1;
    }
  }, [columns, currentBreakpoint, screens]);

  // 获取间距配置
  const spacingConfig = useMemo<SpacingConfig>(() => {
    const baseSpacing = {
      small: 8,
      medium: 16,
      large: 24
    };

    // 1080p上使用更紧凑的间距
    const compactMultiplier = isCompact ? 0.75 : 1;

    return {
      horizontal: baseSpacing[spacing] * compactMultiplier,
      vertical: baseSpacing[spacing] * compactMultiplier
    };
  }, [spacing, isCompact]);

  // 监听断点变化
  useEffect(() => {
    const breakpointKeys = Object.keys(screens) as string[];
    const currentActiveBreakpoint = breakpointKeys.find(
      key => screens[key as keyof typeof screens]
    );

    if (currentActiveBreakpoint) {
      setCurrentBreakpoint(currentActiveBreakpoint);
      // 1080p (lg) 时启用紧凑模式
      setIsCompact(currentActiveBreakpoint === 'lg');
    }
  }, [screens]);

  // 1080p优化的栅格配置
  const gridConfig = useMemo<BreakpointConfig>(() => ({
    // 针对不同断点的列配置
    xs: { span: 24 }, // 超小屏幕：单列
    sm: { span: 24 }, // 小屏幕：单列
    md: { span: actualColumns === 1 ? 24 : 12 }, // 中等屏幕：单列或双列
    lg: { span: actualColumns === 3 ? 8 : actualColumns === 2 ? 12 : 24 }, // 1080p：优化列数
    xl: { span: actualColumns === 3 ? 6 : actualColumns === 2 ? 12 : 8 }, // 大屏幕：更多列
    xxl: { span: actualColumns === 3 ? 6 : actualColumns === 2 ? 8 : 6 } // 超大屏幕
  }), [actualColumns]);

  // 计算容器样式，针对1080p优化
  const containerStyle = useMemo(() => {
    const baseStyle: React.CSSProperties = {
      width: '100%',
      minHeight: '100vh',
      padding: `${spacingConfig.vertical}px`,
      boxSizing: 'border-box',
      ...style
    };

    // 1080p特定优化
    if (currentBreakpoint === 'lg') {
      return {
        ...baseStyle,
        maxWidth: '1920px', // 1080p标准宽度
        margin: '0 auto',
        overflow: 'hidden auto',
        // 减少不必要的边距
        '--form-padding': `${Math.max(12, spacingConfig.horizontal)}px`,
        '--form-gap': `${Math.max(12, spacingConfig.vertical)}px`
      };
    }

    return baseStyle;
  }, [currentBreakpoint, spacingConfig, style]);

  // 渲染子组件，应用响应式布局
  const renderChildren = () => {
    if (Array.isArray(children)) {
      return children.map((child, index) => {
        const colConfig = typeof child === 'object' && child && 'props' in child
          ? child.props
          : {};

        return (
          <Col
            key={index}
            {...gridConfig}
            {...colConfig}
            className={`responsive-form-item responsive-form-item-${index} ${colConfig.className || ''}`}
          >
            {child}
          </Col>
        );
      });
    }

    return (
      <Col {...gridConfig} className="responsive-form-item">
        {children}
      </Col>
    );
  };

  return (
    <ConfigProvider
      theme={{
        token: {
          ...token,
          // 1080p优化的设计令牌
          controlHeight: isCompact ? 28 : 32,
          controlHeightSM: isCompact ? 20 : 24,
          controlHeightLG: isCompact ? 36 : 40,
          paddingXS: isCompact ? 4 : 8,
          paddingSM: isCompact ? 8 : 12,
          padding: isCompact ? 12 : 16,
          paddingLG: isCompact ? 16 : 24,
          paddingXL: isCompact ? 20 : 32,
          marginXS: isCompact ? 4 : 8,
          marginSM: isCompact ? 8 : 12,
          margin: isCompact ? 12 : 16,
          marginLG: isCompact ? 16 : 24,
          marginXL: isCompact ? 20 : 32,
          // 紧凑模式下减少圆角
          borderRadius: isCompact ? 4 : 6,
          borderRadiusSM: isCompact ? 2 : 4,
          borderRadiusLG: isCompact ? 6 : 8
        }
      }}
    >
      <div
        className={`responsive-form-layout ${isCompact ? 'responsive-form-layout-compact' : ''} responsive-form-layout-${currentBreakpoint} ${className}`}
        style={containerStyle}
        {...restProps}
      >
        <Row
          gutter={[spacingConfig.horizontal, spacingConfig.vertical]}
          className="responsive-form-row"
        >
          {renderChildren()}
        </Row>

        {/* 1080p优化提示 */}
        {currentBreakpoint === 'lg' && (
          <div className="responsive-hint">
            <span className="responsive-hint-text">
              📺 已为1080p显示器优化布局
            </span>
          </div>
        )}
      </div>
    </ConfigProvider>
  );
};

export default ResponsiveFormLayout;