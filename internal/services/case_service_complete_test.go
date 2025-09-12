package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
	"law-oa-go/test"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCaseService_CreateCase(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	caseService := services.NewCaseService(testDB.DB)
	
	t.Run("successful case creation", func(t *testing.T) {
		req := &services.CreateCaseRequest{
			CaseNo:      "CASE-2024-001",
			CaseName:    "合同纠纷案",
			CaseType:    "contract",
			Priority:    "high",
			ClientID:    1,
			LawyerID:    2,
			Description: "客户与供应商之间的合同纠纷",
		}
		
		// 模拟检查案件号唯一性
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE case_no = ?").
			WithArgs(req.CaseNo).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// 模拟验证客户存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(req.ClientID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "role"}).AddRow(req.ClientID, "client"))
		
		// 模拟验证律师存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(req.LawyerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "role"}).AddRow(req.LawyerID, "lawyer"))
		
		// 模拟创建案件
		testDB.Mock.ExpectBegin()
		testDB.Mock.ExpectExec("INSERT INTO `cases`").
			WithArgs(
				sqlmock.AnyArg(), // ID
				req.CaseNo,
				req.CaseName,
				req.CaseType,
				req.Priority,
				req.ClientID,
				req.LawyerID,
				"draft", // status
				req.Description,
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		testDB.Mock.ExpectCommit()
		
		case_, err := caseService.CreateCase(context.Background(), req)
		
		require.NoError(t, err)
		assert.Equal(t, req.CaseNo, case_.CaseNo)
		assert.Equal(t, req.CaseName, case_.CaseName)
		assert.Equal(t, req.Priority, case_.Priority)
		assert.Equal(t, "draft", case_.Status)
	})
	
	t.Run("duplicate case number", func(t *testing.T) {
		req := &services.CreateCaseRequest{
			CaseNo:   "CASE-2024-001",
			CaseName: "测试案件",
			ClientID: 1,
			LawyerID: 2,
		}
		
		// 模拟案件号已存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE case_no = ?").
			WithArgs(req.CaseNo).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		
		_, err := caseService.CreateCase(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "case number already exists")
	})
	
	t.Run("client not found", func(t *testing.T) {
		req := &services.CreateCaseRequest{
			CaseNo:   "CASE-2024-002",
			CaseName: "测试案件",
			ClientID: 999,
			LawyerID: 2,
		}
		
		// 模拟案件号唯一
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE case_no = ?").
			WithArgs(req.CaseNo).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// 模拟客户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(req.ClientID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := caseService.CreateCase(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})
}

func TestCaseService_GetCase(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	caseService := services.NewCaseService(testDB.DB)
	
	t.Run("get existing case", func(t *testing.T) {
		caseID := uint(1)
		
		// 模拟案件查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE id = ?").
			WithArgs(caseID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "case_no", "case_name", "case_type", "status", "priority", "client_id", "lawyer_id", "description", "created_at", "updated_at"}).
				AddRow(caseID, "CASE-2024-001", "合同纠纷案", "contract", "active", "high", 1, 2, "合同纠纷描述", test.TestTime(), test.TestTime()))
		
		case_, err := caseService.GetCase(context.Background(), caseID)
		
		require.NoError(t, err)
		assert.Equal(t, caseID, case_.ID)
		assert.Equal(t, "CASE-2024-001", case_.CaseNo)
		assert.Equal(t, "合同纠纷案", case_.CaseName)
		assert.Equal(t, "contract", case_.CaseType)
	})
	
	t.Run("case not found", func(t *testing.T) {
		caseID := uint(999)
		
		// 模拟案件不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE id = ?").
			WithArgs(caseID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := caseService.GetCase(context.Background(), caseID)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "case not found")
	})
}

func TestCaseService_UpdateCase(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	caseService := services.NewCaseService(testDB.DB)
	
	t.Run("update case successfully", func(t *testing.T) {
		caseID := uint(1)
		req := &services.UpdateCaseRequest{
			CaseName:    test.StringPtr("更新后的案件名称"),
			CaseType:    test.StringPtr("tort"),
			Status:      test.StringPtr("in_progress"),
			Priority:    test.StringPtr("medium"),
			LawyerID:    test.UintPtr(3),
			Description: test.StringPtr("更新后的描述"),
		}
		
		// 模拟查询现有案件
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE id = ?").
			WithArgs(caseID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "case_no", "case_name", "status"}).
				AddRow(caseID, "CASE-2024-001", "原案件名称", "active"))
		
		// 模拟验证新律师存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(*req.LawyerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "role"}).AddRow(*req.LawyerID, "lawyer"))
		
		// 模拟更新
		testDB.Mock.ExpectExec("UPDATE `cases`").
			WithArgs(
				*req.CaseName,
				*req.CaseType,
				*req.Status,
				*req.Priority,
				*req.LawyerID,
				*req.Description,
				sqlmock.AnyArg(), // UpdatedAt
				caseID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		case_, err := caseService.UpdateCase(context.Background(), caseID, req)
		
		require.NoError(t, err)
		assert.Equal(t, *req.CaseName, case_.CaseName)
		assert.Equal(t, *req.CaseType, case_.CaseType)
		assert.Equal(t, *req.Status, case_.Status)
	})
	
	t.Run("case not found", func(t *testing.T) {
		caseID := uint(999)
		req := &services.UpdateCaseRequest{
			CaseName: test.StringPtr("更新的名称"),
		}
		
		// 模拟案件不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE id = ?").
			WithArgs(caseID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := caseService.UpdateCase(context.Background(), caseID, req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "case not found")
	})
}

func TestCaseService_ListCases(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	caseService := services.NewCaseService(testDB.DB)
	
	t.Run("list cases with filters", func(t *testing.T) {
		req := &services.ListCasesRequest{
			Page:     1,
			PageSize: 10,
			Status:   "active",
			CaseType: "contract",
			Priority: "high",
		}
		
		// 模拟查询总数
		testDB.Mock.ExpectQuery("SELECT COUNT(.+) FROM `cases`").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))
		
		// 模拟查询案件列表
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "case_no", "case_name", "case_type", "status", "priority", "client_id", "lawyer_id", "created_at"}).
				AddRow(1, "CASE-2024-001", "合同纠纷案1", "contract", "active", "high", 1, 2, test.TestTime()).
				AddRow(2, "CASE-2024-002", "合同纠纷案2", "contract", "active", "high", 3, 4, test.TestTime()))
		
		resp, err := caseService.ListCases(context.Background(), req)
		
		require.NoError(t, err)
		assert.Equal(t, int64(25), resp.Total)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
		assert.Len(t, resp.Cases, 2)
		assert.Equal(t, "CASE-2024-001", resp.Cases[0].CaseNo)
	})
	
	t.Run("empty result", func(t *testing.T) {
		req := &services.ListCasesRequest{
			Page:     1,
			PageSize: 10,
		}
		
		// 模拟查询总数
		testDB.Mock.ExpectQuery("SELECT COUNT(.+) FROM `cases`").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		
		// 模拟查询案件列表
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "case_no", "case_name", "case_type", "status", "priority", "client_id", "lawyer_id", "created_at"}))
		
		resp, err := caseService.ListCases(context.Background(), req)
		
		require.NoError(t, err)
		assert.Equal(t, int64(0), resp.Total)
		assert.Empty(t, resp.Cases)
	})
}

