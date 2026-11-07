package models

import "time"

// ClientCreationRequest stores only purpose-bound HMACs and a result reference.
// No request body, raw key, identity plaintext or cached response is persisted.
type ClientCreationRequest struct {
	ID            uint      `json:"-" gorm:"primaryKey"`
	OperatorID    uint      `json:"-" gorm:"not null;uniqueIndex:uq_client_creation_scope,priority:1"`
	Operation     string    `json:"-" gorm:"size:32;not null;uniqueIndex:uq_client_creation_scope,priority:2"`
	KeyDigest     string    `json:"-" gorm:"size:64;not null;uniqueIndex:uq_client_creation_scope,priority:3"`
	RequestDigest string    `json:"-" gorm:"size:64;not null"`
	ClientID      *uint     `json:"-"`
	Status        string    `json:"-" gorm:"size:20;not null"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

func (ClientCreationRequest) TableName() string { return "client_creation_requests" }
