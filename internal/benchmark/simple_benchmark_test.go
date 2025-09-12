package benchmark

import (
	"testing"
)

// BenchmarkSimple 基准测试示例
func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// 简单的基准测试
		_ = i * 2
	}
}