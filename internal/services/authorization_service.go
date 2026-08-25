package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// AuthActor is the authenticated subject used by handler-level object checks.
type AuthActor struct {
	UserID uint
	Role   string
}

// AuthorizationService centralizes object ownership checks for HTTP handlers.
type AuthorizationService struct {
	caseRepo   repositories.CaseRepository
	clientRepo repositories.ClientRepository
	userRepo   repositories.UserRepository
	docRepo    repositories.DocumentRepository
}

func NewAuthorizationService(
	caseRepo repositories.CaseRepository,
	clientRepo repositories.ClientRepository,
	userRepo repositories.UserRepository,
	docRepo repositories.DocumentRepository,
) *AuthorizationService {
	return &AuthorizationService{
		caseRepo:   caseRepo,
		clientRepo: clientRepo,
		userRepo:   userRepo,
		docRepo:    docRepo,
	}
}

func IsPrivilegedRole(role string) bool {
	switch normalizeRole(role) {
	case "admin", "super_admin":
		return true
	default:
		return false
	}
}

// IsIntakeAssistantRole covers both the legacy assistant role and the
// dedicated intake-assistant role. Neither role may confirm identities,
// execute conflict checks, or make a professional conflict decision.
func IsIntakeAssistantRole(role string) bool {
	switch normalizeRole(role) {
	case "assistant", "intake_assistant":
		return true
	default:
		return false
	}
}

// IsTechnicalAdminRole distinguishes account/configuration administrators
// from business roles that have an explicit professional matter appointment.
// A technical administrator must not receive matter access merely because the
// account has a broad platform role.
func IsTechnicalAdminRole(role string) bool {
	return IsPrivilegedRole(role) && !IsBusinessMatterManagementRole(role)
}

func IsMatterManagementRole(role string) bool {
	switch normalizeRole(role) {
	case "admin", "super_admin", "director", "partner", "compliance", "risk", "risk_control", "management":
		return true
	default:
		return false
	}
}

// IsBusinessMatterManagementRole excludes technical administrators. Technical
// administration is not a professional conflict-review appointment and must
// never grant access to protected historical conflict evidence by implication.
func IsBusinessMatterManagementRole(role string) bool {
	switch normalizeRole(role) {
	case "director", "partner", "compliance", "risk", "risk_control", "management":
		return true
	default:
		return false
	}
}

// IsConflictReviewRole identifies the independent roles allowed to inspect
// historical conflict evidence and record a professional conclusion. A
// conflict officer is deliberately not a general matter-management role.
func IsConflictReviewRole(role string) bool {
	switch normalizeRole(role) {
	case "director", "partner", "compliance", "risk", "risk_control", "management", "conflict_officer":
		return true
	default:
		return false
	}
}

func (s *AuthorizationService) CanReadCase(ctx context.Context, actor AuthActor, caseID uint) (bool, error) {
	if IsTechnicalAdminRole(actor.Role) {
		return false, nil
	}
	caseModel, err := s.findCase(ctx, caseID)
	if err != nil || caseModel == nil {
		return false, err
	}
	visible, err := s.canSeeCaseThroughEthicalWall(ctx, actor, caseModel)
	if err != nil || !visible {
		return false, err
	}
	if IsBusinessMatterManagementRole(actor.Role) {
		return true, nil
	}
	return ownsCase(actor, caseModel), nil
}

// CanReadConflictContext authorizes access to conflict evidence attached to a
// case without granting the caller general matter access. Independent
// conflict reviewers may inspect an unprotected case's conflict evidence, but
// an ethical wall still requires an explicit whitelist entry.
func (s *AuthorizationService) CanReadConflictContext(ctx context.Context, actor AuthActor, caseID uint) (bool, error) {
	if IsTechnicalAdminRole(actor.Role) {
		return false, nil
	}
	caseModel, err := s.findCase(ctx, caseID)
	if err != nil || caseModel == nil {
		return false, err
	}
	visible, err := s.canSeeCaseThroughEthicalWall(ctx, actor, caseModel)
	if err != nil || !visible {
		return false, err
	}
	if IsConflictReviewRole(actor.Role) || IsBusinessMatterManagementRole(actor.Role) {
		return true, nil
	}
	return ownsCase(actor, caseModel), nil
}

