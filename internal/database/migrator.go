package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	if cfg == nil {
		return nil, fmt.Errorf("database configuration is required for migrations")
	}
	driverName := strings.ToLower(cfg.Driver)
	if driverName == "" {
		driverName = "postgres"
	}
	if driverName != "mysql" && driverName != "postgres" && driverName != "postgresql" {
		return nil, fmt.Errorf("unsupported migration database driver %q; use mysql or postgres", cfg.Driver)
	}
	db, databaseName, err := openMigrationDatabase(cfg, driverName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	currentVersion, dirty, initialized, err := readMigrationVersion(db, driverName)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if dirty {
		_ = db.Close()
		return nil, fmt.Errorf("database migration state is dirty at version %d; repair the migration state before running new migrations", currentVersion)
	}
	if (driverName == "postgres" || driverName == "postgresql") && !initialized {
		_ = db.Close()
		return nil, fmt.Errorf("PostgreSQL migration history is not initialized; use -command bootstrap for a fresh or production database, not the historical migration chain")
	}

	if err := validatePendingMigrationDirectory(migrationsPath, driverName, currentVersion); err != nil {
		_ = db.Close()
		return nil, err
	}

	migrationDriver, err := newMigrationDriver(db, driverName)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseName,
		migrationDriver,
	)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate:    m,
		db:         db,
		driverName: driverName,
	}, nil
}

// validateMigrationDirectory prevents a mixed historical directory from
// partially mutating a target database. The repository currently contains
// both PostgreSQL-only and MySQL-only files; accepting that directory and
// relying on the first SQL error leaves golang-migrate in a dirty state after
// earlier migrations may already have committed.
func validateMigrationDirectory(migrationsPath, driverName string) error {
	return validateMigrationFiles(migrationsPath, driverName, 0, false)
}

// validatePendingMigrationDirectory validates only migrations newer than the
// database version. This keeps an old mixed-dialect history from blocking an
// already bootstrapped database while still failing closed before any pending
// SQL is executed.
func validatePendingMigrationDirectory(migrationsPath, driverName string, currentVersion uint) error {
	return validateMigrationFiles(migrationsPath, driverName, currentVersion, true)
}

func validateMigrationFiles(migrationsPath, driverName string, currentVersion uint, pendingOnly bool) error {
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("read migration directory %q: %w", migrationsPath, err)
	}

	postgresOnlyMarkers := []string{
		"jsonb",
		"timestamptz",
		"bigserial",
		"do $$",
		"on conflict",
		"create extension",
		"comment on",
		"::jsonb",
	}
	mysqlOnlyMarkers := []string{
		"engine=innodb",
		"delimiter",
		"auto_increment",
		"create procedure",
		" on update current_timestamp",
	}
	sqlFileCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			continue
		}
		version, hasVersion := migrationVersionFromFilename(entry.Name())
		if pendingOnly && (!hasVersion || version <= currentVersion) {
			continue
		}
		sqlFileCount++
		path := filepath.Join(migrationsPath, entry.Name())
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration file %q: %w", path, err)
		}
		lower := strings.ToLower(string(contents))
		if driverName == "mysql" {
			if marker := firstMigrationMarker(lower, postgresOnlyMarkers); marker != "" {
				return fmt.Errorf("migration directory %q contains PostgreSQL-only syntax %q in %s; provide a MySQL-specific migration directory", migrationsPath, marker, entry.Name())
			}
		}
		if driverName == "postgres" || driverName == "postgresql" {
			if marker := firstMigrationMarker(lower, mysqlOnlyMarkers); marker != "" {
				return fmt.Errorf("migration directory %q contains MySQL-only syntax %q in %s; provide a PostgreSQL-specific migration directory", migrationsPath, marker, entry.Name())
			}
		}
	}
	if sqlFileCount == 0 {
		return fmt.Errorf("migration directory %q contains no SQL migration files", migrationsPath)
	}

	return nil
}

func migrationVersionFromFilename(name string) (uint, bool) {
	if len(name) < 7 || name[6] != '_' {
		return 0, false
	}
	version, err := strconv.ParseUint(name[:6], 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(version), true
}

func readMigrationVersion(db *sql.DB, driverName string) (version uint, dirty bool, initialized bool, err error) {
	var tableCount int
	if driverName == "postgres" || driverName == "postgresql" {
		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = current_schema()
			  AND table_name = 'schema_migrations'`,
		).Scan(&tableCount)
	} else {
		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
			  AND table_name = 'schema_migrations'`,
		).Scan(&tableCount)
	}
	if err != nil {
		return 0, false, false, fmt.Errorf("failed to inspect schema_migrations: %w", err)
	}
	if tableCount == 0 {
		return 0, false, false, nil
	}

	err = db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
	if err == sql.ErrNoRows {
		return 0, false, false, nil
	}
	if err != nil {
		return 0, false, false, fmt.Errorf("failed to read schema_migrations: %w", err)
	}
	return version, dirty, true, nil
}

func firstMigrationMarker(contents string, markers []string) string {
	for _, marker := range markers {
		if strings.Contains(contents, marker) {
			return marker
		}
	}
	return ""
}

func openMigrationDatabase(cfg *config.DatabaseConfig, driverName string) (*sql.DB, string, error) {
	if driverName == "postgres" || driverName == "postgresql" {
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Password,
			cfg.Database,
			sslMode,
		)
		db, err := sql.Open("postgres", dsn)
		return db, "postgres", err
	}

	db, err := openMySQLSQL(*cfg, true)
	return db, "mysql", err
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
		return nil, err
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
