package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"law-oa-go/internal/models"
)

type mockTokenManager struct{}

func (m *mockTokenManager) VerifyToken(_ context.Context, _ string) (*map[string]interface{}, error) {
	return &map[string]interface{}{}, nil
}

func (m *mockTokenManager) ExtractTokenMetadata(_ context.Context, _ string) (*TokenPayload, error) {
	return nil, fmt.Errorf("not implemented by test double")
}

func (m *mockTokenManager) RevokeAllUserTokens(_ context.Context, _ uint) error {
	return nil
}

func (m *mockTokenManager) BlacklistToken(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (m *mockTokenManager) IsTokenBlacklisted(_ context.Context, _ string) bool {
	return false
}

func newSession(user uint, device string) *models.AuthTokenSession {
	now := time.Now()
	return &models.AuthTokenSession{
		ID:                  "session-" + device,
		UserID:              user,
		DeviceID:            device,
		AccessTokenUUID:     "access-" + device,
		RefreshTokenUUID:    "refresh-" + device,
		AccessTokenExpires:  now.Add(time.Hour),
		RefreshTokenExpires: now.Add(24 * time.Hour),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func TestTokenSessionStoreLifecycle(t *testing.T) {
	db := createAuthTokenDB(t)
	store := NewTokenSessionStore(db)
	ctx := context.Background()
	session := newSession(7, "device-a")
	require.NoError(t, store.Create(ctx, session))

	got, err := store.GetByAccessUUID(ctx, session.AccessTokenUUID)
	require.NoError(t, err)
	assert.Equal(t, session.UserID, got.UserID)

	revokedAt := time.Now()
	sessions, err := store.RevokeDevice(ctx, 7, "device-a", revokedAt)
	require.NoError(t, err)
	require.Len(t, sessions, 1)

	got, err = store.GetByRefreshUUID(ctx, session.RefreshTokenUUID)
	require.NoError(t, err)
	assert.NotNil(t, got.DeviceRevokedAt)

	active, err := store.ListActiveDevices(ctx, 7)
	require.NoError(t, err)
	assert.Empty(t, active)
}

func TestTokenManagerWithoutRedisAndCache(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "lawyer@example.test", Role: "lawyer", Status: "active", Username: "lawyer"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	ctx := context.Background()
	details, err := tm.CreateTokens(ctx, user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	payload, err := tm.ValidateAccess(ctx, details.AccessToken, []string{"lawyer"})
	require.NoError(t, err)
	assert.Equal(t, user.ID, payload.UserID)

	require.NoError(t, tm.RevokeToken(ctx, details.AccessToken))
	_, err = tm.ValidateAccess(ctx, details.AccessToken, nil)
	assert.Error(t, err)
}

func TestSessionBackedTokenCarriesTopLevelAuthClaims(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "claims@example.test", Role: "lawyer", Status: "active", Username: "Claims.Account"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	details, err := tm.CreateTokens(context.Background(), user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	claims, err := tm.VerifyToken(context.Background(), details.AccessToken)
	require.NoError(t, err)
	assert.EqualValues(t, user.ID, (*claims)["user_id"])
	assert.Equal(t, user.Username, (*claims)["username"])
	assert.Equal(t, user.Role, (*claims)["role"])
}

func TestTokenRevocationServiceRevokeSingleWithoutRedis(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "logout@example.test", Role: "lawyer", Status: "active", Username: "logout"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	ctx := context.Background()
	details, err := tm.CreateTokens(ctx, user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	service := NewTokenRevocationService(NewTokenManagerAdapter(tm), nil, db)
	result, err := service.RevokeSingle(ctx, details.AccessToken, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, 1, result.RevokedCount)

	_, err = tm.ValidateAccess(ctx, details.AccessToken, nil)
	assert.Error(t, err)
	assert.True(t, tm.IsTokenBlacklisted(ctx, details.AccessToken))
}

func TestDurableSessionCheckerRejectsRevokedSessionWithoutRedis(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "global@example.test", Role: "lawyer", Status: "active", Username: "global"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	ctx := context.Background()
	details, err := tm.CreateTokens(ctx, user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	service := NewTokenRevocationService(NewTokenManagerAdapter(tm), nil, db)
	err = service.tokenManager.RevokeAllUserTokens(ctx, user.ID)
	require.NoError(t, err)

	claims, err := tm.VerifyToken(ctx, details.AccessToken)
	require.NoError(t, err)
	uuid, _ := (*claims)["uuid"].(string)
	_, err = service.sessions.GetActiveByUUID(ctx, uuid)
	require.Error(t, err)
}

func TestSingleLogoutDoesNotRevokeNewConcurrentSession(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "concurrent@example.test", Role: "lawyer", Status: "active", Username: "concurrent"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	ctx := context.Background()
	loggedOut, err := tm.CreateTokens(ctx, user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)
	active, err := tm.CreateTokens(ctx, user, "device-b", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	service := NewTokenRevocationService(NewTokenManagerAdapter(tm), nil, db)
	_, err = service.RevokeSingle(ctx, loggedOut.AccessToken, "127.0.0.1")
	require.NoError(t, err)

	activeClaims, err := tm.VerifyToken(ctx, active.AccessToken)
	require.NoError(t, err)
	issuedAt := time.Unix(int64((*activeClaims)["iat"].(float64)), 0)
	retrieved, err := service.sessions.GetActiveByUUID(ctx, active.AccessUUID)
	require.NoError(t, err)
	assert.Equal(t, active.AccessUUID, retrieved.AccessTokenUUID)
	assert.False(t, service.IsTokenRevokedForClaims(ctx, active.AccessToken, user.ID, issuedAt))
	_, err = tm.ValidateAccess(ctx, active.AccessToken, nil)
	require.NoError(t, err)
}

func TestUserLevelRevocationDoesNotRevokeLaterSession(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "reset@example.test", Role: "lawyer", Status: "active", Username: "reset"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	ctx := context.Background()
	oldToken, err := tm.CreateTokens(ctx, user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	service := NewTokenRevocationService(NewTokenManagerAdapter(tm), nil, db)
	_, err = service.RevokeByUser(ctx, user.ID, "127.0.0.1")
	require.NoError(t, err)
	_, err = tm.ValidateAccess(ctx, oldToken.AccessToken, nil)
	require.Error(t, err)

	// A password reset must invalidate old sessions, but must not block a
	// legitimate login performed after the reset.
	afterReset, err := tm.CreateTokens(ctx, user, "device-b", "127.0.0.1", "test-agent")
	require.NoError(t, err)
	afterClaims, err := tm.VerifyToken(ctx, afterReset.AccessToken)
	require.NoError(t, err)
	issuedAt := time.Unix(int64((*afterClaims)["iat"].(float64)), 0)
	var logs []models.TokenRevocationLog
	require.NoError(t, db.Where("user_id = ? AND revoke_all = ?", user.ID, true).Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, RevokeByUser, RevocationReason(logs[0].RevocationType))
	assert.True(t, logs[0].RevokedAt.Before(issuedAt.Add(time.Second)))
	retrievedAfter, err := service.sessions.GetActiveByUUID(ctx, afterReset.AccessUUID)
	require.NoError(t, err)
	assert.Equal(t, afterReset.AccessUUID, retrievedAfter.AccessTokenUUID)
	assert.True(t, service.sessions.HasUserTokensRevokedAtOrAfter(ctx, user.ID, issuedAt))
	assert.False(t, service.sessions.HasUserTokensRevokedAtOrAfter(ctx, user.ID, issuedAt.Add(time.Second)))
	assert.False(t, service.IsTokenRevokedForClaims(ctx, afterReset.AccessToken, user.ID, issuedAt))
	_, err = tm.ValidateAccess(ctx, afterReset.AccessToken, nil)
	require.NoError(t, err)
}

func TestTokenManagerRefreshRotatesDurableSession(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "refresh@example.test", Role: "lawyer", Status: "active", Username: "refresh"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	ctx := context.Background()
	details, err := tm.CreateTokens(ctx, user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	rotated, err := tm.RefreshTokens(ctx, details.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, details.RefreshUUID, rotated.RefreshUUID)

	var oldSession models.AuthTokenSession
	require.NoError(t, db.Where("refresh_token_uuid = ?", details.RefreshUUID).First(&oldSession).Error)
	assert.NotNil(t, oldSession.RefreshRevokedAt)

	var newSession models.AuthTokenSession
	require.NoError(t, db.Where("refresh_token_uuid = ?", rotated.RefreshUUID).First(&newSession).Error)
	assert.Nil(t, newSession.RevokedAt)
	assert.Equal(t, "device-a", newSession.DeviceID)

	activeDevices, err := NewTokenSessionStore(db).ListActiveDevices(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, activeDevices, 1)
	assert.Equal(t, rotated.RefreshUUID, activeDevices[0].RefreshTokenUUID)
}

func TestTokenManagerRefreshRejectsConcurrentReplay(t *testing.T) {
	db := createAuthTokenDB(t)
	user := &models.User{Name: "律师", Email: "concurrent-refresh@example.test", Role: "lawyer", Status: "active", Username: "concurrent-refresh"}
	require.NoError(t, db.Create(user).Error)

	tm := NewTokenManager(createTestConfig(), nil, nil, db)
	ctx := context.Background()
	details, err := tm.CreateTokens(ctx, user, "device-a", "127.0.0.1", "test-agent")
	require.NoError(t, err)

	const callers = 8
	results := make(chan error, callers)
	var start sync.WaitGroup
	start.Add(1)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Wait()
			_, err := tm.RefreshTokens(ctx, details.RefreshToken)
			results <- err
		}()
	}
	start.Done()
	wg.Wait()
	close(results)

	successes := 0
	refreshErrors := make([]error, 0, callers)
	for err := range results {
		if err == nil {
			successes++
		} else {
			refreshErrors = append(refreshErrors, err)
			require.ErrorContains(t, err, "refresh token not found or expired")
		}
	}
	require.Equal(t, 1, successes, fmt.Sprintf("refresh results: successes=%d errors=%v", successes, refreshErrors))

	var rotatedSessions []models.AuthTokenSession
	require.NoError(t, db.Where("user_id = ?", user.ID).Order("created_at ASC").Find(&rotatedSessions).Error)
	require.Len(t, rotatedSessions, 2)
}

func TestTokenRevocationServiceRevokeDeviceUsesPostgreSQL(t *testing.T) {
	db := createAuthTokenDB(t)
	store := NewTokenSessionStore(db)
	ctx := context.Background()
	session := newSession(8, "device-b")
	require.NoError(t, store.Create(ctx, session))

	service := NewTokenRevocationService(&mockTokenManager{}, nil, db)
	result, err := service.RevokeByDevice(ctx, 8, "device-b", "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, 1, result.RevokedCount)

	var updated models.AuthTokenSession
	require.NoError(t, db.Where("refresh_token_uuid = ?", session.RefreshTokenUUID).First(&updated).Error)
	assert.NotNil(t, updated.DeviceRevokedAt)
}

func TestTokenRevocationServiceRevokeEmptyDeviceOnlyTargetsEmptyDevice(t *testing.T) {
	db := createAuthTokenDB(t)
	store := NewTokenSessionStore(db)
	ctx := context.Background()
	emptyDevice := newSession(9, "")
	emptyDevice.ID = "session-empty"
	emptyDevice.AccessTokenUUID = "access-empty"
	emptyDevice.RefreshTokenUUID = "refresh-empty"
	require.NoError(t, store.Create(ctx, emptyDevice))
	require.NoError(t, store.Create(ctx, newSession(9, "device-a")))

	service := NewTokenRevocationService(&mockTokenManager{}, nil, db)
	result, err := service.RevokeByDevice(ctx, 9, "", "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, 1, result.RevokedCount)

	active, err := store.ListActiveDevices(ctx, 9)
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, "device-a", active[0].DeviceID)
}
