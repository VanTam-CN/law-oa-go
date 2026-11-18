package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

// Continuation acceptance: the read-only decision must be derived from
// server-owned evidence only, must distinguish waiting-for-review from a
// broken detection chain, and must never leak identity digests, ciphertext,
// or plaintext identity numbers through the JSON projection.

func seedContinuationWorld(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/continuation.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE approval_snapshots (id TEXT PRIMARY KEY, approval_request_id TEXT, snapshot_type TEXT, snapshot_data TEXT, source_version INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)").Error)
	require.NoError(t, db.AutoMigrate(&models.ApprovalRequest{}, &models.ConflictCheckRecord{}, &models.ConflictReview{}, &models.User{}))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, client_id TEXT, status TEXT, metadata TEXT, updated_at DATETIME)").Error)
	require.NoError(t, db.Exec("CREATE TABLE clients (id INTEGER PRIMARY KEY, name TEXT, type TEXT, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT DEFAULT '', deleted_at DATETIME DEFAULT NULL)").Error)
	require.NoError(t, db.Exec("CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT DEFAULT '', metadata TEXT DEFAULT '{}')").Error)
	return db
}

type continuationSeedOptions struct {
	factsVersion     string
	committedVer     string
	invalidated      bool
	intakeStatus     string
	reviewDecision   string
	withReview       bool
	clientIdentity   string
	tamperDigest     bool
	frozenClientKey  string
	dropFrozenAnchor bool
	clientAlias      string
}

func seedContinuationIntake(t *testing.T, db *gorm.DB, intakeID string, opts continuationSeedOptions) (clientDigest, partyDigest string) {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("r", 32))
	identity := opts.clientIdentity
	if identity == "" {
		identity = "11010119900307421X"
	}
	ciphertext, digest, err := security.ProtectIdentityNumber(identity)
	require.NoError(t, err)
	if opts.tamperDigest {
		digest = strings.Repeat("0", 64)
	}
	masterAliases := ""
	if opts.clientAlias != "" {
		masterAliases = opts.clientAlias
	}
	require.NoError(t, db.Exec(
		"INSERT INTO clients (id, name, type, identity_type, identity_number_ciphertext, identity_number_digest, aliases) VALUES (9, '客户甲', '个人', 'ID_CARD', ?, ?, ?)",
		ciphertext, digest, masterAliases).Error)
	// A real intake check requires at least one protected opposing party
	// (INTAKE_PARTY_IDENTITY_REQUIRED), so the full-chain fixture freezes a
	// COMPANY party with its own SOCIAL_CREDIT_CODE anchor and stores the
	// matching case_intake_parties row. Name-only parties cannot carry the
	// frozen identity chain, by contract.
	partyCiphertext, partyDigest, err := security.ProtectIdentityNumber("91310000MA1FL0001X")
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, metadata) VALUES (?, '对手公司', 'COMPANY', 'RELATED_PARTY', 'SOCIAL_CREDIT_CODE', ?, ?, ?)",
		intakeID, partyCiphertext, partyDigest, `{"party_type":"OPPOSING_PARTY"}`).Error)
	// A real Batch01 create/update save stores the client as a guarded
	// party row (caseIntakePartyIsClient -> prepareCaseIntakeClientPartyRow),
	// while StartIntakeConflictCheck still freezes the client only from the
	// client master. Seed that mirror row so continuation proves the master
	// identity instead of double-counting it as a party.
	mirrorAlias := ""
	if opts.clientAlias != "" {
		mirrorAlias = ", a-historical-name"
	}
	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, aliases, metadata) VALUES (?, '客户甲', 'INDIVIDUAL', 'client', 'ID_CARD', ?, ?, ?, '{}')",
		intakeID, ciphertext, digest, mirrorAlias).Error)
	factsVersion := opts.factsVersion
	if factsVersion == "" {
		factsVersion = "FV-1"
	}
	committed := opts.committedVer
	if committed == "" {
		committed = factsVersion
	}
	status := opts.intakeStatus
	if status == "" {
		status = "conflict_ready"
	}
	metadata := map[string]interface{}{
		"draft_revision":               "rev-1",
		"conflict_check_id":            "CCT_CONT_1",
		"lawyer_facts_version":         factsVersion,
		"conflict_check_facts_version": committed,
	}
	if opts.invalidated {
		metadata["conflict_invalidated_at"] = time.Now().Format(time.RFC3339)
	}
	raw, err := json.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, client_id, status, metadata, updated_at) VALUES (?, '9', ?, ?, CURRENT_TIMESTAMP)",
		intakeID, status, string(raw)).Error)

	clientAnchor := "hmac-sha256:" + digest
	if opts.dropFrozenAnchor {
		clientAnchor = ""
	}
	partyAnchor := "hmac-sha256:" + partyDigest
	if opts.dropFrozenAnchor {
		partyAnchor = ""
	}
	frozen := map[string]interface{}{
		"intakeId":          intakeID,
		"clientId":          "9",
		"clientName":        "客户甲",
		"clientType":        "PERSON",
		"clientIdentifiers": map[string]string{"id_card": clientAnchor},
		"parties": []map[string]interface{}{
			{"name": "对手公司", "role": "OPPOSING_PARTY", "entityType": "COMPANY",
				"identifiers": map[string]string{"social_credit_code": partyAnchor}},
		},
	}
	frozenRaw, err := json.Marshal(frozen)
	require.NoError(t, err)
	coverage := map[string]interface{}{
		"decision": map[string]interface{}{"coverageStatus": "COMPLETE"},
	}
	coverageRaw, _ := json.Marshal(coverage)
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_check_records (check_id, client_id, client_name, check_status, search_parameters, check_result, user_id, check_time, created_at, updated_at) "+
			"VALUES ('CCT_CONT_1', '9', '客户甲', 'COMPLETED', ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		string(frozenRaw), string(coverageRaw)).Error)
	if opts.withReview {
		decision := opts.reviewDecision
		if decision == "" {
			decision = "no_conflict"
		}
		require.NoError(t, db.Exec(
			"INSERT INTO conflict_reviews (check_id, decision, notes, reviewer_id, reviewer_name, evidence_hash, created_at) "+
				"VALUES ('CCT_CONT_1', ?, '独立复核通过', 42, '核查人', 'hash', CURRENT_TIMESTAMP)",
			decision).Error)
	}
	return digest, partyDigest
}

