package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// TestSuite 测试套件基类
type TestSuite struct {
	suite.Suite
}

// SetupTest 每个测试前的设置
func (s *TestSuite) SetupTest() {
	// 在这里添加测试前的设置
}

// TearDownTest 每个测试后的清理
func (s *TestSuite) TearDownTest() {
	// 在这里添加测试后的清理
}

// BasicTestSuite 基础测试套件
type BasicTestSuite struct {
	TestSuite
}

// TestBasicFunctionality 测试基础功能
func (s *BasicTestSuite) TestBasicFunctionality() {
	s.Equal(1+1, 2, "基础数学计算应该正确")
	s.True(true, "真值应该为真")
	s.NotNil("test", "字符串不应该为空")
}

// TestStringOperations 测试字符串操作
func (s *BasicTestSuite) TestStringOperations() {
	testString := "law-oa-go"
	s.Contains(testString, "law", "字符串应该包含law")
	s.NotContains(testString, "invalid", "字符串不应该包含invalid")
	s.Equal(len("law-oa-go"), 9, "字符串长度应该正确")
}

// TestIntegerOperations 测试整数操作
func (s *BasicTestSuite) TestIntegerOperations() {
	result := 2 * 3
	s.Equal(6, result, "乘法结果应该正确")

	s.Greater(10, 5, "10应该大于5")
	s.Less(3, 8, "3应该小于8")

	s.GreaterOrEqual(5, 5, "5应该大于等于5")
	s.LessOrEqual(5, 5, "5应该小于等于5")
}

// TestArrayOperations 测试数组操作
func (s *BasicTestSuite) TestArrayOperations() {
	array := []int{1, 2, 3, 4, 5}

	s.Equal(5, len(array), "数组长度应该正确")
	s.Equal(1, array[0], "第一个元素应该正确")
	s.Equal(5, array[len(array)-1], "最后一个元素应该正确")

	// 测试切片
	slice := array[1:3]
	s.Equal(2, len(slice), "切片长度应该正确")
	s.Equal(2, slice[0], "切片第一个元素应该正确")
}

// TestMapOperations 测试Map操作
func (s *BasicTestSuite) TestMapOperations() {
	testMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	s.Equal(3, len(testMap), "Map长度应该正确")
	s.Equal(1, testMap["one"], "Map值应该正确")
	s.Equal(2, testMap["two"], "Map值应该正确")

	// 测试键是否存在
	value, exists := testMap["three"]
	s.True(exists, "键应该存在")
	s.Equal(3, value, "值应该正确")

	// 测试不存在的键
	_, exists = testMap["four"]
	s.False(exists, "键不应该存在")
}

// TestErrorHandling 测试错误处理
func (s *BasicTestSuite) TestErrorHandling() {
	// 测试nil错误
	err := error(nil)
	s.NoError(err, "nil错误应该通过NoError检查")

	// 测试非nil错误
	err = fmt.Errorf("test error")
	s.Error(err, "非nil错误应该通过Error检查")
	s.Equal("test error", err.Error(), "错误消息应该正确")
}

// TestPanicHandling 测试Panic处理
func (s *BasicTestSuite) TestPanicHandling() {
	s.NotPanics(func() {
		// 不应该panic的代码
		result := 1 + 1
		s.Equal(2, result)
	}, "正常代码不应该panic")

	s.Panics(func() {
		// 应该panic的代码
		panic("intentional panic")
	}, "故意触发的panic应该被捕获")
}

// TestTimeOperations 测试时间操作
func (s *BasicTestSuite) TestTimeOperations() {
	now := time.Now()

	s.NotZero(now, "当前时间不应该为零")
	s.True(now.After(time.Time{}), "当前时间应该大于零时间")

	// 测试时间计算
	later := now.Add(1 * time.Hour)
	s.True(later.After(now), "加一小时后的时间应该大于现在")
}

// TestInterfaceOperations 测试接口操作
func (s *BasicTestSuite) TestInterfaceOperations() {
	// 测试空接口
	var i interface{}
	s.Nil(i, "空接口应该为nil")

	// 测试类型断言
	i = "test string"
	str, ok := i.(string)
	s.True(ok, "类型断言应该成功")
	s.Equal("test string", str, "断言结果应该正确")

	// 测试错误的类型断言
	_, ok = i.(int)
	s.False(ok, "错误的类型断言应该失败")
}

// TestBasicTestSuite 运行基础测试套件
func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
