package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// VariableResolver 变量解析器
type VariableResolver struct {
	userRepo  repositories.UserRepository
	roleRepo  repositories.RoleRepository
	logger    *logrus.Logger
	functions map[string]VariableFunction
}

// VariableFunction 变量函数类型
type VariableFunction func(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error)

// NewVariableResolver 创建变量解析器
func NewVariableResolver(userRepo repositories.UserRepository, roleRepo repositories.RoleRepository, logger *logrus.Logger) *VariableResolver {
	resolver := &VariableResolver{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		logger:    logger,
		functions: make(map[string]VariableFunction),
	}

	// 注册内置函数
	resolver.registerBuiltinFunctions()

	return resolver
}

// ResolveVariables 解析变量
func (r *VariableResolver) ResolveVariables(ctx context.Context, req interface{}, conditions []DynamicCondition) (map[string]interface{}, error) {
	variables := make(map[string]interface{})

	// 根据请求类型提取基础变量
	switch v := req.(type) {
	case *PolicyEvaluationRequest:
		variables = r.extractPolicyEvaluationVariables(v)
	case *AccessRequest:
		variables = r.extractAccessRequestVariables(v)
	default:
		return nil, fmt.Errorf("unsupported request type: %T", req)
	}

	// 解析条件中的变量
	for _, condition := range conditions {
		for _, varName := range condition.Variables {
			if _, exists := variables[varName]; !exists {
				value, err := r.resolveVariable(ctx, varName, variables)
				if err != nil {
					r.logger.WithError(err).WithField("variable", varName).Warn("Failed to resolve variable")
					variables[varName] = nil
				} else {
					variables[varName] = value
				}
			}
		}
	}

	return variables, nil
}

// extractPolicyEvaluationVariables 提取策略评估变量
func (r *VariableResolver) extractPolicyEvaluationVariables(req *PolicyEvaluationRequest) map[string]interface{} {
	variables := make(map[string]interface{})

	// 用户相关变量
	variables["user.id"] = req.UserID
	variables["user.username"] = req.Username
	variables["user.tenant_id"] = req.TenantID

	// 资源相关变量
	variables["resource.type"] = req.Resource.Type
	variables["resource.id"] = req.Resource.ID
	variables["resource.owner"] = req.Resource.Owner
	variables["resource.tenant_id"] = req.Resource.TenantID
	variables["resource.sensitivity"] = req.Resource.Sensitivity
	variables["resource.category"] = req.Resource.Category
	variables["resource.created_at"] = req.Resource.CreatedAt
	variables["resource.updated_at"] = req.Resource.UpdatedAt

	// 添加资源属性
	for key, value := range req.Resource.Attributes {
		variables[fmt.Sprintf("resource.%s", key)] = value
	}

	// 动作相关变量
	variables["action.type"] = req.Action.Type
	variables["action.method"] = req.Action.Method

	// 添加动作属性
	for key, value := range req.Action.Attributes {
		variables[fmt.Sprintf("action.%s", key)] = value
	}

	// 环境相关变量
	variables["environment.time"] = req.Environment.Time
	variables["environment.ip"] = req.Environment.IP
	variables["environment.user_agent"] = req.Environment.UserAgent
	variables["environment.device"] = req.Environment.Device
	variables["environment.location"] = req.Environment.Location

	// 添加环境属性
	for key, value := range req.Environment.Attributes {
		variables[fmt.Sprintf("environment.%s", key)] = value
	}

	// 上下文变量
	for key, value := range req.Context {
		variables[fmt.Sprintf("context.%s", key)] = value
	}

	// 时间变量
	now := time.Now()
	variables["time.now"] = now
	variables["time.year"] = now.Year()
	variables["time.month"] = int(now.Month())
	variables["time.day"] = now.Day()
	variables["time.hour"] = now.Hour()
	variables["time.minute"] = now.Minute()
	variables["time.second"] = now.Second()
	variables["time.weekday"] = int(now.Weekday())
	variables["time.unix"] = now.Unix()
	variables["time.iso"] = now.Format(time.RFC3339)
	variables["time.date"] = now.Format("2006-01-02")
	variables["time.time"] = now.Format("15:04:05")

	return variables
}

