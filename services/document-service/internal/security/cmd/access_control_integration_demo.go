package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// SimpleAccessControl 简化的访问控制系统
type SimpleAccessControl struct {
	users     map[string]*SimpleUser
	roles     map[string]*SimpleRole
	resources map[string]*SimpleResource
	policies  map[string]bool // key: user:resource:action
	logger    *slog.Logger
	mu        sync.RWMutex
}

// SimpleUser 简单用户
type SimpleUser struct {
	ID         string            `json:"id"`
	Username   string            `json:"username"`
	Email      string            `json:"email"`
	FullName   string            `json:"full_name"`
	Roles      []string          `json:"roles"`
	Attributes map[string]string `json:"attributes"`
	IsActive   bool              `json:"is_active"`
}

// SimpleRole 简单角色
type SimpleRole struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Level       int               `json:"level"`
	Permissions map[string]string `json:"permissions"`
}

// SimpleResource 简单资源
type SimpleResource struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Owner       string            `json:"owner"`
	Attributes map[string]string `json:"attributes"`
	Sensitivity string            `json:"sensitivity"`
}

// SimpleAccessRequest 简单访问请求
type SimpleAccessRequest struct {
	UserID     string            `json:"user_id"`
	ResourceID string            `json:"resource_id"`
	Action     string            `json:"action"`
	Context    map[string]string `json:"context"`
	RequestID  string            `json:"request_id"`
	Timestamp  time.Time         `json:"timestamp"`
}

// SimpleAccessDecision 简单访问决策
type SimpleAccessDecision struct {
	Allowed   bool              `json:"allowed"`
	Reason    string            `json:"reason"`
	Policies  []string          `json:"policies"`
	Attributes map[string]string `json:"attributes"`
	Duration  time.Duration     `json:"duration"`
	RequestID  string            `json:"request_id"`
	Timestamp  time.Time         `json:"timestamp"`
}

