export const TERRAIN_VERTEX_SHADER = `
struct Uniforms {
    viewProjection: mat4x4<f32>,
    world: mat4x4<f32>,
};

@group(0) @binding(0) var<uniform> uniforms : Uniforms;

struct VertexInput {
    @location(0) position : vec3<f32>,
    @location(1) normal : vec3<f32>,
    @location(2) uv : vec2<f32>, // Optional, if we want texture mapping
};

struct VertexOutput {
    @builtin(position) position : vec4<f32>,
    @location(0) vPosition : vec3<f32>,
    @location(1) vNormal : vec3<f32>,
    @location(2) vUV : vec2<f32>,
};

@vertex
fn main(input : VertexInput) -> VertexOutput {
    var output : VertexOutput;
    
    let worldPos = uniforms.world * vec4<f32>(input.position, 1.0);
    output.position = uniforms.viewProjection * worldPos;
    output.vPosition = worldPos.xyz;
    output.vNormal = normalize((uniforms.world * vec4<f32>(input.normal, 0.0)).xyz);
    output.vUV = input.uv;
    
    return output;
}
`;

export const TERRAIN_FRAGMENT_SHADER = `
struct Light {
    direction: vec3<f32>,
    color: vec3<f32>,
};

@group(0) @binding(1) var<uniform> light : Light;

// Biome Colors (Texture or Uniform array?)
// Simplest: array of colors
struct BiomeParams {
    colors: array<vec4<f32>, 16>, // Max 16 biomes
};
@group(0) @binding(2) var<uniform> biomes : BiomeParams;

// Input from Vertex Shader
struct FragmentInput {
    @location(0) vPosition : vec3<f32>,
    @location(1) vNormal : vec3<f32>,
    @location(2) vUV : vec2<f32>,
    // @location(3) biomeID : f32 - we need to pass this attribute!
};

@fragment
fn main(input : FragmentInput) -> @location(0) vec4<f32> {
    // Normal lighting (Lambert)
    let N = normalize(input.vNormal);
    let L = normalize(-light.direction);
    let NdotL = max(dot(N, L), 0.0);
    
    // Biome Color
    // TODO: We need biome ID per vertex or sampled from texture.
    // For now, let's use height-based coloring as placeholder or debug gray.
    
    let baseColor = vec3<f32>(0.5, 0.5, 0.5); // Gray
    
    let diffuse = baseColor * NdotL * light.color;
    let ambient = vec3<f32>(0.1, 0.1, 0.1);
    
    return vec4<f32>(diffuse + ambient, 1.0);
}
`;
