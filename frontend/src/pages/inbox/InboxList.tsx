import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Popconfirm,
  Tooltip,
  Row,
  Col,
  Statistic,
  App,
  DatePicker,
  Badge,
  Dropdown,
} from 'antd'
import {
  CheckOutlined,
  ClockCircleOutlined,
  BellOutlined,
  FilterOutlined,
  ReloadOutlined,
  EyeOutlined,
  CalendarOutlined,
  FireOutlined,
  ExclamationCircleOutlined,
  ThunderboltOutlined,
  MoreOutlined,
} from '@ant-design/icons'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import type { MenuProps } from 'antd'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
import { getToken } from '@/utils/storage'
import WorkflowActions, { workflowActionsOf, reminderCompletionMode, type WorkflowAction } from '@/components/workflow/WorkflowActions'
import { apiRequest } from '@/services/collaborationRequest'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const { Option } = Select
const { RangePicker } = DatePicker

const authHeaders = (extra: Record<string, string> = {}) => {
  const token = getToken()
  return {
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...extra,
  }
}

interface InboxItem {
  actions?: WorkflowAction[]
  workflow_action?: WorkflowAction | null
  completion_mode?: string
  id: number
  user_id: number
  source_type: string
  source_id: number
  title: string
  content: string
  priority: string
  due_date: string | null
  due_date_type: string
  is_read: boolean
  read_at: string | null
  is_completed: boolean
  completed_at: string | null
  reminder_sent: boolean
  reminder_count: number
  escalated: boolean
  escalated_at: string | null
  snoozed_until: string | null
  snoozed_count: number
  created_at: string
  updated_at: string
}

interface InboxStats {
  needs_action?: number
  waiting?: number
  total: number
  unread: number
  pending: number
  completed: number
  critical: number
  high: number
  overdue: number
  due_today: number
  due_this_week: number
}

interface ListResponse {
  items: InboxItem[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}

const priorityConfig: Record<string, { color: string; icon: React.ReactNode; label: string }> = {
  critical: { color: 'red', icon: <FireOutlined />, label: '紧急' },
  high: { color: 'orange', icon: <ExclamationCircleOutlined />, label: '高' },
  medium: { color: 'blue', icon: <ClockCircleOutlined />, label: '中等' },
  low: { color: 'default', icon: <ClockCircleOutlined />, label: '低' },
}

const sourceTypeConfig: Record<string, { label: string; color: string }> = {
  case: { label: '案件任务', color: 'cyan' },
  deadline: { label: '期限提醒', color: 'purple' },
  approval: { label: '审批待办', color: 'blue' },
  task: { label: '任务', color: 'green' },
  reminder: { label: '提醒', color: 'orange' },
  escalation: { label: '升级通知', color: 'red' },
}

const InboxList: React.FC = () => {
  const { message: appMessage } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [items, setItems] = useState<InboxItem[]>([])
  const [stats, setStats] = useState<InboxStats | null>(null)
  const [detailError, setDetailError] = useState('')
  const listRequest = React.useRef(0)
  const detailRequest = React.useRef(0)
  const [detailVisible, setDetailVisible] = useState(false)
  const [snoozeVisible, setSnoozeVisible] = useState(false)
  const [selectedItem, setSelectedItem] = useState<InboxItem | null>(null)
  const [snoozeForm] = Form.useForm()

  // 筛选状态
  const [isReadFilter, setIsReadFilter] = useState<boolean | undefined>(undefined)
  const [isCompletedFilter, setIsCompletedFilter] = useState<boolean | undefined>(undefined)
  const [priorityFilter, setPriorityFilter] = useState<string>('')
  const [workflowBucketFilter, setWorkflowBucketFilter] = useState('')
  const [sourceTypeFilter, setSourceTypeFilter] = useState<string>('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)
  const [searchText, setSearchText] = useState('')

  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  })

  // 排序状态
  const [orderBy, setOrderBy] = useState<string>('due_date')

  useEffect(() => {
    fetchItems()
    fetchStats()
  }, [pagination.current, pagination.pageSize, isReadFilter, isCompletedFilter, priorityFilter, sourceTypeFilter, orderBy, workflowBucketFilter])

