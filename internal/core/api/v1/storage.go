package v1

// KVStore is a small key-value namespace scoped to one extension. Used for
// tiny config (e.g., "lastSyncToken", "preferredView") that doesn't warrant
// SQL tables. Each extension's main SQLite file is opened separately by its
// own internal/<ext>/store.go.
type KVStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
	List(prefix string) ([]string, error)
}

// Secrets is a per-extension secure key-value store for sensitive values
// (passwords, OAuth refresh tokens, etc.). Implementation is keyring-first
// with automatic fallback to an encrypted core table when the OS keyring is
// unavailable. Extensions never see which storage path was used.
//
// Extensions consume this via Storage.Secrets(extensionID). The handle is
// scoped to one extension — keys live in a per-extension namespace, so two
// extensions can use the same key string without colliding.
type Secrets interface {
	// Set stores value under key. Empty value is treated as Delete.
	Set(key, value string) error

	// Get returns the stored value, or "" if no entry exists. The empty
	// string is the "not found" signal — callers distinguish from errors
	// by checking the returned string after a nil error.
	Get(key string) (string, error)

	// Delete removes the entry. Idempotent — deleting a non-existent
	// entry is not an error.
	Delete(key string) error

	// DeleteAll removes every entry for this extension. For uninstall
	// cleanup. Best-effort on individual entry failures; returns an
	// error only on bulk failures (e.g., DB unreachable).
	DeleteAll() error
}

// Storage provides per-extension data services. KV is small string config
// (non-sensitive). Secrets is sensitive credential storage (keyring + AES
// fallback transparently). Per-extension SQLite is implicit (each extension
// opens its own DB via internal/extensions.OpenStore for richer data).
type Storage interface {
	KV(extensionID string) KVStore
	Secrets(extensionID string) Secrets
}
