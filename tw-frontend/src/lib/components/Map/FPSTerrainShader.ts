/**
 * FPSTerrainShader - Data-driven terrain material shader for FPS mode.
 * Phase 4: Visual Polish
 * 
 * Uses simulation data for material selection:
 * - rockHardness → granite/basalt/sandstone texture
 * - sediment → soil/dust/sand appearance
 * - flux → moisture/wetness
 * - temperature → snow/ice coverage
 * - elevation → altitude-based coloring
 */

import { Effect } from "@babylonjs/core/Materials/effect";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import type { Scene } from "@babylonjs/core/scene";
import type { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";

// Register the FPS terrain shaders
Effect.ShadersStore["fpsTerrainVertexShader"] = `
    precision highp float;

    // Attributes
    attribute vec3 position;
    attribute vec3 normal;
    attribute vec2 uv;

    // Uniforms
    uniform mat4 world;
    uniform mat4 viewProjection;
    uniform sampler2D heightmap;
    uniform float scale;
    uniform float seaLevel;

    // Varyings
    varying vec2 vUV;
    varying float vHeight;
    varying float vElevation;
    varying vec3 vNormal;
    varying vec3 vWorldPos;
    varying float vDistanceFromCamera;

    void main(void) {
        vUV = uv;
        vNormal = normal;
        
        // Sample height (0-1 normalized)
        float h = texture2D(heightmap, uv).r;
        vHeight = h;
        
        // Calculate actual elevation for material selection
        vElevation = h * scale;
        
        // Displace along normal
        vec3 p = position + normal * (h * scale);
        vWorldPos = (world * vec4(p, 1.0)).xyz;
        
        vec4 worldPos = world * vec4(p, 1.0);
        gl_Position = viewProjection * worldPos;
        
        // Calculate distance for fog
        vDistanceFromCamera = length(worldPos.xyz);
    }
`;

Effect.ShadersStore["fpsTerrainFragmentShader"] = `
    precision highp float;

    // Varyings
    varying vec2 vUV;
    varying float vHeight;
    varying float vElevation;
    varying vec3 vNormal;
    varying vec3 vWorldPos;
    varying float vDistanceFromCamera;

    // Data textures
    uniform sampler2D heightmap;
    uniform sampler2D rockHardnessTex;   // 0 = soft, 1 = hard
    uniform sampler2D sedimentTex;       // Sediment depth
    uniform sampler2D fluxTex;           // Water flow
    uniform sampler2D temperatureTex;    // Temperature

    // Lighting
    uniform vec3 lightDirection;
    uniform float ambientStrength;

    // Environment
    uniform float seaLevel;
    uniform float snowLine;
    uniform float treeLineElevation;

    // Fog
    uniform vec3 fogColor;
    uniform float fogDensity;
    uniform float fogStart;
    uniform float fogEnd;

    // Material colors (data-driven palette)
    const vec3 DEEP_WATER = vec3(0.02, 0.05, 0.15);
    const vec3 SHALLOW_WATER = vec3(0.05, 0.15, 0.35);
    const vec3 SAND = vec3(0.76, 0.70, 0.50);
    const vec3 SOIL = vec3(0.40, 0.30, 0.20);
    const vec3 GRASS = vec3(0.20, 0.45, 0.15);
    const vec3 WET_GRASS = vec3(0.15, 0.35, 0.10);
    const vec3 ROCK_GRANITE = vec3(0.55, 0.50, 0.48);
    const vec3 ROCK_BASALT = vec3(0.25, 0.25, 0.28);
    const vec3 ROCK_SANDSTONE = vec3(0.72, 0.55, 0.40);
    const vec3 SNOW = vec3(0.95, 0.97, 1.0);
    const vec3 ICE = vec3(0.75, 0.85, 0.95);
    const vec3 MUD = vec3(0.30, 0.25, 0.18);

    void main(void) {
        // Sample data textures
        float rockHardness = texture2D(rockHardnessTex, vUV).r;
        float sediment = texture2D(sedimentTex, vUV).r;
        float flux = texture2D(fluxTex, vUV).r;
        float temperature = texture2D(temperatureTex, vUV).r;

        // Normalize height to actual elevation
        float elevation = vHeight;
        
        // Start with base material based on conditions
        vec3 materialColor = SOIL;
        
        // Water (below sea level)
        float waterLevel = seaLevel;
        if (elevation < waterLevel) {
            float depth = (waterLevel - elevation) / 0.3;
            materialColor = mix(SHALLOW_WATER, DEEP_WATER, clamp(depth, 0.0, 1.0));
        }
        // Land materials
        else {
            float heightAboveWater = elevation - waterLevel;
            
            // Base type from rock hardness
            if (rockHardness > 0.85) {
                // Hard volcanic basalt
                materialColor = ROCK_BASALT;
            } else if (rockHardness > 0.6) {
                // Granite
                materialColor = ROCK_GRANITE;
            } else if (rockHardness < 0.3) {
                // Soft sandstone/sedimentary
                materialColor = ROCK_SANDSTONE;
            } else {
                // Medium - mixed rock
                materialColor = mix(ROCK_SANDSTONE, ROCK_GRANITE, rockHardness);
            }
            
            // Sediment overlay (soil/sand)
            if (sediment > 0.1) {
                float sedimentBlend = clamp(sediment * 2.0, 0.0, 1.0);
                
                // Coastal -> sand, inland -> soil
                if (heightAboveWater < 0.02) {
                    materialColor = mix(materialColor, SAND, sedimentBlend);
                } else {
                    materialColor = mix(materialColor, SOIL, sedimentBlend);
                }
            }
            
            // Vegetation (based on moisture from flux)
            if (flux > 0.1 && sediment > 0.05 && heightAboveWater < 0.3 && temperature > 0.2) {
                float vegetationBlend = clamp(flux * sediment * 4.0, 0.0, 0.7);
                
                if (flux > 0.5) {
                    // Wet/lush vegetation
                    materialColor = mix(materialColor, WET_GRASS, vegetationBlend);
                } else {
                    // Regular grass
                    materialColor = mix(materialColor, GRASS, vegetationBlend);
                }
            }
            
            // Wet rock (high flux, low sediment)
            if (flux > 0.5 && sediment < 0.1) {
                // Darken and add slight reflectivity
                materialColor *= 0.7;
            }
            
            // Mud (high flux + high sediment)
            if (flux > 0.3 && sediment > 0.3) {
                float mudBlend = clamp((flux + sediment - 0.6) * 2.0, 0.0, 0.5);
                materialColor = mix(materialColor, MUD, mudBlend);
            }
            
            // Snow/ice (temperature based)
            if (temperature < 0.15) {
                float snowBlend = clamp((0.15 - temperature) * 10.0, 0.0, 1.0);
                
                if (flux > 0.3) {
                    // Ice where wet
                    materialColor = mix(materialColor, ICE, snowBlend);
                } else {
                    // Snow
                    materialColor = mix(materialColor, SNOW, snowBlend);
                }
            }
            
            // High altitude snow line
            if (heightAboveWater > 0.5) {
                float snowBlend = clamp((heightAboveWater - 0.5) * 3.0, 0.0, 1.0);
                materialColor = mix(materialColor, SNOW, snowBlend);
            }
        }

        // Lighting
        float ndotl = max(0.0, dot(normalize(vNormal), normalize(lightDirection)));
        vec3 litColor = materialColor * (ambientStrength + (1.0 - ambientStrength) * ndotl);

        // Fog
        float fogFactor = clamp((vDistanceFromCamera - fogStart) / (fogEnd - fogStart), 0.0, 1.0);
        fogFactor = fogFactor * fogFactor * fogDensity; // Quadratic falloff
        vec3 finalColor = mix(litColor, fogColor, fogFactor);

        gl_FragColor = vec4(finalColor, 1.0);
    }
`;

export interface FPSTerrainShaderOptions {
    seaLevel?: number;
    snowLine?: number;
    treeLineElevation?: number;
    ambientStrength?: number;
    fogColor?: Vector3;
    fogDensity?: number;
    fogStart?: number;
    fogEnd?: number;
}

/**
 * Data-driven terrain shader for FPS mode.
 */
export class FPSTerrainShader {
    private scene: Scene;
    private material: ShaderMaterial | null = null;
    private options: Required<FPSTerrainShaderOptions>;

    constructor(scene: Scene, options: FPSTerrainShaderOptions = {}) {
        this.scene = scene;
        this.options = {
            seaLevel: options.seaLevel ?? 0.1,
            snowLine: options.snowLine ?? 0.6,
            treeLineElevation: options.treeLineElevation ?? 0.4,
            ambientStrength: options.ambientStrength ?? 0.3,
            fogColor: options.fogColor ?? new Vector3(0.7, 0.8, 0.9),
            fogDensity: options.fogDensity ?? 1.0,
            fogStart: options.fogStart ?? 0.1,
            fogEnd: options.fogEnd ?? 2.0
        };
    }

    /**
     * Create the shader material with data textures.
     */
    createMaterial(
        heightmap: Texture,
        rockHardnessTex: Texture | null,
        sedimentTex: Texture | null,
        fluxTex: Texture | null,
        temperatureTex: Texture | null,
        scale: number = 0.05
    ): ShaderMaterial {
        this.material = new ShaderMaterial(
            "fpsTerrainShader",
            this.scene,
            {
                vertex: "fpsTerrain",
                fragment: "fpsTerrain"
            },
            {
                attributes: ["position", "normal", "uv"],
                uniforms: [
                    "world", "viewProjection", "scale", "seaLevel",
                    "lightDirection", "ambientStrength",
                    "snowLine", "treeLineElevation",
                    "fogColor", "fogDensity", "fogStart", "fogEnd"
                ],
                samplers: ["heightmap", "rockHardnessTex", "sedimentTex", "fluxTex", "temperatureTex"]
            }
        );

        // Set textures
        this.material.setTexture("heightmap", heightmap);

        // Use heightmap as fallback for missing data textures
        this.material.setTexture("rockHardnessTex", rockHardnessTex ?? heightmap);
        this.material.setTexture("sedimentTex", sedimentTex ?? heightmap);
        this.material.setTexture("fluxTex", fluxTex ?? heightmap);
        this.material.setTexture("temperatureTex", temperatureTex ?? heightmap);

        // Set uniforms
        this.material.setFloat("scale", scale);
        this.material.setFloat("seaLevel", this.options.seaLevel);
        this.material.setFloat("snowLine", this.options.snowLine);
        this.material.setFloat("treeLineElevation", this.options.treeLineElevation);
        this.material.setFloat("ambientStrength", this.options.ambientStrength);
        this.material.setVector3("lightDirection", new Vector3(1, 0.5, 0));
        this.material.setVector3("fogColor", this.options.fogColor);
        this.material.setFloat("fogDensity", this.options.fogDensity);
        this.material.setFloat("fogStart", this.options.fogStart);
        this.material.setFloat("fogEnd", this.options.fogEnd);

        this.material.backFaceCulling = true;

        return this.material;
    }

    /**
     * Update light direction.
     */
    setLightDirection(direction: Vector3): void {
        this.material?.setVector3("lightDirection", direction);
    }

    /**
     * Update fog parameters.
     */
    setFog(color: Vector3, density: number, start: number, end: number): void {
        if (this.material) {
            this.material.setVector3("fogColor", color);
            this.material.setFloat("fogDensity", density);
            this.material.setFloat("fogStart", start);
            this.material.setFloat("fogEnd", end);
        }
    }

    /**
     * Update sea level.
     */
    setSeaLevel(level: number): void {
        this.material?.setFloat("seaLevel", level);
    }

    /**
     * Get the material.
     */
    getMaterial(): ShaderMaterial | null {
        return this.material;
    }

    /**
     * Dispose of resources.
     */
    dispose(): void {
        this.material?.dispose();
        this.material = null;
    }
}
