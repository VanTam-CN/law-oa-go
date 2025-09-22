import React, { useState, useEffect } from "react";
import {
  Row,
  Col,
  Card,
  Button,
  Form,
  InputGroup,
  Modal,
  Spinner,
  Dropdown,
  Badge,
  ListGroup,
} from "react-bootstrap";
import {
  getMessages,
  sendMessage,
  markAsRead,
  deleteMessage,
} from "../services/messageService";
import { Message, MessageListRequest, SendMessageRequest } from "../types";

const MessagesPage: React.FC = () => {
  const getMessageIcon = (type: string) => {
    switch (type) {
      case "inbox":
        return "fas fa-inbox";
      case "sent":
        return "fas fa-paper-plane";
      case "draft":
        return "fas fa-file-alt";
      default:
        return "fas fa-envelope";
    }
  };
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [showComposeModal, setShowComposeModal] = useState(false);
  const [replyingTo, setReplyingTo] = useState<Message | null>(null);
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<"all" | "inbox" | "sent" | "draft">("all");
  const [statusFilter, setStatusFilter] = useState<"all" | "read" | "unread">("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [formData, setFormData] = useState({
    recipientId: 0,
    subject: "",
    content: "",
    attachments: [] as File[],
  });

  useEffect(() => {
    loadMessages();
  }, [currentPage, search, typeFilter, statusFilter]);

  const loadMessages = async () => {
    setLoading(true);
    try {
      const params: MessageListRequest = {
        page: currentPage,
        page_size: 10,
        search: search || undefined,
        type: typeFilter !== "all" ? typeFilter : undefined,
        status: statusFilter !== "all" ? statusFilter : undefined,
      };
      const response = await getMessages(params);
      setMessages(response.data);
      setTotalPages(
        Math.ceil(response.pagination.total / response.pagination.page_size),
      );
    } catch (error) {
      console.error("Failed to load messages", error);
    } finally {
      setLoading(false);
    }
  };

  const handleShowComposeModal = (message?: Message) => {
    if (message) {
      setReplyingTo(message);
      setFormData({
        recipientId: message.sender_id,
        subject: `Re: ${message.subject}`,
        content: `

--- Original Message ---
From: ${message.sender_name}
Sent: ${new Date(message.created_at).toLocaleString()}
Subject: ${message.subject}

${message.content}`,
        attachments: [],
      });
    } else {
      setReplyingTo(null);
      setFormData({
        recipientId: 0,
        subject: "",
        content: "",
        attachments: [],
      });
    }
    setShowComposeModal(true);
  };

  const handleCloseComposeModal = () => {
    setShowComposeModal(false);
    setReplyingTo(null);
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      setFormData((prev) => ({
        ...prev,
        attachments: [...prev.attachments, ...files],
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const data: SendMessageRequest = {
        recipient_id: formData.recipientId,
        subject: formData.subject,
        content: formData.content,
      };
      const newMessage = await sendMessage(data);
      setMessages((prev) => [newMessage, ...prev]);
      handleCloseComposeModal();
    } catch (error) {
      console.error("Failed to send message", error);
    }
  };

  const handleMarkAsRead = async (id: number) => {
    try {
      await markAsRead(id);
      setMessages((prev) =>
        prev.map((m) => (m.id === id ? { ...m, read: true } : m)),
      );
    } catch (error) {
      console.error("Failed to mark message as read", error);
    }
  };

  const handleDelete = async (id: number) => {
    if (
      window.confirm(
        "Are you sure you want to delete this message? This action cannot be undone.",
      )
    ) {
      try {
        await deleteMessage(id);
        setMessages((prev) => prev.filter((m) => m.id !== id));
      } catch (error) {
        console.error("Failed to delete message", error);
      }
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    loadMessages();
  };

  const getMessageTypeBadgeClass = (type: string) => {
    switch (type) {
      case "email":
        return "bg-primary";
      case "sms":
        return "bg-success";
      case "notification":
        return "bg-warning";
      case "alert":
        return "bg-danger";
      default:
        return "bg-secondary";
    }
  };

  const getMessageStatusBadgeClass = (status: string) => {
    switch (status) {
      case "sent":
        return "bg-success";
      case "received":
        return "bg-primary";
      case "draft":
        return "bg-warning";
      case "archived":
        return "bg-secondary";
      default:
        return "bg-secondary";
    }
  };

  // Get message type display text
  const getMessageTypeText = (type: string) => {
    switch (type) {
      case "inbox":
        return "Inbox";
      case "sent":
        return "Sent";
      case "draft":
        return "Draft";
      default:
        return type;
    }
  };

  // Get message status display text
  const getMessageStatusText = (status: string) => {
    switch (status) {
      case "read":
        return "Read";
      case "unread":
        return "Unread";
      default:
        return status;
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Messages</h1>
        <Button variant="primary" onClick={() => handleShowComposeModal()}>
          <i className="fas fa-paper-plane me-2"></i>
          Compose Message
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
                    placeholder="Search messages by subject or content..."
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
                  <Dropdown.Toggle
                    variant="outline-secondary"
                    id="type-dropdown"
                  >
                    Type:{" "}
                    {typeFilter === "all"
                      ? "All"
                      : getMessageTypeText(typeFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setTypeFilter("all")}>
                      All
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("inbox")}>
                      {getMessageTypeText("inbox")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("sent")}>
                      {getMessageTypeText("sent")}
                    </Dropdown.Item>
                    <Dropdown.Item
                      onClick={() => setTypeFilter("draft")}
                    >
                      {getMessageTypeText("draft")}
                    </Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
                <Dropdown className="me-2">
                  <Dropdown.Toggle
                    variant="outline-secondary"
                    id="status-dropdown"
                  >
                    Status:{" "}
                    {statusFilter === "all"
                      ? "All"
                      : getMessageStatusText(statusFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setStatusFilter("all")}>
                      All
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter("read")}>
                      {getMessageStatusText("read")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter("unread")}>
                      {getMessageStatusText("unread")}
                    </Dropdown.Item>
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
        <div
          className="d-flex justify-content-center align-items-center"
          style={{ height: "400px" }}
        >
          <Spinner animation="border" />
          <span className="ms-2">Loading messages...</span>
        </div>
      ) : (
        <Card>
          <Card.Body>
            <ListGroup variant="flush">
              {messages.map((message) => (
                <ListGroup.Item
                  key={message.id}
                  className={`border-bottom ${!message.read ? "bg-light" : ""}`}
                >
                  <div className="d-flex">
                    <div
                      className={`message-icon rounded-circle d-flex align-items-center justify-content-center me-3 ${!message.read ? "bg-primary" : "bg-light"}`}
                      style={{ width: "40px", height: "40px" }}
                    >
                      <i
                        className={`${getMessageIcon(message.type)} ${!message.read ? "text-white" : "text-muted"}`}
                      ></i>
                    </div>
                    <div className="flex-grow-1">
                      <div className="d-flex justify-content-between">
                        <div>
                          <h6 className="mb-1">
                            {message.subject}
                            {!message.read && (
                              <Badge bg="danger" className="ms-2">
                                New
                              </Badge>
                            )}
                          </h6>
                          <p className="mb-1 text-muted">
                            {message.content.substring(0, 100)}...
                          </p>
                        </div>
                        <div className="d-flex">
                          <Button
                            variant="outline-primary"
                            size="sm"
                            className="me-2"
                            onClick={() => handleMarkAsRead(message.id)}
                            disabled={message.read}
                          >
                            <i className="fas fa-check"></i>
                          </Button>
                          <Button
                            variant="outline-info"
                            size="sm"
                            className="me-2"
                            onClick={() => handleShowComposeModal(message)}
                          >
                            <i className="fas fa-reply"></i>
                          </Button>
                          <Button
                            variant="outline-danger"
                            size="sm"
                            onClick={() => handleDelete(message.id)}
                          >
                            <i className="fas fa-trash"></i>
                          </Button>
                        </div>
                      </div>
                      <div className="d-flex justify-content-between align-items-center mt-2">
                        <div>
                          <Badge
                            variant={getMessageTypeBadgeClass(message.type)}
                            className="me-2"
                          >
                            {getMessageTypeText(message.type)}
                          </Badge>
                          <Badge
                            variant={message.read ? "bg-secondary" : "bg-primary"}
                            className="me-2"
                          >
                            {message.read ? "Read" : "Unread"}
                          </Badge>
                        </div>
                        <div>
                          <small className="text-muted">
                            <i className="fas fa-clock me-1"></i>
                            {new Date(message.created_at).toLocaleString()}
                          </small>
                        </div>
                      </div>
                    </div>
                  </div>
                </ListGroup.Item>
              ))}
            </ListGroup>

            {messages.length === 0 && (
              <div className="text-center py-5">
                <i className="fas fa-envelope-open-text fa-3x text-muted mb-3"></i>
                <h5>No messages found</h5>
                <p className="text-muted">
                  {search || typeFilter !== "all" || statusFilter !== "all"
                    ? "Try adjusting your search or filter criteria"
                    : "You have no messages at this time"}
                </p>
                <Button
                  variant="primary"
                  onClick={() => handleShowComposeModal()}
                >
                  <i className="fas fa-paper-plane me-2"></i>
                  Compose Your First Message
                </Button>
              </div>
            )}
          </Card.Body>
        </Card>
      )}

      {/* Compose Message Modal */}
      <Modal show={showComposeModal} onHide={handleCloseComposeModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            {replyingTo ? (
              <span>
                <i className="fas fa-reply me-2"></i> Reply to Message
              </span>
            ) : (
              <span>
                <i className="fas fa-paper-plane me-2"></i> Compose Message
              </span>
            )}
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleSubmit}>
          <Modal.Body>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>
                    To <span className="text-danger">*</span>
                  </Form.Label>
                  <Form.Control
                    type="text"
                    name="recipient"
                    value={
                      formData.recipientId
                        ? `User #${formData.recipientId}`
                        : ""
                    }
                    onChange={handleChange}
                    required
                    placeholder="Enter recipient name or email"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>
                    Subject <span className="text-danger">*</span>
                  </Form.Label>
                  <Form.Control
                    type="text"
                    name="subject"
                    value={formData.subject}
                    onChange={handleChange}
                    required
                    placeholder="Enter message subject"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>
                    Message <span className="text-danger">*</span>
                  </Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={6}
                    name="content"
                    value={formData.content}
                    onChange={handleChange}
                    required
                    placeholder="Enter your message content"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>Attachments</Form.Label>
                  <Form.Control
                    type="file"
                    multiple
                    onChange={handleFileChange}
                  />
                  {formData.attachments.length > 0 && (
                    <div className="mt-2">
                      <small className="text-muted">
                        {formData.attachments.length} file(s) selected
                      </small>
                    </div>
                  )}
                </Form.Group>
              </Col>
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={handleCloseComposeModal}>
              <i className="fas fa-times me-2"></i>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              <i className="fas fa-paper-plane me-2"></i>
              Send Message
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default MessagesPage;
