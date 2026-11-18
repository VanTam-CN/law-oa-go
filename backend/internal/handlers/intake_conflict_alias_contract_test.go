package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

// recordingConflictAliasService wraps the REAL detection engine so the
// handler freezes exactly the same audit payload production would: the
// record built here is what the continuation gate later compares against.
type recordingConflictAliasService struct {
	services.ConflictDetectionService
	calls  int32
	record *models.ConflictCheckRecord
}

func (s *recordingConflictAliasService) PrepareConflictCheck(ctx context.Context, r *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	s.calls++
	return s.ConflictDetectionService.PrepareConflictCheck(ctx, r)
}

func (s *recordingConflictAliasService) BuildConflictCheckRecord(ctx context.Context, r *models.ConflictCheckRequest, resp *models.ConflictCheckResponse) *models.ConflictCheckRecord {
	record := s.ConflictDetectionService.BuildConflictCheckRecord(ctx, r, resp)
	s.record = record
	return record
}

type aliasContractCase struct {
	name string
	// DB-facing alias storage.
	masterAliases string // clients.aliases column
	mirrorAliases string // guarded client mirror row aliases column
	metadataJSON  string // case_intakes.metadata JSON
	partyAliases  string // case_intake_parties.aliases column
	partyMetadata string // case_intake_parties.metadata JSON (opposing party)
	// Expected real-handler frozen facts.
	wantClientAliases []string
	wantPartyAliases  []string
	// Branch C reason: a changed master alias normally surfaces as
	// SUBJECT_IDENTITY_CHANGED, but the guarded mirror row scenario
	// fails earlier as SUBJECT_UNKNOWN (the mirror row no longer matches
	// its own master). Both are fail-closed recheck outcomes.
	wantChangedReason string
}

