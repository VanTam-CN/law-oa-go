package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"
)

// 案件验证器
type CaseValidator struct{}

// 验证案件数据
func (v *CaseValidator) ValidateCase(caseItem *models.Case, isUpdate bool) error {
	// 验证案件标题
	if caseItem.Title == "" {
		return errors.New("案件标题不能为空")
	}
	if len(caseItem.Title) < 2 {
		return errors.New("案件标题至少需要2个字符")
	}
	if len(caseItem.Title) > 200 {
		return errors.New("案件标题不能超过200个字符")
	}

	// 验证案件类型
	if caseItem.CaseType == "" {
		return errors.New("案件类型不能为空")
	}
	if len(caseItem.CaseType) > 50 {
		return errors.New("案件类型不能超过50个字符")
	}

	// 验证案件类型是否有效
	validCaseTypes := map[string]bool{
		"civil":          true,
		"criminal":       true,
		"commercial":     true,
		"administrative": true,
	}
	if !validCaseTypes[caseItem.CaseType] {
		return errors.New("无效的案件类型")
	}

	// 验证优先级
	if caseItem.Priority == "" {
		return errors.New("案件优先级不能为空")
	}
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"urgent": true,
	}
	if !validPriorities[caseItem.Priority] {
		return errors.New("无效的案件优先级")
	}

	// 验证状态
	if caseItem.Status == "" {
		return errors.New("案件状态不能为空")
	}
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"closed":    true,
		"suspended": true,
	}
	if !validStatuses[caseItem.Status] {
		return errors.New("无效的案件状态")
	}

	// 验证客户ID
	if caseItem.ClientID == 0 {
		return errors.New("客户ID不能为空")
	}

	// 验证律师ID
	if caseItem.LawyerID == 0 {
		return errors.New("律师ID不能为空")
	}

	// 验证描述长度
	if len(caseItem.Description) > 2000 {
		return errors.New("案件描述不能超过2000个字符")
	}

	// 验证日期逻辑
	if caseItem.StartDate != nil && caseItem.EndDate != nil {
		if caseItem.StartDate.After(*caseItem.EndDate) {
			return errors.New("开始日期不能晚于结束日期")
		}
	}

	return nil
}

// 验证案件搜索参数
func (v *CaseValidator) ValidateSearchParams(searchTerm, caseType, status, priority string) error {
	// 验证搜索词长度
	if len(searchTerm) > 100 {
		return errors.New("搜索词不能超过100个字符")
	}

	// 验证案件类型过滤
	if caseType != "" {
		validCaseTypes := map[string]bool{
			"civil":          true,
			"criminal":       true,
			"commercial":     true,
			"administrative": true,
		}
		if !validCaseTypes[caseType] {
			return errors.New("无效的案件类型过滤")
		}
	}

	// 验证状态过滤
	if status != "" {
		validStatuses := map[string]bool{
			"pending":   true,
			"active":    true,
			"closed":    true,
			"suspended": true,
		}
		if !validStatuses[status] {
			return errors.New("无效的状态过滤")
		}
	}

	// 验证优先级过滤
	if priority != "" {
		validPriorities := map[string]bool{
			"low":    true,
			"medium": true,
			"high":   true,
			"urgent": true,
		}
		if !validPriorities[priority] {
			return errors.New("无效的优先级过滤")
		}
	}

	return nil
}

// 验证分页参数
func (v *CaseValidator) ValidatePagination(page, pageSize int) error {
	if page < 1 {
		return errors.New("页码必须大于0")
	}
	if pageSize < 1 || pageSize > 100 {
		return errors.New("每页数量必须在1-100之间")
	}
	return nil
}

// 验证案件金额
func (v *CaseValidator) ValidateCaseAmount(amount float64) error {
	if amount < 0 {
		return errors.New("案件金额不能为负数")
	}
	if amount > 999999999 {
		return errors.New("案件金额过大，请确认是否正确")
	}
	return nil
}

// 验证日期格式
func (v *CaseValidator) ValidateDateFormat(dateStr string) error {
	if dateStr == "" {
		return nil // 可选字段
	}

	// 尝试多种日期格式
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return nil
		}
	}

	return errors.New("日期格式不正确，请使用 YYYY-MM-DD 格式")
}

// 验证客户名称
func (v *CaseValidator) ValidateClientName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("客户名称不能为空")
	}
	if len(name) < 2 {
		return errors.New("客户名称至少需要2个字符")
	}
	if len(name) > 100 {
		return errors.New("客户名称不能超过100个字符")
	}

	// 验证客户名称不包含特殊字符
	if matched, _ := regexp.MatchString(`[<>\"\'&]`, name); matched {
		return errors.New("客户名称不能包含特殊字符")
	}

	return nil
}

