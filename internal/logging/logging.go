package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var logLevelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// LogEntry 日志条目
type LogEntry struct {
	ID         uint                   `json:"id" gorm:"primaryKey"`
	Timestamp  time.Time              `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	LevelName  string                 `json:"level_name"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Function   string                 `json:"function"`
	File       string                 `json:"file"`
	Line       int                    `json:"line"`
	UserID     *uint                  `json:"user_id,omitempty"`
	Username   string                 `json:"username,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   float64                `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Stacktrace string                 `json:"stacktrace,omitempty"`
	Tags       []string               `json:"tags,omitempty" gorm:"type:json"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" gorm:"type:json"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// LogFilter 日志过滤器
type LogFilter struct {
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Levels      []LogLevel `json:"levels,omitempty"`
	Services    []string   `json:"services,omitempty"`
	Users       []uint     `json:"users,omitempty"`
	Usernames   []string   `json:"usernames,omitempty"`
	RequestIDs  []string   `json:"request_ids,omitempty"`
	IPAddresses []string   `json:"ip_addresses,omitempty"`
	Methods     []string   `json:"methods,omitempty"`
	Paths       []string   `json:"paths,omitempty"`
	StatusCodes []int      `json:"status_codes,omitempty"`
	Keywords    []string   `json:"keywords,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
	SortBy      string     `json:"sort_by,omitempty"`
	SortOrder   string     `json:"sort_order,omitempty"`
}

