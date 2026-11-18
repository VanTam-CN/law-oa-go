package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"

	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

// Idempotency sentinels for client creation (N2 / acceptance C02).
var (
	// ErrClientIdempotencyConflict means the same operator + Idempotency-Key
	// was reused for a different normalized request payload.
	ErrClientIdempotencyConflict = errClientIdempotency("conflict")
	// ErrClientIdempotencyIncomplete means a previous request with the same
	// key is still pending or its client row can no longer be resolved.
	ErrClientIdempotencyIncomplete = errClientIdempotency("incomplete")
	// ErrClientIdempotencyUnavailable means the idempotency store is not
	// wired into the service.
	ErrClientIdempotencyUnavailable = errClientIdempotency("unavailable")
)

type errClientIdempotency string

func (e errClientIdempotency) Error() string {
	switch e {
	case "conflict":
		return "client idempotency key reused for a different request"
	case "incomplete":
		return "client idempotency request is still incomplete"
	default:
		return "client idempotency store is unavailable"
	}
}

// buildCreateClientModel mirrors the column mapping of the legacy
// CreateClient path so both entry points store identical values.
func buildCreateClientModel(req *CreateClientRequest) *models.Client {
	identityNumber := strings.TrimSpace(req.IdentityNumber)
	if identityNumber == "" {
		identityNumber = strings.TrimSpace(req.IDCard)
	}
	identityType := normalizedClientIdentityType(req.Type, req.IdentityType)
	return &models.Client{
		Name:           req.Name,
		Type:           req.Type,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		IdentityType:   identityType,
		IdentityNumber: identityNumber,
		Aliases:        strings.TrimSpace(req.Aliases),
		CreatedBy:      req.CreatedBy,
		Company:        req.Company,
		Industry:       req.Industry,
		ContactPerson:  req.ContactPerson,
		ContactPhone:   req.ContactPhone,
		Source:         req.Source,
		Notes:          req.Notes,
		Status:         "active",
	}
}

// canonicalCreateClientRequest normalizes whitespace so a client retry that
// only changes padding still counts as the same request.
func canonicalCreateClientRequest(req *CreateClientRequest) CreateClientRequest {
	identityNumber := strings.TrimSpace(req.IdentityNumber)
	if identityNumber == "" {
		identityNumber = strings.TrimSpace(req.IDCard)
	}
	return CreateClientRequest{
		Name:           strings.TrimSpace(req.Name),
		Type:           strings.TrimSpace(req.Type),
		Email:          strings.TrimSpace(req.Email),
		Phone:          strings.TrimSpace(req.Phone),
		Address:        strings.TrimSpace(req.Address),
		IDCard:         "",
		IdentityType:   strings.TrimSpace(req.IdentityType),
		IdentityNumber: identityNumber,
		Aliases:        strings.TrimSpace(req.Aliases),
		Company:        strings.TrimSpace(req.Company),
		Industry:       strings.TrimSpace(req.Industry),
		ContactPerson:  strings.TrimSpace(req.ContactPerson),
		ContactPhone:   strings.TrimSpace(req.ContactPhone),
		Source:         strings.TrimSpace(req.Source),
		Notes:          strings.TrimSpace(req.Notes),
	}
}

// CanonicalCreateClientDigest returns a protected HMAC digest of the
// normalized create-client payload. It never contains identity plaintext:
// storage and logs only ever see the hex digest.
func CanonicalCreateClientDigest(req *CreateClientRequest) (string, error) {
	payload, err := json.Marshal(canonicalCreateClientRequest(req))
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, clientIdempotencyDigestKey())
	_, _ = mac.Write([]byte("law-oa-go:client-creation:v1"))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// clientIdempotencyDigestKey prefers the dedicated subject-data key and falls
// back to the JWT secret. Production config requires both, so the fixed
// constant is only reachable in stripped-down local setups.
func clientIdempotencyDigestKey() []byte {
	if key, err := security.SubjectDataKey(); err == nil && len(key) > 0 {
		return key
	}
	if secret := strings.TrimSpace(os.Getenv("JWT_SECRET")); secret != "" {
		key := make([]byte, 32)
		copy(key, []byte(secret))
		return key
	}
	return []byte("law-oa-go:client-creation-idempotency:local")
}

// isDuplicateKeyError recognizes storage-level unique violations across the
// SQLite test driver and the PostgreSQL production driver.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key value violates unique constraint")
}
