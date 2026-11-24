package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type noAuditBodyRead struct{ read bool }

func (b *noAuditBodyRead) Read([]byte) (int, error) { b.read = true; return 0, io.EOF }
func (b *noAuditBodyRead) Close() error             { return nil }

func TestNarrowPilotAuditSkipsSensitiveBodyBeforeRead(t *testing.T) {
	for _, path := range []string{"/api/v1/case-intakes/1", "/api/v1/case-intake-assist/contexts", "/api/v1/cases/intake-workbench/1"} {
		if !containsSensitiveData(path) {
			t.Fatalf("not sensitive: %s", path)
		}
	}
	if containsSensitiveData("/api/v1/case-intakes-other") {
		t.Fatal("prefix boundary too broad")
	}
	body := &noAuditBodyRead{}
	r := httptest.NewRequest("POST", "/api/v1/case-intake-assist/contexts", strings.NewReader(""))
	r.Body = body
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(AuditMiddleware())
	engine.POST("/api/v1/case-intake-assist/contexts", func(c *gin.Context) { c.Status(204) })
	engine.ServeHTTP(httptest.NewRecorder(), r)
	if body.read {
		t.Fatal("audit read sensitive body")
	}
}
