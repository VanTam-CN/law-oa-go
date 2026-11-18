package services

import (
	"context"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDelegationDB 创建测试数据库
func setupTestDelegationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&models.ApprovalDelegation{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	return db
}

// TestCreateDelegation_NoSelfDelegation 测试不能自己代理自己
func TestCreateDelegation_NoSelfDelegation(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	req := &CreateDelegationRequest{
		DelegatorID: "user-1",
		DelegateID:  "user-1", // 同一个用户
		ValidFrom:   time.Now(),
		CreatedBy:   "admin",
	}

	_, err := svc.CreateDelegation(context.Background(), req)
	if err == nil {
		t.Fatal("期望返回错误，但成功创建")
	}
	if err.Error() != "不能为自己配置代理审批" {
		t.Fatalf("错误消息不匹配: %s", err.Error())
	}
}

// TestCreateDelegation_InvalidTimeRange 测试时间范围验证
func TestCreateDelegation_InvalidTimeRange(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	now := time.Now()
	past := now.Add(-24 * time.Hour)

	req := &CreateDelegationRequest{
		DelegatorID: "user-1",
		DelegateID:  "user-2",
		ValidFrom:   now,
		ValidUntil:  &past, // 结束时间早于开始时间
		CreatedBy:   "admin",
	}

	_, err := svc.CreateDelegation(context.Background(), req)
	if err == nil {
		t.Fatal("期望返回错误，但成功创建")
	}
}

// TestCreateDelegation_CircularDelegation 测试循环代理检测
func TestCreateDelegation_CircularDelegation(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	ctx := context.Background()
	now := time.Now()

	// 先创建 A→B 的代理
	err := repo.Create(ctx, &models.ApprovalDelegation{
		ID:          uuid.New().String(),
		DelegatorID: "user-2",
		DelegateID:  "user-1",
		ValidFrom:   now,
		IsActive:    true,
		CreatedBy:   "admin",
	})
	if err != nil {
		t.Fatalf("创建初始代理失败: %v", err)
	}

	// 尝试创建 B→A 的代理（循环）
	req := &CreateDelegationRequest{
		DelegatorID: "user-1",
		DelegateID:  "user-2",
		ValidFrom:   now,
		CreatedBy:   "admin",
	}

	_, err = svc.CreateDelegation(ctx, req)
	if err == nil {
		t.Fatal("期望返回循环代理错误，但成功创建")
	}
	if err.Error() != "不允许循环代理: 形成闭环代理链" {
		t.Fatalf("错误消息不匹配: %s", err.Error())
	}
}

// TestCreateDelegation_Success 测试正常创建
func TestCreateDelegation_Success(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	ctx := context.Background()
	req := &CreateDelegationRequest{
		DelegatorID: "user-1",
		DelegateID:  "user-2",
		ValidFrom:   time.Now(),
		Reason:      "出差一周",
		CreatedBy:   "admin",
	}

	d, err := svc.CreateDelegation(ctx, req)
	if err != nil {
		t.Fatalf("创建代理配置失败: %v", err)
	}
	if d.DelegatorID != "user-1" {
		t.Errorf("DelegatorID 不匹配: %s", d.DelegatorID)
	}
	if d.DelegateID != "user-2" {
		t.Errorf("DelegateID 不匹配: %s", d.DelegateID)
	}
	if d.Reason != "出差一周" {
		t.Errorf("Reason 不匹配: %s", d.Reason)
	}
}

// TestRevokeDelegation 测试撤销代理
func TestRevokeDelegation(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	ctx := context.Background()

	// 先创建代理
	d, err := svc.CreateDelegation(ctx, &CreateDelegationRequest{
		DelegatorID: "user-1",
		DelegateID:  "user-2",
		ValidFrom:   time.Now(),
		CreatedBy:   "admin",
	})
	if err != nil {
		t.Fatalf("创建代理配置失败: %v", err)
	}

	// 撤销
	err = svc.RevokeDelegation(ctx, d.ID)
	if err != nil {
		t.Fatalf("撤销代理配置失败: %v", err)
	}

	// 验证已撤销
	active, err := svc.GetActiveDelegations(ctx, "user-1")
	if err != nil {
		t.Fatalf("获取代理配置失败: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("撤销后仍存在活跃代理配置，数量: %d", len(active))
	}
}

// TestGetEffectiveApprover_WithDelegation 测试有代理时返回代理人
func TestGetEffectiveApprover_WithDelegation(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	ctx := context.Background()

	// 创建代理
	err := repo.Create(ctx, &models.ApprovalDelegation{
		ID:          uuid.New().String(),
		DelegatorID: "original-approver",
		DelegateID:  "delegate-approver",
		ValidFrom:   time.Now().Add(-1 * time.Hour),
		IsActive:    true,
		CreatedBy:   "admin",
	})
	if err != nil {
		t.Fatalf("创建代理配置失败: %v", err)
	}

	effectiveID, _, isDelegated, err := svc.GetEffectiveApprover(ctx, "original-approver")
	if err != nil {
		t.Fatalf("获取有效审批人失败: %v", err)
	}
	if !isDelegated {
		t.Fatal("期望返回代理标记为true")
	}
	if effectiveID != "delegate-approver" {
		t.Errorf("代理人ID不匹配: %s", effectiveID)
	}
}

// TestGetEffectiveApprover_NoDelegation 测试无代理时返回原审批人
func TestGetEffectiveApprover_NoDelegation(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	ctx := context.Background()

	effectiveID, _, isDelegated, err := svc.GetEffectiveApprover(ctx, "original-approver")
	if err != nil {
		t.Fatalf("获取有效审批人失败: %v", err)
	}
	if isDelegated {
		t.Fatal("无代理时不应返回代理标记")
	}
	if effectiveID != "original-approver" {
		t.Errorf("应返回原审批人: %s", effectiveID)
	}
}

// TestGetEffectiveApprover_ExpiredDelegation 测试过期代理不生效
func TestGetEffectiveApprover_ExpiredDelegation(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	ctx := context.Background()
	now := time.Now()

	// 创建已过期的代理
	err := repo.Create(ctx, &models.ApprovalDelegation{
		ID:          uuid.New().String(),
		DelegatorID: "original-approver",
		DelegateID:  "delegate-approver",
		ValidFrom:   now.Add(-48 * time.Hour),
		ValidUntil:  ptrTime(now.Add(-1 * time.Hour)),
		IsActive:    true,
		CreatedBy:   "admin",
	})
	if err != nil {
		t.Fatalf("创建过期代理配置失败: %v", err)
	}

	effectiveID, _, isDelegated, err := svc.GetEffectiveApprover(ctx, "original-approver")
	if err != nil {
		t.Fatalf("获取有效审批人失败: %v", err)
	}
	if isDelegated {
		t.Fatal("过期代理不应生效")
	}
	if effectiveID != "original-approver" {
		t.Errorf("过期代理应返回原审批人: %s", effectiveID)
	}
}

// TestListDelegations 测试列表查询
func TestListDelegations(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	ctx := context.Background()

	// 创建多个代理
	for i := 0; i < 3; i++ {
		err := repo.Create(ctx, &models.ApprovalDelegation{
			ID:          uuid.New().String(),
			DelegatorID: "user-1",
			DelegateID:  "user-2",
			ValidFrom:   time.Now(),
			IsActive:    true,
			CreatedBy:   "admin",
		})
		if err != nil {
			t.Fatalf("创建代理配置失败: %v", err)
		}
	}

	delegations, total, err := svc.ListDelegations(ctx, &repositories.DelegationListParams{
		DelegatorID: "user-1",
		Page:        1,
		PageSize:    10,
	})
	if err != nil {
		t.Fatalf("查询代理配置失败: %v", err)
	}
	if total != 3 {
		t.Errorf("总数不匹配: 期望3，实际%d", total)
	}
	if len(delegations) != 3 {
		t.Errorf("返回数量不匹配: 期望3，实际%d", len(delegations))
	}
}

// TestRevokeDelegation_NotFound 测试撤销不存在的代理
func TestRevokeDelegation_NotFound(t *testing.T) {
	db := setupTestDelegationDB(t)
	repo := repositories.NewApprovalDelegationRepository(db)
	svc := NewApprovalDelegationService(repo)

	err := svc.RevokeDelegation(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("撤销不存在的代理应返回错误")
	}
}

// ptrTime 辅助函数：创建时间指针
func ptrTime(t time.Time) *time.Time {
	return &t
}
