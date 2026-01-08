<script lang="ts">
    /**
     * MUDModeLayout.svelte
     * Text-based MUD interface layout optimized for narrative gameplay.
     * Default for mobile devices. Focus on rich typography and command input.
     */
    import { isMobile } from "$lib/stores/ui";
</script>

<div
    class="mud-layout flex flex-col h-screen bg-gray-950 text-gray-100 overflow-hidden"
>
    <!-- Status Bar (Fixed Top) -->
    <header
        class="h-14 bg-gray-900 border-b border-gray-800 flex items-center px-4 shrink-0 z-20"
    >
        <slot name="status-bar">
            <div class="text-gray-500 text-sm">Status Bar</div>
        </slot>

        <!-- Mode Toggle Button (right side) -->
        <div class="ml-auto">
            <slot name="mode-toggle" />
        </div>
    </header>

    <!-- Main Content Area -->
    <div class="flex flex-1 min-h-0 overflow-hidden">
        <!-- Left Panel: Stats/Inventory (Desktop only) -->
        {#if !$isMobile}
            <aside
                class="w-64 bg-gray-900 border-r border-gray-800 flex flex-col shrink-0 overflow-hidden"
            >
                <div
                    class="p-3 border-b border-gray-800 text-xs font-semibold text-gray-500 uppercase tracking-wider"
                >
                    Character
                </div>
                <div class="flex-1 overflow-y-auto p-3">
                    <slot name="left-panel">
                        <div class="text-gray-600 text-sm">Stats</div>
                    </slot>
                </div>
            </aside>
        {/if}

        <!-- Center: Text Display (Main Game Output) -->
        <main class="flex-1 flex flex-col min-w-0 overflow-hidden bg-gray-950">
            <div class="flex-1 overflow-y-auto">
                <slot name="main-display">
                    <div class="p-6 text-gray-600 italic text-center">
                        Awaiting connection...
                    </div>
                </slot>
            </div>
        </main>

        <!-- Right Panel: Map/Navigation (Desktop only) -->
        {#if !$isMobile}
            <aside
                class="w-56 bg-gray-900 border-l border-gray-800 flex flex-col shrink-0 overflow-hidden"
            >
                <div
                    class="p-3 border-b border-gray-800 text-xs font-semibold text-gray-500 uppercase tracking-wider"
                >
                    Navigation
                </div>
                <div class="flex-1 overflow-y-auto p-3">
                    <slot name="right-panel">
                        <div class="text-gray-600 text-sm">Map</div>
                    </slot>
                </div>
            </aside>
        {/if}
    </div>

    <!-- Command Input (Fixed Bottom) -->
    <footer class="bg-gray-900 border-t border-gray-800 p-3 shrink-0 z-20">
        <slot name="command-input">
            <div class="text-gray-500 text-sm">Input</div>
        </slot>
    </footer>

    <!-- Mobile Controls (e.g., D-pad) - Mobile only -->
    {#if $isMobile}
        <div
            class="h-16 bg-gray-900 border-t border-gray-800 p-2 shrink-0 flex justify-center items-center"
        >
            <slot name="controls">
                <div class="text-gray-600 text-xs">Controls</div>
            </slot>
        </div>
    {/if}
</div>

<style>
    .mud-layout {
        /* Rich typography for MUD feel */
        font-family: "Georgia", "Times New Roman", Times, serif;
    }

    /* Use mono for command input areas */
    :global(.mud-layout input),
    :global(.mud-layout code),
    :global(.mud-layout pre) {
        font-family: "SF Mono", "Monaco", "Inconsolata", "Fira Code",
            "Fira Mono", "Source Code Pro", monospace;
    }
</style>
