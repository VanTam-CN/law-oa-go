package models

import "time"

// AuthTokenSession is the durable source of truth for issued JWT sessions.
// Redis mirrors these records for performance but never defines authentication
// state in the single-instance MVP deployment.
type AuthTokenSession struct {
	ID string `json:"id" gorm:"primaryKey;size:36"`

	UserID    uint   `json:"user_id" gorm:"column:user_id;not null;index"`
	DeviceID  string `json:"device_id" gorm:"column:device_id;size:128;index"`
	IP        string `json:"ip" gorm:"column:ip;size:45"`
	UserAgent string `json:"user_agent" gorm:"column:user_agent;size:512"`

	AccessTokenUUID     string    `json:"-" gorm:"column:access_token_uuid;size:64;not null;uniqueIndex"`
	RefreshTokenUUID    string    `json:"-" gorm:"column:refresh_token_uuid;size:64;not null;uniqueIndex"`
	AccessTokenExpires  time.Time `json:"access_token_expires" gorm:"column:access_token_expires;not null"`
	RefreshTokenExpires time.Time `json:"refresh_token_expires" gorm:"column:refresh_token_expires;not null"`

	AccessRevokedAt  *time.Time `json:"-" gorm:"column:access_revoked_at;index"`
	RefreshRevokedAt *time.Time `json:"-" gorm:"column:refresh_revoked_at;index"`
	RevokedAt        *time.Time `json:"-" gorm:"column:revoked_at;index"`
	DeviceRevokedAt  *time.Time `json:"-" gorm:"column:device_revoked_at;index"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (AuthTokenSession) TableName() string { return "auth_token_sessions" }
