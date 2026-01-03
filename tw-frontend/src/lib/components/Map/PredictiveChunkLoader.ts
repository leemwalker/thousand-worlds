/**
 * PredictiveChunkLoader - Anticipatory chunk loading along descent path.
 * Phase 5: Production Readiness
 * 
 * Handles:
 * - Pre-loading chunks along camera trajectory
 * - Priority queue based on distance and direction
 * - LRU cache management (~50MB budget)
 * - Graceful fallback to lower LOD when high-res not ready
 */

import type { TileCoord } from './interfaces';
import { Vector3 } from "@babylonjs/core/Maths/math.vector";

// Cache budget in estimated chunk count (assuming ~1MB per chunk)
const DEFAULT_CACHE_SIZE = 50;
const PRELOAD_DISTANCE_MULTIPLIER = 2.0; // Load chunks 2x current visibility radius

export interface ChunkCacheEntry {
    coord: TileCoord;
    data: ArrayBuffer | null;       // Heightmap data
    textureData: ArrayBuffer | null; // Texture data
    lodLevel: number;
    lastAccessed: number;
    size: number;                    // Estimated size in bytes
}

export interface LoadPriority {
    coord: TileCoord;
    priority: number;    // Lower = higher priority
    reason: 'below' | 'frustum' | 'adjacent' | 'trajectory' | 'background';
}

export interface PredictiveLoaderOptions {
    maxCacheSize?: number;           // Max chunks in cache
    preloadAhead?: number;           // Seconds ahead to predict
    trajectoryWeight?: number;       // Weight for trajectory-based prioritization
}

/**
 * Calculates future camera position based on velocity.
 */
function predictPosition(
    currentPos: Vector3,
    velocity: Vector3,
    secondsAhead: number
): Vector3 {
    return currentPos.add(velocity.scale(secondsAhead));
}

/**
 * Converts world position to lat/lon.
 */
function worldToLatLon(position: Vector3): { lat: number; lon: number } {
    const radius = position.length();
    if (radius < 0.001) return { lat: 0, lon: 0 };

    const normalized = position.scale(1 / radius);
    const lat = Math.asin(normalized.y) * 180 / Math.PI;
    const lon = Math.atan2(normalized.z, normalized.x) * 180 / Math.PI;

    return { lat, lon };
}

/**
 * Predictive chunk loader with LRU cache.
 */
export class PredictiveChunkLoader {
    private cache: Map<string, ChunkCacheEntry> = new Map();
    private pendingRequests: Set<string> = new Set();
    private maxCacheSize: number;
    private preloadAhead: number;
    private trajectoryWeight: number;
    private onLoadChunk?: (coord: TileCoord, priority: number) => void;

    constructor(options: PredictiveLoaderOptions = {}) {
        this.maxCacheSize = options.maxCacheSize ?? DEFAULT_CACHE_SIZE;
        this.preloadAhead = options.preloadAhead ?? 2.0;
        this.trajectoryWeight = options.trajectoryWeight ?? 0.5;
    }

    /**
     * Set callback for chunk load requests.
     */
    setLoadCallback(callback: (coord: TileCoord, priority: number) => void): void {
        this.onLoadChunk = callback;
    }

    /**
     * Generate cache key from tile coordinates.
     */
    private getKey(coord: TileCoord): string {
        return `${coord.face}_${coord.level}_${coord.x}_${coord.y}`;
    }

    /**
     * Calculate priority queue for chunks based on current state.
     */
    calculatePriorityQueue(
        cameraPosition: Vector3,
        cameraVelocity: Vector3,
        targetLat: number,
        targetLon: number,
        currentLodLevel: number
    ): LoadPriority[] {
        const priorities: LoadPriority[] = [];
        const { lat: camLat, lon: camLon } = worldToLatLon(cameraPosition);

        // 1. Chunks directly below camera (highest priority)
        const belowCoord = this.latLonToTile(camLat, camLon, currentLodLevel);
        priorities.push({
            coord: belowCoord,
            priority: 0,
            reason: 'below'
        });

        // 2. Adjacent chunks to current position
        const adjacentCoords = this.getAdjacentTiles(belowCoord);
        adjacentCoords.forEach((coord, index) => {
            priorities.push({
                coord,
                priority: 10 + index,
                reason: 'adjacent'
            });
        });

        // 3. Chunks along predicted trajectory
        const predictedPos = predictPosition(cameraPosition, cameraVelocity, this.preloadAhead);
        const { lat: predLat, lon: predLon } = worldToLatLon(predictedPos);
        const trajectoryCoord = this.latLonToTile(predLat, predLon, currentLodLevel);

        if (this.getKey(trajectoryCoord) !== this.getKey(belowCoord)) {
            priorities.push({
                coord: trajectoryCoord,
                priority: 5 * (1 - this.trajectoryWeight),
                reason: 'trajectory'
            });
        }

        // 4. Chunks toward descent target (if descending)
        const targetCoord = this.latLonToTile(targetLat, targetLon, currentLodLevel);
        if (this.getKey(targetCoord) !== this.getKey(belowCoord)) {
            priorities.push({
                coord: targetCoord,
                priority: 3,
                reason: 'trajectory'
            });
        }

        // Sort by priority
        priorities.sort((a, b) => a.priority - b.priority);

        return priorities;
    }

    /**
     * Process chunk loading based on priority queue.
     */
    processQueue(queue: LoadPriority[], maxToLoad: number = 5): void {
        let loaded = 0;

        for (const item of queue) {
            if (loaded >= maxToLoad) break;

            const key = this.getKey(item.coord);

            // Skip if already cached or pending
            if (this.cache.has(key) || this.pendingRequests.has(key)) {
                continue;
            }

            // Request load
            this.pendingRequests.add(key);
            this.onLoadChunk?.(item.coord, item.priority);
            loaded++;
        }
    }

