import React from 'react'
import { Alert, Button, Form, Input, Modal, Select } from 'antd'
import { createIntakeClient, type ClientSummary, type CreateIntakeClient } from '@/services/client'

export default function ClientCreateDialog({
  open,
  onCancel,
  onCreated,
}: {
  open: boolean
  onCancel: () => void
  onCreated: (client: ClientSummary) => void
}) {
  const [form] = Form.useForm<CreateIntakeClient>()
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState('')
  const busy = React.useRef(false)
  const mounted = React.useRef(true)
  React.useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
    }
  }, [])
  const key = React.useRef('')
  const type = Form.useWatch('type', form)
  const reset = () => {
    form.resetFields()
    key.current = ''
    setError('')
  }
  React.useEffect(() => {
    if (!open) reset()
    // The mounted form instance is stable; closing clears identity plaintext.
  }, [open])
  const submit = async () => {
    if (busy.current) return
    let values: CreateIntakeClient
    try {
      values = await form.validateFields()
    } catch {
      return
    }
    if (busy.current) return
    busy.current = true
    setSaving(true)
    setError('')
    key.current ||= `client-${crypto.randomUUID()}`
    try {
      const client = await createIntakeClient(values, key.current)
      if (!mounted.current) return
      if (!Number.isSafeInteger(client.id) || client.id <= 0 || !client.name)
        throw new Error('客户创建响应缺少编号，请保持原内容重试确认结果')
      reset()
      onCreated(client)
    } catch (failure) {
      setError(
        failure instanceof Error ? failure.message : '客户创建结果尚未确认，请保留当前内容重试',
      )
    } finally {
      busy.current = false
      setSaving(false)
    }
  }
  return (
    <Modal
      title='新增客户'
      open={open}
      onCancel={() => {
        if (!busy.current) {
          reset()
          onCancel()
        }
      }}
      onOk={() => void submit()}
      okText='保存客户'
      cancelText='取消'
      confirmLoading={saving}
      closable={!saving}
      maskClosable={!saving}
      keyboard={!saving}
      cancelButtonProps={{ disabled: saving }}
      destroyOnHidden
    >
      {error && (
        <Alert
          type='error'
          showIcon
          message={error}
          description='当前立案内容仍保留。响应不确定时请保持内容不变重试；若明确改为另一份创建请求，可选择下方按钮。'
        />
      )}
      <Form
        form={form}
        layout='vertical'
        disabled={saving}
        initialValues={{ type: '企业', identity_type: 'SOCIAL_CREDIT_CODE' }}
      >
        <Form.Item name='type' label='客户类型' rules={[{ required: true }]}>
          <Select
            options={[{ value: '企业' }, { value: '个人' }]}
            onChange={(value) =>
              form.setFieldsValue({
                identity_type: value === '企业' ? 'SOCIAL_CREDIT_CODE' : 'ID_CARD',
                identity_number: '',
              })
            }
          />
        </Form.Item>
        <Form.Item
          name='name'
          label='客户名称'
          rules={[{ required: true, whitespace: true, message: '请填写客户名称' }, { max: 100 }]}
        >
          <Input autoComplete='off' />
        </Form.Item>
        <Form.Item name='identity_type' label='身份类型' rules={[{ required: true }]}>
          <Select
            options={
              type === '个人'
                ? [
                    { value: 'ID_CARD', label: '身份证' },
                    { value: 'PASSPORT', label: '护照' },
                  ]
                : [{ value: 'SOCIAL_CREDIT_CODE', label: '统一社会信用代码' }]
            }
          />
        </Form.Item>
        <Form.Item
          name='identity_number'
          label='身份号码'
          rules={[{ required: true, whitespace: true, message: '请填写身份号码' }]}
        >
          <Input.Password autoComplete='off' />
        </Form.Item>
        <Form.Item
          name='phone'
          label='联系电话'
          rules={[{ max: 20 }, { pattern: /^[\d\s\-+()]*$/, message: '请填写有效联系电话' }]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name='email'
          label='电子邮箱'
          rules={[{ type: 'email', message: '请填写有效邮箱' }]}
        >
          <Input />
        </Form.Item>
        <Form.Item name='address' label='联系地址'>
          <Input />
        </Form.Item>
      </Form>
      {error && (
        <Button
          disabled={saving}
          onClick={() => {
            key.current = ''
            setError('')
          }}
        >
          已改为另一份创建请求
        </Button>
      )}
    </Modal>
  )
}
