/**
 * ConflictCheckInline组件
 * 内联冲突检测组件，支持1080p优化和快速结果显示
 */

import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import {
  Card,
  Alert,
  Button,
  Space,
  Typography,
  Spin,
  Badge,
  Tag,
  Tooltip,
  Divider,
  List,
  Avatar,
  Progress,
  Empty,
  Collapse,
} from 'antd'
import {
  ReloadOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
  SearchOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  RightOutlined,
  DownOutlined,
} from '@ant-design/icons'
import type { CollapseProps } from 'antd/es/collapse'
import type {
  ConflictCheckInlineProps,
  ConflictCheckResult,
  ConflictCheckStatus,
  ConflictCase,
  ConflictSeverity,
  InlineDisplayConfig,
  QuickActionConfig,
  ConflictStatsCardProps,
  ConflictListItemProps,
  QuickActionsProps,
  CheckStatusIndicatorProps,
} from './types/ConflictCheckInline.types'
import {
  DEFAULT_INLINE_DISPLAY_CONFIG,
  DEFAULT_QUICK_ACTION_CONFIG,
  DEFAULT_SEVERITY_CONFIG,
} from './types/ConflictCheckInline.types'
import './ConflictCheckInline.less'

const { Text, Title, Paragraph } = Typography
const { Panel } = Collapse

/**
 * 冲突严重程度颜色获取
 */
const getSeverityColor = (severity: ConflictSeverity): string => {
  return DEFAULT_SEVERITY_CONFIG.colors[severity] || '#d9d9d9'
}

/**
 * 冲突严重程度标签获取
 */
const getSeverityLabel = (severity: ConflictSeverity): string => {
  return DEFAULT_SEVERITY_CONFIG.labels[severity] || '未知'
}

/**
 * 冲突状态文本映射
 */
const STATUS_TEXT: Record<ConflictCheckStatus, string> = {
  idle: '等待检测',
  checking: '检测中...',
  success: '无冲突',
  warning: '发现潜在冲突',
  error: '检测失败',
}

/**
 * 冲突状态颜色映射
 */
const STATUS_COLOR: Record<ConflictCheckStatus, string> = {
  idle: '#d9d9d9',
  checking: '#1890ff',
  success: '#52c41a',
  warning: '#faad14',
  error: '#ff4d4f',
}

/**
 * 检测状态指示器组件
 */
const CheckStatusIndicator: React.FC<CheckStatusIndicatorProps> = ({
  status,
  progress = 0,
  showText = true,
  isCompact,
  statusText = STATUS_TEXT,
  className = '',
  style,
}) => {
  if (status === 'checking') {
    return (
      <div
        className={`check-status-indicator checking ${isCompact ? 'compact' : ''} ${className}`}
        style={style}
      >
        <Spin size={isCompact ? 'small' : 'default'} />
        {showText && (
          <Text type='secondary' style={{ marginLeft: 8 }}>
            {statusText[status]} {progress > 0 && `(${progress}%)`}
          </Text>
        )}
      </div>
    )
  }

  const getStatusIcon = () => {
    switch (status) {
      case 'success':
        return <CheckCircleOutlined style={{ color: STATUS_COLOR[status] }} />
      case 'warning':
        return <WarningOutlined style={{ color: STATUS_COLOR[status] }} />
      case 'error':
        return <CloseCircleOutlined style={{ color: STATUS_COLOR[status] }} />
      default:
        return <InfoCircleOutlined style={{ color: STATUS_COLOR[status] }} />
    }
  }

  return (
    <div
      className={`check-status-indicator ${status} ${isCompact ? 'compact' : ''} ${className}`}
      style={style}
    >
      {getStatusIcon()}
      {showText && (
        <Text type='secondary' style={{ marginLeft: 8 }}>
          {statusText[status]}
        </Text>
      )}
    </div>
  )
}

/**
 * 冲突统计卡片组件
 */
const ConflictStatsCard: React.FC<ConflictStatsCardProps> = ({
  stats,
  total,
  isCompact,
  className = '',
  style,
}) => {
  if (total === 0) {
    return null
  }

  const items = [
    { label: '直接冲突', value: stats.direct, color: '#ff4d4f' },
    { label: '间接冲突', value: stats.indirect, color: '#faad14' },
    { label: '潜在冲突', value: stats.potential, color: '#1890ff' },
  ].filter((item) => item.value > 0)

  return (
    <div className={`conflict-stats-card ${isCompact ? 'compact' : ''} ${className}`} style={style}>
      <Space
        direction={isCompact ? 'vertical' : 'horizontal'}
        size={isCompact ? 'small' : 'middle'}
      >
        <Badge count={total} size={isCompact ? 'small' : 'default'}>
          <Text strong>发现冲突</Text>
        </Badge>
        <Divider type={isCompact ? 'horizontal' : 'vertical'} />
        <Space size={isCompact ? 'small' : 'middle'}>
          {items.map((item, index) => (
            <Tag key={index} color={item.color} style={{ margin: 0 }}>
              {item.label}: {item.value}
            </Tag>
          ))}
        </Space>
      </Space>
    </div>
  )
}

