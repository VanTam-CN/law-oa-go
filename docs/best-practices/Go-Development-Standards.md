# Go开发最佳实践指南

**版本**: v2.1.0
**更新日期**: 2025-09-30
**适用项目**: Law OA Go

---

## 📋 概述

本文档基于Go官方规范和Law OA Go项目的实际需求，提供了Go语言开发的最佳实践指南，包括代码规范、性能优化、安全编程、测试策略等方面。

---

## 🎯 核心原则

### 1. 简洁性 (Simplicity)
- 代码应该简单、易读、易理解
- 避免过度设计和复杂的抽象
- 优先选择最简单的解决方案

### 2. 可读性 (Readability)
- 代码是写给人看的，顺便让机器执行
- 使用有意义的命名
- 保持一致的代码风格

### 3. 性能 (Performance)
- 关注关键路径的性能
- 避免不必要的内存分配
- 合理使用并发

### 4. 安全性 (Security)
- 验证所有输入
- 安全地处理错误
- 避免信息泄露

---

## 📝 代码规范

### 1. 命名规范

#### 包命名
```go
// ✅ 好的包名 - 小写，简短，有意义
package user
package auth
package database

// ❌ 避免的包名
package userService
package auth_service
package pkg_database
```

#### 变量和函数命名
```go
// ✅ 好的命名
var userRepository UserRepository
var isUserActive bool
var maxRetryCount int

func GetUserByID(id int64) (*User, error)
func CreateUser(user *User) error
func ValidateEmail(email string) bool

// ❌ 避免的命名
var userRepo UserRepository
var activeFlag bool
var count int

func getUser(id int64) (*User, error)  // 未导出函数使用驼峰命名
func Create(user *User) error         // 函数名不够具体
func Validate(email string) bool      // 函数名过于通用
```

#### 常量命名
```go
// ✅ 好的常量命名
const (
    DefaultPageSize = 10
    MaxPageSize     = 100
    JWTExpiryHours  = 24

    UserRoleAdmin  = "admin"
    UserRoleLawyer = "lawyer"
    UserRoleClient = "client"
)

// ❌ 避免的命名
const (
    PAGE_SIZE = 10
    MAX_PAGE  = 100
    JWT_HOURS = 24
)
```

#### 接口命名
```go
// ✅ 好的接口命名
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int64) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}

type UserService interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
    GetUserProfile(ctx context.Context, userID int64) (*UserProfile, error)
}

// ❌ 避免的命名
type IUserRepository interface {}  // 不需要I前缀
type UserInterface interface {}    // 过于通用
type Repo interface {}             // 名称不够具体
```

### 2. 文件组织

#### 目录结构
```
internal/
├── handlers/          # HTTP处理器
│   ├── user_handler.go
│   ├── case_handler.go
│   └── middleware.go
├── services/          # 业务逻辑
│   ├── user_service.go
│   ├── case_service.go
│   └── auth_service.go
├── repositories/      # 数据访问
│   ├── user_repository.go
│   ├── case_repository.go
│   └── interfaces.go
├── models/           # 数据模型
│   ├── user.go
│   ├── case.go
│   └── client.go
└── config/           # 配置
    ├── config.go
    └── database.go
```

#### 文件命名
```go
// ✅ 好的文件名
user_handler.go
case_service.go
user_repository.go
auth_middleware.go
database_config.go

// ❌ 避免的文件名
UserHandler.go
case-service.go
user_repository_v2.go
auth_middleware_test.go  // 测试文件应在单独目录
```

### 3. 函数设计

#### 函数长度
```go
// ✅ 好的函数 - 单一职责，长度适中
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 验证输入
    if err := s.validateCreateUserRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // 检查用户是否已存在
    if exists, err := s.userRepository.ExistsByEmail(ctx, req.Email); err != nil {
        return nil, fmt.Errorf("check user existence: %w", err)
    } else if exists {
        return nil, ErrUserAlreadyExists
    }

    // 创建用户
    user := &User{
        Username:  req.Username,
        Email:     req.Email,
        Password:  s.hashPassword(req.Password),
        CreatedAt: time.Now(),
    }

    if err := s.userRepository.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    // 清理敏感信息
    user.Password = ""

    return user, nil
}

// ❌ 避免的函数 - 过长，职责不单一
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 超过100行的函数，包含验证、数据库操作、日志记录、
    // 邮件发送、缓存更新等多个职责
}
```

