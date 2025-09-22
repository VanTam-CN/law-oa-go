import React, { useState } from "react";
import { Nav, Badge } from "react-bootstrap";
import { LinkContainer } from "react-router-bootstrap";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import LanguageSwitcher from "./LanguageSwitcher";
import "./Sidebar.css";

interface SidebarProps {
  collapsed: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed }) => {
  const { user } = useSelector((state: RootState) => state.auth);
  const [activeKey, setActiveKey] = useState("dashboard");

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

  const menuItems = [
    // 基础功能
    {
      key: "dashboard",
      icon: "fas fa-tachometer-alt",
      label: "仪表板",
      path: "/dashboard",
    },

    // 案件管理分组
    {
      key: "case-management",
      icon: "fas fa-gavel",
      label: "案件管理",
      path: "/case-management",
    },
    {
      key: "create-case",
      icon: "fas fa-plus-circle",
      label: "创建案件",
      path: "/create-case",
      indent: true,
    },

    // 客户与律师管理
    {
      key: "client-management",
      icon: "fas fa-users",
      label: "客户管理",
      path: "/client-management",
    },
    {
      key: "lawyer-management",
      icon: "fas fa-user-tie",
      label: "律师管理",
      path: "/lawyer-management",
    },

    // 业务管理
    {
      key: "approval",
      icon: "fas fa-clipboard-check",
      label: "审批流程",
      path: "/approval",
    },
    {
      key: "project",
      icon: "fas fa-project-diagram",
      label: "项目管理",
      path: "/project",
    },
    {
      key: "finance",
      icon: "fas fa-chart-line",
      label: "财务管理",
      path: "/finance",
    },

    // 文档与工具
    {
      key: "file",
      icon: "fas fa-folder-open",
      label: "文件管理",
      path: "/file",
    },
    {
      key: "tools",
      icon: "fas fa-tools",
      label: "专业工具",
      path: "/tools",
      subItems: [
        { key: "fee-calculator", label: "诉讼费计算", path: "/fee-calculator" },
        { key: "law-search", label: "法条查询", path: "/law-search" },
      ],
    },

    // 分析报告
    {
      key: "analytics",
      icon: "fas fa-chart-pie",
      label: "数据分析",
      path: "/analytics",
    },

    // 原有功能
    {
      key: "calendar",
      icon: "fas fa-calendar-alt",
      label: "日历",
      path: "/calendar",
    },
    {
      key: "tasks",
      icon: "fas fa-tasks",
      label: "任务",
      path: "/tasks",
    },
    {
      key: "documents",
      icon: "fas fa-file-contract",
      label: "文档",
      path: "/documents",
    },
    {
      key: "reports",
      icon: "fas fa-file-alt",
      label: "报告",
      path: "/reports",
    },

    // 系统管理
    {
      key: "users",
      icon: "fas fa-user-shield",
      label: "用户管理",
      path: "/users",
      adminOnly: true,
    },
  ];

  return (
    <div
      className={`sidebar bg-dark text-white ${collapsed ? "collapsed" : ""}`}
    >
      <div className="sidebar-header p-3 border-bottom border-secondary">
        {!collapsed && (
          <div className="d-flex align-items-center">
            <div
              className="bg-primary rounded-circle d-flex align-items-center justify-content-center me-3"
              style={{ width: "40px", height: "40px" }}
            >
              <i className="fas fa-balance-scale text-white"></i>
            </div>
            <div>
              <h5 className="mb-0">Law OA</h5>
              <small className="text-muted">Legal Practice Management</small>
            </div>
          </div>
        )}
        {collapsed && (
          <div className="text-center">
            <div
              className="bg-primary rounded-circle d-flex align-items-center justify-content-center mx-auto"
              style={{ width: "40px", height: "40px" }}
            >
              <i className="fas fa-balance-scale text-white"></i>
            </div>
          </div>
        )}
      </div>

      <Nav className="flex-column p-2" activeKey={activeKey}>
        {menuItems.map((item) => {
          if (item.adminOnly && user?.role !== "admin") {
            return null;
          }

          if (item.subItems) {
            // 处理带子菜单的项
            return (
              <div key={item.key} className="sidebar-submenu">
                <LinkContainer
                  to={item.path}
                  onClick={() => setActiveKey(item.key)}
                >
                  <Nav.Link className="text-white sidebar-link">
                    <div className="d-flex align-items-center">
                      <div className="sidebar-icon rounded-circle d-flex align-items-center justify-content-center me-3">
                        <i className={item.icon}></i>
                      </div>
                      {!collapsed && (
                        <span className="sidebar-text">{item.label}</span>
                      )}
                    </div>
                  </Nav.Link>
                </LinkContainer>

                {!collapsed && (
                  <div className="submenu-items ms-4">
                    {item.subItems.map((subItem) => (
                      <LinkContainer
                        key={subItem.key}
                        to={subItem.path}
                        onClick={() => setActiveKey(subItem.key)}
                      >
                        <Nav.Link className="text-white sidebar-sublink">
                          <div className="d-flex align-items-center">
                            <div className="sidebar-subicon rounded-circle d-flex align-items-center justify-content-center me-2">
                              <i className="fas fa-angle-right"></i>
                            </div>
                            <span className="sidebar-text">{subItem.label}</span>
                          </div>
                        </Nav.Link>
                      </LinkContainer>
                    ))}
                  </div>
                )}
              </div>
            );
          }

          // 处理普通菜单项
          return (
            <LinkContainer
              key={item.key}
              to={item.path}
              onClick={() => setActiveKey(item.key)}
            >
              <Nav.Link className={`text-white sidebar-link ${item.indent ? 'sidebar-indented' : ''}`}>
                <div className="d-flex align-items-center">
                  <div className={`sidebar-icon rounded-circle d-flex align-items-center justify-content-center ${item.indent ? 'me-2' : 'me-3'}`}>
                    <i className={item.icon}></i>
                  </div>
                  {!collapsed && (
                    <span className="sidebar-text">{item.label}</span>
                  )}
                </div>
              </Nav.Link>
            </LinkContainer>
          );
        })}
      </Nav>

      {!collapsed && (
        <div className="sidebar-footer mt-auto p-3 border-top border-secondary">
          <div className="user-info mb-3">
            <div className="d-flex align-items-center">
              <div
                className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3"
                style={{ width: "32px", height: "32px" }}
              >
                <i className="fas fa-user text-muted"></i>
              </div>
              <div className="flex-grow-1">
                <div className="fw-bold small">{user?.name}</div>
                <div className="d-flex align-items-center">
                  <Badge
                    bg={getRoleBadgeClass(user?.role || "user")}
                    className="me-2"
                  >
                    {getRoleText(user?.role || "user")}
                  </Badge>
                  <Badge bg="success">
                    {getStatusText(user?.status || "active")}
                  </Badge>
                </div>
              </div>
            </div>
          </div>

          <div className="quick-actions">
            <LinkContainer to="/profile">
              <Nav.Link className="text-white p-2 rounded mb-2">
                <i className="fas fa-user me-2"></i>
                Profile
              </Nav.Link>
            </LinkContainer>
            <LinkContainer to="/settings">
              <Nav.Link className="text-white p-2 rounded mb-2">
                <i className="fas fa-cog me-2"></i>
                Settings
              </Nav.Link>
            </LinkContainer>
            <LinkContainer to="/help">
              <Nav.Link className="text-white p-2 rounded">
                <i className="fas fa-question-circle me-2"></i>
                Help
              </Nav.Link>
            </LinkContainer>

            {/* Language Switcher */}
            <div className="mt-3 pt-3 border-top border-secondary">
              <LanguageSwitcher />
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Sidebar;
