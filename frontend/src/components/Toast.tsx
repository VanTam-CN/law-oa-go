import React, { useState, useEffect } from 'react';
import { Toast, ToastContainer, Button } from 'react-bootstrap';

export interface ToastMessage {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title?: string;
  message: string;
  duration?: number;
  persistent?: boolean;
}

interface ToastContextType {
  showToast: (toast: Omit<ToastMessage, 'id'>) => void;
  hideToast: (id: string) => void;
  clearAllToasts: () => void;
}

const ToastContext = React.createContext<ToastContextType | null>(null);

export const useToast = (): ToastContextType => {
  const context = React.useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};

interface ToastProviderProps {
  children: React.ReactNode;
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const showToast = (toast: Omit<ToastMessage, 'id'>) => {
    const id = Date.now().toString();
    const newToast: ToastMessage = {
      ...toast,
      id,
      duration: toast.duration || (toast.type === 'error' ? 0 : 5000),
      persistent: toast.persistent || toast.type === 'error'
    };

    setToasts(prev => [...prev, newToast]);

    // 非持久化toast自动消失
    if (!newToast.persistent && newToast.duration && newToast.duration > 0) {
      setTimeout(() => {
        hideToast(id);
      }, newToast.duration);
    }
  };

  const hideToast = (id: string) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  };

  const clearAllToasts = () => {
    setToasts([]);
  };

  return (
    <ToastContext.Provider value={{ showToast, hideToast, clearAllToasts }}>
      {children}

      <ToastContainer position="top-end" className="p-3" style={{ zIndex: 9999 }}>
        {toasts.map(toast => (
          <Toast
            key={toast.id}
            onClose={() => hideToast(toast.id)}
            show={true}
            bg={toast.type}
            delay={0}
            autohide={toast.persistent ? false : true}
            className={`toast-${toast.type}`}
          >
            <Toast.Header>
              <strong className="me-auto">
                {toast.title || (toast.type === 'success' ? '成功' :
                           toast.type === 'error' ? '错误' :
                           toast.type === 'warning' ? '警告' : '信息')}
              </strong>
              {!toast.persistent && (
                <Button variant="close" onClick={() => hideToast(toast.id)} size="sm">
                  <span>&times;</span>
                </Button>
              )}
            </Toast.Header>
            <Toast.Body>
              {toast.message}
            </Toast.Body>
          </Toast>
        ))}
      </ToastContainer>
    </ToastContext.Provider>
  );
};

// 便捷的toast方法
export const showSuccessToast = (message: string, title?: string, options?: Partial<ToastMessage>) => {
  const toast = useToast();
  return toast.showToast({
    type: 'success',
    message,
    title,
    ...options
  });
};

export const showErrorToast = (message: string, title?: string, options?: Partial<ToastMessage>) => {
  const toast = useToast();
  return toast.showToast({
    type: 'error',
    message,
    title,
    persistent: true,
    ...options
  });
};

export const showWarningToast = (message: string, title?: string, options?: Partial<ToastMessage>) => {
  const toast = useToast();
  return toast.showToast({
    type: 'warning',
    message,
    title,
    ...options
  });
};

export const showInfoToast = (message: string, title?: string, options?: Partial<ToastMessage>) => {
  const toast = useToast();
  return toast.showToast({
    type: 'info',
    message,
    title,
    ...options
  });
};

export default ToastProvider;