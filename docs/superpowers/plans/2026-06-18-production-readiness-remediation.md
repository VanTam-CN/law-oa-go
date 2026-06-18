# Law OA Go 生产就绪修复实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复审查中列出的 8 个高风险和实际 5 个中风险问题，使系统通过安全、数据一致性、双数据库、前后端构建和发布流水线门禁。

**Architecture:** 先在请求入口实施默认拒绝的认证与授权，再在服务/仓储边界保证资源隔离和资金事务一致性；数据库兼容采用方言中立查询与分方言迁移，工程交付采用可重复构建、真实部署和全量质量门禁。每个任务独立测试、独立提交，不混入现有工作区改动。

**Tech Stack:** Go 1.25、Gin、GORM、MySQL 8、PostgreSQL、Redis、React 18、TypeScript、Vite、ESLint、Docker、GitHub Actions、Kubernetes。

---

## 0. 执行约束与基线

审查标题写“4 个中风险”，正文实际列出 5 个。本计划覆盖正文中的全部 13 项问题。

当前工作区已有未提交改动，至少涉及 `.github/workflows/ci-cd.yml`、冲突检测代码、数据库代码、前端页面和测试。执行者必须先保护这些改动，不得用 `git reset --hard`、`git checkout --`、批量格式化或 `git add .` 覆盖/夹带用户工作。

- [ ] 记录工作区基线：

  ```bash
  git status --short
  git diff --stat
  git diff -- .github/workflows/ci-cd.yml internal/repositories/conflict_repository.go internal/services/conflict_detection_service.go
  ```

  预期：明确哪些差异是执行前已有差异，并保存输出到任务记录，不修改文件。

- [ ] 在用户已提交或明确保存当前改动后，从当前提交创建隔离分支/工作树：

  ```bash
  git switch -c codex/production-readiness-remediation
  ```

  如果当前工作区仍脏，不得自动 stash 或切分用户改动；停止并要求用户先决定如何保存。

- [ ] 记录当前质量基线，但不要把历史失败误判成本任务回归：

  ```bash
  go build ./...
  go test ./...
  go vet ./...
  npm --prefix frontend run type-check
  npm --prefix frontend run lint
  npm --prefix frontend run build
  git diff --check
  ```

  已知基线包括：`internal/testing_disabled` 和 `tests/performance` 编译失败、财务集成测试构造器过期、认证测试未初始化 JWT、若干模型/重试/缓存/Swagger 测试失败、`go vet` 的未导出 JSON 字段和互斥锁复制、前端 142 个 TSConfig 解析错误。

## 合并闸门

| 闸门 | 覆盖任务 | 通过条件 |
|---|---|---|
| A：安全 | 1–4 | 注册不可提权；敏感路由返回正确的 401/403；隔离墙与 OnlyOffice 对错误默认拒绝；SSRF 测试通过 |
| B：数据 | 5–7、9 | 代管款并发审批只生效一次；冲突查询在 MySQL/PostgreSQL 都不漏报；迁移可在两个全新数据库完成；名称短子串不再判 CRITICAL |
| C：发布 | 8、10–12 | `go test ./...`、`go vet ./...`、前端 lint/type-check/build、两个生产镜像和 CI 语法全部通过；流水线包含真实 rollout 与健康检查 |

---

### Task 1：封堵公开注册提权并修复撤销令牌角色读取

**Files:**

- Modify: `internal/handlers/auth_handler.go`
- Modify: `internal/handlers/auth_handler_test.go`
- Modify: `internal/services/user_service.go`
- Modify: `docs/API.md`
- Modify: `docs/openapi.yaml`

**Security contract:**

- 公开注册请求不再接收或信任 `role`，持久化角色固定为 `user`。
- 公开注册生成真实 JWT；不得返回 `simple_token_for_dev`。
- `username` 由规范化邮箱本地部分和规范化完整邮箱的 SHA-256 前 8 位组成，结果截断到 50 字符；同一邮箱生成稳定结果，不留空且避免常见本地部分冲突。
- 管理员创建律师、助理、财务等角色只能走受管理员保护的 `/api/v1/admin/users`。
- 所有撤销令牌处理器统一从 JWT 中间件写入的 `role` 读取角色，最好调用 `middleware.GetCurrentRole`，不再使用 `user_role` 字符串常量。

- [ ] **Step 1：先补失败测试**

  在 `auth_handler_test.go` 增加以下行为测试：

  ```go
  func TestRegisterAlwaysCreatesUnprivilegedUser(t *testing.T) {
      // 请求故意携带 "role":"admin"
      // 捕获传给 UserService/Repository 的用户，断言 Role == "user"
      // 断言响应中的 token 可由 middleware.ValidateToken 解析，claims.Role == "user"
      // 断言 token != "simple_token_for_dev"
  }

  func TestAdminCanRevokeAnotherUsersTokens(t *testing.T) {
      // Gin context 只设置 role=admin，不设置 user_role
      // 断言不是 403，并且 RevokeByUser 被调用一次
  }

  func TestNonAdminCannotRevokeAnotherUsersTokens(t *testing.T) {
      // role=lawyer，current user != target user
      // 断言 403，撤销服务未调用
  }
  ```

- [ ] **Step 2：确认测试先失败**

  ```bash
  go test ./internal/handlers -run 'Test(RegisterAlways|AdminCanRevoke|NonAdminCannotRevoke)' -count=1
  ```

  预期：管理员角色仍被接受、注册 token 仍为假值或撤销返回 403。

