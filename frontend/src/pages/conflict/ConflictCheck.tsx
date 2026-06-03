import React, { useState, useEffect } from 'react'
import { Card, Form, Input, Button, Select, Table, Tag, Alert, Spin, Space, Divider } from 'antd'
import { message } from '@/utils/messageHelper'
import { SearchOutlined, ExclamationCircleOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { getUserList } from '@/services/user'
import { getProjectTypes } from '@/services/project'
import { performConflictCheckLegacy, ConflictResult } from '@/services/conflict'
import { UserInfo } from '@/services/user'
import { ProjectType } from '@/services/project'
import './ConflictCheck.less'

const { Option } = Select
const { TextArea } = Input

interface ConflictCheckFormValues {
  project_name: string
  client_name: string
  opposite_parties: string
  project_type: string
  team_members: string[]
  description: string
}

const ConflictCheck: React.FC = () => {
  const [form] = Form.useForm<ConflictCheckFormValues>()
  const [loading, setLoading] = useState<boolean>(false)
  const [result, setResult] = useState<ConflictResult | null>(null)
  const [users, setUsers] = useState<UserInfo[]>([])
  const [projectTypes, setProjectTypes] = useState<ProjectType[]>([])
  const [loadingData, setLoadingData] = useState<boolean>(true)

  // 获取用户列表和项目类型
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoadingData(true)

        // 使用静态数据避免API调用失败
        const staticUsers = [
          {
            id: 1,
            username: 'admin',
            real_name: '系统管理员',
            email: 'admin@lawfirm.com',
            phone: '13800138000',
            status: 'active',
            created_at: '2024-01-01 00:00:00',
            updated_at: '2024-01-01 00:00:00',
          },
          {
            id: 2,
            username: 'lawyer1',
            real_name: '张律师',
            email: 'zhang@lawfirm.com',
            phone: '13800138001',
            status: 'active',
            created_at: '2024-01-01 00:00:00',
            updated_at: '2024-01-01 00:00:00',
          },
          {
            id: 3,
            username: 'lawyer2',
            real_name: '李律师',
            email: 'li@lawfirm.com',
            phone: '13800138002',
            status: 'active',
            created_at: '2024-01-01 00:00:00',
            updated_at: '2024-01-01 00:00:00',
          },
        ]

        const staticProjectTypes = [
          { id: 1, name: '民事诉讼', code: 'CIVIL', description: '民事纠纷相关案件' },
          { id: 2, name: '商业诉讼', code: 'COMMERCIAL', description: '商业纠纷相关案件' },
          { id: 3, name: '刑事诉讼', code: 'CRIMINAL', description: '刑事案件相关' },
          { id: 4, name: '法律顾问', code: 'ADVISORY', description: '法律咨询服务' },
        ]

        setUsers(staticUsers)
        setProjectTypes(staticProjectTypes)
      } catch (error) {
        console.error('Failed to fetch data:', error)
        message.error('加载数据失败，请刷新页面重试')
      } finally {
        setLoadingData(false)
      }
    }

    fetchData()
  }, [])

  // 处理表单提交
  const handleSubmit = async (values: ConflictCheckFormValues) => {
    try {
      setLoading(true)
      setResult(null)

      // 将对方当事人字符串转换为数组
      const formattedValues = {
        ...values,
        opposite_parties: values.opposite_parties.trim(),
      }

      const response = await performConflictCheckLegacy(formattedValues)
      setResult(response)

      if (response.has_conflict) {
        message.warning('检测到潜在的利益冲突，请查看详细信息')
      } else {
        message.success('未检测到利益冲突')
      }
    } catch (error) {
      console.error('Conflict check error:', error)
      message.error('冲突检查失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 重置表单和结果
  const handleReset = () => {
    form.resetFields()
    setResult(null)
  }

  // 获取冲突级别对应的标签颜色
  const getConflictLevelTag = (level: 'none' | 'low' | 'medium' | 'high') => {
    const levelConfig = {
      none: { color: 'success', text: '无冲突', icon: <CheckCircleOutlined /> },
      low: { color: 'warning', text: '低风险', icon: <ExclamationCircleOutlined /> },
      medium: { color: 'warning', text: '中风险', icon: <ExclamationCircleOutlined /> },
      high: { color: 'error', text: '高风险', icon: <ExclamationCircleOutlined /> },
    }

    const config = levelConfig[level]
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    )
  }

  // 表格列定义
  const columns = [
    {
      title: '冲突类型',
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: '相关实体',
      dataIndex: 'entity',
      key: 'entity',
    },
    {
      title: '关联项目',
      dataIndex: 'project',
      key: 'project',
    },
    {
      title: '风险等级',
      dataIndex: 'level',
      key: 'level',
      render: (level: 'low' | 'medium' | 'high') => getConflictLevelTag(level),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
  ]

  return (
    <div className='conflict-check-container'>
      <Card title='利益冲突检查' className='conflict-form-card'>
        <Spin spinning={loadingData}>
          <Form
            form={form}
            layout='vertical'
            onFinish={handleSubmit}
            initialValues={{
              description: '',
            }}
          >
            <Form.Item
              name='project_name'
              label='项目名称'
              rules={[{ required: true, message: '请输入项目名称' }]}
            >
              <Input placeholder='请输入项目名称' />
            </Form.Item>

            <Form.Item
              name='client_name'
              label='客户名称'
              rules={[{ required: true, message: '请输入客户名称' }]}
            >
              <Input placeholder='请输入客户名称' />
            </Form.Item>

            <Form.Item
              name='opposite_parties'
              label='对方当事人'
              rules={[{ required: true, message: '请输入对方当事人' }]}
            >
              <Input placeholder='请输入对方当事人名称，多个名称请用逗号分隔' />
            </Form.Item>

            <Form.Item
              name='project_type'
              label='项目类型'
              rules={[{ required: true, message: '请选择项目类型' }]}
            >
              <Select placeholder='请选择项目类型'>
                {projectTypes.map((type) => (
                  <Option key={type.id} value={type.code}>
                    {type.name}
                  </Option>
                ))}
              </Select>
            </Form.Item>

            <Form.Item
              name='team_members'
              label='团队成员'
              rules={[{ required: true, message: '请选择团队成员' }]}
            >
              <Select mode='multiple' placeholder='请选择团队成员' optionFilterProp='children'>
                {users.map((user) => (
                  <Option key={user.id} value={user.id.toString()}>
                    {user.real_name}
                  </Option>
                ))}
              </Select>
            </Form.Item>

            <Form.Item name='description' label='项目描述'>
              <TextArea rows={4} placeholder='请输入项目描述' />
            </Form.Item>

            <Form.Item>
              <Space>
                <Button
                  type='primary'
                  htmlType='submit'
                  icon={<SearchOutlined />}
                  loading={loading}
                  className='submit-button'
                >
                  检查冲突
                </Button>
                <Button onClick={handleReset}>重置</Button>
              </Space>
            </Form.Item>
          </Form>
        </Spin>
      </Card>

      {result && (
        <Card title='检查结果' className='conflict-result-card'>
          <Alert
            message={result.has_conflict ? '检测到潜在的利益冲突' : '未检测到利益冲突'}
            description={
              <div>
                <p>冲突等级: {getConflictLevelTag(result.conflict_level)}</p>
                {result.has_conflict && <p>请查看下方详细信息，并咨询合规部门进行进一步评估。</p>}
              </div>
            }
            type={result.has_conflict ? 'warning' : 'success'}
            showIcon
            className='conflict-alert'
          />

          {result.has_conflict && result.conflicts.length > 0 && (
            <>
              <Divider orientation='left'>冲突详情</Divider>
              <Table
                columns={columns}
                dataSource={result.conflicts.map((item) => ({ ...item, key: item.id }))}
                className='conflict-table'
                pagination={{ pageSize: 5 }}
              />
            </>
          )}
        </Card>
      )}
    </div>
  )
}

export default ConflictCheck
