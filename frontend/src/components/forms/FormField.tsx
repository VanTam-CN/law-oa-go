import React from 'react';
import { Form, InputGroup } from 'react-bootstrap';

interface FormFieldProps {
  label: string;
  name: string;
  type?: string;
  value: any;
  onChange: (e: React.ChangeEvent<any>) => void;
  required?: boolean;
  placeholder?: string;
  error?: string;
  helpText?: string;
  icon?: string;
  as?: any;
  rows?: number;
  options?: Array<{ value: string; label: string }>;
  className?: string;
}

const FormField: React.FC<FormFieldProps> = ({ 
  label,
  name,
  type = 'text',
  value,
  onChange,
  required = false,
  placeholder = '',
  error,
  helpText,
  icon,
  as,
  rows,
  options = [],
  className = ''
}) => {
  return (
    <Form.Group className={`mb-3 ${className}`}>
      <Form.Label>
        {label} {required && <span className="text-danger">*</span>}
      </Form.Label>
      
      {icon ? (
        <InputGroup>
          <InputGroup.Text>
            <i className={`${icon}`}></i>
          </InputGroup.Text>
          {type === 'select' ? (
            <Form.Select
              name={name}
              value={value}
              onChange={onChange}
              required={required}
              isInvalid={!!error}
            >
              {!required && (
                <option value="">
                  {placeholder || '请选择...'}
                </option>
              )}
              {options.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Form.Select>
          ) : as === 'textarea' ? (
            <Form.Control
              as="textarea"
              rows={rows || 3}
              name={name}
              value={value}
              onChange={onChange}
              required={required}
              placeholder={placeholder}
              isInvalid={!!error}
            />
          ) : (
            <Form.Control
              type={type}
              name={name}
              value={value}
              onChange={onChange}
              required={required}
              placeholder={placeholder}
              isInvalid={!!error}
            />
          )}
        </InputGroup>
      ) : type === 'select' ? (
        <Form.Select
          name={name}
          value={value}
          onChange={onChange}
          required={required}
          isInvalid={!!error}
        >
          {!required && (
            <option value="">
              {placeholder || '请选择...'}
            </option>
          )}
          {options.map(option => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </Form.Select>
      ) : as === 'textarea' ? (
        <Form.Control
          as="textarea"
          rows={rows || 3}
          name={name}
          value={value}
          onChange={onChange}
          required={required}
          placeholder={placeholder}
          isInvalid={!!error}
        />
      ) : (
        <Form.Control
          type={type}
          name={name}
          value={value}
          onChange={onChange}
          required={required}
          placeholder={placeholder}
          isInvalid={!!error}
        />
      )}
      
      {helpText && <Form.Text className="text-muted">{helpText}</Form.Text>}
      {error && <Form.Control.Feedback type="invalid">{error}</Form.Control.Feedback>}
    </Form.Group>
  );
};

export default FormField;