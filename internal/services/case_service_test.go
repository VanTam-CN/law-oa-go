package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func TestCaseService_CreateCase(t *testing.T) {
	db := setupTestDB(t)
	service := NewCaseService(db)
	ctx := context.Background()

	// 创建测试数据
	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
		Phone: "1234567890",
	}
	db.Create(client)

	lawyer := &models.User{
		Name:  "Test Lawyer",
		Email: "lawyer@example.com",
		Role:  "lawyer",
	}
	db.Create(lawyer)

	tests := []struct {
		name    string
		req     *CreateCaseRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid case creation",
			req: &CreateCaseRequest{
				Title:       "Test Case",
				Description: "Test case description",
				ClientID:    client.ID,
				LawyerID:    lawyer.ID,
				CaseType:    "civil",
				Priority:    "medium",
			},
			wantErr: false,
		},
		{
			name: "client not found",
			req: &CreateCaseRequest{
				Title:       "Test Case",
				Description: "Test case description",
				ClientID:    999,
				LawyerID:    lawyer.ID,
				CaseType:    "civil",
				Priority:    "medium",
			},
			wantErr: true,
			errMsg:  "client not found",
		},
		{
			name: "lawyer not found",
			req: &CreateCaseRequest{
				Title:       "Test Case",
				Description: "Test case description",
				ClientID:    client.ID,
				LawyerID:    999,
				CaseType:    "civil",
				Priority:    "medium",
			},
			wantErr: true,
			errMsg:  "lawyer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caseResp, err := service.CreateCase(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, caseResp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, caseResp)
				assert.Equal(t, tt.req.Title, caseResp.Title)
				assert.Equal(t, tt.req.ClientID, caseResp.ClientID)
				assert.Equal(t, tt.req.LawyerID, caseResp.LawyerID)
				assert.Equal(t, "pending", caseResp.Status) // 默认状态
			}
		})
	}
}

func TestCaseService_GetCaseByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewCaseService(db)
	ctx := context.Background()

	// 创建测试数据
	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
		Phone: "1234567890",
	}
	db.Create(client)

	lawyer := &models.User{
		Name:  "Test Lawyer",
		Email: "lawyer@example.com",
		Role:  "lawyer",
	}
	db.Create(lawyer)

	testCase := &models.Case{
		Title:       "Test Case",
		Description: "Test description",
		ClientID:    client.ID,
		LawyerID:    lawyer.ID,
		CaseType:    "civil",
		Priority:    "medium",
		Status:      "pending",
	}
	db.Create(testCase)

	tests := []struct {
		name    string
		caseID  uint
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid case ID",
			caseID:  testCase.ID,
			wantErr: false,
		},
		{
			name:    "case not found",
			caseID:  999,
			wantErr: true,
			errMsg:  "case not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caseResp, err := service.GetCaseByID(ctx, tt.caseID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, caseResp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, caseResp)
				assert.Equal(t, tt.caseID, caseResp.ID)
				assert.Equal(t, testCase.Title, caseResp.Title)
				assert.Equal(t, client.Name, caseResp.ClientName)
				assert.Equal(t, lawyer.Name, caseResp.LawyerName)
			}
		})
	}
}

func TestCaseService_UpdateCase(t *testing.T) {
	db := setupTestDB(t)
	service := NewCaseService(db)
	ctx := context.Background()

	// 创建测试数据
	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
		Phone: "1234567890",
	}
	db.Create(client)

	lawyer1 := &models.User{
		Name:  "Test Lawyer 1",
		Email: "lawyer1@example.com",
		Role:  "lawyer",
	}
	db.Create(lawyer1)

	lawyer2 := &models.User{
		Name:  "Test Lawyer 2",
		Email: "lawyer2@example.com",
		Role:  "lawyer",
	}
	db.Create(lawyer2)

	testCase := &models.Case{
		Title:       "Original Title",
		Description: "Original description",
		ClientID:    client.ID,
		LawyerID:    lawyer1.ID,
		CaseType:    "civil",
		Priority:    "medium",
		Status:      "pending",
	}
	db.Create(testCase)

	tests := []struct {
		name    string
		caseID  uint
		req     *UpdateCaseRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid update",
			caseID: testCase.ID,
			req: &UpdateCaseRequest{
				Title:    stringPtr("Updated Title"),
				LawyerID: &lawyer2.ID,
				Status:   stringPtr("active"),
			},
			wantErr: false,
		},
		{
			name:   "case not found",
			caseID: 999,
			req: &UpdateCaseRequest{
				Title: stringPtr("Updated Title"),
			},
			wantErr: true,
			errMsg:  "case not found",
		},
		{
			name:   "lawyer not found",
			caseID: testCase.ID,
			req: &UpdateCaseRequest{
				LawyerID: uintPtr(999),
			},
			wantErr: true,
			errMsg:  "lawyer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caseResp, err := service.UpdateCase(ctx, tt.caseID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, caseResp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, caseResp)
				if tt.req.Title != nil {
					assert.Equal(t, *tt.req.Title, caseResp.Title)
				}
				if tt.req.LawyerID != nil {
					assert.Equal(t, *tt.req.LawyerID, caseResp.LawyerID)
				}
				if tt.req.Status != nil {
					assert.Equal(t, *tt.req.Status, caseResp.Status)
				}
			}
		})
	}
}

