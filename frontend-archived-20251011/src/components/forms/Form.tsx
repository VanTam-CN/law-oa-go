import React from 'react';
import { Form as BootstrapForm, Button, Spinner } from 'react-bootstrap';

interface FormProps {
  children: React.ReactNode;
  onSubmit: (e: React.FormEvent) => void;
  loading?: boolean;
  submitText?: string;
  cancelText?: string;
  onCancel?: () => void;
  submitVariant?: string;
  cancelVariant?: string;
  className?: string;
}

const Form: React.FC<FormProps> = ({ 
  children,
  onSubmit,
  loading = false,
  submitText = 'Submit',
  cancelText = 'Cancel',
  onCancel,
  submitVariant = 'primary',
  cancelVariant = 'secondary',
  className = ''
}) => {
  return (
    <BootstrapForm onSubmit={onSubmit} className={className}>
      {children}
      
      <div className="d-flex justify-content-end mt-4">
        {onCancel && (
          <Button 
            variant={cancelVariant} 
            onClick={onCancel}
            className="me-2"
            disabled={loading}
          >
            <i className="fas fa-times me-2"></i>
            {cancelText}
          </Button>
        )}
        <Button 
          variant={submitVariant} 
          type="submit"
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
              <i className="fas fa-save me-2"></i>
              {submitText}
            </>
          )}
        </Button>
      </div>
    </BootstrapForm>
  );
};

export default Form;