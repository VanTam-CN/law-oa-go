package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RuleExecutor 规则执行器
type RuleExecutor struct {
	logger *logrus.Logger
}

// NewRuleExecutor 创建规则执行器
func NewRuleExecutor(logger *logrus.Logger) *RuleExecutor {
	return &RuleExecutor{
		logger: logger,
	}
}

// RuleExecutionResult 规则执行结果
type RuleExecutionResult struct {
	Success    bool                   `json:"success"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
	ExecutedAt time.Time              `json:"executed_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ExecuteRule 执行规则
func (e *RuleExecutor) ExecuteRule(ctx context.Context, rule DynamicRule, variables map[string]interface{}) (*RuleExecutionResult, error) {
	startTime := time.Now()

	e.logger.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"rule_name": rule.Name,
		"rule_type": rule.Type,
	}).Debug("Executing rule")

	result := &RuleExecutionResult{
		Output:     make(map[string]interface{}),
		Duration:   0,
		ExecutedAt: startTime,
		Metadata:   make(map[string]interface{}),
	}

	// 根据规则类型执行
	var err error
	switch rule.Type {
	case "function":
		err = e.executeFunctionRule(ctx, rule, variables, result)
	case "script":
		err = e.executeScriptRule(ctx, rule, variables, result)
	case "api":
		err = e.executeAPIRule(ctx, rule, variables, result)
	case "decision_table":
		err = e.executeDecisionTableRule(ctx, rule, variables, result)
	default:
		err = fmt.Errorf("unknown rule type: %s", rule.Type)
	}

	result.Duration = time.Since(startTime)
	result.Success = err == nil

	if err != nil {
		result.Error = err.Error()
		e.logger.WithError(err).WithField("rule_id", rule.ID).Error("Rule execution failed")
	} else {
		e.logger.WithFields(logrus.Fields{
			"rule_id": rule.ID,
			"duration": result.Duration.Milliseconds(),
		}).Debug("Rule execution completed successfully")
	}

	// 添加元数据
	result.Metadata["rule_type"] = rule.Type
	result.Metadata["input_count"] = len(variables)
	result.Metadata["output_count"] = len(result.Output)

	return result, nil
}

// executeFunctionRule 执行函数规则
func (e *RuleExecutor) executeFunctionRule(ctx context.Context, rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	switch rule.Name {
	case "string_equals":
		return e.executeStringEquals(rule, variables, result)
	case "string_contains":
		return e.executeStringContains(rule, variables, result)
	case "number_compare":
		return e.executeNumberCompare(rule, variables, result)
	case "date_range":
		return e.executeDateRange(rule, variables, result)
	case "ip_range":
		return e.executeIPRange(rule, variables, result)
	case "list_contains":
		return e.executeListContains(rule, variables, result)
	case "regex_match":
		return e.executeRegexMatch(rule, variables, result)
	case "json_path":
		return e.executeJSONPath(rule, variables, result)
	case "custom_function":
		return e.executeCustomFunction(ctx, rule, variables, result)
	default:
		return fmt.Errorf("unknown function: %s", rule.Name)
	}
}

// executeStringEquals 执行字符串相等比较
func (e *RuleExecutor) executeStringEquals(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	// 获取参数
	left, leftExists := rule.Parameters["left"].(string)
	right, rightExists := rule.Parameters["right"].(string)
	caseSensitive, _ := rule.Parameters["case_sensitive"].(bool)
	if !caseSensitive {
		caseSensitive = true
	}

	if !leftExists || !rightExists {
		return fmt.Errorf("missing required parameters: left, right")
	}

	// 解析变量值
	leftValue := e.resolveVariable(left, variables)
	rightValue := e.resolveVariable(right, variables)

	leftStr, ok := leftValue.(string)
	if !ok {
		leftStr = fmt.Sprintf("%v", leftValue)
	}

	rightStr, ok := rightValue.(string)
	if !ok {
		rightStr = fmt.Sprintf("%v", rightValue)
	}

	// 执行比较
	var equal bool
	if caseSensitive {
		equal = leftStr == rightStr
	} else {
		equal = strings.ToLower(leftStr) == strings.ToLower(rightStr)
	}

	result.Output["result"] = equal
	result.Output["left"] = leftStr
	result.Output["right"] = rightStr
	result.Output["case_sensitive"] = caseSensitive

	return nil
}