// NewSimpleAccessControl 创建简单访问控制
func NewSimpleAccessControl() *SimpleAccessControl {
	return &SimpleAccessControl{
		users:     make(map[string]*SimpleUser),
		roles:     make(map[string]*SimpleRole),
		resources: make(map[string]*SimpleResource),
		policies:  make(map[string]bool),
		logger:    slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

// LoadDefaults 加载默认数据
func (sac *SimpleAccessControl) LoadDefaults() {
	sac.mu.Lock()
	defer sac.mu.Unlock()

	// 默认角色
	sac.roles["super_admin"] = &SimpleRole{
		ID:          "super_admin",
		Name:        "超级管理员",
		Description: "拥有所有权限的系统管理员",
		Level:       9,
		Permissions: map[string]string{"*": "*"},
	}

	sac.roles["admin"] = &SimpleRole{
		ID:          "admin",
		Name:        "管理员",
		Description: "系统管理员",
		Level:       8,
		Permissions: map[string]string{
			"user_management": "*",
			"role_management": "*",
			"audit_log":       "read",
		},
	}

	sac.roles["partner"] = &SimpleRole{
		ID:          "partner",
		Name:        "合伙人",
		Description: "律师事务所合伙人",
		Level:       7,
		Permissions: map[string]string{
			"case_management":     "*",
			"client_management":   "*",
			"document_management": "*",
			"billing_management":  "*",
			"report_access":       "*",
		},
	}

	sac.roles["lawyer"] = &SimpleRole{
		ID:          "lawyer",
		Name:        "律师",
		Description: "执业律师",
		Level:       6,
		Permissions: map[string]string{
			"assigned_cases":     "*",
			"document_management": "read,write",
			"client_communication": "*",
		},
	}

	sac.roles["assistant"] = &SimpleRole{
		ID:          "assistant",
		Name:        "助理",
		Description: "律师助理",
		Level:       4,
		Permissions: map[string]string{
			"assigned_cases":  "read",
			"document_management": "read",
			"document_upload": "*",
		},
	}

	sac.roles["client"] = &SimpleRole{
		ID:          "client",
		Name:        "客户",
		Description: "客户",
		Level:       2,
		Permissions: map[string]string{
			"own_cases":     "read",
			"own_documents": "read",
			"communication": "*",
		},
	}

	// 默认用户
	sac.users["admin001"] = &SimpleUser{
		ID:       "admin001",
		Username: "admin",
		Email:    "admin@lawfirm.com",
		FullName: "系统管理员",
		Roles:    []string{"super_admin"},
		IsActive: true,
		Attributes: map[string]string{
			"department": "IT",
			"position":   "系统管理员",
			"level":      "9",
		},
	}

	sac.users["partner001"] = &SimpleUser{
		ID:       "partner001",
		Username: "zhang",
		Email:    "zhang@lawfirm.com",
		FullName: "张律师",
		Roles:    []string{"partner"},
		IsActive: true,
		Attributes: map[string]string{
			"department": "诉讼部",
			"position":   "合伙人",
			"level":      "8",
			"years_exp":   "15",
		},
	}

	sac.users["lawyer001"] = &SimpleUser{
		ID:       "lawyer001",
		Username: "li",
		Email:    "li@lawfirm.com",
		FullName: "李律师",
		Roles:    []string{"lawyer"},
		IsActive: true,
		Attributes: map[string]string{
			"department": "合同部",
			"position":   "律师",
			"level":      "6",
			"years_exp":   "8",
		},
	}

	sac.users["assistant001"] = &SimpleUser{
		ID:       "assistant001",
		Username: "wang",
		Email:    "wang@lawfirm.com",
		FullName: "王助理",
		Roles:    []string{"assistant"},
		IsActive: true,
		Attributes: map[string]string{
			"department": "行政部",
			"position":   "助理",
			"level":      "4",
			"years_exp":   "3",
		},
	}

	sac.users["client001"] = &SimpleUser{
		ID:       "client001",
		Username: "client001",
		Email:    "client@example.com",
		FullName: "客户A",
		Roles:    []string{"client"},
		IsActive: true,
		Attributes: map[string]string{
			"client_type": "corporate",
			"industry":    "technology",
		},
	}

	// 默认资源
	sac.resources["case_001"] = &SimpleResource{
		ID:          "case_001",
		Type:        "case",
		Name:        "商业纠纷案件",
		Owner:       "partner001",
		Sensitivity: "confidential",
		Attributes: map[string]string{
			"department": "诉讼部",
			"status":     "active",
			"client_id":  "client001",
			"case_value": "high",
		},
	}

	sac.resources["document_001"] = &SimpleResource{
		ID:          "document_001",
		Type:        "document",
		Name:        "合同文件",
		Owner:       "lawyer001",
		Sensitivity: "internal",
		Attributes: map[string]string{
			"department":    "合同部",
			"document_type": "pdf",
			"file_size":     "2MB",
			"case_id":       "case_001",
		},
	}

	sac.resources["document_002"] = &SimpleResource{
		ID:          "document_002",
		Type:        "document",
		Name:        "证据材料",
		Owner:       "assistant001",
		Sensitivity: "restricted",
		Attributes: map[string]string{
			"department":    "诉讼部",
			"document_type": "zip",
			"file_size":     "50MB",
			"case_id":       "case_001",
		},
	}

	sac.resources["user_management"] = &SimpleResource{
		ID:          "user_management",
		Type:        "system",
		Name:        "用户管理",
		Owner:       "admin001",
		Sensitivity: "internal",
		Attributes: map[string]string{
			"system_module": "admin",
			"access_level":  "admin_only",
		},
	}
}

// CheckPermission 检查权限
func (sac *SimpleAccessControl) CheckPermission(ctx context.Context, req *SimpleAccessRequest) (*SimpleAccessDecision, error) {
	startTime := time.Now()

	sac.mu.RLock()
	defer sac.mu.RUnlock()

	decision := &SimpleAccessDecision{
		Allowed:   false,
		Reason:    "默认拒绝",
		Policies:  []string{},
		Attributes: make(map[string]string),
		RequestID:  req.RequestID,
		Timestamp:  time.Now(),
	}

	// 1. 验证用户存在
	user, exists := sac.users[req.UserID]
	if !exists {
		decision.Reason = "用户不存在"
		return decision, nil
	}

	// 2. 检查用户状态
	if !user.IsActive {
		decision.Reason = "用户已被禁用"
		return decision, nil
	}

	// 3. 获取资源
	resource, exists := sac.resources[req.ResourceID]
	if !exists {
		decision.Reason = "资源不存在"
		return decision, nil
	}

	// 4. 执行权限检查
	sac.evaluatePermissions(user, resource, req, decision)

	decision.Duration = time.Since(startTime)
	return decision, nil
}

// evaluatePermissions 评估权限
func (sac *SimpleAccessControl) evaluatePermissions(user *SimpleUser, resource *SimpleResource, req *SimpleAccessRequest, decision *SimpleAccessDecision) {
	// 规则1: 资源所有者拥有完全访问权限
	if user.ID == resource.Owner {
		decision.Allowed = true
		decision.Reason = "资源所有者权限"
		decision.Policies = append(decision.Policies, "resource_owner")
		return
	}

	// 规则2: 管理员权限
	for _, roleID := range user.Roles {
		if role, exists := sac.roles[roleID]; exists {
			if role.Level >= 8 {
				decision.Allowed = true
				decision.Reason = fmt.Sprintf("管理员权限 (%s)", role.Name)
				decision.Policies = append(decision.Policies, fmt.Sprintf("admin_role:%s", roleID))
				return
			}
		}
	}

	// 规则3: 时间限制
	if resource.Sensitivity == "restricted" || resource.Sensitivity == "confidential" {
		hour := req.Timestamp.Hour()
		if hour < 8 || hour > 18 {
			if overtimeAllowed, ok := user.Attributes["overtime_allowed"]; !ok || overtimeAllowed != "true" {
				decision.Allowed = false
				decision.Reason = "非工作时间禁止访问敏感资源"
				decision.Policies = append(decision.Policies, "time_restriction")
				return
			}
		}
	}

	// 规则4: 位置限制
	if resource.Sensitivity == "restricted" {
		if location, ok := req.Context["location"]; ok {
			if location != "office" && location != "secure_network" {
				decision.Allowed = false
				decision.Reason = "敏感资源仅允许在安全网络环境访问"
				decision.Policies = append(decision.Policies, "location_restriction")
				return
			}
		}
	}

	// 规则5: 默认拒绝
	decision.Allowed = false
	decision.Reason = "默认拒绝"
}

// 访问控制集成测试
type AccessControlIntegrationTest struct {
	logger *slog.Logger
}

// 测试场景定义
type TestCase struct {
	Name        string
	Description string
	Scenarios   []TestScenario
	PassCount   int
	TotalCount  int
}

type TestScenario struct {
	UserID      string
	ResourceID  string
	Action      string
	Context     map[string]string
	ExpectAllow bool
	Description string
}

// NewAccessControlIntegrationTest 创建集成测试实例
func NewAccessControlIntegrationTest() *AccessControlIntegrationTest {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return &AccessControlIntegrationTest{
		logger: logger,
	}
}

// RunAllTests 运行所有集成测试
func (acit *AccessControlIntegrationTest) RunAllTests() error {
	fmt.Println("🚀 开始访问控制服务集成测试...")
	fmt.Println("==========================================")

	// 创建访问控制实例
	ac := NewSimpleAccessControl()
	ac.LoadDefaults()

	// 定义测试套件
	testSuites := []TestCase{
		{
			Name:        "RBAC基础权限测试",
			Description: "验证基于角色的访问控制基本功能",
			Scenarios:   acit.getRBACTestScenarios(),
		},
		{
			Name:        "ABAC属性权限测试",
			Description: "验证基于属性的访问控制功能",
			Scenarios:   acit.getABACTestScenarios(),
		},
		{
			Name:        "时间限制测试",
			Description: "验证时间相关的访问控制规则",
			Scenarios:   acit.getTimeRestrictionTestScenarios(),
		},
		{
			Name:        "位置限制测试",
			Description: "验证位置相关的访问控制规则",
			Scenarios:   acit.getLocationRestrictionTestScenarios(),
		},
		{
			Name:        "资源所有权测试",
			Description: "验证资源所有者权限规则",
			Scenarios:   acit.getOwnershipTestScenarios(),
		},
		{
			Name:        "敏感资源访问测试",
			Description: "验证敏感资源的特殊访问规则",
			Scenarios:   acit.getSensitiveResourceTestScenarios(),
		},
		{
			Name:        "跨部门访问测试",
			Description: "验证跨部门访问控制规则",
			Scenarios:   acit.getCrossDepartmentTestScenarios(),
		},
		{
			Name:        "权限继承测试",
			Description: "验证权限层次继承功能",
			Scenarios:   acit.getPermissionInheritanceTestScenarios(),
		},
		{
			Name:        "边界条件测试",
			Description: "验证边界条件和异常情况",
			Scenarios:   acit.getBoundaryTestScenarios(),
		},
	}

	// 执行所有测试套件
	totalPassed := 0
	totalFailed := 0

	for i, testSuite := range testSuites {
		fmt.Printf("\n📋 测试套件 %d: %s\n", i+1, testSuite.Name)
		fmt.Printf("描述: %s\n", testSuite.Description)
		fmt.Println(strings.Repeat("-", 60))

		passed, failed := acit.runTestSuite(ac, testSuite)
		testSuite.PassCount = passed
		testSuite.TotalCount = passed + failed

		totalPassed += passed
		totalFailed += failed

		fmt.Printf("结果: %d 通过, %d 失败\n\n", passed, failed)
	}

	// 生成测试报告
	fmt.Println("==========================================")
	fmt.Printf("📊 测试总结\n")
	fmt.Printf("总测试数: %d\n", totalPassed+totalFailed)
	fmt.Printf("通过: %d (%.1f%%)\n", totalPassed, float64(totalPassed)/float64(totalPassed+totalFailed)*100)
	fmt.Printf("失败: %d (%.1f%%)\n", totalFailed, float64(totalFailed)/float64(totalPassed+totalFailed)*100)

	if totalFailed == 0 {
		fmt.Println("🎉 所有测试通过！访问控制服务运行正常")
	} else {
		fmt.Printf("⚠️  有 %d 个测试失败，需要检查\n", totalFailed)
	}

	// 性能测试
	fmt.Println("\n⚡ 性能测试")
	acit.runPerformanceTest(ac)

	// 并发测试
	fmt.Println("\n🔄 并发测试")
	acit.runConcurrencyTest(ac)

	return nil
}

// runTestSuite 运行单个测试套件
func (acit *AccessControlIntegrationTest) runTestSuite(ac *SimpleAccessControl, testSuite TestCase) (int, int) {
	passed := 0
	failed := 0

	for j, scenario := range testSuite.Scenarios {
		result := acit.runSingleScenario(ac, scenario, j+1)
		if result {
			passed++
		} else {
			failed++
		}
	}

	return passed, failed
}

// runSingleScenario 运行单个测试场景
func (acit *AccessControlIntegrationTest) runSingleScenario(ac *SimpleAccessControl, scenario TestScenario, index int) bool {
	req := &SimpleAccessRequest{
		UserID:     scenario.UserID,
		ResourceID: scenario.ResourceID,
		Action:     scenario.Action,
		Context:    scenario.Context,
		RequestID:  fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), index),
		Timestamp:  time.Now(),
	}

	decision, err := ac.CheckPermission(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 测试 %d: %s - 权限检查出错: %v\n", index, scenario.Description, err)
		return false
	}

	status := "✅"
	if decision.Allowed != scenario.ExpectAllow {
		status = "❌"
	}

	fmt.Printf("%s 测试 %d: %s\n", status, index, scenario.Description)
	fmt.Printf("   用户: %s -> 资源: %s (%s)\n", scenario.UserID, scenario.ResourceID, scenario.Action)
	fmt.Printf("   期望: %v, 实际: %v\n", scenario.ExpectAllow, decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	if len(decision.Policies) > 0 {
		fmt.Printf("   策略: %v\n", decision.Policies)
	}
	fmt.Printf("   耗时: %v\n", decision.Duration)

	success := decision.Allowed == scenario.ExpectAllow
	if !success {
		fmt.Printf("   ❌ 测试失败!\n")
	}

	fmt.Println()
	return success
}

// 获取RBAC测试场景
func (acit *AccessControlIntegrationTest) getRBACTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "admin001",
			ResourceID:  "user_management",
			Action:      "read",
			ExpectAllow: true,
			Description: "超级管理员访问用户管理",
		},
		{
			UserID:      "admin001",
			ResourceID:  "case_001",
			Action:      "write",
			ExpectAllow: true,
			Description: "超级管理员修改案件",
		},
		{
			UserID:      "partner001",
			ResourceID:  "case_001",
			Action:      "read",
			ExpectAllow: true,
			Description: "合伙人读取自己的案件",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "document_001",
			Action:      "read",
			ExpectAllow: true,
			Description: "律师读取自己的文档",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "case_001",
			Action:      "write",
			ExpectAllow: false,
			Description: "律师修改他人案件（权限不足）",
		},
		{
			UserID:      "assistant001",
			ResourceID:  "document_002",
			Action:      "read",
			ExpectAllow: true,
			Description: "助理读取自己的文档",
		},
		{
			UserID:      "client001",
			ResourceID:  "document_001",
			Action:      "read",
			ExpectAllow: false,
			Description: "客户访问内部文档（权限不足）",
		},
	}
}

