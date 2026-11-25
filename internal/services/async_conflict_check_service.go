package services

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

const (
	ConflictTaskStatusQueued    = "QUEUED"
	ConflictTaskStatusRunning   = "RUNNING"
	ConflictTaskStatusCompleted = "COMPLETED"
	ConflictTaskStatusFailed    = "FAILED"
)

var ErrConflictTaskNotFound = stderrors.New("conflict check task not found")

type AsyncConflictCheckService interface {
	CreateTask(ctx context.Context, request *models.ConflictCheckRequest) (*ConflictCheckTaskResponse, error)
	GetTask(ctx context.Context, taskID string) (*ConflictCheckTaskResponse, error)
	GetTaskResult(ctx context.Context, taskID string) (*ConflictCheckTaskResultResponse, error)
}

type ConflictCheckTaskResponse struct {
	TaskID                     string    `json:"taskId"`
	CheckID                    string    `json:"checkId"`
	SubjectCaseID              string    `json:"subjectCaseId,omitempty"`
	IntakeID                   string    `json:"intakeId,omitempty"`
	Status                     string    `json:"status"`
	ClientID                   string    `json:"clientId"`
	ClientName                 string    `json:"clientName"`
	CaseName                   string    `json:"caseName"`
	CaseType                   string    `json:"caseType"`
	HasConflict                bool      `json:"hasConflict"`
	RiskLevel                  string    `json:"riskLevel"`
	RecommendedPollingInterval int       `json:"recommendedPollingInterval"`
	CreatedAt                  time.Time `json:"createdAt"`
	UpdatedAt                  time.Time `json:"updatedAt"`
	OwnerID                    uint      `json:"-"`
}

type ConflictCheckTaskResultResponse struct {
	Task   *ConflictCheckTaskResponse `json:"task"`
	Result models.JSON                `json:"result,omitempty"`
	Error  string                     `json:"error,omitempty"`
}

type asyncConflictCheckService struct {
	conflictRepo    repositories.BasicConflictRepository
	conflictService ConflictDetectionService
	taskTimeout     time.Duration
}

func NewAsyncConflictCheckService(conflictRepo repositories.BasicConflictRepository, conflictService ConflictDetectionService) AsyncConflictCheckService {
	return &asyncConflictCheckService{
		conflictRepo:    conflictRepo,
		conflictService: conflictService,
		taskTimeout:     90 * time.Second,
	}
}

