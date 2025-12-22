<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { MapRenderer } from "./MapRenderer";
    import { WebGLMapRenderer } from "./WebGLMapRenderer";
    import type { VisibleTile } from "./MapRenderer";
    import { mapStore } from "$lib/stores/map";
    import { gameWebSocket } from "$lib/services/websocket";
    import { fade } from "svelte/transition";

    export let isOpen = false;
    export let onClose: () => void;

    let canvas: HTMLCanvasElement;
    let renderer: MapRenderer | null = null;
    let webglRenderer: WebGLMapRenderer | null = null;
    let containerWidth = 0;
    let containerHeight = 0;
    let worldMapData: any = null;
    let loading = false;

    // Graphics mode toggle (WebGL vs ASCII)
    let useGraphicsMode = true;

    // Simulation stats
    let simStats = {
        year: 0,
        population: 0,
        species: 0,
        events: [] as string[],
    };

    $: if (isOpen && canvas) {
        initRenderer();
        requestWorldMap();
    }

    $: if (!isOpen) {
        cleanupRenderers();
    }

    // Update map when world map data is received
    $: if ((renderer || webglRenderer) && worldMapData && isOpen) {
        updateWorldMap();
    }

    // Fallback: Use minimap store only if world map data never arrives
    // Wait until loading is complete before falling back
    $: if (renderer && !worldMapData && !loading && $mapStore.data && isOpen) {
        updateFromMinimap();
    }

    function cleanupRenderers() {
        if (renderer) {
            renderer.stop();
            renderer = null;
        }
        if (webglRenderer) {
            webglRenderer.stop();
            webglRenderer.destroy();
            webglRenderer = null;
        }
        worldMapData = null;
    }

    function requestWorldMap() {
        loading = true;
        // Send command to request world map data
        gameWebSocket.sendRawCommand("world map");

        // Timeout: if world_map_data doesn't arrive in 3 seconds, stop loading
        setTimeout(() => {
            if (!worldMapData) {
                loading = false;
            }
        }, 3000);
    }

    function initRenderer() {
        if (!canvas) return;

        // Cleanup existing renderers
        cleanupRenderers();

        if (useGraphicsMode) {
            // Use WebGL renderer for graphics mode
            webglRenderer = new WebGLMapRenderer(canvas);
            webglRenderer.start();
            console.log("[WorldMapModal] Using WebGL graphics mode");
        } else {
            // Use Canvas 2D renderer for text mode
            const ctx = canvas.getContext("2d", { alpha: false });
            if (!ctx) return;

            renderer = new MapRenderer(canvas);
            renderer.setTileSize(4);
            renderer.setViewMode("atlas");
            renderer.setQuality("low");
            renderer.start();
            console.log("[WorldMapModal] Using Canvas 2D text mode");
        }
    }

    // Update from full world map data (Issue 5)
    function updateWorldMap() {
        if (!worldMapData) return;

        // Graphics mode: use WebGL renderer
        if (useGraphicsMode && webglRenderer) {
            webglRenderer.updateData(worldMapData);
            return;
        }

        // Text mode: use Canvas 2D renderer
        if (!renderer) return;

        // Convert player world position to grid position
        const gridWidth = worldMapData.grid_width || 128;
        const gridHeight = worldMapData.grid_height || 64;
        const worldWidth = worldMapData.world_width || 1;
        const worldHeight = worldMapData.world_height || 1;

        // Player position in grid coordinates (not world coordinates)
        const playerGridX = (worldMapData.player_x / worldWidth) * gridWidth;
        const playerGridY = (worldMapData.player_y / worldHeight) * gridHeight;

        const playerPos = {
            x: Math.round(playerGridX),
            y: Math.round(playerGridY),
            z: 0,
        };

        // World size for the renderer is the grid size
        renderer.setWorldSize(Math.max(gridWidth, gridHeight));

        // Convert WorldMapTile to VisibleTile format
        // Use grid coordinates directly - each tile is one grid cell
        const visibleTiles: VisibleTile[] = worldMapData.tiles.map(
            (tile: any) => {
                const vt: VisibleTile = {
                    x: tile.grid_x, // Use grid coordinates directly
                    y: tile.grid_y, // Use grid coordinates directly
                    biome: tile.biome || "Default",
                    elevation: tile.avg_elevation || 0,
                    entities: [],
                };
                if (tile.is_player) vt.is_player = true;
                return vt;
            },
        );

        loading = false;
        renderer.updateData(playerPos, visibleTiles, 1.0);
    }

    // Fallback: Update from minimap data
    function updateFromMinimap() {
        if (!renderer || !$mapStore.data) return;

        console.log(
            "[WorldMapModal] FALLBACK: Using minimap data (no world_map_data received)",
        );

        const playerPos = {
            x: Math.round($mapStore.data.player_x),
            y: Math.round($mapStore.data.player_y),
            z: Math.round($mapStore.data.player_z || 0),
        };

        if ($mapStore.data.grid_size) {
            renderer.setWorldSize($mapStore.data.grid_size);
        }

        const visibleTiles: VisibleTile[] = $mapStore.data.tiles.map(
            (tile: any) => {
                const vt: VisibleTile = {
                    x: tile.x,
                    y: tile.y,
                    biome: tile.biome || "Default",
                    elevation: tile.elevation || 0,
                    entities: tile.entities || [],
                };
                if (tile.is_player) vt.is_player = true;
                if (tile.portal) vt.portal = tile.portal;
                if (tile.occluded) vt.occluded = true;
                return vt;
            },
        );

        renderer.updateData(playerPos, visibleTiles, 1.0);
    }

    // Listen for sim events and world map data
    function handleSimMessage(msg: any) {
        // Backend sends: { type: "game_message", data: { type: "world_map_data", metadata: {...} } }
        const dataType = msg.data?.type || msg.type;

        // Handle world map data from backend
        if (dataType === "world_map_data") {
            // Payload is in metadata for game_message, or directly in data for other message types
            const payload = msg.data?.metadata || msg.data;
            console.log("[WorldMapModal] Received world_map_data:", {
                tiles: payload?.tiles?.length,
                grid: `${payload?.grid_width}x${payload?.grid_height}`,
                worldSize: `${payload?.world_width}x${payload?.world_height}`,
                biomes: [
                    ...new Set(payload?.tiles?.map((t: any) => t.biome) || []),
                ].slice(0, 10),
            });
            worldMapData = payload;
            loading = false;
            return;
        }

        if (dataType === "sim_event") {
            simStats.year = msg.data.year || simStats.year;
            // Add to event log
            const eventText = `Year ${msg.data.year}: ${msg.text || msg.data?.text}`;
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
        // Subscribe to messages for sim stats
        const unsubscribe = gameWebSocket.onMessage(handleSimMessage);

        // Also add window key listeners for controls when open
        window.addEventListener("keydown", handleKeydown);
        window.addEventListener("wheel", handleWheel);

        return () => {
            unsubscribe();
            window.removeEventListener("keydown", handleKeydown);
            window.removeEventListener("wheel", handleWheel);
        };
    });

    function handleKeydown(e: KeyboardEvent) {
        if (!isOpen || !renderer) return;

        const speed = 20; // Pan speed
        switch (e.key.toLowerCase()) {
            case "w":
                renderer.pan(0, speed);
                break;
            case "s":
                renderer.pan(0, -speed);
                break;
            case "a":
                renderer.pan(-speed, 0);
                break;
            case "d":
                renderer.pan(speed, 0);
                break;
        }
    }

    function handleWheel(e: WheelEvent) {
        if (!isOpen || !renderer) return;
        renderer.zoom(e.deltaY);
    }

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
