package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestClientHandler_GetClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Get Client Success", func(t *testing.T) {
		router := gin.New()
		router.GET("/clients/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"id":     1,
				"name":   "Test Client",
				"email":  "client@example.com",
				"status": "active",
			})
		})

		req, _ := http.NewRequest("GET", "/clients/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["id"])
		assert.Equal(t, "Test Client", response["name"])
	})

	t.Run("Get Client Not Found", func(t *testing.T) {
		router := gin.New()
		router.GET("/clients/:id", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Client not found",
				"code":  http.StatusNotFound,
			})
		})

		req, _ := http.NewRequest("GET", "/clients/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.Equal(t, float64(http.StatusNotFound), response["code"])
	})
}
