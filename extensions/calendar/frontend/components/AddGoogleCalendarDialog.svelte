<script lang="ts">
  // AddGoogleCalendarDialog — minimal three-step picker for attaching a Google
  // calendar source to an existing Aerion mail account. Phase 2 Chunk 3.
  //
  // Flow:
  //   1. Account select: dropdown of Aerion accounts filtered to Gmail.
  //   2. Calendar fetch: Calendar_ListGoogleCalendarsForAccount(accountID). If
  //      the broker returns "additional consent required", surface a banner
  //      asking the user to grant calendar access (Chunk 6 will hook this to
  //      a proper consent flow; today we surface the error and the user
  //      retries from account settings).
  //   3. Confirm: checkbox list of writable calendars + name field. Calls
  //      Calendar_AddGoogleSource(accountID, name, selections).
  //
  // This is intentionally low-polish — Chunk 6 ships the production
  // AccountCalendarHookPanelGoogle (post-Gmail-add-flow integration).

  import { _ } from 'svelte-i18n'
  import * as Dialog from '$lib/components/ui/dialog'
  import { Button } from '$lib/components/ui/button'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'
  import { toasts } from '$lib/stores/toast'
  import { dialogGuardOpen, dialogGuardClose } from '$lib/stores/dialogGuard'
  // @ts-ignore - wailsjs bindings
  import { GetAccounts, Calendar_ListGoogleCalendarsForAccount, Calendar_AddGoogleSource } from '$wailsjs/go/app/App.js'
  // @ts-ignore - wailsjs bindings
  import type { account, backend } from '$wailsjs/go/models'

  interface Props {
    open: boolean
    onClose?: () => void
  }

  let { open = $bindable(false), onClose }: Props = $props()

  let accounts = $state<account.Account[]>([])
  let selectedAccountId = $state('')
  let calendars = $state<backend.GoogleCalendarChoice[]>([])
  let selectedIds = $state<Set<string>>(new Set())
  let sourceName = $state('')
  let loadingAccounts = $state(false)
  let loadingCalendars = $state(false)
  let submitting = $state(false)
  let errorMessage = $state('')
  let needsConsent = $state(false)

  const googleAccounts = $derived(
    accounts.filter((a) => (a.imapHost || '').includes('gmail.com') || (a.imapHost || '').includes('googlemail.com')),
  )

  $effect(() => {
    if (!open) return
    dialogGuardOpen()
    return () => dialogGuardClose()
  })

  $effect(() => {
    if (!open) return
    resetState()
    void loadAccounts()
  })

  function resetState() {
    selectedAccountId = ''
    calendars = []
    selectedIds = new Set()
    sourceName = ''
    errorMessage = ''
    needsConsent = false
  }

  async function loadAccounts() {
    loadingAccounts = true
    try {
      const list = await GetAccounts()
      accounts = list || []
      if (googleAccounts.length === 1) {
        selectedAccountId = googleAccounts[0].id
        void onAccountChange()
      }
    } catch (err) {
      errorMessage = String(err)
    } finally {
      loadingAccounts = false
    }
  }

  async function onAccountChange() {
    if (!selectedAccountId) return
    errorMessage = ''
    needsConsent = false
    calendars = []
    selectedIds = new Set()
    sourceName = ''
    loadingCalendars = true
    try {
      const list = await Calendar_ListGoogleCalendarsForAccount(selectedAccountId)
      calendars = list || []
      const acct = accounts.find((a) => a.id === selectedAccountId)
      sourceName = acct?.email ? `Google: ${acct.email}` : 'Google Calendar'
    } catch (err) {
      const msg = String(err)
      needsConsent = msg.includes('additional consent required')
      if (!needsConsent) {
        errorMessage = msg
      }
    } finally {
      loadingCalendars = false
    }
  }

  function toggleCalendar(id: string) {
    const next = new Set(selectedIds)
    if (next.has(id)) {
      next.delete(id)
      selectedIds = next
      return
    }
    next.add(id)
    selectedIds = next
  }

  async function handleSave() {
    if (submitting) return
    if (!selectedAccountId) {
      errorMessage = $_('calendar.settings.addGoogleErrorAccount')
      return
    }
    if (selectedIds.size === 0) {
      errorMessage = $_('calendar.settings.addGoogleErrorPick')
      return
    }
    if (!sourceName.trim()) {
      errorMessage = $_('calendar.settings.addGoogleErrorName')
      return
    }
    submitting = true
    errorMessage = ''
    try {
      const selections = calendars
        .filter((c) => selectedIds.has(c.id))
        .map((c) => ({ id: c.id, displayName: c.summary, color: '' }))
      await Calendar_AddGoogleSource(selectedAccountId, sourceName.trim(), selections)
      toasts.success($_('calendar.settings.addGoogleToastSuccess'))
      close()
    } catch (err) {
      errorMessage = $_('calendar.settings.addGoogleToastError', { values: { message: String(err) } })
    } finally {
      submitting = false
    }
  }

  function close() {
    open = false
    onClose?.()
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title>{$_('calendar.settings.addGoogleTitle')}</Dialog.Title>
      <Dialog.Description>{$_('calendar.settings.addGoogleDescription')}</Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4 py-2">
      <!-- Account picker -->
      <div class="space-y-1">
        <Label>{$_('calendar.settings.addGoogleAccountLabel')}</Label>
        {#if loadingAccounts}
          <p class="text-xs text-muted-foreground">{$_('calendar.common.loading')}</p>
        {/if}
        {#if !loadingAccounts && googleAccounts.length === 0}
          <p class="text-xs text-muted-foreground">{$_('calendar.settings.addGoogleNoAccounts')}</p>
        {/if}
        {#if !loadingAccounts && googleAccounts.length > 0}
          <select
            class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            bind:value={selectedAccountId}
            onchange={() => void onAccountChange()}
          >
            <option value="">{$_('calendar.settings.addGooglePickAccount')}</option>
            {#each googleAccounts as acct (acct.id)}
              <option value={acct.id}>{acct.email}</option>
            {/each}
          </select>
        {/if}
      </div>

      {#if needsConsent}
        <div class="rounded-md border border-yellow-400/40 bg-yellow-400/10 p-3 text-xs text-yellow-700 dark:text-yellow-300">
          {$_('calendar.settings.addGoogleConsentNeeded')}
        </div>
      {/if}

      <!-- Calendar list -->
      {#if loadingCalendars}
        <p class="text-xs text-muted-foreground">{$_('calendar.settings.addGoogleLoadingCalendars')}</p>
      {/if}

      {#if !loadingCalendars && calendars.length > 0}
        <div class="space-y-1">
          <Label>{$_('calendar.settings.addGoogleCalendarsLabel')}</Label>
          <div class="max-h-48 overflow-y-auto rounded-md border border-border">
            {#each calendars as cal (cal.id)}
              <label
                class="flex items-center gap-2 px-3 py-2 text-sm border-b border-border last:border-b-0
                       hover:bg-muted/40 cursor-pointer {!cal.writable ? 'opacity-60' : ''}"
                title={!cal.writable ? $_('calendar.settings.addGoogleReadOnly') : ''}
              >
                <input
                  type="checkbox"
                  checked={selectedIds.has(cal.id)}
                  disabled={!cal.writable}
                  onchange={() => toggleCalendar(cal.id)}
                />
                <span class="truncate flex-1">{cal.summary}</span>
                {#if cal.primary}
                  <span class="text-xs text-muted-foreground">{$_('calendar.settings.addGooglePrimary')}</span>
                {/if}
                {#if !cal.writable}
                  <span class="text-xs text-muted-foreground">{$_('calendar.settings.addGoogleReadOnlyBadge')}</span>
                {/if}
              </label>
            {/each}
          </div>
        </div>

        <!-- Source name -->
        <div class="space-y-1">
          <Label for="source-name">{$_('calendar.settings.addGoogleSourceNameLabel')}</Label>
          <Input id="source-name" bind:value={sourceName} />
        </div>
      {/if}

      {#if errorMessage}
        <p class="text-xs text-destructive">{errorMessage}</p>
      {/if}
    </div>

    <Dialog.Footer>
      <Button variant="ghost" onclick={close}>{$_('calendar.common.cancel')}</Button>
      <Button onclick={handleSave} disabled={submitting || calendars.length === 0 || selectedIds.size === 0}>
        {submitting ? $_('calendar.common.saving') : $_('calendar.common.save')}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
