package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// 从命令行参数获取token
	if len(os.Args) < 2 {
		log.Fatal("请提供JWT token作为参数")
	}
	tokenString := os.Args[1]

	// JWT密钥 (从环境配置文件获取)
	secret := "your-very-secure-jwt-secret-key-for-development-only"

	// 解析token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		log.Printf("Token解析失败: %v", err)
		return
	}

	// 提取claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Printf("Token验证成功!\n")
		fmt.Printf("User ID: %v\n", claims["user_id"])
		fmt.Printf("Username: %v\n", claims["username"])
		fmt.Printf("Role: %v\n", claims["role"])
		fmt.Printf("Issued At: %v\n", claims["iat"])
		fmt.Printf("Expires At: %v\n", claims["exp"])
	} else {
		fmt.Println("Token无效")
	}
}