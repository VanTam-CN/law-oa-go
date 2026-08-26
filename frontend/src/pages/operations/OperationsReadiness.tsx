import React from 'react'
import dayjs from 'dayjs'
import {
  Alert,
  Button,
  Card,
  DatePicker,
  Form,
  Input,
  List,
  Progress,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd'
import {
  getOperationsReadinessSummary,
  OPERATIONS_READINESS_REQUIREMENTS,
  OperationsEvidenceRegistrationInput,
  OperationsEvidenceScope,
  registerOperationsEvidence,
  ServerOperationsReadinessSummary,
} from '@/services/operationsReadiness'
import { useAppStore } from '@/stores/useAppStore'
import { hasPermission } from '@/utils/accessControl'

const { Paragraph, Text, Title } = Typography

const OperationsReadiness: React.FC = () => {
  const { user } = useAppStore()
  const [summary, setSummary] = React.useState<ServerOperationsReadinessSummary | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [scope, setScope] = React.useState<OperationsEvidenceScope>('controlled_pilot')
  type OperationsEvidenceFormValues = Omit<OperationsEvidenceRegistrationInput, 'scope' | 'reviewedAt'> & {
    reviewedAt: dayjs.Dayjs
  }
  const [form] = Form.useForm<OperationsEvidenceFormValues>()
  const canRegister = hasPermission(user, 'operations:register')

  const loadSummary = React.useCallback(async (selectedScope: OperationsEvidenceScope) => {
    setLoading(true)
    setError(null)
    try {
      setSummary(await getOperationsReadinessSummary(selectedScope))
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : '读取运维证据失败')
      setSummary(null)
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void loadSummary(scope)
  }, [loadSummary, scope])

  const submitEvidence = async (values: OperationsEvidenceFormValues) => {
    setLoading(true)
    setError(null)
    try {
      await registerOperationsEvidence({
        ...values,
        scope,
        reviewedAt: values.reviewedAt.toISOString(),
      })
      form.resetFields()
      await loadSummary(scope)
    } catch (registerError) {
      setError(registerError instanceof Error ? registerError.message : '运维证据登记失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Space direction='vertical' size='large' style={{ width: '100%' }}>
      <Card>
        <Title level={3}>运维准备度</Title>
        <Paragraph>
          这个页面给律所主任和行政人员看：系统当前能用，不代表出事后能恢复。每一项都必须留下可复核的证据。
        </Paragraph>
        <Alert
          type='warning'
          showIcon
          message='健康检查通过不等于运维已就绪'
          description='健康检查只能说明当前服务可用。备份、恢复演练、故障负责人、升级和回滚必须逐项登记受控环境证据；本页面不读取健康检查结果。'
        />
        <div style={{ marginTop: 24 }}>
          <Progress
            percent={summary ? Math.round((summary.score / summary.maximumScore) * 100) : 0}
            status='exception'
            format={() =>
              summary
                ? `${summary.score}/${summary.maximumScore}（${summary.verifiedCount}/${summary.total} 项证据）`
                : '读取中'
            }
          />
          <Text type='secondary'>
            五项证据全齐时达到受控运维准备度 7/10；缺任一项仍按已验证项数计 0-5。剩余 3 分需 PR #12 的 production_external_evidence 生产门禁独立判定。
          </Text>
          <Paragraph type='secondary' style={{ marginBottom: 0 }}>
            历史登记全部保留。本表按复核时间取最新记录；复验时追加新登记。
          </Paragraph>
        </div>
        {error ? <Alert type='error' showIcon style={{ marginTop: 16 }} message={error} /> : null}
      </Card>

      <Card
        title={canRegister ? '受控证据登记' : '受控证据'}
        extra={<Select
          value={scope}
          onChange={setScope}
          style={{ width: 160 }}
          options={[
            { value: 'controlled_pilot', label: '受控试点' },
            { value: 'qa', label: 'QA 环境' },
          ]}
        />}
      >
        {canRegister ? (
          <Form form={form} layout='vertical' onFinish={submitEvidence}>
            <Space wrap size='large'>
              <Form.Item name='control' label='事项' rules={[{ required: true, message: '请选择事项' }]} style={{ minWidth: 220 }}>
                <Select options={OPERATIONS_READINESS_REQUIREMENTS.map((item) => ({ value: item.id, label: item.title }))} />
              </Form.Item>
              <Form.Item name='reviewedAt' label='复核时间' rules={[{ required: true, message: '请选择复核时间' }]} style={{ minWidth: 220 }}>
                <DatePicker showTime />
              </Form.Item>
            </Space>
            <Form.Item
              name='evidenceReference'
              label='证据编号或检索位置'
              rules={[{ required: true, message: '请输入可复核证据位置' }]}
              extra='只填写编号、工单或档案位置；不得填写密码、密钥、令牌或联系方式。'
            >
              <Input.TextArea rows={2} maxLength={1000} showCount />
            </Form.Item>
            <Form.Item name='notes' label='复核说明（可选）' extra='记录抽查范围和结论，不替代原始证据。'>
              <Input.TextArea rows={2} maxLength={1000} showCount />
            </Form.Item>
            <Button type='primary' htmlType='submit' loading={loading}>登记受控证据</Button>
          </Form>
        ) : (
          <Alert
            type='info'
            showIcon
            message='当前账号可复核证据'
            description='证据登记由律所主任或系统管理员执行。历史证据和当前缺口对所有运维准备度复核账号可见。'
          />
        )}
      </Card>

      <Card title='证据清单' loading={loading && !summary} data-testid='operations-readiness-requirements'>
        <Table
          dataSource={summary?.items.map((item) => ({
            ...item,
            requirement: OPERATIONS_READINESS_REQUIREMENTS.find((requirement) => requirement.id === item.control),
          })) || []}
          rowKey='control'
          loading={loading}
          pagination={false}
          columns={[
            {
              title: '事项',
              dataIndex: ['requirement', 'title'],
              render: (title: string, record) => (
                <Space direction='vertical' size='small'>
                  <Text strong>{title}</Text>
                  <Text type='secondary'>{record.requirement?.userQuestion}</Text>
                </Space>
              ),
            },
            {
              title: '当前状态',
              key: 'status',
              width: 120,
              render: (_, record) =>
                record.status === 'verified' ? <Tag color='green'>已有证据</Tag> : <Tag color='warning'>待补证据</Tag>,
            },
            {
              title: '可审计来源',
              key: 'source',
              render: (_, record) =>
                record.evidence ? (
                  <Space direction='vertical' size='small'>
                    <Text copyable>{record.evidence.evidenceReference}</Text>
                    <Text type='secondary'>复核人 #{record.evidence.reviewedBy} / {record.evidence.reviewedAt}</Text>
                  </Space>
                ) : (
                <List
                  size='small'
                  dataSource={record.requirement?.evidenceToKeep || []}
                  renderItem={(item) => <List.Item style={{ padding: '2px 0' }}>{item}</List.Item>}
                />
                ),
            },
            {
              title: '下一步动作',
              render: (_, record) =>
                record.status === 'verified' ? (
                  <Text type='secondary'>历史登记全部保留。本表按复核时间取最新记录；复验时追加新登记。</Text>
                ) : (
                  <Text>{record.requirement?.nextAction}</Text>
                ),
            },
          ]}
        />
      </Card>
    </Space>
  )
}

export default OperationsReadiness
