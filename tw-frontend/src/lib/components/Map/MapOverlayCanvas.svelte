<script lang="ts">
    import { onMount, afterUpdate } from "svelte";
    import type { OverlayMode, OverlayData } from "$lib/types/overlays";

    export let width: number;
    export let height: number;
    export let gridWidth: number;
    export let gridHeight: number;

    // New Props
    export let activeLayer: OverlayMode = "none";
    export let overlayData: OverlayData = {};

    // Legacy Props (Backwards Compatibility)
    export let tectonicsData: number[] | null = null;
    export let plateInfo: any[] = [];
    export let mineralsData: any[] | null = null;
    export let showTectonics = false;
    export let showMinerals = false;

    export let cameraX = 0.5;
    export let cameraY = 0.5;
    export let zoom = 1.0;

    let canvas: HTMLCanvasElement;

    // --- Color Palettes ---

    // Biomes: Whittaker-ish mapping
    const BIOME_COLORS: Record<number, string> = {
        0: "#1e3a8a", // Ocean (Dark Blue) - Should verify ID
        1: "#E6CC80", // Desert (Sand)
        2: "#DAA520", // Savanna (Goldenrod)
        3: "#2E8B57", // Jungle (SeaGreen)
        4: "#7CFC00", // Grassland (LawnGreen) - maybe darker?
        5: "#228B22", // Forest (ForestGreen)
        6: "#556B2F", // Taiga (DarkOliveGreen)
        7: "#A0522D", // Tundra (Sienna)
        8: "#E0FFFF", // Ice (LightCyan)
    };

    // Tectonics: 12 distinct colors with low opacity for glass effect
    const PLATE_COLORS = [
        "rgba(147, 51, 234, 0.4)",
        "rgba(59, 130, 246, 0.4)",
        "rgba(16, 185, 129, 0.4)",
        "rgba(245, 158, 11, 0.4)",
        "rgba(239, 68, 68, 0.4)",
        "rgba(236, 72, 153, 0.4)",
        "rgba(6, 182, 212, 0.4)",
        "rgba(132, 204, 22, 0.4)",
        "rgba(168, 85, 247, 0.4)",
        "rgba(251, 191, 36, 0.4)",
        "rgba(99, 102, 241, 0.4)",
        "rgba(14, 165, 233, 0.4)",
    ];

    $: if (canvas) {
        drawOverlays();
    }

    // Redraw on any prop change
    $: activeLayer,
        overlayData,
        showTectonics,
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

        // --- Aspect Ratio Correction ---
        const canvasAspect = width / height;
        const worldAspect = gridWidth / gridHeight;
        let baseScaleX = 1.0;
        let baseScaleY = 1.0;

        if (worldAspect > canvasAspect) {
            baseScaleY = worldAspect / canvasAspect;
        } else {
            baseScaleX = canvasAspect / worldAspect;
        }

        const effectiveScaleX = baseScaleX * zoom;
        const effectiveScaleY = baseScaleY * zoom;

        // Size of the world in screen pixels
        // Note: texScale from WebGL is "Field of View Size".
        // texScale=1 -> View=World. texScale=2 -> View=2xWorld (ZoomOut).
        // So World (in pixels) = Canvas / texScale.
        const worldWidthPx = width / effectiveScaleX;
        const worldHeightPx = height / effectiveScaleY;

        // Cell size on screen
        const cellW = worldWidthPx / gridWidth;
        const cellH = worldHeightPx / gridHeight;

        // Center of the screen in pixels
        const centerScreenX = width / 2;
        const centerScreenY = height / 2;

        // Where the camera is looking in "Grid Space" (float)
        const centerTx = cameraX * gridWidth;
        const centerTy = cameraY * gridHeight;

        // Render based on active layer
        // Legacy Support: map old boolean props to new modes if activeLayer is 'none'
        let mode = activeLayer;
        if (mode === "none") {
            if (showTectonics) mode = "tectonics";
            else if (showMinerals) mode = "resources";
        }

        // Consolidated Draw Loop
        // We iterate grid cells and project them using the Wrapping Logic

        if (mode === "tectonics") {
            // Use legacy data if provided, else overlayData
            const data = overlayData.tectonics || tectonicsData;
            if (data) {
                drawGrid(
                    ctx,
                    data,
                    (val, x, y, w, h) => {
                        if (val <= 0) return;
                        const color =
                            PLATE_COLORS[(val - 1) % PLATE_COLORS.length] ||
                            PLATE_COLORS[0];
                        if (color) {
                            ctx.fillStyle = color;
                            ctx.fillRect(x, y, w, h);
                        }

                        // Simple borders (optional, keeping minimal for perf)
                        // Check neighbors logic is heavy inside generic loop, skipping for now or implement separate pass
                    },
                    centerTx,
                    centerTy,
                    cellW,
                    cellH,
                    centerScreenX,
                    centerScreenY,
                );

                // Draw Labels
                const pInfo = overlayData.plate_info || plateInfo;
                if (pInfo && zoom < 5.0) {
                    drawPlateLabels(
                        ctx,
                        pInfo,
                        centerTx,
                        centerTy,
                        cellW,
                        cellH,
                        centerScreenX,
                        centerScreenY,
                    );
                }
            }
        } else if (mode === "temp" && overlayData.temp) {
            drawGrid(
                ctx,
                overlayData.temp,
                (val, x, y, w, h) => {
                    // Hue: 240 (Blue) -> 0 (Red). Value 0->1.
                    const hue = 240 - val * 240;
                    ctx.fillStyle = `hsla(${hue}, 80%, 50%, 0.6)`;
                    ctx.fillRect(x, y, w, h);
                },
                centerTx,
                centerTy,
                cellW,
                cellH,
                centerScreenX,
                centerScreenY,
            );
        } else if (mode === "moisture" && overlayData.moisture) {
            drawGrid(
                ctx,
                overlayData.moisture,
                (val, x, y, w, h) => {
                    // Hue: 40 (Dry/Brown) -> 200 (Wet/Blue)
                    // Or just Opacity of Blue?
                    // Let's do White->Blue
                    ctx.fillStyle = `rgba(0, 100, 255, ${val * 0.8})`;
                    ctx.fillRect(x, y, w, h);
                },
                centerTx,
                centerTy,
                cellW,
                cellH,
                centerScreenX,
                centerScreenY,
            );
        } else if (mode === "elevation" && overlayData.elevation) {
            drawGrid(
                ctx,
                overlayData.elevation,
                (val, x, y, w, h) => {
                    // Green -> Grey -> White
                    // 0.5 is sea level (usually)?
                    // Let's assume 0-1 full range.
                    const l = Math.floor(val * 100);
                    // Simple greyscale for now
                    ctx.fillStyle = `hsl(120, 0%, ${l}%)`; // Grey
                    if (val < 0.5)
                        ctx.fillStyle = `hsl(200, 80%, ${30 + val * 40}%)`; // Water
                    else if (val < 0.6)
                        ctx.fillStyle = `hsl(100, 60%, ${40 + (val - 0.5) * 100}%)`; // Land
                    else
                        ctx.fillStyle = `hsl(0, 0%, ${50 + (val - 0.6) * 200}%)`; // Mountain

                    ctx.globalAlpha = 0.6;
                    ctx.fillRect(x, y, w, h);
                    ctx.globalAlpha = 1.0;
                },
                centerTx,
                centerTy,
                cellW,
                cellH,
                centerScreenX,
                centerScreenY,
            );
        } else if (mode === "biome" && overlayData.biome) {
            drawGrid(
                ctx,
                overlayData.biome,
                (val, x, y, w, h) => {
                    ctx.fillStyle = BIOME_COLORS[val] || "#000000";
                    ctx.globalAlpha = 0.7;
                    ctx.fillRect(x, y, w, h);
                    ctx.globalAlpha = 1.0;
                },
                centerTx,
                centerTy,
                cellW,
                cellH,
                centerScreenX,
                centerScreenY,
            );
        }

        // Resources (Icons) - drawn on top of any layer if mode is resources or forced enabled
        if (mode === "resources" || showMinerals) {
            const res = overlayData.resources;
            const mins = overlayData.minerals || mineralsData;
            // Merge or handle both?
            // Prioritize new ResourceNode system
            if (res)
                drawResources(
                    ctx,
                    res,
                    centerTx,
                    centerTy,
                    cellW,
                    cellH,
                    centerScreenX,
                    centerScreenY,
                );
            if (mins && !res)
                drawMineralsLegacy(
                    ctx,
                    mins,
                    centerTx,
                    centerTy,
                    cellW,
                    cellH,
                    centerScreenX,
                    centerScreenY,
                );
        }
    }

    // Generic Grid Drawer with Wrapping
    function drawGrid(
        ctx: CanvasRenderingContext2D,
        data: number[],
        renderer: (
            val: number,
            x: number,
            y: number,
            w: number,
            h: number,
        ) => void,
        centerTx: number,
        centerTy: number,
        cellW: number,
        cellH: number,
        centerScreenX: number,
        centerScreenY: number,
    ) {
        // Optimize: Only iterate visible cells?
        // For 128x64 (8k cells), full iteration is fast enough in JS (~1-2ms).
        // Clipping is harder with wrapping logic, simpler to iterate all and early-out content off-screen.

        // To fix floating point gaps, we use ceil for width/height
        const w = Math.ceil(cellW);
        const h = Math.ceil(cellH);

        for (let gy = 0; gy < gridHeight; gy++) {
            // Y wrapping? Map doesn't wrap Y (poles).
            // Calculate screenY normally
            const deltaY = gy - centerTy;
            const screenY = centerScreenY + deltaY * cellH;

            // Y Culling
            if (screenY > height || screenY + h < 0) continue;

            for (let gx = 0; gx < gridWidth; gx++) {
                const idx = gy * gridWidth + gx;
                const val = data[idx];

                // === X WRAPPING LOGIC ===
                let deltaX = gx - centerTx;

                // Shortest path wrapping
                // If map is 100 wide. Camera at 90. Point at 10.
                // 10 - 90 = -80.
                // -80 < -50. Add 100 -> 20. Correct (10 is to the right of 90).
                if (deltaX < -gridWidth / 2) deltaX += gridWidth;
                if (deltaX > gridWidth / 2) deltaX -= gridWidth;

                const screenX = centerScreenX + deltaX * cellW;

                // X Culling
                if (screenX > width || screenX + w < 0) continue;

                // Render
                if (val !== undefined) {
                    renderer(
                        val,
                        Math.floor(screenX),
                        Math.floor(screenY),
                        w,
                        h,
                    );
                }
            }
        }
    }

    function drawPlateLabels(
        ctx: CanvasRenderingContext2D,
        plates: any[],
        centerTx: number,
        centerTy: number,
        cellW: number,
        cellH: number,
        centerScreenX: number,
        centerScreenY: number,
    ) {
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        const fontSize = Math.max(10, Math.min(24, cellW * 2));
        ctx.font = `bold ${fontSize}px sans-serif`;
        ctx.lineWidth = 3;
        ctx.strokeStyle = "rgba(0, 0, 0, 0.8)";
        ctx.fillStyle = "white";

        for (const plate of plates) {
            const cx = plate.center_x;
            const cy = plate.center_y;

            // Y Logic
            const deltaY = cy - centerTy;
            const screenY = centerScreenY + deltaY * cellH;
            if (screenY < -50 || screenY > height + 50) continue;

            // X Logic (Wrapping)
            let deltaX = cx - centerTx;
            if (deltaX < -gridWidth / 2) deltaX += gridWidth;
            if (deltaX > gridWidth / 2) deltaX -= gridWidth;
            const screenX = centerScreenX + deltaX * cellW;
            if (screenX < -100 || screenX > width + 100) continue;

            if (plate.name) {
                ctx.strokeText(plate.name, screenX, screenY);
                ctx.fillText(plate.name, screenX, screenY);
            }
        }
    }

    function drawResources(
        ctx: CanvasRenderingContext2D,
        nodes: any[], // ResourceNode[]
        centerTx: number,
        centerTy: number,
        cellW: number,
        cellH: number,
        centerScreenX: number,
        centerScreenY: number,
    ) {
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        // Scale icon with zoom
        const fontSize = Math.max(12, cellW * 0.8);
        ctx.font = `${fontSize}px serif`; // Emojis often render better with serif fallback or standard

        for (const node of nodes) {
            // Logic identical to labels
            const deltaY = node.y - centerTy;
            const screenY = centerScreenY + deltaY * cellH; // + cellH/2 to center in cell
            if (screenY < -20 || screenY > height + 20) continue;

            let deltaX = node.x - centerTx;
            if (deltaX < -gridWidth / 2) deltaX += gridWidth;
            if (deltaX > gridWidth / 2) deltaX -= gridWidth;
            const screenX = centerScreenX + deltaX * cellW; // + cellW/2
            if (screenX < -20 || screenX > width + 20) continue;

            // Offset to center of cell
            const drawX = screenX + cellW / 2;
            const drawY = screenY + cellH / 2;

            let icon = "❓";
            if (node.type === "gold") icon = "🟡";
            else if (node.type === "iron")
                icon = "⚪"; // Silver/Iron
            else if (node.type === "cave") icon = "🕳️";
            else if (node.type === "coal") icon = "⚫";

            // Shadow for visibility
            ctx.shadowColor = "black";
            ctx.shadowBlur = 2;
            ctx.fillText(icon, drawX, drawY);
            ctx.shadowBlur = 0;
        }
    }

    // Legacy function for old mineralsData format
    function drawMineralsLegacy(
        ctx: CanvasRenderingContext2D,
        deposits: any[],
        centerTx: number,
        centerTy: number,
        cellW: number,
        cellH: number,
        centerScreenX: number,
        centerScreenY: number,
    ) {
        for (const deposit of deposits) {
            const deltaY = deposit.y - centerTy;
            const screenY = centerScreenY + deltaY * cellH;
            if (screenY < -20 || screenY > height + 20) continue;

            let deltaX = deposit.x - centerTx;
            if (deltaX < -gridWidth / 2) deltaX += gridWidth;
            if (deltaX > gridWidth / 2) deltaX -= gridWidth;
            const screenX = centerScreenX + deltaX * cellW;
            if (screenX < -20 || screenX > width + 20) continue;

            const drawX = screenX + cellW / 2;
            const drawY = screenY + cellH / 2;
            const radius = Math.min(cellW, cellH) * 0.4;

            ctx.beginPath();
            ctx.arc(drawX, drawY, radius, 0, Math.PI * 2);
            ctx.fillStyle = deposit.type === "gold" ? "gold" : "brown";
            ctx.fill();
            ctx.strokeStyle = "white";
            ctx.stroke();
        }
    }
</script>

<canvas
    bind:this={canvas}
    {width}
    {height}
    class="absolute inset-0 pointer-events-none"
/>
