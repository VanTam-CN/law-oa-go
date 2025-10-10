import React, { useState, useEffect } from 'react';
import { 
  Card, 
  List, 
  Tag, 
  Button, 
  Space, 
  Popconfirm,
  Tooltip,
  Badge,
  Empty,
  Typography,
  Divider
} from 'antd';
import { 
  HistoryOutlined, 
  DeleteOutlined, 
  SearchOutlined,
  ClockCircleOutlined,
  ReloadOutlined,
  ClearOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Text } = Typography;

interface SearchHistoryItem {
  id: string;
  searchText: string;
  filters: any;
  timestamp: string;
  resultCount: number;
}

interface SearchHistoryProps {
  onSelect: (history: SearchHistoryItem) => void;
  maxItems?: number;
}

const SearchHistory: React.FC<SearchHistoryProps> = ({ 
  onSelect, 
  maxItems = 10 
}) => {
  const [histories, setHistories] = useState<SearchHistoryItem[]>([]);
  const [loading, setLoading] = useState(false);

  // 从localStorage加载搜索历史
  useEffect(() => {
    loadSearchHistory();
  }, []);

  const loadSearchHistory = () => {
    try {
      const stored = localStorage.getItem('case-search-history');
      if (stored) {
        const historyData = JSON.parse(stored);
        setHistories(historyData.slice(0, maxItems));
      }
    } catch (error) {
      console.error('加载搜索历史失败:', error);
    }
  };

  const saveSearchHistory = (newHistories: SearchHistoryItem[]) => {
    try {
      localStorage.setItem('case-search-history', JSON.stringify(newHistories));
      setHistories(newHistories.slice(0, maxItems));
    } catch (error) {
      console.error('保存搜索历史失败:', error);
    }
  };

  // 添加搜索历史
  const addSearchHistory = (searchParams: any, resultCount: number) => {
    const historyItem: SearchHistoryItem = {
      id: `history_${Date.now()}`,
      searchText: searchParams.searchText || '',
      filters: searchParams,
      timestamp: new Date().toISOString(),
      resultCount: resultCount
    };

    const newHistories = [historyItem, ...histories.filter(h => 
      h.searchText !== historyItem.searchText || 
      JSON.stringify(h.filters) !== JSON.stringify(historyItem.filters)
    )];

    saveSearchHistory(newHistories);
  };

  // 删除单条历史
  const deleteHistory = (id: string) => {
    const newHistories = histories.filter(h => h.id !== id);
    saveSearchHistory(newHistories);
  };

  // 清空历史
  const clearAllHistory = () => {
    saveSearchHistory([]);
  };

  // 格式化筛选条件显示
  const formatFilters = (filters: any) => {
    const filterTags = [];
    
    if (filters.searchText) {
      filterTags.push(<Tag key="search" color="blue">关键词: {filters.searchText}</Tag>);
    }
    if (filters.caseType) {
      filterTags.push(<Tag key="caseType" color="orange">类型: {getCaseTypeLabel(filters.caseType)}</Tag>);
    }
    if (filters.status) {
      filterTags.push(<Tag key="status" color="green">状态: {getStatusLabel(filters.status)}</Tag>);
    }
    if (filters.projectType) {
      filterTags.push(<Tag key="projectType" color="purple">项目: {getProjectTypeLabel(filters.projectType)}</Tag>);
    }
    if (filters.lawyerId) {
      filterTags.push(<Tag key="lawyer" color="cyan">律师: {filters.lawyerId}</Tag>);
    }
    if (filters.clientId) {
      filterTags.push(<Tag key="client" color="geekblue">客户: {filters.clientId}</Tag>);
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      filterTags.push(<Tag key="dateRange" color="magenta">日期范围</Tag>);
    }
    if (filters.amountRange && (filters.amountRange[0] || filters.amountRange[1])) {
      filterTags.push(<Tag key="amountRange" color="gold">金额范围</Tag>);
    }

    return filterTags.length > 0 ? filterTags : <Text type="secondary">无筛选条件</Text>;
  };

  const getCaseTypeLabel = (type: string) => {
    const labels: { [key: string]: string } = {
      'CIVIL': '民事',
      'COMMERCIAL': '商事',
      'CRIMINAL': '刑事',
      'ADMINISTRATIVE': '行政',
      'ADVISORY': '咨询',
      'REVIEW': '审查'
    };
    return labels[type] || type;
  };

  const getStatusLabel = (status: string) => {
    const labels: { [key: string]: string } = {
      '0': '未开始',
      '1': '进行中',
      '2': '已结案',
      '3': '已归档'
    };
    return labels[status] || status;
  };

  const getProjectTypeLabel = (type: string) => {
    const labels: { [key: string]: string } = {
      'CIVIL': '民事诉讼',
      'COMMERCIAL': '商事仲裁',
      'CRIMINAL': '刑事辩护',
      'ADMINISTRATIVE': '行政诉讼',
      'ADVISORY': '咨询项目',
      'REVIEW': '审查项目'
    };
    return labels[type] || type;
  };

  // 导出方法供父组件调用
  React.useImperativeHandle(React.useRef<any>(), () => ({
    addSearchHistory
  }));

  return (
    <Card
      title={
        <Space>
          <HistoryOutlined />
          搜索历史
          <Badge count={histories.length} size="small" />
        </Space>
      }
      size="small"
      extra={
        <Space>
          <Tooltip title="刷新">
            <Button 
              type="text" 
              icon={<ReloadOutlined />} 
              onClick={loadSearchHistory}
              size="small"
            />
          </Tooltip>
          {histories.length > 0 && (
            <Popconfirm
              title="确定要清空所有搜索历史吗？"
              onConfirm={clearAllHistory}
              okText="确定"
              cancelText="取消"
            >
              <Tooltip title="清空历史">
                <Button 
                  type="text" 
                  icon={<ClearOutlined />} 
                  size="small"
                  danger
                />
              </Tooltip>
            </Popconfirm>
          )}
        </Space>
      }
    >
      {histories.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="暂无搜索历史"
        />
      ) : (
        <List
          size="small"
          dataSource={histories}
          renderItem={(item) => (
            <List.Item
              key={item.id}
              actions={[
                <Popconfirm
                  key="delete"
                  title="确定要删除这条历史记录吗？"
                  onConfirm={() => deleteHistory(item.id)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Tooltip title="删除">
                    <Button 
                      type="text" 
                      icon={<DeleteOutlined />} 
                      size="small"
                      danger
                    />
                  </Tooltip>
                </Popconfirm>,
                <Tooltip key="reuse" title="重新搜索">
                  <Button 
                    type="text" 
                    icon={<SearchOutlined />} 
                    size="small"
                    onClick={() => onSelect(item)}
                  />
                </Tooltip>
              ]}
            >
              <List.Item.Meta
                title={
                  <Space>
                    <Text strong>{item.searchText || '空关键词搜索'}</Text>
                    <Badge 
                      count={item.resultCount} 
                      style={{ backgroundColor: '#f0f0f0', color: '#999' }}
                    />
                  </Space>
                }
                description={
                  <div>
                    <div style={{ marginBottom: 4 }}>
                      {formatFilters(item.filters)}
                    </div>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      <ClockCircleOutlined /> {dayjs(item.timestamp).format('YYYY-MM-DD HH:mm')}
                    </Text>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      )}
    </Card>
  );
};

// 创建一个HOC来提供搜索历史功能
export const withSearchHistory = (WrappedComponent: React.ComponentType<any>) => {
  return (props: any) => {
    const [searchHistoryRef] = useState(React.useRef<any>());

    const addToHistory = (searchParams: any, resultCount: number) => {
      if (searchHistoryRef.current?.addSearchHistory) {
        searchHistoryRef.current.addSearchHistory(searchParams, resultCount);
      }
    };

    return (
      <>
        <WrappedComponent 
          {...props} 
          addToSearchHistory={addToHistory}
        />
        <div style={{ marginTop: 16 }}>
          <SearchHistory 
            ref={searchHistoryRef}
            onSelect={(history) => {
              props.onSearch && props.onSearch(history.filters);
            }}
          />
        </div>
      </>
    );
  };
};

export default SearchHistory;