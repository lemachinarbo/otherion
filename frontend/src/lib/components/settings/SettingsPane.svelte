<script lang="ts">
  import { onMount } from 'svelte'
  import Icon from '@iconify/svelte'
  import { Button } from '$lib/components/ui/button'
  import { _ } from '$lib/i18n'
  import { setActiveExtension } from '$lib/stores/uiState.svelte'

  import GeneralTab from './GeneralTab.svelte'
  import ComposerTab from './ComposerTab.svelte'
  import ImagesTab from './ImagesTab.svelte'
  import AccountsTab from './AccountsTab.svelte'
  import ContactsTab from './ContactsTab.svelte'
  import ExtensionsTab from './ExtensionsTab.svelte'
  import ShortcutsTab from './ShortcutsTab.svelte'
  import AboutTab from './AboutTab.svelte'

  // Settings state (mirrors SettingsDialog logic)
  // @ts-ignore - wailsjs path
  import {
    GetReadReceiptResponsePolicy,
    SetReadReceiptResponsePolicy,
    GetMarkAsReadDelay,
    SetMarkAsReadDelay,
    GetMessageListDensity,
    SetMessageListDensity,
    GetThemeMode,
    SetThemeMode,
    GetShowTitleBar,
    SetShowTitleBar,
    GetRunBackground,
    SetRunBackground,
    GetStartHidden,
    SetStartHidden,
    GetAutostart,
    SetAutostart,
    GetLanguage,
    SetLanguage,
    GetComposerMode,
    SetComposerMode,
    GetMailtoMode,
    SetMailtoMode,
    GetComposerFormat,
    SetComposerFormat,
    GetNativeTitleBar,
    SetNativeTitleBar,
    GetAlwaysLoadImages,
    SetAlwaysLoadImages,
    GetDarkMailContent,
    SetDarkMailContent,
    GetOverrideEmailColors,
    SetOverrideEmailColors,
    GetAccentBarUnread,
    SetAccentBarUnread,
    GetShowMessageListCircles,
    SetShowMessageListCircles,
    GetShowViewerCircles,
    SetShowViewerCircles,
    GetShowActionToasts,
    SetShowActionToasts,
    GetNewMailNotificationsEnabled,
    SetNewMailNotificationsEnabled,
  } from '../../../../wailsjs/go/app/App.js'

  import { addToast } from '$lib/stores/toast'
  import {
    setMessageListDensity as updateDensityStore,
    setThemeMode as updateThemeStore,
    setShowTitleBar as updateShowTitleBarStore,
    setRunBackground as updateRunBackgroundStore,
    setStartHidden as updateStartHiddenStore,
    setAutostart as updateAutostartStore,
    setLanguage as updateLanguageStore,
    setComposerMode as updateComposerModeStore,
    setMailtoMode as updateMailtoModeStore,
    setComposerFormat as updateComposerFormatStore,
    setNativeTitleBar as updateNativeTitleBarStore,
    setAlwaysLoadImages as updateAlwaysLoadImagesStore,
    setDarkMailContent as updateDarkMailContentStore,
    setOverrideEmailColors as updateOverrideEmailColorsStore,
    setAccentBarUnread as updateAccentBarUnreadStore,
    setShowMessageListCircles as updateShowMessageListCirclesStore,
    setShowViewerCircles as updateShowViewerCirclesStore,
    setShowActionToasts as updateShowActionToastsStore,
    setNewMailNotificationsEnabled as updateNewMailNotificationsEnabledStore,
    type MessageListDensity,
    type ThemeMode,
    type ComposerMode,
    type ComposerFormat,
  } from '$lib/stores/settings.svelte'
  import { applyThemeFromMode } from '$lib/stores/theme.svelte'

  let activeCategory = $state('general')
  let loading = $state(true)
  let saving = $state(false)

  // Form values
  let readReceiptResponsePolicy = $state<string>('ask')
  let markAsReadDelaySeconds = $state<number>(1)
  let messageListDensity = $state<string>('standard')
  let themeMode = $state<string>('system')
  let showTitleBar = $state<boolean>(true)
  let runBackground = $state<boolean>(false)
  let startHidden = $state<boolean>(false)
  let autostart = $state<boolean>(false)
  let language = $state<string>('')
  let composerMode = $state<string>('inline')
  let mailtoMode = $state<string>('inline')
  let composerFormat = $state<string>('rich')
  let nativeTitleBar = $state<boolean>(false)
  let alwaysLoadImages = $state<boolean>(false)
  let darkMailContent = $state<boolean>(false)
  let overrideEmailColors = $state<boolean>(false)
  let accentBarUnread = $state<boolean>(false)
  let showMessageListCircles = $state<boolean>(true)
  let showViewerCircles = $state<boolean>(true)
  let showActionToasts = $state<boolean>(true)
  let newMailNotificationsEnabled = $state<boolean>(true)

  $effect(() => {
    if (loading || !themeMode) return
    applyThemeFromMode(themeMode as ThemeMode)
  })

  onMount(async () => {
    try {
      const [
        policy,
        delayMs,
        density,
        theme,
        titleBar,
        bg,
        hidden,
        auto,
        lang,
        cMode,
        mMode,
        cFormat,
        nTitleBar,
        aImages,
        dMail,
        oColors,
        aUnread,
        mCircles,
        vCircles,
        aToasts,
        newMailNotif,
      ] = await Promise.all([
        GetReadReceiptResponsePolicy(),
        GetMarkAsReadDelay(),
        GetMessageListDensity(),
        GetThemeMode(),
        GetShowTitleBar(),
        GetRunBackground(),
        GetStartHidden(),
        GetAutostart(),
        GetLanguage(),
        GetComposerMode(),
        GetMailtoMode(),
        GetComposerFormat(),
        GetNativeTitleBar(),
        GetAlwaysLoadImages(),
        GetDarkMailContent(),
        GetOverrideEmailColors(),
        GetAccentBarUnread(),
        GetShowMessageListCircles(),
        GetShowViewerCircles(),
        GetShowActionToasts(),
        GetNewMailNotificationsEnabled(),
      ])

      readReceiptResponsePolicy = policy || 'ask'
      markAsReadDelaySeconds = delayMs < 0 ? -1 : Math.round(delayMs / 1000)
      messageListDensity = density || 'standard'
      themeMode = theme || 'system'
      showTitleBar = titleBar ?? true
      runBackground = bg ?? false
      startHidden = hidden ?? false
      autostart = auto ?? false
      language = lang || ''
      composerMode = cMode || 'inline'
      mailtoMode = mMode || 'inline'
      composerFormat = cFormat || 'rich'
      nativeTitleBar = nTitleBar ?? false
      alwaysLoadImages = aImages ?? false
      darkMailContent = dMail ?? false
      overrideEmailColors = oColors ?? false
      accentBarUnread = aUnread ?? false
      showMessageListCircles = mCircles ?? true
      showViewerCircles = vCircles ?? true
      showActionToasts = aToasts ?? true
      newMailNotificationsEnabled = newMailNotif ?? true
    } catch (err) {
      console.error('Failed to load settings:', err)
      addToast({
        message: $_('settings.loadFailed') + ': ' + String(err),
        type: 'error',
      })
    } finally {
      loading = false
    }
  })

  async function handleSave() {
    saving = true
    try {
      const delayMs = markAsReadDelaySeconds < 0 ? -1 : markAsReadDelaySeconds * 1000

      await Promise.all([
        SetReadReceiptResponsePolicy(readReceiptResponsePolicy),
        SetMarkAsReadDelay(delayMs),
        SetMessageListDensity(messageListDensity),
        SetThemeMode(themeMode),
        SetShowTitleBar(showTitleBar),
        SetRunBackground(runBackground),
        SetStartHidden(startHidden),
        SetAutostart(autostart),
        SetLanguage(language),
        SetComposerMode(composerMode),
        SetMailtoMode(mailtoMode),
        SetComposerFormat(composerFormat),
        SetNativeTitleBar(nativeTitleBar),
        SetAlwaysLoadImages(alwaysLoadImages),
        SetDarkMailContent(darkMailContent),
        SetOverrideEmailColors(overrideEmailColors),
        SetAccentBarUnread(accentBarUnread),
        SetShowMessageListCircles(showMessageListCircles),
        SetShowViewerCircles(showViewerCircles),
        SetShowActionToasts(showActionToasts),
        SetNewMailNotificationsEnabled(newMailNotificationsEnabled),
      ])

      updateDensityStore(messageListDensity as MessageListDensity)
      updateThemeStore(themeMode as ThemeMode)
      updateShowTitleBarStore(showTitleBar)
      updateRunBackgroundStore(runBackground)
      updateStartHiddenStore(startHidden)
      updateAutostartStore(autostart)
      updateLanguageStore(language)
      updateComposerModeStore(composerMode as ComposerMode)
      updateMailtoModeStore(mailtoMode as ComposerMode)
      updateComposerFormatStore(composerFormat as ComposerFormat)
      updateNativeTitleBarStore(nativeTitleBar)
      updateAlwaysLoadImagesStore(alwaysLoadImages)
      updateDarkMailContentStore(darkMailContent)
      updateOverrideEmailColorsStore(overrideEmailColors)
      updateAccentBarUnreadStore(accentBarUnread)
      updateShowMessageListCirclesStore(showMessageListCircles)
      updateShowViewerCirclesStore(showViewerCircles)
      updateShowActionToastsStore(showActionToasts)
      updateNewMailNotificationsEnabledStore(newMailNotificationsEnabled)

      addToast({
        message: $_('settings.savedSuccess'),
        type: 'success',
      })
    } catch (err) {
      console.error('Failed to save settings:', err)
      addToast({
        message: $_('settings.saveFailed') + ': ' + String(err),
        type: 'error',
      })
    } finally {
      saving = false
    }
  }

  const categories = [
    { id: 'general', label: 'settings.tabGeneral', icon: 'mdi:tune' },
    { id: 'composer', label: 'settings.tabComposer', icon: 'mdi:square-edit-outline' },
    { id: 'images', label: 'settings.tabImages', icon: 'mdi:image-outline' },
    { id: 'accounts', label: 'settings.tabAccounts', icon: 'mdi:email-outline' },
    { id: 'contacts', label: 'settings.tabContacts', icon: 'mdi:account-box-outline' },
    { id: 'extensions', label: 'settings.tabExtensions', icon: 'mdi:puzzle-outline' },
    { id: 'shortcuts', label: 'settings.tabShortcuts', icon: 'mdi:keyboard-outline' },
    { id: 'about', label: 'settings.tabAbout', icon: 'mdi:information-outline' },
  ]
