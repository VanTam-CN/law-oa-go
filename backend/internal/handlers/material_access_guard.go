package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// materialAccessGuard bundles the object-level permission checks shared by
// the material and case-task handlers: one missing check here would turn
// a guessed ID into a read or write path, so every route entry point goes
// through these helpers first.
type materialAccessGuard struct {
	authz *services.AuthorizationService
	db    *gorm.DB
}

func newMaterialAccessGuard(authz *services.AuthorizationService, db *gorm.DB) *materialAccessGuard {
	return &materialAccessGuard{authz: authz, db: db}
}

// requireCaseReadable passes only when the actor may read the case.
func (s *materialAccessGuard) requireCaseReadable(c *gin.Context, caseID uint) bool {
	if s.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "MATERIAL_AUTHZ_UNAVAILABLE", "材料权限服务未初始化，已拒绝访问")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, err := s.authz.CanReadCase(c.Request.Context(), actor, caseID)
	if err != nil {
		common.APIInternalServerError(c, "材料权限校验失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

// requireCaseManageable passes only when the actor may write the case.
func (s *materialAccessGuard) requireCaseManageable(c *gin.Context, caseID uint) bool {
	if s.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "MATERIAL_AUTHZ_UNAVAILABLE", "材料权限服务未初始化，已拒绝操作")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, err := s.authz.CanManageCase(c.Request.Context(), actor, caseID)
	if err != nil {
		common.APIInternalServerError(c, "材料权限校验失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

// intakeForApproval resolves the intake UUID bound to an approval.
func (s *materialAccessGuard) intakeForApproval(approvalID string) (string, error) {
	var row struct{ Metadata string }
	if err := s.db.Table("approval_requests").Select("metadata").Where("id = ?", approvalID).Take(&row).Error; err != nil {
		return "", fmt.Errorf("读取审批失败：%w", err)
	}
	metadata := map[string]interface{}{}
	if strings.TrimSpace(row.Metadata) != "" {
		if err := json.Unmarshal([]byte(row.Metadata), &metadata); err != nil {
			return "", fmt.Errorf("解析审批元数据失败：%w", err)
		}
	}
	intakeID, _ := metadata["intake_id"].(string)
	return strings.TrimSpace(intakeID), nil
}

// authorizeIntakeMaterial enforces the intake-level boundary for rows
// not yet bound to a case: writes stay creator-only, and the creator\'s
// reads must pass the same client ethical-wall check the client API
// uses. A guessed intake UUID therefore never becomes a back door into
// a walled client context.
//
// write=false (read) additionally admits the current approver or the
// ACTIVE in-window conflict reviewer of an approval bound to this
// intake: a reviewer who returned material items must be able to see
// what he asked for through the direct object binding, with no global
// role bypass and no veto from the applying lawyer\'s client wall
// (review 2026-09-30). write=true stays creator-only.
func (s *materialAccessGuard) authorizeIntakeMaterial(c *gin.Context, intakeID string, write bool) bool {
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	var intake struct {
		ClientID  *uint
		CreatedBy string
	}
	if err := s.db.Table("case_intakes").Select("client_id, created_by").Where("id = ?", intakeID).Take(&intake).Error; err != nil {
		common.APIInternalServerError(c, "读取接案失败", err.Error())
		return false
	}
	isCreator := strings.TrimSpace(intake.CreatedBy) == fmt.Sprint(actor.UserID)
	workflowBound := false
	if !isCreator {
		if write {
			forbidObjectAccess(c)
			return false
		}
		bound, err := s.actorBoundToIntakeWorkflow(intakeID, actor.UserID)
		if err != nil {
			// One decision, one response: the query failure surfaces as a
			// single 500 and the caller never appends a second JSON body.
			common.APIInternalServerError(c, "校验审批绑定失败", err.Error())
			return false
		}
		if !bound {
			forbidObjectAccess(c)
			return false
		}
		workflowBound = true
	}
	// A workflow-bound approver/reviewer was admitted by the direct object
	// binding (current approval or ACTIVE in-window reviewer assignment
	// joined to this exact intake). The applying lawyer's client-level
	// visibility must not veto that professional read appointment, and
	// conflict officers are deliberately outside
	// IsBusinessMatterManagementRole, so a client-wall recount would deny
	// every assigned reviewer. The creator path below still passes the
	// same ethical-wall check as the client API. Read-only: write stays
	// creator-only.
	if !workflowBound && intake.ClientID != nil && *intake.ClientID > 0 {
		return s.canReadClient(c, *intake.ClientID)
	}
	return true
}

// actorBoundToIntakeWorkflow reports whether an approval bound to this
// intake currently names the actor as approver, or the actor is the
// currently assigned independent reviewer of the anchored conflict check
// (conflict_reviewer_assignments, ACTIVE and in window; the check creator
// in conflict_check_records.user_id is the applying lawyer and never
// qualifies). The intake_id comparison is a Go-side exact match against
// the parsed JSON field: no SQL LIKE on a json/jsonb column (Postgres
// rejects json ~~ unknown, SQLSTATE 42883), and no substring fuzz that
// could match a UUID embedded in an unrelated field (trial attempt2
// 2026-09-30). Errors return as (false, err); the caller owns the single
// response write.
func (s *materialAccessGuard) actorBoundToIntakeWorkflow(intakeID string, actorID uint) (bool, error) {
	approverRows := []struct {
		Metadata []byte
	}{}
	if err := s.db.Table("approval_requests").
		Select("metadata").
		Where("current_approver_id = ?", fmt.Sprint(actorID)).
		Find(&approverRows).Error; err != nil {
		return false, fmt.Errorf("读取审批绑定失败：%w", err)
	}
	for _, row := range approverRows {
		if approvalMetadataBindsIntake(row.Metadata, intakeID) {
			return true, nil
		}
	}
	// Reviewer path: the actor must be the CURRENTLY ASSIGNED independent
	// reviewer of the check the approval is anchored to. The check creator
	// (conflict_check_records.user_id) is the applying lawyer and must
	// never qualify (review 2026-09-30). The join mirrors
	// services.GetActiveConflictReviewerAssignment: status ACTIVE and an
	// effective window covering now, and approval metadata still has to
	// bind this exact intake in Go.
	reviewRows := []struct {
		Metadata []byte
	}{}
	now := time.Now()
	if err := s.db.Table("conflict_reviewer_assignments").
		Select("approval_requests.metadata").
		Joins("JOIN approval_requests ON approval_requests.conflict_check_id = conflict_reviewer_assignments.check_id").
		Where("conflict_reviewer_assignments.reviewer_id = ?", actorID).
		Where("conflict_reviewer_assignments.status = ?", models.ConflictReviewerAssignmentActive).
		Where("conflict_reviewer_assignments.effective_from IS NULL OR conflict_reviewer_assignments.effective_from <= ?", now).
		Where("conflict_reviewer_assignments.effective_to IS NULL OR conflict_reviewer_assignments.effective_to > ?", now).
		Scan(&reviewRows).Error; err != nil {
		// A slim schema without the assignment table stays fail-closed
		// (deny) rather than turning the deny into an HTTP error: the
		// approver path above already covered the primary binding.
		return false, nil
	}
	for _, row := range reviewRows {
		if approvalMetadataBindsIntake(row.Metadata, intakeID) {
			return true, nil
		}
	}
	// Pre-approval reviewer read: an assigned reviewer must be able to
	// inspect the intake materials he will be tested against BEFORE any
	// approval exists. The formal commit (CommitIntakeConflictResult)
	// writes metadata["conflict_check_id"] on the intake itself, so the
	// binding is intake.metadata.conflict_check_id = assignment.check_id
	// with the same ACTIVE + in-window conditions as
	// services.GetActiveConflictReviewerAssignment (latest active
	// appointment wins, never the check creator). Read-only: writes
	// stay creator-only in authorizeIntakeMaterial.
	var intakeMeta struct {
		Metadata []byte
	}
	if err := s.db.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&intakeMeta).Error; err != nil {
		// No intake metadata: pre-approval binding unavailable; fail closed.
		return false, nil
	}
	var parsed struct {
		ConflictCheckID string `json:"conflict_check_id"`
	}
	if err := json.Unmarshal(intakeMeta.Metadata, &parsed); err != nil || strings.TrimSpace(parsed.ConflictCheckID) == "" {
		return false, nil
	}
	var assigned int64
	if err := s.db.Table("conflict_reviewer_assignments").
		Where("check_id = ? AND reviewer_id = ?", strings.TrimSpace(parsed.ConflictCheckID), actorID).
		Where("status = ?", models.ConflictReviewerAssignmentActive).
		Where("effective_from IS NULL OR effective_from <= ?", now).
		Where("effective_to IS NULL OR effective_to > ?", now).
		Count(&assigned).Error; err != nil {
		// A slim schema without the assignment table stays fail-closed.
		return false, nil
	}
	return assigned > 0, nil
}

// approvalMetadataBindsIntake parses the approval metadata JSON and
// compares intake_id EXACTLY (string equality). A json/jsonb column is
// read back as raw bytes and decoded here, so the query never asks the
// database to LIKE a non-text column.
func approvalMetadataBindsIntake(metadata []byte, intakeID string) bool {
	if len(metadata) == 0 {
		return false
	}
	var parsed struct {
		IntakeID string `json:"intake_id"`
	}
	if err := json.Unmarshal(metadata, &parsed); err != nil {
		return false
	}
	return strings.TrimSpace(parsed.IntakeID) == intakeID
}

func (s *materialAccessGuard) canReadClient(c *gin.Context, clientID uint) bool {
	if s.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "MATERIAL_AUTHZ_UNAVAILABLE", "客户权限服务未初始化，已拒绝访问")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, err := s.authz.CanReadClient(c.Request.Context(), actor, clientID)
	if err != nil {
		common.APIInternalServerError(c, "客户权限校验失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

// materialScope is the resolved ownership context of one case_materials
// row: it either belongs to a numeric case or to a UUID intake, never to
// a guessed hybrid. The zero scope means the row is orphaned and every
// path fails closed.
type materialScope struct {
	CaseID   uint
	IntakeID string
}

// resolveCaseScope reads the real ownership columns of the material row.
// It never falls back to "smallest case of the client" or any other
// heuristic: a wrong case id here would turn an unrelated lawyer into an
// owner. Intake rows resolve their scope from the intake itself.
func (s *materialAccessGuard) resolveCaseScope(c *gin.Context, materialID uint) (materialScope, bool) {
	var row struct {
		CaseID   *uint
		IntakeID *string
	}
	if err := s.db.Table("case_materials").Select("case_id, intake_id").Where("id = ?", materialID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.APINotFound(c, "材料不存在")
		} else {
			common.APIInternalServerError(c, "读取材料归属失败", err.Error())
		}
		return materialScope{}, false
	}
	scope := materialScope{}
	if row.CaseID != nil {
		scope.CaseID = *row.CaseID
	}
	if row.IntakeID != nil {
		scope.IntakeID = strings.TrimSpace(*row.IntakeID)
	}
	if scope.CaseID == 0 && scope.IntakeID == "" {
		forbidObjectAccess(c)
		return materialScope{}, false
	}
	return scope, true
}

// materialStoragePath returns the safe absolute path for one uploaded
// attachment. Storage lives under <workdir>/uploads/materials/<aa>/; the
// final name is the server-side digest plus a sanitized extension, so a
// malicious multipart filename can never traverse or overwrite a file.
// materialStoragePath returns the safe absolute path for one uploaded
// attachment plus whether THIS call created a fresh file name. Storage
// lives under <workdir>/uploads/materials/<aa>/; the name carries the
// digest prefix for grep-ability plus a random suffix so every upload is
// its own inode: a failed commit may delete its own file without ever
// removing a file another material row already references.
func materialStoragePath(workdir, digest, originalName string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if len(ext) > 12 {
		ext = ""
	}
	for _, r := range ext {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.') {
			ext = ""
			break
		}
	}
	dir := filepath.Join(workdir, "uploads", "materials", digest[:2])
	_ = os.MkdirAll(dir, 0o750)
	unique := digest + "-" + randomUploadToken() + ext
	return filepath.Join(dir, unique), true
}

// randomUploadToken mints the per-upload random suffix without pulling a
// new dependency: crypto/rand hex, same generator family the services
// layer already uses for idempotency keys.
func randomUploadToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// Crypto/rand failure is effectively impossible; fall back to
		// time-based uniqueness rather than silently reusing a name.
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(buf)
}
