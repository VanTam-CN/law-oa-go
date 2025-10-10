package validators

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// SimpleTestValidator 简化的测试验证器
type SimpleTestValidator struct{}

// NewSimpleTestValidator 创建简化的测试验证器
func NewSimpleTestValidator() *SimpleTestValidator {
	return &SimpleTestValidator{}
}

// ValidateTestSuite 验证测试套件
func (v *SimpleTestValidator) ValidateTestSuite(suite interface{}) error {
	// 基础验证逻辑
	if suite == nil {
		return fmt.Errorf("test suite cannot be nil")
	}

	// 这里可以添加更多的验证逻辑
	return nil
}

// ValidateTestCase 验证测试用例
func (v *SimpleTestValidator) ValidateTestCase(testCase interface{}) error {
	// 基础验证逻辑
	if testCase == nil {
		return fmt.Errorf("test case cannot be nil")
	}

	// 这里可以添加更多的验证逻辑
	return nil
}

// ValidateURL 验证URL格式
func (v *SimpleTestValidator) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// 如果是相对URL，先转换为绝对URL
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://localhost:8080" + urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL host cannot be empty")
	}

	return nil
}

// ValidateHTTPMethod 验证HTTP方法
func (v *SimpleTestValidator) ValidateHTTPMethod(method string) error {
	if method == "" {
		return fmt.Errorf("HTTP method cannot be empty")
	}

	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return nil
		}
	}

	return fmt.Errorf("invalid HTTP method: %s", method)
}

// ValidateHeaders 验证HTTP头部
func (v *SimpleTestValidator) ValidateHeaders(headers map[string]string) error {
	if headers == nil {
		return nil
	}

	for key, value := range headers {
		if key == "" {
			return fmt.Errorf("header key cannot be empty")
		}
		if strings.Contains(key, " ") {
			return fmt.Errorf("header key cannot contain spaces: %s", key)
		}
		if value == "" {
			return fmt.Errorf("header value cannot be empty for key: %s", key)
		}
	}

	return nil
}

// ValidateTestName 验证测试名称
func (v *SimpleTestValidator) ValidateTestName(name string) error {
	if name == "" {
		return fmt.Errorf("test name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("test name too long (max 100 characters)")
	}

	// 检查是否包含特殊字符
	if strings.Contains(name, "\n") || strings.Contains(name, "\r") {
		return fmt.Errorf("test name cannot contain line breaks")
	}

	return nil
}

// ValidateTestType 验证测试类型
func (v *SimpleTestValidator) ValidateTestType(testType string) error {
	validTypes := []string{"api", "ui", "performance", "integration", "e2e"}
	for _, validType := range validTypes {
		if testType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid test type: %s", testType)
}

// ValidateEnvironment 验证环境
func (v *SimpleTestValidator) ValidateEnvironment(env string) error {
	if env == "" {
		return nil // 环境可以为空
	}

	validEnvironments := []string{"dev", "test", "staging", "prod"}
	for _, validEnv := range validEnvironments {
		if env == validEnv {
			return nil
		}
	}

	return fmt.Errorf("invalid environment: %s", env)
}

// ValidateTimeout 验证超时时间
func (v *SimpleTestValidator) ValidateTimeout(timeout int) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if timeout > 3600 { // 1小时
		return fmt.Errorf("timeout too long (max 1 hour)")
	}

	return nil
}

// ValidatePriority 验证优先级
func (v *SimpleTestValidator) ValidatePriority(priority int) error {
	if priority < 1 || priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}

	return nil
}

// ValidateTags 验证标签
func (v *SimpleTestValidator) ValidateTags(tags []string) error {
	if tags == nil {
		return nil
	}

	for i, tag := range tags {
		if tag == "" {
			return fmt.Errorf("tag %d cannot be empty", i+1)
		}

		if len(tag) > 50 {
			return fmt.Errorf("tag %d too long (max 50 characters): %s", i+1, tag)
		}

		// 检查标签格式（只允许字母、数字、连字符、下划线）
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, tag)
		if !matched {
			return fmt.Errorf("invalid tag format (must contain only letters, numbers, hyphens and underscores): %s", tag)
		}
	}

	return nil
}

