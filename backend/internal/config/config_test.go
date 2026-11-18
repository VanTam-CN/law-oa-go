package config

import (
	"strings"
	"testing"
)

func TestIsSupportedOriginURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "plain http origin", raw: "http://onlyoffice.example.com", want: true},
		{name: "https origin with port", raw: "https://onlyoffice.example.com:8443/", want: true},
		{name: "reject path", raw: "https://onlyoffice.example.com/editor", want: false},
		{name: "reject query", raw: "https://onlyoffice.example.com?token=1", want: false},
		{name: "reject fragment", raw: "https://onlyoffice.example.com/#editor", want: false},
		{name: "reject credentials", raw: "https://user:pass@onlyoffice.example.com", want: false},
		{name: "reject ftp", raw: "ftp://onlyoffice.example.com", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isSupportedOriginURL(tc.raw); got != tc.want {
				t.Fatalf("isSupportedOriginURL(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestConfigValidateRejectsNonOriginOnlyOfficeURLs(t *testing.T) {
	cfg := &Config{
		Environment: "development",
		Database: DatabaseConfig{
			Host:     "localhost",
			Username: "law",
			Database: "law_oa",
		},
		JWT: JWTConfig{
			Secret: strings.Repeat("a", 32),
		},
		OnlyOffice: OnlyOfficeConfig{
			Enabled:    true,
			URL:        "https://onlyoffice.example.com/editor",
			Secret:     strings.Repeat("b", 32),
			BackendURL: "https://backend.example.com",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want origin validation error")
	}
	if !strings.Contains(err.Error(), "ONLYOFFICE_URL must be a valid HTTP(S) origin") {
		t.Fatalf("Validate() error = %q, want ONLYOFFICE_URL origin validation", err.Error())
	}
}

func TestConfigValidateAcceptsOnlyOfficeOrigins(t *testing.T) {
	cfg := &Config{
		Environment: "development",
		Database: DatabaseConfig{
			Host:     "localhost",
			Username: "law",
			Database: "law_oa",
		},
		JWT: JWTConfig{
			Secret: strings.Repeat("a", 32),
		},
		OnlyOffice: OnlyOfficeConfig{
			Enabled:    true,
			URL:        "https://onlyoffice.example.com",
			Secret:     strings.Repeat("b", 32),
			BackendURL: "https://backend.example.com/",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestConfigValidateRejectsMissingOnlyOfficeBackendURL(t *testing.T) {
	cfg := &Config{
		Environment: "development",
		Database: DatabaseConfig{
			Host:     "localhost",
			Username: "law",
			Database: "law_oa",
		},
		JWT: JWTConfig{
			Secret: strings.Repeat("a", 32),
		},
		OnlyOffice: OnlyOfficeConfig{
			Enabled: true,
			URL:     "https://onlyoffice.example.com",
			Secret:  strings.Repeat("b", 32),
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want backend URL validation error")
	}
	if !strings.Contains(err.Error(), "BACKEND_URL must be configured") {
		t.Fatalf("Validate() error = %q, want BACKEND_URL required validation", err.Error())
	}
}

func TestConfigValidateRejectsBackendOnlyOfficePaths(t *testing.T) {
	cfg := &Config{
		Environment: "development",
		Database: DatabaseConfig{
			Host:     "localhost",
			Username: "law",
			Database: "law_oa",
		},
		JWT: JWTConfig{
			Secret: strings.Repeat("a", 32),
		},
		OnlyOffice: OnlyOfficeConfig{
			Enabled:    true,
			URL:        "https://onlyoffice.example.com",
			Secret:     strings.Repeat("b", 32),
			BackendURL: "https://backend.example.com/callback",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want backend origin validation error")
	}
	if !strings.Contains(err.Error(), "BACKEND_URL must be a valid HTTP(S) origin") {
		t.Fatalf("Validate() error = %q, want BACKEND_URL origin validation", err.Error())
	}
}

func TestConfigValidateRejectsExampleExternalHealthURL(t *testing.T) {
	cfg := &Config{
		Environment: "development",
		Database: DatabaseConfig{
			Host:     "localhost",
			Username: "law",
			Database: "law_oa",
		},
		JWT: JWTConfig{
			Secret: strings.Repeat("a", 32),
		},
		ExternalHealthCheck: ExternalHealthCheckConfig{
			Enabled: true,
			URL:     "https://api.example.com",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want example external health URL validation error")
	}
	if !strings.Contains(err.Error(), "must not use api.example.com") {
		t.Fatalf("Validate() error = %q, want api.example.com rejection", err.Error())
	}
}

func TestConfigValidateAcceptsExplicitExternalHealthURL(t *testing.T) {
	cfg := &Config{
		Environment: "development",
		Database: DatabaseConfig{
			Host:     "localhost",
			Username: "law",
			Database: "law_oa",
		},
		JWT: JWTConfig{
			Secret: strings.Repeat("a", 32),
		},
		ExternalHealthCheck: ExternalHealthCheckConfig{
			Enabled: true,
			URL:     "https://status.example.org",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}
