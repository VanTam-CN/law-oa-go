import React, { useState, useCallback, useMemo } from 'react';
import { 
  Card, 
  List, 
  Avatar, 
  Tag, 
  Button, 
  Input, 
  Select, 
  Space, 
  Typography, 
  Empty, 
  Spin, 
  Pagination, 
  Dropdown, 
  Menu,
  Badge,
  Tooltip,
  Switch,
  Checkbox,
  Row,
  Col
} from 'antd';
import { 
  SearchOutlined, 
  FilterOutlined, 
  MoreOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  EyeOutlined, 
  DownloadOutlined,
  PlusOutlined,
  ReloadOutlined,
  CheckOutlined,
  CloseOutlined
} from '@ant-design/icons';
import type { ListProps } from 'antd/es/list';
import type { PaginationProps } from 'antd/es/pagination';
import './StandardList.less';

const { Text, Paragraph } = Typography;
const { Search } = Input;
const { Option } = Select;

export interface ListItem {
  id: string | number;
  title: string;
  description?: string;
  content?: string;
  avatar?: string;
  icon?: React.ReactNode;
  tags?: string[];
  status?: string;
  priority?: 'high' | 'medium' | 'low';
  datetime?: string;
  author?: string;
  actions?: React.ReactNode[];
  extra?: React.ReactNode;
  cover?: React.ReactNode;
  meta?: Record<string, any>;
  selected?: boolean;
  disabled?: boolean;
}

export interface StandardListProps extends Omit<ListProps<ListItem>, 'dataSource'> {
  data: ListItem[];
  loading?: boolean;
  title?: string;
  extra?: React.ReactNode;
  searchable?: boolean;
  filterable?: boolean;
  selectable?: boolean;
  multiSelect?: boolean;
  showHeader?: boolean;
  showFooter?: boolean;
  showPagination?: boolean;
  pagination?: false | PaginationProps;
  rowSelection?: {
    selectedRowKeys?: (string | number)[];
    onChange?: (selectedRowKeys: (string | number)[]) => void;
  };
  onSearch?: (value: string) => void;
  onFilter?: (filters: Record<string, any>) => void;
  onItemClick?: (item: ListItem) => void;
  onItemDoubleClick?: (item: ListItem) => void;
  renderItem?: (item: ListItem) => React.ReactNode;
  emptyText?: string;
  size?: 'small' | 'default' | 'large';
  grid?: ListProps['grid'];
  actions?: React.ReactNode[];
  filters?: Array<{
    label: string;
    value: string;
    options: Array<{ label: string; value: string }>;
  }>;
  statusColors?: Record<string, string>;
  priorityColors?: Record<string, string>;
}

