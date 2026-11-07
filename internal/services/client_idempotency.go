package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

var (
	ErrClientIdempotencyConflict    = errors.New("client creation key reused for a different request")
	ErrClientIdempotencyIncomplete  = errors.New("client creation result is unavailable")
	ErrClientIdempotencyUnavailable = errors.New("client creation idempotency store is unavailable")
	ErrClientIdempotencyKeyInvalid  = errors.New("client creation key must contain 1 to 120 printable ASCII characters without spaces")
)

const clientCreationOperation = "create_client"

// The production repository implements this optional narrow capability. Existing
// constructors/routes and repositories used by callers without a key are unchanged.
type clientCreationTransactionStore interface {
	WithClientCreationTransaction(context.Context, func(*gorm.DB) error) error
}

func ValidateClientIdempotencyKey(key string) error {
	if len(key) == 0 || len(key) > 120 {
		return ErrClientIdempotencyKeyInvalid
	}
	for _, ch := range key {
		if ch < 0x21 || ch > 0x7e {
			return ErrClientIdempotencyKeyInvalid
		}
	}
	return nil
}

// canonicalCreateClientRequest is used for both the fingerprint and the write.
// Identity aliases/defaults and number normalization match the existing model.
func canonicalCreateClientRequest(req *CreateClientRequest) CreateClientRequest {
	canonical := *req
	fields := []*string{&canonical.Name, &canonical.Type, &canonical.Email,
		&canonical.Phone, &canonical.Address, &canonical.Aliases, &canonical.Company,
		&canonical.Industry, &canonical.ContactPerson, &canonical.ContactPhone,
		&canonical.Source, &canonical.Notes}
	for _, field := range fields {
		*field = strings.TrimSpace(*field)
	}
	canonical.IdentityType = string(normalizedClientIdentityType(canonical.Type, req.IdentityType))
	identity := strings.TrimSpace(req.IdentityNumber)
	if identity == "" {
		identity = req.IDCard
	}
	canonical.IdentityNumber = security.NormalizeIdentityNumber(canonical.IdentityType, identity)
	canonical.IDCard = ""
	return canonical
}

// CreateClientWithIdempotencyKey replays the same persisted client ID. The caller
// must authorize creation before calling, and authorize the returned object before
// serializing it (including replays after an ethical-wall change).
func (s *ClientService) CreateClientWithIdempotencyKey(ctx context.Context, req *CreateClientRequest, operatorID uint, key string) (*ClientResponse, error) {
	if err := ValidateClientIdempotencyKey(key); err != nil {
		return nil, err
	}
	store, ok := s.clientRepo.(clientCreationTransactionStore)
	if !ok || operatorID == 0 {
		return nil, ErrClientIdempotencyUnavailable
	}
	canonical := canonicalCreateClientRequest(req)
	canonical.CreatedBy = operatorID
	if err := s.validateClientRequest(&canonical); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(canonical)
	if err != nil {
		return nil, err
	}
	requestDigest, err := security.IdentityDigest("law-oa-go:client-creation:request:v1\x00" + string(payload))
	if err != nil {
		return nil, err
	}
	keyDigest, err := security.IdentityDigest("law-oa-go:client-creation:key:v1\x00" + key)
	if err != nil {
		return nil, err
	}
	var result *ClientResponse
	err = store.WithClientCreationTransaction(ctx, func(tx *gorm.DB) error {
		claim := models.ClientCreationRequest{
			OperatorID: operatorID, Operation: clientCreationOperation,
			KeyDigest: keyDigest, RequestDigest: requestDigest, Status: "pending",
		}
		// PostgreSQL requires a savepoint before a possible unique violation.
		// The unique index arbitrates across processes; no process-local lock.
		if err := tx.SavePoint("client_creation_claim").Error; err != nil {
			return err
		}
		if err := tx.Create(&claim).Error; err != nil {
			if !isClientCreationDuplicateKey(err) {
				return err
			}
			if err := tx.RollbackTo("client_creation_claim").Error; err != nil {
				return err
			}
			var existing models.ClientCreationRequest
			// A locking/current read also sees the committed winner on MySQL's
			// repeatable-read isolation, rather than a stale snapshot.
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("operator_id = ? AND operation = ? AND key_digest = ?", operatorID, clientCreationOperation, keyDigest).
				First(&existing).Error; err != nil {
				return err
			}
			if existing.RequestDigest != requestDigest {
				return ErrClientIdempotencyConflict
			}
			if existing.Status != "completed" || existing.ClientID == nil {
				return ErrClientIdempotencyIncomplete
			}
			client, err := repositories.NewClientRepository(tx).FindByID(ctx, *existing.ClientID)
			if err != nil {
				return err
			}
			if client == nil {
				return ErrClientIdempotencyIncomplete
			}
			result = s.toClientResponse(client)
			return nil
		}
		// Reuse validation, email uniqueness, NULL email storage, identity
		// encryption hooks and the existing response projection inside this tx.
		created, err := NewClientService(repositories.NewClientRepository(tx)).CreateClient(ctx, &canonical)
		if err != nil {
			return err
		}
		updated := tx.Model(&models.ClientCreationRequest{}).Where("id = ? AND status = ?", claim.ID, "pending").
			Updates(map[string]interface{}{"client_id": created.ID, "status": "completed"})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return ErrClientIdempotencyIncomplete
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func isClientCreationDuplicateKey(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var sqlState interface{ SQLState() string }
	if errors.As(err, &sqlState) && sqlState.SQLState() == "23505" {
		return true
	}
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	// SQLite's non-translated error is used by the isolated local tests.
	return strings.HasPrefix(err.Error(), "UNIQUE constraint failed: client_creation_requests.")
}
