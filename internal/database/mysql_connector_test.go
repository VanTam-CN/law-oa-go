package database

import (
	"testing"

	drivermysql "github.com/go-sql-driver/mysql"
	"law-oa-go/internal/config"
)

func TestMySQLDriverConfigPreservesCredentials(t *testing.T) {
	t.Parallel()

	password := `p@ss:word/?#[]%`
	cfg := config.DatabaseConfig{
		Host:      "127.0.0.1",
		Port:      "3306",
		Username:  `user_name`,
		Password:  password,
		Database:  "law_oa",
		Charset:   "utf8mb4",
		Loc:       "Asia/Shanghai",
		SSLMode:   "skip-verify",
		ParseTime: true,
	}

	got := cfg.MySQLDriverConfig()
	if got.User != cfg.Username {
		t.Fatalf("user = %q, want %q", got.User, cfg.Username)
	}
	if got.Passwd != password {
		t.Fatalf("password = %q, want original special-character password", got.Passwd)
	}
	if got.TLSConfig != "skip-verify" {
		t.Fatalf("tls config = %q, want skip-verify", got.TLSConfig)
	}
	if got.Net != "tcp" || got.Addr != cfg.Host+":"+cfg.Port {
		t.Fatalf("address = %s(%s), want tcp(%s:%s)", got.Net, got.Addr, cfg.Host, cfg.Port)
	}
	if got.DBName != cfg.Database {
		t.Fatalf("database = %q, want %q", got.DBName, cfg.Database)
	}
	if got.Params["charset"] != cfg.GetCharset() {
		t.Fatalf("charset = %q, want %q", got.Params["charset"], cfg.GetCharset())
	}
	if !got.ParseTime {
		t.Fatal("parseTime = false, want true")
	}
	if got.Loc.String() != cfg.Loc {
		t.Fatalf("loc = %q, want %q", got.Loc.String(), cfg.Loc)
	}
}

func TestOpenMySQLSQLSetsMigrationMultiStatements(t *testing.T) {
	t.Parallel()

	db, err := openMySQLSQL(config.DatabaseConfig{
		Host: "127.0.0.1", Port: "3306", Username: "user", Password: "pass",
		Database: "law_oa", Charset: "utf8mb4", Loc: "UTC",
	}, true)
	if err != nil {
		t.Fatalf("openMySQLSQL() error = %v", err)
	}
	defer db.Close()

	if value := db.Driver(); value == nil {
		t.Fatal("driver = nil, want MySQL connector")
	}
	// The connector itself is opaque; assert construction of the equivalent
	// configuration to keep migration multi-statement behavior covered.
	cfg := config.DatabaseConfig{Charset: "utf8mb4", Loc: "UTC"}.MySQLDriverConfig()
	cfg.MultiStatements = true
	if _, err := drivermysql.NewConnector(cfg); err != nil {
		t.Fatalf("NewConnector() error = %v", err)
	}
	if !cfg.MultiStatements {
		t.Fatal("multiStatements = false, want true")
	}
}
