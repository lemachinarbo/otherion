/**
 * Keyboard Shortcuts Store
 *
 * Manages shortcut definitions, custom keybindings, and persistent storage.
 */

export type ShortcutCategory = 'pane' | 'navigation' | 'composer' | 'action'

export interface ShortcutDefinition {
  id: string
  category: ShortcutCategory
  nameKey: string
  descriptionKey: string
  defaultKeys: string[]
}

export const SHORTCUT_DEFINITIONS: ShortcutDefinition[] = [
  // Pane Switching & Focus
  {
    id: 'PANE_FOCUS_NEXT',
    category: 'pane',
    nameKey: 'shortcuts.paneFocusNext',
    descriptionKey: 'shortcuts.paneFocusNextDesc',
    defaultKeys: ['Alt+L', 'Alt+Right'],
  },
  {
    id: 'PANE_FOCUS_PREV',
    category: 'pane',
    nameKey: 'shortcuts.paneFocusPrev',
    descriptionKey: 'shortcuts.paneFocusPrevDesc',
    defaultKeys: ['Alt+H', 'Alt+Left'],
  },
  {
    id: 'SIDEBAR_NEXT',
    category: 'pane',
    nameKey: 'shortcuts.sidebarNext',
    descriptionKey: 'shortcuts.sidebarNextDesc',
    defaultKeys: ['Alt+J', 'Alt+Down'],
  },
  {
    id: 'SIDEBAR_PREV',
    category: 'pane',
    nameKey: 'shortcuts.sidebarPrev',
    descriptionKey: 'shortcuts.sidebarPrevDesc',
    defaultKeys: ['Alt+K', 'Alt+Up'],
  },

  // List & Message Navigation
  {
    id: 'LIST_NEXT',
    category: 'navigation',
    nameKey: 'shortcuts.listNext',
    descriptionKey: 'shortcuts.listNextDesc',
    defaultKeys: ['J', 'Down'],
  },
  {
    id: 'LIST_PREV',
    category: 'navigation',
    nameKey: 'shortcuts.listPrev',
    descriptionKey: 'shortcuts.listPrevDesc',
    defaultKeys: ['K', 'Up'],
  },
  {
    id: 'LIST_NEXT_CHECK',
    category: 'navigation',
    nameKey: 'shortcuts.listNextCheck',
    descriptionKey: 'shortcuts.listNextCheckDesc',
    defaultKeys: ['Shift+J', 'Shift+Down'],
  },
  {
    id: 'LIST_PREV_CHECK',
    category: 'navigation',
    nameKey: 'shortcuts.listPrevCheck',
    descriptionKey: 'shortcuts.listPrevCheckDesc',
    defaultKeys: ['Shift+K', 'Shift+Up'],
  },
  {
    id: 'LIST_TOGGLE_CHECK',
    category: 'navigation',
    nameKey: 'shortcuts.listToggleCheck',
    descriptionKey: 'shortcuts.listToggleCheckDesc',
    defaultKeys: ['Space'],
  },
  {
    id: 'LIST_OPEN',
    category: 'navigation',
    nameKey: 'shortcuts.listOpen',
    descriptionKey: 'shortcuts.listOpenDesc',
    defaultKeys: ['Enter'],
  },
  {
    id: 'LIST_SELECT_ALL',
    category: 'navigation',
    nameKey: 'shortcuts.listSelectAll',
    descriptionKey: 'shortcuts.listSelectAllDesc',
    defaultKeys: ['Ctrl+A'],
  },
  {
    id: 'LIST_DELETE',
    category: 'navigation',
    nameKey: 'shortcuts.listDelete',
    descriptionKey: 'shortcuts.listDeleteDesc',
    defaultKeys: ['Delete', 'Backspace'],
  },
  {
    id: 'SEARCH_MESSAGES',
    category: 'navigation',
    nameKey: 'shortcuts.searchMessages',
    descriptionKey: 'shortcuts.searchMessagesDesc',
    defaultKeys: ['Ctrl+S'],
  },

  // Composer & Mail Actions
  {
    id: 'COMPOSE_NEW',
    category: 'composer',
    nameKey: 'shortcuts.composeNew',
    descriptionKey: 'shortcuts.composeNewDesc',
    defaultKeys: ['Ctrl+N'],
  },
  {
    id: 'REPLY',
    category: 'composer',
    nameKey: 'shortcuts.reply',
    descriptionKey: 'shortcuts.replyDesc',
    defaultKeys: ['Ctrl+R'],
  },
  {
    id: 'REPLY_ALL',
    category: 'composer',
    nameKey: 'shortcuts.replyAll',
    descriptionKey: 'shortcuts.replyAllDesc',
    defaultKeys: ['Ctrl+Shift+R'],
  },
  {
    id: 'FORWARD',
    category: 'composer',
    nameKey: 'shortcuts.forward',
    descriptionKey: 'shortcuts.forwardDesc',
    defaultKeys: ['Ctrl+F'],
  },

  // Actions
  {
    id: 'ARCHIVE',
    category: 'action',
    nameKey: 'shortcuts.archive',
    descriptionKey: 'shortcuts.archiveDesc',
    defaultKeys: ['Ctrl+K'],
  },
  {
    id: 'SPAM',
    category: 'action',
    nameKey: 'shortcuts.spam',
    descriptionKey: 'shortcuts.spamDesc',
    defaultKeys: ['Ctrl+J'],
  },
  {
    id: 'TOGGLE_READ',
    category: 'action',
    nameKey: 'shortcuts.toggleRead',
    descriptionKey: 'shortcuts.toggleReadDesc',
    defaultKeys: ['Ctrl+U'],
  },
  {
    id: 'SYNC_FOLDER',
    category: 'action',
    nameKey: 'shortcuts.syncFolder',
    descriptionKey: 'shortcuts.syncFolderDesc',
    defaultKeys: ['Ctrl+Shift+S'],
  },
  {
    id: 'SYNC_ALL',
    category: 'action',
    nameKey: 'shortcuts.syncAll',
    descriptionKey: 'shortcuts.syncAllDesc',
    defaultKeys: ['Ctrl+Shift+A'],
  },
]