func (s *AuthorizationService) CanManageCase(ctx context.Context, actor AuthActor, caseID uint) (bool, error) {
	if IsTechnicalAdminRole(actor.Role) {
		return false, nil
	}
	caseModel, err := s.findCase(ctx, caseID)
	if err != nil || caseModel == nil {
		return false, err
	}
	visible, err := s.canSeeCaseThroughEthicalWall(ctx, actor, caseModel)
	if err != nil || !visible {
		return false, err
	}
	if IsBusinessMatterManagementRole(actor.Role) {
		return true, nil
	}
	return ownsCase(actor, caseModel), nil
}

func (s *AuthorizationService) CanManageEthicalWall(ctx context.Context, actor AuthActor, caseID uint) (bool, error) {
	if IsBusinessMatterManagementRole(actor.Role) {
		return true, nil
	}
	return s.CanManageCase(ctx, actor, caseID)
}

func (s *AuthorizationService) CanCreateCase(ctx context.Context, actor AuthActor, req *CreateCaseRequest) (bool, error) {
	if IsTechnicalAdminRole(actor.Role) {
		return false, nil
	}
	if IsBusinessMatterManagementRole(actor.Role) {
		if req == nil || req.ClientID == 0 {
			return false, nil
		}
		return s.CanReadClient(ctx, actor, req.ClientID)
	}
	if req == nil || actor.UserID == 0 || req.LawyerID != actor.UserID {
		return false, nil
	}
	if req.AssignedBy != 0 && req.AssignedBy != actor.UserID {
		return false, nil
	}
	client, err := s.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil || client == nil {
		return false, err
	}
	count, err := s.countCasesForClient(ctx, req.ClientID, 0)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return true, nil
	}
	return s.CanReadClient(ctx, actor, req.ClientID)
}

