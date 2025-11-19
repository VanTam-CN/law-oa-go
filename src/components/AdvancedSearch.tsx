import React, { useState, useEffect } from 'react'
import {
  Card,
  Form,
  Input,
  Select,
  DatePicker,
  InputNumber,
  Button,
  Space,
  Tag,
  Row,
  Col,
  Divider,
  Badge,
  Tooltip,
  Collapse,
  Typography,
} from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  FilterOutlined,
  CalendarOutlined,
  DollarOutlined,
  UserOutlined,
  FileTextOutlined,
  TeamOutlined,
  ClockCircleOutlined,
  DownOutlined,
  UpOutlined,
} from '@ant-design/icons'
import type { RangePickerProps } from 'antd/es/date-picker'
import dayjs from 'dayjs'

const { Option } = Select
const { Search } = Input
const { Text } = Typography
const { RangePicker } = DatePicker
const { Panel } = Collapse

interface AdvancedSearchProps {
  onSearch: (params: any) => void
  onReset: () => void
  loading?: boolean
  visible?: boolean
  onVisibleChange?: (visible: boolean) => void
}

interface SearchParams {
  searchText: string
  caseType: string
  status: string
  projectType: string
  lawyerId: string
  clientId: string
  dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null
  amountRange: [number, number] | null
  sortBy: string
  sortOrder: string
}

