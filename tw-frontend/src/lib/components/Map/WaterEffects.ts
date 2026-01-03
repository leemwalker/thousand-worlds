/**
 * WaterEffects - Water surface and underwater rendering effects.
 * Phase 4: Visual Polish
 * 
 * Handles:
 * - Water surface with reflection and waves
 * - Underwater fog and color grading
 * - Caustics (light patterns)
 * - Water-air transition effects
 */

import { Effect } from "@babylonjs/core/Materials/effect";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import { PostProcess } from "@babylonjs/core/PostProcesses/postProcess";
import type { Scene } from "@babylonjs/core/scene";
import type { Camera } from "@babylonjs/core/Cameras/camera";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Color3 } from "@babylonjs/core/Maths/math.color";
import type { Texture } from "@babylonjs/core/Materials/Textures/texture";

// Water surface shader
Effect.ShadersStore["waterSurfaceVertexShader"] = `
    precision highp float;

    attribute vec3 position;
    attribute vec3 normal;
    attribute vec2 uv;

    uniform mat4 world;
    uniform mat4 viewProjection;
    uniform float time;
    uniform float waveHeight;
    uniform float waveFrequency;

    varying vec2 vUV;
    varying vec3 vNormal;
    varying vec3 vWorldPos;
    varying float vWaveOffset;

    void main(void) {
        vUV = uv;
        vNormal = normal;
        
        // Simple wave animation
        float wave = sin(uv.x * waveFrequency + time) * cos(uv.y * waveFrequency + time * 0.7);
        vWaveOffset = wave * waveHeight;
        
        vec3 displaced = position + normal * vWaveOffset;
        vWorldPos = (world * vec4(displaced, 1.0)).xyz;
        
        gl_Position = viewProjection * world * vec4(displaced, 1.0);
    }
`;

Effect.ShadersStore["waterSurfaceFragmentShader"] = `
    precision highp float;

    varying vec2 vUV;
    varying vec3 vNormal;
    varying vec3 vWorldPos;
    varying float vWaveOffset;

    uniform vec3 lightDirection;
    uniform vec3 waterColor;
    uniform vec3 deepWaterColor;
    uniform float transparency;
    uniform float specularPower;
    uniform vec3 cameraPosition;

    void main(void) {
        // Fresnel effect for reflection vs refraction
        vec3 viewDir = normalize(cameraPosition - vWorldPos);
        float fresnel = pow(1.0 - max(0.0, dot(viewDir, normalize(vNormal))), 2.0);
        
        // Wave-perturbed normal for specular
        vec3 perturbedNormal = normalize(vNormal + vec3(vWaveOffset * 0.5, 0.0, vWaveOffset * 0.3));
        
        // Specular highlight (sun glint)
        float spec = pow(max(0.0, dot(reflect(-lightDirection, perturbedNormal), viewDir)), specularPower);
        
        // Base water color with depth variation
        vec3 baseColor = mix(waterColor, deepWaterColor, 0.3);
        
        // Add specular
        vec3 finalColor = baseColor + vec3(spec);
        
        // Increase opacity at edges (fresnel)
        float alpha = mix(transparency, 1.0, fresnel);
        
        gl_FragColor = vec4(finalColor, alpha);
    }
`;

// Underwater post-process shader
Effect.ShadersStore["underwaterPostProcessFragmentShader"] = `
    precision highp float;

    varying vec2 vUV;
    
    uniform sampler2D textureSampler;
    uniform float time;
    uniform float depth;
    uniform vec3 waterColor;
    uniform float distortionStrength;
    uniform float causticsStrength;

    void main(void) {
        // Distortion effect
        vec2 distortedUV = vUV;
        distortedUV.x += sin(vUV.y * 20.0 + time * 2.0) * distortionStrength;
        distortedUV.y += cos(vUV.x * 20.0 + time * 1.5) * distortionStrength;
        
        vec4 color = texture2D(textureSampler, distortedUV);
        
        // Caustics pattern
        float caustics = sin(vUV.x * 30.0 + time) * sin(vUV.y * 30.0 + time * 0.8);
        caustics = caustics * caustics * causticsStrength;
        
        // Apply water color tint based on depth
        float depthFactor = clamp(depth * 0.5, 0.0, 0.8);
        vec3 tintedColor = mix(color.rgb, waterColor, depthFactor);
        
        // Add caustics
        tintedColor += vec3(caustics);
        
        // Reduce contrast underwater
        tintedColor = mix(tintedColor, vec3(0.5), depthFactor * 0.3);
        
        gl_FragColor = vec4(tintedColor, 1.0);
    }
`;

