package mock

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertionHelper 测试断言助手
type AssertionHelper struct {
	t *testing.T
}

// NewAssertionHelper 创建测试断言助手
func NewAssertionHelper(t *testing.T) *AssertionHelper {
	return &AssertionHelper{t: t}
}

// AssertUser 断言用户数据
func (ah *AssertionHelper) AssertUser(expected, actual interface{}, msg ...string) {
	expectedMap, ok1 := expected.(map[string]interface{})
	actualMap, ok2 := actual.(map[string]interface{})
	
	require.True(ah.t, ok1, "Expected must be a map[string]interface{}")
	require.True(ah.t, ok2, "Actual must be a map[string]interface{}")
	
	message := "User data mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	// 检查必填字段
	assert.Equal(ah.t, expectedMap["name"], actualMap["name"], message+" - name")
	assert.Equal(ah.t, expectedMap["email"], actualMap["email"], message+" - email")
	assert.Equal(ah.t, expectedMap["role"], actualMap["role"], message+" - role")
	assert.Equal(ah.t, expectedMap["status"], actualMap["status"], message+" - status")
	
	// 检查时间字段（如果存在）
	if _, ok1 := expectedMap["created_at"]; ok1 {
		if actualCreatedAt, ok2 := actualMap["created_at"]; ok2 {
			assert.NotNil(ah.t, actualCreatedAt, message+" - created_at should not be nil")
		}
	}
}

// AssertCase 断言案件数据
func (ah *AssertionHelper) AssertCase(expected, actual interface{}, msg ...string) {
	expectedMap, ok1 := expected.(map[string]interface{})
	actualMap, ok2 := actual.(map[string]interface{})
	
	require.True(ah.t, ok1, "Expected must be a map[string]interface{}")
	require.True(ah.t, ok2, "Actual must be a map[string]interface{}")
	
	message := "Case data mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	// 检查必填字段
	assert.Equal(ah.t, expectedMap["title"], actualMap["title"], message+" - title")
	assert.Equal(ah.t, expectedMap["description"], actualMap["description"], message+" - description")
	assert.Equal(ah.t, expectedMap["client_id"], actualMap["client_id"], message+" - client_id")
	assert.Equal(ah.t, expectedMap["lawyer_id"], actualMap["lawyer_id"], message+" - lawyer_id")
	assert.Equal(ah.t, expectedMap["case_type"], actualMap["case_type"], message+" - case_type")
	assert.Equal(ah.t, expectedMap["priority"], actualMap["priority"], message+" - priority")
	assert.Equal(ah.t, expectedMap["status"], actualMap["status"], message+" - status")
}

// AssertClient 断言客户数据
func (ah *AssertionHelper) AssertClient(expected, actual interface{}, msg ...string) {
	expectedMap, ok1 := expected.(map[string]interface{})
	actualMap, ok2 := actual.(map[string]interface{})
	
	require.True(ah.t, ok1, "Expected must be a map[string]interface{}")
	require.True(ah.t, ok2, "Actual must be a map[string]interface{}")
	
	message := "Client data mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	// 检查必填字段
	assert.Equal(ah.t, expectedMap["name"], actualMap["name"], message+" - name")
	assert.Equal(ah.t, expectedMap["contact"], actualMap["contact"], message+" - contact")
	assert.Equal(ah.t, expectedMap["phone"], actualMap["phone"], message+" - phone")
	assert.Equal(ah.t, expectedMap["email"], actualMap["email"], message+" - email")
	assert.Equal(ah.t, expectedMap["address"], actualMap["address"], message+" - address")
	assert.Equal(ah.t, expectedMap["type"], actualMap["type"], message+" - type")
}

// AssertError 断言错误
func (ah *AssertionHelper) AssertError(expectedError string, actual error, msg ...string) {
	message := "Error mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	if expectedError == "" {
		assert.NoError(ah.t, actual, message+" - expected no error")
	} else {
		assert.Error(ah.t, actual, message+" - expected error")
		assert.Contains(ah.t, actual.Error(), expectedError, message+" - error message mismatch")
	}
}

