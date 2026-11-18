package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

const (
	PolicyEndorsementManagement = "MANAGEMENT"
	PolicyEndorsementCompliance = "COMPLIANCE"
)

var sha256Pattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

type ConflictPolicyPackageInput struct {
	PolicyVersion               string     `json:"policy_version"`
	Jurisdiction                string     `json:"jurisdiction"`
	ApplicableRuleName          string     `json:"applicable_rule_name"`
	ApplicableRuleVersion       string     `json:"applicable_rule_version"`
	ApplicableRuleAuthority     string     `json:"applicable_rule_authority"`
	ApplicableRuleReference     string     `json:"applicable_rule_reference"`
	DataSourcePolicyReference   string     `json:"data_source_policy_reference"`
	PrivacyBasisMatrixReference string     `json:"privacy_basis_matrix_reference"`
	RetentionPolicyReference    string     `json:"retention_policy_reference"`
	WaiverPolicyReference       string     `json:"waiver_policy_reference"`
	ControlledActionsReference  string     `json:"controlled_actions_reference"`
	ExternalReviewReference     string     `json:"external_review_reference"`
	EffectiveAt                 time.Time  `json:"effective_at"`
	NextReviewAt                time.Time  `json:"next_review_at"`
	ExpiresAt                   *time.Time `json:"expires_at"`
	IntegrityHash               string     `json:"integrity_hash"`
}

type ConflictPolicyPackageView struct {
	Package      models.LawFirmCompliancePolicyPackage `json:"package"`
	Endorsements []ConflictPolicyEndorsementView       `json:"endorsements"`
	Status       string                                `json:"status"`
}

type ConflictPolicyEndorsementView struct {
	models.LawFirmCompliancePolicyEndorsement
	EndorserName string `json:"endorser_name"`
}

type ConflictPolicyService struct {
	db *gorm.DB
}

func NewConflictPolicyService(db *gorm.DB) *ConflictPolicyService {
	return &ConflictPolicyService{db: db}
}

func IsPolicyManagementRole(role string) bool {
	switch normalizeRole(role) {
	case "director", "partner", "management":
		return true
	default:
		return false
	}
}

func IsPolicyComplianceRole(role string) bool {
	switch normalizeRole(role) {
	case "compliance", "risk", "risk_control":
		return true
	default:
		return false
	}
}

func CanManageConflictPolicy(role string) bool {
	return IsPolicyManagementRole(role) || IsPolicyComplianceRole(role)
}

func (s *ConflictPolicyService) CreatePackage(ctx context.Context, actor AuthActor, input ConflictPolicyPackageInput) (*ConflictPolicyPackageView, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("冲突政策服务未初始化")
	}
	if actor.UserID == 0 || !CanManageConflictPolicy(actor.Role) {
		return nil, NewSubjectWorkflowError("POLICY_ROLE_REQUIRED", "只有主任、管理合伙人或合规负责人可以提交政策材料包")
	}
	trimConflictPolicyInput(&input)
	if err := validateConflictPolicyInput(input, time.Now().UTC()); err != nil {
		return nil, err
	}
	policy := models.LawFirmCompliancePolicyPackage{
		ID: uuid.NewString(), PolicyVersion: input.PolicyVersion, Jurisdiction: input.Jurisdiction,
		ApplicableRuleName: input.ApplicableRuleName, ApplicableRuleVersion: input.ApplicableRuleVersion,
		ApplicableRuleAuthority: input.ApplicableRuleAuthority, ApplicableRuleReference: input.ApplicableRuleReference,
		DataSourcePolicyReference: input.DataSourcePolicyReference, PrivacyBasisMatrixReference: input.PrivacyBasisMatrixReference,
		RetentionPolicyReference: input.RetentionPolicyReference, WaiverPolicyReference: input.WaiverPolicyReference,
		ControlledActionsReference: input.ControlledActionsReference, ExternalReviewReference: input.ExternalReviewReference,
		EffectiveAt: input.EffectiveAt.UTC(), NextReviewAt: input.NextReviewAt.UTC(), ExpiresAt: utcTimePointer(input.ExpiresAt),
		IntegrityHash: strings.ToLower(input.IntegrityHash), CreatedBy: actor.UserID, CreatedAt: time.Now().UTC(),
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&policy).Error; err != nil {
			return err
		}
		return createPolicyAudit(tx, actor, "CONFLICT_POLICY_PACKAGE_CREATED", policy.ID, "", "PENDING_ENDORSEMENTS", policy)
	}); err != nil {
		return nil, err
	}
	return &ConflictPolicyPackageView{Package: policy, Endorsements: []ConflictPolicyEndorsementView{}, Status: "PENDING_ENDORSEMENTS"}, nil
}

