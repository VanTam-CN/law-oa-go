//go:build test_compile

package services

import (
	"context"
	"strconv"
)

// 仅用于编译测试，验证类型和方法匹配
func testConflictServiceCompilation() {
	var ctx context.Context

	// 验证string转uint的转换
	clientIDStr := "123"
	clientIDUint, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		return
	}
	_ = uint(clientIDUint)
	_ = ctx
}