const AdvancedSearch: React.FC<AdvancedSearchProps> = ({
  onSearch,
  onReset,
  loading = false,
  visible = false,
  onVisibleChange,
}) => {
  const [form] = Form.useForm()
  const [activeFilters, setActiveFilters] = useState<string[]>([])
  const [searchCount, setSearchCount] = useState(0)

  // 模拟律师数据
  const lawyers = [
    { id: '1', name: '张律师', specialty: '民事诉讼,合同纠纷' },
    { id: '2', name: '李律师', specialty: '刑事辩护,知识产权' },
    { id: '3', name: '王律师', specialty: '商事仲裁,公司法' },
    { id: '4', name: '陈律师', specialty: '婚姻家庭,继承法' },
    { id: '5', name: '刘律师', specialty: '行政诉讼,劳动争议' },
    { id: '6', name: '杨律师', specialty: '房地产,建设工程' },
    { id: '7', name: '赵律师', specialty: '金融证券,投资并购' },
    { id: '8', name: '孙律师', specialty: '知识产权,不正当竞争' },
  ]

  // 模拟客户数据
  const clients = [
    { id: '1', name: '张三', type: 'PERSON' },
    { id: '2', name: 'ABC科技有限公司', type: 'COMPANY' },
    { id: '3', name: '王五', type: 'PERSON' },
    { id: '4', name: 'DEF贸易集团', type: 'COMPANY' },
    { id: '5', name: '赵六', type: 'PERSON' },
    { id: '6', name: 'GHI投资公司', type: 'COMPANY' },
    { id: '7', name: '孙七', type: 'PERSON' },
    { id: '8', name: 'JKL律师事务所', type: 'COMPANY' },
    { id: '9', name: '周八', type: 'PERSON' },
    { id: '10', name: 'MNO咨询集团', type: 'COMPANY' },
  ]

  // 监听表单变化，更新活跃筛选器
  const handleValuesChange = (changedValues: any, allValues: any) => {
    const filters = []

    if (allValues.searchText) {
      filters.push('关键词')
    }
    if (allValues.caseType) {
      filters.push('案件类型')
    }
    if (allValues.status) {
      filters.push('状态')
    }
    if (allValues.projectType) {
      filters.push('项目类型')
    }
    if (allValues.lawyerId) {
      filters.push('负责律师')
    }
    if (allValues.clientId) {
      filters.push('客户')
    }
    if (allValues.dateRange) {
      filters.push('日期范围')
    }
    if (allValues.amountRange) {
      filters.push('金额范围')
    }

    setActiveFilters(filters)
  }

  // 处理搜索
  const handleSearch = async () => {
    try {
      const values = await form.validateFields()

      const params: SearchParams = {
        searchText: values.searchText || '',
        caseType: values.caseType || '',
        status: values.status || '',
        projectType: values.projectType || '',
        lawyerId: values.lawyerId || '',
        clientId: values.clientId || '',
        dateRange: values.dateRange || null,
        amountRange: values.amountRange || null,
        sortBy: values.sortBy || 'createTime',
        sortOrder: values.sortOrder || 'desc',
      }

      setSearchCount((prev) => prev + 1)
      onSearch(params)
    } catch (error) {
      console.error('搜索参数验证失败:', error)
    }
  }

  // 处理重置
  const handleReset = () => {
    form.resetFields()
    setActiveFilters([])
    onReset()
  }

  // 快速搜索预设
  const quickSearchPresets = [
    { label: '今日新增', values: { dateRange: [dayjs(), dayjs()] } },
    {
      label: '本周新增',
      values: { dateRange: [dayjs().startOf('week'), dayjs().endOf('week')] },
    },
    {
      label: '本月新增',
      values: { dateRange: [dayjs().startOf('month'), dayjs().endOf('month')] },
    },
    { label: '高金额案件', values: { amountRange: [1000000, null] } },
    { label: '进行中案件', values: { status: '1' } },
    { label: '已结案案件', values: { status: '2' } },
  ]

  // 应用快速搜索
  const applyQuickSearch = (preset: any) => {
    form.setFieldsValue(preset.values)
    handleValuesChange(preset.values, form.getFieldsValue())
  }

  return (
    <div className='advanced-search'>
      <Card
        title={
          <Space>
            <FilterOutlined />
            高级搜索
            {activeFilters.length > 0 && (
              <Badge count={activeFilters.length} size='small'>
                <Tag color='blue'>已选条件</Tag>
              </Badge>
            )}
            {searchCount > 0 && <Tag color='green'>搜索 {searchCount} 次</Tag>}
          </Space>
        }
        extra={
          <Space>
            <Button type='dashed' icon={<ReloadOutlined />} onClick={handleReset} size='small'>
              重置
            </Button>
            <Button
              type='primary'
              icon={<SearchOutlined />}
              onClick={handleSearch}
              loading={loading}
            >
              搜索
            </Button>
          </Space>
        }
        size='small'
      >
        <Form form={form} layout='vertical' onValuesChange={handleValuesChange}>
          {/* 快速搜索预设 */}
          <div className='quick-search-presets'>
            <div style={{ marginBottom: 16 }}>
              <Text strong>快速搜索：</Text>
              <Space wrap style={{ marginTop: 8 }}>
                {quickSearchPresets.map((preset, index) => (
                  <Tag
                    key={index}
                    color='processing'
                    style={{ cursor: 'pointer' }}
                    onClick={() => applyQuickSearch(preset)}
                  >
                    {preset.label}
                  </Tag>
                ))}
              </Space>
            </div>
          </div>

          <Collapse ghost>
            <Panel header='基本搜索条件' key='1'>
              <Row gutter={16}>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='searchText' label='关键词搜索'>
                    <Search
                      placeholder='搜索案件名称、编号、客户等'
                      allowClear
                      enterButton={<SearchOutlined />}
                      onSearch={handleSearch}
                    />
                  </Form.Item>
                </Col>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='caseType' label='案件类型'>
                    <Select placeholder='请选择案件类型' allowClear>
                      <Option value='CIVIL'>民事案件</Option>
                      <Option value='COMMERCIAL'>商事案件</Option>
                      <Option value='CRIMINAL'>刑事案件</Option>
                      <Option value='ADMINISTRATIVE'>行政案件</Option>
                      <Option value='ADVISORY'>咨询项目</Option>
                      <Option value='REVIEW'>审查项目</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='status' label='案件状态'>
                    <Select placeholder='请选择状态' allowClear>
                      <Option value='0'>未开始</Option>
                      <Option value='1'>进行中</Option>
                      <Option value='2'>已结案</Option>
                      <Option value='3'>已归档</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
            </Panel>

            <Panel header='高级筛选条件' key='2'>
              <Row gutter={16}>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='projectType' label='项目类型'>
                    <Select placeholder='请选择项目类型' allowClear>
                      <Option value='CIVIL'>民事诉讼</Option>
                      <Option value='COMMERCIAL'>商事仲裁</Option>
                      <Option value='CRIMINAL'>刑事辩护</Option>
                      <Option value='ADMINISTRATIVE'>行政诉讼</Option>
                      <Option value='ADVISORY'>咨询项目</Option>
                      <Option value='REVIEW'>审查项目</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='lawyerId' label='负责律师'>
                    <Select placeholder='请选择律师' allowClear>
                      {lawyers.map((lawyer) => (
                        <Option key={lawyer.id} value={lawyer.id}>
                          <Space>
                            <UserOutlined />
                            {lawyer.name}
                            <Text type='secondary' style={{ fontSize: '12px' }}>
                              {lawyer.specialty}
                            </Text>
                          </Space>
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                </Col>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='clientId' label='客户'>
                    <Select placeholder='请选择客户' allowClear>
                      {clients.map((client) => (
                        <Option key={client.id} value={client.id}>
                          <Space>
                            <TeamOutlined />
                            {client.name}
                            <Tag color={client.type === 'COMPANY' ? 'orange' : 'blue'}>
                              {client.type === 'COMPANY' ? '企业' : '个人'}
                            </Tag>
                          </Space>
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
            </Panel>

            <Panel header='范围筛选' key='3'>
              <Row gutter={16}>
                <Col xs={24} sm={12} md={12}>
                  <Form.Item name='dateRange' label='创建日期范围'>
                    <RangePicker
                      style={{ width: '100%' }}
                      presets={[
                        {
                          label: '最近7天',
                          value: [dayjs().subtract(7, 'd'), dayjs()],
                        },
                        {
                          label: '最近30天',
                          value: [dayjs().subtract(30, 'd'), dayjs()],
                        },
                        {
                          label: '最近90天',
                          value: [dayjs().subtract(90, 'd'), dayjs()],
                        },
                      ]}
                    />
                  </Form.Item>
                </Col>
                <Col xs={24} sm={12} md={12}>
                  <Form.Item name='amountRange' label='涉案金额范围（元）'>
                    <div style={{ display: 'flex', width: '100%' }}>
                      <InputNumber
                        style={{ width: '45%' }}
                        placeholder='最小金额'
                        min={0}
                        addonBefore='¥'
                      />
                      <Input
                        style={{
                          width: '10%',
                          borderLeft: 0,
                          borderRight: 0,
                          pointerEvents: 'none',
                        }}
                        placeholder='~'
                        disabled
                      />
                      <InputNumber
                        style={{ width: '45%' }}
                        placeholder='最大金额'
                        min={0}
                        addonBefore='¥'
                      />
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </Panel>

            <Panel header='排序设置' key='4'>
              <Row gutter={16}>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='sortBy' label='排序字段' initialValue='createTime'>
                    <Select>
                      <Option value='createTime'>创建时间</Option>
                      <Option value='updateTime'>更新时间</Option>
                      <Option value='caseAmount'>涉案金额</Option>
                      <Option value='caseName'>案件名称</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item name='sortOrder' label='排序方式' initialValue='desc'>
                    <Select>
                      <Option value='desc'>降序</Option>
                      <Option value='asc'>升序</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
            </Panel>
          </Collapse>
        </Form>
      </Card>
    </div>
  )
}

export default AdvancedSearch