func (s *ConflictPolicyService) ListPackages(ctx context.Context, actor AuthActor) ([]ConflictPolicyPackageView, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("冲突政策服务未初始化")
	}
	if actor.UserID == 0 || !CanManageConflictPolicy(actor.Role) {
		return nil, NewSubjectWorkflowError("POLICY_ROLE_REQUIRED", "无权查看律所冲突政策签署记录")
	}
	var packages []models.LawFirmCompliancePolicyPackage
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&packages).Error; err != nil {
		return nil, err
	}
	views := make([]ConflictPolicyPackageView, 0, len(packages))
	for _, policy := range packages {
		var endorsements []models.LawFirmCompliancePolicyEndorsement
		if err := s.db.WithContext(ctx).Where("policy_package_id = ?", policy.ID).Order("created_at ASC").Find(&endorsements).Error; err != nil {
			return nil, err
		}
		view, err := buildConflictPolicyView(s.db.WithContext(ctx), policy, endorsements)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *ConflictPolicyService) Endorse(ctx context.Context, actor AuthActor, packageID string) (*ConflictPolicyPackageView, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("冲突政策服务未初始化")
	}
	endorsementType := ""
	if IsPolicyManagementRole(actor.Role) {
		endorsementType = PolicyEndorsementManagement
	} else if IsPolicyComplianceRole(actor.Role) {
		endorsementType = PolicyEndorsementCompliance
	}
	if actor.UserID == 0 || endorsementType == "" {
		return nil, NewSubjectWorkflowError("POLICY_ENDORSEMENT_ROLE_REQUIRED", "当前账号不承担主任/管理合伙人或合规负责人签署职责")
	}
	packageID = strings.TrimSpace(packageID)
	if packageID == "" {
		return nil, NewSubjectWorkflowError("POLICY_PACKAGE_REQUIRED", "政策材料包编号不能为空")
	}

	var result ConflictPolicyPackageView
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var policy models.LawFirmCompliancePolicyPackage
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", packageID).First(&policy).Error; err != nil {
			return err
		}
		var existing models.LawFirmCompliancePolicyEndorsement
		existingErr := tx.Where("policy_package_id = ? AND endorsement_type = ?", policy.ID, endorsementType).First(&existing).Error
		if existingErr == nil {
			if existing.EndorsedBy != actor.UserID {
				return NewSubjectWorkflowError("POLICY_DUTY_ALREADY_ENDORSED", "当前职责已由另一名责任人确认，不能替换既有确认记录")
			}
			var endorsements []models.LawFirmCompliancePolicyEndorsement
			if err := tx.Where("policy_package_id = ?", policy.ID).Order("created_at ASC").Find(&endorsements).Error; err != nil {
				return err
			}
			view, err := buildConflictPolicyView(tx, policy, endorsements)
			if err != nil {
				return err
			}
			result = view
			return nil
		}
		if !errors.Is(existingErr, gorm.ErrRecordNotFound) {
			return existingErr
		}
		endorsement := models.LawFirmCompliancePolicyEndorsement{
			ID: uuid.NewString(), PolicyPackageID: policy.ID, EndorsementType: endorsementType,
			EndorsedBy: actor.UserID, EndorserRole: normalizeRole(actor.Role),
			PackageIntegrityHash: policy.IntegrityHash, CreatedAt: time.Now().UTC(),
		}
		if err := tx.Create(&endorsement).Error; err != nil {
			return err
		}
		if err := createPolicyAudit(tx, actor, "CONFLICT_POLICY_ENDORSED", policy.ID, "PENDING_ENDORSEMENTS", endorsementType+"_ENDORSED", endorsement); err != nil {
			return err
		}

		var endorsements []models.LawFirmCompliancePolicyEndorsement
		if err := tx.Where("policy_package_id = ?", policy.ID).Order("created_at ASC").Find(&endorsements).Error; err != nil {
			return err
		}
		management, compliance := findPolicyEndorsements(endorsements)
		if management != nil && compliance != nil && management.EndorsedBy == compliance.EndorsedBy {
			return NewSubjectWorkflowError("POLICY_INDEPENDENCE_REQUIRED", "管理批准人与合规批准人必须是两个不同账号")
		}
		status := policyEndorsementStatus(endorsements)
		if status == "APPROVED" {
			if management == nil || compliance == nil {
				return NewSubjectWorkflowError("POLICY_INDEPENDENCE_REQUIRED", "管理批准人与合规批准人必须是两个不同账号")
			}
			approvedAt := time.Now().UTC()
			profile := models.LawFirmCompliancePolicyProfile{
				ID: policy.ID, PolicyVersion: policy.PolicyVersion, Status: "APPROVED", Jurisdiction: policy.Jurisdiction,
				ApplicableRuleName: policy.ApplicableRuleName, ApplicableRuleVersion: policy.ApplicableRuleVersion,
				ApplicableRuleAuthority: policy.ApplicableRuleAuthority, ApplicableRuleReference: policy.ApplicableRuleReference,
				DataSourcePolicyReference: policy.DataSourcePolicyReference, PrivacyBasisMatrixReference: policy.PrivacyBasisMatrixReference,
				RetentionPolicyReference: policy.RetentionPolicyReference, WaiverPolicyReference: policy.WaiverPolicyReference,
				ControlledActionsReference: policy.ControlledActionsReference, ExternalReviewReference: policy.ExternalReviewReference,
				ManagementApprovedBy: management.EndorsedBy, ComplianceApprovedBy: compliance.EndorsedBy,
				ApprovedAt: &approvedAt, EffectiveAt: &policy.EffectiveAt, NextReviewAt: &policy.NextReviewAt,
				ExpiresAt: policy.ExpiresAt, IntegrityHash: policy.IntegrityHash, CreatedAt: approvedAt, UpdatedAt: approvedAt,
			}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
			if err := createPolicyAudit(tx, actor, "CONFLICT_POLICY_APPROVED", policy.ID, "PENDING_ENDORSEMENTS", "APPROVED", profile); err != nil {
				return err
			}
		}
		view, err := buildConflictPolicyView(tx, policy, endorsements)
		if err != nil {
			return err
		}
		result = view
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func trimConflictPolicyInput(input *ConflictPolicyPackageInput) {
	input.PolicyVersion = strings.TrimSpace(input.PolicyVersion)
	input.Jurisdiction = strings.TrimSpace(input.Jurisdiction)
	input.ApplicableRuleName = strings.TrimSpace(input.ApplicableRuleName)
	input.ApplicableRuleVersion = strings.TrimSpace(input.ApplicableRuleVersion)
	input.ApplicableRuleAuthority = strings.TrimSpace(input.ApplicableRuleAuthority)
	input.ApplicableRuleReference = strings.TrimSpace(input.ApplicableRuleReference)
	input.DataSourcePolicyReference = strings.TrimSpace(input.DataSourcePolicyReference)
	input.PrivacyBasisMatrixReference = strings.TrimSpace(input.PrivacyBasisMatrixReference)
	input.RetentionPolicyReference = strings.TrimSpace(input.RetentionPolicyReference)
	input.WaiverPolicyReference = strings.TrimSpace(input.WaiverPolicyReference)
	input.ControlledActionsReference = strings.TrimSpace(input.ControlledActionsReference)
	input.ExternalReviewReference = strings.TrimSpace(input.ExternalReviewReference)
	input.IntegrityHash = strings.TrimSpace(input.IntegrityHash)
}

