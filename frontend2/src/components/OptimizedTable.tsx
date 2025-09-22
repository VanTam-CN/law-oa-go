import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Table, Spin, Alert, Pagination, Card, Space, Button, Tooltip } from 'antd';
import { ReloadOutlined, ExportOutlined } from '@ant-design/icons';
import type { TableProps, PaginationProps } from 'antd';
import { debounce } from 'lodash';

// 分页参数接口
interface PaginationParams {
  current: number;
  pageSize: number;
  total: number;
}

// 查询参数接口
interface SearchParams {
  [key: string]: any;
}

// 表格数据源接口
interface DataSource {
  list: any[];
  total: number;
}

// 优化的表格组件属性
interface OptimizedTableProps extends Omit<TableProps<any>, 'dataSource'> {
  fetchData: (params: PaginationParams & SearchParams) => Promise<DataSource>;
  searchParams?: SearchParams;
  rowKey?: string;
  showPagination?: boolean;
  showRefresh?: boolean;
  showExport?: boolean;
  exportData?: () => Promise<void>;
  pageSizeOptions?: string[];
  defaultPageSize?: number;
  loading?: boolean;
  onRowClick?: (record: any, index: number) => void;
}

/**
 * 优化的表格组件
 * 支持虚拟滚动、分页、缓存、防抖等功能
 */
const OptimizedTable: React.FC<OptimizedTableProps> = ({
  fetchData,
  searchParams = {},
  rowKey = 'id',
  showPagination = true,
  showRefresh = true,
  showExport = false,
  exportData,
  pageSizeOptions = ['10', '20', '50', '100'],
  defaultPageSize = 20,
  loading = false,
  onRowClick,
  columns,
  ...tableProps
}) => {
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: defaultPageSize,
    total: 0,
  });

  // 数据状态
  const [dataSource, setDataSource] = useState<any[]>([]);
  const [tableLoading, setTableLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 缓存上一次的查询参数
  const [lastSearchParams, setLastSearchParams] = useState<SearchParams>({});
  const [lastPagination, setLastPagination] = useState(pagination);

  // 防抖的查询函数
  const debouncedFetchData = useCallback(
    debounce(async (params: PaginationParams & SearchParams) => {
      try {
        setTableLoading(true);
        setError(null);

        const { list, total } = await fetchData(params);
        
        setDataSource(list);
        setPagination(prev => ({ ...prev, total }));
        
        // 更新缓存
        setLastSearchParams(params);
        setLastPagination(params);
      } catch (err) {
        setError(err instanceof Error ? err.message : '加载数据失败');
        setDataSource([]);
        setPagination(prev => ({ ...prev, total: 0 }));
      } finally {
        setTableLoading(false);
      }
    }, 300),
    [fetchData]
  );

  // 查询数据
  const loadData = useCallback(
    (page: number = pagination.current, size: number = pagination.pageSize) => {
      const params = {
        current: page,
        pageSize: size,
        ...searchParams,
      };
      
      debouncedFetchData(params);
    },
    [pagination.current, pagination.pageSize, searchParams, debouncedFetchData]
  );

  // 初始化加载数据
  useEffect(() => {
    loadData();
  }, []);

  // 搜索参数变化时重新加载数据
  useEffect(() => {
    // 检查搜索参数是否真的发生了变化
    const searchParamsChanged = Object.keys(searchParams).some(key => 
      searchParams[key] !== lastSearchParams[key]
    );

    if (searchParamsChanged) {
      loadData(1, pagination.pageSize);
    }
  }, [searchParams, lastSearchParams, pagination.pageSize, loadData]);

  // 分页变化处理
  const handlePageChange: PaginationProps['onChange'] = (page, pageSize) => {
    setPagination(prev => ({ ...prev, current: page, pageSize }));
    loadData(page, pageSize);
  };

  // 刷新数据
  const handleRefresh = () => {
    loadData(pagination.current, pagination.pageSize);
  };

  // 导出数据
  const handleExport = async () => {
    if (exportData) {
      try {
        await exportData();
      } catch (err) {
        setError(err instanceof Error ? err.message : '导出数据失败');
      }
    }
  };

  // 计算表格高度
  const tableHeight = useMemo(() => {
    const headerHeight = 56;
    const paginationHeight = showPagination ? 56 : 0;
    const actionBarHeight = 40;
    const windowHeight = window.innerHeight;
    
    return windowHeight - headerHeight - paginationHeight - actionBarHeight - 100;
  }, [showPagination]);

  // 渲染操作按钮
  const renderActionBar = () => (
    <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
      <div />
      <Space>
        {showRefresh && (
          <Tooltip title="刷新">
            <Button 
              icon={<ReloadOutlined />} 
              onClick={handleRefresh}
              loading={tableLoading}
            />
          </Tooltip>
        )}
        {showExport && exportData && (
          <Tooltip title="导出">
            <Button 
              icon={<ExportOutlined />} 
              onClick={handleExport}
            >
              导出
            </Button>
          </Tooltip>
        )}
      </Space>
    </div>
  );

  // 分页配置
  const paginationConfig: PaginationProps = showPagination
    ? {
        current: pagination.current,
        pageSize: pagination.pageSize,
        total: pagination.total,
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total, range) => 
          `显示 ${range[0]}-${range[1]} 条，共 ${total} 条`,
        pageSizeOptions,
        onChange: handlePageChange,
        onShowSizeChange: (current, size) => {
          setPagination(prev => ({ ...prev, pageSize: size }));
          loadData(current, size);
        },
      }
    : false;

  // 表格配置
  const tableConfig: TableProps<any> = {
    columns,
    dataSource,
    rowKey,
    loading: tableLoading || loading,
    pagination: paginationConfig,
    scroll: { x: 'max-content', y: tableHeight },
    rowClassName: (record, index) => 
      `table-row ${index % 2 === 0 ? 'table-row-even' : 'table-row-odd'}`,
    onRow: (record, index) => ({
      onClick: () => onRowClick?.(record, index!),
    }),
    ...tableProps,
  };

  return (
    <Card style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {renderActionBar()}
      
      {error && (
        <Alert
          message="错误"
          description={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 16 }}
        />
      )}
      
      <div style={{ flex: 1, overflow: 'hidden' }}>
        <Table {...tableConfig} />
      </div>
      
      <style jsx>{`
        .table-row {
          transition: background-color 0.2s;
        }
        
        .table-row:hover {
          background-color: #f5f5f5;
        }
        
        .table-row-even {
          background-color: #fafafa;
        }
        
        .ant-table-tbody > tr.table-row-odd > td {
          background-color: #fff;
        }
        
        .ant-table-tbody > tr.table-row-even > td {
          background-color: #fafafa;
        }
      `}</style>
    </Card>
  );
};

export default React.memo(OptimizedTable);