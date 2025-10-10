# Design Document

## Architecture Overview

本设计方案专注于修复案件管理界面数据显示问题，通过解决前后端数据格式不匹配、API调用错误和数据显示问题，确保用户能够正常查看和管理案件信息。系统采用单体架构设计，前端使用React + TypeScript，后端使用Go + Gin框架，通过RESTful API进行数据交互。

## System Architecture

### High-Level Architecture

```
┌─────────────────┐    HTTP/HTTPS    ┌─────────────────┐
│                 │ ◄──────────────► │                 │
│   Frontend      │                  │    Backend      │
│   (React)       │                  │    (Go/Gin)     │
│                 │ ◄──────────────► │                 │
└─────────────────┘                  └─────────────────┘
       │                                     │
       │                                     │
       ▼                                     ▼
┌─────────────────┐                  ┌─────────────────┐
│   Browser       │                  │   MySQL DB      │
│   (Chrome)      │                  │                 │
│                 │                  │                 │
└─────────────────┘                  └─────────────────┘
```

### Component Diagram

```
Frontend Components:
┌─────────────────────────────────────────────────────────────┐
│                   CaseManagementPage                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Header    │  │ SearchBar   │  │ FilterPanel │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              CaseTable                             │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │    │
│  │  │ Columns │ │   Rows  │ │ Paging  │ │ Actions │  │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘

Backend Components:
┌─────────────────────────────────────────────────────────────┐
│                    CaseHandler                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ ListCases   │  │ GetCase     │  │ CreateCase  │          │
│  │ UpdateCase  │  │ DeleteCase  │  │ GetStats    │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                   CaseService                               │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Validation  │  │ Business    │  │ Data        │          │
│  │ Logic       │  │ Logic       │  │ Transform   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                CaseRepository                               │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Query       │  │ Pagination   │  │ Filtering    │          │
│  │ Building    │  │ Logic        │  │ Logic        │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
User Request Flow:
1. User loads CaseManagementPage
2. Component mounts → useEffect triggered
3. Frontend calls getCases() API
4. Request: GET /api/cases?page=1&page_size=10&search=...
5. Backend: CaseHandler.ListCases() validates request
6. Backend: CaseService.ListCases() processes business logic
7. Backend: CaseRepository queries database with filters
8. Database returns Case models
9. Service transforms to CaseResponse format
10. Handler returns JSON response with pagination
11. Frontend receives response → updates state
12. Component re-renders with data

Error Handling Flow:
1. API call fails → catch error in caseService
2. Create AppError with details
3. Component catches error → displays user-friendly message
4. Console logs detailed error for debugging
```

## Technical Design

### Backend Design

#### Core Components
- **CaseHandler**: 处理HTTP请求和响应，参数验证
- **CaseService**: 业务逻辑处理，数据转换和验证
- **CaseRepository**: 数据库操作，查询构建和分页
- **ResponseBuilder**: 统一API响应格式

#### API Design

**案件列表API**
```
GET /api/cases
Query Parameters:
- page: int (default: 1)
- page_size: int (default: 20, max: 100)
- status: string (pending|active|closed|suspended)
- case_type: string (civil|criminal|commercial|administrative)
- priority: string (low|medium|high|urgent)
- client_id: int
- lawyer_id: int
- search: string

Response Format:
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "案件标题",
      "description": "案件描述",
      "client_id": 1,
      "client_name": "客户名称",
      "lawyer_id": 1,
      "lawyer_name": "律师名称",
      "case_type": "civil",
      "priority": "medium",
      "status": "pending",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "v1",
    "server": "law-oa-go",
    "environment": "development"
  }
}
```