// ValidateStepType 验证步骤类型
func (v *SimpleTestValidator) ValidateStepType(stepType string) error {
	validTypes := []string{"navigate", "click", "fill", "wait", "assert", "screenshot", "javascript"}
	for _, validType := range validTypes {
		if stepType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid step type: %s", stepType)
}

// ValidateStepAction 验证步骤动作
func (v *SimpleTestValidator) ValidateStepAction(action string) error {
	if action == "" {
		return fmt.Errorf("step action cannot be empty")
	}

	if len(action) > 200 {
		return fmt.Errorf("step action too long (max 200 characters)")
	}

	return nil
}

// ValidateStepTarget 验证步骤目标
func (v *SimpleTestValidator) ValidateStepTarget(target string, stepType string) error {
	// 某些步骤类型需要目标
	if stepType == "navigate" || stepType == "click" || stepType == "fill" {
		if target == "" {
			return fmt.Errorf("%s step requires target", stepType)
		}

		// 对于导航步骤，验证URL
		if stepType == "navigate" {
			return v.ValidateURL(target)
		}
	}

	return nil
}

// ValidateStepValue 验证步骤值
func (v *SimpleTestValidator) ValidateStepValue(value interface{}, stepType string) error {
	// 填充步骤需要值
	if stepType == "fill" && value == nil {
		return fmt.Errorf("fill step requires value")
	}

	// 验证值的大小
	if valueStr, ok := value.(string); ok && len(valueStr) > 10000 {
		return fmt.Errorf("step value too long (max 10000 characters)")
	}

	return nil
}

// ValidateWaitTime 验证等待时间
func (v *SimpleTestValidator) ValidateWaitTime(waitTime int) error {
	if waitTime <= 0 {
		return fmt.Errorf("wait time must be positive")
	}

	if waitTime > 300000 { // 5分钟（毫秒）
		return fmt.Errorf("wait time too long (max 5 minutes)")
	}

	return nil
}

// ValidateJSON 验证JSON格式
func (v *SimpleTestValidator) ValidateJSON(jsonStr string) error {
	if jsonStr == "" {
		return nil // 空字符串是有效的
	}

	// 简单的JSON格式检查
	if !strings.HasPrefix(strings.TrimSpace(jsonStr), "{") && !strings.HasPrefix(strings.TrimSpace(jsonStr), "[") {
		return fmt.Errorf("invalid JSON format")
	}

	return nil
}

// ValidateStatus 验证状态码
func (v *SimpleTestValidator) ValidateStatus(status int) error {
	if status < 100 || status > 599 {
		return fmt.Errorf("invalid HTTP status code: %d", status)
	}

	return nil
}

// ValidateContentType 验证内容类型
func (v *SimpleTestValidator) ValidateContentType(contentType string) error {
	if contentType == "" {
		return nil
	}

	// 简单的内容类型验证
	if !strings.Contains(contentType, "/") {
		return fmt.Errorf("invalid content type format: %s", contentType)
	}

	return nil
}

// ValidateResponseTime 验证响应时间
func (v *SimpleTestValidator) ValidateResponseTime(responseTime int) error {
	if responseTime <= 0 {
		return nil // 响应时间可以为0（表示不设置期望）
	}

	if responseTime > 300000 { // 5分钟（毫秒）
		return fmt.Errorf("response time expectation too long (max 5 minutes)")
	}

	return nil
}

// ValidateBodyContains 验证包含的文本
func (v *SimpleTestValidator) ValidateBodyContains(texts []string) error {
	if texts == nil {
		return nil
	}

	for i, text := range texts {
		if text == "" {
			return fmt.Errorf("body contains text %d cannot be empty", i+1)
		}

		if len(text) > 1000 {
			return fmt.Errorf("body contains text %d too long (max 1000 characters): %s", i+1, text)
		}
	}

	return nil
}

// ValidateExecutionID 验证执行ID
func (v *SimpleTestValidator) ValidateExecutionID(executionID string) error {
	if executionID == "" {
		return fmt.Errorf("execution ID cannot be empty")
	}

	if len(executionID) > 100 {
		return fmt.Errorf("execution ID too long (max 100 characters)")
	}

	// 简单的ID格式验证（字母、数字、连字符、下划线）
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, executionID)
	if !matched {
		return fmt.Errorf("invalid execution ID format (must contain only letters, numbers, hyphens and underscores)")
	}

	return nil
}

