package config

import "strings"

// QuoteLibpqLiteral renders value as a libpq keyword/value connection-string
// literal. Empty values, spaces, single quotes and backslashes all keep
// their literal meaning: the PostgreSQL wire-level DSN parser (pgconn's
// parseKeywordValueSettings) treats a backslash inside single quotes as the
// escape character for both backslash and quote, so both must be escaped.
// A raw unquoted password that is empty silently swallows the following
// keyword pair, which is the bootstrap failure fixed on 2026-09-26.
func QuoteLibpqLiteral(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	return strings.ReplaceAll(value, "'", "\\'")
}
