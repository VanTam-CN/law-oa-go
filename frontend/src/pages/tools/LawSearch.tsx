import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Input,
  Select,
  Table,
  Tag,
  Space,
  Button,
  Spin,
  Empty,
  Pagination,
  Tooltip,
  message,
  Modal,
  Descriptions,
  Typography,
} from 'antd'
import {
  SearchOutlined,
  FileTextOutlined,
  StarOutlined,
  StarFilled,
  TagsOutlined,
  FilterOutlined,
  ImportOutlined,
  DownloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  LawItem,
  LawCategory,
  LawTag,
  LegalSearchRequest,
  LegalSearchResponse,
  LegalStatuteImportItem,
  LegalStatuteImportResponse,
  getLaws,
  searchLaws,
  getLawById,
  getLawCategories,
  getLawTags,
  getPopularSearches,
  addToFavorites,
  removeFromFavorites,
  bulkImportStatutes,
} from '@/services/tools'
import './LawSearch.less'

const { Search } = Input
const { Option } = Select
const { TextArea } = Input
const { Paragraph } = Typography

const LawSearch: React.FC = () => {
  // 状态管理
  const [loading, setLoading] = useState<boolean>(false)
  const [searchResponse, setSearchResponse] = useState<LegalSearchResponse | null>(null)
  const [categories, setCategories] = useState<LawCategory[]>([])
  const [tags, setTags] = useState<LawTag[]>([])
  const [popularSearches, setPopularSearches] = useState<string[]>([])

  const [detailModalVisible, setDetailModalVisible] = useState<boolean>(false)
  const [detailLoading, setDetailLoading] = useState<boolean>(false)
  const [selectedLaw, setSelectedLaw] = useState<LawItem | null>(null)

  // 导入相关状态
  const [importModalVisible, setImportModalVisible] = useState<boolean>(false)
  const [importJson, setImportJson] = useState<string>('')
  const [importing, setImporting] = useState<boolean>(false)
  const [importResult, setImportResult] = useState<LegalStatuteImportResponse | null>(null)

  // 搜索参数
  const [searchParams, setSearchParams] = useState<LegalSearchRequest>({
    page: 1,
    pageSize: 20,
    sortBy: 'relevance',
    sortOrder: 'desc',
  })

  // 加载初始数据
  useEffect(() => {
    loadInitialData()
  }, [])

  // 执行搜索
  useEffect(() => {
    if (searchParams.query || searchParams.categoryId || searchParams.tags?.length) {
      performSearch()
    } else {
      // 没有搜索条件时，获取热门法条
      loadPopularLaws()
    }
  }, [searchParams])

  const loadInitialData = async () => {
    try {
      setLoading(true)
      const [categoriesRes, tagsRes, popularRes] = await Promise.all([
        getLawCategories(),
        getLawTags(),
        getPopularSearches(8),
      ])

      const categoryList = categoriesRes?.data || []
      const tagList = tagsRes?.data || []
      setCategories(categoryList)
      setTags(tagList)

      const terms = popularRes?.data?.length
        ? popularRes.data
        : [...categoryList.map((category) => category.name), ...tagList.map((tag) => tag.name)]

      setPopularSearches(Array.from(new Set(terms.filter(Boolean))).slice(0, 8))
    } catch (error) {
      console.error('Failed to load initial data:', error)
      message.error('加载初始数据失败')
      setPopularSearches([])
    } finally {
      setLoading(false)
    }
  }

  const performSearch = async () => {
    try {
      setLoading(true)
      const response = await searchLaws(searchParams)
      setSearchResponse(response?.data || null)
    } catch (error) {
      console.error('Failed to search laws:', error)
      message.error('搜索失败，请稍后重试')
      setSearchResponse(null)
    } finally {
      setLoading(false)
    }
  }

  const loadPopularLaws = async () => {
    try {
      setLoading(true)
      const response = await getLaws({
        page: 1,
        pageSize: 20,
        sortBy: 'relevance',
        sortOrder: 'desc',
      })
      setSearchResponse(response?.data || null)
    } catch (error) {
      console.error('Failed to load popular laws:', error)
      setSearchResponse(null)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = useCallback((value: string) => {
    setSearchParams((prev) => ({
      ...prev,
      query: value,
      page: 1,
    }))
  }, [])

  const handleCategoryChange = useCallback((categoryId: number) => {
    setSearchParams((prev) => ({
      ...prev,
      categoryId: categoryId || undefined,
      page: 1,
    }))
  }, [])

  const handleTagChange = useCallback((selectedTags: string[]) => {
    setSearchParams((prev) => ({
      ...prev,
      tags: selectedTags,
      page: 1,
    }))
  }, [])

  const handleSortChange = useCallback((value: string) => {
    const [sortBy, sortOrder] = value.split('-')
    setSearchParams((prev) => ({
      ...prev,
      sortBy: sortBy as 'relevance' | 'date' | 'title',
      sortOrder: sortOrder as 'asc' | 'desc',
      page: 1,
    }))
  }, [])

  const handlePageChange = useCallback((page: number, pageSize?: number) => {
    setSearchParams((prev) => ({
      ...prev,
      page,
      pageSize: pageSize || prev.pageSize,
    }))
  }, [])

  const handleFavorite = async (record: LawItem) => {
    try {
      if (record.isFavorited) {
        await removeFromFavorites(record.id)
        message.success('已取消收藏')
        // 更新本地状态
        setSearchResponse((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            statutes: prev.statutes.map((item) =>
              item.id === record.id ? { ...item, isFavorited: false } : item,
            ),
          }
        })
      } else {
        await addToFavorites(record.id)
        message.success('已添加到收藏')
        // 更新本地状态
        setSearchResponse((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            statutes: prev.statutes.map((item) =>
              item.id === record.id ? { ...item, isFavorited: true } : item,
            ),
          }
        })
      }
    } catch (error) {
      console.error('Failed to update favorite status:', error)
      message.error('操作失败，请稍后重试')
    }
  }

  // 打开导入弹窗
  const handleOpenImportModal = () => {
    setImportModalVisible(true)
    setImportJson('')
    setImportResult(null)
  }

  // 导入示例数据
  const handleLoadExample = () => {
    const example = [
      {
        statute_number: '民法典-001',
        title: '民事权利能力',
        content: '自然人从出生时起到死亡时止，具有民事权利能力，依法享有民事权利，承担民事义务。',
        category_code: 'CIVIL_LAW',
        law_name: '民法典',
        chapter: '总则',
        status: 'active',
        tags: ['总则', '权利能力'],
        keywords: ['民事权利', '权利能力', '自然人'],
      },
      {
        statute_number: '民法典-002',
        title: '民事行为能力',
        content: '成年人为完全民事行为能力人，可以独立实施民事法律行为。',
        category_code: 'CIVIL_LAW',
        law_name: '民法典',
        chapter: '总则',
        status: 'active',
        tags: ['总则', '行为能力'],
        keywords: ['民事行为', '行为能力', '成年人'],
      },
      {
        statute_number: '刑法-001',
        title: '故意杀人罪',
        content: '故意杀人的，处死刑、无期徒刑或者十年以上有期徒刑。',
        category_code: 'CRIMINAL_LAW',
        law_name: '刑法',
        chapter: '侵犯公民人身权利、民主权利罪',
        status: 'active',
        tags: ['刑事', '人身权利'],
        keywords: ['故意杀人', '死刑', '有期徒刑'],
      },
    ]
    setImportJson(JSON.stringify(example, null, 2))
  }

  // 下载导入模板
  const handleDownloadTemplate = () => {
    const template = [
      {
        statute_number: '示例-001',
        title: '法条标题示例',
        content: '法条内容示例',
        category_code: 'CIVIL_LAW',
        law_name: '法律名称',
        chapter: '章（可选）',
        section: '节（可选）',
        part: '编（可选）',
        effective_date: '2020-01-01',
        status: 'active',
        tags: ['标签1', '标签2'],
        keywords: ['关键词1', '关键词2'],
      },
    ]
    const blob = new Blob([JSON.stringify(template, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = '法条导入模板.json'
    a.click()
    URL.revokeObjectURL(url)
    message.success('模板下载成功')
  }

  // 执行导入
  const handleImport = async () => {
    if (!importJson.trim()) {
      message.warning('请输入JSON数据')
      return
    }

    let statutes: LegalStatuteImportItem[]
    try {
      statutes = JSON.parse(importJson)
    } catch (err) {
      message.error('JSON格式错误，请检查数据格式')
      return
    }

    if (!Array.isArray(statutes)) {
      message.error('JSON数据格式错误：根节点必须是数组')
      return
    }

    if (statutes.length === 0) {
      message.warning('请至少输入一条法条数据')
      return
    }

    if (statutes.length > 1000) {
      message.warning('单次最多导入1000条法条')
      return
    }

    try {
      setImporting(true)
      const response = await bulkImportStatutes({ statutes })
      const result = response.data
      setImportResult(result)

      if (result.failure_count === 0) {
        message.success(`导入成功！共导入 ${result.success_count} 条法条，耗时 ${result.processing_time_ms}ms`)
        // 刷新搜索结果
        performSearch()
      } else if (result.success_count === 0) {
        message.error('导入失败，请检查数据格式')
      } else {
        message.warning(`导入完成：成功 ${result.success_count} 条，失败 ${result.failure_count} 条`)
        performSearch()
      }
    } catch (error: any) {
      console.error('导入失败:', error)
      message.error(error.response?.data?.message || '导入失败，请稍后重试')
    } finally {
      setImporting(false)
    }
  }

  const handleViewDetail = async (record: LawItem) => {
    setSelectedLaw(record)
    setDetailModalVisible(true)

    try {
      setDetailLoading(true)
      const response = await getLawById(record.id)
      setSelectedLaw(response.data)
    } catch (error) {
      console.error('Failed to load law detail:', error)
      message.error('获取法条详情失败')
    } finally {
      setDetailLoading(false)
    }
  }

  const getCategoryColor = (code: string) => {
    const colors: { [key: string]: string } = {
      CIVIL_LAW: 'blue',
      CRIMINAL_LAW: 'red',
      ADMINISTRATIVE_LAW: 'orange',
      ECONOMIC_LAW: 'green',
      LABOR_LAW: 'purple',
      COMMERCIAL_LAW: 'cyan',
      OTHER: 'default',
    }
    return colors[code] || 'default'
  }

  const columns: ColumnsType<LawItem> = [
    {
      title: '法条编号',
      dataIndex: 'statuteNumber',
      key: 'statuteNumber',
      width: 180,
      render: (text: string) => (
        <Tooltip title={text}>
          <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{text}</span>
        </Tooltip>
      ),
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: LawItem) => (
        <div>
          <div style={{ fontWeight: 500, marginBottom: 4 }}>{text}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>
            {record.lawName}
            {record.chapter && ` · ${record.chapter}`}
          </div>
        </div>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 120,
      render: (category: LawCategory) =>
        category ? <Tag color={getCategoryColor(category.code)}>{category.name}</Tag> : null,
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: 200,
      render: (tags: string[]) => (
        <div>
          {tags.slice(0, 2).map((tag) => (
            <Tag key={tag} style={{ marginBottom: 2, fontSize: '12px' }}>
              {tag}
            </Tag>
          ))}
          {tags.length > 2 && (
            <Tooltip title={tags.slice(2).join(', ')}>
              <Tag style={{ fontSize: '12px' }}>+{tags.length - 2}</Tag>
            </Tooltip>
          )}
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '生效' : '失效'}
        </Tag>
      ),
    },
    {
      title: '收藏',
      key: 'favorite',
      width: 80,
      render: (_, record: LawItem) => (
        <Button
          type='text'
          size='small'
          icon={record.isFavorited ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined />}
          onClick={() => handleFavorite(record)}
        />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record: LawItem) => (
        <Space size='small'>
          <Button
            type='link'
            size='small'
            icon={<FileTextOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            详情
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div className='law-search'>
      <Card
        title='法条查询'
        className='search-card'
        extra={
          <Button
            type='primary'
            icon={<ImportOutlined />}
            onClick={handleOpenImportModal}
          >
            批量导入
          </Button>
        }
      >
        {/* 搜索过滤器 - 单行布局 */}
        <div className='search-filters'>
          <Space size='middle' style={{ width: '100%' }}>
            <Select
              placeholder='选择分类'
              allowClear
              style={{ width: 120 }}
              value={searchParams.categoryId}
              onChange={handleCategoryChange}
              suffixIcon={<FilterOutlined />}
            >
              {categories.map((category) => (
                <Option key={category.id} value={category.id}>
                  {category.name}
                </Option>
              ))}
            </Select>

            <Select
              mode='multiple'
              placeholder='选择标签'
              allowClear
              style={{ width: 150 }}
              value={searchParams.tags}
              onChange={handleTagChange}
              suffixIcon={<TagsOutlined />}
            >
              {tags.map((tag) => (
                <Option key={tag.id} value={tag.name}>
                  {tag.name}
                </Option>
              ))}
            </Select>

            <Select
              defaultValue='relevance-desc'
              style={{ width: 100 }}
              onChange={handleSortChange}
            >
              <Option value='relevance-desc'>相关度 ↓</Option>
              <Option value='relevance-asc'>相关度 ↑</Option>
              <Option value='date-desc'>最新 ↓</Option>
              <Option value='date-asc'>最新 ↑</Option>
              <Option value='title-desc'>标题 ↓</Option>
              <Option value='title-asc'>标题 ↑</Option>
            </Select>

            <Search
              placeholder='搜索法条编号、标题或内容关键词'
              allowClear
              enterButton={<SearchOutlined />}
              value={searchParams.query}
              onChange={(e) => handleSearch(e.target.value)}
              onSearch={handleSearch}
              style={{ width: 500 }}
            />
          </Space>
        </div>

        {popularSearches.length > 0 && (
          <div className='popular-searches'>
            <Space size='small' wrap>
              <span>常用检索：</span>
              {popularSearches.map((term) => (
                <Button key={term} type='link' size='small' onClick={() => handleSearch(term)}>
                  {term}
                </Button>
              ))}
            </Space>
          </div>
        )}

        {/* 搜索结果 */}
        <div className='search-results'>
          <Spin spinning={loading}>
            {searchResponse && searchResponse.statutes.length > 0 ? (
              <>
                <div className='results-header'>
                  <span>
                    找到 {searchResponse.total} 条结果
                    {searchParams.query && ` · 关键词: "${searchParams.query}"`}
                    {searchResponse.searchTime && ` · 耗时: ${searchResponse.searchTime}ms`}
                  </span>
                </div>

                <Table
                  columns={columns}
                  dataSource={searchResponse.statutes}
                  rowKey='id'
                  pagination={false}
                  size='middle'
                  expandable={{
                    expandedRowRender: (record) => (
                      <div className='law-content'>
                        <div className='law-content-header'>
                          <strong>法条内容：</strong>
                        </div>
                        <div className='law-content-text'>{record.content}</div>
                        {record.keywords && record.keywords.length > 0 && (
                          <div className='law-keywords'>
                            <span style={{ marginRight: 8 }}>关键词：</span>
                            {record.keywords.map((keyword) => (
                              <Tag key={keyword} color='blue' style={{ fontSize: '12px' }}>
                                {keyword}
                              </Tag>
                            ))}
                          </div>
                        )}
                      </div>
                    ),
                    rowExpandable: (record) => true,
                  }}
                />

                {searchResponse.total > searchResponse.pageSize && (
                  <div className='pagination-wrapper'>
                    <Pagination
                      current={searchResponse.page}
                      pageSize={searchResponse.pageSize}
                      total={searchResponse.total}
                      showSizeChanger
                      showQuickJumper
                      showTotal={(total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`}
                      onChange={handlePageChange}
                    />
                  </div>
                )}
              </>
            ) : (
              <Empty
                description={
                  loading
                    ? '搜索中...'
                    : searchParams.query
                      ? `未找到与 "${searchParams.query}" 相关的法条`
                      : '暂无数据'
                }
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Spin>
        </div>
      </Card>

      <Modal
        title={selectedLaw?.title || '法条详情'}
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key='close' onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={860}
        destroyOnClose
      >
        <Spin spinning={detailLoading}>
          {selectedLaw && (
            <Space direction='vertical' size='middle' style={{ width: '100%' }}>
              <Descriptions bordered size='small' column={2}>
                <Descriptions.Item label='法条编号'>{selectedLaw.statuteNumber}</Descriptions.Item>
                <Descriptions.Item label='法律名称'>{selectedLaw.lawName}</Descriptions.Item>
                <Descriptions.Item label='分类'>
                  {selectedLaw.category ? (
                    <Tag color={getCategoryColor(selectedLaw.category.code)}>{selectedLaw.category.name}</Tag>
                  ) : (
                    '-'
                  )}
                </Descriptions.Item>
                <Descriptions.Item label='状态'>
                  <Tag color={selectedLaw.status === 'active' ? 'green' : 'red'}>
                    {selectedLaw.status === 'active' ? '生效' : '失效'}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label='章节' span={2}>
                  {[selectedLaw.part, selectedLaw.chapter, selectedLaw.section].filter(Boolean).join(' / ') || '-'}
                </Descriptions.Item>
                <Descriptions.Item label='发布机关' span={2}>
                  {selectedLaw.publishingAuthority || '-'}
                </Descriptions.Item>
              </Descriptions>

              <div className='law-content'>
                <div className='law-content-header'>
                  <strong>法条内容：</strong>
                </div>
                <Paragraph className='law-content-text'>{selectedLaw.content}</Paragraph>
              </div>

              {selectedLaw.tags.length > 0 && (
                <div>
                  <span style={{ marginRight: 8 }}>标签：</span>
                  {selectedLaw.tags.map((tag) => (
                    <Tag key={tag}>{tag}</Tag>
                  ))}
                </div>
              )}

              {selectedLaw.keywords.length > 0 && (
                <div>
                  <span style={{ marginRight: 8 }}>关键词：</span>
                  {selectedLaw.keywords.map((keyword) => (
                    <Tag key={keyword} color='blue'>{keyword}</Tag>
                  ))}
                </div>
              )}
            </Space>
          )}
        </Spin>
      </Modal>

      {/* 批量导入弹窗 */}
      <Modal
        title='批量导入法条'
        open={importModalVisible}
        onCancel={() => setImportModalVisible(false)}
        footer={[
          <Button key='cancel' onClick={() => setImportModalVisible(false)}>
            取消
          </Button>,
          <Button
            key='example'
            onClick={handleLoadExample}
            disabled={importing}
          >
            加载示例
          </Button>,
          <Button
            key='template'
            onClick={handleDownloadTemplate}
            disabled={importing}
            icon={<DownloadOutlined />}
          >
            下载模板
          </Button>,
          <Button
            key='import'
            type='primary'
            onClick={handleImport}
            loading={importing}
            icon={<CheckCircleOutlined />}
          >
            开始导入
          </Button>,
        ]}
        width={800}
        destroyOnClose
      >
        <div className='import-modal-content'>
          <div className='import-description'>
            <p>请输入JSON格式的法条数据。单次最多导入1000条法条。</p>
            <p>您可以点击"加载示例"查看数据格式，或点击"下载模板"获取模板文件。</p>
          </div>

          <TextArea
            value={importJson}
            onChange={(e) => setImportJson(e.target.value)}
            placeholder='请输入JSON格式的法条数据...'
            rows={15}
            style={{ fontFamily: 'monospace', fontSize: '13px' }}
            disabled={importing}
          />

          {importResult && (
            <div className='import-result'>
              <div
                className={`import-result-header ${
                  importResult.failure_count === 0 ? 'success' : importResult.success_count === 0 ? 'error' : 'warning'
                }`}
              >
                {importResult.failure_count === 0 ? (
                  <>
                    <CheckCircleOutlined style={{ marginRight: 8 }} />
                    导入成功！
                  </>
                ) : importResult.success_count === 0 ? (
                  <>
                    <CloseCircleOutlined style={{ marginRight: 8 }} />
                    导入失败
                  </>
                ) : (
                  <>
                    <CheckCircleOutlined style={{ marginRight: 8 }} />
                    部分成功
                  </>
                )}
              </div>
              <div className='import-result-stats'>
                <p>总计: {importResult.total_count} 条</p>
                <p style={{ color: '#52c41a' }}>
                  成功: {importResult.success_count} 条
                </p>
                <p style={{ color: '#ff4d4f' }}>
                  失败: {importResult.failure_count} 条
                </p>
                <p>耗时: {importResult.processing_time_ms}ms</p>
              </div>

              {importResult.errors && importResult.errors.length > 0 && (
                <div className='import-result-errors'>
                  <h4>错误详情:</h4>
                  <ul>
                    {importResult.errors.slice(0, 10).map((err, index) => (
                      <li key={index}>
                        <strong>{err.statute_number}:</strong> {err.message}
                      </li>
                    ))}
                    {importResult.errors.length > 10 && (
                      <li>...还有 {importResult.errors.length - 10} 条错误</li>
                    )}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>
      </Modal>
    </div>
  )
}

export default LawSearch
