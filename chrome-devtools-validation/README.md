# Chrome DevTools Validation Framework

专为律所OA系统设计的全面测试框架，基于Chrome DevTools MCP服务进行端到端自动化验证。

## 功能特性

- 🔧 **Chrome DevTools MCP集成** - 使用MCP服务进行浏览器自动化
- 🎯 **全面测试覆盖** - 覆盖所有业务功能模块
- 📊 **详细测试报告** - 多格式报告输出（HTML、JSON、JUnit）
- 🔄 **智能重试机制** - 自动重试失败的测试用例
- 📸 **失败截图** - 自动捕获失败时的屏幕截图
- 🔍 **完整日志记录** - 结构化日志便于调试
- ⚡ **并发执行** - 支持多测试并行执行
- 🧪 **Page Object模式** - 可维护的测试架构

## 快速开始

### 安装依赖

```bash
npm install
```

### 配置测试环境

创建 `test.config.json` 文件：

```json
{
  "baseUrl": "http://localhost:3000",
  "apiUrl": "http://localhost:8080/api",
  "defaultTimeout": 30000,
  "retryAttempts": 3,
  "headless": false,
  "outputDir": "./test-results"
}
```

### 运行测试

```bash
# 构建项目
npm run build

# 运行测试
npm run test:chrome

# 运行特定测试套件
npm run test:chrome -- --suite="auth"

# 生成测试报告
npm run report
```

## 项目结构

```
chrome-devtools-validation/
├── src/
│   ├── core/           # 核心框架
│   │   ├── config.ts   # 配置管理
│   │   ├── logger.ts   # 日志系统
│   │   └── index.ts
│   ├── types/          # 类型定义
│   │   ├── test-types.ts
│   │   ├── mcp-types.ts
│   │   ├── domain-types.ts
│   │   └── index.ts
│   ├── mcp/            # MCP服务集成
│   ├── pages/          # Page Object模型
│   ├── tests/          # 测试用例
│   ├── reporters/      # 报告生成器
│   └── data/           # 测试数据
├── tests/              # 集成测试
├── test-results/       # 测试结果
└── docs/               # 文档
```

## 测试覆盖范围

### 核心功能模块

1. **用户认证** (`tests/auth/`)
   - 登录/登出功能
   - 密码重置
   - 权限验证
   - 会话管理

2. **案件管理** (`tests/cases/`)
   - 案件CRUD操作
   - 案件状态管理
   - 案件分配
   - 里程碑跟踪

3. **客户管理** (`tests/clients/`)
   - 客户信息管理
   - 客户分类
   - 联系人管理

4. **文档管理** (`tests/documents/`)
   - 文档上传/下载
   - 文档版本控制
   - 文档搜索
   - 权限管理

5. **财务管理** (`tests/finance/`)
   - 费用记录
   - 发票管理
   - 支付处理
   - 财务报告

6. **冲突检测** (`tests/conflicts/`)
   - 冲突检查
   - 冲突解决流程
   - 风险评估

### 集成测试

- **搜索功能** - 全局搜索和模块搜索
- **报表系统** - 各类报表生成
- **工作流** - 端到端业务流程
- **权限系统** - 基于角色的访问控制
- **性能测试** - 页面加载和响应时间
- **安全性测试** - 认证和授权验证

## 配置选项

### 基础配置

```typescript
interface TestConfig {
  baseUrl: string;           // 应用基础URL
  apiUrl: string;            // API基础URL
  defaultTimeout: number;    // 默认超时时间(ms)
  retryAttempts: number;    // 重试次数
  retryDelay: number;        // 重试延迟(ms)
  concurrencyLevel: number; // 并发级别
  screenshotOnFailure: boolean; // 失败时截图
  headless: boolean;         // 无头模式
  slowMo: number;           // 操作延迟(ms)
  outputDir: string;         // 输出目录
  reporting: {               // 报告配置
    formats: ('html' | 'json' | 'junit')[];
    includeScreenshots: boolean;
    includeLogs: boolean;
  };
}
```

### 环境变量

```bash
TEST_BASE_URL=http://localhost:3000
TEST_API_URL=http://localhost:8080/api
TEST_TIMEOUT=30000
TEST_RETRY_ATTEMPTS=3
TEST_HEADLESS=false
TEST_OUTPUT_DIR=./test-results
```

## 编写测试

### 基本测试用例

```typescript
import { TestCase, TestStep, Assertion } from '../types';

const loginTest: TestCase = {
  id: 'auth-login-001',
  name: '用户登录测试',
  description: '验证用户可以成功登录系统',
  priority: 'P0',
  steps: [
    {
      id: 'step-1',
      name: '导航到登录页面',
      type: 'navigate',
      url: '/login',
    },
    {
      id: 'step-2',
      name: '输入用户名',
      type: 'fill',
      selector: '#username',
      value: 'testuser',
    },
    {
      id: 'step-3',
      name: '输入密码',
      type: 'fill',
      selector: '#password',
      value: 'password123',
    },
    {
      id: 'step-4',
      name: '点击登录按钮',
      type: 'click',
      selector: '#login-button',
    },
  ],
  assertions: [
    {
      id: 'assert-1',
      type: 'url-contains',
      expected: '/dashboard',
    },
    {
      id: 'assert-2',
      type: 'element-visible',
      selector: '.user-profile',
      expected: true,
    },
  ],
};
```

### Page Object模式

```typescript
import { PageObject } from '../core/page-object';

class LoginPage extends PageObject {
  private selectors = {
    usernameInput: '#username',
    passwordInput: '#password',
    loginButton: '#login-button',
    errorMessage: '.error-message',
  };

  async login(username: string, password: string): Promise<void> {
    await this.fill(this.selectors.usernameInput, username);
    await this.fill(this.selectors.passwordInput, password);
    await this.click(this.selectors.loginButton);
  }

  async getErrorMessage(): Promise<string> {
    return this.getText(this.selectors.errorMessage);
  }

  async isLoginFormVisible(): Promise<boolean> {
    return this.isVisible(this.selectors.loginButton);
  }
}
```

## 自定义报告

可以扩展报告系统以支持自定义格式：

```typescript
import { TestReporter, TestExecutionResult } from '../types';

class CustomReporter implements TestReporter {
  async generateReport(results: TestExecutionResult): Promise<void> {
    // 自定义报告生成逻辑
  }

  async generateSummary(results: TestExecutionResult): Promise<string> {
    // 自定义摘要生成逻辑
    return `测试完成: ${results.passedTests}/${results.totalTests} 通过`;
  }
}
```

## 故障排除

### 常见问题

1. **Chrome DevTools MCP连接失败**
   - 检查MCP服务是否启动
   - 验证网络连接
   - 检查配置文件中的URL

2. **测试超时**
   - 增加defaultTimeout配置
   - 检查网络延迟
   - 优化测试步骤

3. **元素定位失败**
   - 使用更稳定的选择器
   - 添加等待时间
   - 检查页面加载状态

### 调试模式

启用详细日志输出：

```bash
DEBUG=chrome-devtools* npm run test:chrome
```

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

MIT License