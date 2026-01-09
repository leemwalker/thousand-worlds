import type { IShaderProvider } from "./interfaces";
import { Effect } from "@babylonjs/core/Materials/effect";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import type { Scene } from "@babylonjs/core/scene";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
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
    varying float vHeight;
    varying vec3 vNormal;
    varying vec3 vPosition;
    varying vec2 vUV; // We still pass UV for material textures that might use it (though we should strictly use calc'd UV)

    void main(void) {
        vNormal = normal;
        
        // We will calculate UVs in fragment shader for the diffuse, 
        // BUT we need UVs here for sampling the heightmap if we want displacement.
        // HOWEVER, standard UVs pinch at poles.
        // Ideally, heightmap would also be sampled via 3D position, but vertex texture fetch isn't always cheap/easy with gradients.
        // For now, we will stick to using the Mesh UVs for vertex displacement (geometry is dense there anyway so pinching is less visible in height than in texture),
        // or we could try to compute spherical UVs here too.
        // For Icosphere, 'uv' attribute might not be what we want.
        // Let's compute a basic spherical UV here for the vertex displacement fetch.
        
        vec3 p_norm = normalize(position);
        float long_u = atan(p_norm.z, p_norm.x) / (2.0 * 3.14159265359) + 0.5;
        float lat_v = asin(p_norm.y) / 3.14159265359 + 0.5;
        vec2 sphericalUV = vec2(long_u, lat_v);
        vUV = sphericalUV; // Use this for height fetch

        // Sample height
        // Assuming heightmap is normalized 0.0-1.0
        float h = texture2D(heightmap, sphericalUV).r;
        vHeight = h;
        
        // Displace along normal
        // scale determines max height in world units
        vec3 p = position + normal * (h * scale);
        vPosition = p;
        
        gl_Position = viewProjection * world * vec4(p, 1.0);
    }
