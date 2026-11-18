import React, { useState, useEffect } from 'react'
import {
  Card,
  Row,
  Col,
  Statistic,
  Progress,
  Table,
  Tag,
  Select,
  DatePicker,
  Button,
  Space,
  Tabs,
  List,
  Avatar,
  Typography,
  Tooltip,
} from 'antd'
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  TrophyOutlined,
  TeamOutlined,
  ClockCircleOutlined,
  DollarCircleOutlined,
  FileTextOutlined,
  CalendarOutlined,
  BarChartOutlined,
  PieChartOutlined,
  LineChartOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import moment from 'moment'

const { Option } = Select
const { RangePicker } = DatePicker
const { Text } = Typography

interface AnalyticsData {
  overview: {
    totalCases: number
    ongoingCases: number
    completedCases: number
    totalRevenue: number
    totalHours: number
    clientSatisfaction: number
  }
  caseStats: {
    byType: Array<{ name: string; value: number; color: string }>
    byStatus: Array<{ name: string; value: number; color: string }>
    monthlyTrend: Array<{ month: string; cases: number; revenue: number }>
  }
  lawyerStats: {
    topPerformers: Array<{
      name: string
      avatar: string
      cases: number
      hours: number
      revenue: number
      satisfaction: number
    }>
    workload: Array<{
      name: string
      current: number
      capacity: number
    }>
  }
  clientStats: {
    topClients: Array<{
      name: string
      cases: number
      revenue: number
      satisfaction: number
    }>
    retention: number
    acquisition: number
  }
}

