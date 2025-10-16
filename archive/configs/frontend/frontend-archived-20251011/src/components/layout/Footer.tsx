import React from 'react';
import { Container, Row, Col } from 'react-bootstrap';

const Footer: React.FC = () => {
  return (
    <footer className="footer bg-dark text-white mt-auto py-4">
      <Container fluid>
        <Row>
          <Col md={6}>
            <div className="d-flex align-items-center">
              <div className="bg-primary rounded-circle d-flex align-items-center justify-content-center me-3" style={{ width: '40px', height: '40px' }}>
                <i className="fas fa-balance-scale text-white"></i>
              </div>
              <div>
                <h5 className="mb-0">Law OA System</h5>
                <p className="mb-0 text-muted small">Modern Legal Practice Management</p>
              </div>
            </div>
          </Col>
          <Col md={6}>
            <div className="d-flex justify-content-end align-items-center">
              <div className="me-4">
                <small className="text-muted">
                  <i className="fas fa-code me-1"></i>
                  Version 2.1.0
                </small>
              </div>
              <div className="me-4">
                <small className="text-muted">
                  <i className="fas fa-shield-alt me-1"></i>
                  Secure
                </small>
              </div>
              <div>
                <small className="text-muted">
                  <i className="fas fa-copyright me-1"></i>
                  {new Date().getFullYear()} Law OA System
                </small>
              </div>
            </div>
          </Col>
        </Row>
        <Row className="mt-3">
          <Col md={12}>
            <div className="d-flex justify-content-center">
              <div className="me-4">
                <a href="/privacy" className="text-decoration-none text-muted">
                  <i className="fas fa-lock me-1"></i>
                  Privacy Policy
                </a>
              </div>
              <div className="me-4">
                <a href="/terms" className="text-decoration-none text-muted">
                  <i className="fas fa-file-contract me-1"></i>
                  Terms of Service
                </a>
              </div>
              <div className="me-4">
                <a href="/help" className="text-decoration-none text-muted">
                  <i className="fas fa-question-circle me-1"></i>
                  Help Center
                </a>
              </div>
              <div>
                <a href="/contact" className="text-decoration-none text-muted">
                  <i className="fas fa-envelope me-1"></i>
                  Contact Us
                </a>
              </div>
            </div>
          </Col>
        </Row>
      </Container>
    </footer>
  );
};

export default Footer;