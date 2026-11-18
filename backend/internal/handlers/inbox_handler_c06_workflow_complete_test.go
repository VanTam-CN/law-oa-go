package handlers

import (
	"net/http"
	"strings"
	"testing"

	"law-oa-go/internal/models"
)

// C06 workflow action forgery retest (plan §5.4/§10): manual completion of
// a workflow inbox item must surface as a 409 business conflict with a
// stable error code -- never as a 500 with the raw internal message.
func TestInboxWorkflowItemManualCompleteReturnsConflict(t *testing.T) {
	router, db := newInboxOwnershipTestRouter(t, inboxOwnerUID)

	workflowItem := models.InboxItem{
		ID:              inboxOwnerID + 1,
		UserID:          inboxOwnerUID,
		SourceType:      "approval",
		SourceID:        3,
		CompletionMode:  "workflow",
		WorkflowTaskKey: "approval:approval_decision:APR-C06-RETEST:c0:11",
		Title:           "workflow handoff item",
		Priority:        "high",
	}
	if err := db.Create(&workflowItem).Error; err != nil {
		t.Fatalf("seed workflow inbox item: %v", err)
	}

	recorder := performInboxRequest(router, http.MethodPut, "/inbox/102/complete", "")
	if recorder.Code != http.StatusConflict {
		t.Fatalf("workflow complete status = %d, want %d; body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "INBOX_WORKFLOW_MANUAL_COMPLETE_FORBIDDEN") {
		t.Fatalf("stable business error code missing from response: %s", body)
	}
	if strings.Contains(body, "workflow items cannot be completed manually") {
		t.Fatalf("raw internal error leaked to client: %s", body)
	}

	var retained models.InboxItem
	if err := db.First(&retained, inboxOwnerID+1).Error; err != nil {
		t.Fatalf("load workflow inbox item: %v", err)
	}
	if retained.IsCompleted || retained.CompletedAt != nil {
		t.Fatalf("workflow inbox item was completed by rejected request: %+v", retained)
	}
}
