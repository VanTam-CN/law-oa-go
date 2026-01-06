import React from 'react'
import { Modal, Typography, Tag, Divider, Timeline, Card, Row, Col, Button, Space } from 'antd'
import {
  UserOutlined,
  FileTextOutlined,
  ClockCircleOutlined,
  TeamOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ArrowLeftOutlined,
  CalendarOutlined,
  BankOutlined,
} from '@ant-design/icons'

const { Title, Paragraph, Text } = Typography

// 冲突案例接口
interface ConflictCase {
  id: string
  caseId: string
  caseName: string
  caseNo?: string
  conflictType: string
  riskLevel: string
  description: string
  caseStatus: string
  clientId: string
  clientName: string
  opposingParties: string[]
  conflictDetails: string
  createdAt: string
  lawyerName?: string
  lawyerId?: string
}

// 冲突案例详情属性
interface ConflictCaseDetailProps {
  conflictCase: ConflictCase
  visible: boolean
  onClose: () => void
  onNavigateCase?: (caseId: string) => void
}

// 获取风险等级颜色
const getRiskLevelColor = (level: string) => {
  switch (level.toUpperCase()) {
    case 'CRITICAL':
      return '#dc3545'
    case 'HIGH':
      return '#fd7e14'
    case 'MEDIUM':
      return '#ffc107'
    case 'LOW':
      return '#28a745'
    case 'MINIMAL':
      return '#17a2b8'
    default:
      return '#6c757d'
  }
}

// 获取风险等级标签
const getRiskLevelTag = (level: string) => {
  const color = getRiskLevelColor(level)
  return (
    <Tag color={color} style={{ color: 'white', fontWeight: 'bold' }}>
      {level.toUpperCase()}
    </Tag>
  )
}

// 获取冲突类型图标
const getConflictTypeIcon = (type: string) => {
  switch (type) {
    case '代理冲突':
      return <UserOutlined style={{ color: '#fd7e14' }} />
    case '当事人冲突':
      return <TeamOutlined style={{ color: '#dc3545' }} />
    case '利益关联冲突':
      return <ExclamationCircleOutlined style={{ color: '#ffc107' }} />
    case '商业竞争冲突':
      return <BankOutlined style={{ color: '#17a2b8' }} />
    default:
      return <ExclamationCircleOutlined style={{ color: '#6c757d' }} />
  }
}

// 获取冲突类型描述
const getConflictTypeDescription = (type: string) => {
  switch (type) {
    case '代理冲突':
      return '同一律师代理了存在利益对立的客户案件，可能影响律师的独立判断和代理义务。'
    case '当事人冲突':
      return '新案件的当事人与律师代理的历史案件存在直接利益冲突关系。'
    case '利益关联冲突':
      return '通过公司关联关系或间接利益影响发现的潜在冲突。'
    case '商业竞争冲突':
      return '同一律师代理了同行业竞争企业的案件，可能产生商业利益冲突。'
    default:
      return '需要进一步评估的潜在利益冲突情况。'
  }
}

// 获取风险等级描述
const getRiskLevelDescription = (level: string) => {
  switch (level.toUpperCase()) {
    case 'CRITICAL':
      return '严重冲突：必须立即停止代理，可能违反职业操守规范。'
    case 'HIGH':
      return '高风险冲突：建议谨慎评估，可能需要合规部门审批。'
    case 'MEDIUM':
      return '中等风险：需要注意潜在影响，建议进行风险评估。'
    case 'LOW':
      return '低风险：冲突影响有限，可以继续代理但需保持警惕。'
    case 'MINIMAL':
      return '轻微风险：冲突影响很小，可以正常处理。'
    default:
      return '需要进一步评估风险等级。'
  }
}

