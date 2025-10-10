package executors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// APIExecutor API测试执行器
type APIExecutor struct {
	*testing.BaseExecutor
	client *http.Client
}

// NewAPIExecutor 创建API测试执行器
func NewAPIExecutor(options *testing.TestExecutorOptions, logger testing.TestLogger, metrics testing.TestMetrics) *APIExecutor {
	base := testing.NewBaseExecutor(options, logger, metrics)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: options.Timeout,
		Transport: &loggingTransport{
			Base: &http.Transport{
				MaxIdleConnsPerHost:   10,
				IdleConnTimeout:       30 * time.Second,
				DisableCompression:     false,
				ResponseHeaderTimeout:   options.Timeout,
			},
			Logger: logger,
		},
	}

	return &APIExecutor{
		BaseExecutor: base,
		client:        client,
	}
}

// GetExecutorType 获取执行器类型
func (e *APIExecutor) GetExecutorType() testing.TestType {
	return testing.TestTypeAPI
}

// executeMainTest 执行API测试的主逻辑
func (e *APIExecutor) executeMainTest(ctx context.Context, test *testing.TestCase, result *testing.TestResult, executionCtx *testing.ExecutionContext) error {
	e.logger.Info("Executing API test", "url", test.URL, "method", test.Method)

	// 构建请求
	req, err := e.buildRequest(test, executionCtx)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// 执行请求
	resp, err := e.client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 记录响应信息
	e.logger.Debug("HTTP response received",
		"status", resp.StatusCode,
		"content_length", len(body),
		"headers", resp.Header)

	// 验证响应
	if err := e.validateResponse(test, resp, body, result); err != nil {
		return err
	}

	// 记录响应数据到结果中
	result.Metadata["status_code"] = resp.StatusCode
	result.Metadata["response_headers"] = e.headersToMap(resp.Header)
	result.Metadata["response_body"] = string(body)

	return nil
}

// buildRequest 构建HTTP请求
func (e *APIExecutor) buildRequest(test *testing.TestCase, executionCtx *testing.ExecutionContext) (*http.Request, error) {
	// 解析URL
	url := test.URL
	if url == "" {
		return nil, fmt.Errorf("test URL cannot be empty")
	}

	// 处理相对URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if executionCtx.BaseURL != "" {
			url = executionCtx.BaseURL + url
		} else {
			url = "http://localhost:8080" + url
		}
	}

	// 处理变量替换
	url = e.substituteVariables(url, executionCtx.Variables)

	// 创建请求体
	var body io.Reader
	if test.Body != nil {
		switch v := test.Body.(type) {
		case string:
			bodyData := e.substituteVariables(v.(string), executionCtx.Variables)
			body = strings.NewReader(bodyData)
		case map[string]interface{}:
			jsonData, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = bytes.NewReader(jsonData)
		case []byte:
			body = bytes.NewReader(v)
		default:
			return nil, fmt.Errorf("unsupported body type: %T", v)
		}
	}

	// 创建请求
	req, err := http.NewRequest(test.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	if test.Headers != nil {
		for key, value := range test.Headers {
			value = e.substituteVariables(value, executionCtx.Variables)
			req.Header.Set(key, value)
		}
	}

	// 设置执行上下文头部
	for key, value := range executionCtx.Headers {
		req.Header.Set(key, value)
	}

	// 设置默认头部
	if req.Header.Get("Content-Type") == "" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", e.options.UserAgent)
	}

	return req, nil
}

// validateResponse 验证HTTP响应
func (e *APIExecutor) validateResponse(test *testing.TestCase, resp *http.Response, body []byte, result *testing.TestResult) error {
	// 验证状态码
	if test.Expected != nil && test.Expected.Status > 0 {
		assertion := &testing.TestAssertion{
			Name:     "status_code",
			Type:     "status",
			Actual:   resp.StatusCode,
			Expected: test.Expected.Status,
		}

		if resp.StatusCode != test.Expected.Status {
			assertion.Passed = false
			assertion.ErrorMessage = fmt.Sprintf("expected status %d, got %d", test.Expected.Status, resp.StatusCode)
		} else {
			assertion.Passed = true
		}

		result.Assertions = append(result.Assertions, assertion)

		if !assertion.Passed {
			return fmt.Errorf("status code assertion failed: %s", assertion.ErrorMessage)
		}
	}

	// 验证响应体包含
	if test.Expected != nil && len(test.Expected.BodyContains) > 0 {
		bodyStr := string(body)
		for _, expected := range test.Expected.BodyContains {
			assertion := &testing.TestAssertion{
				Name:     "body_contains",
				Type:     "contains",
				Actual:   bodyStr,
				Expected: expected,
			}

			if !strings.Contains(bodyStr, expected) {
				assertion.Passed = false
				assertion.ErrorMessage = fmt.Sprintf("response body does not contain expected string: %s", expected)
			} else {
				assertion.Passed = true
			}

			result.Assertions = append(result.Assertions, assertion)

			if !assertion.Passed {
				return fmt.Errorf("body contains assertion failed: %s", assertion.ErrorMessage)
			}
		}
	}

	// 验证响应时间
	if test.Expected != nil && test.Expected.ResponseTime > 0 {
		// 这里需要从结果中获取实际响应时间
		// 暂时跳过，因为响应时间在外部计算
	}

	// 验证内容类型
	if test.Expected != nil && test.Expected.ContentType != "" {
		contentType := resp.Header.Get("Content-Type")
		assertion := &testing.TestAssertion{
			Name:     "content_type",
			Type:     "header",
			Actual:   contentType,
			Expected: test.Expected.ContentType,
		}

		if !strings.Contains(contentType, test.Expected.ContentType) {
			assertion.Passed = false
			assertion.ErrorMessage = fmt.Sprintf("expected content type %s, got %s", test.Expected.ContentType, contentType)
		} else {
			assertion.Passed = true
		}

		result.Assertions = append(result.Assertions, assertion)

		if !assertion.Passed {
			return fmt.Errorf("content type assertion failed: %s", assertion.ErrorMessage)
		}
	}

	// 验证响应头
	if test.Expected != nil && len(test.Expected.Headers) > 0 {
		for key, expectedValue := range test.Expected.Headers {
			actualValue := resp.Header.Get(key)
			assertion := &testing.TestAssertion{
				Name:     fmt.Sprintf("header_%s", key),
				Type:     "header",
				Actual:   actualValue,
				Expected: expectedValue,
			}

			if actualValue != expectedValue {
				assertion.Passed = false
				assertion.ErrorMessage = fmt.Sprintf("expected header %s: %s, got: %s", key, expectedValue, actualValue)
			} else {
				assertion.Passed = true
			}

			result.Assertions = append(result.Assertions, assertion)

			if !assertion.Passed {
				return fmt.Errorf("header assertion failed: %s", assertion.ErrorMessage)
			}
		}
	}

	// 验证JSON响应体
	if test.Expected != nil && len(test.Expected.Body) != nil {
		var responseBody interface{}
		if err := json.Unmarshal(body, &responseBody); err == nil {
			// 深度比较JSON
			if err := e.deepEqualJSON(test.Expected.Body, responseBody, assertion); err != nil {
				return err
			}
		} else {
			// 如果不是JSON，直接比较
			bodyStr := string(body)
			expectedStr, ok := test.Expected.Body.(string)
			if ok {
				assertion := &testing.TestAssertion{
					Name:     "body_exact_match",
					Type:     "exact",
					Actual:   bodyStr,
					Expected: expectedStr,
				}

				if bodyStr != expectedStr {
					assertion.Passed = false
					assertion.ErrorMessage = "response body does not match expected body exactly"
				} else {
					assertion.Passed = true
				}

				result.Assertions = append(result.Assertions, assertion)

				if !assertion.Passed {
					return fmt.Errorf("body exact match assertion failed")
				}
			}
		}
	}

	return nil
}

