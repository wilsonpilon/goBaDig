// Package version holds the goBaDig release identity, shared by the
// CLI, TUI and GUI so all three always report the same build.
package version

// Version follows the project's own X.Y.Z scheme (see CHANGELOG.md),
// not SemVer: X is a named block of finished work, Y a named feature
// within that block, Z a build counter within that feature.
const Version = "1.1.3"

// Block is the BASIC-command name attached to X.
const Block = "NEW"

// Feature is the short name attached to Y.
const Feature = "outline"
