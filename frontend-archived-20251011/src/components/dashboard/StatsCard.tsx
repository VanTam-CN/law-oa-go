import React from 'react';
import { Card, Badge } from 'react-bootstrap';

interface StatsCardProps {
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

const StatsCard: React.FC<StatsCardProps> = ({ 
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
    <Card className={`stat-card shadow-sm border-0 ${className}`}>
      <Card.Body>
        <div className="d-flex justify-content-between align-items-center mb-3">
          <div>
            <Card.Title className="mb-0 fs-6">{title}</Card.Title>
            <div className="number fs-2 fw-bold mt-2">{value}</div>
            {description && (
              <div className="small mt-1">{description}</div>
            )}
          </div>
          <div className={`stat-icon rounded-circle d-flex align-items-center justify-content-center ${getColorClasses()}`} style={{ width: '50px', height: '50px' }}>
            <i className={`${icon} fa-lg`}></i>
          </div>
        </div>
        {(trendValue || footer) && (
          <div className="mt-2">
            {trendValue && (
              <div className={`small ${getTrendClass()}`}>
                {getTrendIcon()}
                {trendValue}
              </div>
            )}
            {footer}
          </div>
        )}
      </Card.Body>
    </Card>
  );
};

export default StatsCard;