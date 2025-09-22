import React, { useState, useEffect } from 'react';
import { Spinner, Button, Card, Row, Col, Form, InputGroup, Table, Badge, Modal } from 'react-bootstrap';
import { getUsers, createUser, updateUser, deleteUser } from '../services/userService';
import { UserProfile, UserListRequest, CreateUserRequest, UpdateUserRequest } from '../types';

const UsersPage: React.FC = () => {
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingUser, setEditingUser] = useState<UserProfile | null>(null);
  const [search, setSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    role: 'user',
    status: 'active',
    password: ''
  });

  useEffect(() => {
    loadUsers();
  }, [currentPage, search, roleFilter, statusFilter]);

  const loadUsers = async () => {
    setLoading(true);
    try {
      const params: UserListRequest = {
        page: currentPage,
        page_size: 10,
        search: search || undefined,
        role: roleFilter !== 'all' ? roleFilter : undefined,
        status: statusFilter !== 'all' ? statusFilter : undefined
      };
      const response = await getUsers(params);
      setUsers(response.data);
      setTotalPages(Math.ceil(response.pagination.total / response.pagination.page_size));
    } catch (error) {
      console.error('Failed to load users', error);
    } finally {
      setLoading(false);
    }
  };

  const handleShowModal = (user?: UserProfile) => {
    if (user) {
      setEditingUser(user);
      setFormData({
        name: user.name,
        email: user.email,
        role: user.role,
        status: user.status,
        password: ''
      });
    } else {
      setEditingUser(null);
      setFormData({
        name: '',
        email: '',
        role: 'user',
        status: 'active',
        password: ''
      });
    }
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setEditingUser(null);
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingUser) {
        // Update existing user
        const data: UpdateUserRequest = {
          name: formData.name,
          email: formData.email,
          role: formData.role,
          status: formData.status
        };
        const updatedUser = await updateUser(editingUser.id, data);
        setUsers(prev => prev.map(u => u.id === updatedUser.id ? updatedUser : u));
      } else {
        // Create new user
        const data: CreateUserRequest = {
          name: formData.name,
          email: formData.email,
          password: formData.password,
          role: formData.role
        };
        const newUser = await createUser(data);
        setUsers(prev => [newUser, ...prev]);
      }
      handleCloseModal();
    } catch (error) {
      console.error('Failed to save user', error);
    }
  };

  const handleDelete = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this user? This action cannot be undone.')) {
      try {
        await deleteUser(id);
        setUsers(prev => prev.filter(u => u.id !== id));
      } catch (error) {
        console.error('Failed to delete user', error);
      }
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    loadUsers();
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

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Users</h1>
        <Button variant="primary" onClick={() => handleShowModal()}>
          <i className="fas fa-plus me-2"></i>
          Add User
        </Button>
      </div>

      <Card className="mb-4">
        <Card.Body>
          <Row>
            <Col md={6}>
              <Form onSubmit={handleSearch}>
                <InputGroup>
                  <Form.Control
                    type="text"
                    placeholder="Search users by name or email..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                  />
                  <Button variant="outline-secondary" type="submit">
                    <i className="fas fa-search"></i>
                  </Button>
                </InputGroup>
              </Form>
            </Col>
            <Col md={6}>
              <div className="d-flex justify-content-end">
                <Dropdown className="me-2">
                  <Dropdown.Toggle variant="outline-secondary" id="role-dropdown">
                    Role: {roleFilter === 'all' ? 'All' : getRoleText(roleFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setRoleFilter('all')}>All</Dropdown.Item>
                    <Dropdown.Item onClick={() => setRoleFilter('admin')}>{getRoleText('admin')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setRoleFilter('lawyer')}>{getRoleText('lawyer')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setRoleFilter('user')}>{getRoleText('user')}</Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
                <Dropdown className="me-2">
                  <Dropdown.Toggle variant="outline-secondary" id="status-dropdown">
                    Status: {statusFilter === 'all' ? 'All' : getStatusText(statusFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setStatusFilter('all')}>All</Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter('active')}>{getStatusText('active')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter('inactive')}>{getStatusText('inactive')}</Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
                <Button variant="outline-primary">
                  <i className="fas fa-filter me-2"></i>
                  More Filters
                </Button>
              </div>
            </Col>
          </Row>
        </Card.Body>
      </Card>

      {loading ? (
        <div className="d-flex justify-content-center align-items-center" style={{ height: '400px' }}>
          <Spinner animation="border" />
          <span className="ms-2">Loading users...</span>
        </div>
      ) : (
        <Card>
          <Card.Body>
            <Table striped bordered hover responsive>
              <thead>
                <tr>
                  <th>User</th>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th>Last Login</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map(user => (
                  <tr key={user.id}>
                    <td>
                      <div className="d-flex align-items-center">
                        <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2" style={{ width: '32px', height: '32px' }}>
                          <i className="fas fa-user text-muted"></i>
                        </div>
                        <div>
                          <div className="fw-bold">{user.name}</div>
                          <div className="small text-muted">ID: {user.id}</div>
                        </div>
                      </div>
                    </td>
                    <td>{user.email}</td>
                    <td>
                      <Badge bg={getRoleBadgeClass(user.role)}>
                        {getRoleText(user.role)}
                      </Badge>
                    </td>
                    <td>
                      <Badge bg={getStatusBadgeClass(user.status)}>
                        {getStatusText(user.status)}
                      </Badge>
                    </td>
                    <td>{new Date(user.created_at).toLocaleDateString()}</td>
                    <td>
                      {user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'Never'}
                    </td>
                    <td>
                      <div className="d-flex">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          className="me-2"
                          onClick={() => handleShowModal(user)}
                        >
                          <i className="fas fa-edit"></i>
                        </Button>
                        <Button
                          variant="outline-info"
                          size="sm"
                          className="me-2"
                        >
                          <i className="fas fa-eye"></i>
                        </Button>
                        {user.role !== 'admin' && (
                          <Button
                            variant="outline-danger"
                            size="sm"
                            onClick={() => handleDelete(user.id)}
                          >
                            <i className="fas fa-trash"></i>
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
            
            {users.length === 0 && (
              <div className="text-center py-5">
                <i className="fas fa-user-shield fa-3x text-muted mb-3"></i>
                <h5>No users found</h5>
                <p className="text-muted">Try adjusting your search or filter criteria</p>
                <Button variant="primary" onClick={() => handleShowModal()}>
                  <i className="fas fa-plus me-2"></i>
                  Add Your First User
                </Button>
              </div>
            )}
          </Card.Body>
        </Card>
      )}

      {/* User Modal */}
      <Modal show={showModal} onHide={handleCloseModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            {editingUser ? (
              <span><i className="fas fa-edit me-2"></i> Edit User</span>
            ) : (
              <span><i className="fas fa-user-plus me-2"></i> Add User</span>
            )}
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleSubmit}>
          <Modal.Body>
            <Row>
              <Col md={6}>
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
              <Col md={6}>
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
            {!editingUser && (
              <Row>
                <Col md={12}>
                  <Form.Group className="mb-3">
                    <Form.Label>Password <span className="text-danger">*</span></Form.Label>
                    <Form.Control
                      type="password"
                      name="password"
                      value={formData.password}
                      onChange={handleChange}
                      required={!editingUser}
                      placeholder="Enter password"
                    />
                  </Form.Group>
                </Col>
              </Row>
            )}
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
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={handleCloseModal}>
              <i className="fas fa-times me-2"></i>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              {editingUser ? (
                <span><i className="fas fa-save me-2"></i> Update User</span>
              ) : (
                <span><i className="fas fa-plus me-2"></i> Add User</span>
              )}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default UsersPage;