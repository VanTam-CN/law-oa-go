import React from 'react';
import { Spinner, Button, Table as BootstrapTable } from 'react-bootstrap';

interface TableColumn {
  key: string;
  title: string;
  render?: (value: any, row: any) => React.ReactNode;
  sortable?: boolean;
  width?: string;
}

interface TableProps {
  columns: TableColumn[];
  data: any[];
  loading?: boolean;
  onSort?: (columnKey: string) => void;
  sortColumn?: string;
  sortOrder?: 'asc' | 'desc';
  onRowClick?: (row: any) => void;
  emptyMessage?: string;
  className?: string;
  striped?: boolean;
  bordered?: boolean;
  hover?: boolean;
  responsive?: boolean;
}

const Table: React.FC<TableProps> = ({ 
  columns, 
  data, 
  loading = false, 
  onSort,
  sortColumn,
  sortOrder,
  onRowClick,
  emptyMessage = 'No data found',
  className = '',
  striped = true,
  bordered = true,
  hover = true,
  responsive = true
}) => {
  const handleSort = (columnKey: string) => {
    if (onSort) {
      onSort(columnKey);
    }
  };

  const getSortIcon = (columnKey: string) => {
    if (sortColumn === columnKey) {
      return sortOrder === 'asc' ? (
        <i className="fas fa-sort-up ms-1"></i>
      ) : (
        <i className="fas fa-sort-down ms-1"></i>
      );
    }
    return <i className="fas fa-sort ms-1"></i>;
  };

  if (loading) {
    return (
      <div className="d-flex justify-content-center align-items-center" style={{ height: '400px' }}>
        <Spinner animation="border" />
        <span className="ms-2">Loading data...</span>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div className="text-center py-5">
        <i className="fas fa-database fa-3x text-muted mb-3"></i>
        <h5>{emptyMessage}</h5>
        <p className="text-muted">Try adjusting your search or filter criteria</p>
        <Button variant="primary">
          <i className="fas fa-plus me-2"></i>
          Add New Record
        </Button>
      </div>
    );
  }

  return (
    <div className={className}>
      <BootstrapTable 
        striped={striped} 
        bordered={bordered} 
        hover={hover} 
        responsive={responsive}
      >
        <thead>
          <tr>
            {columns.map(column => (
              <th 
                key={column.key} 
                onClick={() => column.sortable && handleSort(column.key)}
                style={{ cursor: column.sortable ? 'pointer' : 'default', width: column.width }}
              >
                {column.title}
                {column.sortable && getSortIcon(column.key)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, rowIndex) => (
            <tr 
              key={rowIndex} 
              onClick={() => onRowClick && onRowClick(row)}
              className={onRowClick ? 'cursor-pointer' : ''}
            >
              {columns.map(column => (
                <td key={`${rowIndex}-${column.key}`}>
                  {column.render ? column.render(row[column.key], row) : row[column.key]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </BootstrapTable>
    </div>
  );
};

export default Table;