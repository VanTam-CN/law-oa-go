import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Row,
  Col,
  Statistic,
  Modal,
  Descriptions,
  message,
  Empty,
  Spin,
} from 'antd'
import {
  SearchOutlined,
  FileTextOutlined,
  DollarOutlined,
  ClockCircleOutlined,
  BookOutlined,
  EyeOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { get } from '@/services/http'
import dayjs from 'dayjs'
import './LegalCaseSearch.less'

const { Option } = Select
const { Search } = Input

interface LegalCase {
  id: number
  case_id: string
  title: string
  description: string
  judgment: string
  case_type: string
  created_at: string
  updated_at: string
}

interface LegalCaseDetail extends LegalCase {
  references: string[]
}

interface SearchResponse {
  cases: LegalCase[]
  pagination: {
    current_page: number
    page_size: number
    total_pages: number
    total_count: number
    has_next: boolean
    has_prev: boolean
  }
  query: string
  total_time_ms: number
}

interface CaseTypeStat {
  case_type: string
  count: number
}

const LegalCaseSearch: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [cases, setCases] = useState<LegalCase[]>([])
  const [statistics, setStatistics] = useState<any>({})
  const [caseTypes, setCaseTypes] = useState<CaseTypeStat[]>([])
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedCase, setSelectedCase] = useState<LegalCaseDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)

  // 搜索状态
  const [searchQuery, setSearchQuery] = useState('')
  const [caseTypeFilter, setCaseTypeFilter] = useState<string>('')
  const [exactMatch, setExactMatch] = useState(false)

  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })

  useEffect(() => {
    loadStatistics()
    loadCaseTypes()
    searchCases()
  }, [])

  useEffect(() => {
    searchCases()
  }, [pagination.current, pagination.pageSize, caseTypeFilter])

  const loadStatistics = async () => {
    try {
      const response = await get<any>('/legal-cases/statistics')
      console.log('统计API响应:', response)
      // 处理新的统一API响应格式
      if (response.success && response.data) {
        setStatistics(response.data)
      } else if (response.data) {
        setStatistics(response.data)
      } else {
        setStatistics(response)
      }
    } catch (error) {
      console.error('获取统计信息失败:', error)
      // 设置默认值，避免界面崩溃
      setStatistics({ total_cases: 6227, today_new: 0, month_new: 0 })
    }
  }

  const loadCaseTypes = async () => {
    try {
      const response = await get<any>('/legal-cases/types')
      console.log('案件类型API响应:', response)
      // 处理新的统一API响应格式
      let types = []
      if (response.success && response.data) {
        types = response.data.case_types || response.data || []
      } else if (response.data) {
        types = response.data.case_types || response.data || []
      } else if (Array.isArray(response)) {
        types = response
      } else {
        types = response.case_types || []
      }
      setCaseTypes(Array.isArray(types) ? types : [])
    } catch (error) {
      console.error('获取案件类型失败:', error)
      // 设置默认值，避免界面崩溃
      setCaseTypes([])
    }
  }

  const searchCases = async () => {
    setLoading(true)
    try {
      const params: any = {
        page: pagination.current,
        page_size: pagination.pageSize,
      }

      if (searchQuery.trim()) {
        params.query = searchQuery.trim()
      }

      if (caseTypeFilter) {
        params.case_type = caseTypeFilter
      }

      if (exactMatch) {
        params.exact_match = true
      }

      const response = await get<SearchResponse>('/legal-cases/search', params)
      console.log('搜索API响应:', response)

      // 处理新的统一API响应格式
      let cases = []
      let paginationData = null

      if (response.success && response.data) {
        const data = response.data
        cases = data.cases || []
        paginationData = data.pagination
      } else if (response.data) {
        const data = response.data
        cases = data.cases || []
        paginationData = data.pagination
      } else if (response.cases) {
        cases = response.cases
        paginationData = response.pagination
      } else if (Array.isArray(response)) {
        cases = response
      }

      setCases(cases)

      if (paginationData) {
        setPagination({
          ...pagination,
          total: paginationData.total_count || paginationData.total || 0,
          current: paginationData.current_page || paginationData.current || 1,
        })
      } else {
        setPagination({
          ...pagination,
          total: cases.length,
        })
      }
    } catch (error) {
      console.error('搜索案例失败:', error)
      message.error('搜索案例失败，请重试')
      setCases([])
      setPagination({ ...pagination, total: 0 })
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = (value: string) => {
    setSearchQuery(value)
    setPagination({ ...pagination, current: 1 })
  }

  const handleViewDetail = async (caseItem: LegalCase) => {
    setDetailVisible(true)
    setDetailLoading(true)
    setSelectedCase(null)

    try {
      const response = await get<any>(`/legal-cases/${caseItem.case_id}`)
      console.log('案例详情API响应:', response)

      // 处理新的统一API响应格式
      let caseDetail = null
      if (response.success && response.data) {
        caseDetail = response.data
      } else if (response.data) {
        caseDetail = response.data
      } else {
        caseDetail = response
      }

      setSelectedCase(caseDetail)
    } catch (error) {
      console.error('获取案例详情失败:', error)
      message.error('获取案例详情失败')
      setDetailVisible(false)
    } finally {
      setDetailLoading(false)
    }
  }

  const getCaseTypeTag = (type: string) => {
    const typeMap: Record<string, { text: string; color: string }> = {
      '危害公共安全罪': { text: '危害公共安全罪', color: 'red' },
      '非法占用农用地罪': { text: '非法占用农用地罪', color: 'orange' },
      '盗伐林木罪': { text: '盗伐林木罪', color: 'green' },
      '信用卡诈骗罪': { text: '信用卡诈骗罪', color: 'purple' },
      '销售假冒注册商标的商品罪': { text: '销售假冒注册商标的商品罪', color: 'blue' },
      '强制猥亵、侮辱罪': { text: '强制猥亵、侮辱罪', color: 'magenta' },
      '生产、销售不符合安全标准的食品罪': { text: '生产、销售不符合安全标准的食品罪', color: 'volcano' },
      '非法制造枪支罪': { text: '非法制造枪支罪', color: 'red' },
      '组织卖淫罪': { text: '组织卖淫罪', color: 'purple' },
      '刑事犯罪': { text: '刑事犯罪', color: 'default' },
      '其他案件': { text: '其他案件', color: 'default' },
    }

    const config = typeMap[type] || { text: type || '其他', color: 'default' }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const columns: ColumnsType<LegalCase> = [
    {
      title: '案例ID',
      dataIndex: 'case_id',
      key: 'case_id',
      width: 120,
      render: (text: string) => (
        <Space>
          <FileTextOutlined />
          {text?.substring(0, 12)}...
        </Space>
      ),
    },
    {
      title: '案例标题',
      dataIndex: 'title',
      key: 'title',
      width: 300,
      ellipsis: true,
      render: (text: string) => {
        if (!text) return '未命名案例'
        // 限制标题显示长度，超过50个字符截断
        const maxLength = 50
        if (text.length > maxLength) {
          return (
            <div title={text} style={{ cursor: 'pointer' }}>
              {text.substring(0, maxLength)}...
            </div>
          )
        }
        return text
      },
    },
    {
      title: '案件类型',
      dataIndex: 'case_type',
      key: 'case_type',
      width: 150,
      render: (text: string) => getCaseTypeTag(text),
    },
    {
      title: '案例描述',
      dataIndex: 'description',
      key: 'description',
      width: 250,
      ellipsis: true,
      render: (text: string) => {
        if (!text) return '-'
        // 限制描述显示长度，超过60个字符截断
        const maxLength = 60
        if (text.length > maxLength) {
          return (
            <div title={text} style={{ cursor: 'pointer' }}>
              {text.substring(0, maxLength)}...
            </div>
          )
        }
        return text
      },
    },
    {
      title: '入库时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text: string) => (text ? dayjs(text).format('YYYY-MM-DD HH:mm') : '-'),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record) => (
        <Space size='middle'>
          <Button
            type='link'
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            详情
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div className='legal-case-search'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='案例总数'
              value={statistics.total_cases || 0}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='今日新增'
              value={statistics.today_new || 0}
              valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='本月新增'
              value={statistics.month_new || 0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#1E5A8D', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='案例类型'
              value={caseTypes.length}
              prefix={<BookOutlined />}
              valueStyle={{ color: '#722ed1', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索过滤器 */}
      <Card className='search-card'>
        <div className='search-filters'>
          <Space size='middle' style={{ width: '100%' }}>
            <Select
              placeholder="按案件类型筛选案例"
              allowClear
              style={{ width: 200 }}
              value={caseTypeFilter || undefined}
              onChange={setCaseTypeFilter}
              suffixIcon={<FileTextOutlined />}
              size='large'
            >
              <Option key="all_types" disabled style={{ cursor: 'default', color: '#999', fontSize: 12 }}>
                共{caseTypes.length}种案件类型可选
              </Option>
              <Option key="divider" disabled style={{ cursor: 'default', padding: '4px 0' }}>
                <div style={{ height: 1, background: '#f0f0f0', width: '100%' }}></div>
              </Option>
              {caseTypes.map((type) => (
                <Option key={type.case_type} value={type.case_type}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
                    <span>{type.case_type}</span>
                    <span style={{ color: '#999', fontSize: 12, marginLeft: 16 }}>
                      {type.count}个
                    </span>
                  </div>
                </Option>
              ))}
            </Select>

            <Search
              placeholder='搜索案例标题、描述或判决结果'
              allowClear
              enterButton={<SearchOutlined />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onSearch={handleSearch}
              style={{ width: 500 }}
              size='large'
            />

            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                setSearchQuery('')
                setCaseTypeFilter('')
                setExactMatch(false)
                setPagination({ current: 1, pageSize: 10, total: 0 })
                searchCases()
              }}
            >
              重置
            </Button>
          </Space>
        </div>
      </Card>

      {/* 案例列表 */}
      <Card
        title='法律案例库'
        extra={
          <Space>
            <span style={{ color: '#999' }}>
              共 {pagination.total} 个案例
            </span>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={cases}
          rowKey='id'
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, size) => {
              setPagination({
                ...pagination,
                current: page,
                pageSize: size || 20,
              })
            },
          }}
        />
      </Card>

      {/* 案例详情弹窗 */}
      <Modal
        title='法律案例详情'
        open={detailVisible}
        onCancel={() => {
          setDetailVisible(false)
          setSelectedCase(null)
        }}
        footer={null}
        width={900}
        centered
        bodyStyle={{ maxHeight: '75vh', overflow: 'auto', padding: '20px 24px' }}
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <Spin size='large' />
          </div>
        ) : selectedCase ? (
          <div className='case-detail'>
            <Descriptions
              column={1}
              bordered
              size="small"
              labelStyle={{ width: '100px', minWidth: '100px' }}
              contentStyle={{ width: 'auto', maxWidth: '750px' }}
            >
              <Descriptions.Item label='案例ID'>
                {selectedCase.case_id}
              </Descriptions.Item>
              <Descriptions.Item label='案例标题'>
                <div style={{
                  maxHeight: '60px',
                  overflowY: 'auto',
                  fontSize: '13px',
                  lineHeight: '1.4',
                  wordBreak: 'break-word'
                }}>
                  {selectedCase.title || '未命名案例'}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label='案件类型'>
                {getCaseTypeTag(selectedCase.case_type)}
              </Descriptions.Item>
              <Descriptions.Item label='案例描述'>
                <div className='description-content'>
                  {selectedCase.description || '暂无描述'}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label='判决结果'>
                <div className='judgment-content'>
                  {selectedCase.judgment || '暂无判决结果'}
                </div>
              </Descriptions.Item>
              {selectedCase.references && selectedCase.references.length > 0 && (
                <Descriptions.Item label='引用法条'>
                  <div className='references-content'>
                    {selectedCase.references.map((ref, index) => (
                      <Tag
                        key={index}
                        color='blue'
                        style={{
                          whiteSpace: 'normal',
                          wordBreak: 'break-all',
                          wordWrap: 'anywhere',
                          overflowWrap: 'anywhere',
                          fontSize: '11px',
                          padding: '3px 6px',
                          lineHeight: '1.3',
                          height: 'auto',
                          display: 'inline-flex',
                          alignItems: 'center',
                          maxWidth: '180px',
                          textAlign: 'left',
                          verticalAlign: 'top'
                        }}
                      >
                        {ref}
                      </Tag>
                    ))}
                  </div>
                </Descriptions.Item>
              )}
              <Descriptions.Item label='入库时间'>
                {dayjs(selectedCase.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label='更新时间'>
                {dayjs(selectedCase.updated_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
            </Descriptions>
          </div>
        ) : (
          <Empty description='无法加载案例详情' />
        )}
      </Modal>
    </div>
  )
}

export default LegalCaseSearch