`;

Effect.ShadersStore["displacementFragmentShader"] = `
    #extension GL_OES_standard_derivatives : enable
    precision highp float;

    // Varyings
    varying float vHeight;
    varying vec3 vNormal;
    varying vec3 vPosition;

    // Uniforms
    uniform vec3 color; // Base color (fallback)
    uniform vec3 lightDirection; // Dynamic sun direction (normalized)
    uniform float seaLevel; // Normalized sea level (0-1 in heightmap space)
    uniform float minElevation; // Minimum elevation in meters
    uniform float maxElevation; // Maximum elevation in meters
    
    // Data textures
    uniform sampler2D diffuseTex;  // The actual color map (replacing procedural rock)
    uniform sampler2D specularTex; // Specular map (water shininess)
    uniform sampler2D materialTex; // R=hardness, G=continental, B=sediment
    uniform sampler2D iceTex;      // R=ice thickness
    uniform sampler2D normalTex;   // Normal map for 3D shadows
    
    uniform bool hasDiffuseTex;
    uniform bool hasSpecularTex;
    uniform bool hasMaterialTex;
    uniform bool hasIceTex;
    uniform bool hasNormalTex;

    const float PI = 3.14159265359;

    // Calculate Spherical UV from 3D Position
    vec2 calculateUV(vec3 p) {
        vec3 v = normalize(p);
        float u = (atan(v.z, v.x) / (2.0 * PI)) + 0.5;
        float v_coord = (asin(v.y) / PI) + 0.5;
        return vec2(u, v_coord);
    }

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
        // Calculate UVs from 3D position
        vec2 uv = calculateUV(vPosition);

        // --- SEAM FIX ---
        // Calculate analytic derivatives
        vec2 uv_dx = dFdx(uv);
        vec2 uv_dy = dFdy(uv);
        
        // Check for wrapping in U (longitude)
        // If the derivative is too large, it means we wrapped across the Date Line
        if (abs(uv_dx.x) > 0.5) {
            uv_dx.x = -sign(uv_dx.x) * (1.0 - abs(uv_dx.x));
        }
        if (abs(uv_dy.x) > 0.5) {
            uv_dy.x = -sign(uv_dy.x) * (1.0 - abs(uv_dy.x));
        }
        // ----------------

        // Calculate perturbed normal
        vec3 normal = normalize(vNormal);
        
        if (hasNormalTex) {
            // Sample normal map (tangent space)
            // Use textureGrad for seam-safe sampling
            vec3 mapN;
            #ifdef GL_OES_standard_derivatives
                mapN = texture2DGradEXT(normalTex, uv, uv_dx, uv_dy).rgb * 2.0 - 1.0;
            #else
                mapN = texture2D(normalTex, uv).rgb * 2.0 - 1.0;
            #endif
            
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
        vec3 viewDir = normalize(vec3(0.0, 0.0, 0.0) - vPosition); // Simplified View Dir approximation or pass camera pos
        vec3 halfVector = normalize(lightDirection + viewDir);
        float NdotH = max(0.0, dot(normal, halfVector));
        
        // --- BASE SURFACE COLOR ---
        vec3 surfaceColor = vec3(0.5); // Default grey
        
        if (hasDiffuseTex) {
            // If we have a diffuse texture (real map), use it!
            #ifdef GL_OES_standard_derivatives
                surfaceColor = texture2DGradEXT(diffuseTex, uv, uv_dx, uv_dy).rgb;
            #else
                surfaceColor = texture2D(diffuseTex, uv).rgb;
            #endif
        } else {
            // Procedural Coloring (Fallback or Data View)
            // Sample material data if available
            float hardness = 0.5;
            bool isContinental = true;
            float sediment = 0.0;
            
            if (hasMaterialTex) {
                vec4 matData;
                #ifdef GL_OES_standard_derivatives
                    matData = texture2DGradEXT(materialTex, uv, uv_dx, uv_dy);
                #else
                    matData = texture2D(materialTex, uv);
                #endif
                hardness = matData.r;
                isContinental = matData.g > 0.5;
                sediment = matData.b;
            }
            
            // Get base rock color from material data
            surfaceColor = getRockColor(hardness, isContinental);
            
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
        }
        
        // --- OVERLAYS (Ice, etc.) ---
        // Apply ice overlay
        if (hasIceTex) {
            float ice;
            #ifdef GL_OES_standard_derivatives
                ice = texture2DGradEXT(iceTex, uv, uv_dx, uv_dy).r;
            #else
                ice = texture2D(iceTex, uv).r;
            #endif
            
            if (ice > 0.05) {
                vec3 iceCol = getIceColor(ice);
                surfaceColor = mix(surfaceColor, iceCol, min(ice * 2.0, 0.95));
            }
        }
        
        // SPECULAR HIGHLIGHT
        float specularPower = 30.0;
        float specularIntensity = 0.0;
        if (hasSpecularTex) {
             #ifdef GL_OES_standard_derivatives
                specularIntensity = texture2DGradEXT(specularTex, uv, uv_dx, uv_dy).r;
            #else
                specularIntensity = texture2D(specularTex, uv).r;
            #endif
        }
        vec3 specular = vec3(pow(NdotH, specularPower)) * specularIntensity;
        
        vec3 finalColor = surfaceColor * (0.25 + 0.75 * ndotl) + specular; // Ambient + Diffuse + Specular

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

    public getMaterial(): ShaderMaterial | null {
        return this.material;
    }

    public createMaterial(heightmap: Texture | null = null, scale: number = 0.5): ShaderMaterial {
        if (!heightmap) {
            // Create a default flat heightmap (1x1 pixel black)
            const data = new Uint8Array([0, 0, 0, 255]); // Black, full alpha
            heightmap = new Texture(
                "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
                this.scene,
                true, false, Texture.NEAREST_SAMPLINGMODE
            );
        }

        this.heightmap = heightmap;

        this.material = new ShaderMaterial(
            "displacementShader",
            this.scene,
            {
                vertex: "displacement",
                fragment: "displacement",
            },
            {
                attributes: ["position", "normal"], // Removed 'uv' as we don't rely on mesh UVs primarily anymore, but Vertex shader calculates its own
                uniforms: ["world", "viewProjection", "scale", "color", "lightDirection", "seaLevel", "minElevation", "maxElevation",
                    "hasDiffuseTex", "hasSpecularTex", "hasMaterialTex", "hasIceTex", "hasNormalTex"],
                samplers: ["heightmap", "diffuseTex", "specularTex", "materialTex", "iceTex", "normalTex"],
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
        this.material.setInt("hasDiffuseTex", 0);
        this.material.setInt("hasSpecularTex", 0);
        this.material.setInt("hasMaterialTex", 0);
        this.material.setInt("hasIceTex", 0);
        this.material.setInt("hasNormalTex", 0);

        // Render back faces too just in case, though sphere usually doesn't need it
        this.material.backFaceCulling = true;

        return this.material;
    }

    /**
     * Set the main diffuse texture (e.g. the satellite map).
     */
    public setDiffuseTexture(texture: Texture): void {
        if (this.material) {
            this.material.setTexture("diffuseTex", texture);
            this.material.setInt("hasDiffuseTex", 1);
            console.log("[DisplacementShader] Diffuse texture set");
        }
    }

    public setSpecularTexture(texture: Texture): void {
        if (this.material) {
            this.material.setTexture("specularTex", texture);
            this.material.setInt("hasSpecularTex", 1);
            console.log("[DisplacementShader] Specular texture set");
        }
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