#### 参数数量
```go
// ✅ 好的参数设计 - 使用配置结构体
type CreateUserConfig struct {
    Username  string
    Email     string
    Password  string
    Role      string
    IsActive  bool
    CreatedBy int64
}

func (s *UserService) CreateUser(ctx context.Context, config CreateUserConfig) (*User, error) {
    // 实现逻辑
}

// ❌ 避免的参数设计 - 参数过多
func (s *UserService) CreateUser(
    ctx context.Context,
    username, email, password, role string,
    isActive bool,
    createdBy int64,
) (*User, error) {
    // 实现逻辑
}
```

#### 返回值处理
```go
// ✅ 好的错误处理
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    return &user, nil
}

// ✅ 多返回值示例
func ValidateUserInput(user *User) (bool, []string) {
    var errors []string

    if user.Username == "" {
        errors = append(errors, "username is required")
    }

    if user.Email == "" {
        errors = append(errors, "email is required")
    } else if !isValidEmail(user.Email) {
        errors = append(errors, "invalid email format")
    }

    return len(errors) == 0, errors
}
```

---

## 🚀 性能优化最佳实践

### 1. 内存管理

#### 避免内存泄漏
```go
// ✅ 好的内存管理
func (s *UserService) ProcessUsers(users []*User) error {
    // 使用sync.Pool重用对象
    buffer := s.bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buffer.Reset()
        s.bufferPool.Put(buffer)
    }()

    for _, user := range users {
        // 处理用户数据
        if err := s.processUser(buffer, user); err != nil {
            return err
        }
    }
    return nil
}

// ❌ 可能导致内存泄漏
func (s *UserService) ProcessUsers(users []*User) error {
    var buffer bytes.Buffer  // 在循环中重复使用可能导致内存增长

    for _, user := range users {
        buffer.Reset()  // 不会释放已分配的内存
        // 处理用户数据
    }
    return nil
}
```

#### 字符串优化
```go
// ✅ 高效的字符串操作
func BuildUserMessage(user *User) string {
    var builder strings.Builder
    builder.Grow(64)  // 预分配容量

    builder.WriteString("用户: ")
    builder.WriteString(user.Username)
    builder.WriteString(", 邮箱: ")
    builder.WriteString(user.Email)

    return builder.String()
}

// ❌ 低效的字符串操作
func BuildUserMessage(user *User) string {
    message := "用户: " + user.Username + ", 邮箱: " + user.Email  // 多次字符串分配
    return message
}
```

#### 切片预分配
```go
// ✅ 预分配切片容量
func (r *UserRepository) GetActiveUsers(ctx context.Context) ([]*User, error) {
    // 先查询总数
    count, err := r.db.WithContext(ctx).Model(&User{}).Where("is_active = ?", true).Count(&count).Error
    if err != nil {
        return nil, err
    }

    // 预分配切片容量
    users := make([]*User, 0, count)

    err = r.db.WithContext(ctx).Where("is_active = ?", true).Find(&users).Error
    return users, err
}

// ❌ 未预分配容量
func (r *UserRepository) GetActiveUsers(ctx context.Context) ([]*User, error) {
    var users []*User  // 容量为0，可能导致多次重新分配

    err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&users).Error
    return users, err
}
```

### 2. 数据库操作优化

#### 批量操作
```go
// ✅ 批量插入优化
func (r *UserRepository) CreateUsers(ctx context.Context, users []*User) error {
    if len(users) == 0 {
        return nil
    }

    // 使用批量插入
    err := r.db.WithContext(ctx).CreateInBatches(users, 100).Error
    if err != nil {
        return fmt.Errorf("batch create users: %w", err)
    }

    return nil
}

// ❌ 逐条插入效率低
func (r *UserRepository) CreateUsers(ctx context.Context, users []*User) error {
    for _, user := range users {
        if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
            return err
        }
    }
    return nil
}
```

