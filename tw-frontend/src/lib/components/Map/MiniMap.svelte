<script lang="ts">
    import { onDestroy, onMount } from "svelte";
    import { WebGLMapRenderer } from "./WebGLMapRenderer";
    import type { MiniMapData, MiniMapTile } from "./WebGLMapRenderer";
    import { mapStore } from "$lib/stores/map";

    // Fixed size matching D-Pad (3x3 grid of 48px buttons + gaps)
    const MAP_SIZE = 152;

    let canvas: HTMLCanvasElement;
    let renderer: WebGLMapRenderer | null = null;

    $: if (renderer && $mapStore.data) {
        updateMapData();
    }

    onMount(() => {
        console.log("[MiniMap] v3.0 WebGL Mode");
        if (canvas) {
            renderer = new WebGLMapRenderer(canvas);
            renderer.start();

            if ($mapStore.data) {
                updateMapData();
            }
        }
    });

    onDestroy(() => {
        if (renderer) {
            renderer.stop();
            renderer.destroy();
        }
    });

    function updateMapData() {
        if (!renderer || !$mapStore.data) return;

        const data: MiniMapData = {
            tiles: $mapStore.data.tiles.map(
                (tile: any): MiniMapTile => ({
                    x: tile.x,
                    y: tile.y,
                    biome: tile.biome || "Default",
                    elevation: tile.elevation || 0,
                    is_player: tile.is_player || false,
                }),
            ),
            player_x: $mapStore.data.player_x,
            player_y: $mapStore.data.player_y,
            grid_size: $mapStore.data.grid_size || 9,
        };

        renderer.updateMiniMapData(data);
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
</style>
