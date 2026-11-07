package services_test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/database"
)

// Start with an EMPTY private schema and invoke exactly the production
// bootstrap entry point. No AutoMigrate, migration runner or fixture DDL may
// supply the runtime tables/columns that this smoke test is meant to verify.
func ApprovalBootstrapPostgresDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("LAW_OA_COLLAB_PG_DSN")
	if dsn == "" {
		t.Skip("NOT_RUN: LAW_OA_COLLAB_PG_DSN unset; fresh bootstrap PostgreSQL coverage not verified")
	}
	parsed, err := pgconn.ParseConfig(dsn)
	require.NoError(t, err)
	require.Equal(t, "/private/tmp/lawoa-intake-collaboration-runtime/socket", parsed.Host)
	require.EqualValues(t, 55437, parsed.Port)
	require.Equal(t, "lawoa_intake_mvp_test", parsed.Database)
	require.Equal(t, "mac", parsed.User)
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("p", 32))
	config := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), PrepareStmt: true, SkipDefaultTransaction: true, DisableForeignKeyConstraintWhenMigrating: true}
	admin, err := gorm.Open(postgres.Open(dsn), config)
	require.NoError(t, err)
	// The shared local cluster is already bootstrapped. Do not relocate a
	// database-wide extension into a disposable schema during this test.
	var extensionPresent bool
	require.NoError(t, admin.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto')").Scan(&extensionPresent).Error)
	require.True(t, extensionPresent, "local test database must already have pgcrypto")
	schema := "bootstrap_intake_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		require.NoError(t, admin.Exec("DROP SCHEMA "+schema+" CASCADE").Error)
		conn, _ := admin.DB()
		_ = conn.Close()
	})
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema+" application_name=approval_bootstrap_pg_test statement_timeout=30000 lock_timeout=10000"), config)
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	conn.SetMaxOpenConns(6)
	t.Cleanup(func() { _ = conn.Close() })
	var tableCount int64
	require.NoError(t, db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema()").Scan(&tableCount).Error)
	require.Zero(t, tableCount)
	require.NoError(t, database.BootstrapProductionSchema(db))
	require.NoError(t, database.ValidateProductionSchemaContract(db))
	require.NoError(t, db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema()").Scan(&tableCount).Error)
	t.Logf("actual BootstrapProductionSchema version=%s empty_schema=%s created_tables=%d", database.ProductionSchemaVersion, schema, tableCount)
	return db
}

// A restart gets a fresh prepared-statement cache. The v15 test deliberately
// changes a v16 schema back to its old shape, so it must reconnect before
// exercising the upgrade command instead of reusing plans for the old shape.
func reopenApprovalBootstrapPostgres(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()
	var schema string
	require.NoError(t, db.Raw("SELECT current_schema()").Scan(&schema).Error)
	require.True(t, strings.HasPrefix(schema, "bootstrap_intake_"))
	fresh, err := gorm.Open(postgres.Open(os.Getenv("LAW_OA_COLLAB_PG_DSN")+" search_path="+schema+" application_name=approval_bootstrap_pg_restart statement_timeout=30000 lock_timeout=10000"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent), PrepareStmt: true, SkipDefaultTransaction: true, DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	conn, err := fresh.DB()
	require.NoError(t, err)
	conn.SetMaxOpenConns(6)
	t.Cleanup(func() { _ = conn.Close() })
	return fresh
}