// ValidateTestStatus 验证测试状态
func (v *SimpleTestValidator) ValidateTestStatus(status string) error {
	if status == "" {
		return fmt.Errorf("test status cannot be empty")
	}

	validStatuses := []string{"pending", "running", "passed", "failed", "skipped", "error", "timeout"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid test status: %s", status)
}

// ValidateProgress 验证进度
func (v *SimpleTestValidator) ValidateProgress(progress float64) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}

	return nil
}

// ValidateSuiteID 验证套件ID
func (v *SimpleTestValidator) ValidateSuiteID(suiteID string) error {
	if suiteID == "" {
		return fmt.Errorf("suite ID cannot be empty")
	}

	if len(suiteID) > 100 {
		return fmt.Errorf("suite ID too long (max 100 characters)")
	}

	return nil
}

// ValidateTestCaseID 验证测试用例ID
func (v *SimpleTestValidator) ValidateTestCaseID(caseID string) error {
	if caseID == "" {
		return fmt.Errorf("test case ID cannot be empty")
	}

	if len(caseID) > 100 {
		return fmt.Errorf("test case ID too long (max 100 characters)")
	}

	return nil
}

// ValidateResultID 验证结果ID
func (v *SimpleTestValidator) ValidateResultID(resultID string) error {
	if resultID == "" {
		return fmt.Errorf("result ID cannot be empty")
	}

	if len(resultID) > 100 {
		return fmt.Errorf("result ID too long (max 100 characters)")
	}

	return nil
}

// ValidatePagination 验证分页参数
func (v *SimpleTestValidator) ValidatePagination(page, pageSize int) error {
	if page < 1 {
		return fmt.Errorf("page must be at least 1")
	}

	if pageSize < 1 || pageSize > 100 {
		return fmt.Errorf("page size must be between 1 and 100")
	}

	return nil
}

// ValidateTimeRange 验证时间范围
func (v *SimpleTestValidator) ValidateTimeRange(startStr, endStr string) error {
	// 这里可以添加时间范围验证逻辑
	// 暂时返回nil，表示验证通过
	return nil
}

// ValidatePageSize 验证页面大小
func (v *SimpleTestValidator) ValidatePageSize(pageSize int) error {
	if pageSize < 1 || pageSize > 100 {
		return fmt.Errorf("page size must be between 1 and 100")
	}

	return nil
}

// ValidateOffset 验证偏移量
func (v *SimpleTestValidator) ValidateOffset(offset int) error {
	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	return nil
}

// ValidateSortField 验证排序字段
func (v *SimpleTestValidator) ValidateSortField(field string, allowedFields []string) error {
	if field == "" {
		return nil // 空字段表示默认排序
	}

	for _, allowedField := range allowedFields {
		if field == allowedField {
			return nil
		}
	}

	return fmt.Errorf("invalid sort field: %s", field)
}

// ValidateSortOrder 验证排序顺序
func (v *SimpleTestValidator) ValidateSortOrder(order string) error {
	if order == "" {
		return nil // 空顺序表示默认排序
	}

	validOrders := []string{"asc", "desc"}
	for _, validOrder := range validOrders {
		if order == validOrder {
			return nil
		}
	}

	return fmt.Errorf("invalid sort order: %s", order)
}