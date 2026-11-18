package services

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// spyClientRepo wraps the real GORM-backed client writes so idempotency
// transactions run against actual SQL semantics (unique email index, UNIQUE
// constraint errors), while counting Create calls for the replay assertions.
type spyClientRepo struct {
	db      *gorm.DB
	mu      sync.Mutex
	creates int
}

func (s *spyClientRepo) Create(ctx context.Context, client *models.Client) error {
	s.mu.Lock()
	s.creates++
	s.mu.Unlock()
	if client.Email == "" {
		return s.db.WithContext(ctx).Omit("Email").Create(client).Error
	}
	return s.db.WithContext(ctx).Create(client).Error
}

func (s *spyClientRepo) FindByID(ctx context.Context, id uint) (*models.Client, error) {
	var c models.Client
	if err := s.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *spyClientRepo) FindByEmail(ctx context.Context, email string) (*models.Client, error) {
	var c models.Client
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *spyClientRepo) Update(ctx context.Context, client *models.Client) error {
	return s.db.WithContext(ctx).Save(client).Error
}

func (s *spyClientRepo) UpdateWithVersion(ctx context.Context, client *models.Client, expectedVersion uint) error {
	return s.db.WithContext(ctx).Save(client).Error
}

func (s *spyClientRepo) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Client{}, id).Error
}

func (s *spyClientRepo) List(ctx context.Context, params *repositories.ClientListParams) ([]*models.Client, int64, error) {
	return nil, 0, nil
}

func (s *spyClientRepo) GetStats(ctx context.Context) (*repositories.ClientStats, error) {
	return &repositories.ClientStats{}, nil
}

func setupClientIdempotencyDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Client.BeforeSave encrypts identity numbers through SUBJECT_DATA_KEY;
	// tests provide a fixed local key (32 raw bytes) so encryption runs.
	t.Setenv("SUBJECT_DATA_KEY", "0123456789abcdef0123456789abcdef")
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("client_idem_%d.db", time.Now().UnixNano()))
	db, err := gorm.Open(sqlite.Open("file:"+dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Client{}, &models.ClientCreationRequest{}))
	return db
}

func idemService(db *gorm.DB) *ClientService {
	repo := &spyClientRepo{db: db}
	return NewClientService(repo, db)
}

func baseCreateReq(name, email string) *CreateClientRequest {
	return &CreateClientRequest{
		Name:           name,
		Type:           "个人",
		Email:          email,
		IdentityType:   "ID_CARD",
		IdentityNumber: "110101199001011234",
		CreatedBy:      7,
	}
}

func TestCreateClientWithIdempotencyKey_SameRequestReplaysSameClient(t *testing.T) {
	db := setupClientIdempotencyDB(t)
	svc := idemService(db)
	ctx := context.Background()

	first, err := svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("甲客户", "idem-a@example.com"), 7, "key-A")
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("甲客户", "idem-a@example.com"), 7, "key-A")
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)

	var clientCount, claimCount int64
	require.NoError(t, db.Model(&models.Client{}).Count(&clientCount).Error)
	require.NoError(t, db.Model(&models.ClientCreationRequest{}).Count(&claimCount).Error)
	assert.Equal(t, int64(1), clientCount)
	assert.Equal(t, int64(1), claimCount)
}

func TestCreateClientWithIdempotencyKey_SameKeyDifferentPayloadConflicts(t *testing.T) {
	db := setupClientIdempotencyDB(t)
	svc := idemService(db)
	ctx := context.Background()

	_, err := svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("甲客户", "idem-b@example.com"), 7, "key-B")
	require.NoError(t, err)

	_, err = svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("乙客户不同名", "idem-b2@example.com"), 7, "key-B")
	require.ErrorIs(t, err, ErrClientIdempotencyConflict)

	var clientCount int64
	require.NoError(t, db.Model(&models.Client{}).Count(&clientCount).Error)
	assert.Equal(t, int64(1), clientCount)
}

func TestCreateClientWithIdempotencyKey_FailureReleasesKey(t *testing.T) {
	db := setupClientIdempotencyDB(t)
	svc := idemService(db)
	ctx := context.Background()

	existing := &models.Client{Name: "既有客户", Email: "idem-c@example.com"}
	require.NoError(t, db.Create(existing).Error)

	_, err := svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("重名客户", "idem-c@example.com"), 7, "key-C")
	require.Error(t, err)

	var claimCount int64
	require.NoError(t, db.Model(&models.ClientCreationRequest{}).Where("idempotency_key = ?", "key-C").Count(&claimCount).Error)
	assert.Equal(t, int64(0), claimCount, "failed request must roll back the idempotency claim")

	retry, err := svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("新客户", "idem-c2@example.com"), 7, "key-C")
	require.NoError(t, err)
	assert.NotZero(t, retry.ID)
}

func TestCreateClientWithIdempotencyKey_OtherOperatorSameKeyIsolated(t *testing.T) {
	db := setupClientIdempotencyDB(t)
	svc := idemService(db)
	ctx := context.Background()

	first, err := svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("甲客户", "idem-d@example.com"), 7, "shared-key")
	require.NoError(t, err)

	second, err := svc.CreateClientWithIdempotencyKey(ctx, baseCreateReq("乙客户", "idem-d2@example.com"), 8, "shared-key")
	require.NoError(t, err)
	assert.NotEqual(t, first.ID, second.ID)
}

func TestCreateClientWithIdempotencyKey_DigestStableAcrossPadding(t *testing.T) {
	a := baseCreateReq("  甲客户  ", "pad@example.com")
	a.Phone = " 13900000001 "
	b := baseCreateReq("甲客户", "pad@example.com")
	b.Phone = "13900000001"

	da, err := CanonicalCreateClientDigest(a)
	require.NoError(t, err)
	dbv, err := CanonicalCreateClientDigest(b)
	require.NoError(t, err)
	assert.Equal(t, da, dbv)
}