  const fetchItems = async () => {
    const sequence = ++listRequest.current
    setLoading(true)
    try {
      const params: any = {
        page: pagination.current,
        page_size: pagination.pageSize,
        order_by: orderBy,
      }

      if (isReadFilter !== undefined) params.is_read = isReadFilter
      if (isCompletedFilter !== undefined) params.is_completed = isCompletedFilter
      if (priorityFilter) params.priority = priorityFilter
      if (workflowBucketFilter) params.workflow_bucket = workflowBucketFilter
      if (sourceTypeFilter) params.source_type = sourceTypeFilter
      if (searchText) params.search = searchText
      if (dateRange) {
        params.due_after = dateRange[0].format('YYYY-MM-DD')
        params.due_before = dateRange[1].format('YYYY-MM-DD')
      }

      const query = new URLSearchParams(params).toString()
      const response = await fetch(`/api/v1/inbox?${query}`, {
        headers: authHeaders(),
      })
      const data = await response.json()
      if (sequence !== listRequest.current) return
      if (!response.ok) throw new Error(data.error?.message || data.message || '读取待办失败')

      if (data.success || data.data) {
        const result = data.data || data
        setItems(result.items || [])
        if (result.stats) setStats(result.stats)
        setPagination({
          ...pagination,
          total: result.pagination?.total || 0,
        })
      }
    } catch (error) {
      if (sequence === listRequest.current) {
        setItems([])
        appMessage.error(error instanceof Error ? error.message : '获取待办列表失败')
      }
    } finally {
      if (sequence === listRequest.current) setLoading(false)
    }
  }

  const fetchStats = async () => {
    try {
      const response = await fetch('/api/v1/inbox/stats', {
        headers: authHeaders(),
      })
      const data = await response.json()

      if (data.success || data.data) {
        setStats(data.data || data)
      }
    } catch (error) {
      console.error('获取统计失败:', error)
    }
  }

  const handleMarkAsRead = async (id: number) => {
    try {
      const response = await fetch(`/api/v1/inbox/${id}/read`, {
        method: 'PUT',
        headers: authHeaders(),
      })

      if (response.ok) {
        appMessage.success('已标记为已读')
        fetchItems()
        fetchStats()
      }
    } catch (error) {
      appMessage.error('操作失败')
    }
  }

  const handleMarkAsCompleted = async (id: number) => {
    const item = selectedItem?.id === id ? selectedItem : items.find((row) => row.id === id)
    if (!item?.completion_mode || reminderCompletionMode(item) === 'workflow') {
      appMessage.warning('请进入业务页面办理，此事项不能手工完成'); return
    }
    try {
      const operation = reminderCompletionMode(item) === 'manual' ? 'complete' : 'archive'
      const response = await fetch(`/api/v1/inbox/${id}/${operation}`, {
        method: 'PUT',
        headers: authHeaders(),
      })

      if (response.ok) {
        appMessage.success(reminderCompletionMode(item) === 'manual' ? '已标记为完成' : '已归档提醒，业务状态未改变')
        fetchItems()
        fetchStats()
      } else {
        const failure = await response.json().catch(() => ({}))
        appMessage.error(failure.error?.message || failure.message || '事项状态已变化，请刷新后办理')
        fetchItems()
      }
    } catch (error) {
      appMessage.error('操作失败')
    }
  }

  const handleSnooze = async (values: any) => {
    if (!selectedItem) return

    try {
      const response = await fetch(`/api/v1/inbox/${selectedItem.id}/snooze`, {
        method: 'PUT',
        headers: authHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify({
          until: values.until ? values.until.format('YYYY-MM-DD HH:mm:ss') : undefined,
          duration: values.duration,
        }),
      })

      if (response.ok) {
        appMessage.success('已延后提醒')
        setSnoozeVisible(false)
        snoozeForm.resetFields()
        fetchItems()
      }
    } catch (error) {
      appMessage.error('操作失败')
    }
  }

  const showDetail = async (item: InboxItem) => {
    const sequence = ++detailRequest.current
    setSelectedItem(null)
    setDetailError('')
    setDetailVisible(true)
    try {
      const detail = await apiRequest<InboxItem>('/inbox/' + item.id)
      if (sequence !== detailRequest.current) return
      setSelectedItem(detail)
      if (!detail.is_read) void handleMarkAsRead(detail.id)
    } catch (error) { if (sequence === detailRequest.current) setDetailError(error instanceof Error ? error.message : '待办详情读取失败') }
  }

