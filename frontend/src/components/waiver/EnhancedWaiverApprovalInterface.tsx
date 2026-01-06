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
  Timeline,
  Descriptions,
  Divider,
  Tabs,
  Upload,
  message,
  Popconfirm,
  Progress,
  Statistic,
  List,
  Avatar,
  Steps,
  Drawer,
  Switch,
  DatePicker,
  Empty,
  Spin,
} from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  UploadOutlined,
  DownloadOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  UserOutlined,
  FileTextOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  SendOutlined,
  HistoryOutlined,
  SettingOutlined,
  SearchOutlined,
  FilterOutlined,
  SignatureOutlined,
  AuditOutlined,
  TeamOutlined,
  SyncOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { UploadProps } from 'antd'
import dayjs from 'dayjs'
import { waiverApprovalService } from '@/services/waiverApproval'
import {
  conflictDetectionAdapter,
  checkConflicts,
  createWaiverSuggestion,
} from '@/services/conflictDetectionAdapter'
import type {
  WaiverApplication,
  WaiverStatus,
  WaiverType,
  RiskLevel,
  ConflictCheckRequest,
  ConflictCheckResponse,
  ConflictCase,
} from '@/types/waiverApproval'

const { TextArea } = Input
const { Option } = Select
const { Title, Text, Paragraph } = Typography
const { Step } = Steps
const { TabPane } = Tabs

interface EnhancedWaiverApprovalInterfaceProps {
  userRole?: string
  userId?: string
  defaultView?: 'conflicts' | 'pending' | 'history' | 'statistics'
  lawyerId?: string
  lawyerName?: string
}

// 状态配置
const statusConfig: Record<
  WaiverStatus,
  {
    label: string
    color: string
    icon: React.ReactNode
    description: string
  }
> = {
  DRAFT: {
    label: '草稿',
    color: 'default',
    icon: <FileTextOutlined />,
    description: '申请尚未提交',
  },
  SUBMITTED: {
    label: '已提交',
    color: 'processing',
    icon: <SendOutlined />,
    description: '申请已提交，等待审查',
  },
  UNDER_REVIEW: {
    label: '审查中',
    color: 'warning',
    icon: <ClockCircleOutlined />,
    description: '正在审查中',
  },
  APPROVED: {
    label: '已批准',
    color: 'success',
    icon: <CheckCircleOutlined />,
    description: '申请已批准',
  },
  REJECTED: {
    label: '已拒绝',
    icon: <CloseCircleOutlined />,
    color: 'error',
    description: '申请已拒绝',
  },
  ESCALATED: {
    label: '已上报',
    color: 'purple',
    icon: <ExclamationCircleOutlined />,
    description: '已上报给上级',
  },
  EXPIRED: {
    label: '已过期',
    color: 'default',
    icon: <ClockCircleOutlined />,
    description: '申请已过期',
  },
  CANCELLED: {
    label: '已取消',
    color: 'default',
    icon: <CloseCircleOutlined />,
    description: '申请已取消',
  },
}

// 风险等级配置
const riskLevelConfig: Record<
  RiskLevel,
  {
    label: string
    color: string
    priority: number
  }
> = {
  LOW: { label: '低风险', color: 'green', priority: 1 },
  MEDIUM: { label: '中风险', color: 'orange', priority: 2 },
  HIGH: { label: '高风险', color: 'red', priority: 3 },
}

// 冲突类型配置
const conflictTypeConfig: Record<string, { label: string; color: string }> = {
  REPRESENTATION_CONFLICT: { label: '代理冲突', color: 'red' },
  CONFLICT_OF_INTEREST: { label: '利益冲突', color: 'orange' },
  BUSINESS_RELATION: { label: '业务关系', color: 'blue' },
  ORGANIZATIONAL: { label: '组织冲突', color: 'purple' },
  OTHER: { label: '其他', color: 'default' },
}

