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
  getNotifications,
  markAsRead,
  deleteNotification,
} from "../services/notificationService";
import { Notification } from "../types";

const NotificationsPage: React.FC = () => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [unreadCount, setUnreadCount] = useState(0);
  const [filter, setFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    loadNotifications();
  }, [currentPage, search, filter]);

  const loadNotifications = async () => {
    setLoading(true);
    try {
      const params = {
        page: currentPage,
        pageSize: 10,
        search: search || undefined,
        filter: filter !== "all" ? filter : undefined,
      };
      const response = await getNotifications(params);
      setNotifications(response.data);
      setUnreadCount(response.data.filter((n) => !n.read).length);
      setTotalPages(
        Math.ceil(response.pagination.total / response.pagination.page_size),
      );
    } catch (error) {
      console.error("Failed to load notifications", error);
    } finally {
      setLoading(false);
    }
  };

  const handleMarkAsRead = async (id: number) => {
    try {
      await markAsRead(id);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, read: true } : n)),
      );
      setUnreadCount((prev) => prev - 1);
    } catch (error) {
      console.error("Failed to mark notification as read", error);
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      // 暂时简单处理，标记所有为已读
      setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
      setUnreadCount(0);
    } catch (error) {
      console.error("Failed to mark all notifications as read", error);
    }
  };

  const handleDelete = async (id: number) => {
    if (
      window.confirm(
        "Are you sure you want to delete this notification? This action cannot be undone.",
      )
    ) {
      try {
        await deleteNotification(id);
        setNotifications((prev) => prev.filter((n) => n.id !== id));
        const deletedNotification = notifications.find((n) => n.id === id);
        if (deletedNotification && !deletedNotification.read) {
          setUnreadCount((prev) => prev - 1);
        }
      } catch (error) {
        console.error("Failed to delete notification", error);
      }
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    loadNotifications();
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case "info":
        return "fas fa-info-circle";
      case "warning":
        return "fas fa-exclamation-triangle";
      case "error":
        return "fas fa-times-circle";
      case "success":
        return "fas fa-check-circle";
      default:
        return "fas fa-bell";
    }
  };

  const getNotificationBadgeClass = (type: string) => {
    switch (type) {
      case "case_update":
        return "bg-primary";
      case "client_update":
        return "bg-success";
      case "deadline":
        return "bg-danger";
      case "document":
        return "bg-info";
      case "system":
        return "bg-warning";
      case "message":
        return "bg-secondary";
      default:
        return "bg-secondary";
    }
  };

  const getNotificationTypeText = (type: string) => {
    switch (type) {
      case "case_update":
        return "Case Update";
      case "client_update":
        return "Client Update";
      case "deadline":
        return "Deadline";
      case "document":
        return "Document";
      case "system":
        return "System";
      case "message":
        return "Message";
      default:
        return type;
    }
  };

  const filterNotifications = (
    notifications: Notification[],
    filter: string,
    search: string,
  ) => {
    let filtered = notifications;

    if (filter !== "all") {
      filtered = filtered.filter((n) => n.type === filter);
    }

    if (search) {
      filtered = filtered.filter(
        (n) =>
          n.title.toLowerCase().includes(search.toLowerCase()) ||
          n.message.toLowerCase().includes(search.toLowerCase()),
      );
    }

    return filtered;
  };

  const filteredNotifications = filterNotifications(
    notifications,
    filter,
    search,
  );

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Notifications</h1>
        <div className="d-flex">
          <Button
            variant="outline-primary"
            className="me-2"
            onClick={handleMarkAllAsRead}
            disabled={unreadCount === 0}
          >
            <i className="fas fa-check-circle me-2"></i>
            Mark All as Read
          </Button>
          <Dropdown>
            <Dropdown.Toggle variant="outline-secondary" id="filter-dropdown">
              <i className="fas fa-filter me-2"></i>
              Filter:{" "}
              {filter === "all" ? "All" : getNotificationTypeText(filter)}
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item onClick={() => setFilter("all")}>
                All ({notifications.length})
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setFilter("case_update")}>
                {getNotificationTypeText("case_update")} (
                {notifications.filter((n) => n.type === "case_update").length})
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setFilter("client_update")}>
                {getNotificationTypeText("client_update")} (
                {notifications.filter((n) => n.type === "client_update").length}
                )
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setFilter("deadline")}>
                {getNotificationTypeText("deadline")} (
                {notifications.filter((n) => n.type === "deadline").length})
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setFilter("document")}>
                {getNotificationTypeText("document")} (
                {notifications.filter((n) => n.type === "document").length})
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setFilter("system")}>
                {getNotificationTypeText("system")} (
                {notifications.filter((n) => n.type === "system").length})
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setFilter("message")}>
                {getNotificationTypeText("message")} (
                {notifications.filter((n) => n.type === "message").length})
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
        </div>
      </div>

      <Card className="mb-4">
        <Card.Body>
          <Form onSubmit={handleSearch}>
            <InputGroup>
              <Form.Control
                type="text"
                placeholder="Search notifications..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
              <Button variant="outline-secondary" type="submit">
                <i className="fas fa-search"></i>
              </Button>
            </InputGroup>
          </Form>
        </Card.Body>
      </Card>

      <Row>
        <Col md={3}>
          <Card className="mb-4">
            <Card.Header>
              <i className="fas fa-bell me-2"></i>
              Notification Summary
            </Card.Header>
            <Card.Body>
              <div className="notification-summary">
                <div className="summary-item mb-3 p-3 border rounded">
                  <div className="d-flex justify-content-between align-items-center">
                    <div>
                      <div className="fw-bold">Unread</div>
                      <div className="small text-muted">
                        Notifications requiring attention
                      </div>
                    </div>
                    <Badge bg="danger" className="fs-5">
                      {unreadCount}
                    </Badge>
                  </div>
                </div>
                <div className="summary-item mb-3 p-3 border rounded">
                  <div className="d-flex justify-content-between align-items-center">
                    <div>
                      <div className="fw-bold">Total</div>
                      <div className="small text-muted">All notifications</div>
                    </div>
                    <Badge bg="secondary" className="fs-5">
                      {notifications.length}
                    </Badge>
                  </div>
                </div>
                <div className="summary-item p-3 border rounded">
                  <div className="d-flex justify-content-between align-items-center">
                    <div>
                      <div className="fw-bold">Read</div>
                      <div className="small text-muted">
                        Already reviewed notifications
                      </div>
                    </div>
                    <Badge bg="success" className="fs-5">
                      {notifications.length - unreadCount}
                    </Badge>
                  </div>
                </div>
              </div>
            </Card.Body>
          </Card>

          <Card>
            <Card.Header>
              <i className="fas fa-chart-bar me-2"></i>
              Notification Trends
            </Card.Header>
            <Card.Body>
              <div
                className="chart-placeholder"
                style={{
                  height: "200px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                <div className="text-center">
                  <i className="fas fa-chart-line fa-2x text-muted mb-2"></i>
                  <p className="text-muted small">Notification trends chart</p>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>

        <Col md={9}>
          {loading ? (
            <div
              className="d-flex justify-content-center align-items-center"
              style={{ height: "400px" }}
            >
              <Spinner animation="border" />
              <span className="ms-2">Loading notifications...</span>
            </div>
          ) : (
            <Card>
              <Card.Body>
                <ListGroup variant="flush">
                  {filteredNotifications.length > 0 ? (
                    filteredNotifications.map((notification) => (
                      <ListGroup.Item
                        key={notification.id}
                        className={`border-bottom ${!notification.read ? "bg-light" : ""}`}
                      >
                        <div className="d-flex">
                          <div
                            className={`notification-icon rounded-circle d-flex align-items-center justify-content-center me-3 ${!notification.read ? "bg-primary" : "bg-light"}`}
                            style={{ width: "40px", height: "40px" }}
                          >
                            <i
                              className={`${getNotificationIcon(notification.type)} ${!notification.read ? "text-white" : "text-muted"}`}
                            ></i>
                          </div>
                          <div className="flex-grow-1">
                            <div className="d-flex justify-content-between">
                              <div>
                                <h6 className="mb-1">
                                  {notification.title}
                                  {!notification.read && (
                                    <Badge bg="danger" className="ms-2">
                                      New
                                    </Badge>
                                  )}
                                </h6>
                                <p className="mb-1 text-muted">
                                  {notification.message}
                                </p>
                              </div>
                              <div className="d-flex">
                                <Button
                                  variant="outline-primary"
                                  size="sm"
                                  className="me-2"
                                  onClick={() =>
                                    handleMarkAsRead(notification.id)
                                  }
                                  disabled={notification.read}
                                >
                                  <i className="fas fa-check"></i>
                                </Button>
                                <Button
                                  variant="outline-info"
                                  size="sm"
                                  className="me-2"
                                >
                                  <i className="fas fa-eye"></i>
                                </Button>
                                <Button
                                  variant="outline-danger"
                                  size="sm"
                                  onClick={() => handleDelete(notification.id)}
                                >
                                  <i className="fas fa-trash"></i>
                                </Button>
                              </div>
                            </div>
                            <div className="d-flex justify-content-between align-items-center mt-2">
                              <div>
                                <Badge
                                  variant={getNotificationBadgeClass(
                                    notification.type,
                                  )}
                                  className="me-2"
                                >
                                  {getNotificationTypeText(notification.type)}
                                </Badge>
                                <small className="text-muted">
                                  <i className="fas fa-clock me-1"></i>
                                  {new Date(
                                    notification.created_at,
                                  ).toLocaleString()}
                                </small>
                              </div>
                              <div>
                                <small className="text-muted">
                                  {notification.relatedEntity && (
                                    <span>
                                      <i className="fas fa-link me-1"></i>
                                      {notification.relatedEntity.type}: #
                                      {notification.relatedEntity.id}
                                    </span>
                                  )}
                                </small>
                              </div>
                            </div>
                          </div>
                        </div>
                      </ListGroup.Item>
                    ))
                  ) : (
                    <div className="text-center py-5">
                      <i className="fas fa-bell-slash fa-3x text-muted mb-3"></i>
                      <h5>No notifications found</h5>
                      <p className="text-muted">
                        {search || filter !== "all"
                          ? "Try adjusting your search or filter criteria"
                          : "You have no notifications at this time"}
                      </p>
                      {search ||
                        (filter !== "all" && (
                          <Button
                            variant="primary"
                            onClick={() => {
                              setSearch("");
                              setFilter("all");
                            }}
                          >
                            <i className="fas fa-sync me-2"></i>
                            Reset Filters
                          </Button>
                        ))}
                    </div>
                  )}
                </ListGroup>
              </Card.Body>
            </Card>
          )}
        </Col>
      </Row>
    </div>
  );
};

export default NotificationsPage;
