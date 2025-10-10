# Law OA Go 代码贡献流程指南

**版本**: v2.1.0
**更新日期**: 2025-09-30
**适用范围**: 所有项目贡献者

---

## 📋 概述

本文档定义了Law OA Go项目的代码贡献流程，包括开发流程、代码审查、提交规范、分支管理等标准。所有贡献者都必须遵循此流程以确保代码质量和项目一致性。

---

## 🎯 贡献原则

### 1. 质量优先
- 代码必须通过所有测试
- 遵循项目编码规范
- 提供必要的文档和注释

### 2. 协作友好
- 保持提交历史清晰
- 编写有意义的提交信息
- 及时响应代码审查反馈

### 3. 渐进改进
- 小步快跑，频繁提交
- 每个功能独立开发
- 避免大规模重构

---

## 🌳 分支策略

### 分支类型

#### main 分支
- **用途**: 生产环境代码
- **保护**: 只能通过 Pull Request 合并
- **要求**: 必须通过所有 CI 检查
- **更新**: 定期从 release 分支合并

#### develop 分支
- **用途**: 开发环境代码
- **保护**: 只能通过 Pull Request 合并
- **要求**: 必须通过所有测试
- **更新**: 功能分支合并到此

#### feature/* 分支
- **用途**: 新功能开发
- **命名**: `feature/功能描述`
- **来源**: 从 develop 分支创建
- **合并**: 完成后合并到 develop

#### bugfix/* 分支
- **用途**: 错误修复
- **命名**: `bugfix/问题描述`
- **来源**: 从 develop 分支创建
- **合并**: 修复后合并到 develop

#### hotfix/* 分支
- **用途**: 紧急生产修复
- **命名**: `hotfix/修复描述`
- **来源**: 从 main 分支创建
- **合并**: 修复后同时合并到 main 和 develop

#### release/* 分支
- **用途**: 发布准备
- **命名**: `release/版本号`
- **来源**: 从 develop 分支创建
- **合并**: 完成后合并到 main 和 develop

### 分支命名规范

```bash
# 功能分支
feature/user-authentication
feature/case-management-system
feature/email-notifications

# 修复分支
bugfix/login-validation-error
bugfix/database-connection-leak
bugfix/memory-leak-in-reporting

# 热修复分支
hotfix/security-vulnerability-fix
hotfix/critical-production-bug

# 发布分支
release/v2.1.0
release/v2.2.0-beta
```

---

## 🔄 开发流程

### 1. 准备工作

#### 获取最新代码
```bash
# 切换到 develop 分支
git checkout develop

# 拉取最新代码
git pull origin develop

# 确保本地是最新的
git status
```

#### 创建功能分支
```bash
# 创建新分支
git checkout -b feature/user-profile-management

# 或者使用 git switch (Git 2.23+)
git switch -c feature/user-profile-management

# 推送分支到远程
git push -u origin feature/user-profile-management
```

#### 设置开发环境
```bash
# 安装依赖
go mod download

# 运行数据库迁移
make migrate-up

# 启动开发服务器
make dev

# 验证环境
curl http://localhost:8080/health
```

### 2. 开发阶段

#### 编写代码
```bash
# 创建新文件
touch internal/services/user_profile_service.go

# 编辑现有文件
vim internal/handlers/user_handler.go

# 添加测试文件
touch internal/services/user_profile_service_test.go
```

#### 编写测试
```go
// 示例测试文件
package services

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserProfileService_GetProfile(t *testing.T) {
    // 测试逻辑
}
```

#### 运行测试
```bash
# 运行所有测试
make test

# 运行特定测试
go test -v ./internal/services

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 代码检查
```bash
# 代码格式化
gofmt -w .

# 静态分析
make lint

# 安全检查
gosec ./...

# 代码审查
golangci-lint run
```

### 3. 提交代码

#### 暂存更改
```bash
# 查看更改状态
git status

# 暂存特定文件
git add internal/services/user_profile_service.go
git add internal/handlers/user_handler.go

# 暂存所有更改
git add .

