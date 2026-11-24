package handlers

// In-place update contract for legacy native full saves. These tests pin
// write-path behaviors the revision fixtures do not cover yet: duplicate
// payload rows pointing at one stored party row must reject the whole save,
// and a real identity change must update the stored row in place (same
// primary key, preserved opaque columns) instead of deleting and recreating
// it.

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"gorm.io/gorm"

	"law-oa-go/internal/security"
)

func assertIntakeDraftPartiesUntouched(t *testing.T, db *gorm.DB) {
	t.Helper()
	rev, clientDigest, opponentDigest, title := draftRevisionServerState(t, db)
	if rev != "rev-1" || title != "版本守护草稿" {
		t.Fatalf("rejected save must not mutate the intake: rev=%q title=%q", rev, title)
	}
	if clientDigest != draftRevDigestA || opponentDigest != "opp-digest" {
		t.Fatalf("rejected save must not mutate party identities: client=%q opponent=%q", clientDigest, opponentDigest)
	}
	var rows []map[string]interface{}
	if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-rev").Order("id").Find(&rows).Error; err != nil {
		t.Fatalf("reload parties: %v", err)
	}
	if len(rows) != 2 || fmt.Sprint(rows[0]["id"]) != "1" || fmt.Sprint(rows[1]["id"]) != "2" {
		t.Fatalf("rejected save must not duplicate or renumber party rows: %v", rows)
	}
	if fmt.Sprint(rows[1]["aliases"]) != "对手别名" {
		t.Fatalf("rejected save must not touch stored aliases, got %v", rows[1]["aliases"])
	}
}

func TestUpdateIntakeRejectsDuplicatePointersToSamePartyRow(t *testing.T) {
	clientParty := intakeInplacePartyJSON("星河智联科技有限公司", "client", "", "")

	t.Run("identical rows reject the whole save", func(t *testing.T) {
		h, db := newDraftRevisionDB(t)
		r := newDraftRevisionRouter(h)
		stampDraftRevision(t, db, "rev-1")
		opponent := intakeInplacePartyJSON("对手公司", "opposing_party", "", "")
		body := intakeInplaceBody(clientParty, opponent, opponent)
		code, payload := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", body, "")
		// A duplicate pointer is a client-side protocol error, so the save
		// must fail closed with a 4xx before any row is written.
		if code < http.StatusBadRequest || code >= http.StatusInternalServerError {
			t.Fatalf("duplicate pointers must reject the save with a client error, got status=%d payload=%v", code, payload)
		}
		assertIntakeDraftPartiesUntouched(t, db)
	})

	t.Run("canonical and legacy opponent roles point at one row", func(t *testing.T) {
		h, db := newDraftRevisionDB(t)
		r := newDraftRevisionRouter(h)
		stampDraftRevision(t, db, "rev-1")
		legacy := intakeInplacePartyJSON("对手公司", "opponent", "SOCIAL_CREDIT_CODE", "91310000MA1FL0000X")
		canonical := intakeInplacePartyJSON("对手公司", "opposing_party", "SOCIAL_CREDIT_CODE", "91310000MA1FL0000X")
		body := intakeInplaceBody(clientParty, legacy, canonical)
		code, payload := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", body, "")
		if code < http.StatusBadRequest || code >= http.StatusInternalServerError {
			t.Fatalf("role-equivalent duplicate pointers must reject the save with a client error, got status=%d payload=%v", code, payload)
		}
		assertIntakeDraftPartiesUntouched(t, db)
	})
}

