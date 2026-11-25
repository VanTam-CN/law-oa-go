package performance

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"law-oa-go/internal/config"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/router"
)

const (
	npBenchDSN       = "postgres://np_m1:synthetic-runner-only@127.0.0.1:18790/np_benchmark?sslmode=disable"
	npBenchTag       = "np-bench-"
	npBenchEmail     = "np-bench-fixture@synthetic.invalid"
	npBenchPassword  = "synthetic-benchmark-fixture"
	npBenchLoginJSON = `{"email":"` + npBenchEmail + `","password":"` + npBenchPassword + `"}`
)

type benchFixture struct {
	app   *gin.Engine
	token string
}

func benchRedis(b testing.TB) *redis.Client {
	b.Helper()
	require.Equal(b, "127.0.0.1:18980", os.Getenv("NP_BENCH_REDIS_ADDR"), "dedicated benchmark Redis required")
	r := redis.NewClient(&redis.Options{Addr: "127.0.0.1:18980", DB: 12})
	b.Cleanup(func() { require.NoError(b, r.Close()) })
	require.NoError(b, r.Ping(context.Background()).Err())
	return r
}

func setupBenchRouter(b *testing.B) benchFixture {
	b.Helper()
	if os.Getenv("NP_BENCH_PG_DSN") != npBenchDSN {
		b.Fatal("dedicated migrated benchmark database required")
	}
	gin.SetMode(gin.TestMode)
	b.Setenv("JWT_SECRET", "synthetic-benchmark-jwt-secret-fixture-only")
	cfg, err := config.Load()
	require.NoError(b, err)
	middleware.InitJWT(cfg)
	db, err := gorm.Open(postgres.Open(npBenchDSN), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	require.NoError(b, err)
	pool, err := db.DB()
	require.NoError(b, err)
	pool.SetMaxOpenConns(16)
	b.Cleanup(func() { require.NoError(b, pool.Close()) })
	app := gin.New()
	require.NoError(b, router.Init(app, db, benchRedis(b)))
	hash, err := bcrypt.GenerateFromPassword([]byte(npBenchPassword), bcrypt.DefaultCost)
	require.NoError(b, err)
	var user, client, matter int
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw("INSERT INTO users(username,name,email,password,role,status) VALUES (?,?,?,?, 'lawyer','active') RETURNING id", npBenchTag+"lawyer", "Benchmark lawyer", npBenchEmail, string(hash)).Scan(&user).Error; err != nil {
			return err
		}
		if err := tx.Raw("INSERT INTO clients(name,created_by) VALUES (?,?) RETURNING id", npBenchTag+"client", user).Scan(&client).Error; err != nil {
			return err
		}
		return tx.Raw("INSERT INTO cases(case_number,title,client_id,lawyer_id,case_type,created_by) VALUES (?,?,?,?,'civil',?) RETURNING id", "NP-BENCH-CASE", npBenchTag+"case", client, user, fmt.Sprint(user)).Scan(&matter).Error
	})
	require.NoError(b, err)
	b.Cleanup(func() {
		require.NoError(b, db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("DELETE FROM auth_token_sessions WHERE user_id=?", user).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM cases WHERE id=?", matter).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM clients WHERE id=?", client).Error; err != nil {
				return err
			}
			return tx.Exec("DELETE FROM users WHERE id=?", user).Error
		}))
	})
	token, err := loginBenchUser(app, uint(user))
	require.NoError(b, err)
	require.NoError(b, verifyBenchRoutes(app, token))
	return benchFixture{app, token}
}
