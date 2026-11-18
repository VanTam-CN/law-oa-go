package handlers

import (
	"encoding/json"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
	"net/http"
	"testing"
)

// D-item regression: the conflict queue must distinguish a frozen machine
// suggestion (REVIEW_REQUIRED in check_result) from the CURRENT human review
// state. A check with a saved conflict_reviews row must surface review_state
// so the UI stops counting completed reviews as "待人工复核"; the frozen
// check_result JSON must stay byte-identical.
func TestRiskRowsProjectCurrentReviewState(t *testing.T) {
	db := newConflictQueueReviewStateDB(t)
	handler := NewDemoAggregateHandler(db)
	handler.SetAuthorizationService(services.NewAuthorizationService(
		repositories.NewCaseRepository(db), nil, nil, nil,
	))
	context, recorder := conflictApprovalWorkbenchContext(
		"/api/v1/dashboard/command-center?conflict_scope=all", 35,
	)
	handler.CommandCenter(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("command center: %d %s", recorder.Code, recorder.Body.String())
	}
	rows := handler.riskRows(context, 8, true)
	byID := map[string]map[string]interface{}{}
	for _, row := range rows {
		byID[row["id"].(string)] = row
	}

	reviewed, ok := byID["CHECK-REVIEWED"]
	if !ok {
		t.Fatalf("reviewed check missing from queue: %v", byID)
	}
	if state, _ := reviewed["review_state"].(string); state != "reviewed" {
		t.Fatalf("reviewed check review_state=%v want reviewed", reviewed["review_state"])
	}
	if decision, _ := reviewed["review_decision"].(string); decision != "no_conflict" {
		t.Fatalf("reviewed check review_decision=%v want no_conflict", reviewed["review_decision"])
	}

	pending, ok := byID["CHECK-PENDING"]
	if !ok {
		t.Fatalf("pending check missing from queue: %v", byID)
	}
	if state, _ := pending["review_state"].(string); state != "pending_review" {
		t.Fatalf("pending check review_state=%v want pending_review", pending["review_state"])
	}

	// The frozen machine suggestion must stay intact for audit: the queue
	// projection reads it, never writes it.
	var frozen map[string]interface{}
	if err := json.Unmarshal([]byte(byID["CHECK-REVIEWED"]["check_result"].(string)), &frozen); err != nil {
		t.Fatalf("frozen check_result unreadable: %v", err)
	}
	decision := frozen["decision"].(map[string]interface{})
	if decision["status"] != "REVIEW_REQUIRED" {
		t.Fatalf("frozen machine decision changed: %v", decision)
	}
}

// Visibility keeps the pre-existing contract: the queue itself is shared
// within the allowed role scope (see TestConflictOfficerSeesOnlyAssigned...:
// approval links, not queue rows, are assignment-filtered). The projection
// adds no visibility of its own: a lawyer who owns the intake/case sees the
// review_state of their own checks, and nothing here widens the queue.
func TestRiskRowsReviewStateScopedToAssignedReviewer(t *testing.T) {
	db := newConflictQueueReviewStateDB(t)
	handler := NewDemoAggregateHandler(db)
	handler.SetAuthorizationService(services.NewAuthorizationService(
		repositories.NewCaseRepository(db), nil, nil, nil,
	))
	lawyerContext, _ := conflictQueueLawyerContext("/", 34)
	rows := handler.riskRows(lawyerContext, 8, false)
	if len(rows) == 0 {
		t.Fatalf("owning lawyer lost their own queue rows")
	}
	for _, row := range rows {
		state, _ := row["review_state"].(string)
		switch row["id"] {
		case "CHECK-REVIEWED":
			if state != "reviewed" {
				t.Fatalf("own lawyer sees reviewed check as %q", state)
			}
		case "CHECK-PENDING":
			if state != "pending_review" {
				t.Fatalf("own lawyer sees pending check as %q", state)
			}
		}
	}
}
