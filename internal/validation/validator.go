package validation

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	validationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "validation_duration_seconds",
		Help:    "Duration of validation operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})
	
	validationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "validation_errors_total",
		Help: "Total number of validation errors",
	}, []string{"type", "field"})
	
	sanitizationOperations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sanitization_operations_total",
		Help: "Total number of sanitization operations",
	})
)

// Validator 验证器
type Validator struct {
	validator    *validator.Validate
	cacheService *cache.CacheService
	db           *gorm.DB
}

// NewValidator 创建新的验证器
func NewValidator(db *gorm.DB, cacheService *cache.CacheService) *Validator {
	v := validator.New()
	
	// 注册自定义验证函数
	v.RegisterValidation("password_strength", validatePasswordStrength)
	v.RegisterValidation("phone_number", validatePhoneNumber)
	v.RegisterValidation("id_card", validateIDCard)
	v.RegisterValidation("safe_string", validateSafeString)
	v.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	v.RegisterValidation("no_xss", validateNoXSS)
	v.RegisterValidation("file_path", validateFilePath)
	v.RegisterValidation("email_domain", validateEmailDomain)
	
	return &Validator{
		validator:    v,
		cacheService: cacheService,
		db:           db,
	}
}

// ValidationErrors 验证错误集合
type ValidationErrors struct {
	Errors map[string]string `json:"errors"`
}

// Error 实现error接口
func (ve *ValidationErrors) Error() string {
	return fmt.Sprintf("validation failed: %v", ve.Errors)
}

// AddError 添加错误
func (ve *ValidationErrors) AddError(field, message string) {
	if ve.Errors == nil {
		ve.Errors = make(map[string]string)
	}
	ve.Errors[field] = message
}

// HasErrors 检查是否有错误
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// ValidateStruct 验证结构体
func (v *Validator) ValidateStruct(ctx context.Context, s interface{}) *ValidationErrors {
	start := time.Now()
	defer func() {
		validationDuration.WithLabelValues("struct").Observe(time.Since(start).Seconds())
	}()
	
	errors := &ValidationErrors{}
	
	if err := v.validator.Struct(s); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fieldName := err.Field()
			tag := err.Tag()
			
			var message string
			switch tag {
			case "required":
				message = fmt.Sprintf("%s 是必填字段", fieldName)
			case "min":
				message = fmt.Sprintf("%s 长度不能小于 %s", fieldName, err.Param())
			case "max":
				message = fmt.Sprintf("%s 长度不能大于 %s", fieldName, err.Param())
			case "email":
				message = fmt.Sprintf("%s 邮箱格式不正确", fieldName)
			case "password_strength":
				message = fmt.Sprintf("%s 密码强度不足，需要包含大小写字母、数字和特殊字符", fieldName)
			case "phone_number":
				message = fmt.Sprintf("%s 手机号码格式不正确", fieldName)
			case "id_card":
				message = fmt.Sprintf("%s 身份证号码格式不正确", fieldName)
			case "safe_string":
				message = fmt.Sprintf("%s 包含不安全字符", fieldName)
			case "no_sql_injection":
				message = fmt.Sprintf("%s 包含SQL注入风险", fieldName)
			case "no_xss":
				message = fmt.Sprintf("%s 包含XSS风险", fieldName)
			case "file_path":
				message = fmt.Sprintf("%s 文件路径格式不正确", fieldName)
			case "email_domain":
				message = fmt.Sprintf("%s 邮箱域名不被允许", fieldName)
			default:
				message = fmt.Sprintf("%s 验证失败: %s", fieldName, tag)
			}
			
			errors.AddError(fieldName, message)
			validationErrors.WithLabelValues("struct", fieldName).Inc()
		}
	}
	
	if errors.HasErrors() {
		return errors
	}
	
	return nil
}

// ValidatePasswordStrength 验证密码强度
func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	// 密码长度至少8位
	if len(password) < 8 {
		return false
	}
	
	// 包含大写字母
	hasUpper := false
	// 包含小写字母
	hasLower := false
	// 包含数字
	hasNumber := false
	// 包含特殊字符
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// ValidatePhoneNumber 验证手机号码
func validatePhoneNumber(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	
	// 中国手机号码正则
	chinesePhoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return chinesePhoneRegex.MatchString(phone)
}

