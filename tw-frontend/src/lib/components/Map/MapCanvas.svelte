<script lang="ts">
    import { onMount } from "svelte";
    import { MapRenderer } from "./MapRenderer";
    import type { Position, VisibleTile } from "./MapRenderer";

    export let playerPosition: Position = { x: 0, y: 0 };
    export let visibleTiles: VisibleTile[] = [];
    export let width: number = 400;
    export let height: number = 400;

    let canvas: HTMLCanvasElement;
    let renderer: MapRenderer;

    onMount(() => {
        renderer = new MapRenderer(canvas);
        renderMap();
    });

    function renderMap() {
        if (renderer) {
            renderer.render(playerPosition, visibleTiles);
        }
    }

    $: if (renderer && (playerPosition || visibleTiles)) {
        renderMap();
    }
</script>

<canvas
    bind:this={canvas}
    {width}
    {height}
    class="border border-gray-700 rounded bg-black"
/>