#### 查询优化
```go
// ✅ 优化的查询 - 避免N+1问题
func (r *CaseRepository) GetCasesWithClients(ctx context.Context, limit, offset int) ([]Case, error) {
    var cases []Case

    // 使用预加载避免N+1查询
    err := r.db.WithContext(ctx).
        Preload("Client").
        Preload("Lawyer").
        Limit(limit).
        Offset(offset).
        Find(&cases).Error

    return cases, err
}

// ✅ 选择性字段查询
func (r *UserRepository) GetUserSummary(ctx context.Context, id int64) (*UserSummary, error) {
    var summary UserSummary

    err := r.db.WithContext(ctx).
        Model(&User{}).
        Select("id", "username", "email", "created_at").
        Where("id = ?", id).
        First(&summary).Error

    return &summary, err
}

// ❌ 可能导致N+1问题的查询
func (r *CaseRepository) GetCasesWithClients(ctx context.Context, limit, offset int) ([]Case, error) {
    var cases []Case

    err := r.db.WithContext(ctx).
        Limit(limit).
        Offset(offset).
        Find(&cases).Error

    if err != nil {
        return nil, err
    }

    // N+1查询问题
    for i := range cases {
        r.db.WithContext(ctx).First(&cases[i].Client, cases[i].ClientID)
        r.db.WithContext(ctx).First(&cases[i].Lawyer, cases[i].LawyerID)
    }

    return cases, nil
}
```

### 3. 并发编程

#### Goroutine管理
```go
// ✅ 使用Worker Pool管理Goroutine
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    workerPool chan chan Job
    quit       chan bool
    wg         sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers:    workers,
        jobQueue:   make(chan Job, 100),
        workerPool: make(chan chan Job, workers),
        quit:       make(chan bool),
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }

    go p.dispatch()
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()

    for {
        select {
        case job := <-p.jobQueue:
            job.Execute()
        case <-p.quit:
            return
        }
    }
}

// ❌ 无限制创建Goroutine
func ProcessDataConcurrently(data []Data) {
    for _, item := range data {
        go process(item)  // 可能创建过多Goroutine
    }
}
```

#### Channel使用
```go
// ✅ 合理的Channel使用
func (s *Service) ProcessDataStream(ctx context.Context, input <-chan Data) <-chan Result {
    output := make(chan Result, 10)  // 缓冲channel避免阻塞

    go func() {
        defer close(output)

        for {
            select {
            case data, ok := <-input:
                if !ok {
                    return
                }

                result := s.processData(data)

                select {
                case output <- result:
                case <-ctx.Done():
                    return
                }

            case <-ctx.Done():
                return
            }
        }
    }()

    return output
}

// ❌ 可能导致死锁的Channel使用
func ProcessData(input chan Data) chan Result {
    output := make(chan Result)  // 无缓冲channel

    go func() {
        for data := range input {
            result := processData(data)
            output <- result  // 如果没有接收者，会死锁
        }
        close(output)
    }()

    return output
}
```

---

## 🔒 安全编程最佳实践

### 1. 输入验证

#### 结构体验证
```go
// ✅ 全面的输入验证
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Phone    string `json:"phone" binding:"omitempty,len=11"`
}

func (r *CreateUserRequest) Validate() error {
    // 自定义验证规则
    if !isValidUsername(r.Username) {
        return fmt.Errorf("invalid username format")
    }

    if !isStrongPassword(r.Password) {
        return fmt.Errorf("password must contain uppercase, lowercase, number and special character")
    }

    if r.Phone != "" && !isValidPhone(r.Phone) {
        return fmt.Errorf("invalid phone number")
    }

    return nil
}

func isValidUsername(username string) bool {
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
    return matched
}

func isStrongPassword(password string) bool {
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

    return hasUpper && hasLower && hasNumber && hasSpecial
}
```

