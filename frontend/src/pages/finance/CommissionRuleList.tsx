import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Switch,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  Popconfirm,
  message,
  Typography,
  Divider,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { commissionRuleAPI, type CommissionRule } from '@/services/finance'

const { Title } = Typography

const roleMap: Record<string, { label: string; color: string }> = {
  source: { label: '案源', color: 'blue' },
  lawyer: { label: '主办律师', color: 'green' },
  assistant: { label: '协办律师', color: 'orange' },
}

const CommissionRuleList: React.FC = () => {
  const [rules, setRules] = useState<CommissionRule[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<CommissionRule | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm()

  const fetchRules = useCallback(async () => {
    setLoading(true)
    try {
      const response = await commissionRuleAPI.list()
      setRules(Array.isArray(response) ? response : [])
    } catch {
      message.error('加载分成规则失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchRules()
  }, [fetchRules])

  const handleAdd = () => {
    setEditingRule(null)
    form.resetFields()
    form.setFieldsValue({ active: true, priority: 0, min_amount: 0, max_amount: 0, base_rate: 0, performance_rate: 0 })
    setModalVisible(true)
  }

  const handleEdit = (record: CommissionRule) => {
    setEditingRule(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await commissionRuleAPI.delete(id)
      message.success('删除成功')
      fetchRules()
    } catch {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setSubmitting(true)
      if (editingRule) {
        await commissionRuleAPI.update(editingRule.id, values)
        message.success('更新成功')
      } else {
        await commissionRuleAPI.create(values)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchRules()
    } catch {
      message.error('操作失败')
    } finally {
      setSubmitting(false)
    }
  }

  const handleToggleActive = async (record: CommissionRule, checked: boolean) => {
    try {
      await commissionRuleAPI.update(record.id, { active: checked })
      message.success(checked ? '已启用' : '已禁用')
      fetchRules()
    } catch {
      message.error('更新状态失败')
    }
  }

  const columns: ColumnsType<CommissionRule> = [
    {
      title: '规则名称',
      dataIndex: 'name',
      key: 'name',
      width: 160,
    },
    {
      title: '适用角色',
      dataIndex: 'role',
      key: 'role',
      width: 120,
      render: (role: string) => {
        const config = roleMap[role] || { label: role, color: 'default' }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: '金额区间',
      key: 'amount_range',
      width: 200,
      render: (_, record) => (
        <span>
          {record.min_amount.toLocaleString()}
          {' - '}
          {record.max_amount > 0 ? record.max_amount.toLocaleString() : '不限'}
        </span>
      ),
    },
    {
      title: '基础比例',
      dataIndex: 'base_rate',
      key: 'base_rate',
      width: 100,
      render: (rate: number) => <Tag color='blue'>{rate}%</Tag>,
    },
    {
      title: '绩效比例',
      dataIndex: 'performance_rate',
      key: 'performance_rate',
      width: 100,
      render: (rate: number) => rate > 0 ? <Tag color='green'>{rate}%</Tag> : <span style={{ color: '#999' }}>-</span>,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      sorter: (a, b) => a.priority - b.priority,
    },
    {
      title: '状态',
      dataIndex: 'active',
      key: 'active',
      width: 80,
      render: (active: boolean, record: CommissionRule) => (
        <Switch checked={active} size='small' onChange={(checked) => handleToggleActive(record, checked)} />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space size='small'>
          <Button type='link' size='small' icon={<EditOutlined />} onClick={(e) => { e.stopPropagation(); handleEdit(record) }}>
            编辑
          </Button>
          <Popconfirm title='确定删除该规则？' onConfirm={(e) => { e?.stopPropagation(); handleDelete(record.id) }}>
            <Button type='link' size='small' danger icon={<DeleteOutlined />} onClick={(e) => e.stopPropagation()}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <Title level={4} style={{ margin: 0 }}>分成规则配置</Title>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchRules}>刷新</Button>
            <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>新增规则</Button>
          </Space>
        </div>

        <Divider style={{ margin: '0 0 16px' }} />

        <Table
          columns={columns}
          dataSource={rules}
          rowKey='id'
          loading={loading}
          scroll={{ x: 1000 }}
          pagination={{ pageSize: 20 }}
          size='middle'
        />
      </Card>

      <Modal
        title={editingRule ? '编辑分成规则' : '新增分成规则'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={handleSubmit}
        confirmLoading={submitting}
        width={560}
      >
        <Form form={form} layout='vertical'>
          <Form.Item label='规则名称' name='name' rules={[{ required: true, message: '请输入规则名称' }]}>
            <Input placeholder='例如: 案源-中额' />
          </Form.Item>

          <Form.Item label='适用角色' name='role' rules={[{ required: true, message: '请选择角色' }]}>
            <Select placeholder='请选择角色'>
              <Select.Option value='source'>案源</Select.Option>
              <Select.Option value='lawyer'>主办律师</Select.Option>
              <Select.Option value='assistant'>协办律师</Select.Option>
            </Select>
          </Form.Item>

          <Space size='large'>
            <Form.Item label='最小金额' name='min_amount' rules={[{ required: true }]}>
              <InputNumber min={0} step={1000} style={{ width: 160 }} addonBefore='¥' />
            </Form.Item>
            <Form.Item label='最大金额' name='max_amount' help='0表示不限'>
              <InputNumber min={0} step={1000} style={{ width: 160 }} addonBefore='¥' />
            </Form.Item>
          </Space>

          <Space size='large'>
            <Form.Item label='基础比例(%)' name='base_rate' rules={[{ required: true, message: '请输入基础比例' }]}>
              <InputNumber min={0} max={100} step={1} style={{ width: 120 }} addonAfter='%' />
            </Form.Item>
            <Form.Item label='绩效比例(%)' name='performance_rate'>
              <InputNumber min={0} max={100} step={1} style={{ width: 120 }} addonAfter='%' />
            </Form.Item>
          </Space>

          <Space size='large'>
            <Form.Item label='优先级' name='priority' help='越大越优先'>
              <InputNumber min={0} max={99} style={{ width: 120 }} />
            </Form.Item>
            <Form.Item label='启用' name='active' valuePropName='checked'>
              <Switch />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}

export default CommissionRuleList
