import React, { useState } from "react";
import {
  Navbar,
  Nav,
  NavDropdown,
  Container,
  Badge,
  Button,
} from "react-bootstrap";
import { LinkContainer } from "react-router-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import { RootState } from "../../store";
import { logout } from "../../store/slices/authSlice";
import UserAvatar from "../ui/UserAvatar";
import NotificationBadge from "../ui/NotificationBadge";
import SearchBar from "../common/SearchBar";
import "./Header.css";

const Header: React.FC = () => {
  const { user } = useSelector((state: RootState) => state.auth);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [notifications, setNotifications] = useState(3);
  const [messages, setMessages] = useState(5);

  const handleSearch = (query: string, config: any) => {
    console.log("Searching for:", query, config);
    navigate(`/search?q=${encodeURIComponent(query)}`);
  };

  const handleLogout = () => {
    dispatch(logout() as any);
    navigate("/login");
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

  return (
    <Navbar variant="dark" expand="lg" fixed="top" className="header">
      <Container fluid>
        <LinkContainer to="/dashboard">
          <Navbar.Brand className="d-flex align-items-center">
            <div
              className="bg-primary rounded-circle d-flex align-items-center justify-content-center me-2"
              style={{ width: "32px", height: "32px" }}
            >
              <i className="fas fa-balance-scale text-white"></i>
            </div>
            <span className="d-none d-md-block">Law OA System</span>
          </Navbar.Brand>
        </LinkContainer>
        <Navbar.Toggle aria-controls="basic-navbar-nav" />
        <Navbar.Collapse id="basic-navbar-nav">
          <Nav className="me-auto">
            <SearchBar
              onSubmit={handleSearch}
              placeholder="Search clients, cases, documents..."
              enableNavigation={true}
              showAdvancedToggle={true}
              size="sm"
              variant="outline-secondary"
            />
          </Nav>
          <Nav className="d-flex align-items-center">
            <Nav.Link href="/notifications" className="position-relative me-3">
              <i className="fas fa-bell fa-lg"></i>
              <NotificationBadge count={notifications} />
            </Nav.Link>
            <Nav.Link href="/messages" className="position-relative me-3">
              <i className="fas fa-envelope fa-lg"></i>
              <NotificationBadge count={messages} variant="primary" />
            </Nav.Link>
            <NavDropdown
              title={
                <div className="d-flex align-items-center">
                  <UserAvatar
                    name={user?.name || "User"}
                    size="sm"
                    className="me-2"
                  />
                  <span className="d-none d-md-block">{user?.name}</span>
                  <Badge
                    bg={getRoleBadgeClass(user?.role || "user")}
                    className="ms-2 d-none d-md-block"
                  >
                    {getRoleText(user?.role || "user")}
                  </Badge>
                </div>
              }
              id="user-dropdown"
              align="end"
            >
              <LinkContainer to="/profile">
                <NavDropdown.Item>
                  <i className="fas fa-user me-2"></i>
                  My Profile
                </NavDropdown.Item>
              </LinkContainer>
              <LinkContainer to="/settings">
                <NavDropdown.Item>
                  <i className="fas fa-cog me-2"></i>
                  Settings
                </NavDropdown.Item>
              </LinkContainer>
              <NavDropdown.Divider />
              <LinkContainer to="/help">
                <NavDropdown.Item>
                  <i className="fas fa-question-circle me-2"></i>
                  Help Center
                </NavDropdown.Item>
              </LinkContainer>
              <NavDropdown.Item onClick={handleLogout}>
                <i className="fas fa-sign-out-alt me-2"></i>
                Logout
              </NavDropdown.Item>
            </NavDropdown>
          </Nav>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
};

export default Header;
