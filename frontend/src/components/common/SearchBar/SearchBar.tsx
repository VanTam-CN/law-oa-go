import React, { useState, useEffect } from "react";
import {
  Form,
  InputGroup,
  Button,
  Dropdown,
  Badge,
  Card,
  Row,
  Col,
} from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { useSelector, useDispatch } from "react-redux";
import { addToSearchHistory } from "../../../store/slices/uiSlice";

// 直接定义类型避免导入问题
interface RootState {
  ui: {
    searchHistory: string[];
  };
}

// 搜索过滤器类型
export interface SearchFilter {
  key: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: Array<{
    value: string;
    label: string;
  }>;
}

// 搜索配置类型
export interface SearchConfig {
  type: "all" | "clients" | "cases" | "documents" | "users";
  status: "all" | "active" | "inactive" | "pending" | "closed";
  dateRange: "any" | "today" | "this_week" | "this_month" | "this_year";
}

// 搜索模式
export type SearchMode = "simple" | "advanced";

// 搜索结果类型
export interface SearchResult {
  id: number | string;
  type: "client" | "case" | "document" | "user";
  title: string;
  description?: string;
  url: string;
  relevance?: number;
}

// 组件属性
interface SearchBarProps {
  // 基础属性
  value?: string;
  onChange?: (value: string) => void;
  onSubmit?: (query: string, config: SearchConfig) => void;
  placeholder?: string;
  className?: string;

  // 模式控制
  mode?: SearchMode;
  showAdvancedToggle?: boolean;
  enableAdvancedSearch?: boolean;

  // 过滤器
  filters?: SearchFilter[];
  showResetButton?: boolean;
  onReset?: () => void;

  // 高级搜索配置
  initialConfig?: Partial<SearchConfig>;

  // 导航相关
  enableNavigation?: boolean;
  searchRoute?: string;

  // 大小和样式
  size?: "sm" | "md" | "lg";
  variant?: "outline-primary" | "primary" | "outline-secondary";

  // 搜索历史
  showHistory?: boolean;
  maxHistoryItems?: number;
}