// extractAccessRequestVariables 提取访问请求变量
func (r *VariableResolver) extractAccessRequestVariables(req *AccessRequest) map[string]interface{} {
	variables := make(map[string]interface{})

	// 主体相关变量
	variables["subject.id"] = req.Subject.ID
	variables["subject.username"] = req.Subject.Username
	variables["subject.email"] = req.Subject.Email
	variables["subject.tenant_id"] = req.Subject.TenantID
	variables["subject.active"] = req.Subject.Active

	// 添加主体属性
	for key, value := range req.Subject.Attributes {
		variables[fmt.Sprintf("subject.%s", key)] = value
	}

	// 角色变量
	variables["subject.roles"] = req.Subject.Roles
	variables["subject.groups"] = req.Subject.Groups

	// 资源相关变量
	variables["resource.type"] = req.Resource.Type
	variables["resource.id"] = req.Resource.ID
	variables["resource.owner"] = req.Resource.Owner
	variables["resource.tenant_id"] = req.Resource.TenantID
	variables["resource.sensitivity"] = req.Resource.Sensitivity
	variables["resource.category"] = req.Resource.Category
	variables["resource.created_at"] = req.Resource.CreatedAt
	variables["resource.updated_at"] = req.Resource.UpdatedAt

	// 添加资源属性
	for key, value := range req.Resource.Attributes {
		variables[fmt.Sprintf("resource.%s", key)] = value
	}

	// 动作相关变量
	variables["action.type"] = req.Action.Type
	variables["action.method"] = req.Action.Method

	// 添加动作属性
	for key, value := range req.Action.Attributes {
		variables[fmt.Sprintf("action.%s", key)] = value
	}

	// 环境相关变量
	variables["environment.time"] = req.Environment.Time
	variables["environment.ip"] = req.Environment.IP
	variables["environment.user_agent"] = req.Environment.UserAgent
	variables["environment.device"] = req.Environment.Device
	variables["environment.location"] = req.Environment.Location

	// 添加环境属性
	for key, value := range req.Environment.Attributes {
		variables[fmt.Sprintf("environment.%s", key)] = value
	}

	// 时间变量
	now := time.Now()
	variables["time.now"] = now
	variables["time.year"] = now.Year()
	variables["time.month"] = int(now.Month())
	variables["time.day"] = now.Day()
	variables["time.hour"] = now.Hour()
	variables["time.minute"] = now.Minute()
	variables["time.second"] = now.Second()
	variables["time.weekday"] = int(now.Weekday())
	variables["time.unix"] = now.Unix()

	return variables
}

// resolveVariable 解析单个变量
func (r *VariableResolver) resolveVariable(ctx context.Context, varName string, variables map[string]interface{}) (interface{}, error) {
	// 检查是否是函数调用
	if strings.Contains(varName, "(") && strings.HasSuffix(varName, ")") {
		return r.resolveFunctionCall(ctx, varName, variables)
	}

	// 检查是否是嵌套属性访问
	if strings.Contains(varName, ".") {
		return r.resolveNestedVariable(varName, variables)
	}

	// 直接变量查找
	if value, exists := variables[varName]; exists {
		return value, nil
	}

	// 内置变量
	switch varName {
	case "null", "nil":
		return nil, nil
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "empty":
		return "", nil
	case "zero":
		return 0, nil
	}

	return nil, fmt.Errorf("variable not found: %s", varName)
}