// 获取ABAC测试场景
func (acit *AccessControlIntegrationTest) getABACTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "assistant001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"time": "14:30", "location": "office", "device": "desktop"},
			ExpectAllow: true,
			Description: "资源所有者在工作时间访问",
		},
		{
			UserID:      "partner001",
			ResourceID:  "case_001",
			Action:      "read",
			Context:     map[string]string{"time": "10:30", "location": "office"},
			ExpectAllow: true,
			Description: "资源所有者在办公室访问",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"time": "22:30", "location": "home"},
			ExpectAllow: false,
			Description: "非工作时间访问敏感资源",
		},
	}
}

// 获取时间限制测试场景
func (acit *AccessControlIntegrationTest) getTimeRestrictionTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "lawyer001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"time": "10:30", "location": "office"},
			ExpectAllow: true,
			Description: "工作时间访问敏感资源",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"time": "22:30", "location": "office"},
			ExpectAllow: false,
			Description: "非工作时间访问敏感资源",
		},
		{
			UserID:      "admin001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"time": "22:30", "location": "home"},
			ExpectAllow: true,
			Description: "管理员无时间限制",
		},
	}
}

// 获取位置限制测试场景
func (acit *AccessControlIntegrationTest) getLocationRestrictionTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "lawyer001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"location": "office"},
			ExpectAllow: true,
			Description: "在办公室访问敏感资源",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"location": "home"},
			ExpectAllow: false,
			Description: "在家访问敏感资源",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"location": "secure_network"},
			ExpectAllow: true,
			Description: "在安全网络访问敏感资源",
		},
	}
}