// ValidateIDCard 验证身份证号码
func validateIDCard(fl validator.FieldLevel) bool {
	idCard := fl.Field().String()
	
	// 中国身份证号码正则（简化版）
	idCardRegex := regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dX]$`)
	return idCardRegex.MatchString(idCard)
}

// ValidateSafeString 验证安全字符串
func validateSafeString(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	
	// 检查是否包含危险字符
	dangerousChars := []string{"<", ">", "'", "\"", "&", "|", ";", "$", "`", "\\"}
	for _, char := range dangerousChars {
		if strings.Contains(str, char) {
			return false
		}
	}
	
	return true
}

// ValidateNoSQLInjection 验证无SQL注入
func validateNoSQLInjection(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	
	// SQL注入检测模式
	sqlPatterns := []string{
		`(?i)(union\s+select|insert\s+into|delete\s+from|update\s+set|drop\s+table)`,
		`(?i)(or\s+1\s*=\s*1|or\s+1\s*=\s*'1')`,
		`(?i)(;|--|\/\*|\*\/)`,
		`(?i)(exec|execute|sp_|xp_)`,
	}
	
	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, str); matched {
			return false
		}
	}
	
	return true
}

// ValidateNoXSS 验证无XSS
func validateNoXSS(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	
	// XSS检测模式
	xssPatterns := []string{
		`(?i)(<script|javascript:|onload=|onerror=)`,
		`(?i)(<iframe|<object|<embed|<applet)`,
		`(?i)(eval\(|alert\(|confirm\(|prompt\()`,
		`(?i)(document\.|window\.|location\.)`,
	}
	
	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, str); matched {
			return false
		}
	}
	
	return true
}

// ValidateFilePath 验证文件路径
func validateFilePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	
	// 检查路径遍历攻击
	if strings.Contains(path, "..") || strings.Contains(path, "~") {
		return false
	}
	
	// 检查绝对路径
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return false
	}
	
	// 检查设备路径
	devicePatterns := []string{
		`[A-Za-z]:\\`,
		`\\\\`,
		`//`,
	}
	
	for _, pattern := range devicePatterns {
		if matched, _ := regexp.MatchString(pattern, path); matched {
			return false
		}
	}
	
	return true
}

// ValidateEmailDomain 验证邮箱域名
func validateEmailDomain(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	
	// 提取域名
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	
	domain := parts[1]
	
	// 检查域名是否在允许列表中（简化版）
	allowedDomains := []string{
		"gmail.com", "qq.com", "163.com", "126.com", "outlook.com",
		"hotmail.com", "yahoo.com", "sina.com", "sohu.com",
	}
	
	for _, allowedDomain := range allowedDomains {
		if domain == allowedDomain {
			return true
		}
	}
	
	return false
}

// SanitizeInput 清理输入
func (v *Validator) SanitizeInput(input string) string {
	start := time.Now()
	defer func() {
		validationDuration.WithLabelValues("sanitize").Observe(time.Since(start).Seconds())
		sanitizationOperations.Inc()
	}()
	
	// HTML转义
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	
	// 移除危险字符
	dangerousChars := []string{"|", ";", "$", "`", "\\", "\x00", "\x01", "\x02", "\x03", "\x04", "\x05", "\x06", "\x07", "\x08", "\x0b", "\x0c", "\x0e", "\x0f", "\x10", "\x11", "\x12", "\x13", "\x14", "\x15", "\x16", "\x17", "\x18", "\x19", "\x1a", "\x1b", "\x1c", "\x1d", "\x1e", "\x1f"}
	for _, char := range dangerousChars {
		input = strings.ReplaceAll(input, char, "")
	}
	
	return input
}

// SanitizeSQLInput 清理SQL输入
func (v *Validator) SanitizeSQLInput(input string) string {
	start := time.Now()
	defer func() {
		validationDuration.WithLabelValues("sanitize_sql").Observe(time.Since(start).Seconds())
		sanitizationOperations.Inc()
	}()
	
	// 转义SQL特殊字符
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "\"", "\"\"")
	input = strings.ReplaceAll(input, "\\", "\\\\")
	
	return input
}