func TestIntakeContinuationFullChainAllowsSubmitWithoutApproval(t *testing.T) {
	db := seedContinuationWorld(t)
	digest, partyDigest := seedContinuationIntake(t, db, "INT-CONT-A", continuationSeedOptions{withReview: true})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-A")
	require.NoError(t, err)
	require.True(t, decision.CanReuseIdentity)
	require.True(t, decision.CanReuseCheck)
	require.True(t, decision.CanSubmit)
	require.False(t, decision.RecheckRequired)
	require.Empty(t, decision.Reasons)
	require.NotEmpty(t, decision.Subjects)
	require.Contains(t, decision.Subjects[0].IdentityMasked, "*")
	require.GreaterOrEqual(t, len(decision.Subjects), 2, "the full chain must include the protected opposing party row")
	require.Equal(t, "party", decision.Subjects[1].Kind)
	require.Equal(t, "OPPOSING_PARTY", decision.Subjects[1].Role)
	require.Equal(t, "COMPANY", decision.Subjects[1].EntityType)
	require.Contains(t, decision.Subjects[1].IdentityMasked, "*")

	raw, err := json.Marshal(decision)
	require.NoError(t, err)
	body := string(raw)
	require.NotContains(t, body, "identity_digest")
	require.NotContains(t, body, "ciphertext")
	require.NotContains(t, body, "11010119900307421X")
	require.NotContains(t, body, "91310000MA1FL0001X")
	require.NotContains(t, body, "hmac-sha256:")
	require.NotContains(t, body, digest)
	require.NotContains(t, body, partyDigest)
}

func TestIntakeContinuationWaitsForReviewWithoutRecheck(t *testing.T) {
	db := seedContinuationWorld(t)
	seedContinuationIntake(t, db, "INT-CONT-B", continuationSeedOptions{withReview: false})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-B")
	require.NoError(t, err)
	require.True(t, decision.CanReuseIdentity)
	require.True(t, decision.CanReuseCheck)
	require.False(t, decision.CanSubmit)
	require.False(t, decision.RecheckRequired, "missing review must keep the original check and wait, not re-run detection")
	require.Equal(t, []string{ContinuationReasonReviewMissing}, decision.Reasons)
	require.True(t, decision.CanReuseIdentity)
}