// 验证律师名称
func (v *CaseValidator) ValidateLawyerName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("律师名称不能为空")
	}
	if len(name) < 2 {
		return errors.New("律师名称至少需要2个字符")
	}
	if len(name) > 50 {
		return errors.New("律师名称不能超过50个字符")
	}

	// 验证律师名称不包含特殊字符
	if matched, _ := regexp.MatchString(`[<>\"\'&]`, name); matched {
		return errors.New("律师名称不能包含特殊字符")
	}

	return nil
}

// 验证联系方式（电话或邮箱）
func (v *CaseValidator) ValidateContactInfo(contact string) error {
	if contact == "" {
		return nil // 可选字段
	}

	// 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(contact) {
		return nil
	}

	// 验证手机号格式（中国手机号）
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	cleanPhone := regexp.MustCompile(`[\s-]`).ReplaceAllString(contact, "")
	if phoneRegex.MatchString(cleanPhone) {
		return nil
	}

	// 验证固话格式
	landlineRegex := regexp.MustCompile(`^0\d{2,3}-?\d{7,8}$`)
	if landlineRegex.MatchString(cleanPhone) {
		return nil
	}

	return errors.New("联系方式格式不正确，请输入有效的邮箱或电话号码")
}

// 验证案件描述
func (v *CaseValidator) ValidateDescription(description string) error {
	if len(description) > 5000 {
		return errors.New("案件描述不能超过5000个字符")
	}

	// 检查是否包含潜在的脚本注入
	if strings.Contains(strings.ToLower(description), "<script>") {
		return errors.New("案件描述不能包含脚本代码")
	}

	return nil
}

// 验证ID格式
func (v *CaseValidator) ValidateID(id interface{}, fieldName string) error {
	switch v := id.(type) {
	case int:
		if v <= 0 {
			return fmt.Errorf("%s必须是正整数", fieldName)
		}
	case uint:
		if v == 0 {
			return fmt.Errorf("%s不能为0", fieldName)
		}
	case string:
		num, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("%s必须是数字", fieldName)
		}
		if num <= 0 {
			return fmt.Errorf("%s必须是正整数", fieldName)
		}
	default:
		return fmt.Errorf("%s格式不正确", fieldName)
	}
	return nil
}

// 验证排序参数
func (v *CaseValidator) ValidateSortParams(sortBy, sortOrder string) error {
	if sortBy == "" {
		return nil // 可选参数
	}

	validSortFields := map[string]bool{
		"id":         true,
		"title":      true,
		"case_type":  true,
		"priority":   true,
		"status":     true,
		"created_at": true,
		"updated_at": true,
		"start_date": true,
		"end_date":   true,
	}

	if !validSortFields[sortBy] {
		return fmt.Errorf("无效的排序字段: %s", sortBy)
	}

	if sortOrder != "" {
		validSortOrders := map[string]bool{
			"asc":  true,
			"desc": true,
		}
		if !validSortOrders[sortOrder] {
			return fmt.Errorf("无效的排序方向: %s", sortOrder)
		}
	}

	return nil
}

// 验证过滤参数
func (v *CaseValidator) ValidateFilterParams(filters map[string]interface{}) error {
	for key, value := range filters {
		switch key {
		case "client_id", "lawyer_id":
			if err := v.ValidateID(value, key); err != nil {
				return err
			}
		case "case_type":
			if str, ok := value.(string); ok {
				validCaseTypes := map[string]bool{
					"civil":          true,
					"criminal":       true,
					"commercial":     true,
					"administrative": true,
				}
				if !validCaseTypes[str] {
					return fmt.Errorf("无效的案件类型: %s", str)
				}
			} else {
				return fmt.Errorf("案件类型必须是字符串")
			}
		case "status":
			if str, ok := value.(string); ok {
				validStatuses := map[string]bool{
					"pending":   true,
					"active":    true,
					"closed":    true,
					"suspended": true,
				}
				if !validStatuses[str] {
					return fmt.Errorf("无效的状态: %s", str)
				}
			} else {
				return fmt.Errorf("状态必须是字符串")
			}
		case "priority":
			if str, ok := value.(string); ok {
				validPriorities := map[string]bool{
					"low":    true,
					"medium": true,
					"high":   true,
					"urgent": true,
				}
				if !validPriorities[str] {
					return fmt.Errorf("无效的优先级: %s", str)
				}
			} else {
				return fmt.Errorf("优先级必须是字符串")
			}
		default:
			return fmt.Errorf("未知的过滤参数: %s", key)
		}
	}
	return nil
}

// 综合验证案件搜索请求
func (v *CaseValidator) ValidateSearchRequest(searchTerm string, filters map[string]interface{}, page, pageSize int, sortBy, sortOrder string) error {
	// 验证搜索词
	if len(searchTerm) > 100 {
		return errors.New("搜索词不能超过100个字符")
	}

	// 验证过滤参数
	if err := v.ValidateFilterParams(filters); err != nil {
		return err
	}

	// 验证分页参数
	if err := v.ValidatePagination(page, pageSize); err != nil {
		return err
	}

	// 验证排序参数
	if err := v.ValidateSortParams(sortBy, sortOrder); err != nil {
		return err
	}

	return nil
}
