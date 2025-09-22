import React, { useState } from "react";
import { Card as BootstrapCard } from "react-bootstrap";

// 基础Card Props
export interface BaseCardProps {
  id?: string;
  title?: string;
  subtitle?: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
  headerClassName?: string;
  bodyClassName?: string;
  footerClassName?: string;
  borderless?: boolean;
  shadow?: "none" | "sm" | "md" | "lg";
  rounded?: boolean;
  actions?: React.ReactNode;
  footer?: React.ReactNode;
  collapsible?: boolean;
  defaultCollapsed?: boolean;
  onCollapse?: (collapsed: boolean) => void;
  loading?: boolean;
  overlay?: React.ReactNode;
  hover?: boolean;
  clickable?: boolean;
  onClick?: () => void;
}

// 统计Card Props
export interface StatCardProps extends Omit<BaseCardProps, "children"> {
  value: number | string;
  icon?: string;
  iconColor?: string;
  trend?: "up" | "down" | "stable";
  trendValue?: string;
  trendIcon?: string;
  valueColor?: string;
  prefix?: string;
  suffix?: string;
  formatValue?: (value: number | string) => string;
}

// 用户Card Props
export interface UserCardProps extends Omit<BaseCardProps, "children"> {
  user: {
    id: number;
    name: string;
    email?: string;
    role?: string;
    status?: string;
    avatar?: string;
    phone?: string;
    created_at?: string;
  };
  actions?: React.ReactNode;
  showStatus?: boolean;
  showEmail?: boolean;
  showRole?: boolean;
  showDate?: boolean;
}

// 客户Card Props
export interface ClientCardProps extends Omit<BaseCardProps, "children"> {
  client: {
    id: number;
    name: string;
    email: string;
    phone: string;
    company?: string;
    type: "personal" | "company";
    status: string;
    created_at?: string;
  };
  actions?: React.ReactNode;
  showStatus?: boolean;
  showContact?: boolean;
  showCompany?: boolean;
  showType?: boolean;
}

// 案件Card Props
export interface CaseCardProps extends Omit<BaseCardProps, "children"> {
  caseData: {
    id: number;
    title: string;
    case_number?: string;
    client_id: number;
    client_name?: string;
    lawyer_id?: number;
    lawyer_name?: string;
    case_type: string;
    priority: string;
    status: string;
    start_date?: string;
    expected_end_date?: string;
    case_amount?: number;
    description: string;
    created_at: string;
  };
  actions?: React.ReactNode;
  showClient?: boolean;
  showLawyer?: boolean;
  showPriority?: boolean;
  showType?: boolean;
  showProgress?: boolean;
}

