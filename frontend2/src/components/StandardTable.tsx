import React, { useState, useCallback, useMemo } from 'react';
import { 
  Card, 
  Table, 
  Input, 
  Select, 
  Button, 
  Space, 
  Tag, 
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
  Col,
  Avatar,
  Image
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
  CloseOutlined,
  ExportOutlined,
  ImportOutlined,
  SettingOutlined,
  ColumnHeightOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined
} from '@ant-design/icons';
import type { TableProps, ColumnsType } from 'antd/es/table';
import type { PaginationProps } from 'antd/es/pagination';
import './StandardTable.less';

const { Text } = Typography;
const { Search } = Input;
const { Option } = Select;

export interface TableRecord {
  key: string | number;
  [key: string]: any;
}

export interface ColumnConfig {
  title: string;
  dataIndex: string;
  key?: string;
  width?: number | string;
  fixed?: 'left' | 'right';
  sorter?: boolean | ((a: any, b: any) => number);
  filters?: Array<{ text: string; value: string }>;
  filterMultiple?: boolean;
  filterSearch?: boolean;
  render?: (value: any, record: TableRecord, index: number) => React.ReactNode;
  ellipsis?: boolean;
  align?: 'left' | 'center' | 'right';
  className?: string;
  hidden?: boolean;
}

export interface StandardTableProps extends Omit<TableProps<TableRecord>, 'columns' | 'dataSource'> {
  data: TableRecord[];
  columns: ColumnConfig[];
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
  onRowClick?: (record: TableRecord, index: number) => void;
  onRowDoubleClick?: (record: TableRecord, index: number) => void;
  emptyText?: string;
  size?: 'small' | 'middle' | 'large';
  scroll?: TableProps['scroll'];
  actions?: React.ReactNode[];
  filters?: Array<{
    label: string;
    value: string;
    options: Array<{ label: string; value: string }>;
  }>;
  statusColors?: Record<string, string>;
  priorityColors?: Record<string, string>;
  exportable?: boolean;
  importable?: boolean;
  onExport?: () => void;
  onImport?: () => void;
  columnSettings?: boolean;
  density?: boolean;
  fullscreen?: boolean;
}

