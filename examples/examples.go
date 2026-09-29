// Package examples embeds the canonical scenarios so that `chaossql demo`
// works from an installed binary, without a checkout of this repository.
package examples

import "embed"

// FS holds every scenario as <dir>/chaos.yaml with its schema.sql and seed.sql.
//
//go:embed */chaos.yaml */schema.sql */seed.sql
var FS embed.FS
