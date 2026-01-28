package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

func main() {
	// 使用张伟律师的凭据尝试登录
	loginReq := LoginRequest{
		Email:    "zhangwei@law.com",
		Password: "123456",
	}

	jsonData, _ := json.Marshal(loginReq)

	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	if loginResp.Success {
		fmt.Printf("Token: %s\n", loginResp.Data.Token)
	} else {
		fmt.Printf("Login failed: %s\n", string(body))
	}
}