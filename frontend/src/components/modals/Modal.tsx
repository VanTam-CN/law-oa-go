import React from 'react';
import { Modal as BootstrapModal, Button, Spinner } from 'react-bootstrap';

interface ModalProps {
  show: boolean;
  onHide: () => void;
  title: string;
  children: React.ReactNode;
  onConfirm?: (e: React.FormEvent) => void;
  confirmText?: string;
  cancelText?: string;
  confirmVariant?: string;
  cancelVariant?: string;
  loading?: boolean;
  size?: 'sm' | 'lg' | 'xl';
  icon?: string;
  footer?: React.ReactNode;
}

const Modal: React.FC<ModalProps> = ({ 
  show, 
  onHide, 
  title,
  children,
  onConfirm,
  confirmText = 'Save',
  cancelText = 'Cancel',
  confirmVariant = 'primary',
  cancelVariant = 'secondary',
  loading = false,
  size = 'lg',
  icon,
  footer
}) => {
  return (
    <BootstrapModal show={show} onHide={onHide} size={size} centered>
      <BootstrapModal.Header closeButton>
        <BootstrapModal.Title>
          {icon && <i className={`${icon} me-2`}></i>}
          {title}
        </BootstrapModal.Title>
      </BootstrapModal.Header>
      <BootstrapModal.Body>
        {children}
      </BootstrapModal.Body>
      <BootstrapModal.Footer>
        {footer ? (
          footer
        ) : onConfirm ? (
          <>
            <Button variant={cancelVariant} onClick={onHide} disabled={loading}>
              <i className="fas fa-times me-2"></i>
              {cancelText}
            </Button>
            <Button 
              variant={confirmVariant} 
              onClick={onConfirm} 
              disabled={loading}
            >
              {loading ? (
                <>
                  <Spinner 
                    as="span" 
                    animation="border" 
                    size="sm" 
                    role="status" 
                    aria-hidden="true" 
                    className="me-2"
                  />
                  Processing...
                </>
              ) : (
                <>
                  {icon && <i className={`${icon} me-2`}></i>}
                  {confirmText}
                </>
              )}
            </Button>
          </>
        ) : (
          <Button variant={cancelVariant} onClick={onHide}>
            <i className="fas fa-times me-2"></i>
            {cancelText}
          </Button>
        )}
      </BootstrapModal.Footer>
    </BootstrapModal>
  );
};

export default Modal;