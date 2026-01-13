/**
 * AltitudeChunkManager - Manages terrain chunks based on camera altitude.
 * Phase 4: First-Person View Implementation
 * 
 * Handles chunk loading strategy at different altitude stages:
 * - 1000ft (flying): Low-res chunks in 5km radius
 * - 100ft (low): Medium-res chunks in 1km radius
 * - 10ft (ground): High-res chunks in 200m radius
 */

import type { Scene } from "@babylonjs/core/scene";
import type { Camera } from "@babylonjs/core/Cameras/camera";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { FPS_ALTITUDES } from './FPSTransitionController';
import type { TileCoord, CubeFace } from './interfaces';

// LOD configuration per altitude stage
interface LODConfig {
    tileLevel: number;      // Tile LOD level to request
    loadRadius: number;     // Radius around player to load (in planet radius units)
    maxChunks: number;      // Maximum chunks to keep loaded
}

const LOD_CONFIGS: Record<string, LODConfig> = {
    flying: {
        tileLevel: 1,       // Low detail tiles
        loadRadius: 0.5,    // ~3km equivalent
        maxChunks: 16,
    },
    low: {
        tileLevel: 2,       // Medium detail tiles
        loadRadius: 0.15,   // ~1km equivalent
        maxChunks: 25,
    },
    ground: {
        tileLevel: 3,       // High detail tiles
        loadRadius: 0.03,   // ~200m equivalent
        maxChunks: 36,
    }
};

interface ChunkRequest {
    coord: TileCoord;
    priority: number;       // Lower = higher priority
    distance: number;       // Distance from camera
}

export type AltitudeStage = 'orbit' | 'flying' | 'low' | 'ground';

/**
 * Determines altitude stage from camera altitude.
 */
function getAltitudeStage(altitude: number): AltitudeStage {
    if (altitude > FPS_ALTITUDES.FLYING * 1.5) return 'orbit';
    if (altitude > FPS_ALTITUDES.LOW * 1.5) return 'flying';
    if (altitude > FPS_ALTITUDES.GROUND * 1.5) return 'low';
    return 'ground';
}

/**
 * Convert lat/lon to spherical TileCoord.
 */
function latLonToTileCoord(lat: number, lon: number, level: number): TileCoord {
    // Determine cube face from lat/lon
    const latRad = lat * Math.PI / 180;
    const lonRad = lon * Math.PI / 180;

    // Convert to 3D unit vector
    const x = Math.cos(latRad) * Math.cos(lonRad);
    const y = Math.sin(latRad);
    const z = Math.cos(latRad) * Math.sin(lonRad);

    // Determine dominant axis for cube face
    const absX = Math.abs(x);
    const absY = Math.abs(y);
    const absZ = Math.abs(z);

    let face: number;
    let u: number, v: number;

    if (absY >= absX && absY >= absZ) {
        // Top or bottom face
        face = y > 0 ? 4 : 5; // TOP=4, BOTTOM=5
        u = (x / absY + 1) / 2;
        v = (z / absY + 1) / 2;
    } else if (absX >= absZ) {
        // Left or right face
        face = x > 0 ? 3 : 2; // RIGHT=3, LEFT=2
        u = (-z / absX + 1) / 2;
        v = (-y / absX + 1) / 2;
    } else {
        // Front or back face
        face = z > 0 ? 0 : 1; // FRONT=0, BACK=1
        u = (x / absZ + 1) / 2;
        v = (-y / absZ + 1) / 2;
    }

    // Convert UV to tile coordinates at given level
    const tilesPerSide = Math.pow(2, level);
    const tileX = Math.min(tilesPerSide - 1, Math.floor(u * tilesPerSide));
    const tileY = Math.min(tilesPerSide - 1, Math.floor(v * tilesPerSide));

    return {
        face: face as CubeFace,
        level,
        x: tileX,
        y: tileY
    };
}

/**
 * Get tile coordinates in a grid around a center position.
 */