func TestIntakeContinuationInvalidatedFactsForceRecheck(t *testing.T) {
	db := seedContinuationWorld(t)
	seedContinuationIntake(t, db, "INT-CONT-C", continuationSeedOptions{withReview: true, invalidated: true})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-C")
	require.NoError(t, err)
	require.False(t, decision.CanReuseIdentity)
	require.False(t, decision.CanReuseCheck)
	require.False(t, decision.CanSubmit)
	require.True(t, decision.RecheckRequired)
	require.Contains(t, decision.Reasons, ContinuationReasonFactsInvalidated)
}

func TestIntakeContinuationConfirmedConflictBlocksSubmitButKeepsCheck(t *testing.T) {
	db := seedContinuationWorld(t)
	seedContinuationIntake(t, db, "INT-CONT-D", continuationSeedOptions{withReview: true, reviewDecision: "confirmed_conflict"})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-D")
	require.NoError(t, err)
	require.True(t, decision.CanReuseIdentity)
	require.True(t, decision.CanReuseCheck)
	require.False(t, decision.CanSubmit)
	require.False(t, decision.RecheckRequired)
	require.Equal(t, []string{ContinuationReasonReviewConflict}, decision.Reasons)
}

func TestIntakeContinuationIdentityChangedAfterOldCheckFailsClosed(t *testing.T) {
	db := seedContinuationWorld(t)
	_, partyDigest := seedContinuationIntake(t, db, "INT-CONT-E", continuationSeedOptions{withReview: true, clientIdentity: "11010119900307999X", frozenClientKey: "stale"})
	svc := NewIntakeContinuationService(db)
	// Re-freeze the anchors to a different identity so the mismatch is
	// proven, not assumed.
	staleDigest, err := security.IdentityDigest("110101198001011234")
	require.NoError(t, err)
	frozen := map[string]interface{}{
		"intakeId":          "INT-CONT-E",
		"clientId":          "9",
		"clientName":        "客户甲",
		"clientType":        "COMPANY",
		"clientIdentifiers": map[string]string{"id_card": "hmac-sha256:" + staleDigest},
		"parties": []map[string]interface{}{
			{"name": "对手公司", "role": "OPPOSING_PARTY", "entityType": "COMPANY",
				"identifiers": map[string]string{"social_credit_code": "hmac-sha256:" + partyDigest}},
		},
	}
	frozenRaw, err := json.Marshal(frozen)
	require.NoError(t, err)
	require.NoError(t, db.Exec("UPDATE conflict_check_records SET search_parameters = ? WHERE check_id = 'CCT_CONT_1'", string(frozenRaw)).Error)

	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-E")
	require.NoError(t, err)
	require.False(t, decision.CanReuseIdentity)
	require.False(t, decision.CanReuseCheck)
	require.False(t, decision.CanSubmit)
	require.True(t, decision.RecheckRequired, "a changed protected identity must re-run detection")
	require.Contains(t, decision.Reasons, ContinuationReasonSubjectIdentityChanged)
}

func TestIntakeContinuationTamperedDigestFailsClosedWithoutLeak(t *testing.T) {
	db := seedContinuationWorld(t)
	seedContinuationIntake(t, db, "INT-CONT-F", continuationSeedOptions{withReview: true, tamperDigest: true})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-F")
	require.NoError(t, err)
	require.False(t, decision.CanReuseIdentity)
	require.False(t, decision.CanReuseCheck)
	require.False(t, decision.CanSubmit)
	require.Contains(t, decision.Reasons, ContinuationReasonSubjectUnreadable)
	require.Empty(t, decision.Subjects)
	raw, err := json.Marshal(decision)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "identity_digest")
	require.NotContains(t, string(raw), "ciphertext")
}

func TestIntakeContinuationMissingFrozenAnchorFailsClosed(t *testing.T) {
	db := seedContinuationWorld(t)
	seedContinuationIntake(t, db, "INT-CONT-G", continuationSeedOptions{withReview: true, dropFrozenAnchor: true})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-G")
	require.NoError(t, err)
	require.False(t, decision.CanReuseIdentity, "an old check without frozen identity anchors cannot prove the same subject")
	require.False(t, decision.CanReuseCheck)
	require.True(t, decision.RecheckRequired)
	require.Contains(t, decision.Reasons, ContinuationReasonSubjectIdentityChanged)
}

