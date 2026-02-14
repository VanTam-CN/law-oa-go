# Law OA Go - Agile Backlog (TDD Mode)

> 基于2026-02-12项目审阅，遵循TDD原则

---

## TDD 工作流

```
┌─────────────────────────────────────────────────────────────┐
│                     TDD Cycle                               │
│                                                             │
│    ┌─────────┐      ┌─────────┐      ┌───────────┐        │
│    │  RED    │ ──►  │  GREEN  │ ──►  │ REFACTOR  │        │
│    │ 写失败   │      │ 写代码   │      │   优化    │        │
│    │  测试   │      │ 使通过   │      │   代码    │        │
│    └─────────┘      └─────────┘      └───────────┘        │
│         ▲                                  │               │
│         └──────────────────────────────────┘               │
└─────────────────────────────────────────────────────────────┘
```

---

## Sprint 规划 (TDD)

| Sprint | 主题 | 重点 | Story Points |
|--------|------|------|--------------|
| Sprint 1 | 测试基础设施 | 搭建TDD环境 | 8 |
| Sprint 2 | 安全测试 | 先写测试再修复 | 21 |
| Sprint 3 | 财务测试 | TDD开发财务模块 | 21 |
| Sprint 4 | 核心测试 | 补充核心测试 | 21 |
| Sprint 5 | 前端测试保护 | 重构前测试覆盖 | 13 |
| Sprint 6 | 前端重构 | 有测试保护的重构 | 13 |
| Sprint 7 | E2E测试 | 端到端验证 | 8 |

---

## 🧪 Epic 0: 测试基础设施 (TDD前置)

### Story 0.1: Go测试框架配置
**优先级**: P0 | **Story Points**: 3

#### TDD Tasks
- [ ] **0.1.1** 配置testcontainers (PostgreSQL/Redis)
- [ ] **0.1.2** 创建测试数据工厂
- [ ] **0.1.3** 配置mock生成工具 (gomock/mockgen)
- [ ] **0.1.4** 创建测试套件基类

#### 验收标准
- [ ] 可快速启动测试容器
- [ ] 测试数据可复用

---

### Story 0.2: 前端测试框架配置
**优先级**: P0 | **Story Points**: 3

#### TDD Tasks
- [ ] **0.2.1** 配置MSW (Mock Service Worker)
- [ ] **0.2.2** 创建测试数据工厂
- [ ] **0.2.3** 配置testing-library最佳实践
- [ ] **0.2.4** 创建测试工具函数

#### 验收标准
- [ ] API调用可mock
- [ ] 组件测试样板统一

---

### Story 0.3: CI测试配置
**优先级**: P0 | **Story Points**: 2

#### TDD Tasks
- [ ] **0.3.1** 配置测试覆盖率报告
- [ ] **0.3.2** 设置覆盖率阈值 (60%)
- [ ] **0.3.3** 添加PR检查

#### 验收标准
- [ ] CI显示覆盖率
- [ ] 低覆盖率阻止合并

---

## 🔴 Epic 1: 安全加固 (TDD模式)

### Story 1.1: JWT注册测试 → 修复
**优先级**: P0 | **Story Points**: 3

#### 🔴 RED Phase - 先写测试
```go
// internal/handlers/auth_handler_test.go

func TestRegister_ShouldReturnValidJWT(t *testing.T) {
    // Arrange
    handler := setupAuthHandler()
    req := RegisterRequest{...}

    // Act
    resp := handler.Register(req)

    // Assert
    assert.NotEmpty(t, resp.Token)
    assert.NotEqual(t, "simple_token_for_dev", resp.Token)

    // Verify token is valid
    claims, err := middleware.VerifyToken(resp.Token)
    assert.NoError(t, err)
    assert.Equal(t, req.Email, claims.Email)
}
```

#### TDD Tasks
- [ ] **1.1.1** 🔴 写失败测试: 注册返回有效JWT
- [ ] **1.1.2** 🔴 写失败测试: Token可被验证
- [ ] **1.1.3** 🟢 修复 `auth_handler.go:117-118`
- [ ] **1.1.4** 🟢 使用与Login相同的JWT生成
- [ ] **1.1.5** 🔵 重构: 提取JWT生成到独立函数

#### 验收标准
- [ ] 测试先写并通过
- [ ] 注册返回真实JWT

---

