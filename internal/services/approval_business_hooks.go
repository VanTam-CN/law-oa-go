package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ApprovalBusinessHook 审批业务钩子接口
type ApprovalBusinessHook interface {
	// OnApproved 审批通过后触发
	OnApproved(approval *models.ApprovalRequest) error
	// OnRejected 审批拒绝后触发
	OnRejected(approval *models.ApprovalRequest) error
	// OnCancelled 审批撤回后触发
	OnCancelled(approval *models.ApprovalRequest) error
}

// ApprovalBusinessHookService 审批业务钩子服务
type ApprovalBusinessHookService struct {
	db       *gorm.DB
	hooks    map[string]ApprovalBusinessHook
	caseRepo ApprovalCaseRepository
	sealRepo *SealRecordRepository
}

// NewApprovalBusinessHookService 创建业务钩子服务
func NewApprovalBusinessHookService(db *gorm.DB) *ApprovalBusinessHookService {
	service := &ApprovalBusinessHookService{
		db:       db,
		hooks:    make(map[string]ApprovalBusinessHook),
		caseRepo: NewLocalCaseRepo(db),
		sealRepo: NewSealRecordRepository(db),
	}

	// 注册预置钩子
	service.RegisterHook(models.TemplateSealApproval, &SealApprovalHook{db: db})
	service.RegisterHook(models.TemplateCaseFiling, &CaseCreationApprovalHook{
		db:             db,
		caseRepo:       service.caseRepo,
		subjectRecheck: NewSubjectRecheckService(db, nil),
	})

	return service
}

// RegisterHook 注册业务钩子
func (s *ApprovalBusinessHookService) RegisterHook(workflowType string, hook ApprovalBusinessHook) {
	s.hooks[workflowType] = hook
}

// ExecuteOnApproved 执行审批通过钩子
func (s *ApprovalBusinessHookService) ExecuteOnApproved(approval *models.ApprovalRequest) error {
	hook, exists := s.hooks[approval.WorkflowType]
	if !exists {
		// 没有注册的钩子，跳过
		return nil
	}
	return hook.OnApproved(approval)
}

// ExecuteOnRejected 执行审批拒绝钩子
func (s *ApprovalBusinessHookService) ExecuteOnRejected(approval *models.ApprovalRequest) error {
	hook, exists := s.hooks[approval.WorkflowType]
	if !exists {
		return nil
	}
	return hook.OnRejected(approval)
}

// ExecuteOnCancelled 执行审批撤回钩子
func (s *ApprovalBusinessHookService) ExecuteOnCancelled(approval *models.ApprovalRequest) error {
	hook, exists := s.hooks[approval.WorkflowType]
	if !exists {
		return nil
	}
	return hook.OnCancelled(approval)
}

// ==================== 用印审批钩子 ====================

// SealApprovalHook 用印审批钩子
type SealApprovalHook struct {
	db *gorm.DB
}

