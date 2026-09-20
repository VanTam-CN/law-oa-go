package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// terminalBindingReleaser keeps the intake→approval bind row in sync with
// the approval lifecycle. The bind row (approval_case_creation_tracking)
// plus the UNIQUE index on intake_id is the durable duplicate guard: while
// an approval can still produce a case the row must stay. But when the
// approval reaches a state that ends its current lifecycle, the row must
// be released — otherwise the same intake can never be resubmitted
// (probe-reproduced permanent 409 INTAKE_APPROVAL_ALREADY_IN_PROGRESS on
// 58a91ed). Lifecycle coverage:
//
//   - ProcessApprovalDecision reject        → release (resubmit revives the
//     SAME approval, so ResubmitApproval rebinds before any later decision
//     can create a case);
//   - ProcessApprovalDecision requestChanges → release (same revival path);
//   - CancelApproval                        → release (cancelled is a state
//     machine terminal state; the only path back is a NEW approval through
//     the intake submission, which owns the UNIQUE slot itself);
//   - ResubmitApproval                      → rebind (UNIQUE rejects the
//     rebinding if a competing submission took the slot in between, so the
//     revived approval can never create a second case for the intake);
//   - approved                              → keep (the bind row doubles as
//     the durable case link);
//   - expired                               → no runtime transition writes
//     this status today, so there is nothing to hook; documented as an
//     evidence boundary, not a fix.
type terminalBindingReleaser struct {
	integrationRepo *repositories.IntegrationRepository
}

func newTerminalBindingReleaser(db *gorm.DB) *terminalBindingReleaser {
	return &terminalBindingReleaser{integrationRepo: repositories.NewIntegrationRepository(db)}
}

// releaseTerminalBinding drops the bind row for terminal decisions.
// Best-effort by design: a delete failure leaves the row in place, the
// lawyer sees the same 409 contract as before, and the release can be
// retried on the next terminal transition. Deleting the row can never
// widen the duplicate window because the live-approval guard and the
// UNIQUE intake_id index remain in force.
func (r *terminalBindingReleaser) release(ctx context.Context, approvalID string) {
	if r == nil || r.integrationRepo == nil {
		return
	}
	id := strings.TrimSpace(approvalID)
	if id == "" {
		return
	}
	if err := r.integrationRepo.DeleteCaseCreationTrackingByApprovalID(ctx, id); err != nil {
		log.Printf("终态释放接案绑定记录失败（保留绑定以fail-closed，可在下次终态重试）：%v", err)
	}
}

// rebindAfterResubmit restores the bind row when a rejected/needs-revision
// approval is revived. It is fail-closed against a competing submission: if
// another live approval already owns the intake slot, the UNIQUE index
// rejects this insert and the revived approval simply cannot create a case
// for that intake any more. A missing intake binding (approval not created
// through the intake path) is a no-op.
func (r *terminalBindingReleaser) rebindAfterResubmit(ctx context.Context, approval *models.ApprovalRequest) {
	if r == nil || r.integrationRepo == nil || approval == nil {
		return
	}
	metadata := make(map[string]interface{})
	if strings.TrimSpace(approval.Metadata) != "" {
		if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err != nil {
			log.Printf("重新提交时解析审批元数据失败，跳过接案重绑（审批ID=%s）：%v", approval.ID, err)
			return
		}
	}
	bound := metadataString(metadata, "intake_id")
	if bound == "" {
		return
	}
	tracking := &repositories.CaseCreationTracking{
		ID:                  fmt.Sprintf("INTAKE_BIND_%s", approval.ID),
		ApprovalRequestID:   approval.ID,
		IntakeID:            &bound,
		CreationStatus:      "pending",
		ErrorDetails:        "{}",
		DataMapping:         "{}",
		MappedFields:        "{}",
		UnmappedFields:      "{}",
		AppliedConditions:   "{}",
		ImposedRequirements: "{}",
		WorkflowActions:     "[]",
		CreatedBy:           approval.UpdatedBy,
	}
	if err := r.integrationRepo.CreateCaseCreationTracking(ctx, tracking); err != nil {
		log.Printf("重新提交后重绑接案失败（UNIQUE守卫拒绝，fail-closed，审批ID=%s，接案=%s）：%v", approval.ID, bound, err)
	}
}
