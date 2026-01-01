/**
 * TileProvider - WebSocket-based tile provider for cube-face tiles.
 * Phase 3.2: Cube Tile System
 * 
 * Implements ITileProvider interface to request and cache tiles from backend.
 */

import type { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { RawTexture } from "@babylonjs/core/Materials/Textures/rawTexture";
import type { Scene } from "@babylonjs/core/scene";
import { type TileCoord, type ITileProvider, CubeFace } from './interfaces';

interface TileData {
    texture: Texture;
    heightmap: Texture;
}

interface PendingRequest {
    resolve: (data: TileData) => void;
    reject: (error: Error) => void;
    timestamp: number;
}

export class TileProvider implements ITileProvider {
    private scene: Scene;
    private sendCommand: (action: string, message?: string) => void;
    private cache: Map<string, TileData> = new Map();
    private pendingRequests: Map<string, PendingRequest> = new Map();
    private preloadQueue: TileCoord[] = [];
    private isPreloading = false;

    constructor(
        scene: Scene,
        sendCommand: (action: string, message?: string) => void
    ) {
        this.scene = scene;
        this.sendCommand = sendCommand;
    }

    /**
     * Generate cache key for tile coordinates.
     */
    private getTileKey(coord: TileCoord): string {
        return `${coord.face}_${coord.level}_${coord.x}_${coord.y}`;
    }

    /**
     * Request a tile from the backend or return cached version.
     */
    async getTile(coord: TileCoord): Promise<{ texture: Texture; heightmap: Texture }> {
        const key = this.getTileKey(coord);

        // Return cached tile if available
        if (this.cache.has(key)) {
            return this.cache.get(key)!;
        }

        // Check if request is already pending
        if (this.pendingRequests.has(key)) {
            // Wait for existing request
            return new Promise((resolve, reject) => {
                const existing = this.pendingRequests.get(key)!;
                const originalResolve = existing.resolve;
                existing.resolve = (data) => {
                    originalResolve(data);
                    resolve(data);
                };
            });
        }

        // Create new request
        return new Promise((resolve, reject) => {
            this.pendingRequests.set(key, {
                resolve,
                reject,
                timestamp: Date.now()
            });

            // Send request to backend
            const message = `${coord.face},${coord.level},${coord.x},${coord.y},256`;
            this.sendCommand('world_tile', message);

            // Timeout after 30 seconds
            setTimeout(() => {
                if (this.pendingRequests.has(key)) {
                    this.pendingRequests.delete(key);
                    reject(new Error(`Tile request timeout: ${key}`));
                }
            }, 30000);
        });
    }

    /**
     * Handle incoming tile data from WebSocket.
     */
    handleTileResponse(
        metadata: {
            face: number;
            level: number;
            x: number;
            y: number;
            width: number;
            height: number;
            imageSize: number;
            heightmapSize: number;
        },
        imageData: Uint8Array,
        heightmapData: Uint8Array
    ): void {
        const coord: TileCoord = {
            face: metadata.face as CubeFace,
            level: metadata.level,
            x: metadata.x,
            y: metadata.y
        };
        const key = this.getTileKey(coord);

        try {
            // Create textures from the raw data
            // For PNG data, we need to decode it first
            const tileData = this.createTexturesFromPNG(
                imageData,
                heightmapData,
                metadata.width,
                metadata.height
            );

            // Cache the tile
            this.cache.set(key, tileData);

            // Resolve pending request
            const pending = this.pendingRequests.get(key);
            if (pending) {
                pending.resolve(tileData);
                this.pendingRequests.delete(key);
            }

            console.log(`[TileProvider] Tile ${key} loaded and cached`);
        } catch (error) {
            console.error(`[TileProvider] Failed to process tile ${key}:`, error);
            const pending = this.pendingRequests.get(key);
            if (pending) {
                pending.reject(error as Error);
                this.pendingRequests.delete(key);
            }
        }
    }

    /**
     * Create BabylonJS textures from PNG data.
     */
    private createTexturesFromPNG(
        imageData: Uint8Array,
        heightmapData: Uint8Array,
        width: number,
        height: number
    ): TileData {
        // Create blob URLs for the PNG data
        // Slice to create owned ArrayBuffer copies
        const imageBlob = new Blob([imageData.slice()], { type: 'image/png' });
        const heightmapBlob = new Blob([heightmapData.slice()], { type: 'image/png' });

        const imageUrl = URL.createObjectURL(imageBlob);
        const heightmapUrl = URL.createObjectURL(heightmapBlob);

        // Create textures from URLs
        const texture = new (RawTexture as any).LoadFromUrl(
            imageUrl,
            "tileTexture",
            this.scene,
            false, // generateMipMaps
            false, // invertY
            undefined, // samplingMode
            () => URL.revokeObjectURL(imageUrl)
        );

        const heightmap = new (RawTexture as any).LoadFromUrl(
            heightmapUrl,
            "tileHeightmap",
            this.scene,
            false,
            false,
            undefined,
            () => URL.revokeObjectURL(heightmapUrl)
        );

        return { texture, heightmap };
    }

    /**
     * Preload tiles for smooth transitions.
     */
    preloadTiles(coords: TileCoord[]): void {
        this.preloadQueue.push(...coords.filter(c => !this.isCached(c)));
        this.processPreloadQueue();
    }

    /**
     * Process the preload queue (load tiles in background).
     */
    private async processPreloadQueue(): Promise<void> {
        if (this.isPreloading || this.preloadQueue.length === 0) return;

        this.isPreloading = true;

        while (this.preloadQueue.length > 0) {
            const coord = this.preloadQueue.shift()!;
            try {
                await this.getTile(coord);
            } catch (error) {
                console.warn(`[TileProvider] Preload failed for tile:`, coord, error);
            }
            // Small delay to avoid flooding the server
            await new Promise(r => setTimeout(r, 50));
        }

        this.isPreloading = false;
    }

    /**
     * Check if a tile is cached.
     */
    isCached(coord: TileCoord): boolean {
        return this.cache.has(this.getTileKey(coord));
    }

    /**
     * Clear all cached tiles.
     */
    clearCache(): void {
        // Dispose textures
        for (const data of this.cache.values()) {
            data.texture.dispose();
            data.heightmap.dispose();
        }
        this.cache.clear();
    }

    /**
     * Get cache statistics.
     */
    getCacheStats(): { size: number; pending: number } {
        return {
            size: this.cache.size,
            pending: this.pendingRequests.size
        };
    }
}