- [ ] **Step 3：实施最小修复**

  `RegisterRequest` 删除 `Role` 字段；`Register` 构造服务请求时固定角色，并生成真实 token：

  ```go
  const publicRegistrationRole = "user"

  user, err := h.userService.CreateUser(ctx, &services.CreateUserRequest{
      Username: usernameFromEmail(req.Email),
      Name: req.Name,
      Email: strings.ToLower(strings.TrimSpace(req.Email)),
      Password: req.Password,
      Role: publicRegistrationRole,
      Phone: req.Phone,
  })
  // 错误按现有统一错误映射返回，不暴露底层数据库内容。
  token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
  ```

  对请求中多余的 `role` 字段采用“忽略但永不持久化”，以避免旧前端立即中断；OpenAPI 和示例中删除该字段。若项目决定严格拒绝未知字段，必须同步前端后再开启 `DisallowUnknownFields`，不要只对这一接口做不一致处理。

- [ ] **Step 4：统一角色读取**

  把 `RevokeUserTokens`、`RevokeDeviceTokens`、`RevokeAllTokens` 及同类处理器中的：

  ```go
  userRole, _ := c.Get("user_role")
  ```

  替换为：

  ```go
  role, ok := middleware.GetCurrentRole(c)
  if !ok || (role != "admin" && role != "super_admin") {
      common.APIForbidden(c, "权限不足", "只能操作自己的令牌")
      return
  }
  ```

- [ ] **Step 5：验证并提交**

  ```bash
  go test ./internal/handlers -run 'TestAuthHandler_(Register|Login)|Test(RegisterAlways|AdminCanRevoke|NonAdminCannotRevoke)' -count=1
  go test ./internal/services -run TestUserService -count=1
  git diff --check
  git add internal/handlers/auth_handler.go internal/handlers/auth_handler_test.go internal/services/user_service.go docs/API.md docs/openapi.yaml
  git commit -m "fix(auth): prevent privilege escalation during registration"
  ```

  验收：任何公开注册输入都只能创建 `user`；返回 JWT 真实可验证；管理员跨用户撤销成功，普通用户只能撤销自己。

---

### Task 2：为用户、财务和代管款路由补齐服务端授权

**Files:**

- Modify: `internal/router/router.go`
- Modify: `internal/middleware/admin.go`
- Create: `internal/middleware/authorization_test.go`
- Create: `internal/router/router_authorization_test.go`
- Modify: `frontend/src/utils/accessControl.ts`（仅在角色矩阵不一致时同步）
- Modify: `docs/API.md`

**Authorization matrix:**

| 路由组 | 允许角色 |
|---|---|
| `/api/v1/admin/users/**` | `admin`, `super_admin` |
| `/api/v1/admin/roles/**`, `/admin/permissions/**` | `admin`, `super_admin` |
| `/api/v1/finance/**` | `admin`, `super_admin`, `finance` |
| `/api/v1/trust/**` | `admin`, `super_admin`, `finance` |
| `/api/v1/users/me`, `/users/profile`, 修改自己密码/头像 | 任意已认证用户 |

- [ ] **Step 1：编写表驱动路由授权测试**

  ```go
  tests := []struct {
      role string
      method string
      path string
      want int
  }{
      {"lawyer", http.MethodPost, "/api/v1/admin/users", http.StatusForbidden},
      {"assistant", http.MethodPost, "/api/v1/admin/users/2/roles", http.StatusForbidden},
      {"lawyer", http.MethodPost, "/api/v1/trust/transactions/1/approve", http.StatusForbidden},
      {"finance", http.MethodPost, "/api/v1/trust/transactions/1/approve", http.StatusOK},
      {"finance", http.MethodGet, "/api/v1/finance/overview", http.StatusOK},
      {"admin", http.MethodDelete, "/api/v1/admin/users/2", http.StatusOK},
  }
  ```

  测试处理器只返回 200，避免依赖数据库；测试重点是中间件链是否阻止请求。

- [ ] **Step 2：确认未授权请求当前能到达处理器**

  ```bash
  go test ./internal/router ./internal/middleware -run Authorization -count=1
  ```

- [ ] **Step 3：在路由组统一挂载角色中间件**

  ```go
  users := protected.Group("/admin/users")
  users.Use(middleware.RoleMiddleware("admin", "super_admin"))

  finance := protected.Group("/finance")
  finance.Use(middleware.RoleMiddleware("admin", "super_admin", "finance"))

  trust := protected.Group("/trust")
  trust.Use(middleware.RoleMiddleware("admin", "super_admin", "finance"))
  ```

  不要依赖前端菜单隐藏；不要在每个 handler 重复角色判断。`AdminAuthMiddleware` 与 `RoleMiddleware` 的 401/403 语义应一致：没有认证信息为 401，已认证但角色不允许为 403。

- [ ] **Step 4：验证并提交**

  ```bash
  go test ./internal/router ./internal/middleware -run Authorization -count=1
  go test ./internal/handlers -run 'User|Finance|Trust' -count=1
  git diff --check
  git add internal/router/router.go internal/middleware/admin.go internal/middleware/authorization_test.go internal/router/router_authorization_test.go frontend/src/utils/accessControl.ts docs/API.md
  git commit -m "fix(authz): enforce roles on sensitive route groups"
  ```

  验收：JWT 中角色被篡改或角色不在矩阵中时，敏感处理器不会被调用；服务端矩阵与前端一致。

---

### Task 3：重构隔离墙案件解析并改为默认拒绝

**Files:**