func (s *AuthorizationService) CanReadClient(ctx context.Context, actor AuthActor, clientID uint) (bool, error) {
	if IsTechnicalAdminRole(actor.Role) {
		return false, nil
	}
	if IsBusinessMatterManagementRole(actor.Role) {
		return s.canReadClientWithinEthicalWall(ctx, actor, clientID)
	}
	client, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil || client == nil {
		return false, err
	}
	if client.CreatedBy != 0 && client.CreatedBy == actor.UserID {
		// Creation grants access before the first matter exists, but it must
		// never bypass an ethical wall added after matter creation.
		return s.canReadClientWithinEthicalWall(ctx, actor, clientID)
	}
	count, err := s.countCasesForClient(ctx, clientID, actor.UserID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CanCreateClient limits authoritative client-master creation to lawyers and
// firm management roles. Assistants may prepare intake drafts but cannot write
// the identity-bearing client master.
func (s *AuthorizationService) CanCreateClient(_ context.Context, actor AuthActor) (bool, error) {
	if IsBusinessMatterManagementRole(actor.Role) {
		return true, nil
	}
	return normalizeRole(actor.Role) == "lawyer", nil
}

func (s *AuthorizationService) CanManageClient(ctx context.Context, actor AuthActor, clientID uint) (bool, error) {
	if IsTechnicalAdminRole(actor.Role) {
		return false, nil
	}
	if IsBusinessMatterManagementRole(actor.Role) {
		return s.canReadClientWithinEthicalWall(ctx, actor, clientID)
	}
	if normalizeRole(actor.Role) != "lawyer" {
		return false, nil
	}
	return s.CanReadClient(ctx, actor, clientID)
}

func (s *AuthorizationService) CanReadDocument(ctx context.Context, actor AuthActor, documentID uint) (bool, error) {
	return s.canAccessDocument(ctx, actor, documentID, false)
}

func (s *AuthorizationService) CanManageDocument(ctx context.Context, actor AuthActor, documentID uint) (bool, error) {
	return s.canAccessDocument(ctx, actor, documentID, true)
}

func (s *AuthorizationService) CanCreateDocument(ctx context.Context, actor AuthActor, entityType string, entityID uint) (bool, error) {
	if strings.EqualFold(strings.TrimSpace(entityType), "case") && entityID > 0 {
		return s.CanManageCase(ctx, actor, entityID)
	}
	return IsBusinessMatterManagementRole(actor.Role), nil
}

func (s *AuthorizationService) canAccessDocument(ctx context.Context, actor AuthActor, documentID uint, write bool) (bool, error) {
	if IsTechnicalAdminRole(actor.Role) {
		return false, nil
	}
	doc, err := s.docRepo.FindByID(ctx, documentID)
	if err != nil {
		if err == repositories.ErrDocumentNotFound {
			return false, nil
		}
		return false, err
	}
	if doc == nil {
		return false, nil
	}
	if strings.EqualFold(doc.EntityType, "case") && doc.EntityID > 0 {
		if write {
			return s.CanManageCase(ctx, actor, doc.EntityID)
		}
		return s.CanReadCase(ctx, actor, doc.EntityID)
	}
	return false, nil
}

// canSeeCaseThroughEthicalWall is deliberately shared by all object checks.
// Route middleware protects direct case URLs, but conflict, client, approval,
// and document handlers may call the authorization service without that
// middleware. A business role therefore never bypasses a wall merely because
// it is a director or partner; the explicit whitelist remains authoritative.
func (s *AuthorizationService) canSeeCaseThroughEthicalWall(ctx context.Context, actor AuthActor, caseModel *models.Case) (bool, error) {
	if caseModel == nil || !caseModel.EthicalWallEnabled {
		return true, nil
	}
	if actor.UserID == 0 || s.caseRepo == nil || s.caseRepo.GetDB() == nil {
		return false, nil
	}
	var count int64
	err := s.caseRepo.GetDB().WithContext(ctx).
		Table("case_ethical_wall_whitelist").
		Where("case_id = ? AND user_id = ?", caseModel.ID, actor.UserID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("查询案件隔离墙白名单失败: %w", err)
	}
	return count > 0, nil
}

// canReadClientWithinEthicalWall keeps direct client reads consistent with
// the SQL-side client list filter. A client with no matter yet is readable by
// business management; a client whose only matters are wall-protected is not.
func (s *AuthorizationService) canReadClientWithinEthicalWall(ctx context.Context, actor AuthActor, clientID uint) (bool, error) {
	if s.caseRepo == nil || s.caseRepo.GetDB() == nil {
		return false, fmt.Errorf("案件权限数据库未初始化")
	}
	db := s.caseRepo.GetDB().WithContext(ctx)
	var total int64
	if err := db.Model(&models.Case{}).Where("client_id = ? AND deleted_at IS NULL", clientID).Count(&total).Error; err != nil {
		return false, err
	}
	if total == 0 {
		return true, nil
	}
	var visible int64
	if err := db.Model(&models.Case{}).Where(`client_id = ? AND deleted_at IS NULL AND (
		ethical_wall_enabled = ?
		OR EXISTS (
			SELECT 1 FROM case_ethical_wall_whitelist wall_access
			WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
		)
	)`, clientID, false, actor.UserID).Count(&visible).Error; err != nil {
		return false, err
	}
	return visible > 0, nil
}

func (s *AuthorizationService) findCase(ctx context.Context, caseID uint) (*models.Case, error) {
	if s == nil || s.caseRepo == nil {
		return nil, fmt.Errorf("案件权限服务未初始化")
	}
	caseModel, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	return caseModel, nil
}

func (s *AuthorizationService) countCasesForClient(ctx context.Context, clientID uint, lawyerID uint) (int64, error) {
	db := s.caseRepo.GetDB()
	if db == nil {
		return 0, fmt.Errorf("case repository database is not available")
	}
	query := db.WithContext(ctx).Model(&models.Case{}).Where("client_id = ? AND deleted_at IS NULL", clientID)
	if lawyerID > 0 {
		query = query.Where("(lawyer_id = ? OR created_by = ?)", lawyerID, strconv.FormatUint(uint64(lawyerID), 10))
		query = query.Where(`(
			ethical_wall_enabled = ?
			OR EXISTS (
				SELECT 1 FROM case_ethical_wall_whitelist wall_access
				WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
			)
		)`, false, lawyerID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func ownsCase(actor AuthActor, caseModel *models.Case) bool {
	if caseModel == nil || actor.UserID == 0 {
		return false
	}
	if caseModel.LawyerID == actor.UserID {
		return true
	}
	return caseModel.CreatedBy == strconv.FormatUint(uint64(actor.UserID), 10)
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}
