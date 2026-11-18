package models

import "time"

// ClientCreationRequest records one idempotent client-creation request.
// Scope: operator + Idempotency-Key. Same key + same normalized request
// returns the same client; same key + different request returns 409.
// RequestDigest is a protected HMAC digest without identity plaintext.
type ClientCreationRequest struct {
	ID             uint      `json:"id" gorm:"primarykey"`
	OperatorID     uint      `json:"operator_id" gorm:"not null;uniqueIndex:uq_client_creation_requests_operator_key,priority:1"`
	IdempotencyKey string    `json:"idempotency_key" gorm:"column:idempotency_key;size:120;not null;uniqueIndex:uq_client_creation_requests_operator_key,priority:2"`
	RequestDigest  string    `json:"request_digest" gorm:"column:request_digest;size:128;not null"`
	ClientID       *uint     `json:"client_id,omitempty" gorm:"column:client_id"`
	Status         string    `json:"status" gorm:"column:status;size:20;not null;default:'pending'"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ClientCreationRequest) TableName() string {
	return "client_creation_requests"
}
