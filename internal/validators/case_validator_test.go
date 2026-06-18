package validators

import (
	"testing"
	"time"

	"law-oa-go/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCaseValidator_ValidateCase(t *testing.T) {
	validator := &CaseValidator{}

	t.Run("验证有效案件", func(t *testing.T) {
		now := time.Now()
		later := now.Add(24 * time.Hour)

		caseItem := &models.Case{
			Title:       "测试案件",
			CaseType:    "civil",
			Priority:    "medium",
			Status:      "active",
			ClientID:    1,
			LawyerID:    1,
			Description: "这是一个测试案件",
			StartDate:   &now,
			EndDate:     &later,
		}

		err := validator.ValidateCase(caseItem, false)
		assert.NoError(t, err)
	})

	t.Run("验证空标题", func(t *testing.T) {
		caseItem := &models.Case{
			Title:    "",
			CaseType: "civil",
			Priority: "medium",
			Status:   "active",
			ClientID: 1,
			LawyerID: 1,
		}

		err := validator.ValidateCase(caseItem, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "案件标题不能为空")
	})

	t.Run("验证标题过短", func(t *testing.T) {
		caseItem := &models.Case{
			Title:    "A",
			CaseType: "civil",
			Priority: "medium",
			Status:   "active",
			ClientID: 1,
			LawyerID: 1,
		}

		err := validator.ValidateCase(caseItem, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "案件标题至少需要2个字符")
	})

	t.Run("验证无效案件类型", func(t *testing.T) {
		caseItem := &models.Case{
			Title:    "测试案件",
			CaseType: "invalid_type",
			Priority: "medium",
			Status:   "active",
			ClientID: 1,
			LawyerID: 1,
		}

		err := validator.ValidateCase(caseItem, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的案件类型")
	})

	t.Run("验证无效优先级", func(t *testing.T) {
		caseItem := &models.Case{
			Title:    "测试案件",
			CaseType: "civil",
			Priority: "invalid_priority",
			Status:   "active",
			ClientID: 1,
			LawyerID: 1,
		}

		err := validator.ValidateCase(caseItem, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的案件优先级")
	})

	t.Run("验证无效状态", func(t *testing.T) {
		caseItem := &models.Case{
			Title:    "测试案件",
			CaseType: "civil",
			Priority: "medium",
			Status:   "invalid_status",
			ClientID: 1,
			LawyerID: 1,
		}

		err := validator.ValidateCase(caseItem, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的案件状态")
	})

	t.Run("验证日期逻辑", func(t *testing.T) {
		now := time.Now()
		before := now.Add(-24 * time.Hour)

		caseItem := &models.Case{
			Title:       "测试案件",
			CaseType:    "civil",
			Priority:    "medium",
			Status:      "active",
			ClientID:    1,
			LawyerID:    1,
			StartDate:   &now,
			EndDate:     &before,
			Description: "测试",
		}

		err := validator.ValidateCase(caseItem, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "开始日期不能晚于结束日期")
	})
}

func Fuzz_CaseValidator_Validate(f *testing.F) {
	validator := &CaseValidator{}

	// 添加种子语料
	f.Add("测试案件", "civil", "medium", "active")
	f.Add("A", "civil", "medium", "active")
	f.Add("", "civil", "medium", "active")

	f.Fuzz(func(t *testing.T, title, caseType, priority, status string) {
		// 限制输入长度以避免内存问题
		if len(title) > 200 || len(caseType) > 50 || len(priority) > 20 || len(status) > 20 {
			t.Skip()
		}

		caseItem := &models.Case{
			Title:    title,
			CaseType: caseType,
			Priority: priority,
			Status:   status,
			ClientID: 1,
			LawyerID: 1,
		}

		// 测试不会panic
		_ = validator.ValidateCase(caseItem, false)
	})
}

func TestCaseValidator_ValidatePagination(t *testing.T) {
	validator := &CaseValidator{}

	t.Run("验证有效分页参数", func(t *testing.T) {
		err := validator.ValidatePagination(1, 20)
		assert.NoError(t, err)
	})

	t.Run("验证无效页码", func(t *testing.T) {
		err := validator.ValidatePagination(0, 20)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "页码必须大于0")
	})

	t.Run("验证无效页大小", func(t *testing.T) {
		err := validator.ValidatePagination(1, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "每页数量必须在1-100之间")
	})

	t.Run("验证页大小过大", func(t *testing.T) {
		err := validator.ValidatePagination(1, 101)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "每页数量必须在1-100之间")
	})
}

func TestCaseValidator_ValidateContactInfo(t *testing.T) {
	validator := &CaseValidator{}

	t.Run("验证有效邮箱", func(t *testing.T) {
		err := validator.ValidateContactInfo("test@example.com")
		assert.NoError(t, err)
	})

	t.Run("验证有效手机号", func(t *testing.T) {
		err := validator.ValidateContactInfo("13800138000")
		assert.NoError(t, err)
	})

	t.Run("验证有效固话", func(t *testing.T) {
		err := validator.ValidateContactInfo("010-12345678")
		assert.NoError(t, err)
	})

	t.Run("验证无效联系方式", func(t *testing.T) {
		err := validator.ValidateContactInfo("invalid_contact")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "联系方式格式不正确")
	})

	t.Run("验证空联系方式", func(t *testing.T) {
		err := validator.ValidateContactInfo("")
		assert.NoError(t, err)
	})
}

