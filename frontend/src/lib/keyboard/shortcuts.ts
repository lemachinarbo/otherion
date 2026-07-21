// Shared keyboard shortcut predicates — single source of truth for "what key
// combination matches what action." Consumed by both Aerion's mail UI handler
// (App.svelte) and by extension UI components in the kit (frontend/src/lib/
// components/kit/).

export function noMods(e: KeyboardEvent): boolean {
  return !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey
}

export function ctrlOrMeta(e: KeyboardEvent): boolean {
  return (e.ctrlKey || e.metaKey) && !e.altKey
}

export function altOnly(e: KeyboardEvent): boolean {
  return e.altKey && !e.ctrlKey && !e.metaKey
}

// List/pane navigation
export const LIST_NEXT = (e: KeyboardEvent): boolean =>
  (e.key === 'j' || e.key === 'ArrowDown') && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey

export const LIST_PREV = (e: KeyboardEvent): boolean =>
  (e.key === 'k' || e.key === 'ArrowUp') && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey

export const LIST_NEXT_CHECK = (e: KeyboardEvent): boolean =>
  e.shiftKey && !e.ctrlKey && !e.metaKey && !e.altKey &&
  (e.key === 'J' || e.key === 'j' || e.key === 'ArrowDown')

export const LIST_PREV_CHECK = (e: KeyboardEvent): boolean =>
  e.shiftKey && !e.ctrlKey && !e.metaKey && !e.altKey &&
  (e.key === 'K' || e.key === 'k' || e.key === 'ArrowUp')

export const LIST_TOGGLE_CHECK = (e: KeyboardEvent): boolean =>
  e.key === ' ' && noMods(e)

export const LIST_OPEN = (e: KeyboardEvent): boolean =>
  e.key === 'Enter' && noMods(e)

export const LIST_SELECT_ALL = (e: KeyboardEvent): boolean =>
  e.key.toLowerCase() === 'a' && ctrlOrMeta(e) && !e.shiftKey

export const LIST_DELETE = (e: KeyboardEvent): boolean =>
  (e.key === 'Delete' || e.key === 'Backspace') && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey

export const PANE_FOCUS_NEXT = (e: KeyboardEvent): boolean =>
  altOnly(e) && (e.key === 'l' || e.key === 'ArrowRight')

export const PANE_FOCUS_PREV = (e: KeyboardEvent): boolean =>
  altOnly(e) && (e.key === 'h' || e.key === 'ArrowLeft')

export const SIDEBAR_NEXT = (e: KeyboardEvent): boolean =>
  altOnly(e) && (e.key === 'j' || e.key === 'ArrowDown')

export const SIDEBAR_PREV = (e: KeyboardEvent): boolean =>
  altOnly(e) && (e.key === 'k' || e.key === 'ArrowUp')

export const KEY_ARCHIVE = (e: KeyboardEvent): boolean =>
  e.key.toLowerCase() === 'e' || (e.key.toLowerCase() === 'k' && (e.ctrlKey || e.metaKey))

export const KEY_SPAM = (e: KeyboardEvent): boolean =>
  e.key.toLowerCase() === 'j' && (e.ctrlKey || e.metaKey)

export const KEY = {
  LIST_NEXT,
  LIST_PREV,
  LIST_NEXT_CHECK,
  LIST_PREV_CHECK,
  LIST_TOGGLE_CHECK,
  LIST_OPEN,
  LIST_SELECT_ALL,
  LIST_DELETE,
  PANE_FOCUS_NEXT,
  PANE_FOCUS_PREV,
  SIDEBAR_NEXT,
  SIDEBAR_PREV,
  ARCHIVE: KEY_ARCHIVE,
  SPAM: KEY_SPAM,
}

export function matchesAny(e: KeyboardEvent, defs: Array<(e: KeyboardEvent) => boolean>): boolean {
  return defs.some(def => def(e))
}