const StandardList: React.FC<StandardListProps> = ({
  data,
  loading = false,
  title,
  extra,
  searchable = true,
  filterable = true,
  selectable = false,
  multiSelect = false,
  showHeader = true,
  showFooter = true,
  showPagination = true,
  pagination,
  rowSelection,
  onSearch,
  onFilter,
  onItemClick,
  onItemDoubleClick,
  renderItem,
  emptyText = '暂无数据',
  size = 'default',
  grid,
  actions,
  filters = [],
  statusColors = {},
  priorityColors = {},
  ...restProps
}) => {
  const [searchValue, setSearchValue] = useState('');
  const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
  const [selectedItems, setSelectedItems] = useState<(string | number)[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // 默认状态颜色
  const defaultStatusColors: Record<string, string> = {
    '进行中': 'processing',
    '已完成': 'success',
    '已暂停': 'warning',
    '已取消': 'error',
    '待处理': 'default',
    '处理中': 'processing',
    '已处理': 'success',
    '已拒绝': 'error',
    ...statusColors
  };

  // 默认优先级颜色
  const defaultPriorityColors: Record<string, string> = {
    'high': 'error',
    'medium': 'warning',
    'low': 'success',
    ...priorityColors
  };

  // 处理搜索
  const handleSearch = useCallback((value: string) => {
    setSearchValue(value);
    onSearch?.(value);
  }, [onSearch]);

  // 处理过滤
  const handleFilter = useCallback((key: string, value: any) => {
    const newFilters = { ...activeFilters, [key]: value };
    if (value === undefined || value === null || value === '') {
      delete newFilters[key];
    }
    setActiveFilters(newFilters);
    onFilter?.(newFilters);
  }, [activeFilters, onFilter]);

  // 处理选择
  const handleSelect = useCallback((item: ListItem, checked: boolean) => {
    if (!multiSelect) {
      const newSelected = checked ? [item.id] : [];
      setSelectedItems(newSelected);
      rowSelection?.onChange?.(newSelected);
      return;
    }

    let newSelected = [...selectedItems];
    if (checked) {
      newSelected.push(item.id);
    } else {
      newSelected = newSelected.filter(id => id !== item.id);
    }
    setSelectedItems(newSelected);
    rowSelection?.onChange?.(newSelected);
  }, [selectedItems, multiSelect, rowSelection]);

  // 处理全选
  const handleSelectAll = useCallback((checked: boolean) => {
    const newSelected = checked ? data.map(item => item.id) : [];
    setSelectedItems(newSelected);
    rowSelection?.onChange?.(newSelected);
  }, [data, rowSelection]);

  // 过滤后的数据
  const filteredData = useMemo(() => {
    return data.filter(item => {
      // 搜索过滤
      if (searchValue) {
        const searchLower = searchValue.toLowerCase();
        const matchText = item.title.toLowerCase().includes(searchLower) ||
                          item.description?.toLowerCase().includes(searchLower) ||
                          item.content?.toLowerCase().includes(searchLower) ||
                          item.tags?.some(tag => tag.toLowerCase().includes(searchLower));
        if (!matchText) return false;
      }

      // 其他过滤条件
      for (const [key, value] of Object.entries(activeFilters)) {
        if (value && item[key as keyof ListItem] !== value) {
          return false;
        }
      }

      return true;
    });
  }, [data, searchValue, activeFilters]);

  // 分页数据
  const paginatedData = useMemo(() => {
    if (!showPagination || pagination === false) {
      return filteredData;
    }

    const startIndex = (currentPage - 1) * pageSize;
    return filteredData.slice(startIndex, startIndex + pageSize);
  }, [filteredData, currentPage, pageSize, showPagination, pagination]);

  // 列表项操作菜单
  const getItemMenu = (item: ListItem) => (
    <Menu>
      <Menu.Item key="view" icon={<EyeOutlined />} onClick={() => onItemClick?.(item)}>
        查看详情
      </Menu.Item>
      <Menu.Item key="edit" icon={<EditOutlined />} onClick={() => {}}>
        编辑
      </Menu.Item>
      <Menu.Item key="download" icon={<DownloadOutlined />} onClick={() => {}}>
        下载
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item key="delete" icon={<DeleteOutlined />} danger onClick={() => {}}>
        删除
      </Menu.Item>
    </Menu>
  );

  // 默认渲染项
  const defaultRenderItem = (item: ListItem) => (
    <List.Item
      key={item.id}
      className={`list-item ${item.selected ? 'selected' : ''} ${item.disabled ? 'disabled' : ''}`}
      onClick={() => {
        if (!item.disabled) {
          onItemClick?.(item);
        }
      }}
      onDoubleClick={() => {
        if (!item.disabled) {
          onItemDoubleClick?.(item);
        }
      }}
      actions={[
        <Tooltip title="查看详情" key="view">
          <Button 
            type="text" 
            icon={<EyeOutlined />} 
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              onItemClick?.(item);
            }}
          />
        </Tooltip>,
        <Dropdown 
          key="more" 
          overlay={getItemMenu(item)} 
          trigger={['click']}
          placement="bottomRight"
        >
          <Button 
            type="text" 
            icon={<MoreOutlined />} 
            size="small"
            onClick={(e) => e.stopPropagation()}
          />
        </Dropdown>
      ]}
      extra={
        selectable && (
          <Checkbox
            checked={selectedItems.includes(item.id)}
            onChange={(e) => handleSelect(item, e.target.checked)}
            onClick={(e) => e.stopPropagation()}
          />
        )
      }
    >
      <List.Item.Meta
        avatar={
          item.avatar ? (
            <Avatar src={item.avatar} />
          ) : item.icon ? (
            <Avatar 
              style={{ 
                backgroundColor: defaultStatusColors[item.status || 'default'] === 'success' ? 'var(--color-success-100)' :
                               defaultStatusColors[item.status || 'default'] === 'error' ? 'var(--color-error-100)' :
                               defaultStatusColors[item.status || 'default'] === 'warning' ? 'var(--color-warning-100)' :
                               defaultStatusColors[item.status || 'default'] === 'processing' ? 'var(--color-primary-100)' :
                               'var(--color-gray-100)',
                color: defaultStatusColors[item.status || 'default'] === 'success' ? 'var(--color-success-700)' :
                       defaultStatusColors[item.status || 'default'] === 'error' ? 'var(--color-error-700)' :
                       defaultStatusColors[item.status || 'default'] === 'warning' ? 'var(--color-warning-700)' :
                       defaultStatusColors[item.status || 'default'] === 'processing' ? 'var(--color-primary-700)' :
                       'var(--color-gray-700)'
              }}
            >
              {item.icon}
            </Avatar>
          ) : (
            <Avatar 
              style={{ 
                backgroundColor: 'var(--color-primary-100)',
                color: 'var(--color-primary-700)'
              }}
            >
              {item.title.charAt(0)}
            </Avatar>
          )
        }
        title={
          <div className="list-item-title">
            <Text strong>{item.title}</Text>
            {item.priority && (
              <Tag color={defaultPriorityColors[item.priority]} className="priority-tag">
                {item.priority === 'high' ? '高' : item.priority === 'medium' ? '中' : '低'}
              </Tag>
            )}
            {item.status && (
              <Tag color={defaultStatusColors[item.status]} className="status-tag">
                {item.status}
              </Tag>
            )}
          </div>
        }
        description={
          <div className="list-item-description">
            {item.description && (
              <Paragraph 
                ellipsis={{ rows: 2, expandable: false }} 
                className="description-text"
              >
                {item.description}
              </Paragraph>
            )}
            <div className="list-item-meta">
              {item.author && (
                <span className="meta-item">
                  <Text type="secondary">作者: {item.author}</Text>
                </span>
              )}
              {item.datetime && (
                <span className="meta-item">
                  <Text type="secondary">{item.datetime}</Text>
                </span>
              )}
              {item.tags && item.tags.length > 0 && (
                <div className="tags-container">
                  {item.tags.map((tag, index) => (
                    <Tag key={index} size="small" className="item-tag">
                      {tag}
                    </Tag>
                  ))}
                </div>
              )}
            </div>
          </div>
        }
      />
      {item.content && (
        <div className="list-item-content">
          <Paragraph ellipsis={{ rows: 3, expandable: false }}>
            {item.content}
          </Paragraph>
        </div>
      )}
      {item.cover && (
        <div className="list-item-cover">
          {item.cover}
        </div>
      )}
      {item.extra && (
        <div className="list-item-extra">
          {item.extra}
        </div>
      )}
    </List.Item>
  );

  // 分页配置
  const paginationConfig: PaginationProps = pagination === false ? {} as PaginationProps : {
    current: currentPage,
    pageSize,
    total: filteredData.length,
    showSizeChanger: true,
    showQuickJumper: true,
    showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
    pageSizeOptions: ['10', '20', '50', '100'],
    onChange: (page, size) => {
      setCurrentPage(page);
      if (size) setPageSize(size);
    },
    ...pagination
  };

  return (
    <Card 
      className={`standard-list ${size} ${selectable ? 'selectable' : ''}`}
      title={
        showHeader && (
          <div className="list-header">
            <div className="header-left">
              {selectable && (
                <Checkbox
                  checked={selectedItems.length > 0 && selectedItems.length === data.length}
                  indeterminate={selectedItems.length > 0 && selectedItems.length < data.length}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                  className="select-all-checkbox"
                >
                  全选
                </Checkbox>
              )}
              {title && <Text strong className="list-title">{title}</Text>}
            </div>
            <div className="header-right">
              {actions}
              {extra}
            </div>
          </div>
        )
      }
      extra={
        showHeader && (
          <div className="list-toolbar">
            {(searchable || filterable) && (
              <Space wrap>
                {searchable && (
                  <Search
                    placeholder="搜索标题、描述或标签"
                    allowClear
                    enterButton={<SearchOutlined />}
                    size="small"
                    style={{ width: 200 }}
                    value={searchValue}
                    onChange={(e) => setSearchValue(e.target.value)}
                    onSearch={handleSearch}
                  />
                )}
                {filterable && filters.map(filter => (
                  <Select
                    key={filter.label}
                    placeholder={filter.label}
                    allowClear
                    size="small"
                    style={{ width: 120 }}
                    value={activeFilters[filter.value]}
                    onChange={(value) => handleFilter(filter.value, value)}
                  >
                    {filter.options.map(option => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                ))}
                <Button 
                  icon={<ReloadOutlined />} 
                  size="small"
                  onClick={() => {
                    setSearchValue('');
                    setActiveFilters({});
                    onSearch?.('');
                    onFilter?.({});
                  }}
                >
                  重置
                </Button>
              </Space>
            )}
          </div>
        )
      }
    >
      <Spin spinning={loading}>
        {paginatedData.length > 0 ? (
          <List
            {...restProps}
            dataSource={paginatedData}
            renderItem={renderItem || defaultRenderItem}
            grid={grid}
            pagination={showPagination ? paginationConfig : false}
            locale={{
              emptyText: <Empty description={emptyText} />
            }}
          />
        ) : (
          <Empty 
            description={emptyText}
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        )}
      </Spin>
      
      {showFooter && (
        <div className="list-footer">
          <div className="footer-left">
            {selectable && selectedItems.length > 0 && (
              <Text type="secondary">
                已选择 {selectedItems.length} 项
              </Text>
            )}
          </div>
          <div className="footer-right">
            <Text type="secondary">
              共 {filteredData.length} 项
            </Text>
          </div>
        </div>
      )}
    </Card>
  );
};

export default StandardList;