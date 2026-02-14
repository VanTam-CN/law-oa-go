# Law OA Go - Agile Backlog

> 基于2026-02-12项目审阅结果生成的改进任务

---

## Sprint 规划建议

| Sprint | 主题 | 预计Story Points |
|--------|------|------------------|
| Sprint 1 | 安全加固 | 21 |
| Sprint 2 | 测试覆盖-财务模块 | 21 |
| Sprint 3 | 测试覆盖-核心模块 | 21 |
| Sprint 4 | 前端重构-状态管理 | 13 |
| Sprint 5 | 前端重构-组件拆分 | 21 |
| Sprint 6 | 测试覆盖-业务模块 | 13 |

---

## 🔴 Epic 1: 安全加固 (P0)

### Story 1.1: 移除硬编码密码
**优先级**: P0 | **Story Points**: 3 | **模块**: internal/config

#### Tasks
- [ ] **1.1.1** 移除 `config.go` 中的硬编码密码默认值
  - 文件: `internal/config/config.go:214-217`
  - 改为: 强制从环境变量读取，无默认值

- [ ] **1.1.2** 添加配置验证
  - 启动时验证必需配置项
  - 缺失时返回明确错误

- [ ] **1.1.3** 更新 `.env.example` 文档
  - 添加所有必需配置说明

#### 验收标准
- [ ] 无硬编码密码
- [ ] 启动时验证配置完整性
- [ ] 文档更新

---

### Story 1.2: 清理脚本中的凭据
**优先级**: P0 | **Story Points**: 5 | **模块**: scripts/

#### Tasks
- [ ] **1.2.1** 审计所有脚本的硬编码凭据
  - 扫描 `scripts/*.go` 中的数据库连接字符串
  - 生成凭据清单

- [ ] **1.2.2** 创建通用配置加载器
  - 文件: `scripts/common/config.go`
  - 统一从环境变量读取

- [ ] **1.2.3** 重构以下脚本
  - `scripts/update_admin_password.go`
  - `scripts/create_approval_test_data.go`
  - `scripts/create_conflict_test_data.go`
  - `scripts/test_*.go` (约20个文件)

- [ ] **1.2.4** 将脚本目录添加到 `.gitignore` (可选)
  - 或创建 `scripts/.env` 模板

#### 验收标准
- [ ] 所有脚本使用环境变量
- [ ] 无明文密码
- [ ] 脚本文档更新

---

### Story 1.3: 修复JWT注册端点
**优先级**: P0 | **Story Points**: 3 | **模块**: internal/handlers

#### Tasks
- [ ] **1.3.1** 修复 `Register` 方法
  - 文件: `internal/handlers/auth_handler.go:117-118`
  - 使用与 Login 相同的 JWT 生成逻辑

- [ ] **1.3.2** 添加注册流程单元测试
  - 验证返回有效JWT
  - 验证令牌可被验证

- [ ] **1.3.3** 集成测试
  - 完整注册→登录流程

#### 验收标准
- [ ] 注册返回有效JWT
- [ ] 测试覆盖注册流程
- [ ] API文档更新

---

### Story 1.4: 调整JWT过期时间
**优先级**: P0 | **Story Points**: 2 | **模块**: internal/middleware

#### Tasks
- [ ] **1.4.1** 缩短Access Token过期时间
  - 文件: `internal/middleware/jwt.go:71`
  - 从24h改为15-30分钟

- [ ] **1.4.2** 确保Refresh Token机制正常
  - 验证刷新流程
  - 添加刷新端点测试

- [ ] **1.4.3** 配置化过期时间
  - 从配置文件读取
  - 添加环境变量支持

#### 验收标准
- [ ] Access Token ≤ 30分钟
- [ ] Refresh Token 正常工作
- [ ] 配置可调整

---

### Story 1.5: 启用暴力破解保护
**优先级**: P0 | **Story Points**: 3 | **模块**: internal/middleware

#### Tasks
- [ ] **1.5.1** 在登录路由启用 `BruteForceProtectionMiddleware`
  - 文件: `internal/router/router.go`
  - 路由: `/api/v1/auth/login`

- [ ] **1.5.2** 配置保护参数
  - 最大尝试次数: 5
  - 锁定时间: 15分钟
  - 可配置化

- [ ] **1.5.3** 添加测试
  - 验证锁定机制
  - 验证解锁机制

#### 验收标准
- [ ] 登录接口有暴力破解保护
- [ ] 参数可配置
- [ ] 测试覆盖