// deepEqualJSON 深度比较JSON
func (e *APIExecutor) deepEqualJSON(expected, actual interface{}, assertion *testing.TestAssertion) error {
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return fmt.Errorf("failed to marshal expected JSON: %w", err)
	}

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return fmt.Errorf("failed to marshal actual JSON: w")
	}

	assertion.Name = "body_json_deep_equal"
	assertion.Type = "deep_equal"
	assertion.Actual = string(actualJSON)
	assertion.Expected = string(expectedJSON)

	if string(expectedJSON) != string(actualJSON) {
		assertion.Passed = false
		assertion.ErrorMessage = "JSON response body does not match expected JSON"
	} else {
		assertion.Passed = true
	}

	return nil
}

// substituteVariables 替换变量
func (e *APIExecutor) substituteVariables(input string, variables map[string]interface{}) string {
	result := input

	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		var valueStr string

		switch v := value.(type) {
		case string:
			valueStr = v
		case int, int8, int16, int32, int64:
			valueStr = fmt.Sprintf("%d", v)
		case float32, float64:
			valueStr = fmt.Sprintf("%f", v)
		case bool:
			valueStr = fmt.Sprintf("%t", v)
		default:
			valueJSON, _ := json.Marshal(value)
			valueStr = string(valueJSON)
		}

		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	return result
}

// headersToMap 转换HTTP头为map
func (e *APIExecutor) headersToMap(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// loggingTransport 记录HTTP请求的传输层
type loggingTransport struct {
	Base   http.RoundTripper
	Logger testing.TestLogger
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// 记录请求
	t.Logger.Info("HTTP Request",
		"method", req.Method,
		"url", req.URL.String(),
		"headers", req.Header,
		"content_length", req.ContentLength)

	// 执行请求
	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		t.Logger.Error("HTTP request failed", "error", err)
		return nil, err
	}

	// 记录响应
	duration := time.Since(start)
	t.Logger.Info("HTTP Response",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"duration", duration,
		"headers", resp.Header,
		"content_length", resp.ContentLength)

	return resp, err
}

// ExecuteNavigate 执行导航步骤（API测试中不适用）
func (e *APIExecutor) ExecuteNavigate(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	return fmt.Errorf("navigate step not applicable for API tests")
}

// ExecuteClick 执行点击步骤（API测试中不适用）
func (e *APIExecutor) ExecuteClick(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	return fmt.Errorf("click step not applicable for API tests")
}

// ExecuteFill 执行填充步骤（API测试中不适用）
func (e *APIExecutor) ExecuteFill(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	return fmt.Errorf("fill step not applicable for API tests")
}

// ExecuteWait 执行等待步骤
func (e *APIExecutor) ExecuteWait(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	// 检查等待条件
	if step.Target != "" {
		// 这里可以实现轮询等待，暂时使用简单等待
		return nil
	}
	return nil
}

// ExecuteScreenshot 执行截图步骤（API测试中不适用）
func (e *APIExecutor) ExecuteScreenshot(ctx context.Context, step testing.TestStep, result *testing.TestResult) error {
	return fmt.Errorf("screenshot step not applicable for API tests")
}

// ExecuteJavaScript 执行JavaScript步骤（API测试中不适用）
func (e *APIExecutor) ExecuteJavaScript(ctx context.Context, step testing.TestStep, result *testing.TestResult, executionCtx *testing.ExecutionContext) error {
	return fmt.Errorf("javascript step not applicable for API tests")
}