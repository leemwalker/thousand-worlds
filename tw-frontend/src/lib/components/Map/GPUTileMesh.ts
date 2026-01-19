import { Mesh } from "@babylonjs/core/Meshes/mesh";
import type { Scene } from "@babylonjs/core/scene";
import type { WebGPUEngine } from "@babylonjs/core/Engines/webgpuEngine";
import { StorageBuffer } from "@babylonjs/core/Buffers/storageBuffer";
import { VertexBuffer } from "@babylonjs/core/Buffers/buffer";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import type { TileData } from "$lib/game/TileFetcher";
import { TerrainComputeShader } from "./TerrainComputeShader";
import { TERRAIN_VERTEX_SHADER, TERRAIN_FRAGMENT_SHADER } from "./TerrainShaders";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { VertexData } from "@babylonjs/core/Meshes/mesh.vertexData";

export class GPUTileMesh {
    public mesh: Mesh;
    private computeShader: TerrainComputeShader;

    // Buffers
    private heightmapBuffer: StorageBuffer;
    private positionBuffer: StorageBuffer;
    private normalBuffer: StorageBuffer;
    private biomeBuffer: VertexBuffer | null = null;

    constructor(
        private scene: Scene,
        private data: TileData,
        private compute: TerrainComputeShader // Shared compute shader instance
    ) {
        this.computeShader = compute;
        this.mesh = new Mesh(`tile_${data.face}_${data.level}_${data.x}_${data.y}`, scene);

        // Initialize buffers
        const engine = this.scene.getEngine() as WebGPUEngine;
        const vertexCount = this.data.width * this.data.height;

        this.heightmapBuffer = new StorageBuffer(engine, this.data.heightmap.byteLength, 1);
        this.positionBuffer = new StorageBuffer(engine, vertexCount * 3 * 4, 3 | 4);
        this.normalBuffer = new StorageBuffer(engine, vertexCount * 3 * 4, 3 | 4);

        // Process logic
        this.initBuffers();
        this.generateGeometry();
        this.applyMaterial();
    }

    private initBuffers() {
        this.heightmapBuffer.update(this.data.heightmap);
    }

    private generateGeometry() {
        // Dispatch Compute Shader
        // Params: gridSize, face, dispScale, planetRadius
        // TODO: Get real radius and scale from somewhere
        const radius = 6371000.0; // Earth radius approx
        const scale = 1.0;

        this.computeShader.dispatch(
            this.data.width, // Assuming square grid for now
            this.data.face,
            scale,
            radius,
            this.heightmapBuffer,
            this.positionBuffer,
            this.normalBuffer
        );

        // Create Vertex Data and apply to Mesh
        // We use the StorageBuffers as VertexBuffers

        const vertexCount = this.data.width * this.data.height;

        // Positions
        const posBuffer = this.positionBuffer.getBuffer();
        this.mesh.setVerticesBuffer(
            new VertexBuffer(
                this.scene.getEngine(),
                posBuffer,
                VertexBuffer.PositionKind,
                false, // updatable
                false, // poster
                3, // stride
                false, // instanced
                0, // offset
                false, // normalized
                false // useBytes (false = stride/offset in floats if buffer is float) 
                // BUT buffer is GPUBuffer.
            )
        );

        // Normals
        const normBuffer = this.normalBuffer.getBuffer();
        this.mesh.setVerticesBuffer(
            new VertexBuffer(
                this.scene.getEngine(),
                normBuffer,
                VertexBuffer.NormalKind,
                false,
                false,
                3,
                false,
                0,
                3
            )
        );

        // Indices
        // We need to generate indices for the grid logic (CPU side is fine for static grid indices)
        const indices = this.generateIndices(this.data.width, this.data.height);
        this.mesh.setIndices(indices);
    }

    private generateIndices(width: number, height: number): number[] {
        const indices: number[] = [];
        for (let y = 0; y < height - 1; y++) {
            for (let x = 0; x < width - 1; x++) {
                const a = y * width + x;
                const b = y * width + (x + 1);
                const c = (y + 1) * width + x;
                const d = (y + 1) * width + (x + 1);

                // Triangle 1
                indices.push(a, c, b);
                // Triangle 2
                indices.push(b, c, d);
            }
        }
        return indices;
    }

    private applyMaterial() {
        const material = new ShaderMaterial(
            "terrainMat",
            this.scene,
            {
                vertexSource: TERRAIN_VERTEX_SHADER,
                fragmentSource: TERRAIN_FRAGMENT_SHADER,
            },
            {
                attributes: ["position", "normal", "uv"],
                uniforms: ["world", "viewProjection", "lightDir", "colors"],
            }
        );

        this.mesh.material = material;
    }

    public dispose() {
        this.mesh.dispose();
        this.heightmapBuffer.dispose();
        this.positionBuffer.dispose();
        this.normalBuffer.dispose();
    }
}