---

### Story 1.6: 强化CORS配置
**优先级**: P1 | **Story Points**: 2 | **模块**: internal/middleware

#### Tasks
- [ ] **1.6.1** 添加严格的Origin白名单
  - 文件: `internal/middleware/common.go:183-209`
  - 生产环境禁用通配符

- [ ] **1.6.2** 配置化CORS
  - 从环境变量读取允许的域名

- [ ] **1.6.3** 添加安全测试

#### 验收标准
- [ ] CORS配置可配置
- [ ] 生产环境严格模式
- [ ] 测试覆盖

---

### Story 1.7: 修复Redis KEYS命令问题
**优先级**: P1 | **Story Points**: 3 | **模块**: internal/auth

#### Tasks
- [ ] **1.7.1** 使用SCAN替代KEYS
  - 文件: `internal/auth/token_revocation.go:225-227`

- [ ] **1.7.2** 或创建用户设备索引
  - 维护用户→设备的映射

- [ ] **1.7.3** 性能测试

#### 验收标准
- [ ] 不使用KEYS命令
- [ ] 大数据量下性能稳定

---

## 🧪 Epic 2: 测试覆盖 - 财务模块 (P0)

### Story 2.1: 财务Handler测试
**优先级**: P0 | **Story Points**: 5 | **模块**: internal/handlers

#### Tasks
- [ ] **2.1.1** 创建 `finance_handler_test.go`
  - 合同CRUD测试
  - 发票CRUD测试
  - 支付CRUD测试

- [ ] **2.1.2** 创建测试数据工厂
  - Mock合同数据
  - Mock发票数据
  - Mock支付数据

- [ ] **2.1.3** 边界条件测试
  - 金额验证
  - 日期验证
  - 关联验证

#### 验收标准
- [ ] Handler测试覆盖率 > 80%
- [ ] 所有CRUD操作有测试

---

### Story 2.2: 合同服务测试
**优先级**: P0 | **Story Points**: 3 | **模块**: internal/services

#### Tasks
- [ ] **2.2.1** 创建 `contract_service_test.go`
- [ ] **2.2.2** 测试合同状态流转
- [ ] **2.2.3** 测试合同与案件关联

#### 验收标准
- [ ] 服务层测试覆盖 > 70%

---

### Story 2.3: 发票服务测试
**优先级**: P0 | **Story Points**: 3 | **模块**: internal/services

#### Tasks
- [ ] **2.3.1** 创建 `invoice_service_test.go`
- [ ] **2.3.2** 测试发票生成逻辑
- [ ] **2.3.3** 测试发票状态更新

#### 验收标准
- [ ] 服务层测试覆盖 > 70%

---

### Story 2.4: 支付服务测试
**优先级**: P0 | **Story Points**: 3 | **模块**: internal/services

#### Tasks
- [ ] **2.4.1** 创建 `payment_service_test.go`
- [ ] **2.4.2** 测试支付记录创建
- [ ] **2.4.3** 测试支付状态流转

#### 验收标准
- [ ] 服务层测试覆盖 > 70%

---

### Story 2.5: 佣金服务测试
**优先级**: P0 | **Story Points**: 2 | **模块**: internal/services

#### Tasks
- [ ] **2.5.1** 创建 `commission_service_test.go`
- [ ] **2.5.2** 测试佣金计算逻辑
- [ ] **2.5.3** 测试佣金报告生成

#### 验收标准
- [ ] 佣金计算测试覆盖

---

### Story 2.6: 前端财务流程E2E测试
**优先级**: P1 | **Story Points**: 5 | **模块**: frontend/tests

#### Tasks
- [ ] **2.6.1** 完善现有测试
  - 文件: `frontend/tests/integration/FinanceFlow.integration.test.tsx`

- [ ] **2.6.2** 添加合同管理测试
- [ ] **2.6.3** 添加发票管理测试
- [ ] **2.6.4** 添加支付流程测试

#### 验收标准
- [ ] 财务流程E2E测试完整

---

## 🧪 Epic 3: 测试覆盖 - 核心模块 (P1)

### Story 3.1: 通知系统测试
**优先级**: P1 | **Story Points**: 5 | **模块**: internal/services

#### Tasks
- [ ] **3.1.1** 创建 `notification_service_test.go`
- [ ] **3.1.2** 创建 `notification_queue_service_test.go`
- [ ] **3.1.3** 创建 `notification_template_service_test.go`
- [ ] **3.1.4** 创建 `notification_handler_test.go`

