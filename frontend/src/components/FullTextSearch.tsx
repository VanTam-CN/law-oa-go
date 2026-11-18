import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Input,
  Button,
  Select,
  DatePicker,
  InputNumber,
  Space,
  Table,
  Tag,
  message,
  Row,
  Col,
  Divider,
  Tooltip,
  Badge,
} from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  SyncOutlined,
  SettingOutlined,
  ClearOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  fullTextSearch,
  advancedSearch,
  searchSuggestions,
  syncCase,
  batchSyncCases,
  syncAllCases,
} from '@/api/search'
import { SearchHighlight } from '@/components/SearchHighlight'
import { saveSearchHistory, getSearchHistory } from '@/utils/searchHistory'

const { Option } = Select
const { Search } = Input
const { RangePicker } = DatePicker

interface CaseRecord {
  id: number
  caseNumber: string
  caseName: string
  caseType: string
  caseStatus: string
  clientName: string
  responsibleLawyer: string
  caseDescription: string
  caseAmount: number
  startDate: string
  endDate: string
  createTime: string
  updateTime: string
  tags: string[]
  priority: number
}

interface SearchForm {
  keyword: string
  caseType: string
  caseStatus: string
  responsibleLawyer: string
  dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null
  amountRange: [number, number] | null
}

const FullTextSearch: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<CaseRecord[]>([])
  const [total, setTotal] = useState(0)
  const [searchForm, setSearchForm] = useState<SearchForm>({
    keyword: '',
    caseType: '',
    caseStatus: '',
    responsibleLawyer: '',
    dateRange: null,
    amountRange: null,
  })
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [searchHistory, setSearchHistory] = useState<string[]>([])
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })

  // 加载搜索历史
  useEffect(() => {
    const history = getSearchHistory()
    setSearchHistory(history)
  }, [])

  // 执行搜索
  const handleSearch = useCallback(
    async (pageNum: number = 1) => {
      if (!searchForm.keyword.trim() && !showAdvanced) {
        message.warning('请输入搜索关键词')
        return
      }

      setLoading(true)
      try {
        let result
        if (showAdvanced) {
          // 高级搜索
          const params: any = {
            pageNum,
            pageSize: pagination.pageSize,
          }

          if (searchForm.keyword) {
            params.keyword = searchForm.keyword
          }
          if (searchForm.caseType) {
            params.caseType = searchForm.caseType
          }
          if (searchForm.caseStatus) {
            params.caseStatus = searchForm.caseStatus
          }
          if (searchForm.responsibleLawyer) {
            params.responsibleLawyer = searchForm.responsibleLawyer
          }

          if (searchForm.dateRange) {
            params.startDate = searchForm.dateRange[0].format('YYYY-MM-DD HH:mm:ss')
            params.endDate = searchForm.dateRange[1].format('YYYY-MM-DD HH:mm:ss')
          }

          if (searchForm.amountRange) {
            params.minAmount = searchForm.amountRange[0]
            params.maxAmount = searchForm.amountRange[1]
          }

          result = await advancedSearch(params)
        } else {
          // 全文搜索
          result = await fullTextSearch({
            keyword: searchForm.keyword,
            pageNum,
            pageSize: pagination.pageSize,
          })

          // 保存搜索历史
          if (searchForm.keyword.trim()) {
            saveSearchHistory(searchForm.keyword)
            setSearchHistory(getSearchHistory())
          }
        }

        setData(result.data.rows || [])
        setTotal(result.data.total || 0)
        setPagination((prev) => ({
          ...prev,
          current: pageNum,
          total: result.data.total || 0,
        }))
      } catch (error) {
        message.error('搜索失败')
        console.error('搜索错误:', error)
      } finally {
        setLoading(false)
      }
    },
    [searchForm, showAdvanced, pagination.pageSize],
  )

  // 获取搜索建议
  const handleSearchChange = useCallback(async (value: string) => {
    if (value.trim()) {
      try {
        const result = await searchSuggestions({
          prefix: value,
          size: 5,
        })
        setSuggestions(result.data || [])
      } catch (error) {
        console.error('获取搜索建议失败:', error)
      }
    } else {
      setSuggestions([])
    }
  }, [])

  // 同步数据
  const handleSync = useCallback(async (type: 'single' | 'batch' | 'all', caseIds?: number[]) => {
    try {
      setLoading(true)
      switch (type) {
        case 'single':
          if (caseIds && caseIds.length > 0) {
            await syncCase(caseIds[0])
            message.success('同步成功')
          }
          break
        case 'batch':
          if (caseIds && caseIds.length > 0) {
            await batchSyncCases(caseIds)
            message.success('批量同步成功')
          }
          break
        case 'all':
          await syncAllCases()
          message.success('全量同步成功')
          break
      }
    } catch (error) {
      message.error('同步失败')
      console.error('同步错误:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  // 清空搜索条件
  const handleClear = useCallback(() => {
    setSearchForm({
      keyword: '',
      caseType: '',
      caseStatus: '',
      responsibleLawyer: '',
      dateRange: null,
      amountRange: null,
    })
    setSuggestions([])
  }, [])

  // 表格列定义
  const columns: ColumnsType<CaseRecord> = [
    {
      title: '案件编号',
      dataIndex: 'caseNumber',
      key: 'caseNumber',
      width: 120,
      render: (text, record) => (
        <SearchHighlight content={text} searchTerms={[searchForm.keyword]} maxLength={20} />
      ),
    },
    {
      title: '案件名称',
      dataIndex: 'caseName',
      key: 'caseName',
      width: 200,
      render: (text, record) => (
        <SearchHighlight content={text} searchTerms={[searchForm.keyword]} maxLength={30} />
      ),
    },
    {
      title: '案件类型',
      dataIndex: 'caseType',
      key: 'caseType',
      width: 100,
      render: (text) => <Tag color='blue'>{text}</Tag>,
    },
    {
      title: '案件状态',
      dataIndex: 'caseStatus',
      key: 'caseStatus',
      width: 100,
      render: (text) => {
        const statusColors: Record<string, string> = {
          进行中: 'processing',
          已完成: 'success',
          已暂停: 'warning',
          已取消: 'error',
        }
        return <Badge status={statusColors[text] || 'default'} text={text} />
      },
    },
    {
      title: '客户名称',
      dataIndex: 'clientName',
      key: 'clientName',
      width: 120,
      render: (text, record) => (
        <SearchHighlight content={text} searchTerms={[searchForm.keyword]} maxLength={20} />
      ),
    },
    {
      title: '负责律师',
      dataIndex: 'responsibleLawyer',
      key: 'responsibleLawyer',
      width: 100,
      render: (text, record) => (
        <SearchHighlight content={text} searchTerms={[searchForm.keyword]} maxLength={15} />
      ),
    },
    {
      title: '案件金额',
      dataIndex: 'caseAmount',
      key: 'caseAmount',
      width: 100,
      render: (text) => `¥${text?.toLocaleString() || '0'}`,
    },
    {
      title: '开始日期',
      dataIndex: 'startDate',
      key: 'startDate',
      width: 120,
      render: (text) => (text ? dayjs(text).format('YYYY-MM-DD') : '-'),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record) => (
        <Space>
          <Tooltip title='同步该案例'>
            <Button
              type='link'
              size='small'
              icon={<SyncOutlined />}
              onClick={() => handleSync('single', [record.id])}
            />
          </Tooltip>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Card
        title='全文检索'
        extra={
          <Space>
            <Button
              type='primary'
              icon={<SyncOutlined />}
              onClick={() => handleSync('all')}
              loading={loading}
            >
              全量同步
            </Button>
            <Button
              icon={<SettingOutlined />}
              onClick={() => setShowAdvanced(!showAdvanced)}
              type={showAdvanced ? 'primary' : 'default'}
            >
              高级搜索
            </Button>
          </Space>
        }
      >
        {/* 基础搜索 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={16}>
            <Search
              placeholder='输入关键词搜索案件...'
              allowClear
              enterButton={<SearchOutlined />}
              size='large'
              value={searchForm.keyword}
              onChange={(e) => {
                setSearchForm((prev) => ({ ...prev, keyword: e.target.value }))
                handleSearchChange(e.target.value)
              }}
              onSearch={() => handleSearch(1)}
              loading={loading}
            />
          </Col>
          <Col span={8}>
            <Space>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => handleSearch(pagination.current)}
                loading={loading}
              >
                刷新
              </Button>
              <Button icon={<ClearOutlined />} onClick={handleClear}>
                清空
              </Button>
            </Space>
          </Col>
        </Row>

        {/* 搜索建议 */}
        {suggestions.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            <small>搜索建议：</small>
            {suggestions.map((suggestion, index) => (
              <Tag
                key={index}
                style={{ cursor: 'pointer', marginLeft: 8 }}
                onClick={() => {
                  setSearchForm((prev) => ({ ...prev, keyword: suggestion }))
                  setSuggestions([])
                  handleSearch(1)
                }}
              >
                {suggestion}
              </Tag>
            ))}
          </div>
        )}

        {/* 搜索历史 */}
        {searchHistory.length > 0 && !searchForm.keyword && (
          <div style={{ marginBottom: 16 }}>
            <small>搜索历史：</small>
            {searchHistory.map((history, index) => (
              <Tag
                key={index}
                style={{ cursor: 'pointer', marginLeft: 8 }}
                onClick={() => {
                  setSearchForm((prev) => ({ ...prev, keyword: history }))
                  handleSearch(1)
                }}
              >
                {history}
              </Tag>
            ))}
          </div>
        )}

        {/* 高级搜索 */}
        {showAdvanced && (
          <>
            <Divider />
            <Row gutter={16}>
              <Col span={6}>
                <Select
                  placeholder='案件类型'
                  style={{ width: '100%' }}
                  value={searchForm.caseType}
                  onChange={(value) => setSearchForm((prev) => ({ ...prev, caseType: value }))}
                  allowClear
                >
                  <Option value='民事案件'>民事案件</Option>
                  <Option value='刑事案件'>刑事案件</Option>
                  <Option value='行政案件'>行政案件</Option>
                  <Option value='经济案件'>经济案件</Option>
                </Select>
              </Col>
              <Col span={6}>
                <Select
                  placeholder='案件状态'
                  style={{ width: '100%' }}
                  value={searchForm.caseStatus}
                  onChange={(value) => setSearchForm((prev) => ({ ...prev, caseStatus: value }))}
                  allowClear
                >
                  <Option value='进行中'>进行中</Option>
                  <Option value='已完成'>已完成</Option>
                  <Option value='已暂停'>已暂停</Option>
                  <Option value='已取消'>已取消</Option>
                </Select>
              </Col>
              <Col span={6}>
                <Input
                  placeholder='负责律师'
                  value={searchForm.responsibleLawyer}
                  onChange={(e) =>
                    setSearchForm((prev) => ({ ...prev, responsibleLawyer: e.target.value }))
                  }
                  allowClear
                />
              </Col>
              <Col span={6}>
                <RangePicker
                  style={{ width: '100%' }}
                  value={searchForm.dateRange}
                  onChange={(dates) => setSearchForm((prev) => ({ ...prev, dateRange: dates }))}
                  placeholder={['开始日期', '结束日期']}
                />
              </Col>
            </Row>
            <Row gutter={16} style={{ marginTop: 16 }}>
              <Col span={12}>
                <InputNumber
                  placeholder='最小金额'
                  style={{ width: '48%', marginRight: '4%' }}
                  value={searchForm.amountRange?.[0]}
                  onChange={(value) =>
                    setSearchForm((prev) => ({
                      ...prev,
                      amountRange: [value || 0, prev.amountRange?.[1] || 0],
                    }))
                  }
                />
                <InputNumber
                  placeholder='最大金额'
                  style={{ width: '48%' }}
                  value={searchForm.amountRange?.[1]}
                  onChange={(value) =>
                    setSearchForm((prev) => ({
                      ...prev,
                      amountRange: [prev.amountRange?.[0] || 0, value || 0],
                    }))
                  }
                />
              </Col>
              <Col span={12}>
                <Button
                  type='primary'
                  onClick={() => handleSearch(1)}
                  loading={loading}
                  style={{ marginRight: 8 }}
                >
                  开始搜索
                </Button>
                <Button onClick={handleClear}>清空条件</Button>
              </Col>
            </Row>
          </>
        )}

        {/* 搜索结果 */}
        <Divider />
        <Table
          columns={columns}
          dataSource={data}
          rowKey='id'
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: (page, pageSize) => {
              setPagination((prev) => ({ ...prev, current: page, pageSize }))
              handleSearch(page)
            },
          }}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  )
}

export default FullTextSearch
