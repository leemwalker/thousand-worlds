import type { IShaderProvider } from "./interfaces";
import { Effect } from "@babylonjs/core/Materials/effect";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import type { Scene } from "@babylonjs/core/scene";
import type { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";

// Define shaders in-line for Phase 1 simplicity
Effect.ShadersStore["displacementVertexShader"] = `
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

    // Varyings
    varying vec2 vUV;
    varying float vHeight;
    varying vec3 vNormal;
    varying vec3 vPosition;

    void main(void) {
        vUV = uv;
        vNormal = normal;
        
        // Sample height
        // Assuming heightmap is normalized 0.0-1.0
        float h = texture2D(heightmap, uv).r;
        vHeight = h;
        
        // Displace along normal
        // scale determines max height in world units
        vec3 p = position + normal * (h * scale);
        vPosition = p;
        
        gl_Position = viewProjection * world * vec4(p, 1.0);
    }
`;

Effect.ShadersStore["displacementFragmentShader"] = `
    precision highp float;

    // Varyings
    varying vec2 vUV;
    varying float vHeight;
    varying vec3 vNormal;

    // Uniforms
    uniform vec3 color; // Base color
    uniform vec3 lightDirection; // Dynamic sun direction (normalized)
    uniform float seaLevel; // Normalized sea level (0-1 in heightmap space)
    uniform float minElevation; // Minimum elevation in meters
    uniform float maxElevation; // Maximum elevation in meters

    // Bathymetric gradient (underwater)
    vec3 getBathymetricColor(float depthFactor) {
        // depthFactor: 0.0 = at sea level, 1.0 = deepest ocean
        vec3 shallow = vec3(0.0, 0.5, 0.7);   // Turquoise
        vec3 mid = vec3(0.0, 0.25, 0.45);     // Ocean blue  
        vec3 deep = vec3(0.02, 0.08, 0.15);   // Deep navy
        
        if (depthFactor < 0.3) {
            return mix(shallow, mid, depthFactor / 0.3);
        }
        return mix(mid, deep, (depthFactor - 0.3) / 0.7);
    }

    // Hypsometric gradient (land)
    vec3 getHypsometricColor(float heightFactor) {
        // heightFactor: 0.0 = sea level, 1.0 = max elevation
        vec3 coast = vec3(0.35, 0.30, 0.25);     // Low land - basalt
        vec3 lowland = vec3(0.40, 0.35, 0.28);   // Plains
        vec3 highland = vec3(0.48, 0.43, 0.38);  // Highland
        vec3 mountain = vec3(0.55, 0.52, 0.48);  // Mountain rock
        vec3 peak = vec3(0.75, 0.73, 0.70);      // High peaks
        vec3 snow = vec3(0.95, 0.95, 0.95);      // Snow cap
        
        if (heightFactor < 0.1) return mix(coast, lowland, heightFactor / 0.1);
        if (heightFactor < 0.3) return mix(lowland, highland, (heightFactor - 0.1) / 0.2);
        if (heightFactor < 0.5) return mix(highland, mountain, (heightFactor - 0.3) / 0.2);
        if (heightFactor < 0.75) return mix(mountain, peak, (heightFactor - 0.5) / 0.25);
        return mix(peak, snow, (heightFactor - 0.75) / 0.25);
    }

    void main(void) {
        // Lighting
        float ndotl = max(0.0, dot(vNormal, lightDirection));
        
        vec3 surfaceColor;
        
        if (vHeight < seaLevel) {
            // Underwater - calculate depth factor (0 = sea level, 1 = deepest)
            float depthFactor = (seaLevel - vHeight) / max(seaLevel, 0.001);
            surfaceColor = getBathymetricColor(depthFactor);
        } else {
            // Land - calculate height factor (0 = sea level, 1 = max elevation)
            float heightFactor = (vHeight - seaLevel) / max(1.0 - seaLevel, 0.001);
            surfaceColor = getHypsometricColor(heightFactor);
        }
        
        vec3 finalColor = surfaceColor * (0.25 + 0.75 * ndotl); // Ambient + Diffuse

        gl_FragColor = vec4(finalColor, 1.0);
    }
`;

export class DisplacementShader implements IShaderProvider {
    private scene: Scene;
    private material: ShaderMaterial | null = null;
    private heightmap: Texture | null = null;

    constructor(scene: Scene) {
        this.scene = scene;
    }

    public createMaterial(heightmap: Texture, scale: number = 0.5): ShaderMaterial {
        this.heightmap = heightmap;

        this.material = new ShaderMaterial(
            "displacementShader",
            this.scene,
            {
                vertex: "displacement",
                fragment: "displacement",
            },
            {
                attributes: ["position", "normal", "uv"],
                uniforms: ["world", "viewProjection", "scale", "color", "lightDirection", "seaLevel", "minElevation", "maxElevation"],
                samplers: ["heightmap"],
            }
        );

        this.material.setTexture("heightmap", heightmap);
        this.material.setFloat("scale", scale);
        // Default light direction (normalized) - will be updated each frame
        this.material.setVector3("lightDirection", new Vector3(1, 0, 0));

        // Default elevation params (assumes min-max normalization with sea level at ~0.3)
        // These should be updated via setElevationRange() with actual world data
        this.material.setFloat("seaLevel", 0.3);
        this.material.setFloat("minElevation", -6000);
        this.material.setFloat("maxElevation", 8848);

        // Render back faces too just in case, though sphere usually doesn't need it
        this.material.backFaceCulling = true;

        return this.material;
    }

    public updateHeightmap(texture: Texture): void {
        this.heightmap = texture;
        if (this.material) {
            this.material.setTexture("heightmap", texture);
        }
    }

    public setDisplacementScale(scale: number): void {
        if (this.material) {
            this.material.setFloat("scale", scale);
        }
    }

    /**
     * Update the light direction for sun-based shading.
     * @param direction Normalized direction FROM surface TO light source
     */
    public setLightDirection(direction: Vector3): void {
        if (this.material) {
            this.material.setVector3("lightDirection", direction);
        }
    }

    /**
     * Set the elevation range for proper sea level calculation.
     * The heightmap texture uses min-max normalization (0 = minElev, 1 = maxElev).
     * This calculates where sea level falls in that normalized range.
     * 
     * @param minElevation Minimum elevation in meters (e.g., -6000)
     * @param maxElevation Maximum elevation in meters (e.g., 15000)
     * @param seaLevel Actual sea level in meters (typically 0)
     */
    public setElevationRange(minElevation: number, maxElevation: number, seaLevel: number = 0): void {
        if (this.material) {
            const elevRange = maxElevation - minElevation;
            // Calculate normalized sea level (0-1) based on where sea level falls in elevation range
            const normalizedSeaLevel = elevRange > 0 ? (seaLevel - minElevation) / elevRange : 0.5;

            this.material.setFloat("seaLevel", normalizedSeaLevel);
            this.material.setFloat("minElevation", minElevation);
            this.material.setFloat("maxElevation", maxElevation);

            console.log(`[DisplacementShader] Elevation range: ${minElevation}m to ${maxElevation}m, sea level: ${seaLevel}m (normalized: ${normalizedSeaLevel.toFixed(3)})`);
        }
    }

    public dispose(): void {
        if (this.material) {
            this.material.dispose();
            this.material = null;
        }
    }
}
