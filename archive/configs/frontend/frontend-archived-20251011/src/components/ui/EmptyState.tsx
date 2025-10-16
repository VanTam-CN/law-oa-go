import React from 'react';
import { Button, Card } from 'react-bootstrap';

interface EmptyStateProps {
  icon: string;
  title: string;
  description: string;
  actionText?: string;
  onAction?: () => void;
  actionIcon?: string;
  secondaryActionText?: string;
  onSecondaryAction?: () => void;
  secondaryActionIcon?: string;
  className?: string;
}

const EmptyState: React.FC<EmptyStateProps> = ({ 
  icon,
  title,
  description,
  actionText,
  onAction,
  actionIcon,
  secondaryActionText,
  onSecondaryAction,
  secondaryActionIcon,
  className = ''
}) => {
  return (
    <Card className={`text-center border-0 ${className}`}>
      <Card.Body className="py-5">
        <div className="mb-4">
          <div className="bg-light rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3" style={{ width: '80px', height: '80px' }}>
            <i className={`${icon} fa-2x text-muted`}></i>
          </div>
          <h2 className="mb-3">{title}</h2>
          <p className="text-muted mb-0">{description}</p>
        </div>
        
        {(actionText || secondaryActionText) && (
          <div className="d-flex justify-content-center gap-2">
            {actionText && onAction && (
              <Button 
                variant="primary" 
                onClick={onAction}
                className="d-flex align-items-center"
              >
                {actionIcon && <i className={`${actionIcon} me-2`}></i>}
                {actionText}
              </Button>
            )}
            {secondaryActionText && onSecondaryAction && (
              <Button 
                variant="outline-primary" 
                onClick={onSecondaryAction}
                className="d-flex align-items-center"
              >
                {secondaryActionIcon && <i className={`${secondaryActionIcon} me-2`}></i>}
                {secondaryActionText}
              </Button>
            )}
          </div>
        )}
      </Card.Body>
    </Card>
  );
};

export default EmptyState;