package services

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// The completed fast path must consult the REAL approval row (fresh
// CaseCreated=true) and then verify/repair the case's actual party set on
// the same case — never fake a completion over a stripped case.
func TestAstraFixCompletedFreshApprovalRepairsParties(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	first, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	caseID, parseErr := strconv.ParseUint(first.CaseID, 10, 64)
	require.NoError(t, parseErr)
	require.NoError(t, f.db.Where("case_id = ?", caseID).Delete(&models.CaseParty{}).Error)
	// Refresh the fixture's approval pointer from the DB so the probe
	// returns the real post-commit state (CaseCreated=true).
	require.NoError(t, f.db.First(f.approval, "id = ?", "APR-TX-1").Error)
	require.True(t, f.approval.CaseCreated)

	second, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err, "fresh completed approval must repair parties on the same case")
	require.NotNil(t, second)
	assert.Equal(t, "completed", second.Status)
	assert.Equal(t, first.CaseID, second.CaseID, "repair must stay on the same case")
	var partyCount, caseCount int64
	require.NoError(t, f.db.Model(&models.CaseParty{}).Where("case_id = ?", caseID).Count(&partyCount).Error)
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	assert.EqualValues(t, 2, partyCount)
	assert.EqualValues(t, 1, caseCount)
}

// Any rolled-back attempt must leave the approval retryable: the claim now
// lives inside the transaction, so no failure path persists "processing".
func TestAstraFixRolledBackClaimIsRetryable(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	f.tracker.failTrackingCreate = func(_ *repositories.CaseCreationTracking) error {
		return fmt.Errorf("injected tracking failure")
	}
	_, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.Error(t, err)

	var approval models.ApprovalRequest
	require.NoError(t, f.db.First(&approval, "id = ?", "APR-TX-1").Error)
	assert.NotEqual(t, "processing", approval.CaseCreationStatus,
		"a rolled-back transaction must not strand the claim as processing")
	assert.False(t, trackingCompleted(f.db, t))

	// The retry must succeed on the same approval.
	f.tracker.failTrackingCreate = nil
	result, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "completed", result.Status)
}

// Concurrent attempts with a transactional claim: at most one case, and no
// attempt may leave the approval stranded in processing.
func TestAstraFixConcurrentClaimsNoStrand(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	const attempts = 4
	var wg sync.WaitGroup
	errs := make([]error, attempts)
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
		}(i)
	}
	wg.Wait()

	var caseCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	assert.EqualValues(t, 1, caseCount, "transactional claim must yield exactly one case")

	successCount := 0
	for i := range errs {
		if errs[i] == nil {
			successCount++
		}
	}
	assert.GreaterOrEqual(t, successCount, 1)

	var approval models.ApprovalRequest
	require.NoError(t, f.db.First(&approval, "id = ?", "APR-TX-1").Error)
	assert.NotEqual(t, "processing", approval.CaseCreationStatus,
		"completed or failed attempts must never leave a stranded processing claim")
}
