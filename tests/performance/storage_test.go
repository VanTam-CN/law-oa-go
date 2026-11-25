package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

func BenchmarkDatabaseQueries(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(b, err)
	pool, err := db.DB()
	require.NoError(b, err)
	pool.SetMaxOpenConns(1)
	b.Cleanup(func() { require.NoError(b, pool.Close()) })
	require.NoError(b, db.AutoMigrate(&models.User{}))
	users := make([]models.User, 1000)
	for i := range users {
		name := fmt.Sprintf("User%04d", i)
		users[i] = models.User{Username: name, Name: name, Email: name + "@synthetic.invalid", Role: "lawyer"}
	}
	require.NoError(b, db.CreateInBatches(&users, 100).Error)
	b.Run("SimpleQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			want := users[i%len(users)]
			var row models.User
			if err := db.First(&row, want.ID).Error; err != nil || row.Username != want.Username {
				b.Fatalf("wrong seeded user: %v", err)
			}
		}
	})
	b.Run("PaginatedQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			offset := (i % 10) * 10
			var rows []models.User
			if err := db.Order("id").Offset(offset).Limit(10).Find(&rows).Error; err != nil || len(rows) != 10 || rows[0].ID != users[offset].ID {
				b.Fatalf("wrong seeded page: %v", err)
			}
		}
	})
	b.Run("ComplexQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var n int64
			if err := db.Model(&models.User{}).Where("role=? AND name LIKE ?", "lawyer", "User%").Count(&n).Error; err != nil || n != int64(len(users)) {
				b.Fatalf("wrong row count %d: %v", n, err)
			}
		}
	})
}

func BenchmarkRedisOperations(b *testing.B) {
	r := benchRedis(b)
	ctx := context.Background()
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("np-bench:storage:%d", i)
	}
	b.Cleanup(func() { require.NoError(b, r.Del(ctx, keys...).Err()) })
	b.Run("SET", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			if err := r.Set(ctx, key, key, time.Hour).Err(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GET", func(b *testing.B) {
		pipe := r.Pipeline()
		for _, key := range keys {
			pipe.Set(ctx, key, key, time.Hour)
		}
		_, err := pipe.Exec(ctx)
		require.NoError(b, err)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			if got, err := r.Get(ctx, key).Result(); err != nil || got != key {
				b.Fatalf("wrong Redis value: %v", err)
			}
		}
	})
	b.Run("Pipeline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pipe := r.Pipeline()
			for j := 0; j < 10; j++ {
				key := keys[(i*10+j)%len(keys)]
				pipe.Set(ctx, key, key, time.Hour)
			}
			if cmds, err := pipe.Exec(ctx); err != nil || len(cmds) != 10 {
				b.Fatalf("incomplete pipeline: %v", err)
			}
		}
	})
}
