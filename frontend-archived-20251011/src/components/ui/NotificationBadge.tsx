import React from 'react';
import { Badge } from 'react-bootstrap';

interface NotificationBadgeProps {
  count: number;
  variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'info';
  className?: string;
  maxCount?: number;
}

const NotificationBadge: React.FC<NotificationBadgeProps> = ({ 
  count, 
  variant = 'danger',
  className = '',
  maxCount = 99
}) => {
  if (count === 0) {
    return null;
  }

  const displayCount = count > maxCount ? `${maxCount}+` : count;

  return (
    <Badge
      bg={variant}
      className={`position-absolute top-0 start-100 translate-middle rounded-pill ${className}`}
      style={{ fontSize: '0.6rem' }}
    >
      {displayCount}
    </Badge>
  );
};

export default NotificationBadge;