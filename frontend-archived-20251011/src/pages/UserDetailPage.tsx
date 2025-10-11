import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Row, Col, Card, Button, Spinner, Badge, Tabs, Tab } from 'react-bootstrap';
import { getUserById, updateUser, deleteUser } from '../services/userService';
import { UserProfile, UpdateUserRequest } from '../types';

const UserDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');
  const [editing, setEditing] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    role: 'user',
    status: 'active'
  });

  useEffect(() => {
    if (id) {
      loadUser(parseInt(id));
    }
  }, [id]);

  const loadUser = async (userId: number) => {
    setLoading(true);
    try {
      const response = await getUserById(userId);
      setUser(response);
      setFormData({
        name: response.name,
        email: response.email,
        role: response.role,
        status: response.status
      });
    } catch (error) {
      console.error('Failed to load user', error);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = () => {
    setEditing(true);
  };

  const handleCancelEdit = () => {
    setEditing(false);
    if (user) {
      setFormData({
        name: user.name,
        email: user.email,
        role: user.role,
        status: user.status
      });
    }
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;
    
    try {
      const data: UpdateUserRequest = formData;
      const updatedUser = await updateUser(user.id, data);
      setUser(updatedUser);
      setEditing(false);
    } catch (error) {
      console.error('Failed to update user', error);
    }
  };

  const handleDelete = async () => {
    if (!user) return;
    
    if (window.confirm('Are you sure you want to delete this user? This action cannot be undone.')) {
      try {
        await deleteUser(user.id);
        navigate('/users');
      } catch (error) {
        console.error('Failed to delete user', error);
      }
    }
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'active': return 'bg-success';
      case 'inactive': return 'bg-secondary';
      default: return 'bg-secondary';
    }
  };

  const getRoleBadgeClass = (role: string) => {
    switch (role) {
      case 'admin': return 'bg-danger';
      case 'lawyer': return 'bg-primary';
      case 'user': return 'bg-info';
      default: return 'bg-secondary';
    }
  };

  // Get role display text
  const getRoleText = (role: string) => {
    switch (role) {
      case 'admin': return 'Administrator';
      case 'lawyer': return 'Lawyer';
      case 'user': return 'User';
      default: return role;
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return 'Active';
      case 'inactive': return 'Inactive';
      default: return status;
    }
  };

  if (loading) {
    return (
      <div className="d-flex justify-content-center align-items-center" style={{ height: '50vh' }}>
        <Spinner animation="border" />
        <span className="ms-2">Loading user details...</span>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="text-center py-5">
        <i className="fas fa-exclamation-triangle fa-3x text-warning mb-3"></i>
        <h5>User not found</h5>
        <p className="text-muted">The requested user could not be found</p>
        <Button variant="primary" onClick={() => navigate('/users')}>
          <i className="fas fa-arrow-left me-2"></i>
          Back to Users
        </Button>
      </div>
    );
  }

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <Button variant="outline-secondary" onClick={() => navigate('/users')} className="mb-2">
            <i className="fas fa-arrow-left me-2"></i>
            Back to Users
          </Button>
          <h1>User: {user.name}</h1>
        </div>
        <div className="d-flex">
          <Dropdown className="me-2">
            <Dropdown.Toggle variant="outline-secondary" id="actions-dropdown">
              <i className="fas fa-ellipsis-h me-2"></i>
              Actions
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item onClick={handleEdit}>
                <i className="fas fa-edit me-2"></i>
                Edit User
              </Dropdown.Item>
              <Dropdown.Item>
                <i className="fas fa-copy me-2"></i>
                Duplicate User
              </Dropdown.Item>
              <Dropdown.Item>
                <i className="fas fa-history me-2"></i>
                View History
              </Dropdown.Item>
              <Dropdown.Divider />
              <Dropdown.Item className="text-danger" onClick={handleDelete}>
                <i className="fas fa-trash me-2"></i>
                Delete User
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          <Button variant="primary" onClick={handleEdit}>
            <i className="fas fa-edit me-2"></i>
            Edit User
          </Button>
        </div>
      </div>

      <Row>
        <Col md={8}>
          <Card className="mb-4">
            <Card.Header className="d-flex justify-content-between align-items-center">
              <span>User Overview</span>
              <div>
                <Badge bg={getStatusBadgeClass(user.status)} className="me-2">
                  {getStatusText(user.status)}
                </Badge>
                <Badge bg={getRoleBadgeClass(user.role)}>
                  {getRoleText(user.role)}
                </Badge>
              </div>
            </Card.Header>
            <Card.Body>
              {editing ? (
                <Form onSubmit={handleSubmit}>
                  <Row>
                    <Col md={12}>
                      <Form.Group className="mb-3">
                        <Form.Label>Name <span className="text-danger">*</span></Form.Label>
                        <Form.Control
                          type="text"
                          name="name"
                          value={formData.name}
                          onChange={handleChange}
                          required
                          placeholder="Enter user name"
                        />
                      </Form.Group>
                    </Col>
                  </Row>
                  <Row>
                    <Col md={12}>
                      <Form.Group className="mb-3">
                        <Form.Label>Email <span className="text-danger">*</span></Form.Label>
                        <Form.Control
                          type="email"
                          name="email"
                          value={formData.email}
                          onChange={handleChange}
                          required
                          placeholder="Enter user email"
                        />
                      </Form.Group>
                    </Col>
                  </Row>
                  <Row>
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>Role</Form.Label>
                        <Form.Select
                          name="role"
                          value={formData.role}
                          onChange={handleChange}
                        >
                          <option value="user">{getRoleText('user')}</option>
                          <option value="lawyer">{getRoleText('lawyer')}</option>
                          <option value="admin">{getRoleText('admin')}</option>
                        </Form.Select>
                      </Form.Group>
                    </Col>
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>Status</Form.Label>
                        <Form.Select
                          name="status"
                          value={formData.status}
                          onChange={handleChange}
                        >
                          <option value="active">{getStatusText('active')}</option>
                          <option value="inactive">{getStatusText('inactive')}</option>
                        </Form.Select>
                      </Form.Group>
                    </Col>
                  </Row>
                  <div className="d-flex justify-content-end">
                    <Button variant="secondary" onClick={handleCancelEdit} className="me-2">
                      <i className="fas fa-times me-2"></i>
                      Cancel
                    </Button>
                    <Button variant="primary" type="submit">
                      <i className="fas fa-save me-2"></i>
                      Save Changes
                    </Button>
                  </div>
                </Form>
              ) : (
                <Row>
                  <Col md={6}>
                    <div className="mb-3">
                      <small className="text-muted">Name</small>
                      <div className="fw-bold">{user.name}</div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Email</small>
                      <div className="fw-bold">{user.email}</div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Role</small>
                      <div className="fw-bold">
                        <Badge bg={getRoleBadgeClass(user.role)}>
                          {getRoleText(user.role)}
                        </Badge>
                      </div>
                    </div>
                  </Col>
                  <Col md={6}>
                    <div className="mb-3">
                      <small className="text-muted">Status</small>
                      <div className="fw-bold">
                        <Badge bg={getStatusBadgeClass(user.status)}>
                          {getStatusText(user.status)}
                        </Badge>
                      </div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Created</small>
                      <div className="fw-bold">
                        {new Date(user.created_at).toLocaleDateString()} at {new Date(user.created_at).toLocaleTimeString()}
                      </div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Last Login</small>
                      <div className="fw-bold">
                        {user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'Never'}
                      </div>
                    </div>
                  </Col>
                </Row>
              )}
            </Card.Body>
          </Card>

          <Tabs
            activeKey={activeTab}
            onSelect={(k) => setActiveTab(k || 'overview')}
            className="mb-4"
          >
            <Tab eventKey="overview" title="Overview">
              <Card>
                <Card.Body>
                  <h5>User Timeline</h5>
                  <div className="timeline">
                    <div className="timeline-item mb-4">
                      <div className="d-flex">
                        <div className="timeline-icon bg-primary rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-user-plus text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">User Created</h6>
                          <p className="mb-0 text-muted">User account was created in the system</p>
                          <small className="text-muted">
                            {new Date(user.created_at).toLocaleString()}
                          </small>
                        </div>
                      </div>
                    </div>
                    <div className="timeline-item mb-4">
                      <div className="d-flex">
                        <div className="timeline-icon bg-success rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-user-check text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">First Login</h6>
                          <p className="mb-0 text-muted">User logged in to the system for the first time</p>
                          <small className="text-muted">1 day ago</small>
                        </div>
                      </div>
                    </div>
                    <div className="timeline-item">
                      <div className="d-flex">
                        <div className="timeline-icon bg-warning rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-user-edit text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">Profile Updated</h6>
                          <p className="mb-0 text-muted">User profile information was updated</p>
                          <small className="text-muted">2 hours ago</small>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="activity" title="Activity">
              <Card>
                <Card.Body>
                  <h5>User Activity</h5>
                  <div className="activity-list">
                    <div className="activity-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between">
                        <h6 className="mb-1">Client Created</h6>
                        <span className="badge bg-primary">New</span>
                      </div>
                      <p className="mb-1 text-muted">Created new client John Smith</p>
                      <div className="d-flex justify-content-between">
                        <small className="text-muted">
                          <i className="fas fa-calendar me-1"></i>
                          2 hours ago
                        </small>
                        <small className="text-muted">
                          <i className="fas fa-desktop me-1"></i>
                          via Web Interface
                        </small>
                      </div>
                    </div>
                    <div className="activity-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between">
                        <h6 className="mb-1">Case Updated</h6>
                        <span className="badge bg-info">Update</span>
                      </div>
                      <p className="mb-1 text-muted">Updated case #1234 status to Active</p>
                      <div className="d-flex justify-content-between">
                        <small className="text-muted">
                          <i className="fas fa-calendar me-1"></i>
                          1 day ago
                        </small>
                        <small className="text-muted">
                          <i className="fas fa-mobile me-1"></i>
                          via Mobile App
                        </small>
                      </div>
                    </div>
                    <div className="activity-item p-3 border rounded">
                      <div className="d-flex justify-content-between">
                        <h6 className="mb-1">Document Uploaded</h6>
                        <span className="badge bg-success">Upload</span>
                      </div>
                      <p className="mb-1 text-muted">Uploaded contract agreement for case #5678</p>
                      <div className="d-flex justify-content-between">
                        <small className="text-muted">
                          <i className="fas fa-calendar me-1"></i>
                          2 days ago
                        </small>
                        <small className="text-muted">
                          <i className="fas fa-desktop me-1"></i>
                          via Web Interface
                        </small>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="permissions" title="Permissions">
              <Card>
                <Card.Body>
                  <h5>User Permissions</h5>
                  <div className="permission-list">
                    <div className="permission-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div>
                          <h6 className="mb-1">Client Management</h6>
                          <p className="mb-0 text-muted">View, create, edit, and delete client records</p>
                        </div>
                        <Badge bg="success">Granted</Badge>
                      </div>
                    </div>
                    <div className="permission-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div>
                          <h6 className="mb-1">Case Management</h6>
                          <p className="mb-0 text-muted">View, create, edit, and delete case records</p>
                        </div>
                        <Badge bg="success">Granted</Badge>
                      </div>
                    </div>
                    <div className="permission-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div>
                          <h6 className="mb-1">User Management</h6>
                          <p className="mb-0 text-muted">View, create, edit, and delete user accounts</p>
                        </div>
                        <Badge bg={user.role === 'admin' ? 'success' : 'danger'}>
                          {user.role === 'admin' ? 'Granted' : 'Denied'}
                        </Badge>
                      </div>
                    </div>
                    <div className="permission-item p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div>
                          <h6 className="mb-1">System Administration</h6>
                          <p className="mb-0 text-muted">Access to system configuration and settings</p>
                        </div>
                        <Badge bg={user.role === 'admin' ? 'success' : 'danger'}>
                          {user.role === 'admin' ? 'Granted' : 'Denied'}
                        </Badge>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
          </Tabs>
        </Col>
        
        <Col md={4}>
          <Card className="mb-4">
            <Card.Header>User Information</Card.Header>
            <Card.Body>
              <div className="d-flex align-items-center mb-4">
                <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3" style={{ width: '64px', height: '64px' }}>
                  <i className="fas fa-user fa-2x text-muted"></i>
                </div>
                <div>
                  <h5 className="mb-1">{user.name}</h5>
                  <p className="text-muted mb-0">User ID: #{user.id}</p>
                  <Badge bg={getStatusBadgeClass(user.status)} className="mt-1">
                    {getStatusText(user.status)}
                  </Badge>
                </div>
              </div>
              
              <div className="mb-3">
                <small className="text-muted">Email</small>
                <div className="fw-bold">{user.email}</div>
              </div>
              
              <div className="mb-3">
                <small className="text-muted">Role</small>
                <div className="fw-bold">
                  <Badge bg={getRoleBadgeClass(user.role)}>
                    {getRoleText(user.role)}
                  </Badge>
                </div>
              </div>
              
              <div className="mb-3">
                <small className="text-muted">Created</small>
                <div className="fw-bold">
                  {new Date(user.created_at).toLocaleDateString()}
                </div>
              </div>
              
              <div className="mb-3">
                <small className="text-muted">Last Login</small>
                <div className="fw-bold">
                  {user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'Never'}
                </div>
              </div>
              
              <div className="d-grid gap-2">
                <Button variant="primary">
                  <i className="fas fa-envelope me-2"></i>
                  Send Message
                </Button>
                <Button variant="outline-primary">
                  <i className="fas fa-history me-2"></i>
                  View History
                </Button>
              </div>
            </Card.Body>
          </Card>
          
          <Card className="mb-4">
            <Card.Header>Statistics</Card.Header>
            <Card.Body>
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Total Cases</span>
                  <span className="fw-bold">12</span>
                </div>
                <div className="progress mt-1">
                  <div 
                    className="progress-bar bg-primary" 
                    role="progressbar" 
                    style={{ width: '60%' }}
                  ></div>
                </div>
              </div>
              
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Active Cases</span>
                  <span className="fw-bold">8</span>
                </div>
                <div className="progress mt-1">
                  <div 
                    className="progress-bar bg-success" 
                    role="progressbar" 
                    style={{ width: '40%' }}
                  ></div>
                </div>
              </div>
              
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Pending Cases</span>
                  <span className="fw-bold">3</span>
                </div>
                <div className="progress mt-1">
                  <div 
                    className="progress-bar bg-warning" 
                    role="progressbar" 
                    style={{ width: '15%' }}
                  ></div>
                </div>
              </div>
              
              <div className="stat-item">
                <div className="d-flex justify-content-between">
                  <span>Closed Cases</span>
                  <span className="fw-bold">1</span>
                </div>
                <div className="progress mt-1">
                  <div 
                    className="progress-bar bg-info" 
                    role="progressbar" 
                    style={{ width: '5%' }}
                  ></div>
                </div>
              </div>
            </Card.Body>
          </Card>
          
          <Card>
            <Card.Header>Security</Card.Header>
            <Card.Body>
              <div className="security-item mb-3">
                <div className="d-flex justify-content-between align-items-center">
                  <div>
                    <h6 className="mb-1">Two-Factor Authentication</h6>
                    <p className="mb-0 text-muted">Additional security layer for your account</p>
                  </div>
                  <Badge bg="success">Enabled</Badge>
                </div>
              </div>
              <div className="security-item mb-3">
                <div className="d-flex justify-content-between align-items-center">
                  <div>
                    <h6 className="mb-1">Password Strength</h6>
                    <p className="mb-0 text-muted">Strong password recommended</p>
                  </div>
                  <Badge bg="success">Strong</Badge>
                </div>
              </div>
              <div className="security-item">
                <div className="d-flex justify-content-between align-items-center">
                  <div>
                    <h6 className="mb-1">Last Password Change</h6>
                    <p className="mb-0 text-muted">When was your password last updated</p>
                  </div>
                  <Badge bg="secondary">30 days ago</Badge>
                </div>
              </div>
              <div className="d-grid gap-2 mt-3">
                <Button variant="outline-primary">
                  <i className="fas fa-key me-2"></i>
                  Change Password
                </Button>
                <Button variant="outline-secondary">
                  <i className="fas fa-shield-alt me-2"></i>
                  Security Settings
                </Button>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default UserDetailPage;