- Modify: `internal/middleware/ethical_wall.go`
- Create: `internal/middleware/ethical_wall_test.go`
- Modify: `internal/repositories/document_repository.go`
- Modify: `internal/repositories/document_repository_impl.go`
- Modify: `internal/services/document_service.go`
- Modify: `internal/services/document_stats_service.go`
- Modify: `internal/handlers/document_handler_enhanced.go`
- Modify: `internal/router/router.go`

**Required behavior:**

- 使用 `c.FullPath()` 判断参数语义，不能把所有 `:id` 都当案件 ID。
- `/documents/:id`、下载、预览、更新、删除通过文档仓储加载 `Document.EntityType/EntityID`；只有 `EntityType == "case"` 时将 `EntityID` 作为案件 ID。
- `/documents/onlyoffice/open|convert` 从 `document_id` 解析文档，并纳入相同检查。
- `/documents` 列表、统计、回收站不能因为没有 `case_id` 而放行受隔离案件数据；查询层必须排除当前用户无权访问的隔离案件文档。
- `IsEthicalWallEnabled`、白名单查询、文档到案件解析失败时返回 503 并中止，不能 fail-open。
- JSON 请求体用 `ShouldBindBodyWith` 或读取后恢复 `c.Request.Body`，不得导致下游 handler 收到空 body。

- [ ] **Step 1：增加隔离墙回归测试**

  至少覆盖：

  ```go
  // 文档 ID=10，EntityType=case，EntityID=99；路由 /documents/10 必须检查 case 99。
  // 文档列表包含普通文档、白名单案件文档、非白名单隔离案件文档；只返回前两类。
  // IsEthicalWallEnabled 返回错误 -> 503，next handler 未执行。
  // IsUserWhitelisted 返回错误 -> 503，next handler 未执行。
  // 无白名单 -> 403；有白名单 -> 200。
  // POST JSON 经中间件后 handler 仍可读取完整 body。
  // /documents/stats/overview 不得把 "overview" 或其他 :id 当案件 ID。
  ```

- [ ] **Step 2：确认测试先失败**

  ```bash
  go test ./internal/middleware ./internal/repositories ./internal/handlers -run 'EthicalWall|DocumentAccessScope' -count=1
  ```

- [ ] **Step 3：实现路由感知的资源解析器**

  将单一 `extractCaseID` 拆成明确分支：

  ```go
  type CaseIDResolver interface {
      ResolveCaseID(ctx context.Context, c *gin.Context) (caseID uint, applies bool, err error)
  }
  ```

  `applies=false` 只表示资源确实不属于案件；解析失败必须返回 `err`。文档资源解析必须走 `DocumentRepository.FindByID`。避免在中间件中复制文档表查询。

- [ ] **Step 4：为列表与统计增加访问范围**

  `DocumentListParams` 和统计查询接收 `ViewerUserID uint`。仓储在所有文档查询上复用以下逻辑等价的 `NOT EXISTS` 范围：

  ```sql
  NOT EXISTS (
    SELECT 1 FROM cases c
    WHERE documents.entity_type = 'case'
      AND c.id = documents.entity_id
      AND c.ethical_wall_enabled = TRUE
      AND NOT EXISTS (
        SELECT 1 FROM case_ethical_wall_whitelists w
        WHERE w.case_id = c.id AND w.user_id = ?
      )
  )
  ```

  所有 `Count`、分组、合规报告、最大文件和列表必须使用同一 scope，避免条数侧信道。查询构造错误直接返回错误，不返回未过滤数据。

- [ ] **Step 5：保护 OnlyOffice 文档入口**

  在 `onlyoffice` 路由组挂载相同隔离墙中间件，并在 handler 中忽略请求体的 `user_id`，使用 `middleware.GetCurrentUserID(c)` 作为实际操作用户。

- [ ] **Step 6：验证并提交**

  ```bash
  go test ./internal/middleware ./internal/repositories ./internal/services ./internal/handlers -run 'EthicalWall|Document' -count=1
  go test -race ./internal/middleware ./internal/services -run 'EthicalWall|Document' -count=1
  git diff --check
  git add internal/middleware/ethical_wall.go internal/middleware/ethical_wall_test.go internal/repositories/document_repository.go internal/repositories/document_repository_impl.go internal/services/document_service.go internal/services/document_stats_service.go internal/handlers/document_handler_enhanced.go internal/router/router.go
  git commit -m "fix(security): fail closed for ethical wall document access"
  ```

  验收：单文档、列表、统计、预览、下载、编辑和转换均无法泄露非白名单隔离案件信息；依赖故障返回 503。

---

### Task 4：强制 OnlyOffice 回调认证并阻断 SSRF/任意覆盖

**Files:**

- Modify: `internal/config/config.go`
- Modify: `internal/handlers/onlyoffice_handler.go`
- Modify: `internal/handlers/onlyoffice_handler_test.go`
- Modify: `internal/router/router.go`
- Modify: `main.go`
- Modify: `cmd/server/main.go`
- Modify: `.env.example`
- Modify: `docs/CONFIGURATION.md`

**Security contract:**

- `ONLYOFFICE_SECRET` 在生产环境必填；缺失时应用启动失败，不能启动一个无认证回调。
- 回调 token 使用 OnlyOffice 支持的 HS256 JWT 语义，验证签名算法、签名、有效期和 key；不使用“整个含 token 的 body 再做 HMAC”的自指方案。
- 下载 URL 只允许 `http/https`，无 userinfo，主机和有效端口必须与配置的 `ONLYOFFICE_URL` 一致；重定向也重新校验或直接禁止。
- 响应体最大 50 MiB，超限不落盘；只在完整下载、备份和临时文件 `fsync/close` 成功后原子替换。
- 数据库元数据更新失败时返回错误，并保留可恢复备份；日志不得包含 token 或完整敏感 URL 查询串。

