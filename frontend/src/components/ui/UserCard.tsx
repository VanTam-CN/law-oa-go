import React from 'react';
import { Card, Badge } from 'react-bootstrap';

interface UserCardProps {
  name: string;
  email: string;
  role: string;
  status: string;
  avatar?: string;
  lastLogin?: string;
  createdAt?: string;
  className?: string;
  onClick?: () => void;
}

const UserCard: React.FC<UserCardProps> = ({ 
  name, 
  email, 
  role, 
  status, 
  avatar, 
  lastLogin, 
  createdAt,
  className = '',
  onClick
}) => {
  const getRoleBadgeClass = (role: string) => {
    switch (role) {
      case 'admin': return 'bg-danger';
      case 'lawyer': return 'bg-primary';
      case 'user': return 'bg-info';
      default: return 'bg-secondary';
    }
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'active': return 'bg-success';
      case 'inactive': return 'bg-secondary';
      default: return 'bg-secondary';
    }
  };

  // Get role display text
  const getRoleText = (role: string) => {
    switch (role) {
      case 'admin': return 'Administrator';
      case 'lawyer': return 'Lawyer';
      case 'user': return 'User';
      default: return role;
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
      className={`user-card h-100 shadow-sm border-0 ${className} ${onClick ? 'cursor-pointer' : ''}`}
      onClick={onClick}
    >
      <Card.Body className="d-flex flex-column">
        <div className="d-flex align-items-center mb-3">
          <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3" style={{ width: '60px', height: '60px' }}>
            {avatar ? (
              <img src={avatar} alt={name} className="rounded-circle" style={{ width: '60px', height: '60px' }} />
            ) : (
              <i className="fas fa-user fa-2x text-muted"></i>
            )}
          </div>
          <div className="flex-grow-1">
            <h5 className="mb-1">{name}</h5>
            <p className="text-muted mb-0 small">{email}</p>
          </div>
        </div>
        
        <div className="d-flex justify-content-between mb-3">
          <Badge bg={getRoleBadgeClass(role)} className="me-2">
            {getRoleText(role)}
          </Badge>
          <Badge bg={getStatusBadgeClass(status)}>
            {getStatusText(status)}
          </Badge>
        </div>
        
        <div className="mt-auto">
          {lastLogin && (
            <div className="d-flex align-items-center mb-1">
              <i className="fas fa-clock text-muted me-2"></i>
              <small className="text-muted">
                Last login: {new Date(lastLogin).toLocaleDateString()}
              </small>
            </div>
          )}
          {createdAt && (
            <div className="d-flex align-items-center">
              <i className="fas fa-calendar-plus text-muted me-2"></i>
              <small className="text-muted">
                Member since: {new Date(createdAt).toLocaleDateString()}
              </small>
            </div>
          )}
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

export default UserCard;