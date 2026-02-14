# Law OA Go 认证模块重构架构设计

**版本**: 2.0
**日期**: 2026-02-11
**状态**: ✅ 已批准（分阶段实施）
**架构师**: AI Agent (architect-reviewer)

---

## 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| 1.0 | 2026-02-11 | 初始版本 |
| 2.0 | 2026-02-11 | 分阶段实施：Sprint 1 接口统一，Sprint 8 完整功能 |

---

## 一、执行摘要

本文档描述了 Law OA Go 认证模块的重构设计，采用**分阶段实施策略**，确保与 v2.2.0 开发计划的时间线协调。

### 时间线冲突解决

原计划在 Week 2-3 统一 Auth 中间件，但隔离墙 (Sprint 1) 强依赖于它。因此将重构拆分为两部分：

| 阶段 | Sprint | 内容 | 优先级 |
|------|--------|------|--------|
| **Phase 1** | Sprint 1 | 统一 Auth Middleware 接口 | P0 (隔离墙前置依赖) |
| **Phase 2** | Sprint 8 | 完整撤销、设备管理、离职熔断 | P0 |

### 核心目标

1. **Phase 1 (Sprint 1)**: 统一认证接口，为隔离墙提供稳定依赖
2. **Phase 2 (Sprint 8)**: 实现令牌撤销、设备会话、离职熔断
3. **隔离墙集成**: 确保认证中间件与隔离墙中间件正确协同

---

## 二、现有架构分析

### 2.1 当前实现

项目存在两套认证实现：

| 位置 | 描述 | 问题 |
|------|------|------|
| `internal/middleware/jwt.go` | 简单 JWT 中间件 | 无状态，无撤销能力 |
| `internal/auth/` | 增强认证服务 | 未完全集成到路由 |

### 2.2 关键问题

```
┌─────────────────────────────────────────────────────────────────┐
│                        问题诊断                                 │
├─────────────────────────────────────────────────────────────────┤
│ 1. 两套实现并存，维护成本高                                     │
│ 2. 无令牌撤销机制，离职后令牌仍有效                             │
│ 3. 缺少设备会话管理，无法踢出单个设备                           │
│ 4. 审计日志不完整，安全事件难以追溯                             │
│ 5. 与隔离墙中间件集成度低                                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 三、分阶段架构设计

### 3.1 总体架构演进

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        认证架构演进路线                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Sprint 1 (Phase 1)                      Sprint 8 (Phase 2)                  │
│  ┌─────────────────┐                     ┌─────────────────┐                │
│  │  AuthMiddleware │ ───统一接口────────► │  AuthMiddleware │                │
│  │  (简化版)        │                     │  (完整版)        │                │
│  │                 │                     │                 │                │
│  │  • JWT 验证     │                     │  • JWT 验证     │                │
│  │  • 基础中间件   │                     │  • 令牌撤销     │                │
│  │                 │                     │  • 设备管理     │                │
│  └─────────────────┘                     │  • 离职熔断     │                │
│         │                                └─────────────────┘                │
│         ▼                                       │                            │
│  ┌─────────────────┐                           │                            │
│  │  隔离墙中间件    │ ◄──────稳定依赖────────────┘                            │
│  └─────────────────┘                                                        │
│         │                                                                  │
│         ▼                                                                  │
│  ┌─────────────────┐                                                       │
│  │     业务层       │                                                       │
│  └─────────────────┘                                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Phase 1: Sprint 1 接口统一 (AUTH-000)

#### 目标
为隔离墙中间件提供统一的、稳定的认证接口，**暂不实现撤销逻辑**。

#### 接口设计

```go
// internal/auth/middleware.go (统一版本 v2.0 - Sprint 1)

// AuthMiddleware 统一认证中间件 (Phase 1: 简化版)
func AuthMiddleware(config *AuthConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提取令牌
        token := extractToken(c)
        if token == "" {
            c.JSON(401, gin.H{"error": "未提供认证令牌"})
            c.Abort()
            return
        }

        // 2. 解析并验证令牌
        claims, err := validateToken(token, config.SecretKey)
        if err != nil {
            c.JSON(401, gin.H{"error": "令牌无效"})
            c.Abort()
            return
        }

        // 3. 将用户信息注入上下文
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)

        // 4. 设置上下文（隔离墙将使用这些信息）
        c.Set("auth_claims", claims)

        c.Next()
    }
}

