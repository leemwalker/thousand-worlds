/**
 * TileMesh - Individual cube-face tile mesh with displacement shader.
 * Phase 3.2: Cube Tile System
 */

import { Scene } from "@babylonjs/core/scene";
import { Mesh } from "@babylonjs/core/Meshes/mesh";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { CubeFace, TileCoord } from './interfaces';

export interface TileMeshOptions {
    displacementScale?: number;
    segments?: number;
}

/**
 * Represents a single tile mesh on the globe surface.
 */
export class TileMesh {
    private mesh: Mesh;
    private material: StandardMaterial;
    private coord: TileCoord;
    private scene: Scene;
    private disposed: boolean = false;

    constructor(
        scene: Scene,
        coord: TileCoord,
        texture: Texture,
        heightmap: Texture | null,
        options: TileMeshOptions = {}
    ) {
        this.scene = scene;
        this.coord = coord;

        const segments = options.segments ?? 32;
        const displacementScale = options.displacementScale ?? 0.05;

        // Create a plane mesh that will be warped to sphere surface
        this.mesh = MeshBuilder.CreateGround(
            `tile_${coord.face}_${coord.level}_${coord.x}_${coord.y}`,
            {
                width: 1,
                height: 1,
                subdivisions: segments,
                updatable: true
            },
            scene
        );

        // Calculate tile position and orientation on the sphere
        this.positionOnSphere(coord, displacementScale);

        // Create material with texture
        this.material = new StandardMaterial(
            `tileMat_${coord.face}_${coord.level}_${coord.x}_${coord.y}`,
            scene
        );
        this.material.diffuseTexture = texture;
        this.material.specularPower = 64;

        // If heightmap provided, we could use it for per-vertex displacement
        // For now, we rely on the shader-based displacement in DisplacementShader
        if (heightmap) {
            // Store reference for potential CPU displacement fallback
            (this.mesh as any)._heightmapTexture = heightmap;
        }

        this.mesh.material = this.material;
    }

    /**
     * Position the tile mesh on the unit sphere.
     */
    private positionOnSphere(coord: TileCoord, displacementScale: number): void {
        const positions = this.mesh.getVerticesData("position");
        if (!positions) return;

        const tilesPerSide = 1 << coord.level;
        const tileSize = 1.0 / tilesPerSide;

        // UV bounds for this tile
        const uMin = coord.x * tileSize;
        const vMin = coord.y * tileSize;

        const newPositions: number[] = [];

        for (let i = 0; i < positions.length; i += 3) {
            // Original ground position is in [-0.5, 0.5] range
            const localX = positions[i] + 0.5; // Now [0, 1]
            const localZ = positions[i + 2] + 0.5; // Now [0, 1]

            // Map to tile UV within cube face
            const u = uMin + localX * tileSize;
            const v = vMin + localZ * tileSize;

            // Convert to sphere position
            const spherePos = this.cubeToSpherePosition(coord.face, u, v);

            // Apply radius (1.0 is base, add small offset for tile layer)
            const radius = 1.0;
            newPositions.push(
                spherePos.x * radius,
                spherePos.y * radius,
                spherePos.z * radius
            );
        }

        this.mesh.updateVerticesData("position", new Float32Array(newPositions));
        this.mesh.refreshBoundingInfo();
    }

    /**
     * Convert cube face UV to unit sphere position.
     */
    private cubeToSpherePosition(face: CubeFace, u: number, v: number): Vector3 {
        // Convert UV to [-1, 1] cube coordinates
        const cu = 2 * u - 1;
        const cv = 2 * v - 1;

        let x: number, y: number, z: number;

        switch (face) {
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
        return new Vector3(x / length, y / length, z / length);
    }

    /**
     * Get the underlying mesh.
     */
    getMesh(): Mesh {
        return this.mesh;
    }

    /**
     * Get tile coordinates.
     */
    getCoord(): TileCoord {
        return this.coord;
    }

    /**
     * Check if this tile is disposed.
     */
    isDisposed(): boolean {
        return this.disposed;
    }

    /**
     * Update the tile texture.
     */
    updateTexture(texture: Texture): void {
        if (this.disposed) return;
        this.material.diffuseTexture = texture;
    }

    /**
     * Set visibility of the tile.
     */
    setVisible(visible: boolean): void {
        if (this.disposed) return;
        this.mesh.isVisible = visible;
    }

    /**
     * Dispose of the tile mesh and material.
     */
    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;

        if (this.material) {
            if (this.material.diffuseTexture) {
                this.material.diffuseTexture.dispose();
            }
            this.material.dispose();
        }

        if (this.mesh) {
            this.mesh.dispose();
        }
    }
}
