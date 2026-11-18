import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Row,
  Col,
  Space,
  Tag,
  Badge,
  Tooltip,
  Typography,
  Alert,
  Descriptions,
  Divider,
  message,
  Popconfirm,
  Switch,
  Drawer,
  Radio,
  Checkbox,
  List,
  Avatar,
  Statistic,
  Timeline,
} from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  BellOutlined,
  MailOutlined,
  SmsOutlined,
  WechatOutlined,
  WarningOutlined,
  SendOutlined,
  FilterOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  SafetyOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const { TextArea } = Input
const { Option } = Select
const { Title, Text, Paragraph } = Typography

// 通知状态配置
const notificationStatusConfig: Record<
  string,
  { label: string; color: string; icon: React.ReactNode; description: string }
> = {
  pending: {
    label: '待审批',
    color: 'orange',
    icon: <ClockCircleOutlined />,
    description: '等待审批后发送',
  },
  approved: {
    label: '已批准',
    color: 'blue',
    icon: <CheckCircleOutlined />,
    description: '已批准，待发送',
  },
  sent: {
    label: '已发送',
    color: 'green',
    icon: <CheckCircleOutlined />,
    description: '通知已成功发送',
  },
  failed: {
    label: '发送失败',
    color: 'red',
    icon: <CloseCircleOutlined />,
    description: '发送失败，可重试',
  },
  cancelled: {
    label: '已取消',
    color: 'default',
    icon: <CloseCircleOutlined />,
    description: '通知已取消',
  },
}

// 通知渠道配置
const channelConfig: Record<
  string,
  { label: string; icon: React.ReactNode; color: string }
> = {
  email: { label: '邮件', icon: <MailOutlined />, color: '#1890ff' },
  sms: { label: '短信', icon: <SmsOutlined />, color: '#52c41a' },
  wechat: { label: '微信', icon: <WechatOutlined />, color: '#52c41a' },
}

// 优先级配置
const priorityConfig: Record<
  string,
  { label: string; color: string; level: number }
> = {
  urgent: { label: '紧急', color: 'red', level: 3 },
  normal: { label: '普通', color: 'blue', level: 2 },
  low: { label: '低', color: 'default', level: 1 },
}

// 接收人类型配置
const recipientTypeConfig: Record<string, string> = {
  client: '客户',
  lawyer: '律师',
  admin: '管理员',
}

// 通知数据类型
interface NotificationQueue {
  id: number
  trigger_type: string
  trigger_id: number
  case_id?: number
  recipient_type: string
  recipient_id: number
  recipient_name: string
  recipient_contact?: string
  channel: string
  subject?: string
  content: string
  template_id?: string
  status: string
  priority: string
  contains_sensitive_info: boolean
  auto_send: boolean
  created_by: number
  created_at: string
  approved_by?: number
  approved_at?: string
  sent_at?: string
  sent_retry_count: number
  error_message?: string
}

interface NotificationConfirmProps {
  userRole?: string
  userId?: string
}

