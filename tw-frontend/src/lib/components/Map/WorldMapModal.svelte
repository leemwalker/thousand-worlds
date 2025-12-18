<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { MapRenderer } from "./MapRenderer";
    import type { VisibleTile } from "./MapRenderer";
    import { mapStore } from "$lib/stores/map";
    import { gameWebSocket } from "$lib/services/websocket";
    import { fade } from "svelte/transition";

    export let isOpen = false;
    export let onClose: () => void;

    let canvas: HTMLCanvasElement;
    let renderer: MapRenderer | null = null;
    let containerWidth = 0;
    let containerHeight = 0;

    // Simulation stats
    let simStats = {
        year: 0,
        population: 0,
        species: 0,
        events: [] as string[],
    };

    $: if (isOpen && canvas) {
        initRenderer();
    }

    $: if (!isOpen && renderer) {
        renderer.stop();
        renderer = null;
    }

    // Subscribe to map store updates
    $: if (renderer && $mapStore.data && isOpen) {
        updateMap();
    }

    function initRenderer() {
        if (!canvas || renderer) return;

        const ctx = canvas.getContext("2d", { alpha: false });
        if (!ctx) return;

        renderer = new MapRenderer(canvas);
        renderer.setTileSize(4); // Smaller tiles for world view? Or maybe standard.
        // For world map modal, we might want a different zoom level or dynamic zoom.
        // For now, reuse the renderer with potentially different settings.
        renderer.start();

        if ($mapStore.data) {
            updateMap();
        }
    }

    function updateMap() {
        if (!renderer || !$mapStore.data) return;

        const playerPos = {
            x: Math.round($mapStore.data.player_x),
            y: Math.round($mapStore.data.player_y),
        };

        const visibleTiles: VisibleTile[] = $mapStore.data.tiles.map(
            (tile) => ({
                x: tile.x,
                y: tile.y,
                biome: tile.biome || "Default",
                elevation: tile.elevation || 0,
                entities: tile.entities || [],
                is_player: tile.is_player,
                portal: tile.portal,
                occluded: tile.occluded,
            }),
        );

        // Use a smaller scale or different settings for world map if needed
        renderer.updateData(playerPos, visibleTiles, 1.0);
    }

    // Listen for sim events
    function handleSimMessage(msg: any) {
        if (msg.type === "sim_event") {
            simStats.year = msg.data.year || simStats.year;
            // Add to event log
            const eventText = `Year ${msg.data.year}: ${msg.text}`;
            simStats.events = [eventText, ...simStats.events].slice(0, 50); // Keep last 50
        } else if (
            msg.type === "game_message" &&
            msg.data.type === "sim_stats"
        ) {
            // Handle explicit stats update if we add that later
            // For now assume sim_event carries enough info or we rely on mapStore meta
        }
    }

    let unsubscribeWS: (() => void) | null = null;

    onMount(() => {
        // We need to hook into the websocket stream.
        // Ideally +page.svelte passes messages down, or we subscribe to a store.
        // For now, let's assume we can subscribe to gameWebSocket messages directly or a store.
        // gameWebSocket.onMessage is a single handler setter, so we shouldn't overwrite it.
        // We should use a store for messages if available, or event bus.
        // But `gameWebSocket` implementation in `websocket.ts` might allow multiple listeners?
        // Checking `websocket.ts` would be good.
        // Assuming we can't easily hook cleanly without changing websocket service,
        // we will rely on +page.svelte passing props or context.
        // OR: `gameWebSocket` could be an EventTarget?
        // Workaround: We'll poll or rely on mapStore updates for visual,
        // and maybe add a distinct store for sim events later.
        // Actually, let's just use the `messages` store if it existed.
        // For now, simple implementation: visual only + static stats placeholder
    });

    onDestroy(() => {
        if (renderer) renderer.stop();
    });
</script>

{#if isOpen}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm"
        transition:fade
    >
        <div
            class="bg-gray-900 border border-gray-700 rounded-lg shadow-2xl w-[90vw] h-[90vh] flex flex-col overflow-hidden"
        >
            <!-- Header -->
            <div
                class="flex justify-between items-center p-4 border-b border-gray-800 bg-gray-800/50"
            >
                <h2 class="text-xl font-bold text-blue-400">
                    World Map & Simulation
                </h2>
                <div class="flex gap-4">
                    <div class="text-sm text-gray-400">
                        Year: <span class="text-white font-mono"
                            >{simStats.year}</span
                        >
                    </div>
                    <button
                        on:click={onClose}
                        class="text-gray-400 hover:text-white transition-colors"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            class="h-6 w-6"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M6 18L18 6M6 6l12 12"
                            />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Content -->
            <div class="flex-1 flex overflow-hidden">
                <!-- Map Area -->
                <div
                    class="flex-1 bg-black relative"
                    bind:clientWidth={containerWidth}
                    bind:clientHeight={containerHeight}
                >
                    <canvas
                        bind:this={canvas}
                        width={containerWidth}
                        height={containerHeight}
                        class="block w-full h-full"
                    ></canvas>

                    <!-- Overlay Controls -->
                    <div class="absolute bottom-4 right-4 flex gap-2">
                        <div
                            class="bg-gray-800/80 p-2 rounded text-xs text-gray-300"
                        >
                            WASD to Pan • Scroll to Zoom
                        </div>
                    </div>
                </div>

                <!-- Sidebar / Stats -->
                <div
                    class="w-80 bg-gray-850 border-l border-gray-800 flex flex-col"
                >
                    <div class="p-4 border-b border-gray-800">
                        <h3 class="font-bold text-gray-300 mb-2">
                            Ecosystem Stats
                        </h3>
                        <div class="grid grid-cols-2 gap-2 text-sm">
                            <div class="bg-gray-800 p-2 rounded">
                                <div class="text-gray-500 text-xs">
                                    Population
                                </div>
                                <div class="text-green-400 font-mono">--</div>
                            </div>
                            <div class="bg-gray-800 p-2 rounded">
                                <div class="text-gray-500 text-xs">Species</div>
                                <div class="text-yellow-400 font-mono">--</div>
                            </div>
                            <div class="bg-gray-800 p-2 rounded">
                                <div class="text-gray-500 text-xs">
                                    Temperature
                                </div>
                                <div class="text-red-400 font-mono">--°C</div>
                            </div>
                            <div class="bg-gray-800 p-2 rounded">
                                <div class="text-gray-500 text-xs">CO2</div>
                                <div class="text-blue-400 font-mono">
                                    -- ppm
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="flex-1 flex flex-col overflow-hidden">
                        <h3 class="font-bold text-gray-300 p-4 pb-2">
                            Event Log
                        </h3>
                        <div
                            class="flex-1 overflow-y-auto p-4 pt-0 space-y-2 font-mono text-xs"
                        >
                            {#each simStats.events as event}
                                <div
                                    class="text-gray-400 border-l-2 border-gray-700 pl-2 py-1"
                                >
                                    {event}
                                </div>
                            {:else}
                                <div
                                    class="text-gray-600 italic text-center mt-10"
                                >
                                    No recent events
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}
