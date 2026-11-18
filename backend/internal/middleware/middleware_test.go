//go:build ignore
// +build ignore

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"law-oa-go/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建路由
	router := gin.New()
	router.Use(middleware.AuthMiddleware())

	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试未认证的请求
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
