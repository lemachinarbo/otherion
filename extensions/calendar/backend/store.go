package backend

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hkdb/aerion/internal/extensions"
)

// migrations is the per-extension migration sequence for the Calendar
// extension's isolated DB. Each entry runs in version order, idempotent.
var migrations = []extensions.Migration{
	{
		Version: 1,
		SQL: `
			-- Phase 1A placeholder. Kept in place so deployments that ran 1A
			-- before 1B don't see an empty schema-bookkeeping table. Inert
			-- otherwise.
			CREATE TABLE IF NOT EXISTS meta (
				key        TEXT PRIMARY KEY,
				value      TEXT NOT NULL,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`,
	},
	{
		Version: 2,
		SQL: `
			-- Phase 1B: real CalDAV source + calendars schema.
			--
			-- calendar_sources: one row per CalDAV server the user has
			-- connected. Password is NOT stored here — it lives in
			-- coreapi.Storage.Secrets (keyring-first with encrypted-DB
			-- fallback in core's extension_secrets table).
			--
			-- calendars: one row per calendar within a source. CASCADE on
			-- source delete; the source row is the only handle the user has
			-- on this data. ctag + last_synced_at populated by 1C sync.

			CREATE TABLE IF NOT EXISTS calendar_sources (
				id                  TEXT PRIMARY KEY,
				type                TEXT NOT NULL,
				name                TEXT NOT NULL,
				url                 TEXT NOT NULL,
				username            TEXT NOT NULL,
				sync_interval_min   INTEGER NOT NULL DEFAULT 15,
				last_synced_at      INTEGER,
				last_error          TEXT,
				last_error_at       INTEGER,
				account_id          TEXT,
				enabled             INTEGER NOT NULL DEFAULT 1,
				created_at          INTEGER NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_calendar_sources_type ON calendar_sources(type);

			CREATE TABLE IF NOT EXISTS calendars (
				id              TEXT PRIMARY KEY,
				source_id       TEXT NOT NULL REFERENCES calendar_sources(id) ON DELETE CASCADE,
				url             TEXT NOT NULL,
				display_name    TEXT NOT NULL,
				description     TEXT,
				color           TEXT,
				visible         INTEGER NOT NULL DEFAULT 1,
				ctag            TEXT,
				last_synced_at  INTEGER,
				created_at      INTEGER NOT NULL,
				UNIQUE(source_id, url)
			);
			CREATE INDEX IF NOT EXISTS idx_calendars_source ON calendars(source_id);
		`,
	},
}

// Store wraps the per-extension DB for the Calendar extension. Lives in an
// isolated SQLite file at <dataDir>/extensions/calendar/data.db, separate
// from Aerion's main DB. No tables in this file are read or written by
// core code; cross-extension access (none exists yet for Calendar) flows
// through coreapi only.
type Store struct {
	*extensions.Store
}

// NewStore opens the Calendar extension's isolated SQLite DB and applies
// pending migrations. Called eagerly from App.Startup whether or not the
// extension is enabled — keeps the schema valid across enable/disable
// cycles. The same pattern Contacts uses.
func NewStore(dataDir string) (*Store, error) {
	s, err := extensions.OpenStore(dataDir, "calendar", migrations)
	if err != nil {
		return nil, err
	}
	return &Store{Store: s}, nil
}

// Source is the Go type returned by the API + Wails methods for a calendar
// source row. JSON tags drive the TS binding shape — keep stable.
type Source struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Name            string `json:"name"`
	URL             string `json:"url"`
	Username        string `json:"username"`
	SyncIntervalMin int    `json:"syncIntervalMin"`
	LastSyncedAt    int64  `json:"lastSyncedAt"`
	LastError       string `json:"lastError,omitempty"`
	LastErrorAt     int64  `json:"lastErrorAt,omitempty"`
	AccountID       string `json:"accountId,omitempty"`
	Enabled         bool   `json:"enabled"`
	CreatedAt       int64  `json:"createdAt"`
}

// Calendar is the Go type for one calendar row within a source.
type Calendar struct {
	ID           string `json:"id"`
	SourceID     string `json:"sourceId"`
	URL          string `json:"url"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description,omitempty"`
	Color        string `json:"color,omitempty"`
	Visible      bool   `json:"visible"`
	Ctag         string `json:"ctag,omitempty"`
	LastSyncedAt int64  `json:"lastSyncedAt"`
	CreatedAt    int64  `json:"createdAt"`
}

