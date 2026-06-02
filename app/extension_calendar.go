package app

import (
	"github.com/hkdb/aerion/extensions/calendar"
	extcalendarbe "github.com/hkdb/aerion/extensions/calendar/backend"
	"github.com/hkdb/aerion/internal/oauth2"
)

// initCalendarExtension wires the Calendar extension's Bridge into App
// during Startup. All bridge logic lives in extensions/calendar/backend/;
// this file exists ONLY so the host can supply the bridge with its
// host-provided dependencies (settings store, paths, db, coreapi handle)
// and so the embedded-field promotion makes the bridge methods Wails-bindable.
//
// Lightweight-by-default invariant: the Bridge struct allocation is the
// entire footprint until the first enabled `Calendar_*` Wails call. At
// that point, `CalendarBridge.ensureInit()` opens the per-extension
// SQLite, applies pending migrations, and constructs the `API`. Disabled
// extensions contribute zero work.
//
// Per docs/EXT_RULES.md R2, this file holds NO closures wrapping
// `internal/*` calls. The calendar extension's only host touchpoints are
// the standard `coreapi.Core` surfaces (Storage.Secrets for the CalDAV
// password, UI for the rail tab + settings tab).
func (a *App) initCalendarExtension() {
	calendarCore := newCoreForExtension(a, a.calendarExt)

	a.CalendarBridge = extcalendarbe.NewCalendarBridge(extcalendarbe.CalendarBridgeDeps{
		SettingsStore: a.settingsStore,
		Paths:         a.paths,
		DB:            a.db,
		Core:          calendarCore,
	})

	// Register the extension's declared OAuth client configs with the global
	// resolver. Phase 1A: both slots have empty client IDs (unless ldflag-
	// injected at build time); the resolver returns (zero, false) for empty
	// slots and the chain falls through normally. Phase 2 wires Google /
	// Microsoft providers behind these slots.
	oauth2.RegisterCredentialsProvider(extensionOAuthProvider(calendar.OAuthClients()))
}