- [ ] **Step 1：增加认证与 SSRF 测试**

  使用 `httptest.Server` 覆盖：无 secret 的 production 配置启动失败、缺 token 401、错误签名 401、正确签名成功、`127.0.0.1`/云元数据地址/不同端口/不同 host 被拒绝、302 跳向非允许 host 被拒绝、超大响应不覆盖原文件、下载中断不覆盖原文件。

- [ ] **Step 2：确认现状失败**

  ```bash
  go test ./internal/config ./internal/handlers -run 'OnlyOffice|Callback|SSRF' -count=1
  ```

- [ ] **Step 3：将配置纳入正式 Config**

  ```go
  type OnlyOfficeConfig struct {
      URL       string `mapstructure:"url"`
      Secret    string `mapstructure:"secret"`
      BackendURL string `mapstructure:"backend_url"`
  }
  ```

  绑定 `ONLYOFFICE_URL`、`ONLYOFFICE_SECRET`、`BACKEND_URL`。`Config.Validate` 在 production 下拒绝空 secret，并由两个入口将配置显式传给 router/handler。

- [ ] **Step 4：实现受限下载客户端**

  ```go
  client := &http.Client{
      Timeout: 60 * time.Second,
      CheckRedirect: func(req *http.Request, via []*http.Request) error {
          return http.ErrUseLastResponse
      },
  }
  ```

  在每次请求前比较规范化后的 scheme、`Hostname()` 和 `Port()`。不要只做字符串前缀比较，例如 `http://onlyoffice:9090.evil` 必须被拒绝。

- [ ] **Step 5：验证并提交**

  ```bash
  go test ./internal/config ./internal/handlers -run 'OnlyOffice|Callback|SSRF' -count=1
  go test -race ./internal/handlers -run OnlyOffice -count=1
  git diff --check
  git add internal/config/config.go internal/handlers/onlyoffice_handler.go internal/handlers/onlyoffice_handler_test.go internal/router/router.go main.go cmd/server/main.go .env.example docs/CONFIGURATION.md
  git commit -m "fix(onlyoffice): authenticate callbacks and restrict downloads"
  ```

  验收：匿名/伪造回调不能触发网络请求或文件写入；允许 host 的合法签名回调可保存；生产环境空 secret 无法启动。

---

### Task 5：使代管款审批具备事务性、行锁和幂等性

**Files:**

- Modify: `internal/repositories/notification_repository.go`（当前代管款接口定义所在文件）
- Modify: `internal/repositories/trust_account_repository.go`
- Modify: `internal/services/trust_account_service.go`
- Modify: `internal/router/router.go`
- Create: `internal/services/trust_account_service_test.go`
- Create: `tests/integration/trust_transaction_atomicity_test.go`

**Transaction invariant:**

在同一数据库事务中按固定顺序锁定交易行和账户行，重新读取状态与余额，更新余额和交易状态后一起提交。任一步失败全部回滚。重复审批不得再次改变余额。

- [ ] **Step 1：先写服务与集成测试**

  ```go
  func TestApproveTransactionRollsBackWhenTransactionUpdateFails(t *testing.T) {
      // 初始余额 100，withdraw 80；模拟状态更新失败。
      // 断言返回错误、余额仍为 100、交易仍为 pending。
  }

  func TestApproveTransactionConcurrentOnlyAppliesOnce(t *testing.T) {
      // 两个 goroutine 同时审批同一笔 withdraw 80。
      // 断言一个成功、一个得到 ErrTransactionNotPending；最终余额 20。
  }
  ```

  集成测试使用真实 MySQL 8 和 PostgreSQL，不用 SQLite 验证 `FOR UPDATE`。

- [ ] **Step 2：增加事务工作单元与锁定读取**

  ```go
  type TrustUnitOfWork interface {
      WithinTransaction(ctx context.Context, fn func(
          txRepo TrustTransactionRepository,
          accountRepo TrustAccountRepository,
      ) error) error
  }
  ```

  在事务内的 `FindByIDForUpdate` 使用：

  ```go
  db.WithContext(ctx).
      Clauses(clause.Locking{Strength: "UPDATE"}).
      First(&model, id)
  ```

- [ ] **Step 3：重写审批流程**

  顺序固定为：锁交易 → 验证 `pending` → 锁账户 → 验证账户状态和可用余额 → 计算新余额 → 更新账户 → 条件更新交易 `WHERE id=? AND status='pending'` → 检查 `RowsAffected == 1` → 提交。返回结果在提交后重新读取。

- [ ] **Step 4：验证并提交**

  ```bash
  go test ./internal/services ./internal/repositories -run 'Trust.*Approve|Atomicity' -count=1
  go test -race ./internal/services -run Trust -count=1
  go test -tags=integration ./tests/integration -run TrustTransactionAtomicity -count=1
  git diff --check
  git add internal/repositories/notification_repository.go internal/repositories/trust_account_repository.go internal/services/trust_account_service.go internal/router/router.go internal/services/trust_account_service_test.go tests/integration/trust_transaction_atomicity_test.go
  git commit -m "fix(trust): make transaction approval atomic and idempotent"
  ```

  验收：并发审批只扣款一次；余额和状态不可部分提交；MySQL/PostgreSQL 行为一致。

---

### Task 6：修复冲突检测的 MySQL 漏报与错误吞噬

**Files:**

- Modify: `internal/repositories/conflict_repository.go`
- Modify: `internal/services/conflict_detection_service.go`
- Modify: `internal/database/query_builder.go`
- Modify: `internal/repositories/conflict_repository_test.go`
- Modify: `internal/services/conflict_detection_service_test.go`
- Create: `tests/integration/conflict_detection_dialects_test.go`

