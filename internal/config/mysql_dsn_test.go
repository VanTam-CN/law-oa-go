package config

import (
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestGetDatabaseDSNEscapesMySQLCredentials(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Database: DatabaseConfig{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     "3306",
			Username: `user_name`,
			Password: `p@ss:word/?#[]%`,
			Database: "law_oa",
			Charset:  "utf8mb4",
			Loc:      "Asia/Shanghai",
		},
	}

	parsed, err := mysql.ParseDSN(cfg.GetDatabaseDSN())
	if err != nil {
		t.Fatalf("ParseDSN() error = %v", err)
	}
	if parsed.User != cfg.Database.Username {
		t.Fatalf("parsed user = %q, want %q", parsed.User, cfg.Database.Username)
	}
	if parsed.Passwd != cfg.Database.Password {
		t.Fatalf("parsed password = %q, want %q", parsed.Passwd, cfg.Database.Password)
	}
	if parsed.Loc.String() != cfg.Database.Loc {
		t.Fatalf("parsed loc = %q, want %q", parsed.Loc.String(), cfg.Database.Loc)
	}
}
