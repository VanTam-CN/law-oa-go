import React from 'react';
import { Spinner, Table } from 'react-bootstrap';

interface DataTableProps {
  columns: Array<{
    key: string;
    title: string;
    render?: (value: any, row: any) => React.ReactNode;
  }>;
  data: any[];
  loading?: boolean;
  onRowClick?: (row: any) => void;
  emptyMessage?: string;
}

const DataTable: React.FC<DataTableProps> = ({ 
  columns, 
  data, 
  loading = false, 
  onRowClick,
  emptyMessage = 'No data found'
}) => {
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
      </div>
    );
  }

  return (
    <Table striped bordered hover responsive>
      <thead>
        <tr>
          {columns.map(column => (
            <th key={column.key}>{column.title}</th>
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
    </Table>
  );
};

export default DataTable;