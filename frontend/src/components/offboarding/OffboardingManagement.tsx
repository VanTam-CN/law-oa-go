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
  Progress,
  Steps,
  List,
  Avatar,
  message,
  Popconfirm,
  Transfer,
  Checkbox,
  InputNumber,
} from 'antd'
import {
  UserOutlined,
  PlusOutlined,
  EyeOutlined,
  EditOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  FileTextOutlined,
  InboxOutlined,
  LogoutOutlined,
  TeamOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { TransferDirection } from 'antd/es/transfer'
import dayjs from 'dayjs'

const { Option } = Select
const { TextArea } = Input
const { Title, Text, Paragraph } = Typography
const { Step } = Steps

// 离职交接状态配置
const offboardingStatusConfig: Record<
  string,
  { label: string; color: string; icon: React.ReactNode; description: string }
> = {
  pending: {
    label: '待处理',
    color: 'default',
    icon: <ClockCircleOutlined />,
    description: '等待开始交接',
  },
  in_progress: {
    label: '进行中',
    color: 'processing',
    icon: <SyncOutlined spin />,
    description: '交接正在进行',
  },
  completed: {
    label: '已完成',
    color: 'success',
    icon: <CheckCircleOutlined />,
    description: '交接已完成',
  },
  cancelled: {
    label: '已取消',
    color: 'default',
    icon: <CloseCircleOutlined />,
    description: '交接已取消',
  },
}

// 文档处理方式配置
const documentMethodConfig: Record<string, string> = {
  delete: '删除',
  transfer: '转移给新律师',
  revoke_access: '撤销访问权限',
  archive: '归档保存',
}

// 离职交接记录数据类型
interface OffboardingRecord {
  id: number
  user_id: number
  user_name?: string
  initiated_by: number
  initiated_by_name?: string
  initiated_at: string
  new_lawyer_id?: number
  new_lawyer_name?: string
  new_assistant_id?: number
  new_assistant_name?: string
  document_disposal_method: string
  document_disposal_completed_at?: string
  status: string
  completed_at?: string
  notes?: string
  created_at: string
  updated_at: string

  // 统计数据
  case_transfer_count?: number
  case_completed_count?: number
  inbox_transfer_count?: number
  inbox_completed_count?: number
}

// 移交详情数据类型
interface TransferDetail {
  id: number
  offboarding_id: number
  transfer_type: string
  original_owner_id: number
  original_owner_name?: string
  new_owner_id?: number
  new_owner_name?: string
  item_id: number
  item_name: string
  transfer_status: string
  transferred_at?: string
}

// 案件数据类型
interface CaseItem {
  id: number
  title: string
  case_number: string
  status: string
}

// 待办事项数据类型
interface InboxItem {
  id: number
  title: string
  priority: string
  due_date?: string
  is_completed: boolean
}

// 离职交接请求
interface OffboardingRequest {
  user_id: number
  initiated_by: number
  new_lawyer_id?: number
  new_assistant_id?: number
  document_disposal_method: string
  notes?: string
}

interface OffboardingManagementProps {
  userRole?: string
  userId?: string
}

const OffboardingManagement: React.FC<OffboardingManagementProps> = ({
  userRole = 'admin',
  userId,
}) => {
  const [offboardings, setOffboardings] = useState<OffboardingRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState<string>('')

  // Modal states
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [transferModalVisible, setTransferModalVisible] = useState(false)
  const [cancelModalVisible, setCancelModalVisible] = useState(false)

  const [currentRecord, setCurrentRecord] = useState<OffboardingRecord | null>(null)
  const [transferDetails, setTransferDetails] = useState<TransferDetail[]>([])
  const [userCases, setUserCases] = useState<CaseItem[]>([])
  const [userInboxes, setUserInboxes] = useState<InboxItem[]>([])

  const [form] = Form.useForm()
  const [transferForm] = Form.useForm()
  const [cancelForm] = Form.useForm()

  // Transfer 穿梭框状态
  const [caseTargetKeys, setCaseTargetKeys] = useState<React.Key[]>([])
  const [inboxTargetKeys, setInboxTargetKeys] = useState<React.Key[]>([])
  const [selectedCases, setSelectedCases] = useState<React.Key[]>([])
  const [selectedInboxes, setSelectedInboxes] = useState<React.Key[]>([])

  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    inProgress: 0,
    completed: 0,
    thisMonth: 0,
  })

  // 获取交接记录列表
  const fetchOffboardings = useCallback(async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams({
        page: currentPage.toString(),
        page_size: pageSize.toString(),
      })
      if (statusFilter) params.append('status', statusFilter)

      // TODO: 替换为实际的API调用
      // const response = await fetch(`/api/offboarding?${params}`)
      // const data = await response.json()

      // 模拟数据
      await new Promise((resolve) => setTimeout(resolve, 300))

      setOffboardings([])
      setTotal(0)
    } catch (error) {
      message.error('获取交接记录失败')
    } finally {
      setLoading(false)
    }
  }, [currentPage, pageSize, statusFilter])

  // 获取统计数据
  const fetchStats = useCallback(async () => {
    try {
      // TODO: 替换为实际的API调用
      // const response = await fetch('/api/offboarding/stats')
      // const data = await response.json()
      // setStats(data)
    } catch (error) {
      console.error('获取统计数据失败', error)
    }
  }, [])

  // 创建交接
  const handleCreate = async (values: OffboardingRequest) => {
    try {
      // TODO: 替换为实际的API调用
      // await fetch('/api/offboarding', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify(values),
      // })

      message.success('交接流程已发起')
      setCreateModalVisible(false)
      form.resetFields()
      fetchOffboardings()
      fetchStats()
    } catch (error) {
      message.error('发起交接失败')
    }
  }

  // 完成交接
  const handleComplete = async (id: number) => {
    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/offboarding/${id}/complete`, {
      //   method: 'POST',
      // })

      message.success('交接已完成')
      fetchOffboardings()
      fetchStats()
      if (detailModalVisible) {
        setDetailModalVisible(false)
      }
    } catch (error) {
      message.error('完成交接失败')
    }
  }

  // 取消交接
  const handleCancel = async (values: { reason: string }) => {
    if (!currentRecord) return

    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/offboarding/${currentRecord.id}/cancel`, {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ reason: values.reason }),
      // })

      message.success('交接已取消')
      setCancelModalVisible(false)
      cancelForm.resetFields()
      fetchOffboardings()
      fetchStats()
      if (detailModalVisible) {
        setDetailModalVisible(false)
      }
    } catch (error) {
      message.error('取消交接失败')
    }
  }

  // 批量移交案件
  const handleBatchTransferCases = async (caseIds: number[]) => {
    if (!currentRecord || !currentRecord.new_lawyer_id) {
      message.warning('请先指定接手律师')
      return
    }

    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/offboarding/${currentRecord.id}/transfer-cases`, {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({
      //     new_lawyer_id: currentRecord.new_lawyer_id,
      //     case_ids: caseIds,
      //   }),
      // })

      message.success(`成功移交 ${caseIds.length} 个案件`)
      setTransferModalVisible(false)
      fetchOffboardings()
      if (detailModalVisible) {
        fetchOffboardingDetails(currentRecord.id)
      }
    } catch (error) {
      message.error('移交案件失败')
    }
  }

  // 批量移交待办
  const handleBatchTransferInboxes = async (inboxIds: number[]) => {
    if (!currentRecord || !currentRecord.new_assistant_id) {
      message.warning('请先指定接手助理')
      return
    }

    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/offboarding/${currentRecord.id}/transfer-inboxes`, {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({
      //     new_assistant_id: currentRecord.new_assistant_id,
      //     inbox_ids: inboxIds,
      //   }),
      // })

      message.success(`成功移交 ${inboxIds.length} 个待办事项`)
      fetchOffboardings()
      if (detailModalVisible) {
        fetchOffboardingDetails(currentRecord.id)
      }
    } catch (error) {
      message.error('移交待办事项失败')
    }
  }

  // 处理文档
  const handleHandleDocuments = async (method: string) => {
    if (!currentRecord) return

    try {
      // TODO: 替换为实际的API调用
      // await fetch(`/api/offboarding/${currentRecord.id}/handle-documents`, {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ method }),
      // })

      message.success('文档处理完成')
      fetchOffboardings()
      if (detailModalVisible) {
        fetchOffboardingDetails(currentRecord.id)
      }
    } catch (error) {
      message.error('文档处理失败')
    }
  }

  // 获取交接详情
  const fetchOffboardingDetails = async (id: number) => {
    try {
      // TODO: 替换为实际的API调用
      // const response = await fetch(`/api/offboarding/${id}/details`)
      // const data = await response.json()
      // setTransferDetails(data)

      setTransferDetails([])
    } catch (error) {
      console.error('获取交接详情失败', error)
    }
  }

  // 获取用户案件
  const fetchUserCases = async (userId: number) => {
    try {
      // TODO: 替换为实际的API调用
      // const response = await fetch(`/api/users/${userId}/cases`)
      // const data = await response.json()
      // setUserCases(data)

      setUserCases([])
    } catch (error) {
      console.error('获取用户案件失败', error)
    }
  }

  // 获取用户待办
  const fetchUserInboxes = async (userId: number) => {
    try {
      // TODO: 替换为实际的API调用
      // const response = await fetch(`/api/users/${userId}/inbox-items`)
      // const data = await response.json()
      // setUserInboxes(data)

      setUserInboxes([])
    } catch (error) {
      console.error('获取用户待办失败', error)
    }
  }

  // 查看详情
  const handleViewDetail = (record: OffboardingRecord) => {
    setCurrentRecord(record)
    setDetailModalVisible(true)
    fetchOffboardingDetails(record.id)
    fetchUserCases(record.user_id)
    fetchUserInboxes(record.user_id)

    // 设置默认选中的案件和待办
    if (record.new_lawyer_id) {
      const completedCases = transferDetails
        .filter((d) => d.transfer_type === 'case' && d.transfer_status === 'completed')
        .map((d) => d.item_id)
      setCaseTargetKeys(completedCases)
    }

    if (record.new_assistant_id) {
      const completedInboxes = transferDetails
        .filter((d) => d.transfer_type === 'inbox' && d.transfer_status === 'completed')
        .map((d) => d.item_id)
      setInboxTargetKeys(completedInboxes)
    }
  }

  useEffect(() => {
    fetchOffboardings()
    fetchStats()
  }, [fetchOffboardings, fetchStats])

  // 计算进度百分比
  const getProgressPercent = (record: OffboardingRecord) => {
    const totalItems =
      (record.case_transfer_count || 0) + (record.inbox_transfer_count || 0)
    const completedItems =
      (record.case_completed_count || 0) + (record.inbox_completed_count || 0)

    if (totalItems === 0) return 0
    return Math.round((completedItems / totalItems) * 100)
  }

  // 表格列定义
  const columns: ColumnsType<OffboardingRecord> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '离职员工',
      dataIndex: 'user_name',
      key: 'user_name',
      width: 120,
      render: (name: string) => (
        <Space>
          <Avatar size="small" icon={<UserOutlined />} />
          <Text strong>{name}</Text>
        </Space>
      ),
    },
    {
      title: '接手律师',
      dataIndex: 'new_lawyer_name',
      key: 'new_lawyer_name',
      width: 120,
      render: (name: string) => (name ? <TeamOutlined /> : <Text type="secondary">未指定</Text>),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const config = offboardingStatusConfig[status] || {
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
      title: '交接进度',
      key: 'progress',
      width: 180,
      render: (_: unknown, record: OffboardingRecord) => {
        const percent = getProgressPercent(record)
        const completedCases = record.case_completed_count || 0
        const totalCases = record.case_transfer_count || 0
        const completedInboxes = record.inbox_completed_count || 0
        const totalInboxes = record.inbox_transfer_count || 0

        return (
          <Space direction="vertical" size={0} style={{ width: '100%' }}>
            <Progress
              percent={percent}
              size="small"
              status={record.status === 'completed' ? 'success' : 'active'}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              案件 {completedCases}/{totalCases} · 待办 {completedInboxes}/{totalInboxes}
            </Text>
          </Space>
        )
      },
    },
    {
      title: '文档处理',
      dataIndex: 'document_disposal_method',
      key: 'document_disposal_method',
      width: 120,
      render: (method: string, record: OffboardingRecord) => {
        const label = documentMethodConfig[method] || method
        const isCompleted = record.document_disposal_completed_at != null
        return (
          <Tag color={isCompleted ? 'success' : 'warning'}>
            {isCompleted ? <CheckCircleOutlined /> : <ClockCircleOutlined />} {label}
          </Tag>
        )
      },
    },
    {
      title: '发起时间',
      dataIndex: 'initiated_at',
      key: 'initiated_at',
      width: 120,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD'),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right' as const,
      render: (_: unknown, record: OffboardingRecord) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>

          {record.status === 'in_progress' && (
            <Popconfirm
              title="确认完成该交接？"
              description="完成后将自动停用离职员工账号"
              onConfirm={() => handleComplete(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" size="small" icon={<CheckCircleOutlined />} />
            </Popconfirm>
          )}

          {(record.status === 'pending' || record.status === 'in_progress') && (
            <Button
              type="link"
              size="small"
              danger
              icon={<CloseCircleOutlined />}
              onClick={() => {
                setCurrentRecord(record)
                setCancelModalVisible(true)
              }}
            />
          )}
        </Space>
      ),
    },
  ]

  // Transfer 穿梭框渲染
  const renderTransferItem = (item: CaseItem | InboxItem) => {
    if ('case_number' in item) {
      // Case
      return (
        <div className="custom-transfer-item">
          <div>
            <Text strong>{item.title}</Text>
            <br />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {item.case_number}
            </Text>
          </div>
          <Tag color={item.status === 'active' ? 'success' : 'default'}>
            {item.status}
          </Tag>
        </div>
      )
    } else {
      // Inbox
      return (
        <div className="custom-transfer-item">
          <div>
            <Text strong>{item.title}</Text>
            <br />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {item.due_date ? dayjs(item.due_date).format('MM-DD') : '无截止日期'}
            </Text>
          </div>
          <Tag color={item.priority === 'high' ? 'red' : 'blue'}>{item.priority}</Tag>
        </div>
      )
    }
  }

  return (
    <div className="offboarding-management">
      <Card>
        {/* 统计信息 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="总交接数"
                value={stats.total}
                prefix={<LogoutOutlined />}
                valueStyle={{ fontSize: 20 }}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="待处理"
                value={stats.pending}
                valueStyle={{ color: '#faad14' }}
                prefix={<ClockCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="进行中"
                value={stats.inProgress}
                valueStyle={{ color: '#1890ff' }}
                prefix={<SyncOutlined />}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="已完成"
                value={stats.completed}
                valueStyle={{ color: '#52c41a' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={4}>
            <Card size="small">
              <Statistic
                title="本月交接"
                value={stats.thisMonth}
                valueStyle={{ color: '#722ed1' }}
                prefix={<TeamOutlined />}
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
            {Object.entries(offboardingStatusConfig).map(([key, config]) => (
              <Option key={key} value={key}>
                {config.icon} {config.label}
              </Option>
            ))}
          </Select>

          {userRole === 'admin' && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
              发起交接
            </Button>
          )}
        </Space>

        {/* 交接记录列表 */}
        <Table
          rowKey="id"
          columns={columns}
          dataSource={offboardings}
          loading={loading}
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

      {/* 创建交接弹窗 */}
      <Modal
        title={
          <Space>
            <LogoutOutlined />
            <span>发起员工离职交接</span>
          </Space>
        }
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        width={700}
      >
        <Alert
          message="离职交接将撤销该员工的所有登录令牌，并要求完成案件和待办事项的移交"
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />

        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="user_id"
            label="离职员工"
            rules={[{ required: true, message: '请选择离职员工' }]}
          >
            <Select placeholder="选择员工" showSearch filterOption>
              {/* TODO: 从API加载员工列表 */}
              <Option value={1}>示例员工</Option>
            </Select>
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="new_lawyer_id"
                label="接手律师（案件）"
                tooltip="将负责接手该员工的案件"
              >
                <Select placeholder="选择接手律师" allowClear showSearch filterOption>
                  {/* TODO: 从API加载律师列表 */}
                  <Option value={1}>示例律师</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="new_assistant_id"
                label="接手助理（待办）"
                tooltip="将负责接手该员工的待办事项"
              >
                <Select placeholder="选择接手助理" allowClear showSearch filterOption>
                  {/* TODO: 从API加载助理列表 */}
                  <Option value={1}>示例助理</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="document_disposal_method"
            label="文档处理方式"
            rules={[{ required: true, message: '请选择文档处理方式' }]}
          >
            <Select placeholder="选择文档处理方式">
              <Option value="delete">删除文档</Option>
              <Option value="transfer">转移给接手律师</Option>
              <Option value="revoke_access">撤销访问权限</Option>
              <Option value="archive">归档保存</Option>
            </Select>
          </Form.Item>

          <Form.Item name="notes" label="备注说明">
            <TextArea rows={3} placeholder="请输入交接相关的备注说明" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 详情弹窗 */}
      <Modal
        title={
          <Space>
            <UserOutlined />
            <span>交接详情</span>
            {currentRecord && (
              <Tag color={offboardingStatusConfig[currentRecord.status]?.color}>
                {offboardingStatusConfig[currentRecord.status]?.label}
              </Tag>
            )}
          </Space>
        }
        open={detailModalVisible}
        onCancel={() => {
          setDetailModalVisible(false)
          setCurrentRecord(null)
          setTransferDetails([])
          setUserCases([])
          setUserInboxes([])
        }}
        footer={null}
        width={1000}
      >
        {currentRecord && (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            {/* 基本信息 */}
            <Card title="基本信息" size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="交接ID">{currentRecord.id}</Descriptions.Item>
                <Descriptions.Item label="状态">
                  <Tag
                    icon={offboardingStatusConfig[currentRecord.status]?.icon}
                    color={offboardingStatusConfig[currentRecord.status]?.color}
                  >
                    {offboardingStatusConfig[currentRecord.status]?.label}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="离职员工">
                  <Space>
                    <Avatar size="small" icon={<UserOutlined />} />
                    {currentRecord.user_name}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="发起时间">
                  {dayjs(currentRecord.initiated_at).format('YYYY-MM-DD HH:mm')}
                </Descriptions.Item>
                <Descriptions.Item label="接手律师">
                  {currentRecord.new_lawyer_name || <Text type="secondary">未指定</Text>}
                </Descriptions.Item>
                <Descriptions.Item label="接手助理">
                  {currentRecord.new_assistant_name || <Text type="secondary">未指定</Text>}
                </Descriptions.Item>
                <Descriptions.Item label="文档处理方式">
                  {documentMethodConfig[currentRecord.document_disposal_method]}
                </Descriptions.Item>
                <Descriptions.Item label="文档处理状态">
                  {currentRecord.document_disposal_completed_at ? (
                    <Tag color="success" icon={<CheckCircleOutlined />}>
                      已完成
                    </Tag>
                  ) : (
                    <Tag color="warning" icon={<ClockCircleOutlined />}>
                      待处理
                    </Tag>
                  )}
                </Descriptions.Item>
              </Descriptions>

              {currentRecord.notes && (
                <>
                  <Divider style={{ margin: '12px 0' }} />
                  <div>
                    <Text type="secondary">备注：</Text>
                    <Paragraph>{currentRecord.notes}</Paragraph>
                  </div>
                </>
              )}
            </Card>

            {/* 交接进度 */}
            <Card title="交接进度" size="small">
              <Row gutter={16}>
                <Col span={12}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Text strong>案件移交</Text>
                    <Progress
                      percent={
                        currentRecord.case_transfer_count
                          ? Math.round(
                              ((currentRecord.case_completed_count || 0) /
                                currentRecord.case_transfer_count) *
                                100
                            )
                          : 0
                      }
                      status={
                        (currentRecord.case_completed_count || 0) ===
                        (currentRecord.case_transfer_count || 0)
                          ? 'success'
                          : 'active'
                      }
                    />
                    <Text type="secondary">
                      {currentRecord.case_completed_count || 0} /{' '}
                      {currentRecord.case_transfer_count || 0}
                    </Text>
                  </Space>
                </Col>
                <Col span={12}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Text strong>待办移交</Text>
                    <Progress
                      percent={
                        currentRecord.inbox_transfer_count
                          ? Math.round(
                              ((currentRecord.inbox_completed_count || 0) /
                                currentRecord.inbox_transfer_count) *
                                100
                            )
                          : 0
                      }
                      status={
                        (currentRecord.inbox_completed_count || 0) ===
                        (currentRecord.inbox_transfer_count || 0)
                          ? 'success'
                          : 'active'
                      }
                    />
                    <Text type="secondary">
                      {currentRecord.inbox_completed_count || 0} /{' '}
                      {currentRecord.inbox_transfer_count || 0}
                    </Text>
                  </Space>
                </Col>
              </Row>
            </Card>

            {/* 移交操作 */}
            {currentRecord.status === 'in_progress' && (
              <Card title="移交操作" size="small">
                <Space direction="vertical" style={{ width: '100%' }}>
                  {/* 案件移交 */}
                  <div>
                    <Space style={{ marginBottom: 8 }}>
                      <Text strong>案件移交</Text>
                      <Badge
                        count={
                          (currentRecord.case_completed_count || 0) +
                          '/' +
                          (currentRecord.case_transfer_count || 0)
                        }
                      />
                    </Space>

                    {currentRecord.new_lawyer_id ? (
                      <>
                        <Transfer
                          dataSource={userCases}
                          titles={['待移交案件', '已移交案件']}
                          targetKeys={caseTargetKeys}
                          selectedKeys={selectedCases}
                          onChange={setCaseTargetKeys}
                          onSelectChange={(sourceKeys, targetKeys) => {
                            setSelectedCases(
                              targetKeys.filter((key) => caseTargetKeys.includes(key))
                            )
                          }}
                          render={(item) => renderTransferItem(item as CaseItem)}
                          listStyle={{
                            width: 300,
                            height: 300,
                          }}
                          rowKey={(record) => (record as CaseItem).id}
                          showSearch
                          filterOption={(inputValue, item) =>
                            (item as CaseItem).title
                              .toLowerCase()
                              .includes(inputValue.toLowerCase())
                          }
                        />
                        <Button
                          type="primary"
                          block
                          style={{ marginTop: 8 }}
                          onClick={() =>
                            handleBatchTransferCases(caseTargetKeys as number[])
                          }
                          disabled={caseTargetKeys.length === 0}
                        >
                          确认移交选中的 {caseTargetKeys.length} 个案件
                        </Button>
                      </>
                    ) : (
                      <Alert
                        message="请先指定接手律师"
                        type="warning"
                        showIcon
                        style={{ marginTop: 8 }}
                      />
                    )}
                  </div>

                  <Divider />

                  {/* 待办移交 */}
                  <div>
                    <Space style={{ marginBottom: 8 }}>
                      <Text strong>待办事项移交</Text>
                      <Badge
                        count={
                          (currentRecord.inbox_completed_count || 0) +
                          '/' +
                          (currentRecord.inbox_transfer_count || 0)
                        }
                      />
                    </Space>

                    {currentRecord.new_assistant_id ? (
                      <>
                        <Transfer
                          dataSource={userInboxes}
                          titles={['待移交待办', '已移交待办']}
                          targetKeys={inboxTargetKeys}
                          selectedKeys={selectedInboxes}
                          onChange={setInboxTargetKeys}
                          onSelectChange={(sourceKeys, targetKeys) => {
                            setSelectedInboxes(
                              targetKeys.filter((key) => inboxTargetKeys.includes(key))
                            )
                          }}
                          render={(item) => renderTransferItem(item as InboxItem)}
                          listStyle={{
                            width: 300,
                            height: 300,
                          }}
                          rowKey={(record) => (record as InboxItem).id}
                          showSearch
                          filterOption={(inputValue, item) =>
                            (item as InboxItem).title
                              .toLowerCase()
                              .includes(inputValue.toLowerCase())
                          }
                        />
                        <Button
                          type="primary"
                          block
                          style={{ marginTop: 8 }}
                          onClick={() =>
                            handleBatchTransferInboxes(inboxTargetKeys as number[])
                          }
                          disabled={inboxTargetKeys.length === 0}
                        >
                          确认移交选中的 {inboxTargetKeys.length} 个待办
                        </Button>
                      </>
                    ) : (
                      <Alert
                        message="请先指定接手助理"
                        type="warning"
                        showIcon
                        style={{ marginTop: 8 }}
                      />
                    )}
                  </div>

                  <Divider />

                  {/* 文档处理 */}
                  <div>
                    <Text strong>文档处理</Text>
                    {!currentRecord.document_disposal_completed_at ? (
                      <Space wrap style={{ marginTop: 8 }}>
                        <Button
                          icon={<DeleteOutlined />}
                          onClick={() => handleHandleDocuments('delete')}
                        >
                          删除文档
                        </Button>
                        <Button
                          icon={<TeamOutlined />}
                          onClick={() => handleHandleDocuments('transfer')}
                        >
                          转移给接手律师
                        </Button>
                        <Button
                          icon={<LockOutlined />}
                          onClick={() => handleHandleDocuments('revoke_access')}
                        >
                          撤销访问权限
                        </Button>
                        <Button
                          icon={<FileTextOutlined />}
                          onClick={() => handleHandleDocuments('archive')}
                        >
                          归档保存
                        </Button>
                      </Space>
                    ) : (
                      <Alert
                        message="文档已处理完成"
                        type="success"
                        showIcon
                        style={{ marginTop: 8 }}
                      />
                    )}
                  </div>
                </Space>
              </Card>
            )}

            {/* 完成操作 */}
            {currentRecord.status === 'in_progress' && (
              <Card size="small">
                <Space>
                  <Text>确认所有交接工作完成后，可以完成交接流程：</Text>
                  <Popconfirm
                    title="确认完成该交接？"
                    description="完成后将自动停用离职员工账号"
                    onConfirm={() => handleComplete(currentRecord.id)}
                    okText="确认完成"
                    cancelText="取消"
                  >
                    <Button type="primary" icon={<CheckCircleOutlined />}>
                      完成交接
                    </Button>
                  </Popconfirm>
                  <Button
                    danger
                    icon={<CloseCircleOutlined />}
                    onClick={() => setCancelModalVisible(true)}
                  >
                    取消交接
                  </Button>
                </Space>
              </Card>
            )}

            {/* 交接详情列表 */}
            {transferDetails.length > 0 && (
              <Card title="交接明细" size="small">
                <List
                  size="small"
                  dataSource={transferDetails}
                  renderItem={(item) => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={
                          item.transfer_type === 'case' ? (
                            <FileTextOutlined />
                          ) : (
                            <InboxOutlined />
                          )
                        }
                        title={item.item_name}
                        description={
                          <Space>
                            <Text type="secondary">
                              {item.transfer_type === 'case' ? '案件' : '待办'}
                            </Text>
                            <Text type="secondary">→</Text>
                            <Text>{item.new_owner_name || '未分配'}</Text>
                          </Space>
                        }
                      />
                      <Tag
                        color={
                          item.transfer_status === 'completed'
                            ? 'success'
                            : item.transfer_status === 'pending'
                            ? 'warning'
                            : 'default'
                        }
                      >
                        {item.transfer_status === 'completed'
                          ? '已移交'
                          : item.transfer_status === 'pending'
                          ? '待移交'
                          : item.transfer_status}
                      </Tag>
                    </List.Item>
                  )}
                />
              </Card>
            )}
          </Space>
        )}
      </Modal>

      {/* 取消交接弹窗 */}
      <Modal
        title={
          <Space>
            <WarningOutlined />
            <span>取消交接</span>
          </Space>
        }
        open={cancelModalVisible}
        onCancel={() => {
          setCancelModalVisible(false)
          cancelForm.resetFields()
        }}
        onOk={() => cancelForm.submit()}
        okButtonProps={{ danger: true }}
      >
        <Alert
          message="取消交接将停止所有正在进行的移交操作，但已完成的移交不会撤销"
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />

        <Form form={cancelForm} layout="vertical" onFinish={handleCancel}>
          <Form.Item
            name="reason"
            label="取消原因"
            rules={[{ required: true, message: '请输入取消原因' }]}
          >
            <TextArea
              rows={4}
              placeholder="请说明取消交接的原因"
              maxLength={500}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default OffboardingManagement
