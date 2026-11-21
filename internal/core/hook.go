package core

// HookAction tells the pipeline driver what to do with the token a
// platform hook was given. It replaces the original engine's
// tri-state return protocol (Python False/None/Token) with an
// explicit enum, paired with a Token for the Replace case - decisão
// já tomada em OUTLINE.md, seção 6, item 2.
type HookAction int

const (
	// Continue means the driver should keep using the token it already
	// has and fall through to its own default handling (Python: the
	// hook returned False).
	Continue HookAction = iota
	// Skip means the current token must be discarded entirely (Python:
	// the hook returned None).
	Skip
	// Replace means the current token must be substituted with the one
	// returned alongside this action (Python: the hook returned a Token).
	Replace
)
