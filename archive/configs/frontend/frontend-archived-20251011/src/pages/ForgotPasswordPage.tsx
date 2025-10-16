import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { forgotPassword } from '../services/authService';
import './AuthPage.css';

const ForgotPasswordPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess(false);

    try {
      await forgotPassword(email);
      setSuccess(true);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to send reset password email. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page login-page">
      <Container>
        <Row className="justify-content-md-center">
          <Col md={8} lg={6}>
            <div className="text-center mb-5">
              <div className="brand-logo bg-primary rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3">
                <i className="fas fa-balance-scale fa-2x text-white"></i>
              </div>
              <h1 className="display-4">Law OA System</h1>
              <p className="text-muted">Modern Legal Practice Management</p>
            </div>

            <Card className="auth-card">
              <Card.Body className="p-5">
                <div className="text-center mb-4">
                  <h2>Reset Password</h2>
                  <p className="text-muted">Enter your email to receive password reset instructions</p>
                </div>

                {error && <Alert variant="danger">{error}</Alert>}
                {success && (
                  <Alert variant="success">
                    Password reset instructions have been sent to your email. Please check your inbox.
                  </Alert>
                )}

                <Form onSubmit={handleSubmit}>
                  <Form.Group className="mb-4">
                    <Form.Label>Email Address <span className="text-danger">*</span></Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <i className="fas fa-envelope"></i>
                      </InputGroup.Text>
                      <Form.Control
                        type="email"
                        placeholder="Enter your email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        required
                        autoFocus
                        disabled={success}
                      />
                    </InputGroup>
                    <Form.Text className="text-muted">
                      We'll send password reset instructions to this email address.
                    </Form.Text>
                  </Form.Group>

                  <Button
                    variant="primary"
                    type="submit"
                    className="w-100 mb-3 auth-button"
                    disabled={loading || success}
                  >
                    {loading ? (
                      <>
                        <i className="fas fa-spinner fa-spin me-2"></i>
                        Sending Instructions...
                      </>
                    ) : success ? (
                      <>
                        <i className="fas fa-check me-2"></i>
                        Instructions Sent!
                      </>
                    ) : (
                      <>
                        <i className="fas fa-paper-plane me-2"></i>
                        Send Reset Instructions
                      </>
                    )}
                  </Button>

                  <div className="text-center">
                    <p className="mb-0">
                      Remember your password?{' '}
                      <Link to="/login" className="text-decoration-none">
                        Sign in
                      </Link>
                    </p>
                  </div>
                </Form>
              </Card.Body>
            </Card>

            <div className="text-center mt-4">
              <p className="text-muted small">
                <i className="fas fa-lock me-1"></i>
                Your data is securely encrypted
              </p>
            </div>
          </Col>
        </Row>
      </Container>
    </div>
  );
};

export default ForgotPasswordPage;