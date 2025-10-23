package main

import (
	"fmt"
	"reflect"
	"strings"
)

// 模拟前端parseOtherParties函数的逻辑
func parseOtherParties(opponentInfo string) []string {
	if opponentInfo == "" {
		return []string{}
	}

	// 先按换行符分割
	lines := strings.Split(opponentInfo, "\n")

	// 提取可能的当事人名称
	var partyNames []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// 如果是数字开头的行（如"1. "），提取后面的内容
		if strings.HasPrefix(trimmedLine, "1.") || strings.HasPrefix(trimmedLine, "2.") {
			cleanLine := strings.TrimPrefix(trimmedLine, "1.")
			cleanLine = strings.TrimPrefix(cleanLine, "2.")
			cleanLine = strings.TrimSpace(cleanLine)
			partyNames = append(partyNames, cleanLine)
		} else if isLikelyPartyName(trimmedLine) {
			partyNames = append(partyNames, trimmedLine)
		}
	}

	// 如果没有找到任何合理的名称，则使用原始分割方法
	if len(partyNames) == 0 {
		separators := []string{",", ";", "、", "，", "；", "|"}
		var parties []string

		// 尝试各种分隔符进行分割
		for _, separator := range separators {
			if strings.Contains(opponentInfo, separator) {
				parties = strings.Split(opponentInfo, separator)
				break
			}
		}

		// 清理和去重
		uniqueParties := make(map[string]bool)
		var result []string
		for _, party := range parties {
			trimmed := strings.TrimSpace(party)
			if trimmed != "" && !uniqueParties[trimmed] {
				uniqueParties[trimmed] = true
				result = append(result, trimmed)
			}
		}
		return result
	}

	// 清理和去重
	uniqueParties := make(map[string]bool)
	var result []string
	for _, party := range partyNames {
		// 移除常见的无关后缀
		cleaned := strings.TrimSuffix(party, "纠纷")
		cleaned = strings.TrimSuffix(cleaned, "案件")
		cleaned = strings.TrimSpace(cleaned)

		if cleaned != "" && !uniqueParties[cleaned] {
			uniqueParties[cleaned] = true
			result = append(result, cleaned)
		}
	}

	return result
}

func isLikelyPartyName(text string) bool {
	if text == "" || len(text) < 2 {
		return false
	}

	// 包含常见的企业标识符
	companySuffixes := []string{
		"有限公司", "股份有限公司", "集团", "公司", "企业", "实业",
		"科技", "投资", "控股", "发展", "建设", "工程", "贸易",
		"Co., Ltd", "Ltd.", "Inc.", "LLC", "Corp.", "Group",
	}

	hasCompanySuffix := false
	for _, suffix := range companySuffixes {
		if strings.Contains(text, suffix) || strings.Contains(strings.ToLower(text), strings.ToLower(suffix)) {
			hasCompanySuffix = true
			break
		}
	}

	// 如果文本太长，不太可能是纯名称
	if len(text) > 50 {
		return false
	}

	// 如果包含数字编号，很可能是标签而不是名称
	if strings.HasPrefix(text, "1.") || strings.HasPrefix(text, "2.") {
		return false
	}

	// 如果包含常见的描述性词汇，不太可能是纯名称
	descriptionWords := []string{
		"纠纷", "案件", "诉讼", "仲裁", "争议", "地址", "联系", "电话",
		"邮箱", "传真", "邮编", "网址", "说明", "备注",
	}

	hasDescriptionWord := false
	for _, word := range descriptionWords {
		if strings.Contains(text, word) {
			hasDescriptionWord = true
			break
		}
	}

	return hasCompanySuffix || (!hasDescriptionWord && len(text) <= 20)
}

func main() {
	fmt.Println("🧪 测试parseOtherParties函数")
	fmt.Println("=================================")

	// 测试用例
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "简单的公司名称",
			input:    "腾讯",
			expected: []string{"腾讯"},
		},
		{
			name:     "带换行的复杂信息",
			input:    "腾讯\n垄断纠纷案",
			expected: []string{"腾讯"},
		},
		{
			name:     "数字编号格式",
			input:    "1. 腾讯科技有限公司\n2. 深圳市腾讯计算机系统有限公司",
			expected: []string{"腾讯科技有限公司", "深圳市腾讯计算机系统有限公司"},
		},
		{
			name:     "多个分隔符",
			input:    "腾讯、阿里巴巴、字节跳动",
			expected: []string{"腾讯", "阿里巴巴", "字节跳动"},
		},
		{
			name:     "复杂的对方信息",
			input:    "腾讯\n地址:深圳市南山区\n电话:0755-86013388\n垄断纠纷案",
			expected: []string{"腾讯"},
		},
		{
			name:     "实际问题场景",
			input:    "腾讯\n垄断纠纷案",
			expected: []string{"腾讯"},
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n📋 测试用例 %d: %s\n", i+1, tc.name)
		fmt.Printf("输入: %q\n", tc.input)

		result := parseOtherParties(tc.input)
		fmt.Printf("结果: %v\n", result)
		fmt.Printf("期望: %v\n", tc.expected)

		if reflect.DeepEqual(result, tc.expected) {
			fmt.Printf("✅ 测试通过\n")
		} else {
			fmt.Printf("❌ 测试失败\n")
		}
	}

	fmt.Println("\n=================================")
	fmt.Println("测试完成")
}