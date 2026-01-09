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
    uniform vec3 color; // Base color (fallback)
    uniform vec3 lightDirection; // Dynamic sun direction (normalized)
    uniform float seaLevel; // Normalized sea level (0-1 in heightmap space)
    uniform float minElevation; // Minimum elevation in meters
    uniform float maxElevation; // Maximum elevation in meters
    
    // Data textures for data-driven coloring
    uniform sampler2D materialTex; // R=hardness, G=continental, B=sediment
    uniform sampler2D iceTex;      // R=ice thickness
    uniform sampler2D normalTex;   // Normal map for 3D shadows
    uniform bool hasMaterialTex;
    uniform bool hasIceTex;
    uniform bool hasNormalTex;

    // Color palette for rock types based on hardness
    vec3 getRockColor(float hardness, bool isContinental) {
        if (!isContinental) {
            // Oceanic crust: basalt (dark grey)
            return vec3(0.235, 0.235, 0.255);
        }
        
        // Continental crust: sedimentary -> granite based on hardness
        vec3 sandstone = vec3(0.706, 0.549, 0.392);  // Soft sedimentary
        vec3 granite = vec3(0.627, 0.588, 0.569);    // Medium metamorphic
        vec3 hardRock = vec3(0.4, 0.4, 0.42);        // Hard crystalline
        
        if (hardness < 0.5) {
            return mix(sandstone, granite, hardness * 2.0);
        }
        return mix(granite, hardRock, (hardness - 0.5) * 2.0);
    }

    // Sediment overlay color
    vec3 getSedimentColor(float sediment) {
        vec3 sedimentTan = vec3(0.706, 0.627, 0.471);
        return sedimentTan;
    }

    // Ice/snow color based on thickness
    vec3 getIceColor(float thickness) {
        vec3 frost = vec3(0.9, 0.91, 0.93);       // Light frost
        vec3 snow = vec3(0.95, 0.95, 0.97);       // Snow white
        vec3 glacier = vec3(0.75, 0.85, 0.92);    // Blue-white glacier
        
        if (thickness < 0.3) {
            return mix(frost, snow, thickness / 0.3);
        }
        return mix(snow, glacier, (thickness - 0.3) / 0.7);
    }

    // Satellite-style bathymetric (underwater terrain visible through water)
    vec3 getBathymetricColor(vec3 terrainColor, float depthFactor) {
        vec3 shallowWater = vec3(0.314, 0.706, 0.863);  // Turquoise
        vec3 deepWater = vec3(0.02, 0.08, 0.2);         // Deep navy
        
        // Water tint based on depth
        vec3 waterTint = mix(shallowWater, deepWater, depthFactor);
        
        // Blend terrain with water - shallow shows terrain, deep obscures
        float visibility = 1.0 - min(depthFactor * 1.5, 0.85);
        return mix(waterTint, terrainColor * 0.7, visibility);
    }

    void main(void) {
        // Calculate perturbed normal
        vec3 normal = normalize(vNormal);
        
        if (hasNormalTex) {
            // Sample normal map (tangent space)
            vec3 mapN = texture2D(normalTex, vUV).rgb * 2.0 - 1.0;
            
            // Calculate TBN matrix
            // Tangent (East-West)
            vec3 T = normalize(cross(vec3(0.0, 1.0, 0.0), normal));
            if (length(T) < 0.001) T = vec3(1.0, 0.0, 0.0); // Pole fallback
            
            // Bitangent (North-South)
            vec3 B = normalize(cross(normal, T));
            
            // Transform to World Space
            normal = normalize(T * mapN.x + B * mapN.y + normal * mapN.z);
        }

        // Lighting
        float ndotl = max(0.0, dot(normal, lightDirection));
        
        // Sample material data if available
        float hardness = 0.5;
        bool isContinental = true;
        float sediment = 0.0;
        
        if (hasMaterialTex) {
            vec4 matData = texture2D(materialTex, vUV);
            hardness = matData.r;
            isContinental = matData.g > 0.5;
            sediment = matData.b;
        }
        
        // Get base rock color from material data
        vec3 surfaceColor = getRockColor(hardness, isContinental);
        
        // Apply sediment overlay
        if (sediment > 0.1) {
            vec3 sedimentCol = getSedimentColor(sediment);
            surfaceColor = mix(surfaceColor, sedimentCol, min(sediment * 1.5, 0.7));
        }
        
        // Check if underwater
        if (vHeight < seaLevel) {
            // Underwater - satellite-style bathymetry showing underwater terrain
            float depthFactor = (seaLevel - vHeight) / max(seaLevel, 0.001);
            surfaceColor = getBathymetricColor(surfaceColor, depthFactor);
        }
        
        // Apply ice overlay
        if (hasIceTex) {
            float ice = texture2D(iceTex, vUV).r;
            if (ice > 0.05) {
                vec3 iceCol = getIceColor(ice);
                surfaceColor = mix(surfaceColor, iceCol, min(ice * 2.0, 0.95));
            }
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
                uniforms: ["world", "viewProjection", "scale", "color", "lightDirection", "seaLevel", "minElevation", "maxElevation", "hasMaterialTex", "hasIceTex", "hasNormalTex"],
                samplers: ["heightmap", "materialTex", "iceTex", "normalTex"],
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

        // Data texture flags - default to false until textures are provided
        this.material.setInt("hasMaterialTex", 0);
        this.material.setInt("hasIceTex", 0);
        this.material.setInt("hasNormalTex", 0);

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
     * Set the material data texture for data-driven terrain coloring.
     * R=rock hardness (0-1), G=continental (0=oceanic, 1=continental), B=sediment depth
     * @param texture Material data texture (RGB PNG)
     */
    public setMaterialTexture(texture: Texture): void {
        if (this.material) {
            this.material.setTexture("materialTex", texture);
            this.material.setInt("hasMaterialTex", 1);
            console.log("[DisplacementShader] Material texture set");
        }
    }

    /**
     * Set the ice sheet data texture for glacier/polar ice visualization.
     * R=ice thickness (normalized 0-1)
     * @param texture Ice data texture (grayscale PNG)
     */
    public setIceTexture(texture: Texture): void {
        if (this.material) {
            this.material.setTexture("iceTex", texture);
            this.material.setInt("hasIceTex", 1);
            console.log("[DisplacementShader] Ice texture set");
        }
    }

    /**
     * Set the normal map texture for 3D shadows.
     * @param texture Normal map texture (RGB PNG, Tangent Space)
     */
    public setNormalMap(texture: Texture): void {
        if (this.material) {
            this.material.setTexture("normalTex", texture);
            this.material.setInt("hasNormalTex", 1);
            console.log("[DisplacementShader] Normal map set");
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
