<script lang="ts">
  import RailButton from './RailButton.svelte'
  import { getRailTabs, isRailVisible } from '$lib/stores/extensionRegistry.svelte'
  import { getActiveExtension, setActiveExtension } from '$lib/stores/uiState.svelte'

  let active = $derived(getActiveExtension())
  let visible = $derived(isRailVisible())
  let tabs = $derived(getRailTabs())

  function select(name: string) {
    setActiveExtension(name)
  }
</script>

{#if visible}
  <nav
    class="flex flex-col items-stretch w-12 flex-shrink-0 bg-muted/30 border-r border-border py-2 justify-between"
    aria-label="Active extension"
  >
    <div class="flex flex-col items-stretch gap-1">
      <RailButton
        icon="mdi:email"
        label="Mail"
        active={active === 'mail'}
        onclick={() => select('mail')}
      />
      {#each tabs as tab (tab.extensionId)}
        <RailButton
          icon={tab.icon || 'mdi:puzzle'}
          label={tab.label}
          active={active === tab.extensionId}
          onclick={() => select(tab.extensionId)}
        />
      {/each}
    </div>

    <div class="flex flex-col items-stretch">
      <RailButton
        icon="mdi:cog"
        label="Settings"
        active={active === 'settings'}
        onclick={() => select('settings')}
      />
    </div>
  </nav>
{/if}
