<script lang="ts">
    import { gameWebSocket } from "$lib/services/websocket";
    import { haptic } from "$lib/stores/haptic";

    export let commands: Array<{ label: string; command: string }> = [
        { label: "Look", command: "look" },
        { label: "North", command: "north" },
        { label: "South", command: "south" },
        { label: "Inventory", command: "inventory" },
    ];

    function sendQuickCommand(command: string) {
        haptic.selection();
        gameWebSocket.sendRawCommand(command);
    }
</script>

<div class="flex flex-wrap gap-2 w-full">
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
