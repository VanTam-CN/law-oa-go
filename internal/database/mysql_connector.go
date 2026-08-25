package database

import (
	"database/sql"
	"fmt"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
)

// openMySQLGORM opens MySQL from a structured driver configuration. GORM is
// given the resulting database/sql connection pool directly, so credentials
// are never formatted into and re-parsed from a DSN.
func openMySQLGORM(cfg config.DatabaseConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	driverConfig, err := cfg.MySQLDriverConfig()
	if err != nil {
		return nil, err
	}
	conn, err := drivermysql.NewConnector(driverConfig)
	if err != nil {
		return nil, fmt.Errorf("build MySQL connector: %w", err)
	}
	return gorm.Open(mysql.New(mysql.Config{Conn: sql.OpenDB(conn)}), gormConfig)
}

// openMySQLSQL opens MySQL from a structured driver configuration for callers
// that need database/sql rather than GORM.
func openMySQLSQL(cfg config.DatabaseConfig, multiStatements bool) (*sql.DB, error) {
	driverConfig, err := cfg.MySQLDriverConfig()
	if err != nil {
		return nil, err
	}
	driverConfig.MultiStatements = multiStatements
	conn, err := drivermysql.NewConnector(driverConfig)
	if err != nil {
		return nil, fmt.Errorf("build MySQL connector: %w", err)
	}
	return sql.OpenDB(conn), nil
}
