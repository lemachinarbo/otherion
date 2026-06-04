<script lang="ts">
  // AddMicrosoftCalendarDialog — minimal three-step picker for attaching an
  // Outlook / Microsoft 365 calendar source to an existing Aerion mail
  // account. Phase 2 Chunk 4 sibling of AddGoogleCalendarDialog.
  //
  // Flow mirrors the Google variant 1-for-1; the only differences are:
  //   - Filter accounts by imapHost containing "outlook" (catches both
  //     outlook.office365.com for M365 and imap-mail.outlook.com for
  //     consumer outlook.com).
  //   - Calls Calendar_ListMicrosoftCalendarsForAccount /
  //     Calendar_AddMicrosoftSource bridge methods.
  //   - "Read-only" filter looks at MicrosoftCalendarChoice.writable
  //     (derived from Graph's canEdit field).

  import { _ } from 'svelte-i18n'
  import * as Dialog from '$lib/components/ui/dialog'
  import { Button } from '$lib/components/ui/button'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'
  import { toasts } from '$lib/stores/toast'
  import { dialogGuardOpen, dialogGuardClose } from '$lib/stores/dialogGuard'
  // @ts-ignore - wailsjs bindings
  import { GetAccounts, Calendar_ListMicrosoftCalendarsForAccount, Calendar_AddMicrosoftSource } from '$wailsjs/go/app/App.js'
  // @ts-ignore - wailsjs bindings
  import type { account, backend } from '$wailsjs/go/models'

  interface Props {
    open: boolean
    onClose?: () => void
  }

  let { open = $bindable(false), onClose }: Props = $props()

  let accounts = $state<account.Account[]>([])
  let selectedAccountId = $state('')
  let calendars = $state<backend.MicrosoftCalendarChoice[]>([])
  let selectedIds = $state<Set<string>>(new Set())
  let sourceName = $state('')
  let loadingAccounts = $state(false)
  let loadingCalendars = $state(false)
  let submitting = $state(false)
  let errorMessage = $state('')
  let needsConsent = $state(false)

  const microsoftAccounts = $derived(
    accounts.filter((a) => (a.imapHost || '').includes('outlook')),
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
      if (microsoftAccounts.length === 1) {
        selectedAccountId = microsoftAccounts[0].id
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
      const list = await Calendar_ListMicrosoftCalendarsForAccount(selectedAccountId)
      calendars = list || []
      const acct = accounts.find((a) => a.id === selectedAccountId)
      sourceName = acct?.email ? `Outlook: ${acct.email}` : 'Outlook Calendar'
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
      errorMessage = $_('calendar.settings.addOutlookErrorAccount')
      return
    }
    if (selectedIds.size === 0) {
      errorMessage = $_('calendar.settings.addOutlookErrorPick')
      return
    }
    if (!sourceName.trim()) {
      errorMessage = $_('calendar.settings.addOutlookErrorName')
      return
    }
    submitting = true
    errorMessage = ''
    try {
      const selections = calendars
        .filter((c) => selectedIds.has(c.id))
        .map((c) => ({ id: c.id, displayName: c.name, color: '' }))
      await Calendar_AddMicrosoftSource(selectedAccountId, sourceName.trim(), selections)
      toasts.success($_('calendar.settings.addOutlookToastSuccess'))
      close()
    } catch (err) {
      errorMessage = $_('calendar.settings.addOutlookToastError', { values: { message: String(err) } })
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
      <Dialog.Title>{$_('calendar.settings.addOutlookTitle')}</Dialog.Title>
      <Dialog.Description>{$_('calendar.settings.addOutlookDescription')}</Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4 py-2">
      <!-- Account picker -->
      <div class="space-y-1">
        <Label>{$_('calendar.settings.addOutlookAccountLabel')}</Label>
        {#if loadingAccounts}
          <p class="text-xs text-muted-foreground">{$_('calendar.common.loading')}</p>
        {/if}
        {#if !loadingAccounts && microsoftAccounts.length === 0}
          <p class="text-xs text-muted-foreground">{$_('calendar.settings.addOutlookNoAccounts')}</p>
        {/if}
        {#if !loadingAccounts && microsoftAccounts.length > 0}
          <select
            class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            bind:value={selectedAccountId}
            onchange={() => void onAccountChange()}
          >
            <option value="">{$_('calendar.settings.addOutlookPickAccount')}</option>
            {#each microsoftAccounts as acct (acct.id)}
              <option value={acct.id}>{acct.email}</option>
            {/each}
          </select>
        {/if}
      </div>

      {#if needsConsent}
        <div class="rounded-md border border-yellow-400/40 bg-yellow-400/10 p-3 text-xs text-yellow-700 dark:text-yellow-300">
          {$_('calendar.settings.addOutlookConsentNeeded')}
        </div>
      {/if}

      <!-- Calendar list -->
      {#if loadingCalendars}
        <p class="text-xs text-muted-foreground">{$_('calendar.settings.addOutlookLoadingCalendars')}</p>
      {/if}

      {#if !loadingCalendars && calendars.length > 0}
        <div class="space-y-1">
          <Label>{$_('calendar.settings.addOutlookCalendarsLabel')}</Label>
          <div class="max-h-48 overflow-y-auto rounded-md border border-border">
            {#each calendars as cal (cal.id)}
              <label
                class="flex items-center gap-2 px-3 py-2 text-sm border-b border-border last:border-b-0
                       hover:bg-muted/40 cursor-pointer {!cal.writable ? 'opacity-60' : ''}"
                title={!cal.writable ? $_('calendar.settings.addOutlookReadOnly') : ''}
              >
                <input
                  type="checkbox"
                  checked={selectedIds.has(cal.id)}
                  disabled={!cal.writable}
                  onchange={() => toggleCalendar(cal.id)}
                />
                <span class="truncate flex-1">{cal.name}</span>
                {#if cal.isDefaultCalendar}
                  <span class="text-xs text-muted-foreground">{$_('calendar.settings.addOutlookDefault')}</span>
                {/if}
                {#if !cal.writable}
                  <span class="text-xs text-muted-foreground">{$_('calendar.settings.addOutlookReadOnlyBadge')}</span>
                {/if}
              </label>
            {/each}
          </div>
        </div>

        <!-- Source name -->
        <div class="space-y-1">
          <Label for="ms-source-name">{$_('calendar.settings.addOutlookSourceNameLabel')}</Label>
          <Input id="ms-source-name" bind:value={sourceName} />
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
