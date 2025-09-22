import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Row,
  Col,
  Card,
  Button,
  Spinner,
  Badge,
  Tabs, Tab,
  Tab,
  Dropdown,
  Form,
  InputGroup,
} from "react-bootstrap";
import {
  getClient,
  updateClient,
  deleteClient,
} from "../services/clientService";
import { Client, UpdateClientRequest } from "../types";

const ClientDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [client, setClient] = useState<Client | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("overview");
  const [editing, setEditing] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    phone: "",
    address: "",
    company: "",
    status: "active",
  });

  useEffect(() => {
    if (id) {
      loadClient(parseInt(id));
    }
  }, [id]);

  const loadClient = async (client_id: number) => {
    setLoading(true);
    try {
      const response = await getClient(client_id);
      setClient(response);
      setFormData({
        name: response.name,
        email: response.email,
        phone: response.phone,
        address: response.address,
        company: response.company,
        status: response.status,
      });
    } catch (error) {
      console.error("Failed to load client", error);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = () => {
    setEditing(true);
  };

  const handleCancelEdit = () => {
    setEditing(false);
    if (client) {
      setFormData({
        name: client.name,
        email: client.email,
        phone: client.phone,
        address: client.address,
        company: client.company,
        status: client.status,
      });
    }
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!client) return;

    try {
      const data: UpdateClientRequest = formData;
      const updatedClient = await updateClient(client.id, data);
      setClient(updatedClient);
      setEditing(false);
    } catch (error) {
      console.error("Failed to update client", error);
    }
  };

  const handleDelete = async () => {
    if (!client) return;

    if (
      window.confirm(
        "Are you sure you want to delete this client? This action cannot be undone.",
      )
    ) {
      try {
        await deleteClient(client.id);
        navigate("/clients");
      } catch (error) {
        console.error("Failed to delete client", error);
      }
    }
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case "active":
        return "bg-success";
      case "inactive":
        return "bg-secondary";
      default:
        return "bg-secondary";
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

  if (loading) {
    return (
      <div
        className="d-flex justify-content-center align-items-center"
        style={{ height: "50vh" }}
      >
        <Spinner animation="border" />
        <span className="ms-2">Loading client details...</span>
      </div>
    );
  }

  if (!client) {
    return (
      <div className="text-center py-5">
        <i className="fas fa-exclamation-triangle fa-3x text-warning mb-3"></i>
        <h5>Client not found</h5>
        <p className="text-muted">The requested client could not be found</p>
        <Button variant="primary" onClick={() => navigate("/clients")}>
          <i className="fas fa-arrow-left me-2"></i>
          Back to Clients
        </Button>
      </div>
    );
  }

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <Button
            variant="outline-secondary"
            onClick={() => navigate("/clients")}
            className="mb-2"
          >
            <i className="fas fa-arrow-left me-2"></i>
            Back to Clients
          </Button>
          <h1>Client: {client.name}</h1>
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
                Edit Client
              </Dropdown.Item>
              <Dropdown.Item>
                <i className="fas fa-copy me-2"></i>
                Duplicate Client
              </Dropdown.Item>
              <Dropdown.Item>
                <i className="fas fa-history me-2"></i>
                View History
              </Dropdown.Item>
              <Dropdown.Divider />
              <Dropdown.Item className="text-danger" onClick={handleDelete}>
                <i className="fas fa-trash me-2"></i>
                Delete Client
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          <Button variant="primary" onClick={handleEdit}>
            <i className="fas fa-edit me-2"></i>
            Edit Client
          </Button>
        </div>
      </div>

      <Row>
        <Col md={8}>
          <Card className="mb-4">
            <Card.Header className="d-flex justify-content-between align-items-center">
              <span>Client Overview</span>
              <div>
                <Badge bg={getStatusBadgeClass(client.status)} className="me-2">
                  {getStatusText(client.status)}
                </Badge>
              </div>
            </Card.Header>
            <Card.Body>
              {editing ? (
                <Form onSubmit={handleSubmit}>
                  <Row>
                    <Col md={12}>
                      <Form.Group className="mb-3">
                        <Form.Label>
                          Name <span className="text-danger">*</span>
                        </Form.Label>
                        <Form.Control
                          type="text"
                          name="name"
                          value={formData.name}
                          onChange={handleChange}
                          required
                          placeholder="Enter client name"
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
                          name="email"
                          value={formData.email}
                          onChange={handleChange}
                          required
                          placeholder="Enter client email"
                        />
                      </Form.Group>
                    </Col>
                  </Row>
                  <Row>
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>Phone</Form.Label>
                        <Form.Control
                          type="text"
                          name="phone"
                          value={formData.phone}
                          onChange={handleChange}
                          placeholder="Enter phone number"
                        />
                      </Form.Group>
                    </Col>
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>Company</Form.Label>
                        <Form.Control
                          type="text"
                          name="company"
                          value={formData.company}
                          onChange={handleChange}
                          placeholder="Enter company name"
                        />
                      </Form.Group>
                    </Col>
                  </Row>
                  <Row>
                    <Col md={12}>
                      <Form.Group className="mb-3">
                        <Form.Label>Address</Form.Label>
                        <Form.Control
                          as="textarea"
                          rows={3}
                          name="address"
                          value={formData.address}
                          onChange={handleChange}
                          placeholder="Enter client address"
                        />
                      </Form.Group>
                    </Col>
                  </Row>
                  <Row>
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>Status</Form.Label>
                        <Form.Select
                          name="status"
                          value={formData.status}
                          onChange={handleChange}
                        >
                          <option value="active">
                            {getStatusText("active")}
                          </option>
                          <option value="inactive">
                            {getStatusText("inactive")}
                          </option>
                        </Form.Select>
                      </Form.Group>
                    </Col>
                  </Row>
                  <div className="d-flex justify-content-end">
                    <Button
                      variant="secondary"
                      onClick={handleCancelEdit}
                      className="me-2"
                    >
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
                      <div className="fw-bold">{client.name}</div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Email</small>
                      <div className="fw-bold">{client.email}</div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Phone</small>
                      <div className="fw-bold">{client.phone}</div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Company</small>
                      <div className="fw-bold">{client.company}</div>
                    </div>
                  </Col>
                  <Col md={6}>
                    <div className="mb-3">
                      <small className="text-muted">Address</small>
                      <div className="fw-bold">{client.address}</div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Created</small>
                      <div className="fw-bold">
                        {new Date(client.created_at).toLocaleDateString()} at{" "}
                        {new Date(client.created_at).toLocaleTimeString()}
                      </div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Last Updated</small>
                      <div className="fw-bold">
                        {client.updated_at
                          ? new Date(client.updated_at).toLocaleDateString()
                          : "Never"}
                      </div>
                    </div>
                    <div className="mb-3">
                      <small className="text-muted">Client ID</small>
                      <div className="fw-bold">#{client.id}</div>
                    </div>
                  </Col>
                </Row>
              )}
            </Card.Body>
          </Card>

          <Tabs
            activeKey={activeTab}
            onSelect={(k) => setActiveTab(k || "overview")}
            className="mb-4"
          >
            <Tab eventKey="overview" title="Overview">
              <Card>
                <Card.Body>
                  <h5>Client Timeline</h5>
                  <div className="timeline">
                    <div className="timeline-item mb-4">
                      <div className="d-flex">
                        <div className="timeline-icon bg-primary rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-user-plus text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">Client Created</h6>
                          <p className="mb-0 text-muted">
                            Client was added to the system
                          </p>
                          <small className="text-muted">
                            {new Date(client.created_at).toLocaleString()}
                          </small>
                        </div>
                      </div>
                    </div>
                    <div className="timeline-item mb-4">
                      <div className="d-flex">
                        <div className="timeline-icon bg-success rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-gavel text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">First Case Created</h6>
                          <p className="mb-0 text-muted">
                            Case #1234 was created for this client
                          </p>
                          <small className="text-muted">2 days ago</small>
                        </div>
                      </div>
                    </div>
                    <div className="timeline-item">
                      <div className="d-flex">
                        <div className="timeline-icon bg-warning rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-clock text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">Contact Information Updated</h6>
                          <p className="mb-0 text-muted">
                            Client's contact information was updated
                          </p>
                          <small className="text-muted">1 day ago</small>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="cases" title="Cases">
              <Card>
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-center mb-4">
                    <h5>Client Cases</h5>
                    <Button variant="primary">
                      <i className="fas fa-plus me-2"></i>
                      Create New Case
                    </Button>
                  </div>

                  <div className="case-list">
                    <div className="case-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div>
                          <h6 className="mb-1">
                            Case #1234 - Contract Dispute
                          </h6>
                          <p className="mb-1 text-muted">
                            Civil case involving contract disputes between
                            parties
                          </p>
                          <div className="d-flex">
                            <Badge bg="primary" className="me-2">
                              Civil
                            </Badge>
                            <Badge bg="warning">Medium</Badge>
                          </div>
                        </div>
                        <div className="text-end">
                          <div className="fw-bold">Active</div>
                          <small className="text-muted">
                            Assigned to: Jane Smith
                          </small>
                        </div>
                      </div>
                      <div className="d-flex justify-content-between mt-2">
                        <small className="text-muted">
                          <i className="fas fa-calendar me-1"></i>
                          Created: 2 days ago
                        </small>
                        <Button variant="outline-primary" size="sm">
                          <i className="fas fa-eye me-1"></i>
                          View Case
                        </Button>
                      </div>
                    </div>

                    <div className="case-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div>
                          <h6 className="mb-1">
                            Case #5678 - Property Dispute
                          </h6>
                          <p className="mb-1 text-muted">
                            Dispute over property ownership rights
                          </p>
                          <div className="d-flex">
                            <Badge bg="primary" className="me-2">
                              Civil
                            </Badge>
                            <Badge bg="danger">High</Badge>
                          </div>
                        </div>
                        <div className="text-end">
                          <div className="fw-bold">Pending</div>
                          <small className="text-muted">Unassigned</small>
                        </div>
                      </div>
                      <div className="d-flex justify-content-between mt-2">
                        <small className="text-muted">
                          <i className="fas fa-calendar me-1"></i>
                          Created: 1 week ago
                        </small>
                        <Button variant="outline-primary" size="sm">
                          <i className="fas fa-eye me-1"></i>
                          View Case
                        </Button>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="documents" title="Documents">
              <Card>
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-center mb-4">
                    <h5>Client Documents</h5>
                    <Button variant="primary">
                      <i className="fas fa-plus me-2"></i>
                      Upload Document
                    </Button>
                  </div>

                  <div className="document-list">
                    <div className="document-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div className="d-flex align-items-center">
                          <div className="bg-light rounded me-3 p-2">
                            <i className="fas fa-file-pdf fa-2x text-danger"></i>
                          </div>
                          <div>
                            <h6 className="mb-1">
                              Identification Document.pdf
                            </h6>
                            <small className="text-muted">
                              Uploaded 2 days ago
                            </small>
                          </div>
                        </div>
                        <div>
                          <Button
                            variant="outline-primary"
                            size="sm"
                            className="me-2"
                          >
                            <i className="fas fa-download"></i>
                          </Button>
                          <Button variant="outline-danger" size="sm">
                            <i className="fas fa-trash"></i>
                          </Button>
                        </div>
                      </div>
                    </div>

                    <div className="document-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div className="d-flex align-items-center">
                          <div className="bg-light rounded me-3 p-2">
                            <i className="fas fa-file-word fa-2x text-primary"></i>
                          </div>
                          <div>
                            <h6 className="mb-1">Contract Agreement.docx</h6>
                            <small className="text-muted">
                              Uploaded 1 day ago
                            </small>
                          </div>
                        </div>
                        <div>
                          <Button
                            variant="outline-primary"
                            size="sm"
                            className="me-2"
                          >
                            <i className="fas fa-download"></i>
                          </Button>
                          <Button variant="outline-danger" size="sm">
                            <i className="fas fa-trash"></i>
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="communications" title="Communications">
              <Card>
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-center mb-4">
                    <h5>Client Communications</h5>
                    <Button variant="primary">
                      <i className="fas fa-paper-plane me-2"></i>
                      Send Message
                    </Button>
                  </div>

                  <div className="communication-list">
                    <div className="communication-item mb-4 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center mb-2">
                        <div className="fw-bold">Welcome Email</div>
                        <small className="text-muted">2 days ago</small>
                      </div>
                      <p className="mb-2">
                        Welcome to our legal practice management system. You can
                        now access your case information and communicate with
                        your legal team through this portal.
                      </p>
                      <div className="d-flex justify-content-between">
                        <small className="text-muted">
                          <i className="fas fa-envelope me-1"></i>
                          Sent via email
                        </small>
                        <Button variant="outline-primary" size="sm">
                          <i className="fas fa-reply me-1"></i>
                          Reply
                        </Button>
                      </div>
                    </div>

                    <div className="communication-item mb-4 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center mb-2">
                        <div className="fw-bold">Case Update Notification</div>
                        <small className="text-muted">1 day ago</small>
                      </div>
                      <p className="mb-2">
                        Your case #1234 has been assigned to our legal team. We
                        will keep you updated on all developments.
                      </p>
                      <div className="d-flex justify-content-between">
                        <small className="text-muted">
                          <i className="fas fa-envelope me-1"></i>
                          Sent via email
                        </small>
                        <Button variant="outline-primary" size="sm">
                          <i className="fas fa-reply me-1"></i>
                          Reply
                        </Button>
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
            <Card.Header>Contact Information</Card.Header>
            <Card.Body>
              <div className="d-flex align-items-center mb-4">
                <div
                  className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3"
                  style={{ width: "64px", height: "64px" }}
                >
                  <i className="fas fa-user fa-2x text-muted"></i>
                </div>
                <div>
                  <h5 className="mb-1">{client.name}</h5>
                  <p className="text-muted mb-0">Client ID: #{client.id}</p>
                  <Badge
                    variant={getStatusBadgeClass(client.status)}
                    className="mt-1"
                  >
                    {getStatusText(client.status)}
                  </Badge>
                </div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Email</small>
                <div className="fw-bold">{client.email}</div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Phone</small>
                <div className="fw-bold">{client.phone}</div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Address</small>
                <div className="fw-bold">{client.address}</div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Company</small>
                <div className="fw-bold">{client.company}</div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Created</small>
                <div className="fw-bold">
                  {new Date(client.created_at).toLocaleDateString()}
                </div>
              </div>

              <div className="d-grid gap-2">
                <Button variant="primary">
                  <i className="fas fa-envelope me-2"></i>
                  Send Email
                </Button>
                <Button variant="outline-primary">
                  <i className="fas fa-phone me-2"></i>
                  Call Client
                </Button>
              </div>
            </Card.Body>
          </Card>

          <Card className="mb-4">
            <Card.Header>Client Statistics</Card.Header>
            <Card.Body>
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Total Cases</span>
                  <span className="fw-bold">2</span>
                </div>
                <div className="progress mt-1">
                  <div
                    className="progress-bar bg-primary"
                    role="progressbar"
                    style={{ width: "20%" }}
                  ></div>
                </div>
              </div>

              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Active Cases</span>
                  <span className="fw-bold">1</span>
                </div>
                <div className="progress mt-1">
                  <div
                    className="progress-bar bg-success"
                    role="progressbar"
                    style={{ width: "10%" }}
                  ></div>
                </div>
              </div>

              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Pending Cases</span>
                  <span className="fw-bold">1</span>
                </div>
                <div className="progress mt-1">
                  <div
                    className="progress-bar bg-warning"
                    role="progressbar"
                    style={{ width: "10%" }}
                  ></div>
                </div>
              </div>

              <div className="stat-item">
                <div className="d-flex justify-content-between">
                  <span>Closed Cases</span>
                  <span className="fw-bold">0</span>
                </div>
                <div className="progress mt-1">
                  <div
                    className="progress-bar bg-secondary"
                    role="progressbar"
                    style={{ width: "0%" }}
                  ></div>
                </div>
              </div>
            </Card.Body>
          </Card>

          <Card>
            <Card.Header>Quick Actions</Card.Header>
            <Card.Body>
              <div className="d-grid gap-2">
                <Button variant="outline-primary">
                  <i className="fas fa-gavel me-2"></i>
                  Create New Case
                </Button>
                <Button variant="outline-success">
                  <i className="fas fa-file-contract me-2"></i>
                  Upload Document
                </Button>
                <Button variant="outline-info">
                  <i className="fas fa-paper-plane me-2"></i>
                  Send Message
                </Button>
                <Button variant="outline-warning">
                  <i className="fas fa-calendar-plus me-2"></i>
                  Schedule Meeting
                </Button>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ClientDetailPage;