### Story 1.2: 硬编码密码测试 → 移除
**优先级**: P0 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```go
// internal/config/config_test.go

func TestConfig_ShouldNotHaveHardcodedPassword(t *testing.T) {
    // Test: 配置不应有硬编码默认密码
    cfg := Load()

    // 当密码为空时，应该报错而不是使用默认值
    os.Unsetenv("DB_PASSWORD")

    _, err := Load()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "DB_PASSWORD is required")
}

func TestConfig_ShouldValidateRequiredFields(t *testing.T) {
    // Test: 必填字段验证
    testCases := []struct{
        name string
        env map[string]string
        expectError bool
    }{
        {"missing JWT_SECRET", map[string]string{"JWT_SECRET": ""}, true},
        {"missing DB_PASSWORD", map[string]string{"DB_PASSWORD": ""}, true},
        // ...
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ...
        })
    }
}
```

#### TDD Tasks
- [ ] **1.2.1** 🔴 写失败测试: 空密码应报错
- [ ] **1.2.2** 🔴 写失败测试: 必填字段验证
- [ ] **1.2.3** 🟢 移除 `config.go:214-217` 硬编码
- [ ] **1.2.4** 🟢 添加配置验证函数
- [ ] **1.2.5** 🔵 重构: 配置验证器

#### 验收标准
- [ ] 测试先写并通过
- [ ] 无硬编码密码

---

### Story 1.3: 暴力破解保护测试 → 启用
**优先级**: P0 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```go
// internal/middleware/rate_limit_test.go

func TestBruteForceProtection_ShouldBlockAfterMaxAttempts(t *testing.T) {
    // Test: 超过最大尝试次数应锁定
    middleware := NewBruteForceProtection(5, 15*time.Minute)

    // 模拟5次失败
    for i := 0; i < 5; i++ {
        middleware.RecordFailure("192.168.1.1")
    }

    // 第6次应该被阻止
    blocked := middleware.IsBlocked("192.168.1.1")
    assert.True(t, blocked)
}

func TestBruteForceProtection_ShouldUnblockAfterTimeout(t *testing.T) {
    // Test: 锁定时间后应解锁
    // ...
}
```

#### TDD Tasks
- [ ] **1.3.1** 🔴 写失败测试: 锁定机制
- [ ] **1.3.2** 🔴 写失败测试: 解锁机制
- [ ] **1.3.3** 🟢 在登录路由启用中间件
- [ ] **1.3.4** 🟢 配置参数可调
- [ ] **1.3.5** 🔵 重构: 提取到独立包

#### 验收标准
- [ ] 测试先写并通过
- [ ] 登录有保护

---

### Story 1.4: JWT过期时间测试 → 调整
**优先级**: P0 | **Story Points**: 3

#### 🔴 RED Phase - 先写测试
```go
// internal/middleware/jwt_test.go

func TestJWT_AccessTokenExpiry_ShouldBeLessThan30Minutes(t *testing.T) {
    // Test: Access Token过期时间应≤30分钟
    cfg := LoadTestConfig()

    assert.LessOrEqual(t, cfg.JWT.AccessTokenExpiry, 30*time.Minute)
}

func TestJWT_RefreshToken_ShouldWorkAfterAccessTokenExpiry(t *testing.T) {
    // Test: Refresh Token应在Access Token过期后工作
    // ...
}
```

#### TDD Tasks
- [ ] **1.4.1** 🔴 写失败测试: 过期时间验证
- [ ] **1.4.2** 🔴 写失败测试: 刷新流程
- [ ] **1.4.3** 🟢 调整过期时间
- [ ] **1.4.4** 🟢 配置化
- [ ] **1.4.5** 🔵 重构

#### 验收标准
- [ ] 测试先写并通过
- [ ] Access Token ≤ 30min

---

### Story 1.5: CORS配置测试 → 强化
**优先级**: P1 | **Story Points**: 3

#### 🔴 RED Phase - 先写测试
```go
// internal/middleware/cors_test.go

func TestCORS_ShouldOnlyAllowWhitelistedOrigins(t *testing.T) {
    // Test: 只允许白名单域名
    testCases := []struct{
        origin string
        allowed bool
    }{
        {"https://example.com", true},
        {"https://evil.com", false},
        {"*", false}, // 生产环境不应允许通配符
    }
    // ...
}
```

#### TDD Tasks
- [ ] **1.5.1** 🔴 写失败测试: Origin白名单
- [ ] **1.5.2** 🟢 实现白名单
- [ ] **1.5.3** 🟢 配置化

