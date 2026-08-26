package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"law-oa-go/internal/models"
)

// MCPClient MCP服务客户端接口
type MCPClient interface {
	// FetchLatestStandards 获取最新标准
	FetchLatestStandards(ctx context.Context) (*models.MCPStandards, error)
	// GetBestPractices 获取最佳实践
	GetBestPractices(ctx context.Context) ([]string, error)
	// UpdateRules 更新规则
	UpdateRules(ctx context.Context, standards *models.MCPStandards) error
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
	// GetServiceStatus 获取服务状态
	GetServiceStatus(ctx context.Context) (*MCPServiceStatus, error)
}

// MCPServiceStatus MCP服务状态
type MCPServiceStatus struct {
	IsAvailable     bool      `json:"isAvailable"`
	LastChecked     time.Time `json:"lastChecked"`
	ResponseTime    int64     `json:"responseTime"`
	Version         string    `json:"version"`
	LastError       string    `json:"lastError,omitempty"`
	CachedStandards bool      `json:"cachedStandards"`
}

// MCPConfig MCP客户端配置
type MCPConfig struct {
	BaseURL    string        `json:"baseURL"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"maxRetries"`
	RetryDelay time.Duration `json:"retryDelay"`
	CacheTTL   time.Duration `json:"cacheTTL"`
	Enabled    bool          `json:"enabled"`
	APIKey     string        `json:"apiKey"`
}

// mcpClient MCP客户端实现
type mcpClient struct {
	config     *MCPConfig
	httpClient *http.Client
	redis      *redis.Client
}

// NewMCPClient 创建新的MCP客户端
func NewMCPClient(config *MCPConfig, redis *redis.Client) MCPClient {
	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 10,
		},
	}

	return &mcpClient{
		config:     config,
		httpClient: httpClient,
		redis:      redis,
	}
}

// FetchLatestStandards 获取最新标准
func (c *mcpClient) FetchLatestStandards(ctx context.Context) (*models.MCPStandards, error) {
	if !c.config.Enabled {
		return nil, models.ErrMCPServiceUnavailable
	}

	// 检查缓存
	if c.redis != nil {
		if cached, err := c.getCachedStandards(ctx); err == nil {
			return cached, nil
		}
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/api/v1/standards/latest", c.config.BaseURL)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Law-OA-Go/1.0")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	// 发送请求
	var response *http.Response
	err = c.retryRequest(ctx, func() error {
		var err error
		response, err = c.httpClient.Do(req)
		return err
	})
	if err != nil {
		c.updateServiceStatus(ctx, false, err)
		return nil, fmt.Errorf("请求MCP服务失败: %w", err)
	}
	defer response.Body.Close()

	// 检查响应状态
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		errMsg := fmt.Sprintf("MCP服务返回错误状态: %d, 响应: %s", response.StatusCode, string(body))
		c.updateServiceStatus(ctx, false, errors.New(errMsg))
		return nil, errors.New(errMsg)
	}

	// 解析响应
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var standards models.MCPStandards
	if err := json.Unmarshal(body, &standards); err != nil {
		return nil, fmt.Errorf("解析标准失败: %w", err)
	}

	// 更新服务状态
	c.updateServiceStatus(ctx, true, nil)

	// 缓存结果
	if c.redis != nil {
		c.cacheStandards(ctx, &standards)
	}

	return &standards, nil
}

// GetBestPractices 获取最佳实践
func (c *mcpClient) GetBestPractices(ctx context.Context) ([]string, error) {
	standards, err := c.FetchLatestStandards(ctx)
	if err != nil {
		return nil, err
	}

	return standards.BestPractices, nil
}

// UpdateRules 更新规则
func (c *mcpClient) UpdateRules(ctx context.Context, standards *models.MCPStandards) error {
	if !c.config.Enabled {
		return models.ErrMCPServiceUnavailable
	}

	// 这里应该调用规则服务更新规则
	// 由于规则服务尚未实现，暂时只更新MCP标准缓存

	// 缓存更新后的标准
	if c.redis != nil {
		c.cacheStandards(ctx, standards)
	}

	return nil
}