#### SQL注入防护
```go
// ✅ 使用参数化查询
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    var user User

    // 使用GORM的参数化查询，自动防护SQL注入
    err := r.db.WithContext(ctx).
        Where("email = ?", email).  // 参数化查询
        First(&user).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("find user by email: %w", err)
    }

    return &user, nil
}

// ❌ 危险的字符串拼接
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    var user User

    // 危险：容易导致SQL注入
    query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
    err := r.db.WithContext(ctx).Raw(query).Scan(&user).Error

    return &user, err
}
```

### 2. 密码和认证

#### 密码处理
```go
// ✅ 安全的密码处理
type AuthService struct {
    bcryptCost int
    jwtSecret  []byte
}

func (a *AuthService) HashPassword(password string) (string, error) {
    // 使用bcrypt加密密码
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), a.bcryptCost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }

    return string(hashedPassword), nil
}

func (a *AuthService) VerifyPassword(password, hashedPassword string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (a *AuthService) GenerateToken(userID int64, role string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "role":    role,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
        "iat":     time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(a.jwtSecret)
}
```

#### Token验证
```go
// ✅ 安全的Token验证
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // 验证签名方法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return a.jwtSecret, nil
    })

    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        // 检查token是否过期
        if claims.ExpiresAt.Time.Before(time.Now()) {
            return nil, ErrTokenExpired
        }

        return claims, nil
    }

    return nil, ErrInvalidToken
}
```

### 3. 敏感数据处理

#### 日志安全
```go
// ✅ 安全的日志记录
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("创建用户请求解析失败",
            zap.Error(err),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
        return
    }

    // 记录操作日志，但不包含敏感信息
    h.logger.Info("用户注册",
        zap.String("username", req.Username),
        zap.String("email_mask", maskEmail(req.Email)),  // 脱敏处理
        zap.String("client_ip", c.ClientIP()),
    )

    // 业务逻辑处理...
}

func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return email
    }

    username := parts[0]
    domain := parts[1]

    if len(username) <= 3 {
        return email
    }

    masked := username[:2] + "***" + username[len(username)-1:]
    return masked + "@" + domain
}

// ❌ 记录敏感信息
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("创建用户失败",
            zap.Error(err),
            zap.String("password", req.Password),  // 危险：记录了密码
            zap.String("email", req.Email),       // 可能包含敏感信息
        )
        return
    }
}
```

#### 数据脱敏
```go
// ✅ 敏感数据脱敏
type UserResponse struct {
    ID        int64     `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email_mask"`
    Phone     string    `json:"phone_mask"`
    CreatedAt time.Time `json:"created_at"`
}

func (s *UserService) GetUserProfile(ctx context.Context, userID int64) (*UserResponse, error) {
    user, err := s.userRepository.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 脱敏处理
    response := &UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        Email:     maskEmail(user.Email),
        Phone:     maskPhone(user.Phone),
        CreatedAt: user.CreatedAt,
    }

    return response, nil
}

