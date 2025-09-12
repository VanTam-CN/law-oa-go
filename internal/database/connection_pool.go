package database

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/config"
)

// 数据库连接池相关的Prometheus指标
var (
	dbConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "db_connections_active",
		Help: "Number of active database connections",
	}, []string{"database"})

	dbConnectionWaitTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_connection_wait_seconds",
		Help:    "Time spent waiting for database connections",
		Buckets: prometheus.DefBuckets,
	}, []string{"database"})

	dbConnectionErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_connection_errors_total",
		Help: "Total number of database connection errors",
	}, []string{"database"})
)

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
	SlowThreshold   time.Duration // 慢查询阈值
}

// OptimizedDatabase 优化后的数据库连接
type OptimizedDatabase struct {
	DB     *gorm.DB
	Config *ConnectionPoolConfig
}

// NewOptimizedDatabase 创建优化后的数据库连接
func NewOptimizedDatabase(cfg *config.Config) (*OptimizedDatabase, error) {
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%v&loc=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.Charset,
		cfg.Database.ParseTime,
		cfg.Database.Loc,
	)

	// 配置连接池
	poolConfig := &ConnectionPoolConfig{
		MaxOpenConns:    100,                    // 最大打开连接数
		MaxIdleConns:    10,                     // 最大空闲连接数
		ConnMaxLifetime: 30 * time.Minute,       // 连接最大生命周期
		ConnMaxIdleTime: 5 * time.Minute,        // 连接最大空闲时间
		SlowThreshold:   100 * time.Millisecond, // 慢查询阈值
	}

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.New(
			&SlowQueryLogger{threshold: poolConfig.SlowThreshold},
			logger.Config{
				SlowThreshold: poolConfig.SlowThreshold,
				LogLevel:      logger.Warn,
				Colorful:      false,
			},
		),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		dbConnectionErrors.WithLabelValues(cfg.Database.Database).Inc()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)

	// 配置连接池监控
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := sqlDB.Stats()
			dbConnections.WithLabelValues(cfg.Database.Database).Set(float64(stats.OpenConnections))
		}
	}()

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		dbConnectionErrors.WithLabelValues(cfg.Database.Database).Inc()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &OptimizedDatabase{
		DB:     db,
		Config: poolConfig,
	}, nil
}

// SlowQueryLogger 慢查询日志记录器
type SlowQueryLogger struct {
	threshold time.Duration
}

func (l *SlowQueryLogger) Printf(format string, v ...interface{}) {
	// 这里可以实现自定义的慢查询日志记录逻辑
	// 例如发送到监控系统或日志聚合系统
	fmt.Printf("[SLOW QUERY] "+format, v...)
}

// Health 健康检查
func (od *OptimizedDatabase) Health() error {
	sqlDB, err := od.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats 获取连接池统计信息
func (od *OptimizedDatabase) Stats() map[string]interface{} {
	sqlDB, err := od.DB.DB()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_conns":      stats.MaxOpenConnections,
		"open_conns":          stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// Close 关闭数据库连接
func (od *OptimizedDatabase) Close() error {
	sqlDB, err := od.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WithRetry 带重试的数据库操作
func (od *OptimizedDatabase) WithRetry(maxRetries int, operation func() error) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
				continue
			}
		} else {
			return nil
		}
	}

	return lastErr
}

// TransactionWithRetry 带重试的事务操作
func (od *OptimizedDatabase) TransactionWithRetry(maxRetries int, fn func(*gorm.DB) error) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := od.DB.Transaction(fn); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
				continue
			}
		} else {
			return nil
		}
	}

	return lastErr
}

// OptimizedExecute 优化执行SQL
func (od *OptimizedDatabase) OptimizedExecute(sql string, args ...interface{}) error {
	start := time.Now()

	err := od.WithRetry(3, func() error {
		return od.DB.Exec(sql, args...).Error
	})

	duration := time.Since(start)
	if duration > od.Config.SlowThreshold {
		fmt.Printf("[SLOW QUERY] %s took %v\n", sql, duration)
	}

	return err
}

// GetStatsForPrometheus 获取Prometheus格式的统计信息
func (od *OptimizedDatabase) GetStatsForPrometheus() map[string]float64 {
	sqlDB, err := od.DB.DB()
	if err != nil {
		return map[string]float64{"error": 1}
	}
	stats := sqlDB.Stats()

	return map[string]float64{
		"open_connections":    float64(stats.OpenConnections),
		"in_use":              float64(stats.InUse),
		"idle":                float64(stats.Idle),
		"wait_count":          float64(stats.WaitCount),
		"wait_seconds":        float64(stats.WaitDuration) / float64(time.Second),
		"max_idle_closed":     float64(stats.MaxIdleClosed),
		"max_lifetime_closed": float64(stats.MaxLifetimeClosed),
	}
}

// 全局优化数据库实例
var DefaultOptimizedDB *OptimizedDatabase

// InitOptimizedDatabase 初始化优化数据库
func InitOptimizedDatabase(cfg *config.Config) error {
	db, err := NewOptimizedDatabase(cfg)
	if err != nil {
		return err
	}
	DefaultOptimizedDB = db
	return nil
}

// ConnectionPoolOptimizer 连接池优化器
type ConnectionPoolOptimizer struct {
	db *OptimizedDatabase
}

// NewConnectionPoolOptimizer 创建连接池优化器
func NewConnectionPoolOptimizer(db *OptimizedDatabase) *ConnectionPoolOptimizer {
	return &ConnectionPoolOptimizer{db: db}
}

// OptimizeForWorkload 根据工作负载优化连接池
func (cpo *ConnectionPoolOptimizer) OptimizeForWorkload(workloadType string) {
	sqlDB, _ := cpo.db.DB.DB()

	switch workloadType {
	case "read_heavy":
		// 读密集型工作负载
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(20)
	case "write_heavy":
		// 写密集型工作负载
		sqlDB.SetMaxOpenConns(80)
		sqlDB.SetMaxIdleConns(15)
	case "mixed":
		// 混合工作负载
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
	default:
		// 默认配置
		sqlDB.SetMaxOpenConns(cpo.db.Config.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cpo.db.Config.MaxIdleConns)
	}
}

// MonitorConnections 监控连接池
func (cpo *ConnectionPoolOptimizer) MonitorConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sqlDB, _ := cpo.db.DB.DB()
		stats := sqlDB.Stats()

		// 动态调整连接池大小
		if float64(stats.InUse) > float64(stats.MaxOpenConnections)*0.8 {
			// 如果使用率超过80%，增加连接数
			newMax := stats.MaxOpenConnections + 10
			if newMax <= 200 { // 设置上限
				sqlDB.SetMaxOpenConns(newMax)
			}
		}
	}
}
