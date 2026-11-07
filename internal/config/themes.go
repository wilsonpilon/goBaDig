package config

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/wilsonpilon/goBaDig/internal/theme"
)

type Store struct {
	db *sql.DB
}

var ErrNoActiveTheme = errors.New("no active theme is configured")

func OpenDefault() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate user config directory: %w", err)
	}
	configDir = filepath.Join(configDir, "goBaDig")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("create config directory: %w", err)
	}
	return Open(filepath.Join(configDir, "settings.db"))
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open settings database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to settings database: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable SQLite foreign keys: %w", err)
	}
	if err := createThemeSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func createThemeSchema(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS themes (
    id TEXT PRIMARY KEY,
    mode TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    palette_json TEXT NOT NULL,
    position INTEGER NOT NULL,
    UNIQUE (mode, name)
);`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create themes table: %w", err)
	}
	const activeSchema = `
CREATE TABLE IF NOT EXISTS active_themes (
    mode TEXT PRIMARY KEY,
    theme_id TEXT NOT NULL REFERENCES themes(id)
);`
	if _, err := db.Exec(activeSchema); err != nil {
		return fmt.Errorf("create active themes table: %w", err)
	}
	return nil
}

func (s *Store) InstallThemes(themes []theme.Theme) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin theme installation: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
INSERT INTO themes (id, mode, name, description, palette_json, position)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    mode = excluded.mode,
    name = excluded.name,
    description = excluded.description,
    palette_json = excluded.palette_json,
    position = excluded.position`)
	if err != nil {
		return fmt.Errorf("prepare theme installation: %w", err)
	}
	defer stmt.Close()

	for position, item := range themes {
		if err := item.Validate(); err != nil {
			return err
		}
		palette, err := json.Marshal(item.Palette)
		if err != nil {
			return fmt.Errorf("encode palette for theme %q: %w", item.ID, err)
		}
		if _, err := stmt.Exec(item.ID, item.Mode, item.Name, item.Description, string(palette), position); err != nil {
			return fmt.Errorf("install theme %q: %w", item.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit theme installation: %w", err)
	}
	return nil
}

func (s *Store) ListThemes(mode string) ([]theme.Theme, error) {
	rows, err := s.db.Query(`
SELECT id, mode, name, description, palette_json
FROM themes
WHERE mode = ?
ORDER BY position, name`, mode)
	if err != nil {
		return nil, fmt.Errorf("list %s themes: %w", mode, err)
	}
	defer rows.Close()

	var themes []theme.Theme
	for rows.Next() {
		var item theme.Theme
		var paletteJSON string
		if err := rows.Scan(&item.ID, &item.Mode, &item.Name, &item.Description, &paletteJSON); err != nil {
			return nil, fmt.Errorf("read %s theme: %w", mode, err)
		}
		if err := json.Unmarshal([]byte(paletteJSON), &item.Palette); err != nil {
			return nil, fmt.Errorf("decode palette for theme %q: %w", item.ID, err)
		}
		if err := item.Validate(); err != nil {
			return nil, err
		}
		themes = append(themes, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s themes: %w", mode, err)
	}
	return themes, nil
}

func (s *Store) ActiveTheme(mode string) (theme.Theme, error) {
	var item theme.Theme
	var paletteJSON string
	err := s.db.QueryRow(`
SELECT t.id, t.mode, t.name, t.description, t.palette_json
FROM active_themes AS a
JOIN themes AS t ON t.id = a.theme_id AND t.mode = a.mode
WHERE a.mode = ?`, mode).Scan(&item.ID, &item.Mode, &item.Name, &item.Description, &paletteJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return theme.Theme{}, fmt.Errorf("%w for %s", ErrNoActiveTheme, mode)
		}
		return theme.Theme{}, fmt.Errorf("load active %s theme: %w", mode, err)
	}
	if err := json.Unmarshal([]byte(paletteJSON), &item.Palette); err != nil {
		return theme.Theme{}, fmt.Errorf("decode active theme %q palette: %w", item.ID, err)
	}
	if err := item.Validate(); err != nil {
		return theme.Theme{}, err
	}
	return item, nil
}

func (s *Store) SetActiveTheme(mode, themeID string) error {
	var exists int
	err := s.db.QueryRow(`SELECT 1 FROM themes WHERE id = ? AND mode = ?`, themeID, mode).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("theme %q is not installed for mode %q", themeID, mode)
	}
	if err != nil {
		return fmt.Errorf("validate %s theme %q: %w", mode, themeID, err)
	}
	if _, err := s.db.Exec(`
INSERT INTO active_themes (mode, theme_id) VALUES (?, ?)
ON CONFLICT(mode) DO UPDATE SET theme_id = excluded.theme_id`, mode, themeID); err != nil {
		return fmt.Errorf("save active %s theme %q: %w", mode, themeID, err)
	}
	return nil
}

func (s *Store) EnsureActiveTheme(mode, defaultID string) (theme.Theme, error) {
	item, err := s.ActiveTheme(mode)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, ErrNoActiveTheme) {
		return theme.Theme{}, err
	}

	themes, listErr := s.ListThemes(mode)
	if listErr != nil {
		return theme.Theme{}, listErr
	}
	foundDefault := false
	for _, candidate := range themes {
		if candidate.ID == defaultID {
			foundDefault = true
			break
		}
	}
	if !foundDefault {
		return theme.Theme{}, fmt.Errorf("default theme %q is not installed for mode %q", defaultID, mode)
	}
	if err := s.SetActiveTheme(mode, defaultID); err != nil {
		return theme.Theme{}, err
	}
	return s.ActiveTheme(mode)
}
