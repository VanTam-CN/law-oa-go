import React, { useState, useEffect } from "react";
import {
  Row,
  Col,
  Card,
  Form,
  Button,
  Alert,
  Tab,
  Nav,
  Badge,
  Spinner,
} from "react-bootstrap";
import { useDispatch, useSelector } from "react-redux";
import { updateUserProfile, changePassword } from "../services/authService";
import { refreshUser } from "../store/slices/authSlice";
import { AppDispatch, RootState } from "../store";
import { UpdateUserRequest } from "../types";
import "./ProfilePage.css";

const ProfilePage: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { user, loading } = useSelector((state: RootState) => state.auth);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [profileLoading, setProfileLoading] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [profileSuccess, setProfileSuccess] = useState("");
  const [passwordSuccess, setPasswordSuccess] = useState("");
  const [profileError, setProfileError] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [activeTab, setActiveTab] = useState("profile");

  useEffect(() => {
    if (user) {
      setName(user.name);
      setEmail(user.email);
    }
  }, [user]);

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setProfileLoading(true);
    setProfileSuccess("");
    setProfileError("");

    try {
      const data: UpdateUserRequest = {
        name,
        email,
        role: user?.role,
        status: user?.status,
      };
      await updateUserProfile(data);
      await dispatch(refreshUser()).unwrap();
      setProfileSuccess("Profile updated successfully");
    } catch (err: any) {
      setProfileError(
        err.response?.data?.error?.message || "Failed to update profile",
      );
    } finally {
      setProfileLoading(false);
    }
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (newPassword !== confirmPassword) {
      setPasswordError("New passwords do not match");
      return;
    }

    if (newPassword.length < 8) {
      setPasswordError("Password must be at least 8 characters long");
      return;
    }

    setPasswordLoading(true);
    setPasswordSuccess("");
    setPasswordError("");

    try {
      await changePassword(currentPassword, newPassword);
      setPasswordSuccess("Password changed successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err: any) {
      setPasswordError(
        err.response?.data?.error?.message || "Failed to change password",
      );
    } finally {
      setPasswordLoading(false);
    }
  };

  const getRoleBadgeClass = (role: string) => {
    switch (role) {
      case "admin":
        return "bg-danger";
      case "lawyer":
        return "bg-primary";
      case "user":
        return "bg-info";
      default:
        return "bg-secondary";
    }
  };

  // Get role display text
  const getRoleText = (role: string) => {
    switch (role) {
      case "admin":
        return "Administrator";
      case "lawyer":
        return "Lawyer";
      case "user":
        return "User";
      default:
        return role;
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case "active":
        return "Active";
      case "inactive":
        return "Inactive";
      default:
        return status;
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Profile</h1>
        <Button variant="outline-primary">
          <i className="fas fa-download me-2"></i>
          Export Profile
        </Button>
      </div>

      <Row>
        <Col md={4}>
          <Card className="mb-4 profile-sidebar">
            <Card.Body className="text-center">
              <div className="mb-3">
                <div className="profile-avatar bg-light rounded-circle d-flex align-items-center justify-content-center mx-auto">
                  <i className="fas fa-user fa-3x text-muted"></i>
                </div>
              </div>
              <h5>{user?.name}</h5>
              <p className="text-muted">{user?.email}</p>
              <Badge
                variant={getRoleBadgeClass(user?.role || "user")}
                className="mb-3"
              >
                {getRoleText(user?.role || "user")}
              </Badge>
              <div className="d-grid gap-2">
                <Button variant="primary">
                  <i className="fas fa-camera me-2"></i>
                  Change Avatar
                </Button>
                <Button variant="outline-secondary">
                  <i className="fas fa-lock me-2"></i>
                  Two-Factor Authentication
                </Button>
              </div>
            </Card.Body>
          </Card>

          <Card>
            <Card.Header>
              <i className="fas fa-info-circle me-2"></i>
              Account Information
            </Card.Header>
            <Card.Body>
              <div className="mb-3">
                <small className="text-muted">Account Status</small>
                <div className="fw-bold">
                  <i className="fas fa-check-circle text-success me-2"></i>
                  {getStatusText(user?.status || "active")}
                </div>
              </div>
              <div className="mb-3">
                <small className="text-muted">Member Since</small>
                <div className="fw-bold">
                  {user?.created_at
                    ? new Date(user.created_at).toLocaleDateString()
                    : "Unknown"}
                </div>
              </div>
              <div className="mb-3">
                <small className="text-muted">Last Login</small>
                <div className="fw-bold">
                  {user?.updated_at
                    ? new Date(user.updated_at).toLocaleDateString()
                    : "Never"}
                </div>
              </div>
              <div>
                <small className="text-muted">Account ID</small>
                <div className="fw-bold text-muted">#{user?.id}</div>
              </div>
            </Card.Body>
          </Card>
        </Col>

        <Col md={8}>
          <Tab.Container
            activeKey={activeTab}
            onSelect={(k) => setActiveTab(k || "profile")}
          >
            <Card>
              <Card.Header>
                <Nav
                  variant="tabs"
                  activeKey={activeTab}
                  onSelect={(k) => setActiveTab(k || "profile")}
                >
                  <Nav.Item>
                    <Nav.Link eventKey="profile">
                      <i className="fas fa-user-edit me-2"></i>
                      Edit Profile
                    </Nav.Link>
                  </Nav.Item>
                  <Nav.Item>
                    <Nav.Link eventKey="security">
                      <i className="fas fa-shield-alt me-2"></i>
                      Security
                    </Nav.Link>
                  </Nav.Item>
                  <Nav.Item>
                    <Nav.Link eventKey="notifications">
                      <i className="fas fa-bell me-2"></i>
                      Notifications
                    </Nav.Link>
                  </Nav.Item>
                  <Nav.Item>
                    <Nav.Link eventKey="preferences">
                      <i className="fas fa-cog me-2"></i>
                      Preferences
                    </Nav.Link>
                  </Nav.Item>
                </Nav>
              </Card.Header>
              <Card.Body>
                <Tab.Content>
                  <Tab.Pane eventKey="profile">
                    {profileSuccess && (
                      <Alert variant="success">{profileSuccess}</Alert>
                    )}
                    {profileError && (
                      <Alert variant="danger">{profileError}</Alert>
                    )}
                    <Form onSubmit={handleProfileSubmit}>
                      <Row>
                        <Col md={12}>
                          <Form.Group className="mb-3">
                            <Form.Label>
                              Name <span className="text-danger">*</span>
                            </Form.Label>
                            <Form.Control
                              type="text"
                              value={name}
                              onChange={(e) => setName(e.target.value)}
                              required
                              placeholder="Enter your name"
                            />
                          </Form.Group>
                        </Col>
                      </Row>
                      <Row>
                        <Col md={12}>
                          <Form.Group className="mb-3">
                            <Form.Label>
                              Email <span className="text-danger">*</span>
                            </Form.Label>
                            <Form.Control
                              type="email"
                              value={email}
                              onChange={(e) => setEmail(e.target.value)}
                              required
                              placeholder="Enter your email"
                            />
                          </Form.Group>
                        </Col>
                      </Row>
                      <Row>
                        <Col md={12}>
                          <Form.Group className="mb-3">
                            <Form.Label>Role</Form.Label>
                            <Form.Control
                              type="text"
                              value={getRoleText(user?.role || "user")}
                              disabled
                            />
                          </Form.Group>
                        </Col>
                      </Row>
                      <div className="d-flex justify-content-end">
                        <Button
                          variant="primary"
                          type="submit"
                          disabled={profileLoading}
                        >
                          {profileLoading ? (
                            <>
                              <i className="fas fa-spinner fa-spin me-2"></i>
                              Updating...
                            </>
                          ) : (
                            <>
                              <i className="fas fa-save me-2"></i>
                              Update Profile
                            </>
                          )}
                        </Button>
                      </div>
                    </Form>
                  </Tab.Pane>

                  <Tab.Pane eventKey="security">
                    {passwordSuccess && (
                      <Alert variant="success">{passwordSuccess}</Alert>
                    )}
                    {passwordError && (
                      <Alert variant="danger">{passwordError}</Alert>
                    )}
                    <Form onSubmit={handlePasswordSubmit}>
                      <Form.Group className="mb-3">
                        <Form.Label>
                          Current Password{" "}
                          <span className="text-danger">*</span>
                        </Form.Label>
                        <Form.Control
                          type="password"
                          value={currentPassword}
                          onChange={(e) => setCurrentPassword(e.target.value)}
                          required
                          placeholder="Enter current password"
                        />
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Label>
                          New Password <span className="text-danger">*</span>
                        </Form.Label>
                        <Form.Control
                          type="password"
                          value={newPassword}
                          onChange={(e) => setNewPassword(e.target.value)}
                          required
                          placeholder="Enter new password"
                        />
                        <Form.Text className="text-muted">
                          Password must be at least 8 characters long and
                          contain a mix of letters, numbers, and symbols.
                        </Form.Text>
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Label>
                          Confirm New Password{" "}
                          <span className="text-danger">*</span>
                        </Form.Label>
                        <Form.Control
                          type="password"
                          value={confirmPassword}
                          onChange={(e) => setConfirmPassword(e.target.value)}
                          required
                          placeholder="Confirm new password"
                        />
                      </Form.Group>
                      <div className="d-flex justify-content-end">
                        <Button
                          variant="primary"
                          type="submit"
                          disabled={passwordLoading}
                        >
                          {passwordLoading ? (
                            <>
                              <i className="fas fa-spinner fa-spin me-2"></i>
                              Changing...
                            </>
                          ) : (
                            <>
                              <i className="fas fa-key me-2"></i>
                              Change Password
                            </>
                          )}
                        </Button>
                      </div>
                    </Form>
                  </Tab.Pane>

                  <Tab.Pane eventKey="notifications">
                    <h5>
                      <i className="fas fa-bell me-2"></i> Notification
                      Preferences
                    </h5>
                    <p className="text-muted mb-4">
                      Configure how you receive notifications from the system.
                    </p>
                    <Form>
                      <Form.Group className="mb-3">
                        <Form.Check
                          type="switch"
                          id="email-notifications"
                          label="Email Notifications"
                          defaultChecked
                        />
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Check
                          type="switch"
                          id="sms-notifications"
                          label="SMS Notifications"
                        />
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Check
                          type="switch"
                          id="push-notifications"
                          label="Push Notifications"
                          defaultChecked
                        />
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Check
                          type="switch"
                          id="case-updates"
                          label="Case Updates"
                          defaultChecked
                        />
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Check
                          type="switch"
                          id="client-updates"
                          label="Client Updates"
                          defaultChecked
                        />
                      </Form.Group>
                      <div className="d-flex justify-content-end">
                        <Button variant="primary">
                          <i className="fas fa-save me-2"></i>
                          Save Preferences
                        </Button>
                      </div>
                    </Form>
                  </Tab.Pane>

                  <Tab.Pane eventKey="preferences">
                    <h5>
                      <i className="fas fa-cog me-2"></i> System Preferences
                    </h5>
                    <p className="text-muted mb-4">
                      Customize your system experience.
                    </p>
                    <Form>
                      <Form.Group className="mb-3">
                        <Form.Label>Language</Form.Label>
                        <Form.Select defaultValue="en">
                          <option value="en">English</option>
                          <option value="zh">中文</option>
                          <option value="es">Español</option>
                          <option value="fr">Français</option>
                        </Form.Select>
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Label>Theme</Form.Label>
                        <Form.Select defaultValue="light">
                          <option value="light">Light</option>
                          <option value="dark">Dark</option>
                          <option value="auto">Auto</option>
                        </Form.Select>
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Label>Date Format</Form.Label>
                        <Form.Select defaultValue="mm/dd/yyyy">
                          <option value="mm/dd/yyyy">MM/DD/YYYY</option>
                          <option value="dd/mm/yyyy">DD/MM/YYYY</option>
                          <option value="yyyy-mm-dd">YYYY-MM-DD</option>
                        </Form.Select>
                      </Form.Group>
                      <Form.Group className="mb-3">
                        <Form.Label>Time Zone</Form.Label>
                        <Form.Select defaultValue="UTC">
                          <option value="UTC">UTC</option>
                          <option value="America/New_York">Eastern Time</option>
                          <option value="America/Chicago">Central Time</option>
                          <option value="America/Denver">Mountain Time</option>
                          <option value="America/Los_Angeles">
                            Pacific Time
                          </option>
                          <option value="Asia/Shanghai">
                            China Standard Time
                          </option>
                        </Form.Select>
                      </Form.Group>
                      <div className="d-flex justify-content-end">
                        <Button variant="primary">
                          <i className="fas fa-save me-2"></i>
                          Save Preferences
                        </Button>
                      </div>
                    </Form>
                  </Tab.Pane>
                </Tab.Content>
              </Card.Body>
            </Card>
          </Tab.Container>
        </Col>
      </Row>
    </div>
  );
};

export default ProfilePage;