// executeStringContains 执行字符串包含检查
func (e *RuleExecutor) executeStringContains(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	text, textExists := rule.Parameters["text"].(string)
	substring, substringExists := rule.Parameters["substring"].(string)
	caseSensitive, _ := rule.Parameters["case_sensitive"].(bool)
	if !caseSensitive {
		caseSensitive = true
	}

	if !textExists || !substringExists {
		return fmt.Errorf("missing required parameters: text, substring")
	}

	textValue := e.resolveVariable(text, variables)
	substringValue := e.resolveVariable(substring, variables)

	textStr, ok := textValue.(string)
	if !ok {
		textStr = fmt.Sprintf("%v", textValue)
	}

	subStr, ok := substringValue.(string)
	if !ok {
		subStr = fmt.Sprintf("%v", substringValue)
	}

	var contains bool
	if caseSensitive {
		contains = strings.Contains(textStr, subStr)
	} else {
		contains = strings.Contains(strings.ToLower(textStr), strings.ToLower(subStr))
	}

	result.Output["result"] = contains
	result.Output["text"] = textStr
	result.Output["substring"] = subStr
	result.Output["case_sensitive"] = caseSensitive

	return nil
}

// executeNumberCompare 执行数字比较
func (e *RuleExecutor) executeNumberCompare(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	left, leftExists := rule.Parameters["left"].(string)
	right, rightExists := rule.Parameters["right"].(string)
	operator, operatorExists := rule.Parameters["operator"].(string)
	if !operatorExists {
		operator = "=="
	}

	if !leftExists || !rightExists {
		return fmt.Errorf("missing required parameters: left, right")
	}

	leftValue := e.resolveVariable(left, variables)
	rightValue := e.resolveVariable(right, variables)

	leftNum, err := e.toFloat64(leftValue)
	if err != nil {
		return fmt.Errorf("invalid left value: %w", err)
	}

	rightNum, err := e.toFloat64(rightValue)
	if err != nil {
		return fmt.Errorf("invalid right value: %w", err)
	}

	var comparisonResult bool
	switch operator {
	case "==":
		comparisonResult = leftNum == rightNum
	case "!=":
		comparisonResult = leftNum != rightNum
	case ">":
		comparisonResult = leftNum > rightNum
	case ">=":
		comparisonResult = leftNum >= rightNum
	case "<":
		comparisonResult = leftNum < rightNum
	case "<=":
		comparisonResult = leftNum <= rightNum
	default:
		return fmt.Errorf("unknown operator: %s", operator)
	}

	result.Output["result"] = comparisonResult
	result.Output["left"] = leftNum
	result.Output["right"] = rightNum
	result.Output["operator"] = operator

	return nil
}

