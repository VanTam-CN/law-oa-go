package test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "law-oa-go/docs"
)

func TestSwaggerDocumentation(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	r := gin.Default()

	// Add Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Test Swagger JSON endpoint
	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response body
	body, _ := io.ReadAll(w.Body)
	var swaggerDoc map[string]interface{}
	err := json.Unmarshal(body, &swaggerDoc)
	assert.NoError(t, err)
	assert.Equal(t, "示例律师事务所OA API", swaggerDoc["info"].(map[string]interface{})["title"])
	assert.Equal(t, "1.0", swaggerDoc["info"].(map[string]interface{})["version"])
	assert.Equal(t, "/api/v1", swaggerDoc["basePath"])
}

func TestSwaggerUI(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	r := gin.Default()

	// Add Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Test Swagger UI endpoint
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that it's HTML content
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
}

func TestAPIEndpointsDocumented(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	r := gin.Default()

	// Add Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Load the swagger.json file
	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	var swaggerDoc map[string]interface{}
	json.Unmarshal(body, &swaggerDoc)

	paths := swaggerDoc["paths"].(map[string]interface{})

	// Check that key endpoints are documented
	assert.Contains(t, paths, "/auth/login")
	assert.Contains(t, paths, "/auth/register")
	assert.Contains(t, paths, "/users/profile")
	assert.Contains(t, paths, "/cases")
	assert.Contains(t, paths, "/clients")
	assert.Contains(t, paths, "/admin/users")
}

func TestSecurityDefinitions(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	r := gin.Default()

	// Add Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Load the swagger.json file
	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	var swaggerDoc map[string]interface{}
	json.Unmarshal(body, &swaggerDoc)

	// Check that BearerAuth security definition exists
	securityDefs := swaggerDoc["securityDefinitions"].(map[string]interface{})
	assert.Contains(t, securityDefs, "BearerAuth")

	bearerAuth := securityDefs["BearerAuth"].(map[string]interface{})
	assert.Equal(t, "apiKey", bearerAuth["type"])
	assert.Equal(t, "header", bearerAuth["in"])
	assert.Equal(t, "Authorization", bearerAuth["name"])
}