// The real Batch01 save path stores a guarded client mirror row
// (party_role=client) next to opposing-party rows, while
// StartIntakeConflictCheck freezes the client ONLY from the client master.
// Continuation must prove that mirror row against the master (name, type,
// identity, aliases) and then normalize it into the client subject, so the
// current-evidence set matches the frozen tuple set one-to-one. A tampered
// mirror identity must fail closed as SUBJECT_UNKNOWN without swallowing
// the change or leaking protected material.
func TestIntakeContinuationClientMirrorRowNormalizesToClientSubject(t *testing.T) {
	db := seedContinuationWorld(t)
	digest, partyDigest := seedContinuationIntake(t, db, "INT-CONT-H", continuationSeedOptions{withReview: true})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-H")
	require.NoError(t, err)
	require.False(t, decision.hasReason(ContinuationReasonSubjectUnreadable), "a guarded client mirror row must not break the readable subject set: %v", decision.Reasons)
	require.False(t, decision.hasReason(ContinuationReasonSubjectIdentityChanged), "a mirror row matching the client master must not be mistaken for an added party: %v", decision.Reasons)
	require.True(t, decision.CanReuseIdentity)
	require.True(t, decision.CanReuseCheck)
	require.True(t, decision.CanSubmit)
	require.False(t, decision.RecheckRequired)
	require.Empty(t, decision.Reasons)
	require.Len(t, decision.Subjects, 2)
	require.Equal(t, "client", decision.Subjects[0].Kind)
	require.Equal(t, "客户甲", decision.Subjects[0].Name)
	require.Equal(t, "CLIENT", decision.Subjects[0].Role)
	require.Equal(t, []string{}, decision.Subjects[0].Aliases)
	require.Equal(t, "party", decision.Subjects[1].Kind)
	require.Equal(t, "OPPOSING_PARTY", decision.Subjects[1].Role)
	raw, err := json.Marshal(decision)
	require.NoError(t, err)
	body := string(raw)
	require.NotContains(t, body, "hmac-sha256:")
	require.NotContains(t, body, digest)
	require.NotContains(t, body, partyDigest)
	require.NotContains(t, body, "11010119900307421X")
	require.NotContains(t, body, "91310000MA1FL0001X")
}

func TestIntakeContinuationTamperedClientMirrorRowFailsClosed(t *testing.T) {
	db := seedContinuationWorld(t)
	digest, partyDigest := seedContinuationIntake(t, db, "INT-CONT-I", continuationSeedOptions{withReview: true})
	tamperedCiphertext, tamperedDigest, err := security.ProtectIdentityNumber("999999202401019999")
	require.NoError(t, err)
	require.NotEqual(t, digest, tamperedDigest)
	require.NoError(t, db.Exec(
		"UPDATE case_intake_parties SET identity_number_ciphertext = ? WHERE intake_id = ? AND party_role = 'client'",
		tamperedCiphertext, "INT-CONT-I").Error)
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-I")
	require.NoError(t, err)
	require.False(t, decision.CanReuseIdentity)
	require.False(t, decision.CanReuseCheck)
	require.False(t, decision.CanSubmit)
	require.True(t, decision.RecheckRequired)
	require.Empty(t, decision.Subjects)
	require.Contains(t, decision.Reasons, ContinuationReasonSubjectUnreadable)
	require.NotContains(t, decision.Reasons, ContinuationReasonSubjectIdentityChanged, "an unreadable mirror row is evidence corruption, not a proven fact change")
	raw, err := json.Marshal(decision)
	require.NoError(t, err)
	body := string(raw)
	require.NotContains(t, body, "hmac-sha256:")
	require.NotContains(t, body, digest)
	require.NotContains(t, body, partyDigest)
	require.NotContains(t, body, "999999202401019999")
}

