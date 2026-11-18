package models

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/security"
)

// User 用户模型
type User struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Username     string         `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Name         string         `json:"name" gorm:"column:name;size:50"`
	Email        string         `json:"email" gorm:"size:100;not null;uniqueIndex"`
	Password     string         `json:"-" gorm:"size:255;not null"`
	Role         string         `json:"role" gorm:"size:50;not null;default:'user'"`
	Phone        string         `json:"phone" gorm:"size:20"`
	Avatar       string         `json:"avatar" gorm:"size:255"`
	Status       string         `json:"status" gorm:"size:20;default:'active'"`
	Department   string         `json:"department" gorm:"column:department;size:50;default:'综合部'"` // 部门
	DepartmentID *uint          `json:"department_id,omitempty" gorm:"column:department_id;index"`
	// ManagerID is the authoritative direct-management relation used by
	// conflict-review independence checks. It is deliberately an ID only so
	// loading a user never expands a sensitive reviewer hierarchy implicitly.
	ManagerID *uint  `json:"manager_id,omitempty" gorm:"column:manager_id;index"`
	Seniority string `json:"seniority" gorm:"column:seniority;size:20;default:'初级'"` // 职级：初级/中级/高级/合伙人
}

// Client 客户模型
type Client struct {
	ID                       uint           `json:"id" gorm:"primarykey"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `json:"-" gorm:"index"`
	Name                     string         `json:"name" gorm:"column:name;size:100;not null"`
	Type                     string         `json:"type" gorm:"column:type;size:20;not null;default:'个人'"` // 客户类型：个人/企业
	Email                    string         `json:"email" gorm:"size:100;uniqueIndex"`                     // optional; repositories persist an empty value as SQL NULL
	Phone                    string         `json:"-" gorm:"size:20"`                                      // 手机号（json:"-"防止API泄露，通过ToSafeResponse输出脱敏值）
	Address                  string         `json:"address" gorm:"type:text"`
	Company                  string         `json:"company" gorm:"size:100"`
	IDCard                   string         `json:"-" gorm:"column:id_card;size:18"` // legacy plaintext column; production readiness rejects non-empty values
	IDCardDigest             string         `json:"-" gorm:"column:id_card_digest;size:64;index"`
	IDCardCiphertext         string         `json:"-" gorm:"column:id_card_ciphertext;type:text"`
	IdentityType             IdentityType   `json:"identity_type" gorm:"column:identity_type;type:varchar(30);index"`
	IdentityNumber           string         `json:"-" gorm:"-"`
	IdentityNumberDigest     string         `json:"-" gorm:"column:identity_number_digest;size:64;index"`
	IdentityNumberCiphertext string         `json:"-" gorm:"column:identity_number_ciphertext;type:text"`
	Aliases                  string         `json:"aliases,omitempty" gorm:"column:aliases;type:text"`
	CreatedBy                uint           `json:"-" gorm:"column:created_by;index"`
	Industry                 string         `json:"industry" gorm:"column:industry;size:50"`             // 所属行业（企业客户）
	ContactPerson            string         `json:"contact_person" gorm:"column:contact_person;size:50"` // 联系人（企业客户）
	ContactPhone             string         `json:"-" gorm:"column:contact_phone;size:20"`               // 联系电话（json:"-"防止API泄露）
	Source                   string         `json:"source" gorm:"column:source;size:50"`                 // 客户来源
	Notes                    string         `json:"notes" gorm:"column:notes;type:text"`
	Status                   string         `json:"status" gorm:"size:20;default:'active'"`
	Version                  uint           `json:"version" gorm:"not null;default:1"` // 乐观锁版本号
}

