import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  List,
  Button,
  Select,
  Input,
  Tag,
  Space,
  Modal,
  Form,
  InputNumber,
  Switch,
  Divider,
  Typography,
  Alert,
  Tooltip,
  message,
} from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  SearchOutlined,
  UserOutlined,
  SettingOutlined,
} from '@ant-design/icons'

const { Title, Text } = Typography
const { Option } = Select

// 客户角色定义
const CLIENT_ROLES = [
  { value: 'PRIMARY', label: '主要委托人', description: '案件的主要委托人' },
  { value: 'SECONDARY', label: '次要委托人', description: '案件的次要委托人' },
  { value: 'GUARANTOR', label: '担保人', description: '为案件提供担保的个人或机构' },
  { value: 'INTERESTED_PARTY', label: '利益相关方', description: '与案件有利益关系的第三方' },
  { value: 'REPRESENTATIVE', label: '代表人', description: '代表其他主体参与案件的当事人' },
  { value: 'JOINT_APPLICANT', label: '共同申请人', description: '共同提出申请的多方当事人' },
]

// 客户信息接口
interface ClientInfo {
  id: string
  name: string
  type: 'INDIVIDUAL' | 'COMPANY'
  category: string
  registrationNumber?: string
  contactInfo?: string
}

// 已选客户信息接口
interface SelectedClient {
  clientId: string
  clientInfo: ClientInfo
  role: string
  relationshipDescription?: string
  contactInfo?: string
  conflictCheckConfig?: ConflictCheckConfig
}

// 冲突检测配置接口
interface ConflictCheckConfig {
  enabled: boolean
  checkOnCreate: boolean
  searchYears?: number
  includeCorporateRelations?: boolean
  searchDepth?: 'STANDARD' | 'DEEP' | 'COMPREHENSIVE'
  autoWaiverIfPossible?: boolean
}

// 组件Props接口
interface MultiClientSelectorProps {
  value?: SelectedClient[]
  onChange?: (clients: SelectedClient[]) => void
  form?: any
  disabled?: boolean
  maxClients?: number
  showConflictConfig?: boolean
  allowPrimaryOnly?: boolean
}

