package services

import (
 "context"
 "strconv"
 "testing"

 "github.com/stretchr/testify/require"
 "law-oa-go/internal/models"
)

func TestAstraCompletedApprovalMustCheckActualParties(t *testing.T) {
 f := seedApprovalCaseWorld(t)
 first, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
 require.NoError(t, err)
 id, err := strconv.ParseUint(first.CaseID, 10, 64)
 require.NoError(t, err)
 require.NoError(t, f.db.Where("case_id = ?", id).Delete(&models.CaseParty{}).Error)
 // Refresh the approval returned by the fixture, as a real DB lookup would.
 require.NoError(t, f.db.First(f.approval, "id = ?", "APR-TX-1").Error)
 require.True(t, f.approval.CaseCreated)
 result, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
 if err != nil { return } // Explicit rejection is also a safe contract.
 require.NotNil(t, result)
 var count int64
 require.NoError(t, f.db.Model(&models.CaseParty{}).Where("case_id = ?", id).Count(&count).Error)
 require.EqualValues(t, 2, count, "a completed response must not bypass actual party integrity")
}

func TestAstraUnavailableTransactionMustNotStrandClaim(t *testing.T) {
 f := seedApprovalCaseWorld(t)
 f.svc.(*approvalConflictIntegrationService).db = nil
 _, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
 require.Error(t, err)
 var approval models.ApprovalRequest
 require.NoError(t, f.db.First(&approval, "id = ?", "APR-TX-1").Error)
 require.NotEqual(t, "processing", approval.CaseCreationStatus, "failed transaction setup must leave the approval retryable")
}