const AnalyticsDashboard: React.FC = () => {
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [loading, setLoading] = useState(false)
  const [timeRange, setTimeRange] = useState('month')
  const [dateRange, setDateRange] = useState<[moment.Moment, moment.Moment]>([
    moment().subtract(30, 'days'),
    moment(),
  ])

  useEffect(() => {
    fetchAnalyticsData()
  }, [timeRange, dateRange])

  const fetchAnalyticsData = async () => {
    setLoading(true)
    try {
      // 模拟API调用
      await new Promise((resolve) => setTimeout(resolve, 1000))

      const mockData: AnalyticsData = {
        overview: {
          totalCases: 156,
          ongoingCases: 42,
          completedCases: 114,
          totalRevenue: 2850000,
          totalHours: 3240,
          clientSatisfaction: 4.6,
        },
        caseStats: {
          byType: [
            { name: '合同纠纷', value: 45, color: '#8884d8' },
            { name: '劳动争议', value: 32, color: '#82ca9d' },
            { name: '知识产权', value: 28, color: '#ffc658' },
            { name: '婚姻家庭', value: 24, color: '#ff7c7c' },
            { name: '刑事辩护', value: 18, color: '#8dd1e1' },
            { name: '其他', value: 9, color: '#d084d0' },
          ],
          byStatus: [
            { name: '进行中', value: 42, color: '#1890ff' },
            { name: '已完成', value: 84, color: '#52c41a' },
            { name: '已结案', value: 30, color: '#722ed1' },
          ],
          monthlyTrend: [
            { month: '1月', cases: 12, revenue: 280000 },
            { month: '2月', cases: 15, revenue: 320000 },
            { month: '3月', cases: 18, revenue: 380000 },
            { month: '4月', cases: 22, revenue: 450000 },
            { month: '5月', cases: 25, revenue: 520000 },
            { month: '6月', cases: 28, revenue: 580000 },
            { month: '7月', cases: 36, revenue: 720000 },
          ],
        },
        lawyerStats: {
          topPerformers: [
            {
              name: '张律师',
              avatar: 'Z',
              cases: 28,
              hours: 420,
              revenue: 680000,
              satisfaction: 4.8,
            },
            {
              name: '李律师',
              avatar: 'L',
              cases: 24,
              hours: 380,
              revenue: 590000,
              satisfaction: 4.7,
            },
            {
              name: '王律师',
              avatar: 'W',
              cases: 22,
              hours: 350,
              revenue: 520000,
              satisfaction: 4.5,
            },
            {
              name: '赵律师',
              avatar: 'Z',
              cases: 20,
              hours: 320,
              revenue: 480000,
              satisfaction: 4.6,
            },
            {
              name: '陈律师',
              avatar: 'C',
              cases: 18,
              hours: 290,
              revenue: 420000,
              satisfaction: 4.4,
            },
          ],
          workload: [
            { name: '张律师', current: 12, capacity: 15 },
            { name: '李律师', current: 10, capacity: 15 },
            { name: '王律师', current: 8, capacity: 15 },
            { name: '赵律师', current: 14, capacity: 15 },
            { name: '陈律师', current: 6, capacity: 15 },
          ],
        },
        clientStats: {
          topClients: [
            { name: '某科技有限公司', cases: 8, revenue: 180000, satisfaction: 4.9 },
            { name: '某制造企业', cases: 6, revenue: 150000, satisfaction: 4.7 },
            { name: '某投资公司', cases: 5, revenue: 120000, satisfaction: 4.8 },
            { name: '某贸易公司', cases: 4, revenue: 95000, satisfaction: 4.5 },
            { name: '某服务公司', cases: 3, revenue: 80000, satisfaction: 4.6 },
          ],
          retention: 85,
          acquisition: 12,
        },
      }

      setData(mockData)
    } catch (error) {
      console.error('获取数据失败:', error)
    }
    setLoading(false)
  }

  const exportReport = () => {
    // 模拟导出功能
    const reportData = {
      timeRange,
      dateRange: dateRange.map((d) => d.format('YYYY-MM-DD')),
      data,
      exportTime: moment().format('YYYY-MM-DD HH:mm:ss'),
    }

    const blob = new Blob([JSON.stringify(reportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `analytics-report-${moment().format('YYYY-MM-DD')}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  if (!data) {
    return <div>加载中...</div>
  }

  const { overview, caseStats, lawyerStats, clientStats } = data

  return (
    <div style={{ padding: '24px' }}>
      {/* 控制栏 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row justify='space-between' align='middle'>
          <Col>
            <Space>
              <Select value={timeRange} onChange={setTimeRange} style={{ width: 120 }}>
                <Option value='week'>最近一周</Option>
                <Option value='month'>最近一月</Option>
                <Option value='quarter'>最近一季</Option>
                <Option value='year'>最近一年</Option>
              </Select>
              <RangePicker value={dateRange} onChange={(dates) => dates && setDateRange(dates)} />
              <Button type='primary' onClick={fetchAnalyticsData} loading={loading}>
                刷新数据
              </Button>
            </Space>
          </Col>
          <Col>
            <Button icon={<DownloadOutlined />} onClick={exportReport}>
              导出报告
            </Button>
          </Col>
        </Row>
      </Card>

      {/* 概览统计 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title='总案件数'
              value={overview.totalCases}
              prefix={<FileTextOutlined />}
              suffix={
                <Tag color='green'>
                  <ArrowUpOutlined /> 12%
                </Tag>
              }
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title='进行中案件'
              value={overview.ongoingCases}
              prefix={<CalendarOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title='总收入'
              value={overview.totalRevenue}
              prefix={<DollarCircleOutlined />}
              suffix='元'
              precision={0}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title='总工作时长'
              value={overview.totalHours}
              prefix={<ClockCircleOutlined />}
              suffix='小时'
            />
          </Card>
        </Col>
      </Row>

      {/* 详细分析 */}
      <Tabs
        defaultActiveKey='cases'
        type='card'
        items={[
          {
            key: 'cases',
            label: '案件分析',
            children: (
              <Row gutter={16}>
                <Col span={12}>
                  <Card title='案件类型分布'>
                    <ResponsiveContainer width='100%' height={300}>
                      <PieChart>
                        <Pie
                          data={caseStats.byType}
                          cx='50%'
                          cy='50%'
                          labelLine={false}
                          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                          outerRadius={80}
                          fill='#8884d8'
                          dataKey='value'
                        >
                          {caseStats.byType.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={entry.color} />
                          ))}
                        </Pie>
                        <RechartsTooltip />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card title='月度趋势'>
                    <ResponsiveContainer width='100%' height={300}>
                      <LineChart data={caseStats.monthlyTrend}>
                        <CartesianGrid strokeDasharray='3 3' />
                        <XAxis dataKey='month' />
                        <YAxis />
                        <RechartsTooltip />
                        <Legend />
                        <Bar dataKey='cases' fill='#8884d8' name='案件数' />
                        <Line
                          type='monotone'
                          dataKey='revenue'
                          stroke='#82ca9d'
                          name='收入(千元)'
                          yAxisId='right'
                        />
                        <YAxis yAxisId='right' orientation='right' />
                      </LineChart>
                    </ResponsiveContainer>
                  </Card>
                </Col>
              </Row>
            ),
          },
          {
            key: 'lawyers',
            label: '律师绩效',
            children: (
              <Row gutter={16}>
                <Col span={16}>
                  <Card title='律师排行榜'>
                    <List
                      dataSource={lawyerStats.topPerformers}
                      renderItem={(item, index) => (
                        <List.Item>
                          <List.Item.Meta
                            avatar={
                              <Avatar
                                style={{ backgroundColor: index < 3 ? '#f56a00' : '#87d068' }}
                              >
                                {item.avatar}
                              </Avatar>
                            }
                            title={
                              <Space>
                                <Text strong>{item.name}</Text>
                                {index === 0 && <TrophyOutlined style={{ color: '#f56a00' }} />}
                              </Space>
                            }
                            description={
                              <Space size='large'>
                                <span>案件: {item.cases}</span>
                                <span>时长: {item.hours}h</span>
                                <span>收入: ¥{(item.revenue / 1000).toFixed(1)}k</span>
                                <span>满意度: {item.satisfaction}</span>
                              </Space>
                            }
                          />
                        </List.Item>
                      )}
                    />
                  </Card>
                </Col>
                <Col span={8}>
                  <Card title='工作负载'>
                    <Space direction='vertical' style={{ width: '100%' }}>
                      {lawyerStats.workload.map((lawyer, index) => (
                        <div key={index}>
                          <div
                            style={{
                              display: 'flex',
                              justifyContent: 'space-between',
                              marginBottom: '8px',
                            }}
                          >
                            <Text>{lawyer.name}</Text>
                            <Text>
                              {lawyer.current}/{lawyer.capacity}
                            </Text>
                          </div>
                          <Progress
                            percent={(lawyer.current / lawyer.capacity) * 100}
                            showInfo={false}
                            status={
                              lawyer.current >= lawyer.capacity * 0.8 ? 'exception' : 'normal'
                            }
                          />
                        </div>
                      ))}
                    </Space>
                  </Card>
                </Col>
              </Row>
            ),
          },
          {
            key: 'clients',
            label: '客户分析',
            children: (
              <Row gutter={16}>
                <Col span={16}>
                  <Card title='重要客户'>
                    <Table
                      dataSource={clientStats.topClients}
                      pagination={false}
                      columns={[
                        {
                          title: '客户名称',
                          dataIndex: 'name',
                          key: 'name',
                        },
                        {
                          title: '案件数',
                          dataIndex: 'cases',
                          key: 'cases',
                        },
                        {
                          title: '收入',
                          dataIndex: 'revenue',
                          key: 'revenue',
                          render: (value: number) => `¥${value.toLocaleString()}`,
                        },
                        {
                          title: '满意度',
                          dataIndex: 'satisfaction',
                          key: 'satisfaction',
                          render: (value: number) => (
                            <div style={{ display: 'flex', alignItems: 'center' }}>
                              <Progress
                                type='circle'
                                percent={value * 20}
                                width={30}
                                strokeColor='#52c41a'
                                showInfo={false}
                              />
                              <span style={{ marginLeft: '8px' }}>{value}</span>
                            </div>
                          ),
                        },
                      ]}
                    />
                  </Card>
                </Col>
                <Col span={8}>
                  <Card title='客户指标'>
                    <Space direction='vertical' style={{ width: '100%' }} size='large'>
                      <div>
                        <Text>客户保留率</Text>
                        <Progress
                          type='circle'
                          percent={clientStats.retention}
                          strokeColor='#52c41a'
                        />
                      </div>
                      <div>
                        <Text>新客户获取</Text>
                        <Statistic
                          value={clientStats.acquisition}
                          suffix='%'
                          prefix={<TeamOutlined />}
                        />
                      </div>
                    </Space>
                  </Card>
                </Col>
              </Row>
            ),
          },
          {
            key: 'efficiency',
            label: '效率分析',
            children: (
              <Row gutter={16}>
                <Col span={12}>
                  <Card title='案件处理效率'>
                    <ResponsiveContainer width='100%' height={300}>
                      <BarChart data={lawyerStats.topPerformers}>
                        <CartesianGrid strokeDasharray='3 3' />
                        <XAxis dataKey='name' />
                        <YAxis />
                        <RechartsTooltip />
                        <Legend />
                        <Bar dataKey='cases' fill='#8884d8' name='案件数' />
                        <Bar dataKey='hours' fill='#82ca9d' name='工作时长' />
                      </BarChart>
                    </ResponsiveContainer>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card title='收入贡献'>
                    <ResponsiveContainer width='100%' height={300}>
                      <PieChart>
                        <Pie
                          data={lawyerStats.topPerformers.map((lawyer) => ({
                            name: lawyer.name,
                            value: lawyer.revenue,
                            color: `hsl(${Math.random() * 360}, 70%, 50%)`,
                          }))}
                          cx='50%'
                          cy='50%'
                          labelLine={false}
                          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                          outerRadius={80}
                          fill='#8884d8'
                          dataKey='value'
                        >
                          {lawyerStats.topPerformers.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={`hsl(${index * 60}, 70%, 50%)`} />
                          ))}
                        </Pie>
                        <RechartsTooltip />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </Card>
                </Col>
              </Row>
            ),
          },
        ]}
      />
    </div>
  )
}

export default AnalyticsDashboard
