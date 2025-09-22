import React from 'react';
import { Button } from 'react-bootstrap';

interface SidebarToggleProps {
  collapsed: boolean;
  onToggle: () => void;
  className?: string;
}

const SidebarToggle: React.FC<SidebarToggleProps> = ({ 
  collapsed, 
  onToggle,
  className = ''
}) => {
  return (
    <Button
      variant="outline-light"
      className={`sidebar-toggle position-fixed ${className}`}
      onClick={onToggle}
      style={{
        top: '10px',
        left: collapsed ? '70px' : '260px',
        zIndex: 1030,
        transition: 'left 0.3s ease'
      }}
      aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
    >
      <i className={`fas ${collapsed ? 'fa-angle-double-right' : 'fa-angle-double-left'}`}></i>
    </Button>
  );
};

export default SidebarToggle;