#### 验收标准
- [ ] 通知模块测试覆盖 > 60%

---

### Story 3.2: 隔离墙功能测试
**优先级**: P1 | **Story Points**: 3 | **模块**: internal/services

#### Tasks
- [ ] **3.2.1** 创建 `ethical_wall_service_test.go`
- [ ] **3.2.2** 创建 `ethical_wall_handler_test.go`
- [ ] **3.2.3** 测试白名单功能
- [ ] **3.2.4** 测试中间件保护

#### 验收标准
- [ ] 隔离墙测试覆盖 > 60%

---

### Story 3.3: 修复跳过的后端测试
**优先级**: P1 | **Story Points**: 5 | **模块**: internal/

#### Tasks
- [ ] **3.3.1** 审计所有 `t.Skip()` 调用
- [ ] **3.3.2** 修复 `cache_test.go` - Mock Redis
- [ ] **3.3.3** 修复 `case_handler_test.go`
- [ ] **3.3.4** 修复其他跳过测试 (约100处)

#### 验收标准
- [ ] 跳过测试数量 < 10

---

### Story 3.4: 修复跳过的前端测试
**优先级**: P1 | **Story Points**: 5 | **模块**: frontend/tests

#### Tasks
- [ ] **3.4.1** 修复 `CompactCaseForm.integration.test.tsx`
  - 21个test.skip

- [ ] **3.4.2** 修复组件测试
  - `ProgressIndicator.test.tsx` - 12处skip
  - `ConflictCheckInline.test.tsx` - 11处skip
  - `ResponsiveFormLayout.test.tsx` - 7处skip

- [ ] **3.4.3** 完善Mock配置

#### 验收标准
- [ ] 跳过测试数量 < 5

---

### Story 3.5: 缓存测试Mock化
**优先级**: P1 | **Story Points**: 3 | **模块**: internal/cache

#### Tasks
- [ ] **3.5.1** 创建Redis Mock接口
- [ ] **3.5.2** 使用miniredis或go-redis/mock
- [ ] **3.5.3** 重写cache_test.go

#### 验收标准
- [ ] 缓存测试不依赖真实Redis

---

### Story 3.6: 集成测试完善
**优先级**: P1 | **Story Points**: 3 | **模块**: internal/services

#### Tasks
- [ ] **3.6.1** 重写 `integration_test.go`
  - 替换占位符为真实测试

- [ ] **3.6.2** 使用testcontainers
  - PostgreSQL容器
  - Redis容器

#### 验收标准
- [ ] 集成测试有实际验证逻辑

---

## 💻 Epic 4: 前端重构 - 状态管理 (P1)

### Story 4.1: 统一状态管理
**优先级**: P1 | **Story Points**: 8 | **模块**: frontend/src

#### Tasks
- [ ] **4.1.1** 分析当前状态分布
  - `useAppStore` (Zustand)
  - `AuthContext` (Context)

- [ ] **4.1.2** 扩展Zustand Store
  - 合并认证状态
  - 统一权限检查函数

- [ ] **4.1.3** 创建迁移计划
  - 受影响组件清单
  - 迁移顺序

- [ ] **4.1.4** 逐个组件迁移
  - 高优先级组件优先

- [ ] **4.1.5** 移除AuthContext
- [ ] **4.1.6** 更新测试

#### 验收标准
- [ ] 只使用Zustand
- [ ] 无状态重复
- [ ] 测试通过

---

### Story 4.2: 统一localStorage操作
**优先级**: P2 | **Story Points**: 3 | **模块**: frontend/src

#### Tasks
- [ ] **4.2.1** 创建storage工具类
- [ ] **4.2.2** 集中所有localStorage操作
- [ ] **4.2.3** 添加类型安全

#### 验收标准
- [ ] localStorage操作集中管理

---

## 💻 Epic 5: 前端重构 - 组件拆分 (P2)

### Story 5.1: 拆分CaseManagement组件
**优先级**: P2 | **Story Points**: 5 | **模块**: frontend/src/pages/case

#### Tasks
- [ ] **5.1.1** 分析组件结构 (900行)
- [ ] **5.1.2** 提取 `CaseTable.tsx`
- [ ] **5.1.3** 提取 `CaseFilters.tsx`
- [ ] **5.1.4** 提取 `CaseFormModal.tsx`
- [ ] **5.1.5** 更新导入和测试

#### 验收标准
- [ ] 单文件 < 300行
- [ ] 测试通过

---

