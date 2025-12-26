<script context="module" lang="ts">
    // Entity type for overlay rendering
    export interface MapEntity {
        x: number; // Grid X coordinate
        y: number; // Grid Y coordinate
        type: string; // 'player' | 'npc' | 'creature' | 'item' | 'portal'
        name?: string;
        direction?: number; // Facing direction in degrees (for player)
        friendly?: boolean; // For NPC coloring
    }
</script>

<script lang="ts">
    import { onMount, onDestroy } from "svelte";

    /**
     * Entity overlay component - renders entities on a transparent 2D canvas
     * that sits on top of the WebGL map canvas.
     */

    export let width: number = 400;
    export let height: number = 400;

    // Camera state (synchronized from WebGL renderer)
    export let cameraX: number = 0.5;
    export let cameraY: number = 0.5;
    export let texScaleX: number = 1.0;
    export let texScaleY: number = 1.0;

    // Grid dimensions for coordinate conversion
    export let gridWidth: number = 128;
    export let gridHeight: number = 64;

    // Entity data from parent
    export let entities: MapEntity[] = [];

    // Player position in grid coordinates
    export let playerX: number = 0;
    export let playerY: number = 0;

    let canvas: HTMLCanvasElement;
    let ctx: CanvasRenderingContext2D | null = null;
    let animationFrame: number | null = null;

    onMount(() => {
        ctx = canvas.getContext("2d", { alpha: true });
        renderLoop();
    });

    onDestroy(() => {
        if (animationFrame) {
            cancelAnimationFrame(animationFrame);
        }
    });

    function renderLoop() {
        render();
        animationFrame = requestAnimationFrame(renderLoop);
    }

    /**
     * Convert grid coordinates to screen position
     * Uses the same transform as the WebGL shader (inverse of getGridIndexFromScreen)
     */
    function gridToScreen(
        gridX: number,
        gridY: number,
    ): { x: number; y: number } | null {
        // Convert grid to texture coordinates (0-1)
        const texX = (gridX + 0.5) / gridWidth; // Center of grid cell
        const texY = (gridY + 0.5) / gridHeight;

        // Apply camera transform (inverse of shader)
        // texCoord = cameraCenter + (screenNDC - 0.5) * scale
        // screenNDC = (texCoord - cameraCenter) / scale + 0.5
        const ndcX = (texX - cameraX) / texScaleX + 0.5;
        const ndcY = (texY - cameraY) / texScaleY + 0.5;

        // Check if visible on screen
        if (ndcX < 0 || ndcX > 1 || ndcY < 0 || ndcY > 1) {
            return null;
        }

        // Convert to screen pixels
        return {
            x: ndcX * width,
            y: ndcY * height,
        };
    }

    /**
     * Calculate marker size based on zoom level
     */
    function getMarkerSize(): number {
        const zoomFactor = Math.max(texScaleX, texScaleY);
        // Larger markers when zoomed in, smaller when zoomed out
        const baseSize = Math.min(width, height) * 0.02;
        return Math.max(3, baseSize / zoomFactor);
    }

    function render() {
        if (!ctx) return;

        // Clear canvas with transparency
        ctx.clearRect(0, 0, width, height);

        const markerSize = getMarkerSize();

        // Render player marker
        const playerScreen = gridToScreen(playerX, playerY);
        if (playerScreen) {
            renderPlayer(playerScreen.x, playerScreen.y, markerSize);
        }

        // Render all entities
        for (const entity of entities) {
            const screen = gridToScreen(entity.x, entity.y);
            if (!screen) continue;

            switch (entity.type.toLowerCase()) {
                case "player":
                    // Skip if already rendered above
                    break;
                case "npc":
                    renderNPC(screen.x, screen.y, markerSize, entity.friendly);
                    break;
                case "creature":
                    renderCreature(screen.x, screen.y, markerSize);
                    break;
                case "item":
                    renderItem(screen.x, screen.y, markerSize);
                    break;
                case "portal":
                    renderPortal(screen.x, screen.y, markerSize);
                    break;
                default:
                    renderGeneric(screen.x, screen.y, markerSize);
            }
        }
    }

    function renderPlayer(x: number, y: number, size: number) {
        if (!ctx) return;

        // White filled circle with dark outline
        ctx.beginPath();
        ctx.arc(x, y, size * 1.2, 0, Math.PI * 2);
        ctx.fillStyle = "#FFFFFF";
        ctx.fill();
        ctx.strokeStyle = "#333333";
        ctx.lineWidth = 2;
        ctx.stroke();

        // Direction indicator (small triangle pointing up)
        ctx.beginPath();
        ctx.moveTo(x, y - size * 1.8);
        ctx.lineTo(x - size * 0.5, y - size * 0.8);
        ctx.lineTo(x + size * 0.5, y - size * 0.8);
        ctx.closePath();
        ctx.fillStyle = "#FFFFFF";
        ctx.fill();
    }

    function renderNPC(x: number, y: number, size: number, friendly?: boolean) {
        if (!ctx) return;

        // Blue for friendly, red for hostile
        const color = friendly !== false ? "#4A90D9" : "#D94A4A";

        ctx.beginPath();
        ctx.arc(x, y, size, 0, Math.PI * 2);
        ctx.fillStyle = color;
        ctx.fill();
        ctx.strokeStyle = "#FFFFFF";
        ctx.lineWidth = 1;
        ctx.stroke();
    }

    function renderCreature(x: number, y: number, size: number) {
        if (!ctx) return;

        // Orange diamond
        ctx.beginPath();
        ctx.moveTo(x, y - size);
        ctx.lineTo(x + size, y);
        ctx.lineTo(x, y + size);
        ctx.lineTo(x - size, y);
        ctx.closePath();
        ctx.fillStyle = "#E67E22";
        ctx.fill();
        ctx.strokeStyle = "#FFFFFF";
        ctx.lineWidth = 1;
        ctx.stroke();
    }

    function renderItem(x: number, y: number, size: number) {
        if (!ctx) return;

        // Gold square
        const halfSize = size * 0.7;
        ctx.fillStyle = "#F1C40F";
        ctx.fillRect(x - halfSize, y - halfSize, halfSize * 2, halfSize * 2);
        ctx.strokeStyle = "#B7950B";
        ctx.lineWidth = 1;
        ctx.strokeRect(x - halfSize, y - halfSize, halfSize * 2, halfSize * 2);
    }

    function renderPortal(x: number, y: number, size: number) {
        if (!ctx) return;

        // Magenta pulsing circle
        ctx.beginPath();
        ctx.arc(x, y, size * 1.5, 0, Math.PI * 2);
        ctx.fillStyle = "rgba(155, 89, 182, 0.7)";
        ctx.fill();
        ctx.strokeStyle = "#8E44AD";
        ctx.lineWidth = 2;
        ctx.stroke();
    }

    function renderGeneric(x: number, y: number, size: number) {
        if (!ctx) return;

        // Gray circle
        ctx.beginPath();
        ctx.arc(x, y, size * 0.8, 0, Math.PI * 2);
        ctx.fillStyle = "#95A5A6";
        ctx.fill();
    }

    // Re-render when props change
    $: if (ctx && (entities || cameraX || cameraY || texScaleX || texScaleY)) {
        render();
    }
</script>

<canvas bind:this={canvas} {width} {height} class="entity-overlay"></canvas>

<style>
    .entity-overlay {
        position: absolute;
        top: 0;
        left: 0;
        pointer-events: none;
        z-index: 10;
    }
</style>
