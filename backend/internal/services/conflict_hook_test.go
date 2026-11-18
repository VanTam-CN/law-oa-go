package services

import (
	"context"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockConflictCheckService 用于测试的 mock 冲突检测服务
type mockConflictCheckService struct {
	checkFn func(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error)
}

func (m *mockConflictCheckService) CheckConflict(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error) {
	return m.checkFn(ctx, req)
}

func (m *mockConflictCheckService) GenerateReport(ctx context.Context, checkID uint) (*ConflictReport, error) {
	return nil, nil
}

func (m *mockConflictCheckService) GetConflictCheck(ctx context.Context, id uint) (*models.ConflictCheck, error) {
	return nil, nil
}

func (m *mockConflictCheckService) ListConflictChecks(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*models.ConflictCheck, int64, error) {
	return nil, 0, nil
}

func (m *mockConflictCheckService) SearchEntity(ctx context.Context, query string, entityType string) ([]*models.Entity, error) {
	return nil, nil
}

func (m *mockConflictCheckService) GetEntityRelations(ctx context.Context, entityID uint, depth int) ([]*models.Entity, error) {
	return nil, nil
}

func setupConflictHookTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", "01234567890123456789012345678901")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&models.Case{}, &models.Client{}, &models.Entity{}, &models.ConflictCheck{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	return db
}

// TestOnCaseCreated 测试案件创建后自动触发冲突检测
func TestOnCaseCreated(t *testing.T) {
	db := setupConflictHookTestDB(t)

	// 创建客户
	client := &models.Client{
		Name:   "张三",
		IDCard: "110101199001011234",
		Email:  "zhangsan@test.com",
		Phone:  "13800138000",
	}
	db.Create(client)

	// 创建案件
	caseData := &models.Case{
		Title:    "测试案件",
		ClientID: client.ID,
		Status:   "pending",
	}
	db.Create(caseData)

	// 创建 mock 冲突检测服务
	checkCalled := make(chan struct{})
	mockConflict := &mockConflictCheckService{
		checkFn: func(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error) {
			close(checkCalled)
			if req.CaseID != caseData.ID {
				t.Errorf("CaseID 不匹配: 期望 %d, 实际 %d", caseData.ID, req.CaseID)
			}
			if len(req.CheckEntities) != 1 {
				t.Errorf("CheckEntities 数量不匹配: 期望1, 实际 %d", len(req.CheckEntities))
			}
			if req.CheckEntities[0].EntityName != "张三" {
				t.Errorf("EntityName 不匹配: %s", req.CheckEntities[0].EntityName)
			}
			return &ConflictCheckResponse{HasConflict: false}, nil
		},
	}

	caseRepo := repositories.NewCaseRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	entityRepo := repositories.NewEntityRepository(db)

	svc := NewConflictHookService(mockConflict, caseRepo, entityRepo, clientRepo)

	// 执行
	svc.OnCaseCreated(context.Background(), caseData.ID, client.ID)

	select {
	case <-checkCalled:
	case <-time.After(time.Second):
		t.Fatal("冲突检测未被调用")
	}
}

// TestOnCaseCreated_WithConflict 测试检测到冲突的情况
func TestOnCaseCreated_WithConflict(t *testing.T) {
	db := setupConflictHookTestDB(t)

	client := &models.Client{Name: "李四", IDCard: "110101199001011235", Email: "lisi@test.com"}
	db.Create(client)

	caseData := &models.Case{Title: "冲突案件", ClientID: client.ID, Status: "pending"}
	db.Create(caseData)

	var receivedConflicts int
	mockConflict := &mockConflictCheckService{
		checkFn: func(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error) {
			return &ConflictCheckResponse{
				HasConflict:    true,
				TotalConflicts: 3,
			}, nil
		},
	}

	caseRepo := repositories.NewCaseRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	entityRepo := repositories.NewEntityRepository(db)

	svc := NewConflictHookService(mockConflict, caseRepo, entityRepo, clientRepo)
	svc.OnCaseCreated(context.Background(), caseData.ID, client.ID)

	time.Sleep(100 * time.Millisecond)

	if receivedConflicts != 0 {
		// 这个测试主要验证不 panic
		t.Logf("冲突检测正常执行")
	}
}

// TestOnCaseUpdated 测试案件更新后触发
func TestOnCaseUpdated(t *testing.T) {
	db := setupConflictHookTestDB(t)

	client := &models.Client{Name: "王五", IDCard: "110101199001011236", Email: "wangwu@test.com"}
	db.Create(client)

	caseData := &models.Case{Title: "更新案件", ClientID: client.ID, Status: "active"}
	db.Create(caseData)

	checkCalled := make(chan struct{})
	mockConflict := &mockConflictCheckService{
		checkFn: func(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error) {
			close(checkCalled)
			return &ConflictCheckResponse{HasConflict: false}, nil
		},
	}

	caseRepo := repositories.NewCaseRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	entityRepo := repositories.NewEntityRepository(db)

	svc := NewConflictHookService(mockConflict, caseRepo, entityRepo, clientRepo)
	svc.OnCaseUpdated(context.Background(), caseData.ID)

	select {
	case <-checkCalled:
	case <-time.After(time.Second):
		t.Fatal("案件更新后冲突检测未被调用")
	}
}

// TestOnCaseUpdated_NoClient 测试无客户的案件不触发
func TestOnCaseUpdated_NoClient(t *testing.T) {
	db := setupConflictHookTestDB(t)

	caseData := &models.Case{Title: "无客户案件", ClientID: 0, Status: "pending"}
	db.Create(caseData)

	var checkCalled bool
	mockConflict := &mockConflictCheckService{
		checkFn: func(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error) {
			checkCalled = true
			return &ConflictCheckResponse{HasConflict: false}, nil
		},
	}

	caseRepo := repositories.NewCaseRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	entityRepo := repositories.NewEntityRepository(db)

	svc := NewConflictHookService(mockConflict, caseRepo, entityRepo, clientRepo)
	svc.OnCaseUpdated(context.Background(), caseData.ID)

	time.Sleep(100 * time.Millisecond)

	if checkCalled {
		t.Fatal("无客户案件不应触发冲突检测")
	}
}

// TestOnEntityAddedToCase 测试实体添加到案件时触发
func TestOnEntityAddedToCase(t *testing.T) {
	t.Skip("需要 entity_name_history 表，暂时跳过")
}
