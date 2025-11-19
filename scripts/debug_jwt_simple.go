package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func main() {
	// 从测试脚本中获取的真实token
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJ1c2VybmFtZSI6ImFkbWluQGxhd29hLmNvbSIsInJvbGUiOiJhZG1pbiIsImV4cCI6MTczMzYxNDEzOSwiaWF0IjoxNzMzNTI3NzM5fQ.invalid"

	fmt.Println("🔍 手动解析JWT Token...")

	// 分割token
	parts := strings.Split(testToken, ".")
	if len(parts) != 3 {
		fmt.Printf("❌ 无效的JWT格式\n")
		return
	}

	// 解码payload部分
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Printf("❌ 解码payload失败: %v\n", err)
		return
	}

	fmt.Printf("✅ Payload原始数据: %s\n", string(payload))

	// 解析JSON
	var claims JWTClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return
	}

	fmt.Printf("✅ JWT Claims解析成功！\n")
	fmt.Printf("   用户ID: %d (类型: %T)\n", claims.UserID, claims.UserID)
	fmt.Printf("   用户名: %s\n", claims.Username)
	fmt.Printf("   角色: %s\n", claims.Role)

	// 测试类型转换
	fmt.Println("\n🔧 测试类型转换...")

	// 模拟审批处理器中的类型转换
	userID := interface{}(claims.UserID)
	fmt.Printf("   interface{}类型: %T\n", userID)

	switch v := userID.(type) {
	case uint:
		fmt.Printf("   ✅ uint类型转换成功: %d\n", v)
		// 转换为字符串
		userIDStr := fmt.Sprintf("%d", v)
		fmt.Printf("   ✅ 转换为字符串: %s\n", userIDStr)
	case int:
		fmt.Printf("   ✅ int类型转换成功: %d\n", v)
	case string:
		fmt.Printf("   ✅ string类型转换成功: %s\n", v)
	default:
		fmt.Printf("   ❌ 未知类型: %T\n", v)
	}

	fmt.Println("\n🎯 结论:")
	fmt.Println("JWT token中的用户ID是uint类型，在Go中应该能够正确处理。")
	fmt.Println("问题可能不在类型转换，而是在其他地方。")
}