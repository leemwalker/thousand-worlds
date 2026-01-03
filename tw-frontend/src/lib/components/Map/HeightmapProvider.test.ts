/**
 * HeightmapProvider Tests - TDD Red Phase
 * Tests for CPU height sampling at lat/lon coordinates.
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { HeightmapProvider } from './HeightmapProvider';
import { MapDataLayer, type ParsedGridData } from './BinaryDataParser';

/**
 * Create mock grid data for testing
 * Default: 4x2 grid (simple case for verifying coordinate mapping)
 */
function createMockGridData(
    width: number = 4,
    height: number = 2,
    elevationFn: (x: number, y: number) => number = () => 100
): ParsedGridData {
    const size = width * height;
    const elevations = new Float32Array(size);
    const biomes = new Uint8Array(size);

    for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
            elevations[y * width + x] = elevationFn(x, y);
            biomes[y * width + x] = 1; // Ocean default
        }
    }

    return { width, height, elevations, biomes };
}

describe('HeightmapProvider', () => {
    describe('constructor', () => {
        it('should create provider from MapDataLayer', () => {
            const gridData = createMockGridData();
            const dataLayer = new MapDataLayer(gridData);
            const provider = new HeightmapProvider(dataLayer);

            expect(provider).toBeDefined();
        });

        it('should calculate elevation range from data', () => {
            // Grid with varying elevations: 0 to 1000
            const gridData = createMockGridData(4, 2, (x, y) => x * 100 + y * 500);
            const dataLayer = new MapDataLayer(gridData);
            const provider = new HeightmapProvider(dataLayer);

            const range = provider.getElevationRange();
            expect(range.min).toBe(0);   // x=0, y=0 -> 0
            expect(range.max).toBe(800); // x=3, y=1 -> 300 + 500 = 800
        });
    });

    describe('getHeightAt - coordinate conversion', () => {
        let provider: HeightmapProvider;

        beforeEach(() => {
            // 4x2 grid: each cell has unique elevation based on grid position
            // Cell values: [0,1,2,3] for row 0, [4,5,6,7] for row 1
            const gridData = createMockGridData(4, 2, (x, y) => y * 4 + x);
            const dataLayer = new MapDataLayer(gridData);
            provider = new HeightmapProvider(dataLayer);
        });

        it('should return correct height at equator/prime meridian (lat=0, lon=0)', () => {
            // lat=0 -> middle of grid height (y = 0.5 in normalized terms)
            // lon=0 -> middle of grid width (x = 0.5 in normalized terms)
            // For 4x2 grid: lon=0 maps to x=2, lat=0 maps to y=1
            const height = provider.getHeightAt(0, 0);
            // y=1, x=2 -> 1*4 + 2 = 6
            expect(height).toBeCloseTo(6, 1);
        });

        it('should return correct height at north pole (lat=90)', () => {
            // lat=90 -> top of grid (y = 0)
            const height = provider.getHeightAt(90, 0);
            // y=0, x=2 (lon=0) -> 0*4 + 2 = 2
            expect(height).toBeCloseTo(2, 1);
        });

        it('should return correct height at south pole (lat=-90)', () => {
            // lat=-90 -> bottom of grid (y = height-1)
            const height = provider.getHeightAt(-90, 0);
            // y=1, x=2 -> 1*4 + 2 = 6
            expect(height).toBeCloseTo(6, 1);
        });

        it('should wrap at date line (lon=180 and lon=-180 same)', () => {
            const height180 = provider.getHeightAt(0, 180);
            const heightNeg180 = provider.getHeightAt(0, -180);
            expect(height180).toBeCloseTo(heightNeg180, 5);
        });

        it('should handle western hemisphere correctly (lon=-90)', () => {
            // lon=-90 -> x index at 1/4 of grid width
            // For 4-wide grid: x = 1
            const height = provider.getHeightAt(0, -90);
            // y=1, x=1 -> 1*4 + 1 = 5
            expect(height).toBeCloseTo(5, 1);
        });

        it('should handle eastern hemisphere correctly (lon=90)', () => {
            // lon=90 -> x index at 3/4 of grid width
            // For 4-wide grid: x = 3
            const height = provider.getHeightAt(0, 90);
            // y=1, x=3 -> 1*4 + 3 = 7
            expect(height).toBeCloseTo(7, 1);
        });
    });

    describe('getHeightAt - interpolation', () => {
        it('should interpolate smoothly between grid cells', () => {
            // Grid with clear gradient: elevation = x * 100
            const gridData = createMockGridData(10, 5, (x) => x * 100);
            const dataLayer = new MapDataLayer(gridData);
            const provider = new HeightmapProvider(dataLayer);

            // Sample at position between cells
            // With bilinear interpolation, values should change smoothly
            const h1 = provider.getHeightAt(0, -180);  // Left edge
            const h2 = provider.getHeightAt(0, 0);     // Center
            const h3 = provider.getHeightAt(0, 180);   // Right edge (wraps to left)

            // Center should be higher than left (grid increases left to right)
            expect(h2).toBeGreaterThan(h1);
        });
    });

    describe('getElevationRange', () => {
        it('should return correct min and max', () => {
            const gridData = createMockGridData(4, 4, (x, y) => -100 + x * 100 + y * 50);
            const dataLayer = new MapDataLayer(gridData);
            const provider = new HeightmapProvider(dataLayer);

            const range = provider.getElevationRange();
            expect(range.min).toBe(-100); // x=0, y=0
            expect(range.max).toBe(350);  // x=3, y=3 -> -100 + 300 + 150
        });
    });

    describe('getHeightmapTexture', () => {
        it('should return null when no texture set', () => {
            const gridData = createMockGridData();
            const dataLayer = new MapDataLayer(gridData);
            const provider = new HeightmapProvider(dataLayer);

            expect(provider.getHeightmapTexture()).toBeNull();
        });
    });
});
