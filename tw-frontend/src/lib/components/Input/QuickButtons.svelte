<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { haptic } from "$lib/stores/haptic";
    import DPad from "./DPad.svelte";
    import MiniMap from "$lib/components/Map/MiniMap.svelte";

    const dispatch = createEventDispatcher();

    export let commands: Array<{ label: string; command: string }> = [
        { label: "Inventory", command: "inventory" },
        { label: "Help", command: "help" },
    ];

    function sendQuickCommand(command: string) {
        haptic.selection();
        dispatch("submit", command);
    }

    function handleMove(event: CustomEvent<string>) {
        sendQuickCommand(event.detail);
    }
</script>

<div class="flex flex-row gap-4 w-full items-end justify-between">
    <!-- MiniMap on the left, mirroring DPad on the right -->
    <div class="flex-shrink-0">
        <MiniMap />
    </div>

    <!-- Quick command buttons in the center -->
    <div class="flex flex-wrap gap-2 flex-1 justify-center">
        {#each commands as cmd}
            <button
                on:click={() => sendQuickCommand(cmd.command)}
                class="min-w-[44px] min-h-[44px] px-4 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 rounded-lg transition-colors text-sm font-medium"
                aria-label={cmd.label}
            >
                {cmd.label}
            </button>
        {/each}
    </div>

    <!-- DPad on the right -->
    <div class="flex-shrink-0">
        <DPad on:command={handleMove} />
    </div>
</div>
