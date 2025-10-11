import React, { useState } from 'react';
import { useNavigate, Link, useSearchParams } from 'react-router-dom';
import { resetPassword } from '../services/authService';
import './AuthPage.css';

const ResetPasswordPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const navigate = useNavigate();
  const token = searchParams.get('token');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters long');
      return;
    }

    if (!token) {
      setError('Invalid or missing reset token');
      return;
    }

    setLoading(true);
    setError('');
    setSuccess(false);

    try {
      await resetPassword(token, password);
      setSuccess(true);
      setTimeout(() => navigate('/login'), 3000);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to reset password. Please try again.');
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
                  <p className="text-muted">Create a new password for your account</p>
                </div>

                {!token && (
                  <Alert variant="danger">
                    Invalid or missing reset token. Please use the link from your email or{' '}
                    <Link to="/forgot-password" className="text-decoration-none">
                      request a new reset link
                    </Link>.
                  </Alert>
                )}

                {error && <Alert variant="danger">{error}</Alert>}
                {success && (
                  <Alert variant="success">
                    Password reset successful! Redirecting to login page...
                  </Alert>
                )}

                <Form onSubmit={handleSubmit}>
                  <Form.Group className="mb-3">
                    <Form.Label>New Password <span className="text-danger">*</span></Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <i className="fas fa-lock"></i>
                      </InputGroup.Text>
                      <Form.Control
                        type="password"
                        placeholder="Enter new password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                        minLength={8}
                        disabled={success || !token}
                        autoFocus
                      />
                    </InputGroup>
                    <Form.Text className="text-muted">
                      Password must be at least 8 characters long and contain a mix of letters, numbers, and symbols.
                    </Form.Text>
                  </Form.Group>

                  <Form.Group className="mb-4">
                    <Form.Label>Confirm New Password <span className="text-danger">*</span></Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <i className="fas fa-lock"></i>
                      </InputGroup.Text>
                      <Form.Control
                        type="password"
                        placeholder="Confirm new password"
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        required
                        disabled={success || !token}
                      />
                    </InputGroup>
                  </Form.Group>

                  <Button
                    variant="primary"
                    type="submit"
                    className="w-100 mb-3 auth-button"
                    disabled={loading || success || !token}
                  >
                    {loading ? (
                      <>
                        <i className="fas fa-spinner fa-spin me-2"></i>
                        Resetting Password...
                      </>
                    ) : success ? (
                      <>
                        <i className="fas fa-check me-2"></i>
                        Password Reset!
                      </>
                    ) : (
                      <>
                        <i className="fas fa-key me-2"></i>
                        Reset Password
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

export default ResetPasswordPage;