func TestCaseService_ListCases(t *testing.T) {
	db := setupTestDB(t)
	service := NewCaseService(db)
	ctx := context.Background()

	// 创建测试数据
	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
		Phone: "1234567890",
	}
	db.Create(client)

	lawyer := &models.User{
		Name:  "Test Lawyer",
		Email: "lawyer@example.com",
		Role:  "lawyer",
	}
	db.Create(lawyer)

	// 创建多个案件
	cases := []*models.Case{
		{
			Title:    "Case 1",
			ClientID: client.ID,
			LawyerID: lawyer.ID,
			CaseType: "civil",
			Priority: "high",
			Status:   "pending",
		},
		{
			Title:    "Case 2",
			ClientID: client.ID,
			LawyerID: lawyer.ID,
			CaseType: "criminal",
			Priority: "medium",
			Status:   "active",
		},
		{
			Title:    "Case 3",
			ClientID: client.ID,
			LawyerID: lawyer.ID,
			CaseType: "civil",
			Priority: "low",
			Status:   "closed",
		},
	}

	for _, c := range cases {
		db.Create(c)
	}

	tests := []struct {
		name          string
		req           *CaseListRequest
		expectedCount int
	}{
		{
			name:          "list all cases",
			req:           &CaseListRequest{},
			expectedCount: 3,
		},
		{
			name: "filter by status",
			req: &CaseListRequest{
				Status: "pending",
			},
			expectedCount: 1,
		},
		{
			name: "filter by case type",
			req: &CaseListRequest{
				CaseType: "civil",
			},
			expectedCount: 2,
		},
		{
			name: "filter by priority",
			req: &CaseListRequest{
				Priority: "high",
			},
			expectedCount: 1,
		},
		{
			name: "pagination",
			req: &CaseListRequest{
				Page:     1,
				PageSize: 2,
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cases, total, err := service.ListCases(ctx, tt.req)

			assert.NoError(t, err)
			assert.Len(t, cases, tt.expectedCount)
			
			if tt.req.Page == 0 && tt.req.PageSize == 0 {
				assert.Equal(t, int64(tt.expectedCount), total)
			}
		})
	}
}

func TestCaseService_GetCaseStats(t *testing.T) {
	db := setupTestDB(t)
	service := NewCaseService(db)
	ctx := context.Background()

	// 创建测试数据
	client := &models.Client{
		Name:  "Test Client",
		Email: "client@example.com",
		Phone: "1234567890",
	}
	db.Create(client)

	lawyer := &models.User{
		Name:  "Test Lawyer",
		Email: "lawyer@example.com",
		Role:  "lawyer",
	}
	db.Create(lawyer)

	// 创建不同状态和优先级的案件
	cases := []*models.Case{
		{Title: "Case 1", ClientID: client.ID, LawyerID: lawyer.ID, Status: "pending", Priority: "high"},
		{Title: "Case 2", ClientID: client.ID, LawyerID: lawyer.ID, Status: "active", Priority: "medium"},
		{Title: "Case 3", ClientID: client.ID, LawyerID: lawyer.ID, Status: "closed", Priority: "low"},
		{Title: "Case 4", ClientID: client.ID, LawyerID: lawyer.ID, Status: "pending", Priority: "urgent"},
		{Title: "Case 5", ClientID: client.ID, LawyerID: lawyer.ID, Status: "suspended", Priority: "high"},
	}

	for _, c := range cases {
		db.Create(c)
	}

	stats, err := service.GetCaseStats(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(5), stats.TotalCases)
	assert.Equal(t, int64(2), stats.PendingCases)
	assert.Equal(t, int64(1), stats.ActiveCases)
	assert.Equal(t, int64(1), stats.ClosedCases)
	assert.Equal(t, int64(1), stats.SuspendedCases)
	assert.Equal(t, int64(2), stats.HighPriority)
	assert.Equal(t, int64(1), stats.UrgentCases)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func uintPtr(u uint) *uint {
	return &u
}