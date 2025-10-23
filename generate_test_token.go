package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// 生成测试token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 7,
		"email":   "liming@jinchenglaw.com",
		"role":    "lawyer",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	})

	// 使用.env文件中的密钥签名
	secret := "your-very-secure-jwt-secret-key-for-development-only"
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	fmt.Printf("测试Token: %s\n", tokenString)
}