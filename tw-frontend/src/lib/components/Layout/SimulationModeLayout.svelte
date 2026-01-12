<script lang="ts">
    /**
     * SimulationModeLayout.svelte
     * 3D Babylon.js interface with command input overlay.
     * Default for desktop devices. Full-screen 3D with HUD elements.
     */
    import { onMount, onDestroy } from "svelte";
    import { isMobile } from "$lib/stores/ui";
    import { gameStore } from "$lib/stores/game";
    import { gameWebSocket } from "$lib/services/websocket";
    // Layout and UI Components
    import MessageOverlay from "$lib/components/HUD/MessageOverlay.svelte";
    import GameMenuModal from "$lib/components/Layout/GameMenuModal.svelte";

    // Scene Management
    import SceneCanvas from "$lib/components/Scene/SceneCanvas.svelte";
    import {
        sceneManager,
        type GameLocation,
    } from "$lib/components/Scene/SceneManager";
    import { LobbyScene } from "$lib/components/Scene/LobbyScene";
    import WorldController from "$lib/components/Map/WorldController.svelte";
    import { authStore } from "$lib/stores/auth"; // Start logic for menu?

    /** Start with menu closed */
    let isMenuOpen = false;

    // ... (rest of imports are fine, skipping to HTML structure)

    // ... (logic) ...

    function handleMenuClose() {
        isMenuOpen = false;
    }

    function handleResetWorld() {
        if (!gameWebSocket.isConnected()) return;
        gameWebSocket.sendRawCommand("reset_world", {});
        isMenuOpen = false;
    }

    function handleReturnToLobby() {
        gameStore.setGameLocation("LOBBY");
        isMenuOpen = false;
    }

    function handleLogout() {
        authStore.logout();
        isMenuOpen = false;
    }
</script>

<div
    class="simulation-layout relative w-full h-screen bg-black overflow-hidden"
