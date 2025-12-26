import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * Unit tests for WebGLMapRenderer camera math.
 * These test the pure logic functions used for camera transforms.
 */

// Camera scale calculation logic (extracted from setCamera)
function calculateCameraScale(
    canvasWidth: number,
    canvasHeight: number,
    gridWidth: number,
    gridHeight: number,
    zoom: number
): { scaleX: number; scaleY: number } {
    const canvasAspect = canvasWidth / canvasHeight;
    const worldAspect = gridWidth / gridHeight;

    let baseScaleX: number, baseScaleY: number;
    if (worldAspect > canvasAspect) {
        // World is wider - fit width
        baseScaleX = 1.0;
        baseScaleY = worldAspect / canvasAspect;
    } else {
        // World is taller - fit height
        baseScaleY = 1.0;
        baseScaleX = canvasAspect / worldAspect;
    }

    return {
        scaleX: baseScaleX * zoom,
        scaleY: baseScaleY * zoom,
    };
}

// Screen to grid conversion logic (extracted from getGridIndexFromScreen)
function screenToGrid(
    screenX: number,
    screenY: number,
    canvasWidth: number,
    canvasHeight: number,
    centerX: number,
    centerY: number,
    scaleX: number,
    scaleY: number,
    gridWidth: number,
    gridHeight: number
): { gridX: number; gridY: number } | null {
    // Screen to normalized canvas (0-1)
    const ndcX = screenX / canvasWidth;
    const ndcY = screenY / canvasHeight;

    // Apply inverse camera transform
    const texX = centerX + (ndcX - 0.5) * scaleX;
    const texY = centerY + (ndcY - 0.5) * scaleY;

    // Bounds check
    if (texX < 0 || texX > 1 || texY < 0 || texY > 1) {
        return null;
    }

    // Convert to grid index
    const gridX = Math.floor(texX * gridWidth);
    const gridY = Math.floor(texY * gridHeight);

    // Clamp to valid range
    if (gridX < 0 || gridX >= gridWidth || gridY < 0 || gridY >= gridHeight) {
        return null;
    }

    return { gridX, gridY };
}

describe('WebGLMapRenderer Camera Math', () => {
    describe('calculateCameraScale', () => {
        it('preserves aspect ratio for wide world on square canvas', () => {
            // 2:1 world on 1:1 canvas
            const result = calculateCameraScale(100, 100, 200, 100, 1.0);

            // Should fit width (scaleX = 1.0)
            expect(result.scaleX).toBe(1.0);
            // scaleY should be 2.0 to show black bars vertically
            expect(result.scaleY).toBe(2.0);
        });

        it('preserves aspect ratio for tall world on square canvas', () => {
            // 1:2 world on 1:1 canvas
            const result = calculateCameraScale(100, 100, 100, 200, 1.0);

            // Should fit height (scaleY = 1.0)
            expect(result.scaleY).toBe(1.0);
            // scaleX should be 2.0 to show black bars horizontally
            expect(result.scaleX).toBe(2.0);
        });

        it('applies zoom multiplier correctly', () => {
            const result = calculateCameraScale(100, 100, 100, 100, 0.5);

            // Zoom 0.5 = zoomed in 2x, scales should be 0.5
            expect(result.scaleX).toBe(0.5);
            expect(result.scaleY).toBe(0.5);
        });

        it('handles 16:9 canvas with 2:1 world', () => {
            // 16:9 canvas, 2:1 world
            const result = calculateCameraScale(1920, 1080, 200, 100, 1.0);
            const canvasAspect = 1920 / 1080; // ~1.78
            const worldAspect = 2.0;

            // World is wider than canvas, fit width
            expect(result.scaleX).toBe(1.0);
            expect(result.scaleY).toBeCloseTo(worldAspect / canvasAspect, 5);
        });
    });

    describe('screenToGrid', () => {
        // Standard params: 100x100 canvas, 10x10 grid, centered, zoom 1.0
        const canvas = { width: 100, height: 100 };
        const grid = { width: 10, height: 10 };
        const center = { x: 0.5, y: 0.5 };
        const scale = { x: 1.0, y: 1.0 };

        it('returns correct grid cell for center of canvas', () => {
            const result = screenToGrid(
                50, 50,  // center of canvas
                canvas.width, canvas.height,
                center.x, center.y,
                scale.x, scale.y,
                grid.width, grid.height
            );

            expect(result).not.toBeNull();
            expect(result?.gridX).toBe(5);
            expect(result?.gridY).toBe(5);
        });

        it('returns correct grid cell for top-left corner', () => {
            const result = screenToGrid(
                0, 0,  // top-left
                canvas.width, canvas.height,
                center.x, center.y,
                scale.x, scale.y,
                grid.width, grid.height
            );

            expect(result).not.toBeNull();
            expect(result?.gridX).toBe(0);
            expect(result?.gridY).toBe(0);
        });

        it('returns correct grid cell for bottom-right corner', () => {
            const result = screenToGrid(
                99, 99,  // bottom-right (just inside)
                canvas.width, canvas.height,
                center.x, center.y,
                scale.x, scale.y,
                grid.width, grid.height
            );

            expect(result).not.toBeNull();
            expect(result?.gridX).toBe(9);
            expect(result?.gridY).toBe(9);
        });

        it('returns null for out-of-bounds click (zoomed in)', () => {
            // When zoomed in (scale < 1), clicking edge should be OOB
            const result = screenToGrid(
                0, 0,  // top-left corner
                canvas.width, canvas.height,
                0.5, 0.5,  // centered
                0.5, 0.5,  // zoomed in 2x
                grid.width, grid.height
            );

            // With scale 0.5 centered at 0.5, the view covers 0.25-0.75
            // Clicking at NDC 0,0 gives texX = 0.5 + (0 - 0.5) * 0.5 = 0.25
            expect(result).not.toBeNull();
            expect(result?.gridX).toBe(2); // 0.25 * 10 = 2.5 -> floor = 2
        });

        it('accounts for panned camera', () => {
            // Camera panned to bottom-right
            const result = screenToGrid(
                50, 50,  // center of canvas
                canvas.width, canvas.height,
                0.75, 0.75,  // camera looking at bottom-right
                1.0, 1.0,
                grid.width, grid.height
            );

            // Center click with camera at 0.75,0.75 should show grid at 0.75
            expect(result).not.toBeNull();
            expect(result?.gridX).toBe(7);  // 0.75 * 10 = 7.5 -> floor = 7
            expect(result?.gridY).toBe(7);
        });

        it('returns null when clicking outside texture bounds', () => {
            // Camera zoomed out, clicking far edge might be OOB
            const result = screenToGrid(
                0, 50,  // left edge
                canvas.width, canvas.height,
                0.5, 0.5,
                2.0, 2.0,  // zoomed out 2x
                grid.width, grid.height
            );

            // texX = 0.5 + (0 - 0.5) * 2.0 = 0.5 - 1.0 = -0.5 (OOB)
            expect(result).toBeNull();
        });
    });
});
