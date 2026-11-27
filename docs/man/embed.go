// Package docsman embeds the Markdown sources under docs/man/ so
// internal/man can serve them without reading from disk at runtime -
// go:embed cannot reach outside its own directory, so this tiny package
// lives right next to the .md files it embeds, and internal/man imports
// it instead of embedding docs/man/*.md directly. Keeping the actual
// text under docs/ (not internal/) keeps it visible to anyone reading
// the repository's own documentation, same as docs/MANUAL.md.
package docsman

import _ "embed"

// Badig is goBaDig's own command reference (compile/execute/set/
// settings/dialect/theme/version/man), one "# topic" section per
// command - see internal/man/registry.go for how it gets split.
//
//go:embed badig.md
var Badig string
