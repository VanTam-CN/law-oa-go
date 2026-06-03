import React, { useState, useEffect } from 'react'
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Timeline,
  Divider,
  Tabs,
  Steps,
  List,
  Avatar,
  message,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  Popconfirm,
} from 'antd'
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  FileTextOutlined,
  UserOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  DollarCircleOutlined,
  TeamOutlined,
  PhoneOutlined,
  MailOutlined,
  EnvironmentOutlined,
  CommentOutlined,
  PaperClipOutlined,
} from '@ant-design/icons'
import { useNavigate, useParams } from 'react-router'
import dayjs from 'dayjs'
import { caseAPI } from '@/services/lawfirm'
import { caseService } from '@/services/case'
import { deadlineAPI } from '@/services/deadline'
import FeeCalculator from '@/components/fee/FeeCalculator'

const { TextArea } = Input
const { Option } = Select

interface Case {
  caseId: number
  caseNo: string
  caseName: string
  caseType: string
  clientName: string
  clientPhone: string
  clientEmail: string
  clientAddress: string
  lawyerName: string
  lawyerPhone: string
  lawyerEmail: string
  status: string
  description: string
  createTime: string
  updateTime: string
  expectedAmount: number
  actualAmount: number
  principalInfo?: string
  opponentInfo?: string
}

interface CaseDocument {
  id: number
  name: string
  type: string
  size: string
  uploadTime: string
  uploader: string
}

interface CaseTimeline {
  id: number
  time: string
  event: string
  user: string
  description: string
}

