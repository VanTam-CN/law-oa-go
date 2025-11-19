import React, { useState, useEffect } from 'react'
import {
  Modal,
  Form,
  Input,
  Select,
  Button,
  Tabs,
  message,
  Card,
  Row,
  Col,
  Space,
  Divider,
  Alert,
  Badge,
  Tooltip,
} from 'antd'
import {
  PlusOutlined,
  FileTextOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  FolderOpenOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'
import { enhancedCaseAPI } from '@/services/enhancedCase'
import WaiverApplicationForm from './waiver/WaiverApplicationForm'
import MultiClientSelector from './case/MultiClientSelector'
import type { WaiverApplication } from '@/types/waiverApproval'

const { TextArea } = Input
const { Option } = Select
const { TabPane } = Tabs

interface EnhancedCaseWithWaiverProps {
  visible: boolean
  onCancel: () => void
  onSuccess: () => void
}

interface ConflictAlert {
  type: 'CONFLICT' | 'WARNING'
  message: string
  details: string
  severity: 'HIGH' | 'MEDIUM' | 'LOW'
}

const EnhancedCaseWithWaiver: React.FC<EnhancedCaseWithWaiverProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('basic')
  const [selectedClients, setSelectedClients] = useState<any[]>([])
  const [conflictAlerts, setConflictAlerts] = useState<ConflictAlert[]>([])
  const [waiverModalVisible, setWaiverModalVisible] = useState(false)
  const [currentCaseId, setCurrentCaseId] = useState<string>('')
  const [lawyers, setLawyers] = useState<any[]>([])
  const [caseTypes, setCaseTypes] = useState<any[]>([])
  const [billingMethods, setBillingMethods] = useState<any[]>([])

  useEffect(() => {
    if (visible) {
      loadInitialData()
      form.resetFields()
      setSelectedClients([])
      setConflictAlerts([])
    }
  }, [visible])

  const loadInitialData = async () => {
    try {
      // 模拟数据加载
      setLawyers([
        { lawyerId: 1, lawyerName: '张律师', position: '高级合伙人' },
        { lawyerId: 2, lawyerName: '李律师', position: '合伙人' },
        { lawyerId: 3, lawyerName: '王律师', position: '主办律师' },
      ])

      setCaseTypes([
        { id: 'CIVIL', name: '民事案件', code: 'CIVIL' },
        { id: 'COMMERCIAL', name: '商事案件', code: 'COMMERCIAL' },
        { id: 'CRIMINAL', name: '刑事案件', code: 'CRIMINAL' },
        { id: 'ADMINISTRATIVE', name: '行政案件', code: 'ADMINISTRATIVE' },
      ])

      setBillingMethods([
        { id: 'FIXED', name: '定额收费', code: 'FIXED' },
        { id: 'HOURLY', name: '按时收费', code: 'HOURLY' },
        { id: 'RISK', name: '风险代理', code: 'RISK' },
        { id: 'MIXED', name: '混合收费', code: 'MIXED' },
      ])
    } catch (error) {
      message.error('加载初始数据失败')
    }
  }

  // 冲突检测
  const performConflictCheck = async (clients: any[]) => {
    try {
      setLoading(true)

      // 模拟冲突检测逻辑
      const alerts: ConflictAlert[] = []

      clients.forEach((client) => {
        // 检查是否存在潜在冲突
        if (client.type === 'COMPANY' && client.businessNature === '银行') {
          alerts.push({
            type: 'CONFLICT',
            message: '银行客户冲突检测',
            details: '发现与银行客户的潜在利益冲突，需要进一步评估',
            severity: 'HIGH',
          })
        }

        if (client.contactInfo && client.contactInfo.includes('竞争对手')) {
          alerts.push({
            type: 'WARNING',
            message: '竞争对手关联',
            details: '客户与现有客户存在竞争关系，建议进行详细审查',
            severity: 'MEDIUM',
          })
        }
      })

      setConflictAlerts(alerts)

      if (alerts.some((alert) => alert.type === 'CONFLICT')) {
        message.warning('检测到潜在利益冲突，建议申请豁免审批')
      }
    } catch (error) {
      message.error('冲突检测失败')
    } finally {
      setLoading(false)
    }
  }

  // 处理客户选择变化
  const handleClientsChange = (clients: any[]) => {
    setSelectedClients(clients)
    if (clients.length > 0) {
      performConflictCheck(clients)
    } else {
      setConflictAlerts([])
    }
  }

  // 创建案例
  const handleCreateCase = async () => {
    try {
      setLoading(true)

      if (selectedClients.length === 0) {
        message.error('请至少选择一个客户')
        return
      }

      // 检查是否有未解决的冲突
      const hasUnresolvedConflicts = conflictAlerts.some((alert) => alert.type === 'CONFLICT')

      if (hasUnresolvedConflicts) {
        Modal.confirm({
          title: '发现潜在利益冲突',
          content: '检测到潜在利益冲突，建议先创建豁免申请。是否继续创建案例？',
          okText: '继续创建',
          cancelText: '先申请豁免',
          onOk: () => createCase(),
          onCancel: () => {
            setCurrentCaseId(`CASE_${Date.now()}`)
            setWaiverModalVisible(true)
          },
        })
      } else {
        await createCase()
      }
    } catch (error) {
      message.error('创建案例失败')
    } finally {
      setLoading(false)
    }
  }

  // 实际创建案例逻辑
  const createCase = async () => {
    const values = await form.validateFields()

    const caseData = {
      title: values.title,
      description: values.description || '',
      caseType: values.caseType,
      priority: values.priority || 'MEDIUM',
      startDate: values.startDate ? values.startDate.format('YYYY-MM-DD') : undefined,
      practiceArea: values.practiceArea || 'GENERAL',
      estimatedDuration: values.estimatedDuration,
      billingMethod: values.billingMethod,
      clientProfileIds: selectedClients.map((client) => client.clientId),
      clientRoles: selectedClients.reduce((acc, client) => {
        acc[client.clientId] = {
          role: client.role || 'PRIMARY',
          relationshipDescription: client.relationshipDescription,
        }
        return acc
      }, {}),
      lawyerId: values.lawyerId,
      assignedBy: 1,
      isMajorRisk: values.isMajorRisk || false,
    }

    await enhancedCaseAPI.createEnhancedCase(caseData)
    message.success('案例创建成功')
    onSuccess()
    onCancel()
  }

  // 豁免申请成功处理
  const handleWaiverSuccess = (waiverApplication: WaiverApplication) => {
    message.success('豁免申请创建成功')
    setWaiverModalVisible(false)
    // 可以在这里继续创建案例或者等待豁免批准
  }

  // 渲染冲突警报
  const renderConflictAlerts = () => {
    if (conflictAlerts.length === 0) {
      return null
    }

    return (
      <Alert
        message='冲突检测结果'
        description={
          <div>
            {conflictAlerts.map((alert, index) => (
              <div key={index} style={{ marginBottom: 8 }}>
                <Badge
                  color={
                    alert.severity === 'HIGH'
                      ? 'red'
                      : alert.severity === 'MEDIUM'
                        ? 'orange'
                        : 'yellow'
                  }
                  text={alert.message}
                />
                <div style={{ fontSize: '12px', color: '#666', marginTop: 4 }}>{alert.details}</div>
              </div>
            ))}
          </div>
        }
        type={conflictAlerts.some((a) => a.type === 'CONFLICT') ? 'warning' : 'info'}
        showIcon
        action={
          conflictAlerts.some((a) => a.type === 'CONFLICT') && (
            <Button
              size='small'
              type='primary'
              icon={<PlusOutlined />}
              onClick={() => {
                setCurrentCaseId(`CASE_${Date.now()}`)
                setWaiverModalVisible(true)
              }}
            >
              申请豁免
            </Button>
          )
        }
        style={{ marginBottom: 16 }}
      />
    )
  }

  // 渲染基础信息标签页
  const renderBasicInfo = () => (
    <Card title='基本信息' size='small'>
      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label='案件名称'
            name='title'
            rules={[{ required: true, message: '请输入案件名称' }]}
          >
            <Input placeholder='请输入案件名称' />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item
            label='案件类型'
            name='caseType'
            rules={[{ required: true, message: '请选择案件类型' }]}
          >
            <Select placeholder='请选择案件类型'>
              {caseTypes?.map((type) => (
                <Option key={type.id} value={type.code}>
                  {type.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item label='优先级' name='priority' initialValue='MEDIUM'>
            <Select placeholder='请选择优先级'>
              <Option value='HIGH'>高</Option>
              <Option value='MEDIUM'>中</Option>
              <Option value='LOW'>低</Option>
            </Select>
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='业务领域' name='practiceArea' initialValue='GENERAL'>
            <Select placeholder='请选择业务领域'>
              <Option value='GENERAL'>一般法律业务</Option>
              <Option value='CORPORATE'>公司法务</Option>
              <Option value='LITIGATION'>诉讼业务</Option>
              <Option value='COMPLIANCE'>合规业务</Option>
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label='收费方式'
            name='billingMethod'
            rules={[{ required: true, message: '请选择收费方式' }]}
          >
            <Select placeholder='请选择收费方式'>
              {billingMethods?.map((method) => (
                <Option key={method.id} value={method.code}>
                  {method.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='预计持续时间' name='estimatedDuration'>
            <Input placeholder='如：3个月' />
          </Form.Item>
        </Col>
      </Row>

      <Form.Item label='案件描述' name='description'>
        <TextArea rows={4} placeholder='请输入案件的详细描述' />
      </Form.Item>

      <Form.Item
        label='主办律师'
        name='lawyerId'
        rules={[{ required: true, message: '请选择主办律师' }]}
      >
        <Select placeholder='请选择主办律师'>
          {lawyers?.map((lawyer) => (
            <Option key={lawyer.lawyerId} value={lawyer.lawyerId}>
              {lawyer.lawyerName} - {lawyer.position}
            </Option>
          ))}
        </Select>
      </Form.Item>
    </Card>
  )

  return (
    <Modal
      title='创建增强案例'
      open={visible}
      onCancel={onCancel}
      width={1200}
      footer={[
        <Button key='cancel' onClick={onCancel}>
          取消
        </Button>,
        <Button key='submit' type='primary' loading={loading} onClick={handleCreateCase}>
          创建案例
        </Button>,
      ]}
    >
      <Form form={form} layout='vertical'>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane
            tab={
              <span>
                <FileTextOutlined />
                基本信息
              </span>
            }
            key='basic'
          >
            {renderBasicInfo()}
          </TabPane>

          <TabPane
            tab={
              <span>
                <UserOutlined />
                客户选择
                {selectedClients.length > 0 && (
                  <Badge count={selectedClients.length} style={{ marginLeft: 8 }} />
                )}
              </span>
            }
            key='clients'
          >
            <Card title='客户选择与配置' size='small'>
              <MultiClientSelector
                value={selectedClients}
                onChange={handleClientsChange}
                form={form}
                maxClients={10}
                showConflictConfig
                allowPrimaryOnly={false}
              />
            </Card>

            {renderConflictAlerts()}
          </TabPane>

          <TabPane
            tab={
              <span>
                <SafetyOutlined />
                冲突检测
                {conflictAlerts.length > 0 && (
                  <Badge
                    count={conflictAlerts.filter((a) => a.type === 'CONFLICT').length}
                    style={{ marginLeft: 8 }}
                  />
                )}
              </span>
            }
            key='conflict'
          >
            <Card title='冲突检测结果' size='small'>
              {conflictAlerts.length > 0 ? (
                <div>
                  {conflictAlerts.map((alert, index) => (
                    <Alert
                      key={index}
                      message={alert.message}
                      description={alert.details}
                      type={alert.type === 'CONFLICT' ? 'error' : 'warning'}
                      showIcon
                      style={{ marginBottom: 8 }}
                    />
                  ))}

                  {conflictAlerts.some((a) => a.type === 'CONFLICT') && (
                    <div style={{ marginTop: 16 }}>
                      <Alert
                        message='建议操作'
                        description='检测到利益冲突，建议先申请豁免审批，避免后续合规风险。'
                        type='info'
                        showIcon
                        action={
                          <Button
                            type='primary'
                            size='small'
                            icon={<PlusOutlined />}
                            onClick={() => {
                              setCurrentCaseId(`CASE_${Date.now()}`)
                              setWaiverModalVisible(true)
                            }}
                          >
                            申请豁免
                          </Button>
                        }
                      />
                    </div>
                  )}
                </div>
              ) : (
                <div style={{ textAlign: 'center', padding: 40 }}>
                  <CheckCircleOutlined style={{ fontSize: 48, color: '#52c41a' }} />
                  <div style={{ marginTop: 16 }}>
                    <Text type='secondary'>未检测到明显冲突</Text>
                  </div>
                  <div style={{ marginTop: 8 }}>
                    <Text type='secondary' style={{ fontSize: 12 }}>
                      建议定期进行冲突检测以确保合规性
                    </Text>
                  </div>
                </div>
              )}
            </Card>
          </TabPane>
        </Tabs>
      </Form>

      {/* 豁免申请模态框 */}
      <WaiverApplicationForm
        visible={waiverModalVisible}
        onCancel={() => setWaiverModalVisible(false)}
        onSuccess={handleWaiverSuccess}
        caseId={currentCaseId}
        caseTitle={form.getFieldValue('title') || '新建案例'}
      />
    </Modal>
  )
}

export default EnhancedCaseWithWaiver
