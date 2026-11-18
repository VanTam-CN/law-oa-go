package config

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuoteLibpqLiteralEscapesSpecialCharacters(t *testing.T) {
	cases := map[string]string{
		"":        "",
		"p w":     "p w",
		"p'w":     "p\\'w",
		"p\\w":    "p\\\\w",
		"p\\'w":   "p\\\\\\'w",
		"'\\'":    "\\'\\\\\\'",
		"trail\\": "trail\\\\",
		"中文 密'码":  "中文 密\\'码",
	}
	for input, want := range cases {
		assert.Equal(t, want, QuoteLibpqLiteral(input), "input %q", input)
	}
}

func TestGetDatabaseDSNQuotesEmptyPassword(t *testing.T) {
	c := &Config{Database: DatabaseConfig{
		Driver: "postgres", Host: "127.0.0.1", Port: "5432",
		Username: "tester", Password: "", Database: "target_db", SSLMode: "disable",
	}}
	dsn := c.GetDatabaseDSN()
	assert.Contains(t, dsn, "password=''")
	assert.Contains(t, dsn, "dbname=target_db")
	cfg, err := pgconn.ParseConfig(dsn)
	require.NoError(t, err)
	assert.Empty(t, cfg.Password)
	assert.Equal(t, "target_db", cfg.Database)
}

func TestGetDatabaseDSNPasswordRoundtripViaPgconn(t *testing.T) {
	passwords := []string{"", "p w", "p'w", "p\\w", "p\\'w", "'\\'", "trail\\", "中文 密'码"}
	for _, pw := range passwords {
		c := &Config{Database: DatabaseConfig{
			Driver: "postgres", Host: "127.0.0.1", Port: "5432",
			Username: "tester", Password: pw, Database: "target_db", SSLMode: "disable",
		}}
		cfg, err := pgconn.ParseConfig(c.GetDatabaseDSN())
		require.NoError(t, err, "password %q", pw)
		assert.Equal(t, pw, cfg.Password, "password %q", pw)
		assert.Equal(t, "target_db", cfg.Database, "password %q", pw)
	}
}
