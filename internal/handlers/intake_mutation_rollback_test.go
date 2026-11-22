package handlers

// Negative-boundary regressions for the native UpdateCaseIntake write path.
// Both tests reuse newDraftRevisionDB so they exercise the same single-connection
// SQLite fixture as the revision contract suite.

import (
	"fmt"
	"net/http"
	"testing"
)

// A parties insert that fails AFTER the intake row and old party rows have
// already been written must roll the whole save back: no half-saved draft,
// no lost parties, no rotated revision. The failure is forced with a
// test-scoped SQLite trigger, so no production hook is involved.
func TestUpdateIntakeTransactionFailureRollsBackFullSave(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")
	if err := db.Exec("CREATE TRIGGER force_party_insert_fail BEFORE INSERT ON case_intake_parties BEGIN SELECT RAISE(ABORT, 'forced rollback'); END").Error; err != nil {
		t.Fatalf("create rollback trigger: %v", err)
	}

	// The first row updates in place. Inserting the second must roll back
	// that update, the removed client row, and the parent revision.
	body := `{"expected_revision":"rev-1","title":"回滚前改标题","parties":[
        {"entity_name":"对手公司","entity_type":"LEGAL_PERSON","party_role":"opposing_party","identity_type":"SOCIAL_CREDIT_CODE","identity_number":"91310000MA1FL0000X"},
        {"entity_name":"新增合成主体","entity_type":"LEGAL_PERSON","party_role":"related_party","identity_type":"SOCIAL_CREDIT_CODE","identity_number":"SYNTHETIC-NEW"}
    ]}`
	code, payload := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", body, "")
	if code < http.StatusBadRequest {
		t.Fatalf("forced insert failure must reject the save, got status=%d payload=%v", code, payload)
	}

	rev, clientDigest, opponentDigest, title := draftRevisionServerState(t, db)
	if rev != "rev-1" || title != "版本守护草稿" {
		t.Fatalf("failed save must not mutate the intake row: rev=%q title=%q", rev, title)
	}
	if clientDigest != draftRevDigestA || opponentDigest != "opp-digest" {
		t.Fatalf("failed save must not mutate parties: client=%q opponent=%q", clientDigest, opponentDigest)
	}
	var partyCount int64
	if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-rev").Count(&partyCount).Error; err != nil {
		t.Fatalf("reload party count: %v", err)
	}
	if partyCount != 2 {
		t.Fatalf("rollback must restore the original two party rows, got %d", partyCount)
	}
}

// A draft whose client party row is absent stays usable when only client_id
// is switched: the save succeeds, no party rows are touched, and the draft
// can be saved again with the rotated revision.
func TestUpdateIntakeClientSwitchWithoutClientPartyKeepsOldDraftUsable(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")
	if err := db.Table("case_intake_parties").
		Where("intake_id = ? AND LOWER(TRIM(party_role)) = ?", "intake-rev", "client").
		Delete(map[string]interface{}{}).Error; err != nil {
		t.Fatalf("remove client party: %v", err)
	}

	code, saved := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev",
		"{\"expected_revision\":\"rev-1\",\"client_id\":999,\"title\":\"无主体草稿切换客户\"}", "")
	if code != http.StatusOK {
		t.Fatalf("client switch without client party must save, got status=%d payload=%v", code, saved)
	}
	data, _ := saved["data"].(map[string]interface{})
	switchedRev, _ := data["revision"].(string)
	if switchedRev == "" || switchedRev == "rev-1" {
		t.Fatalf("switch save must rotate the revision, got %q", switchedRev)
	}

	var row struct {
		ClientID *uint
	}
	if err := db.Table("case_intakes").Select("client_id").Where("id = ?", "intake-rev").Take(&row).Error; err != nil {
		t.Fatalf("reload intake: %v", err)
	}
	if row.ClientID == nil || *row.ClientID != 999 {
		t.Fatalf("client_id must switch to 999, got %v", row.ClientID)
	}
	rev, _, opponentDigest, _ := draftRevisionServerState(t, db)
	if rev != switchedRev {
		t.Fatalf("stored revision %q must equal response %q", rev, switchedRev)
	}
	if opponentDigest != "opp-digest" {
		t.Fatalf("client-only switch must not touch remaining parties, got opponent=%q", opponentDigest)
	}
	var partyCount int64
	if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-rev").Count(&partyCount).Error; err != nil {
		t.Fatalf("reload party count: %v", err)
	}
	if partyCount != 1 {
		t.Fatalf("only the opposing party row must remain, got %d", partyCount)
	}

	code, second := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev",
		fmt.Sprintf("{\"expected_revision\":%q,\"description\":\"切换后再次保存\"}", switchedRev), "")
	if code != http.StatusOK {
		t.Fatalf("switched draft must remain usable: status=%d payload=%v", code, second)
	}
}