const STORAGE_KEY = 'aerion_custom_shortcuts'

// Persistent reactive state using Svelte 5 runes
let customShortcuts = $state<Record<string, string[]>>(loadSavedShortcuts())

function loadSavedShortcuts(): Record<string, string[]> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      return JSON.parse(raw)
    }
  } catch (err) {
    console.error('Failed to load custom shortcuts:', err)
  }
  return {}
}

function saveShortcuts(map: Record<string, string[]>) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(map))
  } catch (err) {
    console.error('Failed to save custom shortcuts:', err)
  }
}

export function getShortcutKeys(id: string): string[] {
  if (customShortcuts[id] && customShortcuts[id].length > 0) {
    return customShortcuts[id]
  }
  const def = SHORTCUT_DEFINITIONS.find(s => s.id === id)
  return def ? def.defaultKeys : []
}

export function updateShortcutKeys(id: string, keys: string[]) {
  const next = { ...customShortcuts, [id]: keys }
  customShortcuts = next
  saveShortcuts(next)
}

export function resetAllShortcuts() {
  customShortcuts = {}
  saveShortcuts({})
}

export function isShortcutCustomized(id: string): boolean {
  return !!(customShortcuts[id] && customShortcuts[id].length > 0)
}

/**
 * Match a KeyboardEvent against a combo string (e.g. "Alt+L", "Ctrl+Shift+R", "J", "Space", "Down")
 */
export function matchesKeyCombo(combo: string, e: KeyboardEvent): boolean {
  const parts = combo.split('+')

  const reqCtrl = parts.includes('Ctrl') || parts.includes('Cmd')
  const reqAlt = parts.includes('Alt')
  const reqShift = parts.includes('Shift')

  const hasCtrl = e.ctrlKey || e.metaKey
  const hasAlt = e.altKey
  const hasShift = e.shiftKey

  if (reqCtrl !== hasCtrl) return false
  if (reqAlt !== hasAlt) return false
  if (reqShift !== hasShift) return false

  const keyPart = parts[parts.length - 1].toLowerCase()
  const eventKey = e.key.toLowerCase()

  if (keyPart === 'down' || keyPart === 'arrowdown') {
    return e.key === 'ArrowDown' || eventKey === 'down'
  }
  if (keyPart === 'up' || keyPart === 'arrowup') {
    return e.key === 'ArrowUp' || eventKey === 'up'
  }
  if (keyPart === 'left' || keyPart === 'arrowleft') {
    return e.key === 'ArrowLeft' || eventKey === 'left'
  }
  if (keyPart === 'right' || keyPart === 'arrowright') {
    return e.key === 'ArrowRight' || eventKey === 'right'
  }
  if (keyPart === 'space') {
    return e.key === ' ' || eventKey === 'space'
  }
  if (keyPart === 'enter') {
    return e.key === 'Enter'
  }
  if (keyPart === 'delete') {
    return e.key === 'Delete'
  }
  if (keyPart === 'backspace') {
    return e.key === 'Backspace'
  }

  return eventKey === keyPart
}

/**
 * Convert a KeyboardEvent to a standard combo string (e.g. "Ctrl+Shift+R")
 */
export function eventToKeyCombo(e: KeyboardEvent): string | null {
  const key = e.key
  if (['Control', 'Alt', 'Shift', 'Meta'].includes(key)) {
    return null
  }

  const parts: string[] = []
  if (e.ctrlKey || e.metaKey) parts.push('Ctrl')
  if (e.altKey) parts.push('Alt')
  if (e.shiftKey) parts.push('Shift')

  let mainKey = key
  if (key === 'ArrowDown') mainKey = 'Down'
  else if (key === 'ArrowUp') mainKey = 'Up'
  else if (key === 'ArrowLeft') mainKey = 'Left'
  else if (key === 'ArrowRight') mainKey = 'Right'
  else if (key === ' ') mainKey = 'Space'
  else if (key.length === 1) mainKey = key.toUpperCase()

  parts.push(mainKey)
  return parts.join('+')
}