/**
 * 冲突列表项组件
 */
const ConflictListItem: React.FC<ConflictListItemProps> = ({
  conflict,
  index,
  showDetails,
  isCompact,
  onViewDetails,
  onMarkResolved,
  className = '',
  style,
}) => {
  const [expanded, setExpanded] = useState(false)

  const handleViewDetails = useCallback(() => {
    onViewDetails?.(conflict)
  }, [onViewDetails, conflict])

  const handleMarkResolved = useCallback(() => {
    onMarkResolved?.(conflict.id)
  }, [onMarkResolved, conflict.id])

  return (
    <div
      className={`conflict-list-item ${isCompact ? 'compact' : ''} severity-${conflict.severity} ${className}`}
      style={style}
    >
      <div className='conflict-item-header'>
        <Space>
          <Avatar
            size={isCompact ? 'small' : 'default'}
            style={{
              backgroundColor: getSeverityColor(conflict.severity),
            }}
          >
            {index + 1}
          </Avatar>
          <div className='conflict-item-info'>
            <div className='conflict-item-title'>
              <Text strong>{conflict.title}</Text>
              <Tag
                size={isCompact ? 'small' : 'default'}
                color={getSeverityColor(conflict.severity)}
              >
                {getSeverityLabel(conflict.severity)}
              </Tag>
            </div>
            <div className='conflict-item-meta'>
              <Text type='secondary' style={{ fontSize: isCompact ? 12 : 14 }}>
                案件编号: {conflict.caseNumber} | 客户: {conflict.clientName} | 律师:{' '}
                {conflict.lawyerName}
              </Text>
            </div>
          </div>
        </Space>
        <Space>
          {showDetails && (
            <Tooltip title='查看详情'>
              <Button
                type='text'
                size={isCompact ? 'small' : 'middle'}
                icon={<EyeOutlined />}
                onClick={handleViewDetails}
              />
            </Tooltip>
          )}
          {onMarkResolved && (
            <Tooltip title='标记为已解决'>
              <Button
                type='text'
                size={isCompact ? 'small' : 'middle'}
                icon={<CheckCircleOutlined />}
                onClick={handleMarkResolved}
              />
            </Tooltip>
          )}
        </Space>
      </div>

      {showDetails && expanded && (
        <div className='conflict-item-details'>
          <Divider style={{ margin: '8px 0' }} />
          <Paragraph style={{ margin: 0, fontSize: isCompact ? 12 : 14 }}>
            {conflict.description}
          </Paragraph>
          {conflict.recommendation && (
            <div style={{ marginTop: 8 }}>
              <Text type='secondary' style={{ fontSize: isCompact ? 11 : 12 }}>
                建议: {conflict.recommendation}
              </Text>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

/**
 * 快速操作组件
 */
const QuickActions: React.FC<QuickActionsProps> = ({
  status,
  hasConflict,
  config,
  isCompact,
  onRecheck,
  onViewAllConflicts,
  onMarkAllResolved,
  onCustomAction,
  className = '',
  style,
}) => {
  const actions = []

  if (config.allowRecheck && status !== 'checking') {
    actions.push(
      <Tooltip key='recheck' title='重新检测'>
        <Button
          type='text'
          size={isCompact ? 'small' : 'middle'}
          icon={<ReloadOutlined />}
          onClick={onRecheck}
        />
      </Tooltip>,
    )
  }

  if (hasConflict && config.allowViewDetails) {
    actions.push(
      <Tooltip key='view-all' title='查看所有冲突'>
        <Button
          type='text'
          size={isCompact ? 'small' : 'middle'}
          icon={<EyeOutlined />}
          onClick={onViewAllConflicts}
        />
      </Tooltip>,
    )
  }

  if (hasConflict && config.allowMarkResolved) {
    actions.push(
      <Tooltip key='mark-all-resolved' title='标记全部已解决'>
        <Button
          type='text'
          size={isCompact ? 'small' : 'middle'}
          icon={<CheckCircleOutlined />}
          onClick={onMarkAllResolved}
        />
      </Tooltip>,
    )
  }

  // 添加自定义操作
  config.customActions?.forEach((action) => {
    actions.push(
      <Tooltip key={action.key} title={action.label}>
        <Button
          type='text'
          size={isCompact ? 'small' : 'middle'}
          icon={action.icon}
          danger={action.danger}
          onClick={() => {
            action.onClick()
            onCustomAction?.(action.key)
          }}
        />
      </Tooltip>,
    )
  })

  if (actions.length === 0) {
    return null
  }

  return (
    <div className={`quick-actions ${isCompact ? 'compact' : ''} ${className}`} style={style}>
      <Space size={isCompact ? 'small' : 'middle'}>{actions}</Space>
    </div>
  )
}

/**
 * ConflictCheckInline主组件
 */
const ConflictCheckInline: React.FC<ConflictCheckInlineProps> = ({
  checkParams,
  result,
  onStatusChange,
  onCheckComplete,
  onViewDetails,
  onRecheck,
  onMarkResolved,
  displayConfig,
  actionConfig,
  className = '',
  style,
  autoCheck = true,
  debounceDelay = 500,
}) => {
  const [currentResult, setCurrentResult] = useState<ConflictCheckResult | undefined>(result)
  const [status, setStatus] = useState<ConflictCheckStatus>('idle')
  const [checkingProgress, setCheckingProgress] = useState(0)
  const [isCompact, setIsCompact] = useState(false)

  // 合并配置
  const finalDisplayConfig = useMemo<InlineDisplayConfig>(
    () => ({
      ...DEFAULT_INLINE_DISPLAY_CONFIG,
      ...displayConfig,
    }),
    [displayConfig],
  )

  const finalActionConfig = useMemo<QuickActionConfig>(
    () => ({
      ...DEFAULT_QUICK_ACTION_CONFIG,
      ...actionConfig,
    }),
    [actionConfig],
  )

  const checkTimeoutRef = useRef<NodeJS.Timeout>()

  /**
   * 检测1080p分辨率
   */
  useEffect(() => {
    const checkResolution = () => {
      setIsCompact(window.innerWidth <= 1920 && window.innerWidth >= 1600)
    }

    checkResolution()
    window.addEventListener('resize', checkResolution)
    return () => window.removeEventListener('resize', checkResolution)
  }, [])

  /**
   * 执行冲突检测
   */
  const performCheck = useCallback(async () => {
    if (status === 'checking') {
      return
    }

    setStatus('checking')
    setCheckingProgress(0)
    onStatusChange?.('checking')

    try {
      // 模拟检测进度
      const progressInterval = setInterval(() => {
        setCheckingProgress((prev) => {
          if (prev >= 90) {
            clearInterval(progressInterval)
            return 90
          }
          return prev + 10
        })
      }, 100)

      // 这里应该调用实际的冲突检测API
      await new Promise((resolve) => setTimeout(resolve, 2000))

      clearInterval(progressInterval)
      setCheckingProgress(100)

      // 模拟检测结果
      const mockResult: ConflictCheckResult = {
        status: Math.random() > 0.7 ? 'warning' : 'success',
        hasConflict: Math.random() > 0.7,
        conflicts: [],
        checkedAt: new Date().toISOString(),
        stats: {
          total: 0,
          direct: 0,
          indirect: 0,
          potential: 0,
        },
      }

      setCurrentResult(mockResult)
      setStatus(mockResult.status)
      onStatusChange?.(mockResult.status)
      onCheckComplete?.(mockResult)
    } catch (error) {
      const errorResult: ConflictCheckResult = {
        status: 'error',
        hasConflict: false,
        conflicts: [],
        error: error instanceof Error ? error.message : '检测失败',
        stats: {
          total: 0,
          direct: 0,
          indirect: 0,
          potential: 0,
        },
      }

      setCurrentResult(errorResult)
      setStatus('error')
      onStatusChange?.('error')
      onCheckComplete?.(errorResult)
    }
  }, [status, onStatusChange, onCheckComplete])

  /**
   * 自动检测
   */
  useEffect(() => {
    if (!autoCheck || !checkParams.clientName) {
      return
    }

    if (checkTimeoutRef.current) {
      clearTimeout(checkTimeoutRef.current)
    }

    checkTimeoutRef.current = setTimeout(() => {
      performCheck()
    }, debounceDelay)

    return () => {
      if (checkTimeoutRef.current) {
        clearTimeout(checkTimeoutRef.current)
      }
    }
  }, [checkParams, autoCheck, debounceDelay, performCheck])

  /**
   * 处理重新检测
   */
  const handleRecheck = useCallback(() => {
    performCheck()
    onRecheck?.()
  }, [performCheck, onRecheck])

  /**
   * 处理查看详情
   */
  const handleViewDetails = useCallback(
    (conflict: ConflictCase) => {
      onViewDetails?.(conflict)
    },
    [onViewDetails],
  )

  /**
   * 处理标记为已解决
   */
  const handleMarkResolved = useCallback(
    (conflictIds: string | string[]) => {
      const ids = Array.isArray(conflictIds) ? conflictIds : [conflictIds]
      onMarkResolved?.(ids)
    },
    [onMarkResolved],
  )

  /**
   * 渲染主要内容
   */
  const renderContent = () => {
    if (status === 'idle' && !currentResult) {
      return (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description='暂无检测结果'
          style={{ padding: isCompact ? 16 : 24 }}
        >
          <Button
            type='primary'
            icon={<SearchOutlined />}
            onClick={handleRecheck}
            size={isCompact ? 'small' : 'middle'}
          >
            开始检测
          </Button>
        </Empty>
      )
    }

    if (status === 'checking') {
      return (
        <div style={{ padding: isCompact ? 16 : 24, textAlign: 'center' }}>
          <CheckStatusIndicator
            status={status}
            progress={checkingProgress}
            showText
            isCompact={isCompact}
          />
          {checkingProgress > 0 && (
            <Progress
              percent={checkingProgress}
              size={isCompact ? 'small' : 'default'}
              style={{ marginTop: 12 }}
            />
          )}
        </div>
      )
    }

    if (status === 'error' && currentResult?.error) {
      return (
        <Alert
          message='检测失败'
          description={currentResult.error}
          type='error'
          showIcon
          action={
            <Button size='small' type='primary' onClick={handleRecheck}>
              重试
            </Button>
          }
          style={{ margin: isCompact ? 8 : 16 }}
        />
      )
    }

    if (!currentResult || !currentResult.hasConflict) {
      return (
        <Alert
          message='冲突检测完成'
          description='未发现冲突，可以继续创建案件'
          type='success'
          showIcon
          style={{ margin: isCompact ? 8 : 16 }}
        />
      )
    }

    const { conflicts, stats } = currentResult
    const displayConflicts = conflicts.slice(0, finalDisplayConfig.maxDisplayCount)

    return (
      <div className='conflict-check-content'>
        {/* 统计信息 */}
        {finalDisplayConfig.showStats && (
          <ConflictStatsCard
            stats={stats}
            total={stats.total}
            isCompact={isCompact}
            style={{ marginBottom: 16 }}
          />
        )}

        {/* 冲突列表 */}
        <div className='conflict-list'>
          <Collapse
            ghost
            size={isCompact ? 'small' : 'middle'}
            items={
              finalDisplayConfig.displayMode === 'card'
                ? [
                    {
                      key: 'conflicts',
                      label: (
                        <Space>
                          <Text strong>冲突项目 ({displayConflicts.length})</Text>
                          {conflicts.length > displayConflicts.length && (
                            <Text type='secondary'>
                              (显示前{displayConflicts.length}项，共{conflicts.length}项)
                            </Text>
                          )}
                        </Space>
                      ),
                      children: (
                        <List
                          size={isCompact ? 'small' : 'default'}
                          dataSource={displayConflicts}
                          renderItem={(conflict, index) => (
                            <ConflictListItem
                              key={conflict.id}
                              conflict={conflict}
                              index={index}
                              showDetails={finalDisplayConfig.showDetails}
                              isCompact={isCompact}
                              onViewDetails={handleViewDetails}
                              onMarkResolved={(conflictId) => handleMarkResolved(conflictId)}
                            />
                          )}
                        />
                      ),
                    },
                  ]
                : undefined
            }
          />
        </div>
      </div>
    )
  }

  const containerClassName = [
    'conflict-check-inline',
    isCompact ? 'compact' : '',
    status,
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div className={containerClassName} style={style}>
      <Card
        size={isCompact ? 'small' : 'default'}
        className='conflict-check-card'
        title={
          <div className='conflict-check-header'>
            <Space>
              <CheckStatusIndicator status={status} showText={!isCompact} isCompact={isCompact} />
              {finalDisplayConfig.showActions && (
                <QuickActions
                  status={status}
                  hasConflict={currentResult?.hasConflict || false}
                  config={finalActionConfig}
                  isCompact={isCompact}
                  onRecheck={handleRecheck}
                  onViewAllConflicts={() => onViewDetails?.(currentResult?.conflicts[0]!)}
                  onMarkAllResolved={() =>
                    handleMarkResolved(currentResult?.conflicts.map((c) => c.id) || [])
                  }
                />
              )}
            </Space>
          </div>
        }
        extra={
          !isCompact &&
          currentResult?.checkedAt && (
            <Text type='secondary' style={{ fontSize: 12 }}>
              检测时间: {new Date(currentResult.checkedAt).toLocaleString()}
            </Text>
          )
        }
      >
        {renderContent()}
      </Card>
    </div>
  )
}

export default ConflictCheckInline