// resolveFunctionCall 解析函数调用
func (r *VariableResolver) resolveFunctionCall(ctx context.Context, funcCall string, variables map[string]interface{}) (interface{}, error) {
	// 解析函数名和参数
	parts := strings.SplitN(funcCall, "(", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid function call format: %s", funcCall)
	}

	funcName := strings.TrimSpace(parts[1])
	funcName = funcName[:len(funcName)-1] // 移除右括号

	// 解析参数
	var params []string
	paramStr := strings.TrimSpace(parts[0])
	if strings.Contains(paramStr, ".") {
		// 处理对象方法调用，如 "user.hasRole"
		methodParts := strings.Split(paramStr, ".")
		if len(methodParts) >= 2 {
			objectName := methodParts[0]
			methodName := strings.Join(methodParts[1:], ".")
			return r.resolveObjectMethod(ctx, objectName, methodName, variables)
		}
	}

	// 查找函数
	if fn, exists := r.functions[funcName]; exists {
		return fn(ctx, map[string]interface{}{}, variables)
	}

	return nil, fmt.Errorf("function not found: %s", funcName)
}

// resolveObjectMethod 解析对象方法
func (r *VariableResolver) resolveObjectMethod(ctx context.Context, objectName, methodName string, variables map[string]interface{}) (interface{}, error) {
	switch objectName {
	case "user":
		return r.resolveUserMethod(ctx, methodName, variables)
	case "resource":
		return r.resolveResourceMethod(ctx, methodName, variables)
	case "time":
		return r.resolveTimeMethod(ctx, methodName, variables)
	default:
		return nil, fmt.Errorf("unknown object: %s", objectName)
	}
}

// resolveUserMethod 解析用户方法
func (r *VariableResolver) resolveUserMethod(ctx context.Context, methodName string, variables map[string]interface{}) (interface{}, error) {
	userID, exists := variables["user.id"]
	if !exists {
		return nil, fmt.Errorf("user.id not found")
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		return nil, fmt.Errorf("invalid user.id type")
	}

	user, err := r.userRepo.GetByID(ctx, userIDUint)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	switch methodName {
	case "hasRole":
		roleName, exists := variables["role_name"]
		if !exists {
			return false, nil
		}
		roleNameStr, ok := roleName.(string)
		if !ok {
			return false, nil
		}
		return r.userHasRole(ctx, user.ID, roleNameStr), nil
	case "roles":
		return r.getUserRoles(ctx, user.ID), nil
	case "isActive":
		return user.Active, nil
	case "email":
		return user.Email, nil
	case "profile":
		return map[string]interface{}{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"active":     user.Active,
		}, nil
	default:
		return nil, fmt.Errorf("unknown user method: %s", methodName)
	}
}

// resolveResourceMethod 解析资源方法
func (r *VariableResolver) resolveResourceMethod(ctx context.Context, methodName string, variables map[string]interface{}) (interface{}, error) {
	switch methodName {
	case "isOwner":
		ownerID, exists := variables["resource.owner"]
		if !exists {
			return false, nil
		}
		userID, exists := variables["user.id"]
		if !exists {
			return false, nil
		}
		return fmt.Sprintf("%v", ownerID) == fmt.Sprintf("%v", userID), nil
	case "sensitivityLevel":
		sensitivity, exists := variables["resource.sensitivity"]
		if !exists {
			return 0, nil
		}
		return r.getSensitivityLevel(fmt.Sprintf("%v", sensitivity)), nil
	case "age":
		createdAt, exists := variables["resource.created_at"]
		if !exists {
			return 0, nil
		}
		if createdTime, ok := createdAt.(time.Time); ok {
			return time.Since(createdTime).Hours() / 24, nil // 返回天数
		}
		return 0, nil
	default:
		return nil, fmt.Errorf("unknown resource method: %s", methodName)
	}
}

