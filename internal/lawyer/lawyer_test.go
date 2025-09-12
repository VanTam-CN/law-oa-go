package lawyer

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

// LawyerServiceTest 律师服务测试
func TestLawyerService(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("CreateLawyer", func(t *testing.T) {
		// 测试创建律师
		lawyer := &models.Lawyer{
			FirstName:    "Michael",
			LastName:     "Johnson",
			Email:        "michael.johnson@example.com",
			Phone:        "13800138001",
			Specialty:    "Criminal Law",
			Experience:   10,
			Status:       "active",
			LicenseNumber: "LIC123456",
			JoinDate:     time.Now(),
		}
		
		createdLawyer, err := lawyerService.CreateLawyer(lawyer)
		require.NoError(t, err)
		assert.NotZero(t, createdLawyer.ID)
		assert.Equal(t, lawyer.FirstName, createdLawyer.FirstName)
		assert.Equal(t, lawyer.LastName, createdLawyer.LastName)
		assert.Equal(t, lawyer.Specialty, createdLawyer.Specialty)
	})
	
	t.Run("GetLawyer", func(t *testing.T) {
		// 测试获取律师
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		retrievedLawyer, err := lawyerService.GetLawyer(lawyer.ID)
		require.NoError(t, err)
		assert.Equal(t, lawyer.ID, retrievedLawyer.ID)
		assert.Equal(t, lawyer.FirstName, retrievedLawyer.FirstName)
		assert.Equal(t, lawyer.LastName, retrievedLawyer.LastName)
		assert.Equal(t, lawyer.Specialty, retrievedLawyer.Specialty)
	})
	
	t.Run("GetLawyerNotFound", func(t *testing.T) {
		// 测试获取不存在的律师
		_, err := lawyerService.GetLawyer(99999)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
	
	t.Run("UpdateLawyer", func(t *testing.T) {
		// 测试更新律师
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		updatedLawyer := *lawyer
		updatedLawyer.FirstName = "Updated"
		updatedLawyer.Specialty = "Updated Specialty"
		updatedLawyer.Experience = 15
		
		result, err := lawyerService.UpdateLawyer(lawyer.ID, &updatedLawyer)
		require.NoError(t, err)
		assert.Equal(t, "Updated", result.FirstName)
		assert.Equal(t, "Updated Specialty", result.Specialty)
		assert.Equal(t, 15, result.Experience)
	})
	
	t.Run("DeleteLawyer", func(t *testing.T) {
		// 测试删除律师
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		err := lawyerService.DeleteLawyer(lawyer.ID)
		require.NoError(t, err)
		
		// 验证律师已被删除
		_, err = lawyerService.GetLawyer(lawyer.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
	
	t.Run("ListLawyers", func(t *testing.T) {
		// 测试律师列表
		// 创建多个律师
		for i := 0; i < 5; i++ {
			config.CreateTestLawyer(t, suite.DB)
		}
		
		lawyers, total, err := lawyerService.ListLawyers(1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(lawyers), 5)
		assert.GreaterOrEqual(t, total, 5)
	})
	
	t.Run("SearchLawyers", func(t *testing.T) {
		// 测试搜索律师
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		searchTerm := lawyer.FirstName
		lawyers, err := lawyerService.SearchLawyers(searchTerm)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(lawyers), 1)
	})
	
	t.Run("GetLawyerBySpecialty", func(t *testing.T) {
		// 测试根据专业获取律师
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		lawyers, err := lawyerService.GetLawyersBySpecialty(lawyer.Specialty)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(lawyers), 1)
	})
	
	t.Run("GetLawyerCases", func(t *testing.T) {
		// 测试获取律师案件
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		cases, err := lawyerService.GetLawyerCases(lawyer.ID)
		require.NoError(t, err)
		assert.Len(t, cases, 1)
		assert.Equal(t, case.ID, cases[0].ID)
	})
	
	t.Run("GetLawyerStats", func(t *testing.T) {
		// 测试获取律师统计
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		stats, err := lawyerService.GetLawyerStats(lawyer.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalCases, 3)
		assert.GreaterOrEqual(t, stats.ActiveCases, 3)
	})
}

// LawyerHandlerTest 律师处理器测试
func TestLawyerHandler(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	handler := NewLawyerHandler(suite.DB)
	
	// 设置路由
	router := gin.New()
	router.POST("/lawyers", handler.CreateLawyer)
	router.GET("/lawyers/:id", handler.GetLawyer)
	router.PUT("/lawyers/:id", handler.UpdateLawyer)
	router.DELETE("/lawyers/:id", handler.DeleteLawyer)
	router.GET("/lawyers", handler.ListLawyers)
	router.GET("/lawyers/search", handler.SearchLawyers)
	router.GET("/lawyers/specialty/:specialty", handler.GetLawyersBySpecialty)
	router.GET("/lawyers/:id/cases", handler.GetLawyerCases)
	router.GET("/lawyers/:id/stats", handler.GetLawyerStats)
	
	suite.Router = router
	
	t.Run("CreateLawyerAPI", func(t *testing.T) {
		// 测试创建律师API
		lawyer := map[string]interface{}{
			"first_name":    "Sarah",
			"last_name":     "Williams",
			"email":         "sarah.williams@example.com",
			"phone":         "13700137001",
			"specialty":     "Family Law",
			"experience":    8,
			"status":        "active",
			"license_number": "LIC789012",
		}
		
		w := suite.PerformAuthRequest("POST", "/lawyers", lawyer)
		data := suite.AssertSuccess(t, w)
		
		assert.NotZero(t, data["id"])
		assert.Equal(t, "Sarah", data["first_name"])
		assert.Equal(t, "Williams", data["last_name"])
		assert.Equal(t, "Family Law", data["specialty"])
	})
	
	t.Run("GetLawyerAPI", func(t *testing.T) {
		// 测试获取律师API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		w := suite.PerformAuthRequest("GET", "/lawyers/"+string(lawyer.ID), nil)
		data := suite.AssertSuccess(t, w)
		
		assert.Equal(t, lawyer.ID, uint(data["id"].(float64)))
		assert.Equal(t, lawyer.FirstName, data["first_name"])
		assert.Equal(t, lawyer.LastName, data["last_name"])
		assert.Equal(t, lawyer.Specialty, data["specialty"])
	})
	
	t.Run("UpdateLawyerAPI", func(t *testing.T) {
		// 测试更新律师API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		updateData := map[string]interface{}{
			"first_name": "Updated",
			"specialty":  "Updated Specialty",
			"experience": 12,
		}
		
		w := suite.PerformAuthRequest("PUT", "/lawyers/"+string(lawyer.ID), updateData)
		data := suite.AssertSuccess(t, w)
		
		assert.Equal(t, "Updated", data["first_name"])
		assert.Equal(t, "Updated Specialty", data["specialty"])
		assert.Equal(t, 12, int(data["experience"].(float64)))
	})
	
	t.Run("DeleteLawyerAPI", func(t *testing.T) {
		// 测试删除律师API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		w := suite.PerformAuthRequest("DELETE", "/lawyers/"+string(lawyer.ID), nil)
		suite.AssertSuccess(t, w)
		
		// 验证律师已被删除
		w = suite.PerformAuthRequest("GET", "/lawyers/"+string(lawyer.ID), nil)
		suite.AssertError(t, w, 404)
	})
	
	t.Run("ListLawyersAPI", func(t *testing.T) {
		// 测试律师列表API
		// 创建多个律师
		for i := 0; i < 3; i++ {
			config.CreateTestLawyer(t, suite.DB)
		}
		
		w := suite.PerformAuthRequest("GET", "/lawyers?page=1&size=10", nil)
		data := suite.AssertSuccess(t, w)
		
		lawyers := data["lawyers"].([]interface{})
		total := data["total"].(float64)
		
		assert.GreaterOrEqual(t, len(lawyers), 3)
		assert.GreaterOrEqual(t, total, 3)
	})
	
	t.Run("SearchLawyersAPI", func(t *testing.T) {
		// 测试搜索律师API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		w := suite.PerformAuthRequest("GET", "/lawyers/search?q="+lawyer.FirstName, nil)
		data := suite.AssertSuccess(t, w)
		
		lawyers := data["lawyers"].([]interface{})
		assert.GreaterOrEqual(t, len(lawyers), 1)
	})
	
	t.Run("GetLawyersBySpecialtyAPI", func(t *testing.T) {
		// 测试根据专业获取律师API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		w := suite.PerformAuthRequest("GET", "/lawyers/specialty/"+lawyer.Specialty, nil)
		data := suite.AssertSuccess(t, w)
		
		lawyers := data["lawyers"].([]interface{})
		assert.GreaterOrEqual(t, len(lawyers), 1)
	})
	
	t.Run("GetLawyerCasesAPI", func(t *testing.T) {
		// 测试获取律师案件API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		w := suite.PerformAuthRequest("GET", "/lawyers/"+string(lawyer.ID)+"/cases", nil)
		data := suite.AssertSuccess(t, w)
		
		cases := data["cases"].([]interface{})
		assert.Len(t, cases, 1)
		assert.Equal(t, case.ID, uint(cases[0].(map[string]interface{})["id"].(float64)))
	})
	
	t.Run("GetLawyerStatsAPI", func(t *testing.T) {
		// 测试获取律师统计API
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 2; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		w := suite.PerformAuthRequest("GET", "/lawyers/"+string(lawyer.ID)+"/stats", nil)
		data := suite.AssertSuccess(t, w)
		
		assert.GreaterOrEqual(t, data["total_cases"].(float64), 2)
		assert.GreaterOrEqual(t, data["active_cases"].(float64), 2)
	})
	
	t.Run("CreateLawyerValidation", func(t *testing.T) {
		// 测试创建律师验证
		invalidLawyer := map[string]interface{}{
			"first_name": "", // 空名字
			"email":      "invalid-email", // 无效邮箱
			"experience": -1, // 无效经验
		}
		
		w := suite.PerformAuthRequest("POST", "/lawyers", invalidLawyer)
		suite.AssertError(t, w, 400)
	})
	
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		// 测试未授权访问
		w := suite.PerformRequest("GET", "/lawyers", nil, nil)
		suite.AssertError(t, w, 401)
	})
}

