//go:build np_m1_pg

package m1_test

// I15 conflict write-back regression on the dedicated M1 PostgreSQL instance.
// Contract under test: ConflictSubjectAssociation.IntakeRevision, the
// conflict_check_records.search_parameters JSON key "intakeRevision", and the
// repositories.ErrIntakeRevisionConflict sentinel.
//
// A matching input revision links intake and cases in one protected
// transaction and advances the intake draft revision exactly once; stale or
// missing revisions fail closed without writing anything. The big JSON
// integer must survive the metadata merge byte-for-byte. Run with:
//
//	go test -tags np_m1_pg ./tests/integration/m1 -run TestConflictPG -count=1

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// Same literal the native draft test guards: a JSON number above the
// float64-precise range that must come back as the exact decimal string.
const conflictPGBigPrecision = "9007199254740993"

type conflictPGFixture struct {
	intakeID string
	caseID   string
	clientID string
	userID   uint
	// revision is the input revision recorded in the check request;
	// nextRevision is the deterministic post-completion draft revision.
	revision     string
	nextRevision string
}

func seedConflictPGFixture(t *testing.T, db *gorm.DB, withCase bool, caseCheckID string) conflictPGFixture {
	t.Helper()
	name := uuid.NewString()
	user := models.User{Username: "m1c-" + name, Name: "M1 conflict synthetic lawyer", Email: name + "@np.invalid", Password: "disabled-synthetic", Role: "lawyer", Status: "active"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	client := models.Client{Name: "M1 conflict synthetic client", Email: name + "-client@np.invalid", CreatedBy: user.ID, Type: "企业"}
	if err := db.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	intakeID := uuid.NewString()
	revision := "rev-" + name
	metadata := fmt.Sprintf("{\"draft_revision\":%q,\"precision\":%s,\"note\":\"seed\"}", revision, conflictPGBigPrecision)
	if err := db.Exec(
		"INSERT INTO case_intakes(id,intake_code,title,client_id,status,case_type,priority,description,metadata,created_by) VALUES(?,?,?,?,'lawyer_facts_confirmed','civil','medium','seed',?::jsonb,?)",
		intakeID, "IC-"+name, "M1 conflict seed", client.ID, metadata, user.ID,
	).Error; err != nil {
		t.Fatal(err)
	}
	fixture := conflictPGFixture{
		intakeID:     intakeID,
		clientID:     fmt.Sprintf("%d", client.ID),
		userID:       user.ID,
		revision:     revision,
		nextRevision: "next-" + name,
	}
	if withCase {
		// cases.id is bigint and LinkConflictCheckToCase parses the
		// association case ID as uint32, so stay inside that range.
		caseIDNum := int64(1000000000) + time.Now().UnixNano()%3000000000
		caseID := fmt.Sprintf("%d", caseIDNum)
		if err := db.Exec(
			"INSERT INTO cases(id,case_number,title,client_id,lawyer_id,case_type,status,conflict_check_id) VALUES(?,?,?,?,?,'civil','pending',?)",
			caseIDNum, "CN-"+name, "M1 conflict seed case", client.ID, user.ID, caseCheckID,
		).Error; err != nil {
			t.Fatal(err)
		}
		fixture.caseID = caseID
	}
	return fixture
}

func insertConflictPGRecord(t *testing.T, db *gorm.DB, fixture conflictPGFixture, checkID, intakeRevision string) {
	t.Helper()
	params := fmt.Sprintf("{\"clientId\":%q,\"intakeId\":%q,\"subjectCaseId\":%q,\"intakeRevision\":%q,\"precision\":%s}",
		fixture.clientID, fixture.intakeID, fixture.caseID, intakeRevision, conflictPGBigPrecision)
	if err := db.Exec(
		"INSERT INTO conflict_check_records(check_id,client_id,client_name,case_name,case_type,check_status,has_conflict,risk_level,search_parameters,check_result,user_id,check_time,duration) VALUES(?,?,?,?,?,'COMPLETED',false,'LOW',?::json,'{}'::json,?,?,?)",
		checkID, fixture.clientID, "seed client", "seed case", "civil", params, fixture.userID, time.Now(), int64(1),
	).Error; err != nil {
		t.Fatal(err)
	}
}

func conflictPGIntakeState(t *testing.T, db *gorm.DB, intakeID string) (revision, status, linkedCheck, precision string) {
	t.Helper()
	var row struct {
		Revision    string
		Status      string
		LinkedCheck string
		Precision   string
	}
	if err := db.Raw(
		"SELECT coalesce(metadata->>'draft_revision','') AS revision, status, coalesce(metadata->>'conflict_check_id','') AS linked_check, coalesce(metadata->>'precision','') AS precision FROM case_intakes WHERE id=?",
		intakeID,
	).Scan(&row).Error; err != nil {
		t.Fatal(err)
	}
	return row.Revision, row.Status, row.LinkedCheck, row.Precision
}

func conflictPGCaseCheckID(t *testing.T, db *gorm.DB, caseID string) string {
	t.Helper()
	var row struct {
		LinkedCheck *string
	}
	if err := db.Raw("SELECT conflict_check_id AS linked_check FROM cases WHERE id=?", caseID).Scan(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.LinkedCheck == nil {
		return ""
	}
	return *row.LinkedCheck
}

func TestConflictPGLinkAdvancesRevisionOnce(t *testing.T) {
	db := isolatedDB(t)
	ctx := context.Background()
	fixture := seedConflictPGFixture(t, db, true, "")
	checkID := "chk-" + uuid.NewString()
	insertConflictPGRecord(t, db, fixture, checkID, fixture.revision)
	association := repositories.ConflictSubjectAssociation{
		CheckID:            checkID,
		SubjectCaseID:      fixture.caseID,
		IntakeID:           fixture.intakeID,
		ClientID:           fixture.clientID,
		IntakeRevision:     fixture.revision,
		CompletionRevision: fixture.nextRevision,
	}
	if err := repositories.LinkConflictCheckToCase(ctx, db, association); err != nil {
		t.Fatalf("matching revision must link atomically: %v", err)
	}
	revision, status, linked, precision := conflictPGIntakeState(t, db, fixture.intakeID)
	if revision != fixture.nextRevision {
		t.Fatalf("intake revision must advance exactly once to %q, got %q", fixture.nextRevision, revision)
	}
	if status != "conflict_ready" || linked != checkID {
		t.Fatalf("intake link incomplete: status=%q check=%q", status, linked)
	}
	if precision != conflictPGBigPrecision {
		t.Fatalf("big integer precision lost on metadata merge: got %q want %s", precision, conflictPGBigPrecision)
	}
	if got := conflictPGCaseCheckID(t, db, fixture.caseID); got != checkID {
		t.Fatalf("case pointer mismatch after link: %q", got)
	}
	// The same matching-version input replayed after completion must fail
	// closed: the stored revision no longer matches, and nothing may advance
	// or overwrite a second time.
	if err := repositories.LinkConflictCheckToCase(ctx, db, association); !errors.Is(err, repositories.ErrIntakeRevisionConflict) {
		t.Fatalf("replay of completed revision must reject with ErrIntakeRevisionConflict, got %v", err)
	}
	revision2, status2, linked2, precision2 := conflictPGIntakeState(t, db, fixture.intakeID)
	if revision2 != revision || status2 != status || linked2 != linked || precision2 != precision {
		t.Fatalf("rejected replay mutated intake: rev %q/%q status %q/%q check %q/%q precision %q/%q",
			revision, revision2, status, status2, linked, linked2, precision, precision2)
	}
	if got := conflictPGCaseCheckID(t, db, fixture.caseID); got != checkID {
		t.Fatalf("rejected replay mutated case pointer: %q", got)
	}
}

func TestConflictPGRejectsStaleIntakeRevision(t *testing.T) {
	db := isolatedDB(t)
	ctx := context.Background()
	fixture := seedConflictPGFixture(t, db, true, "")
	// The first record completes with the matching revision; the second
	// record still carries that now-stale revision and must be rejected
	// without writing anything.
	firstCheck := "chk-" + uuid.NewString()
	staleCheck := "chk-" + uuid.NewString()
	insertConflictPGRecord(t, db, fixture, firstCheck, fixture.revision)
	if err := repositories.LinkConflictCheckToCase(ctx, db, repositories.ConflictSubjectAssociation{
		CheckID:            firstCheck,
		SubjectCaseID:      fixture.caseID,
		IntakeID:           fixture.intakeID,
		ClientID:           fixture.clientID,
		IntakeRevision:     fixture.revision,
		CompletionRevision: fixture.nextRevision,
	}); err != nil {
		t.Fatalf("baseline link failed: %v", err)
	}
	revision, status, linked, precision := conflictPGIntakeState(t, db, fixture.intakeID)
	insertConflictPGRecord(t, db, fixture, staleCheck, fixture.revision)
	if err := repositories.LinkConflictCheckToCase(ctx, db, repositories.ConflictSubjectAssociation{
		CheckID:        staleCheck,
		SubjectCaseID:  fixture.caseID,
		IntakeID:       fixture.intakeID,
		ClientID:       fixture.clientID,
		IntakeRevision: fixture.revision,
	}); !errors.Is(err, repositories.ErrIntakeRevisionConflict) {
		t.Fatalf("stale revision must reject with ErrIntakeRevisionConflict, got %v", err)
	}
	revision2, status2, linked2, precision2 := conflictPGIntakeState(t, db, fixture.intakeID)
	if revision2 != revision || status2 != status || linked2 != linked || precision2 != precision {
		t.Fatalf("stale rejection mutated intake: rev %q/%q status %q/%q check %q/%q precision %q/%q",
			revision, revision2, status, status2, linked, linked2, precision, precision2)
	}
}

func TestConflictPGRejectsMissingIntakeRevision(t *testing.T) {
	db := isolatedDB(t)
	ctx := context.Background()
	fixture := seedConflictPGFixture(t, db, false, "")
	// Legacy record without the intakeRevision key and an empty association
	// revision: ambiguity fails closed, the draft revision must not advance.
	checkID := "chk-" + uuid.NewString()
	params := fmt.Sprintf("{\"clientId\":%q,\"intakeId\":%q}", fixture.clientID, fixture.intakeID)
	if err := db.Exec(
		"INSERT INTO conflict_check_records(check_id,client_id,client_name,case_name,case_type,check_status,has_conflict,risk_level,search_parameters,check_result,user_id,check_time,duration) VALUES(?,?,?,?,?,'COMPLETED',false,'LOW',?::json,'{}'::json,?,?,?)",
		checkID, fixture.clientID, "seed client", "seed case", "civil", params, fixture.userID, time.Now(), int64(1),
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := repositories.LinkConflictCheckToCase(ctx, db, repositories.ConflictSubjectAssociation{
		CheckID:  checkID,
		IntakeID: fixture.intakeID,
		ClientID: fixture.clientID,
	}); !errors.Is(err, repositories.ErrIntakeRevisionConflict) {
		t.Fatalf("missing input revision must reject with ErrIntakeRevisionConflict, got %v", err)
	}
	revision, status, linked, precision := conflictPGIntakeState(t, db, fixture.intakeID)
	if revision != fixture.revision || status != "lawyer_facts_confirmed" || linked != "" || precision != conflictPGBigPrecision {
		t.Fatalf("missing-revision rejection mutated intake: rev %q status %q check %q precision %q", revision, status, linked, precision)
	}
}

func TestConflictPGLinkRollsBackAcrossIntakeAndCases(t *testing.T) {
	db := isolatedDB(t)
	ctx := context.Background()
	// Case already bound to a different check: the whole association must
	// roll back, leaving intake status, pointer and revision untouched.
	fixture := seedConflictPGFixture(t, db, true, "chk-prebound-"+uuid.NewString())
	checkID := "chk-" + uuid.NewString()
	insertConflictPGRecord(t, db, fixture, checkID, fixture.revision)
	if err := repositories.LinkConflictCheckToCase(ctx, db, repositories.ConflictSubjectAssociation{
		CheckID:            checkID,
		SubjectCaseID:      fixture.caseID,
		IntakeID:           fixture.intakeID,
		ClientID:           fixture.clientID,
		IntakeRevision:     fixture.revision,
		CompletionRevision: fixture.nextRevision,
	}); err == nil {
		t.Fatal("case bound to another check must fail the whole link")
	}
	revision, status, linked, precision := conflictPGIntakeState(t, db, fixture.intakeID)
	if revision != fixture.revision || status != "lawyer_facts_confirmed" || linked != "" || precision != conflictPGBigPrecision {
		t.Fatalf("failed link mutated intake: rev %q status %q check %q precision %q", revision, status, linked, precision)
	}
	if got := conflictPGCaseCheckID(t, db, fixture.caseID); got == checkID || got == "" {
		t.Fatalf("pre-bound case pointer changed: %q", got)
	}
}