const StandardTable: React.FC<StandardTableProps> = ({
  data,
  columns,
  loading = false,
  title,
  extra,
  searchable = true,
  filterable = true,
  selectable = false,
  multiSelect = true,
  showHeader = true,
  showFooter = true,
  showPagination = true,
  pagination,
  rowSelection,
  onSearch,
  onFilter,
  onRowClick,
  onRowDoubleClick,
  emptyText = '暂无数据',
  size = 'middle',
  scroll,
  actions,
  filters = [],
  statusColors = {},
  priorityColors = {},
  exportable = false,
  importable = false,
  onExport,
  onImport,
  columnSettings = true,
  density = true,
  fullscreen = false,
  ...restProps
}) => {
  const [searchValue, setSearchValue] = useState('');
  const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
  const [selectedRows, setSelectedRows] = useState<(string | number)[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [tableSize, setTableSize] = useState<'small' | 'middle' | 'large'>(size);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [visibleColumns, setVisibleColumns] = useState(
    columns.map(col => ({ ...col, visible: !col.hidden }))
  );

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
  const handleSelect = useCallback((keys: (string | number)[]) => {
    setSelectedRows(keys);
    rowSelection?.onChange?.(keys);
  }, [rowSelection]);

  // 处理全屏
  const handleFullscreen = useCallback(() => {
    if (!isFullscreen) {
      const elem = document.documentElement;
      if (elem.requestFullscreen) {
        elem.requestFullscreen();
      }
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
    setIsFullscreen(!isFullscreen);
  }, [isFullscreen]);

  // 处理列显示/隐藏
  const handleColumnVisibility = useCallback((key: string, visible: boolean) => {
    setVisibleColumns(prev => 
      prev.map(col => col.key === key ? { ...col, visible } : col)
    );
  }, []);

  // 过滤后的数据
  const filteredData = useMemo(() => {
    return data.filter(record => {
      // 搜索过滤
      if (searchValue) {
        const searchLower = searchValue.toLowerCase();
        const matchText = Object.values(record).some(value => 
          String(value).toLowerCase().includes(searchLower)
        );
        if (!matchText) return false;
      }

      // 其他过滤条件
      for (const [key, value] of Object.entries(activeFilters)) {
        if (value && record[key] !== value) {
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

  // 处理列配置
  const tableColumns: ColumnsType<TableRecord> = useMemo(() => {
    return visibleColumns
      .filter(col => col.visible)
      .map(col => {
        const column: ColumnsType<TableRecord>[0] = {
          title: col.title,
          dataIndex: col.dataIndex,
          key: col.key || col.dataIndex,
          width: col.width,
          fixed: col.fixed,
          sorter: col.sorter,
          filters: col.filters,
          filterMultiple: col.filterMultiple,
          filterSearch: col.filterSearch,
          ellipsis: col.ellipsis,
          align: col.align,
          className: col.className
        };

        // 自定义渲染
        if (col.render) {
          column.render = col.render;
        } else {
          // 默认渲染逻辑
          column.render = (value, record) => {
            // 状态标签
            if (col.dataIndex === 'status' && defaultStatusColors[value]) {
              return <Tag color={defaultStatusColors[value]}>{value}</Tag>;
            }
            
            // 优先级标签
            if (col.dataIndex === 'priority' && defaultPriorityColors[value]) {
              const priorityText = value === 'high' ? '高' : value === 'medium' ? '中' : '低';
              return <Tag color={defaultPriorityColors[value]}>{priorityText}</Tag>;
            }
            
            // 头像
            if (col.dataIndex === 'avatar' && value) {
              return <Avatar src={value} size="small" />;
            }
            
            // 图片
            if (col.dataIndex === 'image' && value) {
              return <Image src={value} width={50} height={50} style={{ objectFit: 'cover' }} />;
            }
            
            // 金额
            if (col.dataIndex === 'amount' || col.dataIndex.includes('price') || col.dataIndex.includes('money')) {
              return value ? `¥${Number(value).toLocaleString()}` : '-';
            }
            
            // 日期
            if (col.dataIndex.includes('date') || col.dataIndex.includes('time')) {
              return value ? new Date(value).toLocaleDateString() : '-';
            }
            
            // 布尔值
            if (typeof value === 'boolean') {
              return value ? <CheckOutlined style={{ color: 'var(--color-success)' }} /> : <CloseOutlined style={{ color: 'var(--color-error)' }} />;
            }
            
            return value || '-';
          };
        }

        return column;
      });
  }, [visibleColumns, defaultStatusColors, defaultPriorityColors]);

  // 行操作菜单
  const getRowMenu = (record: TableRecord) => (
    <Menu>
      <Menu.Item key="view" icon={<EyeOutlined />} onClick={() => onRowClick?.(record, 0)}>
        查看详情
      </Menu.Item>
      <Menu.Item key="edit" icon={<EditOutlined />} onClick={() => {}}>
        编辑
      </Menu.Item>
      <Menu.Item key="delete" icon={<DeleteOutlined />} danger onClick={() => {}}>
        删除
      </Menu.Item>
    </Menu>
  );

  // 列设置菜单
  const getColumnSettingsMenu = () => (
    <Menu>
      {visibleColumns.map(col => (
        <Menu.Item key={col.key}>
          <Checkbox
            checked={col.visible}
            onChange={(e) => handleColumnVisibility(col.key || col.dataIndex, e.target.checked)}
          >
            {col.title}
          </Checkbox>
        </Menu.Item>
      ))}
    </Menu>
  );

  // 密度设置菜单
  const getDensityMenu = () => (
    <Menu>
      <Menu.Item key="large" onClick={() => setTableSize('large')}>
        <span>宽松</span>
      </Menu.Item>
      <Menu.Item key="middle" onClick={() => setTableSize('middle')}>
        <span>中等</span>
      </Menu.Item>
      <Menu.Item key="small" onClick={() => setTableSize('small')}>
        <span>紧凑</span>
      </Menu.Item>
    </Menu>
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

  // 行选择配置
  const rowSelectionConfig = selectable ? {
    selectedRowKeys: selectedRows,
    onChange: handleSelect,
    type: multiSelect ? 'checkbox' : 'radio' as const,
    columnWidth: 50,
    fixed: 'left'
  } : undefined;

  return (
    <Card 
      className={`standard-table ${isFullscreen ? 'fullscreen' : ''}`}
      title={
        showHeader && (
          <div className="table-header">
            <div className="header-left">
              {title && <Text strong className="table-title">{title}</Text>}
            </div>
            <div className="header-right">
              {actions}
              {extra}
            </div>
          </div>
        )
      }
      extra={
        <div className="table-toolbar">
          <Space wrap>
            {/* 搜索 */}
            {searchable && (
              <Search
                placeholder="搜索表格内容"
                allowClear
                enterButton={<SearchOutlined />}
                size="small"
                style={{ width: 200 }}
                value={searchValue}
                onChange={(e) => setSearchValue(e.target.value)}
                onSearch={handleSearch}
              />
            )}
            
            {/* 过滤器 */}
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
            
            {/* 刷新 */}
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
            
            {/* 导入导出 */}
            {importable && (
              <Button 
                icon={<ImportOutlined />} 
                size="small"
                onClick={onImport}
              >
                导入
              </Button>
            )}
            
            {exportable && (
              <Button 
                icon={<ExportOutlined />} 
                size="small"
                onClick={onExport}
              >
                导出
              </Button>
            )}
            
            {/* 列设置 */}
            {columnSettings && (
              <Dropdown overlay={getColumnSettingsMenu()} trigger={['click']}>
                <Button icon={<SettingOutlined />} size="small">
                  列设置
                </Button>
              </Dropdown>
            )}
            
            {/* 密度设置 */}
            {density && (
              <Dropdown overlay={getDensityMenu()} trigger={['click']}>
                <Button icon={<ColumnHeightOutlined />} size="small">
                  密度
                </Button>
              </Dropdown>
            )}
            
            {/* 全屏 */}
            {fullscreen && (
              <Button 
                icon={isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />} 
                size="small"
                onClick={handleFullscreen}
              >
                {isFullscreen ? '退出全屏' : '全屏'}
              </Button>
            )}
          </Space>
        </div>
      }
    >
      <Spin spinning={loading}>
        {paginatedData.length > 0 ? (
          <Table
            {...restProps}
            columns={tableColumns}
            dataSource={paginatedData}
            rowSelection={rowSelectionConfig}
            size={tableSize}
            scroll={scroll}
            pagination={showPagination ? paginationConfig : false}
            onRow={(record, index) => ({
              onClick: () => onRowClick?.(record, index || 0),
              onDoubleClick: () => onRowDoubleClick?.(record, index || 0),
              className: selectable ? 'selectable-row' : ''
            })}
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
        <div className="table-footer">
          <div className="footer-left">
            {selectable && selectedRows.length > 0 && (
              <Text type="secondary">
                已选择 {selectedRows.length} 项
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

export default StandardTable;