func TestContinuationIntakeEntityTypeMatchesMasterContract(t *testing.T) {
	cases := []struct {
		name         string
		entityType   string
		identityType string
		want         bool
	}{
		{"person aliases", "PERSON", "ID_CARD", true},
		{"individual canonical", "INDIVIDUAL", "PASSPORT", true},
		{"chinese person alias", "自然人", "ID_CARD", true},
		{"company canonical", "LEGAL_PERSON", "SOCIAL_CREDIT_CODE", true},
		{"company alias", "COMPANY", "BUSINESS_LICENSE", true},
		{"chinese company alias", "公司", "ORGANIZATION_CODE", true},
		{"organization bucket cannot mirror an org-code master", "ORGANIZATION", "ORGANIZATION_CODE", false},
		{"empty uses legal-person save default on company identity", "", "SOCIAL_CREDIT_CODE", true},
		{"empty rejects individual identity", "", "ID_CARD", false},
		{"unknown non-empty fails closed", "TAMPERED", "SOCIAL_CREDIT_CODE", false},
		{"unknown spelling fails closed", "UNKNOWN", "ID_CARD", false},
		{"any is not a known bucket", "ANY", "SOCIAL_CREDIT_CODE", false},
		{"person never matches company identity", "PERSON", "SOCIAL_CREDIT_CODE", false},
		{"company never matches personal identity", "LEGAL_PERSON", "ID_CARD", false},
		{"unknown identity type fails closed", "LEGAL_PERSON", "TAMPERED", false},
		{"empty identity type keeps legacy fallback", "COMPANY", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, continuationIntakeEntityTypeMatchesMaster(tc.entityType, tc.identityType))
		})
	}
}

// Every guarded mirror fact must be proven before the row may be folded
// into the client subject. A broken protected identity (missing ciphertext,
// missing digest, digest drift) or a proven mirror-fact drift (type bucket,
// identity type, alias drift in the column OR in row metadata) fails closed
// as SUBJECT_UNKNOWN. A renamed or de-cliented row falls back to the party
// path and fails the frozen one-to-one comparison as an identity change.
func TestIntakeContinuationClientMirrorNegativeCases(t *testing.T) {
	cases := make([]struct {
		name        string
		mirrorSQL   string
		args        []interface{}
		wantChanged bool
	}, 0)
	cases = append(cases,
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "missing mirror ciphertext", mirrorSQL: "UPDATE case_intake_parties SET identity_number_ciphertext = '' WHERE intake_id = ? AND party_role = 'client'"},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "mirror digest no longer matches ciphertext", mirrorSQL: "UPDATE case_intake_parties SET identity_number_digest = ? WHERE intake_id = ? AND party_role = 'client'", args: []interface{}{strings.Repeat("a", 64)}},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "missing mirror digest", mirrorSQL: "UPDATE case_intake_parties SET identity_number_digest = '' WHERE intake_id = ? AND party_role = 'client'"},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "mirror entity type bucket drift", mirrorSQL: "UPDATE case_intake_parties SET entity_type = 'LEGAL_PERSON' WHERE intake_id = ? AND party_role = 'client'"},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "mirror unknown entity type fails closed", mirrorSQL: "UPDATE case_intake_parties SET entity_type = 'TAMPERED' WHERE intake_id = ? AND party_role = 'client'"},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "mirror identity type drift", mirrorSQL: "UPDATE case_intake_parties SET identity_type = 'SOCIAL_CREDIT_CODE' WHERE intake_id = ? AND party_role = 'client'"},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "mirror alias column drift", mirrorSQL: "UPDATE case_intake_parties SET aliases = '旧用名' WHERE intake_id = ? AND party_role = 'client'"},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "mirror metadata-only alias cannot be swallowed", mirrorSQL: "UPDATE case_intake_parties SET metadata = ? WHERE intake_id = ? AND party_role = 'client'", args: []interface{}{`{"aliases":["元数据旧名"]}`}},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "renamed client-role row is an extra party", mirrorSQL: "UPDATE case_intake_parties SET entity_name = '改名客户' WHERE intake_id = ? AND party_role = 'client'", wantChanged: true},
		struct {
			name        string
			mirrorSQL   string
			args        []interface{}
			wantChanged bool
		}{name: "de-cliented row with client name is an extra party", mirrorSQL: "UPDATE case_intake_parties SET party_role = 'related_party' WHERE intake_id = ? AND party_role = 'client'", wantChanged: true},
	)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := seedContinuationWorld(t)
			digest, partyDigest := seedContinuationIntake(t, db, "INT-CONT-NEG", continuationSeedOptions{withReview: true})
			args := append([]interface{}{}, tc.args...)
			args = append(args, "INT-CONT-NEG")
			require.NoError(t, db.Exec(tc.mirrorSQL, args...).Error)
			svc := NewIntakeContinuationService(db)
			decision, err := svc.EvaluateIntakeContinuation("INT-CONT-NEG")
			require.NoError(t, err)
			require.False(t, decision.CanReuseIdentity)
			require.False(t, decision.CanReuseCheck)
			require.False(t, decision.CanSubmit)
			require.True(t, decision.RecheckRequired)
			if tc.wantChanged {
				// The renamed/de-cliented row stays a readable party tuple;
				// the frozen one-to-one comparison proves the identity change.
				require.Len(t, decision.Subjects, 3)
				require.Contains(t, decision.Subjects[2].IdentityMasked, "*")
				require.Contains(t, decision.Reasons, ContinuationReasonSubjectIdentityChanged)
			} else {
				require.Empty(t, decision.Subjects)
				require.Contains(t, decision.Reasons, ContinuationReasonSubjectUnreadable)
			}
			raw, err := json.Marshal(decision)
			require.NoError(t, err)
			body := string(raw)
			require.NotContains(t, body, "hmac-sha256:")
			require.NotContains(t, body, "ciphertext")
			require.NotContains(t, body, digest)
			require.NotContains(t, body, partyDigest)
			require.NotContains(t, body, "11010119900307421X")
			require.NotContains(t, body, "91310000MA1FL0001X")
		})
	}
}