// WithTx runs fn inside a SQLite transaction on the per-extension DB.
// Rolls back on any error returned from fn; commits on nil. Matches the
// transaction helper carddav.Store uses for multi-row atomic operations.
func (s *Store) WithTx(fn func(*sql.Tx) error) error {
	tx, err := s.DB().Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// CreateSourceTx inserts a calendar_sources row inside an existing
// transaction. The caller is responsible for committing.
func (s *Store) CreateSourceTx(tx *sql.Tx, src Source) error {
	if src.ID == "" {
		return errors.New("calendar.Store: source ID required")
	}
	if src.CreatedAt == 0 {
		src.CreatedAt = time.Now().Unix()
	}
	if src.SyncIntervalMin == 0 {
		src.SyncIntervalMin = 15
	}
	_, err := tx.Exec(`
		INSERT INTO calendar_sources (
			id, type, name, url, username, sync_interval_min,
			last_synced_at, last_error, last_error_at,
			account_id, enabled, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		src.ID, src.Type, src.Name, src.URL, src.Username, src.SyncIntervalMin,
		nullIfZero(src.LastSyncedAt), nullIfEmpty(src.LastError), nullIfZero(src.LastErrorAt),
		nullIfEmpty(src.AccountID), boolToInt(src.Enabled), src.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert calendar_source: %w", err)
	}
	return nil
}

// GetSource returns one source by id, or sql.ErrNoRows when missing.
func (s *Store) GetSource(id string) (*Source, error) {
	row := s.DB().QueryRow(`
		SELECT id, type, name, url, username, sync_interval_min,
		       COALESCE(last_synced_at, 0), COALESCE(last_error, ''), COALESCE(last_error_at, 0),
		       COALESCE(account_id, ''), enabled, created_at
		FROM calendar_sources WHERE id = ?`, id)
	src := &Source{}
	var enabled int
	if err := row.Scan(
		&src.ID, &src.Type, &src.Name, &src.URL, &src.Username, &src.SyncIntervalMin,
		&src.LastSyncedAt, &src.LastError, &src.LastErrorAt,
		&src.AccountID, &enabled, &src.CreatedAt,
	); err != nil {
		return nil, err
	}
	src.Enabled = enabled != 0
	return src, nil
}

// ListSources returns all configured calendar sources ordered by created_at.
func (s *Store) ListSources() ([]Source, error) {
	rows, err := s.DB().Query(`
		SELECT id, type, name, url, username, sync_interval_min,
		       COALESCE(last_synced_at, 0), COALESCE(last_error, ''), COALESCE(last_error_at, 0),
		       COALESCE(account_id, ''), enabled, created_at
		FROM calendar_sources ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("query calendar_sources: %w", err)
	}
	defer rows.Close()

	var out []Source
	for rows.Next() {
		var src Source
		var enabled int
		if err := rows.Scan(
			&src.ID, &src.Type, &src.Name, &src.URL, &src.Username, &src.SyncIntervalMin,
			&src.LastSyncedAt, &src.LastError, &src.LastErrorAt,
			&src.AccountID, &enabled, &src.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan calendar_source: %w", err)
		}
		src.Enabled = enabled != 0
		out = append(out, src)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate calendar_sources: %w", err)
	}
	return out, nil
}

// DeleteSource removes a calendar_sources row. CASCADE removes the
// associated calendars rows. Idempotent — deleting a non-existent source
// is not an error.
func (s *Store) DeleteSource(id string) error {
	_, err := s.DB().Exec(`DELETE FROM calendar_sources WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete calendar_source: %w", err)
	}
	return nil
}

// CreateCalendarTx inserts a calendars row inside an existing transaction.
func (s *Store) CreateCalendarTx(tx *sql.Tx, cal Calendar) error {
	if cal.ID == "" {
		return errors.New("calendar.Store: calendar ID required")
	}
	if cal.SourceID == "" {
		return errors.New("calendar.Store: source ID required")
	}
	if cal.CreatedAt == 0 {
		cal.CreatedAt = time.Now().Unix()
	}
	_, err := tx.Exec(`
		INSERT INTO calendars (
			id, source_id, url, display_name, description, color,
			visible, ctag, last_synced_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cal.ID, cal.SourceID, cal.URL, cal.DisplayName, nullIfEmpty(cal.Description),
		nullIfEmpty(cal.Color), boolToInt(cal.Visible), nullIfEmpty(cal.Ctag),
		nullIfZero(cal.LastSyncedAt), cal.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert calendar: %w", err)
	}
	return nil
}

// ListCalendars returns all calendar rows for a source, ordered by display
// name for stable UI rendering.
func (s *Store) ListCalendars(sourceID string) ([]Calendar, error) {
	rows, err := s.DB().Query(`
		SELECT id, source_id, url, display_name,
		       COALESCE(description, ''), COALESCE(color, ''),
		       visible, COALESCE(ctag, ''), COALESCE(last_synced_at, 0), created_at
		FROM calendars WHERE source_id = ? ORDER BY display_name ASC`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("query calendars: %w", err)
	}
	defer rows.Close()

	var out []Calendar
	for rows.Next() {
		var cal Calendar
		var visible int
		if err := rows.Scan(
			&cal.ID, &cal.SourceID, &cal.URL, &cal.DisplayName,
			&cal.Description, &cal.Color,
			&visible, &cal.Ctag, &cal.LastSyncedAt, &cal.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan calendar: %w", err)
		}
		cal.Visible = visible != 0
		out = append(out, cal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate calendars: %w", err)
	}
	return out, nil
}

// nullIfEmpty returns the input wrapped in sql.NullString — empty strings
// become SQL NULL so COALESCE works on read.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullIfZero returns SQL NULL for zero int64 inputs.
func nullIfZero(n int64) any {
	if n == 0 {
		return nil
	}
	return n
}

// boolToInt converts a Go bool to SQLite's 0/1 integer convention.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