- [ ] **Step 1：增加双数据库回归测试**

  使用相同 fixture 在 MySQL 8 和 PostgreSQL 插入标题、描述、客户名大小写不同的案件，断言对方当事人查询都命中；将表名临时改错或关闭连接，断言服务返回错误，而不是 `0 conflicts`。

- [ ] **Step 2：替换 PostgreSQL 专属查询**

  把运行时 `ILIKE` 改为两库通用表达式：

  ```sql
  LOWER(c.title) LIKE LOWER(?)
  ```

  `CASE` 相关度和 `WHERE` 条件必须同时改，`internal/database/query_builder.go` 也不能继续生成 `ILIKE`。

- [ ] **Step 3：传播查询和扫描错误**

  将：

  ```go
  func (...) checkOpponentConflicts(...) []*models.ConflictCase
  ```

  改为：

  ```go
  func (...) checkOpponentConflicts(...) ([]*models.ConflictCase, error)
  ```

  `Rows()`、`Scan`、`rows.Err()` 任一失败都包装上下文并返回；上层把检查标记为失败，不保存“无冲突”成功记录。

- [ ] **Step 4：验证并提交**

  ```bash
  go test ./internal/repositories ./internal/services ./internal/database -run Conflict -count=1
  go test -tags=integration ./tests/integration -run ConflictDetectionDialects -count=1
  rg -n '\bILIKE\b' internal --glob '*.go'
  git diff --check
  git add internal/repositories/conflict_repository.go internal/services/conflict_detection_service.go internal/database/query_builder.go internal/repositories/conflict_repository_test.go internal/services/conflict_detection_service_test.go tests/integration/conflict_detection_dialects_test.go
  git commit -m "fix(conflict): use portable queries and propagate database errors"
  ```

  验收：`rg` 不再发现运行时代码生成 `ILIKE`；数据库故障得到 5xx/失败状态，不产生“无冲突”结论。

---

### Task 7：收紧冲突名称匹配，短子串不得升级为 CRITICAL

**Files:**

- Modify: `internal/repositories/conflict_repository.go`
- Modify: `internal/repositories/conflict_repository_test.go`
- Modify: `internal/services/conflict_detection_service.go`
- Modify: `internal/services/conflict_detection_service_test.go`

**Matching policy:**

- 去除大小写、空白和明确公司类型后完全相等：直接命中，可判 `CRITICAL`。
- 单向/双向包含、简称或模糊相似：只能作为候选，最高 `HIGH`，必须人工复核。
- 空值、单字、过短通用词不得命中。
- 不在代码中维护大型公司的硬编码品牌关系来直接判定法律冲突；关联企业应来自客户关系数据。

- [ ] **Step 1：增加边界测试**

  ```go
  // true:  "上海示例科技有限公司" vs "上海示例科技"
  // false: "上海示例科技有限公司" vs "示例"
  // false: "华为技术有限公司" vs "华"
  // false: "北京甲科技有限公司" vs "上海甲贸易有限公司"
  // false: 空字符串或清洗后仅剩公司类型词
  ```

- [ ] **Step 2：实现分类结果，不用一个 bool 承载所有语义**

  ```go
  type PartyMatchKind int
  const (
      PartyNoMatch PartyMatchKind = iota
      PartyCandidateMatch
      PartyExactNormalizedMatch
  )
  ```

  只有 `PartyExactNormalizedMatch` 能把风险提升到 `CRITICAL`。候选匹配保留命中理由和原始名称，风险最高 `HIGH`。

- [ ] **Step 3：验证并提交**

  ```bash
  go test ./internal/repositories ./internal/services -run 'PartyName|DirectOpposing' -count=1
  git diff --check
  git add internal/repositories/conflict_repository.go internal/repositories/conflict_repository_test.go internal/services/conflict_detection_service.go internal/services/conflict_detection_service_test.go
  git commit -m "fix(conflict): prevent critical matches from short substrings"
  ```

  验收：短名称不会直接产生 `CRITICAL`；规范化全称仍可直接命中。

---

### Task 8：修复异步事件 nil 解引用与 goroutine panic

**Files:**

- Modify: `internal/services/event_dispatcher.go`
- Create: `internal/services/event_dispatcher_test.go`

- [ ] **Step 1：增加 panic 与 nil 测试**

  ```go
  func TestDispatchRecoversHandlerPanic(t *testing.T) {
      // 注册会 panic 的 handler，再注册成功 handler。
      // 断言进程不崩溃，成功 handler 仍执行，panic 被记录。
  }

  func TestCaseHandlersRejectNilCase(t *testing.T) {
      // repo 返回 nil, nil；逐个调用涉及 case_ 的默认 handler。
      // 断言返回明确 error，不 panic。
  }
  ```

- [ ] **Step 2：实现统一安全执行器**

  ```go
  func (d *EventDispatcher) runHandler(ctx context.Context, event Event, h EventHandler) {
      defer func() {
          if recovered := recover(); recovered != nil {
              slog.Error("event handler panic", "type", event.Type, "panic", recovered)
          }
      }()
      handlerCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
      defer cancel()
      if err := h(handlerCtx, &event); err != nil {
          slog.Error("event handler failed", "type", event.Type, "error", err)
      }
  }
  ```

  每个 goroutine 获取事件值副本；所有 `FindByID` 后显式检查 `case_ == nil`。不要用 `%w` 包装 nil error。