#### 验收标准
- [ ] 测试先写并通过
- [ ] CORS严格

---

## 🧪 Epic 2: 财务模块测试 (TDD开发)

### Story 2.1: 合同服务TDD
**优先级**: P0 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```go
// internal/services/contract_service_test.go

func TestContractService_Create(t *testing.T) {
    t.Run("should create contract with valid data", func(t *testing.T) {
        // ...
    })

    t.Run("should reject contract with invalid amount", func(t *testing.T) {
        // 金额必须 > 0
    })

    t.Run("should reject contract with past end date", func(t *testing.T) {
        // 结束日期不能在过去
    })

    t.Run("should link contract to case", func(t *testing.T) {
        // 必须关联案件
    })
}

func TestContractService_StatusTransition(t *testing.T) {
    t.Run("draft -> active is valid", func(t *testing.T) {})
    t.Run("active -> completed is valid", func(t *testing.T) {})
    t.Run("draft -> completed is invalid", func(t *testing.T) {})
    t.Run("cancelled contract cannot be reactivated", func(t *testing.T) {})
}
```

#### TDD Tasks
- [ ] **2.1.1** 🔴 写失败测试: Create CRUD
- [ ] **2.1.2** 🔴 写失败测试: 状态转换
- [ ] **2.1.3** 🔴 写失败测试: 业务规则
- [ ] **2.1.4** 🟢 实现Create
- [ ] **2.1.5** 🟢 实现状态机
- [ ] **2.1.6** 🔵 重构

#### 验收标准
- [ ] 测试覆盖 > 80%
- [ ] 所有测试通过

---

### Story 2.2: 发票服务TDD
**优先级**: P0 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```go
// internal/services/invoice_service_test.go

func TestInvoiceService_Create(t *testing.T) {
    t.Run("should generate invoice number automatically", func(t *testing.T) {})
    t.Run("should calculate total from items", func(t *testing.T) {})
    t.Run("should require contract reference", func(t *testing.T) {})
}

func TestInvoiceService_PaymentStatus(t *testing.T) {
    t.Run("unpaid when no payments", func(t *testing.T) {})
    t.Run("partial when payments < total", func(t *testing.T) {})
    t.Run("paid when payments >= total", func(t *testing.T) {})
}
```

#### TDD Tasks
- [ ] **2.2.1** 🔴 写失败测试: CRUD
- [ ] **2.2.2** 🔴 写失败测试: 自动编号
- [ ] **2.2.3** 🔴 写失败测试: 支付状态
- [ ] **2.2.4** 🟢 实现
- [ ] **2.2.5** 🔵 重构

#### 验收标准
- [ ] 测试覆盖 > 80%

---

### Story 2.3: 支付服务TDD
**优先级**: P0 | **Story Points**: 3

#### 🔴 RED Phase - 先写测试
```go
// internal/services/payment_service_test.go

func TestPaymentService_Create(t *testing.T) {
    t.Run("should validate amount > 0", func(t *testing.T) {})
    t.Run("should link to invoice", func(t *testing.T) {})
    t.Run("should update invoice status", func(t *testing.T) {})
}
```

#### TDD Tasks
- [ ] **2.3.1** 🔴 写失败测试
- [ ] **2.3.2** 🟢 实现
- [ ] **2.3.3** 🔵 重构

#### 验收标准
- [ ] 测试覆盖 > 80%

---

### Story 2.4: 佣金服务TDD
**优先级**: P0 | **Story Points**: 3

#### 🔴 RED Phase - 先写测试
```go
// internal/services/commission_service_test.go

func TestCommissionService_Calculate(t *testing.T) {
    t.Run("should calculate based on rate", func(t *testing.T) {})
    t.Run("should handle tiered rates", func(t *testing.T) {})
    t.Run("should generate report", func(t *testing.T) {})
}
```

#### TDD Tasks
- [ ] **2.4.1** 🔴 写失败测试
- [ ] **2.4.2** 🟢 实现
- [ ] **2.4.3** 🔵 重构

#### 验收标准
- [ ] 测试覆盖 > 80%

---

### Story 2.5: 财务Handler TDD
**优先级**: P0 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```go
// internal/handlers/finance_handler_test.go

func TestFinanceHandler_CreateContract(t *testing.T) {
    t.Run("should return 201 on success", func(t *testing.T) {})
    t.Run("should return 400 on invalid input", func(t *testing.T) {})
    t.Run("should return 401 without auth", func(t *testing.T) {})
}

// 类似测试用于 Invoice, Payment 等
```

