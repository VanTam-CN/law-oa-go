import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Row,
  Col,
  Statistic,
  Select,
  Input,
  message,
  Modal,
  Drawer,
  Form,
  Radio,
  Alert,
  Popconfirm,
} from 'antd'
import {
  BellOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SendOutlined,
  EyeOutlined,
  SearchOutlined,
  ReloadOutlined,
  FilterOutlined,
  DeleteOutlined,
  StopOutlined,
  ExclamationCircleOutlined,
  MailOutlined,
  PhoneOutlined,
  WechatOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  notificationQueueService,
  notificationTemplateService,
  notificationStatusMap,
  notificationPriorityMap,
  notificationChannelMap,
  type NotificationQueue,
  type NotificationQueueStats,
} from '@/services/notification'
import './NotificationQueue.less'

const { Option } = Select
const { TextArea } = Input

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
    key: 'recipient',
    width: 150,
    render: (_, record) => (
      <div>
        <div style={{ fontWeight: 600 }}>{record.recipient_name}</div>
        <Tag color='blue' style={{ marginTop: 4 }}>
          {record.recipient_type === 'client' ? '客户' : record.recipient_type === 'lawyer' ? '律师' : '管理员'}
        </Tag>
      </div>
    ),
  },
  {
    title: '渠道',
    dataIndex: 'channel',
    key: 'channel',
    width: 80,
    render: (channel: keyof typeof notificationChannelMap) => (
      <span style={{ fontSize: 16 }}>
        {notificationChannelMap[channel]?.icon} {notificationChannelMap[channel]?.text}
      </span>
    ),
  },
  {
    title: '标题/内容',
    key: 'content',
    width: 250,
    render: (_, record) => (
      <div>
        {record.subject && (
          <div style={{ fontWeight: 600, marginBottom: 4 }}>{record.subject}</div>
        )}
        <div style={{ fontSize: 12, color: '#666' }}>
          {record.content.length > 50
            ? record.content.substring(0, 50) + '...'
            : record.content}
        </div>
      </div>
    ),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    render: (status: keyof typeof notificationStatusMap) => {
      const config = notificationStatusMap[status] || { text: status, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '优先级',
    dataIndex: 'priority',
    key: 'priority',
    width: 80,
    render: (priority: keyof typeof notificationPriorityMap) => {
      const config = notificationPriorityMap[priority] || { text: priority, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '敏感信息',
    dataIndex: 'contains_sensitive_info',
    key: 'contains_sensitive_info',
    width: 100,
    render: (contains: boolean) => (
      contains ? (
        <Tag color='orange' icon={<ExclamationCircleOutlined />}>
          包含
        </Tag>
      ) : (
        <Tag color='green'>无</Tag>
      )
    ),
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 110,
    render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
  },
  {
    title: '发送时间',
    dataIndex: 'sent_at',
    key: 'sent_at',
    width: 110,
    render: (date: string | null | undefined) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-',
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
    fixed: 'right',
    render: (_, record) => (
      <Space size='small'>
        <Button
          type='link'
          size='small'
          icon={<EyeOutlined />}
          onClick={() => handleView(record)}
        >
          详情
        </Button>
        {record.status === 'pending' && (
          <>
            <Button
              type='link'
              size='small'
              icon={<CheckCircleOutlined />}
              onClick={() => handleApprove(record)}
            >
              审批
            </Button>
            <Button
              type='link'
              size='small'
              danger
              icon={<StopOutlined />}
              onClick={() => handleReject(record)}
            >
              拒绝
            </Button>
          </>
        )}
        {record.status === 'approved' && (
          <Button
            type='link'
            size='small'
            icon={<SendOutlined />}
            onClick={() => handleSend(record)}
          >
            发送
          </Button>
        )}
      </Space>
    ),
  },
]

let handleView: (record: NotificationQueue) => void
let handleApprove: (record: NotificationQueue) => void
let handleReject: (record: NotificationQueue) => void
let handleSend: (record: NotificationQueue) => void

const NotificationQueue: React.FC = () => {
  const [notifications, setNotifications] = useState<NotificationQueue[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [stats, setStats] = useState<NotificationQueueStats | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState<boolean>(false)
  const [rejectModalVisible, setRejectModalVisible] = useState<boolean>(false)
  const [selectedNotification, setSelectedNotification] = useState<NotificationQueue | null>(null)
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  const [form] = Form.useForm()

  // 查询参数
  const [queryParams, setQueryParams] = useState({
    page: 1,
    page_size: 10,
    status: '',
    priority: '',
    channel: '',
    recipient_type: '',
    search: '',
  })

  // 搜索表单状态
  const [searchForm, setSearchForm] = useState({
    status: '',
    priority: '',
    channel: '',
    recipient_type: '',
    search: '',
  })

  const [total, setTotal] = useState<number>(0)

  // 获取通知队列列表
  const fetchNotifications = async () => {
    setLoading(true)
    try {
      const res = await notificationQueueService.getQueue(queryParams)
      setNotifications(res.data || [])
      setTotal(res.pagination?.total || 0)

      // 同时获取统计
      const statsRes = await notificationQueueService.getQueueStats()
      setStats(statsRes)
    } catch (error) {
      message.error('获取通知列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchNotifications()
  }, [queryParams])

  // 查看详情
  handleView = (record: NotificationQueue) => {
    setSelectedNotification(record)
    setDetailDrawerVisible(true)
  }

  // 审批通过
  handleApprove = async (record: NotificationQueue) => {
    try {
      await notificationQueueService.approveNotification(record.id)
      message.success('审批通过')
      fetchNotifications()
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 打开拒绝弹窗
  handleReject = (record: NotificationQueue) => {
    setSelectedNotification(record)
    form.resetFields()
    setRejectModalVisible(true)
  }

  // 确认拒绝
  const handleRejectConfirm = async () => {
    try {
      const values = await form.validateFields()
      if (selectedNotification) {
        await notificationQueueService.rejectNotification(selectedNotification.id, values.reason)
        message.success('已拒绝该通知')
        setRejectModalVisible(false)
        fetchNotifications()
      }
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 发送通知
  handleSend = async (record: NotificationQueue) => {
    try {
      await notificationQueueService.sendNotification(record.id)
      message.success('发送成功')
      fetchNotifications()
    } catch (error) {
      message.error('发送失败')
    }
  }

  // 批量确认
  const handleBatchConfirm = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要操作的通知')
      return
    }
    try {
      await notificationQueueService.batchConfirm(selectedRowKeys as number[])
      message.success(`已确认 ${selectedRowKeys.length} 条通知`)
      setSelectedRowKeys([])
      fetchNotifications()
    } catch (error) {
      message.error('批量操作失败')
    }
  }

  // 批量取消
  const handleBatchCancel = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要操作的通知')
      return
    }
    try {
      await notificationQueueService.batchCancel(selectedRowKeys as number[])
      message.success(`已取消 ${selectedRowKeys.length} 条通知`)
      setSelectedRowKeys([])
      fetchNotifications()
    } catch (error) {
      message.error('批量操作失败')
    }
  }

  // 搜索
  const handleSearch = () => {
    setQueryParams({
      ...queryParams,
      ...searchForm,
      page: 1,
    })
  }

  // 重置搜索
  const handleReset = () => {
    setSearchForm({
      status: '',
      priority: '',
      channel: '',
      recipient_type: '',
      search: '',
    })
    setQueryParams({
      page: 1,
      page_size: 10,
      status: '',
      priority: '',
      channel: '',
      recipient_type: '',
      search: '',
    })
  }

  return (
    <div className='notification-queue'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待处理'
              value={stats?.pending || 0}
              valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
              prefix={<BellOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待审批'
              value={stats?.pending_approval || 0}
              valueStyle={{ color: '#ff4d4f', fontSize: '24px', fontWeight: 700 }}
              prefix={<ExclamationCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='已发送'
              value={stats?.sent || 0}
              valueStyle={{ color: '#52c41a', fontSize: '24px', fontWeight: 700 }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='发送失败'
              value={stats?.failed || 0}
              valueStyle={{ color: '#ff4d4f', fontSize: '24px', fontWeight: 700 }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 敏感信息提示 */}
      {stats && stats.pending_approval > 0 && (
        <Alert
          message={`有 ${stats.pending_approval} 条包含敏感信息的通知需要审批`}
          type='warning'
          showIcon
          closable
          style={{ marginBottom: 16 }}
        />
      )}

      {/* 搜索栏 */}
      <Card className='search-card'>
        <Space size='middle' wrap style={{ width: '100%' }}>
          <Select
            placeholder='筛选状态'
            style={{ width: 120 }}
            value={searchForm.status || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, status: value || '' })}
            allowClear
          >
            <Option value='pending'>待处理</Option>
            <Option value='approved'>已审批</Option>
            <Option value='sent'>已发送</Option>
            <Option value='cancelled'>已取消</Option>
            <Option value='failed'>发送失败</Option>
          </Select>

          <Select
            placeholder='优先级'
            style={{ width: 100 }}
            value={searchForm.priority || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, priority: value || '' })}
            allowClear
          >
            <Option value='urgent'>紧急</Option>
            <Option value='normal'>普通</Option>
            <Option value='low'>低</Option>
          </Select>

          <Select
            placeholder='渠道'
            style={{ width: 100 }}
            value={searchForm.channel || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, channel: value || '' })}
            allowClear
          >
            <Option value='email'>邮件</Option>
            <Option value='sms'>短信</Option>
            <Option value='wechat'>微信</Option>
          </Select>

          <Input
            placeholder='搜索标题或内容'
            style={{ width: 200 }}
            value={searchForm.search}
            onChange={(e) => setSearchForm({ ...searchForm, search: e.target.value })}
            prefix={<SearchOutlined />}
          />

          <Button type='primary' icon={<SearchOutlined />} onClick={handleSearch}>
            搜索
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>
        </Space>
      </Card>

      {/* 批量操作 */}
      {selectedRowKeys.length > 0 && (
        <Card style={{ marginBottom: 12 }}>
          <Space>
            <span>已选择 {selectedRowKeys.length} 条</span>
            <Button type='primary' size='small' icon={<CheckCircleOutlined />} onClick={handleBatchConfirm}>
              批量确认
            </Button>
            <Button danger size='small' icon={<StopOutlined />} onClick={handleBatchCancel}>
              批量取消
            </Button>
          </Space>
        </Card>
      )}

      {/* 通知列表 */}
      <Card title='通知队列' className='table-card'>
        <Table
          rowSelection={{
            selectedRowKeys,
            onChange: setSelectedRowKeys,
            getCheckboxProps: (record) => ({
              disabled: record.status === 'sent' || record.status === 'cancelled',
            }),
          }}
          columns={columns}
          dataSource={notifications}
          rowKey='id'
          loading={loading}
          scroll={{ x: 1400 }}
          pagination={{
            current: queryParams.page,
            pageSize: queryParams.page_size,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, size) => {
              setQueryParams({
                ...queryParams,
                page,
                page_size: size || 10,
              })
            },
          }}
        />
      </Card>

      {/* 详情抽屉 */}
      <Drawer
        title='通知详情'
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        width={600}
      >
        {selectedNotification && (
          <div className='notification-detail'>
            <Card title='基本信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>通知ID：</span>
                  {selectedNotification.id}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>状态：</span>
                  {(() => {
                    const config = notificationStatusMap[selectedNotification.status] || {
                      text: selectedNotification.status,
                      color: 'default',
                    }
                    return <Tag color={config.color}>{config.text}</Tag>
                  })()}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>渠道：</span>
                  {notificationChannelMap[selectedNotification.channel]?.text}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>优先级：</span>
                  {notificationPriorityMap[selectedNotification.priority]?.text}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>创建时间：</span>
                  {dayjs(selectedNotification.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>发送时间：</span>
                  {selectedNotification.sent_at
                    ? dayjs(selectedNotification.sent_at).format('YYYY-MM-DD HH:mm:ss')
                    : '-'}
                </Col>
              </Row>
            </Card>

            <Card title='接收人信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>姓名：</span>
                  {selectedNotification.recipient_name}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>类型：</span>
                  {selectedNotification.recipient_type === 'client'
                    ? '客户'
                    : selectedNotification.recipient_type === 'lawyer'
                    ? '律师'
                    : '管理员'}
                </Col>
                {selectedNotification.recipient_contact && (
                  <Col span={24}>
                    <span className='detail-label'>联系方式：</span>
                    {selectedNotification.recipient_contact}
                  </Col>
                )}
              </Row>
            </Card>

            <Card title='通知内容' size='small' style={{ marginBottom: 16 }}>
              {selectedNotification.subject && (
                <div style={{ marginBottom: 12 }}>
                  <span className='detail-label'>标题：</span>
                  <div style={{ marginTop: 4, fontWeight: 600 }}>{selectedNotification.subject}</div>
                </div>
              )}
              <div>
                <span className='detail-label'>内容：</span>
                <div
                  style={{
                    marginTop: 4,
                    padding: 12,
                    background: '#f5f5f5',
                    borderRadius: 4,
                    whiteSpace: 'pre-wrap',
                  }}
                >
                  {selectedNotification.content}
                </div>
              </div>
            </Card>

            {selectedNotification.contains_sensitive_info && (
              <Alert
                message='此通知包含敏感信息，需要审批后才能发送'
                type='warning'
                showIcon
              />
            )}

            {selectedNotification.error_message && (
              <Alert
                message='发送错误'
                description={selectedNotification.error_message}
                type='error'
                showIcon
              />
            )}
          </div>
        )}
      </Drawer>

      {/* 拒绝弹窗 */}
      <Modal
        title='拒绝通知'
        open={rejectModalVisible}
        onOk={handleRejectConfirm}
        onCancel={() => setRejectModalVisible(false)}
        okText='确认拒绝'
        cancelText='取消'
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            label='拒绝原因'
            name='reason'
            rules={[{ required: true, message: '请输入拒绝原因' }]}
          >
            <TextArea rows={4} placeholder='请输入拒绝原因...' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default NotificationQueue