**案件详情API**
```
GET /api/cases/{id}

Response Format:
{
  "success": true,
  "data": {
    "id": 1,
    "title": "案件标题",
    "description": "案件描述",
    "client_id": 1,
    "client": {
      "id": 1,
      "name": "客户名称",
      "email": "client@example.com",
      "phone": "1234567890"
    },
    "lawyer_id": 1,
    "lawyer": {
      "id": 1,
      "name": "律师名称",
      "email": "lawyer@example.com"
    },
    "case_type": "civil",
    "priority": "medium",
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "v1"
  }
}
```

#### Database Design

**Cases Table Structure**
```sql
CREATE TABLE cases (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    client_id BIGINT UNSIGNED NOT NULL,
    lawyer_id BIGINT UNSIGNED NOT NULL,
    case_type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'pending',
    start_date TIMESTAMP NULL,
    end_date TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    INDEX idx_client_id (client_id),
    INDEX idx_lawyer_id (lawyer_id),
    INDEX idx_status (status),
    INDEX idx_case_type (case_type),
    INDEX idx_priority (priority),
    INDEX idx_created_at (created_at),

    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (lawyer_id) REFERENCES users(id)
);
```

**Related Tables**
```sql
-- Clients Table
CREATE TABLE clients (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    company VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20),
    address TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Users Table (for lawyers)
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### Service Layer

**CaseService 核心方法**
```go
// ListCases 获取案件列表（支持分页和过滤）
func (s *CaseService) ListCases(ctx context.Context, req *CaseListRequest) ([]*CaseResponse, int64, error)

// GetCaseByID 获取单个案件详情
func (s *CaseService) GetCaseByID(ctx context.Context, id uint) (*CaseResponse, error)

// CreateCase 创建新案件
func (s *CaseService) CreateCase(ctx context.Context, req *CreateCaseRequest) (*CaseResponse, error)

// UpdateCase 更新案件信息
func (s *CaseService) UpdateCase(ctx context.Context, id uint, req *UpdateCaseRequest) (*CaseResponse, error)

// DeleteCase 删除案件
func (s *CaseService) DeleteCase(ctx context.Context, id uint) error

// toCaseResponse 数据模型转换
func (s *CaseService) toCaseResponse(caseModel *models.Case) *CaseResponse
```

**数据转换逻辑**
```go
func (s *CaseService) toCaseResponse(caseModel *models.Case) *CaseResponse {
    response := &CaseResponse{
        ID:          caseModel.ID,
        Title:       caseModel.Title,
        Description: caseModel.Description,
        ClientID:    caseModel.ClientID,
        LawyerID:    caseModel.LawyerID,
        CaseType:    caseModel.CaseType,
        Priority:    caseModel.Priority,
        Status:      caseModel.Status,
        CreatedAt:   caseModel.CreatedAt,
        UpdatedAt:   caseModel.UpdatedAt,
    }

    // 处理客户信息显示
    if caseModel.Client != nil {
        if caseModel.Client.Company != "" {
            response.ClientName = caseModel.Client.Company
        } else {
            response.ClientName = caseModel.Client.Name
        }
    }

    // 处理律师信息
    if caseModel.Lawyer != nil {
        response.LawyerName = caseModel.Lawyer.Name
    }

    return response
}
```

#### Security Model
- **认证**: JWT Token验证，通过AuthMiddleware
- **授权**: 基于角色的访问控制，律师只能查看自己负责的案件
- **数据保护**: 输入验证，SQL注入防护，XSS防护
- **审计**: 操作日志记录，请求追踪

### Frontend Design

#### Component Architecture

**CaseManagementPage 组件结构**
```typescript
interface CaseManagementPageProps {}

interface Case {
  id: number;
  title: string;
  description: string;
  client_id: number;
  client_name?: string;
  lawyer_id: number | null;
  lawyer_name?: string;
  case_type: string;
  priority: string;
  status: string;
  created_at: string;
  updated_at: string;
}