func (s *asyncConflictCheckService) CreateTask(ctx context.Context, request *models.ConflictCheckRequest) (*ConflictCheckTaskResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	if request.CheckID == "" {
		request.CheckID = "CCT_" + uuid.NewString()
	}
	if (strings.TrimSpace(request.SubjectCaseID) != "" || strings.TrimSpace(request.IntakeID) != "") && !hasConflictSubjectLinker(s.conflictRepo) {
		return nil, fmt.Errorf("冲突检测上下文关联服务未初始化，已阻止创建无案件关联任务")
	}

	userID, _ := strconv.ParseUint(request.UserID, 10, 32)
	now := time.Now()
	record := &models.ConflictCheckRecord{
		CheckID:          request.CheckID,
		ClientID:         request.ClientID,
		ClientName:       request.ClientName,
		CaseName:         request.CaseName,
		CaseType:         request.CaseType,
		CheckStatus:      ConflictTaskStatusQueued,
		RiskLevel:        "LOW",
		SearchParameters: toModelJSON(auditSafeConflictRequest(request)),
		UserID:           uint(userID),
		CheckTime:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.conflictRepo.SaveCheckRecord(ctx, record); err != nil {
		// Reusing a caller-supplied check ID is the supported idempotency path.
		// Return the existing task only to its owner; the HTTP handler performs a
		// second case/intake authorization check before serializing it.
		existing, getErr := s.conflictRepo.GetCheckRecord(ctx, record.CheckID)
		if getErr != nil || existing == nil || (existing.UserID != uint(userID) && !IsConflictReviewRole(request.ActorRole)) {
			return nil, err
		}
		return s.toTaskResponse(existing), nil
	}

	taskRequest := *request
	go s.runTask(record.CheckID, &taskRequest)

	return s.toTaskResponse(record), nil
}

func (s *asyncConflictCheckService) GetTask(ctx context.Context, taskID string) (*ConflictCheckTaskResponse, error) {
	record, err := s.conflictRepo.GetCheckRecord(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrConflictTaskNotFound
	}
	return s.toTaskResponse(record), nil
}

func (s *asyncConflictCheckService) GetTaskResult(ctx context.Context, taskID string) (*ConflictCheckTaskResultResponse, error) {
	record, err := s.conflictRepo.GetCheckRecord(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrConflictTaskNotFound
	}

	response := &ConflictCheckTaskResultResponse{
		Task:   s.toTaskResponse(record),
		Result: record.CheckResult,
	}
	if record.CheckStatus == ConflictTaskStatusFailed {
		if value, ok := record.CheckResult["error"].(string); ok {
			response.Error = value
		}
	}
	return response, nil
}

func (s *asyncConflictCheckService) runTask(taskID string, request *models.ConflictCheckRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), s.taskTimeout)
	defer cancel()

	startTime := time.Now()
	record, err := s.conflictRepo.GetCheckRecord(ctx, taskID)
	if err != nil || record == nil {
		return
	}

	if err := s.conflictRepo.ClaimConflictCheckRun(ctx, taskID); err != nil {
		return
	}

	initialRisk := record.RiskLevel
	fail := func(cause error) {
		// After a detector timeout its context cannot persist the failure receipt.
		receiptCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer done()
		record.CheckStatus = ConflictTaskStatusFailed
		record.HasConflict = false
		record.RiskLevel = initialRisk
		record.CheckResult = models.JSON{"error": cause.Error()}
		record.Duration = time.Since(startTime).Milliseconds()
		if err := s.conflictRepo.FinalizeConflictCheck(receiptCtx, record); err != nil {
			log.Print("无法写入冲突检测失败终态，请核对任务记录")
		}
	}
	if s.conflictService == nil {
		fail(stderrors.New("conflict detection service is not initialized"))
		return
	}

	var result *models.ConflictCheckResponse
	if formal, ok := s.conflictService.(*conflictDetectionService); ok {
		result, err = formal.performConflictCheck(ctx, request, true)
	} else {
		result, err = s.conflictService.PerformConflictCheck(ctx, request)
	}
	if err != nil {
		fail(err)
		return
	}
	if ctx.Err() != nil {
		fail(ctx.Err())
		return
	}
	if result == nil {
		fail(stderrors.New("conflict detector returned no result"))
		return
	}
	record.Duration = time.Since(startTime).Milliseconds()
	record.CheckStatus = ConflictTaskStatusCompleted
	record.HasConflict = result.HasConflict
	auditResult := *result
	auditResult.NormalizedSubjects = auditSafeNormalizedSubjects(result.NormalizedSubjects)
	record.CheckResult = toModelJSON(&auditResult)
	if result.RiskAssessment != nil {
		record.RiskLevel = result.RiskAssessment.OverallRisk
	}
	record.CheckTime = result.CheckTime
	if strings.TrimSpace(request.SubjectCaseID) != "" || strings.TrimSpace(request.IntakeID) != "" {
		coverageStatus := ""
		if result.Decision != nil {
			coverageStatus = result.Decision.CoverageStatus
		}
		association := repositories.ConflictSubjectAssociation{
			CheckID:           record.CheckID,
			SubjectCaseID:     request.SubjectCaseID,
			SubjectCaseNumber: request.SubjectCaseNumber,
			IntakeID:          request.IntakeID,
			IntakeRevision:    request.IntakeRevision,
			ClientID:          request.ClientID,
			CoverageStatus:    coverageStatus,
			CheckedAt:         record.CheckTime,
		}
		err = s.conflictRepo.FinalizeConflictCheckWithAssociation(ctx, record, association)
	} else {
		err = s.conflictRepo.FinalizeConflictCheck(ctx, record)
	}
	if err != nil {
		fail(err)
	}
}

func hasConflictSubjectLinker(repo repositories.BasicConflictRepository) bool {
	_, ok := repo.(repositories.ConflictSubjectLinker)
	return ok
}

func (s *asyncConflictCheckService) toTaskResponse(record *models.ConflictCheckRecord) *ConflictCheckTaskResponse {
	subjectCaseID := subjectCaseIDFromSearchParameters(record.SearchParameters)
	subjectCaseIDText := ""
	if subjectCaseID > 0 {
		subjectCaseIDText = strconv.FormatUint(uint64(subjectCaseID), 10)
	}
	return &ConflictCheckTaskResponse{
		TaskID:                     record.CheckID,
		CheckID:                    record.CheckID,
		SubjectCaseID:              subjectCaseIDText,
		IntakeID:                   SubjectIntakeIDFromSearchParameters(record.SearchParameters),
		Status:                     record.CheckStatus,
		ClientID:                   record.ClientID,
		ClientName:                 record.ClientName,
		CaseName:                   record.CaseName,
		CaseType:                   record.CaseType,
		HasConflict:                record.HasConflict,
		RiskLevel:                  record.RiskLevel,
		RecommendedPollingInterval: 2,
		CreatedAt:                  record.CreatedAt,
		UpdatedAt:                  record.UpdatedAt,
		OwnerID:                    record.UserID,
	}
}

func toModelJSON(value interface{}) models.JSON {
	if value == nil {
		return models.JSON{}
	}
	if data, ok := value.(models.JSON); ok {
		return data
	}
	if data, ok := value.(map[string]interface{}); ok {
		return models.JSON(data)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return models.JSON{"marshal_error": err.Error()}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err == nil {
		return models.JSON(result)
	}

	var generic interface{}
	if err := json.Unmarshal(data, &generic); err != nil {
		return models.JSON{"unmarshal_error": err.Error(), "raw": fmt.Sprintf("%s", data)}
	}
	return models.JSON{"items": generic}
}