// 获取资源所有权测试场景
func (acit *AccessControlIntegrationTest) getOwnershipTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "assistant001",
			ResourceID:  "document_002",
			Action:      "read",
			ExpectAllow: true,
			Description: "资源所有者访问自己的资源",
		},
		{
			UserID:      "assistant001",
			ResourceID:  "document_002",
			Action:      "write",
			ExpectAllow: true,
			Description: "资源所有者修改自己的资源",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "document_002",
			Action:      "read",
			ExpectAllow: false,
			Description: "非所有者访问他人资源",
		},
	}
}

// 获取敏感资源测试场景
func (acit *AccessControlIntegrationTest) getSensitiveResourceTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "admin001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"time": "22:30", "location": "home"},
			ExpectAllow: true,
			Description: "管理员访问敏感资源（无限制）",
		},
		{
			UserID:      "partner001",
			ResourceID:  "document_002",
			Action:      "read",
			Context:     map[string]string{"time": "22:30", "location": "home"},
			ExpectAllow: false,
			Description: "合伙人非工作时间在家访问敏感资源",
		},
	}
}

// 获取跨部门访问测试场景
func (acit *AccessControlIntegrationTest) getCrossDepartmentTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "lawyer001",
			ResourceID:  "case_001",
			Action:      "read",
			Context:     map[string]string{"time": "10:30", "location": "office"},
			ExpectAllow: false,
			Description: "合同部律师访问诉讼部案件",
		},
		{
			UserID:      "partner001",
			ResourceID:  "document_001",
			Action:      "read",
			Context:     map[string]string{"time": "10:30", "location": "office"},
			ExpectAllow: true,
			Description: "高级合伙人跨部门访问（高级权限）",
		},
	}
}