// resolveTimeMethod 解析时间方法
func (r *VariableResolver) resolveTimeMethod(ctx context.Context, methodName string, variables map[string]interface{}) (interface{}, error) {
	now := time.Now()

	switch methodName {
	case "now":
		return now, nil
	case "format":
		format, exists := variables["format"]
		if !exists {
			format = "2006-01-02 15:04:05"
		}
		formatStr, ok := format.(string)
		if !ok {
			formatStr = "2006-01-02 15:04:05"
		}
		return now.Format(formatStr), nil
	case "unix":
		return now.Unix(), nil
	case "iso":
		return now.Format(time.RFC3339), nil
	case "date":
		return now.Format("2006-01-02"), nil
	case "time":
		return now.Format("15:04:05"), nil
	case "isBusinessHours":
		hour := now.Hour()
		return hour >= 9 && hour < 17, nil
	case "isWeekend":
		weekday := now.Weekday()
		return weekday == time.Saturday || weekday == time.Sunday, nil
	default:
		return nil, fmt.Errorf("unknown time method: %s", methodName)
	}
}

// resolveNestedVariable 解析嵌套变量
func (r *VariableResolver) resolveNestedVariable(varName string, variables map[string]interface{}) (interface{}, error) {
	parts := strings.Split(varName, ".")
	current := variables

	for i, part := range parts {
		if value, exists := current[part]; exists {
			if i == len(parts)-1 {
				return value, nil
			}
			if nextMap, ok := value.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, fmt.Errorf("path segment '%s' is not an object", part)
			}
		} else {
			return nil, fmt.Errorf("path segment '%s' not found", part)
		}
	}

	return current, nil
}

// registerBuiltinFunctions 注册内置函数
func (r *VariableResolver) registerBuiltinFunctions() {
	// 字符串函数
	r.functions["upper"] = r.builtinUpper
	r.functions["lower"] = r.builtinLower
	r.functions["length"] = r.builtinLength
	r.functions["contains"] = r.builtinContains
	r.functions["startsWith"] = r.builtinStartsWith
	r.functions["endsWith"] = r.builtinEndsWith
	r.functions["split"] = r.builtinSplit
	r.functions["join"] = r.builtinJoin
	r.functions["replace"] = r.builtinReplace
	r.functions["trim"] = r.builtinTrim

	// 数学函数
	r.functions["add"] = r.builtinAdd
	r.functions["subtract"] = r.builtinSubtract
	r.functions["multiply"] = r.builtinMultiply
	r.functions["divide"] = r.builtinDivide
	r.functions["mod"] = r.builtinMod
	r.functions["abs"] = r.builtinAbs
	r.functions["max"] = r.builtinMax
	r.functions["min"] = r.builtinMin

	// 时间函数
	r.functions["parseDate"] = r.builtinParseDate
	r.functions["formatDate"] = r.builtinFormatDate
	r.functions["dateDiff"] = r.builtinDateDiff
	r.functions["dateAdd"] = r.builtinDateAdd

	// 类型转换函数
	r.functions["toString"] = r.builtinToString
	r.functions["toInt"] = r.builtinToInt
	r.functions["toFloat"] = r.builtinToFloat
	r.functions["toBool"] = r.builtinToBool

	// 集合函数
	r.functions["size"] = r.builtinSize
	r.functions["isEmpty"] = r.builtinIsEmpty
	r.functions["contains"] = r.builtinContains
	r.functions["union"] = r.builtinUnion
	r.functions["intersect"] = r.builtinIntersect
	r.functions["difference"] = r.builtinDifference
}

// 内置函数实现
func (r *VariableResolver) builtinUpper(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if str, exists := params["str"]; exists {
		return strings.ToUpper(fmt.Sprintf("%v", str)), nil
	}
	return "", nil
}

func (r *VariableResolver) builtinLower(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if str, exists := params["str"]; exists {
		return strings.ToLower(fmt.Sprintf("%v", str)), nil
	}
	return "", nil
}

