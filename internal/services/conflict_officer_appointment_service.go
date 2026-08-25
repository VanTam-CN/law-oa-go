package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

type ConflictOfficerAppointmentInput struct {
	OfficerID                  uint      `json:"officer_id" binding:"required"`
	DeputyID                   *uint     `json:"deputy_id,omitempty"`
	EffectiveFrom              time.Time `json:"effective_from" binding:"required"`
	EffectiveTo                time.Time `json:"effective_to" binding:"required"`
	RecusalDeclaration         string    `json:"recusal_declaration" binding:"required"`
	ExternalMechanismReference string    `json:"external_mechanism_reference"`
}

type ConflictOfficerAppointmentView struct {
	models.ConflictOfficerAppointment
	OfficerName   string `json:"officer_name"`
	DeputyName    string `json:"deputy_name,omitempty"`
	AppointerName string `json:"appointer_name"`
	Current       bool   `json:"current"`
}

type ConflictOfficerAppointmentService struct {
	db *gorm.DB
}

func NewConflictOfficerAppointmentService(db *gorm.DB) *ConflictOfficerAppointmentService {
	return &ConflictOfficerAppointmentService{db: db}
}

func (s *ConflictOfficerAppointmentService) List(ctx context.Context, actor AuthActor) ([]ConflictOfficerAppointmentView, error) {
	if s == nil || s.db == nil {
		return nil, newConflictReviewerError("OFFICER_APPOINTMENT_UNAVAILABLE", "核查人任命服务未初始化")
	}
	if actor.UserID == 0 || (!IsConflictReviewRole(actor.Role) && !IsPolicyManagementRole(actor.Role) && !IsPolicyComplianceRole(actor.Role)) {
		return nil, newConflictReviewerError("OFFICER_APPOINTMENT_FORBIDDEN", "当前账号无权查看冲突核查人任命记录")
	}
	var appointments []models.ConflictOfficerAppointment
	if err := s.db.WithContext(ctx).Order("effective_from DESC, created_at DESC").Find(&appointments).Error; err != nil {
		return nil, err
	}
	return buildAppointmentViews(s.db.WithContext(ctx), appointments)
}

