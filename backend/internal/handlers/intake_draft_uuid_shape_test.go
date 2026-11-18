package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"gorm.io/gorm"
)

// uuidShapeDialector reports a dialect name without executing any query, so
// only expression building is observable here.
type uuidShapeDialector struct{ gorm.Dialector }

func (d uuidShapeDialector) Name() string { return "postgres" }

// A uuid-shaped identifier must behave exactly like the repaired real A3:
// the completed tracking row hides its own intake and nothing else.
func TestListIntakeDraftsKeepsUuidShapedIdsRecoverable(t *testing.T) {
	db := newIntakeDraftListDB(t)
	statements := []string{
		"CREATE TABLE approval_requests (id TEXT PRIMARY KEY, type TEXT, status TEXT, metadata TEXT, deleted_at DATETIME)",
		"CREATE TABLE approval_case_creation_tracking (id TEXT PRIMARY KEY, approval_request_id TEXT, intake_id TEXT, creation_status TEXT)",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create approval table: %v", err)
		}
	}
	if err := db.Exec("INSERT INTO approval_requests (id, type, status, metadata) VALUES ('ap-uuid', 'case_creation', 'approved', '{}')").Error; err != nil {
		t.Fatalf("seed approved approval: %v", err)
	}
	const uuidID = "a572241f-1111-4111-8111-111111111111"
	if err := db.Exec("INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, creation_status) VALUES ('tr-uuid', 'ap-uuid', ?, 'completed')", uuidID).Error; err != nil {
		t.Fatalf("seed uuid-shaped tracking: %v", err)
	}
	if err := db.Exec("UPDATE case_intakes SET id = ? WHERE id = 'i-1'", uuidID).Error; err != nil {
		t.Fatalf("uuid-shape intake id: %v", err)
	}
	r := newIntakeDraftListRouter(db)
	code, data := listIntakeDrafts(t, r, "lawyer", "7", "")
	if code != http.StatusOK {
		t.Fatalf("uuid-shaped list: status=%d want 200", code)
	}
	ids := map[string]bool{}
	for _, raw := range data["items"].([]interface{}) {
		row, _ := raw.(map[string]interface{})
		ids[fmt.Sprint(row["id"])] = true
	}
	if ids[uuidID] {
		t.Fatalf("uuid-shaped completed creation must hide its intake: ids=%v", ids)
	}
	if !ids["i-2"] || !ids["i-4"] {
		t.Fatalf("uuid-shaped comparison must not hide unrelated drafts: ids=%v", ids)
	}
}
