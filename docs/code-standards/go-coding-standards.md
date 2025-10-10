# Go 代码规范
## Law OA Go 项目编码标准

**版本**: 1.0  
**创建日期**: 2025-09-30  
**适用项目**: Law OA Go v2.1.0  

---

## 📋 目录

1. [概述](#概述)
2. [基础规范](#基础规范)
3. [命名规范](#命名规范)
4. [代码结构](#代码结构)
5. [错误处理](#错误处理)
6. [并发编程](#并发编程)
7. [性能优化](#性能优化)
8. [安全编程](#安全编程)
9. [测试规范](#测试规范)
10. [文档规范](#文档规范)

---

## 🎯 概述

本文档定义了 Law OA Go 项目的 Go 代码编写标准，旨在确保代码的一致性、可读性、可维护性和安全性。所有开发人员必须遵循这些标准。

### 设计原则

- **简洁性**: 代码应该简洁明了，避免不必要的复杂性
- **一致性**: 整个项目保持统一的编码风格
- **可读性**: 代码应该自解释，易于理解
- **安全性**: 优先考虑安全性，防范常见漏洞
- **性能**: 在保证可读性的前提下优化性能

---

## 🔧 基础规范

### 1. 代码格式化

**必须使用 `gofmt` 和 `goimports` 格式化代码**

```bash
# 格式化所有 Go 文件
gofmt -w .
goimports -w .
```

**配置编辑器自动格式化**
```json
// VS Code settings.json
{
    "go.formatTool": "goimports",
    "editor.formatOnSave": true
}
```

### 2. 导入规范

**导入顺序**：
1. 标准库
2. 第三方库
3. 项目内部包

```go
// ✅ 正确的导入顺序
import (
    "context"
    "fmt"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "gorm.io/gorm"

    "github.com/your-org/law-oa-go/internal/models"
    "github.com/your-org/law-oa-go/internal/services"
)

// ❌ 错误的导入顺序
import (
    "github.com/gin-gonic/gin"
    "fmt"
    "github.com/your-org/law-oa-go/internal/models"
    "time"
)
```

### 3. 行长度限制

- **最大行长度**: 120 字符
- **建议行长度**: 80 字符
- 超长行应适当换行

```go
// ✅ 适当的换行
user, err := s.userService.CreateUser(ctx, &CreateUserRequest{
    Name:     req.Name,
    Email:    req.Email,
    Password: req.Password,
    Role:     req.Role,
})

// ❌ 行过长
user, err := s.userService.CreateUser(ctx, &CreateUserRequest{Name: req.Name, Email: req.Email, Password: req.Password, Role: req.Role})
```

---

## 📝 命名规范

### 1. 包命名

- 使用小写字母
- 简短且有意义
- 避免下划线和驼峰

```go
// ✅ 好的包名
package handlers
package models
package auth

// ❌ 不好的包名
package user_handlers
package UserModels
package authenticationService
```

### 2. 变量命名

**局部变量**：使用驼峰命名法
```go
// ✅ 好的变量名
var userName string
var userCount int
var isActive bool

// ❌ 不好的变量名
var user_name string
var cnt int
var flag bool
```

**全局变量**：使用驼峰命名法，首字母大写（如果需要导出）
```go
// ✅ 导出的全局变量
var DefaultTimeout = 30 * time.Second
var MaxRetryCount = 3

// ✅ 未导出的全局变量
var defaultConfig = &Config{}
```

### 3. 函数命名

**公共函数**：首字母大写，使用驼峰命名法
```go
// ✅ 公共函数
func CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
func ValidateEmail(email string) bool
func GetUserByID(id uint64) (*User, error)
```

**私有函数**：首字母小写，使用驼峰命名法
```go
// ✅ 私有函数
func validateUserInput(req *CreateUserRequest) error
func hashPassword(password string) (string, error)
func generateToken(user *User) (string, error)
```

### 4. 常量命名

**公共常量**：全大写，使用下划线分隔
```go
const (
    MAX_USERNAME_LENGTH = 50
    MIN_PASSWORD_LENGTH = 8
    DEFAULT_PAGE_SIZE   = 20
)
```

**枚举类型**：使用类型前缀
```go
type UserRole string

const (
    UserRoleAdmin   UserRole = "admin"
    UserRoleLawyer  UserRole = "lawyer"
    UserRoleClient  UserRole = "client"
)
```

### 5. 接口命名

- 单方法接口通常以 `-er` 结尾
- 多方法接口使用描述性名称

```go
// ✅ 单方法接口
type Reader interface {
    Read([]byte) (int, error)
}

type UserValidator interface {
    Validate(*User) error
}

// ✅ 多方法接口
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uint64) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uint64) error
}
```

---

## 🏗️ 代码结构

### 1. 文件组织

**目录结构**：
```
internal/
├── handlers/          # HTTP 处理器
├── services/          # 业务逻辑
├── repositories/      # 数据访问
├── models/           # 数据模型
├── middleware/       # 中间件
├── validators/       # 验证器
├── utils/           # 工具函数
└── config/          # 配置
```

### 2. 结构体定义

**字段顺序**：
1. 导出字段（按重要性排序）
2. 未导出字段
3. 嵌入字段放在最前面

```go
// ✅ 好的结构体定义
type User struct {
    // 嵌入字段
    BaseModel

    // 导出字段（按重要性）
    ID       uint64    `json:"id" gorm:"primaryKey"`
    Name     string    `json:"name" gorm:"not null"`
    Email    string    `json:"email" gorm:"uniqueIndex;not null"`
    Role     UserRole  `json:"role" gorm:"not null"`
    IsActive bool      `json:"is_active" gorm:"default:true"`
    
    // 时间字段
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // 未导出字段
    password string `gorm:"not null"`
}
```

### 3. 方法定义

**接收器命名**：使用类型名的首字母（小写）
```go
// ✅ 好的接收器命名
func (u *User) SetPassword(password string) error {
    // 实现
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 实现
}

// ❌ 不好的接收器命名
func (user *User) SetPassword(password string) error {
    // 实现
}
```

### 4. 函数参数

**参数顺序**：
1. `context.Context` (如果需要)
2. 主要参数
3. 选项参数

```go
// ✅ 好的参数顺序
func (s *UserService) GetUsers(ctx context.Context, filter *UserFilter, opts ...GetUsersOption) ([]*User, error) {
    // 实现
}

// ✅ 使用选项模式
type GetUsersOption func(*GetUsersOptions)

func WithPagination(page, size int) GetUsersOption {
    return func(opts *GetUsersOptions) {
        opts.Page = page
        opts.Size = size
    }
}
```

---

## ⚠️ 错误处理

### 1. 错误定义

**使用 `errors` 包创建错误**：
```go
import (
    "errors"
    "fmt"
)

// ✅ 定义错误变量
var (
    ErrUserNotFound     = errors.New("user not found")
    ErrInvalidEmail     = errors.New("invalid email format")
    ErrPasswordTooWeak  = errors.New("password does not meet requirements")
)

// ✅ 定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}
```

### 2. 错误包装

**使用 `fmt.Errorf` 包装错误**：
```go
// ✅ 错误包装
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    if err := s.validateUser(req); err != nil {
        return nil, fmt.Errorf("user validation failed: %w", err)
    }

    user, err := s.userRepo.Create(ctx, req.ToUser())
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}
```

### 3. 错误检查

**使用 `errors.Is` 和 `errors.As`**：
```go
// ✅ 错误检查
func (h *UserHandler) GetUser(c *gin.Context) {
    user, err := h.userService.GetUserByID(ctx, userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        
        var validationErr *ValidationError
        if errors.As(err, &validationErr) {
            c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Message})
            return
        }
        
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }
    
    c.JSON(http.StatusOK, user)
}
```

### 4. 错误日志

**记录错误上下文**：
```go
// ✅ 记录错误上下文
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    logger := s.logger.WithFields(logrus.Fields{
        "operation": "CreateUser",
        "email":     req.Email,
        "role":      req.Role,
    })

    user, err := s.userRepo.Create(ctx, req.ToUser())
    if err != nil {
        logger.WithError(err).Error("Failed to create user in database")
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    logger.Info("User created successfully")
    return user, nil
}
```

---

## 🔄 并发编程

### 1. Goroutine 使用

**避免 Goroutine 泄漏**：
```go
// ✅ 正确的 Goroutine 使用
func (s *NotificationService) SendNotifications(ctx context.Context, users []*User) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(users))
    
    for _, user := range users {
        wg.Add(1)
        go func(u *User) {
            defer wg.Done()
            
            select {
            case <-ctx.Done():
                errChan <- ctx.Err()
                return
            default:
                if err := s.sendNotification(ctx, u); err != nil {
                    errChan <- err
                }
            }
        }(user)
    }
    
    wg.Wait()
    close(errChan)
    
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### 2. Channel 使用

**正确关闭 Channel**：
```go
// ✅ 正确的 Channel 使用
func (s *EventProcessor) ProcessEvents(ctx context.Context) error {
    eventChan := make(chan *Event, 100)
    
    // 启动生产者
    go func() {
        defer close(eventChan)
        for {
            select {
            case <-ctx.Done():
                return
            default:
                event, err := s.getNextEvent(ctx)
                if err != nil {
                    s.logger.WithError(err).Error("Failed to get event")
                    continue
                }
                if event == nil {
                    return
                }
                
                select {
                case eventChan <- event:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    
    // 处理事件
    for event := range eventChan {
        if err := s.processEvent(ctx, event); err != nil {
            s.logger.WithError(err).Error("Failed to process event")
        }
    }
    
    return nil
}
```

### 3. 互斥锁使用

**避免死锁**：
```go
// ✅ 正确的锁使用
type UserCache struct {
    mu    sync.RWMutex
    users map[uint64]*User
}

func (c *UserCache) Get(id uint64) (*User, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    user, exists := c.users[id]
    return user, exists
}

func (c *UserCache) Set(id uint64, user *User) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.users[id] = user
}
```

---

## ⚡ 性能优化

### 1. 内存分配优化

**预分配切片容量**：
```go
// ✅ 预分配容量
func (s *UserService) GetUsersByIDs(ctx context.Context, ids []uint64) ([]*User, error) {
    users := make([]*User, 0, len(ids)) // 预分配容量
    
    for _, id := range ids {
        user, err := s.GetUserByID(ctx, id)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, nil
}

// ❌ 未预分配容量
func (s *UserService) GetUsersByIDs(ctx context.Context, ids []uint64) ([]*User, error) {
    var users []*User // 容量为 0，会多次重新分配
    
    for _, id := range ids {
        user, err := s.GetUserByID(ctx, id)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, nil
}
```

### 2. 字符串拼接优化

**使用 `strings.Builder`**：
```go
// ✅ 使用 strings.Builder
func buildQuery(conditions []string) string {
    var builder strings.Builder
    builder.WriteString("SELECT * FROM users WHERE ")
    
    for i, condition := range conditions {
        if i > 0 {
            builder.WriteString(" AND ")
        }
        builder.WriteString(condition)
    }
    
    return builder.String()
}

// ❌ 使用字符串拼接
func buildQuery(conditions []string) string {
    query := "SELECT * FROM users WHERE "
    
    for i, condition := range conditions {
        if i > 0 {
            query += " AND "
        }
        query += condition
    }
    
    return query
}
```

### 3. 避免不必要的分配

**重用对象**：
```go
// ✅ 使用对象池
var requestPool = sync.Pool{
    New: func() interface{} {
        return &ProcessRequest{}
    },
}

func (s *Service) ProcessData(data []byte) error {
    req := requestPool.Get().(*ProcessRequest)
    defer requestPool.Put(req)
    
    req.Reset()
    req.Data = data
    
    return s.process(req)
}
```

---

## 🔒 安全编程

### 1. 输入验证

**验证所有外部输入**：
```go
// ✅ 输入验证
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 验证邮箱格式
    if !isValidEmail(req.Email) {
        return nil, &ValidationError{
            Field:   "email",
            Message: "invalid email format",
        }
    }
    
    // 验证密码强度
    if !isStrongPassword(req.Password) {
        return nil, &ValidationError{
            Field:   "password",
            Message: "password does not meet security requirements",
        }
    }
    
    // 清理输入
    req.Name = strings.TrimSpace(req.Name)
    req.Email = strings.ToLower(strings.TrimSpace(req.Email))
    
    return s.createUser(ctx, req)
}

func isValidEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}
```

### 2. SQL 注入防护

**使用参数化查询**：
```go
// ✅ 参数化查询
func (r *UserRepository) GetUsersByRole(ctx context.Context, role string) ([]*User, error) {
    var users []*User
    
    err := r.db.WithContext(ctx).
        Where("role = ?", role).
        Find(&users).Error
    
    return users, err
}

// ❌ 字符串拼接（SQL 注入风险）
func (r *UserRepository) GetUsersByRole(ctx context.Context, role string) ([]*User, error) {
    var users []*User
    
    query := fmt.Sprintf("SELECT * FROM users WHERE role = '%s'", role)
    err := r.db.WithContext(ctx).Raw(query).Scan(&users).Error
    
    return users, err
}
```

### 3. 密码处理

**安全的密码处理**：
```go
import "golang.org/x/crypto/bcrypt"

// ✅ 安全的密码哈希
func (s *AuthService) HashPassword(password string) (string, error) {
    // 使用适当的成本参数
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    
    return string(hashedPassword), nil
}

func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ❌ 不安全的密码处理
func (s *AuthService) HashPassword(password string) (string, error) {
    h := sha256.New()
    h.Write([]byte(password))
    return fmt.Sprintf("%x", h.Sum(nil)), nil // 不安全：没有盐值
}
```

### 4. 敏感信息处理

**避免敏感信息泄漏**：
```go
// ✅ 安全的用户结构
type User struct {
    ID       uint64    `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Role     UserRole  `json:"role"`
    password string    `json:"-"` // 不序列化密码
}

func (u *User) ToResponse() *UserResponse {
    return &UserResponse{
        ID:    u.ID,
        Name:  u.Name,
        Email: u.Email,
        Role:  u.Role,
        // 不包含密码等敏感信息
    }
}
```

---

## 🧪 测试规范

### 1. 测试文件命名

**测试文件以 `_test.go` 结尾**：
```
user_service.go      -> user_service_test.go
auth_handler.go      -> auth_handler_test.go
user_repository.go   -> user_repository_test.go
```

### 2. 测试函数命名

**使用描述性的测试名称**：
```go
// ✅ 好的测试名称
func TestUserService_CreateUser_Success(t *testing.T) {}
func TestUserService_CreateUser_InvalidEmail_ReturnsError(t *testing.T) {}
func TestUserService_CreateUser_DuplicateEmail_ReturnsError(t *testing.T) {}

// ❌ 不好的测试名称
func TestCreateUser(t *testing.T) {}
func TestCreateUser2(t *testing.T) {}
```

### 3. 测试结构

**使用表驱动测试**：
```go
func TestUserValidator_ValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {
            name:    "valid email",
            email:   "user@example.com",
            wantErr: false,
        },
        {
            name:    "invalid email - missing @",
            email:   "userexample.com",
            wantErr: true,
        },
        {
            name:    "invalid email - missing domain",
            email:   "user@",
            wantErr: true,
        },
    }

    validator := NewUserValidator()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 4. Mock 使用

**使用接口进行测试**：
```go
// ✅ 可测试的设计
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uint64) (*User, error)
}

type UserService struct {
    userRepo UserRepository
}

// 测试中使用 mock
func TestUserService_CreateUser(t *testing.T) {
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    user, err := service.CreateUser(context.Background(), &CreateUserRequest{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "password123",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    mockRepo.AssertExpectations(t)
}
```

---

## 📚 文档规范

### 1. 包文档

**每个包都应该有文档**：
```go
// Package handlers provides HTTP request handlers for the Law OA Go application.
// It includes handlers for user management, authentication, case management,
// and other core functionalities.
//
// All handlers follow RESTful conventions and return JSON responses.
// Error handling is standardized across all handlers.
package handlers
```

### 2. 函数文档

**公共函数必须有文档**：
```go
// CreateUser creates a new user in the system with the provided information.
// It validates the input, hashes the password, and stores the user in the database.
//
// The function returns an error if:
//   - The email is already in use
//   - The input validation fails
//   - The database operation fails
//
// Example:
//   user, err := service.CreateUser(ctx, &CreateUserRequest{
//       Name:     "John Doe",
//       Email:    "john@example.com",
//       Password: "securePassword123",
//       Role:     UserRoleLawyer,
//   })
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 实现
}
```

### 3. 类型文档

**重要类型需要文档**：
```go
// User represents a user in the Law OA system.
// Users can have different roles (admin, lawyer, client) and different permissions.
type User struct {
    ID       uint64    `json:"id" gorm:"primaryKey"`
    Name     string    `json:"name" gorm:"not null"`
    Email    string    `json:"email" gorm:"uniqueIndex;not null"`
    Role     UserRole  `json:"role" gorm:"not null"`
    IsActive bool      `json:"is_active" gorm:"default:true"`
    
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    password string `gorm:"not null"`
}

// UserRole defines the possible roles for users in the system.
type UserRole string

const (
    // UserRoleAdmin has full system access
    UserRoleAdmin UserRole = "admin"
    // UserRoleLawyer can manage cases and clients
    UserRoleLawyer UserRole = "lawyer"
    // UserRoleClient has limited access to their own data
    UserRoleClient UserRole = "client"
)
```

---

## 🔍 代码审查检查点

### 1. 基础检查
- [ ] 代码已格式化（gofmt, goimports）
- [ ] 导入顺序正确
- [ ] 命名符合规范
- [ ] 没有未使用的变量和导入

### 2. 错误处理
- [ ] 所有错误都被正确处理
- [ ] 错误信息有意义且安全
- [ ] 使用了适当的错误包装
- [ ] 错误日志包含足够的上下文

### 3. 并发安全
- [ ] 正确使用了锁和 channel
- [ ] 没有 goroutine 泄漏
- [ ] 正确处理了 context 取消

### 4. 性能考虑
- [ ] 避免了不必要的内存分配
- [ ] 使用了适当的数据结构
- [ ] 数据库查询已优化

### 5. 安全性
- [ ] 输入已验证和清理
- [ ] 使用了参数化查询
- [ ] 敏感信息不会泄漏
- [ ] 密码安全处理

### 6. 测试
- [ ] 有足够的测试覆盖率
- [ ] 测试名称描述性强
- [ ] 使用了适当的测试技术

---

## 📊 评分标准

### 代码质量评分 (总分 100 分)

| 类别 | 权重 | 评分标准 |
|------|------|----------|
| **代码规范** | 20% | 格式化、命名、结构 |
| **错误处理** | 25% | 错误定义、包装、处理 |
| **安全性** | 25% | 输入验证、SQL 注入防护、密码安全 |
| **性能** | 15% | 内存使用、算法效率、并发安全 |
| **可维护性** | 10% | 代码清晰度、文档完整性 |
| **测试** | 5% | 测试覆盖率、测试质量 |

### 评分等级

- **优秀 (90-100分)**: 完全符合标准，代码质量极高
- **良好 (80-89分)**: 基本符合标准，有少量改进空间
- **一般 (70-79分)**: 部分符合标准，需要改进
- **需要改进 (60-69分)**: 不符合多项标准，需要重构
- **不合格 (<60分)**: 严重不符合标准，需要重写

---

## 🛠️ 工具配置

### 1. golangci-lint 配置

参考项目根目录的 `.golangci.yml` 配置文件。

### 2. 编辑器配置

**VS Code 推荐设置**：
```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.vetOnSave": "package",
    "go.buildOnSave": "package",
    "go.testOnSave": true,
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    }
}
```

### 3. Git Hooks

使用项目提供的 pre-commit hook 确保代码质量。

---

## 📞 支持与反馈

如有疑问或建议，请联系：
- 技术负责人：[技术负责人邮箱]
- 开发团队：[团队邮箱]
- 项目仓库：[GitHub Issues]

---

**文档版本**: 1.0  
**最后更新**: 2025-09-30  
**下次审查**: 2025-12-30