// OnApproved 用印审批通过后生成用印记录
func (h *SealApprovalHook) OnApproved(approval *models.ApprovalRequest) error {
	// 解析元数据
	var metadata models.SealApprovalMetadata
	if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err != nil {
		return fmt.Errorf("解析用印元数据失败: %v", err)
	}

	// 创建用印记录
	record := &SealRecord{
		ApprovalID:     approval.ID,
		ApprovalNumber: approval.RequestNumber,
		DocumentTitle:  metadata.DocumentTitle,
		DocumentType:   metadata.DocumentType,
		SealType:       metadata.SealType,
		SealCount:      metadata.SealCount,
		SealImportance: metadata.SealImportance,
		ContractValue:  metadata.ContractValue,
		UsePurpose:     metadata.UsePurpose,
		ApplicantID:    approval.ApplicantID,
		ApplicantName:  approval.ApplicantName,
		DepartmentID:   approval.DepartmentID,
		DepartmentName: approval.DepartmentName,
		SealDate:       time.Now(),
		Status:         "completed",
		CreatedBy:      approval.ApplicantID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.db.Create(record).Error; err != nil {
		return fmt.Errorf("创建用印记录失败: %v", err)
	}

	return nil
}

// OnRejected 用印审批拒绝
func (h *SealApprovalHook) OnRejected(approval *models.ApprovalRequest) error {
	// 记录拒绝原因，不需要特殊处理
	return nil
}

// OnCancelled 用印审批撤回
func (h *SealApprovalHook) OnCancelled(approval *models.ApprovalRequest) error {
	return nil
}

// ==================== 立案审批钩子 ====================

// CaseCreationApprovalHook 立案审批钩子
type CaseCreationApprovalHook struct {
	db             *gorm.DB
	caseRepo       ApprovalCaseRepository
	subjectRecheck *SubjectRecheckService
}

// OnApproved 立案审批通过后创建案件
func (h *CaseCreationApprovalHook) OnApproved(approval *models.ApprovalRequest) error {
	// 解析元数据
	var metadata models.CaseCreationMetadata
	if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err != nil {
		return fmt.Errorf("解析立案元数据失败: %v", err)
	}
	checkID := conflictCheckIDFromMetadata(parseMetadata(approval.Metadata))
	if h.subjectRecheck == nil {
		return NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "主体版本服务未初始化，已阻止审批成案")
	}
	if checkID == "" {
		return NewSubjectWorkflowError("CONFLICT_CHECK_REQUIRED", "审批成案必须绑定已独立复核的利益冲突检测记录")
	}
	if err := h.subjectRecheck.RequireConflictDispositionForCase(context.Background(), checkID, metadata.ClientID, metadata.LawyerID, "approval_case_creation"); err != nil {
		return err
	}

	// 生成案件编号
	caseNumber, err := h.generateCaseNumber()
	if err != nil {
		return fmt.Errorf("生成案件编号失败: %v", err)
	}

	// 创建案件
	now := time.Now()
	newCase := &models.Case{
		CaseNumber:             caseNumber,
		Title:                  metadata.CaseTitle,
		Description:            metadata.CaseDescription,
		ClientID:               metadata.ClientID,
		LawyerID:               metadata.LawyerID,
		CaseType:               metadata.CaseType,
		Priority:               h.mapPriority(metadata.Urgency),
		Status:                 "active",
		StartDate:              &now,
		CreatedBy:              approval.ApplicantID,
		SubjectVersion:         1,
		SubjectState:           models.SubjectStateEffective,
		ConflictCheckID:        checkID,
		ConflictCoverageStatus: "COMPLETE",
	}

	// 开启事务创建案件
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newCase).Error; err != nil {
			return fmt.Errorf("创建案件失败: %v", err)
		}

		// 更新审批元数据，关联案件ID
		updatedMetadataMap := parseMetadata(approval.Metadata)
		updatedMetadataMap["conflict_check_id"] = checkID
		updatedMetadataMap["case_id"] = newCase.ID
		updatedMetadataMap["case_number"] = caseNumber
		updatedMetadataMap["created_at"] = time.Now().Format(time.RFC3339)
		metadataBytes, _ := json.Marshal(updatedMetadataMap)
		approval.Metadata = string(metadataBytes)
		approval.UpdatedBy = approval.ApplicantID

		if err := tx.Save(approval).Error; err != nil {
			return fmt.Errorf("更新审批元数据失败: %v", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// OnRejected 立案审批拒绝
func (h *CaseCreationApprovalHook) OnRejected(approval *models.ApprovalRequest) error {
	return nil
}

// OnCancelled 立案审批撤回
func (h *CaseCreationApprovalHook) OnCancelled(approval *models.ApprovalRequest) error {
	return nil
}

// generateCaseNumber 生成案件编号
func (h *CaseCreationApprovalHook) generateCaseNumber() (string, error) {
	now := time.Now()
	year := now.Year()
	prefix := fmt.Sprintf("LH%d", year)

	// 添加3次重试机制解决并发竞争
	for i := 0; i < 3; i++ {
		var count int64
		if err := h.db.Model(&models.Case{}).
			Where("case_number LIKE ?", prefix+"%").
			Count(&count).Error; err != nil {
			return "", err
		}

		caseNumber := fmt.Sprintf("%s%04d", prefix, count+1)

		// 检查生成的编号是否已存在
		var existing int64
		if err := h.db.Model(&models.Case{}).Where("case_number = ?", caseNumber).Count(&existing).Error; err == nil && existing == 0 {
			return caseNumber, nil
		}
		// 存在冲突，短暂等待后重试
		time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
	}

	// 多次重试失败，使用高精度时间戳作为后缀保证唯一性
	return fmt.Sprintf("%s%s", prefix, time.Now().Format("150405.000")), nil
}

// mapPriority 映射紧急程度到优先级
func (h *CaseCreationApprovalHook) mapPriority(urgency string) string {
	switch urgency {
	case "非常紧急":
		return "critical"
	case "紧急":
		return "high"
	default:
		return "medium"
	}
}

// ==================== 用印记录模型 ====================

// SealRecord 用印记录
type SealRecord struct {
	ID             uint           `gorm:"primarykey"`
	ApprovalID     string         `gorm:"column:approval_id;type:varchar(36);not null;index"`
	ApprovalNumber string         `gorm:"column:approval_number;type:varchar(50);not null"`
	DocumentTitle  string         `gorm:"column:document_title;type:varchar(255);not null"`
	DocumentType   string         `gorm:"column:document_type;type:varchar(100)"`
	SealType       string         `gorm:"column:seal_type;type:varchar(50);not null"`
	SealCount      int            `gorm:"column:seal_count;default:1;not null"`
	SealImportance string         `gorm:"column:seal_importance;type:varchar(20)"`
	ContractValue  float64        `gorm:"column:contract_value"`
	UsePurpose     string         `gorm:"column:use_purpose;type:text"`
	ApplicantID    string         `gorm:"column:applicant_id;type:varchar(36);not null"`
	ApplicantName  string         `gorm:"column:applicant_name;type:varchar(255);not null"`
	DepartmentID   string         `gorm:"column:department_id;type:varchar(36)"`
	DepartmentName string         `gorm:"column:department_name;type:varchar(255)"`
	SealDate       time.Time      `gorm:"column:seal_date;not null"`
	Status         string         `gorm:"column:status;type:varchar(20);default:'completed'"`
	CreatedBy      string         `gorm:"column:created_by;type:varchar(36);not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName 指定表名
func (SealRecord) TableName() string {
	return "seal_records"
}

// SealRecordRepository 用印记录仓库
type SealRecordRepository struct {
	db *gorm.DB
}

// NewSealRecordRepository 创建用印记录仓库
func NewSealRecordRepository(db *gorm.DB) *SealRecordRepository {
	return &SealRecordRepository{db: db}
}

// Create 创建用印记录
func (r *SealRecordRepository) Create(record *SealRecord) error {
	return r.db.Create(record).Error
}

// FindByApprovalID 根据审批ID查找用印记录
func (r *SealRecordRepository) FindByApprovalID(approvalID string) (*SealRecord, error) {
	var record SealRecord
	err := r.db.Where("approval_id = ?", approvalID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ==================== 案件仓库 ====================

// ApprovalCaseRepository 审批用案件仓库接口（避免与冲突检测的CaseRepository冲突）
type ApprovalCaseRepository interface {
	Create(newCase *models.Case) error
	FindByID(id uint) (*models.Case, error)
	FindByCaseNumber(caseNumber string) (*models.Case, error)
}

// LocalCaseRepo 案件仓库实现
type LocalCaseRepo struct {
	db *gorm.DB
}

// NewLocalCaseRepo 创建案件仓库
func NewLocalCaseRepo(db *gorm.DB) ApprovalCaseRepository {
	return &LocalCaseRepo{db: db}
}

// Create 创建案件
func (r *LocalCaseRepo) Create(newCase *models.Case) error {
	return r.db.Create(newCase).Error
}

// FindByID 根据ID查找案件
func (r *LocalCaseRepo) FindByID(id uint) (*models.Case, error) {
	var caseModel models.Case
	err := r.db.First(&caseModel, id).Error
	if err != nil {
		return nil, err
	}
	return &caseModel, nil
}

// FindByCaseNumber 根据案件编号查找案件
func (r *LocalCaseRepo) FindByCaseNumber(caseNumber string) (*models.Case, error) {
	var caseModel models.Case
	err := r.db.Where("case_number = ?", caseNumber).First(&caseModel).Error
	if err != nil {
		return nil, err
	}
	return &caseModel, nil
}
