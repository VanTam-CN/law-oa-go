import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Row,
  Col,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  Space,
  Tag,
  Badge,
  Progress,
  Statistic,
  Alert,
  Timeline,
  List,
  Avatar,
  Tabs,
  Typography,
  Tooltip,
  Divider,
  Upload,
  message,
  Popconfirm,
  Descriptions,
  Drawer,
  Switch,
  InputNumber,
} from 'antd'
import {
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  FileTextOutlined,
  UploadOutlined,
  DownloadOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  BellOutlined,
  SettingOutlined,
  SearchOutlined,
  FilterOutlined,
  ReloadOutlined,
  CalendarOutlined,
  BarChartOutlined,
  TeamOutlined,
  SafetyOutlined,
  AlertOutlined,
  MonitorOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { UploadProps } from 'antd'
import dayjs from 'dayjs'
import { waiverApprovalService } from '@/services/waiverApproval'
import type {
  WaiverApplication,
  WaiverStatus,
  WaiverType,
  RiskLevel,
  WaiverMonitoringTask,
  MonitoringTaskCompletion,
  CreateMonitoringTaskRequest,
  WaiverStatistics,
} from '@/types/waiverApproval'

const { TextArea } = Input
const { Option } = Select
const { Title, Text, Paragraph } = Typography
const { TabPane } = Tabs
const { RangePicker } = DatePicker

interface WaiverMonitoringDashboardProps {
  userRole?: string
  userId?: string
}

// 监控任务状态配置
const taskStatusConfig: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  PENDING: { label: '待处理', color: 'orange', icon: <ClockCircleOutlined /> },
  IN_PROGRESS: { label: '进行中', color: 'blue', icon: <MonitorOutlined /> },
  COMPLETED: { label: '已完成', color: 'green', icon: <CheckCircleOutlined /> },
  OVERDUE: { label: '已逾期', color: 'red', icon: <WarningOutlined /> },
  CANCELLED: { label: '已取消', color: 'default', icon: <ExclamationCircleOutlined /> },
}

// 监控频率配置
const frequencyConfig = {
  DAILY: { label: '每日', days: 1 },
  WEEKLY: { label: '每周', days: 7 },
  MONTHLY: { label: '每月', days: 30 },
  QUARTERLY: { label: '每季度', days: 90 },
}