func TestStartIntakeConflictCheckFreezesAndContinuesAliasContract(t *testing.T) {
	cases := []aliasContractCase{
		{
			name:              "client_master_column_only",
			masterAliases:     "客户历史名称",
			wantClientAliases: []string{"客户历史名称"},
			wantPartyAliases:  []string{},
			wantChangedReason: services.ContinuationReasonSubjectIdentityChanged,
		},
		{
			name:              "client_master_plus_guarded_mirror_row",
			masterAliases:     "客户历史名称",
			mirrorAliases:     "客户历史名称",
			wantClientAliases: []string{"客户历史名称"},
			wantPartyAliases:  []string{},
			wantChangedReason: services.ContinuationReasonSubjectUnreadable,
		},
		{
			name:              "intake_metadata_unions_master_column",
			metadataJSON:      "{\"client_aliases\":[\"客户登记名A\"]}",
			masterAliases:     "客户历史名称",
			wantClientAliases: []string{"客户历史名称", "客户登记名A"},
			wantPartyAliases:  []string{},
			wantChangedReason: services.ContinuationReasonSubjectIdentityChanged,
		},
		{
			name:              "opposing_party_column_alias",
			partyAliases:      "对手历史名称",
			wantClientAliases: []string{},
			wantPartyAliases:  []string{"对手历史名称"},
			wantChangedReason: services.ContinuationReasonSubjectIdentityChanged,
		},
		{
			// The handler stringListValue accepts a delimited STRING in
			// metadata and splits it with the shared separator contract; the
			// continuation gate must compare against the same split set, so
			// this shape freezes and re-verifies instead of failing closed.
			name:              "intake_metadata_string_splits_and_freezes",
			metadataJSON:      "{\"client_aliases\":\"客户旧名甲；客户旧名乙\"}",
			wantClientAliases: []string{"客户旧名甲", "客户旧名乙"},
			wantPartyAliases:  []string{},
			wantChangedReason: services.ContinuationReasonSubjectIdentityChanged,
		},
		{
			name:              "opposing_party_metadata_string_splits_and_freezes",
			partyMetadata:     "{\"aliases\":\"对方旧名甲；对方旧名乙\"}",
			wantClientAliases: []string{},
			wantPartyAliases:  []string{"对方旧名甲", "对方旧名乙"},
			wantChangedReason: services.ContinuationReasonSubjectIdentityChanged,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newAliasContractDB(t)
			seedAliasContractIntake(t, db, tc)
			if strings.TrimSpace(tc.partyMetadata) != "" {
				require.NoError(t, db.Exec("UPDATE case_intake_parties SET metadata = ? WHERE party_role = 'opposing_party'", tc.partyMetadata).Error)
			}

			svc := newAliasContractHandlerService(t, db)
			r := newAliasContractRouter(t, db, svc)
			runAliasContractCheck(t, r)
			require.NotNil(t, svc.record)

			// Branch A: the real handler froze the master column into the
			// audit record; the frozen aliases must match the current facts.
			parameters := svc.record.SearchParameters.ToMap()
			// Alias facts are compared as sorted sets: the gate compares
			// sorted sets, so frozen insertion order is not part of the
			// contract and either storage order must freeze equivalently.
			require.Equal(t, sortedAliasContractStrings(tc.wantClientAliases), sortedAliasContractStrings(aliasContractStringSet(parameters["clientAliases"])))
			parties := aliasContractFrozenParties(t, parameters["parties"])
			require.Len(t, parties, 1)
			require.Equal(t, tc.wantPartyAliases, aliasContractStringSet(parties[0]["aliases"]))

			// The intake commit moved the intake to conflict_ready and bound
			// the committed facts version, so the chain is committable.
			status, meta := aliasContractIntakeState(t, db, "intake-alias")
			require.Equal(t, "conflict_ready", status)
			require.Equal(t, "FV-alias-1", meta["conflict_check_facts_version"])

			seedAliasContractReview(t, db, svc.record.CheckID)
			decision, err := services.NewIntakeContinuationService(db).EvaluateIntakeContinuation("intake-alias")
			require.NoError(t, err)
			require.True(t, decision.CanSubmit, "reasons: %v", decision.Reasons)
			require.False(t, decision.RecheckRequired)
			require.Empty(t, decision.Reasons)

			// Branch B: a legacy audit snapshot missing the alias fact the
			// NEW contract freezes must keep failing closed and demand a
			// fresh check - a missing alias may never be dismissed as "the
			// old snapshot had none". The omission is constructed at the
			// storage copy this scenario actually exercises: client-side
			// scenarios drop clientAliases; the opposing-party scenario
			// drops the party aliases key, so both legacy shapes are
			// explicit instead of vacuous.
			// The legacy map is a SHALLOW copy: for party scenarios the
			// deletion below mutates the nested party map that parameters
			// still references, so the original snapshot bytes must be
			// captured here, before that mutation, for the Branch C restore.
			originalSnapshotRaw, err := json.Marshal(parameters)
			require.NoError(t, err)
			legacy := map[string]interface{}{}
			for key, value := range parameters {
				legacy[key] = value
			}
			if len(tc.wantClientAliases) > 0 {
				delete(legacy, "clientAliases")
			} else if len(tc.wantPartyAliases) > 0 {
				rawParties, ok := legacy["parties"].([]interface{})
				require.True(t, ok)
				require.NotEmpty(t, rawParties)
				firstParty, ok := rawParties[0].(map[string]interface{})
				require.True(t, ok)
				delete(firstParty, "aliases")
			}
			legacyRaw, err := json.Marshal(legacy)
			require.NoError(t, err)
			require.NoError(t, db.Exec("UPDATE conflict_check_records SET search_parameters = ? WHERE check_id = ?",
				string(legacyRaw), svc.record.CheckID).Error)
			legacyDecision, err := services.NewIntakeContinuationService(db).EvaluateIntakeContinuation("intake-alias")
			require.NoError(t, err)
			require.False(t, legacyDecision.CanSubmit)
			require.True(t, legacyDecision.RecheckRequired)
			require.Contains(t, legacyDecision.Reasons, services.ContinuationReasonSubjectIdentityChanged)

			// Branch C restores the REAL frozen record first: Branch B
			// deliberately mangled it into the legacy shape, and the
			// changed-fact run must start from the snapshot the real handler
			// produced, not from the legacy deletion. A genuinely changed
			// master alias on top of the intact snapshot must still be
			// refused, so the unified contract cannot be satisfied by
			// weakening the comparison.
			require.NoError(t, db.Exec("UPDATE conflict_check_records SET search_parameters = ? WHERE check_id = ?",
				string(originalSnapshotRaw), svc.record.CheckID).Error)
			restoredDecision, err := services.NewIntakeContinuationService(db).EvaluateIntakeContinuation("intake-alias")
			require.NoError(t, err)
			require.True(t, restoredDecision.CanSubmit, "restored real snapshot reasons: %v", restoredDecision.Reasons)
			require.False(t, restoredDecision.RecheckRequired)
			require.NoError(t, db.Exec("UPDATE clients SET aliases = ? WHERE id = ?",
				"客户改名后的别名", mustAliasContractClientID(t, db)).Error)
			changedDecision, err := services.NewIntakeContinuationService(db).EvaluateIntakeContinuation("intake-alias")
			require.NoError(t, err)
			require.False(t, changedDecision.CanSubmit)
			require.True(t, changedDecision.RecheckRequired)
			require.Contains(t, changedDecision.Reasons, tc.wantChangedReason)
		})
	}
}

