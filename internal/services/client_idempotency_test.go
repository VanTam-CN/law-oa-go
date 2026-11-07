package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

type clientCreationTestLog struct {
	mu sync.Mutex
	strings.Builder
}

func (l *clientCreationTestLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.Builder.Write(p)
}

func (l *clientCreationTestLog) content() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.Builder.String()
}

// PostgreSQL is opt-in and limited to this task's local socket/test database.
// Every case gets a fresh private schema, leaving other workers' data alone.
func clientCreationFixture(t *testing.T, dialect string) (*gorm.DB, func() *gorm.DB, *clientCreationTestLog) {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("c", 32))
	logs := &clientCreationTestLog{}
	config := &gorm.Config{Logger: logger.New(log.New(logs, "", 0), logger.Config{LogLevel: logger.Info})}
	var dialector func() gorm.Dialector
	if dialect == "postgres" {
		dsn := os.Getenv("CLIENT_IDEMPOTENCY_TEST_DSN")
		if dsn == "" {
			t.Skip("CLIENT_IDEMPOTENCY_TEST_DSN not set; real PostgreSQL behavior not verified in this run")
		}
		parsed, err := pgconn.ParseConfig(dsn)
		require.NoError(t, err)
		require.Equal(t, "/private/tmp/lawoa-intake-collaboration-runtime/socket", parsed.Host)
		require.True(t, strings.HasPrefix(parsed.Database, "lawoa_client_idempotency_"), "test database required")
		admin, err := gorm.Open(postgres.Open(dsn), config)
		require.NoError(t, err)
		schema := fmt.Sprintf("client_slice_%d", time.Now().UnixNano())
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() {
			require.NoError(t, admin.Exec("DROP SCHEMA "+schema+" CASCADE").Error)
			conn, _ := admin.DB()
			_ = conn.Close()
		})
		dialector = func() gorm.Dialector { return postgres.Open(dsn + " search_path=" + schema) }
	} else {
		dsn := filepath.Join(t.TempDir(), "clients.db") + "?_busy_timeout=10000&_journal_mode=WAL"
		dialector = func() gorm.Dialector { return sqlite.Open(dsn) }
	}
	open := func() *gorm.DB {
		db, err := gorm.Open(dialector(), config)
		require.NoError(t, err)
		conn, err := db.DB()
		require.NoError(t, err)
		conn.SetMaxOpenConns(12)
		t.Cleanup(func() { _ = conn.Close() })
		return db
	}
	db := open()
	require.NoError(t, db.AutoMigrate(&models.Client{}))
	if dialect == "postgres" {
		runClientCreationMigration(t, db, "up")
	} else {
		require.NoError(t, db.AutoMigrate(&models.ClientCreationRequest{}))
	}
	return db, open, logs
}

func runClientCreationMigration(t *testing.T, db *gorm.DB, direction string) {
	t.Helper()
	contents, err := os.ReadFile("../../migrations/000083_client_creation_idempotency." + direction + ".sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(contents)).Error)
}

func clientCreationService(db *gorm.DB) *ClientService {
	return NewClientService(repositories.NewClientRepository(db))
}

func clientCreationRequest() *CreateClientRequest {
	return &CreateClientRequest{Name: "虚构同名客户", Type: "个人", IDCard: "31011019900101001x", Phone: "13900000000"}
}

func clientCreationCounts(t *testing.T, db *gorm.DB, clients, claims int64) {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&models.Client{}).Count(&n).Error)
	require.Equal(t, clients, n)
	require.NoError(t, db.Model(&models.ClientCreationRequest{}).Count(&n).Error)
	require.Equal(t, claims, n)
}

