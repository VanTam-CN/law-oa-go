import React from 'react';
import { Card, Button, Dropdown } from 'react-bootstrap';

interface ChartCardProps {
  title: string;
  children: React.ReactNode;
  onRefresh?: () => void;
  onViewDetails?: () => void;
  className?: string;
}

const ChartCard: React.FC<ChartCardProps> = ({ 
  title, 
  children, 
  onRefresh, 
  onViewDetails,
  className = ''
}) => {
  return (
    <Card className={`chart-card shadow-sm border-0 ${className}`}>
      <Card.Header className="d-flex justify-content-between align-items-center">
        <span className="fw-bold">{title}</span>
        <div className="d-flex">
          {onViewDetails && (
            <Button 
              variant="outline-primary" 
              size="sm" 
              className="me-2"
              onClick={onViewDetails}
            >
              <i className="fas fa-external-link-alt me-1"></i>
              View Details
            </Button>
          )}
          <Dropdown>
            <Dropdown.Toggle variant="outline-secondary" size="sm" id="chart-actions-dropdown">
              <i className="fas fa-ellipsis-h"></i>
            </Dropdown.Toggle>
            <Dropdown.Menu>
              {onRefresh && (
                <Dropdown.Item onClick={onRefresh}>
                  <i className="fas fa-sync-alt me-2"></i>
                  Refresh Data
                </Dropdown.Item>
              )}
              <Dropdown.Item>
                <i className="fas fa-download me-2"></i>
                Export Chart
              </Dropdown.Item>
              <Dropdown.Item>
                <i className="fas fa-cog me-2"></i>
                Chart Settings
              </Dropdown.Item>
              <Dropdown.Divider />
              <Dropdown.Item>
                <i className="fas fa-expand me-2"></i>
                Fullscreen View
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
        </div>
      </Card.Header>
      <Card.Body>
        {children}
      </Card.Body>
    </Card>
  );
};

export default ChartCard;