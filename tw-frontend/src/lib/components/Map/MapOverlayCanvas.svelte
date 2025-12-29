<script lang="ts">
    import { onMount, afterUpdate, createEventDispatcher } from "svelte";
    import type {
        OverlayMode,
        OverlayData,
        ResourceNode,
    } from "$lib/types/overlays";

    const dispatch = createEventDispatcher();

    export let width: number;
    export let height: number;
    export let gridWidth: number;
    export let gridHeight: number;

    // Multi-Layer Support
    // We accept a Set of strings. If passing from parent, ensure it's reactive.
    export let activeLayers: Set<OverlayMode> = new Set();

    // Legacy single activeLayer support (mapped to Set internally if needed, but easier to just use new prop)
    // We will ignore `activeLayer` prop here and expect parent to pass `activeLayers`.

    export let overlayData: OverlayData = {};

    // Legacy Props (Backwards Compatibility - Mapped to layers in parent or ignored)
    export let tectonicsData: number[] | null = null;
    export let plateInfo: any[] = [];
    export let mineralsData: any[] | null = null;
    export let showTectonics = false;
    export let showMinerals = false;

    export let cameraX = 0.5;
    export let cameraY = 0.5;
    export let zoom = 1.0;

    let canvas: HTMLCanvasElement;
    let hoveredItem: any = null; // Track hovered item for tooltip

    // Store render positions for hit detection
    // Simple list of {x, y, radius, data}
    let hitTargets: { x: number; y: number; r: number; node: ResourceNode }[] =
        [];

    // --- Color Palettes ---
    const BIOME_COLORS: Record<number, string> = {
        0: "#1e3a8a", // Ocean (Dark Blue)
        1: "#E6CC80", // Desert (Sand)
        2: "#DAA520", // Savanna (Goldenrod)
        3: "#2E8B57", // Jungle (SeaGreen)
        4: "#7CFC00", // Grassland (LawnGreen)
        5: "#228B22", // Forest (ForestGreen)
        6: "#556B2F", // Taiga (DarkOliveGreen)
        7: "#A0522D", // Tundra (Sienna)
        8: "#E0FFFF", // Ice (LightCyan)
    };

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

    // React to changes
    $: activeLayers,
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

        hitTargets = []; // Reset hit targets

        // Clear canvas
        ctx.clearRect(0, 0, width, height);

        // Aspect Ratio & Scale Logic
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
        const worldWidthPx = width / effectiveScaleX;
        const worldHeightPx = height / effectiveScaleY;
        const cellW = worldWidthPx / gridWidth;
        const cellH = worldHeightPx / gridHeight;
        const centerScreenX = width / 2;
        const centerScreenY = height / 2;
        const centerTx = cameraX * gridWidth;
        const centerTy = cameraY * gridHeight;

        // --- Render Layer Z-Order ---
        // 1. Biome OR Elevation (Opaque base)
        if (activeLayers.has("biome") && overlayData.biome) {
            drawGrid(
                ctx,
                overlayData.biome,
                (val: number, x: number, y: number, w: number, h: number) => {
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
        } else if (activeLayers.has("elevation") && overlayData.elevation) {
            drawGrid(
                ctx,
                overlayData.elevation,
                (val: number, x: number, y: number, w: number, h: number) => {
                    const waterLevel = overlayData.globalWaterLevel ?? 0.5;

                    let color = "#000000";

                    if (val <= waterLevel) {
                        // Water Gradient
                        // Normalize 0..waterLevel -> 0..1
                        const depthRatio = val / waterLevel;
                        color = getColor(depthRatio, [
                            { t: 0.0, hex: "#0a1a2f" }, // Abyss (Dark Navy)
                            { t: 0.3, hex: "#1e3a8a" }, // Ocean (Standard Blue)
                            { t: 0.8, hex: "#60a5fa" }, // Shallows (Light Blue)
                            { t: 1.0, hex: "#60a5fa" }, // Coast (Keep Light Blue)
                        ]);
                    } else {
                        // Land Gradient
                        // Normalize waterLevel..1 -> 0..1
                        const landRatio =
                            (val - waterLevel) / (1.0 - waterLevel);
                        color = getColor(landRatio, [
                            { t: 0.0, hex: "#fde047" }, // Beach (Yellow)
                            { t: 0.05, hex: "#166534" }, // Lowland (Green)
                            { t: 0.55, hex: "#854d0e" }, // Highland (Brown)
                            { t: 0.85, hex: "#525252" }, // Mountain (Grey)
                            { t: 0.95, hex: "#ffffff" }, // Snow (White)
                            { t: 1.0, hex: "#ffffff" },
                        ]);
                    }

                    ctx.fillStyle = color;
                    // Opaque for distinct viewing
                    ctx.globalAlpha = 1.0;
                    ctx.fillRect(x, y, w, h);
                },
                centerTx,
                centerTy,
                cellW,
                cellH,
                centerScreenX,
                centerScreenY,
            );
        }

        // 2. Tectonics (Semi-transparent)
        if (
            (activeLayers.has("tectonics") || showTectonics) &&
            (overlayData.tectonics || tectonicsData)
        ) {
            const data = overlayData.tectonics || tectonicsData;
            if (data) {
                drawGrid(
                    ctx,
                    data,
                    (
                        val: number,
                        x: number,
                        y: number,
                        w: number,
                        h: number,
                    ) => {
                        if (val <= 0) return;
                        const color =
                            PLATE_COLORS[(val - 1) % PLATE_COLORS.length] ||
                            PLATE_COLORS[0];
                        if (color) {
                            ctx.fillStyle = color;
                            ctx.fillRect(x, y, w, h);
                        }
                    },
                    centerTx,
                    centerTy,
                    cellW,
                    cellH,
                    centerScreenX,
                    centerScreenY,
                );

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
        }

        // 3. Heatmaps (Alpha blended)
        if (activeLayers.has("temp") && overlayData.temp) {
            drawGrid(
                ctx,
                overlayData.temp,
                (val: number, x: number, y: number, w: number, h: number) => {
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
        }
        if (activeLayers.has("moisture") && overlayData.moisture) {
            drawGrid(
                ctx,
                overlayData.moisture,
                (val: number, x: number, y: number, w: number, h: number) => {
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
        }

        // 4. Features & Resources (Icons)
        if (activeLayers.has("features") && overlayData.features) {
            drawIcons(
                ctx,
                overlayData.features,
                centerTx,
                centerTy,
                cellW,
                cellH,
                centerScreenX,
                centerScreenY,
                "feature",
            );
        }
        if (
            (activeLayers.has("resources") || showMinerals) &&
            overlayData.resources
        ) {
            drawIcons(
                ctx,
                overlayData.resources,
                centerTx,
                centerTy,
                cellW,
                cellH,
                centerScreenX,
                centerScreenY,
                "resource",
            );
        }
    }

    // Generic Grid Drawer
    function drawGrid(
        ctx: CanvasRenderingContext2D,
        data: number[],
        renderer: any,
        centerTx: number,
        centerTy: number,
        cellW: number,
        cellH: number,
        centerScreenX: number,
        centerScreenY: number,
    ) {
        const w = Math.ceil(cellW);
        const h = Math.ceil(cellH);
        for (let gy = 0; gy < gridHeight; gy++) {
            const deltaY = gy - centerTy;
            const screenY = centerScreenY + deltaY * cellH;
            if (screenY > height || screenY + h < 0) continue;

            for (let gx = 0; gx < gridWidth; gx++) {
                const idx = gy * gridWidth + gx;
                const val = data[idx];
                let deltaX = gx - centerTx;
                if (deltaX < -gridWidth / 2) deltaX += gridWidth;
                if (deltaX > gridWidth / 2) deltaX -= gridWidth;
                const screenX = centerScreenX + deltaX * cellW;
                if (screenX > width || screenX + w < 0) continue;

                if (val !== undefined)
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
            const deltaY = plate.center_y - centerTy;
            const screenY = centerScreenY + deltaY * cellH;
            if (screenY < -50 || screenY > height + 50) continue;

            let deltaX = plate.center_x - centerTx;
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

    function drawIcons(
        ctx: CanvasRenderingContext2D,
        nodes: ResourceNode[],
        centerTx: number,
        centerTy: number,
        cellW: number,
        cellH: number,
        centerScreenX: number,
        centerScreenY: number,
        category: "resource" | "feature",
    ) {
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        // Scale icon logic: Features bigger than Resources
        const scale = category === "feature" ? 1.5 : 0.8;
        const fontSize = Math.min(48, Math.max(12, cellW * scale));
        ctx.font = `${fontSize}px serif`;

        for (const node of nodes) {
            const deltaY = node.y - centerTy;
            const screenY = centerScreenY + deltaY * cellH;
            if (screenY < -50 || screenY > height + 50) continue;

            let deltaX = node.x - centerTx;
            if (deltaX < -gridWidth / 2) deltaX += gridWidth;
            if (deltaX > gridWidth / 2) deltaX -= gridWidth;
            const screenX = centerScreenX + deltaX * cellW;
            if (screenX < -50 || screenX > width + 50) continue;

            const drawX = screenX + cellW / 2;
            const drawY = screenY + cellH / 2;

            let icon = "❓";
            if (node.type === "gold") icon = "🟡";
            else if (node.type === "iron") icon = "⚪";
            else if (node.type === "cave") icon = "🕳️";
            else if (node.type === "coal") icon = "⚫";
            else if (node.type === "volcano") icon = "🌋";
            else if (node.type === "peak") icon = "🏔️";
            else if (node.type === "trench") icon = "🕳️";

            ctx.shadowColor = "black";
            ctx.shadowBlur = category === "feature" ? 4 : 2;
            ctx.fillText(icon, drawX, drawY);
            ctx.shadowBlur = 0;

            // Register Hit Target
            // Use slightly larger radius for easier hovering
            hitTargets.push({
                x: drawX,
                y: drawY,
                r: fontSize / 1.5,
                node: node,
            });
        }
    }

    function handleMouseMove(e: MouseEvent) {
        if (!canvas) return;
        const rect = canvas.getBoundingClientRect();
        const mouseX = e.clientX - rect.left;
        const mouseY = e.clientY - rect.top;

        let hit = null;
        // Check list in reverse (draw order: top on top)
        for (let i = hitTargets.length - 1; i >= 0; i--) {
            const t = hitTargets[i];
            if (!t) continue;
            const dx = mouseX - t.x;
            const dy = mouseY - t.y;
            if (dx * dx + dy * dy < t.r * t.r) {
                hit = t.node;
                break;
            }
        }

        if (hit !== hoveredItem) {
            hoveredItem = hit;
            dispatch("hover", hit); // Parent handles popup
            canvas.style.cursor = hit ? "help" : "default";
        }
    }

    // --- Helper Functions ---

    function hexToRgb(hex: string): { r: number; g: number; b: number } {
        const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
        return result
            ? {
                  r: parseInt(result[1], 16),
                  g: parseInt(result[2], 16),
                  b: parseInt(result[3], 16),
              }
            : { r: 0, g: 0, b: 0 };
    }

    function getColor(t: number, stops: { t: number; hex: string }[]): string {
        if (stops.length === 0) return "#000000";
        // Clamp t
        if (t <= 0) return stops[0].hex;
        if (t >= 1) return stops[stops.length - 1].hex;

        // Find stops
        for (let i = 0; i < stops.length - 1; i++) {
            const s1 = stops[i];
            const s2 = stops[i + 1];
            if (s1 && s2 && t >= s1.t && t <= s2.t) {
                // Interpolate
                const localT = (t - s1.t) / (s2.t - s1.t);
                const c1 = hexToRgb(s1.hex);
                const c2 = hexToRgb(s2.hex);
                const r = Math.round(c1.r + (c2.r - c1.r) * localT);
                const g = Math.round(c1.g + (c2.g - c1.g) * localT);
                const b = Math.round(c1.b + (c2.b - c1.b) * localT);
                return `rgb(${r}, ${g}, ${b})`;
            }
        }
        return stops[stops.length - 1].hex;
    }
</script>

<canvas
    bind:this={canvas}
    {width}
    {height}
    class="absolute inset-0 pointer-events-auto"
    on:mousemove={handleMouseMove}
    on:mouseleave={() => {
        hoveredItem = null;
        dispatch("hover", null);
    }}
/>
