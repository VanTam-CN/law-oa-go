package casemgmt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa/test"
)

// CaseServiceTest 案件服务测试
func TestCaseService(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("CreateCase", func(t *testing.T) {
		// 测试创建案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		case := &models.Case{
			Title:       "Test Civil Case",
			Description: "This is a test civil case",
			CaseNumber:  "TC-001",
			CaseType:    "Civil",
			Status:      "active",
			Priority:    "medium",
			LawyerID:    lawyer.ID,
			ClientID:    client.ID,
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 6, 0),
			Budget:      100000.00,
		}
		
		createdCase, err := caseService.CreateCase(case)
		require.NoError(t, err)
		assert.NotZero(t, createdCase.ID)
		assert.Equal(t, case.Title, createdCase.Title)
		assert.Equal(t, case.Description, createdCase.Description)
		assert.Equal(t, case.CaseNumber, createdCase.CaseNumber)
		assert.Equal(t, case.CaseType, createdCase.CaseType)
	})
	
	t.Run("GetCase", func(t *testing.T) {
		// 测试获取案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		retrievedCase, err := caseService.GetCase(case.ID)
		require.NoError(t, err)
		assert.Equal(t, case.ID, retrievedCase.ID)
		assert.Equal(t, case.Title, retrievedCase.Title)
		assert.Equal(t, case.CaseNumber, retrievedCase.CaseNumber)
	})
	
	t.Run("GetCaseNotFound", func(t *testing.T) {
		// 测试获取不存在的案件
		_, err := caseService.GetCase(99999)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
	
	t.Run("UpdateCase", func(t *testing.T) {
		// 测试更新案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		updatedCase := *case
		updatedCase.Title = "Updated Title"
		updatedCase.Status = "completed"
		updatedCase.Priority = "high"
		
		result, err := caseService.UpdateCase(case.ID, &updatedCase)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", result.Title)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, "high", result.Priority)
	})
	
	t.Run("DeleteCase", func(t *testing.T) {
		// 测试删除案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		err := caseService.DeleteCase(case.ID)
		require.NoError(t, err)
		
		// 验证案件已被删除
		_, err = caseService.GetCase(case.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
	
	t.Run("ListCases", func(t *testing.T) {
		// 测试案件列表
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 5; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		cases, total, err := caseService.ListCases(1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 5)
		assert.GreaterOrEqual(t, total, 5)
	})
	
	t.Run("SearchCases", func(t *testing.T) {
		// 测试搜索案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		searchTerm := case.Title
		cases, err := caseService.SearchCases(searchTerm)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 1)
	})
	
	t.Run("GetCasesByLawyer", func(t *testing.T) {
		// 测试根据律师获取案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		cases, err := caseService.GetCasesByLawyer(lawyer.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 3)
	})
	
	t.Run("GetCasesByClient", func(t *testing.T) {
		// 测试根据客户获取案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		cases, err := caseService.GetCasesByClient(client.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 3)
	})
	
	t.Run("GetCaseDocuments", func(t *testing.T) {
		// 测试获取案件文档
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		// 创建测试文档
		document := &models.Document{
			Title:      "Test Document",
			FileName:   "test.pdf",
			FilePath:   "/path/to/test.pdf",
			FileSize:   1024,
			CaseID:     case.ID,
			UploadedBy: lawyer.ID,
		}
		err := suite.DB.Create(document).Error
		require.NoError(t, err)
		
		documents, err := caseService.GetCaseDocuments(case.ID)
		require.NoError(t, err)
		assert.Len(t, documents, 1)
		assert.Equal(t, document.ID, documents[0].ID)
	})
	
	t.Run("GetCaseTimeline", func(t *testing.T) {
		// 测试获取案件时间线
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		// 创建测试时间线事件
		event := &models.CaseEvent{
			CaseID:  case.ID,
			EventType: "created",
			Title:    "Case Created",
			Description: "Case was created",
			EventDate: time.Now(),
			UserID:   lawyer.ID,
		}
		err := suite.DB.Create(event).Error
		require.NoError(t, err)
		
		timeline, err := caseService.GetCaseTimeline(case.ID)
		require.NoError(t, err)
		assert.Len(t, timeline, 1)
		assert.Equal(t, event.ID, timeline[0].ID)
	})
	
	t.Run("GetCaseStats", func(t *testing.T) {
		// 测试获取案件统计
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 5; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		stats, err := caseService.GetCaseStats()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalCases, 5)
		assert.GreaterOrEqual(t, stats.ActiveCases, 5)
	})
}

// CaseHandlerTest 案件处理器测试
func TestCaseHandler(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	handler := NewCaseHandler(suite.DB)
	
	// 设置路由
	router := gin.New()
	router.POST("/cases", handler.CreateCase)
	router.GET("/cases/:id", handler.GetCase)
	router.PUT("/cases/:id", handler.UpdateCase)
	router.DELETE("/cases/:id", handler.DeleteCase)
	router.GET("/cases", handler.ListCases)
	router.GET("/cases/search", handler.SearchCases)
	router.GET("/cases/lawyer/:lawyer_id", handler.GetCasesByLawyer)
	router.GET("/cases/client/:client_id", handler.GetCasesByClient)
	router.GET("/cases/:id/documents", handler.GetCaseDocuments)
	router.GET("/cases/:id/timeline", handler.GetCaseTimeline)
	router.GET("/cases/stats", handler.GetCaseStats)
	
	suite.Router = router
	
	t.Run("CreateCaseAPI", func(t *testing.T) {
		// 测试创建案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		case := map[string]interface{}{
			"title":       "Test Criminal Case",
			"description": "This is a test criminal case",
			"case_number": "TC-002",
			"case_type":   "Criminal",
			"status":      "active",
			"priority":    "high",
			"lawyer_id":   lawyer.ID,
			"client_id":   client.ID,
			"start_date":  time.Now().Format("2006-01-02"),
			"budget":      50000.00,
		}
		
		w := suite.PerformAuthRequest("POST", "/cases", case)
		data := suite.AssertSuccess(t, w)
		
		assert.NotZero(t, data["id"])
		assert.Equal(t, "Test Criminal Case", data["title"])
		assert.Equal(t, "TC-002", data["case_number"])
		assert.Equal(t, "Criminal", data["case_type"])
	})
	
	t.Run("GetCaseAPI", func(t *testing.T) {
		// 测试获取案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		w := suite.PerformAuthRequest("GET", "/cases/"+string(case.ID), nil)
		data := suite.AssertSuccess(t, w)
		
		assert.Equal(t, case.ID, uint(data["id"].(float64)))
		assert.Equal(t, case.Title, data["title"])
		assert.Equal(t, case.CaseNumber, data["case_number"])
	})
	
	t.Run("UpdateCaseAPI", func(t *testing.T) {
		// 测试更新案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		updateData := map[string]interface{}{
			"title":    "Updated Title",
			"status":   "completed",
			"priority": "low",
		}
		
		w := suite.PerformAuthRequest("PUT", "/cases/"+string(case.ID), updateData)
		data := suite.AssertSuccess(t, w)
		
		assert.Equal(t, "Updated Title", data["title"])
		assert.Equal(t, "completed", data["status"])
		assert.Equal(t, "low", data["priority"])
	})
	
	t.Run("DeleteCaseAPI", func(t *testing.T) {
		// 测试删除案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		w := suite.PerformAuthRequest("DELETE", "/cases/"+string(case.ID), nil)
		suite.AssertSuccess(t, w)
		
		// 验证案件已被删除
		w = suite.PerformAuthRequest("GET", "/cases/"+string(case.ID), nil)
		suite.AssertError(t, w, 404)
	})
	
	t.Run("ListCasesAPI", func(t *testing.T) {
		// 测试案件列表API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		w := suite.PerformAuthRequest("GET", "/cases?page=1&size=10", nil)
		data := suite.AssertSuccess(t, w)
		
		cases := data["cases"].([]interface{})
		total := data["total"].(float64)
		
		assert.GreaterOrEqual(t, len(cases), 3)
		assert.GreaterOrEqual(t, total, 3)
	})
	
	t.Run("SearchCasesAPI", func(t *testing.T) {
		// 测试搜索案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		w := suite.PerformAuthRequest("GET", "/cases/search?q="+case.Title, nil)
		data := suite.AssertSuccess(t, w)
		
		cases := data["cases"].([]interface{})
		assert.GreaterOrEqual(t, len(cases), 1)
	})
	
	t.Run("GetCasesByLawyerAPI", func(t *testing.T) {
		// 测试根据律师获取案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 2; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		w := suite.PerformAuthRequest("GET", "/cases/lawyer/"+string(lawyer.ID), nil)
		data := suite.AssertSuccess(t, w)
		
		cases := data["cases"].([]interface{})
		assert.GreaterOrEqual(t, len(cases), 2)
	})
	
	t.Run("GetCasesByClientAPI", func(t *testing.T) {
		// 测试根据客户获取案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 2; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		w := suite.PerformAuthRequest("GET", "/cases/client/"+string(client.ID), nil)
		data := suite.AssertSuccess(t, w)
		
		cases := data["cases"].([]interface{})
		assert.GreaterOrEqual(t, len(cases), 2)
	})
	
	t.Run("GetCaseDocumentsAPI", func(t *testing.T) {
		// 测试获取案件文档API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		// 创建测试文档
		document := &models.Document{
			Title:      "Test Document",
			FileName:   "test.pdf",
			FilePath:   "/path/to/test.pdf",
			FileSize:   1024,
			CaseID:     case.ID,
			UploadedBy: lawyer.ID,
		}
		err := suite.DB.Create(document).Error
		require.NoError(t, err)
		
		w := suite.PerformAuthRequest("GET", "/cases/"+string(case.ID)+"/documents", nil)
		data := suite.AssertSuccess(t, w)
		
		documents := data["documents"].([]interface{})
		assert.Len(t, documents, 1)
		assert.Equal(t, document.ID, uint(documents[0].(map[string]interface{})["id"].(float64)))
	})
	
	t.Run("GetCaseTimelineAPI", func(t *testing.T) {
		// 测试获取案件时间线API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		// 创建测试时间线事件
		event := &models.CaseEvent{
			CaseID:     case.ID,
			EventType:  "created",
			Title:      "Case Created",
			Description: "Case was created",
			EventDate:  time.Now(),
			UserID:     lawyer.ID,
		}
		err := suite.DB.Create(event).Error
		require.NoError(t, err)
		
		w := suite.PerformAuthRequest("GET", "/cases/"+string(case.ID)+"/timeline", nil)
		data := suite.AssertSuccess(t, w)
		
		timeline := data["timeline"].([]interface{})
		assert.Len(t, timeline, 1)
		assert.Equal(t, event.ID, uint(timeline[0].(map[string]interface{})["id"].(float64)))
	})
	
	t.Run("GetCaseStatsAPI", func(t *testing.T) {
		// 测试获取案件统计API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		w := suite.PerformAuthRequest("GET", "/cases/stats", nil)
		data := suite.AssertSuccess(t, w)
		
		assert.GreaterOrEqual(t, data["total_cases"].(float64), 3)
		assert.GreaterOrEqual(t, data["active_cases"].(float64), 3)
	})
	
	t.Run("CreateCaseValidation", func(t *testing.T) {
		// 测试创建案件验证
		invalidCase := map[string]interface{}{
			"title":       "", // 空标题
			"case_number": "", // 空案件编号
			"case_type":   "", // 空案件类型
		}
		
		w := suite.PerformAuthRequest("POST", "/cases", invalidCase)
		suite.AssertError(t, w, 400)
	})
	
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		// 测试未授权访问
		w := suite.PerformRequest("GET", "/cases", nil, nil)
		suite.AssertError(t, w, 401)
	})
}

