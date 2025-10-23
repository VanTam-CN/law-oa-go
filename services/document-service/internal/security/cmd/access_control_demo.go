package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AccessControlDemo 访问控制演示程序
type AccessControlDemo struct {
	accessControl *AccessControlService
	logger         *slog.Logger
}

// main 主函数
func main() {
	fmt.Println("🛡️ 开始访问控制服务演示...")

	// 初始化演示
	demo, err := NewAccessControlDemo()
	if err != nil {
		fmt.Printf("❌ 初始化演示失败: %v\n", err)
		os.Exit(1)
	}

	// 运行演示
	if err := demo.Run(); err != nil {
		fmt.Printf("❌ 演示运行失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎉 访问控制服务演示完成！")
}

// NewAccessControlDemo 创建访问控制演示
func NewAccessControlDemo() (*AccessControlDemo, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 连接数据库
	dsn := "host=localhost user=postgres password=postgres dbname=law_oa port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// 如果连接失败，使用SQLite作为备选
		logger.Warn("PostgreSQL连接失败，使用SQLite作为备选", "error", err)
		db, err = gorm.Open(sqlite.Open("access_control_demo.db"), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("SQLite连接失败: %w", err)
		}
	}

	// 配置访问控制
	config := &AccessControlConfig{
		EnableRBAC:       true,
		EnableABAC:       true,
		EnableCache:      true,
		CacheTTL:         5 * time.Minute,
		EnableAudit:      true,
		DefaultDeny:      true,
		MaxHierarchyDepth: 5,
	}

	// 创建访问控制服务
	accessControl, err := NewAccessControlService(db, logger, config)
	if err != nil {
		return nil, fmt.Errorf("创建访问控制服务失败: %w", err)
	}

	return &AccessControlDemo{
		accessControl: accessControl,
		logger:         logger,
	}, nil
}

// Run 运行演示
func (acd *AccessControlDemo) Run() error {
	// 演示1: 创建测试数据
	if err := acd.demonstrateDataCreation(); err != nil {
		return fmt.Errorf("数据创建演示失败: %w", err)
	}

	// 演示2: RBAC权限检查
	if err := acd.demonstrateRBAC(); err != nil {
		return fmt.Errorf("RBAC演示失败: %w", err)
	}

	// 演示3: ABAC权限检查
	if err := acd.demonstrateABAC(); err != nil {
		return fmt.Errorf("ABAC演示失败: %w", err)
	}

	// 演示4: 动态权限管理
	if err := acd.demonstrateDynamicPermissionManagement(); err != nil {
		return fmt.Errorf("动态权限管理演示失败: %w", err)
	}

	// 演示5: 缓存性能测试
	if err := acd.demonstrateCachePerformance(); err != nil {
		return fmt.Errorf("缓存性能演示失败: %w", err)
	}

	// 演示6: 审计日志查询
	if err := acd.demonstrateAuditLogging(); err != nil {
		return fmt.Errorf("审计日志演示失败: %w", err)
	}

	return nil
}

// demonstrateDataCreation 演示数据创建
func (acd *AccessControlDemo) demonstrateDataCreation() error {
	acd.logger.Info("开始演示数据创建")

	// 创建测试用户
	users := []*User{
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
	}

	for _, user := range users {
		if err := acd.accessControl.attributeStore.CreateUser(user); err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}
	}

	// 创建测试资源
	resources := []*Resource{
		{
			ID:          "case_001",
			Type:        "case",
			Name:        "商业纠纷案件",
			Owner:       "partner001",
			Sensitivity: "confidential",
			Category:    "commercial",
			Attributes: map[string]string{
				"department":    "诉讼部",
				"status":        "active",
				"client_id":     "client001",
				"case_value":    "high",
			},
		},
		{
			ID:          "document_001",
			Type:        "document",
			Name:        "合同文件",
			Owner:       "lawyer001",
			Sensitivity: "internal",
			Category:    "contract",
			Attributes: map[string]string{
				"department":    "合同部",
				"document_type": "pdf",
				"file_size":     "2MB",
				"case_id":       "case_001",
			},
		},
		{
			ID:          "document_002",
			Type:        "document",
			Name:        "证据材料",
			Owner:       "assistant001",
			Sensitivity: "restricted",
			Category:    "evidence",
			Attributes: map[string]string{
				"department":    "诉讼部",
				"document_type": "zip",
				"file_size":     "50MB",
				"case_id":       "case_001",
			},
		},
		{
			ID:          "user_management",
			Type:        "system",
			Name:        "用户管理",
			Owner:       "admin001",
			Sensitivity: "internal",
			Category:    "admin",
			Attributes: map[string]string{
				"system_module": "admin",
				"access_level":  "admin_only",
			},
		},
	}

	for _, resource := range resources {
		if err := acd.accessControl.attributeStore.CreateResource(resource); err != nil {
			return fmt.Errorf("创建资源失败: %w", err)
		}
	}

	acd.logger.Info("数据创建演示完成", "users_count", len(users), "resources_count", len(resources))
	return nil
}

