package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// Test-only fixture: the HTTP test must perform the actual client edit before
// PrepareReviewedSubmission supplies the next detector/reviewer fixture inputs.
type InboxRevisionPostgresFixture struct {
	DB                        *gorm.DB
	Approval                  *models.ApprovalRequest
	IntakeID                  string
	OldClientID, NewClientID  uint
	State                     func() map[string]string
	PrepareReviewedSubmission func() *models.ResubmitApprovalRequest
}

func NewInboxRevisionPostgresFixture(t *testing.T) InboxRevisionPostgresFixture {
	t.Helper()
	f := newApprovalCollaborationPostgresFixture(t)
	request := f.returnForRevision(t)
	oldClient := f.clientID
	client := &models.Client{Name: "R1补正新客户", Type: "个人", CreatedBy: 11, IdentityType: models.IdentityTypeIDCard, IdentityNumber: "310110199001010044"}
	require.NoError(t, repositories.NewClientRepository(f.db).Create(context.Background(), client))
	return InboxRevisionPostgresFixture{DB: f.db, Approval: f.reload(t, f.approval.ID), IntakeID: f.intakeID,
		OldClientID: oldClient, NewClientID: client.ID, State: func() map[string]string { return f.state(t) },
		PrepareReviewedSubmission: func() *models.ResubmitApprovalRequest {
			f.reviseReviewedSubject(t, request)
			f.clientID, f.clientDigest = client.ID, client.IdentityNumberDigest
			var check models.ConflictCheckRecord
			require.NoError(t, f.db.First(&check, "check_id = ?", f.checkID).Error)
			check.ClientID, check.ClientName = fmt.Sprint(client.ID), client.Name
			check.SearchParameters["clientIdentifiers"] = map[string]interface{}{"id_card": "hmac-sha256:" + client.IdentityNumberDigest}
			require.NoError(t, f.db.Save(&check).Error)
			require.NoError(t, f.db.Table("case_intakes").Where("id = ?", f.intakeID).Update("status", "conflict_ready").Error)
			return request
		},
	}
}