func maskPhone(phone string) string {
    if len(phone) != 11 {
        return phone
    }
    return phone[:3] + "****" + phone[7:]
}
```

---

## 🧪 测试最佳实践

### 1. 单元测试

#### 测试结构
```go
// ✅ 良好的测试结构
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name     string
        req      *CreateUserRequest
        setup    func(*mocks.MockUserRepository)
        wantErr  bool
        wantUser *User
    }{
        {
            name: "成功创建用户",
            req: &CreateUserRequest{
                Username: "testuser",
                Email:    "test@example.com",
                Password: "Password123!",
            },
            setup: func(repo *mocks.MockUserRepository) {
                repo.EXPECT().
                    ExistsByEmail(gomock.Any(), "test@example.com").
                    Return(false, nil)
                repo.EXPECT().
                    Create(gomock.Any(), gomock.Any()).
                    DoAndReturn(func(ctx context.Context, user *User) error {
                        user.ID = 1
                        user.CreatedAt = time.Now()
                        return nil
                    })
            },
            wantErr: false,
            wantUser: &User{
                ID:       1,
                Username: "testuser",
                Email:    "test@example.com",
            },
        },
        {
            name: "用户已存在",
            req: &CreateUserRequest{
                Username: "testuser",
                Email:    "test@example.com",
                Password: "Password123!",
            },
            setup: func(repo *mocks.MockUserRepository) {
                repo.EXPECT().
                    ExistsByEmail(gomock.Any(), "test@example.com").
                    Return(true, nil)
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            repo := mocks.NewMockUserRepository(ctrl)
            tt.setup(repo)

            service := NewUserService(repo, &Config{})

            user, err := service.CreateUser(context.Background(), tt.req)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantUser.ID, user.ID)
                assert.Equal(t, tt.wantUser.Username, user.Username)
                assert.Equal(t, tt.wantUser.Email, user.Email)
                assert.Empty(t, user.Password)  // 密码应被清空
            }
        })
    }
}
```

#### Mock使用
```go
// ✅ 有效的Mock使用
func TestUserHandler_GetUser(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockService := mocks.NewMockUserService(ctrl)
    mockLogger := zaptest.NewLogger(t)

    handler := NewUserHandler(mockService, mockLogger)

    expectedUser := &User{
        ID:       1,
        Username: "testuser",
        Email:    "test@example.com",
    }

    mockService.EXPECT().
        GetUserProfile(gomock.Any(), int64(1)).
        Return(expectedUser, nil)

    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.GET("/users/:id", handler.GetUser)

    req, _ := http.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)

    data := response["data"].(map[string]interface{})
    assert.Equal(t, float64(1), data["id"])
    assert.Equal(t, "testuser", data["username"])
}
```

### 2. 集成测试

#### 数据库集成测试
```go
// ✅ 数据库集成测试
func TestUserRepository_Integration(t *testing.T) {
    // 设置测试数据库
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := NewUserRepository(db)

    t.Run("创建和获取用户", func(t *testing.T) {
        ctx := context.Background()

        // 创建用户
        user := &User{
            Username:  "testuser",
            Email:     "test@example.com",
            Password:  "hashed_password",
            CreatedAt: time.Now(),
        }

        err := repo.Create(ctx, user)
        assert.NoError(t, err)
        assert.NotZero(t, user.ID)

        // 获取用户
        found, err := repo.GetByID(ctx, user.ID)
        assert.NoError(t, err)
        assert.Equal(t, user.Username, found.Username)
        assert.Equal(t, user.Email, found.Email)
    })

    t.Run("查询不存在的用户", func(t *testing.T) {
        ctx := context.Background()

        _, err := repo.GetByID(ctx, 99999)
        assert.Error(t, err)
        assert.True(t, errors.Is(err, ErrUserNotFound))
    })
}

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(mysql.Open("testuser:testpass@tcp(localhost:3306)/law_oa_test?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
    require.NoError(t, err)

    // 自动迁移
    err = db.AutoMigrate(&User{})
    require.NoError(t, err)

    return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
    sqlDB, err := db.DB()
    require.NoError(t, err)

    // 清理测试数据
    db.Exec("DELETE FROM users")

    sqlDB.Close()
}
```

### 3. 性能测试

#### Benchmark测试
```go
// ✅ 性能测试
func BenchmarkUserService_CreateUser(b *testing.B) {
    service := setupUserService(b)

    req := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "Password123!",
    }

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := service.CreateUser(context.Background(), req)
            if err != nil && !errors.Is(err, ErrUserAlreadyExists) {
                b.Fatal(err)
            }
        }
    })
}

