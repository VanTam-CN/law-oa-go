//go:build np_i11_redis

package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"law-oa-go/internal/middleware"
)

func TestI11ResponseCacheRealRedis(t *testing.T) {
	addr := os.Getenv("NP_I11_REDIS")
	host, _, err := net.SplitHostPort(addr)
	if err != nil || host != "127.0.0.1" {
		t.Fatal("disposable loopback Redis required")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	info, err := client.Info(ctx, "server").Result()
	if err != nil || !strings.Contains(info, "redis_version:7.") {
		t.Fatal("fixture requires Redis 7")
	}
	commandStats := func() string {
		info, err := client.Info(ctx, "commandstats").Result()
		if err != nil {
			t.Fatal(err)
		}
		var counted []string
		for _, line := range strings.Split(info, "\n") {
			if strings.HasPrefix(line, "cmdstat_get:") || strings.HasPrefix(line, "cmdstat_set:") {
				counted = append(counted, line)
			}
		}
		return strings.Join(counted, "\n")
	}
	engine := newHTTPEngine()
	engine.Use(middleware.CacheMiddleware(middleware.CacheConfig{RedisClient: client, KeyPrefix: "i11"}))
	current := "revision-one"
	engine.GET("/*resource", func(c *gin.Context) {
		if c.GetHeader("X-Fixture-Actor") != "granted" {
			c.Status(403)
			return
		}
		c.String(200, current)
	})
	stale, err := json.Marshal(middleware.CacheResponse{Status: 200, Body: []byte("old")})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/case-intakes/1/source-snapshots/2", "/api/v1/case-intakes/write-results/1", "/api/v1/case-intake-assist/contexts/1", "/api/v1/cases/intake-workbench/1"} {
		if err := client.Set(ctx, "i11:"+path, stale, 0).Err(); err != nil {
			t.Fatal(err)
		}
		before := commandStats()
		for _, tc := range []struct {
			actor, version string
			status         int
		}{{"granted", "revision-one", 200}, {"revoked", "", 403}, {"granted", "revision-two", 200}} {
			current = tc.version
			req := httptest.NewRequest("GET", path, nil)
			req.Header.Set("X-Fixture-Actor", tc.actor)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)
			if rec.Code != tc.status || rec.Body.String() != tc.version || rec.Header().Get("Cache-Control") != "private, no-store" {
				t.Fatal("incorrect access, version or cache headers")
			}
		}
		if commandStats() != before {
			t.Fatal("business cache touched Redis")
		}
	}
}
