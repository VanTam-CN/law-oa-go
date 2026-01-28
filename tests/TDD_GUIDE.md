# 财务模块 TDD 测试配置指南

## 🎯 测试策略

### 测试金字塔
```
        /\
       /  \        E2E Tests (少量)
      /____\       集成测试 (适中)
     /      \      单元测试 (大量)
    /________\
```

## 📋 测试覆盖率目标

### 总体目标
- **单元测试覆盖率**: ≥ 90%
- **集成测试覆盖率**: ≥ 80%
- **关键业务逻辑**: 100%
- **API接口测试**: ≥ 85%

### 分类目标
- **Models层**: 95% (数据模型验证)
- **Repository层**: 90% (数据访问逻辑)
- **Service层**: 95% (业务逻辑)
- **Handler层**: 85% (API接口)
- **Utils层**: 90% (工具函数)

## 🛠️ 测试工具配置

### 1. Go测试配置

#### go.test 配置
```yaml
# .go.test.yml
coverage:
  profile: full
  coverpackage: ./...
  covermode: atomic
  include: []
  exclude:
    - "*_test.go"
    - "*_mock.go"
    - "*/mock/*"
    - "*/test/*"
```

### 2. Makefile测试命令

```makefile
# 测试相关命令
.PHONY: test test-cover test-bench test-race test-integration

# 运行所有测试
test:
	@echo "运行所有测试..."
	go test -v -race -coverprofile=coverage.out ./...

# 运行测试并生成覆盖率报告
test-cover:
	@echo "运行测试并生成覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 运行单元测试
test-unit:
	@echo "运行单元测试..."
	go test -v -run "^Test[^I]" ./...

# 运行集成测试
test-integration:
	@echo "运行集成测试..."
	go test -v -run "^TestI" ./...

# 运行性能测试
test-bench:
	@echo "运行性能测试..."
	go test -bench=. -benchmem ./...

# 运行竞态检测
test-race:
	@echo "运行竞态检测..."
	go test -race ./...

# 生成覆盖率报告
coverage-report:
	@echo "生成覆盖率报告..."
	go tool cover -func=coverage.out | grep total
	go tool cover -html=coverage.out -o coverage.html

# 清理测试文件
test-clean:
	@echo "清理测试文件..."
	rm -f coverage.out coverage.html
	go clean -test
```

### 3. GitHub Actions CI配置

```yaml
# .github/workflows/test.yml
name: 测试

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: 运行测试
    runs-on: ubuntu-latest

    strategy:
      matrix:
        go-version: ['1.23', '1.22']

    steps:
    - name: 检出代码
      uses: actions/checkout@v3

    - name: 设置Go环境
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: 安装依赖
      run: go mod download

    - name: 运行测试
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: 上传覆盖率到Codecov
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
        flags: unittests
        name: codecov-umbrella

    - name: 生成覆盖率报告
      if: matrix.go-version == '1.23'
      run: |
        go tool cover -html=coverage.out -o coverage.html

    - name: 上传覆盖率报告
      uses: actions/upload-artifact@v3
      with:
        name: 覆盖率报告
        path: coverage.html
```

## 🧪 TDD开发流程

### Red-Green-Refactor循环

#### 1. RED - 编写失败的测试
```bash
# 1. 先写测试，确保失败
go test -v -run TestInvoice_CreateInvoice
# 预期: FAIL
```

#### 2. GREEN - 让测试通过
```bash
# 2. 实现最小代码让测试通过
go test -v -run TestInvoice_CreateInvoice
# 预期: PASS
```

#### 3. REFACTOR - 重构代码
```bash
# 3. 优化代码，保持测试通过
go test -v -race -run TestInvoice_CreateInvoice
# 预期: PASS
```

## 📊 测试文件组织

### 目录结构
```
internal/
├── models/
│   ├── financial_models.go
│   └── financial_models_test.go         # 模型单元测试
├── repositories/
│   ├── financial_repository.go
│   └── financial_repository_test.go     # 仓储集成测试
├── services/
│   ├── financial_service.go
│   └── financial_service_test.go        # 服务层测试
├── handlers/
│   ├── financial_handler.go
│   └── financial_handler_test.go        # API测试
└── utils/
    ├── financial_calculator.go
    └── financial_calculator_test.go     # 工具函数测试

tests/
├── integration/
│   └── financial_integration_test.go    # 集成测试
├── e2e/
│   └── financial_e2e_test.go           # 端到端测试
└── fixtures/
    └── financial_test_data.go          # 测试数据
```

## 🎯 测试命名规范

### 函数命名
```go
// 单元测试
func Test{FunctionName}_{Scenario}_{ExpectedResult}(t *testing.T)

// 集成测试
func TestI{ComponentName}_{Workflow}_{ExpectedResult}(t *testing.T)

// 基准测试
func Benchmark{FunctionName}_{Scenario}(b *testing.B)

// 表格驱动测试
func Test{FunctionName}_{Scenarios}(t *testing.T)
```