func (r *VariableResolver) builtinLength(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		switch v := value.(type) {
		case string:
			return len(v), nil
		case []interface{}:
			return len(v), nil
		case map[string]interface{}:
			return len(v), nil
		default:
			return len(fmt.Sprintf("%v", v)), nil
		}
	}
	return 0, nil
}

func (r *VariableResolver) builtinContains(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	container, containerExists := params["container"]
	item, itemExists := params["item"]
	if !containerExists || !itemExists {
		return false, nil
	}

	containerStr := fmt.Sprintf("%v", container)
	itemStr := fmt.Sprintf("%v", item)

	return strings.Contains(containerStr, itemStr), nil
}

func (r *VariableResolver) builtinStartsWith(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	str, strExists := params["str"]
	prefix, prefixExists := params["prefix"]
	if !strExists || !prefixExists {
		return false, nil
	}

	strStr := fmt.Sprintf("%v", str)
	prefixStr := fmt.Sprintf("%v", prefix)

	return strings.HasPrefix(strStr, prefixStr), nil
}

func (r *VariableResolver) builtinEndsWith(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	str, strExists := params["str"]
	suffix, suffixExists := params["suffix"]
	if !strExists || !suffixExists {
		return false, nil
	}

	strStr := fmt.Sprintf("%v", str)
	suffixStr := fmt.Sprintf("%v", suffix)

	return strings.HasSuffix(strStr, suffixStr), nil
}

func (r *VariableResolver) builtinSplit(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	str, strExists := params["str"]
	sep, sepExists := params["separator"]
	if !strExists {
		return []string{}, nil
	}

	strStr := fmt.Sprintf("%v", str)
	sepStr := ","
	if sepExists {
		sepStr = fmt.Sprintf("%v", sep)
	}

	return strings.Split(strStr, sepStr), nil
}

func (r *VariableResolver) builtinJoin(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	items, itemsExists := params["items"]
	sep, sepExists := params["separator"]
	if !itemsExists {
		return "", nil
	}

	sepStr := ","
	if sepExists {
		sepStr = fmt.Sprintf("%v", sep)
	}

	switch itemsVal := items.(type) {
	case []interface{}:
		var strItems []string
		for _, item := range itemsVal {
			strItems = append(strItems, fmt.Sprintf("%v", item))
		}
		return strings.Join(strItems, sepStr), nil
	case []string:
		return strings.Join(itemsVal, sepStr), nil
	default:
		return fmt.Sprintf("%v", items), nil
	}
}

func (r *VariableResolver) builtinReplace(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	str, strExists := params["str"]
	old, oldExists := params["old"]
	new, newExists := params["new"]
	if !strExists || !oldExists {
		return "", nil
	}

	strStr := fmt.Sprintf("%v", str)
	oldStr := fmt.Sprintf("%v", old)
	newStr := ""
	if newExists {
		newStr = fmt.Sprintf("%v", new)
	}

	return strings.ReplaceAll(strStr, oldStr, newStr), nil
}

func (r *VariableResolver) builtinTrim(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if str, exists := params["str"]; exists {
		return strings.TrimSpace(fmt.Sprintf("%v", str)), nil
	}
	return "", nil
}

func (r *VariableResolver) builtinAdd(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return 0, nil
	}

	aFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, err
	}

	bFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
	if err != nil {
		return 0, err
	}

	return aFloat + bFloat, nil
}

func (r *VariableResolver) builtinSubtract(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return 0, nil
	}

	aFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, err
	}

	bFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
	if err != nil {
		return 0, err
	}

	return aFloat - bFloat, nil
}

func (r *VariableResolver) builtinMultiply(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return 0, nil
	}

	aFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, err
	}

	bFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
	if err != nil {
		return 0, err
	}

	return aFloat * bFloat, nil
}

func (r *VariableResolver) builtinDivide(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return 0, nil
	}

	aFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, err
	}

	bFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
	if err != nil {
		return 0, err
	}

	if bFloat == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	return aFloat / bFloat, nil
}