// demonstrateRBAC 演示RBAC权限检查
func (acd *AccessControlDemo) demonstrateRBAC() error {
	acd.logger.Info("开始演示RBAC权限检查")

	// 测试场景1: 超级管理员访问系统管理
	req := &AccessRequest{
		UserID:     "admin001",
		ResourceID: "user_management",
		Action:     "read",
		Context:    map[string]string{},
		IPAddress:  "192.168.1.100",
		UserAgent:  "Demo/1.0",
		RequestID:  "rbac_test_001",
		Timestamp:  time.Now(),
	}

	decision, err := acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ RBAC测试1 - 超级管理员访问系统管理\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, "系统管理员")
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, "用户管理")
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   策略: %v\n", decision.Policies)
	fmt.Printf("   耗时: %v\n\n", decision.Duration)

	// 测试场景2: 合伙人访问案件
	req = &AccessRequest{
		UserID:     "partner001",
		ResourceID: "case_001",
		Action:     "read",
		Context:    map[string]string{},
		IPAddress:  "192.168.1.100",
		UserAgent:  "Demo/1.0",
		RequestID:  "rbac_test_002",
		Timestamp:  time.Now(),
	}

	decision, err = acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ RBAC测试2 - 合伙人访问案件\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, "合伙人")
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, "商业纠纷案件")
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   策略: %v\n", decision.Policies)
	fmt.Printf("   耗时: %v\n\n", decision.Duration)

	// 测试场景3: 助理访问敏感文档（应该被拒绝）
	req = &AccessRequest{
		UserID:     "assistant001",
		ResourceID: "document_002",
		Action:     "read",
		Context:    map[string]string{},
		IPAddress:  "192.168.1.100",
		UserAgent:  "Demo/1.0",
		RequestID:  "rbac_test_003",
		Timestamp:  time.Now(),
	}

	decision, err = acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("❌ RBAC测试3 - 助理访问敏感文档\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, "助理")
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, "证据材料")
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   策略: %v\n", decision.Policies)
	fmt.Printf("   耗时: %v\n\n", decision.Duration)

	acd.logger.Info("RBAC权限检查演示完成")
	return nil
}

// demonstrateABAC 演示ABAC权限检查
func (acd *AccessControlDemo) demonstrateABAC() error {
	acd.logger.Info("开始演示ABAC权限检查")

	// 测试场景1: 工作时间访问
	currentTime := time.Now()
	req := &AccessRequest{
		UserID:     "lawyer001",
		ResourceID: "document_001",
		Action:     "read",
		Context: map[string]string{
			"time":       currentTime.Format("15:04"),
			"location":   "office",
			"device":     "desktop",
			"session_id": "sess_123",
		},
		IPAddress:  "192.168.1.100",
		UserAgent:  "Demo/1.0",
		RequestID:  "abac_test_001",
		Timestamp:  currentTime,
	}

	decision, err := acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ ABAC测试1 - 工作时间访问\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, "李律师")
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, "合同文件")
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   时间: %s\n", req.Context["time"])
	fmt.Printf("   位置: %s\n", req.Context["location"])
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   属性: %v\n", decision.Attributes)
	fmt.Printf("   耗时: %v\n\n", decision.Duration)

	// 测试场景2: 非工作时间访问敏感资源
	req = &AccessRequest{
		UserID:     "lawyer001",
		ResourceID: "document_002",
		Action:     "read",
		Context: map[string]string{
			"time":       "22:30", // 晚上10:30
			"location":   "home",
			"device":     "laptop",
			"session_id": "sess_124",
		},
		IPAddress:  "203.0.113.1",
		UserAgent:  "Demo/1.0",
		RequestID:  "abac_test_002",
		Timestamp:  time.Now(),
	}

	decision, err = acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("❌ ABAC测试2 - 非工作时间访问敏感资源\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, "李律师")
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, "证据材料")
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   时间: %s\n", req.Context["time"])
	fmt.Printf("   位置: %s\n", req.Context["location"])
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   属性: %v\n", decision.Attributes)
	fmt.Printf("   耗时: %v\n\n", decision.Duration)

	// 测试场景3: 资源所有者访问
	req = &AccessRequest{
		UserID:     "assistant001",
		ResourceID: "document_002",
		Action:     "read",
		Context: map[string]string{
			"time":       "14:30",
			"location":   "office",
			"device":     "desktop",
			"session_id": "sess_125",
		},
		IPAddress:  "192.168.1.100",
		UserAgent:  "Demo/1.0",
		RequestID:  "abac_test_003",
		Timestamp:  time.Now(),
	}

	decision, err = acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ ABAC测试3 - 资源所有者访问\n")
	fmt.Printf("   用户: %s (%s)\n", req.UserID, "王助理")
	fmt.Printf("   资源: %s (%s)\n", req.ResourceID, "证据材料")
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   时间: %s\n", req.Context["time"])
	fmt.Printf("   位置: %s\n", req.Context["location"])
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n", decision.Reason)
	fmt.Printf("   策略: %v\n", decision.Policies)
	fmt.Printf("   耗时: %v\n\n", decision.Duration)

	acd.logger.Info("ABAC权限检查演示完成")
	return nil
}