// executeDateRange 执行日期范围检查
func (e *RuleExecutor) executeDateRange(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	date, dateExists := rule.Parameters["date"].(string)
	startDate, startExists := rule.Parameters["start_date"].(string)
	endDate, endExists := rule.Parameters["end_date"].(string)
	format, _ := rule.Parameters["format"].(string)
	if format == "" {
		format = "2006-01-02T15:04:05Z"
	}

	if !dateExists {
		return fmt.Errorf("missing required parameter: date")
	}

	dateValue := e.resolveVariable(date, variables)
	dateStr, ok := dateValue.(string)
	if !ok {
		dateStr = fmt.Sprintf("%v", dateValue)
	}

	parsedDate, err := time.Parse(format, dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	var inRange bool
	if startExists && endExists {
		startValue := e.resolveVariable(startDate, variables)
		endValue := e.resolveVariable(endDate, variables)

		startStr, ok := startValue.(string)
		if !ok {
			startStr = fmt.Sprintf("%v", startValue)
		}

		endStr, ok := endValue.(string)
		if !ok {
			endStr = fmt.Sprintf("%v", endValue)
		}

		parsedStart, err := time.Parse(format, startStr)
		if err != nil {
			return fmt.Errorf("invalid start_date format: %w", err)
		}

		parsedEnd, err := time.Parse(format, endStr)
		if err != nil {
			return fmt.Errorf("invalid end_date format: %w", err)
		}

		inRange = parsedDate.After(parsedStart) && parsedDate.Before(parsedEnd)
	} else if startExists {
		startValue := e.resolveVariable(startDate, variables)
		startStr, ok := startValue.(string)
		if !ok {
			startStr = fmt.Sprintf("%v", startValue)
		}

		parsedStart, err := time.Parse(format, startStr)
		if err != nil {
			return fmt.Errorf("invalid start_date format: %w", err)
		}

		inRange = parsedDate.After(parsedStart)
	} else if endExists {
		endValue := e.resolveVariable(endDate, variables)
		endStr, ok := endValue.(string)
		if !ok {
			endStr = fmt.Sprintf("%v", endValue)
		}

		parsedEnd, err := time.Parse(format, endStr)
		if err != nil {
			return fmt.Errorf("invalid end_date format: %w", err)
		}

		inRange = parsedDate.Before(parsedEnd)
	} else {
		return fmt.Errorf("at least one of start_date or end_date must be provided")
	}

	result.Output["result"] = inRange
	result.Output["date"] = parsedDate
	result.Output["format"] = format

	if startExists {
		startValue := e.resolveVariable(startDate, variables)
		startStr, _ := startValue.(string)
		if parsedStart, err := time.Parse(format, startStr); err == nil {
			result.Output["start_date"] = parsedStart
		}
	}

	if endExists {
		endValue := e.resolveVariable(endDate, variables)
		endStr, _ := endValue.(string)
		if parsedEnd, err := time.Parse(format, endStr); err == nil {
			result.Output["end_date"] = parsedEnd
		}
	}

	return nil
}

// executeIPRange 执行IP范围检查
func (e *RuleExecutor) executeIPRange(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	ip, ipExists := rule.Parameters["ip"].(string)
	cidr, cidrExists := rule.Parameters["cidr"].(string)

	if !ipExists || !cidrExists {
		return fmt.Errorf("missing required parameters: ip, cidr")
	}

	ipValue := e.resolveVariable(ip, variables)
	cidrValue := e.resolveVariable(cidr, variables)

	ipStr, ok := ipValue.(string)
	if !ok {
		ipStr = fmt.Sprintf("%v", ipValue)
	}

	cidrStr, ok := cidrValue.(string)
	if !ok {
		cidrStr = fmt.Sprintf("%v", cidrValue)
	}

	// 简化实现，实际应该使用net包进行IP网络检查
	inRange := strings.Contains(cidrStr, ipStr) || ipStr == cidrStr

	result.Output["result"] = inRange
	result.Output["ip"] = ipStr
	result.Output["cidr"] = cidrStr

	return nil
}

// executeListContains 执行列表包含检查
func (e *RuleExecutor) executeListContains(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	list, listExists := rule.Parameters["list"].(string)
	item, itemExists := rule.Parameters["item"].(string)

	if !listExists || !itemExists {
		return fmt.Errorf("missing required parameters: list, item")
	}

	listValue := e.resolveVariable(list, variables)
	itemValue := e.resolveVariable(item, variables)

	listSlice, ok := listValue.([]interface{})
	if !ok {
		// 尝试解析JSON数组
		if listStr, ok := listValue.(string); ok {
			var parsedList []interface{}
			if err := json.Unmarshal([]byte(listStr), &parsedList); err == nil {
				listSlice = parsedList
			} else {
				return fmt.Errorf("invalid list format: %w", err)
			}
		} else {
			return fmt.Errorf("list must be an array or JSON string")
		}
	}

	itemStr := fmt.Sprintf("%v", itemValue)
	contains := false

	for _, listItem := range listSlice {
		if fmt.Sprintf("%v", listItem) == itemStr {
			contains = true
			break
		}
	}

	result.Output["result"] = contains
	result.Output["list"] = listSlice
	result.Output["item"] = itemStr

	return nil
}

// executeRegexMatch 执行正则表达式匹配
func (e *RuleExecutor) executeRegexMatch(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	text, textExists := rule.Parameters["text"].(string)
	pattern, patternExists := rule.Parameters["pattern"].(string)

	if !textExists || !patternExists {
		return fmt.Errorf("missing required parameters: text, pattern")
	}

	textValue := e.resolveVariable(text, variables)
	patternValue := e.resolveVariable(pattern, variables)

	textStr, ok := textValue.(string)
	if !ok {
		textStr = fmt.Sprintf("%v", textValue)
	}

	patternStr, ok := patternValue.(string)
	if !ok {
		patternStr = fmt.Sprintf("%v", patternValue)
	}

	// 简化实现，实际应该使用regexp包
	matched := strings.Contains(textStr, patternStr) || textStr == patternStr

	result.Output["result"] = matched
	result.Output["text"] = textStr
	result.Output["pattern"] = patternStr

	return nil
}

// executeJSONPath 执行JSON路径查询
func (e *RuleExecutor) executeJSONPath(rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	jsonData, jsonExists := rule.Parameters["json"].(string)
	path, pathExists := rule.Parameters["path"].(string)

	if !jsonExists || !pathExists {
		return fmt.Errorf("missing required parameters: json, path")
	}

	jsonValue := e.resolveVariable(jsonData, variables)
	pathValue := e.resolveVariable(path, variables)

	jsonStr, ok := jsonValue.(string)
	if !ok {
		jsonBytes, err := json.Marshal(jsonValue)
		if err != nil {
			return fmt.Errorf("invalid json data: %w", err)
		}
		jsonStr = string(jsonBytes)
	}

	pathStr, ok := pathValue.(string)
	if !ok {
		pathStr = fmt.Sprintf("%v", pathValue)
	}

	// 简化实现，实际应该使用gjson或类似库
	var jsonDataMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonDataMap); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	value, err := e.getJSONPathValue(jsonDataMap, pathStr)
	if err != nil {
		return fmt.Errorf("failed to get json path value: %w", err)
	}

	exists := value != nil

	result.Output["result"] = exists
	result.Output["value"] = value
	result.Output["path"] = pathStr

	return nil
}

