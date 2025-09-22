import React from 'react';
import { Card, Badge } from 'react-bootstrap';

interface DataCardProps {
  title: string;
  value: number | string;
  icon: string;
  color: string;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  description?: string;
  footer?: React.ReactNode;
  className?: string;
}

const DataCard: React.FC<DataCardProps> = ({ 
  title, 
  value, 
  icon, 
  color,
  trend,
  trendValue,
  description,
  footer,
  className = ''
}) => {
  const getColorClasses = () => {
    switch (color) {
      case 'primary': return 'bg-primary text-white';
      case 'success': return 'bg-success text-white';
      case 'warning': return 'bg-warning text-white';
      case 'danger': return 'bg-danger text-white';
      case 'info': return 'bg-info text-white';
      case 'secondary': return 'bg-secondary text-white';
      default: return 'bg-secondary text-white';
    }
  };

  const getTrendIcon = () => {
    if (trend === 'up') {
      return <i className="fas fa-arrow-up me-1"></i>;
    } else if (trend === 'down') {
      return <i className="fas fa-arrow-down me-1"></i>;
    } else if (trend === 'neutral') {
      return <i className="fas fa-minus me-1"></i>;
    }
    return null;
  };

  const getTrendClass = () => {
    if (trend === 'up') {
      return 'text-success';
    } else if (trend === 'down') {
      return 'text-danger';
    } else if (trend === 'neutral') {
      return 'text-muted';
    }
    return 'text-muted';
  };

  return (
    <Card className={`data-card shadow-sm border-0 ${className}`}>
      <Card.Body>
        <div className="d-flex justify-content-between align-items-center mb-3">
          <div className={`rounded-circle d-flex align-items-center justify-content-center ${getColorClasses()}`} style={{ width: '50px', height: '50px' }}>
            <i className={`${icon} fa-lg`}></i>
          </div>
          <div className="text-end">
            <div className="fs-4 fw-bold">{value}</div>
            {trendValue && (
              <div className={`small ${getTrendClass()}`}>
                {getTrendIcon()}
                {trendValue}
              </div>
            )}
          </div>
        </div>
        <div>
          <h6 className="mb-1">{title}</h6>
          {description && <p className="text-muted small mb-0">{description}</p>}
        </div>
      </Card.Body>
      {footer && (
        <Card.Footer className="bg-transparent border-top-0 p-0 pt-2">
          {footer}
        </Card.Footer>
      )}
    </Card>
  );
};

export default DataCard;