const NotificationConfirm: React.FC<NotificationConfirmProps> = ({
  userRole = 'admin',
  userId,
}) => {
  const [notifications, setNotifications] = useState<NotificationQueue[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [channelFilter, setChannelFilter] = useState<string>('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [detailVisible, setDetailVisible] = useState(false)
  const [currentNotification, setCurrentNotification] = useState<NotificationQueue | null>(null)
  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    approved: 0,
    sent: 0,
    failed: 0,
    sensitive: 0,
  })

  // 获取通知列表
  const fetchNotifications = useCallback(async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams({
        page: currentPage.toString(),
        page_size: pageSize.toString(),
      })
      if (statusFilter) params.append('status', statusFilter)
      if (channelFilter) params.append('channel', channelFilter)

      // TODO: 替换为实际的API调用
      // const response = await fetch(`/api/notifications?${params}`)
      // const data = await response.json()

      // 模拟数据
      await new Promise((resolve) => setTimeout(resolve, 300))

      setNotifications([])
      setTotal(0)
    } catch (error) {
      message.error('获取通知列表失败')
    } finally {
      setLoading(false)
    }
  }, [currentPage, pageSize, statusFilter, channelFilter])

  // 获取统计数据
  const fetchStats = useCallback(async () => {
    try {
      // TODO: 替换为实际的API调用
      // const response = await fetch('/api/notifications/stats')
      // const data = await response.json()
      // setStats(data)
    } catch (error) {
      console.error('获取统计数据失败', error)
    }
  }, [])

  // 审批通过
  const handleApprove = async (id: number) => {
    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/notifications/${id}/approve`, { method: 'POST' })
      message.success('审批通过')
      fetchNotifications()
      fetchStats()
    } catch (error) {
      message.error('审批失败')
    }
  }

  // 审批拒绝
  const handleReject = async (id: number) => {
    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/notifications/${id}/reject`, { method: 'POST' })
      message.success('已拒绝')
      fetchNotifications()
      fetchStats()
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 批量审批
  const handleBatchApprove = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要审批的通知')
      return
    }

    try {
      // TODO: 替换为实际的API调用
      // await fetch('/api/notifications/batch-approve', {
      //   method: 'POST',
      //   body: JSON.stringify({ ids: selectedRowKeys }),
      // })
      message.success(`成功审批 ${selectedRowKeys.length} 条通知`)
      setSelectedRowKeys([])
      fetchNotifications()
      fetchStats()
    } catch (error) {
      message.error('批量审批失败')
    }
  }

  // 删除通知
  const handleDelete = async (id: number) => {
    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/notifications/${id}`, { method: 'DELETE' })
      message.success('删除成功')
      fetchNotifications()
    } catch (error) {
      message.error('删除失败')
    }
  }

  // 查看详情
  const handleViewDetail = (record: NotificationQueue) => {
    setCurrentNotification(record)
    setDetailVisible(true)
  }

  // 重试发送
  const handleRetry = async (id: number) => {
    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/notifications/${id}/send`, { method: 'POST' })
      message.success('发送成功')
      fetchNotifications()
    } catch (error) {
      message.error('发送失败')
    }
  }

  useEffect(() => {
    fetchNotifications()
    fetchStats()
  }, [fetchNotifications, fetchStats])

  // 表格列定义
  const columns: ColumnsType<NotificationQueue> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '接收人',
      dataIndex: 'recipient_name',
      key: 'recipient_name',
      width: 120,
      render: (text: string, record: NotificationQueue) => (
        <Space direction="vertical" size={0}>
          <Text strong>{text}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {recipientTypeConfig[record.recipient_type] || record.recipient_type}
          </Text>
        </Space>
      ),
    },
    {
      title: '渠道',
      dataIndex: 'channel',
      key: 'channel',
      width: 80,
      render: (channel: string) => {
        const config = channelConfig[channel] || { label: channel, icon: null, color: 'default' }
        return (
          <Tag icon={config.icon} color={config.color}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '主题',
      dataIndex: 'subject',
      key: 'subject',
      ellipsis: true,
      render: (subject?: string) => subject || '-',
    },
    {
      title: '敏感信息',
      dataIndex: 'contains_sensitive_info',
      key: 'contains_sensitive_info',
      width: 100,
      render: (contains: boolean) => {
        if (contains) {
          return (
            <Tooltip title="包含敏感信息">
              <Tag icon={<SafetyOutlined />} color="warning">
                敏感
              </Tag>
            </Tooltip>
          )
        }
        return <Tag color="success">安全</Tag>
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      render: (priority: string) => {
        const config = priorityConfig[priority] || { label: priority, color: 'default', level: 0 }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const config = notificationStatusConfig[status] || {
          label: status,
          color: 'default',
          icon: null,
          description: '',
        }
        return (
          <Tooltip title={config.description}>
            <Tag icon={config.icon} color={config.color}>
              {config.label}
            </Tag>
          </Tooltip>
        )
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right' as const,
      render: (_: unknown, record: NotificationQueue) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>

          {record.status === 'pending' && (
            <>
              <Popconfirm
                title="确认批准该通知？"
                onConfirm={() => handleApprove(record.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button type="link" size="small" icon={<CheckCircleOutlined />} />
              </Popconfirm>
              <Popconfirm
                title="确认拒绝该通知？"
                onConfirm={() => handleReject(record.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button type="link" size="small" danger icon={<CloseCircleOutlined />} />
              </Popconfirm>
            </>
          )}

          {record.status === 'failed' && (
            <Tooltip title="重试发送">
              <Button
                type="link"
                size="small"
                icon={<SendOutlined />}
                onClick={() => handleRetry(record.id)}
              />
            </Tooltip>
          )}

          {(record.status === 'pending' || record.status === 'cancelled') && (
            <Popconfirm
              title="确认删除该通知？"
              onConfirm={() => handleDelete(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => {
      setSelectedRowKeys(keys)
    },
    getCheckboxProps: (record: NotificationQueue) => ({
      disabled: record.status !== 'pending',
    }),
  }

  return (
    <div className="notification-confirm">
      <Card>
        {/* 统计信息 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="总通知"
                value={stats.total}
                prefix={<BellOutlined />}
                valueStyle={{ fontSize: 20 }}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="待审批"
                value={stats.pending}
                valueStyle={{ color: '#faad14' }}
                prefix={<ClockCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="已批准"
                value={stats.approved}
                valueStyle={{ color: '#1890ff' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="发送失败"
                value={stats.failed}
                valueStyle={{ color: '#ff4d4f' }}
                prefix={<CloseCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="含敏感信息"
                value={stats.sensitive}
                valueStyle={{ color: '#faad14' }}
                prefix={<SafetyOutlined />}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="已发送"
                value={stats.sent}
                valueStyle={{ color: '#52c41a' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
        </Row>

        {/* 筛选栏 */}
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            placeholder="筛选状态"
            allowClear
            style={{ width: 120 }}
            value={statusFilter || undefined}
            onChange={(value) => {
              setStatusFilter(value || '')
              setCurrentPage(1)
            }}
          >
            <Option value="">全部状态</Option>
            {Object.entries(notificationStatusConfig).map(([key, config]) => (
              <Option key={key} value={key}>
                {config.icon} {config.label}
              </Option>
            ))}
          </Select>

          <Select
            placeholder="筛选渠道"
            allowClear
            style={{ width: 120 }}
            value={channelFilter || undefined}
            onChange={(value) => {
              setChannelFilter(value || '')
              setCurrentPage(1)
            }}
          >
            <Option value="">全部渠道</Option>
            {Object.entries(channelConfig).map(([key, config]) => (
              <Option key={key} value={key}>
                {config.icon} {config.label}
              </Option>
            ))}
          </Select>

          {selectedRowKeys.length > 0 && (
            <Button
              type="primary"
              icon={<CheckCircleOutlined />}
              onClick={handleBatchApprove}
            >
              批量审批 ({selectedRowKeys.length})
            </Button>
          )}
        </Space>

        {/* 敏感信息警告 */}
        {stats.sensitive > 0 && (
          <Alert
            message={`发现 ${stats.sensitive} 条包含敏感信息的通知，请仔细审核后再发送`}
            type="warning"
            showIcon
            closable
            style={{ marginBottom: 16 }}
          />
        )}

        {/* 通知列表 */}
        <Table
          rowKey="id"
          columns={columns}
          dataSource={notifications}
          loading={loading}
          rowSelection={rowSelection}
          scroll={{ x: 1200 }}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size || 10)
            },
          }}
        />
      </Card>

      {/* 详情抽屉 */}
      <Drawer
        title="通知详情"
        placement="right"
        width={600}
        open={detailVisible}
        onClose={() => setDetailVisible(false)}
      >
        {currentNotification && (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            {/* 基本信息 */}
            <Card title="基本信息" size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="ID">{currentNotification.id}</Descriptions.Item>
                <Descriptions.Item label="触发类型">{currentNotification.trigger_type}</Descriptions.Item>
                <Descriptions.Item label="触发ID">{currentNotification.trigger_id}</Descriptions.Item>
                <Descriptions.Item label="接收人">{currentNotification.recipient_name}</Descriptions.Item>
                <Descriptions.Item label="渠道">
                  {channelConfig[currentNotification.channel]?.label || currentNotification.channel}
                </Descriptions.Item>
                <Descriptions.Item label="状态">
                  {notificationStatusConfig[currentNotification.status]?.label || currentNotification.status}
                </Descriptions.Item>
                <Descriptions.Item label="优先级">
                  {priorityConfig[currentNotification.priority]?.label || currentNotification.priority}
                </Descriptions.Item>
                <Descriptions.Item label="敏感信息">
                  {currentNotification.contains_sensitive_info ? (
                    <Tag color="warning" icon={<SafetyOutlined />}>
                      包含敏感信息
                    </Tag>
                  ) : (
                    <Tag color="success">安全</Tag>
                  )}
                </Descriptions.Item>
                <Descriptions.Item label="自动发送">
                  {currentNotification.auto_send ? '是' : '否'}
                </Descriptions.Item>
                <Descriptions.Item label="创建时间">
                  {dayjs(currentNotification.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {/* 通知内容 */}
            <Card title="通知内容" size="small">
              {currentNotification.subject && (
                <div style={{ marginBottom: 12 }}>
                  <Text strong>主题：</Text>
                  <Paragraph>{currentNotification.subject}</Paragraph>
                </div>
              )}
              <div style={{ marginBottom: 12 }}>
                <Text strong>内容：</Text>
                <Paragraph
                  copyable
                  style={{
                    background: currentNotification.contains_sensitive_info ? '#fff7e6' : '#f5f5f5',
                    padding: 12,
                    borderRadius: 4,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-all',
                  }}
                >
                  {currentNotification.content}
                </Paragraph>
              </div>
              {currentNotification.contains_sensitive_info && (
                <Alert
                  message="此通知包含敏感信息，请确保接收人已获得相关授权"
                  type="warning"
                  showIcon
                />
              )}
            </Card>

            {/* 操作记录 */}
            <Card title="操作记录" size="small">
              <Timeline>
                <Timeline.Item color="green">
                  <Text>创建于 {dayjs(currentNotification.created_at).format('YYYY-MM-DD HH:mm:ss')}</Text>
                </Timeline.Item>
                {currentNotification.approved_at && (
                  <Timeline.Item color="blue">
                    <Text>审批通过于 {dayjs(currentNotification.approved_at).format('YYYY-MM-DD HH:mm:ss')}</Text>
                  </Timeline.Item>
                )}
                {currentNotification.sent_at && (
                  <Timeline.Item color="green">
                    <Text>发送于 {dayjs(currentNotification.sent_at).format('YYYY-MM-DD HH:mm:ss')}</Text>
                  </Timeline.Item>
                )}
                {currentNotification.error_message && (
                  <Timeline.Item color="red">
                    <Text type="danger">错误：{currentNotification.error_message}</Text>
                  </Timeline.Item>
                )}
              </Timeline>
            </Card>

            {/* 操作按钮 */}
            {currentNotification.status === 'pending' && (
              <Space>
                <Button
                  type="primary"
                  icon={<CheckCircleOutlined />}
                  onClick={() => {
                    handleApprove(currentNotification.id)
                    setDetailVisible(false)
                  }}
                >
                  审批通过
                </Button>
                <Button
                  danger
                  icon={<CloseCircleOutlined />}
                  onClick={() => {
                    handleReject(currentNotification.id)
                    setDetailVisible(false)
                  }}
                >
                  拒绝
                </Button>
              </Space>
            )}
          </Space>
        )}
      </Drawer>
    </div>
  )
}

export default NotificationConfirm
