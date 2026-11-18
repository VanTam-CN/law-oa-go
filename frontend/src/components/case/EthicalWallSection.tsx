/**
 * EthicalWallSection 组件
 * 隔离墙状态显示组件
 * 用于案件详情页显示和管理隔离墙状态
 */

import React, { useCallback, useState, useMemo } from 'react'
import {
  Card,
  Space,
  Tag,
  Button,
  Alert,
  Typography,
  Switch,
  Tooltip,
  Popconfirm,
  message,
  Spin,
  Divider,
} from 'antd'
import {
  SafetyOutlined,
  SafetyCertificateOutlined,
  UnlockOutlined,
  TeamOutlined,
  InfoCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { EthicalWall, EthicalWallStatus } from '../../types/ethicalWall'
import { useEthicalWall, useEnableEthicalWall, useDisableEthicalWall } from '../../hooks/useEthicalWall'
import WhitelistModal from './WhitelistModal'
import './EthicalWallSection.less'

const { Text, Paragraph } = Typography

interface EthicalWallSectionProps {
  caseId: string
  caseName?: string
  className?: string
  style?: React.CSSProperties
  compact?: boolean
}

/**
 * 隔离墙状态标签组件
 */
const WallStatusTag: React.FC<{
  enabled: boolean
  compact?: boolean
}> = ({ enabled, compact }) => {
  if (enabled) {
    return (
      <Tag
        icon={<SafetyCertificateOutlined />}
        color="success"
        style={{ fontSize: compact ? 12 : 14 }}
      >
        隔离墙已启用
      </Tag>
    )
  }

  return (
    <Tag
      icon={<UnlockOutlined />}
      color="default"
      style={{ fontSize: compact ? 12 : 14 }}
    >
      隔离墙未启用
    </Tag>
  )
}

/**
 * 隔离墙说明组件
 */
const WallDescription: React.FC<{ compact?: boolean }> = ({ compact }) => {
  const description = compact
    ? '启用后，只有白名单用户可访问此案件'
    : '隔离墙用于限制案件访问权限，只有白名单中的用户才能查看和操作该案件。适用于敏感案件或需要保密的场景。'

  return (
    <Paragraph type="secondary" style={{ marginBottom: 0, fontSize: compact ? 12 : 14 }}>
      <InfoCircleOutlined style={{ marginRight: 4 }} />
      {description}
    </Paragraph>
  )
}

/**
 * 主组件
 */
const EthicalWallSection: React.FC<EthicalWallSectionProps> = ({
  caseId,
  caseName,
  className = '',
  style,
  compact = false,
}) => {
  const [whitelistModalVisible, setWhitelistModalVisible] = useState(false)

  // 查询隔离墙状态
  const { data: wallData, isLoading, error } = useEthicalWall(caseId)

  // 启用/禁用隔离墙
  const enableMutation = useEnableEthicalWall()
  const disableMutation = useDisableEthicalWall()

  const isMutating = enableMutation.isPending || disableMutation.isPending
  const isEnabled = wallData?.enabled ?? false

  /**
   * 处理切换隔离墙状态
   */
  const handleToggleWall = useCallback(
    (checked: boolean) => {
      if (checked) {
        // 启用隔离墙
        enableMutation.mutate(caseId, {
          onSuccess: () => {
            message.success('隔离墙已启用')
          },
          onError: (err: Error) => {
            message.error(`启用失败: ${err.message}`)
          },
        })
      } else {
        // 禁用隔离墙
        disableMutation.mutate(caseId, {
          onSuccess: () => {
            message.success('隔离墙已禁用')
          },
          onError: (err: Error) => {
            message.error(`禁用失败: ${err.message}`)
          },
        })
      }
    },
    [caseId, enableMutation, disableMutation],
  )

  /**
   * 打开白名单管理弹窗
   */
  const handleOpenWhitelist = useCallback(() => {
    setWhitelistModalVisible(true)
  }, [])

  /**
   * 关闭白名单管理弹窗
   */
  const handleCloseWhitelist = useCallback(() => {
    setWhitelistModalVisible(false)
  }, [])

  /**
   * 渲染加载状态
   */
  if (isLoading) {
    return (
      <Card
        className={`ethical-wall-section ${className}`}
        style={style}
        size={compact ? 'small' : 'default'}
      >
        <div style={{ textAlign: 'center', padding: compact ? 16 : 24 }}>
          <Spin size="small" />
          <Text type="secondary" style={{ marginLeft: 8 }}>
            加载中...
          </Text>
        </div>
      </Card>
    )
  }

  /**
   * 渲染错误状态
   */
  if (error) {
    return (
      <Card
        className={`ethical-wall-section ${className}`}
        style={style}
        size={compact ? 'small' : 'default'}
      >
        <Alert
          message="加载失败"
          description={error instanceof Error ? error.message : '未知错误'}
          type="error"
          showIcon
        />
      </Card>
    )
  }

  /**
   * 卡片标题
   */
  const cardTitle = (
    <Space>
      <SafetyOutlined />
      <Text strong>隔离墙管理</Text>
      <WallStatusTag enabled={isEnabled} compact={compact} />
    </Space>
  )

  /**
   * 卡片额外操作
   */
  const cardExtra = isEnabled && (
    <Tooltip title="管理白名单用户">
      <Button
        type="text"
        icon={<TeamOutlined />}
        onClick={handleOpenWhitelist}
        size={compact ? 'small' : 'middle'}
      >
        {!compact && '白名单'}
      </Button>
    </Tooltip>
  )

  return (
    <>
      <Card
        className={`ethical-wall-section ${isEnabled ? 'enabled' : 'disabled'} ${className}`}
        style={style}
        size={compact ? 'small' : 'default'}
        title={cardTitle}
        extra={cardExtra}
      >
        <Space direction="vertical" size={compact ? 'small' : 'middle'} style={{ width: '100%' }}>
          {/* 说明 */}
          <WallDescription compact={compact} />

          {!compact && <Divider style={{ margin: '8px 0' }} />}

          {/* 操作区 */}
          <div className="wall-actions">
            <Space>
              <Text type="secondary">隔离墙状态:</Text>
              {isEnabled ? (
                <Tag color="success" icon={<SafetyCertificateOutlined />}>
                  已启用
                </Tag>
              ) : (
                <Tag color="default" icon={<UnlockOutlined />}>
                  未启用
                </Tag>
              )}
            </Space>

            <Space>
              {isEnabled ? (
                <Popconfirm
                  title="确定要禁用隔离墙吗？"
                  description="禁用后，所有用户都可以访问此案件。"
                  onConfirm={() => handleToggleWall(false)}
                  okText="确定"
                  cancelText="取消"
                  okButtonProps={{ danger: true }}
                >
                  <Button
                    danger
                    icon={<CloseCircleOutlined />}
                    loading={isMutating}
                    size={compact ? 'small' : 'middle'}
                  >
                    禁用隔离墙
                  </Button>
                </Popconfirm>
              ) : (
                <Popconfirm
                  title="确定要启用隔离墙吗？"
                  description="启用后，只有白名单用户才能访问此案件。"
                  onConfirm={() => handleToggleWall(true)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button
                    type="primary"
                    icon={<SafetyOutlined />}
                    loading={isMutating}
                    size={compact ? 'small' : 'middle'}
                  >
                    启用隔离墙
                  </Button>
                </Popconfirm>
              )}

              {isEnabled && (
                <Button
                  icon={<TeamOutlined />}
                  onClick={handleOpenWhitelist}
                  size={compact ? 'small' : 'middle'}
                >
                  管理白名单
                </Button>
              )}
            </Space>
          </div>

          {/* 启用时的提示 */}
          {isEnabled && (
            <Alert
              message="隔离墙保护中"
              description="当前案件处于隔离墙保护状态，仅白名单用户可访问。"
              type="info"
              showIcon
              icon={<SafetyCertificateOutlined />}
              style={{ marginTop: 8 }}
            />
          )}
        </Space>
      </Card>

      {/* 白名单管理弹窗 */}
      <WhitelistModal
        open={whitelistModalVisible}
        caseId={caseId}
        caseName={caseName}
        onClose={handleCloseWhitelist}
      />
    </>
  )
}

export default EthicalWallSection
