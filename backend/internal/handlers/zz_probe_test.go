package handlers

import (
	"context"
	"testing"

	"law-oa-go/internal/services"
)

func TestZZProbeReviewerBinding(t *testing.T) {
	db := newMatterCollabDB(t)
	intake := "aaaa1111-0000-0000-0000-000000000001"
	checkID := "CCT_PROBE"
	meta := `{"intake_id":"` + intake + `"}`
	db.Exec("INSERT INTO approval_requests (id, metadata, current_approver_id, status, applicant_id) VALUES ('ap-probe', ?, '4', 'submitted', '1')", meta)
	db.Exec("UPDATE approval_requests SET conflict_check_id = ? WHERE id = 'ap-probe'", checkID)
	db.Exec("INSERT INTO users (id, role) VALUES (6, 'conflict_officer')")
	db.Exec("INSERT INTO conflict_reviewer_assignments (id, check_id, reviewer_id, status, recusal_declared) VALUES ('asg-p', ?, 6, 'ACTIVE', 1)", checkID)

	g := &materialAccessGuard{db: db}
	bound, err := g.actorBoundToIntakeWorkflow(intake, 6)
	t.Logf("bound=%v err=%v", bound, err)
	rows, _ := db.Raw("SELECT conflict_reviewer_assignments.id, approval_requests.metadata FROM conflict_reviewer_assignments JOIN approval_requests ON approval_requests.conflict_check_id = conflict_reviewer_assignments.check_id WHERE conflict_reviewer_assignments.reviewer_id = 6 AND conflict_reviewer_assignments.status = 'ACTIVE'").Rows()
	defer rows.Close()
	for rows.Next() {
		var id string
		var meta string
		_ = rows.Scan(&id, &meta)
		t.Logf("raw row id=%s meta=%q", id, meta)
	}

	authz := services.NewAuthorizationServiceWithDB(db)
	ok, err := authz.CanReadClient(context.Background(), services.AuthActor{UserID: 6, Role: "conflict_officer"}, 10)
	t.Logf("CanReadClient ok=%v err=%v", ok, err)
}
