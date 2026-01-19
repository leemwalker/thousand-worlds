/**
 * TileGlobeManager - Orchestrates cube-face tile loading for globe rendering.
 * Phase 3.2: Cube Tile System - BabylonGlobe Integration
 * 
 * This class manages the lifecycle of tile meshes based on camera position,
 * handling loading, caching, and disposal of tiles as the user zooms in/out.
 */

import type { Scene } from "@babylonjs/core/scene";
import type { Camera } from "@babylonjs/core/Cameras/camera";
import type { TransformNode } from "@babylonjs/core/Meshes/transformNode";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { TileManager } from './TileManager';
import { TileProvider } from './TileProvider';
import { TileMesh } from './TileMesh';
import { GPUTileMesh } from './GPUTileMesh';
import type { TerrainComputeShader } from './TerrainComputeShader';
import { CubeFace, type TileCoord, type ITileProvider } from './interfaces';

type TileMeshType = TileMesh | GPUTileMesh;

interface ActiveTile {
    mesh: TileMeshType;
    lastUsed: number;
}

export interface TileGlobeManagerOptions {
    maxLevel?: number;
    tileSize?: number;
    maxActiveTiles?: number;
    computeShader?: TerrainComputeShader;
    radius?: number;
    forceLevel?: number;
}

/**
 * Manages cube-face tiles for the 3D globe, handling loading and disposal.
 */
export class TileGlobeManager {
    private scene: Scene;
    private parentNode: TransformNode;
    private tileManager: TileManager;
    private tileProvider: TileProvider;
    private computeShader?: TerrainComputeShader;
    private activeTiles: Map<string, ActiveTile> = new Map();
    private loadingTiles: Set<string> = new Set();
    private maxActiveTiles: number;
    private enabled: boolean = false;
    private currentLevel: number = 0;
    private radius: number;
    private forceLevel?: number;

    constructor(
        scene: Scene,
        parentNode: TransformNode,
        sendCommand: (action: string, message?: string) => void,
        options: TileGlobeManagerOptions = {}
    ) {
        this.scene = scene;
        this.parentNode = parentNode;
        this.maxActiveTiles = options.maxActiveTiles ?? 50;
        this.computeShader = options.computeShader;
        this.radius = options.radius ?? 6371000.0;
        this.forceLevel = options.forceLevel;

        // Initialize tile provider with WebSocket command sender
        this.tileProvider = new TileProvider(scene, sendCommand);

        // Initialize tile manager for visibility calculations
        this.tileManager = new TileManager(this.tileProvider, {
            maxLevel: options.maxLevel ?? 4,
            tileSize: options.tileSize ?? 256
        });
    }

    /**
     * Generate cache key for tile coordinates.
     */
    private getTileKey(coord: TileCoord): string {
        return `${coord.face}_${coord.level}_${coord.x}_${coord.y}`;
    }

    /**
     * Enable tile loading (called when zoom crosses threshold).
     */
    enable(): void {
        this.enabled = true;
        console.log('[TileGlobeManager] Tile system enabled');
    }

    /**
     * Disable tile loading and dispose all tiles.
     */
    disable(): void {
        this.enabled = false;
        this.disposeAllTiles();
        console.log('[TileGlobeManager] Tile system disabled');
    }

    /**
     * Update tiles based on camera position.
     * Called each frame from the render loop.
     */
    update(camera: Camera): void {
        if (!this.enabled) return;

        // Calculate appropriate LOD level based on camera distance
        if (this.forceLevel !== undefined) {
            this.currentLevel = this.forceLevel;
        } else {
            this.currentLevel = this.tileManager.calculateLevel(camera);
        }

        // Get priority queue of visible tiles
        // Note: TileManager logic currently assumes standard globe (radius=1?). 
        // If forceLevel is set, we assume the camera is "visible" to these tiles.
        // For FPV sky, the planet is distant. TileManager's simple frustum check works if we set up the context right.
        const priorityQueue = this.tileManager.getPriorityQueue(camera, this.currentLevel);

        // Mark all tiles as potentially unused
        const now = Date.now();

        // Load tiles in priority order
        for (const { coord, distance } of priorityQueue) {
            const key = this.getTileKey(coord);

            if (this.activeTiles.has(key)) {
                // Update last used time
                this.activeTiles.get(key)!.lastUsed = now;
            } else if (!this.loadingTiles.has(key) && this.activeTiles.size < this.maxActiveTiles) {
                // Start loading this tile
                this.loadTile(coord);
            }
        }

        // Dispose tiles that haven't been used recently
        this.cleanupOldTiles(now - 10000); // 10 second cutoff
    }

    /**
     * Load a tile asynchronously.
     */
    private async loadTile(coord: TileCoord): Promise<void> {
        const key = this.getTileKey(coord);
        this.loadingTiles.add(key);

        try {
            const tileData = await this.tileProvider.getTile(coord);

            if (!this.enabled || this.activeTiles.has(key)) {
                // Disabled or already loaded (race condition)
                this.loadingTiles.delete(key);
                return;
            }

            // Create tile mesh
            let tileMesh: TileMeshType;

            if (this.computeShader) {
                tileMesh = new GPUTileMesh(
                    this.scene,
                    tileData.raw,
                    this.computeShader,
                    this.radius
                );
            } else {
                tileMesh = new TileMesh(
                    this.scene,
                    coord,
                    tileData.texture,
                    tileData.heightmap,
                    { displacementScale: 0.05 }
                );
            }

            // Access inner mesh for parenting
            tileMesh.mesh.parent = this.parentNode;

            // Track active tile
            this.activeTiles.set(key, {
                mesh: tileMesh,
                lastUsed: Date.now()
            });

            console.log(`[TileGlobeManager] Loaded tile ${key}`);
        } catch (error) {
            console.error(`[TileGlobeManager] Failed to load tile ${key}:`, error);
        } finally {
            this.loadingTiles.delete(key);
        }
    }

    /**
     * Dispose tiles that haven't been used since cutoff time.
     */
    private cleanupOldTiles(cutoffTime: number): void {
        for (const [key, tile] of this.activeTiles.entries()) {
            if (tile.lastUsed < cutoffTime) {
                // Both implementation have dispose method
                tile.mesh.dispose();
                this.activeTiles.delete(key);
                console.log(`[TileGlobeManager] Disposed tile ${key}`);
            }
        }
    }

    /**
     * Dispose all active tiles.
     */
    private disposeAllTiles(): void {
        for (const [key, tile] of this.activeTiles.entries()) {
            tile.mesh.dispose();
        }
        this.activeTiles.clear();
        this.loadingTiles.clear();
    }

    /**
     * Handle incoming tile response from WebSocket.
     */

    /**
     * Get current LOD level.
     */
    getCurrentLevel(): number {
        return this.currentLevel;
    }

    /**
     * Get statistics for debugging.
     */
    getStats(): { active: number; loading: number; level: number } {
        return {
            active: this.activeTiles.size,
            loading: this.loadingTiles.size,
            level: this.currentLevel
        };
    }

    /**
     * Dispose the manager and all resources.
     */
    dispose(): void {
        this.disable();
        this.tileProvider.clearCache();
    }
}
