// Package shutdown 提供优雅关闭功能
// 基于最新Go应用优雅关闭最佳实践

package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// ShutdownHook 关闭钩子函数
type ShutdownHook func(ctx context.Context) error

// HookInfo 钩子信息
type HookInfo struct {
	Name     string
	Hook     ShutdownHook
	Timeout  time.Duration
	Priority int // 优先级，数字越小优先级越高
}

// ShutdownManager 优雅关闭管理器
type ShutdownManager struct {
	logger     *zap.Logger
	hooks      []HookInfo
	timeout    time.Duration
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	signals    []os.Signal
}

// ShutdownConfig 关闭配置
type ShutdownConfig struct {
	Timeout            time.Duration
	GracePeriod        time.Duration
	ForceCloseAfter   time.Duration
	SilenceTimeout     bool
	EnableSignals      bool
	EnableHTTPShutdown bool
	EnableDBShutdown   bool
	EnableCacheShutdown bool
}

// DefaultShutdownConfig 默认关闭配置
var DefaultShutdownConfig = ShutdownConfig{
	Timeout:            30 * time.Second,
	GracePeriod:        10 * time.Second,
	ForceCloseAfter:    60 * time.Second,
	SilenceTimeout:     false,
	EnableSignals:      true,
	EnableHTTPShutdown: true,
	EnableDBShutdown:   true,
	EnableCacheShutdown: true,
}

// NewShutdownManager 创建新的关闭管理器
func NewShutdownManager(logger *zap.Logger, config *ShutdownConfig) *ShutdownManager {
	if config == nil {
		config = &DefaultShutdownConfig
	}

	ctx, cancel := context.WithCancel(context.Background())

	sm := &ShutdownManager{
		logger:  logger,
		hooks:   make([]HookInfo, 0),
		timeout: config.Timeout,
		ctx:     ctx,
		cancel:  cancel,
		signals: []os.Signal{
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // 来自Kubernetes的终止信号
			syscall.SIGQUIT, // Ctrl+\
		},
	}

	// 注册默认钩子
	sm.registerDefaultHooks(config)

	return sm
}

// registerDefaultHooks 注册默认钩子
func (sm *ShutdownManager) registerDefaultHooks(config *ShutdownConfig) {
	// HTTP服务器关闭钩子
	if config.EnableHTTPShutdown {
		sm.AddHook("http-server", sm.httpServerShutdown, 5*time.Second, 1)
	}

	// 数据库关闭钩子
	if config.EnableDBShutdown {
		sm.AddHook("database", sm.databaseShutdown, 10*time.Second, 2)
	}

	// 缓存关闭钩子
	if config.EnableCacheShutdown {
		sm.AddHook("cache", sm.cacheShutdown, 5*time.Second, 3)
	}

	// 监控关闭钩子
	sm.AddHook("monitoring", sm.monitoringShutdown, 5*time.Second, 4)

	// 日志关闭钩子
	sm.AddHook("logging", sm.loggingShutdown, 2*time.Second, 5)
}

