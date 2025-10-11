import React from 'react';
import { Spinner, OverlayTrigger, Tooltip } from 'react-bootstrap';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'info' | 'light' | 'dark';
  message?: string;
  fullscreen?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ 
  size = 'md',
  variant = 'primary',
  message = 'Loading...',
  fullscreen = false,
  className = '',
  style = {}
}) => {
  const getSizeClass = () => {
    switch (size) {
      case 'sm': return 'spinner-border-sm';
      case 'lg': return 'spinner-border-lg';
      default: return '';
    }
  };

  const getWrapperStyle = (): React.CSSProperties => {
    if (fullscreen) {
      return {
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        backgroundColor: 'rgba(255, 255, 255, 0.8)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 9999,
        ...style
      };
    }
    return style;
  };

  return (
    <div 
      className={className}
      style={getWrapperStyle()}
    >
      <div className="text-center">
        <Spinner 
          animation="border" 
          variant={variant}
          className={getSizeClass()}
        />
        {message && (
          <div className="mt-2">
            <OverlayTrigger
              placement="bottom"
              overlay={<Tooltip id="loading-tooltip">{message}</Tooltip>}
            >
              <span className="text-muted">{message}</span>
            </OverlayTrigger>
          </div>
        )}
      </div>
    </div>
  );
};

export default LoadingSpinner;