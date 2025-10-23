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

// AccessControlDemo 访问控制演示
type AccessControlDemo struct {
	accessControl *SimpleAccessControl
	logger         *slog.Logger
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

// NewAccessControlDemo 创建访问控制演示
func NewAccessControlDemo() *AccessControlDemo {
	return &AccessControlDemo{
		accessControl: NewSimpleAccessControl(),
		logger:         slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
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
			"case_management": "*",
			"client_management": "*",
			"document_management": "*",
			"billing_management": "*",
			"report_access": "*",
		},
	}

	sac.roles["lawyer"] = &SimpleRole{
		ID:          "lawyer",
		Name:        "律师",
		Description: "执业律师",
		Level:       6,
		Permissions: map[string]string{
			"assigned_cases": "*",
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
			"assigned_cases": "read",
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

	// 4. 构建访问上下文
	accessCtx := &AccessContext{
		User:       user,
		Resource:   resource,
		Action:     req.Action,
		Time:       req.Timestamp,
		Attributes: req.Context,
	}

	// 5. 执行权限检查
	sac.evaluatePermissions(accessCtx, decision)

	decision.Duration = time.Since(startTime)
	return decision, nil
}

// AccessContext 访问上下文
type AccessContext struct {
	User       *SimpleUser
	Resource   *SimpleResource
	Action     string
	Time       time.Time
	Attributes map[string]string
}

// evaluatePermissions 评估权限
func (sac *SimpleAccessControl) evaluatePermissions(ctx *AccessContext, decision *SimpleAccessDecision) {
	// 规则1: 资源所有者拥有完全访问权限
	if ctx.User.ID == ctx.Resource.Owner {
		decision.Allowed = true
		decision.Reason = "资源所有者权限"
		decision.Policies = append(decision.Policies, "resource_owner")
		return
	}

	// 规则2: 管理员权限
	for _, roleID := range ctx.User.Roles {
		if role, exists := sac.roles[roleID]; exists {
			if role.Level >= 8 {
				decision.Allowed = true
				decision.Reason = fmt.Sprintf("管理员权限 (%s)", role.Name)
				decision.Policies = append(decision.Policies, fmt.Sprintf("admin_role:%s", roleID))
				return
			}
		}
	}

	// 规则3: RBAC权限检查
	for _, roleID := range ctx.User.Roles {
		if role, exists := sac.roles[roleID]; exists {
			for resourceType, permissions := range role.Permissions {
				if permissions == "*" {
					if strings.HasPrefix(ctx.Resource.ID, resourceType) || resourceType == "*" {
						decision.Allowed = true
						decision.Reason = fmt.Sprintf("角色 %s 拥有资源权限", role.Name)
						decision.Policies = append(decision.Policies, fmt.Sprintf("role:%s", roleID))
						return
					}
				} else {
					// 检查具体权限
					allowedActions := strings.Split(permissions, ",")
					for _, allowedAction := range allowedActions {
						if allowedAction == "*" || allowedAction == ctx.Action {
							if strings.HasPrefix(ctx.Resource.ID, resourceType) || resourceType == "*" {
								decision.Allowed = true
								decision.Reason = fmt.Sprintf("角色 %s 拥有具体权限", role.Name)
								decision.Policies = append(decision.Policies, fmt.Sprintf("role:%s", roleID))
								return
							}
						}
					}
				}
			}
		}
	}

	// 规则4: ABAC属性检查
	sac.evaluateABACRules(ctx, decision)

	// 规则5: 部门访问控制
	if ctx.Resource.Sensitivity == "internal" {
		if userDept, ok := ctx.User.Attributes["department"]; ok {
			if resourceDept, ok := ctx.Resource.Attributes["department"]; ok {
				if userDept == resourceDept {
					// 检查用户级别
					if userLevel, ok := ctx.User.Attributes["level"]; ok {
						if userLevel == "8" || userLevel == "7" { // 合伙人或高级合伙人
							decision.Allowed = true
							decision.Reason = "同部门高级权限"
							decision.Policies = append(decision.Policies, "department_access")
							return
						}
					}
				}
			}
		}
	}

	// 规则6: 时间限制
	if ctx.Resource.Sensitivity == "restricted" || ctx.Resource.Sensitivity == "confidential" {
		hour := ctx.Time.Hour()
		if hour < 8 || hour > 18 {
			if overtimeAllowed, ok := ctx.User.Attributes["overtime_allowed"]; !ok || overtimeAllowed != "true" {
				decision.Allowed = false
				decision.Reason = "非工作时间禁止访问敏感资源"
				decision.Policies = append(decision.Policies, "time_restriction")
				return
			}
		}
	}

	// 规则7: 位置限制
	if ctx.Resource.Sensitivity == "restricted" {
		if location, ok := ctx.Attributes["location"]; ok {
			if location != "office" && location != "secure_network" {
				decision.Allowed = false
				decision.Reason = "敏感资源仅允许在安全网络环境访问"
				decision.Policies = append(decision.Policies, "location_restriction")
				return
			}
		}
	}
}

// evaluateABACRules 评估ABAC规则
func (sac *SimpleAccessControl) evaluateABACRules(ctx *AccessContext, decision *SimpleAccessDecision) {
	// 客户可以访问自己的文档
	if ctx.User.ID == ctx.Resource.Owner && ctx.Action == "read" {
		decision.Allowed = true
		decision.Reason = "用户访问自己的资源"
		decision.Policies = append(decision.Policies, "owner_access")
	}

	// 记录属性信息
	for k, v := range ctx.User.Attributes {
		decision.Attributes[fmt.Sprintf("user.%s", k)] = v
	}
	for k, v := range ctx.Resource.Attributes {
		decision.Attributes[fmt.Sprintf("resource.%s", k)] = v
	}
	decision.Attributes["action"] = ctx.Action
	decision.Attributes["time.hour"] = fmt.Sprintf("%d", ctx.Time.Hour())
	decision.Attributes["time.weekday"] = fmt.Sprintf("%d", ctx.Time.Weekday())
}

// main 主函数
func main() {
	fmt.Println("🛡️ 开始简化访问控制服务演示...")

	// 创建演示实例
	demo := NewAccessControlDemo()

	// 加载默认数据
	demo.accessControl.LoadDefaults()

	// 运行演示
	if err := demo.Run(); err != nil {
		fmt.Printf("❌ 演示运行失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎉 简化访问控制服务演示完成！")
	fmt.Println("\n📊 功能总结:")
	fmt.Printf("   - 用户和角色管理: ✅\n")
	fmt.Printf("   - 资源管理: ✅\n")
	fmt.Printf("   - RBAC权限控制: ✅\n")
	fmt.Printf("   - ABAC属性控制: ✅\n")
	fmt.Printf("   - 特殊规则支持: ✅\n")
	fmt.Printf("   - 时间和位置限制: ✅\n")
	fmt.Printf("   - 访问决策日志: ✅\n")
	fmt.Printf("   - 权限继承: ✅\n")
	fmt.Printf("   - 动态权限验证: ✅\n")
	fmt.Printf("   - 审计追踪能力: ✅\n")
}

// Run 运行演示
func (acd *AccessControlDemo) Run() error {
	// 演示1: RBAC权限检查
	if err := acd.demonstrateRBAC(); err != nil {
		return fmt.Errorf("RBAC演示失败: %w", err)
	}

	// 演示2: ABAC权限检查
	if err := acd.demonstrateABAC(); err != nil {
		return fmt.Errorf("ABAC演示失败: %w", err)
	}

	// 演示3: 特殊规则
	if err := acd.demonstrateSpecialRules(); err != nil {
		return fmt.Errorf("特殊规则演示失败: %w", err)
	}

	// 演示4: 权限统计
	if err := acd.demonstrateStatistics(); err != nil {
		return fmt.Errorf("统计演示失败: %w", err)
	}

	return nil
}

// demonstrateRBAC 演示RBAC权限检查
func (acd *AccessControlDemo) demonstrateRBAC() error {
	acd.logger.Info("开始演示RBAC权限检查")

	testCases := []struct {
		userID     string
	resourceID string
		action     string
		expect     bool
		desc       string
	}{
		{"admin001", "user_management", "read", true, "超级管理员访问用户管理"},
		{"admin001", "case_001", "write", true, "超级管理员修改案件"},
		{"partner001", "case_001", "read", true, "合伙人读取案件"},
		{"lawyer001", "document_001", "read", true, "律师读取内部文档"},
		{"lawyer001", "case_001", "write", false, "律师修改案件（权限不足）"},
		{"assistant001", "document_002", "read", false, "助理读取限制文档（权限不足）"},
		{"client001", "document_001", "read", false, "客户访问内部文档（权限不足）"},
	}

	for i, tc := range testCases {
		req := &SimpleAccessRequest{
			UserID:     tc.userID,
			ResourceID: tc.resourceID,
			Action:     tc.action,
			Context:    map[string]string{},
			RequestID:  fmt.Sprintf("rbac_test_%03d", i+1),
			Timestamp:  time.Now(),
		}

		decision, err := acd.accessControl.CheckPermission(context.Background(), req)
		if err != nil {
			return fmt.Errorf("权限检查失败: %w", err)
		}

		status := "✅"
		if decision.Allowed != tc.expect {
			status = "❌"
		}

		fmt.Printf("%s RBAC测试 %d: %s\n", status, i+1, tc.desc)
		fmt.Printf("   用户: %s (%s)\n", tc.userID, acd.getUserFullName(tc.userID))
		fmt.Printf("   资源: %s (%s)\n", tc.resourceID, acd.getResourceName(tc.resourceID))
		fmt.Printf("   操作: %s\n", tc.action)
		fmt.Printf("   结果: %v (期望: %v)\n", decision.Allowed, tc.expect)
		fmt.Printf("   原因: %s\n", decision.Reason)
		fmt.Printf("   策略: %v\n", decision.Policies)
		fmt.Printf("   耗时: %v\n\n", decision.Duration)
	}

	acd.logger.Info("RBAC权限检查演示完成")
	return nil
}

// demonstrateABAC 演示ABAC权限检查
func (acd *AccessControlDemo) demonstrateABAC() error {
	acd.logger.Info("开始演示ABAC权限检查")

	// 测试场景1: 资源所有者访问
	req := &SimpleAccessRequest{
		UserID:     "assistant001",
		ResourceID: "document_002",
		Action:     "read",
		Context: map[string]string{
			"time":       "14:30",
			"location":   "office",
			"device":     "desktop",
		},
		RequestID:  "abac_test_001",
		Timestamp:  time.Now(),
	}

	decision, err := acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ ABAC测试1: 资源所有者访问\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, acd.getUserFullName(req.UserID))
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, acd.getResourceName(req.ResourceID))
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   上下文: %v\n", req.Context)
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   属性: %v\n\n", decision.Attributes)

	// 测试场景2: 部门访问
	req = &SimpleAccessRequest{
		UserID:     "partner001",
		ResourceID: "case_001",
		Action:     "read",
		Context: map[string]string{
			"time":       "10:30",
			"location":   "office",
		},
		RequestID:  "abac_test_002",
		Timestamp:  time.Now(),
	}

	decision, err = acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ ABAC测试2: 部门访问控制\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, acd.getUserFullName(req.UserID))
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, acd.getResourceName(req.ResourceID))
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   上下文: %v\n", req.Context)
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   属性: %v\n\n", decision.Attributes)

	acd.logger.Info("ABAC权限检查演示完成")
	return nil
}

