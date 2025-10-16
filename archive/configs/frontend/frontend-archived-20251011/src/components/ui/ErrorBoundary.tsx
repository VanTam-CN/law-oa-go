import React from 'react';
import { Button, Card } from 'react-bootstrap';

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { 
      hasError: false, 
      error: null, 
      errorInfo: null 
    };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({ error, errorInfo });
    console.error('Error caught by boundary:', error, errorInfo);
  }

  handleReload = () => {
    window.location.reload();
  };

  handleGoHome = () => {
    window.location.href = '/';
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="d-flex align-items-center justify-content-center min-vh-100 bg-light">
          <Card className="text-center shadow-lg border-0" style={{ maxWidth: '600px' }}>
            <Card.Body className="p-5">
              <div className="mb-4">
                <div className="bg-danger rounded-circle d-flex align-items-center justify-content-center mx-auto mb-4" style={{ width: '100px', height: '100px' }}>
                  <i className="fas fa-exclamation-triangle fa-3x text-white"></i>
                </div>
                <h1 className="display-4 fw-bold text-danger">Oops!</h1>
                <h2 className="mb-3">Something went wrong</h2>
                <p className="text-muted mb-4">
                  We're sorry, but something unexpected happened. Our team has been notified and we're working to fix the issue.
                </p>
              </div>
              
              <div className="d-grid gap-3 d-md-flex justify-content-md-center">
                <Button variant="primary" onClick={this.handleReload} className="me-md-2">
                  <i className="fas fa-sync-alt me-2"></i>
                  Reload Page
                </Button>
                <Button variant="outline-primary" onClick={this.handleGoHome}>
                  <i className="fas fa-home me-2"></i>
                  Go Home
                </Button>
              </div>
              
              {this.state.error && (
                <details className="mt-4 text-start">
                  <summary className="text-muted small mb-2">Error details</summary>
                  <div className="bg-light p-3 rounded">
                    <h6 className="text-danger">{this.state.error.toString()}</h6>
                    {this.state.errorInfo && (
                      <pre className="text-muted small">
                        {this.state.errorInfo.componentStack}
                      </pre>
                    )}
                  </div>
                </details>
              )}
            </Card.Body>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;