func (r *VariableResolver) builtinMod(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return 0, nil
	}

	aFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, err
	}

	bFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
	if err != nil {
		return 0, err
	}

	if bFloat == 0 {
		return 0, fmt.Errorf("modulo by zero")
	}

	return int(aFloat) % int(bFloat), nil
}

func (r *VariableResolver) builtinAbs(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		valueFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
		if err != nil {
			return 0, err
		}
		if valueFloat < 0 {
			return -valueFloat, nil
		}
		return valueFloat, nil
	}
	return 0, nil
}

func (r *VariableResolver) builtinMax(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return 0, nil
	}

	aFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, err
	}

	bFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
	if err != nil {
		return 0, err
	}

	if aFloat > bFloat {
		return aFloat, nil
	}
	return bFloat, nil
}

func (r *VariableResolver) builtinMin(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return 0, nil
	}

	aFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", a), 64)
	if err != nil {
		return 0, err
	}

	bFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", b), 64)
	if err != nil {
		return 0, err
	}

	if aFloat < bFloat {
		return aFloat, nil
	}
	return bFloat, nil
}

func (r *VariableResolver) builtinParseDate(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	dateStr, dateExists := params["date"]
	format, formatExists := params["format"]
	if !dateExists {
		return time.Time{}, fmt.Errorf("date parameter required")
	}

	dateStrValue := fmt.Sprintf("%v", dateStr)
	formatStr := "2006-01-02T15:04:05Z"
	if formatExists {
		formatStr = fmt.Sprintf("%v", format)
	}

	return time.Parse(formatStr, dateStrValue)
}

func (r *VariableResolver) builtinFormatDate(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	date, dateExists := params["date"]
	format, formatExists := params["format"]
	if !dateExists {
		return "", fmt.Errorf("date parameter required")
	}

	var dateValue time.Time
	switch v := date.(type) {
	case time.Time:
		dateValue = v
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return "", err
		}
		dateValue = parsed
	default:
		return "", fmt.Errorf("invalid date type")
	}

	formatStr := "2006-01-02 15:04:05"
	if formatExists {
		formatStr = fmt.Sprintf("%v", format)
	}

	return dateValue.Format(formatStr), nil
}

func (r *VariableResolver) builtinDateDiff(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	start, startExists := params["start"]
	end, endExists := params["end"]
	unit, unitExists := params["unit"]
	if !startExists || !endExists {
		return 0, fmt.Errorf("start and end parameters required")
	}

	unitStr := "days"
	if unitExists {
		unitStr = fmt.Sprintf("%v", unit)
	}

	// 简化实现，只支持字符串时间格式
	startStr := fmt.Sprintf("%v", start)
	endStr := fmt.Sprintf("%v", end)

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		startTime, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return 0, err
		}
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		endTime, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return 0, err
		}
	}

	duration := endTime.Sub(startTime)

	switch unitStr {
	case "seconds":
		return duration.Seconds(), nil
	case "minutes":
		return duration.Minutes(), nil
	case "hours":
		return duration.Hours(), nil
	case "days":
		return duration.Hours() / 24, nil
	default:
		return duration.Seconds(), nil
	}
}

func (r *VariableResolver) builtinDateAdd(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	date, dateExists := params["date"]
	duration, durationExists := params["duration"]
	if !dateExists || !durationExists {
		return time.Time{}, fmt.Errorf("date and duration parameters required")
	}

	var dateValue time.Time
	switch v := date.(type) {
	case time.Time:
		dateValue = v
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, err
		}
		dateValue = parsed
	default:
		return time.Time{}, fmt.Errorf("invalid date type")
	}

	durationStr := fmt.Sprintf("%v", duration)
	durationValue, err := time.ParseDuration(durationStr)
	if err != nil {
		return time.Time{}, err
	}

	return dateValue.Add(durationValue), nil
}

func (r *VariableResolver) builtinToString(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		return fmt.Sprintf("%v", value), nil
	}
	return "", nil
}

