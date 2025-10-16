import React from 'react';
import { Modal, Button } from 'react-bootstrap';

interface ConfirmModalProps {
  show: boolean;
  onHide: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  confirmVariant?: string;
  cancelVariant?: string;
  icon?: string;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({ 
  show, 
  onHide, 
  onConfirm,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  confirmVariant = 'primary',
  cancelVariant = 'secondary',
  icon = 'fa-exclamation-triangle'
}) => {
  return (
    <Modal show={show} onHide={onHide} centered>
      <Modal.Header closeButton>
        <Modal.Title>
          <i className={`fas ${icon} me-2`}></i>
          {title}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <p>{message}</p>
      </Modal.Body>
      <Modal.Footer>
        <Button variant={cancelVariant} onClick={onHide}>
          <i className="fas fa-times me-2"></i>
          {cancelText}
        </Button>
        <Button variant={confirmVariant} onClick={onConfirm}>
          <i className="fas fa-check me-2"></i>
          {confirmText}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default ConfirmModal;