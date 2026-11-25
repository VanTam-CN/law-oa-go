package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"law-oa-go/internal/logger"
	"law-oa-go/internal/middleware"
)

// benchRedis 由 fixture_test.go 提供（b testing.TB -> *redis.Client）。

// benchCacheConfig 生产 CacheMiddleware 的基准配置（无跳过规则，全 GET 可缓存）。
func benchCacheConfig(client *redis.Client) middleware.CacheConfig {
	return middleware.CacheConfig{
		TTL:         5 * time.Minute,
		RedisClient: client,
		KeyPrefix:   npBenchTag + "cache-mech",
		MaxBodySize: 1024 * 1024,
	}
}

func BenchmarkCacheMiddleware(b *testing.B) {
	client := benchRedis(b)
	previous := logger.Logger
	logger.Logger = zap.NewNop()
	b.Cleanup(func() { logger.Logger = previous })
	b.Run("CacheHit", func(b *testing.B) {
		cfg := benchCacheConfig(client)
		path := "/api/bench-hit"
		key := cfg.KeyPrefix + ":" + path
		payload := []byte(`{"mechanism":"cached-payload"}`)
		cached, err := json.Marshal(middleware.CacheResponse{
			Status: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"X-Mechanism":  "hit",
			},
			Body: payload,
		})
		if err != nil {
			b.Fatalf("marshal CacheResponse: %v", err)
		}
		ctx := context.Background()
		if err := client.Set(ctx, key, cached, 5*time.Minute).Err(); err != nil {
			b.Fatalf("prewrite cache: %v", err)
		}

		b.Cleanup(func() { require.NoError(b, client.Del(ctx, key).Err()) })
		// Measure only the cache read. Downstream short-circuiting is a separate integration check.
		app := gin.New()
		app.GET(path, middleware.CacheMiddleware(cfg))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != http.StatusOK || w.Header().Get("X-Mechanism") != "hit" || !bytes.Equal(w.Body.Bytes(), payload) {
				b.Fatal("expected preloaded response")
			}
		}
	})

	b.Run("CacheMiss", func(b *testing.B) {
		cfg := benchCacheConfig(client)
		cfg.MaxBodySize = 1 // Response is deliberately too large to cache.
		path := "/api/bench-miss"
		key := cfg.KeyPrefix + ":" + path
		require.NoError(b, client.Del(context.Background(), key).Err())
		app := gin.New()
		app.GET(path, middleware.CacheMiddleware(cfg), func(c *gin.Context) { c.String(http.StatusOK, "miss") })
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != http.StatusOK || w.Body.String() != "miss" {
				b.Fatal("expected uncached response")
			}
		}
		b.StopTimer()
		require.ErrorIs(b, client.Get(context.Background(), key).Err(), redis.Nil)
	})
}

// benchHTTPRequestsTotal 读取生产 PrometheusMiddleware 暴露在默认注册表中的
// http_requests_total 在指定 endpoint 上的当前计数值。
func benchHTTPRequestsTotal(b *testing.B, endpoint string) float64 {
	b.Helper()
	families, err := prom.DefaultGatherer.Gather()
	if err != nil {
		b.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != "law_oa_http_requests_total" {
			continue
		}
		for _, m := range family.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "endpoint" && lp.GetValue() == endpoint {
					return m.GetCounter().GetValue()
				}
			}
		}
	}
	return 0
}

func BenchmarkMiddlewarePerformance(b *testing.B) {
	b.Run("PerformanceMiddleware", func(b *testing.B) {
		endpoint := "/api/bench-perf"
		app := gin.New()
		app.Use(middleware.PrometheusMiddleware())
		app.GET(endpoint, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		before := benchHTTPRequestsTotal(b, endpoint)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, endpoint, nil))
		}
		b.StopTimer()

		// 终端校验：生产性能标记中间件按迭代次数精确计数。
		delta := benchHTTPRequestsTotal(b, endpoint) - before
		if delta != float64(b.N) {
			b.Fatalf("http_requests_total delta = %v, want %v", delta, b.N)
		}
	})

	b.Run("RateLimitMiddleware", func(b *testing.B) {
		client := benchRedis(b)
		ctx := context.Background()
		prefix := npBenchTag + "rl-bench"
		limit := int64(1_000_000_000)
		cfg := middleware.RateLimitConfig{
			RedisClient: client,
			KeyPrefix:   prefix,
			Limit:       limit,
			Window:      time.Hour,
		}
		app := gin.New()
		app.Use(middleware.RateLimitMiddleware(cfg))
		app.GET("/api/bench-rl", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/bench-rl", nil))
		}
		b.StopTimer()

		// 终端校验 1：放行路径携带生产限流标记头。
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/bench-rl", nil))
		if w.Code != http.StatusOK {
			b.Fatalf("rate limit allowed status = %d", w.Code)
		}
		if got := w.Header().Get("X-RateLimit-Limit"); got != strconv.FormatInt(limit, 10) {
			b.Fatalf("X-RateLimit-Limit = %q", got)
		}
		if remaining, err := strconv.ParseInt(w.Header().Get("X-RateLimit-Remaining"), 10, 64); err != nil || remaining >= limit {
			b.Fatalf("X-RateLimit-Remaining = %q err=%v", w.Header().Get("X-RateLimit-Remaining"), err)
		}

		// 终端校验 2：独立 key 证明超限后生产中间件返回 429。
		proofPrefix := npBenchTag + "rl-proof"
		proofCfg := middleware.RateLimitConfig{
			RedisClient: client,
			KeyPrefix:   proofPrefix,
			Limit:       2,
			Window:      time.Hour,
		}
		proofApp := gin.New()
		proofApp.Use(middleware.RateLimitMiddleware(proofCfg))
		proofApp.GET("/api/bench-rl-proof", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		codes := make([]int, 0, 3)
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			proofApp.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/bench-rl-proof", nil))
			codes = append(codes, w.Code)
		}
		if codes[0] != http.StatusOK || codes[1] != http.StatusOK || codes[2] != http.StatusTooManyRequests {
			b.Fatalf("rate limit proof codes = %v, want [200 200 429]", codes)
		}

		// 清理基准与证明使用的 Redis key。
		keys := []string{prefix + ":ip:192.0.2.1", proofPrefix + ":ip:192.0.2.1"}
		require.NoError(b, client.Del(ctx, keys...).Err())
	})

	b.Run("CORSMiddleware", func(b *testing.B) {
		origin := "http://localhost:3003"
		app := gin.New()
		app.Use(middleware.CORS())
		app.OPTIONS("/api/bench-cors", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodOptions, "/api/bench-cors", nil)
			req.Header.Set("Origin", origin)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
		b.StopTimer()

		// 终端校验 1：放行 origin 回显 ACAO 并返回预检 204。
		req := httptest.NewRequest(http.MethodOptions, "/api/bench-cors", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			b.Fatalf("preflight status = %d, want 204", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
			b.Fatalf("ACAO = %q, want %q", got, origin)
		}

		// 终端校验 2：非放行 origin 不下发 ACAO。
		req2 := httptest.NewRequest(http.MethodOptions, "/api/bench-cors", nil)
		req2.Header.Set("Origin", "http://evil.example")
		w2 := httptest.NewRecorder()
		app.ServeHTTP(w2, req2)
		if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "" {
			b.Fatalf("ACAO for disallowed origin = %q, want empty", got)
		}
	})
}
