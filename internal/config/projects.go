package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/wilsonpilon/goBaDig/internal/options"
)

// Project is one entry in the "projects" table - a folder, a dialect
// and a name, saved so the planned two-panel file manager
// (internal/filemanager) can list/reopen it quickly. Dialect is one of
// options.CategoryMSX/CategoryZX81/CategoryZXSpectrum - reusing that
// package's own category constants rather than inventing a parallel
// enum, since options.go's own doc comment already named "a real
// compile command" as the anchor project-scoping was waiting for;
// that anchor now exists (internal/dialect). Per-project SETTINGS
// (line start/step, default emulator, etc.) are deliberately NOT part
// of this table yet - confirmed with the user (2026-11-26): a project
// is just name+path+dialect for now.
type Project struct {
	Name      string
	Path      string
	Dialect   string
	CreatedAt time.Time
}

// projectDialects are the only values CreateProject accepts for
// dialectID - the three real dialects, never a bare string the
// caller could typo without a clear error.
var projectDialects = map[string]bool{
	options.CategoryMSX:        true,
	options.CategoryZX81:       true,
	options.CategoryZXSpectrum: true,
}

func createProjectSchema(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS projects (
    name TEXT PRIMARY KEY,
    path TEXT NOT NULL,
    dialect TEXT NOT NULL,
    created_at TEXT NOT NULL
);`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create projects table: %w", err)
	}
	return nil
}

// CreateProject registers a new project: validates dialectID, creates
// path if it doesn't exist yet (os.MkdirAll - safe to call on an
// already-existing directory), and inserts a row keyed by name. name
// must be unique - creating a second project with the same name is a
// real error (a UNIQUE constraint violation bubbling up from SQLite),
// not a silent overwrite - unlike themes' own install-or-update
// convention (ON CONFLICT DO UPDATE), a project is a user-named thing
// the user creates once, not a catalog entry this project keeps
// pre-installed and in sync on every startup.
func (s *Store) CreateProject(name, path, dialectID string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if !projectDialects[dialectID] {
		return fmt.Errorf("unknown project dialect %q (expected %q, %q or %q)",
			dialectID, options.CategoryMSX, options.CategoryZX81, options.CategoryZXSpectrum)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create project directory %s: %w", path, err)
	}
	if _, err := s.db.Exec(`INSERT INTO projects (name, path, dialect, created_at) VALUES (?, ?, ?, ?)`,
		name, path, dialectID, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("save project %q: %w", name, err)
	}
	return nil
}

// ListProjects returns every saved project, most recently created
// first - the "recent projects" quick-list the file manager will
// show. Ties on created_at (two projects created within the same
// clock tick) break by rowid DESC (SQLite's own implicit insertion-
// order column), so "most recent first" stays deterministic rather
// than relying on created_at's own resolution alone.
func (s *Store) ListProjects() ([]Project, error) {
	rows, err := s.db.Query(`SELECT name, path, dialect, created_at FROM projects ORDER BY created_at DESC, rowid DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		var createdAt string
		if err := rows.Scan(&p.Name, &p.Path, &p.Dialect, &createdAt); err != nil {
			return nil, fmt.Errorf("read project: %w", err)
		}
		p.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse created_at for project %q: %w", p.Name, err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}
