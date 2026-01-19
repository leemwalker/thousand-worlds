import type { WebGPUEngine } from "@babylonjs/core/Engines/webgpuEngine";
import { ComputeShader } from "@babylonjs/core/Compute/computeShader";
import { StorageBuffer } from "@babylonjs/core/Buffers/storageBuffer";
import { UniformBuffer } from "@babylonjs/core/Materials/uniformBuffer";

const TERRAIN_COMPUTE_SOURCE = `
struct Params {
    gridSize: u32,
    face: u32,
    displacementScale: f32,
    planetRadius: f32,
};

@group(0) @binding(0) var<uniform> params : Params;
@group(0) @binding(1) var<storage, read> heightmap : array<f32>;
@group(0) @binding(2) var<storage, read_write> positions : array<f32>;
@group(0) @binding(3) var<storage, read_write> normals : array<f32>;

const PI: f32 = 3.14159265359;

fn cubeToSphere(face: u32, u: f32, v: f32) -> vec3<f32> {
    let cu = 2.0 * u - 1.0;
    let cv = 2.0 * v - 1.0;
    var p = vec3<f32>(0.0);

    // 0:Front, 1:Back, 2:Left, 3:Right, 4:Top, 5:Bottom
    switch (face) {
        case 0u: { p = vec3<f32>(cu, -cv, 1.0); }
        case 1u: { p = vec3<f32>(-cu, -cv, -1.0); }
        case 2u: { p = vec3<f32>(-1.0, -cv, -cu); }
        case 3u: { p = vec3<f32>(1.0, -cv, cu); }
        case 4u: { p = vec3<f32>(cu, 1.0, cv); }
        default: { p = vec3<f32>(cu, -1.0, -cv); } // 5u
    }

    return normalize(p);
}

@compute @workgroup_size(8, 8, 1)
fn main(@builtin(global_invocation_id) global_id : vec3<u32>) {
    let x = global_id.x;
    let y = global_id.y;
    let size = params.gridSize;

    if (x >= size || y >= size) {
        return;
    }

    let idx = y * size + x;
    
    // UV coordinates (0.0 to 1.0)
    let u = f32(x) / f32(size - 1u);
    let v = f32(y) / f32(size - 1u);

    // Get height
    let height = heightmap[idx];
    
    // Calculate base sphere position
    let sphereNormal = cubeToSphere(params.face, u, v);
    
    // Apply displacement
    let radius = params.planetRadius + (height * params.displacementScale);
    let pos = sphereNormal * radius;

    // Write position (stride 3)
    let pIdx = idx * 3u;
    positions[pIdx] = pos.x;
    positions[pIdx + 1u] = pos.y;
    positions[pIdx + 2u] = pos.z;

    // Calculate Normal (finite difference)
    // We need neighbors. Check boundary conditions.
    // For simplicity in this step, use sphere normal or compute proper gradients.
    // Proper gradient requires accessing neighbors (x+1, y+1).
    
    // Simple gradient approach:
    // This requires handling edge cases carefully. 
    // If x+1 >= size, clamp or wrap? For individual tiles, we typically need ghost cells or just clamp.
    
    let hR = heightmap[y * size + min(x + 1u, size - 1u)];
    let hL = heightmap[y * size + max(x, 1u) - 1u]; // if x=0, max(0,1)-1 = 0
    let hD = heightmap[min(y + 1u, size - 1u) * size + x];
    let hU = heightmap[(max(y, 1u) - 1u) * size + x];

    // Compute tangent vectors on sphere surface is complex.
    // Simplifying: Perturb normal by gradient.
    let dx = (hR - hL) * params.displacementScale;
    let dy = (hD - hU) * params.displacementScale;
    
    // This is a rough approximation. 
    // Ideally we compute 3 positions and cross product.
    
    normals[pIdx] = sphereNormal.x;
    normals[pIdx + 1u] = sphereNormal.y;
    normals[pIdx + 2u] = sphereNormal.z;
    // todo: proper normal computation
}
`;

export class TerrainComputeShader {
    private computeShader: ComputeShader;
    private paramsBuffer: UniformBuffer;

    constructor(engine: WebGPUEngine, gridSize: number) {
        this.computeShader = new ComputeShader(
            "terrainCompute",
            engine,
            { computeSource: TERRAIN_COMPUTE_SOURCE },
            {
                bindingsMapping: {
                    "params": { group: 0, binding: 0 },
                    "heightmap": { group: 0, binding: 1 },
                    "positions": { group: 0, binding: 2 },
                    "normals": { group: 0, binding: 3 },
                }
            }
        );

        this.paramsBuffer = new UniformBuffer(engine);
        this.paramsBuffer.addUniform("gridSize", 1);
        this.paramsBuffer.addUniform("face", 1);
        this.paramsBuffer.addUniform("displacementScale", 1);
        this.paramsBuffer.addUniform("planetRadius", 1);
        this.paramsBuffer.update();
    }

    public dispatch(
        gridSize: number,
        face: number,
        displacementScale: number,
        planetRadius: number,
        heightmap: StorageBuffer,
        positions: StorageBuffer,
        normals: StorageBuffer
    ) {
        this.paramsBuffer.updateUInt("gridSize", gridSize);
        this.paramsBuffer.updateUInt("face", face);
        this.paramsBuffer.updateFloat("displacementScale", displacementScale);
        this.paramsBuffer.updateFloat("planetRadius", planetRadius);
        this.paramsBuffer.update();

        this.computeShader.setUniformBuffer("params", this.paramsBuffer);
        this.computeShader.setStorageBuffer("heightmap", heightmap);
        this.computeShader.setStorageBuffer("positions", positions);
        this.computeShader.setStorageBuffer("normals", normals);

        const groupSize = 8;
        const groups = Math.ceil(gridSize / groupSize);
        this.computeShader.dispatch(groups, groups, 1);
    }
}
