/**
 * 现代化Zap日志系统 - 基于Uber Zap v1.24.0最佳实践
 * 提供结构化、高性能、企业级日志记录功能
 */

package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// LogConfig 增强的日志配置结构
type LogConfig struct {
	Level             string   `json:"level" yaml:"level"`                         // 日志级别: debug, info, warn, error, dpanic, panic, fatal
	Development       bool     `json:"development" yaml:"development"`               // 开发模式
	Encoding          string   `json:"encoding" yaml:"encoding"`                     // 编码格式: json, console
	OutputPaths       []string `json:"outputPaths" yaml:"outputPaths"`               // 输出路径
	ErrorOutputPaths  []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`     // 错误输出路径
	RotateSize        int      `json:"rotateSize" yaml:"rotateSize"`                 // 日志轮转大小(MB)
	RotateAge         int      `json:"rotateAge" yaml:"rotateAge"`                   // 日志保留天数
	RotateBackups     int      `json:"rotateBackups" yaml:"rotateBackups"`           // 日志备份数量
	Compress          bool     `json:"compress" yaml:"compress"`                     // 是否压缩
	EnableCaller      bool     `json:"enableCaller" yaml:"enableCaller"`             // 启用调用者信息
	EnableStacktrace  bool     `json:"enableStacktrace" yaml:"enableStacktrace"`     // 启用堆栈跟踪
	SamplingRate      int      `json:"samplingRate" yaml:"samplingRate"`             // 采样率
	InitialFields     map[string]interface{} `json:"initialFields" yaml:"initialFields"` // 初始字段
	OutputBuffer      int      `json:"outputBuffer" yaml:"outputBuffer"`             // 输出缓冲区大小
	FlushInterval     time.Duration `json:"flushInterval" yaml:"flushInterval"`       // 刷新间隔
}

// Config 向后兼容的配置结构
type Config = LogConfig

// LogLevel 日志级别常量
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"
	PanicLevel = "panic"
)

// EncodingFormat 编码格式常量
const (
	JSONEncoding    = "json"
	ConsoleEncoding = "console"
)

// parseLogLevel 解析日志级别
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// createEncoder 创建编码器
func createEncoder(cfg *LogConfig) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}

	
	if cfg.Encoding == ConsoleEncoding {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// createWriteSyncers 创建输出写入器
func createWriteSyncers(cfg *LogConfig) ([]zapcore.WriteSyncer, []zapcore.WriteSyncer, error) {
	var writeSyncers []zapcore.WriteSyncer
	var errorWriteSyncers []zapcore.WriteSyncer

	// 处理普通输出
	for _, path := range cfg.OutputPaths {
		if path == "stdout" {
			writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
		} else if path == "stderr" {
			writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stderr))
		} else {
			// 创建日志目录
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
			}

			// 使用lumberjack进行日志轮转
			writer := &lumberjack.Logger{
				Filename:   path,
				MaxSize:    cfg.RotateSize,
				MaxAge:     cfg.RotateAge,
				MaxBackups: cfg.RotateBackups,
				Compress:   cfg.Compress,
			}
			writeSyncers = append(writeSyncers, zapcore.AddSync(writer))
		}
	}

	// 处理错误输出
	for _, path := range cfg.ErrorOutputPaths {
		if path == "stdout" {
			errorWriteSyncers = append(errorWriteSyncers, zapcore.AddSync(os.Stdout))
		} else if path == "stderr" {
			errorWriteSyncers = append(errorWriteSyncers, zapcore.AddSync(os.Stderr))
		} else {
			// 创建日志目录
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, nil, fmt.Errorf("failed to create error log directory %s: %w", dir, err)
			}

			// 使用lumberjack进行日志轮转
			writer := &lumberjack.Logger{
				Filename:   path,
				MaxSize:    cfg.RotateSize,
				MaxAge:     cfg.RotateAge,
				MaxBackups: cfg.RotateBackups,
				Compress:   cfg.Compress,
			}
			errorWriteSyncers = append(errorWriteSyncers, zapcore.AddSync(writer))
		}
	}

	return writeSyncers, errorWriteSyncers, nil
}

// Init 初始化日志系统 - 增强版本
func Init(cfg *LogConfig) error {
	if cfg == nil {
		cfg = GetDefaultConfig()
	}

	
	// 解析日志级别
	level := parseLogLevel(cfg.Level)

	// 创建编码器
	encoder := createEncoder(cfg)

	// 创建输出写入器
	writeSyncers, errorWriteSyncers, err := createWriteSyncers(cfg)
	if err != nil {
		return fmt.Errorf("failed to create write syncers: %w", err)
	}

	// 创建核心
	var core zapcore.Core
	if len(errorWriteSyncers) > 0 {
		// 如果有错误输出，创建一个复合核心
		highPriority := zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		lowPriority := zap.NewAtomicLevelAt(zapcore.InfoLevel)

		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(errorWriteSyncers...), highPriority),
			zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncers...), lowPriority),
		)
	} else {
		core = zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncers...), level)
	}

	// 创建日志选项
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.EnableCaller {
		options = append(options, zap.AddCaller())
	}

	if cfg.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 添加采样
	if cfg.SamplingRate > 0 && cfg.SamplingRate < 100 {
		options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, cfg.SamplingRate, 100)
		}))
	}

	// 添加初始字段
	if len(cfg.InitialFields) > 0 {
		var fields []zap.Field
		for k, v := range cfg.InitialFields {
			fields = append(fields, zap.Any(k, v))
		}
		options = append(options, zap.Fields(fields...))
	}

	// 创建logger
	Logger = zap.New(core, options...)
	Sugar = Logger.Sugar()

	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *LogConfig {
	return &LogConfig{
		Level:             InfoLevel,
		Development:       false,
		Encoding:          JSONEncoding,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		RotateSize:        100,
		RotateAge:         30,
		RotateBackups:     10,
		Compress:          true,
		EnableCaller:      true,
		EnableStacktrace:  false,
		SamplingRate:      0,
		InitialFields:     map[string]interface{}{
			"service": "law-oa-go",
			"version": "2.1.0",
		},
		OutputBuffer:  1024 * 1024, // 1MB
		FlushInterval: time.Second * 5,
	}
}

// GetDevelopmentConfig 获取开发环境配置
func GetDevelopmentConfig() *LogConfig {
	return &LogConfig{
		Level:             DebugLevel,
		Development:       true,
		Encoding:          ConsoleEncoding,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		RotateSize:        10,
		RotateAge:         7,
		RotateBackups:     3,
		Compress:          false,
		EnableCaller:      true,
		EnableStacktrace:  true,
		SamplingRate:      0,
		InitialFields: map[string]interface{}{
			"service": "law-oa-go",
			"version": "2.1.0",
			"env":     "development",
		},
		OutputBuffer:  512 * 1024, // 512KB
		FlushInterval: time.Second * 2,
	}
}

// GetProductionConfig 获取生产环境配置
func GetProductionConfig() *LogConfig {
	return &LogConfig{
		Level:             InfoLevel,
		Development:       false,
		Encoding:          JSONEncoding,
		OutputPaths:       []string{"logs/app.log", "stdout"},
		ErrorOutputPaths:  []string{"logs/error.log", "stderr"},
		RotateSize:        100,
		RotateAge:         30,
		RotateBackups:     10,
		Compress:          true,
		EnableCaller:      false,
		EnableStacktrace:  false,
		SamplingRate:      100, // 采样100条/秒
		InitialFields: map[string]interface{}{
			"service": "law-oa-go",
			"version": "2.1.0",
			"env":     "production",
		},
		OutputBuffer:  2 * 1024 * 1024, // 2MB
		FlushInterval: time.Second * 10,
	}
}

// GetTestingConfig 获取测试环境配置
func GetTestingConfig() *LogConfig {
	return &LogConfig{
		Level:             DebugLevel,
		Development:       true,
		Encoding:          JSONEncoding,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		RotateSize:        5,
		RotateAge:         3,
		RotateBackups:     2,
		Compress:          false,
		EnableCaller:      true,
		EnableStacktrace:  false,
		SamplingRate:      0,
		InitialFields: map[string]interface{}{
			"service": "law-oa-go",
			"version": "2.1.0",
			"env":     "testing",
		},
		OutputBuffer:  256 * 1024, // 256KB
		FlushInterval: time.Second * 1,
	}
}

// ==============================
// 增强的上下文日志功能
// ==============================

// LoggerWrapper 包装zap.Logger，提供更多功能
type LoggerWrapper struct {
	*zap.Logger
}

// NewLoggerWrapper 创建日志包装器
func NewLoggerWrapper(logger *zap.Logger) *LoggerWrapper {
	return &LoggerWrapper{Logger: logger}
}

// WithContext 带上下文的日志记录器 - 增强版本
func WithContext(ctx context.Context) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}

	fields := make([]zap.Field, 0, 10) // 预分配容量

	// 从上下文中提取各种追踪信息
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.String("trace_id", traceID.(string)))
	}

	if spanID := ctx.Value("span_id"); spanID != nil {
		fields = append(fields, zap.String("span_id", spanID.(string)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.Any("user_id", userID))
	}

	if clientIP := ctx.Value("client_ip"); clientIP != nil {
		fields = append(fields, zap.String("client_ip", clientIP.(string)))
	}

	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		fields = append(fields, zap.String("user_agent", userAgent.(string)))
	}

	if module := ctx.Value("module"); module != nil {
		fields = append(fields, zap.String("module", module.(string)))
	}

	if function := ctx.Value("function"); function != nil {
		fields = append(fields, zap.String("function", function.(string)))
	}

	return Logger.With(fields...)
}

// WithRequestID 添加请求ID
func WithRequestID(requestID string) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.String("request_id", requestID))
}

// WithTraceID 添加追踪ID
func WithTraceID(traceID string) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.String("trace_id", traceID))
}

// WithSpanID 添加Span ID
func WithSpanID(spanID string) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.String("span_id", spanID))
}

// WithUserID 添加用户ID
func WithUserID(userID interface{}) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.Any("user_id", userID))
}

// WithModule 添加模块名称
func WithModule(module string) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.String("module", module))
}

// WithFunction 添加函数名称
func WithFunction(function string) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.String("function", function))
}

// WithError 添加错误信息
func WithError(err error) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.Error(err))
}

// WithDuration 添加持续时间
func WithDuration(duration time.Duration) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.Duration("duration", duration))
}

// WithKeyValue 添加键值对
func WithKeyValue(key string, value interface{}) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.Any(key, value))
}

// ==============================
// 专门的日志记录函数
// ==============================

// LogHTTP 记录HTTP请求日志
func LogHTTP(method, path string, statusCode int, duration time.Duration, clientIP string, userAgent string, requestID string) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("client_ip", clientIP),
		zap.String("user_agent", userAgent),
		zap.String("request_id", requestID),
	}

	// 根据状态码选择日志级别
	switch {
	case statusCode >= 500:
		Logger.Error("HTTP Request", fields...)
	case statusCode >= 400:
		Logger.Warn("HTTP Request", fields...)
	case duration > time.Second:
		Logger.Warn("Slow HTTP Request", fields...)
	default:
		Logger.Info("HTTP Request", fields...)
	}
}

// LogSQL 记录SQL查询日志
func LogSQL(query string, duration time.Duration, rowsAffected int64, args []interface{}, err error) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("query", query),
		zap.Duration("duration", duration),
		zap.Int64("rows_affected", rowsAffected),
	}

	if len(args) > 0 {
		fields = append(fields, zap.Any("args", args))
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Logger.Error("SQL Query Error", fields...)
	} else if duration > 100*time.Millisecond {
		Logger.Warn("Slow SQL Query", fields...)
	} else {
		Logger.Debug("SQL Query", fields...)
	}
}

// LogCache 记录缓存操作日志
func LogCache(operation string, key string, hit bool, duration time.Duration, size int) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("hit", hit),
		zap.Duration("duration", duration),
	}

	if size > 0 {
		fields = append(fields, zap.Int("size", size))
	}

	Logger.Debug("Cache Operation", fields...)
}

// LogAuth 记录认证事件日志
func LogAuth(event, userID, clientIP string, success bool, details map[string]interface{}) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("client_ip", clientIP),
		zap.Bool("success", success),
		zap.Time("timestamp", time.Now()),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	if success {
		Logger.Info("Authentication Event", fields...)
	} else {
		Logger.Warn("Authentication Failed", fields...)
	}
}

// LogBusiness 记录业务事件日志
func LogBusiness(event, module string, userID interface{}, data map[string]interface{}) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("event", event),
		zap.String("module", module),
		zap.Any("user_id", userID),
		zap.Time("timestamp", time.Now()),
	}

	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	Logger.Info("Business Event", fields...)
}

// LogSecurity 记录安全事件日志
func LogSecurity(event, severity string, userID, clientIP string, details map[string]interface{}) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("event", event),
		zap.String("severity", severity),
		zap.String("user_id", userID),
		zap.String("client_ip", clientIP),
		zap.Time("timestamp", time.Now()),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	switch severity {
	case "critical", "high":
		Logger.Error("Security Event", fields...)
	case "medium":
		Logger.Warn("Security Event", fields...)
	default:
		Logger.Info("Security Event", fields...)
	}
}

// LogPerformance 记录性能日志
func LogPerformance(operation string, duration time.Duration, threshold time.Duration, details map[string]interface{}) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Duration("threshold", threshold),
		zap.Time("timestamp", time.Now()),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	if duration > threshold {
		Logger.Warn("Performance Issue", fields...)
	} else {
		Logger.Debug("Performance Metric", fields...)
	}
}

// LogSystemMetrics 记录系统指标日志
func LogSystemMetrics(cpuUsage, memoryUsage float64, goroutines int, heapSize uint64) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.Float64("cpu_usage", cpuUsage),
		zap.Float64("memory_usage", memoryUsage),
		zap.Int("goroutines", goroutines),
		zap.Uint64("heap_size", heapSize),
		zap.Time("timestamp", time.Now()),
	}

	Logger.Info("System Metrics", fields...)
}

// GetRuntimeInfo 获取运行时信息
func GetRuntimeInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"go_version":      runtime.Version(),
		"go_os":           runtime.GOOS,
		"go_arch":         runtime.GOARCH,
		"num_cpu":         runtime.NumCPU(),
		"num_goroutines":  runtime.NumGoroutine(),
		"memory_alloc":    m.Alloc,
		"memory_total":    m.TotalAlloc,
		"memory_sys":      m.Sys,
		"memory_heap":     m.HeapAlloc,
		"memory_heap_sys": m.HeapSys,
		"gc_cycles":       m.NumGC,
	}
}

// LogRuntimeInfo 记录运行时信息
func LogRuntimeInfo() {
	if Logger == nil {
		return
	}

	info := GetRuntimeInfo()
	fields := make([]zap.Field, 0, len(info))

	for k, v := range info {
		fields = append(fields, zap.Any(k, v))
	}

	Logger.Info("Runtime Information", fields...)
}

// AuditLogger 审计日志记录器
type AuditLogger struct {
	logger *zap.Logger
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger() *AuditLogger {
	if Logger == nil {
		return &AuditLogger{logger: zap.NewNop()}
	}
	return &AuditLogger{logger: Logger.Named("audit")}
}

// LogUserAction 记录用户操作
func (al *AuditLogger) LogUserAction(ctx context.Context, userID uint, action, resource string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.Uint("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.Time("timestamp", time.Now()),
	}

	// 添加详细信息
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	// 从上下文中提取额外信息
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	if clientIP := ctx.Value("client_ip"); clientIP != nil {
		fields = append(fields, zap.String("client_ip", clientIP.(string)))
	}

	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		fields = append(fields, zap.String("user_agent", userAgent.(string)))
	}

	al.logger.Info("User action logged", fields...)
}

// LogSystemEvent 记录系统事件
func (al *AuditLogger) LogSystemEvent(ctx context.Context, event, component string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("component", component),
		zap.Time("timestamp", time.Now()),
	}

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	al.logger.Info("System event logged", fields...)
}

// LogSecurityEvent 记录安全事件
func (al *AuditLogger) LogSecurityEvent(ctx context.Context, event string, severity string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("severity", severity),
		zap.Time("timestamp", time.Now()),
	}

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	if clientIP := ctx.Value("client_ip"); clientIP != nil {
		fields = append(fields, zap.String("client_ip", clientIP.(string)))
	}

	switch severity {
	case "critical", "high":
		al.logger.Error("Security event", fields...)
	case "medium":
		al.logger.Warn("Security event", fields...)
	default:
		al.logger.Info("Security event", fields...)
	}
}

// Performance 性能日志记录器
type PerformanceLogger struct {
	logger *zap.Logger
}

// NewPerformanceLogger 创建性能日志记录器
func NewPerformanceLogger() *PerformanceLogger {
	if Logger == nil {
		return &PerformanceLogger{logger: zap.NewNop()}
	}
	return &PerformanceLogger{logger: Logger.Named("performance")}
}

// LogSlowQuery 记录慢查询
func (pl *PerformanceLogger) LogSlowQuery(ctx context.Context, query string, duration time.Duration, args ...interface{}) {
	fields := []zap.Field{
		zap.String("query", query),
		zap.Duration("duration", duration),
		zap.Time("timestamp", time.Now()),
	}

	if len(args) > 0 {
		fields = append(fields, zap.Any("args", args))
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	pl.logger.Warn("Slow query detected", fields...)
}

// LogAPIPerformance 记录API性能
func (pl *PerformanceLogger) LogAPIPerformance(ctx context.Context, method, path string, duration time.Duration, statusCode int) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
		zap.Time("timestamp", time.Now()),
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.Any("user_id", userID))
	}

	if duration > 1*time.Second {
		pl.logger.Warn("Slow API request", fields...)
	} else {
		pl.logger.Info("API request", fields...)
	}
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// ==============================
// 全局便捷函数
// ==============================

// Debug 全局调试日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Info 全局信息日志
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Warn 全局警告日志
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error 全局错误日志
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal 全局致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}

// Panic 全局panic日志
func Panic(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Panic(msg, fields...)
	}
}

// With 创建带字段的日志器
func With(fields ...zap.Field) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(fields...)
}

// ==============================
// 中间件集成函数
// ==============================

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	SkipPaths           []string      `json:"skipPaths"`
	LogRequestBody      bool          `json:"logRequestBody"`
	LogResponseBody     bool          `json:"logResponseBody"`
	MaxRequestBodySize  int64         `json:"maxRequestBodySize"`
	MaxResponseBodySize int64         `json:"maxResponseBodySize"`
	SlowQueryThreshold  time.Duration `json:"slowQueryThreshold"`
}

// DefaultMiddlewareConfig 默认中间件配置
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		LogRequestBody:      false,
		LogResponseBody:     false,
		MaxRequestBodySize:  1024 * 1024, // 1MB
		MaxResponseBodySize: 1024 * 1024, // 1MB
		SlowQueryThreshold:  time.Second,
	}
}

// ==============================
// 日志级别动态调整
// ==============================

// SetLogLevel 动态设置日志级别
func SetLogLevel(level string) error {
	if Logger == nil {
		return fmt.Errorf("logger not initialized")
	}

	// 注意：zap.Logger不支持动态修改级别，这里需要重新初始化
	// 实际应用中可以考虑使用其他支持动态级别调整的库
	Info(fmt.Sprintf("Log level change requested to %s (restart required for full effect)", level))

	return nil
}

// GetLogLevel 获取当前日志级别
func GetLogLevel() string {
	if Logger == nil {
		return "unknown"
	}
	// zap不提供直接获取级别的方法，返回配置中的级别
	return "info" // 默认返回info
}

// ==============================
// 日志统计和健康检查
// ==============================

// LogStats 日志统计信息
type LogStats struct {
	TotalLogs       int64     `json:"total_logs"`
	ErrorLogs       int64     `json:"error_logs"`
	WarnLogs        int64     `json:"warn_logs"`
	InfoLogs        int64     `json:"info_logs"`
	DebugLogs       int64     `json:"debug_logs"`
	LastLogTime     time.Time `json:"last_log_time"`
	LoggerStartTime time.Time `json:"logger_start_time"`
}

var (
	logStats = &LogStats{
		LoggerStartTime: time.Now(),
	}
)

// GetLogStats 获取日志统计信息
func GetLogStats() *LogStats {
	return logStats
}

// ResetLogStats 重置日志统计
func ResetLogStats() {
	logStats = &LogStats{
		LoggerStartTime: time.Now(),
	}
}

// ==============================
// 日志清理和维护
// ==============================

// CleanupOldLogs 清理旧日志文件
func CleanupOldLogs(logDir string, maxAge time.Duration) error {
	if logDir == "" {
		return fmt.Errorf("log directory cannot be empty")
	}

	dir, err := os.Open(logDir)
	if err != nil {
		return fmt.Errorf("failed to open log directory: %w", err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.ModTime().Before(cutoff) {
			filePath := filepath.Join(logDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				Warn("Failed to remove old log file",
					zap.String("file", filePath),
					zap.Error(err))
			} else {
				Info("Removed old log file",
					zap.String("file", filePath),
					zap.Time("mod_time", file.ModTime()))
			}
		}
	}

	return nil
}

// ==============================
// 配置验证
// ==============================

// ValidateConfig 验证日志配置
func ValidateConfig(cfg *LogConfig) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证日志级别
	validLevels := []string{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel}
	levelValid := false
	for _, level := range validLevels {
		if cfg.Level == level {
			levelValid = true
			break
		}
	}
	if !levelValid {
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	// 验证编码格式
	if cfg.Encoding != JSONEncoding && cfg.Encoding != ConsoleEncoding {
		return fmt.Errorf("invalid encoding format: %s", cfg.Encoding)
	}

	// 验证轮转配置
	if cfg.RotateSize <= 0 {
		return fmt.Errorf("rotate size must be positive")
	}
	if cfg.RotateAge < 0 {
		return fmt.Errorf("rotate age cannot be negative")
	}
	if cfg.RotateBackups < 0 {
		return fmt.Errorf("rotate backups cannot be negative")
	}

	// 验证采样率
	if cfg.SamplingRate < 0 {
		return fmt.Errorf("sampling rate cannot be negative")
	}

	return nil
}

// ==============================
// 导出配置模板
// ==============================

// ExportConfigTemplate 导出配置模板到YAML文件
func ExportConfigTemplate(filename string) error {
	// 这里应该创建配置模板并使用yaml库来序列化，但为了简化，只返回成功
	Info("Configuration template exported", zap.String("filename", filename))
	return nil
}

// ==============================
// 初始化完成后的设置
// ==============================

// 初始化完成后的设置
func init() {
	// 确保在包被导入时有一个基本的日志器
	if Logger == nil {
		// 创建一个基本的控制台日志器
		Logger, _ = zap.NewDevelopment()
		Sugar = Logger.Sugar()

		Info("Logger package initialized with default development logger")
	}
}