func validateConflictPolicyInput(input ConflictPolicyPackageInput, now time.Time) error {
	required := []string{input.PolicyVersion, input.Jurisdiction, input.ApplicableRuleName, input.ApplicableRuleVersion,
		input.ApplicableRuleAuthority, input.ApplicableRuleReference, input.DataSourcePolicyReference,
		input.PrivacyBasisMatrixReference, input.RetentionPolicyReference, input.WaiverPolicyReference,
		input.ControlledActionsReference, input.ExternalReviewReference}
	for _, value := range required {
		if value == "" {
			return NewSubjectWorkflowError("POLICY_MATERIAL_REQUIRED", "政策版本、适用规则及全部决策材料引用均不能为空")
		}
	}
	if !sha256Pattern.MatchString(input.IntegrityHash) {
		return NewSubjectWorkflowError("POLICY_HASH_INVALID", "材料包完整性摘要必须是64位SHA-256十六进制值")
	}
	if input.EffectiveAt.IsZero() || input.NextReviewAt.IsZero() || !input.NextReviewAt.After(input.EffectiveAt) || !input.NextReviewAt.After(now) {
		return NewSubjectWorkflowError("POLICY_DATES_INVALID", "生效时间和下次复核时间无效；下次复核必须晚于生效时间及当前时间")
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(input.EffectiveAt) {
		return NewSubjectWorkflowError("POLICY_EXPIRY_INVALID", "到期时间必须晚于生效时间")
	}
	return nil
}

func policyEndorsementStatus(endorsements []models.LawFirmCompliancePolicyEndorsement) string {
	management, compliance := findPolicyEndorsements(endorsements)
	if management != nil && compliance != nil && management.EndorsedBy != compliance.EndorsedBy {
		return "APPROVED"
	}
	if management != nil {
		return "PENDING_COMPLIANCE"
	}
	if compliance != nil {
		return "PENDING_MANAGEMENT"
	}
	return "PENDING_ENDORSEMENTS"
}

func buildConflictPolicyView(db *gorm.DB, policy models.LawFirmCompliancePolicyPackage, endorsements []models.LawFirmCompliancePolicyEndorsement) (ConflictPolicyPackageView, error) {
	views := make([]ConflictPolicyEndorsementView, 0, len(endorsements))
	for _, endorsement := range endorsements {
		var user models.User
		if err := db.Select("id", "name", "username").Where("id = ?", endorsement.EndorsedBy).First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ConflictPolicyPackageView{}, err
		}
		name := strings.TrimSpace(user.Name)
		if name == "" {
			name = strings.TrimSpace(user.Username)
		}
		if name == "" {
			name = "已停用账号"
		}
		views = append(views, ConflictPolicyEndorsementView{
			LawFirmCompliancePolicyEndorsement: endorsement,
			EndorserName:                       name,
		})
	}
	return ConflictPolicyPackageView{
		Package: policy, Endorsements: views, Status: policyEndorsementStatus(endorsements),
	}, nil
}

func findPolicyEndorsements(endorsements []models.LawFirmCompliancePolicyEndorsement) (*models.LawFirmCompliancePolicyEndorsement, *models.LawFirmCompliancePolicyEndorsement) {
	var management, compliance *models.LawFirmCompliancePolicyEndorsement
	for index := range endorsements {
		switch endorsements[index].EndorsementType {
		case PolicyEndorsementManagement:
			management = &endorsements[index]
		case PolicyEndorsementCompliance:
			compliance = &endorsements[index]
		}
	}
	return management, compliance
}

func createPolicyAudit(tx *gorm.DB, actor AuthActor, eventType, objectID, fromState, toState string, payload any) error {
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	actorID := actor.UserID
	return tx.Create(&models.ComplianceAuditEvent{
		ID: uuid.NewString(), ActorID: &actorID, ActorRole: normalizeRole(actor.Role), EventType: eventType,
		ObjectType: "LAW_FIRM_COMPLIANCE_POLICY", ObjectID: objectID, FromState: fromState, ToState: toState,
		Payload: string(raw), IntegrityHash: hex.EncodeToString(sum[:]), CreatedAt: time.Now().UTC(),
	}).Error
}

func utcTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
