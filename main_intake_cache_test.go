package main

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIntakeEarlyResponsesNoStore(t *testing.T) {
	engine := newHTTPEngine()
	engine.GET("/*path", func(c *gin.Context) { c.Status(200) })
	request := func(method, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.RemoteAddr = "192.0.2.218:24100"
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		return rec
	}
	// Exhaust the real limiter using an unrelated endpoint.
	for i := 0; i < 100; i++ {
		request("GET", "/health")
	}
	for _, path := range []string{"/api/v1/case-intakes", "/api/v1/case-intake-assist/capabilities", "/api/v1/cases/intake-workbench/1"} {
		for method, status := range map[string]int{"GET": 429, "OPTIONS": 204} {
			rec := request(method, path)
			if got := rec.Header().Get("Cache-Control"); rec.Code != status || got != "private, no-store" {
				t.Errorf("%s %s: status=%d cache-control=%q", method, path, rec.Code, got)
			}
		}
	}
	if rec := request("GET", "/api/v1/case-intakes-other"); rec.Code != 429 || rec.Header().Get("Cache-Control") != "" {
		t.Error("adjacent route unexpectedly received intake cache policy")
	}
}
