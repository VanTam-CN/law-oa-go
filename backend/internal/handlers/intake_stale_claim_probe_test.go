package handlers

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sync/atomic"
)

// TestStaleClaimStealWhileCreatorInProgress: the only interleave that
// reaches the claim while another creator is mid-flight. B must pass
// validate while the intake is still conflict_ready, then be delayed
// between validate and claim. A then claims (a pre-fix claim leaves the
// pre-claim updated_at untouched, which still looks stale), passes the
// post-claim recheck, and parks before approval creation. When B resumes,
// its claim classifies the live claim of A as a stale orphan and steals
// it; the recheck of B sees no live approval yet (A has not created), so
// B creates a second approval for the same intake. Post-fix the claim
// refreshes updated_at, so the claim of B finds 0 eligible rows and B
// gets 409. The injected approvalClaimClock (no backdating, no real
// sleep) puts this on the REAL timeout takeover path: the stale window
// elapses in test time while A holds a fresh claim. Channel close order
// (never re-closed) keeps cleanup safe.
func TestStaleClaimStealWhileCreatorInProgress(t *testing.T) {
	db := newDuplicateApprovalGuardDB(t)
	r := newDuplicateApprovalGuardRouter(db)
	// The intake sat in conflict_ready for a long time: any claim made
	// now starts from an already-stale updated_at on pre-fix code.
	if err := db.Exec("UPDATE case_intakes SET updated_at = ? WHERE id = 'intake-dup'",
		time.Now().Add(-30*time.Minute)).Error; err != nil {
		t.Fatalf("backdate intake: %v", err)
	}
	body := `{"type":"case_creation","title":"陈旧claim抢占样本","content":"陈旧claim抢占样本","applicant_name":"qa_lawyer_a","department_name":"公司业务部","workflow_type":"CASE_APPROVAL","metadata":{"intake_id":"intake-dup","conflict_check_id":"CHK-DUP-1"}}`

	bParked := make(chan struct{})    // B parked between validate and claim
	bMayClaim := make(chan struct{})  // closed once: releases B after A parked
	aParked := make(chan struct{})    // A parked after recheck, before create
	aMayCreate := make(chan struct{}) // closed once: lets parked creators run
	var claimArrivals int32
	var createArrivals int32
	origClaimGate := approvalClaimGate
	origCreateGate := approvalCreateGate
	approvalClaimGate = func() {
		if atomic.AddInt32(&claimArrivals, 1) == 1 {
			// First arrival is B: park before its claim until A parked.
			close(bParked)
			<-bMayClaim
		}
		// A (second arrival) passes through and claims immediately.
	}
	// Advance the claim clock far past the 5-minute stale window between
	// A parking and B resuming, so B goes through the REAL timeout
	// takeover path (the stale branch of the claim predicate). The intake
	// updated_at is not backdated here: the window elapses in test time
	// via the injected clock.
	fakeNow := time.Now()
	approvalClaimClock = func() time.Time { return fakeNow.Add(10 * time.Minute) }
	approvalCreateGate = func() {
		if atomic.AddInt32(&createArrivals, 1) == 1 {
			// First creator (A) parks after its claim+recheck.
			close(aParked)
			<-aMayCreate
		}
		// Any later creator (only possible if B stole the claim) runs.
	}
	releaseAll := func() {
		select {
		case <-bMayClaim:
		default:
			close(bMayClaim)
		}
		select {
		case <-aMayCreate:
		default:
			close(aMayCreate)
		}
		approvalClaimGate = origClaimGate
		approvalCreateGate = origCreateGate
		approvalClaimClock = func() time.Time { return time.Now() }
	}
	t.Cleanup(releaseAll)

	type result struct {
		code int
	}
	results := make([]result, 2) // 0=B (delayed), 1=A (creator)
	bDone := make(chan struct{}, 1)
	aDone := make(chan struct{}, 1)
	var wg sync.WaitGroup
	// B first: passes validate, parks before claim.
	wg.Add(1)
	go func() {
		defer wg.Done()
		code, _ := postWithConflict(t, r, body)
		results[0] = result{code}
		bDone <- struct{}{}
	}()
	select {
	case <-bParked:
	case <-time.After(20 * time.Second):
		releaseAll()
		wg.Wait()
		t.Fatalf("B never parked between validate and claim")
	}
	// A now: validate, claim, recheck, park pre-create.
	wg.Add(1)
	go func() {
		defer wg.Done()
		code, _ := postWithConflict(t, r, body)
		results[1] = result{code}
		aDone <- struct{}{}
	}()
	select {
	case <-aParked:
	case <-time.After(20 * time.Second):
		releaseAll()
		wg.Wait()
		t.Fatalf("A never parked before create; results: %v", results)
	}
	// Release B: its claim races the stale-looking live claim of A.
	close(bMayClaim)
	<-bDone
	// Only now let A finish creating.
	close(aMayCreate)
	<-aDone
	wg.Wait()

	var count int64
	db.Table("approval_requests").Count(&count)
	t.Logf("status codes: B=%d A=%d approvals=%d", results[0].code, results[1].code, count)
	assert.Equal(t, int64(1), count, "stale-claim steal must not create a second live approval")
	assert.Equal(t, 200, results[1].code, "A must succeed with exactly one approval")
	assert.Equal(t, 409, results[0].code, "B must be rejected while A holds a fresh claim")
}

// TestOrphanClaimNotSilentlyTakenOver documents the contract for a
// stranded claim (process died mid-flight): validate runs BEFORE the
// claim and requires status conflict_ready, so a new submission cannot
// take over an approval_submitting intake through the API no matter how
// old it is. The stale branch inside the claim is only reachable for
// requests that passed validate BEFORE the stranded claim existed (the
// delayed interleave above). Recovery from a crashed submit goes through
// the facts re-confirm + re-check workflow, which resets the intake to
// conflict_ready by design (F06).
func TestOrphanClaimNotSilentlyTakenOver(t *testing.T) {
	db := newDuplicateApprovalGuardDB(t)
	r := newDuplicateApprovalGuardRouter(db)
	if err := db.Exec("UPDATE case_intakes SET status = 'approval_submitting', updated_at = ? WHERE id = 'intake-dup'",
		time.Now().Add(-10*time.Minute)).Error; err != nil {
		t.Fatalf("seed orphan claim: %v", err)
	}
	body := `{"type":"case_creation","title":"孤儿claim接管样本","content":"孤儿claim接管样本","applicant_name":"qa_lawyer_a","department_name":"公司业务部","workflow_type":"CASE_APPROVAL","metadata":{"intake_id":"intake-dup","conflict_check_id":"CHK-DUP-1"}}`
	code, respBody := postWithConflict(t, r, body)
	t.Logf("orphan takeover attempt status: %d body=%v", code, respBody)
	assert.Equal(t, 409, code, "an approval_submitting intake must not accept new submissions")
	errMap, _ := respBody["error"].(map[string]interface{})
	errCode, _ := errMap["code"].(string)
	assert.Equal(t, "INTAKE_CONFLICT_RESULT_STALE", errCode)
}