// A master alias stored on the client record must be mirrored on the client
// party row: the frozen audit only carries the metadata alias list, so a
// master-only alias would otherwise widen the current subject beyond the
// frozen tuple and must fail closed as an unreadable subject set, never
// pass.
func TestIntakeContinuationClientMasterAliasWithoutMirrorAliasFailsClosed(t *testing.T) {
	db := seedContinuationWorld(t)
	_, partyDigest := seedContinuationIntake(t, db, "INT-CONT-ALIAS", continuationSeedOptions{withReview: true, clientAlias: "主档案旧名"})
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-ALIAS")
	require.NoError(t, err)
	require.False(t, decision.CanReuseIdentity)
	require.False(t, decision.CanReuseCheck)
	require.False(t, decision.CanSubmit)
	require.Empty(t, decision.Subjects)
	require.Contains(t, decision.Reasons, ContinuationReasonSubjectUnreadable)
	require.NotContains(t, decision.Reasons, ContinuationReasonSubjectIdentityChanged)
	raw, err := json.Marshal(decision)
	require.NoError(t, err)
	body := string(raw)
	require.NotContains(t, body, "hmac-sha256:")
	require.NotContains(t, body, partyDigest)
}

// A company client saved through the real guarded path carries aliases in
// several copies at once: the client master column, the intake metadata
// list, and the mirror row column plus mirror metadata. The exact stored
// spellings duplicate values, pad them with whitespace, and order them
// differently; the guarded-save contract treats those copies as one alias
// fact, so a mirror proving the same trimmed alias set must still fold into
// the client subject and reuse the old check.
func TestIntakeContinuationCompanyClientMirrorWithAliasesNormalizes(t *testing.T) {
	db := seedContinuationWorld(t)
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("r", 32))
	masterCiphertext, masterDigest, err := security.ProtectIdentityNumber("91310000MA1FL9999X")
	require.NoError(t, err)
	partyCiphertext, partyDigest, err := security.ProtectIdentityNumber("91310000MA1FL0001X")
	require.NoError(t, err)
	// Real guarded-save shapes: master aliases with padding, duplicates and
	// an empty segment; metadata list keeping duplicates; mirror column and
	// mirror metadata repeating the same facts in a different order.
	require.NoError(t, db.Exec(
		"INSERT INTO clients (id, name, type, identity_type, identity_number_ciphertext, identity_number_digest, aliases) "+
			"VALUES (9, '澄明练习甲公司', '企业', 'SOCIAL_CREDIT_CODE', ?, ?, ' 昭宇咨询 ,澄明旧称,昭宇咨询, ')",
		masterCiphertext, masterDigest).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, metadata) "+
			"VALUES ('INT-CONT-ALIAS-OK', '对手公司', 'COMPANY', 'RELATED_PARTY', 'SOCIAL_CREDIT_CODE', ?, ?, '{\"party_type\":\"OPPOSING_PARTY\"}')",
		partyCiphertext, partyDigest).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, aliases, metadata) "+
			"VALUES ('INT-CONT-ALIAS-OK', '澄明练习甲公司', 'LEGAL_PERSON', 'client', 'SOCIAL_CREDIT_CODE', ?, ?, '昭宇咨询, 澄明旧称 , , 昭宇咨询', '{\"aliases\":[\" 澄明旧称 \"],\"client_aliases\":[\"昭宇咨询\"]}')",
		masterCiphertext, masterDigest).Error)
	metadata := map[string]interface{}{
		"draft_revision":               "rev-1",
		"conflict_check_id":            "CCT_CONT_1",
		"lawyer_facts_version":         "FV-1",
		"conflict_check_facts_version": "FV-1",
		"client_aliases":               []string{"昭宇咨询", "昭宇咨询", " 澄明旧称 "},
	}
	metadataRaw, err := json.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, client_id, status, metadata, updated_at) VALUES ('INT-CONT-ALIAS-OK', '9', 'conflict_ready', ?, CURRENT_TIMESTAMP)",
		string(metadataRaw)).Error)
	frozen := map[string]interface{}{
		"intakeId":          "INT-CONT-ALIAS-OK",
		"clientId":          "9",
		"clientName":        "澄明练习甲公司",
		"clientType":        "COMPANY",
		"clientAliases":     []string{"昭宇咨询", "昭宇咨询", " 澄明旧称 "},
		"clientIdentifiers": map[string]string{"social_credit_code": "hmac-sha256:" + masterDigest},
		"parties": []map[string]interface{}{
			{"name": "对手公司", "role": "OPPOSING_PARTY", "entityType": "COMPANY",
				"identifiers": map[string]string{"social_credit_code": "hmac-sha256:" + partyDigest}},
		},
	}
	frozenRaw, err := json.Marshal(frozen)
	require.NoError(t, err)
	coverageRaw, _ := json.Marshal(map[string]interface{}{
		"decision": map[string]interface{}{"coverageStatus": "COMPLETE"},
	})
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_check_records (check_id, client_id, client_name, check_status, search_parameters, check_result, user_id, check_time, created_at, updated_at) "+
			"VALUES ('CCT_CONT_1', '9', '澄明练习甲公司', 'COMPLETED', ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		string(frozenRaw), string(coverageRaw)).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_reviews (check_id, decision, notes, reviewer_id, reviewer_name, evidence_hash, created_at) "+
			"VALUES ('CCT_CONT_1', 'no_conflict', '独立复核通过', 42, '核查人', 'hash', CURRENT_TIMESTAMP)").Error)
	svc := NewIntakeContinuationService(db)
	decision, err := svc.EvaluateIntakeContinuation("INT-CONT-ALIAS-OK")
	require.NoError(t, err)
	require.Empty(t, decision.Reasons, "the same alias fact stored in guarded-save copies must still continue: %v", decision.Reasons)
	require.True(t, decision.CanReuseIdentity)
	require.True(t, decision.CanReuseCheck)
	require.True(t, decision.CanSubmit)
	require.False(t, decision.RecheckRequired)
	require.Len(t, decision.Subjects, 2, "the mirror row must fold into the client subject, not double-count")
	require.Equal(t, "client", decision.Subjects[0].Kind)
	require.Equal(t, "澄明练习甲公司", decision.Subjects[0].Name)
	require.Equal(t, "party", decision.Subjects[1].Kind)
	require.Equal(t, "对手公司", decision.Subjects[1].Name)
	require.Contains(t, decision.Subjects[0].IdentityMasked, "*")
	raw, err := json.Marshal(decision)
	require.NoError(t, err)
	body := string(raw)
	require.NotContains(t, body, "hmac-sha256:")
	require.NotContains(t, body, "ciphertext")
	require.NotContains(t, body, masterDigest)
	require.NotContains(t, body, partyDigest)
	require.NotContains(t, body, "91310000MA1FL9999X")
	require.NotContains(t, body, "91310000MA1FL0001X")
}
