<script lang="ts">
  import { onMount } from 'svelte'
  import Icon from '@iconify/svelte'
  // @ts-ignore - wailsjs path
  import { GetAppInfo } from '../../../../wailsjs/go/app/App.js'
  import { BrowserOpenURL } from '../../../../wailsjs/runtime/runtime'
  import logo from '../../../assets/images/logo-universal.png'
  import { _ } from '$lib/i18n'

  interface AppInfo {
    name: string
    version: string
    description: string
    website: string
    license: string
  }

  let appInfo = $state<AppInfo | null>(null)
  let loading = $state(true)

  onMount(async () => {
    try {
      appInfo = await GetAppInfo()
    } catch (err) {
      console.error('Failed to load app info:', err)
    } finally {
      loading = false
    }
  })

  function openWebsite() {
    if (appInfo?.website) {
      BrowserOpenURL(appInfo.website)
    }
  }

  function openUpstream() {
    BrowserOpenURL('https://github.com/hkdb/aerion')
  }
</script>

<div class="flex flex-col items-center justify-center py-6 space-y-6">
  {#if loading}
    <Icon icon="mdi:loading" class="w-8 h-8 animate-spin text-muted-foreground" />
  {:else if appInfo}
    <!-- Logo + App Name & Version -->
    <div class="flex flex-col items-center space-y-2">
      <img src={logo} alt="{appInfo.name} Logo" class="w-24 h-24" />
      <div class="text-center space-y-1">
        <h2 class="text-2xl font-bold text-foreground">{appInfo.name}</h2>
        <p class="text-sm text-muted-foreground">{$_('settingsAbout.version', { values: { version: appInfo.version } })}</p>
      </div>
    </div>

    <!-- Description & Attribution -->
    <div class="text-center space-y-2 max-w-sm">
      <p class="text-sm text-muted-foreground">
        {appInfo.description}
      </p>
      <p class="text-xs text-muted-foreground/80">
        Based on <button onclick={openUpstream} class="underline hover:text-foreground inline-flex items-center gap-0.5">Aerion by HKDB</button> ({appInfo.license})
      </p>
    </div>

    <!-- Links -->
    <div class="flex flex-col items-center gap-2">
      <button
        onclick={openWebsite}
        class="flex items-center gap-2 text-sm text-primary hover:underline transition-colors"
      >
        <Icon icon="mdi:github" class="w-5 h-5" />
        <span>{$_('settingsAbout.github')}</span>
      </button>
    </div>
  {:else}
    <p class="text-muted-foreground">{$_('settingsAbout.failedToLoad')}</p>
  {/if}
</div>
