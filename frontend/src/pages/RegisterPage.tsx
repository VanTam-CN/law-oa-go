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
import { useDispatch, useSelector } from "react-redux";
import { register } from "../store/slices/authSlice";
import { AppDispatch, RootState } from "../store";
import { RegisterRequest } from "../types";
import "./AuthPage.css";

const RegisterPage: React.FC = () => {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [role, setRole] = useState("user");
  const [localError, setLocalError] = useState("");
  const [agreeToTerms, setAgreeToTerms] = useState(false);
  const dispatch = useDispatch<AppDispatch>();
  const {
    loading,
    error: reduxError,
    isAuthenticated,
  } = useSelector((state: RootState) => state.auth);
  const navigate = useNavigate();

  // 如果已经认证，重定向到首页
  React.useEffect(() => {
    if (isAuthenticated) {
      navigate("/");
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (password !== confirmPassword) {
      setLocalError("Passwords do not match");
      return;
    }

    if (password.length < 8) {
      setLocalError("Password must be at least 8 characters long");
      return;
    }

    if (!agreeToTerms) {
      setLocalError(
        "You must agree to the Terms of Service and Privacy Policy",
      );
      return;
    }

    try {
      const data: RegisterRequest = { name, email, password, role };
      const result = await dispatch(register(data)).unwrap();
      if (result) {
        navigate("/");
      }
    } catch (err: any) {
      // 错误已经在Redux slice中处理
    }
  };

  return (
    <div className="auth-page register-page">
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
                  <h2>Create Account</h2>
                  <p className="text-muted">
                    Join our legal practice management platform
                  </p>
                </div>

                {localError && <Alert variant="danger">{localError}</Alert>}
                {reduxError && <Alert variant="danger">{reduxError}</Alert>}

                <Form onSubmit={handleSubmit}>
                  <Form.Group className="mb-3">
                    <Form.Label>
                      Full Name <span className="text-danger">*</span>
                    </Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <i className="fas fa-user"></i>
                      </InputGroup.Text>
                      <Form.Control
                        type="text"
                        placeholder="Enter your full name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        required
                        autoFocus
                      />
                    </InputGroup>
                  </Form.Group>

                  <Form.Group className="mb-3">
                    <Form.Label>
                      Email Address <span className="text-danger">*</span>
                    </Form.Label>
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
                      />
                    </InputGroup>
                  </Form.Group>

                  <Form.Group className="mb-3">
                    <Form.Label>
                      Password <span className="text-danger">*</span>
                    </Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <i className="fas fa-lock"></i>
                      </InputGroup.Text>
                      <Form.Control
                        type="password"
                        placeholder="Create a password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                      />
                    </InputGroup>
                    <Form.Text className="text-muted">
                      Password must be at least 8 characters long and contain a
                      mix of letters, numbers, and symbols.
                    </Form.Text>
                  </Form.Group>

                  <Form.Group className="mb-3">
                    <Form.Label>
                      Confirm Password <span className="text-danger">*</span>
                    </Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <i className="fas fa-lock"></i>
                      </InputGroup.Text>
                      <Form.Control
                        type="password"
                        placeholder="Confirm your password"
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        required
                      />
                    </InputGroup>
                  </Form.Group>

                  <Form.Group className="mb-3">
                    <Form.Label>Role</Form.Label>
                    <Form.Select
                      value={role}
                      onChange={(e) => setRole(e.target.value)}
                    >
                      <option value="user">User</option>
                      <option value="lawyer">Lawyer</option>
                      <option value="admin">Administrator</option>
                    </Form.Select>
                  </Form.Group>

                  <Form.Group className="mb-4">
                    <Form.Check
                      required
                      type="checkbox"
                      id="terms"
                      label={
                        <>
                          I agree to the{" "}
                          <Link
                            to="/terms"
                            target="_blank"
                            className="text-decoration-none"
                          >
                            Terms of Service
                          </Link>{" "}
                          and{" "}
                          <Link
                            to="/privacy"
                            target="_blank"
                            className="text-decoration-none"
                          >
                            Privacy Policy
                          </Link>
                        </>
                      }
                      checked={agreeToTerms}
                      onChange={(e) => setAgreeToTerms(e.target.checked)}
                    />
                  </Form.Group>

                  <Button
                    variant="primary"
                    type="submit"
                    className="w-100 mb-3 auth-button"
                    disabled={loading}
                  >
                    {loading ? (
                      <>
                        <i className="fas fa-spinner fa-spin me-2"></i>
                        Creating Account...
                      </>
                    ) : (
                      <>
                        <i className="fas fa-user-plus me-2"></i>
                        Create Account
                      </>
                    )}
                  </Button>

                  <div className="text-center">
                    <p className="mb-0">
                      Already have an account?{" "}
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
                <i className="fas fa-shield-alt me-1"></i>
                Your data is securely encrypted
              </p>
            </div>
          </Col>
        </Row>
      </Container>
    </div>
  );
};

export default RegisterPage;
