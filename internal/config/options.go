package config

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/wilsonpilon/goBaDig/internal/options"
)

// ErrOptionNotSet mirrors ErrNoActiveTheme: returned by GetOption when
// no row exists yet for a (category, key) pair - the caller is
// expected to have run EnsureOptionDefaults first, same as every
// theme.go caller runs EnsureActiveTheme before ActiveTheme.
var ErrOptionNotSet = errors.New("option is not set")

func createOptionSchema(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS options (
    category TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (category, key)
);`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create options table: %w", err)
	}
	return nil
}

// EnsureOptionDefaults seeds every one of fields' own Default value,
// but ONLY for a (category, key) pair that has no row yet - unlike
// InstallThemes (which always upserts the whole catalog, since a
// theme's content IS the catalog), an option's catalog entry is
// capability/shape, not user data: a value the user already changed
// must never be reverted back to its Default just because this ran
// again on the next startup.
func (s *Store) EnsureOptionDefaults(category string, fields []options.Field) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin option defaults: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
INSERT INTO options (category, key, value) VALUES (?, ?, ?)
ON CONFLICT(category, key) DO NOTHING`)
	if err != nil {
		return fmt.Errorf("prepare option defaults: %w", err)
	}
	defer stmt.Close()

	for _, f := range fields {
		if f.Category != category {
			continue
		}
		if _, err := stmt.Exec(f.Category, f.Key, f.Default); err != nil {
			return fmt.Errorf("seed default for option %s.%s: %w", f.Category, f.Key, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit option defaults: %w", err)
	}
	return nil
}

// SetOption validates raw against category/key's catalog Field (same
// principle as InstallThemes calling Theme.Validate before writing)
// and, if valid, stores it under the Field's own canonical Key
// spelling (not necessarily raw's own case - Find/Validate both match
// Key case-insensitively).
func (s *Store) SetOption(category, key, raw string) error {
	field, ok := options.Find(category, key)
	if !ok {
		return fmt.Errorf("unknown option %s.%s", category, key)
	}
	if err := options.Validate(category, field.Key, raw); err != nil {
		return err
	}
	if _, err := s.db.Exec(`
INSERT INTO options (category, key, value) VALUES (?, ?, ?)
ON CONFLICT(category, key) DO UPDATE SET value = excluded.value`, category, field.Key, raw); err != nil {
		return fmt.Errorf("save option %s.%s: %w", category, field.Key, err)
	}
	return nil
}

// GetOption reads back one option's current value. Returns
// ErrOptionNotSet if EnsureOptionDefaults was never run for this
// (category, key) - there is no implicit fallback to the catalog
// Default here, same as ActiveTheme never implicitly falls back to
// some theme without EnsureActiveTheme having run first.
func (s *Store) GetOption(category, key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM options WHERE category = ? AND key = ?`, category, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: %s.%s", ErrOptionNotSet, category, key)
	}
	if err != nil {
		return "", fmt.Errorf("load option %s.%s: %w", category, key, err)
	}
	return value, nil
}

// ListOptions returns every option currently stored for category, as
// key -> value.
func (s *Store) ListOptions(category string) (map[string]string, error) {
	rows, err := s.db.Query(`SELECT key, value FROM options WHERE category = ?`, category)
	if err != nil {
		return nil, fmt.Errorf("list options for %s: %w", category, err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("read option for %s: %w", category, err)
		}
		out[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate options for %s: %w", category, err)
	}
	return out, nil
}