// executeCustomFunction 执行自定义函数
func (e *RuleExecutor) executeCustomFunction(ctx context.Context, rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	// 简化实现，实际应该支持注册的自定义函数
	functionName, exists := rule.Parameters["function"].(string)
	if !exists {
		return fmt.Errorf("missing required parameter: function")
	}

	e.logger.WithFields(logrus.Fields{
		"function": functionName,
		"parameters": rule.Parameters,
	}).Warn("Custom function execution not implemented")

	result.Output["result"] = false
	result.Output["function"] = functionName
	result.Output["message"] = "Custom function execution not implemented"

	return nil
}

// executeScriptRule 执行脚本规则
func (e *RuleExecutor) executeScriptRule(ctx context.Context, rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	// 简化实现，实际应该集成脚本引擎
	e.logger.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"logic": rule.Logic,
	}).Warn("Script rule execution not implemented")

	result.Output["result"] = false
	result.Output["message"] = "Script rule execution not implemented"

	return nil
}

// executeAPIRule 执行API规则
func (e *RuleExecutor) executeAPIRule(ctx context.Context, rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	// 简化实现，实际应该实现HTTP客户端
	e.logger.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"definition": rule.Definition,
	}).Warn("API rule execution not implemented")

	result.Output["result"] = false
	result.Output["message"] = "API rule execution not implemented"

	return nil
}

