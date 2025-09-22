import React from 'react';
import { Card, Badge } from 'react-bootstrap';

interface ClientCardProps {
  id: number;
  name: string;
  email: string;
  phone: string;
  company: string;
  address: string;
  status: string;
  totalCases: number;
  activeCases: number;
  createdAt: string;
  updatedAt: string;
  className?: string;
  onClick?: () => void;
}

const ClientCard: React.FC<ClientCardProps> = ({ 
  id,
  name, 
  email, 
  phone, 
  company, 
  address, 
  status, 
  totalCases, 
  activeCases,
  createdAt,
  updatedAt,
  className = '',
  onClick
}) => {
  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'active': return 'bg-success';
      case 'inactive': return 'bg-secondary';
      default: return 'bg-secondary';
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return 'Active';
      case 'inactive': return 'Inactive';
      default: return status;
    }
  };

  return (
    <Card 
      className={`client-card h-100 shadow-sm border-0 ${className} ${onClick ? 'cursor-pointer' : ''}`}
      onClick={onClick}
    >
      <Card.Body className="d-flex flex-column">
        <div className="d-flex justify-content-between align-items-start mb-3">
          <div className="d-flex align-items-center">
            <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3" style={{ width: '60px', height: '60px' }}>
              <i className="fas fa-user fa-2x text-muted"></i>
            </div>
            <div>
              <h5 className="mb-1">{name}</h5>
              <div className="d-flex align-items-center">
                <Badge bg={getStatusBadgeClass(status)} className="me-2">
                  {getStatusText(status)}
                </Badge>
                <small className="text-muted">ID: #{id}</small>
              </div>
            </div>
          </div>
        </div>
        
        <div className="mb-3">
          <div className="d-flex align-items-center mb-2">
            <i className="fas fa-envelope text-muted me-2"></i>
            <small className="text-muted">{email}</small>
          </div>
          <div className="d-flex align-items-center mb-2">
            <i className="fas fa-phone text-muted me-2"></i>
            <small className="text-muted">{phone}</small>
          </div>
          <div className="d-flex align-items-center mb-2">
            <i className="fas fa-building text-muted me-2"></i>
            <small className="text-muted">{company}</small>
          </div>
          <div className="d-flex align-items-start">
            <i className="fas fa-map-marker-alt text-muted mt-1 me-2"></i>
            <small className="text-muted">{address}</small>
          </div>
        </div>
        
        <div className="mt-auto">
          <div className="d-flex justify-content-between mb-2">
            <div className="text-center">
              <div className="fw-bold">{totalCases}</div>
              <div className="small text-muted">Total Cases</div>
            </div>
            <div className="text-center">
              <div className="fw-bold text-primary">{activeCases}</div>
              <div className="small text-muted">Active Cases</div>
            </div>
            <div className="text-center">
              <div className="fw-bold text-success">{totalCases - activeCases}</div>
              <div className="small text-muted">Closed Cases</div>
            </div>
          </div>
          
          <div className="d-flex justify-content-between">
            <small className="text-muted">
              <i className="fas fa-calendar-plus me-1"></i>
              {new Date(createdAt).toLocaleDateString()}
            </small>
            <small className="text-muted">
              <i className="fas fa-history me-1"></i>
              {new Date(updatedAt).toLocaleDateString()}
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

export default ClientCard;