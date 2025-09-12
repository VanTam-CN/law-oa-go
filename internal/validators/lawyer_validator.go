package validators

import (
	"errors"
	"law-oa-go/internal/models"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 律师验证器
type LawyerValidator struct{}

// 验证律师数据
func (v *LawyerValidator) ValidateLawyer(lawyer *models.User, isUpdate bool) error {
	// 验证律师姓名
	if lawyer.Name == "" {
		return errors.New("律师姓名不能为空")
	}
	if len(lawyer.Name) < 2 {
		return errors.New("律师姓名至少需要2个字符")
	}
	if len(lawyer.Name) > 100 {
		return errors.New("律师姓名不能超过100个字符")
	}
	
	// 验证角色
	if lawyer.Role != "lawyer" {
		return errors.New("用户角色必须为律师")
	}
	
	// 验证邮箱
	if lawyer.Email == "" {
		return errors.New("律师邮箱不能为空")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(lawyer.Email) {
		return errors.New("律师邮箱格式不正确")
	}
	
	// 验证手机号（如果提供）
	if lawyer.Phone != "" {
		phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
		if !phoneRegex.MatchString(lawyer.Phone) {
			return errors.New("律师手机号格式不正确")
		}
	}
	
	// 验证状态
	if lawyer.Status == "" {
		return errors.New("律师状态不能为空")
	}
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
	}
	if !validStatuses[lawyer.Status] {
		return errors.New("无效的律师状态")
	}
	
	return nil
}

// 验证律师搜索参数
func (v *LawyerValidator) ValidateSearchParams(searchTerm, status string) error {
	// 验证搜索词长度
	if len(searchTerm) > 100 {
		return errors.New("搜索词不能超过100个字符")
	}
	
	// 验证状态过滤
	if status != "" {
		validStatuses := map[string]bool{
			"active":   true,
			"inactive": true,
		}
		if !validStatuses[status] {
			return errors.New("无效的状态过滤")
		}
	}
	
	return nil
}

// 验证分页参数
func (v *LawyerValidator) ValidatePagination(page, pageSize int) error {
	if page < 1 {
		return errors.New("页码必须大于0")
	}
	if pageSize < 1 || pageSize > 100 {
		return errors.New("每页数量必须在1-100之间")
	}
	return nil
}