- [ ] **Step 3：验证并提交**

  ```bash
  go test ./internal/services -run EventDispatcher -count=1
  go test -race ./internal/services -run EventDispatcher -count=10
  git diff --check
  git add internal/services/event_dispatcher.go internal/services/event_dispatcher_test.go
  git commit -m "fix(events): contain handler panics and nil resources"
  ```

  验收：任一事件 handler panic 不会终止进程；请求结束后异步任务有独立超时上下文；nil 资源返回错误。

---

### Task 9：建立真正的 MySQL/PostgreSQL 双方言迁移

**Files:**

- Move: `migrations/*.up.sql`, `migrations/*.down.sql` → `migrations/postgres/`
- Create: `migrations/mysql/` 下与 PostgreSQL 相同版本号的迁移对
- Move: `migrations/001_schema_v2.2.0.sql` → `migrations/legacy/001_schema_v2.2.0.sql`
- Modify: `internal/database/migrator.go`
- Modify: `internal/database/migrator_test.go`
- Modify: `cmd/migrate/main.go`
- Modify: `Makefile`
- Modify: `docker-compose.yml`
- Modify: `.env.example`
- Modify: `docs/CONFIGURATION.md`
- Create: `.github/workflows/migrations.yml`

**Dialect mapping:**

| PostgreSQL | MySQL 8 |
|---|---|
| `SERIAL/BIGSERIAL` | `BIGINT UNSIGNED AUTO_INCREMENT` |
| `UUID`, `uuid_generate_v4()` | `CHAR(36)`, `UUID()` |
| `JSONB`, `::jsonb` | `JSON`, 去除 cast |
| `TIMESTAMPTZ` | `DATETIME(6)`，应用统一 UTC |
| `ON CONFLICT DO NOTHING` | `INSERT IGNORE` |
| `ON CONFLICT ... DO UPDATE` | `ON DUPLICATE KEY UPDATE` |
| PostgreSQL enum/DO block/partial index | `VARCHAR + CHECK` 或普通索引，并显式写 MySQL DDL |

- [ ] **Step 1：先增加迁移路径选择测试**

  `NewMigrator(cfg, "./migrations")` 必须自动选择 `migrations/postgres` 或 `migrations/mysql`；未知 driver 返回错误，不能默认为 MySQL。

- [ ] **Step 2：迁移现有 PostgreSQL 文件并逐版本创建 MySQL 等价文件**

  两个目录的版本集合必须完全相等：

  ```bash
  diff \
    <(find migrations/postgres -name '*.up.sql' -exec basename {} \; | cut -d_ -f1 | sort) \
    <(find migrations/mysql -name '*.up.sql' -exec basename {} \; | cut -d_ -f1 | sort)
  ```

  不得用文本替换机械转换后直接通过；逐个核对主键/外键类型、唯一约束、默认值和 seed 幂等语义。`001_schema_v2.2.0.sql` 不符合 golang-migrate 命名，放入 legacy 并在文档注明不自动执行。

- [ ] **Step 3：统一环境变量名**

  Docker Compose 与 `.env.example` 改用代码实际绑定的 `DB_DRIVER`、`DB_USERNAME`、`DB_PASSWORD`、`DB_DATABASE`；保留旧变量兼容时必须在配置加载层显式映射并标记弃用，不能让 Compose 静默使用默认账号。

- [ ] **Step 4：添加双数据库 CI**

  `migrations.yml` 使用 MySQL 8 和 PostgreSQL 16 两个 job，各自在空库执行：`up` → 检查关键表/约束 → `down` → `up`。只操作 CI 临时数据库。

- [ ] **Step 5：验证并提交**

  ```bash
  go test ./internal/database ./cmd/migrate -count=1
  DB_DRIVER=mysql go run ./cmd/migrate -migrations ./migrations up
  DB_DRIVER=postgres go run ./cmd/migrate -migrations ./migrations up
  git diff --check
  git add migrations internal/database/migrator.go internal/database/migrator_test.go cmd/migrate/main.go Makefile docker-compose.yml .env.example docs/CONFIGURATION.md .github/workflows/migrations.yml
  git commit -m "fix(database): add verified mysql and postgres migrations"
  ```

  验收：两种全新数据库都能完整 up/down/up；版本集合一致；应用配置与 Compose 变量一致。由于这是持久数据边界，真实环境迁移必须另行确认并先备份，本任务只在临时库验证。

---

### Task 10：恢复前端全量 TypeScript 与 ESLint 门禁

**Files:**

- Modify: `frontend/tsconfig.json`
- Modify: `frontend/tsconfig.test.json`
- Modify: `frontend/tsconfig.node.json`
- Modify: `frontend/.eslintrc.cjs`
- Modify: `frontend/package.json`
- Modify: lint/type-check 实际暴露错误涉及的 `frontend/src/**/*.ts(x)`

- [ ] **Step 1：扩大全量源码范围**

  `tsconfig.json` 使用：

  ```json
  {
    "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.d.ts"],
    "exclude": [
      "src/**/*.test.ts", "src/**/*.test.tsx",
      "src/**/*.spec.ts", "src/**/*.spec.tsx",
      "src/**/__tests__/**", "src/test/**"
    ]
  }
  ```

  `tsconfig.test.json` 负责测试和 `src/test/**`；`tsconfig.node.json` 负责 Vite/Jest 等配置文件。

- [ ] **Step 2：让 ESLint 正确选择项目**

  ```js
  parserOptions: {
    project: ['./tsconfig.json', './tsconfig.test.json', './tsconfig.node.json'],
    tsconfigRootDir: __dirname,
  }
  ```

  不得通过移除 `parserOptions.project`、扩大 ignore、关闭核心规则或加批量 disable 注释制造假绿。

