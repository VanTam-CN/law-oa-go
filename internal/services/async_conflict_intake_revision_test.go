package services

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// intakeRevisionAsyncRepo is a self-contained in-memory double for the async
// service-level regression. linkRevision models the intake's current server
// revision: a detection finishing after a native edit must be rejected by the
// same revision comparison the transactional repository performs.
type intakeRevisionAsyncRepo struct {
	mu           sync.Mutex
	records      map[string]*models.ConflictCheckRecord
	linkRevision string
	accepted     []repositories.ConflictSubjectAssociation
}

func newIntakeRevisionAsyncRepo(linkRevision string) *intakeRevisionAsyncRepo {
	return &intakeRevisionAsyncRepo{
		records:      make(map[string]*models.ConflictCheckRecord),
		linkRevision: linkRevision,
	}
}

func (r *intakeRevisionAsyncRepo) rotateLinkRevision(revision string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.linkRevision = revision
}

func (r *intakeRevisionAsyncRepo) acceptedAssociations() []repositories.ConflictSubjectAssociation {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]repositories.ConflictSubjectAssociation(nil), r.accepted...)
}

func (r *intakeRevisionAsyncRepo) recordByID(checkID string) *models.ConflictCheckRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[checkID]
	if !ok {
		return nil
	}
	copyRecord := *record
	return &copyRecord
}

func (r *intakeRevisionAsyncRepo) SaveCheckRecord(_ context.Context, record *models.ConflictCheckRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.records[record.CheckID]; exists {
		return &duplicateIntakeRevisionCheckIDError{checkID: record.CheckID}
	}
	stored := *record
	r.records[record.CheckID] = &stored
	return nil
}

func (r *intakeRevisionAsyncRepo) GetCheckRecord(_ context.Context, checkID string) (*models.ConflictCheckRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[checkID]
	if !ok {
		// Production GetCheckRecord maps gorm.ErrRecordNotFound to (nil, nil);
		// the async service depends on that shape to detect a lost task.
		return nil, nil
	}
	copyRecord := *record
	return &copyRecord, nil
}

func (r *intakeRevisionAsyncRepo) UpdateCheckRecord(_ context.Context, record *models.ConflictCheckRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Production UpdateCheckRecord is a Save upsert.
	stored := *record
	r.records[record.CheckID] = &stored
	return nil
}

func (r *intakeRevisionAsyncRepo) LinkConflictCheckToCase(_ context.Context, association repositories.ConflictSubjectAssociation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if association.IntakeRevision == "" || association.IntakeRevision != r.linkRevision {
		return repositories.ErrIntakeRevisionConflict
	}
	r.accepted = append(r.accepted, association)
	return nil
}

func (r *intakeRevisionAsyncRepo) GetCheckHistory(_ context.Context, _ string, _ int) ([]*models.ConflictCheckRecord, error) {
	return nil, nil
}

func (r *intakeRevisionAsyncRepo) GetConflictCases(context.Context, *repositories.ConflictSearchParams) ([]*models.ConflictCase, error) {
	return nil, nil
}

func (r *intakeRevisionAsyncRepo) GetPotentialConflicts(context.Context, string, uint, []string, time.Time) ([]*models.ConflictCase, error) {
	return nil, nil
}

func (r *intakeRevisionAsyncRepo) GetClientRelations(context.Context, string) ([]*models.ClientRelation, error) {
	return nil, nil
}

func (r *intakeRevisionAsyncRepo) SaveConflictCases(context.Context, []*models.ConflictCase) error {
	return nil
}

func (r *intakeRevisionAsyncRepo) GetConflictRules(context.Context, bool) ([]*models.ConflictRule, error) {
	return nil, nil
}

func (r *intakeRevisionAsyncRepo) SaveConflictRule(context.Context, *models.ConflictRule) error {
	return nil
}

func (r *intakeRevisionAsyncRepo) UpdateConflictRule(context.Context, *models.ConflictRule) error {
	return nil
}

func (r *intakeRevisionAsyncRepo) GetMCPStandards(context.Context, bool) (*models.MCPStandards, error) {
	return nil, nil
}

func (r *intakeRevisionAsyncRepo) SaveMCPStandards(context.Context, *models.MCPStandards) error {
	return nil
}

func (r *intakeRevisionAsyncRepo) GetConflictStats(context.Context, string) (*repositories.ConflictStats, error) {
	return &repositories.ConflictStats{}, nil
}

type duplicateIntakeRevisionCheckIDError struct {
	checkID string
}

func (e *duplicateIntakeRevisionCheckIDError) Error() string {
	return "duplicate conflict check id: " + e.checkID
}

// intakeRevisionAsyncDetector records the request the async service actually
// forwards and supports a during-detection hook, which stands in for a native
// intake edit racing the in-flight check.
type intakeRevisionAsyncDetector struct {
	mu              sync.Mutex
	performed       []models.ConflictCheckRequest
	duringDetection func()
}

func (d *intakeRevisionAsyncDetector) PerformConflictCheck(_ context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	d.mu.Lock()
	d.performed = append(d.performed, *request)
	hook := d.duringDetection
	d.mu.Unlock()
	if hook != nil {
		hook()
	}
	return &models.ConflictCheckResponse{
		CheckID:     request.CheckID,
		HasConflict: false,
		CheckTime:   time.Now(),
		Decision: &models.ConflictDecisionSummary{
			Status:         "CLEAR",
			CoverageStatus: "COMPLETE",
		},
	}, nil
}

