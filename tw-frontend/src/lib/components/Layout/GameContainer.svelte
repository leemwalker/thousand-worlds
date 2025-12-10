<script lang="ts">
  import { onMount } from 'svelte';
  import { isMobile, setScreenWidth } from '$lib/stores/ui';
  import MobileLayout from './MobileLayout.svelte';
  import DesktopLayout from './DesktopLayout.svelte';

  onMount(() => {
    const handleResize = () => {
      setScreenWidth(window.innerWidth);
    };
    
    // Initial set
    handleResize();
    
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  });
</script>

<div class="game-container w-full h-full">
  {#if $isMobile}
    <MobileLayout>
      <slot name="status-bar" slot="status-bar" />
      <slot name="main-display" slot="main-display" />
      <slot name="command-input" slot="command-input" />
      <slot name="controls" slot="controls" />
    </MobileLayout>
  {:else}
    <DesktopLayout>
      <slot name="status-bar" slot="status-bar" />
      <slot name="main-display" slot="main-display" />
      <slot name="command-input" slot="command-input" />
      <slot name="left-panel" slot="left-panel" />
      <slot name="right-panel" slot="right-panel" />
    </DesktopLayout>
  {/if}
</div>
