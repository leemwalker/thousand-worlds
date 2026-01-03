/**
 * HorizonRenderer - Atmosphere and horizon rendering for flying view.
 * Phase 4: Visual Polish
 * 
 * Handles:
 * - Atmospheric scattering for realistic sky
 * - Horizon gradient blending
 * - Distance fog matching atmosphere
 * - Day/night sky color transitions
 */

import { Effect } from "@babylonjs/core/Materials/effect";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import type { Scene } from "@babylonjs/core/scene";
import type { Mesh } from "@babylonjs/core/Meshes/mesh";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Color3 } from "@babylonjs/core/Maths/math.color";

// Atmosphere shader (skybox replacement)
Effect.ShadersStore["atmosphereVertexShader"] = `
    precision highp float;

    attribute vec3 position;
    
    uniform mat4 world;
    uniform mat4 viewProjection;

    varying vec3 vPosition;

    void main(void) {
        vPosition = position;
        gl_Position = viewProjection * world * vec4(position, 1.0);
    }
`;

Effect.ShadersStore["atmosphereFragmentShader"] = `
    precision highp float;

    varying vec3 vPosition;

    uniform vec3 sunDirection;
    uniform vec3 zenithColor;      // Color at top of sky
    uniform vec3 horizonColor;     // Color at horizon
    uniform vec3 groundColor;      // Color below horizon
    uniform vec3 sunColor;
    uniform float sunSize;
    uniform float atmosphereHeight;
    uniform float sunIntensity;

    void main(void) {
        vec3 viewDir = normalize(vPosition);
        
        // Calculate altitude angle (-1 = down, 0 = horizon, 1 = up)
        float altitude = viewDir.y;
        
        // Sky gradient based on altitude
        vec3 skyColor;
        
        if (altitude > 0.0) {
            // Above horizon - blend zenith to horizon
            float t = pow(altitude, 0.5); // Non-linear curve for more horizon color
            skyColor = mix(horizonColor, zenithColor, t);
        } else {
            // Below horizon - blend to ground color
            float t = clamp(-altitude * 3.0, 0.0, 1.0);
            skyColor = mix(horizonColor, groundColor, t);
        }
        
        // Sun disc
        float sunAngle = dot(viewDir, sunDirection);
        float sunDisc = smoothstep(1.0 - sunSize * 0.01, 1.0, sunAngle);
        
        // Sun glow (halo around sun)
        float sunGlow = pow(max(0.0, sunAngle), 8.0) * 0.5;
        
        // Atmospheric scattering near horizon (brighter near sun)
        float horizonGlow = 1.0 - abs(altitude);
        horizonGlow = pow(horizonGlow, 3.0);
        float nearSun = max(0.0, sunAngle);
        vec3 horizonScatter = horizonColor * horizonGlow * (1.0 + nearSun * 2.0);
        
        // Combine
        vec3 finalColor = skyColor + horizonScatter * 0.3;
        finalColor += sunColor * (sunDisc + sunGlow) * sunIntensity;
        
        gl_FragColor = vec4(finalColor, 1.0);
    }
`;

export interface HorizonRendererOptions {
    zenithColor?: Color3;       // Top of sky
    horizonColor?: Color3;      // At horizon
    groundColor?: Color3;       // Below horizon
    sunColor?: Color3;
    sunSize?: number;
    sunIntensity?: number;
}

/**
 * Renders atmospheric horizon for ground-level and flying views.
 */
export class HorizonRenderer {
    private scene: Scene;
    private skyDome: Mesh | null = null;
    private skyMaterial: ShaderMaterial | null = null;
    private sunDirection: Vector3 = new Vector3(1, 0.3, 0);
    private options: Required<HorizonRendererOptions>;

    constructor(scene: Scene, options: HorizonRendererOptions = {}) {
        this.scene = scene;
        this.options = {
            zenithColor: options.zenithColor ?? new Color3(0.15, 0.35, 0.65),
            horizonColor: options.horizonColor ?? new Color3(0.6, 0.75, 0.9),
            groundColor: options.groundColor ?? new Color3(0.2, 0.25, 0.3),
            sunColor: options.sunColor ?? new Color3(1.0, 0.95, 0.8),
            sunSize: options.sunSize ?? 3.0,
            sunIntensity: options.sunIntensity ?? 1.0
        };
    }

