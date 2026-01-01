import type { IShaderProvider } from "./interfaces";
import { Effect } from "@babylonjs/core/Materials/effect";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import type { Scene } from "@babylonjs/core/scene";
import type { Texture } from "@babylonjs/core/Materials/Textures/texture";

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

    // Simple directional light (approximate sun)
    const vec3 lightDir = normalize(vec3(1.0, 0.0, 0.0));

    void main(void) {
        // Simple Lambertian lighting
        // Note: For correct lighting on displaced surface, we ideally need to 
        // recompute normals from the heightmap gradient, but that's expensive.
        // For Phase 1, we'll use original sphere normals which is "okay" for smooth planets
        // but not great for mountains. Phase 2 improvement: normal map or derivative normals.
        
        float ndotl = max(0.0, dot(vNormal, lightDir));
        
        // Coloring based on height (simple ramp)
        vec3 surfaceColor = vec3(0.1, 0.1, 0.3); // Deep ocean
        
        if (vHeight > 0.01) {
            surfaceColor = vec3(0.2, 0.5, 0.2); // Land
        }
        if (vHeight > 0.6) {
            surfaceColor = vec3(0.5, 0.5, 0.5); // Mountain
        }
        if (vHeight > 0.8) {
            surfaceColor = vec3(1.0, 1.0, 1.0); // Snow
        }
        
        vec3 finalColor = surfaceColor * (0.2 + 0.8 * ndotl); // Ambient + Diffuse

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
                uniforms: ["world", "viewProjection", "scale", "color"],
                samplers: ["heightmap"],
            }
        );

        this.material.setTexture("heightmap", heightmap);
        this.material.setFloat("scale", scale);

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

    public dispose(): void {
        if (this.material) {
            this.material.dispose();
            this.material = null;
        }
    }
}