  const showSnooze = (item: InboxItem) => {
    setSelectedItem(item)
    setSnoozeVisible(true)
  }

  const getDueDateStatus = (item: InboxItem) => {
    if (!item.due_date || item.is_completed) return null

    const now = dayjs()
    const dueDate = dayjs(item.due_date)
    const diff = dueDate.diff(now, 'day')

    if (diff < 0) {
      return <Tag color="red">已超期 {Math.abs(diff)} 天</Tag>
    } else if (diff === 0) {
      return <Tag color="orange">今天到期</Tag>
    } else if (diff === 1) {
      return <Tag color="gold">明天到期</Tag>
    } else if (diff <= 7) {
      return <Tag color="blue">剩余 {diff} 天</Tag>
    } else {
      return <Tag color="default">剩余 {diff} 天</Tag>
    }
  }

  const columns: ColumnsType<InboxItem> = [
    {
      title: '状态',
      dataIndex: 'is_read',
      width: 60,
      render: (isRead: boolean, record: InboxItem) => (
        <Space size={4}>
          {!isRead && <Badge status="error" />}
          {record.escalated && <Tooltip title="已升级"><ExclamationCircleOutlined style={{ color: 'red' }} /></Tooltip>}
          {record.snoozed_until && dayjs(record.snoozed_until).isAfter(dayjs()) && (
            <Tooltip title={`延后至 ${dayjs(record.snoozed_until).format('YYYY-MM-DD HH:mm')}`}>
              <ClockCircleOutlined style={{ color: 'orange' }} />
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 80,
      render: (priority: string) => {
        const config = priorityConfig[priority] || priorityConfig.medium
        return (
          <Tag color={config.color} icon={config.icon}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '标题',
      dataIndex: 'title',
      ellipsis: true,
      render: (title: string, record: InboxItem) => (
        <Space direction="vertical" size={0}>
          <span style={{ fontWeight: !record.is_read ? 600 : 400 }}>
            {title}
          </span>
          {record.content && (
            <span style={{ fontSize: 12, color: '#999' }}>
              {record.content}
            </span>
          )}
        </Space>
      ),
    },
    {
      title: '来源',
      dataIndex: 'source_type',
      width: 100,
      render: (sourceType: string) => {
        const config = sourceTypeConfig[sourceType] || { label: sourceType, color: 'default' }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: '到期时间',
      dataIndex: 'due_date',
      width: 140,
      render: (dueDate: string | null, record: InboxItem) => {
        if (!dueDate) return '-'
        return (
          <Space direction="vertical" size={0}>
            <span>{dayjs(dueDate).format('YYYY-MM-DD HH:mm')}</span>
            {getDueDateStatus(record)}
          </Space>
        )
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 100,
      render: (createdAt: string) => (
        <Tooltip title={dayjs(createdAt).format('YYYY-MM-DD HH:mm:ss')}>
          <span>{dayjs(createdAt).fromNow()}</span>
        </Tooltip>
      ),
    },
    {
      title: '操作',
      width: 120,
      render: (_, record: InboxItem) => {
        const canComplete = !record.is_completed && Boolean(record.completion_mode) && reminderCompletionMode(record) !== 'workflow'
        const items: MenuProps['items'] = [
          {
            key: 'detail',
            label: '查看详情',
            icon: <EyeOutlined />,
            onClick: () => showDetail(record),
          },
          {
            key: 'snooze',
            label: '延后提醒',
            icon: <ClockCircleOutlined />,
            onClick: () => showSnooze(record),
            disabled: record.is_completed,
          },
          {
            type: 'divider',
          },
          {
            key: 'complete',
            label: reminderCompletionMode(record) === 'manual' ? '标记完成' : '归档提醒',
            icon: <CheckOutlined />,
            onClick: () => handleMarkAsCompleted(record.id),
            disabled: !canComplete,
          },
        ]

        return (
          <Space wrap>
            <WorkflowActions actions={workflowActionsOf(record)} />
            {canComplete && (
              <Tooltip title={reminderCompletionMode(record) === 'manual' ? '标记完成' : '归档提醒'}>
                <Button
                  type="text"
                  icon={<CheckOutlined />}
                  aria-label={`${reminderCompletionMode(record) === 'manual' ? '标记完成' : '归档提醒'}：${record.title}`}
                  title="标记完成"
                  onClick={() => handleMarkAsCompleted(record.id)}
                />
              </Tooltip>
            )}
            <Dropdown menu={{ items }} trigger={['click']}>
              <Button type="text" icon={<MoreOutlined />} aria-label={`更多操作：${record.title}`} title="更多操作" />
            </Dropdown>
          </Space>
        )
      },
    },
  ]

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination({
      ...pagination,
      current: newPagination.current || 1,
      pageSize: newPagination.pageSize || 20,
    })
  }

  const handleFilterReset = () => {
    setIsReadFilter(undefined)
    setIsCompletedFilter(undefined)
    setPriorityFilter('')
    setSourceTypeFilter('')
    setWorkflowBucketFilter('')
    setDateRange(null)
    setSearchText('')
  }

  return (
    <div style={{ padding: 24 }}>
      {/* 统计卡片 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="待办总数"
                value={stats.total}
                prefix={<BellOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="未读"
                value={stats.unread}
                prefix={<Badge count={stats.unread} />}
                valueStyle={{ color: stats.unread > 0 ? '#ff4d4f' : '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="需要我处理"
                value={stats.needs_action ?? stats.pending}
                prefix={<ClockCircleOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已超期"
                value={stats.overdue}
                prefix={<ExclamationCircleOutlined />}
                valueStyle={{ color: stats.overdue > 0 ? '#ff4d4f' : '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>
      )}

      <Card
        title={
          <Space>
            <BellOutlined />
            <span>我的待办</span>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchItems}>
              刷新
            </Button>
          </Space>
        }
      >
        {/* 筛选栏 */}
        <Space wrap style={{ marginBottom: 16 }} size="middle">
          <Input
            placeholder="搜索待办..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onPressEnter={fetchItems}
            style={{ width: 200 }}
            allowClear
          />
          <Select
            aria-label="完成状态"
            placeholder="完成状态"
            value={isCompletedFilter === undefined ? undefined : (isCompletedFilter ? '已处理' : '待处理')}
            onChange={(value) => setIsCompletedFilter(value === undefined ? undefined : value === '已处理')}
            style={{ width: 120 }}
            allowClear
          >
            <Option value="待处理">待处理</Option>
            <Option value="已处理">已处理</Option>
          </Select>
          <Select
            aria-label="优先级"
            placeholder="优先级"
            value={priorityFilter || undefined}
            onChange={setPriorityFilter}
            style={{ width: 120 }}
            allowClear
          >
            <Option value="critical">紧急</Option>
            <Option value="high">高</Option>
            <Option value="medium">中等</Option>
            <Option value="low">低</Option>
          </Select>
          <Select
            aria-label="来源类型"
            placeholder="来源类型"
            value={sourceTypeFilter || undefined}
            onChange={setSourceTypeFilter}
            style={{ width: 120 }}
            allowClear
          >
            <Option value="deadline">期限提醒</Option>
            <Option value="approval">审批待办</Option>
            <Option value="task">任务</Option>
            <Option value="reminder">提醒</Option>
          </Select>
          <RangePicker
            value={dateRange}
            onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
            placeholder={['开始日期', '结束日期']}
          />
          <Select aria-label='办理分类' value={workflowBucketFilter} onChange={(value) => { setWorkflowBucketFilter(value); setPagination((old) => ({ ...old, current: 1 })) }} style={{ width: 150 }} options={[{ value: '', label: '全部事项' }, { value: 'needs_action', label: '需要我处理' }, { value: 'waiting', label: '等待他人' }, { value: 'completed', label: '已处理／结果提醒' }]} />
          <Select
            aria-label="排序方式"
            placeholder="排序方式"
            value={orderBy}
            onChange={setOrderBy}
            style={{ width: 120 }}
          >
            <Option value="due_date">按到期时间</Option>
            <Option value="priority">按优先级</Option>
            <Option value="created_at">按创建时间</Option>
          </Select>
          <Button icon={<FilterOutlined />} onClick={handleFilterReset}>
            重置筛选
          </Button>
        </Space>

        <Table
          columns={columns}
          dataSource={items}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
          onChange={handleTableChange}
          onRow={(record) => ({
            onDoubleClick: () => showDetail(record),
            style: { cursor: 'pointer' },
          })}
          rowClassName={(record) => (!record.is_read ? 'unread-row' : '')}
        />
      </Card>

      {/* 详情弹窗 */}
      <Modal
        title="待办详情"
        open={detailVisible}
        onCancel={() => { ++detailRequest.current; setDetailVisible(false) }}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
          selectedItem && !selectedItem.is_completed && selectedItem.completion_mode && reminderCompletionMode(selectedItem) !== 'workflow' && (
            <Button
              key="complete"
              type="primary"
              icon={<CheckOutlined />}
              onClick={() => {
                if (selectedItem) {
                  handleMarkAsCompleted(selectedItem.id)
                  setDetailVisible(false)
                }
              }}
            >
              {selectedItem && reminderCompletionMode(selectedItem) === 'manual' ? '标记完成' : '归档提醒'}
            </Button>
          ),
        ]}
        width={600}
      >
        {detailError && <p role='alert'>{detailError}</p>}
        {!selectedItem && !detailError && <p>正在读取最新事项</p>}
        {selectedItem && (
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <WorkflowActions actions={workflowActionsOf(selectedItem)} />
            {reminderCompletionMode(selectedItem) === 'workflow' && <p>请进入业务页面办理。延后仅改变提醒时间，不改变办理期限或业务状态。</p>}
            <div>
              <Space>
                <Tag color={priorityConfig[selectedItem.priority]?.color}>
                  {priorityConfig[selectedItem.priority]?.label}
                </Tag>
                <Tag color={sourceTypeConfig[selectedItem.source_type]?.color}>
                  {sourceTypeConfig[selectedItem.source_type]?.label}
                </Tag>
              </Space>
            </div>
            <div>
              <h3>{selectedItem.title}</h3>
            </div>
            {selectedItem.content && (
              <div>
                <p style={{ whiteSpace: 'pre-wrap' }}>{selectedItem.content}</p>
              </div>
            )}
            <Row gutter={16}>
              <Col span={12}>
                <div style={{ color: '#666' }}>创建时间</div>
                <div>{dayjs(selectedItem.created_at).format('YYYY-MM-DD HH:mm:ss')}</div>
              </Col>
              {selectedItem.due_date && (
                <Col span={12}>
                  <div style={{ color: '#666' }}>到期时间</div>
                  <div>
                    {dayjs(selectedItem.due_date).format('YYYY-MM-DD HH:mm:ss')}
                    {getDueDateStatus(selectedItem)}
                  </div>
                </Col>
              )}
            </Row>
            {selectedItem.escalated && (
              <div>
                <Tag color="red" icon={<ExclamationCircleOutlined />}>
                  已升级通知 ({dayjs(selectedItem.escalated_at).fromNow()})
                </Tag>
              </div>
            )}
          </Space>
        )}
      </Modal>

      {/* 延后弹窗 */}
      <Modal
        title="延后提醒"
        open={snoozeVisible}
        onCancel={() => setSnoozeVisible(false)}
        onOk={() => snoozeForm.submit()}
      >
        <Form form={snoozeForm} onFinish={handleSnooze} layout="vertical">
          <Form.Item
            name="duration"
            label="延后天数"
            extra="或者指定具体时间"
          >
            <Select placeholder="选择延后天数" allowClear>
              <Option value={1}>延后 1 天</Option>
              <Option value={3}>延后 3 天</Option>
              <Option value={7}>延后 7 天</Option>
              <Option value={14}>延后 14 天</Option>
              <Option value={30}>延后 30 天</Option>
            </Select>
          </Form.Item>
          <Form.Item name="until" label="或指定具体时间">
            <DatePicker
              showTime
              format="YYYY-MM-DD HH:mm:ss"
              style={{ width: '100%' }}
              placeholder="选择提醒时间"
            />
          </Form.Item>
          {selectedItem && selectedItem.snoozed_count > 0 && (
            <div style={{ color: '#faad14' }}>
              已延后 {selectedItem.snoozed_count} 次
            </div>
          )}
        </Form>
      </Modal>

      <style>{`
        .unread-row {
          background-color: #f0f7ff;
          font-weight: 500;
        }
        .unread-row:hover {
          background-color: #e6f4ff !important;
        }
      `}</style>
    </div>
  )
}

export default InboxList
