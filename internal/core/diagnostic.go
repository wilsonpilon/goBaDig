package core

import (
	"fmt"
	"strings"
)

// Severity classifies a Diagnostic. SeverityError is fatal: a hook
// that hits one returns it directly as an `error`, and the pipeline
// stops immediately - replacing the original engine's
// `Infolog.log(1, ...)`, which always printed and then raised
// SystemExit, with its exit code stored as a mutable class attribute
// set per loaded platform module (a real footgun: see OUTLINE.md,
// seção 3.5). Here the exit code is decided exactly once, in main().
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

// Diagnostic is one reported issue, optionally pinned to a Position.
type Diagnostic struct {
	Severity Severity
	Message  string
	Pos      *Position
}

// Error implements the error interface so a fatal Diagnostic can be
// returned directly as a Go error. The rendering mirrors the original
// engine's error format: the message, the offending source line, and
// a "^" pointer under the exact column (helper.py's Infolog.log).
func (d Diagnostic) Error() string {
	if d.Pos == nil {
		return d.Message
	}
	text := strings.TrimRight(d.Pos.Text(), "\r\n")
	pointerCol := d.Pos.Column - d.Pos.Offset - 1
	if pointerCol < 0 {
		pointerCol = 0
	}
	return fmt.Sprintf("%s\n    %s\n    %s^", d.Message, strings.TrimSpace(text), strings.Repeat("-", pointerCol))
}

// DiagnosticSink collects non-fatal diagnostics (warnings/info) as the
// pipeline runs. What happens to them - printed to stderr, shown in a
// TUI dialog, collected for a GUI panel - is a decision for whoever
// consumes the sink, not for this package: keeps internal/core
// independent of any particular interface (CLI/TUI/GUI parity,
// OUTLINE.md seção 6, item 11).
type DiagnosticSink interface {
	Report(Diagnostic)
}
