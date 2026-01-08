<script lang="ts">
    import { onMount, onDestroy, createEventDispatcher } from "svelte";

    export let width: number | string = "100%";
    export let height: number | string = "100%";

    const dispatch = createEventDispatcher<{
        canvasReady: HTMLCanvasElement;
        resize: { width: number; height: number };
    }>();

    let canvas: HTMLCanvasElement;
    let container: HTMLDivElement;
    let resizeObserver: ResizeObserver | null = null;

    onMount(() => {
        if (!canvas) return;

        // Notify parent that canvas is ready
        dispatch("canvasReady", canvas);

        // Set up resize observer
        resizeObserver = new ResizeObserver((entries) => {
            for (const entry of entries) {
                const { width, height } = entry.contentRect;
                // Update canvas size to match container
                canvas.width = width;
                canvas.height = height;
                dispatch("resize", { width, height });
            }
        });

        resizeObserver.observe(container);
    });

    onDestroy(() => {
        resizeObserver?.disconnect();
    });

    // Expose canvas for external access
    export function getCanvas(): HTMLCanvasElement {
        return canvas;
    }
</script>

/** * SceneCanvas.svelte * Owns the canvas element and provides it to
SceneManager. * Part of the Engine Hoist refactor - separates canvas from scene
logic. */
<div
    bind:this={container}
    class="scene-canvas-container"
    style="width: {typeof width === 'number'
        ? `${width}px`
        : width}; height: {typeof height === 'number'
        ? `${height}px`
        : height};"
>
    <canvas bind:this={canvas} class="scene-canvas" tabindex="0"></canvas>
</div>

<style>
    .scene-canvas-container {
        position: relative;
        overflow: hidden;
    }

    .scene-canvas {
        display: block;
        width: 100%;
        height: 100%;
        outline: none;
        touch-action: none; /* Prevent browser zoom/pan on touch */
    }
</style>
