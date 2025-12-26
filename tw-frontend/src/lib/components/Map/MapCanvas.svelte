<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { WebGLMapRenderer } from "./WebGLMapRenderer";
    import type { WorldMapData } from "./WebGLMapRenderer";
    import MapEntityOverlay from "./MapEntityOverlay.svelte";
    import type { MapEntity } from "./MapEntityOverlay.svelte";

    export let playerPosition: { x: number; y: number } = { x: 0, y: 0 };
    export let mapData: WorldMapData | null = null;
    export let entities: MapEntity[] = [];
    export let width: number = 400;
    export let height: number = 400;

    let canvas: HTMLCanvasElement;
    let renderer: WebGLMapRenderer | null = null;

    // Camera state (for overlay synchronization)
    let zoom = 1.0;
    let cameraX = 0.5;
    let cameraY = 0.5;
    let texScaleX = 1.0;
    let texScaleY = 1.0;
    let gridWidth = 128;
    let gridHeight = 64;

    // Drag state
    let isDragging = false;
    let lastMouseX = 0;
    let lastMouseY = 0;

    onMount(() => {
        renderer = new WebGLMapRenderer(canvas);
        renderer.start();

        if (mapData) {
            renderer.updateData(mapData);
            renderer.fitToWorld();
            syncCameraState();
        }
    });

    onDestroy(() => {
        if (renderer) {
            renderer.stop();
            renderer.destroy();
        }
    });

    /**
     * Sync camera state from renderer to overlay
     */
    function syncCameraState() {
        if (!renderer) return;

        const pos = renderer.getCameraPosition();
        const scale = renderer.getTexScale();
        const grid = renderer.getGridDimensions();

        cameraX = pos.x;
        cameraY = pos.y;
        texScaleX = scale.x;
        texScaleY = scale.y;
        gridWidth = grid.width;
        gridHeight = grid.height;
        zoom = renderer.getZoom();
    }

    // React to data changes
    $: if (renderer && mapData) {
        renderer.updateData(mapData);
        renderer.fitToWorld();
        syncCameraState();
    }

    // Zoom handler (wheel)
    function handleWheel(e: WheelEvent) {
        e.preventDefault();
        if (!renderer) return;

        // Zoom in/out based on scroll direction
        const zoomDelta = e.deltaY > 0 ? 1.1 : 0.9;
        zoom = Math.max(0.1, Math.min(10.0, zoom * zoomDelta));

        // Apply new zoom
        renderer.setCamera(cameraX, cameraY, zoom);
        syncCameraState();
    }

    // Pan handlers (mouse drag)
    function handleMouseDown(e: MouseEvent) {
        if (e.button !== 0) return; // Left click only
        isDragging = true;
        lastMouseX = e.clientX;
        lastMouseY = e.clientY;
        canvas.style.cursor = "grabbing";
    }

    function handleMouseMove(e: MouseEvent) {
        if (!isDragging || !renderer) return;

        const deltaX = e.clientX - lastMouseX;
        const deltaY = e.clientY - lastMouseY;
        lastMouseX = e.clientX;
        lastMouseY = e.clientY;

        // Convert pixel delta to texture coordinate delta
        const texDeltaX = (-deltaX / canvas.width) * renderer.getZoom();
        const texDeltaY = (-deltaY / canvas.height) * renderer.getZoom();

        cameraX += texDeltaX;
        cameraY += texDeltaY;

        renderer.setCamera(cameraX, cameraY, zoom);
        syncCameraState();
    }

    function handleMouseUp() {
        isDragging = false;
        canvas.style.cursor = "grab";
    }

    function handleMouseLeave() {
        isDragging = false;
        canvas.style.cursor = "default";
    }

    function handleMouseEnter() {
        canvas.style.cursor = "grab";
    }
</script>

<div class="map-container" style="width: {width}px; height: {height}px;">
    <canvas
        bind:this={canvas}
        {width}
        {height}
        class="map-canvas"
        on:wheel|preventDefault={handleWheel}
        on:mousedown={handleMouseDown}
        on:mousemove={handleMouseMove}
        on:mouseup={handleMouseUp}
        on:mouseleave={handleMouseLeave}
        on:mouseenter={handleMouseEnter}
    ></canvas>

    <MapEntityOverlay
        {width}
        {height}
        {cameraX}
        {cameraY}
        {texScaleX}
        {texScaleY}
        {gridWidth}
        {gridHeight}
        {entities}
        playerX={playerPosition.x}
        playerY={playerPosition.y}
    />
</div>

<style>
    .map-container {
        position: relative;
        display: inline-block;
    }

    .map-canvas {
        border: 1px solid #374151;
        border-radius: 4px;
        background: #000;
        cursor: grab;
        display: block;
    }

    .map-canvas:active {
        cursor: grabbing;
    }
</style>
