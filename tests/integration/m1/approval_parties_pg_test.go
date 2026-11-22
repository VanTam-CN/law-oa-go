//go:build np_m1_pg

package m1_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"law-oa-go/internal/models"
)

func TestM1R3ApprovalPartySnapshot(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	checkID := uuid.NewString()
	insertConflictPGRecord(t, db, f, checkID, f.revision)
	if err := db.Exec("UPDATE case_intakes SET status='conflict_ready', metadata=metadata||jsonb_build_object('conflict_check_id',?::text) WHERE id=?", checkID, f.intakeID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO case_intake_parties(intake_id,entity_name,party_role,entity_type,metadata,identity_number_ciphertext)
 SELECT ?, 'Party '||n, CASE WHEN n=1 THEN 'opponent' ELSE 'related_party' END,'LEGAL_PERSON','{"party_type":"affiliate","private":"must-stay-private"}','must-stay-private' FROM generate_series(1,23) n`, f.intakeID).Error; err != nil {
		t.Fatal(err)
	}
	admin := models.User{Username: "review-" + checkID, Name: "Fixture approver", Email: checkID + "@invalid.example", Password: "disabled", Role: "admin", Status: "active"}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	metadata := map[string]any{"intake_id": f.intakeID, "intake_revision": f.revision, "conflict_check_id": checkID, "parties": []any{map[string]any{"role": "related_party", "name": "Browser placeholder"}}}
	req := map[string]any{"type": "case_creation", "workflow_type": "CASE_APPROVAL", "title": "Fixture approval", "content": "Fixture content", "department_name": "Fixture", "applicant_name": "Fixture", "case_creation_config": map[string]any{}, "metadata": metadata}
	const path = "/api/v1/integration/approvals/with-conflict"
	w := send("POST", path, req)
	if w.Code != 200 {
		t.Fatalf("approval HTTP %d: %.300s", w.Code, w.Body.String())
	}
	var response struct {
		Data struct {
			ID string `json:"approval_id"`
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := db.Table("approval_requests").Select("metadata").Where("id=?", response.Data.ID).Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Parties          []map[string]string
		ApprovalSnapshot struct {
			Metadata struct{ Parties []map[string]string }
		} `json:"approval_snapshot"`
	}
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		t.Fatal(err)
	}
	for _, parties := range [][]map[string]string{snapshot.Parties, snapshot.ApprovalSnapshot.Metadata.Parties} {
		names := map[string]string{}
		for _, p := range parties {
			names[p["name"]] = p["role"]
		}
		if len(parties) != 24 || names["Party 23"] != "related_party" || names["Party 1"] != "opposing_party" || names["M1 conflict synthetic client"] != "client" || strings.Contains(raw, "must-stay-private") {
			t.Fatalf("stored approval has wrong or unsafe party list, total=%d", len(parties))
		}
	}
	if !reflect.DeepEqual(snapshot.Parties, snapshot.ApprovalSnapshot.Metadata.Parties) {
		t.Fatal("frozen list differs from approval metadata")
	}
	for n := 1; n <= 23; n++ {
		if snapshot.Parties[n]["name"] != fmt.Sprintf("Party %d", n) || snapshot.Parties[n]["party_type"] != "affiliate" {
			t.Fatal("party ordering or label lost", n)
		}
	}
	metadata["intake_revision"] = "stale"
	if w := send("POST", path, req); w.Code != 409 {
		t.Errorf("stale approval HTTP %d", w.Code)
	}
	var count int64
	if err := db.Table("approval_requests").Where("metadata::jsonb->>'intake_id' = ?", f.intakeID).Count(&count).Error; err != nil || count != 1 {
		t.Fatal("unexpected persisted approval count", count, err)
	}
}
