import React from 'react';
import { Badge } from 'react-bootstrap';

interface UserAvatarProps {
  name: string;
  size?: 'sm' | 'md' | 'lg';
  status?: 'online' | 'offline' | 'away' | 'busy';
  className?: string;
}

const UserAvatar: React.FC<UserAvatarProps> = ({ 
  name, 
  size = 'md',
  status,
  className = ''
}) => {
  const getSizeClasses = () => {
    switch (size) {
      case 'sm': return 'avatar-sm';
      case 'md': return 'avatar-md';
      case 'lg': return 'avatar-lg';
      default: return 'avatar-md';
    }
  };

  const getStatusClasses = () => {
    if (!status) return '';
    
    switch (status) {
      case 'online': return 'bg-success';
      case 'offline': return 'bg-secondary';
      case 'away': return 'bg-warning';
      case 'busy': return 'bg-danger';
      default: return 'bg-secondary';
    }
  };

  const getInitials = (name: string) => {
    const names = name.split(' ');
    let initials = names[0].substring(0, 1).toUpperCase();
    
    if (names.length > 1) {
      initials += names[names.length - 1].substring(0, 1).toUpperCase();
    }
    
    return initials;
  };

  return (
    <div className={`user-avatar ${getSizeClasses()} ${className} position-relative d-inline-block`}>
      <div className="avatar bg-light rounded-circle d-flex align-items-center justify-content-center">
        <span className="text-muted fw-bold">{getInitials(name)}</span>
      </div>
      {status && (
        <Badge
          bg={getStatusClasses()}
          className="position-absolute bottom-0 end-0 rounded-circle border border-white"
          style={{ width: '10px', height: '10px', minWidth: '10px' }}
        ></Badge>
      )}
    </div>
  );
};

export default UserAvatar;