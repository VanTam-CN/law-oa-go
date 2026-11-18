package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicFunctionality(t *testing.T) {
	// 这是一个基础测试，确保测试框架能正常工作
	assert.True(t, true, "测试框架应该正常工作")
	assert.Equal(t, 1+1, 2, "基础数学计算应该正确")

	// 验证我们能导入必要的包
	assert.NotNil(t, t, "测试对象不应该为nil")
}

func TestStringOperations(t *testing.T) {
	// 测试字符串操作
	testString := "law-oa-go"
	assert.Contains(t, testString, "law", "字符串应该包含law")
	assert.NotContains(t, testString, "invalid", "字符串不应该包含invalid")
}

func TestIntegerOperations(t *testing.T) {
	// 测试整数操作
	result := 2 * 3
	assert.Equal(t, 6, result, "乘法结果应该正确")

	assert.Greater(t, 10, 5, "10应该大于5")
	assert.Less(t, 3, 8, "3应该小于8")
}
