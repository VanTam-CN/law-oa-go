package executors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"law-oa-go/internal/testing"

	"github.com/playwright-community/playwright-go"
)

// UIExecutor UI测试执行器
type UIExecutor struct {
	*testing.BaseExecutor
	browser playwright.Browser
	page    playwright.Page
	options *UIExecutorOptions
}

// UIExecutorOptions UI执行器选项
type UIExecutorOptions struct {
	ScreenshotDir     string
	VideoDir          string
	DownloadsDir      string
	DefaultTimeout    time.Duration
	NavigationTimeout time.Duration
	ActionTimeout     time.Duration
	WaitTimeout       time.Duration
	Viewport          playwright.ViewportSize
	EmulateMedia      *playwright.Media
	Headless          bool
	SlowMo            time.Duration
	TraceDir          string
	UserAgent         string
	IgnoreHTTPSErrors bool
}

// NewUIExecutor 创建UI测试执行器
func NewUIExecutor(options *testing.TestExecutorOptions, logger testing.TestLogger, metrics testing.TestMetrics) *UIExecutor {
	base := testing.NewBaseExecutor(options, logger, metrics)

	uiOptions := &UIExecutorOptions{
		ScreenshotDir:     "screenshots",
		VideoDir:          "videos",
		DownloadsDir:      "downloads",
		TraceDir:          "traces",
		DefaultTimeout:    options.Timeout,
		NavigationTimeout: options.Timeout,
		ActionTimeout:     options.Timeout,
		WaitTimeout:       options.Timeout,
		Headless:          options.Headless,
		SlowMo:            options.Timeout / 10,
		IgnoreHTTPSErrors: !options.VerifySSL,
		Viewport: playwright.ViewportSize{
			Width:  int64(options.WindowSize["width"]),
			Height: int64(options.WindowSize["height"]),
		},
		UserAgent: options.UserAgent,
	}

	return &UIExecutor{
		BaseExecutor: base,
		options:      uiOptions,
	}
}

// GetExecutorType 获取执行器类型
func (e *UIExecutor) GetExecutorType() testing.TestType {
	return testing.TestTypeUI
}

// Setup 设置UI测试执行器
func (e *UIExecutor) Setup(ctx context.Context, executionCtx *testing.ExecutionContext) error {
	// 调用基础设置
	if err := e.BaseExecutor.Setup(ctx, executionCtx); err != nil {
		return err
	}

	// 初始化Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start Playwright: %w", err)
	}

	// 创建浏览器实例
	browserOptions := playwright.BrowserTypeLaunchOptions{
		Headless:          &e.options.Headless,
		SlowMo:            e.options.SlowMo.Milliseconds(),
		IgnoreHTTPSErrors: &e.options.IgnoreHTTPSErrors,
	}

	var browser playwright.Browser
	switch e.options.BrowserType {
	case "chromium":
		browser, err = pw.Chromium.Launch(browserOptions)
	case "firefox":
		browser, err = pw.Firefox.Launch(browserOptions)
	case "webkit":
		browser, err = pw.Webkit.Launch(browserOptions)
	default:
		browser, err = pw.Chromium.Launch(browserOptions)
	}

	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	e.browser = browser
	e.logger.Info("Browser launched", "type", e.options.BrowserType, "headless", e.options.Headless)

	// 创建新页面
	page, err := e.browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport:         &e.options.Viewport,
		UserAgent:        &e.options.UserAgent,
		DefaultTimeout:   &e.options.DefaultTimeout,
		DefaultNavigationTimeout: &e.options.NavigationTimeout,
	})
	if err != nil {
		e.browser.Close()
		return fmt.Errorf("failed to create page: %w", err)
	}

	e.page = page
	e.logger.Info("Page created", "viewport", e.options.Viewport)

	// 设置媒体模拟
	if e.options.EmulateMedia != nil {
		if err := e.page.EmulateMedia(playwright.PageEmulateMediaOptions{
			Media: e.options.EmulateMedia,
		}); err != nil {
			e.logger.Warn("Failed to emulate media", "error", err)
		}
	}

	// 创建必要的目录
	if err := e.ensureDirectories(); err != nil {
		e.logger.Warn("Failed to create directories", "error", err)
	}

	// 开始追踪
	if e.options.TraceDir != "" {
		if err := e.page.Context().Tracing.Start(playwright.TracingStartOptions{
			Screenshots: playwright.Bool(true),
			Snapshots:   playwright.Bool(true),
			Sources:     playwright.Bool(true),
		}); err != nil {
			e.logger.Warn("Failed to start tracing", "error", err)
		}
	}

	return nil
}