func TestUpdateIntakeRealIdentityChangeUpdatesSameRow(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")
	if err := db.Exec(`UPDATE case_intake_parties SET relation_depth = 3, metadata = '{"opaque":9007199254740993}' WHERE id = 2`).Error; err != nil {
		t.Fatalf("enrich stored opponent: %v", err)
	}
	var before map[string]interface{}
	if err := db.Table("case_intake_parties").Where("id = ?", 2).Take(&before).Error; err != nil {
		t.Fatalf("load stored opponent: %v", err)
	}
	// An enforced foreign key makes delete/reinsert fail even though the old
	// write path copies the original id back into the recreated row.
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}
	if err := db.Exec("CREATE TABLE intake_party_inplace_pins (party INTEGER NOT NULL REFERENCES case_intake_parties(id))").Error; err != nil {
		t.Fatalf("create pin table: %v", err)
	}
	if err := db.Exec("INSERT INTO intake_party_inplace_pins (party) VALUES (2)").Error; err != nil {
		t.Fatalf("pin opponent row: %v", err)
	}

	const newNumber = "91310000MA1FL7777X"
	clientParty := intakeInplacePartyJSON("星河智联科技有限公司", "client", "", "")
	opponent := intakeInplacePartyJSON("对手公司", "opposing_party", "SOCIAL_CREDIT_CODE", newNumber)
	body := intakeInplaceBody(clientParty, opponent)
	code, payload := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", body, "")
	if code != http.StatusOK {
		t.Fatalf("a real identity change must save in place, got status=%d payload=%v", code, payload)
	}

	var pins int64
	if err := db.Table("intake_party_inplace_pins").Count(&pins).Error; err != nil {
		t.Fatalf("count pins: %v", err)
	}
	if pins != 1 {
		t.Fatalf("the pinned party row must never be deleted, got %d pins", pins)
	}
	var rows []map[string]interface{}
	if err := db.Table("case_intake_parties").Where("intake_id = ? AND party_role = ?", "intake-rev", "opposing_party").Order("id").Limit(2).Find(&rows).Error; err != nil {
		t.Fatalf("reload opponent: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("exactly one opponent row must remain, got %d", len(rows))
	}
	row := rows[0]
	if fmt.Sprint(row["id"]) != fmt.Sprint(before["id"]) {
		t.Fatalf("identity update must keep the stored row id, got %v want %v", row["id"], before["id"])
	}
	wantDigest, err := security.IdentityDigest(security.NormalizeIdentityNumber("SOCIAL_CREDIT_CODE", newNumber))
	if err != nil {
		t.Fatalf("derive digest: %v", err)
	}
	if fmt.Sprint(row["identity_number_digest"]) != wantDigest {
		t.Fatalf("identity update must store the new lookup digest, got %v want %s", row["identity_number_digest"], wantDigest)
	}
	cipher := fmt.Sprint(row["identity_number_ciphertext"])
	if cipher == "" || cipher == "opp-enc" {
		t.Fatalf("identity update must re-encrypt the protected value, got %q", cipher)
	}
	if fmt.Sprint(row["relation_depth"]) != "3" {
		t.Fatalf("unchanged relation depth must survive the update, got %v", row["relation_depth"])
	}
	if fmt.Sprint(row["aliases"]) != "对手别名" {
		t.Fatalf("unchanged aliases must survive the update, got %v", row["aliases"])
	}
	if !strings.Contains(fmt.Sprint(row["metadata"]), "9007199254740993") {
		t.Fatalf("opaque metadata must survive the update, got %v", row["metadata"])
	}
	if fmt.Sprint(row["created_at"]) != fmt.Sprint(before["created_at"]) {
		t.Fatalf("created_at must survive the update, got %v want %v", row["created_at"], before["created_at"])
	}
	var clientRow map[string]interface{}
	if err := db.Table("case_intake_parties").Where("intake_id = ? AND party_role = ?", "intake-rev", "client").Take(&clientRow).Error; err != nil {
		t.Fatalf("reload client party: %v", err)
	}
	if fmt.Sprint(clientRow["identity_number_digest"]) != draftRevDigestA {
		t.Fatalf("client snapshot identity must stay untouched, got %v", clientRow["identity_number_digest"])
	}
	rev, _, _, _ := draftRevisionServerState(t, db)
	if rev == "" || rev == "rev-1" {
		t.Fatalf("successful save must rotate the revision, got %q", rev)
	}
}

// intakeInplacePartyJSON builds one payload row for the tests above.
// identityType and identityNumber may be empty for an unchanged echo.
func intakeInplacePartyJSON(name, role, identityType, identityNumber string) string {
	parts := []string{
		`"entity_name":"` + name + `"`,
		`"entity_type":"LEGAL_PERSON"`,
		`"party_role":"` + role + `"`,
	}
	if identityType != "" {
		parts = append(parts, `"identity_type":"`+identityType+`"`, `"identity_number":"`+identityNumber+`"`)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// intakeInplaceBody wraps payload rows in the native full-save request.
func intakeInplaceBody(parties ...string) string {
	return `{"expected_revision":"rev-1","parties":[` + strings.Join(parties, ",") + `]}`
}