// executeDecisionTableRule 执行决策表规则
func (e *RuleExecutor) executeDecisionTableRule(ctx context.Context, rule DynamicRule, variables map[string]interface{}, result *RuleExecutionResult) error {
	// 解析决策表定义
	definition, ok := rule.Definition.(string)
	if !ok {
		return fmt.Errorf("invalid decision table definition")
	}

	var decisionTable DecisionTable
	if err := json.Unmarshal([]byte(definition), &decisionTable); err != nil {
		return fmt.Errorf("failed to parse decision table: %w", err)
	}

	// 评估决策表
	for _, ruleRow := range decisionTable.Rules {
		if e.evaluateDecisionRule(ruleRow, variables) {
			result.Output["result"] = ruleRow.Output.Value
			result.Output["rule_matched"] = ruleRow.ID
			result.Output["description"] = ruleRow.Description
			return nil
		}
	}

	// 如果没有匹配的规则，使用默认输出
	if decisionTable.DefaultOutput != nil {
		result.Output["result"] = decisionTable.DefaultOutput.Value
		result.Output["rule_matched"] = "default"
		result.Output["description"] = "Default rule matched"
	} else {
		result.Output["result"] = false
		result.Output["rule_matched"] = "none"
		result.Output["description"] = "No rule matched"
	}

	return nil
}

// DecisionTable 决策表
type DecisionTable struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Inputs        []DecisionInput  `json:"inputs"`
	Outputs       []DecisionOutput `json:"outputs"`
	Rules         []DecisionRule   `json:"rules"`
	DefaultOutput *DecisionOutput  `json:"default_output"`
}

// DecisionInput 决策表输入
type DecisionInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// DecisionOutput 决策表输出
type DecisionOutput struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
}

// DecisionRule 决策规则
type DecisionRule struct {
	ID          string                    `json:"id"`
	Description string                    `json:"description"`
	Conditions  map[string]interface{}    `json:"conditions"`
	Output      DecisionOutput            `json:"output"`
	Priority    int                       `json:"priority"`
}

// evaluateDecisionRule 评估决策规则
func (e *RuleExecutor) evaluateDecisionRule(rule DecisionRule, variables map[string]interface{}) bool {
	for inputID, condition := range rule.Conditions {
		variableValue, exists := variables[inputID]
		if !exists {
			return false
		}

		// 简化的条件评估
		if conditionValue, ok := condition.(string); ok {
			if fmt.Sprintf("%v", variableValue) != conditionValue {
				return false
			}
		} else if conditionMap, ok := condition.(map[string]interface{}); ok {
			if operator, opExists := conditionMap["operator"].(string); opExists {
				if value, valExists := conditionMap["value"]; valExists {
					if !e.evaluateConditionOperation(operator, variableValue, value) {
						return false
					}
				}
			}
		}
	}

	return true
}

// evaluateConditionOperation 评估条件操作
func (e *RuleExecutor) evaluateConditionOperation(operator string, variableValue, conditionValue interface{}) bool {
	switch operator {
	case "equals", "==":
		return fmt.Sprintf("%v", variableValue) == fmt.Sprintf("%v", conditionValue)
	case "not_equals", "!=":
		return fmt.Sprintf("%v", variableValue) != fmt.Sprintf("%v", conditionValue)
	case "greater_than", ">":
		return e.compareGreaterThan(variableValue, conditionValue)
	case "less_than", "<":
		return e.compareLessThan(variableValue, conditionValue)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", variableValue), fmt.Sprintf("%v", conditionValue))
	default:
		return false
	}
}

// 辅助方法
func (e *RuleExecutor) resolveVariable(varName string, variables map[string]interface{}) interface{} {
	if strings.HasPrefix(varName, "$") {
		actualVarName := varName[1:]
		if value, exists := variables[actualVarName]; exists {
			return value
		}
	}
	return varName
}

func (e *RuleExecutor) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (e *RuleExecutor) getJSONPathValue(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(strings.Trim(path, "."), ".")
	current := data

	for i, part := range parts {
		if next, exists := current[part]; exists {
			if i == len(parts)-1 {
				return next, nil
			}
			if nextMap, ok := next.(map[string]interface{}); ok {
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

func (e *RuleExecutor) compareGreaterThan(a, b interface{}) bool {
	aFloat, err := e.toFloat64(a)
	if err != nil {
		return false
	}
	bFloat, err := e.toFloat64(b)
	if err != nil {
		return false
	}
	return aFloat > bFloat
}

func (e *RuleExecutor) compareLessThan(a, b interface{}) bool {
	aFloat, err := e.toFloat64(a)
	if err != nil {
		return false
	}
	bFloat, err := e.toFloat64(b)
	if err != nil {
		return false
	}
	return aFloat < bFloat
}