package validators

import (
	"errors"
	"law-oa-go/internal/models"
	"regexp"
)

type ClientValidator struct{}

func NewClientValidator() *ClientValidator {
	return &ClientValidator{}
}

// 验证客户数据
func (v *ClientValidator) ValidateClient(client *models.Client, isUpdate bool) error {
	// 验证客户名称
	if client.Name == "" {
		return errors.New("客户名称不能为空")
	}
	if len(client.Name) < 2 {
		return errors.New("客户名称至少需要2个字符")
	}
	if len(client.Name) > 100 {
		return errors.New("客户名称不能超过100个字符")
	}

	// 验证邮箱（如果提供）
	if client.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(client.Email) {
			return errors.New("客户邮箱格式不正确")
		}
	}

	// 验证手机号（如果提供）
	if client.Phone != "" {
		phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
		if !phoneRegex.MatchString(client.Phone) {
			return errors.New("客户手机号格式不正确")
		}
	}

	// 验证地址长度
	if len(client.Address) > 255 {
		return errors.New("客户地址不能超过255个字符")
	}

	// 验证公司名称长度
	if len(client.Company) > 100 {
		return errors.New("公司名称不能超过100个字符")
	}

	// 验证备注长度
	if len(client.Notes) > 2000 {
		return errors.New("备注不能超过2000个字符")
	}

	// 验证状态
	if client.Status == "" {
		return errors.New("客户状态不能为空")
	}
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
	}
	if !validStatuses[client.Status] {
		return errors.New("无效的客户状态")
	}

	return nil
}

// 验证客户搜索参数
func (v *ClientValidator) ValidateSearchParams(searchTerm, status string) error {
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
func (v *ClientValidator) ValidatePagination(page, pageSize int) error {
	if page < 1 {
		return errors.New("页码必须大于0")
	}
	if pageSize < 1 || pageSize > 100 {
		return errors.New("每页数量必须在1-100之间")
	}
	return nil
}