// LawyerValidatorTest 律师验证器测试
func TestLawyerValidator(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	validator := NewLawyerValidator()
	
	t.Run("ValidLawyer", func(t *testing.T) {
		// 测试有效律师验证
		lawyer := &models.Lawyer{
			FirstName:    "John",
			LastName:     "Doe",
			Email:        "john.doe@example.com",
			Phone:        "13800138000",
			Specialty:    "Corporate Law",
			Experience:   5,
			LicenseNumber: "LIC123456",
		}
		
		err := validator.ValidateLawyer(lawyer)
		assert.NoError(t, err)
	})
	
	t.Run("InvalidEmail", func(t *testing.T) {
		// 测试无效邮箱
		lawyer := &models.Lawyer{
			FirstName:    "John",
			LastName:     "Doe",
			Email:        "invalid-email",
			Specialty:    "Corporate Law",
			Experience:   5,
			LicenseNumber: "LIC123456",
		}
		
		err := validator.ValidateLawyer(lawyer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})
	
	t.Run("MissingRequiredFields", func(t *testing.T) {
		// 测试缺失必填字段
		lawyer := &models.Lawyer{
			FirstName: "",
			LastName:  "",
			Email:     "",
		}
		
		err := validator.ValidateLawyer(lawyer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first_name")
		assert.Contains(t, err.Error(), "last_name")
		assert.Contains(t, err.Error(), "email")
	})
	
	t.Run("InvalidExperience", func(t *testing.T) {
		// 测试无效经验
		lawyer := &models.Lawyer{
			FirstName:    "John",
			LastName:     "Doe",
			Email:        "john.doe@example.com",
			Specialty:    "Corporate Law",
			Experience:   -1, // 负数
			LicenseNumber: "LIC123456",
		}
		
		err := validator.ValidateLawyer(lawyer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "experience")
	})
	
	t.Run("InvalidLicenseNumber", func(t *testing.T) {
		// 测试无效执照号码
		lawyer := &models.Lawyer{
			FirstName:    "John",
			LastName:     "Doe",
			Email:        "john.doe@example.com",
			Specialty:    "Corporate Law",
			Experience:   5,
			LicenseNumber: "", // 空执照号码
		}
		
		err := validator.ValidateLawyer(lawyer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "license_number")
	})
}

// LawyerRepositoryTest 律师仓库测试
func TestLawyerRepository(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	repo := NewLawyerRepository(suite.DB)
	
	t.Run("CreateLawyer", func(t *testing.T) {
		// 测试创建律师
		lawyer := &models.Lawyer{
			FirstName:    "Test",
			LastName:     "Lawyer",
			Email:        "test@example.com",
			Phone:        "13800138000",
			Specialty:    "Test Specialty",
			Experience:   5,
			LicenseNumber: "TEST123",
		}
		
		err := repo.Create(lawyer)
		require.NoError(t, err)
		assert.NotZero(t, lawyer.ID)
	})
	
	t.Run("FindByID", func(t *testing.T) {
		// 测试根据ID查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		foundLawyer, err := repo.FindByID(lawyer.ID)
		require.NoError(t, err)
		assert.Equal(t, lawyer.ID, foundLawyer.ID)
		assert.Equal(t, lawyer.FirstName, foundLawyer.FirstName)
	})
	
	t.Run("FindByEmail", func(t *testing.T) {
		// 测试根据邮箱查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		foundLawyer, err := repo.FindByEmail(lawyer.Email)
		require.NoError(t, err)
		assert.Equal(t, lawyer.ID, foundLawyer.ID)
		assert.Equal(t, lawyer.Email, foundLawyer.Email)
	})
	
	t.Run("FindByLicenseNumber", func(t *testing.T) {
		// 测试根据执照号码查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		foundLawyer, err := repo.FindByLicenseNumber(lawyer.LicenseNumber)
		require.NoError(t, err)
		assert.Equal(t, lawyer.ID, foundLawyer.ID)
		assert.Equal(t, lawyer.LicenseNumber, foundLawyer.LicenseNumber)
	})
	
	t.Run("Update", func(t *testing.T) {
		// 测试更新
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		lawyer.FirstName = "Updated"
		lawyer.Specialty = "Updated Specialty"
		lawyer.Experience = 10
		
		err := repo.Update(lawyer)
		require.NoError(t, err)
		
		// 验证更新
		updatedLawyer, err := repo.FindByID(lawyer.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updatedLawyer.FirstName)
		assert.Equal(t, "Updated Specialty", updatedLawyer.Specialty)
		assert.Equal(t, 10, updatedLawyer.Experience)
	})
	
	t.Run("Delete", func(t *testing.T) {
		// 测试删除
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		err := repo.Delete(lawyer.ID)
		require.NoError(t, err)
		
		// 验证删除
		_, err = repo.FindByID(lawyer.ID)
		assert.Error(t, err)
	})
	
	t.Run("List", func(t *testing.T) {
		// 测试列表
		// 创建多个律师
		for i := 0; i < 5; i++ {
			config.CreateTestLawyer(t, suite.DB)
		}
		
		lawyers, total, err := repo.List(1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(lawyers), 5)
		assert.GreaterOrEqual(t, total, 5)
	})
	
	t.Run("Search", func(t *testing.T) {
		// 测试搜索
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		lawyers, err := repo.Search(lawyer.FirstName)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(lawyers), 1)
	})
	
	t.Run("FindBySpecialty", func(t *testing.T) {
		// 测试根据专业查找
		lawyer := config.CreateTestLawyer(t, suite.DB)
		
		lawyers, err := repo.FindBySpecialty(lawyer.Specialty)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(lawyers), 1)
	})
	
	t.Run("GetLawyerStats", func(t *testing.T) {
		// 测试获取律师统计
		lawyer := config.CreateTestLawyer(t, suite.DB)
		client := config.CreateTestClient(t, suite.DB)
		
		// 创建多个案件
		for i := 0; i < 3; i++ {
			config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		}
		
		stats, err := repo.GetLawyerStats(lawyer.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalCases, 3)
	})
}