// TokenClaims 统一的令牌声明结构
type TokenClaims struct {
    UserID      string `json:"user_id"`
    Username    string `json:"username"`
    Role        string `json:"role"`
    IssuedAt    int64  `json:"iat"`
    ExpiresAt   int64  `json:"exp"`
}

// 隔离墙可使用的辅助函数
func GetUserID(c *gin.Context) string {
    if userID, exists := c.Get("user_id"); exists {
        return userID.(string)
    }
    return ""
}
```

#### Sprint 1 交付物

| 文件 | 内容 |
|------|------|
| `internal/auth/middleware.go` | 统一的 AuthMiddleware |
| `internal/auth/claims.go` | TokenClaims 结构定义 |
| `internal/auth/context.go` | 上下文辅助函数 |
| `internal/auth/token_manager.go` | 简化的令牌管理（仅验证） |

#### 移除的旧代码

- `internal/middleware/jwt.go` → 删除或标记为 deprecated

### 3.3 Phase 2: Sprint 8 完整功能

#### 新增组件

```go
// internal/auth/token_revocation.go (Sprint 8 新增)
type TokenRevocationService struct {
    db    *gorm.DB
    cache *redis.Client
}

// 支持四种撤销策略
type RevocationStrategy int

const (
    RevokeAll        RevocationStrategy = iota // 撤销所有令牌（离职）
    RevokeByUser                               // 撤销用户所有令牌（密码重置）
    RevokeByDevice                             // 撤销设备令牌（安全事件）
    RevokeSingle                               // 撤销单个会话（登出）
)

// internal/auth/device_session.go (Sprint 8 新增)
type DeviceSession struct {
    ID           uint      `gorm:"primaryKey"`
    UserID       string    `gorm:"index"`
    DeviceID     string    `gorm:"index"`
    DeviceName   string    // 设备名称
    DeviceType   string    // mobile, desktop, tablet
    UserAgent    string
    IPAddress    string
    LastSeenAt   time.Time
    CreatedAt    time.Time
    IsActive     bool      `gorm:"index"`
}

// internal/auth/offboarding_auth.go (Sprint 8 新增)
type OffboardingAuthService struct {
    tokenRevocation *TokenRevocationService
    deviceManager   *DeviceSessionManager
    logger          *AuditLogger
}

// 一键离职
func (s *OffboardingAuthService) OffboardUser(ctx context.Context, req OffboardRequest) error {
    // 1. 撤销所有活动令牌
    // 2. 使所有设备会话失效
    // 3. 记录审计日志
    // 4. 触发业务数据移交
}
```

#### Sprint 8 交付物

| 文件 | 内容 |
|------|------|
| `internal/auth/token_revocation.go` | 令牌撤销服务 |
| `internal/auth/device_session.go` | 设备会话管理 |
| `internal/auth/offboarding_auth.go` | 离职认证服务 |
| `internal/auth/middleware_v2.go` | 增强版中间件（含撤销检查） |

---

## 四、数据库变更

### 4.1 Sprint 1 变更

**无数据库变更** - Sprint 1 仅做接口统一。

### 4.2 Sprint 8 变更

```sql
-- migrations/001_schema_v2.2.0.sql (Sprint 8 添加)

-- 令牌撤销表
CREATE TABLE revoked_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(64) NOT NULL,
    jti VARCHAR(64) NOT NULL UNIQUE,
    revoked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    reason VARCHAR(255),
    device_id VARCHAR(64)
);

CREATE INDEX idx_revoked_tokens_user ON revoked_tokens(user_id);
CREATE INDEX idx_revoked_tokens_jti ON revoked_tokens(jti);
CREATE INDEX idx_revoked_tokens_expires ON revoked_tokens(expires_at);

-- 设备会话表
CREATE TABLE device_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    device_name VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45),
    last_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX idx_device_sessions ON device_sessions(user_id, device_id);
CREATE INDEX idx_device_sessions_active ON device_sessions(user_id, is_active);
```

---

## 五、API 接口设计

### 5.1 Sprint 1 接口 (保持不变)

现有的登录/登出接口继续工作，无需变更。

### 5.2 Sprint 8 新增接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/auth/devices | 获取活动设备列表 |
| DELETE | /api/v1/auth/devices/:id | 踢出指定设备 |
| POST | /api/v1/auth/offboard/initiate | 发起离职流程 |
| POST | /api/v1/auth/offboard/complete | 完成离职（撤销所有令牌） |

---

## 六、中间件集成

### 6.1 Sprint 1 集成

```go
// 路由配置 (Sprint 1)
authMiddleware := auth.AuthMiddleware(&auth.AuthConfig{
    SecretKey: config.JWTSecret,
})

