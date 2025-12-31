<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { MapRenderer } from "./MapRenderer";
    import { WebGLMapRenderer } from "./WebGLMapRenderer";
    import { parseBinaryGridData, MapDataLayer } from "./BinaryDataParser";
    import WorldMapLegend from "./WorldMapLegend.svelte";
    import MapOverlayCanvas from "./MapOverlayCanvas.svelte";
    import BabylonGlobe from "./BabylonGlobe.svelte";
    import { fade, fly } from "svelte/transition";
    import { mapStore } from "$lib/stores/map";
    import { gameWebSocket } from "$lib/services/websocket";
    import type { OverlayMode } from "$lib/types/overlays";
    import type { VisibleTile } from "./MapRenderer";

    export let isOpen = false;
    export let onClose: () => void;

    let canvas: HTMLCanvasElement;
    let renderer: MapRenderer | null = null;
    let webglRenderer: WebGLMapRenderer | null = null;
    let containerWidth = 0;
    let containerHeight = 0;
    let worldMapData: any = null;
    let loading = false;
    let mapDataLayer: MapDataLayer | null = null; // Binary grid data for tooltips

    // Hover state for tile inspection
    let hoveredTile: { gridX: number; gridY: number } | null = null;
    let hoverScreenX = 0;
    let hoverScreenY = 0;
    let tooltipData: {
        x: number;
        y: number;
        elevation: number;
        biome: string;
        customText?: string;
    } | null = null;

    // WebGL camera state
    let cameraZoom = 1.0;
    let cameraX = 0.5;
    let cameraY = 0.5;
    let isDragging = false;
    let lastMouseX = 0;
    let lastMouseY = 0;

    // Graphics mode toggle (WebGL vs ASCII)
    let useGraphicsMode = true;

    // Globe view toggle (3D sphere vs 2D flat map)
    let useGlobeView = false;
    let globeTextureBlob: Blob | null = null;

    // Overlay state
    let activeLayers: Set<OverlayMode> = new Set();
    let showMineralsOverlay = false; // Kept as separate toggle for resources (legacy? or just add to set?)
    // Actually, let's keep showMineralsOverlay for now but sync it or just use the set?
    // The previous code had it separate. Let's just use the Set for everything in the future,
    // but for now, let's keep the existing `showMineralsOverlay` binding if it was used elsewhere.
    // However, the previous `MapOverlayCanvas` usage passed `showMinerals`.
    // My new `MapOverlayCanvas` respects `activeLayers.has("resources")`.

    // Simulation stats (from events or world data)
    let simStats = {
        year: 0,
        population: 0,
        species: 0,
        events: [] as string[],
    };

    // Derived stats for HUD
    $: displayStats = {
        age: worldMapData?.simulated_years
            ? (worldMapData.simulated_years / 1_000_000).toFixed(1) + "M Years"
            : simStats.year > 0
              ? `Year ${simStats.year}`
              : "--",
        temp:
            worldMapData?.avg_temperature !== undefined
                ? worldMapData.avg_temperature.toFixed(1) + "°C"
                : "--",
        elev:
            worldMapData?.max_elevation !== undefined
                ? (worldMapData.max_elevation / 1000).toFixed(1) + "km"
                : "--",
        sea:
            worldMapData?.sea_level !== undefined
                ? worldMapData.sea_level.toFixed(0) + "m"
                : "--",
        land:
            worldMapData?.land_coverage !== undefined
                ? worldMapData.land_coverage.toFixed(1) + "%"
                : "--",
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
        mapDataLayer = null; // Clear tooltip data layer
    }

    function requestWorldMap(highRes = false) {
        if (!highRes) {
            loading = true;
        } else {
            // Background loading for 4K
            console.log("[WorldMap] Proceeding to load 4K background map...");
        }

        const payload = highRes ? { width: 4096, height: 2048 } : {}; // Empty = defaults (2048x1024)

        // Send command to request world map data
        gameWebSocket.sendRawCommand("world_map_image", payload);

        // Timeout logic only for initial load
        if (!highRes) {
            setTimeout(() => {
                if (!worldMapData) {
                    loading = false;
                }
            }, 3000);
        }
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
            // Fit to world view initially
            webglRenderer.fitToWorld();
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

        // Handle World Map Image Response (Phase 2 Integration)
        if (msg.type === "world_map_image_response") {
            const payload = msg.data;
            const isHighRes = payload.width > 2048;
            const isInitialLoad = !worldMapData;
            worldMapData = payload; // MUST set this to track state
            console.log(
                `[WorldMapModal] Received world_map_image_response (${payload.width}x${payload.height}) blob size:`,
                payload.imageBlob.size,
            );

            // Store blob for globe view (Babylon.js)
            if (payload.imageBlob) {
                globeTextureBlob = payload.imageBlob;
            }

            if (
                useGraphicsMode &&
                webglRenderer &&
                payload.imageBlob &&
                !useGlobeView
            ) {
                // Update metadata (sets grid/world size, player pos)
                webglRenderer.updateData(payload);
                // Then upload image texture override
                webglRenderer.updateTextureFromBlob(payload.imageBlob);

                // Parse binary grid data for tooltips (Sprint 2)
                if (payload.gridData && payload.gridData.byteLength > 0) {
                    const parsed = parseBinaryGridData(payload.gridData);
                    if (parsed) {
                        mapDataLayer = new MapDataLayer(parsed);
                        console.log(
                            `[WorldMapModal] MapDataLayer ready: ${parsed.width}x${parsed.height}`,
                        );
                    }
                }

                // Only auto-fit on the very first successful load
                if (isInitialLoad) {
                    webglRenderer.fitToWorld();
                }
            }

            loading = false;

            // Progressive Loading: If this was the initial low-res load, immediately kickoff 4K load
            if (!isHighRes) {
                console.log(
                    "[WorldMapModal] Initial map loaded. Requesting 4K background load...",
                );
                requestWorldMap(true);
            }
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

        // Subscribe to reconnection events to refresh map after connection recovery
        const unsubscribeReconnect = gameWebSocket.onReconnect(() => {
            if (isOpen && worldMapData) {
                console.log(
                    "[WorldMapModal] Reconnected, refreshing world map...",
                );
                requestWorldMap();
            }
        });

        // Also add window key listeners for controls when open
        window.addEventListener("keydown", handleKeydown);
        window.addEventListener("wheel", handleWheel);

        return () => {
            unsubscribe();
            unsubscribeReconnect();
            window.removeEventListener("keydown", handleKeydown);
            window.removeEventListener("wheel", handleWheel);
        };
    });

    function handleKeydown(e: KeyboardEvent) {
        if (!isOpen) return;

        // WebGL mode: handle camera
        if (useGraphicsMode && webglRenderer) {
            const panDelta = 0.05 * cameraZoom;
            switch (e.key.toLowerCase()) {
                case "w":
                    cameraY -= panDelta;
                    webglRenderer.setCamera(cameraX, cameraY, cameraZoom);
                    syncCameraState();
                    break;
                case "s":
                    cameraY += panDelta;
                    webglRenderer.setCamera(cameraX, cameraY, cameraZoom);
                    syncCameraState();
                    break;
                case "a":
                    cameraX -= panDelta;
                    webglRenderer.setCamera(cameraX, cameraY, cameraZoom);
                    syncCameraState();
                    break;
                case "d":
                    cameraX += panDelta;
                    webglRenderer.setCamera(cameraX, cameraY, cameraZoom);
                    syncCameraState();
                    break;
            }
            return;
        }

        // Canvas2D mode
        if (!renderer) return;
        const speed = 20;
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
        if (!isOpen) return;

        // WebGL mode
        if (useGraphicsMode && webglRenderer) {
            const zoomDelta = e.deltaY > 0 ? 1.1 : 0.9;
            cameraZoom = Math.max(0.1, Math.min(10.0, cameraZoom * zoomDelta));
            webglRenderer.setCamera(cameraX, cameraY, cameraZoom);
            syncCameraState();
            return;
        }

        // Canvas2D mode
        if (!renderer) return;
        renderer.zoom(e.deltaY);
    }

    function syncCameraState() {
        if (!webglRenderer) return;
        const pos = webglRenderer.getCameraPosition();
        cameraX = pos.x;
        cameraY = pos.y;
        cameraZoom = webglRenderer.getZoom();
    }

    // Hover handling for tooltip
    function handleMapMouseMove(e: MouseEvent) {
        if (!useGraphicsMode || !webglRenderer || !worldMapData) return;

        // Handle dragging
        if (isDragging) {
            const deltaX = e.clientX - lastMouseX;
            const deltaY = e.clientY - lastMouseY;
            lastMouseX = e.clientX;
            lastMouseY = e.clientY;

            const texDeltaX = (-deltaX / containerWidth) * cameraZoom;
            const texDeltaY = (-deltaY / containerHeight) * cameraZoom;

            cameraX += texDeltaX;
            cameraY += texDeltaY;
            webglRenderer.setCamera(cameraX, cameraY, cameraZoom);
            syncCameraState();
            return;
        }

        // Get grid position under mouse
        const rect = canvas.getBoundingClientRect();
        const mouseX = e.clientX - rect.left;
        const mouseY = e.clientY - rect.top;

        hoverScreenX = e.clientX;
        hoverScreenY = e.clientY;

        const gridPos = webglRenderer.getGridIndexFromScreen(mouseX, mouseY);
        if (!gridPos) {
            hoveredTile = null;
            tooltipData = null;
            return;
        }

        hoveredTile = gridPos;

        // Use MapDataLayer (binary grid) if available - more accurate for high-res mode
        if (mapDataLayer) {
            // gridPos already contains grid indices from getGridIndexFromScreen
            // which accounts for zoom/pan camera transformations
            const elevation = mapDataLayer.getElevation(
                gridPos.gridX,
                gridPos.gridY,
            );
            const biome = mapDataLayer.getBiomeName(
                gridPos.gridX,
                gridPos.gridY,
            );

            tooltipData = {
                x: gridPos.gridX,
                y: gridPos.gridY,
                elevation: elevation,
                biome: biome,
            };
            return;
        }

        // Fallback: Find tile data from JSON tiles (legacy mode)
        let tile: any = null;
        if (worldMapData.tiles) {
            tile = worldMapData.tiles.find(
                (t: any) =>
                    t.grid_x === gridPos.gridX && t.grid_y === gridPos.gridY,
            );
        }

        if (tile) {
            tooltipData = {
                x: gridPos.gridX,
                y: gridPos.gridY,
                elevation: tile.avg_elevation || 0,
                biome: tile.biome || "Unknown",
            };
        } else {
            tooltipData = null;
        }
    }

    function handleMapMouseDown(e: MouseEvent) {
        if (!useGraphicsMode || e.button !== 0) return;
        isDragging = true;
        lastMouseX = e.clientX;
        lastMouseY = e.clientY;
    }

    function handleMapMouseUp() {
        isDragging = false;
    }

    function handleMapMouseLeave() {
        isDragging = false;
        hoveredTile = null;
        tooltipData = null;
    }

    function handleLayerToggle(id: string) {
        toggleLayer(id);
    }

    function toggleLayer(id: string) {
        if (id === "none") {
            activeLayers.clear();
            activeLayers = activeLayers;
            showMineralsOverlay = false;
            return;
        }

        const mode = id as OverlayMode;
        if (activeLayers.has(mode)) {
            activeLayers.delete(mode);
        } else {
            activeLayers.add(mode);
        }
        activeLayers = activeLayers; // Trigger reactivity
    }

    function handleOverlayHover(e: CustomEvent) {
        const hit = e.detail; // ResourceNode | null
        if (!hit) {
            // If we are inspecting a tile via mouse move, keep showing that?
            // Or should overlay tooltip take precedence?
            // Current `handleMapMouseMove` sets `tooltipData`.
            // If `hit` is null, we do nothing and let `handleMapMouseMove` control it (tile info).
            // But if `hit` is present, we should override `tooltipData` with overlay info.
            if (!webglRenderer) {
                tooltipData = null;
            }
            return;
        }

        // Create tooltip data from ResourceNode
        // We need screen coordinates. `handleMapMouseMove` gives us mouse coords.
        // But `MapOverlayCanvas` doesn't pass screen coords in the event, only data.
        // Wait, `handleMouseMove` in MapOverlayCanvas tracks `hoveredItem`.

        // Actually, let's update `tooltipData` ONLY if we have a hit.
        // We'll need a way to distinguish Tile Tooltip vs Overlay Tooltip.
        // Let's add specific fields for Overlay info.

        let label = "Unknown Feature";
        if (hit.type === "volcano") label = "Volcano 🌋";
        else if (hit.type === "peak") label = "Peak 🏔️";
        else if (hit.type === "trench") label = "Trench 🕳️";
        else if (hit.type === "gold") label = "Gold Deposit 🟡";
        else if (hit.type === "iron") label = "Iron Deposit ⚪";
        else if (hit.type === "coal") label = "Coal Deposit ⚫";
        else if (hit.type === "cave") label = "Cave Entrance 🕳️";

        // Enrich with `hit.data` if available
        let details = "";
        if (hit.data) {
            if (hit.data.height)
                details += `Height: ${hit.data.height.toFixed(0)}m\n`;
            if (hit.data.depth)
                details += `Depth: ${hit.data.depth.toFixed(0)}m\n`;
            if (hit.data.pressure)
                details += `Pressure: ${(hit.data.pressure * 100).toFixed(1)}%\n`;
            if (hit.data.age)
                details += `Age: ${(hit.data.age / 1000000).toFixed(1)}M yr\n`;
        }

        // We can hijack tooltipData or genericize it.
        // Existing tooltipData: { x, y, elevation, biome }
        // Let's force it.
        tooltipData = {
            x: Math.round(hit.x), // ResourceNode has x/y in World Coords? Or Grid?
            // ResourceNode x/y are grid coords usually.
            y: Math.round(hit.y),
            biome: label,
            elevation: hit.val ? hit.val * 3000 : 0, // Mock elevation from val if needed, or just display raw
            // We really should change the Tooltip UI to be more generic.
            // For now, let's just piggyback.
            customText: details, // Add this field to tooltip handling?
        } as any;
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
                <div class="flex gap-4 items-center">
                    <!-- View Toggle: Globe / Map -->
                    <button
                        on:click={() => (useGlobeView = !useGlobeView)}
                        class="px-3 py-1 rounded text-sm font-medium transition-colors
                               {useGlobeView
                            ? 'bg-blue-600 text-white'
                            : 'bg-gray-700 text-gray-300 hover:bg-gray-600'}"
                    >
                        {useGlobeView ? "🌍 Globe" : "🗺️ Map"}
                    </button>
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
                <div
                    class="flex-1 bg-black relative"
                    bind:clientWidth={containerWidth}
                    bind:clientHeight={containerHeight}
                >
                    {#if useGlobeView}
                        <!-- 3D Globe View (Babylon.js) -->
                        <BabylonGlobe textureBlob={globeTextureBlob} />
                    {:else}
                        <!-- 2D Flat Map View -->
                        <canvas
                            bind:this={canvas}
                            width={containerWidth}
                            height={containerHeight}
                            class="block w-full h-full cursor-grab"
                            class:cursor-grabbing={isDragging}
                            on:mousemove={handleMapMouseMove}
                            on:mousedown={handleMapMouseDown}
                            on:mouseup={handleMapMouseUp}
                            on:mouseleave={handleMapMouseLeave}
                        ></canvas>
                    {/if}

                    <!-- Overlay Canvas (Tectonics / Minerals / Env) -->
                    {#if worldMapData?.overlays && (activeLayers.size > 0 || showMineralsOverlay)}
                        <MapOverlayCanvas
                            width={containerWidth}
                            height={containerHeight}
                            gridWidth={worldMapData.grid_width}
                            gridHeight={worldMapData.grid_height}
                            overlayData={worldMapData.overlays}
                            {activeLayers}
                            showMinerals={showMineralsOverlay}
                            {cameraX}
                            {cameraY}
                            zoom={cameraZoom}
                            on:hover={handleOverlayHover}
                        />
                    {/if}

                    <!-- Tile Inspection Tooltip -->
                    {#if tooltipData}
                        <div
                            class="fixed z-50 pointer-events-none"
                            style="left: {hoverScreenX +
                                16}px; top: {hoverScreenY + 16}px;"
                        >
                            <div
                                class="bg-gray-900/95 border border-gray-600 rounded-lg p-3 shadow-xl text-sm min-w-[150px]"
                            >
                                <div
                                    class="text-gray-400 text-xs mb-2 font-mono"
                                >
                                    ({tooltipData.x}, {tooltipData.y})
                                </div>
                                <div class="space-y-1">
                                    <div class="flex justify-between">
                                        <span class="text-gray-400">Biome</span>
                                        <span class="text-white font-medium"
                                            >{tooltipData.biome}</span
                                        >
                                    </div>
                                    <div class="flex justify-between">
                                        <span class="text-gray-400"
                                            >Elevation</span
                                        >
                                        <span class="text-yellow-400"
                                            >{tooltipData.elevation.toFixed(
                                                0,
                                            )}m</span
                                        >
                                    </div>
                                    {#if tooltipData.customText}
                                        <div
                                            class="text-xs text-blue-300 whitespace-pre-line mt-2 pt-2 border-t border-gray-700"
                                        >
                                            {tooltipData.customText}
                                        </div>
                                    {/if}
                                </div>
                            </div>
                        </div>
                    {/if}

                    <!-- Stats HUD (Top Left) -->
                    <div
                        class="absolute top-4 left-4 p-4 rounded-lg bg-gray-900/80 backdrop-blur border border-gray-700 shadow-xl min-w-[200px]"
                        transition:fade
                    >
                        <h3
                            class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-3 border-b border-gray-700 pb-2"
                        >
                            Planetary Data
                        </h3>

                        <div class="space-y-3 font-mono text-sm">
                            <div class="flex justify-between items-center">
                                <span class="text-gray-400">Age</span>
                                <span class="text-purple-400 font-bold"
                                    >{displayStats.age}</span
                                >
                            </div>
                            <div class="flex justify-between items-center">
                                <span class="text-gray-400">Avg Temp</span>
                                <span class="text-red-400 font-bold"
                                    >{displayStats.temp}</span
                                >
                            </div>
                            <div class="flex justify-between items-center">
                                <span class="text-gray-400">Max Elev</span>
                                <span class="text-yellow-400"
                                    >{displayStats.elev}</span
                                >
                            </div>
                            <div class="flex justify-between items-center">
                                <span class="text-gray-400">Sea Level</span>
                                <span class="text-blue-400"
                                    >{displayStats.sea}</span
                                >
                            </div>
                            <div class="flex justify-between items-center">
                                <span class="text-gray-400">Land Mass</span>
                                <span class="text-green-400"
                                    >{displayStats.land}</span
                                >
                            </div>

                            {#if worldMapData?.seed}
                                <div
                                    class="pt-2 mt-2 border-t border-gray-700 text-xs flex justify-between"
                                >
                                    <span class="text-gray-500">Seed</span>
                                    <span class="text-gray-400"
                                        >{worldMapData.seed}</span
                                    >
                                </div>
                            {/if}
                        </div>
                    </div>

                    <!-- Legend (Bottom Left) -->
                    <div class="absolute bottom-4 left-4" transition:fade>
                        <WorldMapLegend
                            mode={(activeLayers.size === 0
                                ? worldMapData?.is_simulated
                                    ? "terrain"
                                    : "biome"
                                : Array.from(activeLayers).pop()) || "terrain"}
                            {activeLayers}
                        />
                    </div>

                    <!-- Overlay Controls (Bottom Right) -->
                    <div
                        class="absolute bottom-4 right-4 flex flex-col gap-2 items-end"
                    >
                        <!-- Overlay Toggles -->
                        {#if worldMapData?.overlays}
                            <div
                                class="bg-gray-800/90 p-3 rounded-lg border border-gray-700 space-y-3 min-w-[200px]"
                            >
                                <div
                                    class="text-xs text-gray-400 font-bold uppercase border-b border-gray-700 pb-1"
                                >
                                    Data Layers
                                </div>

                                <!-- Layer Selection (Radio-like behavior) -->
                                <div class="space-y-1">
                                    {#each [{ id: "none", label: "Clear All", icon: "🚫" }, { id: "tectonics", label: "Tectonics", icon: "📐" }, { id: "elevation", label: "Elevation", icon: "🏔️" }, { id: "temp", label: "Temperature", icon: "🌡️" }, { id: "moisture", label: "Moisture", icon: "💧" }, { id: "biome", label: "Biomes", icon: "🌿" }, { id: "features", label: "Terrain Features", icon: "📍" }] as layer}
                                        <button
                                            class="w-full text-left px-2 py-1.5 rounded text-xs flex items-center justify-between transition-colors {activeLayers.has(
                                                layer.id,
                                            ) ||
                                            (layer.id === 'none' &&
                                                activeLayers.size === 0)
                                                ? 'bg-blue-600/30 text-blue-200 border border-blue-500/30'
                                                : 'hover:bg-gray-700 text-gray-300'}"
                                            on:click={() =>
                                                handleLayerToggle(layer.id)}
                                        >
                                            <span
                                                class="flex items-center gap-2"
                                            >
                                                <span>{layer.icon}</span>
                                                {layer.label}
                                            </span>
                                            {#if activeLayers.has(layer.id)}
                                                <span
                                                    class="w-1.5 h-1.5 rounded-full bg-blue-400"
                                                ></span>
                                            {/if}
                                        </button>
                                    {/each}
                                </div>

                                <!-- Resources Toggle (Independent) -->
                                {#if worldMapData.overlays.resources || worldMapData.overlays.minerals}
                                    <div class="pt-2 border-t border-gray-700">
                                        <label
                                            class="flex items-center gap-2 cursor-pointer hover:bg-gray-700/50 px-2 py-1 rounded transition-colors group"
                                        >
                                            <input
                                                type="checkbox"
                                                bind:checked={
                                                    showMineralsOverlay
                                                }
                                                class="w-4 h-4 accent-yellow-500 rounded border-gray-600 bg-gray-700"
                                            />
                                            <div class="flex flex-col">
                                                <span
                                                    class="text-xs text-gray-300 group-hover:text-white"
                                                    >Show Resources</span
                                                >
                                                {#if worldMapData.overlays.resources}
                                                    <span
                                                        class="text-[10px] text-gray-500"
                                                    >
                                                        {worldMapData.overlays
                                                            .resources.length} nodes
                                                    </span>
                                                {/if}
                                            </div>
                                        </label>
                                    </div>
                                {/if}
                            </div>
                        {/if}
                        <div
                            class="bg-gray-800/80 p-2 rounded text-xs text-gray-300"
                        >
                            WASD to Pan • Scroll to Zoom
                        </div>
                    </div>
                </div>

                <!-- Sidebar / Event Log Only -->
                <div
                    class="w-72 bg-gray-850 border-l border-gray-800 flex flex-col"
                >
                    <!-- Natural Satellites Section -->
                    {#if worldMapData?.satellites?.length > 0}
                        <div class="p-4 border-b border-gray-800">
                            <h3
                                class="font-bold text-gray-300 mb-3 flex items-center gap-2"
                            >
                                <span class="text-lg">🌙</span>
                                Natural Satellites
                            </h3>
                            <div class="space-y-2 text-sm">
                                {#each worldMapData.satellites as sat}
                                    <div
                                        class="flex justify-between items-center"
                                    >
                                        <span class="text-gray-300"
                                            >{sat.name}</span
                                        >
                                        <span class="text-gray-500 text-xs">
                                            {sat.mass.toFixed(2)}x Luna
                                        </span>
                                    </div>
                                {/each}
                            </div>
                            <!-- Climate Stability -->
                            <div
                                class="mt-3 pt-3 border-t border-gray-700 text-xs"
                            >
                                <div class="flex justify-between">
                                    <span class="text-gray-500"
                                        >Climate Stability</span
                                    >
                                    <span
                                        class={worldMapData.satellites.reduce(
                                            (a, s) => a + s.mass,
                                            0,
                                        ) > 0.01
                                            ? "text-green-400"
                                            : "text-yellow-400"}
                                    >
                                        {worldMapData.satellites.reduce(
                                            (a, s) => a + s.mass,
                                            0,
                                        ) > 0.01
                                            ? "Stable"
                                            : "Variable"}
                                    </span>
                                </div>
                                <div class="flex justify-between mt-1">
                                    <span class="text-gray-500"
                                        >Impact Shield</span
                                    >
                                    <span class="text-blue-400">
                                        {Math.min(
                                            worldMapData.satellites.length * 5,
                                            20,
                                        )}%
                                    </span>
                                </div>
                            </div>
                        </div>
                    {:else}
                        <div class="p-4 border-b border-gray-800">
                            <h3
                                class="font-bold text-gray-300 mb-2 flex items-center gap-2"
                            >
                                <span class="text-lg">🌙</span>
                                Natural Satellites
                            </h3>
                            <div class="text-gray-500 text-sm italic">
                                No moons detected
                            </div>
                            <div class="mt-2 text-xs flex justify-between">
                                <span class="text-gray-500"
                                    >Climate Stability</span
                                >
                                <span class="text-red-400">Chaotic</span>
                            </div>
                        </div>
                    {/if}

                    <!-- Event Log -->
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