### Story 5.2: 拆分AuthContext组件
**优先级**: P2 | **Story Points**: 3 | **模块**: frontend/src/context

#### Tasks
- [ ] **5.2.1** 分析组件结构 (624行)
- [ ] **5.2.2** 提取认证逻辑到hooks
- [ ] **5.2.3** 提取权限检查到工具函数

#### 验收标准
- [ ] 单文件 < 300行

---

### Story 5.3: 清理测试页面
**优先级**: P2 | **Story Points**: 2 | **模块**: frontend/src/pages

#### Tasks
- [ ] **5.3.1** 移动测试页面
  - `TestPage.tsx`
  - `MinimalTest.tsx`
  - `SimpleTest.tsx`
  - `ApiTest.tsx`
  - 等

- [ ] **5.3.2** 移动到 `__tests__/pages/` 或删除

#### 验收标准
- [ ] 生产代码无测试页面

---

## 💻 Epic 6: 前端重构 - API层 (P2)

### Story 6.1: 统一HTTP客户端
**优先级**: P2 | **Story Points**: 5 | **模块**: frontend/src/services

#### Tasks
- [ ] **6.1.1** 审计当前HTTP客户端
  - `services/api.ts` (fetch)
  - `services/http.ts` (axios)

- [ ] **6.1.2** 选择统一方案 (建议axios)
- [ ] **6.1.3** 迁移所有fetch调用
- [ ] **6.1.4** 删除重复文件

#### 验收标准
- [ ] 只使用一个HTTP客户端
- [ ] 无重复代码

---

### Story 6.2: 配置化API地址
**优先级**: P2 | **Story Points**: 2 | **模块**: frontend/src

#### Tasks
- [ ] **6.2.1** 创建环境变量配置
  - `.env.development`
  - `.env.production`

- [ ] **6.2.2** 移除硬编码URL
  - `services/api.ts:5`

#### 验收标准
- [ ] API地址可配置

---

### Story 6.3: 统一错误处理
**优先级**: P2 | **Story Points**: 3 | **模块**: frontend/src/services

#### Tasks
- [ ] **6.3.1** 创建统一错误处理器
- [ ] **6.3.2** 标准化错误消息格式
- [ ] **6.3.3** 更新所有服务

#### 验收标准
- [ ] 错误处理一致

---

## 📋 Epic 7: 测试基础设施 (P2)

### Story 7.1: CI测试覆盖率
**优先级**: P2 | **Story Points**: 3 | **模块**: .github

#### Tasks
- [ ] **7.1.1** 配置Go测试覆盖率
- [ ] **7.1.2** 配置前端测试覆盖率
- [ ] **7.1.3** 添加覆盖率阈值 (60%)
- [ ] **7.1.4** CI失败条件

#### 验收标准
- [ ] CI显示覆盖率
- [ ] 低于阈值失败

---

### Story 7.2: E2E测试引入
**优先级**: P2 | **Story Points**: 5 | **模块**: e2e/

#### Tasks
- [ ] **7.2.1** 配置Playwright
- [ ] **7.2.2** 登录流程E2E
- [ ] **7.2.3** 案件创建E2E
- [ ] **7.2.4** 审批流程E2E

#### 验收标准
- [ ] 5-10个关键E2E测试

---

## 📊 Summary

### 按优先级统计

| 优先级 | Story数 | Story Points |
|--------|---------|--------------|
| P0 | 11 | 38 |
| P1 | 11 | 44 |
| P2 | 8 | 28 |
| **总计** | **30** | **110** |

### 按Epic统计

| Epic | Stories | Points |
|------|---------|--------|
| 安全加固 | 7 | 21 |
| 测试-财务 | 6 | 21 |
| 测试-核心 | 6 | 24 |
| 前端-状态 | 2 | 11 |
| 前端-组件 | 3 | 10 |
| 前端-API | 3 | 10 |
| 测试基础 | 2 | 8 |

### 建议Sprint规划

```
Sprint 1 (21 pts): Epic 1 - 安全加固
Sprint 2 (21 pts): Epic 2 - 测试-财务
Sprint 3 (24 pts): Epic 3 - 测试-核心
Sprint 4 (11 pts): Epic 4 - 前端状态管理
Sprint 5 (10 pts): Epic 5 - 前端组件拆分
Sprint 6 (10 pts): Epic 6 - 前端API层
Sprint 7 (8 pts):  Epic 7 - 测试基础设施
```

---

*生成日期: 2026-02-12*
*基于项目审阅报告 v1.0*