func (r *VariableResolver) builtinToInt(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		return strconv.Atoi(fmt.Sprintf("%v", value))
	}
	return 0, nil
}

func (r *VariableResolver) builtinToFloat(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		return strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	}
	return 0.0, nil
}

func (r *VariableResolver) builtinToBool(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		return strconv.ParseBool(fmt.Sprintf("%v", value))
	}
	return false, nil
}

func (r *VariableResolver) builtinSize(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		switch v := value.(type) {
		case string:
			return len(v), nil
		case []interface{}:
			return len(v), nil
		case map[string]interface{}:
			return len(v), nil
		default:
			return 0, nil
		}
	}
	return 0, nil
}

func (r *VariableResolver) builtinIsEmpty(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	if value, exists := params["value"]; exists {
		switch v := value.(type) {
		case string:
			return len(v) == 0, nil
		case []interface{}:
			return len(v) == 0, nil
		case map[string]interface{}:
			return len(v) == 0, nil
		case nil:
			return true, nil
		default:
			return false, nil
		}
	}
	return true, nil
}

func (r *VariableResolver) builtinUnion(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return []interface{}{}, nil
	}

	var result []interface{}

	if aSlice, ok := a.([]interface{}); ok {
		result = append(result, aSlice...)
	}

	if bSlice, ok := b.([]interface{}); ok {
		result = append(result, bSlice...)
	}

	return result, nil
}

func (r *VariableResolver) builtinIntersect(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return []interface{}{}, nil
	}

	aSlice, aOk := a.([]interface{})
	bSlice, bOk := b.([]interface{})
	if !aOk || !bOk {
		return []interface{}{}, nil
	}

	var result []interface{}
	seen := make(map[string]bool)

	for _, item := range aSlice {
		itemStr := fmt.Sprintf("%v", item)
		seen[itemStr] = true
	}

	for _, item := range bSlice {
		itemStr := fmt.Sprintf("%v", item)
		if seen[itemStr] {
			result = append(result, item)
		}
	}

	return result, nil
}

func (r *VariableResolver) builtinDifference(ctx context.Context, params map[string]interface{}, variables map[string]interface{}) (interface{}, error) {
	a, aExists := params["a"]
	b, bExists := params["b"]
	if !aExists || !bExists {
		return []interface{}{}, nil
	}

	aSlice, aOk := a.([]interface{})
	bSlice, bOk := b.([]interface{})
	if !aOk || !bOk {
		return []interface{}{}, nil
	}

	var result []interface{}
	seen := make(map[string]bool)

	for _, item := range bSlice {
		itemStr := fmt.Sprintf("%v", item)
		seen[itemStr] = true
	}

	for _, item := range aSlice {
		itemStr := fmt.Sprintf("%v", item)
		if !seen[itemStr] {
			result = append(result, item)
		}
	}

	return result, nil
}

// 辅助方法
func (r *VariableResolver) userHasRole(ctx context.Context, userID uint, roleName string) bool {
	roles, err := r.userRepo.GetRoles(ctx, userID, "")
	if err != nil {
		return false
	}

	for _, userRole := range roles {
		if userRole.Role.Name == roleName {
			return true
		}
	}

	return false
}

func (r *VariableResolver) getUserRoles(ctx context.Context, userID uint) []string {
	roles, err := r.userRepo.GetRoles(ctx, userID, "")
	if err != nil {
		return []string{}
	}

	var roleNames []string
	for _, userRole := range roles {
		roleNames = append(roleNames, userRole.Role.Name)
	}

	return roleNames
}

func (r *VariableResolver) getSensitivityLevel(sensitivity string) int {
	switch strings.ToLower(sensitivity) {
	case "public":
		return 1
	case "internal":
		return 2
	case "confidential":
		return 3
	case "secret":
		return 4
	case "top_secret":
		return 5
	default:
		return 0
	}
}