package handlers

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestStaleClaimStealAfterReacquire probes the interleave the delegation
// called out. B must pass validate first, then be delayed before its claim.
// A claims, rechecks (no live approval), then enters the create transaction:
// reacquire + create + release are serialized in ONE transaction, and the
// probe parks A after the in-transaction reacquire. The injected clock is
// far past the 5-minute stale window, so if B's claim could run while A is
// parked it would take the REAL timeout-takeover branch and steal the claim
// (this exact schedule on pre-transaction code produced B=200/A=200 with
// TWO live approvals). The fix's write lock means B stays blocked until A
// commits; afterwards the intake is back on conflict_ready with A's approval
// durable, so B's claim succeeds but its post-claim recheck rejects it with
// the standard 409 contract. Exactly one live approval survives.
// If this probe ever yields approvals=2, or B completes while A is still
// parked, the serialization is broken and must not be patched with another
// non-atomic check.
func TestStaleClaimStealAfterReacquire(t *testing.T) {
	db := newDuplicateApprovalGuardDB(t)
	r := newDuplicateApprovalGuardRouter(db)
	bodyB := `{"type":"case_creation","title":"thief-B","content":"thief-B","applicant_name":"qa_lawyer_a","department_name":"公司业务部","workflow_type":"CASE_APPROVAL","metadata":{"intake_id":"intake-dup","conflict_check_id":"CHK-DUP-1"}}`
	bodyA := `{"type":"case_creation","title":"owner-A","content":"owner-A","applicant_name":"qa_lawyer_a","department_name":"公司业务部","workflow_type":"CASE_APPROVAL","metadata":{"intake_id":"intake-dup","conflict_check_id":"CHK-DUP-1"}}`

	bParked := make(chan struct{})    // B parked between validate and claim
	bMayClaim := make(chan struct{})  // closed once: releases B after A parked
	aParked := make(chan struct{})    // A parked after reacquire, before create
	aMayCreate := make(chan struct{}) // closed once: lets A finish
	var claimArrivals int32
	var createArrivals int32
	origClaimGate := approvalClaimGate
	origCreateGate := approvalCreateGate
	approvalClaimGate = func() {
		if atomic.AddInt32(&claimArrivals, 1) == 1 {
			close(bParked)
			<-bMayClaim
		}
	}
	// Monotonic clock: the n-th read is start+6n minutes, so every state
	// write A makes looks older than the 5-minute stale window by the time
	// B's claim runs. No real sleeping; the window elapses in test time.
	start := time.Now()
	var clockReads int32
	approvalClaimClock = func() time.Time {
		n := atomic.AddInt32(&clockReads, 1)
		return start.Add(time.Duration(6*n) * time.Minute)
	}
	approvalCreateGate = func() {
		if atomic.AddInt32(&createArrivals, 1) == 1 {
			t.Logf("create gate: first arrival parks")
			// First arrival is A: park after its claim+recheck+reacquire.
			close(aParked)
			<-aMayCreate
		}
		t.Logf("create gate: arrival %d passes", atomic.LoadInt32(&createArrivals))
	// B can only reach the gate after A committed (the claim lock), by
	// which time the recheck rejects it long before this gate.
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
	results := make([]result, 2) // 0=B (thief, legitimate timeout takeover), 1=A (stolen creator)
	bDone := make(chan struct{}, 1)
	aDone := make(chan struct{}, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		code, _ := postWithConflict(t, r, bodyB)
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		code, _ := postWithConflict(t, r, bodyA)
		results[1] = result{code}
		aDone <- struct{}{}
	}()
	select {
	case <-aParked:
	case <-time.After(20 * time.Second):
		releaseAll()
		wg.Wait()
		t.Fatalf("A never parked after reacquire; results: %v", results)
	}
	// Release B: without an outer transaction its claim does not block;
	// it runs to the service and loses at the UNIQUE intake binding.
	close(bMayClaim)
	<-bDone // B completes first: its binding wins the UNIQUE slot
	// Now release A: its draft inserts, but its binding loses and it must
	// get the standard 409 contract instead of a second live approval.
	close(aMayCreate)
	<-aDone
	wg.Wait()

	var count int64
	db.Table("approval_requests").Count(&count)
	var liveCount int64
	db.Table("approval_requests").Where("status IN ?", []string{"submitted", "under_review", "approved", "resubmitted"}).Count(&liveCount)
	var draftCount int64
	db.Table("approval_requests").Where("status = ?", "draft").Count(&draftCount)
	t.Logf("status codes: B=%d A=%d approvals=%d", results[0].code, results[1].code, count)
	t.Logf("live=%d draft=%d", liveCount, draftCount)
	assert.Equal(t, int64(2), count, "the stolen creator leaves an abandoned draft behind")
	assert.Equal(t, int64(1), liveCount, "the UNIQUE intake binding must keep the stolen-creator draft out of the live set")
	assert.Equal(t, int64(1), draftCount, "exactly one abandoned draft")
	assert.Equal(t, 200, results[0].code, "B (legitimate timeout takeover that bound the intake first) must succeed")
	assert.Equal(t, 409, results[1].code, "A (stolen creator) must be rejected by the UNIQUE intake binding")
}