#### TDD Tasks
- [ ] **2.5.1** 🔴 写失败测试: Contract endpoints
- [ ] **2.5.2** 🔴 写失败测试: Invoice endpoints
- [ ] **2.5.3** 🔴 写失败测试: Payment endpoints
- [ ] **2.5.4** 🟢 实现
- [ ] **2.5.5** 🔵 重构

#### 验收标准
- [ ] Handler测试覆盖 > 80%

---

## 🧪 Epic 3: 核心模块测试 (TDD补充)

### Story 3.1: 通知服务TDD
**优先级**: P1 | **Story Points**: 5

#### TDD Tasks
- [ ] **3.1.1** 🔴 写失败测试: 通知队列
- [ ] **3.1.2** 🔴 写失败测试: 模板管理
- [ ] **3.1.3** 🔴 写失败测试: 发送流程
- [ ] **3.1.4** 🟢 实现
- [ ] **3.1.5** 🔵 重构

#### 验收标准
- [ ] 测试覆盖 > 60%

---

### Story 3.2: 隔离墙TDD
**优先级**: P1 | **Story Points**: 3

#### TDD Tasks
- [ ] **3.2.1** 🔴 写失败测试: 白名单
- [ ] **3.2.2** 🔴 写失败测试: 访问控制
- [ ] **3.2.3** 🔴 写失败测试: 中间件
- [ ] **3.2.4** 🟢 实现
- [ ] **3.2.5** 🔵 重构

#### 验收标准
- [ ] 测试覆盖 > 60%

---

### Story 3.3: 跳过测试修复
**优先级**: P1 | **Story Points**: 5

#### TDD Tasks
- [ ] **3.3.1** 审计所有 `t.Skip()`
- [ ] **3.3.2** 修复: 使用mock替代真实依赖
- [ ] **3.3.3** 验证所有测试通过

#### 验收标准
- [ ] Skip数量 < 10

---

## 💻 Epic 4: 前端测试保护 (重构前置)

### Story 4.1: 状态管理测试
**优先级**: P1 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```typescript
// src/stores/__tests__/appStore.test.ts

describe('useAppStore', () => {
  describe('auth', () => {
    it('should store user info on login', () => {})
    it('should clear user info on logout', () => {})
    it('should check permissions correctly', () => {})
  })

  describe('persistence', () => {
    it('should persist to localStorage', () => {})
    it('should restore from localStorage', () => {})
  })
})
```

#### TDD Tasks
- [ ] **4.1.1** 🔴 写失败测试: Zustand store
- [ ] **4.1.2** 🔴 写失败测试: 权限检查
- [ ] **4.1.3** 🔴 写失败测试: 持久化
- [ ] **4.1.4** 🟢 确保测试通过 (为重构做准备)

#### 验收标准
- [ ] Store测试覆盖 > 80%
- [ ] 为重构提供安全网

---

### Story 4.2: AuthContext组件测试
**优先级**: P1 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```typescript
// src/context/__tests__/AuthContext.test.tsx

describe('AuthContext', () => {
  it('should provide auth state', () => {})
  it('should handle login', () => {})
  it('should handle logout', () => {})
  it('should refresh token', () => {})
  it('should check permissions', () => {})
})
```

#### TDD Tasks
- [ ] **4.2.1** 🔴 写失败测试: 完整覆盖
- [ ] **4.2.2** 🟢 确保测试通过

#### 验收标准
- [ ] AuthContext测试覆盖 > 80%
- [ ] 为迁移提供安全网

---

### Story 4.3: 大组件快照测试
**优先级**: P1 | **Story Points**: 3

#### 🔴 RED Phase - 先写测试
```typescript
// src/pages/case/__tests__/CaseManagement.test.tsx

describe('CaseManagement', () => {
  it('should render table', () => {})
  it('should render filters', () => {})
  it('should handle search', () => {})
  // 快照测试确保重构不破坏UI
  it('should match snapshot', () => {})
})
```

#### TDD Tasks
- [ ] **4.3.1** 🔴 写失败测试: CaseManagement
- [ ] **4.3.2** 🔴 写失败测试: 其他大组件
- [ ] **4.3.3** 🟢 确保测试通过

#### 验收标准
- [ ] 大组件有快照测试
- [ ] 为拆分提供安全网

---

## 💻 Epic 5: 前端重构 (有测试保护)