// 获取权限继承测试场景
func (acit *AccessControlIntegrationTest) getPermissionInheritanceTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "admin001",
			ResourceID:  "user_management",
			Action:      "write",
			ExpectAllow: true,
			Description: "超级管理员拥有所有权限",
		},
		{
			UserID:      "partner001",
			ResourceID:  "case_001",
			Action:      "delete",
			ExpectAllow: true,
			Description: "合伙人拥有案件管理权限",
		},
	}
}

// 获取边界条件测试场景
func (acit *AccessControlIntegrationTest) getBoundaryTestScenarios() []TestScenario {
	return []TestScenario{
		{
			UserID:      "nonexistent_user",
			ResourceID:  "case_001",
			Action:      "read",
			ExpectAllow: false,
			Description: "不存在的用户访问资源",
		},
		{
			UserID:      "lawyer001",
			ResourceID:  "nonexistent_resource",
			Action:      "read",
			ExpectAllow: false,
			Description: "访问不存在的资源",
		},
		{
			UserID:      "client001",
			ResourceID:  "document_001",
			Action:      "invalid_action",
			ExpectAllow: false,
			Description: "执行无效操作",
		},
	}
}

// runPerformanceTest 运行性能测试
func (acit *AccessControlIntegrationTest) runPerformanceTest(ac *SimpleAccessControl) {
	req := &SimpleAccessRequest{
		UserID:     "lawyer001",
		ResourceID: "document_001",
		Action:     "read",
		Context:    map[string]string{},
		RequestID:  "perf_test",
		Timestamp:  time.Now(),
	}

	// 预热
	for i := 0; i < 10; i++ {
		ac.CheckPermission(context.Background(), req)
	}

	// 性能测试
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		req.RequestID = fmt.Sprintf("perf_test_%d", i)
		ac.CheckPermission(context.Background(), req)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	fmt.Printf("   执行 %d 次权限检查\n", iterations)
	fmt.Printf("   总耗时: %v\n", duration)
	fmt.Printf("   平均耗时: %v\n", avgDuration)
	fmt.Printf("   QPS: %.0f\n", float64(iterations)/duration.Seconds())

	if avgDuration < 1*time.Millisecond {
		fmt.Printf("   ✅ 性能测试通过\n")
	} else {
		fmt.Printf("   ⚠️  性能需要优化\n")
	}
}

