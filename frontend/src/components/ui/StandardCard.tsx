/**
 * 统一卡片组件
 * 基于设计系统，提供一致的卡片样式
 */
import React from 'react';
import { Card, CardProps } from 'antd';
import { DESIGN_TOKENS, createComponentStyles } from '@/constants/design-system';

export interface StandardCardProps extends Omit<CardProps, 'style' | 'className'> {
  variant?: 'default' | 'hoverable' | 'bordered' | 'statistic';
  color?: string;
  padding?: 'sm' | 'md' | 'lg' | 'xl';
  children: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

const StandardCard: React.FC<StandardCardProps> = ({
  variant = 'default',
  color,
  padding = 'lg',
  children,
  className,
  style,
  ...cardProps
}) => {
  // 获取基础样式
  const getBaseStyles = () => {
    const baseStyles = {
      background: DESIGN_TOKENS.colors.bgCard,
      borderRadius: DESIGN_TOKENS.radius.lg,
      border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
      transition: `all ${DESIGN_TOKENS.animation.durationNormal} ${DESIGN_TOKENS.animation.easeOut}`,
    };

    // 根据变体添加样式
    switch (variant) {
      case 'hoverable':
        return {
          ...baseStyles,
          boxShadow: DESIGN_TOKENS.shadows.md,
          cursor: 'pointer',
          '&:hover': {
            boxShadow: DESIGN_TOKENS.shadows.lg,
            transform: 'translateY(-2px)',
          },
        };

      case 'statistic':
        return {
          ...baseStyles,
          borderLeft: `4px solid ${color || DESIGN_TOKENS.colors.primary}`,
          boxShadow: DESIGN_TOKENS.shadows.sm,
          '&:hover': {
            boxShadow: DESIGN_TOKENS.shadows.md,
            transform: 'translateY(-1px)',
          },
        };

      case 'bordered':
        return {
          ...baseStyles,
          border: `2px solid ${DESIGN_TOKENS.colors.borderBase}`,
        };

      default:
        return {
          ...baseStyles,
          boxShadow: DESIGN_TOKENS.shadows.sm,
        };
    }
  };

  // 获取内边距
  const getPadding = () => {
    const paddingMap = {
      sm: DESIGN_TOKENS.spacing.md,
      md: DESIGN_TOKENS.spacing.lg,
      lg: DESIGN_TOKENS.spacing.xl,
      xl: DESIGN_TOKENS.spacing.xxl,
    };
    return paddingMap[padding];
  };

  // 合并样式
  const mergedStyle = React.useMemo(() => {
    const baseStyles = getBaseStyles();
    const paddingStyle = {
      padding: getPadding(),
    };

    return {
      ...baseStyles,
      ...paddingStyle,
      ...style,
    };
  }, [variant, color, padding, style]);

  return (
    <Card
      className={className}
      style={mergedStyle}
      {...cardProps}
    >
      {children}
    </Card>
  );
};

export default StandardCard;