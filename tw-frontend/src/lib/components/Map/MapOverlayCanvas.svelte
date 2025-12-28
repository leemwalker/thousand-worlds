<script lang="ts">
    import { onMount, afterUpdate } from "svelte";

    export let width: number;
    export let height: number;
    export let gridWidth: number;
    export let gridHeight: number;
    export let tectonicsData: number[] | null = null;
    export let plateInfo: any[] = [];
    export let mineralsData: any[] | null = null;
    export let showTectonics = false;
    export let showMinerals = false;
    export let cameraX = 0.5;
    export let cameraY = 0.5;
    export let zoom = 1.0;

    let canvas: HTMLCanvasElement;

    // Plate color palette (10 distinct colors with low opacity for shading)
    const PLATE_COLORS = [
        "rgba(147, 51, 234, 0.2)", // Purple
        "rgba(59, 130, 246, 0.2)", // Blue
        "rgba(16, 185, 129, 0.2)", // Green
        "rgba(245, 158, 11, 0.2)", // Orange
        "rgba(239, 68, 68, 0.2)", // Red
        "rgba(236, 72, 153, 0.2)", // Pink
        "rgba(6, 182, 212, 0.2)", // Cyan
        "rgba(132, 204, 22, 0.2)", // Lime
        "rgba(168, 85, 247, 0.2)", // Violet
        "rgba(251, 191, 36, 0.2)", // Amber
        "rgba(99, 102, 241, 0.2)", // Indigo
        "rgba(14, 165, 233, 0.2)", // Sky
    ];

    // Border colors (same palette but fully opaque)
    const PLATE_BORDERS = [
        "rgb(147, 51, 234)",
        "rgb(59, 130, 246)",
        "rgb(16, 185, 129)",
        "rgb(245, 158, 11)",
        "rgb(239, 68, 68)",
        "rgb(236, 72, 153)",
        "rgb(6, 182, 212)",
        "rgb(132, 204, 22)",
        "rgb(168, 85, 247)",
        "rgb(251, 191, 36)",
        "rgb(99, 102, 241)",
        "rgb(14, 165, 233)",
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
        plateInfo,
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

        // Draw tectonic plates (fill and labels)
        if (showTectonics && tectonicsData && tectonicsData.length > 0) {
            drawTectonicPlates(ctx, startX, startY, cellW, cellH);
        }

        // Draw mineral deposits
        if (showMinerals && mineralsData && mineralsData.length > 0) {
            drawMineralDeposits(ctx, startX, startY, cellW, cellH);
        }
    }

    function drawTectonicPlates(
        ctx: CanvasRenderingContext2D,
        startX: number,
        startY: number,
        cellW: number,
        cellH: number,
    ) {
        if (!tectonicsData) return;

        // 1. Draw Plate Fills (Shaded Area)
        for (let gy = 0; gy < gridHeight; gy++) {
            for (let gx = 0; gx < gridWidth; gx++) {
                const idx = gy * gridWidth + gx;
                const plateId = tectonicsData[idx];

                if (plateId === undefined || plateId < 0) continue;

                const screenX = (gx - startX) * cellW;
                const screenY = (gy - startY) * cellH;

                // Optimization: skip if completely off-screen
                if (
                    screenX + cellW < 0 ||
                    screenX > width ||
                    screenY + cellH < 0 ||
                    screenY > height
                ) {
                    continue;
                }

                // Fill color
                const color =
                    PLATE_COLORS[(plateId - 1) % PLATE_COLORS.length] ||
                    PLATE_COLORS[0] ||
                    "rgba(0,0,0,0.5)";
                ctx.fillStyle = color;
                ctx.fillRect(
                    Math.floor(screenX),
                    Math.floor(screenY),
                    Math.ceil(cellW),
                    Math.ceil(cellH),
                );

                // Draw borders if edge
                let isEdge = false;
                if (gx < gridWidth - 1 && tectonicsData[idx + 1] !== plateId)
                    isEdge = true;
                if (
                    gy < gridHeight - 1 &&
                    tectonicsData[idx + gridWidth] !== plateId
                )
                    isEdge = true;

                // Optional: Draw explicit border lines for sharper definition
                if (isEdge) {
                    const borderColor =
                        PLATE_BORDERS[(plateId - 1) % PLATE_BORDERS.length] ||
                        PLATE_BORDERS[0] ||
                        "black";
                    ctx.fillStyle = borderColor;
                    // Draw a thin border on the edge side
                    if (
                        gx < gridWidth - 1 &&
                        tectonicsData[idx + 1] !== plateId
                    ) {
                        ctx.fillRect(screenX + cellW - 1, screenY, 1, cellH);
                    }
                    if (
                        gy < gridHeight - 1 &&
                        tectonicsData[idx + gridWidth] !== plateId
                    ) {
                        ctx.fillRect(screenX, screenY + cellH - 1, cellW, 1);
                    }
                }
            }
        }

        // 2. Draw Labels
        if (plateInfo && plateInfo.length > 0 && zoom < 5.0) {
            // Hide labels if zoomed out too far? No, assume reasonable zoom.
            ctx.textAlign = "center";
            ctx.textBaseline = "middle";
            // Scale font size
            const fontSize = Math.max(10, Math.min(24, cellW * 2));
            ctx.font = `bold ${fontSize}px sans-serif`;

            for (const plate of plateInfo) {
                if (!plate.name) continue;

                const screenX = (plate.center_x - startX) * cellW + cellW / 2;
                const screenY = (plate.center_y - startY) * cellH + cellH / 2;

                if (
                    screenX < -100 ||
                    screenX > width + 100 ||
                    screenY < -50 ||
                    screenY > height + 50
                ) {
                    continue;
                }

                // Stroke text for readability over map
                ctx.lineWidth = 3;
                ctx.strokeStyle = "rgba(0, 0, 0, 0.8)";
                ctx.strokeText(plate.name, screenX, screenY);

                ctx.fillStyle = "white";
                ctx.fillText(plate.name, screenX, screenY);
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
