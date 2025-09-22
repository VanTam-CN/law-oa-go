import React from 'react';
import { ProgressBar as BootstrapProgressBar } from 'react-bootstrap';

interface ProgressBarProps {
  value: number;
  max?: number;
  label?: string;
  variant?: 'success' | 'info' | 'warning' | 'danger' | 'primary' | 'secondary';
  striped?: boolean;
  animated?: boolean;
  className?: string;
  showPercentage?: boolean;
}

const ProgressBar: React.FC<ProgressBarProps> = ({ 
  value, 
  max = 100, 
  label, 
  variant = 'primary',
  striped = false,
  animated = false,
  className = '',
  showPercentage = true
}) => {
  const percentage = Math.round((value / max) * 100);

  return (
    <div className={className}>
      {label && (
        <div className="d-flex justify-content-between mb-1">
          <span className="small fw-bold">{label}</span>
          {showPercentage && <span className="small">{percentage}%</span>}
        </div>
      )}
      <BootstrapProgressBar 
        variant={variant}
        now={percentage}
        striped={striped}
        animated={animated}
        className="mb-2"
      />
      {!label && showPercentage && (
        <div className="d-flex justify-content-end">
          <span className="small">{percentage}%</span>
        </div>
      )}
    </div>
  );
};

export default ProgressBar;