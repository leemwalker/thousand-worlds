<script lang="ts">
    import { afterUpdate } from "svelte";

    export let events: string[] = [];

    let container: HTMLElement;

    // Auto-scroll to bottom when new events arrive
    afterUpdate(() => {
        if (container) {
            container.scrollTop = container.scrollHeight;
        }
    });

    // Format event with emoji based on type
    function formatEvent(event: string): string {
        if (event.includes("Impact") || event.includes("Asteroid"))
            return `☄️ ${event}`;
        if (event.includes("Extinction")) return `💀 ${event}`;
        if (event.includes("Volcanic") || event.includes("Eruption"))
            return `🌋 ${event}`;
        if (event.includes("Ice") || event.includes("Glaciation"))
            return `❄️ ${event}`;
        if (event.includes("Flood") || event.includes("Deluge"))
            return `🌊 ${event}`;
        if (event.includes("Tectonic") || event.includes("Plate"))
            return `🏔️ ${event}`;
        return `📍 ${event}`;
    }
</script>

<div class="simulation-log" bind:this={container}>
    <h3>Geological Event Log</h3>
    {#if events.length === 0}
        <p class="placeholder">No events recorded yet...</p>
    {:else}
        <ul>
            {#each events as event}
                <li>{formatEvent(event)}</li>
            {/each}
        </ul>
    {/if}
</div>

<style>
    .simulation-log {
        background: rgba(0, 0, 0, 0.75);
        border: 1px solid rgba(255, 255, 255, 0.15);
        height: 180px;
        overflow-y: auto;
        padding: 0.75rem;
        font-family: "JetBrains Mono", "Courier New", monospace;
        font-size: 0.75rem;
        color: #e0e0e0;
        border-radius: 6px;
        backdrop-filter: blur(8px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    }

    h3 {
        margin: 0 0 0.5rem 0;
        font-size: 0.7rem;
        color: #888;
        text-transform: uppercase;
        letter-spacing: 0.1em;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        padding-bottom: 0.4rem;
    }

    .placeholder {
        color: #666;
        font-style: italic;
        margin: 0;
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }

    li {
        padding: 3px 0;
        border-bottom: 1px solid rgba(255, 255, 255, 0.03);
        animation: fadeIn 0.3s ease-out;
    }

    li:last-child {
        border-bottom: none;
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(-4px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    /* Custom scrollbar */
    .simulation-log::-webkit-scrollbar {
        width: 6px;
    }

    .simulation-log::-webkit-scrollbar-track {
        background: rgba(0, 0, 0, 0.2);
    }

    .simulation-log::-webkit-scrollbar-thumb {
        background: rgba(255, 255, 255, 0.15);
        border-radius: 3px;
    }

    .simulation-log::-webkit-scrollbar-thumb:hover {
        background: rgba(255, 255, 255, 0.25);
    }
</style>
