package database

import "strings"

func mysqlTLSParam(sslMode string) string {
	mode := strings.TrimSpace(strings.ToLower(sslMode))
	switch mode {
	case "", "disable", "disabled", "false", "0":
		return ""
	case "skip-verify", "preferred", "true":
		return "&tls=" + mode
	default:
		return "&tls=" + sslMode
	}
}