// AssertSuccess 断言成功响应
func (ah *AssertionHelper) AssertSuccess(actual interface{}, msg ...string) {
	message := "Expected success response"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.NotNil(ah.t, actual, message+" - response should not be nil")
	
	// 检查是否是map类型
	if actualMap, ok := actual.(map[string]interface{}); ok {
		assert.NotContains(ah.t, actualMap, "error", message+" - should not contain error field")
		if code, exists := actualMap["code"]; exists {
			assert.Equal(ah.t, float64(200), code, message+" - should have success code")
		}
	}
}

// AssertHTTPResponse 断言HTTP响应
func (ah *AssertionHelper) AssertHTTPResponse(expectedStatusCode int, actual interface{}, msg ...string) {
	message := "HTTP response mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	actualMap, ok := actual.(map[string]interface{})
	require.True(ah.t, ok, "HTTP response must be a map[string]interface{}")
	
	if code, exists := actualMap["code"]; exists {
		assert.Equal(ah.t, float64(expectedStatusCode), code, message+" - status code mismatch")
	}
}

// AssertTimeRecent 断言时间是最近的
func (ah *AssertionHelper) AssertTimeRecent(actualTime interface{}, maxDuration time.Duration, msg ...string) {
	message := "Time is not recent"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	var t time.Time
	switch v := actualTime.(type) {
	case time.Time:
		t = v
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		require.NoError(ah.t, err, message+" - time parsing failed")
		t = parsed
	default:
		ah.t.Fatal(message+" - unsupported time type")
	}
	
	now := time.Now()
	duration := now.Sub(t)
	if duration < 0 {
		duration = -duration
	}
	
	assert.True(ah.t, duration <= maxDuration, 
		fmt.Sprintf("%s - time difference %v exceeds max duration %v", message, duration, maxDuration))
}

// AssertMapContains 断言map包含指定键值对
func (ah *AssertionHelper) AssertMapContains(expected, actual map[string]interface{}, msg ...string) {
	message := "Map does not contain expected key-value pairs"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		assert.True(ah.t, exists, fmt.Sprintf("%s - key '%s' not found", message, key))
		assert.Equal(ah.t, expectedValue, actualValue, fmt.Sprintf("%s - value mismatch for key '%s'", message, key))
	}
}

// AssertJSONEqual 断言JSON相等
func (ah *AssertionHelper) AssertJSONEqual(expected, actual interface{}, msg ...string) {
	message := "JSON mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	expectedJSON, err := json.Marshal(expected)
	require.NoError(ah.t, err, message+" - failed to marshal expected")
	
	actualJSON, err := json.Marshal(actual)
	require.NoError(ah.t, err, message+" - failed to marshal actual")
	
	assert.JSONEq(ah.t, string(expectedJSON), string(actualJSON), message)
}

// AssertSliceLength 断言切片长度
func (ah *AssertionHelper) AssertSliceLength(expectedLength int, actual interface{}, msg ...string) {
	message := "Slice length mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	val := reflect.ValueOf(actual)
	require.True(ah.t, val.Kind() == reflect.Slice || val.Kind() == reflect.Array, 
		message+" - actual must be a slice or array")
	
	assert.Equal(ah.t, expectedLength, val.Len(), message+" - length mismatch")
}

// AssertMapHasKey 断言map包含指定键
func (ah *AssertionHelper) AssertMapHasKey(key string, actual interface{}, msg ...string) {
	message := fmt.Sprintf("Map does not contain key '%s'", key)
	if len(msg) > 0 {
		message = msg[0]
	}
	
	actualMap, ok := actual.(map[string]interface{})
	require.True(ah.t, ok, message+" - actual must be a map[string]interface{}")
	
	_, exists := actualMap[key]
	assert.True(ah.t, exists, message)
}

// AssertMapNotHasKey 断言map不包含指定键
func (ah *AssertionHelper) AssertMapNotHasKey(key string, actual interface{}, msg ...string) {
	message := fmt.Sprintf("Map should not contain key '%s'", key)
	if len(msg) > 0 {
		message = msg[0]
	}
	
	actualMap, ok := actual.(map[string]interface{})
	require.True(ah.t, ok, message+" - actual must be a map[string]interface{}")
	
	_, exists := actualMap[key]
	assert.False(ah.t, exists, message)
}

