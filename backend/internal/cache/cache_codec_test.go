package cache

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeInto covers the pure read contract without any network access.
func TestDecodeInto(t *testing.T) {
	t.Run("string destination keeps raw value", func(t *testing.T) {
		var dest string
		require.NoError(t, decodeInto(`{"json":"payload"}`, &dest))
		assert.Equal(t, `{"json":"payload"}`, dest)
	})
	t.Run("bytes destination keeps raw value", func(t *testing.T) {
		var dest []byte
		require.NoError(t, decodeInto(`legacy`, &dest))
		assert.Equal(t, []byte("legacy"), dest)
	})
	t.Run("map destination decodes JSON", func(t *testing.T) {
		var dest map[string]interface{}
		require.NoError(t, decodeInto(`{"device_id":"d1","user_id":7}`, &dest))
		assert.Equal(t, "d1", dest["device_id"])
		assert.Equal(t, float64(7), dest["user_id"])
	})
	t.Run("int destination decodes JSON", func(t *testing.T) {
		var dest int
		require.NoError(t, decodeInto(`3`, &dest))
		assert.Equal(t, 3, dest)
	})
	t.Run("nil destination fails closed", func(t *testing.T) {
		require.Error(t, decodeInto(`anything`, nil))
	})
	t.Run("non-pointer destination fails closed", func(t *testing.T) {
		var dest map[string]interface{}
		require.Error(t, decodeInto(`{}`, dest))
	})
	t.Run("nil pointer destination fails closed", func(t *testing.T) {
		var dest *map[string]interface{}
		require.Error(t, decodeInto(`{}`, dest))
	})
	t.Run("nil string pointer destination fails closed", func(t *testing.T) {
		var dest *string
		require.Error(t, decodeInto(`x`, dest))
	})
	t.Run("nil bytes pointer destination fails closed", func(t *testing.T) {
		var dest *[]byte
		require.Error(t, decodeInto(`x`, dest))
	})
	t.Run("invalid JSON fails closed", func(t *testing.T) {
		var dest map[string]interface{}
		require.Error(t, decodeInto(`not-json`, &dest))
	})
}

// TestSetRejectsNullValues pins the fail-closed contract for untyped and
// typed nil values. The nil check runs before any Redis call, so this
// needs no reachable instance and never writes to the isolated cache.
func TestSetRejectsNullValues(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:18770"})
	defer client.Close()
	svc := NewCacheService(client, "c15_cache_codec")

	t.Run("untyped nil fails closed", func(t *testing.T) {
		err := svc.Set("null:untyped", nil, time.Minute)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cache: unsupported value type")
	})
	t.Run("typed nil map fails closed", func(t *testing.T) {
		var value map[string]interface{}
		err := svc.Set("null:map", value, time.Minute)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cache: unsupported value type")
	})
	t.Run("typed nil pointer fails closed", func(t *testing.T) {
		value := (*map[string]interface{})(nil)
		err := svc.Set("null:pointer", value, time.Minute)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cache: unsupported value type")
	})
}

// TestCacheServiceSetGetRoundTrip exercises the real client against the
// isolated cache instance on 127.0.0.1:18770. It is skipped when that
// instance is not reachable so the suite stays hermetic elsewhere.
func TestCacheServiceSetGetRoundTrip(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:18770"})
	defer client.Close()
	svc := NewCacheService(client, "c15_cache_codec")
	if !svc.Ping() {
		t.Skip("isolated cache instance on 127.0.0.1:18770 not reachable")
	}

	t.Run("device map round trip", func(t *testing.T) {
		want := map[string]interface{}{
			"device_id": "device-1",
			"user_id":   42,
			"ip":        "10.0.0.8",
		}
		require.NoError(t, svc.Set("device:1", want, time.Minute))

		var got map[string]interface{}
		require.NoError(t, svc.Get("device:1", &got))
		assert.Equal(t, want["device_id"], got["device_id"])
		assert.Equal(t, want["ip"], got["ip"])
		// JSON numbers decode as float64; the middleware only reads
		// string fields off the device map, so this is contract-consistent.
		assert.Equal(t, float64(42), got["user_id"])
	})

	t.Run("rate limit counter round trip", func(t *testing.T) {
		require.NoError(t, svc.Set("rate:1", 3, time.Minute))
		var count int
		require.NoError(t, svc.Get("rate:1", &count))
		assert.Equal(t, 3, count)

		require.NoError(t, svc.Set("rate:1", count+1, time.Minute))
		require.NoError(t, svc.Get("rate:1", &count))
		assert.Equal(t, 4, count)
	})

	t.Run("legacy string and bytes stay raw", func(t *testing.T) {
		require.NoError(t, svc.Set("raw:1", "plain-value", time.Minute))

		var raw string
		require.NoError(t, svc.Get("raw:1", &raw))
		assert.Equal(t, "plain-value", raw)

		var rawBytes []byte
		require.NoError(t, svc.Get("raw:1", &rawBytes))
		assert.Equal(t, []byte("plain-value"), rawBytes)
	})
}
