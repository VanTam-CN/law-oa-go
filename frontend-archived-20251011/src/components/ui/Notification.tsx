import React, { useState, useEffect } from 'react';
import { Toast, ToastContainer } from 'react-bootstrap';

interface NotificationProps {
  id: string;
  title: string;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
  duration?: number;
  onClose: (id: string) => void;
}

const Notification: React.FC<NotificationProps> = ({ 
  id,
  title,
  message,
  type,
  duration = 5000,
  onClose
}) => {
  const [show, setShow] = useState(true);

  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        setShow(false);
        setTimeout(() => onClose(id), 300);
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [duration, id, onClose]);

  const getIconClass = () => {
    switch (type) {
      case 'success': return 'fas fa-check-circle text-success';
      case 'error': return 'fas fa-exclamation-circle text-danger';
      case 'warning': return 'fas fa-exclamation-triangle text-warning';
      case 'info': return 'fas fa-info-circle text-info';
      default: return 'fas fa-info-circle text-info';
    }
  };

  const getBgClass = () => {
    switch (type) {
      case 'success': return 'bg-success';
      case 'error': return 'bg-danger';
      case 'warning': return 'bg-warning';
      case 'info': return 'bg-info';
      default: return 'bg-info';
    }
  };

  return (
    <Toast 
      show={show} 
      onClose={() => {
        setShow(false);
        setTimeout(() => onClose(id), 300);
      }}
      className={`mb-2 border-0 shadow-sm ${getBgClass()}`}
      style={{ maxWidth: '400px' }}
    >
      <Toast.Header className="d-flex justify-content-between align-items-center">
        <div className="d-flex align-items-center">
          <i className={`${getIconClass()} me-2`}></i>
          <strong className="me-auto">{title}</strong>
        </div>
        <small className="text-muted">Just now</small>
      </Toast.Header>
      <Toast.Body className="text-white">
        {message}
      </Toast.Body>
    </Toast>
  );
};

export default Notification;