# 交互式暂存
git add -p
```

#### 提交代码
```bash
# 提交暂存的更改
git commit -m "feat: add user profile management service

- Implement user profile CRUD operations
- Add input validation for profile updates
- Include comprehensive unit tests
- Update API documentation

Closes #123"

# 或者使用更详细的多行提交信息
git commit
```

#### 推送代码
```bash
# 推送到远程分支
git push origin feature/user-profile-management

# 强制推送（谨慎使用）
git push --force-with-lease origin feature/user-profile-management
```

### 4. 创建 Pull Request

#### 准备 Pull Request
```bash
# 确保分支是最新的
git fetch origin
git rebase origin/develop

# 解决冲突（如果有）
# 手动编辑冲突文件
git add .
git rebase --continue

# 推送更新
git push origin feature/user-profile-management
```

#### 创建 Pull Request (GitHub)
1. 访问 GitHub 项目页面
2. 点击 "New pull request"
3. 选择正确的分支：
   - Base: `develop`
   - Head: `feature/user-profile-management`
4. 填写 PR 模板
5. 添加审查者
6. 提交 PR

#### PR 模板
```markdown
## 📝 变更描述
简要描述本次变更的内容和目的。

## 🎯 变更类型
- [ ] 新功能 (feature)
- [ ] 错误修复 (bugfix)
- [ ] 文档更新 (docs)
- [ ] 样式调整 (style)
- [ ] 重构代码 (refactor)
- [ ] 性能优化 (perf)
- [ ] 测试相关 (test)
- [ ] 构建相关 (build)
- [ ] CI/CD 相关 (ci)

## 🧪 测试
- [ ] 单元测试已通过
- [ ] 集成测试已通过
- [ ] 手动测试已完成
- [ ] 测试覆盖率达标

## 📋 检查清单
- [ ] 代码遵循项目规范
- [ ] 已添加必要的测试
- [ ] 文档已更新
- [ ] 无安全漏洞
- [ ] 性能影响可接受

## 🔗 相关 Issue
Closes #(issue number)

## 📸 截图（如适用）
添加相关的截图或 GIF。

## 💬 其他说明
其他需要说明的内容。
```

---

## 📝 提交规范

### 提交信息格式

#### Conventional Commits 规范
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### 提交类型
- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试相关
- `build`: 构建系统或依赖更新
- `ci`: CI/CD 配置更新
- `chore`: 其他不重要的修改

#### 提交示例
```bash
# 好的提交信息
git commit -m "feat(auth): add JWT token refresh mechanism

- Implement automatic token refresh
- Add refresh token storage in Redis
- Update middleware to handle expired tokens
- Add comprehensive tests for token refresh flow

Closes #456"

# 修复提交
git commit -m "fix(database): resolve connection pool exhaustion

- Add connection timeout configuration
- Implement connection health checks
- Add proper connection cleanup in shutdown"

# 文档提交
git commit -m "docs: update API documentation for user endpoints

- Add request/response examples
- Update authentication requirements
- Fix typos in endpoint descriptions"

# 小的修复
git commit -m "fix: correct typo in error message"
```

### 提交频率

#### 推荐频率
- **小功能**: 每天提交 2-5 次
- **大功能**: 每天提交 1-3 次
- **Bug 修复**: 每天提交 3-10 次

#### 提交时机
```bash
# ✅ 适当的提交时机
- 完成一个小功能点
- 修复一个 bug
- 完成一个测试用例
- 更新文档
- 重构完成一个模块

# ❌ 避免的提交时机
- 代码不能编译时
- 测试失败时
- 代码审查未通过时
- 包含敏感信息时
```

---

## 👥 代码审查

### 审查流程

#### 1. 自我审查
```bash
# 检查代码差异
git diff origin/develop...HEAD

# 运行完整测试
make test

# 检查代码质量
make lint

# 检查安全性
gosec ./...