func BenchmarkUserRepository_GetByID(b *testing.B) {
    repo := setupUserRepository(b)

    // 预先创建测试数据
    var userIDs []int64
    for i := 0; i < 1000; i++ {
        user := &User{
            Username:  fmt.Sprintf("user%d", i),
            Email:     fmt.Sprintf("user%d@example.com", i),
            CreatedAt: time.Now(),
        }
        err := repo.Create(context.Background(), user)
        if err != nil {
            b.Fatal(err)
        }
        userIDs = append(userIDs, user.ID)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        userID := userIDs[i%len(userIDs)]
        _, err := repo.GetByID(context.Background(), userID)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## 📊 代码质量工具

### 1. 静态分析

#### golangci-lint配置
```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gosec
    - misspell
    - unconvert
    - dupl
    - goconst
    - gocyclo
    - gocritic
    - lll
    - maligned
    - prealloc
    - scopelint
    - gocognit

linters-settings:
  gofmt:
    simplify: true

  goimports:
    local-prefixes: github.com/your-org/law-oa-go

  gocyclo:
    min-complexity: 15

  dupl:
    threshold: 100

  goconst:
    min-len: 3
    min-occurrences: 3

  lll:
    line-length: 120

  prealloc:
    simple: true
    range-loops: true
    for-loops: false

  maligned:
    suggest-new: true

  gosec:
    excludes:
      - G204  # 子进程执行，在测试中可能需要

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
```

### 2. 代码覆盖率

#### 覆盖率配置
```bash
#!/bin/bash
# coverage.sh

# 运行测试并生成覆盖率报告
go test -v -race -coverprofile=coverage.out ./...

# 生成HTML报告
go tool cover -html=coverage.out -o coverage.html

# 显示覆盖率统计
go tool cover -func=coverage.out

# 设置覆盖率目标
COVERAGE_THRESHOLD=70

# 获取总覆盖率
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

# 检查是否达到目标
if (( $(echo "$TOTAL_COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
    echo "✅ 覆盖率 $TOTAL_COVERAGE% 达到目标 $COVERAGE_THRESHOLD%"
else
    echo "❌ 覆盖率 $TOTAL_COVERAGE% 未达到目标 $COVERAGE_THRESHOLD%"
    exit 1
fi
```

---

## 📚 开发工具配置

### 1. VS Code配置

```json
// .vscode/settings.json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "go.testOnSave": true,
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "go.buildTags": "",
    "go.buildFlags": [],
    "go.testFlags": ["-v", "-race"],
    "go.testTimeout": "30s",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    }
}
```

### 2. Git Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "运行代码检查..."

# 格式化检查
if ! gofmt -l . | grep -q .; then
    echo "✅ 代码格式检查通过"
else
    echo "❌ 代码格式检查失败，请运行 gofmt -w ."
    exit 1
fi

# 静态分析
if ! golangci-lint run; then
    echo "❌ 静态分析检查失败"
    exit 1
fi

# 运行测试
if ! go test -race ./...; then
    echo "❌ 测试失败"
    exit 1
fi

# 检查覆盖率
COVERAGE=$(go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print $3}' | sed 's/%//' || echo "0")
if (( $(echo "$COVERAGE < 70" | bc -l) )); then
    echo "❌ 测试覆盖率 $COVERAGE% 低于要求 70%"
    exit 1
fi

echo "✅ 所有检查通过"
```

---

## 🎯 最佳实践检查清单

### 代码质量
- [ ] 代码遵循Go官方规范
- [ ] 使用有意义的命名
- [ ] 函数职责单一，长度适中
- [ ] 避免代码重复
- [ ] 添加适当的注释和文档

### 性能优化
- [ ] 避免内存泄漏
- [ ] 合理使用并发
- [ ] 优化数据库查询
- [ ] 预分配切片和映射容量
- [ ] 使用对象池重用对象

### 安全编程
- [ ] 验证所有输入
- [ ] 安全处理密码
- [ ] 防护SQL注入
- [ ] 日志不包含敏感信息
- [ ] 使用HTTPS传输

### 测试覆盖
- [ ] 单元测试覆盖核心逻辑
- [ ] 集成测试验证数据流
- [ ] 性能测试确保性能指标
- [ ] 测试覆盖率达到70%以上
- [ ] Mock外部依赖

### 工具配置
- [ ] 配置静态分析工具
- [ ] 设置代码格式化
- [ ] 配置Git hooks
- [ ] 集成CI/CD检查
- [ ] 监控代码质量指标

---

**文档版本**: v2.1.0
**最后更新**: 2025-09-30
**下次审查**: 2025-12-30
**维护团队**: 开发团队