package backend

import (
	"errors"
	"sync"

	coreapi "github.com/hkdb/aerion/internal/core/api/v1"
	"github.com/hkdb/aerion/internal/database"
	"github.com/hkdb/aerion/internal/platform"
)

// CalendarBridge is the Wails-bindable surface for the Calendar extension.
// It's embedded into the host `*app.App` struct; Go's method-promotion
// makes every CalendarBridge method appear on App so Wails' reflection-based
// bind generator picks them up. All Calendar-specific logic lives here, not
// in the host. The host's `app/extension_calendar.go` is reduced to a
// dozen lines of construction wiring.
//
// Method naming: all Wails-bound bridge methods use the `Calendar_` prefix
// so they can't collide with another extension's methods after embedding
// into the same App. See docs/EXT_RULES.md R19.
//
// Lightweight-by-default invariant: when the user has the Calendar
// extension disabled, NOTHING is loaded beyond the ~80-byte CalendarBridge
// struct itself. The per-extension SQLite is opened eagerly at Startup
// (schema-validity invariant), but the `API` wrapper that holds the
// caldav client + secrets handle is lazy-init via sync.Once inside
// `ensureInit`. The first enabled method call triggers init; subsequent
// calls are fast. See docs/EXT_RULES.md §4.
type CalendarBridge struct {
	deps CalendarBridgeDeps

	// Lazy-initialized API. Constructed on first enabled bridge call so
	// disabled extensions contribute zero work.
	initOnce sync.Once
	initErr  error
	api      *API
}

// CalendarBridgeDeps bundles the host-provided dependencies the bridge needs.
// Grouped into a struct so adding a new dep doesn't churn every call site
// in the host. Per docs/EXT_RULES.md R2, this struct holds NO closures
// wrapping `internal/*` calls — anything the extension needs from the host
// goes through `coreapi.Core` directly.
type CalendarBridgeDeps struct {
	// SettingsStore is consulted on every bridge call for the enabled
	// gate (lightweight invariant — disabled calls short-circuit before
	// any work).
	SettingsStore SettingsStore

	// Paths gives the bridge access to the OS-appropriate data directory
	// for opening the extension's per-extension SQLite.
	Paths *platform.Paths

	// DB is the shared application database. Not used by the calendar
	// extension's primary data paths (calendar data lives in its own
	// per-extension SQLite, opened via Paths). Kept here for symmetry
	// with Contacts and forward-compat with Phase 2 cross-extension
	// queries that may need it.
	DB *database.DB

	// Core is the coreapi.Core handle. The bridge uses it to reach the
	// host-implemented surfaces — currently `coreapi.Storage.Secrets`
	// for the CalDAV password storage. Per-extension scoped at Core
	// construction time in `newCoreForExtension`.
	Core coreapi.Core
}

// SettingsStore is the narrow interface the bridge needs from the host's
// settings store. Defined here (rather than importing the concrete type)
// so 3rd-party extensions can swap in their own implementation for tests
// and so this file doesn't grow a host-package dependency.
type SettingsStore interface {
	IsExtensionEnabled(id string) (bool, error)
}

// NewCalendarBridge constructs the bridge with its dependencies. Does NOT
// touch the DB or open any extension state — that's the Store's job
// (called eagerly from app/extension_calendar.go to keep schema valid
// across enable/disable cycles).
func NewCalendarBridge(deps CalendarBridgeDeps) *CalendarBridge {
	return &CalendarBridge{deps: deps}
}

// extensionID is the key the bridge looks up in settings for the
// enabled-state check, AND the scope passed to coreapi.Storage.Secrets.
// Kept as a const so a typo doesn't silently disable every bridge
// method or store secrets in the wrong namespace.
const extensionID = "calendar"

// gateEnabled returns true when the extension is currently enabled AND
// the host gave us a SettingsStore. Returns false (silently) when the
// store is nil or when the settings read errors out.
func (b *CalendarBridge) gateEnabled() bool {
	if b.deps.SettingsStore == nil {
		return false
	}
	enabled, err := b.deps.SettingsStore.IsExtensionEnabled(extensionID)
	if err != nil {
		return false
	}
	return enabled
}

// ensureInit lazily constructs the API on the first enabled bridge call.
// Reuses the Store the host opened eagerly at Startup (passed via Paths
// + opened by app/extension_calendar.go). Secrets handle is fetched
// from coreapi.Core, pre-scoped to this extension's ID.
func (b *CalendarBridge) ensureInit() error {
	b.initOnce.Do(func() {
		if b.deps.DB == nil || b.deps.Paths == nil {
			b.initErr = errors.New("calendar.CalendarBridge: missing DB or Paths in deps")
			return
		}
		if b.deps.Core == nil {
			b.initErr = errors.New("calendar.CalendarBridge: missing Core in deps")
			return
		}

		store, err := NewStore(b.deps.Paths.Data)
		if err != nil {
			b.initErr = err
			return
		}

		secrets := b.deps.Core.Storage().Secrets(extensionID)
		b.api = NewAPI(store, secrets)
	})
	return b.initErr
}

// --- Wails-bound surface (Calendar_*) ----------------------------------------
//
// All methods gate on gateEnabled() so disabled extensions short-circuit
// before any work. ensureInit runs once per process; subsequent calls are
// the cost of one sync.Once.Done() check.

// Calendar_AddCalDAVSource probes the user-entered server URL with the
// supplied credentials, persists the source + discovered calendars, and
// stores the password via coreapi.Storage.Secrets. Returns the new
// source's ID, or an error describing where discovery failed (auth /
// principal / home-set / list).
func (b *CalendarBridge) Calendar_AddCalDAVSource(name, url, username, password string) (string, error) {
	if !b.gateEnabled() {
		return "", errors.New("calendar: extension disabled")
	}
	if err := b.ensureInit(); err != nil {
		return "", err
	}
	return b.api.AddCalDAVSource(name, url, username, password)
}

// Calendar_ListSources returns all configured calendar sources. Returns
// nil (empty result) when the extension is disabled — consistent with
// Contacts_ListSources's behavior.
func (b *CalendarBridge) Calendar_ListSources() ([]Source, error) {
	if !b.gateEnabled() {
		return nil, nil
	}
	if err := b.ensureInit(); err != nil {
		return nil, err
	}
	return b.api.ListSources()
}

// Calendar_ListCalendars returns the calendars for a single source.
// Returns nil when the extension is disabled.
func (b *CalendarBridge) Calendar_ListCalendars(sourceID string) ([]Calendar, error) {
	if !b.gateEnabled() {
		return nil, nil
	}
	if err := b.ensureInit(); err != nil {
		return nil, err
	}
	return b.api.ListCalendars(sourceID)
}

// Calendar_DeleteSource removes a calendar source and all its associated
// data (calendars via CASCADE, stored password via coreapi.Secrets).
// Idempotent — deleting a non-existent source is not an error.
func (b *CalendarBridge) Calendar_DeleteSource(sourceID string) error {
	if !b.gateEnabled() {
		return nil
	}
	if err := b.ensureInit(); err != nil {
		return err
	}
	return b.api.DeleteSource(sourceID)
}