// CaseValidatorTest 案件验证器测试
func TestCaseValidator(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	validator := NewCaseValidator()
	
	t.Run("ValidCase", func(t *testing.T) {
		// 测试有效案件验证
		case := &models.Case{
			Title:      "Test Case",
			CaseNumber: "TC-001",
			CaseType:   "Civil",
			Status:     "active",
			Priority:   "medium",
			LawyerID:   1,
			ClientID:   1,
			StartDate:  time.Now(),
		}
		
		err := validator.ValidateCase(case)
		assert.NoError(t, err)
	})
	
	t.Run("MissingRequiredFields", func(t *testing.T) {
		// 测试缺失必填字段
		case := &models.Case{
			Title:      "",
			CaseNumber: "",
			CaseType:   "",
		}
		
		err := validator.ValidateCase(case)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title")
		assert.Contains(t, err.Error(), "case_number")
		assert.Contains(t, err.Error(), "case_type")
	})
	
	t.Run("InvalidCaseType", func(t *testing.T) {
		// 测试无效案件类型
		case := &models.Case{
			Title:      "Test Case",
			CaseNumber: "TC-001",
			CaseType:   "InvalidType", // 无效类型
			Status:     "active",
			Priority:   "medium",
			LawyerID:   1,
			ClientID:   1,
		}
		
		err := validator.ValidateCase(case)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "case_type")
	})
	
	t.Run("InvalidStatus", func(t *testing.T) {
		// 测试无效状态
		case := &models.Case{
			Title:      "Test Case",
			CaseNumber: "TC-001",
			CaseType:   "Civil",
			Status:     "InvalidStatus", // 无效状态
			Priority:   "medium",
			LawyerID:   1,
			ClientID:   1,
		}
		
		err := validator.ValidateCase(case)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status")
	})
	
	t.Run("InvalidPriority", func(t *testing.T) {
		// 测试无效优先级
		case := &models.Case{
			Title:      "Test Case",
			CaseNumber: "TC-001",
			CaseType:   "Civil",
			Status:     "active",
			Priority:   "InvalidPriority", // 无效优先级
			LawyerID:   1,
			ClientID:   1,
		}
		
		err := validator.ValidateCase(case)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "priority")
	})
	
	t.Run("MissingLawyerID", func(t *testing.T) {
		// 测试缺失律师ID
		case := &models.Case{
			Title:      "Test Case",
			CaseNumber: "TC-001",
			CaseType:   "Civil",
			Status:     "active",
			Priority:   "medium",
			LawyerID:   0, // 无效ID
			ClientID:   1,
		}
		
		err := validator.ValidateCase(case)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lawyer_id")
	})
	
	t.Run("MissingClientID", func(t *testing.T) {
		// 测试缺失客户ID
		case := &models.Case{
			Title:      "Test Case",
			CaseNumber: "TC-001",
			CaseType:   "Civil",
			Status:     "active",
			Priority:   "medium",
			LawyerID:   1,
			ClientID:   0, // 无效ID
		}
		
		err := validator.ValidateCase(case)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client_id")
	})
}