const WaiverMonitoringDashboard: React.FC<WaiverMonitoringDashboardProps> = ({
  userRole = 'LAWYER',
  userId,
}) => {
  const [activeTab, setActiveTab] = useState('overview')
  const [monitoringTasks, setMonitoringTasks] = useState<WaiverMonitoringTask[]>([])
  const [approvedWaivers, setApprovedWaivers] = useState<WaiverApplication[]>([])
  const [statistics, setStatistics] = useState<WaiverStatistics | null>(null)
  const [loading, setLoading] = useState(false)
  const [selectedTask, setSelectedTask] = useState<WaiverMonitoringTask | null>(null)
  const [taskDetailVisible, setTaskDetailVisible] = useState(false)
  const [createTaskModalVisible, setCreateTaskModalVisible] = useState(false)
  const [completionModalVisible, setCompletionModalVisible] = useState(false)
  const [evidenceFiles, setEvidenceFiles] = useState<File[]>([])
  const [createTaskForm] = Form.useForm()
  const [completionForm] = Form.useForm()

  // 加载监控任务
  const loadMonitoringTasks = useCallback(async () => {
    try {
      setLoading(true)
      const response = await waiverApprovalService.getMonitoringTasks()
      if (response.success) {
        setMonitoringTasks(response.data)
      }
    } catch (error) {
      message.error('加载监控任务失败')
    } finally {
      setLoading(false)
    }
  }, [])

  // 加载已批准的豁免申请
  const loadApprovedWaivers = useCallback(async () => {
    try {
      setLoading(true)
      const response = await waiverApprovalService.getWaiverApplications({
        status: ['APPROVED'],
        page: 1,
        pageSize: 100,
      })
      if (response.success) {
        setApprovedWaivers(response.data.items)
      }
    } catch (error) {
      message.error('加载已批准豁免失败')
    } finally {
      setLoading(false)
    }
  }, [])

  // 加载统计数据
  const loadStatistics = useCallback(async () => {
    try {
      setLoading(true)
      const response = await waiverApprovalService.getWaiverStatistics()
      if (response.success) {
        setStatistics(response.data)
      }
    } catch (error) {
      message.error('加载统计数据失败')
    } finally {
      setLoading(false)
    }
  }, [])

  // 初始加载
  useEffect(() => {
    loadMonitoringTasks()
    loadApprovedWaivers()
    loadStatistics()
  }, [loadMonitoringTasks, loadApprovedWaivers, loadStatistics])

  // 创建监控任务
  const handleCreateTask = async (values: any) => {
    try {
      setLoading(true)
      const request: CreateMonitoringTaskRequest = {
        applicationId: values.applicationId,
        taskTitle: values.taskTitle,
        taskDescription: values.taskDescription,
        assignedTo: values.assignedTo,
        dueDate: values.dueDate.format('YYYY-MM-DD'),
        frequency: values.frequency,
      }

      const response = await waiverApprovalService.createMonitoringTask(request)
      if (response.success) {
        message.success('监控任务创建成功')
        setCreateTaskModalVisible(false)
        createTaskForm.resetFields()
        loadMonitoringTasks()
      }
    } catch (error: any) {
      message.error(error.message || '创建监控任务失败')
    } finally {
      setLoading(false)
    }
  }

  // 完成监控任务
  const handleCompleteTask = async (values: any) => {
    if (!selectedTask) {
      return
    }

    try {
      setLoading(true)
      await waiverApprovalService.updateMonitoringTaskStatus(
        selectedTask.id,
        values.status,
        values.notes,
        evidenceFiles,
      )

      message.success('任务状态更新成功')
      setCompletionModalVisible(false)
      completionForm.resetFields()
      setEvidenceFiles([])
      setSelectedTask(null)
      loadMonitoringTasks()
    } catch (error: any) {
      message.error(error.message || '更新任务状态失败')
    } finally {
      setLoading(false)
    }
  }

  // 查看任务详情
  const viewTaskDetail = (task: WaiverMonitoringTask) => {
    setSelectedTask(task)
    setTaskDetailVisible(true)
  }

  // 计算任务完成率
  const calculateCompletionRate = (task: WaiverMonitoringTask) => {
    if (task.completionHistory.length === 0) {
      return 0
    }
    const completedCount = task.completionHistory.filter(
      (completion) => completion.status === 'COMPLETED',
    ).length
    return (completedCount / task.completionHistory.length) * 100
  }

  // 检查任务是否逾期
  const isTaskOverdue = (task: WaiverMonitoringTask) => {
    return dayjs().isAfter(dayjs(task.nextDueDate)) && task.isActive
  }

  // 监控任务表格列
  const monitoringTaskColumns: ColumnsType<WaiverMonitoringTask> = [
    {
      title: '任务标题',
      dataIndex: 'taskTitle',
      key: 'taskTitle',
      ellipsis: true,
      render: (text, record) => (
        <Button type='link' onClick={() => viewTaskDetail(record)}>
          {text}
        </Button>
      ),
    },
    {
      title: '相关豁免',
      dataIndex: 'applicationId',
      key: 'applicationId',
      width: 120,
      render: (applicationId) => {
        const waiver = approvedWaivers.find((w) => w.id === applicationId)
        return waiver ? waiver.caseId : applicationId
      },
    },
    {
      title: '负责人',
      dataIndex: 'assignedToName',
      key: 'assignedToName',
      width: 100,
    },
    {
      title: '频率',
      dataIndex: 'frequency',
      key: 'frequency',
      width: 80,
      render: (frequency) => frequencyConfig[frequency as keyof typeof frequencyConfig].label,
    },
    {
      title: '下次截止',
      dataIndex: 'nextDueDate',
      key: 'nextDueDate',
      width: 100,
      render: (date, record) => {
        const isOverdue = isTaskOverdue(record)
        return (
          <Tooltip title={isOverdue ? '已逾期' : '正常'}>
            <Tag color={isOverdue ? 'red' : 'green'}>{dayjs(date).format('MM-DD')}</Tag>
          </Tooltip>
        )
      },
    },
    {
      title: '完成率',
      key: 'completionRate',
      width: 120,
      render: (_, record) => {
        const rate = calculateCompletionRate(record)
        return <Progress percent={rate} size='small' status={rate === 100 ? 'success' : 'active'} />
      },
    },
    {
      title: '状态',
      key: 'status',
      width: 100,
      render: (_, record) => {
        const isOverdue = isTaskOverdue(record)
        const status = isOverdue ? 'OVERDUE' : record.isActive ? 'IN_PROGRESS' : 'CANCELLED'
        const config = taskStatusConfig[status]
        return <Badge color={config.color} text={config.label} />
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space size='small'>
          <Tooltip title='查看详情'>
            <Button type='text' icon={<EyeOutlined />} onClick={() => viewTaskDetail(record)} />
          </Tooltip>
          <Tooltip title='完成监控'>
            <Button
              type='text'
              icon={<CheckCircleOutlined />}
              onClick={() => {
                setSelectedTask(record)
                setCompletionModalVisible(true)
              }}
            />
          </Tooltip>
          <Tooltip title='编辑任务'>
            <Button
              type='text'
              icon={<EditOutlined />}
              // onClick={() => editTask(record)}
            />
          </Tooltip>
          <Tooltip title='停用任务'>
            <Button
              type='text'
              danger
              icon={<DeleteOutlined />}
              // onClick={() => deactivateTask(record.id)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ]

  // 渲染概览页面
  const renderOverview = () => (
    <div>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title='活跃监控任务'
              value={monitoringTasks.filter((t) => t.isActive).length}
              prefix={<MonitorOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title='逾期任务'
              value={monitoringTasks.filter((t) => isTaskOverdue(t)).length}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title='本月完成'
              value={monitoringTasks.reduce((total, task) => {
                const thisMonth = dayjs().format('YYYY-MM')
                return (
                  total +
                  task.completionHistory.filter(
                    (completion) =>
                      dayjs(completion.completedAt).format('YYYY-MM') === thisMonth &&
                      completion.status === 'COMPLETED',
                  ).length
                )
              }, 0)}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title='监控覆盖率'
              value={
                approvedWaivers.length > 0
                  ? Math.round((monitoringTasks.length / approvedWaivers.length) * 100)
                  : 0
              }
              suffix='%'
              prefix={<SafetyOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 逾期任务提醒 */}
      {monitoringTasks.filter((t) => isTaskOverdue(t)).length > 0 && (
        <Alert
          message='逾期任务提醒'
          description={`您有 ${monitoringTasks.filter((t) => isTaskOverdue(t)).length} 个监控任务已逾期，请及时处理`}
          type='warning'
          showIcon
          style={{ marginBottom: 16 }}
          action={
            <Button size='small' danger>
              立即处理
            </Button>
          }
        />
      )}

      {/* 今日待办 */}
      <Row gutter={16}>
        <Col span={12}>
          <Card title='今日待办' extra={<Button icon={<CalendarOutlined />}>查看日历</Button>}>
            <List
              dataSource={monitoringTasks
                .filter((task) => dayjs(task.nextDueDate).isSame(dayjs(), 'day') && task.isActive)
                .slice(0, 5)}
              renderItem={(task) => (
                <List.Item
                  actions={[
                    <Button type='link' size='small' onClick={() => viewTaskDetail(task)}>
                      处理
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<Avatar icon={<ClockCircleOutlined />} />}
                    title={task.taskTitle}
                    description={
                      <Space>
                        <Text type='secondary'>{task.assignedToName}</Text>
                        <Tag color='orange'>今日到期</Tag>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title='监控活动' extra={<Button icon={<ReloadOutlined />}>刷新</Button>}>
            <Timeline>
              {monitoringTasks
                .flatMap((task) => task.completionHistory)
                .sort((a, b) => dayjs(b.completedAt).unix() - dayjs(a.completedAt).unix())
                .slice(0, 5)
                .map((completion, index) => (
                  <Timeline.Item
                    key={index}
                    color={
                      completion.status === 'COMPLETED'
                        ? 'green'
                        : completion.status === 'PARTIALLY_COMPLETED'
                          ? 'orange'
                          : 'red'
                    }
                  >
                    <div>
                      <Text strong>{completion.completedByName}</Text>
                      <Text> 完成了监控任务</Text>
                    </div>
                    <div style={{ fontSize: '12px', color: '#666' }}>
                      {dayjs(completion.completedAt).format('YYYY-MM-DD HH:mm')}
                    </div>
                    <div style={{ fontSize: '12px', color: '#666' }}>{completion.notes}</div>
                  </Timeline.Item>
                ))}
            </Timeline>
          </Card>
        </Col>
      </Row>
    </div>
  )

  // 渲染监控任务管理
  const renderTaskManagement = () => (
    <Card
      title='监控任务管理'
      extra={
        <Space>
          <Button icon={<FilterOutlined />}>筛选</Button>
          <Button icon={<SearchOutlined />}>搜索</Button>
          <Button
            type='primary'
            icon={<PlusOutlined />}
            onClick={() => setCreateTaskModalVisible(true)}
          >
            创建任务
          </Button>
        </Space>
      }
    >
      <Table
        columns={monitoringTaskColumns}
        dataSource={monitoringTasks}
        loading={loading}
        pagination={{
          total: monitoringTasks.length,
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        rowKey='id'
        size='small'
      />
    </Card>
  )

  // 渲染合规分析
  const renderComplianceAnalysis = () => {
    if (!statistics) {
      return null
    }

    return (
      <div>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={12}>
            <Card title='豁免状态分布'>
              <div style={{ height: 200 }}>
                {Object.entries({
                  pending: statistics.pendingApplications,
                  approved: statistics.approvedApplications,
                  rejected: statistics.rejectedApplications,
                  escalated: statistics.escalatedApplications,
                }).map(([key, value]) => (
                  <div key={key} style={{ marginBottom: 12 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Text>
                        {key === 'pending'
                          ? '待审批'
                          : key === 'approved'
                            ? '已批准'
                            : key === 'rejected'
                              ? '已拒绝'
                              : '已上报'}
                      </Text>
                      <Text strong>{value}</Text>
                    </div>
                    <Progress
                      percent={
                        statistics.totalApplications > 0
                          ? (value / statistics.totalApplications) * 100
                          : 0
                      }
                      showInfo={false}
                      size='small'
                    />
                  </div>
                ))}
              </div>
            </Card>
          </Col>
          <Col span={12}>
            <Card title='风险等级监控'>
              <div style={{ height: 200 }}>
                {Object.entries(statistics.riskLevelDistribution).map(([level, count]) => (
                  <div key={level} style={{ marginBottom: 12 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Text>
                        {level === 'LOW'
                          ? '低风险'
                          : level === 'MEDIUM'
                            ? '中风险'
                            : level === 'HIGH'
                              ? '高风险'
                              : '关键风险'}
                      </Text>
                      <Text strong>{count}</Text>
                    </div>
                    <Progress
                      percent={
                        statistics.totalApplications > 0
                          ? (count / statistics.totalApplications) * 100
                          : 0
                      }
                      strokeColor={
                        level === 'LOW'
                          ? '#52c41a'
                          : level === 'MEDIUM'
                            ? '#fa8c16'
                            : level === 'HIGH'
                              ? '#f5222d'
                              : '#722ed1'
                      }
                      showInfo={false}
                      size='small'
                    />
                  </div>
                ))}
              </div>
            </Card>
          </Col>
        </Row>

        <Card title='监控合规性报告'>
          <Descriptions column={2} bordered>
            <Descriptions.Item label='平均审批时间'>
              {statistics.averageReviewTime} 天
            </Descriptions.Item>
            <Descriptions.Item label='批准率'>
              {statistics.approvalRate.toFixed(1)}%
            </Descriptions.Item>
            <Descriptions.Item label='拒绝率'>
              {statistics.rejectionRate.toFixed(1)}%
            </Descriptions.Item>
            <Descriptions.Item label='上报率'>
              {statistics.escalationRate.toFixed(1)}%
            </Descriptions.Item>
            <Descriptions.Item label='监控任务覆盖率'>
              {approvedWaivers.length > 0
                ? Math.round((monitoringTasks.length / approvedWaivers.length) * 100)
                : 0}
              %
            </Descriptions.Item>
            <Descriptions.Item label='任务完成率'>
              {monitoringTasks.length > 0
                ? Math.round(
                    monitoringTasks.reduce(
                      (total, task) => total + calculateCompletionRate(task),
                      0,
                    ) / monitoringTasks.length,
                  )
                : 0}
              %
            </Descriptions.Item>
          </Descriptions>
        </Card>
      </div>
    )
  }

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>豁免监控看板</Title>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <BarChartOutlined />
              监控概览
            </span>
          }
          key='overview'
        >
          {renderOverview()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <MonitorOutlined />
              任务管理
            </span>
          }
          key='tasks'
        >
          {renderTaskManagement()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <SafetyOutlined />
              合规分析
            </span>
          }
          key='compliance'
        >
          {renderComplianceAnalysis()}
        </TabPane>
      </Tabs>

      {/* 创建监控任务模态框 */}
      <Modal
        title='创建监控任务'
        open={createTaskModalVisible}
        onCancel={() => setCreateTaskModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={createTaskForm} layout='vertical' onFinish={handleCreateTask}>
          <Form.Item
            label='关联豁免'
            name='applicationId'
            rules={[{ required: true, message: '请选择关联的豁免申请' }]}
          >
            <Select placeholder='请选择豁免申请'>
              {approvedWaivers.map((waiver) => (
                <Option key={waiver.id} value={waiver.id}>
                  {waiver.caseId} - {waiver.caseTitle}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label='任务标题'
            name='taskTitle'
            rules={[{ required: true, message: '请输入任务标题' }]}
          >
            <Input placeholder='请输入监控任务标题' />
          </Form.Item>

          <Form.Item
            label='任务描述'
            name='taskDescription'
            rules={[{ required: true, message: '请输入任务描述' }]}
          >
            <TextArea rows={3} placeholder='请详细描述监控任务内容' />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label='负责人'
                name='assignedTo'
                rules={[{ required: true, message: '请选择负责人' }]}
              >
                <Select placeholder='请选择负责人'>
                  <Option value='lawyer_001'>张律师</Option>
                  <Option value='lawyer_002'>李律师</Option>
                  <Option value='assistant_001'>王助理</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label='截止日期'
                name='dueDate'
                rules={[{ required: true, message: '请选择截止日期' }]}
              >
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label='监控频率'
            name='frequency'
            rules={[{ required: true, message: '请选择监控频率' }]}
          >
            <Select placeholder='请选择监控频率'>
              <Option value='DAILY'>每日</Option>
              <Option value='WEEKLY'>每周</Option>
              <Option value='MONTHLY'>每月</Option>
              <Option value='QUARTERLY'>每季度</Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button onClick={() => setCreateTaskModalVisible(false)}>取消</Button>
              <Button type='primary' htmlType='submit' loading={loading}>
                创建任务
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 完成任务模态框 */}
      <Modal
        title='完成监控任务'
        open={completionModalVisible}
        onCancel={() => setCompletionModalVisible(false)}
        footer={null}
        width={600}
      >
        {selectedTask && (
          <Form form={completionForm} layout='vertical' onFinish={handleCompleteTask}>
            <div style={{ marginBottom: 16, padding: 16, background: '#f5f5f5', borderRadius: 6 }}>
              <Text strong>任务：{selectedTask.taskTitle}</Text>
              <br />
              <Text type='secondary'>
                截止日期：{dayjs(selectedTask.nextDueDate).format('YYYY-MM-DD')}
              </Text>
            </div>

            <Form.Item
              label='完成状态'
              name='status'
              rules={[{ required: true, message: '请选择完成状态' }]}
              initialValue='COMPLETED'
            >
              <Select>
                <Option value='COMPLETED'>完全完成</Option>
                <Option value='PARTIALLY_COMPLETED'>部分完成</Option>
              </Select>
            </Form.Item>

            <Form.Item
              label='完成说明'
              name='notes'
              rules={[{ required: true, message: '请输入完成说明' }]}
            >
              <TextArea rows={4} placeholder='请详细说明监控执行情况和发现' />
            </Form.Item>

            <Form.Item label='证据材料'>
              <Upload
                multiple
                beforeUpload={() => false}
                onChange={({ fileList }) => {
                  setEvidenceFiles(
                    fileList.map((file) => file.originFileObj).filter(Boolean) as File[],
                  )
                }}
              >
                <Button icon={<UploadOutlined />}>上传证据材料</Button>
              </Upload>
              <div style={{ color: '#666', fontSize: '12px', marginTop: 8 }}>
                支持上传监控报告、检查记录等证据材料
              </div>
            </Form.Item>

            <Form.Item>
              <Space>
                <Button onClick={() => setCompletionModalVisible(false)}>取消</Button>
                <Button type='primary' htmlType='submit' loading={loading}>
                  提交完成
                </Button>
              </Space>
            </Form.Item>
          </Form>
        )}
      </Modal>

      {/* 任务详情抽屉 */}
      <Drawer
        title='监控任务详情'
        placement='right'
        onClose={() => setTaskDetailVisible(false)}
        open={taskDetailVisible}
        width={800}
      >
        {selectedTask && (
          <div>
            <Descriptions column={1} bordered>
              <Descriptions.Item label='任务标题'>{selectedTask.taskTitle}</Descriptions.Item>
              <Descriptions.Item label='任务描述'>{selectedTask.taskDescription}</Descriptions.Item>
              <Descriptions.Item label='负责人'>{selectedTask.assignedToName}</Descriptions.Item>
              <Descriptions.Item label='监控频率'>
                {frequencyConfig[selectedTask.frequency].label}
              </Descriptions.Item>
              <Descriptions.Item label='下次截止'>
                {dayjs(selectedTask.nextDueDate).format('YYYY-MM-DD HH:mm')}
              </Descriptions.Item>
              <Descriptions.Item label='完成率'>
                <Progress percent={calculateCompletionRate(selectedTask)} />
              </Descriptions.Item>
            </Descriptions>

            <Divider />

            <Title level={5}>完成历史</Title>
            <Timeline>
              {selectedTask.completionHistory.map((completion, index) => (
                <Timeline.Item
                  key={index}
                  color={
                    completion.status === 'COMPLETED'
                      ? 'green'
                      : completion.status === 'PARTIALLY_COMPLETED'
                        ? 'orange'
                        : 'red'
                  }
                >
                  <div>
                    <Text strong>{completion.completedByName}</Text>
                    <Tag color='blue' style={{ marginLeft: 8 }}>
                      {completion.status === 'COMPLETED'
                        ? '完全完成'
                        : completion.status === 'PARTIALLY_COMPLETED'
                          ? '部分完成'
                          : '未完成'}
                    </Tag>
                  </div>
                  <div>{completion.notes}</div>
                  <div style={{ fontSize: '12px', color: '#666', marginTop: 4 }}>
                    {dayjs(completion.completedAt).format('YYYY-MM-DD HH:mm:ss')}
                  </div>
                  {completion.evidenceAttachments.length > 0 && (
                    <div style={{ marginTop: 8 }}>
                      <Text type='secondary'>
                        证据材料：{completion.evidenceAttachments.length} 个
                      </Text>
                    </div>
                  )}
                </Timeline.Item>
              ))}
            </Timeline>
          </div>
        )}
      </Drawer>
    </div>
  )
}

export default WaiverMonitoringDashboard