// LogStats 日志统计
type LogStats struct {
	TotalCount    int64              `json:"total_count"`
	LevelCounts   map[LogLevel]int64 `json:"level_counts"`
	ServiceCounts map[string]int64   `json:"service_counts"`
	UserCounts    map[string]int64   `json:"user_counts"`
	IPCounts      map[string]int64   `json:"ip_counts"`
	MethodCounts  map[string]int64   `json:"method_counts"`
	ErrorCounts   map[string]int64   `json:"error_counts"`
	HourlyStats   map[string]int64   `json:"hourly_stats"`
	DailyStats    map[string]int64   `json:"daily_stats"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level         LogLevel `json:"level"`
	Format        string   `json:"format"` // json, text
	Output        string   `json:"output"` // console, file, both
	FilePath      string   `json:"file_path"`
	MaxSize       int      `json:"max_size"` // MB
	MaxBackups    int      `json:"max_backups"`
	MaxAge        int      `json:"max_age"` // days
	Compress      bool     `json:"compress"`
	EnableConsole bool     `json:"enable_console"`
	EnableFile    bool     `json:"enable_file"`
	EnableDB      bool     `json:"enable_db"`
	EnableES      bool     `json:"enable_es"`
}

// LoggingService 日志服务
type LoggingService struct {
	config    *config.Config
	db        *gorm.DB
	zapLogger *zap.Logger
	logConfig *LogConfig

	// 日志队列
	logQueue   chan *LogEntry
	bufferPool sync.Pool

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 正则缓存
	keywordRegex *regexp.Regexp
}

// NewLoggingService 创建日志服务
func NewLoggingService(cfg *config.Config, db *gorm.DB) (*LoggingService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 默认日志配置
	logConfig := &LogConfig{
		Level:         INFO,
		Format:        "json",
		Output:        "both",
		FilePath:      "logs/app.log",
		MaxSize:       100,
		MaxBackups:    10,
		MaxAge:        30,
		Compress:      true,
		EnableConsole: true,
		EnableFile:    true,
		EnableDB:      true,
		EnableES:      false,
	}

	// 创建Zap日志器
	zapLogger, err := createZapLogger(logConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("创建日志器失败: %w", err)
	}

	service := &LoggingService{
		config:    cfg,
		db:        db,
		zapLogger: zapLogger,
		logConfig: logConfig,
		logQueue:  make(chan *LogEntry, 10000),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 初始化缓冲池
	service.bufferPool.New = func() interface{} {
		return make([]byte, 0, 1024)
	}

	// 自动创建表
	if err := service.autoMigrate(); err != nil {
		zapLogger.Error("数据库迁移失败", zap.Error(err))
	}

	// 启动日志处理器
	service.startLogProcessor()

	return service, nil
}

// createZapLogger 创建Zap日志器
func createZapLogger(config *LogConfig) (*zap.Logger, error) {
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var core zapcore.Core

	// 创建多个核心
	var cores []zapcore.Core

	// 控制台输出
	if config.EnableConsole {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.InfoLevel
			}),
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if config.EnableFile {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.DebugLevel
			}),
		)
		cores = append(cores, fileCore)
	}

	// 组合核心
	if len(cores) > 0 {
		core = zapcore.NewTee(cores...)
	} else {
		core = zapcore.NewNopCore()
	}

	// 创建日志器
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger, nil
}

// startLogProcessor 启动日志处理器
func (s *LoggingService) startLogProcessor() {
	// 启动多个工作协程
	for i := 0; i < 3; i++ {
		s.wg.Add(1)
		go s.logWorker()
	}

	// 启动清理协程
	s.wg.Add(1)
	go s.cleanupWorker()
}

// logWorker 日志工作协程
func (s *LoggingService) logWorker() {
	defer s.wg.Done()

	for {
		select {
		case entry := <-s.logQueue:
			s.processLogEntry(entry)
		case <-s.ctx.Done():
			return
		}
	}
}

// cleanupWorker 清理工作协程
func (s *LoggingService) cleanupWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupOldLogs()
		case <-s.ctx.Done():
			return
		}
	}
}

// processLogEntry 处理日志条目
func (s *LoggingService) processLogEntry(entry *LogEntry) {
	// 保存到数据库
	if s.logConfig.EnableDB {
		go s.saveLogToDB(entry)
	}

	// 发送到Elasticsearch
	if s.logConfig.EnableES {
		go s.sendLogToES(entry)
	}
}

// saveLogToDB 保存日志到数据库
func (s *LoggingService) saveLogToDB(entry *LogEntry) {
	if err := s.db.Create(entry).Error; err != nil {
		s.zapLogger.Error("保存日志到数据库失败",
			zap.Error(err),
			zap.Uint("log_id", entry.ID),
			zap.String("message", entry.Message))
	}
}

// sendLogToES 发送日志到Elasticsearch
func (s *LoggingService) sendLogToES(entry *LogEntry) {
	// 实现发送到Elasticsearch的逻辑
	// 这里只是示例，需要根据实际的ES客户端实现
}

// cleanupOldLogs 清理旧日志
func (s *LoggingService) cleanupOldLogs() {
	// 清理30天前的日志
	cutoff := time.Now().AddDate(0, 0, -30)

	if err := s.db.Where("created_at < ?", cutoff).Delete(&LogEntry{}).Error; err != nil {
		s.zapLogger.Error("清理旧日志失败", zap.Error(err))
	}
}

// Log 记录日志
func (s *LoggingService) Log(level LogLevel, message string, fields ...zap.Field) {
	// 获取调用者信息
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	fn := runtime.FuncForPC(pc)
	function := "unknown"
	if fn != nil {
		function = fn.Name()
	}

	// 创建日志条目
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		LevelName: logLevelNames[level],
		Message:   message,
		Service:   "law-oa",
		Function:  function,
		File:      file,
		Line:      line,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 发送到Zap
	switch level {
	case DEBUG:
		s.zapLogger.Debug(message, fields...)
	case INFO:
		s.zapLogger.Info(message, fields...)
	case WARN:
		s.zapLogger.Warn(message, fields...)
	case ERROR:
		s.zapLogger.Error(message, fields...)
	case FATAL:
		s.zapLogger.Fatal(message, fields...)
	}

	// 异步处理
	select {
	case s.logQueue <- entry:
	default:
		s.zapLogger.Warn("日志队列已满，丢弃日志", zap.String("message", message))
	}
}

// Debug 记录调试日志
func (s *LoggingService) Debug(message string, fields ...zap.Field) {
	s.Log(DEBUG, message, fields...)
}

// Info 记录信息日志
func (s *LoggingService) Info(message string, fields ...zap.Field) {
	s.Log(INFO, message, fields...)
}

// Warn 记录警告日志
func (s *LoggingService) Warn(message string, fields ...zap.Field) {
	s.Log(WARN, message, fields...)
}

// Error 记录错误日志
func (s *LoggingService) Error(message string, fields ...zap.Field) {
	s.Log(ERROR, message, fields...)
}

// Fatal 记录致命日志
func (s *LoggingService) Fatal(message string, fields ...zap.Field) {
	s.Log(FATAL, message, fields...)
}

// WithContext 创建带上下文的日志器
func (s *LoggingService) WithContext(c *gin.Context) *ContextLogger {
	return &ContextLogger{
		service:   s,
		context:   c,
		requestID: c.GetString("request_id"),
		userID:    c.GetUint("user_id"),
		username:  c.GetString("username"),
		ipAddress: c.ClientIP(),
		userAgent: c.Request.UserAgent(),
		method:    c.Request.Method,
		path:      c.Request.URL.Path,
		startTime: time.Now(),
	}
}

// QueryLogs 查询日志
func (s *LoggingService) QueryLogs(filter LogFilter) ([]*LogEntry, int64, error) {
	query := s.db.Model(&LogEntry{})

	// 应用过滤器
	if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}

	if len(filter.Levels) > 0 {
		query = query.Where("level IN ?", filter.Levels)
	}

	if len(filter.Services) > 0 {
		query = query.Where("service IN ?", filter.Services)
	}

	if len(filter.Users) > 0 {
		query = query.Where("user_id IN ?", filter.Users)
	}

	if len(filter.Usernames) > 0 {
		query = query.Where("username IN ?", filter.Usernames)
	}

	if len(filter.RequestIDs) > 0 {
		query = query.Where("request_id IN ?", filter.RequestIDs)
	}

	if len(filter.IPAddresses) > 0 {
		query = query.Where("ip_address IN ?", filter.IPAddresses)
	}

	if len(filter.Methods) > 0 {
		query = query.Where("method IN ?", filter.Methods)
	}

	if len(filter.Paths) > 0 {
		query = query.Where("path IN ?", filter.Paths)
	}

	if len(filter.StatusCodes) > 0 {
		query = query.Where("status_code IN ?", filter.StatusCodes)
	}

	if len(filter.Keywords) > 0 {
		conditions := make([]string, 0, len(filter.Keywords))
		args := make([]interface{}, 0, len(filter.Keywords))
		for _, keyword := range filter.Keywords {
			conditions = append(conditions, "message LIKE ?")
			args = append(args, "%"+keyword+"%")
		}
		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	if len(filter.Tags) > 0 {
		// JSON查询支持PostgreSQL
		if s.db.Dialector.Name() == "postgres" {
			query = query.Where("tags::jsonb ?| ARRAY[?]", strings.Join(filter.Tags, ","))
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	if filter.SortBy != "" {
		order := filter.SortOrder
		if order == "" {
			order = "desc"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.SortBy, strings.ToUpper(order)))
	} else {
		query = query.Order("timestamp desc")
	}

	// 分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// 查询数据
	var logs []*LogEntry
	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLogStats 获取日志统计
func (s *LoggingService) GetLogStats(startTime, endTime time.Time) (*LogStats, error) {
	stats := &LogStats{
		LevelCounts:   make(map[LogLevel]int64),
		ServiceCounts: make(map[string]int64),
		UserCounts:    make(map[string]int64),
		IPCounts:      make(map[string]int64),
		MethodCounts:  make(map[string]int64),
		ErrorCounts:   make(map[string]int64),
		HourlyStats:   make(map[string]int64),
		DailyStats:    make(map[string]int64),
	}

	// 基础查询
	baseQuery := s.db.Model(&LogEntry{}).Where("timestamp BETWEEN ? AND ?", startTime, endTime)

	// 总数
	if err := baseQuery.Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	// 按级别统计
	var levelStats []struct {
		Level LogLevel
		Count int64
	}
	if err := baseQuery.Select("level, count(*) as count").Group("level").Scan(&levelStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range levelStats {
		stats.LevelCounts[stat.Level] = stat.Count
	}

	// 按服务统计
	var serviceStats []struct {
		Service string
		Count   int64
	}
	if err := baseQuery.Select("service, count(*) as count").Group("service").Scan(&serviceStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range serviceStats {
		stats.ServiceCounts[stat.Service] = stat.Count
	}

	// 按用户统计
	var userStats []struct {
		Username string
		Count    int64
	}
	if err := baseQuery.Select("username, count(*) as count").Where("username != ''").Group("username").Scan(&userStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range userStats {
		stats.UserCounts[stat.Username] = stat.Count
	}

	// 按IP统计
	var ipStats []struct {
		IPAddress string
		Count     int64
	}
	if err := baseQuery.Select("ip_address, count(*) as count").Where("ip_address != ''").Group("ip_address").Scan(&ipStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range ipStats {
		stats.IPCounts[stat.IPAddress] = stat.Count
	}

	// 按方法统计
	var methodStats []struct {
		Method string
		Count  int64
	}
	if err := baseQuery.Select("method, count(*) as count").Where("method != ''").Group("method").Scan(&methodStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range methodStats {
		stats.MethodCounts[stat.Method] = stat.Count
	}

	// 按错误统计
	var errorStats []struct {
		Error string
		Count int64
	}
	if err := baseQuery.Select("error, count(*) as count").Where("error != '' AND level = ?", ERROR).Group("error").Scan(&errorStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range errorStats {
		stats.ErrorCounts[stat.Error] = stat.Count
	}

	// 按小时统计
	var hourlyStats []struct {
		Hour  string
		Count int64
	}
	if err := baseQuery.Select("DATE_FORMAT(timestamp, '%Y-%m-%d %H:00:00') as hour, count(*) as count").Group("hour").Scan(&hourlyStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range hourlyStats {
		stats.HourlyStats[stat.Hour] = stat.Count
	}

	// 按天统计
	var dailyStats []struct {
		Day   string
		Count int64
	}
	if err := baseQuery.Select("DATE_FORMAT(timestamp, '%Y-%m-%d') as day, count(*) as count").Group("day").Scan(&dailyStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range dailyStats {
		stats.DailyStats[stat.Day] = stat.Count
	}

	return stats, nil
}

// autoMigrate 自动迁移
func (s *LoggingService) autoMigrate() error {
	return s.db.AutoMigrate(&LogEntry{})
}

// Close 关闭日志服务
func (s *LoggingService) Close() {
	s.cancel()
	s.wg.Wait()

	if err := s.zapLogger.Sync(); err != nil {
		log.Printf("同步日志器失败: %v", err)
	}
}

// ContextLogger 上下文日志器
type ContextLogger struct {
	service   *LoggingService
	context   *gin.Context
	requestID string
	userID    uint
	username  string
	ipAddress string
	userAgent string
	method    string
	path      string
	startTime time.Time
	fields    []zap.Field
}

// WithField 添加字段
func (cl *ContextLogger) WithField(key string, value interface{}) *ContextLogger {
	cl.fields = append(cl.fields, zap.Any(key, value))
	return cl
}

// WithFields 添加多个字段
func (cl *ContextLogger) WithFields(fields map[string]interface{}) *ContextLogger {
	for k, v := range fields {
		cl.fields = append(cl.fields, zap.Any(k, v))
	}
	return cl
}

// WithError 添加错误
func (cl *ContextLogger) WithError(err error) *ContextLogger {
	cl.fields = append(cl.fields, zap.Error(err))
	return cl
}

// WithUser 添加用户信息
func (cl *ContextLogger) WithUser(userID uint, username string) *ContextLogger {
	cl.userID = userID
	cl.username = username
	return cl
}

// WithRequestID 添加请求ID
func (cl *ContextLogger) WithRequestID(requestID string) *ContextLogger {
	cl.requestID = requestID
	return cl
}

// WithTags 添加标签
func (cl *ContextLogger) WithTags(tags ...string) *ContextLogger {
	cl.fields = append(cl.fields, zap.Strings("tags", tags))
	return cl
}

// Debug 记录调试日志
func (cl *ContextLogger) Debug(message string) {
	cl.log(DEBUG, message)
}

// Info 记录信息日志
func (cl *ContextLogger) Info(message string) {
	cl.log(INFO, message)
}

// Warn 记录警告日志
func (cl *ContextLogger) Warn(message string) {
	cl.log(WARN, message)
}

// Error 记录错误日志
func (cl *ContextLogger) Error(message string) {
	cl.log(ERROR, message)
}

// Fatal 记录致命日志
func (cl *ContextLogger) Fatal(message string) {
	cl.log(FATAL, message)
}

// log 记录日志
func (cl *ContextLogger) log(level LogLevel, message string) {
	// 获取调用者信息
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	fn := runtime.FuncForPC(pc)
	function := "unknown"
	if fn != nil {
		function = fn.Name()
	}

	// 创建日志条目
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		LevelName: logLevelNames[level],
		Message:   message,
		Service:   "law-oa",
		Function:  function,
		File:      file,
		Line:      line,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置上下文字段
	if cl.requestID != "" {
		entry.RequestID = cl.requestID
	}
	if cl.userID > 0 {
		entry.UserID = &cl.userID
	}
	if cl.username != "" {
		entry.Username = cl.username
	}
	if cl.ipAddress != "" {
		entry.IPAddress = cl.ipAddress
	}
	if cl.userAgent != "" {
		entry.UserAgent = cl.userAgent
	}
	if cl.method != "" {
		entry.Method = cl.method
	}
	if cl.path != "" {
		entry.Path = cl.path
	}

	// 计算请求持续时间
	if !cl.startTime.IsZero() {
		entry.Duration = time.Since(cl.startTime).Seconds()
	}

	// 从上下文获取状态码
	if cl.context != nil {
		entry.StatusCode = cl.context.Writer.Status()
	}

	// 发送到Zap
	cl.service.zapLogger.Info(message, cl.fields...)

	// 异步处理
	select {
	case cl.service.logQueue <- entry:
	default:
		cl.service.zapLogger.Warn("日志队列已满，丢弃日志", zap.String("message", message))
	}
}

// LoggingMiddleware 日志中间件
func (s *LoggingService) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("request_id", requestID)
		}

		// 创建上下文日志器
		logger := s.WithContext(c)
		c.Set("logger", logger)

		// 记录请求开始
		logger.WithFields(map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"user_agent": c.Request.UserAgent(),
			"ip":         c.ClientIP(),
		}).Info("请求开始")

		// 处理请求
		c.Next()

		// 记录请求结束
		duration := time.Since(start)
		status := c.Writer.Status()

		fields := map[string]interface{}{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   status,
			"duration": duration.Seconds(),
			"size":     c.Writer.Size(),
		}

		if len(c.Errors) > 0 {
			fields["error"] = c.Errors.String()
		}

		if status >= 400 {
			logger.WithFields(fields).Error("请求失败")
		} else {
			logger.WithFields(fields).Info("请求完成")
		}
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GlobalLogger 全局日志器
var GlobalLogger *LoggingService

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger(cfg *config.Config, db *gorm.DB) error {
	logger, err := NewLoggingService(cfg, db)
	if err != nil {
		return err
	}
	GlobalLogger = logger
	return nil
}

// Log 全局日志函数
func Log(level LogLevel, message string, fields ...zap.Field) {
	if GlobalLogger != nil {
		GlobalLogger.Log(level, message, fields...)
	}
}

// Debug 全局调试日志
func Debug(message string, fields ...zap.Field) {
	Log(DEBUG, message, fields...)
}

// Info 全局信息日志
func Info(message string, fields ...zap.Field) {
	Log(INFO, message, fields...)
}

// Warn 全局警告日志
func Warn(message string, fields ...zap.Field) {
	Log(WARN, message, fields...)
}

// Error 全局错误日志
func Error(message string, fields ...zap.Field) {
	Log(ERROR, message, fields...)
}

// Fatal 全局致命日志
func Fatal(message string, fields ...zap.Field) {
	Log(FATAL, message, fields...)
}
