import React, { useState, useEffect } from 'react'
import {
  Card,
  Form,
  Input,
  Button,
  Select,
  Table,
  Tag,
  Alert,
  Spin,
  Space,
  Divider,
  Row,
  Col,
  Statistic,
  List,
  Modal,
  Tooltip,
  Switch,
  Progress,
  message,
} from 'antd'
import {
  SearchOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  DownloadOutlined,
  HistoryOutlined,
  WarningOutlined,
  FileTextOutlined,
  SyncOutlined,
  SafetyCertificateOutlined,
  SendOutlined,
  AuditOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { post, get } from '@/services/http'
import './ConflictCheck.less'

const { Option } = Select

// 类型定义
interface ConflictMatchV2 {
  matchId: string
  matchType: 'direct' | 'indirect' | 'related' | 'opposing' | 'api'
  lawyerId: number
  lawyerName: string
  caseId: number
  caseTitle: string
  caseType: string
  relationship: string
  matchReason: string
  entityInfo: {
    name: string
    standardName: string
    taxId: string
    type: string
  }
  riskLevel: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'PASS'
  riskFactors: string[]
}

interface ConflictCheckResultV2 {
  checkId: string
  riskLevel: string
  riskScore: number
  matchCount: number
  matches: ConflictMatchV2[]
  checkTime: string
  durationMs: number
  searchScope: string
  recommendations: string[]
}

interface ConflictReport {
  id: number
  reportNumber: string
  clientName: string
  riskLevel: string
  reportUrl: string
  createdAt: string
}

interface ConflictScanJob {
  id: number
  scanType: string
  status: string
  scannedCases: number
  foundConflicts: number
  createdAt: string
  completedAt?: string
}

const ConflictDetectionV2: React.FC = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState<boolean>(false)
  const [result, setResult] = useState<ConflictCheckResultV2 | null>(null)
  const [generatingReport, setGeneratingReport] = useState<boolean>(false)
  const [showHistory, setShowHistory] = useState<boolean>(false)
  const [historyLoading, setHistoryLoading] = useState<boolean>(false)
  const [reports, setReports] = useState<ConflictReport[]>([])
  const [scanJobs, setScanJobs] = useState<ConflictScanJob[]>([])
  const [detailModalVisible, setDetailModalVisible] = useState<boolean>(false)
  const [selectedMatch, setSelectedMatch] = useState<ConflictMatchV2 | null>(null)
  const [searchDepth, setSearchDepth] = useState<'basic' | 'standard' | 'deep'>('standard')
  const [approvalStatus, setApprovalStatus] = useState<'none' | 'pending' | 'approved' | 'rejected' | 'waived'>('none')
  const [submittingApproval, setSubmittingApproval] = useState(false)

  // 风险等级配置
  const riskConfig = {
    CRITICAL: { color: 'error', text: '严重风险', icon: <ExclamationCircleOutlined />, score: 80 },
    HIGH: { color: 'warning', text: '高风险', icon: <WarningOutlined />, score: 60 },
    MEDIUM: { color: 'processing', text: '中风险', icon: <ExclamationCircleOutlined />, score: 40 },
    LOW: { color: 'default', text: '低风险', icon: <ExclamationCircleOutlined />, score: 20 },
    PASS: { color: 'success', text: '无冲突', icon: <CheckCircleOutlined />, score: 0 },
  }

  // 审批状态配置
  const approvalStatusConfig: Record<string, { color: string; text: string }> = {
    none: { color: 'default', text: '未提交' },
    pending: { color: 'processing', text: '审批中' },
    approved: { color: 'success', text: '已批准' },
    rejected: { color: 'error', text: '已拒绝' },
    waived: { color: 'warning', text: '已豁免' },
  }

  // 执行冲突检测
  const handleCheck = async (values: any) => {
    try {
      setLoading(true)
      setResult(null)
      setApprovalStatus('none')

      const requestData = {
        lawyerId: values.lawyerId,
        clientName: values.clientName,
        clientTaxId: values.clientTaxId || undefined,
        caseId: values.caseId || undefined,
        opposingParties: values.opposingParties
          ? values.opposingParties.split(',').map((s: string) => s.trim()).filter(Boolean)
          : [],
        searchDepth,
        includeRelated: values.includeRelated ?? true,
      }

      const response = await post<any>('/api/v2/conflict/check', requestData)

      if (response.data) {
        setResult(response.data)

        if (response.data.riskLevel === 'PASS') {
          message.success('未检测到利益冲突')
        } else {
          message.warning(`检测到${riskConfig[response.data.riskLevel as keyof typeof riskConfig]?.text || '潜在'}冲突`)
        }
      }
    } catch (error: any) {
      console.error('Conflict check error:', error)
      message.error(error.message || '冲突检测失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 生成 PDF 报告
  const handleGenerateReport = async () => {
    if (!result) return

    try {
      setGeneratingReport(true)

      await post<any>('/api/conflict/report', {
        checkedBy: 1, // TODO: 从用户信息获取
        checkTime: result.checkTime,
        checkDurationMs: result.durationMs,
        clientName: form.getFieldValue('clientName'),
        clientTaxId: form.getFieldValue('clientTaxId'),
        opposingParty: form.getFieldValue('opposingParties'),
        riskLevel: result.riskLevel,
        matchedCases: result.matches,
        relatedCompanies: [],
        conflictDetails: {},
        templateType: 'standard',
      })

      message.success('报告生成成功')
    } catch (error: any) {
      console.error('Generate report error:', error)
      message.error('报告生成失败')
    } finally {
      setGeneratingReport(false)
    }
  }

  // 查看历史记录
  const handleViewHistory = async () => {
    try {
      setShowHistory(true)
      setHistoryLoading(true)

      const [reportsRes, jobsRes] = await Promise.all([
        get<any>('/api/conflict/reports'),
        get<any>('/api/conflict/scan-jobs'),
      ])

      setReports(reportsRes.data?.list || [])
      setScanJobs(jobsRes.data?.list || [])
    } catch (error: any) {
      console.error('Load history error:', error)
      message.error('加载历史记录失败')
    } finally {
      setHistoryLoading(false)
    }
  }

  // 提交利益冲突审批
  const handleSubmitApproval = () => {
    if (!result) return

    Modal.confirm({
      title: '提交利益冲突审批',
      icon: <AuditOutlined />,
      content: (
        <div>
          <p>确认将此次冲突检测结果提交审批？</p>
          <p style={{ color: '#999', fontSize: 12 }}>
            风险等级: {riskConfig[result.riskLevel as keyof typeof riskConfig]?.text}，
            匹配数量: {result.matchCount}
          </p>
        </div>
      ),
      okText: '确认提交',
      cancelText: '取消',
      onOk: async () => {
        try {
          setSubmittingApproval(true)

          await post('/api/integration/approvals/with-conflict', {
            checkId: result.checkId,
            riskLevel: result.riskLevel,
            riskScore: result.riskScore,
            matchCount: result.matchCount,
            clientName: form.getFieldValue('clientName'),
            caseId: form.getFieldValue('caseId') || undefined,
            matches: result.matches.map((m) => ({
              matchId: m.matchId,
              matchType: m.matchType,
              lawyerName: m.lawyerName,
              caseTitle: m.caseTitle,
              entityName: m.entityInfo.name,
              riskLevel: m.riskLevel,
              matchReason: m.matchReason,
            })),
            checkTime: result.checkTime,
            searchScope: result.searchScope,
            recommendations: result.recommendations,
          })

          setApprovalStatus('pending')
          message.success('审批申请已提交，请等待审批结果')
        } catch (error: unknown) {
          const errMsg = error instanceof Error ? error.message : '提交审批失败，请稍后重试'
          message.error(errMsg)
        } finally {
          setSubmittingApproval(false)
        }
      },
    })
  }

  // 查看冲突详情
  const handleViewDetail = (match: ConflictMatchV2) => {
    setSelectedMatch(match)
    setDetailModalVisible(true)
  }

  // 获取匹配类型标签
  const getMatchTypeTag = (type: string) => {
    const config: Record<string, { color: string; text: string }> = {
      direct: { color: 'error', text: '直接匹配' },
      indirect: { color: 'warning', text: '间接匹配' },
      related: { color: 'processing', text: '关联匹配' },
      opposing: { color: 'error', text: '对方当事人' },
      api: { color: 'blue', text: 'API 匹配' },
    }
    const { color, text } = config[type] || { color: 'default', text: type }
    return <Tag color={color}>{text}</Tag>
  }

  // 表格列定义
  const columns: ColumnsType<ConflictMatchV2> = [
    {
      title: '匹配类型',
      dataIndex: 'matchType',
      key: 'matchType',
      width: 100,
      render: (type: string) => getMatchTypeTag(type),
    },
    {
      title: '实体信息',
      key: 'entity',
      render: (_, record) => (
        <div>
          <div>{record.entityInfo.name}</div>
          {record.entityInfo.taxId && (
            <div style={{ fontSize: '12px', color: '#999' }}>{record.entityInfo.taxId}</div>
          )}
        </div>
      ),
    },
    {
      title: '关联案件',
      dataIndex: 'caseTitle',
      key: 'caseTitle',
      ellipsis: true,
    },
    {
      title: '律师',
      dataIndex: 'lawyerName',
      key: 'lawyerName',
      width: 80,
    },
    {
      title: '关系',
      dataIndex: 'relationship',
      key: 'relationship',
      width: 80,
      render: (rel: string) => {
        const config: Record<string, string> = {
          client: '客户',
          opposing: '对方',
          witness: '证人',
        }
        return config[rel] || rel
      },
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      width: 100,
      render: (level: keyof typeof riskConfig) => {
        const config = riskConfig[level]
        return (
          <Tag color={config.color} icon={config.icon}>
            {config.text}
          </Tag>
        )
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
          详情
        </Button>
      ),
    },
  ]

  return (
    <div className="conflict-check-container">
      <Row gutter={[16, 16]}>
        {/* 左侧：检测表单 */}
        <Col xs={24} lg={8}>
          <Card title="利益冲突检测" extra={<SafetyCertificateOutlined />}>
            <Form
              form={form}
              layout="vertical"
              onFinish={handleCheck}
              initialValues={{
                includeRelated: true,
                searchDepth: 'standard',
              }}
            >
              <Form.Item
                name="lawyerId"
                label="律师"
                rules={[{ required: true, message: '请选择律师' }]}
              >
                <Select placeholder="请选择律师" showSearch>
                  <Option value="1">系统管理员</Option>
                  <Option value="2">张律师</Option>
                  <Option value="3">李律师</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="clientName"
                label="客户名称"
                rules={[{ required: true, message: '请输入客户名称' }]}
              >
                <Input placeholder="请输入客户/公司名称" />
              </Form.Item>

              <Form.Item name="clientTaxId" label="客户税号（可选）">
                <Input placeholder="请输入统一社会信用代码" />
              </Form.Item>

              <Form.Item name="caseId" label="关联案件（可选）">
                <Select placeholder="请选择案件" allowClear>
                  <Option value="1">案件A</Option>
                  <Option value="2">案件B</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="opposingParties"
                label="对方当事人（可选）"
                tooltip="多个当事人用逗号分隔"
              >
                <Input placeholder="如：公司A, 个人B" />
              </Form.Item>

              <Form.Item label="搜索深度">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Select value={searchDepth} onChange={setSearchDepth}>
                    <Option value="basic">基础搜索（直接匹配）</Option>
                    <Option value="standard">标准搜索（含模糊匹配）</Option>
                    <Option value="deep">深度搜索（含 API 查询）</Option>
                  </Select>
                </Space>
              </Form.Item>

              <Form.Item name="includeRelated" valuePropName="checked">
                <Switch /> <span style={{ marginLeft: 8 }}>包含关联实体</span>
              </Form.Item>

              <Form.Item>
                <Space style={{ width: '100%' }}>
                  <Button
                    type="primary"
                    htmlType="submit"
                    icon={<SearchOutlined />}
                    loading={loading}
                    block
                  >
                    开始检测
                  </Button>
                  <Button onClick={() => form.resetFields()}>重置</Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>

          {/* 统计信息卡片 */}
          {result && (
            <Card title="检测统计" style={{ marginTop: 16 }}>
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic
                    title="风险分数"
                    value={result.riskScore}
                    precision={2}
                    suffix="/ 1.00"
                    valueStyle={{ color: result.riskScore > 0.6 ? '#cf1322' : '#3f8600' }}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="检测耗时"
                    value={result.durationMs}
                    suffix="ms"
                  />
                </Col>
              </Row>
              <Divider style={{ margin: '12px 0' }} />
              <div style={{ fontSize: '12px', color: '#666' }}>
                <div>搜索范围: {result.searchScope}</div>
                <div>检测时间: {new Date(result.checkTime).toLocaleString()}</div>
              </div>
            </Card>
          )}
        </Col>

        {/* 右侧：检测结果 */}
        <Col xs={24} lg={16}>
          {result && (
            <>
              {/* 风险评估 */}
              <Card
                title={
                  <Space>
                    {riskConfig[result.riskLevel as keyof typeof riskConfig]?.icon}
                    <span>风险评估结果</span>
                  </Space>
                }
                extra={
                  <Space>
                    <Tag
                      color={approvalStatusConfig[approvalStatus].color}
                      icon={<AuditOutlined />}
                    >
                      审批状态: {approvalStatusConfig[approvalStatus].text}
                    </Tag>
                    <Button
                      icon={<DownloadOutlined />}
                      onClick={handleGenerateReport}
                      loading={generatingReport}
                      size="small"
                    >
                      生成报告
                    </Button>
                    <Button
                      icon={<HistoryOutlined />}
                      onClick={handleViewHistory}
                      size="small"
                    >
                      历史记录
                    </Button>
                  </Space>
                }
              >
                <Alert
                  message={
                    <Space>
                      <span>风险等级: </span>
                      <Tag
                        color={riskConfig[result.riskLevel as keyof typeof riskConfig]?.color}
                        icon={riskConfig[result.riskLevel as keyof typeof riskConfig]?.icon}
                      >
                        {riskConfig[result.riskLevel as keyof typeof riskConfig]?.text}
                      </Tag>
                    </Space>
                  }
                  description={
                    <div>
                      {result.recommendations.map((rec, idx) => (
                        <div key={idx}>• {rec}</div>
                      ))}
                    </div>
                  }
                  type={
                    result.riskLevel === 'PASS'
                      ? 'success'
                      : result.riskLevel === 'CRITICAL' || result.riskLevel === 'HIGH'
                        ? 'error'
                        : 'warning'
                  }
                  showIcon
                />

                {result.matchCount > 0 && (
                  <>
                    <Divider />
                    <Progress
                      percent={Math.round(result.riskScore * 100)}
                      status={result.riskLevel === 'PASS' ? 'success' : 'exception'}
                      strokeColor={result.riskLevel === 'PASS' ? '#52c41a' : '#ff4d4f'}
                    />
                  </>
                )}

                {result.riskLevel !== 'PASS' && (
                  <div style={{ marginTop: 16, textAlign: 'right' }}>
                    <Button
                      type="primary"
                      icon={<SendOutlined />}
                      onClick={handleSubmitApproval}
                      loading={submittingApproval}
                      disabled={approvalStatus === 'pending' || approvalStatus === 'approved'}
                    >
                      {approvalStatus === 'none'
                        ? '提交利益冲突审批'
                        : approvalStatus === 'pending'
                          ? '审批中...'
                          : approvalStatus === 'approved'
                            ? '已批准'
                            : approvalStatus === 'waived'
                              ? '已豁免'
                              : '重新提交审批'}
                    </Button>
                  </div>
                )}
              </Card>

              {/* 匹配结果列表 */}
              {result.matchCount > 0 && (
                <Card title={`匹配结果 (${result.matchCount})`} style={{ marginTop: 16 }}>
                  <Table
                    columns={columns}
                    dataSource={result.matches}
                    rowKey="matchId"
                    pagination={{ pageSize: 10 }}
                    size="small"
                  />
                </Card>
              )}

              {/* 无冲突时的显示 */}
              {result.matchCount === 0 && (
                <Card style={{ marginTop: 16 }}>
                  <div style={{ textAlign: 'center', padding: '40px 0' }}>
                    <CheckCircleOutlined style={{ fontSize: 64, color: '#52c41a' }} />
                    <div style={{ marginTop: 16, fontSize: 16, color: '#52c41a' }}>
                      未检测到利益冲突
                    </div>
                    <div style={{ marginTop: 8, color: '#999' }}>
                      可以正常处理此案件
                    </div>
                  </div>
                </Card>
              )}
            </>
          )}

          {/* 初始状态 */}
          {!result && (
            <Card>
              <div style={{ textAlign: 'center', padding: '60px 0', color: '#999' }}>
                <SearchOutlined style={{ fontSize: 64, marginBottom: 16 }} />
                <div>请填写检测信息并点击"开始检测"</div>
              </div>
            </Card>
          )}
        </Col>
      </Row>

      {/* 历史记录弹窗 */}
      <Modal
        title={
          <Space>
            <HistoryOutlined />
            历史记录
          </Space>
        }
        open={showHistory}
        onCancel={() => setShowHistory(false)}
        width={800}
        footer={[
          <Button key="close" onClick={() => setShowHistory(false)}>
            关闭
          </Button>,
        ]}
      >
        <Spin spinning={historyLoading}>
          <Divider orientation="left">检测报告</Divider>
          {reports.length > 0 ? (
            <List
              dataSource={reports}
              renderItem={(item) => (
                <List.Item
                  actions={[
                    <Button type="link" size="small" icon={<DownloadOutlined />}>
                      下载
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    title={item.reportNumber}
                    description={
                      <Space split={<Divider type="vertical" />}>
                        <span>{item.clientName}</span>
                        <Tag>{item.riskLevel}</Tag>
                        <span>{new Date(item.createdAt).toLocaleString()}</span>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          ) : (
            <div style={{ textAlign: 'center', padding: 20, color: '#999' }}>暂无报告记录</div>
          )}

          <Divider orientation="left">扫描任务</Divider>
          {scanJobs.length > 0 ? (
            <List
              dataSource={scanJobs}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={
                      item.status === 'running' ? (
                        <SyncOutlined spin />
                      ) : item.status === 'completed' ? (
                        <CheckCircleOutlined style={{ color: '#52c41a' }} />
                      ) : (
                        <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
                      )
                    }
                    title={
                      <Space>
                        <span>{item.scanType === 'daily' ? '每日扫描' : item.scanType === 'weekly' ? '每周扫描' : '手动扫描'}</span>
                        <Tag>{item.status}</Tag>
                      </Space>
                    }
                    description={
                      <Space split={<Divider type="vertical" />}>
                        <span>扫描案件: {item.scannedCases}</span>
                        <span>发现冲突: {item.foundConflicts}</span>
                        <span>{item.completedAt ? new Date(item.completedAt).toLocaleString() : new Date(item.createdAt).toLocaleString()}</span>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          ) : (
            <div style={{ textAlign: 'center', padding: 20, color: '#999' }}>暂无扫描记录</div>
          )}
        </Spin>
      </Modal>

      {/* 详情弹窗 */}
      <Modal
        title="冲突详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
      >
        {selectedMatch && (
          <div>
            <Divider orientation="left">基本信息</Divider>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <div style={{ color: '#666' }}>匹配类型</div>
                <div>{getMatchTypeTag(selectedMatch.matchType)}</div>
              </Col>
              <Col span={12}>
                <div style={{ color: '#666' }}>风险等级</div>
                <div>
                  <Tag
                    color={riskConfig[selectedMatch.riskLevel]?.color}
                    icon={riskConfig[selectedMatch.riskLevel]?.icon}
                  >
                    {riskConfig[selectedMatch.riskLevel]?.text}
                  </Tag>
                </div>
              </Col>
              <Col span={24}>
                <div style={{ color: '#666' }}>实体信息</div>
                <div>{selectedMatch.entityInfo.name}</div>
                {selectedMatch.entityInfo.taxId && (
                  <div style={{ color: '#999', fontSize: 12 }}>{selectedMatch.entityInfo.taxId}</div>
                )}
              </Col>
              <Col span={12}>
                <div style={{ color: '#666' }}>关联案件</div>
                <div>{selectedMatch.caseTitle}</div>
              </Col>
              <Col span={12}>
                <div style={{ color: '#666' }}>负责律师</div>
                <div>{selectedMatch.lawyerName}</div>
              </Col>
              <Col span={24}>
                <div style={{ color: '#666' }}>匹配原因</div>
                <div>{selectedMatch.matchReason}</div>
              </Col>
            </Row>

            {selectedMatch.riskFactors.length > 0 && (
              <>
                <Divider orientation="left">风险因素</Divider>
                <div>
                  {selectedMatch.riskFactors.map((factor, idx) => (
                    <Tag key={idx} color="warning" style={{ marginBottom: 8 }}>
                      {factor}
                    </Tag>
                  ))}
                </div>
              </>
            )}
          </div>
        )}
      </Modal>
    </div>
  )
}

export default ConflictDetectionV2