const BaseCard: React.FC<BaseCardProps> = ({
  id,
  title,
  subtitle,
  description,
  children,
  className = "",
  headerClassName = "",
  bodyClassName = "",
  footerClassName = "",
  borderless = false,
  shadow = "sm",
  rounded = true,
  actions,
  footer,
  collapsible = false,
  defaultCollapsed = false,
  onCollapse,
  loading = false,
  overlay,
  hover = false,
  clickable = false,
  onClick,
}) => {
  const [collapsed, setCollapsed] = useState(defaultCollapsed);

  const toggleCollapse = () => {
    const newCollapsed = !collapsed;
    setCollapsed(newCollapsed);
    onCollapse?.(newCollapsed);
  };

  const shadowClass = shadow !== "none" ? `shadow-${shadow}` : "";
  const roundedClass = rounded ? "rounded" : "";
  const borderClass = borderless ? "border-0" : "";

  const cardClasses = [
    "base-card",
    shadowClass,
    roundedClass,
    borderClass,
    hover ? "base-card-hover" : "",
    clickable ? "base-card-clickable" : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  const handleClick = () => {
    if (clickable && onClick) {
      onClick();
    }
  };

  return (
    <BootstrapCard
      className={cardClasses}
      onClick={handleClick}
      style={clickable ? { cursor: "pointer" } : undefined}
    >
      {(title || subtitle || actions || collapsible) && (
        <BootstrapCard.Header
          className={`bg-white ${borderless ? "border-0" : ""} ${headerClassName}`}
        >
          <div className="d-flex justify-content-between align-items-center">
            <div className="flex-grow-1">
              {title && (
                <BootstrapCard.Title className="mb-0 d-flex align-items-center">
                  {title}
                  {description && (
                    <small className="text-muted ms-2">{description}</small>
                  )}
                </BootstrapCard.Title>
              )}
              {subtitle && (
                <BootstrapCard.Subtitle className="text-muted">
                  {subtitle}
                </BootstrapCard.Subtitle>
              )}
            </div>
            <div className="d-flex align-items-center ms-3">
              {actions && <div className="me-2">{actions}</div>}
              {collapsible && (
                <button
                  className="btn btn-sm btn-outline-secondary rounded-circle p-1"
                  onClick={toggleCollapse}
                  aria-label={collapsed ? "展开" : "收起"}
                >
                  <i
                    className={`fas ${collapsed ? "fa-chevron-down" : "fa-chevron-up"}`}
                  ></i>
                </button>
              )}
            </div>
          </div>
        </BootstrapCard.Header>
      )}

      {!collapsed && (
        <BootstrapCard.Body className={bodyClassName}>
          {loading && (
            <div className="text-center py-4">
              <div className="spinner-border text-primary" role="status">
                <span className="visually-hidden">加载中...</span>
              </div>
            </div>
          )}
          {!loading && children}
          {overlay && !loading && <div className="card-overlay">{overlay}</div>}
        </BootstrapCard.Body>
      )}

      {footer && !collapsed && (
        <BootstrapCard.Footer className={footerClassName}>
          {footer}
        </BootstrapCard.Footer>
      )}
    </BootstrapCard>
  );
};

// 统计Card组件
const StatCard: React.FC<StatCardProps> = ({
  value,
  icon,
  iconColor = "primary",
  trend,
  trendValue,
  trendIcon,
  valueColor,
  prefix = "",
  suffix = "",
  formatValue,
  ...cardProps
}) => {
  const displayValue = formatValue
    ? formatValue(value)
    : `${prefix}${value}${suffix}`;

  const getTrendIcon = () => {
    if (trendIcon) return <i className={`${trendIcon} me-1`}></i>;
    if (trend === "up") return <i className="fas fa-arrow-up me-1"></i>;
    if (trend === "down") return <i className="fas fa-arrow-down me-1"></i>;
    if (trend === "stable") return <i className="fas fa-minus me-1"></i>;
    return null;
  };

  const getTrendClass = () => {
    if (trend === "up") return "text-success";
    if (trend === "down") return "text-danger";
    if (trend === "stable") return "text-warning";
    return "text-muted";
  };

  const getValueColorClass = () => {
    if (valueColor) return `text-${valueColor}`;
    return "";
  };

  return (
    <BaseCard
      {...cardProps}
      className={`stat-card ${cardProps.className || ""}`}
    >
      <div className="d-flex justify-content-between align-items-center">
        <div className="flex-grow-1">
          {cardProps.title && (
            <div className="text-muted small mb-2">{cardProps.title}</div>
          )}
          <div className={`stat-value h4 mb-0 ${getValueColorClass()}`}>
            {displayValue}
          </div>
          {cardProps.subtitle && (
            <div className="text-muted small mt-1">{cardProps.subtitle}</div>
          )}
          {trendValue && (
            <div className="stat-trend mt-2">
              <span className={`small ${getTrendClass()}`}>
                {getTrendIcon()}
                {trendValue}
              </span>
            </div>
          )}
        </div>
        {icon && (
          <div
            className={`stat-icon bg-${iconColor} bg-opacity-10 rounded-circle p-3`}
          >
            <i className={`${icon} text-${iconColor} fa-2x`}></i>
          </div>
        )}
      </div>
    </BaseCard>
  );
};

// 用户Card组件
const UserCard: React.FC<UserCardProps> = ({
  user,
  actions,
  showStatus = true,
  showEmail = true,
  showRole = true,
  showDate = false,
  ...cardProps
}) => {
  const getRoleBadgeClass = (role: string) => {
    switch (role) {
      case "admin":
        return "bg-danger";
      case "manager":
        return "bg-warning";
      case "user":
        return "bg-info";
      default:
        return "bg-secondary";
    }
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case "active":
        return "bg-success";
      case "inactive":
        return "bg-secondary";
      case "pending":
        return "bg-warning";
      default:
        return "bg-secondary";
    }
  };

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((word) => word[0])
      .join("")
      .substring(0, 2)
      .toUpperCase();
  };

  return (
    <BaseCard
      {...cardProps}
      title={user.name}
      subtitle={user.email}
      actions={actions}
    >
      <div className="d-flex align-items-center mb-3">
        {user.avatar ? (
          <img
            src={user.avatar}
            alt={user.name}
            className="rounded-circle me-3"
            width="48"
            height="48"
          />
        ) : (
          <div
            className="avatar-circle bg-primary text-white rounded-circle me-3 d-flex align-items-center justify-content-center"
            style={{ width: "48px", height: "48px" }}
          >
            {getInitials(user.name)}
          </div>
        )}
        <div className="flex-grow-1">
          <div className="d-flex align-items-center gap-2 mb-1">
            {showRole && user.role && (
              <span
                className={`badge ${getRoleBadgeClass(user.role)} text-white`}
              >
                {user.role}
              </span>
            )}
            {showStatus && user.status && (
              <span
                className={`badge ${getStatusBadgeClass(user.status)} text-white`}
              >
                {user.status}
              </span>
            )}
          </div>
          {showEmail && user.email && (
            <div className="text-muted small">{user.email}</div>
          )}
          {user.phone && (
            <div className="text-muted small">
              <i className="fas fa-phone me-1"></i>
              {user.phone}
            </div>
          )}
        </div>
      </div>
      {showDate && user.created_at && (
        <div className="text-muted small">
          <i className="fas fa-calendar me-1"></i>
          创建于 {new Date(user.created_at).toLocaleDateString()}
        </div>
      )}
    </BaseCard>
  );
};

