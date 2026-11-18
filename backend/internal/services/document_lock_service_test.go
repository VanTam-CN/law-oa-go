package services

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

func newDocumentLockTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "document-locks.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Document{}, &models.User{}, &models.DocumentLock{}))
	return db
}

func TestDocumentLockLifecycleWithoutRedis(t *testing.T) {
	db := newDocumentLockTestDB(t)
	require.NoError(t, db.Create(&models.Document{Name: "合同", Status: "active"}).Error)
	service := NewDocumentLockService(db, nil)
	ctx := context.Background()

	first, err := service.AcquireLock(ctx, &AcquireLockRequest{DocumentID: 1, UserID: 10, UserName: "律师A"})
	require.NoError(t, err)
	assert.True(t, first.CanEdit)

	sameUser, err := service.AcquireLock(ctx, &AcquireLockRequest{DocumentID: 1, UserID: 10, UserName: "律师A"})
	require.NoError(t, err)
	assert.True(t, sameUser.CanEdit)
	assert.Equal(t, first.LockedBy, sameUser.LockedBy)

	second, err := service.AcquireLock(ctx, &AcquireLockRequest{DocumentID: 1, UserID: 11, UserName: "律师B"})
	require.NoError(t, err)
	assert.False(t, second.CanEdit)
	assert.Equal(t, uint(10), second.LockedBy)

	renewed, err := service.RenewLock(ctx, &RenewLockRequest{DocumentID: 1, UserID: 11})
	require.Error(t, err)

	renewed, err = service.RenewLock(ctx, &RenewLockRequest{DocumentID: 1, UserID: 10})
	require.NoError(t, err)
	assert.True(t, renewed.CanEdit)
	assert.True(t, renewed.ExpiresAt.After(first.ExpiresAt.Add(-time.Second)))

	require.Error(t, service.ReleaseLock(ctx, &ReleaseLockRequest{DocumentID: 1, UserID: 11}))
	require.NoError(t, service.ReleaseLock(ctx, &ReleaseLockRequest{DocumentID: 1, UserID: 10}))

	status, err := service.GetLockStatus(ctx, 1, 11)
	require.NoError(t, err)
	assert.True(t, status.CanEdit)
	assert.False(t, status.IsLocked)
}

func TestDocumentLockConcurrentAcquireAllowsOneOwner(t *testing.T) {
	db := newDocumentLockTestDB(t)
	require.NoError(t, db.Create(&models.Document{Name: "并发合同", Status: "active"}).Error)
	service := NewDocumentLockService(db, nil)
	ctx := context.Background()

	const workers = 12
	results := make([]*LockStatus, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			status, _ := service.AcquireLock(ctx, &AcquireLockRequest{DocumentID: 1, UserID: uint(i + 1)})
			results[i] = status
		}(i)
	}
	close(start)
	wg.Wait()

	owners := 0
	for _, status := range results {
		if status != nil && status.CanEdit {
			owners++
		}
	}
	assert.Equal(t, 1, owners)
}

func TestDocumentLockExpiredRowCanBeTakenOver(t *testing.T) {
	db := newDocumentLockTestDB(t)
	require.NoError(t, db.Create(&models.Document{Name: "过期合同", Status: "active"}).Error)
	now := time.Now()
	require.NoError(t, db.Create(&models.DocumentLock{
		DocumentID: 1, LockedBy: 10, LockedAt: now.Add(-time.Hour),
		ExpiresAt: now.Add(-time.Minute), LastActivity: &now,
	}).Error)
	service := NewDocumentLockService(db, nil)

	status, err := service.AcquireLock(context.Background(), &AcquireLockRequest{DocumentID: 1, UserID: 11})
	require.NoError(t, err)
	assert.True(t, status.CanEdit)
	assert.Equal(t, uint(11), status.LockedBy)
}
