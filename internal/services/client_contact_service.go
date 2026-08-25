package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

const clientContactCipherPurpose = "client-contact-v1"

var ErrClientContactVersionConflict = errors.New("client contact version conflict")

type ClientContactService struct {
	db *gorm.DB
}

type SavePrimaryContactRequest struct {
	Version  uint   `json:"version"`
	Name     string `json:"name" binding:"required,max=100"`
	Position string `json:"position" binding:"omitempty,max=100"`
	Phone    string `json:"phone" binding:"omitempty,max=30"`
	Email    string `json:"email" binding:"omitempty,email,max=150"`
}

type ClientContactResponse struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	Name      string    `json:"name"`
	Position  string    `json:"position"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	IsPrimary bool      `json:"is_primary"`
	Version   uint      `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Legacy    bool      `json:"legacy,omitempty"`
}

func NewClientContactService(db *gorm.DB) *ClientContactService {
	return &ClientContactService{db: db}
}

func (s *ClientContactService) GetPrimaryContact(ctx context.Context, clientID uint) (*ClientContactResponse, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("客户联系人服务未初始化")
	}
	var contact models.ClientContact
	err := s.db.WithContext(ctx).Where("client_id = ? AND is_primary = ?", clientID, true).First(&contact).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return contactResponse(&contact)
}

func (s *ClientContactService) SavePrimaryContact(ctx context.Context, clientID, actorID uint, req SavePrimaryContactRequest) (*ClientContactResponse, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("客户联系人服务未初始化")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Position = strings.TrimSpace(req.Position)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" {
		return nil, errors.New("联系人姓名不能为空")
	}
	if len([]rune(req.Name)) > 100 || len([]rune(req.Position)) > 100 || len([]rune(req.Phone)) > 30 || len([]rune(req.Email)) > 150 {
		return nil, errors.New("联系人字段长度超过限制")
	}
	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return nil, errors.New("联系人邮箱格式无效")
		}
	}
	phoneCiphertext, err := security.ProtectSensitiveValue(clientContactCipherPurpose, req.Phone)
	if err != nil {
		return nil, err
	}
	emailCiphertext, err := security.ProtectSensitiveValue(clientContactCipherPurpose, req.Email)
	if err != nil {
		return nil, err
	}

	var saved models.ClientContact
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.ClientContact
		findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("client_id = ? AND is_primary = ?", clientID, true).
			First(&existing).Error
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			saved = models.ClientContact{
				ClientID: clientID, Name: req.Name, Position: req.Position,
				PhoneCiphertext: phoneCiphertext, EmailCiphertext: emailCiphertext,
				IsPrimary: true, Version: 1, CreatedBy: actorID, UpdatedBy: actorID,
			}
			return tx.Create(&saved).Error
		}
		if req.Version == 0 || req.Version != existing.Version {
			return ErrClientContactVersionConflict
		}
		result := tx.Model(&models.ClientContact{}).
			Where("id = ? AND version = ?", existing.ID, req.Version).
			Updates(map[string]interface{}{
				"name": req.Name, "position": req.Position,
				"phone_ciphertext": phoneCiphertext, "email_ciphertext": emailCiphertext,
				"updated_by": actorID, "version": req.Version + 1, "updated_at": time.Now(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrClientContactVersionConflict
		}
		if err := tx.First(&saved, existing.ID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	response, err := contactResponse(&saved)
	if err != nil {
		return nil, fmt.Errorf("读取已保存联系人失败: %w", err)
	}
	return response, nil
}

func contactResponse(contact *models.ClientContact) (*ClientContactResponse, error) {
	phone, err := security.DecryptSensitiveValue(clientContactCipherPurpose, contact.PhoneCiphertext)
	if err != nil {
		return nil, err
	}
	email, err := security.DecryptSensitiveValue(clientContactCipherPurpose, contact.EmailCiphertext)
	if err != nil {
		return nil, err
	}
	return &ClientContactResponse{
		ID: contact.ID, ClientID: contact.ClientID, Name: contact.Name, Position: contact.Position,
		Phone: phone, Email: email, IsPrimary: contact.IsPrimary, Version: contact.Version,
		CreatedAt: contact.CreatedAt, UpdatedAt: contact.UpdatedAt,
	}, nil
}
