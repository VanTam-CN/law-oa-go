import React from 'react';
import { Link } from 'react-router-dom';

const NotFoundPage: React.FC = () => {
  return (
    <div className="d-flex align-items-center justify-content-center min-vh-100 bg-light">
      <Container>
        <Row className="justify-content-md-center">
          <Col md={8} lg={6}>
            <Card className="text-center shadow-lg border-0">
              <Card.Body className="p-5">
                <div className="mb-4">
                  <div className="bg-primary rounded-circle d-flex align-items-center justify-content-center mx-auto mb-4" style={{ width: '100px', height: '100px' }}>
                    <i className="fas fa-exclamation-triangle fa-3x text-white"></i>
                  </div>
                  <h1 className="display-1 fw-bold text-primary">404</h1>
                  <h2 className="mb-3">Page Not Found</h2>
                  <p className="text-muted mb-4">
                    Sorry, the page you are looking for could not be found. It might have been removed, renamed, or is temporarily unavailable.
                  </p>
                </div>
                
                <div className="d-grid gap-3 d-md-flex justify-content-md-center">
                  <Link to="/">
                    <Button variant="primary" size="lg" className="me-md-2">
                      <i className="fas fa-home me-2"></i>
                      Go to Homepage
                    </Button>
                  </Link>
                  <Link to="/dashboard">
                    <Button variant="outline-primary" size="lg">
                      <i className="fas fa-tachometer-alt me-2"></i>
                      Go to Dashboard
                    </Button>
                  </Link>
                </div>
                
                <div className="mt-5">
                  <h5 className="mb-3">Need Help?</h5>
                  <div className="d-flex justify-content-center gap-3">
                    <Link to="/help" className="text-decoration-none">
                      <i className="fas fa-question-circle me-2"></i>
                      Visit Help Center
                    </Link>
                    <Link to="/support" className="text-decoration-none">
                      <i className="fas fa-headset me-2"></i>
                      Contact Support
                    </Link>
                  </div>
                </div>
              </Card.Body>
            </Card>
            
            <div className="text-center mt-4">
              <p className="text-muted small">
                <i className="fas fa-shield-alt me-1"></i>
                Law Office Automation System
              </p>
            </div>
          </Col>
        </Row>
      </Container>
    </div>
  );
};

export default NotFoundPage;