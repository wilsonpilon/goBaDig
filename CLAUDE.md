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

## Release at every finished step/block

Whenever a planned step or block from `OUTLINE.md` (a "Fase" or
"Bloco") is actually finished and verified (build/vet/test clean), cut
a real release instead of just committing: bump the version, update
every doc, package, tag, push, and publish to GitHub — all of it,
every time, not just the commit. Pedido explícito do usuário em
2026-11-21.

1. **Decide the version bump** per the `X.Y.Z` scheme in
   `CHANGELOG.md` (not SemVer): `Z` for every build, `Y` for a new
   named feature (resets `Z` to 0), `X` for a big finished block
   (resets `Y`/`Z` to 0). Judgment call, not mechanical — when several
   unrelated things finished in the same sitting (as happened
   2026-11-21: a backlog of UI/infra work, then a separate
   `internal/core` first slice), prefer **one release per distinct
   finished thing**, in the order they actually finished, over
   cramming everything into a single bump. Ask the user if the
   boundary is genuinely unclear; don't guess silently on something
   that becomes permanent public history.
2. **Update every doc**, not just `OUTLINE.md`/`resume.txt`:
   `CHANGELOG.md` (close the `[Não lançado]`/previous section, add the
   new `## [X.Y.Z] — data — bloco X, feature Y (build Z)` section),
   `docs/MANUAL.md` (anything the release changes for an actual user:
   new commands/menus, status of pending features), `docs/releases/
   vX.Y.Z.md` (new file, same content that goes on the GitHub release
   - see `docs/releases/README.md`), `internal/version/version.go`
   (`Version`/`Block`/`Feature` consts).
3. **Grep for hardcoded version/feature-name strings in tests**
   (`grep -rn "feature <old-name>" internal`) before running the suite
   - a literal like `"bloco NEW, feature themes"` in a test breaks on
     every bump. Fix it to read `version.Block`/`version.Feature`
     dynamically instead of patching the string each time (done
     2026-11-21 in `internal/cli/console_test.go`).
4. **Build the real package**: `.\build.ps1` (PowerShell - it needs
   `gcc` in PATH for the GUI/CGO build) from the repo root. It runs
   `go vet`/`go test`, compiles, and produces
   `dist/goBaDig-vX.Y.Z.zip` reading the version straight from
   `version.go`. Fails loudly (non-zero exit) if a test breaks -
   fix the test, don't skip it with `-SkipTests` just to get past this
   step.
5. **Commit, tag, push, release** - as separate steps, each confirmed
   before the next, since bundling `git push` together with other git
   commands in one Bash call has been denied by the permission
   classifier in this environment before:
   ```
   git add <the doc/version files for this release>
   git commit -m "..."
   git tag -a vX.Y.Z -m "vX.Y.Z - Block / feature (build Z)"
   git push origin main
   git push origin vX.Y.Z
   gh release create vX.Y.Z "dist/goBaDig-vX.Y.Z.zip" \
     --title "vX.Y.Z — Block / feature (build Z)" \
     --notes-file docs/releases/vX.Y.Z.md
   ```
6. **Rebuild `current/`** (see above) after the version bump, so the
   always-fresh dev build reports the new version too.

The commit that bumps the version should generally be its own commit,
separate from the commit that adds the actual feature code - makes it
possible to tag the release boundary precisely and keeps the release
commit's diff reviewable (version.go + docs only, no code).