    /**
     * Store loaded chunk in cache.
     */
    cacheChunk(
        coord: TileCoord,
        heightmapData: ArrayBuffer,
        textureData: ArrayBuffer,
        lodLevel: number
    ): void {
        const key = this.getKey(coord);
        this.pendingRequests.delete(key);

        // Evict if cache is full
        while (this.cache.size >= this.maxCacheSize) {
            this.evictLRU();
        }

        const entry: ChunkCacheEntry = {
            coord,
            data: heightmapData,
            textureData,
            lodLevel,
            lastAccessed: Date.now(),
            size: heightmapData.byteLength + textureData.byteLength
        };

        this.cache.set(key, entry);
    }

    /**
     * Get cached chunk, updating access time.
     */
    getChunk(coord: TileCoord): ChunkCacheEntry | null {
        const key = this.getKey(coord);
        const entry = this.cache.get(key);

        if (entry) {
            entry.lastAccessed = Date.now();
            return entry;
        }

        return null;
    }

    /**
     * Check if chunk is cached (at any LOD).
     */
    hasChunk(coord: TileCoord): boolean {
        return this.cache.has(this.getKey(coord));
    }

    /**
     * Get fallback lower-LOD chunk if high-res not available.
     */
    getFallbackChunk(coord: TileCoord): ChunkCacheEntry | null {
        // Try current LOD first
        const entry = this.getChunk(coord);
        if (entry) return entry;

        // Try lower LOD levels
        for (let level = coord.level - 1; level >= 0; level--) {
            // Calculate parent tile coordinates
            const parentX = Math.floor(coord.x / 2);
            const parentY = Math.floor(coord.y / 2);
            const parentCoord: TileCoord = {
                ...coord,
                level,
                x: parentX,
                y: parentY
            };

            const parentEntry = this.getChunk(parentCoord);
            if (parentEntry) return parentEntry;
        }

        return null;
    }

    /**
     * Evict least recently used entry.
     */
    private evictLRU(): void {
        let oldestKey: string | null = null;
        let oldestTime = Infinity;

        for (const [key, entry] of this.cache.entries()) {
            if (entry.lastAccessed < oldestTime) {
                oldestTime = entry.lastAccessed;
                oldestKey = key;
            }
        }

        if (oldestKey) {
            this.cache.delete(oldestKey);
        }
    }

    /**
     * Convert lat/lon to tile coordinate.
     */
    private latLonToTile(lat: number, lon: number, level: number): TileCoord {
        // Simplified tile calculation (same as AltitudeChunkManager)
        const latRad = lat * Math.PI / 180;
        const lonRad = lon * Math.PI / 180;

        const x = Math.cos(latRad) * Math.cos(lonRad);
        const y = Math.sin(latRad);
        const z = Math.cos(latRad) * Math.sin(lonRad);

        const absX = Math.abs(x);
        const absY = Math.abs(y);
        const absZ = Math.abs(z);

        let face: number;
        let u: number, v: number;

        if (absY >= absX && absY >= absZ) {
            face = y > 0 ? 4 : 5;
            u = (x / absY + 1) / 2;
            v = (z / absY + 1) / 2;
        } else if (absX >= absZ) {
            face = x > 0 ? 3 : 2;
            u = (-z / absX + 1) / 2;
            v = (-y / absX + 1) / 2;
        } else {
            face = z > 0 ? 0 : 1;
            u = (x / absZ + 1) / 2;
            v = (-y / absZ + 1) / 2;
        }

        const tilesPerSide = Math.pow(2, level);
        const tileX = Math.min(tilesPerSide - 1, Math.floor(u * tilesPerSide));
        const tileY = Math.min(tilesPerSide - 1, Math.floor(v * tilesPerSide));

        return { face: face as any, level, x: tileX, y: tileY };
    }

    /**
     * Get adjacent tile coordinates.
     */
    private getAdjacentTiles(coord: TileCoord): TileCoord[] {
        const adjacent: TileCoord[] = [];
        const tilesPerSide = Math.pow(2, coord.level);

        const offsets = [
            { dx: 1, dy: 0 }, { dx: -1, dy: 0 },
            { dx: 0, dy: 1 }, { dx: 0, dy: -1 },
            { dx: 1, dy: 1 }, { dx: -1, dy: -1 },
            { dx: 1, dy: -1 }, { dx: -1, dy: 1 }
        ];

        for (const { dx, dy } of offsets) {
            const newX = coord.x + dx;
            const newY = coord.y + dy;

            // Skip out of bounds (cross-face would need more complex logic)
            if (newX >= 0 && newX < tilesPerSide && newY >= 0 && newY < tilesPerSide) {
                adjacent.push({
                    ...coord,
                    x: newX,
                    y: newY
                });
            }
        }

        return adjacent;
    }

    /**
     * Get cache statistics.
     */
    getStats(): { cached: number; pending: number; memoryMB: number } {
        let totalBytes = 0;
        for (const entry of this.cache.values()) {
            totalBytes += entry.size;
        }

        return {
            cached: this.cache.size,
            pending: this.pendingRequests.size,
            memoryMB: totalBytes / (1024 * 1024)
        };
    }

    /**
     * Clear all cached chunks.
     */
    clear(): void {
        this.cache.clear();
        this.pendingRequests.clear();
    }

    /**
     * Dispose of resources.
     */
    dispose(): void {
        this.clear();
    }
}
