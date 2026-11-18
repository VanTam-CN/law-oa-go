package services

import (
	"encoding/json"
	"strings"
	"testing"

	"law-oa-go/internal/models"
)

// Independent proof: an identity digest alone is not the complete party
// facts. Role, type, name and the exact subject set belong to the check.
func TestIndependentContinuationRequiresSameSubjectFacts(t *testing.T) {
	clientDigest := strings.Repeat("a", 64)
	partyDigest := strings.Repeat("b", 64)
	frozen, err := json.Marshal(map[string]interface{}{
		"clientId": "9", "clientName": "虚构客户甲", "clientType": "PERSON",
		"clientIdentifiers": map[string]string{"id_card": "hmac-sha256:" + clientDigest},
		"parties": []map[string]interface{}{{
			"name": "虚构对手公司", "role": "OPPOSING_PARTY", "entityType": "COMPANY",
			"identifiers": map[string]string{"social_credit_code": "hmac-sha256:" + partyDigest},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	base := []continuationSubjectEvidence{
		{Subject: ContinuationSubject{Kind: "client", Name: "虚构客户甲", EntityType: "PERSON", IdentityType: "ID_CARD"}, IdentityDigest: clientDigest},
		{Subject: ContinuationSubject{Kind: "party", Name: "虚构对手公司", Role: "OPPOSING_PARTY", EntityType: "COMPANY", IdentityType: "SOCIAL_CREDIT_CODE"}, IdentityDigest: partyDigest},
	}
	if err := verifyContinuationFrozenIdentity(string(frozen), base); err != nil {
		t.Fatalf("same frozen facts must pass: %v", err)
	}
	tests := []struct {
		name   string
		change func([]continuationSubjectEvidence) []continuationSubjectEvidence
	}{
		{"changed_party_role", func(s []continuationSubjectEvidence) []continuationSubjectEvidence {
			s[1].Subject.Role = "CO_CLIENT"
			return s
		}},
		{"changed_party_name", func(s []continuationSubjectEvidence) []continuationSubjectEvidence {
			s[1].Subject.Name = "另一个虚构名称"
			return s
		}},
		{"changed_party_type", func(s []continuationSubjectEvidence) []continuationSubjectEvidence {
			s[1].Subject.EntityType = "PERSON"
			return s
		}},
		{"missing_original_party", func(s []continuationSubjectEvidence) []continuationSubjectEvidence { return s[:1] }},
		{"changed_client_name", func(s []continuationSubjectEvidence) []continuationSubjectEvidence {
			s[0].Subject.Name = "另一个虚构客户"
			return s
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := append([]continuationSubjectEvidence(nil), base...)
			if err := verifyContinuationFrozenIdentity(string(frozen), test.change(current)); err == nil {
				t.Fatal("changed subject facts must not reuse the old check")
			}
		})
	}
}

func TestIndependentContinuationDatabaseFailureNeverGrantsReuse(t *testing.T) {
	db := seedContinuationWorld(t)
	seedContinuationIntake(t, db, "INT-INDEP-DB", continuationSeedOptions{withReview: true})
	if err := db.Exec("DROP TABLE conflict_reviews").Error; err != nil {
		t.Fatal(err)
	}
	decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation("INT-INDEP-DB")
	if err == nil && decision != nil && (decision.CanReuseCheck || decision.CanSubmit) {
		t.Fatalf("database failure must not grant continuation: %+v", decision)
	}
}

func TestIndependentContinuationUsesActiveSnapshotNotNewest(t *testing.T) {
	db := seedContinuationWorld(t)
	seedContinuationIntake(t, db, "INT-INDEP-SNAP", continuationSeedOptions{withReview: true})
	active := "independent-active-round"
	approval := models.ApprovalRequest{
		ID: "independent-approval", RequestNumber: "independent-number",
		Title: "虚构接案审批", Type: "CASE_CREATION", Status: models.ApprovalStatusNeedsRevision,
		Metadata:                   `{"intake_id":"INT-INDEP-SNAP"}`,
		ActiveSubmissionSnapshotID: &active,
	}
	if err := db.Create(&approval).Error; err != nil {
		t.Fatal(err)
	}
	for round, id := range []string{active, "independent-newer-unactivated-round"} {
		if err := db.Exec("INSERT INTO approval_snapshots (id, approval_request_id, snapshot_type, snapshot_data, source_version) VALUES (?, ?, ?, ?, ?)",
			id, approval.ID, models.ApprovalSnapshotTypeIntakeSubmission, `{"intake_id":"INT-INDEP-SNAP"}`, round+1).Error; err != nil {
			t.Fatal(err)
		}
	}
	decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation("INT-INDEP-SNAP")
	if err != nil {
		t.Fatal(err)
	}
	if decision.Approval == nil || decision.Approval.SnapshotID != active || decision.Approval.Round != 1 {
		t.Fatalf("must use active pointer, not newer row: %+v", decision.Approval)
	}
	if err := db.Model(&approval).Update("active_submission_snapshot_id", "missing-frozen-round").Error; err != nil {
		t.Fatal(err)
	}
	decision, err = NewIntakeContinuationService(db).EvaluateIntakeContinuation("INT-INDEP-SNAP")
	if err == nil && decision != nil && decision.CanSubmit {
		t.Fatal("broken active pointer must not be replaced with the latest snapshot")
	}
}

func TestIndependentContinuationResolvesFormattedJSONApprovalBinding(t *testing.T) {
	db := seedContinuationWorld(t)
	intakeID := "INT-INDEPENDENT-FORMATTED"
	seedContinuationIntake(t, db, intakeID, continuationSeedOptions{withReview: true})
	active := "independent-formatted-snapshot"
	approval := models.ApprovalRequest{ID: "independent-formatted-approval", RequestNumber: "independent-formatted-number", Title: "虚构原审批", Type: "CASE_CREATION", Status: models.ApprovalStatusNeedsRevision,
		Metadata: "{\n  \"intake_id\": \"" + intakeID + "\"\n}", ActiveSubmissionSnapshotID: &active}
	if err := db.Create(&approval).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO approval_snapshots (id, approval_request_id, snapshot_type, snapshot_data, source_version) VALUES (?, ?, ?, ?, ?)", active, approval.ID, models.ApprovalSnapshotTypeIntakeSubmission, `{"intake_id":"`+intakeID+`"}`, 1).Error; err != nil {
		t.Fatal(err)
	}
	decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation(intakeID)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Approval == nil || decision.Approval.ApprovalID != approval.ID || decision.Approval.SnapshotID != active {
		t.Fatalf("valid JSON whitespace must not hide the original approval: %+v", decision.Approval)
	}
}

func TestIndependentContinuationDoesNotResubmitAnAlreadyOwnedApproval(t *testing.T) {
	for _, status := range []string{models.ApprovalStatusNeedsRevision, models.ApprovalStatusSubmitted, models.ApprovalStatusUnderReview, models.ApprovalStatusResubmitted, models.ApprovalStatusApproved, models.ApprovalStatusRejected, models.ApprovalStatusCancelled, models.ApprovalStatusExpired, models.ApprovalStatusDraft, "unrecognized_phase"} {
		t.Run(status, func(t *testing.T) {
			db := seedContinuationWorld(t)
			intakeID := "INT-INDEPENDENT-PHASE"
			seedContinuationIntake(t, db, intakeID, continuationSeedOptions{withReview: true})
			active := "independent-phase-snapshot"
			approval := models.ApprovalRequest{ID: "independent-phase-approval", RequestNumber: "independent-phase-number", Title: "虚构已有审批", Type: "CASE_CREATION", Status: status,
				Metadata: `{"intake_id":"` + intakeID + `"}`, ActiveSubmissionSnapshotID: &active}
			if err := db.Create(&approval).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("INSERT INTO approval_snapshots (id, approval_request_id, snapshot_type, snapshot_data, source_version) VALUES (?, ?, ?, ?, ?)", active, approval.ID, models.ApprovalSnapshotTypeIntakeSubmission, `{"intake_id":"`+intakeID+`"}`, 1).Error; err != nil {
				t.Fatal(err)
			}
			decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation(intakeID)
			if err != nil {
				t.Fatal(err)
			}
			if !decision.CanReuseCheck || decision.RecheckRequired {
				t.Fatalf("an unchanged check must not be forced to rerun just because approval is %s", status)
			}
			wantSubmit := status == models.ApprovalStatusNeedsRevision
			if decision.CanSubmit != wantSubmit {
				t.Fatalf("existing approval phase %q has CanSubmit=%v, want %v", status, decision.CanSubmit, wantSubmit)
			}
		})
	}
}