const SearchBar: React.FC<SearchBarProps> = ({
  value = "",
  onChange,
  onSubmit,
  placeholder = "Search...",
  className = "",
  mode = "simple",
  showAdvancedToggle = true,
  enableAdvancedSearch = true,
  filters = [],
  showResetButton = true,
  onReset,
  initialConfig = {},
  enableNavigation = false,
  searchRoute = "/search",
  size = "md",
  variant = "outline-secondary",
  showHistory = false,
  maxHistoryItems = 5,
}) => {
  // 内部状态
  const [query, setQuery] = useState(value);
  const [showAdvanced, setShowAdvanced] = useState(mode === "advanced");
  const [config, setConfig] = useState<SearchConfig>({
    type: "all",
    status: "all",
    dateRange: "any",
    ...initialConfig,
  });

  // 外部状态
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { searchHistory } = useSelector((state: RootState) => state.ui);

  // 同步外部value
  useEffect(() => {
    setQuery(value);
  }, [value]);

  // 处理搜索提交
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!query.trim()) return;

    // 添加到搜索历史
    dispatch(addToSearchHistory(query));

    // 构建搜索查询
    const searchQuery = buildSearchQuery(query, config);

    // 调用回调
    if (onSubmit) {
      onSubmit(searchQuery, config);
    }

    // 导航到搜索页面
    if (enableNavigation) {
      navigate(`${searchRoute}?q=${encodeURIComponent(searchQuery)}`);
    }
  };

  // 构建搜索查询
  const buildSearchQuery = (
    baseQuery: string,
    config: SearchConfig,
  ): string => {
    let query = baseQuery;

    if (config.type !== "all") {
      query += ` type:${config.type}`;
    }

    if (config.status !== "all") {
      query += ` status:${config.status}`;
    }

    if (config.dateRange !== "any") {
      query += ` date:${config.dateRange}`;
    }

    return query.trim();
  };

  // 重置搜索
  const handleReset = () => {
    setQuery("");
    setConfig({
      type: "all",
      status: "all",
      dateRange: "any",
      ...initialConfig,
    });

    if (onReset) {
      onReset();
    }
  };

  // 处理历史搜索
  const handleHistoryClick = (historyItem: string) => {
    setQuery(historyItem);
    if (onChange) {
      onChange(historyItem);
    }
  };

  // 获取按钮大小类
  const getButtonSizeClass = () => {
    switch (size) {
      case "sm":
        return "btn-sm";
      case "lg":
        return "btn-lg";
      default:
        return "";
    }
  };

  // 获取输入框大小类
  const getInputSizeClass = () => {
    switch (size) {
      case "sm":
        return "form-control-sm";
      case "lg":
        return "form-control-lg";
      default:
        return "";
    }
  };

  return (
    <div className={`search-bar-component ${className}`}>
      {/* 主搜索表单 */}
      <Form onSubmit={handleSubmit}>
        <InputGroup>
          <Form.Control
            type="text"
            placeholder={placeholder}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              if (onChange) {
                onChange(e.target.value);
              }
            }}
            className={getInputSizeClass()}
          />

          {/* 搜索按钮 */}
          <Button
            variant={variant}
            type="submit"
            className={getButtonSizeClass()}
          >
            <i className="fas fa-search"></i>
          </Button>

          {/* 高级搜索按钮 */}
          {showAdvancedToggle && enableAdvancedSearch && (
            <Dropdown>
              <Dropdown.Toggle
                variant={variant}
                className={getButtonSizeClass()}
              >
                <i className="fas fa-caret-down"></i>
              </Dropdown.Toggle>
              <Dropdown.Menu>
                <Dropdown.Item onClick={() => setShowAdvanced(!showAdvanced)}>
                  <i className="fas fa-sliders-h me-2"></i>
                  Advanced Search
                </Dropdown.Item>
                {showHistory && (
                  <>
                    <Dropdown.Item onClick={() => navigate("/search")}>
                      <i className="fas fa-history me-2"></i>
                      Search History
                    </Dropdown.Item>
                    {searchHistory
                      .slice(0, maxHistoryItems)
                      .map((item: string, index: number) => (
                        <Dropdown.Item
                          key={index}
                          onClick={() => handleHistoryClick(item)}
                        >
                          <i className="fas fa-clock me-2"></i>
                          {item}
                        </Dropdown.Item>
                      ))}
                  </>
                )}
              </Dropdown.Menu>
            </Dropdown>
          )}
        </InputGroup>
      </Form>

      {/* 高级搜索面板 */}
      {showAdvanced && enableAdvancedSearch && (
        <Card className="mt-3">
          <Card.Body>
            <h5 className="mb-3">
              <i className="fas fa-sliders-h me-2"></i>
              Advanced Search
            </h5>
            <Form onSubmit={handleSubmit}>
              <Row>
                <Col md={4}>
                  <Form.Group className="mb-3">
                    <Form.Label>Search Type</Form.Label>
                    <Form.Select
                      value={config.type}
                      onChange={(e) =>
                        setConfig({
                          ...config,
                          type: e.target.value as SearchConfig["type"],
                        })
                      }
                    >
                      <option value="all">All Types</option>
                      <option value="clients">Clients</option>
                      <option value="cases">Cases</option>
                      <option value="documents">Documents</option>
                      <option value="users">Users</option>
                    </Form.Select>
                  </Form.Group>
                </Col>
                <Col md={4}>
                  <Form.Group className="mb-3">
                    <Form.Label>Status</Form.Label>
                    <Form.Select
                      value={config.status}
                      onChange={(e) =>
                        setConfig({
                          ...config,
                          status: e.target.value as SearchConfig["status"],
                        })
                      }
                    >
                      <option value="all">All Statuses</option>
                      <option value="active">Active</option>
                      <option value="inactive">Inactive</option>
                      <option value="pending">Pending</option>
                      <option value="closed">Closed</option>
                    </Form.Select>
                  </Form.Group>
                </Col>
                <Col md={4}>
                  <Form.Group className="mb-3">
                    <Form.Label>Date Range</Form.Label>
                    <Form.Select
                      value={config.dateRange}
                      onChange={(e) =>
                        setConfig({
                          ...config,
                          dateRange: e.target
                            .value as SearchConfig["dateRange"],
                        })
                      }
                    >
                      <option value="any">Any Time</option>
                      <option value="today">Today</option>
                      <option value="this_week">This Week</option>
                      <option value="this_month">This Month</option>
                      <option value="this_year">This Year</option>
                    </Form.Select>
                  </Form.Group>
                </Col>
              </Row>

              {/* 自定义过滤器 */}
              {filters.length > 0 && (
                <Row className="mt-3">
                  {filters.map((filter) => (
                    <Col md={4} key={filter.key}>
                      <Form.Group className="mb-3">
                        <Form.Label>{filter.label}</Form.Label>
                        <Form.Select
                          value={filter.value}
                          onChange={(e) => filter.onChange(e.target.value)}
                        >
                          {filter.options.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          ))}
                        </Form.Select>
                      </Form.Group>
                    </Col>
                  ))}
                </Row>
              )}

              {/* 操作按钮 */}
              <div className="d-flex justify-content-end">
                <Button
                  variant="outline-secondary"
                  className="me-2"
                  onClick={() => setShowAdvanced(false)}
                >
                  <i className="fas fa-times me-2"></i>
                  Cancel
                </Button>

                {showResetButton && (
                  <Button
                    variant="outline-primary"
                    className="me-2"
                    onClick={handleReset}
                  >
                    <i className="fas fa-undo me-2"></i>
                    Reset
                  </Button>
                )}

                <Button variant="primary" type="submit">
                  <i className="fas fa-search me-2"></i>
                  Search
                </Button>
              </div>
            </Form>
          </Card.Body>
        </Card>
      )}

      {/* 简单过滤器 */}
      {!showAdvanced && filters.length > 0 && (
        <div className="d-flex mt-3">
          {filters.map((filter) => (
            <Dropdown key={filter.key} className="me-2">
              <Dropdown.Toggle variant="outline-secondary" size="sm">
                {filter.label}:{" "}
                {filter.options.find((opt) => opt.value === filter.value)
                  ?.label || "All"}
              </Dropdown.Toggle>
              <Dropdown.Menu>
                {filter.options.map((option) => (
                  <Dropdown.Item
                    key={option.value}
                    onClick={() => filter.onChange(option.value)}
                    active={filter.value === option.value}
                  >
                    {option.label}
                  </Dropdown.Item>
                ))}
              </Dropdown.Menu>
            </Dropdown>
          ))}

          {showResetButton && (
            <Button
              variant="outline-primary"
              size="sm"
              className="me-2"
              onClick={handleReset}
            >
              <i className="fas fa-undo me-1"></i>
              Reset
            </Button>
          )}
        </div>
      )}
    </div>
  );
};

export default SearchBar;
