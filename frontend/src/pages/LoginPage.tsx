import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  Form,
  Button,
  Container,
  Row,
  Col,
  Alert,
  Card,
  InputGroup,
} from "react-bootstrap";
import { useDispatch } from "react-redux";
import { login } from "../store/slices/authSlice";
import { LoginRequest } from "../types";
import "./AuthPage.css";

const LoginPage: React.FC = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const data: LoginRequest = { email, password };
      await dispatch(login(data) as any);
      navigate("/");
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
          "Login failed. Please check your credentials and try again.",
      );
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

            <Card className="auth-card shadow-lg">
              <Card.Body className="p-5">
                <div className="text-center mb-4">
                  <h2>Welcome Back</h2>
                  <p className="text-muted">Sign in to your account</p>
                </div>

                {error && <Alert variant="danger">{error}</Alert>}

                <Form onSubmit={handleSubmit}>
                  <Form.Group className="mb-3">
                    <Form.Label>Email Address</Form.Label>
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
                      />
                    </InputGroup>
                  </Form.Group>

                  <Form.Group className="mb-3">
                    <Form.Label>Password</Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <i className="fas fa-lock"></i>
                      </InputGroup.Text>
                      <Form.Control
                        type="password"
                        placeholder="Enter your password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                      />
                    </InputGroup>
                  </Form.Group>

                  <div className="d-flex justify-content-between align-items-center mb-4">
                    <Form.Check
                      type="checkbox"
                      id="remember-me"
                      label="Remember me"
                      checked={rememberMe}
                      onChange={(e) => setRememberMe(e.target.checked)}
                    />
                    <Link
                      to="/forgot-password"
                      className="text-decoration-none"
                    >
                      Forgot password?
                    </Link>
                  </div>

                  <Button
                    variant="primary"
                    type="submit"
                    className="w-100 mb-3 auth-button"
                    disabled={loading}
                  >
                    {loading ? (
                      <>
                        <i className="fas fa-spinner fa-spin me-2"></i>
                        Signing in...
                      </>
                    ) : (
                      <>
                        <i className="fas fa-sign-in-alt me-2"></i>
                        Sign In
                      </>
                    )}
                  </Button>

                  <div className="text-center">
                    <p className="mb-0">
                      Don't have an account?{" "}
                      <Link to="/register" className="text-decoration-none">
                        Sign up
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

export default LoginPage;