const CaseDetail: React.FC = () => {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [loading, setLoading] = useState(true)
  const [caseDetail, setCaseDetail] = useState<Case | null>(null)
  const [documents, setDocuments] = useState<CaseDocument[]>([])
  const [timeline, setTimeline] = useState<CaseTimeline[]>([])
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    if (id) {
      fetchCaseDetail()
    }
  }, [id])

  const fetchCaseDetail = async () => {
    setLoading(true)
    try {
      // 直接使用案件详情API
      const caseResponse = await caseAPI.getById(Number(id))

      if (caseResponse) {
        // 将API返回的数据转换为详情页面需要的格式
        const caseDetail: Case = {
          caseId: caseResponse.id || 0,
          caseNo: '', // 后端暂无案件编号字段
          caseName: caseResponse.title || '',
          caseType: caseResponse.caseType || '',
          clientName: caseResponse.clientName || '',
          clientPhone: caseResponse.client?.phone || '',
          clientEmail: caseResponse.client?.email || '',
          clientAddress: caseResponse.client?.address || '',
          lawyerName: caseResponse.lawyerName || '',
          lawyerPhone: caseResponse.lawyer?.phone || '',
          lawyerEmail: caseResponse.lawyer?.email || '',
          status: caseResponse.status || '',
          description: caseResponse.description || '',
          createTime: caseResponse.createdAt
            ? dayjs(caseResponse.createdAt).format('YYYY-MM-DD HH:mm:ss')
            : '',
          updateTime: caseResponse.updatedAt
            ? dayjs(caseResponse.updatedAt).format('YYYY-MM-DD HH:mm:ss')
            : '',
          expectedAmount: 0, // 后端暂无此字段
          actualAmount: 0,
          principalInfo: '', // 后端暂无此字段
          opponentInfo: '', // 后端暂无此字段
        }

        setCaseDetail(caseDetail)
        // 暂时使用空数组的文档和时间线，后续可以添加相应的API
        setDocuments([])
        setTimeline([])
      } else {
        message.error('未找到该案件')
      }
    } catch (error) {
      console.error('获取案件详情失败:', error)
      message.error('获取案件详情失败')
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = () => {
    if (caseDetail) {
      form.setFieldsValue(caseDetail)
      setEditModalVisible(true)
    }
  }

  const handleDelete = async () => {
    if (!caseDetail) {
      return
    }

    try {
      await caseService.deleteCase(caseDetail.caseId)
      message.success('删除成功')
      navigate('/case')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleUpdate = async (values: any) => {
    if (!caseDetail) {
      return
    }

    try {
      await caseService.updateCase(caseDetail.caseId, values)
      await fetchCaseDetail()
      setEditModalVisible(false)
      message.success('更新成功')
    } catch (error) {
      message.error('更新失败')
    }
  }

  const getStatusBadge = (status: string) => {
    const statusMap: Record<string, { text: string; color: string }> = {
      lead: { text: '线索', color: 'default' },
      risk_review: { text: '风控审查', color: 'orange' },
      signed: { text: '已签约', color: 'blue' },
      in_progress: { text: '办案中', color: 'processing' },
      trial: { text: '庭审中', color: 'purple' },
      closed: { text: '已结案', color: 'success' },
      archived: { text: '已归档', color: 'default' },
      '0': { text: '未开始', color: 'default' },
      '1': { text: '进行中', color: 'processing' },
      '2': { text: '已结案', color: 'success' },
      '3': { text: '已归档', color: 'default' },
      pending: { text: '待处理', color: 'orange' },
      active: { text: '进行中', color: 'green' },
      suspended: { text: '已暂停', color: 'red' },
    }
    const config = statusMap[status] || { text: '未知', color: 'default' }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  // 7阶段状态定义
  const caseStages = [
    { key: 'lead', title: '线索', description: '案件线索收集与初步评估' },
    { key: 'risk_review', title: '风控审查', description: '利益冲突排查与风险评估' },
    { key: 'signed', title: '已签约', description: '委托合同签署完成' },
    { key: 'in_progress', title: '办案中', description: '案件实质性办理阶段' },
    { key: 'trial', title: '庭审中', description: '开庭审理阶段' },
    { key: 'closed', title: '已结案', description: '案件审结' },
    { key: 'archived', title: '已归档', description: '案件材料归档保存' },
  ]

  const currentStageIndex = caseStages.findIndex((s) => s.key === caseDetail?.status)

  const handleTransitionStage = async (targetStage: string) => {
    try {
      await caseService.updateStatus(Number(id), targetStage)
      message.success(`案件状态已更新为: ${caseStages.find((s) => s.key === targetStage)?.title}`)
      fetchCaseDetail()
    } catch {
      message.error('状态更新失败')
    }
  }

  const [activeTabKey, setActiveTabKey] = useState('basic')

  // 时效管理状态
  const [deadlineTypes, setDeadlineTypes] = useState<Array<{ type: string; name: string; description: string; default_days: number; category: string }>>([])
  const [caseDeadlines, setCaseDeadlines] = useState<Array<{ id: number; deadline_type: string; type_name: string; start_date: string; end_date: string; remaining_days: number; status: string }>>([])
  const [deadlineLoading, setDeadlineLoading] = useState(false)

  const fetchDeadlineData = async () => {
    setDeadlineLoading(true)
    try {
      const [types, deadlines] = await Promise.all([
        deadlineAPI.getTypes(),
        deadlineAPI.getCaseDeadlines(Number(id)),
      ])
      setDeadlineTypes(types || [])
      setCaseDeadlines(deadlines || [])
    } catch {
      message.error('获取时效数据失败')
    } finally {
      setDeadlineLoading(false)
    }
  }

  useEffect(() => {
    if (activeTabKey === 'deadlines' && id) {
      fetchDeadlineData()
    }
  }, [activeTabKey, id])

  const getCaseTypeTag = (type: string) => {
    const typeMap = {
      CIVIL: { text: '民事案件', color: 'blue' },
      COMMERCIAL: { text: '商事案件', color: 'orange' },
      CRIMINAL: { text: '刑事案件', color: 'red' },
      ADMINISTRATIVE: { text: '行政案件', color: 'purple' },
    }
    const config = typeMap[type as keyof typeof typeMap] || { text: '其他', color: 'default' }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const tabItems = [
    {
      key: 'basic',
      label: '基本信息',
      children: (
        <Card title='案件信息' loading={loading}>
          <Descriptions bordered column={2}>
            <Descriptions.Item label='案件编号'>
              <Space>
                <FileTextOutlined />
                {caseDetail?.caseNo}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='案件名称'>{caseDetail?.caseName}</Descriptions.Item>
            <Descriptions.Item label='案件类型'>
              {getCaseTypeTag(caseDetail?.caseType || '')}
            </Descriptions.Item>
            <Descriptions.Item label='状态'>
              {getStatusBadge(caseDetail?.status || '')}
            </Descriptions.Item>
            <Descriptions.Item label='预计金额'>
              <Space>
                <DollarCircleOutlined />¥{caseDetail?.expectedAmount?.toLocaleString()}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='实际金额'>
              <Space>
                <DollarCircleOutlined />¥{caseDetail?.actualAmount?.toLocaleString()}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='创建时间' span={2}>
              <Space>
                <CalendarOutlined />
                {caseDetail?.createTime}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='更新时间' span={2}>
              <Space>
                <ClockCircleOutlined />
                {caseDetail?.updateTime}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='案件描述' span={2}>
              {caseDetail?.description}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ),
    },
    {
      key: 'client',
      label: '客户信息',
      children: (
        <Card title='客户信息' loading={loading}>
          <Descriptions bordered column={2}>
            <Descriptions.Item label='客户姓名'>
              <Space>
                <UserOutlined />
                {caseDetail?.clientName}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='联系电话'>
              <Space>
                <PhoneOutlined />
                {caseDetail?.clientPhone}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='邮箱地址'>
              <Space>
                <MailOutlined />
                {caseDetail?.clientEmail}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='联系地址'>
              <Space>
                <EnvironmentOutlined />
                {caseDetail?.clientAddress}
              </Space>
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ),
    },
    {
      key: 'parties',
      label: '当事人信息',
      children: (
        <Card title='当事人信息' loading={loading}>
          <Descriptions bordered column={1}>
            <Descriptions.Item label='委托人信息'>
              <div style={{ whiteSpace: 'pre-wrap' }}>
                {caseDetail?.principalInfo || '暂无委托人信息'}
              </div>
            </Descriptions.Item>
            <Descriptions.Item label='对方当事人信息'>
              <div style={{ whiteSpace: 'pre-wrap' }}>
                {caseDetail?.opponentInfo || '暂无对方当事人信息'}
              </div>
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ),
    },
    {
      key: 'lawyer',
      label: '律师信息',
      children: (
        <Card title='律师信息' loading={loading}>
          <Descriptions bordered column={2}>
            <Descriptions.Item label='负责律师'>
              <Space>
                <TeamOutlined />
                {caseDetail?.lawyerName}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='联系电话'>
              <Space>
                <PhoneOutlined />
                {caseDetail?.lawyerPhone}
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label='邮箱地址'>
              <Space>
                <MailOutlined />
                {caseDetail?.lawyerEmail}
              </Space>
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ),
    },
    {
      key: 'documents',
      label: '案件文档',
      children: (
        <Card title='案件文档' loading={loading}>
          <List
            dataSource={documents}
            renderItem={(item) => (
              <List.Item
                actions={[
                  <Button type='link' icon={<PaperClipOutlined />}>
                    下载
                  </Button>,
                  <Button type='link' icon={<CommentOutlined />}>
                    查看
                  </Button>,
                ]}
              >
                <List.Item.Meta
                  avatar={<Avatar icon={<PaperClipOutlined />} />}
                  title={item.name}
                  description={
                    <Space>
                      <Tag>{item.type}</Tag>
                      <span>{item.size}</span>
                      <span>{item.uploadTime}</span>
                      <span>上传者: {item.uploader}</span>
                    </Space>
                  }
                />
              </List.Item>
            )}
          />
        </Card>
      ),
    },
    {
      key: 'fee-calculation',
      label: '费用测算',
      children: (
        <FeeCalculator
          caseType={caseDetail?.caseType || ''}
          expectedAmount={caseDetail?.expectedAmount || 0}
        />
      ),
    },
    {
      key: 'timeline',
      label: '案件进展',
      children: (
        <Card title='案件进展' loading={loading}>
          <Timeline>
            {timeline.map((item) => (
              <Timeline.Item key={item.id}>
                <div>
                  <div style={{ fontWeight: 'bold', marginBottom: 4 }}>{item.event}</div>
                  <div style={{ color: '#666', marginBottom: 4 }}>{item.description}</div>
                  <div style={{ color: '#999', fontSize: '12px' }}>
                    {item.time} · {item.user}
                  </div>
                </div>
              </Timeline.Item>
            ))}
          </Timeline>
        </Card>
      ),
    },
    {
      key: 'stage-flow',
      label: '状态流转',
      children: (
        <Card title='案件状态流转' loading={loading}>
          <div style={{ marginBottom: 24 }}>
            <Steps
              current={currentStageIndex >= 0 ? currentStageIndex : 0}
              items={caseStages.map((stage) => ({
                title: stage.title,
                description: stage.description,
              }))}
            />
          </div>
          <Divider orientation='left'>状态变更操作</Divider>
          <Space wrap style={{ marginBottom: 16 }}>
            {caseStages.map((stage, idx) => {
              const isCurrent = caseDetail?.status === stage.key
              const isPast = idx < currentStageIndex
              const isNext = idx === currentStageIndex + 1
              return (
                <Button
                  key={stage.key}
                  type={isCurrent ? 'primary' : 'default'}
                  disabled={isCurrent || isPast}
                  onClick={() => handleTransitionStage(stage.key)}
                  style={{ opacity: isPast ? 0.4 : 1 }}
                >
                  {isCurrent ? `${stage.title} (当前)` : isNext ? `→ ${stage.title}` : stage.title}
                </Button>
              )
            })}
          </Space>
          <div style={{ color: '#999', fontSize: 12 }}>
            当前状态: {getStatusBadge(caseDetail?.status || '')}
          </div>
        </Card>
      ),
    },
    {
      key: 'deadlines',
      label: '时效管理',
      children: (
        <Card title='诉讼时效管理' loading={deadlineLoading}>
          <div style={{ marginBottom: 24 }}>
            <h4>支持的期限类型</h4>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 8 }}>
              {deadlineTypes.map((dt) => (
                <Tag key={dt.type} color='blue'>
                  {dt.name}: {dt.description} (默认{dt.default_days}天)
                </Tag>
              ))}
            </div>
          </div>
          <Divider orientation='left'>案件期限</Divider>
          {caseDeadlines.length === 0 ? (
            <div style={{ textAlign: 'center', color: '#999', padding: 24 }}>
              暂无期限数据。请通过后端API添加案件期限。
            </div>
          ) : (
            <List
              dataSource={caseDeadlines}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    title={item.type_name}
                    description={`截止日期: ${item.end_date} | 剩余: ${item.remaining_days}天`}
                  />
                  <Tag color={item.remaining_days <= 7 ? 'red' : item.remaining_days <= 30 ? 'orange' : 'green'}>
                    {item.status}
                  </Tag>
                </List.Item>
              )}
            />
          )}
        </Card>
      ),
    },
  ]

  if (loading) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <div>加载中...</div>
      </div>
    )
  }

  if (!caseDetail) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <div>案件不存在</div>
      </div>
    )
  }

  return (
    <div className='case-detail'>
      <Card>
        <div className='detail-header'>
          <Space>
            <Button type='text' icon={<ArrowLeftOutlined />} onClick={() => navigate('/case')}>
              返回
            </Button>
            <h2>{caseDetail.caseName}</h2>
          </Space>
          <Space>
            <Button type='primary' icon={<EditOutlined />} onClick={handleEdit}>
              编辑
            </Button>
            <Popconfirm
              title='确定要删除这个案件吗？'
              onConfirm={handleDelete}
              okText='确定'
              cancelText='取消'
            >
              <Button danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          </Space>
        </div>

        <Tabs items={tabItems} activeKey={activeTabKey} onChange={setActiveTabKey} />
      </Card>

      <Modal
        title='编辑案件'
        open={editModalVisible}
        onCancel={() => setEditModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout='vertical' onFinish={handleUpdate}>
          <Form.Item
            label='案件名称'
            name='caseName'
            rules={[{ required: true, message: '请输入案件名称' }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            label='案件类型'
            name='caseType'
            rules={[{ required: true, message: '请选择案件类型' }]}
          >
            <Select>
              <Option value='CIVIL'>民事案件</Option>
              <Option value='COMMERCIAL'>商事案件</Option>
              <Option value='CRIMINAL'>刑事案件</Option>
              <Option value='ADMINISTRATIVE'>行政案件</Option>
            </Select>
          </Form.Item>

          <Form.Item label='状态' name='status' rules={[{ required: true, message: '请选择状态' }]}>
            <Select>
              <Option value='0'>未开始</Option>
              <Option value='1'>进行中</Option>
              <Option value='2'>已结案</Option>
              <Option value='3'>已归档</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label='预计金额'
            name='expectedAmount'
            rules={[{ required: true, message: '请输入预计金额' }]}
          >
            <Input type='number' addonBefore='¥' />
          </Form.Item>

          <Form.Item label='实际金额' name='actualAmount'>
            <Input type='number' addonBefore='¥' />
          </Form.Item>

          <Form.Item
            label='案件描述'
            name='description'
            rules={[{ required: true, message: '请输入案件描述' }]}
          >
            <TextArea rows={4} />
          </Form.Item>

          <Form.Item>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={() => setEditModalVisible(false)}>取消</Button>
              <Button type='primary' htmlType='submit'>
                更新
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default CaseDetail
