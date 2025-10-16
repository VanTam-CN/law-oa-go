<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# Law OA Go - AI Context Template

## 1. Project Overview
- **Vision**: 构建现代化的律师事务所办公自动化系统，为中小型律师事务所提供完整的数字化解决方案
- **Current Phase**: v2.1.0 - 生产就绪阶段，核心功能已完成，搜索和文档管理功能开发中
- **Key Architecture**: 单体架构设计，Go后端 + React前端，MySQL数据库 + Redis缓存
- **Development Strategy**: 采用分层架构，注重性能优化、安全性和可维护性

## 2. Project Structure

项目遵循标准的Go项目结构，采用分层架构设计。完整的技术栈和文件树结构参考[项目结构文档](.claude/steering/structure.md)。

## 3. Coding Standards & AI Instructions

### General Instructions
- 你的首要任务是管理好自己的上下文。在规划变更之前总是先读取相关文件。
- 更新文档时要简洁扼要，避免文档膨胀。
- 遵循KISS、YAGNI、DRY原则编写代码。
- 如有疑问，遵循经过验证的最佳实践进行实现。
- 未经用户批准，不要提交到git。
- 不要运行任何服务器，而是告诉用户运行服务器进行测试。
- 优先考虑行业标准的库和框架，而不是自定义实现。
- 绝对不要模拟任何内容，不要使用占位符，不要省略代码。
- 在相关的地方应用SOLID原则，使用现代框架特性而不是重新发明解决方案。
- 对一个想法是否好坏要诚实。
- 让副作用明确且最小化。
- 设计支持多服务语音处理的数据库模式 - 使用会话ID进行状态协调，避免在服务器内存中存储对话数据。
- 保持对话历史轻量级（文本，而非音频）。

### File Organization & Modularity
- 默认创建多个小而专注的文件，而不是单一的大文件
- 每个文件应该有单一职责和明确目的
- 尽可能保持文件在350行以下 - 通过提取工具、常量、类型或逻辑组件到单独的模块来分割大文件
- 分离关注点：将工具、常量、类型、组件和业务逻辑分离到不同的文件中
- 优先使用组合而不是继承 - 只对真正的'is-a'关系使用继承，对'has-a'或行为混合使用组合

- 遵循现有的项目结构和约定 - 将文件放在适当的目录中。如有必要，创建新目录并移动文件。
- 使用定义明确的子目录来保持组织性和可扩展性
- 使用清晰的文件夹层次结构和一致的命名约定构建项目
- 正确导入/导出 - 设计为可重用和可维护

### Type Hints (REQUIRED)
- **总是**为函数参数和返回值使用类型提示
- 使用`from typing import`来处理复杂类型
- 优先使用`Optional[T]`而不是`Union[T, None]`
- 对数据结构使用Pydantic模型

### Naming Conventions
- **Classes**: PascalCase (e.g., `VoicePipeline`)
- **Functions/Methods**: snake_case (e.g., `process_audio`)
- **Constants**: UPPER_SNAKE_CASE (e.g., `MAX_AUDIO_SIZE`)
- **Private methods**: 前导下划线 (e.g., `_validate_input`)
- **Pydantic Models**: PascalCase with `Schema` suffix (e.g., `ChatRequestSchema`, `UserSchema`)

### Documentation Requirements
- 每个模块都需要文档字符串
- 每个公共函数都需要文档字符串
- 使用Google风格的文档字符串
- 在文档字符串中包含类型信息

### Security First
- 永远不要信任外部输入 - 在边界验证所有内容
- 将机密信息保存在环境变量中，不要在代码中保存
- 记录安全事件（登录尝试、身份验证失败、速率限制、权限拒绝），但永远不要记录敏感数据（音频、对话内容、令牌、个人信息）
- 在API网关级别认证用户 - 永远不要信任客户端令牌
- 使用行级安全性(RLS)来强制用户之间的数据隔离
- 设计跨所有客户端类型一致工作的认证
- 在创建会话之前验证服务器端的所有身份验证令牌
- 在存储或处理之前清理所有用户输入

### Error Handling
- 使用具体的异常而不是通用异常
- 总是用上下文记录错误
- 提供有用的错误信息
- 安全地失败 - 错误不应该暴露系统内部

### Observable Systems & Logging Standards
- 每个请求都需要关联ID用于调试
- 构建机器可读的结构化日志，而不是人类可读的 - 使用JSON格式和一致的字段（时间戳、级别、关联ID、事件、上下文）进行自动化分析
- 使跨服务边界的调试成为可能

### State Management
- 每个数据只有一个真实来源
- 使状态变更明确和可追踪
- 设计多服务语音处理 - 使用会话ID进行状态协调，避免在服务器内存中存储对话数据
- 保持对话历史轻量级（文本，而不是音频）