// AddHook 添加关闭钩子
func (sm *ShutdownManager) AddHook(name string, hook ShutdownHook, timeout time.Duration, priority int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	hookInfo := HookInfo{
		Name:     name,
		Hook:     hook,
		Timeout:  timeout,
		Priority: priority,
	}

	// 按优先级插入
	inserted := false
	for i, existing := range sm.hooks {
		if priority < existing.Priority {
			sm.hooks = append(sm.hooks[:i], append([]HookInfo{hookInfo}, sm.hooks[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		sm.hooks = append(sm.hooks, hookInfo)
	}

	sm.logger.Info("注册关闭钩子",
		zap.String("name", name),
		zap.Duration("timeout", timeout),
		zap.Int("priority", priority),
	)
}

// RemoveHook 移除关闭钩子
func (sm *ShutdownManager) RemoveHook(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, hook := range sm.hooks {
		if hook.Name == name {
			sm.hooks = append(sm.hooks[:i], sm.hooks[i+1:]...)
			sm.logger.Info("移除关闭钩子", zap.String("name", name))
			return
		}
	}
}

// Start 开始监听关闭信号
func (sm *ShutdownManager) Start() {
	if len(sm.signals) == 0 {
		return
	}

	sm.wg.Add(1)
	go sm.signalHandler()
}

// Shutdown 执行优雅关闭
func (sm *ShutdownManager) Shutdown(ctx context.Context) error {
	sm.logger.Info("开始优雅关闭")
	startTime := time.Now()

	// 取消所有正在进行的操作
	sm.cancel()

	// 创建关闭上下文
	shutdownCtx := ctx
	if shutdownCtx == nil {
		shutdownCtx = context.Background()
	}

	// 设置总体超时
	shutdownCtx, cancel := context.WithTimeout(shutdownCtx, sm.timeout)
	defer cancel()

	// 等待所有钩子完成
	err := sm.runHooks(shutdownCtx)

	duration := time.Since(startTime)
	if err != nil {
		sm.logger.Error("优雅关闭失败",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		return err
	}

	sm.logger.Info("优雅关闭完成", zap.Duration("duration", duration))
	return nil
}

// signalHandler 信号处理器
func (sm *ShutdownManager) signalHandler() {
	defer sm.wg.Done()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sm.signals...)

	for {
		select {
		case sig, ok := <-sigChan:
			if !ok {
				return
			}

			sm.logger.Info("收到关闭信号", zap.String("signal", sig.String()))

			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				// 执行优雅关闭
				go func() {
					if err := sm.Shutdown(context.Background()); err != nil {
						sm.logger.Error("优雅关闭失败", zap.Error(err))
						os.Exit(1)
					}
					os.Exit(0)
				}()

			case syscall.SIGQUIT:
				// 立即关闭
				sm.logger.Warn("收到强制关闭信号")
				go func() {
					time.Sleep(2 * time.Second) // 给点时间清理
					os.Exit(1)
				}()
			}

		case <-sm.ctx.Done():
			return
		}
	}
}

// runHooks 运行所有关闭钩子
func (sm *ShutdownManager) runHooks(ctx context.Context) error {
	sm.mu.Lock()
	hooks := make([]HookInfo, len(sm.hooks))
	copy(hooks, sm.hooks)
	sm.mu.Unlock()

	// 创建错误通道
	errChan := make(chan error, len(hooks))
	var wg sync.WaitGroup

	// 并发执行所有钩子
	for _, hook := range hooks {
		wg.Add(1)
		go func(h HookInfo) {
			defer wg.Done()

			hookCtx, cancel := context.WithTimeout(ctx, h.Timeout)
			defer cancel()

			sm.logger.Info("执行关闭钩子", zap.String("name", h.Name))
			startTime := time.Now()

			if err := h.Hook(hookCtx); err != nil {
				duration := time.Since(startTime)
				sm.logger.Error("关闭钩子执行失败",
					zap.String("name", h.Name),
					zap.Error(err),
					zap.Duration("duration", duration),
				)
				errChan <- fmt.Errorf("钩子 %s 失败: %w", h.Name, err)
			} else {
				duration := time.Since(startTime)
				sm.logger.Info("关闭钩子执行成功",
					zap.String("name", h.Name),
					zap.Duration("duration", duration),
				)
			}
		}(hook)
	}

	// 等待所有钩子完成
	wg.Wait()
	close(errChan)

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("关闭钩子执行失败: %v", errors)
	}

	return nil
}

// GetContext 获取关闭上下文
func (sm *ShutdownManager) GetContext() context.Context {
	return sm.ctx
}

// IsShuttingDown 检查是否正在关闭
func (sm *ShutdownManager) IsShuttingDown() bool {
	select {
	case <-sm.ctx.Done():
		return true
	default:
		return false
	}
}

// Wait 等待关闭完成
func (sm *ShutdownManager) Wait() {
	sm.wg.Wait()
}

// 默认钩子实现

// httpServerShutdown HTTP服务器关闭钩子
func (sm *ShutdownManager) httpServerShutdown(ctx context.Context) error {
	// 这里应该关闭HTTP服务器
	// 由于我们没有直接的服务器引用，这里只是一个示例实现
	sm.logger.Info("HTTP服务器关闭钩子执行")
	return nil
}

// databaseShutdown 数据库关闭钩子
func (sm *ShutdownManager) databaseShutdown(ctx context.Context) error {
	// 这里应该关闭数据库连接
	sm.logger.Info("数据库关闭钩子执行")
	return nil
}

// cacheShutdown 缓存关闭钩子
func (sm *ShutdownManager) cacheShutdown(ctx context.Context) error {
	// 这里应该关闭缓存服务
	sm.logger.Info("缓存关闭钩子执行")
	return nil
}

// monitoringShutdown 监控关闭钩子
func (sm *ShutdownManager) monitoringShutdown(ctx context.Context) error {
	// 这里应该关闭监控服务
	sm.logger.Info("监控关闭钩子执行")
	return nil
}

// loggingShutdown 日志关闭钩子
func (sm *ShutdownManager) loggingShutdown(ctx context.Context) error {
	// 这里应该关闭日志服务
	sm.logger.Info("日志关闭钩子执行")
	return nil
}

// 全局关闭管理器实例
var globalShutdownManager *ShutdownManager

// InitGlobalShutdownManager 初始化全局关闭管理器
func InitGlobalShutdownManager(logger *zap.Logger, config *ShutdownConfig) {
	globalShutdownManager = NewShutdownManager(logger, config)
}

// GetGlobalShutdownManager 获取全局关闭管理器
func GetGlobalShutdownManager() *ShutdownManager {
	return globalShutdownManager
}

// AddGlobalHook 添加全局关闭钩子
func AddGlobalHook(name string, hook ShutdownHook, timeout time.Duration, priority int) {
	if globalShutdownManager != nil {
		globalShutdownManager.AddHook(name, hook, timeout, priority)
	}
}

// StartGlobalShutdown 启动全局关闭监听
func StartGlobalShutdown() {
	if globalShutdownManager != nil {
		globalShutdownManager.Start()
	}
}

// IsGlobalShuttingDown 检查全局是否正在关闭
func IsGlobalShuttingDown() bool {
	if globalShutdownManager != nil {
		return globalShutdownManager.IsShuttingDown()
	}
	return false
}

// GetGlobalShutdownContext 获取全局关闭上下文
func GetGlobalShutdownContext() context.Context {
	if globalShutdownManager != nil {
		return globalShutdownManager.GetContext()
	}
	return context.Background()
}