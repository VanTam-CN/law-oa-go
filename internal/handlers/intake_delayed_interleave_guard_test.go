package handlers

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"law-oa-go/internal/models"
)

// Reproduction of the delayed-interleaving gap (delayed-claim-review.md):
// request B passes the durable live-approval guard and then pauses
// between validate and claim (the gate). Request A arrives second, waits
// until B is parked (so validate of B strictly precedes the claim of A),
// then fully claims+creates+releases its approval AND returns its
// response. Only then is B released. B must be rejected 409; two 200s
// would mean a second live approval for the same intake (duplicate case
// creation).
func TestIntakeDelayedInterleaveDuplicateApproval(t *testing.T) {
	db := newDuplicateApprovalGuardDB(t)
	r := newDuplicateApprovalGuardRouter(db)
	body := `{"type":"case_creation","title":"延迟交错样本","content":"延迟交错样本","applicant_name":"qa_lawyer_a","department_name":"公司业务部","workflow_type":"CASE_APPROVAL","metadata":{"intake_id":"intake-dup","conflict_check_id":"CHK-DUP-1"}}`

	parked := make(chan struct{})      // closed when B parks at the gate
	aMayProceed := make(chan struct{}) // opened for A once B is parked
	resume := make(chan struct{})      // opened when A fully finished
	var resumeOnce sync.Once
	origGate := approvalClaimGate
	approvalClaimGate = func() {
		select {
		case <-parked:
			// This is A (B already parked): wait for permission, then run.
			<-aMayProceed
		default:
			// This is B: park immediately after validate, before claim.
			close(parked)
			<-resume
		}
	}
	t.Cleanup(func() {
		resumeOnce.Do(func() { close(resume) })
		approvalClaimGate = origGate
	})

	type result struct {
		code int
	}
	results := make([]result, 2)
	aDone := make(chan struct{}, 1)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			code, _ := postWithConflict(t, r, body)
			results[idx] = result{code}
			select {
			case aDone <- struct{}{}:
			default:
			}
		}(i)
	}

	// 1. B parks after its validate, before its claim.
	select {
	case <-parked:
	case <-time.After(20 * time.Second):
		resumeOnce.Do(func() { close(resume) })
		wg.Wait()
		t.Fatalf("B never reached the gate; codes: %d %d", results[0].code, results[1].code)
	}
	// 2. A may now run its full claim+create+release.
	close(aMayProceed)
	// 3. wait until A fully finished (response returned).
	<-aDone
	// 4. only now release B.
	resumeOnce.Do(func() { close(resume) })
	wg.Wait()

	var codes []int
	for _, res := range results {
		codes = append(codes, res.code)
	}
	t.Logf("status codes: %v", codes)

	var count int64
	db.Table("approval_requests").Count(&count)
	if count != 1 {
		t.Fatalf("delayed interleave must not create a second live approval, got %d", count)
	}
}

// TestIntakeConcurrentDistinctIntakesAllSucceed asserts that the
// duplicate-approval gate is per-intake: distinct intakes submitted at
// the same moment must all create their approval, never blocking each
// other.
func TestIntakeConcurrentDistinctIntakesAllSucceed(t *testing.T) {
	db := newDuplicateApprovalGuardDB(t)
	// Clone the baseline seeded intake/check (valid shape already proven
	// by the sequential guard test) into three more intakes.
	for i := 2; i <= 4; i++ {
		if err := db.Exec(
			`INSERT INTO case_intakes (id, title, case_type, client_id, created_by, status, metadata)
			  SELECT 'intake-dup-' || ?, 'parallel-' || ?, case_type, client_id, created_by, status,
			         replace(metadata, 'CHK-DUP-1', 'CHK-DUP-' || ?)
			  FROM case_intakes WHERE id = 'intake-dup'`,
			fmt.Sprintf("%d", i), fmt.Sprintf("%d", i), fmt.Sprintf("%d", i),
		).Error; err != nil {
			t.Fatalf("clone intake %d: %v", i, err)
		}
		if err := db.Exec(
			`INSERT INTO conflict_check_records (check_id, check_status, check_result)
			  SELECT 'CHK-DUP-' || ?, check_status, check_result
			  FROM conflict_check_records WHERE check_id = 'CHK-DUP-1'`,
			fmt.Sprintf("%d", i),
		).Error; err != nil {
			t.Fatalf("clone check %d: %v", i, err)
		}
	}
	r := newDuplicateApprovalGuardRouter(db)

	var wg sync.WaitGroup
	codes := make([]int, 4)
	errCodes := make([]string, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			n := idx + 1
			intakeID, checkID := "intake-dup", "CHK-DUP-1"
			if n > 1 {
				intakeID = fmt.Sprintf("intake-dup-%d", n)
				checkID = fmt.Sprintf("CHK-DUP-%d", n)
			}
			body := fmt.Sprintf(
				`{"type":"case_creation","title":"parallel","content":"parallel","applicant_name":"qa_lawyer_a","department_name":"公司业务部","workflow_type":"CASE_APPROVAL","metadata":{"intake_id":"%s","conflict_check_id":"%s"}}`,
				intakeID, checkID)
			code, resp := postWithConflict(t, r, body)
			codes[idx] = code
			if errObj, ok := resp["error"].(map[string]interface{}); ok {
				errCodes[idx], _ = errObj["code"].(string)
			}
		}(i)
	}
	wg.Wait()
	for i, code := range codes {
		if code != 200 {
			t.Fatalf("distinct intake %d must succeed concurrently, got %d (%s)", i+1, code, errCodes[i])
		}
	}
	var count int64
	db.Table("approval_requests").Count(&count)
	if count != 4 {
		t.Fatalf("four distinct intakes must yield four approvals, got %d", count)
	}
}

// TestIntakeResidualDraftRetryAfterFailedSubmit covers the failed-submit
// path: the claim is taken, creation fails, the intake is released back
// to conflict_ready, and no live approval exists, so a legitimate retry
// must succeed. A residual draft (not a live status) must not block it.
func TestIntakeResidualDraftRetryAfterFailedSubmit(t *testing.T) {
	db := newDuplicateApprovalGuardDB(t)
	r := newDuplicateApprovalGuardRouter(db)
	body := `{"type":"case_creation","title":"retry","content":"retry","applicant_name":"qa_lawyer_a","department_name":"公司业务部","workflow_type":"CASE_APPROVAL","metadata":{"intake_id":"intake-dup","conflict_check_id":"CHK-DUP-1"}}`

	// Force CreateApproval to fail after the claim by dropping its table.
	if err := db.Migrator().DropTable("approval_requests"); err != nil {
		t.Fatalf("drop approval_requests: %v", err)
	}
	code, _ := postWithConflict(t, r, body)
	if code == 200 {
		t.Fatal("submit must fail when approval creation fails")
	}

	// Recreate the table before retrying, as a deploy or migration would.
	if err := db.AutoMigrate(&models.ApprovalRequest{}); err != nil {
		t.Fatalf("recreate approval_requests: %v", err)
	}

	// No live approval exists (a draft is not live), so the released
	// intake must accept a legitimate retry and create exactly one live
	// approval.
	code2, resp2 := postWithConflict(t, r, body)
	if code2 != 200 {
		t.Fatalf("retry after failed submit must succeed, got %d: %v", code2, resp2)
	}
	var count int64
	db.Table("approval_requests").Where("status <> 'draft'").Count(&count)
	if count != 1 {
		t.Fatalf("retry must create exactly one non-draft approval, got %d", count)
	}
}