// demonstrateDynamicPermissionManagement 演示动态权限管理
func (acd *AccessControlDemo) demonstrateDynamicPermissionManagement() error {
	acd.logger.Info("开始演示动态权限管理")

	// 创建新角色
	newRole := &Role{
		ID:          "senior_partner",
		Name:        "高级合伙人",
		Description: "高级合伙人角色，拥有特殊权限",
		Level:       9,
		Department:  "管理层",
		IsSystem:    false,
		Attributes: map[string]string{
			"special_privileges": "true",
			"approval_required":  "false",
		},
	}

	if err := acd.accessControl.policyManager.AddRole(newRole); err != nil {
		return fmt.Errorf("添加角色失败: %w", err)
	}

	fmt.Printf("✅ 创建新角色: %s\n", newRole.Name)

	// 为用户分配新角色
	if err := acd.accessControl.policyManager.AssignRoleToUser("partner001", "senior_partner"); err != nil {
		return fmt.Errorf("分配角色失败: %w", err)
	}

	fmt.Printf("✅ 为用户 %s 分配角色 %s\n", "partner001", "senior_partner")

	// 测试新角色的权限
	req := &AccessRequest{
		UserID:     "partner001",
		ResourceID: "user_management",
		Action:     "read",
		Context:    map[string]string{},
		IPAddress:  "192.168.1.100",
		UserAgent:  "Demo/1.0",
		RequestID:  "dynamic_test_001",
		Timestamp:  time.Now(),
	}

	decision, err := acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ 测试新角色权限\n")
	fmt.Printf("   用户: %s\n", req.UserID)
	fmt.Printf("   角色: senior_partner\n")
	fmt.Printf("   资源: %s\n", req.ResourceID)
	fmt.Printf("   操作: %s\n", req.Action)
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n\n", decision.Reason)

	// 添加权限策略
	if err := acd.accessControl.policyManager.AssignPermissionToRole("senior_partner", "*", "read"); err != nil {
		return fmt.Errorf("分配权限失败: %w", err)
	}

	fmt.Printf("✅ 为角色 %s 添加权限策略\n", "senior_partner")

	// 再次测试权限
	decision, err = acd.accessControl.CheckPermission(context.Background(), req)
	if err != nil {
		return fmt.Errorf("权限检查失败: %w", err)
	}

	fmt.Printf("✅ 测试更新后的权限\n")
	fmt.Printf("   结果: %v\n", decision.Allowed)
	fmt.Printf("   原因: %s\n\n", decision.Reason)

	// 获取策略统计
	stats := acd.accessControl.policyManager.GetPolicyStats()
	fmt.Printf("📊 策略统计信息\n")
	for key, value := range stats {
		fmt.Printf("   %s: %v\n", key, value)
	}
	fmt.Println()

	acd.logger.Info("动态权限管理演示完成")
	return nil
}