// CaseRepositoryTest 案件仓库测试
func TestCaseRepository(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	repo := NewCaseRepository(suite.DB)
	
	t.Run("CreateCase", func(t *testing.T) {
		// 测试创建案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		case := &models.Case{
			Title:      "Test Case",
			CaseNumber: "TC-001",
			CaseType:   "Civil",
			Status:     "active",
			Priority:   "medium",
			LawyerID:   lawyer.ID,
			ClientID:   client.ID,
			StartDate:  time.Now(),
		}
		
		err := repo.Create(case)
		require.NoError(t, err)
		assert.NotZero(t, case.ID)
	})
	
	t.Run("FindByID", func(t *testing.T) {
		// 测试根据ID查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		foundCase, err := repo.FindByID(case.ID)
		require.NoError(t, err)
		assert.Equal(t, case.ID, foundCase.ID)
		assert.Equal(t, case.Title, foundCase.Title)
	})
	
	t.Run("FindByCaseNumber", func(t *testing.T) {
		// 测试根据案件编号查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		foundCase, err := repo.FindByCaseNumber(case.CaseNumber)
		require.NoError(t, err)
		assert.Equal(t, case.ID, foundCase.ID)
		assert.Equal(t, case.CaseNumber, foundCase.CaseNumber)
	})
	
	t.Run("Update", func(t *testing.T) {
		// 测试更新
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		case.Title = "Updated Title"
		case.Status = "completed"
		
		err := repo.Update(case)
		require.NoError(t, err)
		
		// 验证更新
		updatedCase, err := repo.FindByID(case.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", updatedCase.Title)
		assert.Equal(t, "completed", updatedCase.Status)
	})
	
	t.Run("Delete", func(t *testing.T) {
		// 测试删除
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		err := repo.Delete(case.ID)
		require.NoError(t, err)
		
		// 验证删除
		_, err = repo.FindByID(case.ID)
		assert.Error(t, err)
	})
	
	t.Run("List", func(t *testing.T) {
		// 测试列表
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 5; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		cases, total, err := repo.List(1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 5)
		assert.GreaterOrEqual(t, total, 5)
	})
	
	t.Run("Search", func(t *testing.T) {
		// 测试搜索
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		cases, err := repo.Search(case.Title)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 1)
	})
	
	t.Run("FindByLawyer", func(t *testing.T) {
		// 测试根据律师查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		cases, err := repo.FindByLawyer(lawyer.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 3)
	})
	
	t.Run("FindByClient", func(t *testing.T) {
		// 测试根据客户查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		cases, err := repo.FindByClient(client.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cases), 3)
	})
	
	t.Run("GetCaseStats", func(t *testing.T) {
		// 测试获取案件统计
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 5; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		stats, err := repo.GetCaseStats()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalCases, 5)
	})
}