import React from 'react';
import { ToastContainer } from 'react-bootstrap';
import Notification from './Notification';

interface NotificationContainerProps {
  notifications: Array<{
    id: string;
    title: string;
    message: string;
    type: 'success' | 'error' | 'warning' | 'info';
    duration?: number;
  }>;
  onRemove: (id: string) => void;
}

const NotificationContainer: React.FC<NotificationContainerProps> = ({ 
  notifications, 
  onRemove 
}) => {
  return (
    <ToastContainer 
      position="top-end" 
      className="p-3"
      style={{ zIndex: 9999 }}
    >
      {notifications.map(notification => (
        <Notification
          key={notification.id}
          id={notification.id}
          title={notification.title}
          message={notification.message}
          type={notification.type}
          duration={notification.duration}
          onClose={onRemove}
        />
      ))}
    </ToastContainer>
  );
};

export default NotificationContainer;