package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"law-oa-go/internal/config"
)

// Migrator 数据库迁移器
type Migrator struct {
	migrate    *migrate.Migrate
	db         *sql.DB
	driverName string
}

// NewMigrator 创建迁移器
func NewMigrator(cfg *config.DatabaseConfig, migrationsPath string) (*Migrator, error) {
	driverName := strings.ToLower(cfg.Driver)
	if driverName == "" {
		driverName = "postgres"
	}

	sqlDriver, databaseName, dsn := buildMigrationDSN(cfg, driverName)
	db, err := sql.Open(sqlDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	migrationDriver, err := newMigrationDriver(db, driverName)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseName,
		migrationDriver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate:    m,
		db:         db,
		driverName: driverName,
	}, nil
}

func buildMigrationDSN(cfg *config.DatabaseConfig, driverName string) (sqlDriver string, databaseName string, dsn string) {
	if driverName == "postgres" || driverName == "postgresql" {
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		return "postgres", "postgres", fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Password,
			cfg.Database,
			sslMode,
		)
	}

	return "mysql", "mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%v&loc=%s&multiStatements=true&tls=skip-verify",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)
}

func newMigrationDriver(db *sql.DB, driverName string) (migratedb.Driver, error) {
	if driverName == "postgres" || driverName == "postgresql" {
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres driver: %w", err)
		}
		return driver, nil
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create mysql driver: %w", err)
	}
	return driver, nil
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up() error {
	log.Println("开始执行数据库迁移...")

	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("数据库已是最新版本，无需迁移")
	} else {
		log.Println("数据库迁移完成")
	}

	return nil
}

// Down 回滚所有迁移
func (m *Migrator) Down() error {
	log.Println("开始回滚数据库迁移...")

	err := m.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("没有可回滚的迁移")
	} else {
		log.Println("数据库迁移回滚完成")
	}

	return nil
}

// Steps 执行指定步数的迁移
func (m *Migrator) Steps(n int) error {
	if n > 0 {
		log.Printf("开始执行 %d 步迁移...", n)
	} else {
		log.Printf("开始回滚 %d 步迁移...", -n)
	}

	err := m.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d steps: %w", n, err)
	}

	if err == migrate.ErrNoChange {
		log.Println("没有可执行的迁移步骤")
	} else {
		log.Printf("迁移步骤执行完成")
	}

	return nil
}

// Goto 迁移到指定版本
func (m *Migrator) Goto(version uint) error {
	log.Printf("开始迁移到版本 %d...", version)

	err := m.migrate.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	if err == migrate.ErrNoChange {
		log.Printf("数据库已是版本 %d", version)
	} else {
		log.Printf("成功迁移到版本 %d", version)
	}

	return nil
}

// Version 获取当前数据库版本
func (m *Migrator) Version() (uint, bool, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}

// Force 强制设置数据库版本（用于修复损坏的迁移状态）
func (m *Migrator) Force(version int) error {
	log.Printf("强制设置数据库版本为 %d...", version)

	err := m.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}

	log.Printf("成功强制设置版本为 %d", version)
	return nil
}

// Drop 删除所有表（危险操作）
func (m *Migrator) Drop() error {
	log.Println("警告：开始删除所有数据库表...")

	err := m.migrate.Drop()
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	log.Println("所有数据库表已删除")
	return nil
}

// Close 关闭迁移器
func (m *Migrator) Close() error {
	if m.migrate != nil {
		sourceErr, dbErr := m.migrate.Close()
		if sourceErr != nil {
			return fmt.Errorf("failed to close source: %w", sourceErr)
		}
		if dbErr != nil {
			return fmt.Errorf("failed to close database: %w", dbErr)
		}
	}

	if m.db != nil {
		return m.db.Close()
	}

	return nil
}

// Status 获取迁移状态信息
func (m *Migrator) Status() (*MigrationStatus, error) {
	version, dirty, err := m.Version()
	if err != nil {
		return nil, err
	}

	return &MigrationStatus{
		Version: version,
		Dirty:   dirty,
	}, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Version uint `json:"version"`
	Dirty   bool `json:"dirty"`
}

// ValidateDatabase 验证数据库连接和基本结构
func (m *Migrator) ValidateDatabase() error {
	// 检查数据库连接
	if err := m.db.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// 检查迁移表是否存在
	var count int
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'schema_migrations'"
	if m.driverName == "postgres" || m.driverName == "postgresql" {
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'schema_migrations'"
	}
	err := m.db.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration table: %w", err)
	}

	if count == 0 {
		log.Println("迁移表不存在，这可能是首次运行")
	}

	return nil
}

// GetAppliedMigrations 获取已应用的迁移列表
func (m *Migrator) GetAppliedMigrations() ([]uint, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var versions []uint
	for rows.Next() {
		var version uint
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}