// Teardown 清理UI测试执行器
func (e *UIExecutor) Teardown(ctx context.Context, executionCtx *testing.ExecutionContext) error {
	defer e.BaseExecutor.Teardown(ctx, executionCtx)

	// 停止追踪并保存
	if e.page != nil && e.options.TraceDir != "" {
		traceFile := filepath.Join(e.options.TraceDir, fmt.Sprintf("trace_%s_%d.zip", executionCtx.ExecutionID, time.Now().Unix()))
		if err := e.page.Context().Tracing.Stop(playwright.TracingStopOptions{
			Path: playwright.String(traceFile),
		}); err != nil {
			e.logger.Warn("Failed to save trace", "error", err)
		}
	}

	// 关闭页面
	if e.page != nil {
		if err := e.page.Close(); err != nil {
			e.logger.Warn("Failed to close page", "error", err)
		}
	}

	// 关闭浏览器
	if e.browser != nil {
		if err := e.browser.Close(); err != nil {
			e.logger.Warn("Failed to close browser", "error", err)
		}
	}

	e.logger.Info("UI executor teardown completed")
	return nil
}

// executeMainTest 执行UI测试的主逻辑
func (e *UIExecutor) executeMainTest(ctx context.Context, test *testing.TestCase, result *testing.TestResult, executionCtx *testing.ExecutionContext) error {
	e.logger.Info("Executing UI test", "name", test.Name, "url", test.URL)

	// 执行步骤
	for i, step := range test.Steps {
		if err := e.executeStep(ctx, step, result, executionCtx); err != nil {
			return fmt.Errorf("step %d (%s) failed: %w", i+1, step.Name, err)
		}

		// 检查是否需要截图
		if step.Screenshot {
			screenshotName := fmt.Sprintf("%s_step_%d", test.Name, i+1)
			if err := e.takeScreenshot(screenshotName, result); err != nil {
				e.logger.Warn("Failed to take step screenshot", "step", i+1, "error", err)
			}
		}
	}

	// 记录UI特定指标
	if e.page != nil {
		metrics, err := e.collectUIMetrics()
		if err != nil {
			e.logger.Warn("Failed to collect UI metrics", "error", err)
		} else {
			result.Metadata["ui_metrics"] = metrics
		}
	}

	return nil
}

// ExecuteNavigate 执行导航步骤
func (e *UIExecutor) ExecuteNavigate(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	targetURL := e.substituteVariables(step.Target, executionCtx.Variables)

	// 处理相对URL
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		if executionCtx.BaseURL != "" {
			targetURL = executionCtx.BaseURL + targetURL
		} else {
			targetURL = "http://localhost:8080" + targetURL
		}
	}

	e.logger.Info("Navigating to", "url", targetURL)

	options := playwright.PageGotoOptions{
		Timeout: playwright.Float(e.options.NavigationTimeout.Seconds() * 1000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}

	_, err := e.page.Goto(targetURL, options)
	if err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	// 等待页面加载
	if err := e.page.WaitForLoadState(playwright.PageWaitForLoadStateStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	}); err != nil {
		e.logger.Warn("Failed to wait for DOM content loaded", "error", err)
	}

	return nil
}

