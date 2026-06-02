package backend

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-webdav"
	extcaldav "github.com/emersion/go-webdav/caldav"
)

// DiscoveredCalendar is the calendar-shaped result of CalDAV discovery,
// suitable for persisting as a row in the `calendars` table.
type DiscoveredCalendar struct {
	Path        string
	DisplayName string
	Description string
}

// DiscoverCalendars probes a user-entered CalDAV server URL with the
// provided basic-auth credentials and returns the resolved calendar-home-set
// path plus the list of calendars found.
//
// Mirrors `internal/carddav/client.go::DiscoverAddressbooks` exactly:
//   1. URL as-is (the user may have entered a full home-set path).
//   2. `<scheme>://<host>/.well-known/caldav` (RFC 6764).
//   3. Nextcloud paths (`/remote.php/dav`, `/remote.php/caldav`).
//   4. Common paths (`/dav`, `/caldav`, `/principals/<username>`).
//
// Each attempt does `FindCurrentUserPrincipal` → `FindCalendarHomeSet` →
// `FindCalendars`. The first attempt that yields >0 calendars wins. On
// total failure the last seen error is returned (most-specific reason).
//
// The XML-fix transport that the CardDAV client carries to work around
// quirky servers (Purelymail's RFC 1123Z dates, mailbox.org's unquoted
// ETags) is intentionally OMITTED here. Discovery PROPFIND responses don't
// hit the affected XML elements — 1C will factor / inline the fix when
// ETag-based sync needs it.
func DiscoverCalendars(ctx context.Context, baseURL, username, password string) (string, []DiscoveredCalendar, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	httpClient := webdav.HTTPClientWithBasicAuth(
		newCalDAVHTTPClient(30*time.Second),
		username, password,
	)

	var lastErr error

	// Attempt 1: URL as-is.
	homePath, calendars, err := tryDiscoverCalDAVFromURL(ctx, httpClient, parsedURL.String())
	if err == nil && len(calendars) > 0 {
		return homePath, calendars, nil
	}
	if err != nil {
		lastErr = err
	}

	// Attempt 2: .well-known/caldav.
	wellKnownURL := fmt.Sprintf("%s://%s/.well-known/caldav", parsedURL.Scheme, parsedURL.Host)
	homePath, calendars, err = tryDiscoverCalDAVFromURL(ctx, httpClient, wellKnownURL)
	if err == nil && len(calendars) > 0 {
		return homePath, calendars, nil
	}
	if err != nil {
		lastErr = err
	}

	// Attempt 3 + 4: common paths.
	commonPaths := []string{
		"/remote.php/dav",                                          // Nextcloud / ownCloud
		"/remote.php/caldav",                                       // older Nextcloud
		fmt.Sprintf("/remote.php/dav/calendars/%s/", username),     // Nextcloud direct calendar home
		"/dav",
		"/caldav",
		"/principals/" + username,
	}
	for _, path := range commonPaths {
		tryURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, path)
		homePath, calendars, err = tryDiscoverCalDAVFromURL(ctx, httpClient, tryURL)
		if err == nil && len(calendars) > 0 {
			return homePath, calendars, nil
		}
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return "", nil, fmt.Errorf("no calendars found at %s: %w", baseURL, lastErr)
	}
	return "", nil, fmt.Errorf("no calendars found at %s", baseURL)
}

// tryDiscoverCalDAVFromURL runs the per-attempt discovery sequence against a
// single candidate URL. Order: FindCurrentUserPrincipal → FindCalendarHomeSet
// → FindCalendars. If the principal lookup fails, we treat the URL itself as
// the home-set and try to list calendars directly.
func tryDiscoverCalDAVFromURL(ctx context.Context, httpClient webdav.HTTPClient, urlStr string) (string, []DiscoveredCalendar, error) {
	client, err := extcaldav.NewClient(httpClient, urlStr)
	if err != nil {
		return "", nil, fmt.Errorf("new caldav client: %w", err)
	}

	principal, err := client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		// Principal lookup unsupported / unauthorized — try the URL itself as
		// the calendar home set. Some servers (or pre-supplied home-set URLs)
		// work that way.
		cals, lerr := tryListCalendarsAt(ctx, httpClient, urlStr)
		if lerr != nil {
			return "", nil, fmt.Errorf("find principal: %w", err)
		}
		return urlStr, cals, nil
	}

	homeSet, err := client.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return "", nil, fmt.Errorf("find calendar home set: %w", err)
	}

	resolvedHome := resolveCalDAVURL(urlStr, homeSet)
	cals, lerr := tryListCalendarsAt(ctx, httpClient, resolvedHome)
	if lerr != nil {
		return "", nil, fmt.Errorf("list calendars: %w", lerr)
	}
	return resolvedHome, cals, nil
}

// tryListCalendarsAt issues FindCalendars against the given URL. Filters out
// non-VEVENT collections (some servers expose VTODO-only or VJOURNAL
// collections in the same home-set) so the calendar UI doesn't render
// surprise rows.
func tryListCalendarsAt(ctx context.Context, httpClient webdav.HTTPClient, urlStr string) ([]DiscoveredCalendar, error) {
	client, err := extcaldav.NewClient(httpClient, urlStr)
	if err != nil {
		return nil, fmt.Errorf("new caldav client: %w", err)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	cals, err := client.FindCalendars(ctx, parsedURL.Path)
	if err != nil {
		return nil, err
	}

	out := make([]DiscoveredCalendar, 0, len(cals))
	for _, cal := range cals {
		if !supportsVEVENT(cal.SupportedComponentSet) {
			continue
		}
		name := cal.Name
		if name == "" {
			// Fall back to the last path segment.
			parts := strings.Split(strings.Trim(cal.Path, "/"), "/")
			if len(parts) > 0 {
				name = parts[len(parts)-1]
			}
		}
		out = append(out, DiscoveredCalendar{
			Path:        cal.Path,
			DisplayName: name,
			Description: cal.Description,
		})
	}
	return out, nil
}

// supportsVEVENT returns true when the calendar's
// supported-calendar-component-set explicitly includes VEVENT, OR when the
// set is empty (per RFC 4791, an empty set means "all components" by some
// server interpretations; we accept rather than skip).
func supportsVEVENT(set []string) bool {
	if len(set) == 0 {
		return true
	}
	for _, c := range set {
		if strings.EqualFold(c, "VEVENT") {
			return true
		}
	}
	return false
}

// resolveCalDAVURL resolves a (possibly relative) href against a base URL.
// Mirrors `internal/carddav/client.go::resolveURL`.
func resolveCalDAVURL(baseURL, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

// newCalDAVHTTPClient returns the plain *http.Client used for discovery.
// Mirrors `internal/carddav/client.go::newHTTPClient` MINUS the
// xmlFixTransport. The XML-fix is needed for the ETag / lastmodified parsing
// quirks in some servers; discovery PROPFIND doesn't touch those fields.
// If 1C's sync layer hits the same compat issues, the transport gets factored
// at that point (likely as a shared internal/davutil package the host
// exposes, since calendar can't import internal/carddav).
func newCalDAVHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}
