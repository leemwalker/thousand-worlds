<script lang="ts">
    import { onDestroy, onMount } from "svelte";
    import { MapRenderer } from "./MapRenderer";
    import type { VisibleTile, RenderQuality } from "./MapRenderer";
    import { mapStore } from "$lib/stores/map";

    // Fixed size matching D-Pad (3x3 grid of 48px buttons + gaps)
    const MAP_SIZE = 152;

    // Tile size adjusts based on grid size (fractional for sub-pixel rendering at high altitude)
    $: gridSize = $mapStore.data?.grid_size || 9;
    $: tileSize = MAP_SIZE / gridSize;

    let canvas: HTMLCanvasElement;
    let renderer: MapRenderer | null = null;

    $: if (renderer && $mapStore.data) {
        console.log("[MiniMap] Data received:", {
            tiles: $mapStore.data.tiles?.length,
            quality: $mapStore.data.render_quality,
            grid_size: $mapStore.data.grid_size,
            player: { x: $mapStore.data.player_x, y: $mapStore.data.player_y },
            sample_tile: $mapStore.data.tiles?.[0],
        });
        renderer.setQuality($mapStore.data.render_quality || "low");
        renderer.setStyle("standard"); // Use biome-based rendering instead of geology
        renderer.setTileSize(tileSize);
        updateMapData();
    }

    onMount(() => {
        console.log("[MiniMap] v2.3 Debug Mode");
        if (canvas) {
            renderer = new MapRenderer(canvas);
            renderer.setTileSize(tileSize);
            renderer.start(); // Start render loop

            if ($mapStore.data) {
                renderer.setQuality($mapStore.data.render_quality || "low");
                renderer.setStyle("standard"); // Use biome-based rendering
                updateMapData();
            }
        }
    });

    onDestroy(() => {
        if (renderer) {
            renderer.stop();
        }
    });

    function updateMapData() {
        if (!renderer || !$mapStore.data) return;

        const playerPos = {
            x: Math.round($mapStore.data.player_x),
            y: Math.round($mapStore.data.player_y),
            z: Math.round($mapStore.data.player_z || 0), // Pass Player Z
        };

        // Convert tiles to VisibleTile format
        const visibleTiles: VisibleTile[] = $mapStore.data.tiles.map((tile) => {
            const vt: VisibleTile = {
                x: tile.x,
                y: tile.y,
                biome: tile.biome || "Default",
                elevation: tile.elevation || 0,
                entities: tile.entities || [],
                is_player: tile.is_player || false,
                occluded: tile.occluded || false,
            };
            if (tile.portal) vt.portal = tile.portal;
            return vt;
        });

        renderer.updateData(playerPos, visibleTiles, $mapStore.data.scale || 1);
    }

    function getQualityLabel(quality: RenderQuality | undefined): string {
        switch (quality) {
            case "high":
                return "◆◆◆";
            case "medium":
                return "◆◆○";
            default:
                return "◆○○";
        }
    }
</script>

<div class="mini-map-container" data-testid="mini-map">
    <canvas
        bind:this={canvas}
        width={MAP_SIZE}
        height={MAP_SIZE}
        class="map-canvas"
    />

    <!-- Compass indicator -->
    <div class="compass">N</div>

    <!-- Quality indicator -->
    <div class="quality-indicator" title="Perception Quality">
        {getQualityLabel($mapStore.data?.render_quality)}
    </div>
</div>

<style>
    .mini-map-container {
        position: relative;
        width: 160px;
        height: 160px;
        background: rgba(31, 41, 55, 0.9);
        border: 1px solid #374151;
        border-radius: 12px;
        padding: 4px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
    }

    .map-canvas {
        width: 100%;
        height: 100%;
        border-radius: 8px;
        image-rendering: pixelated;
        image-rendering: crisp-edges;
    }

    .compass {
        position: absolute;
        top: -8px;
        left: 50%;
        transform: translateX(-50%);
        background: rgba(31, 41, 55, 0.95);
        color: #9ca3af;
        font-size: 10px;
        font-weight: bold;
        padding: 2px 6px;
        border-radius: 4px;
        border: 1px solid #374151;
    }

    .quality-indicator {
        position: absolute;
        bottom: -6px;
        left: 50%;
        transform: translateX(-50%);
        background: rgba(31, 41, 55, 0.95);
        color: #ffd700;
        font-size: 8px;
        padding: 2px 6px;
        border-radius: 4px;
        border: 1px solid #374151;
        letter-spacing: 2px;
    }
</style>
