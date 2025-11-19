package main

import (
	"fmt"

	"law-oa-go/internal/middleware"
)

func main() {
	// 测试token解析
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJ1c2VybmFtZSI6ImFkbWluQGxhd29hLmNvbSIsInJvbGUiOiJhZG1pbiIsImV4cCI6MTczMzYxNDEzOSwiaWF0IjoxNzMzNTI3NzM5fQ.invalid"

	fmt.Println("🔍 调试JWT Token解析...")

	// 解析token
	claims, err := middleware.ParseToken(testToken)
	if err != nil {
		fmt.Printf("❌ Token解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ Token解析成功！\n")
	fmt.Printf("   用户ID: %d (类型: %T)\n", claims.UserID, claims.UserID)
	fmt.Printf("   用户名: %s\n", claims.Username)
	fmt.Printf("   角色: %s\n", claims.Role)

	// 检查类型转换
	fmt.Println("\n🔧 测试类型转换...")

	// 模拟审批处理器中的类型转换
	userID := interface{}(claims.UserID)
	fmt.Printf("   interface{}类型: %T\n", userID)

	switch v := userID.(type) {
	case uint:
		fmt.Printf("   ✅ uint类型转换成功: %d\n", v)
	case int:
		fmt.Printf("   ✅ int类型转换成功: %d\n", v)
	case string:
		fmt.Printf("   ✅ string类型转换成功: %s\n", v)
	default:
		fmt.Printf("   ❌ 未知类型: %T\n", v)
	}

	// 测试GORM的用户ID类型
	fmt.Println("\n🗄️ 测试数据库用户ID类型...")
	testGormTypes()
}

func testGormTypes() {
	// 模拟从数据库获取的用户结构体
	type User struct {
		ID       uint   `gorm:"primaryKey"`
		Username string
		Email    string
		Role     string
	}

	user := User{
		ID:       3,
		Username: "admin@lawoa.com",
		Email:    "admin@lawoa.com",
		Role:     "admin",
	}

	fmt.Printf("   数据库用户ID: %d (类型: %T)\n", user.ID, user.ID)

	// 测试生成JWT
	token, _, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		fmt.Printf("   ❌ 生成JWT失败: %v\n", err)
		return
	}

	fmt.Printf("   ✅ 生成JWT成功: %s...\n", token[:50])

	// 解析刚生成的token
	claims, err := middleware.ParseToken(token)
	if err != nil {
		fmt.Printf("   ❌ 解析生成的JWT失败: %v\n", err)
		return
	}

	fmt.Printf("   解析后用户ID: %d (类型: %T)\n", claims.UserID, claims.UserID)
}