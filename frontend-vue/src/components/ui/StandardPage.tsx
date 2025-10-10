/**
 * 统一页面布局组件
 * 基于设计系统，提供一致的页面布局和样式
 */
import React from 'react';
import { Row, Col, Space, Button, Tooltip } from 'antd';
import { ReloadOutlined, PlusOutlined } from '@ant-design/icons';
import { DESIGN_TOKENS } from '@/constants/design-system';

export interface StandardPageProps {
  title?: string;
  subtitle?: string;
  children: React.ReactNode;
  actions?: React.ReactNode;
  showRefresh?: boolean;
  showAdd?: boolean;
  onRefresh?: () => void;
  onAdd?: () => void;
  extra?: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

const StandardPage: React.FC<StandardPageProps> = ({
  title,
  subtitle,
  children,
  actions,
  showRefresh = true,
  showAdd = false,
  onRefresh,
  onAdd,
  extra,
  className,
  style,
}) => {
  // 页面容器样式
  const pageContainerStyle: React.CSSProperties = {
    background: DESIGN_TOKENS.colors.bgPage,
    minHeight: '100vh',
    padding: DESIGN_TOKENS.spacing.lg,
  };

  // 内容区域样式
  const contentAreaStyle: React.CSSProperties = {
    background: DESIGN_TOKENS.colors.bgContainer,
    borderRadius: DESIGN_TOKENS.radius.xl,
    padding: DESIGN_TOKENS.spacing.xxl,
    boxShadow: DESIGN_TOKENS.shadows.md,
    ...style,
  };

  // 头部区域样式
  const headerStyle: React.CSSProperties = {
    marginBottom: DESIGN_TOKENS.spacing.xxl,
    paddingBottom: DESIGN_TOKENS.spacing.xl,
    borderBottom: `1px solid ${DESIGN_TOKENS.colors.borderSplit}`,
  };

  // 标题样式
  const titleStyle: React.CSSProperties = {
    fontSize: DESIGN_TOKENS.typography.xxxl.fontSize,
    fontWeight: '600',
    color: DESIGN_TOKENS.colors.textPrimary,
    margin: 0,
    marginBottom: DESIGN_TOKENS.spacing.sm,
  };

  // 副标题样式
  const subtitleStyle: React.CSSProperties = {
    fontSize: DESIGN_TOKENS.typography.base.fontSize,
    color: DESIGN_TOKENS.colors.textSecondary,
    margin: 0,
  };

  // 操作按钮样式
  const actionButtonStyle = {
    borderRadius: DESIGN_TOKENS.radius.md,
    boxShadow: DESIGN_TOKENS.shadows.sm,
    transition: `all ${DESIGN_TOKENS.animation.durationNormal} ${DESIGN_TOKENS.animation.easeOut}`,
  };

  return (
    <div className={className} style={pageContainerStyle}>
      <div style={contentAreaStyle}>
        {/* 页面头部 */}
        {(title || subtitle || actions || showRefresh || showAdd) && (
          <div style={headerStyle}>
            <Row justify="space-between" align="middle">
              <Col>
                {title && (
                  <h1 style={titleStyle}>
                    {title}
                  </h1>
                )}
                {subtitle && (
                  <p style={subtitleStyle}>
                    {subtitle}
                  </p>
                )}
              </Col>
              <Col>
                <Space size={DESIGN_TOKENS.spacing.md}>
                  {/* 默认操作按钮 */}
                  {showRefresh && (
                    <Tooltip title="刷新数据">
                      <Button
                        icon={<ReloadOutlined />}
                        onClick={onRefresh}
                        style={actionButtonStyle}
                      >
                        刷新
                      </Button>
                    </Tooltip>
                  )}
                  {showAdd && (
                    <Button
                      type="primary"
                      icon={<PlusOutlined />}
                      onClick={onAdd}
                      style={{
                        ...actionButtonStyle,
                        background: `linear-gradient(135deg, ${DESIGN_TOKENS.colors.primary} 0%, ${DESIGN_TOKENS.colors.primaryHover} 100%)`,
                        border: 'none',
                      }}
                    >
                      新增
                    </Button>
                  )}

                  {/* 自定义操作按钮 */}
                  {actions}
                </Space>
              </Col>
            </Row>

            {/* 额外内容 */}
            {extra && (
              <div style={{ marginTop: DESIGN_TOKENS.spacing.lg }}>
                {extra}
              </div>
            )}
          </div>
        )}

        {/* 页面内容 */}
        <div>
          {children}
        </div>
      </div>
    </div>
  );
};

export default StandardPage;