// ValidationMiddleware 验证中间件
func (v *Validator) ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			validationDuration.WithLabelValues("middleware").Observe(time.Since(start).Seconds())
		}()
		
		// 验证查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if !v.validateQueryParam(key, value) {
					validationErrors.WithLabelValues("query_param", key).Inc()
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    400,
						"message": "查询参数验证失败",
						"error":   fmt.Sprintf("Invalid query parameter: %s", key),
					})
					c.Abort()
					return
				}
			}
		}
		
		// 验证请求头
		for key, values := range c.Request.Header {
			if v.shouldValidateHeader(key) {
				for _, value := range values {
					if !v.validateHeader(key, value) {
						validationErrors.WithLabelValues("header", key).Inc()
						c.JSON(http.StatusBadRequest, gin.H{
							"code":    400,
							"message": "请求头验证失败",
							"error":   fmt.Sprintf("Invalid header: %s", key),
						})
						c.Abort()
						return
					}
				}
			}
		}
		
		c.Next()
	}
}

// SQLInjectionProtectionMiddleware SQL注入防护中间件
func (v *Validator) SQLInjectionProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			validationDuration.WithLabelValues("sql_injection_protection").Observe(time.Since(start).Seconds())
		}()
		
		// 检查URL
		if v.detectSQLInjection(c.Request.URL.String()) {
			validationErrors.WithLabelValues("sql_injection", "url").Inc()
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "检测到SQL注入攻击",
				"error":   "SQL injection detected",
			})
			c.Abort()
			return
		}
		
		// 检查查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if v.detectSQLInjection(value) {
					validationErrors.WithLabelValues("sql_injection", "query").Inc()
					c.JSON(http.StatusForbidden, gin.H{
						"code":    403,
						"message": "检测到SQL注入攻击",
						"error":   "SQL injection detected",
					})
					c.Abort()
					return
				}
			}
		}
		
		// 检查POST数据
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for _, value := range values {
						if v.detectSQLInjection(value) {
							validationErrors.WithLabelValues("sql_injection", "post").Inc()
							c.JSON(http.StatusForbidden, gin.H{
								"code":    403,
								"message": "检测到SQL注入攻击",
								"error":   "SQL injection detected",
							})
							c.Abort()
							return
						}
					}
				}
			}
		}
		
		c.Next()
	}
}

// XSSProtectionMiddleware XSS防护中间件
func (v *Validator) XSSProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			validationDuration.WithLabelValues("xss_protection").Observe(time.Since(start).Seconds())
		}()
		
		// 检查URL
		if v.detectXSS(c.Request.URL.String()) {
			validationErrors.WithLabelValues("xss", "url").Inc()
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "检测到XSS攻击",
				"error":   "XSS detected",
			})
			c.Abort()
			return
		}
		
		// 检查查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if v.detectXSS(value) {
					validationErrors.WithLabelValues("xss", "query").Inc()
					c.JSON(http.StatusForbidden, gin.H{
						"code":    403,
						"message": "检测到XSS攻击",
						"error":   "XSS detected",
					})
					c.Abort()
					return
				}
			}
		}
		
		// 检查POST数据
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for _, value := range values {
						if v.detectXSS(value) {
							validationErrors.WithLabelValues("xss", "post").Inc()
							c.JSON(http.StatusForbidden, gin.H{
								"code":    403,
								"message": "检测到XSS攻击",
								"error":   "XSS detected",
							})
							c.Abort()
							return
						}
					}
				}
			}
		}
		
		c.Next()
	}
}

// FileUploadValidationMiddleware 文件上传验证中间件
func (v *Validator) FileUploadValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			validationDuration.WithLabelValues("file_upload_validation").Observe(time.Since(start).Seconds())
		}()
		
		// 检查Content-Type
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			validationErrors.WithLabelValues("file_upload", "content_type").Inc()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "不支持的文件上传格式",
				"error":   "Unsupported content type",
			})
			c.Abort()
			return
		}
		
		// 检查文件大小
		if c.Request.ContentLength > 50*1024*1024 { // 50MB限制
			validationErrors.WithLabelValues("file_upload", "size").Inc()
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":    413,
				"message": "文件大小超过限制",
				"error":   "File size exceeded",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// DatabaseInjectionProtection 数据库注入防护
func (v *Validator) DatabaseInjectionProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			validationDuration.WithLabelValues("database_injection_protection").Observe(time.Since(start).Seconds())
		}()
		
		// 检查数据库查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if v.detectDatabaseInjection(value) {
					validationErrors.WithLabelValues("database_injection", "query").Inc()
					c.JSON(http.StatusForbidden, gin.H{
						"code":    403,
						"message": "检测到数据库注入攻击",
						"error":   "Database injection detected",
					})
					c.Abort()
					return
				}
			}
		}
		
		c.Next()
	}
}

