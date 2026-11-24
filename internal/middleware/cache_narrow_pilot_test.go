package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestNarrowPilotCacheZeroReadWrite(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cachePrefix := "m4"
	cached, err := json.Marshal(CacheResponse{Status: 200, Body: []byte("old synthetic body")})
	if err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	engine.Use(CacheMiddleware(CacheConfig{KeyPrefix: cachePrefix, RedisClient: client}))
	engine.GET("/*resource", func(c *gin.Context) {
		if c.GetHeader("X-Fixture-Access") != "granted" {
			c.String(403, "denied")
			return
		}
		c.String(200, "current synthetic version")
	})

	serve := func(path, access string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("X-Fixture-Access", access)
		r := httptest.NewRecorder()
		engine.ServeHTTP(r, req)
		return r
	}
	for _, path := range []string{
		"/api/v1/case-intakes", "/api/v1/case-intakes/1", "/api/v1/case-intakes/1/source-snapshots/2",
		"/api/v1/case-intakes/write-results/1", "/api/v1/case-intake-assist/contexts/1",
		"/api/v1/cases/intake-workbench/1",
	} {
		if err := server.Set(cachePrefix+":"+path, string(cached)); err != nil {
			t.Fatal(err)
		}
		before := server.CommandCount()
		for _, test := range []struct {
			access string
			status int
			body   string
		}{
			{"granted", 200, "current synthetic version"}, {"revoked", 403, "denied"},
		} {
			r := serve(path, test.access)
			if r.Code != test.status || r.Body.String() != test.body || r.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatalf("%s %s: code=%d, body=%q, cache-control=%q", path, test.access, r.Code, r.Body.String(), r.Header().Get("Cache-Control"))
			}
		}
		if after := server.CommandCount(); after != before {
			t.Fatalf("%s: cache commands %d -> %d", path, before, after)
		}
	}
	for _, path := range []string{"/api/v1/cases/1", "/api/v1/case-intakes-other"} {
		if err := server.Set(cachePrefix+":"+path, string(cached)); err != nil {
			t.Fatal(err)
		}
		r := serve(path, "granted")
		if !strings.HasPrefix(r.Body.String(), "old synthetic body") || r.Header().Get("Cache-Control") != "" {
			t.Fatalf("neighbor %s lost normal cached response", path)
		}
	}
}
