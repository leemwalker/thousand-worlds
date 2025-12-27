<script lang="ts">
    import { onMount, afterUpdate } from "svelte";

    export let width: number;
    export let height: number;
    export let gridWidth: number;
    export let gridHeight: number;
    export let tectonicsData: number[] | null = null;
    export let mineralsData: any[] | null = null;
    export let showTectonics = false;
    export let showMinerals = false;
    export let cameraX = 0.5;
    export let cameraY = 0.5;
    export let zoom = 1.0;

    let canvas: HTMLCanvasElement;

    // Plate color palette (10 distinct colors with semi-transparency)
    const PLATE_COLORS = [
        "rgba(147, 51, 234, 0.4)", // Purple
        "rgba(59, 130, 246, 0.4)", // Blue
        "rgba(16, 185, 129, 0.4)", // Green
        "rgba(245, 158, 11, 0.4)", // Orange
        "rgba(239, 68, 68, 0.4)", // Red
        "rgba(236, 72, 153, 0.4)", // Pink
        "rgba(6, 182, 212, 0.4)", // Cyan
        "rgba(132, 204, 22, 0.4)", // Lime
        "rgba(168, 85, 247, 0.4)", // Violet
        "rgba(251, 191, 36, 0.4)", // Amber
        "rgba(99, 102, 241, 0.4)", // Indigo
        "rgba(14, 165, 233, 0.4)", // Sky
    ];

    // Mineral type colors
    const MINERAL_COLORS: Record<string, string> = {
        iron: "rgba(139, 69, 19, 0.9)", // Brown
        gold: "rgba(255, 215, 0, 0.9)", // Gold
        coal: "rgba(50, 50, 50, 0.9)", // Dark gray
        fossil: "rgba(210, 180, 140, 0.9)", // Tan
        oil: "rgba(20, 20, 20, 0.9)", // Black
        default: "rgba(251, 191, 36, 0.8)", // Amber
    };

    $: if (canvas) {
        drawOverlays();
    }

    // Redraw when any relevant prop changes
    $: showTectonics,
        showMinerals,
        cameraX,
        cameraY,
        zoom,
        width,
        height,
        tectonicsData,
        mineralsData,
        drawOverlays();

    function drawOverlays() {
        if (!canvas) return;
        const ctx = canvas.getContext("2d");
        if (!ctx) return;

        // Clear canvas
        ctx.clearRect(0, 0, width, height);

        if (!showTectonics && !showMinerals) return;

        // Calculate visible area based on camera and zoom
        // zoom = 1.0 means full world visible, zoom < 1.0 = zoomed in
        const visibleWidth = gridWidth * zoom;
        const visibleHeight = gridHeight * zoom;

        // Calculate the top-left corner of visible area in grid coordinates
        const startX = cameraX * gridWidth - visibleWidth / 2;
        const startY = cameraY * gridHeight - visibleHeight / 2;

        // Cell size on screen
        const cellW = width / visibleWidth;
        const cellH = height / visibleHeight;

        // Draw tectonic plate boundaries
        if (showTectonics && tectonicsData && tectonicsData.length > 0) {
            drawTectonicBoundaries(ctx, startX, startY, cellW, cellH);
        }

        // Draw mineral deposits
        if (showMinerals && mineralsData && mineralsData.length > 0) {
            drawMineralDeposits(ctx, startX, startY, cellW, cellH);
        }
    }

    function drawTectonicBoundaries(
        ctx: CanvasRenderingContext2D,
        startX: number,
        startY: number,
        cellW: number,
        cellH: number,
    ) {
        if (!tectonicsData) return;

        // Draw boundary lines where plate IDs differ
        ctx.lineWidth = Math.max(1, cellW * 0.1);

        for (let gy = 0; gy < gridHeight; gy++) {
            for (let gx = 0; gx < gridWidth; gx++) {
                const idx = gy * gridWidth + gx;
                const plateId = tectonicsData[idx];

                if (plateId === undefined || plateId < 0) continue;

                const screenX = (gx - startX) * cellW;
                const screenY = (gy - startY) * cellH;

                // Skip if completely off-screen
                if (
                    screenX + cellW < 0 ||
                    screenX > width ||
                    screenY + cellH < 0 ||
                    screenY > height
                ) {
                    continue;
                }

                // Check right neighbor for boundary
                if (gx < gridWidth - 1) {
                    const rightPlate = tectonicsData[idx + 1];
                    if (
                        rightPlate !== undefined &&
                        rightPlate >= 0 &&
                        rightPlate !== plateId
                    ) {
                        const color =
                            PLATE_COLORS[plateId % PLATE_COLORS.length] ??
                            PLATE_COLORS[0];
                        ctx.strokeStyle = color;
                        ctx.beginPath();
                        ctx.moveTo(screenX + cellW, screenY);
                        ctx.lineTo(screenX + cellW, screenY + cellH);
                        ctx.stroke();
                    }
                }

                // Check bottom neighbor for boundary
                if (gy < gridHeight - 1) {
                    const bottomPlate = tectonicsData[idx + gridWidth];
                    if (
                        bottomPlate !== undefined &&
                        bottomPlate >= 0 &&
                        bottomPlate !== plateId
                    ) {
                        const color =
                            PLATE_COLORS[plateId % PLATE_COLORS.length] ??
                            PLATE_COLORS[0];
                        ctx.strokeStyle = color;
                        ctx.beginPath();
                        ctx.moveTo(screenX, screenY + cellH);
                        ctx.lineTo(screenX + cellW, screenY + cellH);
                        ctx.stroke();
                    }
                }
            }
        }
    }

    function drawMineralDeposits(
        ctx: CanvasRenderingContext2D,
        startX: number,
        startY: number,
        cellW: number,
        cellH: number,
    ) {
        if (!mineralsData) return;

        for (const deposit of mineralsData) {
            const screenX = (deposit.x - startX) * cellW + cellW / 2;
            const screenY = (deposit.y - startY) * cellH + cellH / 2;

            // Skip if off-screen
            if (
                screenX < -10 ||
                screenX > width + 10 ||
                screenY < -10 ||
                screenY > height + 10
            ) {
                continue;
            }

            // Marker size based on zoom and quantity
            const baseSize = Math.max(4, Math.min(cellW, cellH) * 0.5);
            const quantityBonus =
                Math.min(deposit.quantity / 100, 1) * baseSize * 0.5;
            const radius = baseSize + quantityBonus;

            // Get color based on type - ensure non-undefined with fallback
            const color: string =
                MINERAL_COLORS[deposit.type] ??
                MINERAL_COLORS["default"] ??
                "rgba(251, 191, 36, 0.8)";

            // Draw deposit marker
            ctx.beginPath();
            ctx.arc(screenX, screenY, radius, 0, Math.PI * 2);

            // Discovered deposits are solid, undiscovered are semi-transparent
            if (deposit.discovered) {
                ctx.fillStyle = color;
                ctx.fill();
                ctx.strokeStyle = "rgba(255, 255, 255, 0.8)";
                ctx.lineWidth = 2;
                ctx.stroke();
            } else {
                // Undiscovered - dashed outline
                ctx.setLineDash([3, 3]);
                ctx.strokeStyle = color;
                ctx.lineWidth = 2;
                ctx.stroke();
                ctx.setLineDash([]);
            }
        }
    }
</script>

<canvas
    bind:this={canvas}
    {width}
    {height}
    class="absolute inset-0 pointer-events-none"
/>