function getTilesAround(centerLat: number, centerLon: number, level: number, radius: number): TileCoord[] {
    const tiles: TileCoord[] = [];
    const tilesPerSide = Math.pow(2, level);

    // Calculate angular step per tile at this level
    const tileAngle = 360 / tilesPerSide / 6; // Rough approximation

    // Calculate number of tiles to load based on radius
    const tileRadius = Math.ceil(radius / tileAngle);

    // Get center tile
    const centerTile = latLonToTileCoord(centerLat, centerLon, level);
    tiles.push(centerTile);

    // Get surrounding tiles
    for (let dx = -tileRadius; dx <= tileRadius; dx++) {
        for (let dy = -tileRadius; dy <= tileRadius; dy++) {
            if (dx === 0 && dy === 0) continue;

            // Offset lat/lon
            const offsetLat = centerLat + dy * tileAngle;
            const offsetLon = centerLon + dx * tileAngle;

            // Skip if out of bounds
            if (offsetLat < -90 || offsetLat > 90) continue;

            // Wrap longitude
            const wrappedLon = ((offsetLon + 180) % 360) - 180;

            const tile = latLonToTileCoord(offsetLat, wrappedLon, level);

            // Avoid duplicates
            const key = `${tile.face}_${tile.level}_${tile.x}_${tile.y}`;
            if (!tiles.some(t => `${t.face}_${t.level}_${t.x}_${t.y}` === key)) {
                tiles.push(tile);
            }
        }
    }

    return tiles;
}

export interface AltitudeChunkManagerOptions {
    onChunkRequest?: (coords: TileCoord[]) => void;
    onStageChange?: (stage: AltitudeStage) => void;
}

/**
 * Manages terrain chunk loading based on camera altitude and position.
 */
export class AltitudeChunkManager {
    private scene: Scene;
    private currentStage: AltitudeStage = 'orbit';
    private targetLat: number = 0;
    private targetLon: number = 0;
    private loadedChunks: Set<string> = new Set();
    private pendingChunks: Set<string> = new Set();
    private onChunkRequest?: ((coords: TileCoord[]) => void) | undefined;
    private onStageChange?: ((stage: AltitudeStage) => void) | undefined;

    constructor(scene: Scene, options: AltitudeChunkManagerOptions = {}) {
        this.scene = scene;
        this.onChunkRequest = options.onChunkRequest;
        this.onStageChange = options.onStageChange;
    }

    /**
     * Set target position for chunk loading.
     */
    setTarget(lat: number, lon: number): void {
        this.targetLat = lat;
        this.targetLon = lon;
    }

    /**
     * Update chunk loading based on camera position.
     */
    update(camera: Camera): void {
        // Calculate altitude (distance from surface)
        const pos = camera.position;
        const distance = Math.sqrt(pos.x * pos.x + pos.y * pos.y + pos.z * pos.z);
        const altitude = distance - 1.0; // Assuming planet radius = 1.0

        // Determine stage
        const newStage = getAltitudeStage(altitude);

        if (newStage !== this.currentStage) {
            this.currentStage = newStage;
            this.onStageChange?.(newStage);
            console.log(`[AltitudeChunkManager] Stage changed to: ${newStage}`);
        }

        // Skip loading in orbit mode
        if (this.currentStage === 'orbit') {
            return;
        }

        // Get LOD config for current stage
        const config = LOD_CONFIGS[this.currentStage];
        if (!config) return;

        // Calculate chunks to load
        const neededTiles = getTilesAround(
            this.targetLat,
            this.targetLon,
            config.tileLevel,
            config.loadRadius * 360 // Convert to degrees
        );

        // Filter to max chunks
        const tilesToLoad = neededTiles.slice(0, config.maxChunks);

        // Find new chunks to request
        const newChunks: TileCoord[] = [];
        for (const tile of tilesToLoad) {
            const key = `${tile.face}_${tile.level}_${tile.x}_${tile.y}`;
            if (!this.loadedChunks.has(key) && !this.pendingChunks.has(key)) {
                newChunks.push(tile);
                this.pendingChunks.add(key);
            }
        }

        // Request new chunks
        if (newChunks.length > 0 && this.onChunkRequest) {
            this.onChunkRequest(newChunks);
        }
    }

    /**
     * Mark a chunk as loaded.
     */
    markLoaded(coord: TileCoord): void {
        const key = `${coord.face}_${coord.level}_${coord.x}_${coord.y}`;
        this.pendingChunks.delete(key);
        this.loadedChunks.add(key);
    }

    /**
     * Get current altitude stage.
     */
    getCurrentStage(): AltitudeStage {
        return this.currentStage;
    }

    /**
     * Get loading statistics.
     */
    getStats(): { stage: AltitudeStage; loaded: number; pending: number } {
        return {
            stage: this.currentStage,
            loaded: this.loadedChunks.size,
            pending: this.pendingChunks.size
        };
    }

    /**
     * Clear all loaded chunks (for stage transitions).
     */
    clearChunks(): void {
        this.loadedChunks.clear();
        this.pendingChunks.clear();
    }

    /**
     * Dispose manager resources.
     */
    dispose(): void {
        this.clearChunks();
    }
}

// Export helper functions for testing
export { getAltitudeStage, latLonToTileCoord, getTilesAround };