// ExecuteClick 执行点击步骤
func (e *UIExecutor) ExecuteClick(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	locator := e.substituteVariables(step.Target, executionCtx.Variables)
	e.logger.Info("Clicking on", "locator", locator)

	timeout := e.options.ActionTimeout
	if step.Wait > 0 {
		timeout = step.Wait
	}

	options := playwright.PageClickOptions{
		Timeout: playwright.Float(timeout.Seconds() * 1000),
	}

	// 处理双击
	if step.Action == "double_click" {
		return e.page.Dblclick(locator, options)
	}

	// 处理右键点击
	if step.Action == "right_click" {
		return e.page.Click(locator, playwright.PageClickOptions{
			Button: playwright.MouseButtonRight,
			Timeout: playwright.Float(timeout.Seconds() * 1000),
		})
	}

	return e.page.Click(locator, options)
}

// ExecuteFill 执行填充步骤
func (e *UIExecutor) ExecuteFill(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	locator := e.substituteVariables(step.Target, executionCtx.Variables)
	value := e.substituteValue(step.Value, executionCtx.Variables)

	e.logger.Info("Filling field", "locator", locator, "value", value)

	timeout := e.options.ActionTimeout
	if step.Wait > 0 {
		timeout = step.Wait
	}

	// 先清空字段（如果需要）
	if step.Action == "fill_clear" || step.Action == "clear_and_fill" {
		if err := e.page.Fill(locator, ""); err != nil {
			e.logger.Warn("Failed to clear field", "locator", locator, "error", err)
		}
	}

	return e.page.Fill(locator, value, playwright.PageFillOptions{
		Timeout: playwright.Float(timeout.Seconds() * 1000),
	})
}

// ExecuteWait 执行等待步骤
func (e *UIExecutor) ExecuteWait(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	waitTime := step.Wait
	if waitTime == 0 {
		waitTime = e.options.WaitTimeout
	}

	// 如果有目标选择器，等待元素出现
	if step.Target != "" {
		locator := e.substituteVariables(step.Target, executionCtx.Variables)
		e.logger.Info("Waiting for element", "locator", locator, "timeout", waitTime)

		timeout := playwright.Float(waitTime.Seconds() * 1000)

		if step.Action == "wait_hidden" {
			if err := e.page.WaitForSelector(locator, playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: timeout,
			}); err != nil {
				return fmt.Errorf("element not hidden: %w", err)
			}
		} else if step.Action == "wait_visible" {
			if err := e.page.WaitForSelector(locator, playwright.PageWaitForSelectorOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: timeout,
			}); err != nil {
				return fmt.Errorf("element not visible: %w", err)
			}
		} else {
			if err := e.page.WaitForSelector(locator, playwright.PageWaitForSelectorOptions{
				Timeout: timeout,
			}); err != nil {
				return fmt.Errorf("element not found: %w", err)
			}
		}
	} else {
		// 简单的时间等待
		e.logger.Info("Waiting for", "duration", waitTime)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// 等待完成
		}
	}

	return nil
}

// ExecuteScreenshot 执行截图步骤
func (e *UIExecutor) ExecuteScreenshot(ctx context.Context, step testing.TestStep, result *testing.TestResult) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	screenshotName := step.Name
	if screenshotName == "" {
		screenshotName = fmt.Sprintf("screenshot_%d", time.Now().Unix())
	}

	return e.takeScreenshot(screenshotName, result)
}

// ExecuteJavaScript 执行JavaScript步骤
func (e *UIExecutor) ExecuteJavaScript(ctx context.Context, step testing.TestStep, result *testing.TestResult, executionCtx *testing.ExecutionContext) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	script := e.substituteVariables(step.Action, executionCtx.Variables)
	e.logger.Info("Executing JavaScript", "script", script)

	// 准备参数
	var args []interface{}
	if step.Value != nil {
		switch v := step.Value.(type) {
		case []interface{}:
			args = v
		default:
			args = []interface{}{v}
		}
	}

	// 执行脚本
	jsResult, err := e.page.Evaluate(script, args...)
	if err != nil {
		return fmt.Errorf("JavaScript execution failed: %w", err)
	}

	// 记录结果
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["javascript_result"] = jsResult
	result.Metadata["javascript_step"] = step.Name

	return nil
}

