# CLAUDE.md

Instructions for Claude Code sessions working in this repository. See
`OUTLINE.md` for the full living plan and architecture, `resume.txt`
(untracked) for session-to-session handoff notes.

## `current/` — always-fresh local dev build

Whenever a change is made to any goBaDig program (CLI/TUI/GUI, any
`internal/*` package), rebuild the executable into `current/` before
considering the change done, so there is always a ready-to-run build
for manual testing on this machine. This directory is gitignored
(`/current/` in `.gitignore`) — nothing in it is ever committed.

Recipe (PowerShell or Bash, from the repo root):

```
# 1. Make sure current/ and current/resource/ exist (idempotent).
mkdir -Force current/resource   # PowerShell
mkdir -p current/resource       # Bash

# 2. Clear everything in current/ EXCEPT resource/, then rebuild.
#    (Bash example; do the equivalent in PowerShell.)
find current -mindepth 1 -maxdepth 1 ! -name resource -exec rm -rf {} +
go build -o current/goBaDig.exe ./cmd/goBaDig
```

Add any other file the program needs at runtime for testing (there are
none today — settings live in the OS user-config dir, not next to the
executable) to the same copy step, alongside the binary.

**Never delete or modify anything under `current/resource/`.** That
subdirectory holds the user's own working test files (sample BASIC
programs, fixtures, whatever they're using to exercise the program by
hand) and persists across rebuilds — it is the one thing in `current/`
that survives the "clear everything" step above. Do not create files
there yourself unless asked; it's the user's workspace, not build
output.

Do not confuse `current/` with `dist/` (versioned release packaging,
built by `build.ps1`, zipped for GitHub Releases) or the top-level
`resource/` (third-party material the project studies/repackages —
unrelated to `current/resource/`'s purpose). See `dist/README.md` and
`resource/README.md`.
