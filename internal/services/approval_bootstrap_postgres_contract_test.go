package services_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/database"
)

func bootstrapSlot(t *testing.T, db *gorm.DB, intakeID interface{}) error {
	t.Helper()
	return db.Exec(`INSERT INTO approval_case_creation_tracking(id,approval_request_id,intake_id,created_by)
		VALUES (?,?,?,'11')`, uuid.NewString(), uuid.NewString(), intakeID).Error
}

func requireBootstrapUniqueViolation(t *testing.T, err error) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23505", pgErr.Code)
}

func TestApprovalBootstrapPostgresClaimAndSlotParity(t *testing.T) {
	db := ApprovalBootstrapPostgresDatabase(t)
	// Schema validation checks the exact 000080 nullable VARCHAR(64) shape and
	// the canonical 000081 full unique index. Exercise its actual NULL semantics.
	intakeID := uuid.NewString()
	require.NoError(t, bootstrapSlot(t, db, intakeID))
	requireBootstrapUniqueViolation(t, bootstrapSlot(t, db, intakeID))
	require.NoError(t, bootstrapSlot(t, db, nil))
	require.NoError(t, bootstrapSlot(t, db, nil))
	var count int64
	require.NoError(t, db.Table("approval_case_creation_tracking").Count(&count).Error)
	require.EqualValues(t, 3, count)
	// A second source bootstrap must neither rewrite nor delete existing slots.
	require.NoError(t, database.BootstrapProductionSchema(db))
	require.NoError(t, db.Table("approval_case_creation_tracking").Count(&count).Error)
	require.EqualValues(t, 3, count)
}

func TestApprovalBootstrapPostgresV15UpgradeFirstHTTPSubmission(t *testing.T) {
	db := ApprovalBootstrapPostgresDatabase(t)
	// Reproduce the installed v15 runtime shape, not an empty CREATE TABLE path.
	require.NoError(t, db.Exec("ALTER TABLE case_intakes DROP COLUMN approval_claim_token").Error)
	require.NoError(t, db.Exec("DROP INDEX idx_case_creation_tracking_intake_unique").Error)
	require.NoError(t, db.Exec("UPDATE schema_bootstrap_state SET version = 'postgres-mvp-2026-11-07-v15'").Error)
	intakeID := uuid.NewString()
	require.NoError(t, db.Exec(`INSERT INTO case_intakes(id,intake_code,title,status,metadata)
		VALUES (?,'V15-PRESERVE','upgrade preservation','draft','{"retained":true}')`, intakeID).Error)
	db = reopenApprovalBootstrapPostgres(t, db)
	require.ErrorContains(t, database.ValidateProductionSchemaContract(db), "approval_claim_token", "old v15 must not pass the new startup contract")
	db = reopenApprovalBootstrapPostgres(t, db)
	require.NoError(t, database.BootstrapProductionSchema(db))
	var version string
	require.NoError(t, db.Raw("SELECT version FROM schema_bootstrap_state WHERE id = 1").Scan(&version).Error)
	require.Equal(t, "postgres-mvp-2026-11-07-v16", version)
	var retained bool
	require.NoError(t, db.Raw("SELECT (metadata->>'retained')::boolean FROM case_intakes WHERE id = ?", intakeID).Scan(&retained).Error)
	require.True(t, retained)
	// Repeated bootstrap preserves a claim owned by an in-flight request.
	require.NoError(t, db.Exec("UPDATE case_intakes SET approval_claim_token = 'retained-owner' WHERE id = ?", intakeID).Error)
	require.NoError(t, database.BootstrapProductionSchema(db))
	var token string
	require.NoError(t, db.Raw("SELECT approval_claim_token FROM case_intakes WHERE id = ?", intakeID).Scan(&token).Error)
	require.Equal(t, "retained-owner", token)
	assertApprovalBootstrapFirstHTTPSubmission(t, db)
}

func TestApprovalBootstrapPostgresStartupRejectsBrokenClaimContract(t *testing.T) {
	db := ApprovalBootstrapPostgresDatabase(t)
	for _, scenario := range []struct{ name, damage string }{
		{"missing_claim", "ALTER TABLE case_intakes DROP COLUMN approval_claim_token"},
		{"short_claim", "ALTER TABLE case_intakes ALTER COLUMN approval_claim_token TYPE VARCHAR(32)"},
		{"nonnull_claim", "ALTER TABLE case_intakes ALTER COLUMN approval_claim_token SET NOT NULL"},
		{"missing_index", "DROP INDEX idx_case_creation_tracking_intake_unique"},
		{"nonunique_index", "DROP INDEX idx_case_creation_tracking_intake_unique; CREATE INDEX idx_case_creation_tracking_intake_unique ON approval_case_creation_tracking(intake_id)"},
		{"partial_index", "DROP INDEX idx_case_creation_tracking_intake_unique; CREATE UNIQUE INDEX idx_case_creation_tracking_intake_unique ON approval_case_creation_tracking(intake_id) WHERE creation_status = 'pending'"},
		{"wrong_column_index", "DROP INDEX idx_case_creation_tracking_intake_unique; CREATE UNIQUE INDEX idx_case_creation_tracking_intake_unique ON approval_case_creation_tracking(approval_request_id)"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			tx := reopenApprovalBootstrapPostgres(t, db).Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()
			// These are deliberate corruptions of this test's private schema.
			for _, statement := range strings.Split(scenario.damage, ";") {
				require.NoError(t, tx.Exec(statement).Error)
			}
			want := "唯一占槽"
			if strings.Contains(scenario.name, "claim") {
				want = "approval_claim_token"
			}
			require.ErrorContains(t, database.ValidateProductionSchemaContract(tx), want)
		})
	}
	require.NoError(t, database.ValidateProductionSchemaContract(db))
}

func TestApprovalBootstrapPostgresDuplicateLegacySlotsFailAtomically(t *testing.T) {
	db := ApprovalBootstrapPostgresDatabase(t)
	// Removing this column also removes every implicit/explicit slot index.
	// Legacy duplicates must be reported, never silently repaired or discarded.
	require.NoError(t, db.Exec("ALTER TABLE approval_case_creation_tracking DROP COLUMN intake_id").Error)
	require.NoError(t, db.Exec("ALTER TABLE approval_case_creation_tracking ADD COLUMN intake_id VARCHAR(36)").Error)
	require.NoError(t, db.Exec("ALTER TABLE case_intakes DROP COLUMN approval_claim_token").Error)
	require.NoError(t, db.Exec("UPDATE schema_bootstrap_state SET version = 'postgres-mvp-2026-11-07-v15'").Error)
	intakeID := uuid.NewString()
	require.NoError(t, bootstrapSlot(t, db, intakeID))
	require.NoError(t, bootstrapSlot(t, db, intakeID))
	err := database.BootstrapProductionSchema(db)
	requireBootstrapUniqueViolation(t, err)
	require.False(t, db.Migrator().HasColumn("case_intakes", "approval_claim_token"), "the additive column must roll back with the failed unique index")
	var version string
	require.NoError(t, db.Raw("SELECT version FROM schema_bootstrap_state WHERE id = 1").Scan(&version).Error)
	require.Equal(t, "postgres-mvp-2026-11-07-v15", version)
	var count int64
	require.NoError(t, db.Table("approval_case_creation_tracking").Where("intake_id = ?", intakeID).Count(&count).Error)
	require.EqualValues(t, 2, count)
}
