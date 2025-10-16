import React from 'react';
import { Pagination as BootstrapPagination } from 'react-bootstrap';

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  className?: string;
}

const Pagination: React.FC<PaginationProps> = ({ 
  currentPage, 
  totalPages, 
  onPageChange,
  className = ''
}) => {
  const getPageNumbers = () => {
    const pages = [];
    const maxVisiblePages = 5;
    
    if (totalPages <= maxVisiblePages) {
      // Show all pages
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Show first page, current page, and last page with ellipses
      if (currentPage <= 3) {
        // Near the beginning
        for (let i = 1; i <= Math.min(maxVisiblePages - 1, totalPages); i++) {
          pages.push(i);
        }
        if (totalPages > maxVisiblePages) {
          pages.push('...');
          pages.push(totalPages);
        }
      } else if (currentPage >= totalPages - 2) {
        // Near the end
        pages.push(1);
        pages.push('...');
        for (let i = totalPages - (maxVisiblePages - 2); i <= totalPages; i++) {
          pages.push(i);
        }
      } else {
        // In the middle
        pages.push(1);
        pages.push('...');
        for (let i = currentPage - 1; i <= currentPage + 1; i++) {
          pages.push(i);
        }
        pages.push('...');
        pages.push(totalPages);
      }
    }
    
    return pages;
  };

  const handlePageChange = (page: number | string) => {
    if (typeof page === 'number' && page >= 1 && page <= totalPages) {
      onPageChange(page);
    }
  };

  if (totalPages <= 1) {
    return null;
  }

  return (
    <div className={`d-flex justify-content-center ${className}`}>
      <BootstrapPagination>
        <BootstrapPagination.First 
          onClick={() => handlePageChange(1)} 
          disabled={currentPage === 1}
        />
        <BootstrapPagination.Prev 
          onClick={() => handlePageChange(currentPage - 1)} 
          disabled={currentPage === 1}
        />
        
        {getPageNumbers().map((page, index) => (
          <BootstrapPagination.Item
            key={index}
            active={page === currentPage}
            onClick={() => typeof page === 'number' && handlePageChange(page)}
            disabled={page === '...'}
          >
            {page}
          </BootstrapPagination.Item>
        ))}
        
        <BootstrapPagination.Next 
          onClick={() => handlePageChange(currentPage + 1)} 
          disabled={currentPage === totalPages}
        />
        <BootstrapPagination.Last 
          onClick={() => handlePageChange(totalPages)} 
          disabled={currentPage === totalPages}
        />
      </BootstrapPagination>
    </div>
  );
};

export default Pagination;