// 辅助方法
func (v *Validator) validateQueryParam(key, value string) bool {
	// 跳过一些无害的查询参数
	skipParams := []string{"page", "size", "sort", "order", "search"}
	for _, param := range skipParams {
		if key == param {
			return true
		}
	}
	
	// 基本验证
	if len(value) > 1000 { // 限制查询参数长度
		return false
	}
	
	// 检查SQL注入
	if v.detectSQLInjection(value) {
		return false
	}
	
	// 检查XSS
	if v.detectXSS(value) {
		return false
	}
	
	return true
}

func (v *Validator) validateHeader(key, value string) bool {
	// 跳过一些无害的请求头
	skipHeaders := []string{"User-Agent", "Accept", "Accept-Language", "Accept-Encoding", "Connection", "Cache-Control"}
	for _, header := range skipHeaders {
		if key == header {
			return true
		}
	}
	
	// 基本验证
	if len(value) > 500 { // 限制请求头长度
		return false
	}
	
	// 检查SQL注入
	if v.detectSQLInjection(value) {
		return false
	}
	
	// 检查XSS
	if v.detectXSS(value) {
		return false
	}
	
	return true
}

func (v *Validator) shouldValidateHeader(key string) bool {
	// 需要验证的请求头
	validateHeaders := []string{"Authorization", "X-API-Key", "X-Request-ID", "Referer", "Origin"}
	for _, header := range validateHeaders {
		if key == header {
			return true
		}
	}
	return false
}

