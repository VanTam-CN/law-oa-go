import React from 'react';
import { Card } from 'react-bootstrap';

interface StatCardProps {
  title: string;
  value: number | string;
  icon: string;
  color: string;
  trend?: 'up' | 'down';
  trendValue?: string;
  description?: string;
}

const StatCard: React.FC<StatCardProps> = ({ 
  title, 
  value, 
  icon, 
  color,
  trend,
  trendValue,
  description
}) => {
  const getColorClasses = () => {
    switch (color) {
      case 'primary': return 'bg-primary text-white';
      case 'success': return 'bg-success text-white';
      case 'warning': return 'bg-warning text-white';
      case 'danger': return 'bg-danger text-white';
      case 'info': return 'bg-info text-white';
      default: return 'bg-secondary text-white';
    }
  };

  const getTrendIcon = () => {
    if (trend === 'up') {
      return <i className="fas fa-arrow-up me-1"></i>;
    } else if (trend === 'down') {
      return <i className="fas fa-arrow-down me-1"></i>;
    }
    return null;
  };

  const getTrendClass = () => {
    if (trend === 'up') {
      return 'text-success';
    } else if (trend === 'down') {
      return 'text-danger';
    }
    return 'text-muted';
  };

  return (
    <Card className={`stat-card ${getColorClasses()}`}>
      <Card.Body>
        <div className="d-flex justify-content-between align-items-center">
          <div>
            <Card.Title className="mb-0 fs-6">{title}</Card.Title>
            <div className="number fs-2 fw-bold mt-2">{value}</div>
            {description && (
              <div className="small mt-1">{description}</div>
            )}
          </div>
          <div className="stat-icon">
            <i className={`${icon} fa-2x`}></i>
          </div>
        </div>
        {trendValue && (
          <div className="mt-3">
            <span className={`small ${getTrendClass()}`}>
              {getTrendIcon()}
              {trendValue}
            </span>
          </div>
        )}
      </Card.Body>
    </Card>
  );
};

export default StatCard;