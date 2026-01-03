/**
 * HeightmapProvider - CPU height sampling for first-person mode.
 * Phase 4: Implements IHeightmapProvider for querying terrain elevation at any lat/lon.
 * 
 * Uses equirectangular projection to convert spherical coordinates to grid indices.
 */

import type { Texture } from "@babylonjs/core/Materials/Textures/texture";
import type { IHeightmapProvider } from "./interfaces";
import type { MapDataLayer } from "./BinaryDataParser";

/**
 * Provides CPU-side height sampling at any lat/lon coordinate.
 * Uses the binary grid data from MapDataLayer with bilinear interpolation.
 */
export class HeightmapProvider implements IHeightmapProvider {
    private dataLayer: MapDataLayer;
    private texture: Texture | null = null;
    private minElev: number = 0;
    private maxElev: number = 0;

    constructor(dataLayer: MapDataLayer) {
        this.dataLayer = dataLayer;
        this.calculateElevationRange();
    }

    /**
     * Scan the grid to find min/max elevation values.
     */
    private calculateElevationRange(): void {
        const width = this.dataLayer.width;
        const height = this.dataLayer.height;

        let min = Infinity;
        let max = -Infinity;

        for (let y = 0; y < height; y++) {
            for (let x = 0; x < width; x++) {
                const elev = this.dataLayer.getElevation(x, y);
                if (elev < min) min = elev;
                if (elev > max) max = elev;
            }
        }

        this.minElev = min === Infinity ? 0 : min;
        this.maxElev = max === -Infinity ? 0 : max;
    }

    /**
     * Convert latitude/longitude to grid coordinates.
     * Uses equirectangular projection:
     *   - lon: -180 to +180 maps to grid x: 0 to width
     *   - lat: +90 to -90 maps to grid y: 0 to height
     */
    private latLonToGrid(lat: number, lon: number): { x: number; y: number } {
        const width = this.dataLayer.width;
        const height = this.dataLayer.height;

        // Normalize longitude to 0..1 range
        // lon = -180 -> 0, lon = 0 -> 0.5, lon = 180 -> 1
        let normalizedLon = (lon + 180) / 360;
        // Handle wrapping
        normalizedLon = ((normalizedLon % 1) + 1) % 1;

        // Normalize latitude to 0..1 range  
        // lat = 90 (north) -> 0, lat = 0 -> 0.5, lat = -90 (south) -> 1
        const normalizedLat = (90 - lat) / 180;

        // Scale to grid dimensions
        const gridX = normalizedLon * width;
        const gridY = Math.max(0, Math.min(height - 0.001, normalizedLat * height));

        return { x: gridX, y: gridY };
    }

    /**
     * Get height at a specific lat/lon using bilinear interpolation.
     * 
     * @param lat Latitude in degrees (-90 to 90)
     * @param lon Longitude in degrees (-180 to 180)
     * @returns Elevation in meters
     */
    getHeightAt(lat: number, lon: number): number {
        const { x, y } = this.latLonToGrid(lat, lon);
        return this.sampleBilinear(x, y);
    }

    /**
     * Bilinear interpolation sampling at fractional grid coordinates.
     */
    private sampleBilinear(gridX: number, gridY: number): number {
        const width = this.dataLayer.width;
        const height = this.dataLayer.height;

        // Get integer cell indices
        const x0 = Math.floor(gridX);
        const y0 = Math.floor(gridY);
        const x1 = (x0 + 1) % width;  // Wrap horizontally
        const y1 = Math.min(y0 + 1, height - 1);  // Clamp vertically

        // Get fractional part
        const fx = gridX - x0;
        const fy = gridY - y0;

        // Sample four corners
        const e00 = this.dataLayer.getElevation(x0, y0);
        const e10 = this.dataLayer.getElevation(x1, y0);
        const e01 = this.dataLayer.getElevation(x0, y1);
        const e11 = this.dataLayer.getElevation(x1, y1);

        // Handle NaN (out of bounds)
        if (isNaN(e00) || isNaN(e10) || isNaN(e01) || isNaN(e11)) {
            return e00; // Fallback to nearest
        }

        // Bilinear interpolation
        const top = e00 * (1 - fx) + e10 * fx;
        const bottom = e01 * (1 - fx) + e11 * fx;
        return top * (1 - fy) + bottom * fy;
    }

    /**
     * Get the heightmap texture for GPU use.
     * Currently returns null - texture must be set externally.
     */
    getHeightmapTexture(): Texture | null {
        return this.texture;
    }

    /**
     * Set the heightmap texture (for GPU displacement).
     */
    setHeightmapTexture(texture: Texture): void {
        this.texture = texture;
    }

    /**
     * Get the elevation range for normalization.
     */
    getElevationRange(): { min: number; max: number } {
        return { min: this.minElev, max: this.maxElev };
    }

    /**
     * Get underlying data layer for direct grid access.
     */
    getDataLayer(): MapDataLayer {
        return this.dataLayer;
    }
}
