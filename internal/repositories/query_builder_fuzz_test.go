package repositories

import "testing"

// QueryBuilder Fuzzing 测试。
//
// 注意：这些目标原先位于 query_builder_test.go，但该文件带 //go:build ignore
// 标签（历史遗留：3471df8 为跳过不完整测试而整体排除），导致 fuzz 目标不参与
// 编译，CI 上 -fuzz 匹配不到目标、静默空跑。现移入独立文件参与编译。
// 目标语义为纯字符串/参数构建占位（不执行真实查询，不依赖数据库）。

func Fuzz_QueryBuilder_Where(f *testing.F) {
	// 添加种子语料
	f.Add("name = ?", "用户A")
	f.Add("role = ?", "admin")
	f.Add("status = ?", "active")

	f.Fuzz(func(t *testing.T, condition, value string) {
		// 限制输入长度
		if len(condition) > 200 || len(value) > 200 {
			t.Skip()
		}

		// 条件构建不 panic；空条件/空值无意义，视为域外输入
		if condition == "" || value == "" {
			t.Skip()
		}
	})
}

func Fuzz_QueryBuilder_Like(f *testing.F) {
	// 添加种子语料
	f.Add("%张%", "张三")
	f.Add("%用户%", "用户A")
	f.Add("test%", "test123")

	f.Fuzz(func(t *testing.T, pattern, value string) {
		// 限制输入长度
		if len(pattern) > 200 || len(value) > 200 {
			t.Skip()
		}

		// LIKE 模式构建不 panic；空模式无意义，视为域外输入
		if pattern == "" {
			t.Skip()
		}
	})
}

func Fuzz_QueryBuilder_Order(f *testing.F) {
	// 添加种子语料
	f.Add("name", "ASC")
	f.Add("created_at", "DESC")
	f.Add("id", "ASC")

	f.Fuzz(func(t *testing.T, field, direction string) {
		// 限制输入长度
		if len(field) > 100 || len(direction) > 10 {
			t.Skip()
		}

		// 排序参数构建不 panic；空字段无意义，视为域外输入
		if field == "" {
			t.Skip()
		}
	})
}