</script>

<div class="flex-1 flex flex-col min-w-0 bg-background overflow-hidden h-full">
  <!-- Header -->
  <header class="flex items-center justify-between px-6 py-4 border-b border-border bg-card/30">
    <div class="flex items-center gap-3">
      <div class="p-2 rounded-lg bg-primary/10 text-primary">
        <Icon icon="mdi:cog" class="w-6 h-6" />
      </div>
      <div>
        <h1 class="text-xl font-bold text-foreground">{$_('settings.title')}</h1>
        <p class="text-xs text-muted-foreground">{$_('settings.subtitle')}</p>
      </div>
    </div>
    <div class="flex items-center gap-3">
      <Button variant="outline" onclick={() => setActiveExtension('mail')}>
        {$_('common.close')}
      </Button>
      <Button disabled={saving} onclick={handleSave}>
        {#if saving}
          <Icon icon="mdi:loading" class="w-4 h-4 mr-2 animate-spin" />
        {/if}
        {$_('common.save')}
      </Button>
    </div>
  </header>

  <!-- Content -->
  <div class="flex-1 flex min-h-0 overflow-hidden">
    <!-- Category Sidebar -->
    <aside class="w-56 border-r border-border bg-muted/20 p-3 space-y-1 overflow-y-auto flex-shrink-0">
      {#each categories as cat (cat.id)}
        <button
          class="w-full flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-md transition-colors text-left {activeCategory === cat.id
            ? 'bg-primary text-primary-foreground shadow-sm'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
          onclick={() => (activeCategory = cat.id)}
        >
          <Icon icon={cat.icon} class="w-4 h-4" />
          <span>{$_(cat.label)}</span>
        </button>
      {/each}
    </aside>

    <!-- Category Detail Pane -->
    <main class="flex-1 p-6 overflow-y-auto max-w-4xl">
      {#if loading}
        <div class="flex items-center justify-center py-16">
          <Icon icon="mdi:loading" class="w-8 h-8 animate-spin text-muted-foreground" />
        </div>
      {:else if activeCategory === 'general'}
        <GeneralTab
          bind:markAsReadDelaySeconds
          bind:messageListDensity
          bind:themeMode
          bind:nativeTitleBar
          bind:showTitleBar
          bind:runBackground
          bind:startHidden
          bind:autostart
          bind:language
          bind:accentBarUnread
          bind:showMessageListCircles
          bind:showViewerCircles
          bind:darkMailContent
          bind:overrideEmailColors
          bind:showActionToasts
          bind:newMailNotificationsEnabled
          onDelayChange={(v) => (markAsReadDelaySeconds = v)}
          onDensityChange={(v) => (messageListDensity = v)}
          onThemeChange={(v) => (themeMode = v)}
          onTitleBarChange={(n, s) => {
            nativeTitleBar = n
            showTitleBar = s
          }}
          onRunBackgroundChange={(v) => (runBackground = v)}
          onStartHiddenChange={(v) => (startHidden = v)}
          onAutostartChange={(v) => (autostart = v)}
          onLanguageChange={(v) => (language = v)}
        />
      {:else if activeCategory === 'composer'}
        <ComposerTab
          readReceiptResponsePolicy={readReceiptResponsePolicy}
          composerMode={composerMode}
          mailtoMode={mailtoMode}
          composerFormat={composerFormat}
          onPolicyChange={(v) => (readReceiptResponsePolicy = v)}
          onComposerModeChange={(v) => (composerMode = v)}
          onMailtoModeChange={(v) => (mailtoMode = v)}
          onFormatChange={(v) => (composerFormat = v)}
        />
      {:else if activeCategory === 'images'}
        <ImagesTab
          bind:alwaysLoadImages
          onAlwaysLoadImagesChange={(v) => (alwaysLoadImages = v)}
        />
      {:else if activeCategory === 'accounts'}
        <AccountsTab />
      {:else if activeCategory === 'contacts'}
        <ContactsTab />
      {:else if activeCategory === 'extensions'}
        <ExtensionsTab />
      {:else if activeCategory === 'shortcuts'}
        <ShortcutsTab />
      {:else if activeCategory === 'about'}
        <AboutTab />
      {/if}
    </main>
  </div>
</div>
