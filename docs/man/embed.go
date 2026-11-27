// Package docsman embeds the Markdown sources under docs/man/ so
// internal/man can serve them without reading from disk at runtime.
// The embed directive cannot reach outside its own directory, so this
// tiny package lives right next to the .md files it embeds, and
// internal/man imports it instead of embedding docs/man/*.md
// directly. Keeping the actual text under docs/ (not internal/) keeps
// it visible to anyone reading the repository's own documentation,
// same as docs/MANUAL.md.
package docsman

import _ "embed"

// Badig is goBaDig's own command reference (compile/execute/set/
// settings/dialect/theme/version/man), one "# topic" section per
// command - see internal/man/registry.go for how it gets split.
//
//go:embed badig.md
var Badig string

// Asmsx is the asMSX Z80 cross-assembler's own manual (vendored at
// resource/asMSX/doc/asmsx.md, copied here since resource/ is
// gitignored and not distributed - see the attribution note at the
// top of asmsx.md itself), registered as a single "asmsx" topic
// (internal/man/registry.go) regardless of its own heading text.
//
//go:embed asmsx.md
var Asmsx string