# 检查性能影响
go test -bench=. ./...
```

#### 2. 同行审查
- **审查者**: 至少需要 1-2 名团队成员审查
- **审查时间**: 24-48 小时内响应
- **审查内容**: 代码质量、功能正确性、安全性、性能

#### 3. 技术负责人审查
- **时机**: 重大变更或复杂功能
- **审查者**: 技术负责人或架构师
- **关注点**: 架构设计、技术选型、长期影响

### 审查清单

#### 代码质量
- [ ] 代码符合项目规范
- [ ] 函数和变量命名清晰
- [ ] 代码结构合理，职责单一
- [ ] 无重复代码
- [ ] 错误处理完善

#### 功能正确性
- [ ] 功能实现符合需求
- [ ] 边界条件处理正确
- [ ] 异常情况处理合理
- [ ] 数据验证完整

#### 安全性
- [ ] 输入验证完整
- [ ] 无 SQL 注入风险
- [ ] 无 XSS 风险
- [ ] 敏感信息保护
- [ ] 权限检查正确

#### 性能
- [ ] 无明显性能问题
- [ ] 数据库查询优化
- [ ] 内存使用合理
- [ ] 并发安全

#### 测试
- [ ] 测试覆盖率充足
- [ ] 测试用例全面
- [ ] 测试可读性好
- [ ] 集成测试通过

#### 文档
- [ ] 代码注释充分
- [ ] API 文档更新
- [ ] README 更新（如需要）
- [ ] 变更日志更新

### 审查反馈

#### 反馈类型
```markdown
# 建议性反馈 (Suggestion)
"建议将这个函数拆分成更小的函数，提高可读性。"

# 必须修改 (Required)
"这里存在 SQL 注入风险，必须使用参数化查询。"

# 讨论性反馈 (Discussion)
"这个实现方式可能存在性能问题，我们讨论一下其他方案。"

# 赞扬 (Praise)
"这个错误处理设计得很棒！"
```

#### 响应时间
- **紧急问题**: 4 小时内响应
- **一般问题**: 24 小时内响应
- **建议性反馈**: 48 小时内响应

---

## 🚀 部署流程

### 1. 预发布检查

#### 功能验证
```bash
# 本地测试
make test

# 集成测试
make test-integration

# 端到端测试
make test-e2e

# 性能测试
make test-performance
```

#### 安全检查
```bash
# 安全扫描
gosec ./...

# 依赖漏洞检查
go list -json -m all | nancy sleuth

# 代码质量检查
make lint
```

### 2. 合并到 develop

#### 合并要求
- [ ] 所有 CI 检查通过
- [ ] 至少一个审查者批准
- [ ] 无合并冲突
- [ ] 功能测试完成

#### 合并操作
```bash
# 通过 GitHub UI 合并
# 或使用命令行（需要权限）
git checkout develop
git pull origin develop
git merge --no-ff feature/user-profile-management
git push origin develop
```

### 3. 发布准备

#### 创建发布分支
```bash
# 从 develop 创建发布分支
git checkout -b release/v2.1.0

# 更新版本信息
vim VERSION
vim CHANGELOG.md

# 提交版本更新
git add VERSION CHANGELOG.md
git commit -m "chore: prepare release v2.1.0"
git push origin release/v2.1.0
```

#### 发布测试
```bash
# 部署到测试环境
make deploy-staging

# 运行冒烟测试
make test-smoke

# 手动验证功能
```

### 4. 发布到生产

#### 合并到 main
```bash
# 合并发布分支到 main
git checkout main
git merge --no-ff release/v2.1.0
git tag -a v2.1.0 -m "Release version 2.1.0"
git push origin main --tags
```

#### 合并回 develop
```bash
# 合并回 develop 分支
git checkout develop
git merge --no-ff release/v2.1.0
git push origin develop
```

#### 部署生产
```bash
# 部署到生产环境
make deploy-production

# 验证部署
make verify-production
```

---

## 🛠️ 开发工具配置

### Git 配置

#### 全局配置
```bash
# 设置用户信息
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# 设置编辑器
git config --global core.editor "vim"

# 设置默认分支
git config --global init.defaultBranch main

# 设置推送策略
git config --global push.default simple

