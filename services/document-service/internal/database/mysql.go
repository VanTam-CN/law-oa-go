package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
}

// DefaultMySQLConfig 返回默认MySQL配置
func DefaultMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "password",
		Database:        "document_service",
		Charset:         "utf8mb4",
		ParseTime:       true,
		Loc:             "Local",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 3600, // 1小时
		ConnMaxIdleTime: 600,  // 10分钟
	}
}

// GetDSN 获取数据源名称
func (c *MySQLConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, c.ParseTime, c.Loc)
}

// Database 数据库管理器
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建新的数据库实例
func NewDatabase(config *MySQLConfig) (*Database, error) {
	dsn := config.GetDSN()

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)

	return &Database{DB: db}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Ping 检查数据库连接
func (d *Database) Ping(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// GetDB 获取GORM DB实例
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Health 数据库健康检查
func (d *Database) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.Ping(ctx)
}

// Stats 获取数据库连接统计信息
func (d *Database) Stats() map[string]interface{} {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":   stats.MaxOpenConnections,
		"open_connections":       stats.OpenConnections,
		"in_use":                stats.InUse,
		"idle":                  stats.Idle,
		"wait_count":            stats.WaitCount,
		"wait_duration":         stats.WaitDuration.String(),
		"max_idle_closed":       stats.MaxIdleClosed,
		"max_idle_time_closed":  stats.MaxIdleTimeClosed,
		"max_lifetime_closed":   stats.MaxLifetimeClosed,
	}
}

// AutoMigrate 自动迁移数据库表
func (d *Database) AutoMigrate(models ...interface{}) error {
	log.Println("开始数据库迁移...")

	if err := d.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("数据库迁移完成")
	return nil
}

// CreateIndexes 创建数据库索引
func (d *Database) CreateIndexes() error {
	db := d.DB

	// 创建复合索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_documents_tenant_category ON documents(tenant_id, category)",
		"CREATE INDEX IF NOT EXISTS idx_documents_entity_type_id ON documents(entity_type, entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_documents_status_created ON documents(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_document_versions_document_version ON document_versions(document_id, version)",
		"CREATE INDEX IF NOT EXISTS idx_document_permissions_tenant_user ON document_permissions(tenant_id, user_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_permissions_tenant_role ON document_permissions(tenant_id, role_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_audits_tenant_action ON document_audits(tenant_id, action)",
		"CREATE INDEX IF NOT EXISTS idx_document_audits_created_at ON document_audits(created_at)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			log.Printf("创建索引失败: %s, 错误: %v", index, err)
		} else {
			log.Printf("成功创建索引: %s", index)
		}
	}

	return nil
}

// BeginTx 开始事务
func (d *Database) BeginTx(ctx context.Context, opts ...*sql.TxOptions) (*gorm.DB, error) {
	tx := d.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	return tx, nil
}