### 测试用例组织
```go
func TestInvoice_CalculateTaxAmount(t *testing.T) {
    tests := []struct {
        name          string
        amount        float64
        taxRate       float64
        expectedTax   float64
        expectedTotal float64
    }{
        {
            name:          "13%税率计算",
            amount:        10000.00,
            taxRate:       0.13,
            expectedTax:   1300.00,
            expectedTotal: 11300.00,
        },
        // ... 更多测试用例
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

## 🔧 Mock和Fixture

### 使用testify/mock
```go
package services_test

import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock Repository
type MockFinancialRepository struct {
    mock.Mock
}

func (m *MockFinancialRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
    args := m.Called(ctx, invoice)
    return args.Error(0)
}

// 使用Mock测试
func TestInvoiceService_CreateInvoice(t *testing.T) {
    mockRepo := new(MockFinancialRepository)
    service := NewFinancialService(mockRepo)

    invoice := &models.Invoice{
        InvoiceNumber: "TEST-001",
        Amount:        10000.00,
    }

    mockRepo.On("CreateInvoice", mock.Anything, invoice).Return(nil)

    err := service.CreateInvoice(context.Background(), invoice)
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### 测试Fixture
```go
package testfixtures

func CreateTestInvoice() *models.Invoice {
    return &models.Invoice{
        InvoiceNumber: "TEST-001",
        InvoiceType:   "vat_special",
        ProjectName:   "测试项目",
        Amount:        10000.00,
        TaxRate:       0.13,
        IssueDate:     time.Now(),
        DueDate:       time.Now().AddDate(0, 1, 0),
        CreatedBy:     1,
        Status:        "pending",
    }
}

func CreateTestExpense() *models.Expense {
    return &models.Expense{
        ExpenseNumber: "TEST-001",
        Description:   "测试费用",
        Amount:        500.00,
        TotalAmount:   500.00,
        ApplicantID:   1,
        Status:        "pending",
    }
}
```

## 📈 覆盖率报告解读

### 覆盖率报告示例
```
mode: atomic
internal/models/financial_models.go:45:    models.Invoice     100.0%
internal/models/financial_models.go:123:   models.Expense     95.0%
internal/repositories/financial_repository.go:67:   FinancialRepository.CreateInvoice    90.0%
total: statements: 1250, 85.6%
```

### 覆盖率标准
- **< 70%**: 需要改进
- **70-80%**: 基本合格
- **80-90%**: 良好
- **90-95%**: 优秀
- **≥ 95%**: 卓越

## ⚠️ 常见TDD陷阱

### 1. 测试脆弱性
❌ **错误**: 过度依赖实现细节
✅ **正确**: 测试行为而非实现

### 2. 测试覆盖质量
❌ **错误**: 追求数量忽视质量
✅ **正确**: 关注关键路径和边界条件

### 3. Mock滥用
❌ **错误**: 过度Mock导致测试脱离实际
✅ **正确**: 合理使用Mock隔离依赖

### 4. 测试数据
❌ **错误**: 使用生产数据
✅ **正确**: 使用专门构建的测试数据

## 🚀 快速开始

### 运行所有测试
```bash
# 克隆项目后
cd law-oa-go

# 运行所有测试
make test

# 生成覆盖率报告
make test-cover

# 查看报告
open coverage.html
```

### TDD开发新功能
```bash
# 1. 创建测试文件
touch internal/services/financial_service_test.go

# 2. 编写测试（失败）
# 3. 实现功能（通过）
# 4. 重构代码（保持通过）
make test-watch
```

## 📚 最佳实践

### ✅ DO - 应该做的
1. **先写测试** - TDD的核心原则
2. **测试隔离** - 每个测试独立运行
3. **明确断言** - 使用具体的错误消息
4. **边界测试** - 测试正常、边界、异常情况
5. **性能基准** - 关键算法需要基准测试
6. **持续集成** - 每次提交都运行测试

### ❌ DON'T - 不应该做的
1. **跳过测试** - 不要写 `t.Skip()`
2. **忽略失败** - 不要忽视失败的测试
3. **过度Mock** - 保持测试的真实性
4. **测试私有方法** - 测试公共接口
5. **测试依赖外部** - 使用Mock隔离外部依赖
6. **忽略覆盖率** - 保持高覆盖率

## 🎯 当前财务模块测试状态

### 已完成的测试
- ✅ **财务模型单元测试** - `financial_models_test.go`
- ✅ **仓储层集成测试** - `financial_repository_test.go`

### 待完成的测试
- 🔄 **Service层测试** - 业务逻辑测试
- 🔄 **Handler层测试** - API接口测试
- 🔄 **工具函数测试** - 财务计算测试

### 测试统计
- **单元测试**: 150+ 测试用例
- **集成测试**: 50+ 测试场景
- **基准测试**: 10+ 性能测试
- **预期覆盖率**: 90%+

## 🔄 持续改进

### 下一步行动
1. 运行现有测试确保通过
2. 补充Service层测试
3. 添加Handler层测试
4. 集成到CI/CD流程
5. 定期审查测试质量