// 客户Card组件
const ClientCard: React.FC<ClientCardProps> = ({
  client,
  actions,
  showStatus = true,
  showContact = true,
  showCompany = true,
  showType = true,
  ...cardProps
}) => {
  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case "active":
        return "bg-success";
      case "inactive":
        return "bg-secondary";
      case "pending":
        return "bg-warning";
      default:
        return "bg-secondary";
    }
  };

  const getTypeBadgeClass = (type: string) => {
    switch (type) {
      case "personal":
        return "bg-info";
      case "company":
        return "bg-primary";
      default:
        return "bg-secondary";
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case "active":
        return "活跃";
      case "inactive":
        return "未活跃";
      case "pending":
        return "待审核";
      default:
        return status;
    }
  };

  const getTypeText = (type: string) => {
    switch (type) {
      case "personal":
        return "个人";
      case "company":
        return "企业";
      default:
        return type;
    }
  };

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((word) => word[0])
      .join("")
      .substring(0, 2)
      .toUpperCase();
  };

  const defaultActions = (
    <div className="d-flex gap-1">
      <button className="btn btn-sm btn-outline-primary" title="编辑">
        <i className="fas fa-edit"></i>
      </button>
      <button className="btn btn-sm btn-outline-info" title="查看">
        <i className="fas fa-eye"></i>
      </button>
      <button className="btn btn-sm btn-outline-danger" title="删除">
        <i className="fas fa-trash"></i>
      </button>
    </div>
  );

  return (
    <BaseCard
      {...cardProps}
      title={client.name}
      subtitle={client.email}
      actions={actions || defaultActions}
      className={`client-card ${cardProps.className || ""}`}
    >
      <div className="d-flex align-items-center mb-3">
        <div
          className="avatar-circle bg-primary text-white rounded-circle me-3 d-flex align-items-center justify-content-center"
          style={{ width: "48px", height: "48px" }}
        >
          {getInitials(client.name)}
        </div>
        <div className="flex-grow-1">
          <div className="d-flex align-items-center gap-2 mb-1">
            {showStatus && (
              <span
                className={`badge ${getStatusBadgeClass(client.status)} text-white`}
              >
                {getStatusText(client.status)}
              </span>
            )}
            {showType && (
              <span
                className={`badge ${getTypeBadgeClass(client.type)} text-white`}
              >
                {getTypeText(client.type)}
              </span>
            )}
          </div>
          {showCompany && client.company && (
            <div className="text-muted small">
              <i className="fas fa-building me-1"></i>
              {client.company}
            </div>
          )}
        </div>
      </div>

      {showContact && (
        <div className="contact-info mb-3">
          <div className="d-flex align-items-center mb-2">
            <i className="fas fa-envelope text-muted me-2"></i>
            <small className="text-muted">{client.email}</small>
          </div>
          <div className="d-flex align-items-center">
            <i className="fas fa-phone text-muted me-2"></i>
            <small className="text-muted">{client.phone}</small>
          </div>
        </div>
      )}

      {client.created_at && (
        <div className="text-muted small">
          <i className="fas fa-calendar me-1"></i>
          创建于 {new Date(client.created_at).toLocaleDateString()}
        </div>
      )}
    </BaseCard>
  );
};