// 应用顺序：认证 -> 隔离墙 -> 业务
router.Use(authMiddleware)
router.Use(ethicalWallMiddleware) // EW-003 依赖稳定的 authMiddleware
```

### 6.2 Sprint 8 升级

```go
// 路由配置 (Sprint 8 - 无缝升级)
authMiddleware := auth.AuthMiddlewareV2(&auth.AuthConfigV2{
    SecretKey: config.JWTSecret,
    RevocationService: revocationSvc,  // 新增
    DeviceManager: deviceMgr,          // 新增
})

// 中间件顺序保持不变，内部逻辑增强
router.Use(authMiddleware)
router.Use(ethicalWallMiddleware)
```

---

## 七、实施计划

### 7.1 Sprint 1 (AUTH-000) 任务清单

| 任务ID | 任务 | SP | 依赖 |
|--------|------|----|------|
| AUTH-000-1 | 分析现有两套认证代码 | 1 | - |
| AUTH-000-2 | 设计统一接口规范 | 1 | AUTH-000-1 |
| AUTH-000-3 | 实现 AuthMiddleware 统一版本 | 2 | AUTH-000-2 |
| AUTH-000-4 | 迁移路由到新中间件 | 1 | AUTH-000-3 |
| AUTH-000-5 | 删除/标记废弃旧代码 | 1 | AUTH-000-4 |
| AUTH-000-6 | 编写单元测试 | 1 | AUTH-000-3 |
| AUTH-000-7 | 与隔离墙集成验证 | 2 | AUTH-000-3 |

**总计**: 9 SP

### 7.2 Sprint 8 完整任务清单

| 任务ID | 任务 | SP | 依赖 |
|--------|------|----|------|
| AUTH-001 | 实现令牌撤销服务 | 3 | AUTH-000 |
| AUTH-002 | 实现设备会话管理 | 3 | AUTH-000 |
| AUTH-003 | 实现离职认证服务 | 3 | AUTH-001, AUTH-002 |
| AUTH-004 | 增强中间件（撤销检查） | 2 | AUTH-001 |
| AUTH-005 | 实现设备管理 API | 2 | AUTH-002 |
| AUTH-006 | 前端设备管理界面 | 3 | AUTH-005 |

**总计**: 16 SP (与 Sprint 8 其他任务合并)

---

## 八、风险与回退

| 风险 | 影响 | 缓解措施 | 回退策略 |
|------|------|----------|----------|
| 接口不兼容 | 隔离墙无法工作 | 保持向后兼容 | 恢复旧中间件 |
| Sprint 1 延期 | 隔离墙开发受阻 | 优先级 P0，立即处理 | 临时使用旧实现 |
| Sprint 8 变更破坏稳定性 | 用户体验受影响 | 灰度发布 | 回滚到 V1 中间件 |

---

## 九、审批检查清单

### Sprint 1 (AUTH-000) 审批

- [x] 接口设计符合隔离墙需求
- [x] 向后兼容现有路由
- [x] 不依赖新增数据库表
- [x] 可在 2 周内完成

### Sprint 8 完整功能审批

- [ ] 令牌撤销机制完整
- [ ] 设备管理功能完整
- [ ] 离职熔断与业务流程集成
- [ ] 性能测试通过

---

## 十、附录

### 10.1 关键文件清单

**Sprint 1 (AUTH-000)**:
- `internal/auth/middleware.go` - 统一中间件
- `internal/auth/claims.go` - 令牌声明
- `internal/auth/context.go` - 上下文辅助
- `internal/auth/token_manager.go` - 简化版

**Sprint 8**:
- `internal/auth/token_revocation.go` - 新增
- `internal/auth/device_session.go` - 新增
- `internal/auth/offboarding_auth.go` - 新增
- `migrations/001_schema_v2.2.0.sql` - 添加新表

### 10.2 依赖变更

```go
// Sprint 8 新增依赖
require (
    github.com/go-redis/redis/v9 v9.0.0  // 撤销缓存
    github.com/google/uuid v1.3.0        // 设备ID生成
)
```

---

**文档版本**: 2.0
**创建日期**: 2026-02-11
**创建者**: AI Architect Agent
**状态**: ✅ 已批准（分阶段实施）
**最后更新**: 2026-02-11