// 冲突案例详情组件
const ConflictCaseDetail: React.FC<ConflictCaseDetailProps> = ({
  conflictCase,
  visible,
  onClose,
  onNavigateCase,
}) => {
  // 如果冲突案例为空，显示加载或错误状态
  if (!conflictCase) {
    return (
      <Modal
        title='冲突案例详情'
        open={visible}
        onCancel={onClose}
        width={800}
        footer={[
          <Button key='back' icon={<ArrowLeftOutlined />} onClick={onClose}>
            返回
          </Button>,
        ]}
      >
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <ExclamationCircleOutlined style={{ fontSize: '64px', color: '#d9d9d9' }} />
          <div style={{ marginTop: '16px' }}>
            <Text type='secondary'>冲突案例数据不存在或已被删除</Text>
          </div>
        </div>
      </Modal>
    )
  }

  // 格式化时间
  const formatTime = (timeString: string) => {
    if (!timeString) {
      return '未知时间'
    }
    try {
      return new Date(timeString).toLocaleString('zh-CN')
    } catch (error) {
      return '时间格式错误'
    }
  }

  return (
    <Modal
      title={
        <Space>
          {getConflictTypeIcon(conflictCase.conflictType || '')}
          <span>冲突案例详情</span>
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={800}
      footer={[
        <Button key='back' icon={<ArrowLeftOutlined />} onClick={onClose}>
          返回
        </Button>,
        onNavigateCase && (
          <Button key='view' type='primary' onClick={() => onNavigateCase(conflictCase.caseId)}>
            查看案件详情
          </Button>
        ),
      ]}
    >
      <div style={{ padding: '20px 0' }}>
        {/* 基本信息 */}
        <Card title='基本信息' style={{ marginBottom: '20px' }}>
          <Row gutter={[16, 16]}>
            <Col span={12}>
              <div>
                <Text type='secondary'>案件名称</Text>
                <div style={{ fontSize: '16px', fontWeight: 'bold', marginTop: '4px' }}>
                  {conflictCase.caseName || '未知案件'}
                </div>
              </div>
            </Col>
            <Col span={6}>
              <div>
                <Text type='secondary'>案件编号</Text>
                <div style={{ marginTop: '4px' }}>{conflictCase.caseNo || '未设置'}</div>
              </div>
            </Col>
            <Col span={6}>
              <div>
                <Text type='secondary'>案件状态</Text>
                <div style={{ marginTop: '4px' }}>
                  <Tag color='green'>{conflictCase.caseStatus || '未知状态'}</Tag>
                </div>
              </div>
            </Col>
          </Row>
        </Card>

        {/* 冲突信息 */}
        <Card title='冲突信息' style={{ marginBottom: '20px' }}>
          <Row gutter={[16, 16]}>
            <Col span={8}>
              <div>
                <Text type='secondary'>冲突类型</Text>
                <div style={{ display: 'flex', alignItems: 'center', marginTop: '4px' }}>
                  {getConflictTypeIcon(conflictCase.conflictType || '')}
                  <span style={{ marginLeft: '8px' }}>
                    {conflictCase.conflictType || '未知冲突'}
                  </span>
                </div>
              </div>
            </Col>
            <Col span={8}>
              <div>
                <Text type='secondary'>风险等级</Text>
                <div style={{ marginTop: '4px' }}>
                  {getRiskLevelTag(conflictCase.riskLevel || 'LOW')}
                </div>
              </div>
            </Col>
            <Col span={8}>
              <div>
                <Text type='secondary'>创建时间</Text>
                <div style={{ marginTop: '4px' }}>
                  <ClockCircleOutlined style={{ marginRight: '4px' }} />
                  {formatTime(conflictCase.createdAt || '')}
                </div>
              </div>
            </Col>
          </Row>

          <Divider />

          <Row gutter={[16, 16]}>
            <Col span={24}>
              <div>
                <Text type='secondary'>客户信息</Text>
                <div style={{ marginTop: '4px' }}>
                  <TeamOutlined style={{ marginRight: '4px', color: '#1890ff' }} />
                  <Text strong>{conflictCase.clientName || '未知客户'}</Text>
                </div>
              </div>
            </Col>
          </Row>

          {conflictCase.lawyerName && (
            <Row gutter={[16, 16]}>
              <Col span={24}>
                <div>
                  <Text type='secondary'>代理律师</Text>
                  <div style={{ marginTop: '4px' }}>
                    <UserOutlined style={{ marginRight: '4px', color: '#52c41a' }} />
                    <Text>{conflictCase.lawyerName || '未知律师'}</Text>
                  </div>
                </div>
              </Col>
            </Row>
          )}
        </Card>

        {/* 详细说明 */}
        <Card title='详细说明' style={{ marginBottom: '20px' }}>
          <div style={{ marginBottom: '16px' }}>
            <Title level={5} style={{ marginBottom: '8px' }}>
              冲突类型说明
            </Title>
            <Paragraph style={{ color: '#374151', fontSize: '14px' }}>
              {getConflictTypeDescription(conflictCase.conflictType || '')}
            </Paragraph>
          </div>

          <div style={{ marginBottom: '16px' }}>
            <Title level={5} style={{ marginBottom: '8px' }}>
              案件描述
            </Title>
            <Paragraph>{conflictCase.description || '无描述'}</Paragraph>
          </div>

          {conflictCase.conflictDetails && (
            <div>
              <Title level={5} style={{ marginBottom: '8px' }}>
                冲突详情
              </Title>
              <Paragraph>{conflictCase.conflictDetails || '无详情'}</Paragraph>
            </div>
          )}
        </Card>

        {/* 对方当事人信息 */}
        {conflictCase.opposingParties && conflictCase.opposingParties.length > 0 && (
          <Card title='对方当事人' style={{ marginBottom: '20px' }}>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
              {(conflictCase.opposingParties || []).map((party, index) => (
                <Tag key={index} color='blue' style={{ marginBottom: '8px' }}>
                  {party || '未知当事人'}
                </Tag>
              ))}
            </div>
          </Card>
        )}

        {/* 风险评估 */}
        <Card title='风险评估' style={{ marginBottom: '20px' }}>
          <div style={{ display: 'flex', alignItems: 'center', marginBottom: '12px' }}>
            <ExclamationCircleOutlined
              style={{
                marginRight: '8px',
                color: getRiskLevelColor(conflictCase.riskLevel || 'LOW'),
              }}
            />
            <Text style={{ fontWeight: 'bold' }}>风险等级评估</Text>
          </div>
          <Paragraph style={{ color: '#666' }}>
            {getRiskLevelDescription(conflictCase.riskLevel || 'LOW')}
          </Paragraph>
        </Card>

        {/* 处理建议 */}
        <Card title='处理建议'>
          <Timeline>
            <Timeline.Item color='blue' dot={<CheckCircleOutlined style={{ fontSize: '16px' }} />}>
              <Text strong>立即评估</Text>
              <div style={{ marginTop: '4px' }}>建议立即评估该冲突对当前案件代理的影响程度</div>
            </Timeline.Item>
            <Timeline.Item
              color='orange'
              dot={<ExclamationCircleOutlined style={{ fontSize: '16px' }} />}
            >
              <Text strong>风险控制</Text>
              <div style={{ marginTop: '4px' }}>制定相应的风险控制措施，必要时寻求合规部门意见</div>
            </Timeline.Item>
            <Timeline.Item color='green' dot={<CheckCircleOutlined style={{ fontSize: '16px' }} />}>
              <Text strong>记录存档</Text>
              <div style={{ marginTop: '4px' }}>将冲突检查结果和处理决定记录存档，以备后续查证</div>
            </Timeline.Item>
          </Timeline>
        </Card>
      </div>
    </Modal>
  )
}

export default ConflictCaseDetail