func (v *Validator) detectSQLInjection(input string) bool {
	sqlPatterns := []string{
		`(?i)(union\s+select|insert\s+into|delete\s+from|update\s+set|drop\s+table|alter\s+table)`,
		`(?i)(or\s+1\s*=\s*1|or\s+1\s*=\s*'1'|and\s+1\s*=\s*1)`,
		`(?i)(;|--|\/\*|\*\/|@@|@)`,
		`(?i)(exec|execute|sp_|xp_|select|insert|update|delete|drop|create|alter)`,
		`(?i)(waitfor\s+delay|sleep\(|benchmark\()`,
		`(?i)(load_file|into\s+outfile|into\s+dumpfile)`,
		`(?i)(information_schema|mysql\.|pg_|sys\.)`,
		`(?i)(0x[0-9a-f]+|char\(|ascii\(|concat\(|group_concat\()`,
	}
	
	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

func (v *Validator) detectXSS(input string) bool {
	xssPatterns := []string{
		`(?i)(<script|javascript:|onload=|onerror=|onclick=|onmouseover=|onfocus=)`,
		`(?i)(<iframe|<object|<embed|<applet|<meta|<link)`,
		`(?i)(eval\(|alert\(|confirm\(|prompt\(|setInterval\(|setTimeout\()`,
		`(?i)(document\.|window\.|location\.|self\.|top\.|parent\.)`,
		`(?i)(expression\(|script:|data:text/html|vbscript:)`,
		`(?i)(fromCharCode|String\.fromCharCode)`,
		`(?i)(\x00|\x01|\x02|\x03|\x04|\x05|\x06|\x07|\x08|\x0b|\x0c|\x0e|\x0f|\x10|\x11|\x12|\x13|\x14|\x15|\x16|\x17|\x18|\x19|\x1a|\x1b|\x1c|\x1d|\x1e|\x1f)`,
	}
	
	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

func (v *Validator) detectDatabaseInjection(input string) bool {
	// 更广泛的数据库注入检测
	injectionPatterns := []string{
		`(?i)(union\s+select|union\s+all\s+select)`,
		`(?i)(insert\s+into|update\s+set|delete\s+from)`,
		`(?i)(drop\s+table|truncate\s+table|alter\s+table)`,
		`(?i)(create\s+table|create\s+view|create\s+index)`,
		`(?i)(exec|execute|sp_executesql)`,
		`(?i)(waitfor\s+delay|pg_sleep|benchmark)`,
		`(?i)(xp_cmdshell|xp_regread|xp_regwrite)`,
		`(?i)(information_schema|sys.tables|sys.columns)`,
		`(?i)(load_file|bulk\s+insert|bcp)`,
		`(?i)(or\s+1\s*=\s*1|and\s+1\s*=\s*1)`,
		`(?i)(\|\||&&|like\s'%|%')`,
	}
	
	for _, pattern := range injectionPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

// SafeQueryParam 安全查询参数
func (v *Validator) SafeQueryParam(c *gin.Context, key string) (string, error) {
	value := c.Query(key)
	if value == "" {
		return "", nil
	}
	
	// 验证参数
	if !v.validateQueryParam(key, value) {
		return "", fmt.Errorf("invalid query parameter: %s", key)
	}
	
	// 清理输入
	return v.SanitizeInput(value), nil
}

// SafeQueryParamWithDefault 安全查询参数（带默认值）
func (v *Validator) SafeQueryParamWithDefault(c *gin.Context, key, defaultValue string) string {
	value, err := v.SafeQueryParam(c, key)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}

// SafeQueryParamInt 安全查询参数（整型）
func (v *Validator) SafeQueryParamInt(c *gin.Context, key string) (int, error) {
	value, err := v.SafeQueryParam(c, key)
	if err != nil {
		return 0, err
	}
	
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer parameter: %s", key)
	}
	
	return result, nil
}

// SafeQueryParamIntWithDefault 安全查询参数（整型，带默认值）
func (v *Validator) SafeQueryParamIntWithDefault(c *gin.Context, key string, defaultValue int) int {
	value, err := v.SafeQueryParamInt(c, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// ValidatePagination 验证分页参数
func (v *Validator) ValidatePagination(c *gin.Context) (page, size int, err error) {
	page = v.SafeQueryParamIntWithDefault(c, "page", 1)
	size = v.SafeQueryParamIntWithDefault(c, "size", 10)
	
	// 验证分页参数范围
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	
	return page, size, nil
}

// RegisterCustomValidation 注册自定义验证函数
func (v *Validator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// ValidateField 验证单个字段
func (v *Validator) ValidateField(field interface{}, tag string) bool {
	err := v.validator.Var(field, tag)
	return err == nil
}

// GetValidationErrors 获取验证错误详情
func (v *Validator) GetValidationErrors(err error) *ValidationErrors {
	if err == nil {
		return nil
	}
	
	errors := &ValidationErrors{}
	
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			fieldName := fieldErr.Field()
			tag := fieldErr.Tag()
			
			var message string
			switch tag {
			case "required":
				message = fmt.Sprintf("%s 是必填字段", fieldName)
			case "email":
				message = fmt.Sprintf("%s 邮箱格式不正确", fieldName)
			case "min":
				message = fmt.Sprintf("%s 长度不能小于 %s", fieldName, fieldErr.Param())
			case "max":
				message = fmt.Sprintf("%s 长度不能大于 %s", fieldName, fieldErr.Param())
			default:
				message = fmt.Sprintf("%s 验证失败: %s", fieldName, tag)
			}
			
			errors.AddError(fieldName, message)
		}
	}
	
	if errors.HasErrors() {
		return errors
	}
	
	return nil
}

// GetFieldValidationRules 获取字段的验证规则
func (v *Validator) GetFieldValidationRules(structType reflect.Type, fieldName string) []string {
	field, found := structType.Elem().FieldByName(fieldName)
	if !found {
		return nil
	}
	
	tag := field.Tag.Get("validate")
	if tag == "" {
		return nil
	}
	
	return strings.Split(tag, ",")
}

// IsValidType 检查数据类型是否有效
func (v *Validator) IsValidType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "int":
		_, ok := value.(int)
		return ok
	case "float64":
		_, ok := value.(float64)
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "time.Time":
		_, ok := value.(time.Time)
		return ok
	default:
		return false
	}
}

// ValidateLength 验证长度
func (v *Validator) ValidateLength(value string, min, max int) bool {
	length := len(value)
	return length >= min && length <= max
}

// ValidateRegex 验证正则表达式
func (v *Validator) ValidateRegex(value, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

// ValidateIn 验证值是否在允许的列表中
func (v *Validator) ValidateIn(value interface{}, allowedValues []interface{}) bool {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return true
		}
	}
	return false
}

// ValidateRequired 验证必填字段
func (v *Validator) ValidateRequired(value interface{}) bool {
	if value == nil {
		return false
	}
	
	switch v := value.(type) {
	case string:
		return v != ""
	case int:
		return true
	case float64:
		return true
	case bool:
		return true
	case []interface{}:
		return len(v) > 0
	default:
		return true
	}
}