    /**
     * Create the sky dome mesh with atmosphere shader.
     */
    createSkyDome(radius: number = 100): Mesh {
        // Create inverted sphere (normals facing inward)
        this.skyDome = MeshBuilder.CreateSphere(
            "skyDome",
            { diameter: radius * 2, segments: 32, sideOrientation: 1 }, // 1 = inside
            this.scene
        );

        // Create atmosphere material
        this.skyMaterial = new ShaderMaterial(
            "atmosphereMaterial",
            this.scene,
            {
                vertex: "atmosphere",
                fragment: "atmosphere"
            },
            {
                attributes: ["position"],
                uniforms: [
                    "world", "viewProjection",
                    "sunDirection", "zenithColor", "horizonColor", "groundColor",
                    "sunColor", "sunSize", "sunIntensity", "atmosphereHeight"
                ]
            }
        );

        // Set initial values
        this.skyMaterial.setVector3("sunDirection", this.sunDirection);
        this.skyMaterial.setColor3("zenithColor", this.options.zenithColor);
        this.skyMaterial.setColor3("horizonColor", this.options.horizonColor);
        this.skyMaterial.setColor3("groundColor", this.options.groundColor);
        this.skyMaterial.setColor3("sunColor", this.options.sunColor);
        this.skyMaterial.setFloat("sunSize", this.options.sunSize);
        this.skyMaterial.setFloat("sunIntensity", this.options.sunIntensity);
        this.skyMaterial.setFloat("atmosphereHeight", 0.1);

        this.skyMaterial.backFaceCulling = false;
        this.skyMaterial.disableDepthWrite = true;

        this.skyDome.material = this.skyMaterial;
        this.skyDome.infiniteDistance = true; // Always renders behind everything

        return this.skyDome;
    }

    /**
     * Update sun direction (for day/night cycle).
     */
    setSunDirection(direction: Vector3): void {
        this.sunDirection = direction.clone();
        this.sunDirection.normalize();

        this.skyMaterial?.setVector3("sunDirection", this.sunDirection);

        // Adjust colors based on sun altitude (day/night)
        const sunAltitude = this.sunDirection.y;

        if (sunAltitude < 0) {
            // Night time - darken sky
            const nightFactor = Math.min(1.0, -sunAltitude * 2);
            const nightZenith = this.options.zenithColor.scale(1 - nightFactor * 0.9);
            const nightHorizon = this.options.horizonColor.scale(1 - nightFactor * 0.8);

            this.skyMaterial?.setColor3("zenithColor", nightZenith);
            this.skyMaterial?.setColor3("horizonColor", nightHorizon);
            this.skyMaterial?.setFloat("sunIntensity", 0.1); // Dim sun
        } else if (sunAltitude < 0.2) {
            // Sunrise/sunset - warm colors
            const dawnFactor = sunAltitude / 0.2;
            const dawnHorizon = new Color3(
                0.9 * (1 - dawnFactor) + this.options.horizonColor.r * dawnFactor,
                0.5 * (1 - dawnFactor) + this.options.horizonColor.g * dawnFactor,
                0.3 * (1 - dawnFactor) + this.options.horizonColor.b * dawnFactor
            );

            this.skyMaterial?.setColor3("horizonColor", dawnHorizon);
            this.skyMaterial?.setFloat("sunIntensity", 0.5 + dawnFactor * 0.5);
        } else {
            // Day time - normal colors
            this.skyMaterial?.setColor3("zenithColor", this.options.zenithColor);
            this.skyMaterial?.setColor3("horizonColor", this.options.horizonColor);
            this.skyMaterial?.setFloat("sunIntensity", this.options.sunIntensity);
        }
    }

    /**
     * Get fog color matching current horizon.
     */
    getFogColor(): Color3 {
        // Return current horizon color for consistent fog
        if (this.skyMaterial) {
            return this.options.horizonColor;
        }
        return new Color3(0.7, 0.8, 0.9);
    }

    /**
     * Update sky dome position to follow camera.
     */
    followCamera(cameraPosition: Vector3): void {
        if (this.skyDome) {
            this.skyDome.position = cameraPosition;
        }
    }

    /**
     * Set custom atmosphere colors.
     */
    setAtmosphereColors(zenith: Color3, horizon: Color3, ground: Color3): void {
        this.options.zenithColor = zenith;
        this.options.horizonColor = horizon;
        this.options.groundColor = ground;

        this.skyMaterial?.setColor3("zenithColor", zenith);
        this.skyMaterial?.setColor3("horizonColor", horizon);
        this.skyMaterial?.setColor3("groundColor", ground);
    }

    /**
     * Get sky dome mesh.
     */
    getSkyDome(): Mesh | null {
        return this.skyDome;
    }

    /**
     * Dispose of resources.
     */
    dispose(): void {
        this.skyDome?.dispose();
        this.skyMaterial?.dispose();
        this.skyDome = null;
        this.skyMaterial = null;
    }
}
