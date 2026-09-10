package repositories

import (
	"context"
	"errors"
	"fmt"
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

// setupSingleConnCommitDB is the same minimal schema as setupIntakeCommitDB
// but pinned to one pooled connection. The SQLite write lock then serializes
// the two racing transactions: the loser cannot begin reading until the
// winner committed, so it observes the rotated commit token and fails with
// the typed race error instead of a low-level BUSY_SNAPSHOT error.
func setupSingleConnCommitDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("intake_race_%d.db", time.Now().UnixNano()))
	dsn := fmt.Sprintf("file:%s?_busy_timeout=30000&_journal_mode=WAL", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&models.ConflictCheckRecord{}, &models.ConflictCase{}, &models.Client{}))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes ("+
		"id TEXT PRIMARY KEY, client_id INTEGER, status TEXT, metadata TEXT, updated_at DATETIME)").Error)
	return db
}

func seedCommitRaceClient(t *testing.T, db *gorm.DB, name string) models.Client {
	t.Helper()
	client := models.Client{Name: name, Type: "企业"}
	require.NoError(t, db.Create(&client).Error)
	return client
}

// 并发提交竞争：同一接案两次并发 commit 必须恰好成功一次。失败方收到
// IntakeConflictCommitRaceError 并携带获胜检测编号；审计历史只保留获胜
// 记录，接案 metadata 只指向获胜结果，失败方审计记录整体回滚。
func TestCommitIntakeConflictResultConcurrentCommitSingleWinner(t *testing.T) {
	db := setupSingleConnCommitDB(t)
	client := seedCommitRaceClient(t, db, "并发竞争客户")
	seedCommitIntake(t, db, "intake-race-001", client.ID, "lawyer_facts_confirmed",
		"{\"lawyer_facts_confirmed_at\":\"2026-09-10T01:00:00Z\"}")

	const racers = 2
	type commitOutcome struct {
		checkID string
		err     error
	}
	outcomes := make(chan commitOutcome, racers)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(racers)
	for i := 0; i < racers; i++ {
		checkID := fmt.Sprintf("CCT-RACE-%d", i)
		go func() {
			ready.Done()
			<-start
			err := CommitIntakeConflictResult(context.Background(), db,
				buildCommitRecord(checkID, client.ID), nil, nil,
				&models.ConflictCheckResponse{CheckID: checkID},
				commitAssociation(checkID, "intake-race-001", client.ID),
				"2026-09-10T01:00:00Z")
			outcomes <- commitOutcome{checkID: checkID, err: err}
		}()
	}
	ready.Wait()
	close(start)

	var winners, losers []commitOutcome
	for i := 0; i < racers; i++ {
		o := <-outcomes
		if o.err == nil {
			winners = append(winners, o)
		} else {
			losers = append(losers, o)
		}
	}
	require.Len(t, winners, 1, "exactly one concurrent commit may win")
	require.Len(t, losers, 1)
	var raceErr *IntakeConflictCommitRaceError
	require.True(t, errors.As(losers[0].err, &raceErr),
		"loser must fail with the typed race error, got: %v", losers[0].err)
	assert.Equal(t, winners[0].checkID, raceErr.WinningCheckID,
		"race error must point at the winning check id")

	status, metadata := commitIntakeRow(t, db, "intake-race-001")
	assert.Equal(t, "conflict_ready", status)
	assert.Contains(t, metadata, winners[0].checkID)
	assert.NotContains(t, metadata, losers[0].checkID)
	assert.EqualValues(t, 1, countCommitRecords(t, db),
		"losing commit must roll back its audit record entirely")
}

// 过期事实保护：状态离开 lawyer_facts_confirmed 后，即使确认时间戳仍匹配，
// 也不再允许提交，防止用旧检测覆盖新流程状态。
func TestCommitIntakeConflictResultRejectsStaleConfirmedState(t *testing.T) {
	db := setupSingleConnCommitDB(t)
	client := seedCommitRaceClient(t, db, "过期状态客户")
	seedCommitIntake(t, db, "intake-race-002", client.ID, "case_created",
		"{\"lawyer_facts_confirmed_at\":\"2026-09-10T01:00:00Z\"}")

	err := CommitIntakeConflictResult(context.Background(), db,
		buildCommitRecord("CCT-RACE-STALE", client.ID), nil, nil,
		&models.ConflictCheckResponse{CheckID: "CCT-RACE-STALE"},
		commitAssociation("CCT-RACE-STALE", "intake-race-002", client.ID),
		"2026-09-10T01:00:00Z")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "状态")
	var raceErr *IntakeConflictCommitRaceError
	assert.False(t, errors.As(err, &raceErr), "state gate must not surface as a race error")
	status, metadata := commitIntakeRow(t, db, "intake-race-002")
	assert.Equal(t, "case_created", status, "state must stay unchanged on rejection")
	assert.NotContains(t, metadata, "CCT-RACE-STALE")
	assert.EqualValues(t, 0, countCommitRecords(t, db),
		"no orphan audit record may survive a rejected commit")
}