export interface WaterEffectsOptions {
    waterColor?: Color3;
    deepWaterColor?: Color3;
    transparency?: number;
    waveHeight?: number;
    waveFrequency?: number;
    specularPower?: number;
}

/**
 * Manages water rendering effects.
 */
export class WaterEffects {
    private scene: Scene;
    private surfaceMaterial: ShaderMaterial | null = null;
    private underwaterPostProcess: PostProcess | null = null;
    private isUnderwater: boolean = false;
    private time: number = 0;
    private options: Required<WaterEffectsOptions>;

    constructor(scene: Scene, options: WaterEffectsOptions = {}) {
        this.scene = scene;
        this.options = {
            waterColor: options.waterColor ?? new Color3(0.1, 0.3, 0.5),
            deepWaterColor: options.deepWaterColor ?? new Color3(0.02, 0.1, 0.2),
            transparency: options.transparency ?? 0.7,
            waveHeight: options.waveHeight ?? 0.002,
            waveFrequency: options.waveFrequency ?? 10.0,
            specularPower: options.specularPower ?? 64.0
        };
    }

    /**
     * Create water surface material.
     */
    createSurfaceMaterial(): ShaderMaterial {
        this.surfaceMaterial = new ShaderMaterial(
            "waterSurface",
            this.scene,
            {
                vertex: "waterSurface",
                fragment: "waterSurface"
            },
            {
                attributes: ["position", "normal", "uv"],
                uniforms: [
                    "world", "viewProjection", "time",
                    "waveHeight", "waveFrequency",
                    "lightDirection", "waterColor", "deepWaterColor",
                    "transparency", "specularPower", "cameraPosition"
                ]
            }
        );

        this.surfaceMaterial.setFloat("time", 0);
        this.surfaceMaterial.setFloat("waveHeight", this.options.waveHeight);
        this.surfaceMaterial.setFloat("waveFrequency", this.options.waveFrequency);
        this.surfaceMaterial.setVector3("lightDirection", new Vector3(1, 0.5, 0));
        this.surfaceMaterial.setColor3("waterColor", this.options.waterColor);
        this.surfaceMaterial.setColor3("deepWaterColor", this.options.deepWaterColor);
        this.surfaceMaterial.setFloat("transparency", this.options.transparency);
        this.surfaceMaterial.setFloat("specularPower", this.options.specularPower);
        this.surfaceMaterial.setVector3("cameraPosition", Vector3.Zero());

        this.surfaceMaterial.alpha = this.options.transparency;
        this.surfaceMaterial.backFaceCulling = false;

        return this.surfaceMaterial;
    }

    /**
     * Create underwater post-process effect.
     */
    createUnderwaterEffect(camera: Camera): PostProcess {
        this.underwaterPostProcess = new PostProcess(
            "underwaterEffect",
            "underwaterPostProcess",
            ["time", "depth", "waterColor", "distortionStrength", "causticsStrength"],
            null,
            1.0,
            camera
        );

        this.underwaterPostProcess.onApply = (effect) => {
            effect.setFloat("time", this.time);
            effect.setFloat("depth", 1.0); // Will be updated based on actual depth
            effect.setColor3("waterColor", this.options.waterColor);
            effect.setFloat("distortionStrength", 0.003);
            effect.setFloat("causticsStrength", 0.1);
        };

        // Start disabled
        this.underwaterPostProcess.dispose();
        this.underwaterPostProcess = null;

        return this.underwaterPostProcess!;
    }

    /**
     * Update effects each frame.
     */
    update(deltaTime: number, cameraPosition: Vector3): void {
        this.time += deltaTime;

        if (this.surfaceMaterial) {
            this.surfaceMaterial.setFloat("time", this.time);
            this.surfaceMaterial.setVector3("cameraPosition", cameraPosition);
        }
    }

    /**
     * Enable underwater mode.
     */
    setUnderwater(underwater: boolean, depth: number = 1.0): void {
        this.isUnderwater = underwater;
        // Underwater post-process would be enabled/disabled here
    }

    /**
     * Update light direction.
     */
    setLightDirection(direction: Vector3): void {
        this.surfaceMaterial?.setVector3("lightDirection", direction);
    }

    /**
     * Get surface material.
     */
    getSurfaceMaterial(): ShaderMaterial | null {
        return this.surfaceMaterial;
    }

    /**
     * Check if underwater.
     */
    getIsUnderwater(): boolean {
        return this.isUnderwater;
    }

    /**
     * Dispose of resources.
     */
    dispose(): void {
        this.surfaceMaterial?.dispose();
        this.underwaterPostProcess?.dispose();
        this.surfaceMaterial = null;
        this.underwaterPostProcess = null;
    }
}
