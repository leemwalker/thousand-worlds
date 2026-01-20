<script lang="ts">
    /**
     * ModeToggle.svelte
     * Button to toggle between TEXT and VISUAL interface modes.
     */
    import { interfaceMode, toggleInterfaceMode } from "$lib/stores/ui";

    /** Optional: Smaller variant for tight spaces */
    export let compact: boolean = false;
</script>

<button
    data-testid="mode-toggle"
    on:click={toggleInterfaceMode}
    class="mode-toggle flex items-center gap-2 px-3 py-2 rounded-lg transition-all duration-200
         bg-gray-800/80 hover:bg-gray-700/80 border border-gray-600/50 hover:border-gray-500/50
         text-gray-300 hover:text-white shadow-md backdrop-blur-sm
         {compact ? 'text-sm' : 'text-base'}"
    title={$interfaceMode === "TEXT"
        ? "Switch to 3D View"
        : "Switch to Text Mode"}
    aria-label={$interfaceMode === "TEXT"
        ? "Switch to 3D simulation view"
        : "Switch to text-based TXT view"}
>
    {#if $interfaceMode === "TEXT"}
        <!-- Currently in TEXT mode, show icon to switch to VISUAL -->
        <svg
            class="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
        >
            <!-- Globe/3D icon -->
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
            />
        </svg>
        {#if !compact}
            <span>3D View</span>
        {/if}
    {:else}
        <!-- Currently in VISUAL mode, show icon to switch to TEXT -->
        <svg
            class="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
        >
            <!-- Document/Text icon -->
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
        </svg>
        {#if !compact}
            <span>Text Mode</span>
        {/if}
    {/if}
</button>

<style>
    .mode-toggle {
        /* Ensure clickable on overlay layers */
        z-index: 50;
    }

    .mode-toggle:focus {
        outline: none;
        box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.5);
    }

    .mode-toggle:active {
        transform: scale(0.98);
    }
</style>
