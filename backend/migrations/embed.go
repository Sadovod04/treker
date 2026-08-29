// Package migrations embeds the SQL migration files so the server can run them
// on startup without a separate goose binary.
package migrations

import "embed"

// FS holds every *.sql migration in lexical order.
//
//go:embed *.sql
var FS embed.FS