const CaseManagementPage: React.FC<CaseManagementPageProps> = () => {
  // 状态管理
  const [cases, setCases] = useState<Case[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // 搜索和过滤状态
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');

  // 组件生命周期
  useEffect(() => {
    fetchCases();
  }, [pagination.current, searchText, statusFilter, typeFilter]);

  // API调用
  const fetchCases = async () => {
    setLoading(true);
    try {
      const response = await getCases({
        page: pagination.current,
        page_size: pagination.pageSize,
        search: searchText,
        status: statusFilter,
        case_type: typeFilter
      });

      setCases(response.data);
      setPagination({
        current: response.pagination.page,
        pageSize: response.pagination.page_size,
        total: response.pagination.total
      });
    } catch (error) {
      console.error('获取案件列表失败:', error);
      // 错误处理逻辑
    } finally {
      setLoading(false);
    }
  };

  // 渲染逻辑
  return (
    <div className="case-management">
      <SearchBar />
      <FilterPanel />
      <CaseTable />
      <PaginationComponent />
    </div>
  );
};
```

#### State Management
- **本地状态**: 使用React useState管理组件状态
- **全局状态**: 使用Context API管理用户认证信息
- **缓存策略**: 使用React Query或SWR进行数据缓存和同步
- **错误状态**: 统一的错误处理和用户反馈机制

#### UI/UX Design
- **响应式设计**: 支持桌面和移动端访问
- **加载状态**: 骨架屏和加载指示器
- **错误状态**: 友好的错误提示和重试机制
- **空状态**: 无数据时的提示和引导
- **交互反馈**: 按钮状态、悬停效果、过渡动画

#### Performance Optimization
- **虚拟滚动**: 大数据量时的性能优化
- **防抖搜索**: 搜索输入的防抖处理
- **懒加载**: 分页数据的按需加载
- **缓存优化**: API响应缓存和状态缓存

## Integration Design

### External Integrations
- **认证服务**: JWT Token验证
- **文件存储**: 本地文件系统或云存储
- **通知服务**: 邮件和站内通知

### Internal Integrations
- **用户管理**: 律师和客户数据获取
- **权限管理**: RBAC权限验证
- **审计日志**: 操作记录和追踪

### Data Migration
- **数据同步**: 确保前后端数据格式一致性
- **数据验证**: 完整的数据验证规则
- **向后兼容**: 保持API版本兼容性

## Technology Stack

### Backend Technologies
- **语言/框架**: Go 1.23+ + Gin v1.9.1
- **数据库**: MySQL 8.0+ with GORM ORM
- **缓存**: Redis 7+ (go-redis/v9)
- **认证**: JWT (golang-jwt/jwt v5.0.0)
- **日志**: Zap v1.24.0 (结构化日志)

### Frontend Technologies
- **框架**: React 18.2.0 + TypeScript 5.9.2
- **UI库**: Bootstrap 5.3.1 + React Bootstrap 2.8.0
- **状态管理**: React Context + useState/useReducer
- **HTTP客户端**: Axios 1.5.0
- **构建工具**: CRACO 7.1.0

### Infrastructure
- **部署**: Docker + Docker Compose
- **监控**: Prometheus metrics
- **日志**: JSON结构化日志
- **健康检查**: 内置健康检查端点

## Implementation Strategy

### Development Phases
- **阶段1**: 数据格式对齐和API修复（核心问题解决）
- **阶段2**: 前端组件优化和用户体验提升
- **阶段3**: 性能优化和错误处理完善

### Testing Strategy
- **单元测试**: 后端服务层和Repository层测试
- **集成测试**: API接口测试和数据流测试
- **端到端测试**: 使用Chrome DevTools验证完整流程
- **性能测试**: API响应时间和并发测试

### Deployment Strategy
- **开发环境**: 本地Docker容器部署
- **测试环境**: 持续集成和自动化测试
- **生产环境**: 蓝绿部署和回滚机制

## Performance Considerations

### Scalability
- **数据库优化**: 索引优化和查询优化
- **缓存策略**: Redis缓存热点数据
- **分页优化**: 游标分页支持大数据量

### Performance Optimization
- **查询优化**: 减少N+1查询问题
- **响应压缩**: Gzip压缩API响应
- **CDN加速**: 静态资源CDN分发

### Caching Strategy
- **浏览器缓存**: 静态资源长期缓存
- **API缓存**: 不频繁变动的数据缓存
- **数据库缓存**: 查询结果缓存

### Load Balancing
- **水平扩展**: 支持多实例部署
- **负载均衡**: Nginx反向代理
- **健康检查**: 实例健康状态监控

## Security Considerations

### Authentication & Authorization
- **JWT验证**: 无状态Token认证
- **权限控制**: 基于角色的访问控制
- **会话管理**: Token过期和刷新机制

### Data Protection
- **输入验证**: 严格的参数验证
- **SQL注入防护**: 参数化查询
- **XSS防护**: 输出编码和CSP策略

### Security Monitoring
- **审计日志**: 完整的操作记录
- **异常监控**: 异常请求检测
- **安全扫描**: 定期安全漏洞扫描

### Compliance
- **数据保护**: 敏感数据加密存储
- **访问控制**: 最小权限原则
- **合规审计**: 操作审计追踪

## Monitoring & Observability

### Logging Strategy
- **结构化日志**: JSON格式日志
- **日志级别**: DEBUG/INFO/WARN/ERROR
- **上下文追踪**: Request ID关联

### Metrics Collection
- **业务指标**: 案件创建、更新、删除数量
- **性能指标**: API响应时间、错误率
- **系统指标**: CPU、内存、磁盘使用率

### Alerting
- **错误告警**: 异常错误率告警
- **性能告警**: 响应时间超时告警
- **业务告警**: 关键业务异常告警

### Health Checks
- **健康检查**: /health端点
- **依赖检查**: 数据库、Redis连接检查
- **就绪检查**: 服务就绪状态检查

## Error Handling

### Error Scenarios
- **网络错误**: 连接超时、网络中断
- **数据错误**: 数据格式不匹配、验证失败
- **权限错误**: 未授权访问、权限不足
- **系统错误**: 服务不可用、内部错误

### Error Recovery
- **重试机制**: 指数退避重试
- **降级策略**: 核心功能优先保证
- **错误恢复**: 自动错误恢复机制

### User Experience
- **友好提示**: 用户可理解的错误信息
- **操作指导**: 错误解决建议
- **状态反馈**: 实时操作状态反馈

## Quality Assurance

### Code Quality
- **代码规范**: Go官方规范、ESLint/Prettier
- **代码审查**: Pull Request审查机制
- **静态分析**: 代码质量检查工具

### Testing Coverage
- **单元测试覆盖率**: ≥80%
- **集成测试覆盖**: 主要业务流程
- **端到端测试**: 关键用户路径

### Code Review Process
- **技术审查**: 架构和设计合理性
- **安全审查**: 安全漏洞检查
- **性能审查**: 性能影响评估

## Risk Mitigation

### Technical Risks
- **数据不一致**: 强数据一致性验证
- **性能问题**: 性能监控和优化
- **兼容性问题**: 版本兼容性测试

### Operational Risks
- **部署失败**: 蓝绿部署和回滚
- **数据丢失**: 数据备份和恢复
- **服务中断**: 高可用架构设计

### Business Risks
- **用户体验**: 用户反馈收集和改进
- **数据安全**: 安全审计和加固
- **合规风险**: 合规性检查和改进

## Future Considerations

### Extensibility
- **微服务拆分**: 模块化服务设计
- **API版本化**: 向后兼容的API设计
- **插件机制**: 功能扩展插件化

### Maintenance
- **文档维护**: 技术文档和用户文档
- **代码维护**: 重构和优化
- **依赖管理**: 依赖更新和安全补丁

### Upgrade Path
- **数据库升级**: 平滑数据库升级策略
- **框架升级**: 框架版本升级计划
- **架构演进**: 架构现代化演进路径