// ClientContact stores a customer's operational contact separately from the
// customer master record. Phone and email are encrypted at rest and are only
// decrypted after object-level client authorization.
type ClientContact struct {
	ID              uint      `json:"id" gorm:"primarykey"`
	ClientID        uint      `json:"client_id" gorm:"not null;index;uniqueIndex:uq_client_primary_contact,where:is_primary = true"`
	Name            string    `json:"name" gorm:"size:100;not null"`
	Position        string    `json:"position" gorm:"size:100"`
	PhoneCiphertext string    `json:"-" gorm:"column:phone_ciphertext;type:text"`
	EmailCiphertext string    `json:"-" gorm:"column:email_ciphertext;type:text"`
	IsPrimary       bool      `json:"is_primary" gorm:"column:is_primary;not null;default:false"`
	Version         uint      `json:"version" gorm:"not null;default:1"`
	CreatedBy       uint      `json:"-" gorm:"column:created_by;not null;index"`
	UpdatedBy       uint      `json:"-" gorm:"column:updated_by;not null;index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ClientContact) TableName() string { return "client_contacts" }

// BeforeSave converts newly supplied client identity numbers into encrypted
// storage. Existing plaintext rows remain visible to the readiness audit until
// an explicit backfill has verified their digest and ciphertext.
func (c *Client) BeforeSave(tx *gorm.DB) error {
	identityNumber := strings.TrimSpace(c.IdentityNumber)
	if identityNumber == "" {
		identityNumber = strings.TrimSpace(c.IDCard)
	}
	if identityNumber == "" {
		return nil
	}
	identityType := c.EffectiveIdentityType()
	normalized := security.NormalizeIdentityNumber(string(identityType), identityNumber)
	ciphertext, digest, err := security.ProtectIdentityNumber(normalized)
	if err != nil {
		return fmt.Errorf("保存客户身份信息失败: %w", err)
	}
	c.IdentityType = identityType
	c.IdentityNumberCiphertext = ciphertext
	c.IdentityNumberDigest = digest
	c.IdentityNumber = ""
	// Keep the legacy protected columns synchronized during the migration
	// window. Runtime reads prefer the generic fields.
	c.IDCardCiphertext = ciphertext
	c.IDCardDigest = digest
	c.IDCard = ""
	return nil
}

// Case 案件模型
type Case struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CaseNumber  string         `json:"case_number" gorm:"column:case_number;size:50;uniqueIndex;not null"` // 案件编号
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	ClientID    uint           `json:"client_id" gorm:"not null;index"`
	Client      *Client        `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	LawyerID    uint           `json:"lawyer_id" gorm:"not null;index"`
	Lawyer      *User          `json:"lawyer,omitempty" gorm:"foreignKey:LawyerID"`
	CaseType    string         `json:"case_type" gorm:"size:50;not null"`
	Priority    string         `json:"priority" gorm:"size:20;default:'medium'"`
	Status      string         `json:"status" gorm:"size:20;default:'pending'"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	CreatedBy   string         `json:"created_by" gorm:"column:created_by;size:36"` // 创建人ID

	// P0 冲突主体版本。subject_snapshot 只在独立复核通过后替换，
	// pending_subject_revision_id 用于阻止绕过重检直接推进案件。
	SubjectVersion           int    `json:"subject_version" gorm:"column:subject_version;default:1"`
	SubjectState             string `json:"subject_state" gorm:"column:subject_state;size:50;default:EFFECTIVE"`
	SubjectSnapshot          string `json:"subject_snapshot,omitempty" gorm:"column:subject_snapshot;type:text"`
	PendingSubjectRevisionID string `json:"pending_subject_revision_id,omitempty" gorm:"column:pending_subject_revision_id;size:36"`
	ConflictCheckID          string `json:"conflict_check_id,omitempty" gorm:"column:conflict_check_id;size:100"`
	ConflictCoverageStatus   string `json:"conflict_coverage_status" gorm:"column:conflict_coverage_status;size:30;default:COVERAGE_LIMITED"`

	// 隔离墙相关字段 (v2.2.0)
	EthicalWallEnabled       bool       `json:"ethical_wall_enabled" gorm:"column:ethical_wall_enabled;default:false;index:idx_ethical_wall;comment:是否启用隔离墙"`
	EthicalWallDescription   string     `json:"ethical_wall_description" gorm:"column:ethical_wall_description;type:text;comment:隔离墙说明"`
	EthicalWallEnabledBy     *uint      `json:"ethical_wall_enabled_by,omitempty" gorm:"column:ethical_wall_enabled_by;comment:启用人ID"`
	EthicalWallEnabledByUser *User      `json:"ethical_wall_enabled_by_user,omitempty" gorm:"foreignKey:EthicalWallEnabledBy"`
	EthicalWallEnabledAt     *time.Time `json:"ethical_wall_enabled_at,omitempty" gorm:"column:ethical_wall_enabled_at;comment:启用时间"`

	// 临时字段，不映射到数据库，仅用于增强冲突检测
	OpposingParty string `json:"opposing_party" gorm:"-"`
	ClientName    string `json:"client_name" gorm:"-"`
}