// takeScreenshot 截图
func (e *UIExecutor) takeScreenshot(name string, result *testing.TestResult) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.png", name, timestamp)
	screenshotPath := filepath.Join(e.options.ScreenshotDir, filename)

	// 确保目录存在
	if err := os.MkdirAll(e.options.ScreenshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshot directory: %w", err)
	}

	// 截图
	options := playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
		Path:     playwright.String(screenshotPath),
	}

	screenshot, err := e.page.Screenshot(options)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	// 如果需要base64编码（用于嵌入报告）
	base64Image := base64.StdEncoding.EncodeToString(screenshot)
	result.Metadata["screenshot_path"] = screenshotPath
	result.Metadata["screenshot_base64"] = base64Image
	result.Screenshots = append(result.Screenshots, screenshotPath)

	e.logger.Info("Screenshot taken", "path", screenshotPath)
	return nil
}

// collectUIMetrics 收集UI指标
func (e *UIExecutor) collectUIMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 获取页面性能指标
	perfMetrics, err := e.page.Evaluate("() => JSON.stringify(window.performance.timing)")
	if err == nil && perfMetrics != nil {
		var timing map[string]interface{}
		if err := json.Unmarshal([]byte(perfMetrics.(string)), &timing); err == nil {
			metrics["performance_timing"] = timing
		}
	}

	// 获取页面大小信息
	if e.page != nil {
		viewport := e.page.ViewportSize()
		metrics["viewport"] = map[string]int64{
			"width":  viewport.Width,
			"height": viewport.Height,
		}
	}

	// 获取URL和标题
	if title, err := e.page.Title(); err == nil {
		metrics["page_title"] = title
	}
	if url := e.page.URL(); url != "" {
		metrics["page_url"] = url
	}

	return metrics, nil
}

// ensureDirectories 确保必要的目录存在
func (e *UIExecutor) ensureDirectories() error {
	dirs := []string{
		e.options.ScreenshotDir,
		e.options.VideoDir,
		e.options.DownloadsDir,
		e.options.TraceDir,
	}

	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}

	return nil
}

// substituteVariables 替换字符串中的变量
func (e *UIExecutor) substituteVariables(input string, variables map[string]interface{}) string {
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

// substituteValue 替换值中的变量
func (e *UIExecutor) substituteValue(value interface{}, variables map[string]interface{}) interface{} {
	if str, ok := value.(string); ok {
		return e.substituteVariables(str, variables)
	}
	return value
}

// GetPage 获取页面实例（用于高级操作）
func (e *UIExecutor) GetPage() playwright.Page {
	return e.page
}

// GetBrowser 获取浏览器实例（用于高级操作）
func (e *UIExecutor) GetBrowser() playwright.Browser {
	return e.browser
}

// SetContext 设置测试上下文
func (e *UIExecutor) SetContext(ctx context.Context, executionCtx *testing.ExecutionContext) error {
	// 可以在这里设置额外的页面上下文
	return nil
}

// ValidatePage 验证页面状态
func (e *UIExecutor) ValidatePage(expectedURL string) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	currentURL := e.page.URL()
	if expectedURL != "" && !strings.Contains(currentURL, expectedURL) {
		return fmt.Errorf("expected URL containing %s, got %s", expectedURL, currentURL)
	}

	return nil
}

// WaitForNetworkIdle 等待网络空闲
func (e *UIExecutor) WaitForNetworkIdle(timeout time.Duration) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	return e.page.WaitForLoadState(playwright.PageWaitForLoadStateStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(timeout.Seconds() * 1000),
	})
}

// ScrollToElement 滚动到元素
func (e *UIExecutor) ScrollToElement(selector string) error {
	if e.page == nil {
		return fmt.Errorf("page not initialized")
	}

	_, err := e.page.EvalOnSelector(selector, "element => element.scrollIntoView({behavior: 'smooth', block: 'center'})")
	return err
}