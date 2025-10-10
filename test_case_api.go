package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT密钥 - 从配置中获取
const jwtSecret = "your-very-secure-jwt-secret-key-that-is-at-least-32-characters-long"

func generateToken(userID uint, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func main() {
	fmt.Println("=== 生成测试JWT Token ===")

	// 使用admin用户信息生成token
	token, err := generateToken(1, "admin@law-oa.com", "admin")
	if err != nil {
		fmt.Printf("生成token失败: %v\n", err)
		return
	}

	fmt.Printf("Generated Token: %s\n", token)

	// 保存token到文件
	fmt.Printf("\n测试命令:\n")
	fmt.Printf("curl -X GET \"http://localhost:8080/api/cases\" -H \"Authorization: Bearer %s\" -H \"Content-Type: application/json\" -s | jq '.'\n", token)
	fmt.Printf("\n新建案件测试:\n")
	fmt.Printf(`curl -X POST "http://localhost:8080/api/cases" -H "Authorization: Bearer %s" -H "Content-Type: application/json" -d '{"title":"测试案件","description":"API测试创建的案件","client_id":1,"lawyer_id":2,"case_type":"civil","priority":"medium"}' -s | jq '.'`+"\n", token)
}