- [ ] **Step 3：把 type-check 拆成明确门禁**

  ```json
  {
    "type-check:app": "tsc --noEmit -p tsconfig.json",
    "type-check:test": "tsc --noEmit -p tsconfig.test.json",
    "type-check": "npm run type-check:app && npm run type-check:test"
  }
  ```

- [ ] **Step 4：修复新暴露的真实错误**

  按“类型定义/接口调用/React hooks/未处理 null/测试环境类型”分类逐批修复，每批运行对应命令。不得使用 `any`、`@ts-ignore` 或删除测试绕过；确需例外必须单点说明原因并有测试覆盖。

- [ ] **Step 5：验证并提交**

  ```bash
  npm --prefix frontend ci
  npm --prefix frontend run type-check
  npm --prefix frontend run lint
  npm --prefix frontend run test:ci
  npm --prefix frontend run build
  git diff --check
  git add frontend/tsconfig.json frontend/tsconfig.test.json frontend/tsconfig.node.json frontend/.eslintrc.cjs frontend/package.json frontend/src
  git commit -m "fix(frontend): restore full typecheck and lint coverage"
  ```

  验收：220 个现有 TS/TSX 文件均属于恰当 TSConfig；lint 不再出现 “TSConfig does not include this file”；所有命令退出码为 0。

---

### Task 11：修复后端和前端生产镜像

**Files:**

- Modify: `Dockerfile`
- Modify: `frontend/Dockerfile`
- Modify: `.dockerignore`
- Modify: `frontend/.dockerignore`
- Modify: `docker-compose.yml`

- [ ] **Step 1：增加可重复镜像构建验证脚本/CI 步骤**

  目标命令：

  ```bash
  docker build --target production -t law-oa-go:test .
  docker build --target production -t law-oa-frontend:test ./frontend
  docker image inspect law-oa-go:test --format '{{.Config.User}}'
  docker run --rm --entrypoint /bin/sh law-oa-frontend:test -c 'test -f /usr/share/nginx/html/index.html'
  ```

- [ ] **Step 2：修复后端多阶段构建**

  - 基础 Go 镜像与 `go.mod` 对齐为 Go 1.25。
  - 测试阶段显式 `CGO_ENABLED=1 go test -race`；静态生产构建显式 `CGO_ENABLED=0`。
  - 生产阶段改为可创建目录且支持可靠健康检查的最小 Alpine 运行时，或在 builder 预建目录后复制到 scratch；本项目当前还需要 shell/健康检查，优先 Alpine。
  - 不向生产镜像复制 `.env.example` 为 `.env`。
  - 以非 root 用户运行，预建 `/app/uploads`、`/app/recycle`、`/app/logs`，并验证可写。
  - Docker build 只负责可重复构建；完整测试由前置 CI job 执行，避免同一套测试在镜像内重复且行为不同。

- [ ] **Step 3：修复前端输出路径**

  builder/production 统一使用 Vite 的 `/app/dist`：

  ```dockerfile
  RUN npm run build && test -f /app/dist/index.html
  COPY --from=builder /app/dist /usr/share/nginx/html
  ```

  Node 基础镜像与 CI 对齐为 Node 20；构建参数改用 Vite 可读取的 `VITE_*` 名称，并同步前端取值。

- [ ] **Step 4：验证并提交**

  ```bash
  docker build --target production -t law-oa-go:test .
  docker build --target production -t law-oa-frontend:test ./frontend
  docker compose config
  git diff --check
  git add Dockerfile frontend/Dockerfile .dockerignore frontend/.dockerignore docker-compose.yml
  git commit -m "fix(build): make production images reproducible"
  ```

  验收：两个生产 target 构建成功；后端非 root 且目录可写；前端镜像包含 `dist/index.html`。

---

### Task 12：修复 CI 依赖图并实现真实部署与最终发布门禁

**Files:**

- Modify: `.github/workflows/ci-cd.yml`
- Modify: `scripts/ci-static-analysis.sh`
- Modify: `docs/ci-cd-usage-guide.md`
- Modify: 现有失败测试对应文件（仅修原因，不删、不 skip、不弱化）

**Pipeline policy:**

- `docker-build` 依赖测试/质量门禁，不依赖只在 main 运行的 `pgo-build`；PGO 是 main 的可选优化产物，不阻断 develop/release 镜像。
- develop push 构建并推送后部署 staging；published release/tag 构建并推送后部署 production。
- 使用镜像 digest 部署，不能依赖可变 `latest`。
- 部署必须执行 `kubectl set image`/`helm upgrade`、`rollout status` 和真实 HTTP 健康检查；echo 占位不算部署。
- production GitHub Environment 必须配置人工审批。计划/代码修改不授权触发真实部署。

- [ ] **Step 1：重画 job 依赖**

  ```yaml
  docker-build:
    needs: [integrated-static-analysis, test]
    if: github.event_name == 'push' || github.event_name == 'release'

  deploy-staging:
    needs: [docker-build]
    if: github.event_name == 'push' && github.ref == 'refs/heads/develop'

  deploy-production:
    needs: [docker-build, fuzzing]
    if: github.event_name == 'release' && github.event.action == 'published'
  ```

  给构建步骤设置 `id: build` 并输出 `digest`。release job 删除对不存在的 `pgo-build` artifact 的下载。