### Story 5.1: 统一状态管理
**优先级**: P1 | **Story Points**: 8

#### 前置条件: Story 4.1, 4.2 测试通过

#### TDD Tasks
- [ ] **5.1.1** 🔴 写失败测试: 新的统一接口
- [ ] **5.1.2** 🟢 扩展Zustand store
- [ ] **5.1.3** 🟢 迁移AuthContext逻辑
- [ ] **5.1.4** 🔵 重构: 逐个组件迁移
- [ ] **5.1.5** 🔵 删除AuthContext
- [ ] **5.1.6** ✅ 所有测试仍然通过

#### 验收标准
- [ ] 只使用Zustand
- [ ] 所有测试通过 (证明重构安全)

---

### Story 5.2: 拆分CaseManagement
**优先级**: P2 | **Story Points**: 5

#### 前置条件: Story 4.3 测试通过

#### TDD Tasks
- [ ] **5.2.1** 🔴 为子组件写测试
- [ ] **5.2.2** 🟢 提取 CaseTable.tsx
- [ ] **5.2.3** 🟢 提取 CaseFilters.tsx
- [ ] **5.2.4** 🟢 提取 CaseFormModal.tsx
- [ ] **5.2.5** ✅ 快照测试匹配

#### 验收标准
- [ ] 单文件 < 300行
- [ ] 快照测试通过

---

### Story 5.3: 统一API层
**优先级**: P2 | **Story Points**: 5

#### 🔴 RED Phase - 先写测试
```typescript
// src/services/__tests__/api.test.ts

describe('API Layer', () => {
  it('should use configured baseURL', () => {})
  it('should handle errors consistently', () => {})
  it('should add auth header', () => {})
})
```

#### TDD Tasks
- [ ] **5.3.1** 🔴 写失败测试: API配置
- [ ] **5.3.2** 🟢 统一到axios
- [ ] **5.3.3** 🟢 删除重复代码
- [ ] **5.3.4** 🔵 配置化baseURL
- [ ] **5.3.5** ✅ 测试通过

#### 验收标准
- [ ] 只用一个HTTP客户端
- [ ] 测试覆盖API配置

---

## 📋 Epic 6: E2E测试

### Story 6.1: Playwright配置
**优先级**: P2 | **Story Points**: 3

#### TDD Tasks
- [ ] **6.1.1** 配置Playwright
- [ ] **6.1.2** 创建测试工具

---

### Story 6.2: 关键流程E2E
**优先级**: P2 | **Story Points**: 5

#### TDD Tasks
- [ ] **6.2.1** 🔴 写E2E: 登录流程
- [ ] **6.2.2** 🔴 写E2E: 案件创建
- [ ] **6.2.3** 🔴 写E2E: 审批流程
- [ ] **6.2.4** 🔴 写E2E: 财务流程

---

## 📊 TDD Backlog 统计

### 按Sprint

| Sprint | 主题 | 重点 | Points |
|--------|------|------|--------|
| 0 | 测试基础 | TDD环境 | 8 |
| 1 | 安全 | 测试→修复 | 21 |
| 2 | 财务 | TDD开发 | 21 |
| 3 | 核心 | TDD补充 | 13 |
| 4 | 前端测试 | 重构保护 | 13 |
| 5 | 前端重构 | 安全重构 | 18 |
| 6 | E2E | 端到端 | 8 |

### TDD任务分布

| 阶段 | 任务数 | 说明 |
|------|--------|------|
| 🔴 RED | 35 | 先写失败测试 |
| 🟢 GREEN | 30 | 写代码通过 |
| 🔵 REFACTOR | 15 | 优化代码 |
| ✅ VERIFY | 10 | 验证测试通过 |

### TDD检查清单

每个Story必须满足:
- [ ] 先写失败测试 (RED)
- [ ] 最小代码通过 (GREEN)
- [ ] 重构优化 (REFACTOR)
- [ ] 所有测试通过 (VERIFY)

---

## 📝 TDD命令速查

### Go测试
```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -run TestRegister ./internal/handlers/...

# 查看覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 监视模式
go test -v -count=1 -run TestName ./...
```

### 前端测试
```bash
# 运行所有测试
npm test

# 监视模式
npm test -- --watch

# 覆盖率
npm test -- --coverage

# 特定文件
npm test -- CaseManagement.test.tsx
```

---

*生成日期: 2026-02-12*
*遵循TDD原则: Red → Green → Refactor*