// 案件Card组件
const CaseCard: React.FC<CaseCardProps> = ({
  caseData,
  actions,
  showClient = true,
  showLawyer = true,
  showPriority = true,
  showType = true,
  showProgress = false,
  ...cardProps
}) => {
  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case "pending":
        return "bg-warning";
      case "active":
        return "bg-primary";
      case "closed":
        return "bg-success";
      case "suspended":
        return "bg-secondary";
      default:
        return "bg-secondary";
    }
  };

  const getPriorityBadgeClass = (priority: string) => {
    switch (priority) {
      case "low":
        return "bg-info";
      case "medium":
        return "bg-warning text-dark";
      case "high":
        return "bg-danger";
      case "urgent":
        return "bg-danger";
      default:
        return "bg-secondary";
    }
  };

  const getTypeBadgeClass = (caseType: string) => {
    switch (caseType) {
      case "civil":
        return "bg-primary";
      case "criminal":
        return "bg-danger";
      case "commercial":
        return "bg-success";
      case "administrative":
        return "bg-info";
      default:
        return "bg-secondary";
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case "pending":
        return "待处理";
      case "active":
        return "进行中";
      case "closed":
        return "已结案";
      case "suspended":
        return "已暂停";
      default:
        return status;
    }
  };

  const getPriorityText = (priority: string) => {
    switch (priority) {
      case "low":
        return "低";
      case "medium":
        return "中";
      case "high":
        return "高";
      case "urgent":
        return "紧急";
      default:
        return priority;
    }
  };

  const getTypeText = (caseType: string) => {
    switch (caseType) {
      case "civil":
        return "民事";
      case "criminal":
        return "刑事";
      case "commercial":
        return "商业";
      case "administrative":
        return "行政";
      default:
        return caseType;
    }
  };

  const getProgressPercentage = () => {
    if (caseData.status === "closed") return 100;
    if (caseData.status === "suspended") return 0;
    if (caseData.status === "pending") return 10;
    if (caseData.status === "active") {
      if (caseData.expected_end_date) {
        const startDate = caseData.start_date || caseData.created_at;
        if (!startDate) return 50;
        const start = new Date(startDate).getTime();
        const end = new Date(caseData.expected_end_date).getTime();
        const now = Date.now();
        return Math.min(
          Math.max(Math.round(((now - start) / (end - start)) * 100), 0),
          100,
        );
      }
      return 50; // 默认进度
    }
    return 0;
  };

  const defaultActions = (
    <div className="d-flex gap-1">
      <button className="btn btn-sm btn-outline-primary" title="编辑">
        <i className="fas fa-edit"></i>
      </button>
      <button className="btn btn-sm btn-outline-info" title="查看">
        <i className="fas fa-eye"></i>
      </button>
      <button className="btn btn-sm btn-outline-danger" title="删除">
        <i className="fas fa-trash"></i>
      </button>
    </div>
  );

  const progressPercentage = getProgressPercentage();

  return (
    <BaseCard
      {...cardProps}
      title={caseData.title}
      subtitle={
        caseData.case_number ? `案件号: ${caseData.case_number}` : undefined
      }
      actions={actions || defaultActions}
      className={`case-card ${cardProps.className || ""}`}
    >
      <div className="mb-3">
        <p className="text-muted small mb-0">
          {caseData.description?.substring(0, 100)}
          {caseData.description && caseData.description.length > 100
            ? "..."
            : ""}
        </p>
      </div>

      <div className="d-flex flex-wrap gap-2 mb-3">
        <span
          className={`badge ${getTypeBadgeClass(caseData.case_type)} text-white`}
        >
          {getTypeText(caseData.case_type)}
        </span>
        {showPriority && (
          <span
            className={`badge ${getPriorityBadgeClass(caseData.priority)} text-white`}
          >
            {getPriorityText(caseData.priority)}
          </span>
        )}
        <span
          className={`badge ${getStatusBadgeClass(caseData.status)} text-white`}
        >
          {getStatusText(caseData.status)}
        </span>
      </div>

      <div className="d-flex align-items-center mb-3">
        {showClient && caseData.client_name && (
          <div className="d-flex align-items-center me-3">
            <div
              className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2"
              style={{ width: "32px", height: "32px" }}
            >
              <i className="fas fa-user text-muted"></i>
            </div>
            <div>
              <div className="small fw-bold">{caseData.client_name}</div>
              <div className="small text-muted">委托人</div>
            </div>
          </div>
        )}

        {showLawyer && caseData.lawyer_name && (
          <div className="d-flex align-items-center">
            <div
              className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2"
              style={{ width: "32px", height: "32px" }}
            >
              <i className="fas fa-user-tie text-muted"></i>
            </div>
            <div>
              <div className="small fw-bold">{caseData.lawyer_name}</div>
              <div className="small text-muted">律师</div>
            </div>
          </div>
        )}
      </div>

      {showProgress && (
        <div className="progress-info mb-3">
          <div className="d-flex justify-content-between mb-1">
            <small className="text-muted">进度</small>
            <small className="text-muted">{progressPercentage}%</small>
          </div>
          <div className="progress" style={{ height: "6px" }}>
            <div
              className={`progress-bar ${
                progressPercentage === 100
                  ? "bg-success"
                  : progressPercentage > 70
                    ? "bg-primary"
                    : progressPercentage > 30
                      ? "bg-warning"
                      : "bg-danger"
              }`}
              role="progressbar"
              style={{ width: `${progressPercentage}%` }}
            ></div>
          </div>
        </div>
      )}

      {caseData.case_amount && (
        <div className="text-muted small mb-2">
          <i className="fas fa-dollar-sign me-1"></i>
          案件金额: ¥{caseData.case_amount.toLocaleString()}
        </div>
      )}

      <div className="d-flex justify-content-between text-muted small">
        {caseData.start_date && (
          <span>
            <i className="fas fa-calendar-plus me-1"></i>
            {new Date(caseData.start_date).toLocaleDateString()}
          </span>
        )}
        {caseData.expected_end_date && (
          <span>
            <i className="fas fa-calendar-check me-1"></i>
            {new Date(caseData.expected_end_date).toLocaleDateString()}
          </span>
        )}
      </div>
    </BaseCard>
  );
};

// 导出所有Card组件
export { BaseCard, StatCard, UserCard, ClientCard, CaseCard };
export default BaseCard;
