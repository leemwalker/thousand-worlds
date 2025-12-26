<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { WebGLMapRenderer } from "./WebGLMapRenderer";
    import type { WorldMapData } from "./WebGLMapRenderer";

    export let playerPosition: { x: number; y: number } = { x: 0, y: 0 };
    export let mapData: WorldMapData | null = null;
    export let width: number = 400;
    export let height: number = 400;

    let canvas: HTMLCanvasElement;
    let renderer: WebGLMapRenderer | null = null;

    // Camera state
    let zoom = 1.0;
    let cameraX = 0.5;
    let cameraY = 0.5;

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
        }
    });

    onDestroy(() => {
        if (renderer) {
            renderer.stop();
            renderer.destroy();
        }
    });

    // React to data changes
    $: if (renderer && mapData) {
        renderer.updateData(mapData);
        renderer.fitToWorld();
        // Reset camera when new data arrives
        zoom = 1.0;
        cameraX = 0.5;
        cameraY = 0.5;
    }

    // Zoom handler (wheel)
    function handleWheel(e: WheelEvent) {
        e.preventDefault();
        if (!renderer) return;

        // Zoom in/out based on scroll direction
        const zoomDelta = e.deltaY > 0 ? 1.1 : 0.9;
        zoom = Math.max(0.1, Math.min(10.0, zoom * zoomDelta));

        // Zoom centered on mouse position
        const rect = canvas.getBoundingClientRect();
        const mouseX = e.clientX - rect.left;
        const mouseY = e.clientY - rect.top;

        // Get grid position under mouse before zoom
        const gridBefore = renderer.getGridIndexFromScreen(mouseX, mouseY);

        // Apply new zoom
        renderer.setCamera(cameraX, cameraY, zoom);

        // Adjust camera to keep mouse over same grid position (if valid)
        if (gridBefore) {
            const gridAfter = renderer.getGridIndexFromScreen(mouseX, mouseY);
            if (gridAfter) {
                // Calculate offset needed to keep grid position under cursor
                const pos = renderer.getCameraPosition();
                cameraX = pos.x;
                cameraY = pos.y;
            }
        }
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
        // Negative because dragging right should move view left (camera right)
        const texDeltaX = (-deltaX / canvas.width) * renderer.getZoom();
        const texDeltaY = (-deltaY / canvas.height) * renderer.getZoom();

        cameraX += texDeltaX;
        cameraY += texDeltaY;

        renderer.setCamera(cameraX, cameraY, zoom);

        // Update camera position from clamped values
        const pos = renderer.getCameraPosition();
        cameraX = pos.x;
        cameraY = pos.y;
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

<style>
    .map-canvas {
        border: 1px solid #374151;
        border-radius: 4px;
        background: #000;
        cursor: grab;
    }

    .map-canvas:active {
        cursor: grabbing;
    }
</style>
