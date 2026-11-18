import React, { useState, useEffect } from 'react'
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Timeline,
  Divider,
  List,
  Avatar,
  message,
  Spin,
  Tabs,
  Badge,
} from 'antd'
import {
  ArrowLeftOutlined,
  FileTextOutlined,
  UserOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  TeamOutlined,
  PhoneOutlined,
  MailOutlined,
  PaperClipOutlined,
  BankOutlined as GavelOutlined,
  MedicineBoxOutlined,
  TrophyOutlined,
  EyeOutlined,
} from '@ant-design/icons'
import { useNavigate, useParams } from 'react-router'
import { lawyerService, Lawyer } from '@/services/lawyer'
import { caseService } from '@/services/case'

interface LawyerCase {
  id: number
  caseNo: string
  caseName: string
  caseType: string
  clientName: string
  status: string
  amount: number
  createTime: string
}

interface LawyerDocument {
  id: number
  name: string
  type: string
  size: string
  uploadTime: string
  uploader: string
}

interface LawyerTimeline {
  id: number
  time: string
  event: string
  description: string
}

const LawyerDetail: React.FC = () => {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [loading, setLoading] = useState(true)
  const [lawyerDetail, setLawyerDetail] = useState<Lawyer | null>(null)
  const [cases, setCases] = useState<LawyerCase[]>([])
  const [documents, setDocuments] = useState<LawyerDocument[]>([])
  const [timeline, setTimeline] = useState<LawyerTimeline[]>([])

  useEffect(() => {
    if (id) {
      fetchLawyerDetail()
    }
  }, [id])

  const fetchLawyerDetail = async () => {
    setLoading(true)
    try {
      const lawyerID = Number(id)
      const [lawyer, caseResponse] = await Promise.all([
        lawyerService.getLawyerDetail(lawyerID),
        caseService.getCaseList({ lawyer_id: lawyerID, page: 1, page_size: 20 }),
      ])

      setLawyerDetail(lawyer)
      setCases(normalizeCaseRows(caseResponse))
      setDocuments([])
      setTimeline(buildTimeline(lawyer))
    } catch (error) {
      message.error('获取律师详情失败')
    } finally {
      setLoading(false)
    }
  }

  const normalizeCaseRows = (response: any): LawyerCase[] => {
    const rows = Array.isArray(response)
      ? response
      : response?.cases || response?.list || response?.data?.cases || response?.data?.list || []

    return rows.map((item: any) => ({
      id: Number(item.id),
      caseNo: item.case_number || item.caseNumber || `CASE-${item.id}`,
      caseName: item.title || item.caseName || item.case_name || '未命名案件',
      caseType: item.case_type || item.caseType || item.type || 'OTHER',
      clientName: item.client?.name || item.clientName || item.client_name || '-',
      status: item.status || 'pending',
      amount: Number(item.amount || item.contract_amount || 0),
      createTime: item.created_at || item.createdAt || '',
    }))
  }

  const buildTimeline = (lawyer: Lawyer): LawyerTimeline[] => {
    const createdAt = lawyer.joinDate || ''
    return [
      {
        id: 1,
        time: createdAt,
        event: '账号创建',
        description: `${lawyer.name || lawyer.lawyerName} 的律师账号已在系统中创建。`,
      },
    ].filter((item) => item.time)
  }

  const getStatusBadge = (status: string) => {
    const statusMap = {
      active: { text: '在职', color: 'success' },
      inactive: { text: '离职', color: 'default' },
      on_leave: { text: '休假', color: 'warning' },
    }
    const config = statusMap[status as keyof typeof statusMap] || { text: '未知', color: 'default' }
    return <Badge status={config.color as any} text={config.text} />
  }

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

  const getSpecialtyTags = (specialties?: string | string[]) => {
    const specialtyMap: Record<string, { icon: React.ReactNode; color: string }> = {
      民事案件: { icon: <GavelOutlined />, color: 'blue' },
      刑事案件: { icon: <MedicineBoxOutlined />, color: 'red' },
      商事案件: { icon: <TrophyOutlined />, color: 'orange' },
      行政案件: { icon: <TrophyOutlined />, color: 'purple' },
      知识产权: { icon: <TrophyOutlined />, color: 'cyan' },
      劳动纠纷: { icon: <MedicineBoxOutlined />, color: 'green' },
    }

    const specialtyList = Array.isArray(specialties)
      ? specialties
      : (specialties || '综合法律服务')
          .split(/[、,，/]/)
          .map((item) => item.trim())
          .filter(Boolean)

    return specialtyList.map((spec) => {
      const config = specialtyMap[spec] || { icon: <GavelOutlined />, color: 'default' }
      return (
        <Tag key={spec} color={config.color} icon={config.icon}>
          {spec}
        </Tag>
      )
    })
  }

  if (loading) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Spin size='large' />
        <div style={{ marginTop: 16 }}>加载中...</div>
      </div>
    )
  }

  if (!lawyerDetail) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <div>律师不存在</div>
      </div>
    )
  }

  return (
    <div className='lawyer-detail'>
      <Card>
        <div
          className='detail-header'
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '24px',
          }}
        >
          <Space>
            <Button type='text' icon={<ArrowLeftOutlined />} onClick={() => navigate('/lawyer')}>
              返回
            </Button>
            <h2 style={{ margin: 0 }}>{lawyerDetail.name}</h2>
          </Space>
        </div>

        <Tabs
          defaultActiveKey='basic'
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: (
                <Card title='律师信息' loading={loading}>
                  <Descriptions bordered column={2}>
                    <Descriptions.Item label='姓名'>
                      <Space>
                        <UserOutlined />
                        {lawyerDetail.name}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='性别'>
                      <Tag color={lawyerDetail.gender === 'male' ? 'blue' : 'pink'}>
                        {lawyerDetail.gender === 'male' ? '男' : '女'}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label='联系电话'>
                      <Space>
                        <PhoneOutlined />
                        {lawyerDetail.phone}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='邮箱地址'>
                      <Space>
                        <MailOutlined />
                        {lawyerDetail.email}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='执业证号'>
                      {lawyerDetail.licenseNumber || lawyerDetail.licenseNo || '-'}
                    </Descriptions.Item>
                    <Descriptions.Item label='工作经验'>
                      <Space>
                        <ClockCircleOutlined />
                        {lawyerDetail.experience || 0}年
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='部门'>
                      <TeamOutlined />
                      {lawyerDetail.department}
                    </Descriptions.Item>
                    <Descriptions.Item label='职位'>{lawyerDetail.position}</Descriptions.Item>
                    <Descriptions.Item label='入职日期' span={2}>
                      <Space>
                        <CalendarOutlined />
                        {lawyerDetail.joinDate || '-'}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='状态' span={2}>
                      {getStatusBadge(lawyerDetail.status || 'active')}
                    </Descriptions.Item>
                    <Descriptions.Item label='专业领域' span={2}>
                      <Space size='small' wrap>
                        {getSpecialtyTags(lawyerDetail.specialty)}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='个人简介' span={2}>
                      {lawyerDetail.profile || '暂无个人简介'}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              ),
            },
            {
              key: 'cases',
              label: '负责案件',
              children: (
                <Card title='负责案件列表' loading={loading}>
                  <List
                    dataSource={cases}
                    locale={{ emptyText: '暂无负责案件' }}
                    renderItem={(item) => (
                      <List.Item
                        actions={[
                          <Button
                            type='link'
                            icon={<EyeOutlined />}
                            onClick={() => navigate(`/case/${item.id}`)}
                          >
                            查看
                          </Button>,
                        ]}
                      >
                        <List.Item.Meta
                          avatar={<Avatar icon={<FileTextOutlined />} />}
                          title={
                            <Space>
                              {item.caseNo}
                              {getCaseTypeTag(item.caseType)}
                            </Space>
                          }
                          description={
                            <div>
                              <div>{item.caseName}</div>
                              <div style={{ color: '#666', fontSize: '12px' }}>
                                客户: {item.clientName} · 金额: ¥{item.amount.toLocaleString()}
                              </div>
                            </div>
                          }
                        />
                      </List.Item>
                    )}
                  />
                </Card>
              ),
            },
            {
              key: 'documents',
              label: '相关文档',
              children: (
                <Card title='律师文档' loading={loading}>
                  <List
                    dataSource={documents}
                    locale={{ emptyText: '暂无关联文档' }}
                    renderItem={(item) => (
                      <List.Item
                        actions={[
                          <Button type='link' icon={<PaperClipOutlined />}>
                            下载
                          </Button>,
                          <Button type='link' icon={<EyeOutlined />}>
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
              key: 'timeline',
              label: '职业发展',
              children: (
                <Card title='职业发展时间线' loading={loading}>
                  <Timeline>
                    {timeline.map((item) => (
                      <Timeline.Item key={item.id}>
                        <div>
                          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>{item.event}</div>
                          <div style={{ color: '#666', marginBottom: 4 }}>{item.description}</div>
                          <div style={{ color: '#999', fontSize: '12px' }}>{item.time}</div>
                        </div>
                      </Timeline.Item>
                    ))}
                  </Timeline>
                </Card>
              ),
            },
          ]}
        />
      </Card>
    </div>
  )
}

export default LawyerDetail