### API Design Principles
- RESTful设计，一致的URL模式
- 正确使用HTTP状态码
- 从第一天开始版本化API（/v1/, /v2/）
- 支持列表端点的分页
- 使用一致的JSON响应格式：
  - 成功: `{ "data": {...}, "error": null }`
  - 错误: `{ "data": null, "error": {"message": "...", "code": "..."} }`

## 4. Multi-Agent Workflows & Context Injection

### Automatic Context Injection for Sub-Agents
当使用Task工具生成子代理时，核心项目上下文（CLAUDE.md、project-structure.md、docs-overview.md）会通过subagent-context-injector钩子自动注入到它们的提示中。这确保所有子代理都能立即访问基本的项目文档，无需在每个Task提示中手动指定。

## 5. MCP Server Integrations

### Gemini Consultation Server
**何时使用：**
- 需要深度分析或多种方法的复杂编码问题
- 代码审查和架构讨论
- 调试跨多个文件的复杂问题
- 性能优化和重构指导
- 复杂实现的详细解释
- 高安全相关任务

**自动上下文注入：**
- kit的`gemini-context-injector.sh`钩子会自动包含两个关键文件用于新会话：
  - `docs/ai-context/project-structure.md` - 完整的项目结构和技术栈
  - `MCP-ASSISTANT-RULES.md` - 你的项目特定的编码标准和指南
- 这确保Gemini始终对你的技术栈、架构和项目标准有全面的了解

**使用模式：**
```python
# 新咨询会话（项目结构通过钩子自动附加）
mcp__gemini__consult_gemini(
    specific_question="我应该如何优化这个语音管道？",
    problem_description="需要减少实时音频处理的延迟",
    code_context="当前管道顺序处理音频...",
    attached_files=[
        "src/core/pipelines/voice_pipeline.py"  # 你的特定文件
    ],
    preferred_approach="optimize"
)

# 现有会话中的跟进
mcp__gemini__consult_gemini(
    specific_question="内存使用情况怎么样？",
    session_id="session_123",
    additional_context="已经实施你的建议，现在看到高内存使用"
)
```

**关键能力：**
- 具有上下文保留的持久会话
- 文件附加和缓存用于多文件分析
- 专门化协助模式（解决方案、审查、调试、优化、解释）
- 复杂、多步骤问题的会话管理

**重要提示：** 将Gemini的响应视为建议反馈。批判性地评估建议，将有价值的见解整合到你的解决方案中，然后继续你的实现。

### Context7 Documentation Server
**Repository**: [Context7 MCP Server](https://github.com/upstash/context7)

**何时使用：**
- 使用外部库/框架（React、FastAPI、Next.js等）工作时
- 需要超出训练截止日期的当前文档
- 实现与第三方工具的新集成或功能
- 排查库特定问题

**使用模式：**
```python
# 将库名解析为Context7 ID
mcp__context7__resolve_library_id(libraryName="react")

# 获取专注的文档
mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/facebook/react",
    topic="hooks",
    tokens=8000
)
```

**关键能力：**
- 最新库文档访问
- 主题专注的文档检索
- 支持特定库版本
- 与当前开发实践集成

## 6. Post-Task Completion Protocol
完成任何编码任务后，遵循此清单：

### 1. Type Safety & Quality Checks
根据修改的内容运行适当的命令：
- **Python项目**: 运行mypy类型检查
- **TypeScript项目**: 运行tsc --noEmit
- **其他语言**: 运行适当的linting/类型检查工具

### 2. Verification
- 确保所有类型检查通过，然后才认为任务完成
- 如果发现类型错误，在将任务标记为完成之前修复它们

## Steering Documents

以下是为AI助手提供项目指导的 steering documents：

### 📋 产品指南
- **文件**: `.claude/steering/product.md`
- **内容**: 产品概述、核心价值主张、功能模块、业务规则、技术特色、用户价值
- **用途**: 帮助AI理解产品定位和业务需求

### 🔧 技术指南
- **文件**: `.claude/steering/tech.md`
- **内容**: 技术栈详情、构建系统、测试策略、部署配置、性能优化、安全配置、开发规范
- **用途**: 指导AI进行技术实现和架构决策

### 📁 结构指南
- **文件**: `.claude/steering/structure.md`
- **内容**: 目录组织、文件命名约定、关键文件位置、分层架构、模块化原则
- **用途**: 帮助AI理解项目结构和代码组织

### 使用方法
这些steering documents会在AI会话开始时自动注入，为AI助手提供项目背景和指导原则。确保在开发过程中参考这些文档以保持一致性和质量。