const EnhancedWaiverApprovalInterface: React.FC<EnhancedWaiverApprovalInterfaceProps> = ({
  userRole = 'LAWYER',
  userId = '',
  defaultView = 'conflicts',
  lawyerId = '',
  lawyerName = '',
}) => {
  // 状态管理
  const [activeTab, setActiveTab] = useState(defaultView)
  const [conflicts, setConflicts] = useState<ConflictCase[]>([])
  const [waiverApplications, setWaiverApplications] = useState<WaiverApplication[]>([])
  const [selectedConflict, setSelectedConflict] = useState<ConflictCase | null>(null)
  const [selectedWaiver, setSelectedWaiver] = useState<WaiverApplication | null>(null)
  const [loading, setLoading] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)

  // 模态框状态
  const [waiverModalVisible, setWaiverModalVisible] = useState(false)
  const [conflictModalVisible, setConflictModalVisible] = useState(false)
  const [approvalModalVisible, setApprovalModalVisible] = useState(false)

  // 表单状态
  const [approvalForm] = Form.useForm()
  const [searchForm] = Form.useForm()

  // 统计数据
  const [statistics, setStatistics] = useState({
    totalConflicts: 0,
    highRiskConflicts: 0,
    pendingWaivers: 0,
    approvedWaivers: 0,
    rejectedWaivers: 0,
  })

  // 刷新数据
  const refreshData = useCallback(() => {
    setRefreshKey((prev) => prev + 1)
  }, [])

  // 加载冲突数据
  useEffect(() => {
    if (lawyerId) {
      loadConflicts()
    }
  }, [lawyerId, refreshKey])

  // 加载豁免申请数据
  useEffect(() => {
    if (userId) {
      loadWaiverApplications()
    }
  }, [userId, refreshKey])

  const loadConflicts = async () => {
    if (!lawyerId) {
      return
    }

    setLoading(true)
    try {
      const conflictData = await conflictDetectionAdapter.getLawyerConflicts(lawyerId)
      setConflicts(conflictData)

      // 更新统计
      const highRiskCount = conflictData.filter((c) => c.riskLevel === 'HIGH').length
      setStatistics((prev) => ({
        ...prev,
        totalConflicts: conflictData.length,
        highRiskConflicts: highRiskCount,
      }))
    } catch (error) {
      console.error('加载冲突数据失败:', error)
      message.error('加载冲突数据失败')
    } finally {
      setLoading(false)
    }
  }

  const loadWaiverApplications = async () => {
    setLoading(true)
    try {
      const response = await waiverApprovalService.getWaiverApplications({
        page: 1,
        pageSize: 50,
        applicantId: userId,
      })

      setWaiverApplications(response.data)

      // 更新统计
      const pending = response.data.filter(
        (w) => w.status === 'SUBMITTED' || w.status === 'UNDER_REVIEW',
      ).length
      const approved = response.data.filter((w) => w.status === 'APPROVED').length
      const rejected = response.data.filter((w) => w.status === 'REJECTED').length

      setStatistics((prev) => ({
        ...prev,
        pendingWaivers: pending,
        approvedWaivers: approved,
        rejectedWaivers: rejected,
      }))
    } catch (error) {
      console.error('加载豁免申请失败:', error)
      message.error('加载豁免申请失败')
    } finally {
      setLoading(false)
    }
  }

  // 从冲突创建豁免申请
  const handleCreateWaiverFromConflict = async (conflict: ConflictCase) => {
    try {
      // 构建冲突检测响应
      const conflictResponse: ConflictCheckResponse = {
        checkId: `CHECK_${Date.now()}`,
        hasConflict: true,
        conflictCases: [conflict],
        checkTime: new Date().toISOString(),
        riskAssessment: {
          overallRiskLevel: conflict.riskLevel,
          highRiskCount: conflict.riskLevel === 'HIGH' ? 1 : 0,
          mediumRiskCount: conflict.riskLevel === 'MEDIUM' ? 1 : 0,
          lowRiskCount: conflict.riskLevel === 'LOW' ? 1 : 0,
          riskFactors: [conflict.conflictType],
        },
      }

      // 创建豁免申请建议
      const suggestion = createWaiverSuggestion(conflictResponse)

      if (suggestion.recommendWaiver) {
        // 构建豁免申请数据
        const waiverData: Partial<WaiverApplication> = {
          caseId: conflict.caseId,
          caseTitle: conflict.caseName,
          waiverType: suggestion.waiverType as WaiverType,
          riskLevel: suggestion.riskLevel,
          description: `基于冲突检测生成的豁免申请：${conflict.description}`,
          justification: `检测到${conflictTypeConfig[conflict.conflictType]?.label || conflict.conflictType}，需要申请豁免以继续代理工作。`,
          mitigationMeasures: suggestion.mitigationMeasures,
          applicantId: userId,
          applicantName: lawyerName,
        }

        // 创建豁免申请
        const waiver = await waiverApprovalService.createWaiverApplication(waiverData)
        message.success('豁免申请创建成功')
        refreshData()
        setSelectedWaiver(waiver)
        setApprovalModalVisible(true)
      } else {
        message.info('该冲突不需要申请豁免')
      }
    } catch (error) {
      console.error('创建豁免申请失败:', error)
      message.error('创建豁免申请失败')
    }
  }

  // 批量创建豁免申请
  const handleBatchCreateWaivers = async () => {
    try {
      const activeConflicts = conflicts.filter(
        (c) => c.status === 'ACTIVE' && !waiverApplications.some((w) => w.caseId === c.caseId),
      )

      if (activeConflicts.length === 0) {
        message.info('没有需要处理的冲突')
        return
      }

      Modal.confirm({
        title: '批量创建豁免申请',
        content: `将为 ${activeConflicts.length} 个冲突案例创建豁免申请，确定继续吗？`,
        onOk: async () => {
          for (const conflict of activeConflicts) {
            await handleCreateWaiverFromConflict(conflict)
          }
          message.success(`成功创建 ${activeConflicts.length} 个豁免申请`)
        },
      })
    } catch (error) {
      console.error('批量创建失败:', error)
      message.error('批量创建豁免申请失败')
    }
  }

  // 冲突案例表格列定义
  const conflictColumns: ColumnsType<ConflictCase> = [
    {
      title: '案件信息',
      key: 'case',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>{record.caseName}</div>
          <div style={{ fontSize: '13px', color: '#374151' }}>案件编号: {record.caseId}</div>
        </div>
      ),
    },
    {
      title: '冲突类型',
      dataIndex: 'conflictType',
      key: 'conflictType',
      render: (type) => (
        <Tag color={conflictTypeConfig[type]?.color || 'default'}>
          {conflictTypeConfig[type]?.label || type}
        </Tag>
      ),
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      render: (level) => (
        <Tag color={riskLevelConfig[level]?.color}>{riskLevelConfig[level]?.label}</Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Badge
          status={status === 'ACTIVE' ? 'processing' : 'default'}
          text={status === 'ACTIVE' ? '进行中' : status}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => {
        const hasWaiver = waiverApplications.some((w) => w.caseId === record.caseId)

        return (
          <Space>
            <Button
              size='small'
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedConflict(record)
                setConflictModalVisible(true)
              }}
            >
              详情
            </Button>
            {!hasWaiver && (
              <Button
                type='primary'
                size='small'
                icon={<FileTextOutlined />}
                onClick={() => handleCreateWaiverFromConflict(record)}
              >
                申请豁免
              </Button>
            )}
            {hasWaiver && (
              <Button size='small' icon={<CheckCircleOutlined />} disabled>
                已申请
              </Button>
            )}
          </Space>
        )
      },
    },
  ]

  // 豁免申请表格列定义
  const waiverColumns: ColumnsType<WaiverApplication> = [
    {
      title: '案件信息',
      key: 'case',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>{record.caseTitle}</div>
          <div style={{ fontSize: '13px', color: '#374151' }}>申请编号: {record.id}</div>
        </div>
      ),
    },
    {
      title: '豁免类型',
      dataIndex: 'waiverType',
      key: 'waiverType',
      render: (type) => <Tag color='blue'>{type}</Tag>,
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      render: (level) => (
        <Tag color={riskLevelConfig[level]?.color}>{riskLevelConfig[level]?.label}</Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const config = statusConfig[status]
        return (
          <Tag color={config.color} icon={config.icon}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '申请时间',
      dataIndex: 'submittedAt',
      key: 'submittedAt',
      render: (time) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            size='small'
            icon={<EyeOutlined />}
            onClick={() => {
              setSelectedWaiver(record)
              setApprovalModalVisible(true)
            }}
          >
            详情
          </Button>
          {record.status === 'DRAFT' && (
            <Button
              type='primary'
              size='small'
              icon={<SendOutlined />}
              onClick={() => handleSubmitWaiver(record.id)}
            >
              提交
            </Button>
          )}
        </Space>
      ),
    },
  ]

  // 提交豁免申请
  const handleSubmitWaiver = async (waiverId: string) => {
    try {
      await waiverApprovalService.submitWaiverApplication(waiverId)
      message.success('豁免申请提交成功')
      refreshData()
    } catch (error) {
      console.error('提交失败:', error)
      message.error('提交失败')
    }
  }

  // 渲染冲突检测页面
  const renderConflictTab = () => (
    <div>
      <Card
        title='冲突检测管理'
        extra={
          <Space>
            <Button type='primary' icon={<SyncOutlined />} onClick={refreshData}>
              刷新数据
            </Button>
            {conflicts.filter(
              (c) =>
                c.status === 'ACTIVE' && !waiverApplications.some((w) => w.caseId === c.caseId),
            ).length > 0 && (
              <Button type='primary' icon={<FileTextOutlined />} onClick={handleBatchCreateWaivers}>
                批量申请豁免
              </Button>
            )}
          </Space>
        }
      >
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card size='small'>
              <Statistic
                title='总冲突数'
                value={statistics.totalConflicts}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card size='small'>
              <Statistic
                title='高风险冲突'
                value={statistics.highRiskConflicts}
                valueStyle={{ color: '#f5222d' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card size='small'>
              <Statistic
                title='待申请豁免'
                value={
                  conflicts.filter(
                    (c) =>
                      c.status === 'ACTIVE' &&
                      !waiverApplications.some((w) => w.caseId === c.caseId),
                  ).length
                }
                valueStyle={{ color: '#fa8c16' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card size='small'>
              <Statistic
                title='已申请豁免'
                value={waiverApplications.length}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>

        <Table
          columns={conflictColumns}
          dataSource={conflicts}
          rowKey='id'
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
          }}
        />
      </Card>
    </div>
  )

  // 渲染豁免申请页面
  const renderWaiverTab = () => (
    <div>
      <Card
        title='豁免申请管理'
        extra={
          <Space>
            <Button icon={<SyncOutlined />} onClick={refreshData}>
              刷新数据
            </Button>
          </Space>
        }
      >
        <Table
          columns={waiverColumns}
          dataSource={waiverApplications}
          rowKey='id'
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
          }}
        />
      </Card>
    </div>
  )

  // 渲染统计页面
  const renderStatisticsTab = () => (
    <div>
      <Card title='数据统计'>
        <Row gutter={16}>
          <Col span={12}>
            <Card size='small' title='冲突统计'>
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic
                    title='总冲突数'
                    value={statistics.totalConflicts}
                    valueStyle={{ color: '#1890ff' }}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title='高风险冲突'
                    value={statistics.highRiskConflicts}
                    valueStyle={{ color: '#f5222d' }}
                  />
                </Col>
              </Row>
              <Divider />
              <div>
                <Text strong>冲突类型分布：</Text>
                <div style={{ marginTop: 8 }}>
                  {Object.entries(
                    conflicts.reduce(
                      (acc, conflict) => {
                        acc[conflict.conflictType] = (acc[conflict.conflictType] || 0) + 1
                        return acc
                      },
                      {} as Record<string, number>,
                    ),
                  ).map(([type, count]) => (
                    <Tag key={type} color={conflictTypeConfig[type]?.color}>
                      {conflictTypeConfig[type]?.label || type}: {count}
                    </Tag>
                  ))}
                </div>
              </div>
            </Card>
          </Col>
          <Col span={12}>
            <Card size='small' title='豁免申请统计'>
              <Row gutter={16}>
                <Col span={8}>
                  <Statistic
                    title='待审批'
                    value={statistics.pendingWaivers}
                    valueStyle={{ color: '#fa8c16' }}
                  />
                </Col>
                <Col span={8}>
                  <Statistic
                    title='已批准'
                    value={statistics.approvedWaivers}
                    valueStyle={{ color: '#52c41a' }}
                  />
                </Col>
                <Col span={8}>
                  <Statistic
                    title='已拒绝'
                    value={statistics.rejectedWaivers}
                    valueStyle={{ color: '#f5222d' }}
                  />
                </Col>
              </Row>
              <Divider />
              <div>
                <Text strong>审批状态分布：</Text>
                <div style={{ marginTop: 8 }}>
                  {Object.entries(
                    waiverApplications.reduce(
                      (acc, waiver) => {
                        acc[waiver.status] = (acc[waiver.status] || 0) + 1
                        return acc
                      },
                      {} as Record<string, number>,
                    ),
                  ).map(([status, count]) => (
                    <Tag key={status} color={statusConfig[status as WaiverStatus]?.color}>
                      {statusConfig[status as WaiverStatus]?.label}: {count}
                    </Tag>
                  ))}
                </div>
              </div>
            </Card>
          </Col>
        </Row>
      </Card>
    </div>
  )

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>
        增强型豁免审批管理
        {lawyerName && (
          <Text style={{ fontSize: 16, fontWeight: 'normal', marginLeft: 16 }}>- {lawyerName}</Text>
        )}
      </Title>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <WarningOutlined />
              冲突检测
            </span>
          }
          key='conflicts'
        >
          {renderConflictTab()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <FileTextOutlined />
              豁免申请
            </span>
          }
          key='waivers'
        >
          {renderWaiverTab()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <InfoCircleOutlined />
              数据统计
            </span>
          }
          key='statistics'
        >
          {renderStatisticsTab()}
        </TabPane>
      </Tabs>

      {/* 冲突详情模态框 */}
      <Modal
        title='冲突详情'
        open={conflictModalVisible}
        onCancel={() => setConflictModalVisible(false)}
        footer={[
          <Button key='close' onClick={() => setConflictModalVisible(false)}>
            关闭
          </Button>,
          selectedConflict &&
            !waiverApplications.some((w) => w.caseId === selectedConflict.caseId) && (
              <Button
                key='create'
                type='primary'
                icon={<FileTextOutlined />}
                onClick={() => {
                  if (selectedConflict) {
                    handleCreateWaiverFromConflict(selectedConflict)
                    setConflictModalVisible(false)
                  }
                }}
              >
                申请豁免
              </Button>
            ),
        ]}
        width={800}
      >
        {selectedConflict && (
          <div>
            <Descriptions column={2} bordered>
              <Descriptions.Item label='案件名称' span={2}>
                {selectedConflict.caseName}
              </Descriptions.Item>
              <Descriptions.Item label='案件编号'>{selectedConflict.caseId}</Descriptions.Item>
              <Descriptions.Item label='冲突类型'>
                <Tag color={conflictTypeConfig[selectedConflict.conflictType]?.color}>
                  {conflictTypeConfig[selectedConflict.conflictType]?.label}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label='风险等级'>
                <Tag color={riskLevelConfig[selectedConflict.riskLevel]?.color}>
                  {riskLevelConfig[selectedConflict.riskLevel]?.label}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label='当前状态'>
                <Badge
                  status={selectedConflict.status === 'ACTIVE' ? 'processing' : 'default'}
                  text={selectedConflict.status === 'ACTIVE' ? '进行中' : selectedConflict.status}
                />
              </Descriptions.Item>
              <Descriptions.Item label='创建时间'>
                {dayjs(selectedConflict.createdAt).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label='冲突描述' span={2}>
                <Paragraph>{selectedConflict.description}</Paragraph>
              </Descriptions.Item>
              <Descriptions.Item label='相关方' span={2}>
                {selectedConflict.parties.map((party, index) => (
                  <Tag key={index} style={{ margin: '4px 4px 4px 0' }}>
                    {party.name} ({party.type === 'LAWYER' ? '律师' : '客户'})
                  </Tag>
                ))}
              </Descriptions.Item>
              <Descriptions.Item label='风险控制措施' span={2}>
                <Paragraph>{selectedConflict.mitigationMeasures}</Paragraph>
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>

      {/* 豁免申请详情模态框 */}
      <Modal
        title='豁免申请详情'
        open={approvalModalVisible}
        onCancel={() => setApprovalModalVisible(false)}
        footer={[
          <Button key='close' onClick={() => setApprovalModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={1000}
      >
        {selectedWaiver && (
          <div>
            <Descriptions column={2} bordered>
              <Descriptions.Item label='申请编号' span={2}>
                {selectedWaiver.id}
              </Descriptions.Item>
              <Descriptions.Item label='案件标题' span={2}>
                {selectedWaiver.caseTitle}
              </Descriptions.Item>
              <Descriptions.Item label='豁免类型'>
                <Tag color='blue'>{selectedWaiver.waiverType}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label='风险等级'>
                <Tag color={riskLevelConfig[selectedWaiver.riskLevel]?.color}>
                  {riskLevelConfig[selectedWaiver.riskLevel]?.label}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label='申请状态'>
                {(() => {
                  const config = statusConfig[selectedWaiver.status]
                  return (
                    <Tag color={config.color} icon={config.icon}>
                      {config.label}
                    </Tag>
                  )
                })()}
              </Descriptions.Item>
              <Descriptions.Item label='申请人'>{selectedWaiver.applicantName}</Descriptions.Item>
              <Descriptions.Item label='申请时间'>
                {selectedWaiver.submittedAt
                  ? dayjs(selectedWaiver.submittedAt).format('YYYY-MM-DD HH:mm:ss')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label='申请描述' span={2}>
                <Paragraph>{selectedWaiver.description}</Paragraph>
              </Descriptions.Item>
              <Descriptions.Item label='申请理由' span={2}>
                <Paragraph>{selectedWaiver.justification}</Paragraph>
              </Descriptions.Item>
              <Descriptions.Item label='风险控制措施' span={2}>
                <Paragraph>{selectedWaiver.mitigationMeasures}</Paragraph>
              </Descriptions.Item>
            </Descriptions>

            {selectedWaiver.approvalRecords && selectedWaiver.approvalRecords.length > 0 && (
              <>
                <Divider />
                <Title level={4}>审批记录</Title>
                <Timeline>
                  {selectedWaiver.approvalRecords.map((record, index) => (
                    <Timeline.Item
                      key={index}
                      color={
                        record.action === 'APPROVE'
                          ? 'green'
                          : record.action === 'REJECT'
                            ? 'red'
                            : 'blue'
                      }
                    >
                      <div>
                        <Text strong>{record.reviewerName}</Text>
                        <Tag color='blue' style={{ marginLeft: 8 }}>
                          {record.action === 'APPROVE'
                            ? '批准'
                            : record.action === 'REJECT'
                              ? '拒绝'
                              : '提交'}
                        </Tag>
                        <br />
                        <Text type='secondary'>
                          {dayjs(record.reviewTime).format('YYYY-MM-DD HH:mm:ss')}
                        </Text>
                        {record.comments && (
                          <div style={{ marginTop: 4 }}>
                            <Text>{record.comments}</Text>
                          </div>
                        )}
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
              </>
            )}
          </div>
        )}
      </Modal>
    </div>
  )
}

export default EnhancedWaiverApprovalInterface
