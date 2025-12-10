<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { haptic } from "$lib/stores/haptic";
    import DPad from "./DPad.svelte";

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
    <div class="flex flex-wrap gap-2 w-full justify-start">
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

    <div class="flex-shrink-0">
        <DPad on:command={handleMove} />
    </div>
</div>
