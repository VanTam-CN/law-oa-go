package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 生成密码哈希
	password := "123456"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}

	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", string(hash))

	// 验证密码
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Printf("Password verification failed: %v\n", err)
	} else {
		fmt.Printf("Password verification successful\n")
	}
}