package validators

import (
	"errors"
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
		"civil":         true,
		"criminal":      true,
		"commercial":    true,
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
			"civil":         true,
			"criminal":      true,
			"commercial":    true,
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