const MultiClientSelector: React.FC<MultiClientSelectorProps> = ({
  value = [],
  onChange,
  form,
  disabled = false,
  maxClients = 10,
  showConflictConfig = true,
  allowPrimaryOnly = false,
}) => {
  const [clients, setClients] = useState<SelectedClient[]>(value)
  const [searchModalVisible, setSearchModalVisible] = useState(false)
  const [configModalVisible, setConfigModalVisible] = useState(false)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchResults, setSearchResults] = useState<ClientInfo[]>([])
  const [selectedClientId, setSelectedClientId] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [configForm] = Form.useForm()

  // 同步外部value到内部state
  useEffect(() => {
    setClients(value)
  }, [value])

  // 通知外部变化
  const notifyChange = useCallback(
    (newClients: SelectedClient[]) => {
      setClients(newClients)
      onChange?.(newClients)
    },
    [onChange],
  )

  // 模拟搜索客户
  const searchClients = useCallback(async (keyword: string) => {
    if (!keyword.trim()) {
      setSearchResults([])
      return
    }

    setLoading(true)
    try {
      // 模拟API调用 - 实际项目中替换为真实API
      const mockResults: ClientInfo[] = [
        {
          id: 'client_001',
          name: '张三科技有限公司',
          type: 'COMPANY',
          category: '一般企业',
          registrationNumber: '91110000123456789X',
          contactInfo: '13800138000',
        },
        {
          id: 'client_002',
          name: '李四',
          type: 'INDIVIDUAL',
          category: '个人客户',
          contactInfo: '13900139000',
        },
        {
          id: 'client_003',
          name: '王五集团有限公司',
          type: 'COMPANY',
          category: '大型企业',
          registrationNumber: '91110000987654321Y',
          contactInfo: '13700137000',
        },
      ].filter(
        (client) =>
          client.name.toLowerCase().includes(keyword.toLowerCase()) ||
          client.registrationNumber?.includes(keyword),
      )

      setSearchResults(mockResults)
    } catch (error) {
      console.error('搜索客户失败:', error)
      message.error('搜索客户失败，请重试')
    } finally {
      setLoading(false)
    }
  }, [])

  // 处理搜索
  const handleSearch = useCallback(
    (keyword: string) => {
      setSearchKeyword(keyword)
      searchClients(keyword)
    },
    [searchClients],
  )

  // 添加客户
  const addClient = useCallback(
    (clientInfo: ClientInfo) => {
      if (clients.length >= maxClients) {
        message.warning(`最多只能添加${maxClients}个客户`)
        return
      }

      // 检查是否已存在
      if (clients.some((c) => c.clientId === clientInfo.id)) {
        message.warning('该客户已经添加')
        return
      }

      // 检查主要委托人限制
      if (allowPrimaryOnly && clients.some((c) => c.role === 'PRIMARY')) {
        message.warning('只能有一个主要委托人')
        return
      }

      // 确定默认角色
      let defaultRole = 'SECONDARY'
      if (clients.length === 0 || !clients.some((c) => c.role === 'PRIMARY')) {
        defaultRole = 'PRIMARY'
      }

      const newClient: SelectedClient = {
        clientId: clientInfo.id,
        clientInfo,
        role: defaultRole,
        relationshipDescription: '',
        contactInfo: clientInfo.contactInfo,
        conflictCheckConfig: {
          enabled: true,
          checkOnCreate: true,
          searchYears: clientInfo.type === 'COMPANY' ? 7 : 5,
          includeCorporateRelations: clientInfo.type === 'COMPANY',
          searchDepth: 'STANDARD',
          autoWaiverIfPossible: true,
        },
      }

      const newClients = [...clients, newClient]
      notifyChange(newClients)
      setSearchModalVisible(false)
      setSearchKeyword('')
      setSearchResults([])
      message.success('客户添加成功')
    },
    [clients, maxClients, allowPrimaryOnly, notifyChange],
  )

  // 移除客户
  const removeClient = useCallback(
    (clientId: string) => {
      const newClients = clients.filter((c) => c.clientId !== clientId)

      // 如果移除的是主要委托人，需要重新指定
      const hasPrimary = newClients.some((c) => c.role === 'PRIMARY')
      if (!hasPrimary && newClients.length > 0) {
        newClients[0].role = 'PRIMARY'
      }

      notifyChange(newClients)
      message.success('客户移除成功')
    },
    [clients, notifyChange],
  )

  // 更新客户角色
  const updateClientRole = useCallback(
    (clientId: string, role: string) => {
      // 检查主要委托人限制
      if (role === 'PRIMARY' && allowPrimaryOnly) {
        const currentPrimary = clients.find((c) => c.role === 'PRIMARY')
        if (currentPrimary && currentPrimary.clientId !== clientId) {
          message.warning('只能有一个主要委托人')
          return
        }
      }

      const newClients = clients.map((c) => (c.clientId === clientId ? { ...c, role } : c))
      notifyChange(newClients)
    },
    [clients, allowPrimaryOnly, notifyChange],
  )

  // 更新客户信息
  const updateClientInfo = useCallback(
    (clientId: string, updates: Partial<SelectedClient>) => {
      const newClients = clients.map((c) => (c.clientId === clientId ? { ...c, ...updates } : c))
      notifyChange(newClients)
    },
    [clients, notifyChange],
  )

  // 打开冲突检测配置
  const openConflictConfig = useCallback(
    (clientId: string) => {
      setSelectedClientId(clientId)
      const client = clients.find((c) => c.clientId === clientId)
      if (client?.conflictCheckConfig) {
        configForm.setFieldsValue(client.conflictCheckConfig)
      }
      setConfigModalVisible(true)
    },
    [clients, configForm],
  )

  // 保存冲突检测配置
  const saveConflictConfig = useCallback(() => {
    configForm.validateFields().then((values: ConflictCheckConfig) => {
      updateClientInfo(selectedClientId, { conflictCheckConfig: values })
      setConfigModalVisible(false)
      configForm.resetFields()
      message.success('冲突检测配置已保存')
    })
  }, [selectedClientId, updateClientInfo, configForm])

  // 获取角色标签颜色
  const getRoleTagColor = (role: string) => {
    const colors: Record<string, string> = {
      PRIMARY: 'red',
      SECONDARY: 'blue',
      GUARANTOR: 'orange',
      INTERESTED_PARTY: 'green',
      REPRESENTATIVE: 'purple',
      JOINT_APPLICANT: 'cyan',
    }
    return colors[role] || 'default'
  }

  // 获取角色名称
  const getRoleLabel = (role: string) => {
    const roleInfo = CLIENT_ROLES.find((r) => r.value === role)
    return roleInfo?.label || role
  }

  return (
    <div className='multi-client-selector'>
      <Card
        title={
          <Space>
            <UserOutlined />
            <span>客户选择</span>
            <Tag color='blue'>
              {clients.length}/{maxClients}
            </Tag>
          </Space>
        }
        extra={
          <Button
            type='primary'
            icon={<PlusOutlined />}
            onClick={() => setSearchModalVisible(true)}
            disabled={disabled || clients.length >= maxClients}
          >
            添加客户
          </Button>
        }
      >
        {clients.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <UserOutlined style={{ fontSize: '48px', color: '#ccc' }} />
            <div style={{ marginTop: '16px', color: '#999' }}>暂未添加客户，点击上方按钮添加</div>
          </div>
        ) : (
          <List
            dataSource={clients}
            renderItem={(client) => (
              <List.Item
                key={client.clientId}
                actions={[
                  showConflictConfig && (
                    <Tooltip title='冲突检测配置'>
                      <Button
                        type='text'
                        icon={<SettingOutlined />}
                        onClick={() => openConflictConfig(client.clientId)}
                        disabled={disabled}
                      />
                    </Tooltip>
                  ),
                  <Tooltip title='移除客户'>
                    <Button
                      type='text'
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => removeClient(client.clientId)}
                      disabled={disabled}
                    />
                  </Tooltip>,
                ].filter(Boolean)}
              >
                <List.Item.Meta
                  title={
                    <Space>
                      <span>{client.clientInfo.name}</span>
                      <Tag color={getRoleTagColor(client.role)}>{getRoleLabel(client.role)}</Tag>
                      <Tag color='default'>
                        {client.clientInfo.type === 'COMPANY' ? '企业' : '个人'}
                      </Tag>
                    </Space>
                  }
                  description={
                    <div>
                      <div style={{ marginBottom: '8px' }}>
                        <Text type='secondary'>{client.clientInfo.category}</Text>
                        {client.clientInfo.registrationNumber && (
                          <Text type='secondary' style={{ marginLeft: '16px' }}>
                            注册号: {client.clientInfo.registrationNumber}
                          </Text>
                        )}
                      </div>

                      <Space direction='vertical' style={{ width: '100%' }}>
                        <div>
                          <Text strong>角色: </Text>
                          <Select
                            value={client.role}
                            onChange={(role) => updateClientRole(client.clientId, role)}
                            style={{ width: 200 }}
                            disabled={disabled}
                          >
                            {CLIENT_ROLES.map((role) => (
                              <Option key={role.value} value={role.value}>
                                {role.label}
                              </Option>
                            ))}
                          </Select>
                        </div>

                        {client.relationshipDescription && (
                          <div>
                            <Text strong>关系描述: </Text>
                            <Text>{client.relationshipDescription}</Text>
                          </div>
                        )}

                        {client.contactInfo && (
                          <div>
                            <Text strong>联系方式: </Text>
                            <Text>{client.contactInfo}</Text>
                          </div>
                        )}
                      </Space>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        )}

        {allowPrimaryOnly && !clients.some((c) => c.role === 'PRIMARY') && clients.length > 0 && (
          <Alert
            message='警告'
            description='案件必须有一个主要委托人'
            type='warning'
            showIcon
            style={{ marginTop: '16px' }}
          />
        )}
      </Card>

      {/* 添加客户模态框 */}
      <Modal
        title='添加客户'
        open={searchModalVisible}
        onCancel={() => {
          setSearchModalVisible(false)
          setSearchKeyword('')
          setSearchResults([])
        }}
        footer={null}
        width={800}
      >
        <Space direction='vertical' style={{ width: '100%' }}>
          <Input.Search
            placeholder='输入客户名称或注册号搜索'
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            onSearch={handleSearch}
            style={{ width: '100%' }}
            loading={loading}
          />

          {searchResults.length > 0 && (
            <List
              dataSource={searchResults}
              renderItem={(client) => (
                <List.Item
                  actions={[
                    <Button
                      type='primary'
                      icon={<PlusOutlined />}
                      onClick={() => addClient(client)}
                    >
                      添加
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<UserOutlined />}
                    title={client.name}
                    description={
                      <Space direction='vertical' style={{ width: '100%' }}>
                        <div>
                          <Tag color='default'>{client.type === 'COMPANY' ? '企业' : '个人'}</Tag>
                          <Text type='secondary'>{client.category}</Text>
                        </div>
                        {client.registrationNumber && (
                          <div>
                            <Text type='secondary'>注册号: {client.registrationNumber}</Text>
                          </div>
                        )}
                        {client.contactInfo && (
                          <div>
                            <Text type='secondary'>联系方式: {client.contactInfo}</Text>
                          </div>
                        )}
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          )}

          {searchKeyword && searchResults.length === 0 && !loading && (
            <div style={{ textAlign: 'center', padding: '40px 0' }}>
              <SearchOutlined style={{ fontSize: '48px', color: '#ccc' }} />
              <div style={{ marginTop: '16px', color: '#999' }}>未找到匹配的客户</div>
            </div>
          )}
        </Space>
      </Modal>

      {/* 冲突检测配置模态框 */}
      {showConflictConfig && (
        <Modal
          title='冲突检测配置'
          open={configModalVisible}
          onOk={saveConflictConfig}
          onCancel={() => {
            setConfigModalVisible(false)
            configForm.resetFields()
          }}
          width={600}
        >
          <Form
            form={configForm}
            layout='vertical'
            initialValues={{
              enabled: true,
              checkOnCreate: true,
              searchYears: 5,
              includeCorporateRelations: false,
              searchDepth: 'STANDARD',
              autoWaiverIfPossible: true,
            }}
          >
            <Form.Item name='enabled' label='启用冲突检测' valuePropName='checked'>
              <Switch />
            </Form.Item>

            <Form.Item name='checkOnCreate' label='创建时自动检测' valuePropName='checked'>
              <Switch />
            </Form.Item>

            <Form.Item
              name='searchYears'
              label='搜索年限'
              rules={[{ required: true, message: '请输入搜索年限' }]}
            >
              <InputNumber min={1} max={20} placeholder='搜索年限' style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name='includeCorporateRelations'
              label='包含企业关联关系'
              valuePropName='checked'
            >
              <Switch />
            </Form.Item>

            <Form.Item
              name='searchDepth'
              label='搜索深度'
              rules={[{ required: true, message: '请选择搜索深度' }]}
            >
              <Select placeholder='选择搜索深度'>
                <Option value='STANDARD'>标准</Option>
                <Option value='DEEP'>深度</Option>
                <Option value='COMPREHENSIVE'>全面</Option>
              </Select>
            </Form.Item>

            <Form.Item
              name='autoWaiverIfPossible'
              label='自动豁免（如可能）'
              valuePropName='checked'
            >
              <Switch />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </div>
  )
}

export default MultiClientSelector
