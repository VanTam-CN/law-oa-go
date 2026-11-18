package handlers

import (
	"testing"

	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// Regression for the independent read-failure probe: when the conflict_reviews
// table exists but the row cannot be read (fixture breaks the decision column
// contract), the queue must fail closed — review_state stays unknown, rows
// stay visible, and the frozen check_result bytes are untouched. Silently
// prefilling pending_review would dress an unknown state up as a verified
// "waiting for human review" conclusion.
func TestRiskRowsUnreadableReviewsAreNotKnownPending(t *testing.T) {
	db := newConflictQueueReviewStateDB(t)
	var frozenBefore string
	if err := db.Table("conflict_check_records").Select("check_result").Where("check_id = ?", "CHECK-REVIEWED").Row().Scan(&frozenBefore); err != nil {
		t.Fatal(err)
	}
	// Break only the local fixture's read contract, never a real database.
	if err := db.Exec("ALTER TABLE conflict_reviews RENAME COLUMN decision TO unavailable_decision").Error; err != nil {
		t.Fatal(err)
	}
	handler := NewDemoAggregateHandler(db)
	handler.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), nil, nil, nil))
	ctx, _ := conflictQueueLawyerContext("/", 34)
	rows := handler.riskRows(ctx, 8, false)
	if len(rows) != 2 {
		t.Fatalf("read failure must not hide visible queue rows: got %d", len(rows))
	}
	for _, row := range rows {
		state, _ := row["review_state"].(string)
		if state == "pending_review" || state == "reviewed" {
			t.Errorf("unreadable review for %v falsely projected as verified state %q", row["id"], state)
		}
	}
	var frozenAfter string
	if err := db.Table("conflict_check_records").Select("check_result").Where("check_id = ?", "CHECK-REVIEWED").Row().Scan(&frozenAfter); err != nil {
		t.Fatal(err)
	}
	if frozenBefore != frozenAfter {
		t.Fatal("review-state projection changed frozen machine evidence")
	}
}