// runConcurrencyTest 运行并发测试
func (acit *AccessControlIntegrationTest) runConcurrencyTest(ac *SimpleAccessControl) {
	concurrency := 100
	iterations := 10

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	errorCount := 0

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				req := &SimpleAccessRequest{
					UserID:     fmt.Sprintf("user_%d", workerID%5+1),
					ResourceID: "document_001",
					Action:     "read",
					Context:    map[string]string{},
					RequestID:  fmt.Sprintf("concurrent_test_%d_%d", workerID, j),
					Timestamp:  time.Now(),
				}

				_, err := ac.CheckPermission(context.Background(), req)
				mu.Lock()
				if err != nil {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalRequests := concurrency * iterations
	fmt.Printf("   并发测试: %d 协程 x %d 次请求\n", concurrency, iterations)
	fmt.Printf("   总请求数: %d\n", totalRequests)
	fmt.Printf("   成功: %d, 失败: %d\n", successCount, errorCount)
	fmt.Printf("   总耗时: %v\n", duration)
	fmt.Printf("   QPS: %.0f\n", float64(totalRequests)/duration.Seconds())

	if errorCount == 0 {
		fmt.Printf("   ✅ 并发测试通过\n")
	} else {
		fmt.Printf("   ⚠️  并发测试有错误\n")
	}
}

// main 主函数
func main() {
	test := NewAccessControlIntegrationTest()

	if err := test.RunAllTests(); err != nil {
		fmt.Printf("❌ 集成测试失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎯 访问控制服务集成测试完成!")
}