func TestCaseService_UpdateCaseStatus(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	caseService := services.NewCaseService(testDB.DB)
	
	t.Run("update status successfully", func(t *testing.T) {
		caseID := uint(1)
		newStatus := "closed"
		
		// 模拟查询现有案件
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE id = ?").
			WithArgs(caseID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "case_no", "status"}).
				AddRow(caseID, "CASE-2024-001", "active"))
		
		// 模拟状态验证（检查是否允许从active到closed）
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `case_status_transitions`").
			WithArgs("active", "closed").
			WillReturnRows(sqlmock.NewRows([]string{"allowed"}).AddRow(true))
		
		// 模拟更新状态
		testDB.Mock.ExpectExec("UPDATE `cases`").
			WithArgs(newStatus, sqlmock.AnyArg(), caseID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		err := caseService.UpdateCaseStatus(context.Background(), caseID, newStatus)
		
		require.NoError(t, err)
	})
	
	t.Run("invalid status transition", func(t *testing.T) {
		caseID := uint(1)
		newStatus := "invalid_status"
		
		// 模拟查询现有案件
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `cases` WHERE id = ?").
			WithArgs(caseID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "case_no", "status"}).
				AddRow(caseID, "CASE-2024-001", "active"))
		
		// 模拟状态验证失败
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `case_status_transitions`").
			WithArgs("active", newStatus).
			WillReturnError(gorm.ErrRecordNotFound)
		
		err := caseService.UpdateCaseStatus(context.Background(), caseID, newStatus)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")
	})
}