- [ ] **Step 2：实现真实 Kubernetes rollout**

  两个 GitHub Environment 分别配置：

  - Secret：`KUBE_CONFIG_STAGING` / `KUBE_CONFIG_PRODUCTION`（base64 kubeconfig）
  - Variable：`KUBE_NAMESPACE`、`KUBE_DEPLOYMENT`、`KUBE_CONTAINER`、`HEALTHCHECK_URL`

  部署核心步骤：

  ```yaml
  - run: echo "${KUBE_CONFIG}" | base64 --decode > "$RUNNER_TEMP/kubeconfig"
  - run: >-
      kubectl --kubeconfig "$RUNNER_TEMP/kubeconfig" -n "$KUBE_NAMESPACE"
      set image "deployment/$KUBE_DEPLOYMENT"
      "$KUBE_CONTAINER=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ needs.docker-build.outputs.digest }}"
  - run: >-
      kubectl --kubeconfig "$RUNNER_TEMP/kubeconfig" -n "$KUBE_NAMESPACE"
      rollout status "deployment/$KUBE_DEPLOYMENT" --timeout=10m
  - run: curl --fail --show-error --retry 12 --retry-delay 5 "$HEALTHCHECK_URL/health/ready"
  ```

  failure 分支收集 `describe` 和最近日志；不要输出 kubeconfig 或 Secret。

- [ ] **Step 3：修复全量 Go 基线，建立硬门禁**

  当前已知失败必须逐类修复根因：

  - 给认证测试初始化隔离 JWT manager，并更新注册契约。
  - 更新 `tests/integration/finance_workflow_test.go` 的服务构造器和过期字段。
  - 修复 `tests/performance` 对已删除 `router.NewRouter/RouterConfig` 的引用。
  - 恢复 `internal/testing_disabled` 缺失类型或将仍有价值的测试迁回实际 package；不得仅加 skip/build tag 隐藏。
  - Redis 相关测试使用 CI Redis 服务，连接失败必须明确，不接受随机 EOF。
  - 修复模型验证、retry 行为、folder fixture 和 Swagger 路由的实际契约差异。
  - 导出 `enhanced_errors.go` 中带 JSON tag 的字段或删除无意义 tag；修复 `content_filter_service.go` 的 mutex 复制。

  每类单独提交；不得把测试预期改成当前错误行为来“修绿”。

- [ ] **Step 4：最终验证**

  ```bash
  mapfile -t changed_go < <(git diff --name-only --diff-filter=ACM "$(git merge-base HEAD origin/main)"..HEAD -- '*.go')
  if ((${#changed_go[@]})); then gofmt -w "${changed_go[@]}"; fi
  test -z "$(gofmt -l .)"
  go build ./...
  go test -race ./...
  go vet ./...
  npm --prefix frontend ci
  npm --prefix frontend run type-check
  npm --prefix frontend run lint
  npm --prefix frontend run test:ci
  npm --prefix frontend run build
  docker build --target production -t law-oa-go:${GITHUB_SHA:-local} .
  docker build --target production -t law-oa-frontend:${GITHUB_SHA:-local} ./frontend
  docker compose config
  git diff --check
  ```

  在 CI 中固定执行 `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7`，并用 workflow 单元检查或测试分支验证 develop、main、release 三种事件的 job 不会因 `needs` 的 skipped 状态被连带跳过。

- [ ] **Step 5：提交 CI 修复**

  ```bash
  git add .github/workflows/ci-cd.yml scripts/ci-static-analysis.sh docs/ci-cd-usage-guide.md
  git commit -m "fix(ci): build and deploy every supported release path"
  ```

  验收：develop 会产生 staging rollout；published release 会产生 production rollout；两者都使用 digest、等待 rollout 并执行健康检查；所有质量命令为绿色。

---

## GLM-5.2 执行协议

将以下指令与本计划文件一起交给 GLM-5.2：

```text
严格按 docs/superpowers/plans/2026-06-18-production-readiness-remediation.md 执行。
一次只处理一个 Task；先写失败测试，再实现，再运行该 Task 的窄验证，再提交。
每个 Task 开始前检查 git status，禁止覆盖执行前已有改动，禁止 git add .。
安全检查失败一律 fail-closed；数据库错误不得转换为“无冲突”；资金更新必须在真实事务和行锁下完成。
不得删除、skip、弱化测试来获得绿色结果，不得新增依赖，除非先说明必要性并获得确认。
Task 9 只允许对临时数据库执行迁移验证；Task 12 只修改流水线，不得触发真实 staging/production 部署。
每完成一个 Task，报告：修改文件、测试命令与结果、剩余风险、提交哈希。任一验收条件未满足，不得进入下一 Task。
```

## 最终验收清单

- [ ] 公开注册即使携带 `role=admin`，数据库和 JWT 仍为 `user`。
- [ ] 普通登录用户无法调用用户管理、财务、代管款审批接口。
- [ ] 文档 ID 正确解析到案件；列表/统计不泄露隔离案件；仓储错误默认拒绝。
- [ ] OnlyOffice 匿名、错误签名、恶意 URL、重定向和超大响应全部被拒绝。
- [ ] 两个并发审批只有一次余额变更；任一写入失败全部回滚。
- [ ] 冲突检测在 MySQL/PostgreSQL 命中一致，查询错误显式失败。
- [ ] 短子串不再生成 `CRITICAL` 直接冲突。
- [ ] 异步事件 nil 与 panic 不会终止进程。
- [ ] 两套迁移可在空库执行 up/down/up。
- [ ] 全量前端源码被 type-check/lint 覆盖。
- [ ] 后端和前端生产镜像可重复构建。
- [ ] develop/release 流水线均能进入真实部署 job，production 有人工审批。
- [ ] `go test -race ./...`、`go vet ./...`、前端质量命令和 `git diff --check` 全部通过。
