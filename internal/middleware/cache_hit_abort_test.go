package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 命中分支必须 Abort：缓存响应写出后，Gin 处理链不得继续执行下游 handler；
// 未命中路径仍由下游 handler 提供正文。
func TestCacheMiddlewareHitAbortsDownstream(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	prefix := "cache-hit-abort"
	body := []byte(`{"mechanism":"cached"}`)
	cached, err := json.Marshal(CacheResponse{
		Status:  http.StatusNonAuthoritativeInfo,
		Headers: map[string]string{"Content-Type": "application/json", "X-Mech": "hit"},
		Body:    body,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Set(prefix+":/api/bench-hit", string(cached)); err != nil {
		t.Fatal(err)
	}

	downstream := 0
	engine := gin.New()
	engine.Use(CacheMiddleware(CacheConfig{KeyPrefix: prefix, RedisClient: client}))
	engine.GET("/api/bench-hit", func(c *gin.Context) {
		downstream++
		c.String(http.StatusOK, "live")
	})
	engine.GET("/api/bench-miss", func(c *gin.Context) {
		downstream++
		c.String(http.StatusOK, "miss")
	})

	hit := httptest.NewRecorder()
	engine.ServeHTTP(hit, httptest.NewRequest(http.MethodGet, "/api/bench-hit", nil))
	if hit.Code != http.StatusNonAuthoritativeInfo || hit.Body.String() != string(body) || hit.Header().Get("X-Mech") != "hit" {
		t.Fatalf("hit: code=%d body=%q x-mech=%q", hit.Code, hit.Body.String(), hit.Header().Get("X-Mech"))
	}
	if downstream != 0 {
		t.Fatalf("hit downstream runs = %d, want 0", downstream)
	}

	miss := httptest.NewRecorder()
	engine.ServeHTTP(miss, httptest.NewRequest(http.MethodGet, "/api/bench-miss", nil))
	if miss.Code != http.StatusOK || miss.Body.String() != "miss" {
		t.Fatalf("miss: code=%d body=%q", miss.Code, miss.Body.String())
	}
	if downstream != 1 {
		t.Fatalf("miss downstream runs = %d, want 1", downstream)
	}
}