func sortedAliasContractStrings(values []string) []string {
	out := append([]string{}, values...)
	sort.Strings(out)
	return out
}

func newAliasContractDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("r", 32))
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "alias-contract.db")), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(
		&models.ConflictCheckRecord{}, &models.ConflictCase{}, &models.ConflictReview{},
		&models.Client{}, &models.User{}, &models.Case{},
		&models.Entity{}, &models.CaseParty{}, &models.EntityNameHistory{}, &models.EntityRelation{},
		&models.ConflictSearchScope{}, &models.ConflictIndexBuildRun{},
	))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT, client_id INTEGER, created_by TEXT, priority TEXT, status TEXT, metadata TEXT, created_at DATETIME, updated_at DATETIME)").Error)
	require.NoError(t, db.Exec("CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, case_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, relation_depth INTEGER DEFAULT 0, metadata TEXT, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT, created_at DATETIME)").Error)
	require.NoError(t, db.Exec("CREATE TABLE client_relations (id TEXT PRIMARY KEY, client_id TEXT, related_client_id TEXT, relation_type TEXT, relation_detail TEXT, active BOOLEAN, created_at DATETIME, updated_at DATETIME)").Error)
	return db
}

func seedAliasContractIntake(t *testing.T, db *gorm.DB, tc aliasContractCase) {
	t.Helper()
	identity := "91310000ALIASTEST1"
	ciphertext, digest, err := security.ProtectIdentityNumber(identity)
	require.NoError(t, err)
	client := &models.Client{
		Name: "别名客户甲", Type: "个人", Status: "active", CreatedBy: 7,
		Email:        "alias-client@test.local",
		IdentityType: models.IdentityTypeIDCard, IdentityNumberCiphertext: ciphertext, IdentityNumberDigest: digest,
		Aliases: tc.masterAliases,
	}
	require.NoError(t, db.Create(client).Error)
	require.NoError(t, db.Exec("INSERT INTO users (id, username, email, password, role, status, created_at, updated_at) VALUES (7, 'lawyer7', 'lawyer7@test.local', 'x', 'lawyer', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)").Error)
	// lawyer_facts_version is pre-seeded exactly as ConfirmIntakeFacts
	// writes it; the handler then commits conflict_check_facts_version from
	// the same value, so the continuation version gate can pass honestly.
	metadataMap := map[string]interface{}{}
	if strings.TrimSpace(tc.metadataJSON) != "" {
		require.NoError(t, json.Unmarshal([]byte(tc.metadataJSON), &metadataMap))
	}
	if _, exists := metadataMap["lawyer_facts_version"]; !exists {
		metadataMap["lawyer_facts_version"] = "FV-alias-1"
	}
	metadataBytes, err := json.Marshal(metadataMap)
	require.NoError(t, err)
	metadata := string(metadataBytes)
	require.NoError(t, db.Exec("INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, status, metadata) VALUES ('intake-alias', 'INT-AL', '别名冻结练习（虚构）', 'civil', ?, '7', 'lawyer_facts_confirmed', ?)", client.ID, metadata).Error)
	partyCipher, partyDigest, err := security.ProtectIdentityNumber("91310000MA1FL0002X")
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, aliases) VALUES ('intake-alias', '对手公司', 'LEGAL_PERSON', 'opposing_party', 'SOCIAL_CREDIT_CODE', ?, ?, ?)",
		partyCipher, partyDigest, tc.partyAliases).Error)
	if tc.mirrorAliases != "" {
		require.NoError(t, db.Exec(
			"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, aliases) VALUES ('intake-alias', '别名客户甲', 'INDIVIDUAL', 'client', 'ID_CARD', ?, ?, ?)",
			ciphertext, digest, tc.mirrorAliases).Error)
	}
	seedAliasContractCoverage(t, db)
}