# 设置合并策略
git config --global pull.rebase false
```

#### 项目配置
```bash
# 设置项目级别的配置
git config core.autocrlf input  # Linux/Mac
git config core.autocrlf true   # Windows

# 设置文件权限
git config core.filemode false

# 设置忽略文件权限
git config core.ignorecase false
```

### Git Hooks

#### Pre-commit Hook
```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "运行 pre-commit 检查..."

# 代码格式检查
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

echo "✅ 所有检查通过"
```

#### Pre-push Hook
```bash
#!/bin/sh
# .git/hooks/pre-push

echo "运行 pre-push 检查..."

# 检查是否有未提交的更改
if ! git diff --cached --exit-code; then
    echo "❌ 有未提交的更改，请先提交"
    exit 1
fi

# 运行完整测试套件
make test

# 检查代码覆盖率
COVERAGE=$(go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print $3}' | sed 's/%//' || echo "0")
if (( $(echo "$COVERAGE < 70" | bc -l) )); then
    echo "❌ 测试覆盖率 $COVERAGE% 低于要求 70%"
    exit 1
fi

echo "✅ Pre-push 检查通过"
```

### IDE 配置

#### VS Code 设置
```json
// .vscode/settings.json
{
    "git.autofetch": true,
    "git.enableSmartCommit": true,
    "git.postCommitCommand": "none",
    "git.showInlineOpenFileAction": false,
    "git.suggestSmartCommit": false,
    "git.confirmSync": false,
    "git.branchProtection": ["main", "develop"],
    "git.branchProtectionPrompt": "alwaysCommitToNewBranch"
}
```

---

## 📊 质量指标

### 代码质量标准

#### 测试覆盖率
- **最低要求**: 70%
- **推荐目标**: 80%
- **优秀标准**: 90%+

#### 代码复杂度
- **圈复杂度**: < 15
- **函数长度**: < 100 行
- **文件长度**: < 500 行

#### 性能指标
- **API 响应时间**: < 200ms (P95)
- **数据库查询时间**: < 100ms
- **内存使用**: < 512MB

### CI/CD 检查

#### 自动化检查
- [ ] 代码编译通过
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 代码质量检查通过
- [ ] 安全扫描通过
- [ ] 性能测试通过

#### 手动检查
- [ ] 功能验证完成
- [ ] 用户验收测试通过
- [ ] 文档更新完成

---

## 🚨 紧急流程

### 热修复流程

#### 1. 创建热修复分支
```bash
# 从 main 分支创建
git checkout main
git pull origin main
git checkout -b hotfix/critical-security-fix

# 推送到远程
git push -u origin hotfix/critical-security-fix
```

#### 2. 快速修复
```bash
# 进行最小化修复
# 只修复关键问题，避免其他更改

# 运行必要测试
go test ./...

# 快速代码检查
golangci-lint run --fast
```

#### 3. 紧急发布
```bash
# 提交修复
git commit -m "hotfix: fix critical security vulnerability"

# 推送到远程
git push origin hotfix/critical-security-fix

# 创建紧急 PR
# 要求立即审查和合并

# 合并到 main
git checkout main
git merge --no-ff hotfix/critical-security-fix
git tag -a v2.1.1 -m "Hotfix v2.1.1"
git push origin main --tags

# 合并回 develop
git checkout develop
git merge --no-ff hotfix/critical-security-fix
git push origin develop

# 立即部署到生产
make deploy-emergency
```

---

## 📞 技术支持

### 寻求帮助
- **技术问题**: 在相关的 Issue 中提问
- **流程问题**: 联系项目维护者
- **紧急情况**: 直接联系技术负责人

### 联系方式
- **技术负责人**: tech-lead@law-oa.com
- **项目维护者**: maintainers@law-oa.com
- **开发团队**: dev-team@law-oa.com

### 社区资源
- **项目文档**: https://docs.law-oa.com
- **开发者论坛**: https://forum.law-oa.com
- **企业微信群**: Law OA Go 开发群

---

**文档版本**: v2.1.0
**最后更新**: 2025-09-30
**下次审查**: 2025-12-30
**维护团队**: 开发团队