func (d *intakeRevisionAsyncDetector) GetCheckHistory(context.Context, string, int) ([]*models.ConflictCheckRecord, error) {
	return nil, nil
}

func (d *intakeRevisionAsyncDetector) GetConflictStats(context.Context, string) (*repositories.ConflictStats, error) {
	return &repositories.ConflictStats{}, nil
}

func (d *intakeRevisionAsyncDetector) performedRequests() []models.ConflictCheckRequest {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]models.ConflictCheckRequest(nil), d.performed...)
}

func intakeRevisionAsyncRequest(checkID, revision string) *models.ConflictCheckRequest {
	return &models.ConflictCheckRequest{
		CheckID:        checkID,
		IntakeID:       "intake-revision",
		IntakeRevision: revision,
		ClientID:       "33",
		ClientName:     "星河智联科技有限公司",
		ClientType:     "COMPANY",
		OtherParties:   []string{"对手公司"},
		CaseName:       "版本守护案件",
		CaseType:       "civil",
		SearchDepth:    "STANDARD",
		UserID:         "7",
		ActorUserID:    7,
		ActorRole:      "lawyer",
		RequestTime:    time.Now(),
	}
}

func waitForIntakeRevisionAsyncStatus(t *testing.T, repo *intakeRevisionAsyncRepo, taskID, want string) *models.ConflictCheckRecord {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		record := repo.recordByID(taskID)
		if record != nil && record.CheckStatus == want {
			return record
		}
		if time.Now().After(deadline) {
			status := "missing"
			if record != nil {
				status = record.CheckStatus
			}
			t.Fatalf("task %s status=%q, want %q", taskID, status, want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestAsyncConflictCheckServiceLinksFreshRevisionOnce(t *testing.T) {
	repo := newIntakeRevisionAsyncRepo("rev-A")
	detector := &intakeRevisionAsyncDetector{}
	service := NewAsyncConflictCheckService(repo, detector)

	task, err := service.CreateTask(context.Background(), intakeRevisionAsyncRequest("CCT-ASYNC-FRESH", "rev-A"))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task == nil || task.TaskID != "CCT-ASYNC-FRESH" || task.Status != ConflictTaskStatusQueued {
		t.Fatalf("unexpected create response: %+v", task)
	}
	record := waitForIntakeRevisionAsyncStatus(t, repo, "CCT-ASYNC-FRESH", ConflictTaskStatusCompleted)
	accepted := repo.acceptedAssociations()
	if len(accepted) != 1 {
		t.Fatalf("accepted associations=%d want 1: %+v", len(accepted), accepted)
	}
	association := accepted[0]
	if association.CheckID != "CCT-ASYNC-FRESH" || association.IntakeID != "intake-revision" || association.IntakeRevision != "rev-A" || association.ClientID != "33" {
		t.Fatalf("association fields mismatch: %+v", association)
	}
	if association.CoverageStatus != "COMPLETE" {
		t.Fatalf("coverage status=%q want COMPLETE", association.CoverageStatus)
	}
	performed := detector.performedRequests()
	if len(performed) != 1 || performed[0].IntakeRevision != "rev-A" {
		t.Fatalf("detector requests mismatch: %+v", performed)
	}
	result, err := service.GetTaskResult(context.Background(), "CCT-ASYNC-FRESH")
	if err != nil || result == nil || result.Error != "" || result.Task.Status != ConflictTaskStatusCompleted {
		t.Fatalf("task result: err=%v result=%+v", err, result)
	}
	if record.CheckStatus != ConflictTaskStatusCompleted {
		t.Fatalf("stored status=%q", record.CheckStatus)
	}
}

func TestAsyncConflictCheckServiceFailsStaleInFlightRevision(t *testing.T) {
	repo := newIntakeRevisionAsyncRepo("rev-A")
	detector := &intakeRevisionAsyncDetector{
		// A native facts edit completes while the check is in flight, so the
		// intake revision rotates before the result link is attempted.
		duringDetection: func() { repo.rotateLinkRevision("rev-B") },
	}
	service := NewAsyncConflictCheckService(repo, detector)

	task, err := service.CreateTask(context.Background(), intakeRevisionAsyncRequest("CCT-ASYNC-STALE", "rev-A"))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task == nil || task.TaskID != "CCT-ASYNC-STALE" {
		t.Fatalf("unexpected create response: %+v", task)
	}
	waitForIntakeRevisionAsyncStatus(t, repo, "CCT-ASYNC-STALE", ConflictTaskStatusFailed)
	if accepted := repo.acceptedAssociations(); len(accepted) != 0 {
		t.Fatalf("stale revision advanced the link: %+v", accepted)
	}
	performed := detector.performedRequests()
	if len(performed) != 1 || performed[0].IntakeRevision != "rev-A" {
		t.Fatalf("detector requests mismatch: %+v", performed)
	}
	result, err := service.GetTaskResult(context.Background(), "CCT-ASYNC-STALE")
	if err != nil || result == nil {
		t.Fatalf("task result: err=%v result=%+v", err, result)
	}
	if result.Task.Status != ConflictTaskStatusFailed {
		t.Fatalf("result status=%q want FAILED", result.Task.Status)
	}
	if !strings.Contains(result.Error, "接案事实版本已变化") {
		t.Fatalf("result error=%q want revision conflict message", result.Error)
	}
	if stored := repo.recordByID("CCT-ASYNC-STALE"); stored == nil || stored.CheckStatus != ConflictTaskStatusFailed {
		t.Fatalf("stored task not failed: %+v", stored)
	}
}