// seedAliasContractCoverage registers the four authoritative archive scopes
// plus completed reconciliation runs, so the real engine's P0 decision
// freezes coverageStatus=COMPLETE (the continuation coverage gate).
func seedAliasContractCoverage(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now()
	coveredFrom := now.AddDate(-10, 0, 0)
	coveredTo := now.AddDate(1, 0, 0)
	syncedAt := now.Add(-time.Minute)
	reviewedAt := now.Add(-time.Minute)
	for index, scopeType := range []string{"CASE_ARCHIVE", "CLIENT_ARCHIVE", "SUBJECT_REGISTRY", "RELATION_ARCHIVE"} {
		runID := fmt.Sprintf("run-alias-%d", index)
		evidence := fmt.Sprintf("evidence://alias/%d", index)
		hash := strings.Repeat(fmt.Sprintf("%02x", index), 2) + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		require.NoError(t, db.Create(&models.ConflictIndexBuildRun{
			ID: runID, ScopeType: scopeType, SourceVersion: "alias-v1", Status: models.ConflictIndexBuildComplete,
			SourceRecordCount: 10, IndexedRecordCount: 10, MissingRecordCount: 0,
			ReconciliationHash: hash,
			EvidenceReference:  evidence,
			StartedAt:          now.Add(-time.Hour), CompletedAt: &now,
			CreatedAt: now, UpdatedAt: now,
		}).Error)
		require.NoError(t, db.Create(&models.ConflictSearchScope{
			ID: fmt.Sprintf("scope-alias-%d", index), ScopeType: scopeType,
			Status: services.ConflictScopeActive, CoverageStatus: services.ConflictCoverageComplete,
			SourceVersion: "alias-v1", EvidenceReference: evidence,
			CoveredFrom: &coveredFrom, CoveredTo: &coveredTo, MissingSources: "[]",
			IndexRunID: runID, SourceOfTruth: true, SyncMode: "BATCH",
			MaxSyncLagMinutes: 1440, LastSuccessfulSyncAt: &syncedAt,
			MinimumFieldCoverageBPS: 10000, MeasuredFieldCoverageBPS: 10000,
			MaximumDuplicateRateBPS: 10000, MeasuredDuplicateRateBPS: 0,
			QualityOwnerID: uintPtrAliasContract(7), QualityReviewedAt: &reviewedAt,
			MaxQualityReviewAgeDays: 365, FailureAlertReference: "alert://alias",
			CorrectionProcedureReference: "procedure://alias",
			CreatedAt:                    now, UpdatedAt: now,
		}).Error)
	}
}

func uintPtrAliasContract(value uint) *uint { return &value }

func newAliasContractHandlerService(t *testing.T, db *gorm.DB) *recordingConflictAliasService {
	t.Helper()
	impl := services.NewConflictDetectionService(
		repositories.NewConflictRepository(db, nil), services.NewRiskAssessor(nil, nil),
		repositories.NewUserRepository(db), repositories.NewClientRepository(db),
		repositories.NewCaseRepository(db),
	)
	return &recordingConflictAliasService{ConflictDetectionService: impl}
}

func newAliasContractRouter(t *testing.T, db *gorm.DB, svc *recordingConflictAliasService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	h.SetConflictDetectionService(svc)
	h.SetAuthorizationService(services.NewAuthorizationService(
		repositories.NewCaseRepository(db), repositories.NewClientRepository(db),
		repositories.NewUserRepository(db), repositories.NewDocumentRepository(db),
	))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		c.Next()
	})
	r.POST("/api/v1/case-intakes/:id/conflict-check", h.StartIntakeConflictCheck)
	return r
}

func runAliasContractCheck(t *testing.T, r *gin.Engine) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/case-intakes/intake-alias/conflict-check", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "handler response: %s", w.Body.String())
}

func seedAliasContractReview(t *testing.T, db *gorm.DB, checkID string) {
	t.Helper()
	review := &models.ConflictReview{
		CheckID: checkID, Decision: "no_conflict", Notes: "独立复核通过（虚构练习）",
		ReviewerID: 42, ReviewerName: "核查人", EvidenceHash: strings.Repeat("a", 64),
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.Create(review).Error)
}

func aliasContractIntakeState(t *testing.T, db *gorm.DB, intakeID string) (string, map[string]interface{}) {
	t.Helper()
	var row struct {
		Status   string
		Metadata string
	}
	require.NoError(t, db.Table("case_intakes").Select("status, metadata").Where("id = ?", intakeID).Take(&row).Error)
	meta := map[string]interface{}{}
	require.NoError(t, json.Unmarshal([]byte(row.Metadata), &meta))
	return row.Status, meta
}

func mustAliasContractClientID(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	var client models.Client
	require.NoError(t, db.Where("email = ?", "alias-client@test.local").Take(&client).Error)
	return client.ID
}

func aliasContractStringSet(value interface{}) []string {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var list []string
	if json.Unmarshal(raw, &list) != nil {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(list))
	for _, item := range list {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func aliasContractFrozenParties(t *testing.T, value interface{}) []map[string]interface{} {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	parties := []map[string]interface{}{}
	require.NoError(t, json.Unmarshal(raw, &parties))
	return parties
}
