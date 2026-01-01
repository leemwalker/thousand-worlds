/**
 * TerrainChunk - High-resolution terrain mesh for first-person mode.
 * Phase 3.3: Terrain Placeholders for FPS Mode
 * 
 * Manages a single chunk of terrain with mesh generation and optional physics collider.
 */

import type { Scene } from "@babylonjs/core/scene";
import type { Mesh } from "@babylonjs/core/Meshes/mesh";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
import { Color3 } from "@babylonjs/core/Maths/math.color";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import type { ITerrainChunk } from './interfaces';

export interface TerrainChunkOptions {
    size?: number;
    segments?: number;
    enablePhysics?: boolean;
}

interface Position3D {
    x: number;
    y: number;
    z: number;
}

/**
 * Represents a single terrain chunk for ground-level rendering.
 */
export class TerrainChunk implements ITerrainChunk {
    private scene: Scene;
    private mesh: Mesh | null = null;
    private material: StandardMaterial | null = null;
    private collider: any | null = null; // PhysicsBody when physics enabled
    private disposed: boolean = false;
    private _position: Position3D;
    private _lodLevel: number;
    private _size: number;
    private segments: number;

    // ITerrainChunk interface properties
    position: { lat: number; lon: number };
    lodLevel: number;

    constructor(
        scene: Scene,
        position: Position3D,
        lodLevel: number,
        options: TerrainChunkOptions = {}
    ) {
        this.scene = scene;
        this._position = position;
        this._lodLevel = lodLevel;
        this._size = options.size ?? 256;
        this.segments = options.segments ?? 64;

        // Interface-compatible position (lat/lon derived from world position)
        this.position = { lat: position.z, lon: position.x };
        this.lodLevel = lodLevel;

        // Create placeholder mesh
        this.createPlaceholderMesh();
    }

    /**
     * Create a placeholder ground mesh until height data is loaded.
     */
    private createPlaceholderMesh(): void {
        this.mesh = MeshBuilder.CreateGround(
            `terrain_${this._position.x}_${this._position.z}_lod${this._lodLevel}`,
            {
                width: this._size,
                height: this._size,
                subdivisions: this.segments,
                updatable: true
            },
            this.scene
        );

        // Position in world space
        this.mesh.position = new Vector3(
            this._position.x,
            this._position.y,
            this._position.z
        );

        // Create material
        this.material = new StandardMaterial(
            `terrainMat_${this._position.x}_${this._position.z}`,
            this.scene
        );
        this.material.diffuseColor = new Color3(0.4, 0.5, 0.3); // Terrain green
        this.material.specularColor = new Color3(0.1, 0.1, 0.1);
        this.mesh.material = this.material;
    }

    /**
     * Get the terrain mesh.
     */
    getMesh(): Mesh | null {
        return this.mesh;
    }

    /**
     * Get physics collider (null if physics not enabled).
     */
    getCollider(): any | null {
        return this.collider;
    }

    /**
     * Get chunk world position.
     */
    getPosition(): Position3D {
        return this._position;
    }

    /**
     * Get LOD level.
     */
    getLODLevel(): number {
        return this._lodLevel;
    }

    /**
     * Get chunk size.
     */
    getSize(): number {
        return this._size;
    }

    /**
     * Check if chunk is disposed.
     */
    isDisposed(): boolean {
        return this.disposed;
    }

    /**
     * Generate terrain mesh from height data.
     */
    generateFromHeightData(
        heightData: Float32Array,
        width: number,
        height: number,
        heightScale: number = 1.0
    ): void {
        if (this.disposed || !this.mesh) return;

        const positions = this.mesh.getVerticesData("position");
        if (!positions) return;

        const newPositions = [...positions];

        // Update Y positions based on height data
        for (let i = 0; i < newPositions.length; i += 3) {
            // Map vertex position to height data index
            const localX = (newPositions[i] / this._size + 0.5);
            const localZ = (newPositions[i + 2] / this._size + 0.5);

            const hx = Math.floor(localX * (width - 1));
            const hz = Math.floor(localZ * (height - 1));
            const idx = Math.min(hz * width + hx, heightData.length - 1);

            newPositions[i + 1] = heightData[idx] * heightScale;
        }

        this.mesh.updateVerticesData("position", new Float32Array(newPositions));
        this.mesh.refreshBoundingInfo();
    }

    /**
     * Update material with texture.
     */
    setTexture(textureUrl: string): void {
        if (this.disposed || !this.material) return;
        // TODO: Load and apply texture
    }

    /**
     * Dispose of all resources.
     */
    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;

        if (this.collider) {
            // Dispose physics body if exists
            this.collider = null;
        }

        if (this.material) {
            this.material.dispose();
            this.material = null;
        }

        if (this.mesh) {
            this.mesh.dispose();
            this.mesh = null;
        }
    }
}