func TestCaseValidator_ValidateDescription(t *testing.T) {
	validator := &CaseValidator{}

	t.Run("验证有效描述", func(t *testing.T) {
		err := validator.ValidateDescription("这是一个有效的案件描述")
		assert.NoError(t, err)
	})

	t.Run("验证描述过长", func(t *testing.T) {
		longDesc := string(make([]byte, 5001))
		for i := range longDesc {
			longDesc = "a" + longDesc[:i]
		}
		err := validator.ValidateDescription(longDesc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "案件描述不能超过5000个字符")
	})

	t.Run("验证包含脚本标签", func(t *testing.T) {
		err := validator.ValidateDescription("正常文本<script>alert('xss')</script>")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "案件描述不能包含脚本代码")
	})
}

func TestClientValidator_Validate(t *testing.T) {
	validator := &ClientValidator{}

	t.Run("验证有效客户", func(t *testing.T) {
		client := &models.Client{
			Name:   "测试客户",
			Email:  "test@example.com",
			Phone:  "13800138000",
			Status: "active",
		}

		err := validator.ValidateClient(client, false)
		assert.NoError(t, err)
	})

	t.Run("验证空客户名", func(t *testing.T) {
		client := &models.Client{
			Name:   "",
			Email:  "test@example.com",
			Phone:  "13800138000",
			Status: "active",
		}

		err := validator.ValidateClient(client, false)
		assert.Error(t, err)
	})

	t.Run("验证客户名过短", func(t *testing.T) {
		client := &models.Client{
			Name:   "A",
			Email:  "test@example.com",
			Phone:  "13800138000",
			Status: "active",
		}

		err := validator.ValidateClient(client, false)
		assert.Error(t, err)
	})
}

func Fuzz_ClientValidator_Validate(f *testing.F) {
	validator := &ClientValidator{}

	// 添加种子语料
	f.Add("测试客户", "test@example.com", "13800138000")
	f.Add("", "test@example.com", "13800138000")
	f.Add("A", "test@example.com", "13800138000")

	f.Fuzz(func(t *testing.T, name, email, phone string) {
		// 限制输入长度
		if len(name) > 100 || len(email) > 100 || len(phone) > 20 {
			t.Skip()
		}

		client := &models.Client{
			Name:   name,
			Email:  email,
			Phone:  phone,
			Status: "active",
		}

		// 测试不会panic
		_ = validator.ValidateClient(client, false)
	})
}

func TestLawyerValidator_Validate(t *testing.T) {
	validator := &LawyerValidator{}

	t.Run("验证有效律师", func(t *testing.T) {
		lawyer := &models.User{
			Name:   "张律师",
			Email:  "lawyer@example.com",
			Phone:  "13800138000",
			Role:   "lawyer",
			Status: "active",
		}

		err := validator.ValidateLawyer(lawyer, false)
		assert.NoError(t, err)
	})

	t.Run("验证空律师名", func(t *testing.T) {
		lawyer := &models.User{
			Name:   "",
			Email:  "lawyer@example.com",
			Phone:  "13800138000",
			Role:   "lawyer",
			Status: "active",
		}

		err := validator.ValidateLawyer(lawyer, false)
		assert.Error(t, err)
	})
}

func Fuzz_LawyerValidator_Validate(f *testing.F) {
	validator := &LawyerValidator{}

	// 添加种子语料
	f.Add("张律师", "lawyer@example.com", "13800138000", "LICENSE123")
	f.Add("", "lawyer@example.com", "13800138000", "LICENSE123")
	f.Add("A", "lawyer@example.com", "13800138000", "LICENSE123")

	f.Fuzz(func(t *testing.T, name, email, phone, role string) {
		// 限制输入长度
		if len(name) > 100 || len(email) > 100 || len(phone) > 20 || len(role) > 50 {
			t.Skip()
		}

		lawyer := &models.User{
			Name:   name,
			Email:  email,
			Phone:  phone,
			Role:   role,
			Status: "active",
		}

		// 测试不会panic
		_ = validator.ValidateLawyer(lawyer, false)
	})
}
