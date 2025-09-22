import React from 'react';
import { Card, Badge } from 'react-bootstrap';

interface CaseCardProps {
  id: number;
  title: string;
  description: string;
  clientName: string;
  caseType: string;
  priority: string;
  status: string;
  lawyerName?: string;
  createdAt: string;
  updatedAt: string;
  className?: string;
  onClick?: () => void;
}

const CaseCard: React.FC<CaseCardProps> = ({ 
  id,
  title, 
  description, 
  clientName, 
  caseType, 
  priority, 
  status, 
  lawyerName,
  createdAt,
  updatedAt,
  className = '',
  onClick
}) => {
  const getCaseTypeBadgeClass = (caseType: string) => {
    switch (caseType) {
      case 'civil': return 'bg-primary';
      case 'criminal': return 'bg-danger';
      case 'commercial': return 'bg-success';
      case 'administrative': return 'bg-info';
      default: return 'bg-secondary';
    }
  };

  const getPriorityBadgeClass = (priority: string) => {
    switch (priority) {
      case 'low': return 'bg-info';
      case 'medium': return 'bg-warning';
      case 'high': return 'bg-danger';
      case 'urgent': return 'bg-danger';
      default: return 'bg-secondary';
    }
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-warning';
      case 'active': return 'bg-primary';
      case 'closed': return 'bg-success';
      case 'suspended': return 'bg-secondary';
      default: return 'bg-secondary';
    }
  };

  // Get case type display text
  const getCaseTypeText = (caseType: string) => {
    switch (caseType) {
      case 'civil': return 'Civil';
      case 'criminal': return 'Criminal';
      case 'commercial': return 'Commercial';
      case 'administrative': return 'Administrative';
      default: return caseType;
    }
  };

  // Get priority display text
  const getPriorityText = (priority: string) => {
    switch (priority) {
      case 'low': return 'Low';
      case 'medium': return 'Medium';
      case 'high': return 'High';
      case 'urgent': return 'Urgent';
      default: return priority;
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return 'Pending';
      case 'active': return 'Active';
      case 'closed': return 'Closed';
      case 'suspended': return 'Suspended';
      default: return status;
    }
  };

  return (
    <Card 
      className={`case-card h-100 shadow-sm border-0 ${className} ${onClick ? 'cursor-pointer' : ''}`}
      onClick={onClick}
    >
      <Card.Body className="d-flex flex-column">
        <div className="d-flex justify-content-between align-items-start mb-3">
          <div>
            <h5 className="mb-1">#{id} {title}</h5>
            <p className="text-muted mb-0 small">{description.substring(0, 100)}{description.length > 100 ? '...' : ''}</p>
          </div>
          <div className="d-flex flex-column">
            <Badge bg={getCaseTypeBadgeClass(caseType)} className="mb-1">
              {getCaseTypeText(caseType)}
            </Badge>
            <Badge bg={getPriorityBadgeClass(priority)}>
              {getPriorityText(priority)}
            </Badge>
          </div>
        </div>
        
        <div className="d-flex justify-content-between mb-3">
          <div className="d-flex align-items-center">
            <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2" style={{ width: '32px', height: '32px' }}>
              <i className="fas fa-user text-muted"></i>
            </div>
            <div>
              <div className="small fw-bold">{clientName}</div>
              <div className="small text-muted">Client</div>
            </div>
          </div>
          
          {lawyerName && (
            <div className="d-flex align-items-center">
              <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2" style={{ width: '32px', height: '32px' }}>
                <i className="fas fa-user-tie text-muted"></i>
              </div>
              <div>
                <div className="small fw-bold">{lawyerName}</div>
                <div className="small text-muted">Lawyer</div>
              </div>
            </div>
          )}
        </div>
        
        <div className="mt-auto">
          <div className="d-flex justify-content-between align-items-center mb-2">
            <Badge bg={getStatusBadgeClass(status)}>
              {getStatusText(status)}
            </Badge>
            <small className="text-muted">
              Updated: {new Date(updatedAt).toLocaleDateString()}
            </small>
          </div>
          
          <div className="d-flex justify-content-between">
            <small className="text-muted">
              <i className="fas fa-calendar-plus me-1"></i>
              {new Date(createdAt).toLocaleDateString()}
            </small>
            <small className="text-muted">
              <i className="fas fa-history me-1"></i>
              {Math.floor((new Date().getTime() - new Date(updatedAt).getTime()) / (1000 * 60 * 60 * 24))} days ago
            </small>
          </div>
        </div>
      </Card.Body>
      
      <Card.Footer className="bg-transparent border-top-0 d-flex justify-content-end">
        <div className="d-flex">
          <button className="btn btn-sm btn-outline-primary me-2">
            <i className="fas fa-edit me-1"></i>
            Edit
          </button>
          <button className="btn btn-sm btn-outline-info me-2">
            <i className="fas fa-eye me-1"></i>
            View
          </button>
          <button className="btn btn-sm btn-outline-danger">
            <i className="fas fa-trash me-1"></i>
            Delete
          </button>
        </div>
      </Card.Footer>
    </Card>
  );
};

export default CaseCard;