func (s *ConflictOfficerAppointmentService) Create(ctx context.Context, actor AuthActor, input ConflictOfficerAppointmentInput) (*ConflictOfficerAppointmentView, error) {
	if s == nil || s.db == nil {
		return nil, newConflictReviewerError("OFFICER_APPOINTMENT_UNAVAILABLE", "核查人任命服务未初始化")
	}
	if actor.UserID == 0 || !IsPolicyManagementRole(actor.Role) {
		return nil, newConflictReviewerError("OFFICER_APPOINTMENT_FORBIDDEN", "只有主任或管理合伙人可以任命律所冲突核查人及代理人")
	}
	input.RecusalDeclaration = strings.TrimSpace(input.RecusalDeclaration)
	input.ExternalMechanismReference = strings.TrimSpace(input.ExternalMechanismReference)
	input.EffectiveFrom = input.EffectiveFrom.UTC()
	input.EffectiveTo = input.EffectiveTo.UTC()
	if input.OfficerID == 0 || input.EffectiveFrom.IsZero() || input.EffectiveTo.IsZero() || !input.EffectiveTo.After(input.EffectiveFrom) {
		return nil, newConflictReviewerError("OFFICER_APPOINTMENT_INPUT_INVALID", "必须选择核查人并填写有效的任期起止时间")
	}
	if len([]rune(input.RecusalDeclaration)) < 10 {
		return nil, newConflictReviewerError("OFFICER_RECUSAL_DECLARATION_REQUIRED", "回避与独立性声明不得少于10个字")
	}
	if actor.UserID == input.OfficerID || (input.DeputyID != nil && (actor.UserID == *input.DeputyID || input.OfficerID == *input.DeputyID)) {
		return nil, newConflictReviewerError("OFFICER_APPOINTMENT_SEPARATION_REQUIRED", "任命人、主核查人和代理人必须相互独立且使用不同账号")
	}

	appointment := models.ConflictOfficerAppointment{
		ID: uuid.NewString(), OfficerID: input.OfficerID, DeputyID: input.DeputyID, AppointedBy: actor.UserID,
		EffectiveFrom: input.EffectiveFrom, EffectiveTo: input.EffectiveTo,
		RecusalDeclaration: input.RecusalDeclaration, ExternalMechanismReference: input.ExternalMechanismReference,
		CreatedAt: time.Now().UTC(),
	}
	var result ConflictOfficerAppointmentView
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		candidateIDs := []uint{input.OfficerID}
		if input.DeputyID != nil {
			candidateIDs = append(candidateIDs, *input.DeputyID)
		}
		var users []models.User
		if err := tx.Select("id", "name", "username", "role", "status").Where("id IN ? AND deleted_at IS NULL", candidateIDs).Find(&users).Error; err != nil {
			return err
		}
		if len(users) != len(candidateIDs) {
			return newConflictReviewerError("OFFICER_APPOINTMENT_CANDIDATE_INVALID", "主核查人或代理账号不存在")
		}
		for _, user := range users {
			if !strings.EqualFold(strings.TrimSpace(user.Status), "active") || !IsConflictReviewRole(user.Role) {
				return newConflictReviewerError("OFFICER_APPOINTMENT_CANDIDATE_INVALID", "主核查人和代理人必须是启用中的专业冲突复核角色")
			}
		}
		var overlap int64
		if err := tx.Model(&models.ConflictOfficerAppointment{}).
			Where("effective_from < ? AND effective_to > ?", input.EffectiveTo, input.EffectiveFrom).
			Where("officer_id IN ? OR deputy_id IN ?", candidateIDs, candidateIDs).
			Count(&overlap).Error; err != nil {
			return err
		}
		if overlap > 0 {
			return newConflictReviewerError("OFFICER_APPOINTMENT_OVERLAP", "所选核查人或代理人在该任期内已有任命记录，请调整任期")
		}
		if err := tx.Create(&appointment).Error; err != nil {
			return err
		}
		if err := createOfficerAppointmentAudit(tx, actor, appointment); err != nil {
			return err
		}
		views, err := buildAppointmentViews(tx, []models.ConflictOfficerAppointment{appointment})
		if err != nil {
			return err
		}
		result = views[0]
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func buildAppointmentViews(db *gorm.DB, appointments []models.ConflictOfficerAppointment) ([]ConflictOfficerAppointmentView, error) {
	ids := make([]uint, 0, len(appointments)*3)
	for _, item := range appointments {
		ids = append(ids, item.OfficerID, item.AppointedBy)
		if item.DeputyID != nil {
			ids = append(ids, *item.DeputyID)
		}
	}
	var users []models.User
	if len(ids) > 0 {
		if err := db.Select("id", "name", "username").Where("id IN ?", ids).Find(&users).Error; err != nil {
			return nil, err
		}
	}
	names := make(map[uint]string, len(users))
	for _, user := range users {
		name := strings.TrimSpace(user.Name)
		if name == "" {
			name = strings.TrimSpace(user.Username)
		}
		names[user.ID] = name
	}
	now := time.Now().UTC()
	views := make([]ConflictOfficerAppointmentView, 0, len(appointments))
	for _, item := range appointments {
		view := ConflictOfficerAppointmentView{
			ConflictOfficerAppointment: item, OfficerName: names[item.OfficerID], AppointerName: names[item.AppointedBy],
			Current: !item.EffectiveFrom.After(now) && item.EffectiveTo.After(now),
		}
		if item.DeputyID != nil {
			view.DeputyName = names[*item.DeputyID]
		}
		views = append(views, view)
	}
	return views, nil
}

func createOfficerAppointmentAudit(tx *gorm.DB, actor AuthActor, appointment models.ConflictOfficerAppointment) error {
	raw, err := json.Marshal(appointment)
	if err != nil {
		return fmt.Errorf("序列化核查人任命证据失败: %w", err)
	}
	sum := sha256.Sum256(raw)
	actorID := actor.UserID
	return tx.Create(&models.ComplianceAuditEvent{
		ID: uuid.NewString(), ActorID: &actorID, ActorRole: normalizeRole(actor.Role),
		EventType: "CONFLICT_OFFICER_APPOINTED", ObjectType: "CONFLICT_OFFICER_APPOINTMENT", ObjectID: appointment.ID,
		FromState: "", ToState: "ACTIVE_TERM", Payload: string(raw), IntegrityHash: hex.EncodeToString(sum[:]), CreatedAt: time.Now().UTC(),
	}).Error
}