// AssertType 断言类型匹配
func (ah *AssertionHelper) AssertType(expectedType reflect.Type, actual interface{}, msg ...string) {
	message := "Type mismatch"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	actualType := reflect.TypeOf(actual)
	assert.Equal(ah.t, expectedType, actualType, 
		fmt.Sprintf("%s - expected %v, got %v", message, expectedType, actualType))
}

// AssertNotNil 断言值不为nil
func (ah *AssertionHelper) AssertNotNil(actual interface{}, msg ...string) {
	message := "Value should not be nil"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.NotNil(ah.t, actual, message)
}

// AssertNil 断言值为nil
func (ah *AssertionHelper) AssertNil(actual interface{}, msg ...string) {
	message := "Value should be nil"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.Nil(ah.t, actual, message)
}

// AssertNotEmpty 断言字符串不为空
func (ah *AssertionHelper) AssertNotEmpty(actual string, msg ...string) {
	message := "String should not be empty"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.NotEmpty(ah.t, actual, message)
}

// AssertEmpty 断言字符串为空
func (ah *AssertionHelper) AssertEmpty(actual string, msg ...string) {
	message := "String should be empty"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.Empty(ah.t, actual, message)
}

// AssertPositive 断言数值为正
func (ah *AssertionHelper) AssertPositive(actual interface{}, msg ...string) {
	message := "Value should be positive"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.Positive(ah.t, actual, message)
}

// AssertNegative 断言数值为负
func (ah *AssertionHelper) AssertNegative(actual interface{}, msg ...string) {
	message := "Value should be negative"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.Negative(ah.t, actual, message)
}

// AssertZero 断言数值为零
func (ah *AssertionHelper) AssertZero(actual interface{}, msg ...string) {
	message := "Value should be zero"
	if len(msg) > 0 {
		message = msg[0]
	}
	
	assert.Zero(ah.t, actual, message)
}

// MockExpectationHelper Mock期望助手
type MockExpectationHelper struct {
	t *testing.T
}

// NewMockExpectationHelper 创建Mock期望助手
func NewMockExpectationHelper(t *testing.T) *MockExpectationHelper {
	return &MockExpectationHelper{t: t}
}

// SetupMockDBExpectations 设置数据库Mock期望
func (meh *MockExpectationHelper) SetupMockDBExpectations(mockDB *MockDB, expectations []DBExpectation) {
	for _, expectation := range expectations {
		switch expectation.Operation {
		case "select":
			mockDB.Mock.ExpectQuery(expectation.Query).
				WithArgs(toDriverValues(expectation.Args)...).
				WillReturnRows(expectation.Rows)
		case "insert":
			mockDB.Mock.ExpectBegin()
			mockDB.Mock.ExpectExec(expectation.Query).
				WithArgs(toDriverValues(expectation.Args)...).
				WillReturnResult(expectation.Result)
			mockDB.Mock.ExpectCommit()
		case "update":
			mockDB.Mock.ExpectBegin()
			mockDB.Mock.ExpectExec(expectation.Query).
				WithArgs(toDriverValues(expectation.Args)...).
				WillReturnResult(expectation.Result)
			mockDB.Mock.ExpectCommit()
		case "delete":
			mockDB.Mock.ExpectBegin()
			mockDB.Mock.ExpectExec(expectation.Query).
				WithArgs(toDriverValues(expectation.Args)...).
				WillReturnResult(expectation.Result)
			mockDB.Mock.ExpectCommit()
		}
	}
}

// toDriverValues 将interface{}切片转换为driver.Value切片
func toDriverValues(args []interface{}) []driver.Value {
	result := make([]driver.Value, len(args))
	for i, arg := range args {
		result[i] = arg
	}
	return result
}

// DBExpectation 数据库期望配置
type DBExpectation struct {
	Operation string
	Query     string
	Args      []interface{}
	Rows      *sqlmock.Rows
	Result    driver.Result
}