// demonstrateCachePerformance 演示缓存性能
func (acd *AccessControlDemo) demonstrateCachePerformance() error {
	acd.logger.Info("开始演示缓存性能")

	req := &AccessRequest{
		UserID:     "lawyer001",
		ResourceID: "document_001",
		Action:     "read",
		Context:    map[string]string{},
		IPAddress:  "192.168.1.100",
		UserAgent:  "Demo/1.0",
		RequestID:  "cache_test_001",
		Timestamp:  time.Now(),
	}

	// 第一次访问（无缓存）
	start := time.Now()
	decision1, err := acd.accessControl.CheckPermission(context.Background(), req)
	firstAccessTime := time.Since(start)

	if err != nil {
		return fmt.Errorf("第一次权限检查失败: %w", err)
	}

	// 第二次访问（有缓存）
	start = time.Now()
	decision2, err := acd.accessControl.CheckPermission(context.Background(), req)
	secondAccessTime := time.Since(start)

	if err != nil {
		return fmt.Errorf("第二次权限检查失败: %w", err)
	}

	fmt.Printf("🚀 缓存性能测试\n")
	fmt.Printf("   第一次访问耗时: %v\n", firstAccessTime)
	fmt.Printf("   第二次访问耗时: %v\n", secondAccessTime)
	fmt.Printf("   性能提升: %.2fx\n", float64(firstAccessTime)/float64(secondAccessTime))
	fmt.Printf("   结果一致性: %v\n", decision1.Allowed == decision2.Allowed)

	// 批量测试
	batchSize := 100
	totalTime := time.Duration(0)

	for i := 0; i < batchSize; i++ {
		req.RequestID = fmt.Sprintf("cache_test_%03d", i)
		start := time.Now()
		_, err := acd.accessControl.CheckPermission(context.Background(), req)
		totalTime += time.Since(start)
		if err != nil {
			return fmt.Errorf("批量权限检查失败: %w", err)
		}
	}

	avgTime := totalTime / time.Duration(batchSize)
	fmt.Printf("   批量测试 - 次数: %d\n", batchSize)
	fmt.Printf("   批量测试 - 平均耗时: %v\n", avgTime)

	// 缓存统计
	cacheStats := acd.accessControl.cache.GetStats()
	fmt.Printf("   缓存统计: %v\n\n", cacheStats)

	acd.logger.Info("缓存性能演示完成")
	return nil
}

// demonstrateAuditLogging 演示审计日志
func (acd *AccessControlDemo) demonstrateAuditLogging() error {
	acd.logger.Info("开始演示审计日志")

	// 执行多次访问检查以生成日志
	testCases := []struct {
		userID     string
		resourceID string
		action     string
		expectAllow bool
	}{
		{"admin001", "user_management", "read", true},
		{"admin001", "user_management", "write", true},
		{"partner001", "case_001", "read", true},
		{"lawyer001", "document_001", "read", true},
		{"assistant001", "document_002", "read", false},
		{"client001", "document_001", "read", false},
	}

	for i, tc := range testCases {
		req := &AccessRequest{
			UserID:     tc.userID,
			ResourceID: tc.resourceID,
			Action:     tc.action,
			Context:    map[string]string{},
			IPAddress:  "192.168.1.100",
			UserAgent:  "Demo/1.0",
			RequestID:  fmt.Sprintf("audit_test_%03d", i+1),
			Timestamp:  time.Now(),
		}

		decision, err := acd.accessControl.CheckPermission(context.Background(), req)
		if err != nil {
			return fmt.Errorf("权限检查失败: %w", err)
		}

		fmt.Printf("📝 审计日志测试 %d\n", i+1)
		fmt.Printf("   用户: %s\n", tc.userID)
		fmt.Printf("   资源: %s\n", tc.resourceID)
		fmt.Printf("   操作: %s\n", tc.action)
		fmt.Printf("   期望: %v, 实际: %v\n", tc.expectAllow, decision.Allowed)
		fmt.Printf("   匹配: %v\n", tc.expectAllow == decision.Allowed)
	}

	// 获取访问统计
	stats, err := acd.accessControl.auditLogger.GetAccessStats("hour")
	if err != nil {
		return fmt.Errorf("获取访问统计失败: %w", err)
	}

	fmt.Printf("\n📊 访问统计信息\n")
	for key, value := range stats {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// 获取可疑活动
	suspicious, err := acd.accessControl.auditLogger.GetSuspiciousActivities(1)
	if err != nil {
		return fmt.Errorf("获取可疑活动失败: %w", err)
	}

	fmt.Printf("\n🚨 可疑活动数量: %d\n", len(suspicious))
	if len(suspicious) > 0 {
		fmt.Printf("   最新可疑活动: %s\n", suspicious[0].Reason)
	}

	fmt.Println()

	acd.logger.Info("审计日志演示完成")
	return nil
}