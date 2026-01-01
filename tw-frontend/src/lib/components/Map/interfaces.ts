/**
 * Shared interfaces for 3D Planet Exploration system.
 * Phase 1: Shader-Based Displacement
 * Phase 2: Cube Tile System
 * Phase 3: Chunked Spherical Terrain
 */

import type { Scene, Texture, Mesh, Camera, ShaderMaterial } from "@babylonjs/core";

// =============================================================================
// Phase 1: Shader-Based Displacement
// =============================================================================

/**
 * Provider for heightmap data used in GPU displacement mapping.
 */
export interface IHeightmapProvider {
    /** Get the heightmap texture for shader use */
    getHeightmapTexture(): Texture | null;

    /** Get height value at a specific lat/lon (for CPU fallback) */
    getHeightAt(lat: number, lon: number): number;

    /** Get min/max elevation for normalization */
    getElevationRange(): { min: number; max: number };
}

/**
 * Provider for displacement shader materials.
 */
export interface IShaderProvider {
    /** Create a displacement shader material for the given scene */
    createMaterial(scene: Scene, heightmap: Texture): ShaderMaterial;

    /** Update the heightmap texture (for streaming) */
    updateHeightmap(texture: Texture): void;

    /** Set terrain displacement scale */
    setDisplacementScale(scale: number): void;
}

/**
 * Manager for Level of Detail mesh swapping.
 */


export interface ILODManager {
    /** Get appropriate LOD level for given camera distance */
    getLODLevel(distance: number): number;

    /** Get mesh for specific LOD level */
    getMesh(level: number): Mesh | null;

    /** Update LOD based on current camera */
    update(camera: Camera): void;
}

// =============================================================================
// Phase 2: Cube Tile System (Placeholders)
// =============================================================================

/**
 * Cube face enumeration for tile system.
 */
export enum CubeFace {
    FRONT = 0,
    BACK = 1,
    LEFT = 2,
    RIGHT = 3,
    TOP = 4,
    BOTTOM = 5,
}

/**
 * Tile coordinate in the quadtree pyramid.
 */
export interface TileCoord {
    face: CubeFace;
    level: number;
    x: number;
    y: number;
}

/**
 * Provider for tile data.
 */
export interface ITileProvider {
    /** Get tile texture and heightmap */
    getTile(coord: TileCoord): Promise<{ texture: Texture; heightmap: Texture }>;

    /** Preload tiles for smooth transitions */
    preloadTiles(coords: TileCoord[]): void;

    /** Check if tile is cached */
    isCached(coord: TileCoord): boolean;
}

// =============================================================================
// Phase 3: First-Person Terrain (Placeholders)
// =============================================================================

/**
 * Interface for terrain chunk mesh management.
 */
export interface ITerrainChunk {
    /** Get the mesh for rendering */
    getMesh(): Mesh;

    /** Dispose of resources */
    dispose(): void;

    /** Chunk world position */
    position: { lat: number; lon: number };

    /** Current LOD level */
    lodLevel: number;
}

/**
 * Interface for first-person camera controller.
 */
export interface IPlayerController {
    /** Handle input for movement */
    handleInput(deltaTime: number): void;

    /** Get current world position */
    getPosition(): { lat: number; lon: number; altitude: number };

    /** Get the camera instance */
    getCamera(): Camera;
}

/**
 * View mode for transitioning between orbit and ground views.
 */
export enum ViewMode {
    ORBIT = "orbit",
    TRANSITION = "transition",
    TERRAIN = "terrain",
}
