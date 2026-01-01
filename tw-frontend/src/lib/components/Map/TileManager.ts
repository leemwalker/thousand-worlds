/**
 * TileManager - Cube-face tile visibility culling and priority loading.
 * Phase 3.2: Cube Tile System
 */

import type { Camera } from "@babylonjs/core/Cameras/camera";
import { CubeFace, TileCoord, ITileProvider } from './interfaces';

export interface TileManagerOptions {
    maxLevel?: number;
    tileSize?: number;
}

export interface PriorityTile {
    coord: TileCoord;
    distance: number;
}

interface Vector3Like {
    x: number;
    y: number;
    z: number;
}

/**
 * Manages cube-face tiles for streaming high-resolution terrain.
 */
export class TileManager {
    private provider: ITileProvider;
    private maxLevel: number;
    private tileSize: number;

    constructor(provider: ITileProvider, options: TileManagerOptions = {}) {
        this.provider = provider;
        this.maxLevel = options.maxLevel ?? 4;
        this.tileSize = options.tileSize ?? 256;
    }

    /**
     * Get maximum supported LOD level.
     */
    getMaxLevel(): number {
        return this.maxLevel;
    }

    /**
     * Get all tiles for a given level (unfiltered).
     */
    getVisibleTiles(camera: Camera, level: number): TileCoord[] {
        const tiles: TileCoord[] = [];
        const tilesPerSide = 1 << level; // 2^level

        for (let face = 0; face < 6; face++) {
            for (let y = 0; y < tilesPerSide; y++) {
                for (let x = 0; x < tilesPerSide; x++) {
                    tiles.push({
                        face: face as CubeFace,
                        level,
                        x,
                        y
                    });
                }
            }
        }

        return tiles;
    }

    /**
     * Get tiles filtered by camera frustum (back-face culling).
     */
    getVisibleTilesFiltered(camera: Camera, level: number): TileCoord[] {
        const allTiles = this.getVisibleTiles(camera, level);
        const cameraPos = camera.position as Vector3Like;

        // Get camera direction
        const ray = camera.getForwardRay(1);
        const cameraDir = ray.direction as Vector3Like;

        return allTiles.filter(coord => {
            const center = this.getTileCenter(coord);

            // Vector from camera to tile center
            const toTile = {
                x: center.x - cameraPos.x,
                y: center.y - cameraPos.y,
                z: center.z - cameraPos.z
            };

            // Face normal (for back-face culling)
            const faceNormal = this.getFaceNormal(coord.face);

            // Tile is visible if facing camera (dot product > 0)
            const dot = toTile.x * faceNormal.x + toTile.y * faceNormal.y + toTile.z * faceNormal.z;

            return dot < 0; // Tile faces camera if normal points toward it
        });
    }

    /**
     * Get tiles sorted by distance to camera (priority queue).
     */
    getPriorityQueue(camera: Camera, level: number): PriorityTile[] {
        const tiles = this.getVisibleTiles(camera, level);
        const cameraPos = camera.position as Vector3Like;

        const priorityTiles: PriorityTile[] = tiles.map(coord => {
            const center = this.getTileCenter(coord);
            const dx = center.x - cameraPos.x;
            const dy = center.y - cameraPos.y;
            const dz = center.z - cameraPos.z;
            const distance = Math.sqrt(dx * dx + dy * dy + dz * dz);

            return { coord, distance };
        });

        // Sort by distance (closest first)
        priorityTiles.sort((a, b) => a.distance - b.distance);

        return priorityTiles;
    }

    /**
     * Calculate appropriate LOD level based on camera distance.
     */
    calculateLevel(camera: Camera): number {
        const cameraPos = camera.position as Vector3Like;
        const distance = Math.sqrt(
            cameraPos.x * cameraPos.x +
            cameraPos.y * cameraPos.y +
            cameraPos.z * cameraPos.z
        );

        // Distance thresholds for LOD levels (assuming globe radius = 1)
        // Far (>5): Level 0
        // Medium (2-5): Level 1
        // Close (1.5-2): Level 2
        // Very close (1.2-1.5): Level 3
        // Surface (<1.2): Level 4+

        if (distance > 5) return 0;
        if (distance > 2) return 1;
        if (distance > 1.5) return 2;
        if (distance > 1.2) return 3;

        // For very close, calculate based on distance
        const level = Math.floor(4 + (1.2 - distance) * 10);
        return Math.min(level, this.maxLevel);
    }

    /**
     * Get 3D center position of a tile on the unit sphere.
     */
    getTileCenter(coord: TileCoord): Vector3Like {
        const tilesPerSide = 1 << coord.level;
        const tileSize = 1.0 / tilesPerSide;

        // UV center within face
        const u = (coord.x + 0.5) * tileSize;
        const v = (coord.y + 0.5) * tileSize;

        // Convert UV to [-1, 1] cube coordinates
        const cu = 2 * u - 1;
        const cv = 2 * v - 1;

        // Map face UV to 3D cube position
        let x: number, y: number, z: number;
        switch (coord.face) {
            case CubeFace.FRONT:
                x = cu; y = -cv; z = 1;
                break;
            case CubeFace.BACK:
                x = -cu; y = -cv; z = -1;
                break;
            case CubeFace.LEFT:
                x = -1; y = -cv; z = -cu;
                break;
            case CubeFace.RIGHT:
                x = 1; y = -cv; z = cu;
                break;
            case CubeFace.TOP:
                x = cu; y = 1; z = cv;
                break;
            case CubeFace.BOTTOM:
                x = cu; y = -1; z = -cv;
                break;
            default:
                x = 0; y = 0; z = 1;
        }

        // Normalize to unit sphere
        const length = Math.sqrt(x * x + y * y + z * z);
        return {
            x: x / length,
            y: y / length,
            z: z / length
        };
    }

    /**
     * Get outward-facing normal for a cube face.
     */
    private getFaceNormal(face: CubeFace): Vector3Like {
        switch (face) {
            case CubeFace.FRONT: return { x: 0, y: 0, z: 1 };
            case CubeFace.BACK: return { x: 0, y: 0, z: -1 };
            case CubeFace.LEFT: return { x: -1, y: 0, z: 0 };
            case CubeFace.RIGHT: return { x: 1, y: 0, z: 0 };
            case CubeFace.TOP: return { x: 0, y: 1, z: 0 };
            case CubeFace.BOTTOM: return { x: 0, y: -1, z: 0 };
            default: return { x: 0, y: 0, z: 1 };
        }
    }
}
