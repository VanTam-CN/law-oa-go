package handlers

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func conflictQueueLawyerContext(path string, userID uint) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, path, nil)
	context.Set("user_id", userID)
	context.Set("role", "lawyer")
	context.Set("username", "律师甲")
	return context, recorder
}