// HealthCheck 健康检查
func (c *mcpClient) HealthCheck(ctx context.Context) error {
	if !c.config.Enabled {
		return models.ErrMCPServiceUnavailable
	}

	url := fmt.Sprintf("%s/health", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建健康检查请求失败: %w", err)
	}

	start := time.Now()
	response, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.updateServiceStatus(ctx, false, err)
		return fmt.Errorf("健康检查失败: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("健康检查返回状态码: %d", response.StatusCode)
		c.updateServiceStatus(ctx, false, err)
		return err
	}

	c.updateServiceStatus(ctx, true, nil)

	// 更新响应时间
	status := &MCPServiceStatus{
		IsAvailable:  true,
		LastChecked:  time.Now(),
		ResponseTime: duration.Milliseconds(),
	}

	c.saveServiceStatus(ctx, status)

	return nil
}

// GetServiceStatus 获取服务状态
func (c *mcpClient) GetServiceStatus(ctx context.Context) (*MCPServiceStatus, error) {
	if c.redis == nil {
		return &MCPServiceStatus{
			IsAvailable: false,
			LastChecked: time.Now(),
			LastError:   "Redis不可用",
		}, nil
	}

	statusKey := "mcp:service:status"
	data, err := c.redis.Get(ctx, statusKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			// 没有缓存的状态，执行健康检查
			healthErr := c.HealthCheck(ctx)
			if healthErr != nil {
				return &MCPServiceStatus{
					IsAvailable: false,
					LastChecked: time.Now(),
					LastError:   healthErr.Error(),
				}, nil
			}
			// 重新获取状态
			return c.GetServiceStatus(ctx)
		}
		return nil, fmt.Errorf("获取服务状态失败: %w", err)
	}

	var status MCPServiceStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("解析服务状态失败: %w", err)
	}

	return &status, nil
}

// retryRequest 重试请求
func (c *mcpClient) retryRequest(ctx context.Context, requestFunc func() error) error {
	var lastErr error

	for i := 0; i < c.config.MaxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		err := requestFunc()
		if err == nil {
			return nil
		}

		lastErr = err

		// 如果是上下文取消，直接返回
		if errors.Is(err, context.Canceled) {
			return err
		}
	}

	return fmt.Errorf("重试%d次后仍然失败: %w", c.config.MaxRetries, lastErr)
}

// getCachedStandards 获取缓存的标准
func (c *mcpClient) getCachedStandards(ctx context.Context) (*models.MCPStandards, error) {
	cacheKey := "mcp:standards:latest"
	data, err := c.redis.Get(ctx, cacheKey).Bytes()
	if err != nil {
		return nil, err
	}

	var standards models.MCPStandards
	if err := json.Unmarshal(data, &standards); err != nil {
		return nil, err
	}

	return &standards, nil
}

// cacheStandards 缓存标准
func (c *mcpClient) cacheStandards(ctx context.Context, standards *models.MCPStandards) {
	cacheKey := "mcp:standards:latest"
	data, err := json.Marshal(standards)
	if err != nil {
		return
	}

	c.redis.Set(ctx, cacheKey, data, c.config.CacheTTL)
}

// updateServiceStatus 更新服务状态
func (c *mcpClient) updateServiceStatus(ctx context.Context, isAvailable bool, err error) {
	status := &MCPServiceStatus{
		IsAvailable: isAvailable,
		LastChecked: time.Now(),
	}

	if !isAvailable && err != nil {
		status.LastError = err.Error()
	}

	c.saveServiceStatus(ctx, status)
}

// saveServiceStatus 保存服务状态
func (c *mcpClient) saveServiceStatus(ctx context.Context, status *MCPServiceStatus) {
	if c.redis == nil {
		return
	}

	data, err := json.Marshal(status)
	if err != nil {
		return
	}

	c.redis.Set(ctx, "mcp:service:status", data, 5*time.Minute)
}

// GetMCPConfig 获取MCP配置（从环境变量）
func GetMCPConfig() *MCPConfig {
	return &MCPConfig{
		BaseURL:    getEnv("MCP_BASE_URL", "https://api.mcp.example.com"),
		Timeout:    getDurationEnv("MCP_TIMEOUT", 30*time.Second),
		MaxRetries: getIntEnv("MCP_MAX_RETRIES", 3),
		RetryDelay: getDurationEnv("MCP_RETRY_DELAY", 1*time.Second),
		CacheTTL:   getDurationEnv("MCP_CACHE_TTL", 1*time.Hour),
		Enabled:    getBoolEnv("MCP_ENABLED", false),
		APIKey:     getEnv("MCP_API_KEY", ""),
	}
}

// 环境变量辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