func TestClientCreationIdempotency(t *testing.T) {
	for _, dialect := range []string{"sqlite", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			t.Run("lost_response_reconnect_canonical_replay_and_no_plaintext", func(t *testing.T) {
				db, reopen, logs := clientCreationFixture(t, dialect)
				req := clientCreationRequest()
				original := *req
				key := "request-key-must-not-be-persisted"
				first, err := clientCreationService(db).CreateClientWithIdempotencyKey(context.Background(), req, 7, key)
				require.NoError(t, err)
				require.Equal(t, original, *req, "caller-owned request must not be mutated")
				conn, _ := db.DB()
				require.NoError(t, conn.Close())
				db = reopen()
				retry := *req
				retry.Name = "  " + req.Name + "  "
				retry.IDCard = ""
				retry.IdentityType = " id_card "
				retry.IdentityNumber = "310110-19900101-001X"
				retry.CreatedBy = 999 // must not spoof the authenticated operator
				again, err := clientCreationService(db).CreateClientWithIdempotencyKey(context.Background(), &retry, 7, key)
				require.NoError(t, err)
				require.Equal(t, first.ID, again.ID)
				require.Equal(t, first.IdentityNumber, again.IdentityNumber)
				require.True(t, first.CreatedAt.Equal(again.CreatedAt))
				clientCreationCounts(t, db, 1, 1)
				var stored models.Client
				require.NoError(t, db.First(&stored, first.ID).Error)
				require.Equal(t, uint(7), stored.CreatedBy)
				require.Empty(t, stored.IDCard)
				require.NotEmpty(t, stored.IdentityNumberCiphertext)
				digest, err := security.IdentityDigest("31011019900101001X")
				require.NoError(t, err)
				require.Equal(t, digest, stored.IdentityNumberDigest)
				var rows []map[string]interface{}
				require.NoError(t, db.Table("client_creation_requests").Find(&rows).Error)
				ledger, err := json.Marshal(rows)
				require.NoError(t, err)
				response, err := json.Marshal(again)
				require.NoError(t, err)
				for _, forbidden := range []string{req.IDCard, "31011019900101001X", key} {
					require.NotContains(t, string(ledger), forbidden)
					require.NotContains(t, string(response), forbidden)
					require.NotContains(t, logs.content(), forbidden)
				}
			})

			t.Run("different_payload_conflicts_and_scope_isolation", func(t *testing.T) {
				db, _, _ := clientCreationFixture(t, dialect)
				svc := clientCreationService(db)
				ctx := context.Background()
				req := clientCreationRequest()
				first, err := svc.CreateClientWithIdempotencyKey(ctx, req, 7, "same-key")
				require.NoError(t, err)
				for _, mutate := range []func(*CreateClientRequest){
					func(r *CreateClientRequest) { r.Name = "另一客户" },
					func(r *CreateClientRequest) { r.IDCard = "310110199001010022" },
					func(r *CreateClientRequest) { r.Email = "fictional@example.invalid" },
					func(r *CreateClientRequest) { r.Notes = "新的备注" },
					func(r *CreateClientRequest) { r.ContactPhone = "13900000001" },
				} {
					changed := *req
					mutate(&changed)
					result, err := svc.CreateClientWithIdempotencyKey(ctx, &changed, 7, "same-key")
					require.ErrorIs(t, err, ErrClientIdempotencyConflict)
					require.Nil(t, result)
				}
				second, err := svc.CreateClientWithIdempotencyKey(ctx, req, 8, "same-key")
				require.NoError(t, err)
				require.NotEqual(t, first.ID, second.ID)
				third, err := svc.CreateClientWithIdempotencyKey(ctx, req, 7, "different-key")
				require.NoError(t, err)
				require.NotEqual(t, first.ID, third.ID, "same name/identity is not request-level deduplication")
				// The storage scope includes operation as well as actor and key.
				require.NoError(t, db.Model(&models.ClientCreationRequest{}).Where("operator_id = ?", 8).Update("operation", "other_operation").Error)
				fourth, err := svc.CreateClientWithIdempotencyKey(ctx, req, 8, "same-key")
				require.NoError(t, err)
				require.NotEqual(t, second.ID, fourth.ID)
				clientCreationCounts(t, db, 4, 4)
			})

			for _, different := range []bool{false, true} {
				t.Run(fmt.Sprintf("concurrent_requests_different_payload_%t", different), func(t *testing.T) {
					db, _, _ := clientCreationFixture(t, dialect)
					const workers = 12
					type outcome struct {
						client *ClientResponse
						err    error
						name   string
					}
					outcomes := make(chan outcome, workers)
					start := make(chan struct{})
					for i := 0; i < workers; i++ {
						go func(i int) {
							req := clientCreationRequest()
							if different && i%2 == 0 {
								req.Name = "另一份并发资料"
							}
							<-start
							client, err := clientCreationService(db).CreateClientWithIdempotencyKey(context.Background(), req, 7, "concurrent-key")
							outcomes <- outcome{client, err, req.Name}
						}(i)
					}
					close(start)
					var id uint
					conflicts := 0
					for i := 0; i < workers; i++ {
						out := <-outcomes
						if different && errors.Is(out.err, ErrClientIdempotencyConflict) {
							conflicts++
							continue
						}
						require.NoError(t, out.err)
						if id == 0 {
							id = out.client.ID
						}
						require.Equal(t, id, out.client.ID)
						require.Equal(t, out.name, out.client.Name)
					}
					if different {
						require.Equal(t, workers/2, conflicts)
					}
					clientCreationCounts(t, db, 1, 1)
				})
			}

			t.Run("claim_and_client_rollback_on_late_failure", func(t *testing.T) {
				db, _, _ := clientCreationFixture(t, dialect)
				injected := errors.New("injected result-write failure")
				require.NoError(t, db.Callback().Update().Before("gorm:update").Register("client-test:fail-completion", func(tx *gorm.DB) {
					if tx.Statement.Table == "client_creation_requests" {
						tx.AddError(injected)
					}
				}))
				_, err := clientCreationService(db).CreateClientWithIdempotencyKey(context.Background(), clientCreationRequest(), 7, "rollback-key")
				require.ErrorIs(t, err, injected)
				clientCreationCounts(t, db, 0, 0)
				require.NoError(t, db.Callback().Update().Remove("client-test:fail-completion"))
				_, err = clientCreationService(db).CreateClientWithIdempotencyKey(context.Background(), clientCreationRequest(), 7, "rollback-key")
				require.NoError(t, err)
				clientCreationCounts(t, db, 1, 1)
			})

			t.Run("existing_email_rule_and_failed_claim_release", func(t *testing.T) {
				db, _, _ := clientCreationFixture(t, dialect)
				svc := clientCreationService(db)
				req := clientCreationRequest()
				req.Email = "fictional@example.invalid"
				_, err := svc.CreateClient(context.Background(), req)
				require.NoError(t, err)
				_, err = svc.CreateClientWithIdempotencyKey(context.Background(), req, 7, "retry-after-failure")
				require.Error(t, err)
				clientCreationCounts(t, db, 1, 0)
				req.Email = ""
				_, err = svc.CreateClientWithIdempotencyKey(context.Background(), req, 7, "retry-after-failure")
				require.NoError(t, err)
				clientCreationCounts(t, db, 2, 1)
			})

			t.Run("missing_or_invalid_key_material_fails_closed", func(t *testing.T) {
				db, _, _ := clientCreationFixture(t, dialect)
				for _, secret := range []string{"", "too-short"} {
					t.Setenv("SUBJECT_DATA_KEY", secret)
					_, err := clientCreationService(db).CreateClientWithIdempotencyKey(context.Background(), clientCreationRequest(), 7, "protected-request")
					require.True(t, errors.Is(err, security.ErrSubjectDataKeyMissing) || errors.Is(err, security.ErrSubjectDataKeyInvalid))
					clientCreationCounts(t, db, 0, 0)
				}
			})

			t.Run("deleted_result_never_creates_replacement", func(t *testing.T) {
				db, _, _ := clientCreationFixture(t, dialect)
				svc := clientCreationService(db)
				first, err := svc.CreateClientWithIdempotencyKey(context.Background(), clientCreationRequest(), 7, "deleted-result")
				require.NoError(t, err)
				require.NoError(t, db.Delete(&models.Client{}, first.ID).Error)
				_, err = svc.CreateClientWithIdempotencyKey(context.Background(), clientCreationRequest(), 7, "deleted-result")
				require.ErrorIs(t, err, ErrClientIdempotencyIncomplete)
				clientCreationCounts(t, db, 0, 1)
				var count int64
				require.NoError(t, db.Unscoped().Model(&models.Client{}).Count(&count).Error)
				require.EqualValues(t, 1, count)
			})

			t.Run("no_key_preserves_optional_identity_and_no_ledger_dependency", func(t *testing.T) {
				db, _, _ := clientCreationFixture(t, dialect)
				require.NoError(t, db.Migrator().DropTable(&models.ClientCreationRequest{}))
				t.Setenv("SUBJECT_DATA_KEY", "")
				req := &CreateClientRequest{Name: "  原有名称空格保留  ", Type: "个人"}
				first, err := clientCreationService(db).CreateClient(context.Background(), req)
				require.NoError(t, err)
				second, err := clientCreationService(db).CreateClient(context.Background(), req)
				require.NoError(t, err)
				require.NotEqual(t, first.ID, second.ID)
				require.Equal(t, req.Name, first.Name)
			})
		})
	}
}

func TestClientCreationPostgresMigrationUpDownUp(t *testing.T) {
	db, _, _ := clientCreationFixture(t, "postgres")
	require.True(t, db.Migrator().HasIndex(&models.ClientCreationRequest{}, "uq_client_creation_scope"))
	runClientCreationMigration(t, db, "down")
	require.False(t, db.Migrator().HasTable(&models.ClientCreationRequest{}))
	runClientCreationMigration(t, db, "up")
	_, err := clientCreationService(db).CreateClientWithIdempotencyKey(context.Background(), clientCreationRequest(), 7, "after-up-down-up")
	require.NoError(t, err)
	clientCreationCounts(t, db, 1, 1)
}