>
    <!-- 3D Canvas (unchanged) -->
    <div class="absolute inset-0 z-0">
        <SceneCanvas on:canvasReady={handleCanvasReady} />
        <!-- WorldController block ... -->
        {#if $gameStore.gameLocation === "WORLD" && activeScene}
            <!-- ... existing WorldController ... -->
            <!-- We need to preserve the implementation details here -->
            <WorldController
                scene={activeScene}
                globeTextureBlob={$gameStore.world.textureBlob}
                globeHeightmapBlob={$gameStore.world.heightmapBlob}
                materialBlob={$gameStore.world.materialBlob}
                iceBlob={$gameStore.world.iceBlob}
                seaLevel={$gameStore.world.geo.seaLevel}
                maxElevation={$gameStore.world.geo.maxElevation}
                minElevation={$gameStore.world.geo.minElevation}
                satellites={$gameStore.world.sim.satellites}
                pois={$gameStore.world.sim.pois}
                onSendCommand={(action, payload) => {
                    if (typeof payload === "object") {
                        gameWebSocket.sendRawCommand(action, payload);
                    } else {
                        try {
                            const p = payload ? JSON.parse(payload) : {};
                            gameWebSocket.sendRawCommand(action, p);
                        } catch (e) {
                            gameWebSocket.sendRawCommand(action, {
                                data: payload,
                            });
                        }
                    }
                }}
            />
        {/if}
    </div>

    <!-- HUD Overlay Layer -->
    <div class="absolute inset-0 z-10 pointer-events-none">
        <!-- Game Menu Modal (Pointer events auto when open) -->
        <div class="pointer-events-auto">
            <GameMenuModal
                isOpen={isMenuOpen}
                on:close={handleMenuClose}
                on:resetWorld={handleResetWorld}
                on:returnToLobby={handleReturnToLobby}
                on:logout={handleLogout}
            />
        </div>

        <!-- Fading Messages Overlay -->
        <MessageOverlay />

        <!-- Top Bar: Menu Button + Mode Toggle -->
        <header
            class="absolute top-0 left-0 right-0 h-14 flex items-center px-4 pointer-events-auto bg-gradient-to-b from-black/60 to-transparent"
        >
            <!-- Menu Button (Top Left) -->
            <button
                class="flex items-center gap-2 px-3 py-1.5 bg-gray-900/80 hover:bg-gray-800 border border-gray-700 rounded-md text-gray-200 transition-colors shadow-sm backdrop-blur-sm"
                on:click={() => (isMenuOpen = true)}
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="h-5 w-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M4 6h16M4 12h16M4 18h16"
                    />
                </svg>
                <span class="text-sm font-medium">Menu</span>
            </button>

            <!-- Mode Toggle Button (right side) -->
            <div class="ml-auto">
                <slot name="mode-toggle" />
            </div>
        </header>

        <!-- Left Side: Mini Stats (Desktop only) -->
        {#if !$isMobile}
            <aside class="absolute top-16 left-4 w-48 pointer-events-auto">
                <div
                    class="bg-gray-900/80 backdrop-blur-sm rounded-lg border border-gray-700/50 p-3 shadow-lg"
                >
                    <slot name="hud-stats">
                        <div class="text-gray-400 text-xs">Stats Overlay</div>
                    </slot>
                </div>
            </aside>
        {/if}

        <!-- Right Side: Minimap (Desktop only) -->
        {#if !$isMobile}
            <aside class="absolute top-16 right-4 pointer-events-auto">
                <div
                    class="w-40 h-40 bg-gray-900/80 backdrop-blur-sm rounded-lg border border-gray-700/50 overflow-hidden shadow-lg"
                >
                    <slot name="minimap">
                        <div
                            class="w-full h-full flex items-center justify-center text-gray-500 text-xs"
                        >
                            Minimap
                        </div>
                    </slot>
                </div>
            </aside>
        {/if}

        <!-- Bottom: Command Input + Text Log -->
        <div class="absolute bottom-0 left-0 right-0 pointer-events-auto">
            <!-- Collapsible Text Log -->
            {#if textLogExpanded}
                <div
                    class="mx-4 mb-2 max-h-64 bg-gray-900/90 backdrop-blur-sm rounded-t-lg border border-b-0 border-gray-700/50 overflow-hidden shadow-lg"
                >
                    <div
                        class="flex items-center justify-between px-3 py-2 border-b border-gray-700/50"
                    >
                        <span
                            class="text-xs font-semibold text-gray-400 uppercase tracking-wider"
                            >Game Log</span
                        >
                        <button
                            class="text-gray-500 hover:text-gray-300 p-1"
                            on:click={() => (textLogExpanded = false)}
                            aria-label="Collapse log"
                        >
                            <svg
                                class="w-4 h-4"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M19 9l-7 7-7-7"
                                />
                            </svg>
                        </button>
                    </div>
                    <div class="h-48 overflow-y-auto p-3">
                        <slot name="text-log">
                            <div class="text-gray-500 text-sm italic">
                                No messages yet...
                            </div>
                        </slot>
                    </div>
                </div>
            {/if}

            <!-- Command Input Bar -->
            <div
                class="bg-gray-900/90 backdrop-blur-sm border-t border-gray-700/50 p-3"
            >
                <div class="flex items-center gap-2">
                    <!-- Expand Log Button -->
                    {#if !textLogExpanded}
                        <button
                            class="p-2 text-gray-400 hover:text-gray-200 hover:bg-gray-800/50 rounded transition-colors"
                            on:click={() => (textLogExpanded = true)}
                            aria-label="Expand game log"
                            title="Show game log"
                        >
                            <svg
                                class="w-5 h-5"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M4 6h16M4 12h16M4 18h16"
                                />
                            </svg>
                        </button>
                    {/if}

                    <!-- Command Input -->
                    <div class="flex-1">
                        <slot name="command-input">
                            <div class="text-gray-500 text-sm">Input</div>
                        </slot>
                    </div>
                </div>
            </div>
        </div>

        <!-- Mobile: Touch Controls Overlay -->
        {#if $isMobile}
            <div class="absolute bottom-24 right-4 pointer-events-auto">
                <slot name="controls">
                    <div
                        class="w-24 h-24 bg-gray-900/60 rounded-full flex items-center justify-center"
                    >
                        <span class="text-gray-500 text-xs">D-Pad</span>
                    </div>
                </slot>
            </div>
        {/if}
    </div>
</div>

<style>
    .simulation-layout {
        /* Prevent text selection on HUD */
        user-select: none;
        -webkit-user-select: none;
    }

    /* Allow text selection in text log and input */
    :global(.simulation-layout input),
    :global(.simulation-layout .text-log) {
        user-select: text;
        -webkit-user-select: text;
    }
</style>
