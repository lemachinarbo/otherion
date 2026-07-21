<script lang="ts">
  import Icon from '@iconify/svelte'
  import { Button } from '$lib/components/ui/button'
  import { Input } from '$lib/components/ui/input'
  import { _ } from '$lib/i18n'
  import { addToast } from '$lib/stores/toast'
  import {
    SHORTCUT_DEFINITIONS,
    getShortcutKeys,
    updateShortcutKeys,
    resetAllShortcuts,
    isShortcutCustomized,
    eventToKeyCombo,
    type ShortcutCategory,
  } from '$lib/stores/shortcuts.svelte'

  let searchQuery = $state('')
  let recordingId = $state<string | null>(null)

  const categories: { id: ShortcutCategory; label: string; icon: string }[] = [
    { id: 'pane', label: $_('shortcuts.categories.pane') || 'Pane Switching & Focus', icon: 'lucide:layout-grid' },
    { id: 'navigation', label: $_('shortcuts.categories.navigation') || 'List & Message Navigation', icon: 'lucide:arrow-down-up' },
    { id: 'composer', label: $_('shortcuts.categories.composer') || 'Composer & Email Actions', icon: 'lucide:square-pen' },
    { id: 'action', label: $_('shortcuts.categories.action') || 'Sync & System Actions', icon: 'lucide:zap' },
  ]

  let filteredDefinitions = $derived(
    SHORTCUT_DEFINITIONS.filter(def => {
      if (!searchQuery.trim()) return true
      const q = searchQuery.toLowerCase()
      const name = ($_ (def.nameKey) || def.id).toLowerCase()
      const desc = ($_ (def.descriptionKey) || '').toLowerCase()
      const keys = getShortcutKeys(def.id).join(' ').toLowerCase()
      return name.includes(q) || desc.includes(q) || keys.includes(q)
    })
  )

  function startRecording(id: string) {
    recordingId = id
  }

  function stopRecording() {
    recordingId = null
  }

  // Window-level capture during key recording: intercepts any key combo & stops Escape from closing dialog
  $effect(() => {
    if (!recordingId) return

    const handleGlobalKeyCapture = (e: KeyboardEvent) => {
      e.preventDefault()
      e.stopPropagation()
      e.stopImmediatePropagation()

      if (e.key === 'Escape') {
        stopRecording()
        return
      }

      const combo = eventToKeyCombo(e)
      if (combo) {
        updateShortcutKeys(recordingId!, [combo])
        stopRecording()
        addToast({
          type: 'success',
          message: $_('toast.settingsSaved') || 'Shortcut updated',
        })
      }
    }

    window.addEventListener('keydown', handleGlobalKeyCapture, { capture: true })
    return () => {
      window.removeEventListener('keydown', handleGlobalKeyCapture, { capture: true })
    }
  })

  function handleResetOne(id: string) {
    const def = SHORTCUT_DEFINITIONS.find(s => s.id === id)
    if (def) {
      updateShortcutKeys(id, def.defaultKeys)
      addToast({
        type: 'info',
        message: 'Shortcut reset to default',
      })
    }
  }

  function handleResetAll() {
    resetAllShortcuts()
    addToast({
      type: 'info',
      message: 'All shortcuts reset to defaults',
    })
  }
</script>

<div class="space-y-6 pb-2">
  <!-- Search & Reset Header -->
  <div class="flex items-center justify-between gap-4">
    <div class="relative flex-1">
      <Icon icon="lucide:search" class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />
      <Input
        type="text"
        placeholder={$_('common.search') + ' shortcuts...'}
        bind:value={searchQuery}
        class="pl-9 text-xs h-9"
      />
    </div>
    <Button variant="outline" size="sm" onclick={handleResetAll} class="text-xs h-9 gap-1.5 shrink-0">
      <Icon icon="lucide:rotate-ccw" class="w-3.5 h-3.5" />
      {$_('shortcuts.resetDefaults') || 'Reset Defaults'}
    </Button>
  </div>

  <!-- Shortcut Categories -->
  {#each categories as cat (cat.id)}
    {@const items = filteredDefinitions.filter(d => d.category === cat.id)}
    {#if items.length > 0}
      <div class="space-y-3">
        <div class="flex items-center gap-2 border-b border-border pb-1.5">
          <Icon icon={cat.icon} class="w-4 h-4 text-primary" />
          <h3 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{cat.label}</h3>
          <span class="text-[10px] px-1.5 py-0.2 rounded-full bg-muted text-muted-foreground font-mono">{items.length}</span>
        </div>

        <div class="grid grid-cols-1 gap-2">
          {#each items as item (item.id)}
            {@const currentKeys = getShortcutKeys(item.id)}
            {@const isCustom = isShortcutCustomized(item.id)}
            {@const isRecording = recordingId === item.id}

            <div
              class="flex items-center justify-between p-2.5 rounded-lg border border-border bg-card/50 hover:bg-accent/40 transition-colors gap-3"
            >
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-foreground truncate">
                    {$_ (item.nameKey) !== item.nameKey ? $_ (item.nameKey) : item.id}
                  </span>
                  {#if isCustom}
                    <span class="text-[10px] px-1.5 py-0.5 rounded bg-primary/10 text-primary font-medium">Custom</span>
                  {/if}
                </div>
                <p class="text-xs text-muted-foreground truncate mt-0.5">
                  {$_ (item.descriptionKey) !== item.descriptionKey ? $_ (item.descriptionKey) : ''}
                </p>
              </div>

              <div class="flex items-center gap-2 shrink-0">
                {#if isRecording}
                  <button
                    type="button"
                    onclick={stopRecording}
                    class="flex items-center gap-2 px-3 py-1.5 rounded-md border-2 border-primary bg-primary/15 text-primary text-xs font-mono animate-pulse outline-none cursor-pointer"
                  >
                    <Icon icon="lucide:radio" class="w-3.5 h-3.5 animate-spin" />
                    <span>{$_('shortcuts.pressKeyToRecord') || 'Press keys... (Esc to cancel)'}</span>
                  </button>
                {:else}
                  <div class="flex items-center gap-1">
                    {#each currentKeys as keyCombo (keyCombo)}
                      <kbd class="px-2 py-1 bg-muted/80 text-foreground border border-border/80 rounded text-xs font-mono shadow-xs font-semibold">
                        {keyCombo}
                      </kbd>
                    {/each}
                  </div>

                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-7 px-2 text-xs"
                    onclick={() => startRecording(item.id)}
                    title="Rebind shortcut"
                  >
                    <Icon icon="lucide:pencil" class="w-3.5 h-3.5" />
                  </Button>

                  {#if isCustom}
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-7 px-2 text-xs text-muted-foreground hover:text-foreground"
                      onclick={() => handleResetOne(item.id)}
                      title="Reset to default"
                    >
                      <Icon icon="lucide:undo-2" class="w-3.5 h-3.5" />
                    </Button>
                  {/if}
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/each}
</div>
