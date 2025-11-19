/**
 * 统一表格组件
 * 基于设计系统，提供一致的表格样式和功能
 */
import React from 'react'
import { Table, TableProps, Tag, Space, Tooltip } from 'antd'
import { DESIGN_TOKENS, BUSINESS_STATUS, designUtils } from '@/constants/design-system'
import type { ColumnsType } from 'antd/es/table'

export interface StandardTableProps<T = any> extends Omit<TableProps<T>, 'style' | 'className'> {
  variant?: 'default' | 'bordered' | 'striped'
  showHeader?: boolean
  size?: 'small' | 'middle' | 'large'
  className?: string
  style?: React.CSSProperties
}

const StandardTable = <T extends Record<string, any>>({
  variant = 'default',
  showHeader = true,
  size = 'middle',
  className,
  style,
  ...tableProps
}: StandardTableProps<T>) => {
  // 表格容器样式
  const tableContainerStyle: React.CSSProperties = {
    background: DESIGN_TOKENS.colors.bgCard,
    borderRadius: DESIGN_TOKENS.radius.lg,
    border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
    overflow: 'hidden',
    boxShadow: DESIGN_TOKENS.shadows.sm,
    ...style,
  }

  // 表格样式
  const tableStyle: React.CSSProperties = {
    fontSize: DESIGN_TOKENS.typography.base.fontSize,
  }

  // 根据变体添加自定义类名
  const getClassName = () => {
    const classNames = ['standard-table']

    if (variant === 'striped') {
      classNames.push('standard-table--striped')
    }

    if (variant === 'bordered') {
      classNames.push('standard-table--bordered')
    }

    if (className) {
      classNames.push(className)
    }

    return classNames.join(' ')
  }

  return (
    <div style={tableContainerStyle}>
      <Table<T>
        {...tableProps}
        className={getClassName()}
        style={tableStyle}
        size={size}
        showHeader={showHeader}
      />
    </div>
  )
}

// =============================================================================
// 标准列生成器
// =============================================================================

export const createStandardColumns = {
  // 创建状态列
  createStatusColumn: <T extends Record<string, any>>(
    dataIndex: keyof T,
    title: string = '状态',
    statusType: 'case' | 'user' | 'priority' | 'approval' = 'case',
    width: number = 100,
  ): ColumnsType<T>[number] => ({
    title,
    dataIndex: dataIndex as string,
    key: dataIndex as string,
    width,
    render: (status: string) => {
      const statusMap = {
        case: BUSINESS_STATUS.CASE_STATUS,
        user: BUSINESS_STATUS.USER_TYPE,
        priority: BUSINESS_STATUS.PRIORITY,
        approval: BUSINESS_STATUS.APPROVAL_STATUS,
      }

      const currentStatusMap = statusMap[statusType]
      const statusKey = Object.keys(currentStatusMap).find(
        (key) =>
          key === status?.toUpperCase() ||
          currentStatusMap[key as keyof typeof currentStatusMap].text === status,
      )

      if (statusKey) {
        const statusConfig = currentStatusMap[statusKey as keyof typeof currentStatusMap]
        return (
          <Tag
            color={statusConfig.color}
            style={{
              borderRadius: DESIGN_TOKENS.radius.sm,
              fontSize: DESIGN_TOKENS.typography.xs.fontSize,
              fontWeight: '500',
            }}
          >
            {statusConfig.text}
          </Tag>
        )
      }

      return (
        <Tag
          color={DESIGN_TOKENS.colors.textTertiary}
          style={{
            borderRadius: DESIGN_TOKENS.radius.sm,
            fontSize: DESIGN_TOKENS.typography.xs.fontSize,
          }}
        >
          {status || '未知'}
        </Tag>
      )
    },
  }),

  // 创建操作列
  createActionColumn: <T extends Record<string, any>>(
    actions: Array<{
      key: string
      label: string
      icon?: React.ReactNode
      onClick: (record: T) => void
      danger?: boolean
      disabled?: boolean
      tooltip?: string
    }>,
    width: number = 150,
    fixed?: 'left' | 'right',
  ): ColumnsType<T>[number] => ({
    title: '操作',
    key: 'action',
    width,
    fixed,
    render: (_, record) => (
      <Space size={DESIGN_TOKENS.spacing.sm}>
        {actions.map((action) => {
          const button = (
            <Button
              key={action.key}
              type='link'
              size='small'
              icon={action.icon}
              onClick={() => action.onClick(record)}
              danger={action.danger}
              disabled={action.disabled}
              style={{
                borderRadius: DESIGN_TOKENS.radius.sm,
                fontSize: DESIGN_TOKENS.typography.sm.fontSize,
                padding: `${DESIGN_TOKENS.spacing.xs} ${DESIGN_TOKENS.spacing.sm}`,
              }}
            >
              {action.label}
            </Button>
          )

          return action.tooltip ? (
            <Tooltip key={action.key} title={action.tooltip}>
              {button}
            </Tooltip>
          ) : (
            button
          )
        })}
      </Space>
    ),
  }),

  // 创建时间列
  createTimeColumn: <T extends Record<string, any>>(
    dataIndex: keyof T,
    title: string = '时间',
    width: number = 180,
    format?: string,
  ): ColumnsType<T>[number] => ({
    title,
    dataIndex: dataIndex as string,
    key: dataIndex as string,
    width,
    render: (time: string) => {
      if (!time) {
        return '-'
      }

      try {
        const date = new Date(time)
        const formatString = format || 'YYYY-MM-DD HH:mm'
        return date.toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
        })
      } catch {
        return time
      }
    },
  }),

  // 创建用户信息列
  createUserColumn: <T extends Record<string, any>>(
    dataIndex: keyof T,
    title: string = '用户',
    width: number = 150,
    showAvatar: boolean = true,
  ): ColumnsType<T>[number] => ({
    title,
    dataIndex: dataIndex as string,
    key: dataIndex as string,
    width,
    render: (user: string | { name: string; avatar?: string }) => {
      const userName = typeof user === 'string' ? user : user?.name
      const userAvatar = typeof user === 'object' ? user?.avatar : undefined

      return (
        <Space size={DESIGN_TOKENS.spacing.sm}>
          {showAvatar && (
            <div
              style={{
                width: '24px',
                height: '24px',
                borderRadius: DESIGN_TOKENS.radius.full,
                background: DESIGN_TOKENS.colors.primaryLight,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: DESIGN_TOKENS.typography.xs.fontSize,
                color: DESIGN_TOKENS.colors.primary,
                fontWeight: '500',
              }}
            >
              {userName?.charAt(0)?.toUpperCase()}
            </div>
          )}
          <span
            style={{
              fontSize: DESIGN_TOKENS.typography.base.fontSize,
              color: DESIGN_TOKENS.colors.textPrimary,
            }}
          >
            {userName || '-'}
          </span>
        </Space>
      )
    },
  }),
}

export default StandardTable
