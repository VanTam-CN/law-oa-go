import React, { useState, useEffect, useCallback } from 'react';
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
  AutoComplete,
  Tooltip,
  Badge,
  message
} from 'antd';
import {
  SearchOutlined,
  FileTextOutlined,
  StarOutlined,
  StarFilled,
  HistoryOutlined,
  TagsOutlined,
  FilterOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import {
  LawItem,
  LawCategory,
  LawTag,
  LegalSearchRequest,
  LegalSearchResponse,
  getLaws,
  searchLaws,
  getLawCategories,
  getLawTags,
  getPopularSearches,
  addToFavorites,
  removeFromFavorites
} from '@/services/tools';
import './LawSearch.less';

const { Search } = Input;
const { Option } = Select;
const { TextArea } = Input;

const LawSearch: React.FC = () => {
  // 状态管理
  const [loading, setLoading] = useState<boolean>(false);
  const [searchResponse, setSearchResponse] = useState<LegalSearchResponse | null>(null);
  const [categories, setCategories] = useState<LawCategory[]>([]);
  const [tags, setTags] = useState<LawTag[]>([]);
  const [popularSearches, setPopularSearches] = useState<string[]>([]);
  const [searchSuggestions, setSearchSuggestions] = useState<string[]>([]);

  // 搜索参数
  const [searchParams, setSearchParams] = useState<LegalSearchRequest>({
    page: 1,
    pageSize: 20,
    sortBy: 'relevance',
    sortOrder: 'desc'
  });

  // 加载初始数据
  useEffect(() => {
    loadInitialData();
  }, []);

  // 执行搜索
  useEffect(() => {
    if (searchParams.query || searchParams.categoryId || searchParams.tags?.length) {
      performSearch();
    } else {
      // 没有搜索条件时，获取热门法条
      loadPopularLaws();
    }
  }, [searchParams]);

  const loadInitialData = async () => {
    try {
      setLoading(true);
      const [categoriesRes, tagsRes, searchesRes] = await Promise.all([
        getLawCategories(),
        getLawTags(),
        getPopularSearches(10)
      ]);

      setCategories(categoriesRes.data || []);
      setTags(tagsRes.data || []);
      setPopularSearches(searchesRes.data || []);
    } catch (error) {
      console.error('Failed to load initial data:', error);
      message.error('加载初始数据失败');
    } finally {
      setLoading(false);
    }
  };

  const performSearch = async () => {
    try {
      setLoading(true);
      const response = await searchLaws(searchParams);
      setSearchResponse(response.data);
    } catch (error) {
      console.error('Failed to search laws:', error);
      message.error('搜索失败，请稍后重试');
      setSearchResponse(null);
    } finally {
      setLoading(false);
    }
  };

  const loadPopularLaws = async () => {
    try {
      setLoading(true);
      const response = await getLaws({
        page: 1,
        pageSize: 20,
        sortBy: 'relevance',
        sortOrder: 'desc'
      });
      setSearchResponse(response.data);
    } catch (error) {
      console.error('Failed to load popular laws:', error);
      setSearchResponse(null);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = useCallback((value: string) => {
    setSearchParams(prev => ({
      ...prev,
      query: value,
      page: 1
    }));
  }, []);

  const handleCategoryChange = useCallback((categoryId: number) => {
    setSearchParams(prev => ({
      ...prev,
      categoryId: categoryId || undefined,
      page: 1
    }));
  }, []);

  const handleTagChange = useCallback((selectedTags: string[]) => {
    setSearchParams(prev => ({
      ...prev,
      tags: selectedTags,
      page: 1
    }));
  }, []);

  const handleSortChange = useCallback((value: string) => {
    const [sortBy, sortOrder] = value.split('-');
    setSearchParams(prev => ({
      ...prev,
      sortBy: sortBy as 'relevance' | 'date' | 'title',
      sortOrder: sortOrder as 'asc' | 'desc',
      page: 1
    }));
  }, []);

  const handlePageChange = useCallback((page: number, pageSize?: number) => {
    setSearchParams(prev => ({
      ...prev,
      page,
      pageSize: pageSize || prev.pageSize
    }));
  }, []);

  const handleFavorite = async (record: LawItem) => {
    try {
      if (record.isFavorited) {
        await removeFromFavorites(record.id);
        message.success('已取消收藏');
        // 更新本地状态
        if (searchResponse) {
          setSearchResponse(prev => ({
            ...prev!,
            statutes: prev.statutes.map(item =>
              item.id === record.id ? { ...item, isFavorited: false } : item
            )
          }));
        }
      } else {
        await addToFavorites(record.id);
        message.success('已添加到收藏');
        // 更新本地状态
        if (searchResponse) {
          setSearchResponse(prev => ({
            ...prev!,
            statutes: prev.statutes.map(item =>
              item.id === record.id ? { ...item, isFavorited: true } : item
            )
          }));
        }
      }
    } catch (error) {
      console.error('Failed to update favorite status:', error);
      message.error('操作失败，请稍后重试');
    }
  };

  const handleSuggestionSelect = useCallback((value: string) => {
    handleSearch(value);
  }, [handleSearch]);

  const getCategoryColor = (code: string) => {
    const colors: { [key: string]: string } = {
      'CIVIL_LAW': 'blue',
      'CRIMINAL_LAW': 'red',
      'ADMINISTRATIVE_LAW': 'orange',
      'ECONOMIC_LAW': 'green',
      'LABOR_LAW': 'purple',
      'COMMERCIAL_LAW': 'cyan',
      'OTHER': 'default'
    };
    return colors[code] || 'default';
  };

  const columns: ColumnsType<LawItem> = [
    {
      title: '法条编号',
      dataIndex: 'statuteNumber',
      key: 'statuteNumber',
      width: 180,
      render: (text: string) => (
        <Tooltip title={text}>
          <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>
            {text}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: LawItem) => (
        <div>
          <div style={{ fontWeight: 500, marginBottom: 4 }}>
            {text}
          </div>
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
      render: (category: LawCategory) => category ? (
        <Tag color={getCategoryColor(category.code)}>
          {category.name}
        </Tag>
      ) : null,
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: 200,
      render: (tags: string[]) => (
        <div>
          {tags.slice(0, 2).map(tag => (
            <Tag key={tag} size="small" style={{ marginBottom: 2 }}>
              {tag}
            </Tag>
          ))}
          {tags.length > 2 && (
            <Tooltip title={tags.slice(2).join(', ')}>
              <Tag size="small">+{tags.length - 2}</Tag>
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
          type="text"
          size="small"
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
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => {
              // TODO: 实现查看详情功能
              console.log('查看详情:', record);
            }}
          >
            详情
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="law-search">
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <span>法条查询</span>
            <Badge count={popularSearches.length} showZero>
              <Button
                type="text"
                size="small"
                icon={<HistoryOutlined />}
                title="热门搜索"
              />
            </Badge>
          </div>
        }
        className="search-card"
      >
        {/* 搜索过滤器 */}
        <div className="search-filters">
          <div className="search-main">
            <AutoComplete
              style={{ width: '100%', maxWidth: 500 }}
              options={searchSuggestions.map(suggestion => ({ value: suggestion }))}
              onSelect={handleSuggestionSelect}
              onSearch={(value) => {
                // 这里可以实现搜索建议API调用
                setSearchSuggestions(popularSearches.filter(search =>
                  search.toLowerCase().includes(value.toLowerCase())
                ));
              }}
            >
              <Search
                placeholder="搜索法条编号、标题或内容关键词"
                allowClear
                enterButton={<SearchOutlined />}
                size="large"
                value={searchParams.query}
                onChange={(e) => handleSearch(e.target.value)}
                onSearch={handleSearch}
                style={{ marginBottom: 16 }}
              />
            </AutoComplete>
          </div>

          <div className="search-options">
            <Space wrap>
              <Select
                placeholder="选择分类"
                allowClear
                style={{ width: 150 }}
                value={searchParams.categoryId}
                onChange={handleCategoryChange}
                suffixIcon={<FilterOutlined />}
              >
                {categories.map(category => (
                  <Option key={category.id} value={category.id}>
                    {category.name}
                  </Option>
                ))}
              </Select>

              <Select
                mode="multiple"
                placeholder="选择标签"
                allowClear
                style={{ width: 200 }}
                value={searchParams.tags}
                onChange={handleTagChange}
                suffixIcon={<TagsOutlined />}
              >
                {tags.map(tag => (
                  <Option key={tag.id} value={tag.name}>
                    {tag.name}
                  </Option>
                ))}
              </Select>

              <Select
                defaultValue="relevance-desc"
                style={{ width: 120 }}
                onChange={handleSortChange}
              >
                <Option value="relevance-desc">相关度 ↓</Option>
                <Option value="relevance-asc">相关度 ↑</Option>
                <Option value="date-desc">最新 ↓</Option>
                <Option value="date-asc">最新 ↑</Option>
                <Option value="title-desc">标题 ↓</Option>
                <Option value="title-asc">标题 ↑</Option>
              </Select>
            </Space>
          </div>
        </div>

        {/* 热门搜索 */}
        {popularSearches.length > 0 && !searchParams.query && (
          <div className="popular-searches">
            <span style={{ marginRight: 8, color: '#666' }}>热门搜索：</span>
            <Space wrap>
              {popularSearches.map(search => (
                <Button
                  key={search}
                  type="link"
                  size="small"
                  onClick={() => handleSearch(search)}
                >
                  {search}
                </Button>
              ))}
            </Space>
          </div>
        )}

        {/* 搜索结果 */}
        <div className="search-results">
          <Spin spinning={loading}>
            {searchResponse && searchResponse.statutes.length > 0 ? (
              <>
                <div className="results-header">
                  <span>
                    找到 {searchResponse.total} 条结果
                    {searchParams.query && ` · 关键词: "${searchParams.query}"`}
                    {searchResponse.searchTime && ` · 耗时: ${searchResponse.searchTime}ms`}
                  </span>
                </div>

                <Table
                  columns={columns}
                  dataSource={searchResponse.statutes}
                  rowKey="id"
                  pagination={false}
                  size="middle"
                  expandable={{
                    expandedRowRender: (record) => (
                      <div className="law-content">
                        <div className="law-content-header">
                          <strong>法条内容：</strong>
                        </div>
                        <div className="law-content-text">
                          {record.content}
                        </div>
                        {record.keywords && record.keywords.length > 0 && (
                          <div className="law-keywords">
                            <span style={{ marginRight: 8 }}>关键词：</span>
                            {record.keywords.map(keyword => (
                              <Tag key={keyword} size="small" color="blue">
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
                  <div className="pagination-wrapper">
                    <Pagination
                      current={searchResponse.page}
                      pageSize={searchResponse.pageSize}
                      total={searchResponse.total}
                      showSizeChanger
                      showQuickJumper
                      showTotal={(total, range) =>
                        `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
                      }
                      onChange={handlePageChange}
                    />
                  </div>
                )}
              </>
            ) : (
              <Empty
                description={
                  loading ? '搜索中...' : (
                    searchParams.query ?
                      `未找到与 "${searchParams.query}" 相关的法条` :
                      '暂无数据'
                  )
                }
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Spin>
        </div>
      </Card>
    </div>
  );
};

export default LawSearch;