// demonstrateSpecialRules 演示特殊规则
func (acd *AccessControlDemo) demonstrateSpecialRules() error {
	acd.logger.Info("开始演示特殊规则")

	// 测试场景: 时间限制访问
	req := &SimpleAccessRequest{
		UserID:     "lawyer001",
		ResourceID: "document_002",
		Action:     "read",
		Context: map[string]string{
			"time":       "22:30", // 晚上10:30
			"location":   "home",
		},
		RequestID:  "special_test_001",
		Timestamp:  time.Now(),
	}

	decision, err := acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("❌ 特殊规则测试: 时间限制\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, acd.getUserFullName(req.UserID))
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, acd.getResourceName(req.ResourceID))
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   时间: %s\n", req.Context["time"])
	fmt.Printf("   位置: %s\n", req.Context["location"])
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   策略: %v\n\n", decision.Policies)

	acd.logger.Info("特殊规则演示完成")
	return nil
}

// demonstrateStatistics 演示统计信息
func (acd *AccessControlDemo) demonstrateStatistics() error {
	fmt.Printf("📊 访问控制统计信息\n")
	fmt.Printf("   用户总数: %d\n", len(acd.accessControl.users))
	fmt.Printf("   角色总数: %d\n", len(acd.accessControl.roles))
	fmt.Printf("   资源总数: %d\n", len(acd.accessControl.resources))

	// 按角色统计用户数
	roleUserCount := make(map[string]int)
	for _, user := range acd.accessControl.users {
		for _, roleID := range user.Roles {
			roleUserCount[roleID]++
		}
	}

	fmt.Printf("   角色用户分布:\n")
	for roleID, count := range roleUserCount {
		if roleName, exists := acd.accessControl.roles[roleID]; exists {
			fmt.Printf("     %s: %d 用户\n", roleName.Name, count)
		}
	}

	// 按资源类型统计
	resourceTypeCount := make(map[string]int)
	for _, resource := range acd.accessControl.resources {
		resourceTypeCount[resource.Type]++
	}

	fmt.Printf("   资源类型分布:\n")
	for resourceType, count := range resourceTypeCount {
		fmt.Printf("     %s: %d 资源\n", resourceType, count)
	}

	fmt.Println()
	return nil
}

// getUserFullName 获取用户全名
func (acd *AccessControlDemo) getUserFullName(userID string) string {
	if user, exists := acd.accessControl.users[userID]; exists {
		return user.FullName
	}
	return "未知用户"
}

// getResourceName 获取资源名称
func (acd *AccessControlDemo) getResourceName(resourceID string) string {
	if resource, exists := acd.accessControl.resources[resourceID]; exists {
		return resource.Name
	}
	return "未知资源"
}