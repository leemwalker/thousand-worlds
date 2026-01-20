<script lang="ts">
  /**
   * GameContainer.svelte
   * Main container that switches between TXT and Simulation layouts based on InterfaceMode.
   */
  import { onMount } from "svelte";
  import { interfaceMode, setScreenWidth } from "$lib/stores/ui";
  import TXTModeLayout from "./TXTModeLayout.svelte";
  import SimulationModeLayout from "./SimulationModeLayout.svelte";

  onMount(() => {
    const handleResize = () => {
      setScreenWidth(window.innerWidth);
    };

    // Initial set
    handleResize();

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  });
</script>

<div
  class="game-container w-full h-full"
  data-testid="game-container"
  data-mode={$interfaceMode}
>
  {#if $interfaceMode === "TEXT"}
    <TXTModeLayout>
      <slot name="status-bar" slot="status-bar" />
      <slot name="main-display" slot="main-display" />
      <slot name="command-input" slot="command-input" />
      <slot name="left-panel" slot="left-panel" />
      <slot name="right-panel" slot="right-panel" />
      <slot name="controls" slot="controls" />
      <slot name="mode-toggle" slot="mode-toggle" />
    </TXTModeLayout>
  {:else}
    <SimulationModeLayout>
      <slot name="canvas" slot="canvas" />
      <slot name="status-bar" slot="status-bar" />
      <slot name="command-input" slot="command-input" />
      <slot name="text-log" slot="text-log" />
      <slot name="hud-stats" slot="hud-stats" />
      <slot name="minimap" slot="minimap" />
      <slot name="controls" slot="controls" />
      <slot name="mode-toggle" slot="mode-toggle" />
    </SimulationModeLayout>
  {/if}
</div>
