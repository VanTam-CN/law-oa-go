import React from 'react';
import { Card as BootstrapCard } from 'react-bootstrap';

interface CardProps {
  title?: string;
  subtitle?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  headerClassName?: string;
  bodyClassName?: string;
  footer?: React.ReactNode;
  footerClassName?: string;
  collapsible?: boolean;
  defaultCollapsed?: boolean;
  onCollapse?: (collapsed: boolean) => void;
}

const Card: React.FC<CardProps> = ({ 
  title,
  subtitle,
  actions,
  children,
  className = '',
  headerClassName = '',
  bodyClassName = '',
  footer,
  footerClassName = '',
  collapsible = false,
  defaultCollapsed = false,
  onCollapse
}) => {
  const [collapsed, setCollapsed] = React.useState(defaultCollapsed);

  const toggleCollapse = () => {
    const newCollapsed = !collapsed;
    setCollapsed(newCollapsed);
    if (onCollapse) {
      onCollapse(newCollapsed);
    }
  };

  return (
    <BootstrapCard className={`shadow-sm border-0 ${className}`}>
      {(title || subtitle || actions || collapsible) && (
        <BootstrapCard.Header className={`bg-white border-0 ${headerClassName}`}>
          <div className="d-flex justify-content-between align-items-center">
            <div>
              {title && <BootstrapCard.Title className="mb-0">{title}</BootstrapCard.Title>}
              {subtitle && <BootstrapCard.Subtitle className="mb-0 text-muted">{subtitle}</BootstrapCard.Subtitle>}
            </div>
            <div className="d-flex align-items-center">
              {actions && <div className="me-2">{actions}</div>}
              {collapsible && (
                <button 
                  className="btn btn-sm btn-outline-secondary rounded-circle p-1" 
                  onClick={toggleCollapse}
                  aria-label={collapsed ? 'Expand' : 'Collapse'}
                >
                  <i className={`fas ${collapsed ? 'fa-chevron-down' : 'fa-chevron-up'}`}></i>
                </button>
              )}
            </div>
          </div>
        </BootstrapCard.Header>
      )}
      
      {!collapsed && (
        <BootstrapCard.Body className={bodyClassName}>
          {children}
        </BootstrapCard.Body>
      )}
      
      {footer && !collapsed && (
        <BootstrapCard.Footer className={footerClassName}>
          {footer}
        </BootstrapCard.Footer>
      )}
    </BootstrapCard>
  );
};

export default Card;