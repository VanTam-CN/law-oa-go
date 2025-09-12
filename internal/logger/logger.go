package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
	Level      string `json:"level" yaml:"level"`           // 日志级别
	Format     string `json:"format" yaml:"format"`         // 格式：json 或 console
	Output     string `json:"output" yaml:"output"`         // 输出：stdout 或文件路径
	MaxSize    int    `json:"maxSize" yaml:"maxSize"`       // 单个文件最大大小(MB)
	MaxBackups int    `json:"maxBackups" yaml:"maxBackups"` // 保留的旧文件数量
	MaxAge     int    `json:"maxAge" yaml:"maxAge"`         // 保留天数
	Compress   bool   `json:"compress" yaml:"compress"`     // 是否压缩
}

// Init 初始化日志系统
func Init(cfg *Config) error {
	// 设置日志级别
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// 设置编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 设置输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "stdout" || cfg.Output == "" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		// 文件输出，支持日志轮转
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Output,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = Logger.Sugar()

	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
}

// WithContext 带上下文的日志记录器
func WithContext(ctx context.Context) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}

	// 从上下文中提取请求ID等信息
	if requestID := ctx.Value("request_id"); requestID != nil {
		return Logger.With(zap.String("request_id", requestID.(string)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		return Logger.With(zap.Any("user_id", userID))
	}

	return Logger
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
