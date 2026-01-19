/**
 * MoonFPVController - First-person view controller for moon surfaces.
 * 
 * Extends the concept of FirstPersonController with moon-specific features:
 * - Dynamic gravity calculated from moon's mass and radius
 * - Horizon-based fog distance (smaller moons = closer horizon)
 * - Procedural crater terrain generation
 * - Observation mode only (no editing/interaction)
 */

import { TileGlobeManager } from "./TileGlobeManager";
import { TransformNode } from "@babylonjs/core/Meshes/transformNode";
import { TerrainComputeShader } from "./TerrainComputeShader";

export interface MoonFPVOptions {
    /** Starting position on moon surface */
    startPosition?: Vector3;
    /** Eye height above ground */
    eyeHeight?: number;
    /** Planet diameter in km for sky display */
    planetDiameterKm?: number;
    /** Moon orbital distance in km for angular size calculation */
    moonDistanceKm?: number;
    /** List of all moons for sibling rendering */
    siblingMoons?: MoonData[];
    /** Callback to send data requests */
    sendCommand?: (action: string, message?: string) => void;
    /** Compute shader for high-detail planet rendering */
    computeShader?: TerrainComputeShader;
}

// ... imports remain the same ...

export class MoonFPVController implements IPlayerController {
    // ... params ...
    private sendCommand?: (action: string, message?: string) => void;
    private computeShader?: TerrainComputeShader;
    private tileGlobeManager: TileGlobeManager | null = null;
    private planetNode: TransformNode | null = null;

    // ... existing props ...

    constructor(scene: Scene, moon: MoonData, options: MoonFPVOptions = {}) {
        this.scene = scene;
        this.moon = moon;
        this.moonParams = getMoonFPVParams(moon);
        this.eyeHeight = options.eyeHeight ?? 1.778;
        this.planetDiameterKm = options.planetDiameterKm ?? 12742;
        this.moonDistanceKm = options.moonDistanceKm ?? (moon.distance / 1e6);
        this.siblingMoons = options.siblingMoons ?? [];
        this.sendCommand = options.sendCommand;
        this.computeShader = options.computeShader;

        // ... rest of constructor ...
        // ... call createPlanetInSky() ...
    }

    /**
     * Create the parent planet visible in the sky.
     * Uses TileGlobeManager for high-detail WebGPU rendering.
     */
    private createPlanetInSky(): void {
        // Calculate angular size of planet as seen from moon
        const planetRadiusKm = this.planetDiameterKm / 2;
        const distanceKm = Math.max(this.moonDistanceKm, planetRadiusKm * 1.1);

        // Angular radius (alpha)
        let angularRadiusRad = Math.asin(planetRadiusKm / distanceKm);

        // Visual Scale Factor (Cinematic View)
        const PLANET_ANGULAR_SCALE = 15.0;
        angularRadiusRad *= PLANET_ANGULAR_SCALE;

        const maxRad = (70 * Math.PI) / 180;
        if (angularRadiusRad > maxRad) angularRadiusRad = maxRad;

        const angularSizeDeg = (angularRadiusRad * 2) * (180 / Math.PI);
        const skyDistance = this.terrainSize * 1.5;

        // Calculate radius needed to subtend this angle at skyDistance
        // sin(alpha) = R / (R + skyDistance) => R = skyDistance * sin(alpha) / (1 - sin(alpha))
        // But simpler: Place center at 'centerDistance' and radius 'sceneRadius'.
        // Surface is at skyDistance.

        const sinAlpha = Math.sin(angularRadiusRad);
        const relativeScale = (1 / sinAlpha) - 1;
        const safeScale = Math.max(relativeScale, 0.1);

        const sceneRadius = skyDistance / safeScale;
        const centerDistance = sceneRadius + skyDistance;

        console.log(`[MoonFPV] Planet: ${this.planetDiameterKm}km @ ${distanceKm.toFixed(1)}km = ${angularSizeDeg.toFixed(2)}°`);
        console.log(`[MoonFPV] Rendering: R=${sceneRadius.toFixed(1)}, D=${centerDistance.toFixed(1)}`);

        // Create TransformNode as parent for TileGlobeManager
        this.planetNode = new TransformNode("planetNode", this.scene);
        this.planetNode.position = new Vector3(0, centerDistance * 0.8, centerDistance * 0.5);
        this.planetNode.renderingGroupId = 0;

        // Store for update linkage
        this.planetMesh = this.planetNode as any as Mesh; // Treat node as mesh for positioning logic
        (this.planetMesh as any).orbitDistance = centerDistance; // Store for updateCelestialBodies

        if (this.sendCommand && this.computeShader) {
            console.log(`[MoonFPV] Initializing WebGPU Planet with Radius=${sceneRadius}`);

            this.tileGlobeManager = new TileGlobeManager(
                this.scene,
                this.planetNode,
                this.sendCommand,
                {
                    maxLevel: 4,
                    maxActiveTiles: 50,
                    computeShader: this.computeShader,
                    radius: sceneRadius,
                    forceLevel: 3 // High detail for sky view
                }
            );
            this.tileGlobeManager.enable();
        } else {
            // Fallback to simple sphere if no backend connection or WebGPU
            console.warn("[MoonFPV] WebGPU/Backend not available - Falling back to simple sphere");
            const sphere = MeshBuilder.CreateSphere("moonFPVPlanet", {
                diameter: sceneRadius * 2,
                segments: 64
            }, this.scene);
            sphere.parent = this.planetNode;

            this.planetMaterial = new StandardMaterial("moonFPVPlanetMat", this.scene);
            this.planetMaterial.diffuseColor = new Color3(0.3, 0.5, 0.7);
            this.planetMaterial.emissiveColor = new Color3(0.1, 0.15, 0.2);
            this.planetMaterial.specularColor = new Color3(0, 0, 0);
            sphere.material = this.planetMaterial;
        }

        console.log(`[MoonFPV] Planet node created: enabled=${this.planetNode.isEnabled()}`);
    }

    // update method needs to call tileGlobeManager.update

    public update(): void {
        this.tileGlobeManager?.update(this.camera);
    }

    public dispose() {
        this.disposed = true;
        this.tileGlobeManager?.dispose();
        // ... existing disposal ...
        if (this.scene.onBeforeRenderObservable.hasObservers()) {
            this.scene.onBeforeRenderObservable.remove(this.wrapObserver);
        }
    }

    /**
     * Create flat moon terrain with craters.
     */
    private createFlatTerrain(): void {
        const subdivisions = 128;

        this.terrain = MeshBuilder.CreateGround("moonTerrain", {
            width: this.terrainSize,
            height: this.terrainSize,
            subdivisions: subdivisions,
            updatable: true
        }, this.scene);

        // Apply crater deformations
        this.applyFlatCraters(subdivisions);

        // Material based on moon color
        const terrainMat = new StandardMaterial("moonTerrainMat", this.scene);
        const moonColorHex = this.moon.color || "#888888";
        const r = parseInt(moonColorHex.slice(1, 3), 16) / 255;
        const g = parseInt(moonColorHex.slice(3, 5), 16) / 255;
        const b = parseInt(moonColorHex.slice(5, 7), 16) / 255;

        terrainMat.diffuseColor = new Color3(r * 0.8, g * 0.8, b * 0.8);
        terrainMat.specularColor = new Color3(0.1, 0.1, 0.1);
        this.terrain.material = terrainMat;
        this.terrain.checkCollisions = true;

        console.log(`[MoonFPV] Created flat terrain, size=${this.terrainSize}`);
    }

    /**
     * Apply crater deformations to flat terrain.
     */
    private applyFlatCraters(subdivisions: number): void {
        if (!this.terrain) return;

        const positions = this.terrain.getVerticesData('position');
        if (!positions) return;

        // Create random craters
        const craters: Array<{ x: number, z: number, radius: number, depth: number }> = [];
        const craterCount = 15 + Math.floor(Math.random() * 10);
        const halfSize = this.terrainSize / 2;

        for (let i = 0; i < craterCount; i++) {
            craters.push({
                x: (Math.random() - 0.5) * this.terrainSize * 0.8,
                z: (Math.random() - 0.5) * this.terrainSize * 0.8,
                radius: 10 + Math.random() * 30,
                depth: 0.5 + Math.random() * 2
            });
        }

        // Apply curvature and crater deformations
        for (let i = 0; i < positions.length; i += 3) {
            const x = positions[i] ?? 0;
            const z = positions[i + 2] ?? 0;

            // Calculate distance from terrain center
            const distFromCenter = Math.sqrt(x * x + z * z);

            // Apply curvature: drop = R - sqrt(R² - d²)
            // For large R, this approximates d²/(2R)
            let y = 0;
            if (distFromCenter < this.moonRadiusM) {
                y = -(distFromCenter * distFromCenter) / (2 * this.moonRadiusM);
            }

            for (const crater of craters) {
                const dx = x - crater.x;
                const dz = z - crater.z;
                const dist = Math.sqrt(dx * dx + dz * dz);

                if (dist < crater.radius) {
                    const t = dist / crater.radius;
                    y -= crater.depth * Math.cos(t * Math.PI * 0.5);
                    // Add rim
                    if (t > 0.7) {
                        y += crater.depth * 0.2 * (t - 0.7) / 0.3;
                    }
                }
            }

            // Add noise
            y += (Math.random() - 0.5) * 0.3;
            positions[i + 1] = y;
        }

        this.terrain.updateVerticesData('position', positions);
        this.terrain.createNormals(true);
    }

    /**
     * Create skybox for star background with photorealistic stars.
     */
    private createStarfield(): void {
        // Create skybox (large box that follows camera)
        const skyboxSize = 45000;
        this.starfieldMesh = MeshBuilder.CreateBox("skybox", {
            size: skyboxSize
        }, this.scene);

        // Create high-res dynamic texture for stars
        // 4096 resolution for crisp stars on large displays
        const textureSize = 4096;
        const starTexture = new DynamicTexture("starTexture", { width: textureSize, height: textureSize }, this.scene, false);
        const ctx = starTexture.getContext();

        // 1. Black space background
        ctx.fillStyle = "#000000";
        ctx.fillRect(0, 0, textureSize, textureSize);

        // 2. Generate Nebula/Galaxy Dust (Subtle noise)
        // Simple procedural noise simulation using semi-transparent gradients
        for (let i = 0; i < 20; i++) {
            const x = Math.random() * textureSize;
            const y = Math.random() * textureSize;
            const radius = 400 + Math.random() * 600;

            const gradient = ctx.createRadialGradient(x, y, 0, x, y, radius);
            const hue = 200 + Math.random() * 60; // Blues and Purples
            gradient.addColorStop(0, `hsla(${hue}, 60%, 10%, 0.04)`);
            gradient.addColorStop(0.5, `hsla(${hue}, 60%, 5%, 0.02)`);
            gradient.addColorStop(1, "transparent");

            ctx.fillStyle = gradient;
            ctx.beginPath();
            ctx.arc(x, y, radius, 0, Math.PI * 2);
            ctx.fill();
        }

        // 3. Generate Stars
        const numStars = 6000;
        for (let i = 0; i < numStars; i++) {
            const x = Math.random() * textureSize;
            const y = Math.random() * textureSize;
            const r = Math.random();

            let size = 0.5; // Tiny specks (defaults)
            let opacity = Math.random() * 0.8 + 0.2;
            let color = "#FFFFFF";

            // Star distribution
            if (r > 0.99) {
                // Bright stars (0.1%)
                size = Math.random() * 2.0 + 1.5;
                opacity = 1.0;
                // Tint slightly blue or yellow
                color = Math.random() > 0.5 ? "#DDEEFF" : "#FFF8DD";
            } else if (r > 0.9) {
                // Medium stars (10%)
                size = Math.random() * 1.2 + 0.8;
                opacity = 0.8;
            } else {
                // Dim/Distant stars (90%)
                size = Math.random() * 0.8 + 0.2;
                opacity = Math.random() * 0.5 + 0.1;
            }

            // Draw star
            ctx.fillStyle = color;
            ctx.globalAlpha = opacity;
            ctx.beginPath();
            ctx.arc(x, y, size, 0, Math.PI * 2);
            ctx.fill();
        }

        ctx.globalAlpha = 1.0; // Reset
        starTexture.update();

        // Skybox material
        const skyboxMat = new StandardMaterial("skyboxMat", this.scene);
        skyboxMat.backFaceCulling = false;
        skyboxMat.diffuseColor = new Color3(0, 0, 0);
        skyboxMat.specularColor = new Color3(0, 0, 0);
        skyboxMat.emissiveTexture = starTexture;
        skyboxMat.disableLighting = true;

        this.starfieldMesh.material = skyboxMat;
        this.starfieldMesh.renderingGroupId = 0; // Background layer

        // Skybox stays centered on camera at infinite distance
        this.starfieldMesh.infiniteDistance = true;
        this.starfieldMesh.checkCollisions = false;

        console.log(`[MoonFPV] Created photorealistic skybox, size=${skyboxSize}, resolution=${textureSize}`);
    }

    /**
     * Create sun visual in the sky.
     */
    private createSun(): void {
        const sunDistance = this.terrainSize * 1.5 * 5; // Match updateCelestialBodies (skyDistance * 5)
        const sunSize = 30;

        this.sunMesh = MeshBuilder.CreateSphere("fpvSun", {
            diameter: sunSize,
            segments: 16
        }, this.scene);

        // Position sun above and to the side
        this.sunMesh.position = new Vector3(sunDistance, sunDistance * 0.7, 0);

        // Bright emissive material
        const sunMat = new StandardMaterial("sunMat", this.scene);
        sunMat.diffuseColor = new Color3(1, 1, 0.8);
        sunMat.emissiveColor = new Color3(1, 0.9, 0.7);
        sunMat.specularColor = new Color3(0, 0, 0);
        this.sunMesh.material = sunMat;

        console.log(`[MoonFPV] Created sun at distance ${sunDistance}`);
    }

    /**
     * Setup position wrapping for circumnavigation and celestial body updates.
     */
    private setupPositionWrapping(): void {
        const halfSize = this.terrainSize / 2;
        const wrapBuffer = 10;
        // moonRadiusM circumference / terrainSize = how many terrain widths = 360°
        const degreesPerMeter = 360 / (2 * Math.PI * this.moonRadiusM);

        this.wrapObserver = this.scene.onBeforeRenderObservable.add(() => {
            if (this.disposed || !this.camera) return;

            const pos = this.camera.position;
            let wrapped = false;

            // Wrap X position (represents longitude)
            if (pos.x > halfSize - wrapBuffer) {
                pos.x = -halfSize + wrapBuffer * 2;
                wrapped = true;
            } else if (pos.x < -halfSize + wrapBuffer) {
                pos.x = halfSize - wrapBuffer * 2;
                wrapped = true;
            }

            // Wrap Z position  
            if (pos.z > halfSize - wrapBuffer) {
                pos.z = -halfSize + wrapBuffer * 2;
                wrapped = true;
            } else if (pos.z < -halfSize + wrapBuffer) {
                pos.z = halfSize - wrapBuffer * 2;
                wrapped = true;
            }

            if (wrapped) {
                this.camera.position = pos;
            }

            // Track player longitude (x position maps to longitude)
            // Terrain center (x=0) = directly facing planet (lon=0)
            // For real circumference: lon = x * degreesPerMeter, but we scale to terrainSize
            this.playerLongitude = (pos.x / halfSize) * 90; // -90 to +90 degrees

            // Update celestial body positions based on longitude
            this.updateCelestialBodies();

            // Lock camera Y to eye height above terrain (terrain has curvature)
            const distFromCenter = Math.sqrt(pos.x * pos.x + pos.z * pos.z);
            const curvatureDrop = (distFromCenter * distFromCenter) / (2 * this.moonRadiusM);
            const targetY = -curvatureDrop + this.eyeHeight;
            pos.y = targetY;
            this.camera.position = pos;

            // Update Tile System
            this.tileGlobeManager?.update(this.camera);
        });

        console.log(`[MoonFPV] Position wrapping enabled with celestial tracking`);
    }

    /**
     * Update sun and planet positions based on time and player position.
     * 
     * Tidal locking mechanics:
     * - Planet stays at fixed position in sky (always facing the moon's near side)
     * - Planet rotates slowly (simulating moon orbiting around it)
     * - Sun moves across sky based on orbital phase (lunar day = orbital period)
     * - Walking to far side makes planet disappear below horizon
     */
    private updateCelestialBodies(): void {
        const skyDistance = this.terrainSize * 1.5;

        // Calculate orbital phase based on elapsed time
        const elapsedSeconds = (performance.now() - this.startTime) / 1000;
        this.orbitalPhase = (elapsedSeconds / this.lunarDaySeconds) * 2 * Math.PI;

        // === PLANET (tidally locked - fixed position, but rotates) ===
        if (this.planetMesh) {
            // Planet altitude based on player longitude (near side vs far side)
            // lon=0: planet overhead, lon=±90: planet at horizon, beyond: below horizon
            const planetAltitude = 90 - Math.abs(this.playerLongitude);

            if (planetAltitude > -5) { // Allow slightly below horizon for atmospheric glow if added
                this.planetMesh.setEnabled(true);
                const altRad = (planetAltitude / 90) * (Math.PI / 2);

                // Position relative to CAMERA to simulate infinite distance (no parallax)
                const cameraPos = this.camera.position;

                // Calculate position on the sky sphere
                // Note: We use Z for North/South alignment here simplistically
                // Use calculated orbit distance if available
                const dist = (this.planetMesh as any).orbitDistance || skyDistance;

                const skyPos = new Vector3(
                    0,
                    dist * Math.sin(altRad),
                    dist * Math.cos(altRad)
                );

                this.planetMesh.position = cameraPos.add(skyPos);

                // Rotate planet on Y axis to show moon orbiting around it
                // Negative rotation because we're viewing from moon's perspective
                this.planetMesh.rotation.y = -this.orbitalPhase;

                // Force rendering on top of skybox
                this.planetMesh.renderingGroupId = 0; // Background layer (behind terrain)

                // Update sibling moons relative to planet
                if (this.siblingMeshes.length > 0) {
                    this.updateSiblingMoons(skyPos, dist);
                }
            } else {
                // Far side of moon - planet below horizon
                this.planetMesh.setEnabled(false);
            }
        }

        // === SUN (moves across sky based on orbital phase) ===
        if (this.sunMesh) {
            // Sun position based on orbital phase (time-based, not walk-based)
            // Sun rises in east, sets in west over the course of a lunar day
            const sunAngle = this.orbitalPhase;

            // Sun traces a great circle across the sky
            const sunAltitude = Math.sin(sunAngle);
            const sunAzimuth = Math.cos(sunAngle);

            // Position relative to CAMERA
            const cameraPos = this.camera.position;

            // Sun distance needs to be significantly further than planet surface (skyDistance)
            // to avoid Z-fighting and ensure planet occludes sun
            const sunDistance = skyDistance * 5;

            const sunOffset = new Vector3(
                sunDistance * sunAzimuth,          // East-West
                sunDistance * Math.max(sunAltitude, -0.3), // Altitude (clamp for visibility)
                0                                  // North-South
            );

            const sunPos = cameraPos.add(sunOffset);
            this.sunMesh.position = sunPos;

            // Force rendering on top of skybox
            this.sunMesh.renderingGroupId = 0; // Background layer

            // Update directional light direction
            if (this.sunLight) {
                // Light direction points FROM sun toward origin (camera)
                // We use the offset direction
                this.sunLight.direction = sunOffset.negate().normalize();

                // Dim light when sun is below horizon (lunar night)
                this.sunLight.intensity = sunAltitude > 0 ? 1.5 : 0.1;
            }
        }
    }

    /**
     * Create the parent planet visible in the sky.
     * Size and position based on orbital distance.
     */
    private createPlanetInSky(): void {
        // Calculate angular size of planet as seen from moon
        const planetRadiusKm = this.planetDiameterKm / 2;
        const distanceKm = Math.max(this.moonDistanceKm, planetRadiusKm * 1.1); // Ensure we aren't inside the planet logic-wise

        // Angular radius (alpha)
        let angularRadiusRad = Math.asin(planetRadiusKm / distanceKm);

        // Visual Scale Factor (Cinematic View)
        // Real earth-moon size is ~1.9 degrees (too small for game feel)
        // 15x scale makes it ~28 degrees (Cinematic, fills nice portion of sky)
        const PLANET_ANGULAR_SCALE = 15.0;
        angularRadiusRad *= PLANET_ANGULAR_SCALE;

        // Clamp to prevent rendering glitches (max 140 degrees total size = 70 deg radius)
        const maxRad = (70 * Math.PI) / 180;
        if (angularRadiusRad > maxRad) angularRadiusRad = maxRad;

        const angularSizeDeg = (angularRadiusRad * 2) * (180 / Math.PI);

        // Place planet in sky
        // We want the SURFACE of the planet to be at skyDistance, not the center
        // This ensures the camera is never "inside" the mesh
        const skyDistance = this.terrainSize * 1.5; // e.g. 750

        // Math: sin(alpha) = R / D
        // where R is scene radius of planet, D is distance to center
        // We want D - R = skyDistance (surface is at skyDistance)
        // D = R + skyDistance
        // sin(alpha) = R / (R + skyDistance)
        // R (1/sin - 1) = skyDistance
        // R = skyDistance / (1/sin(alpha) - 1)

        const sinAlpha = Math.sin(angularRadiusRad);
        const relativeScale = (1 / sinAlpha) - 1;

        // If relativeScale is too small (alpha -> 90), clamp it
        const safeScale = Math.max(relativeScale, 0.1);

        const sceneRadius = skyDistance / safeScale;
        const sceneDiameter = sceneRadius * 2;
        const centerDistance = sceneRadius + skyDistance;

        console.log(`[MoonFPV] Planet: ${this.planetDiameterKm}km @ ${distanceKm.toFixed(1)}km = ${angularSizeDeg.toFixed(2)}°`);
        console.log(`[MoonFPV] Rendering: R=${sceneRadius.toFixed(1)}, D=${centerDistance.toFixed(1)} (Surface @ ${skyDistance})`);

        // Create planet sphere
        this.planetMesh = MeshBuilder.CreateSphere("moonFPVPlanet", {
            diameter: sceneDiameter,
            segments: 64 // Increased segments for large planet
        }, this.scene);

        // Position planet above horizon initially (will be updated by updateCelestialBodies)
        // Note: updateCelestialBodies needs to know the correct distance (centerDistance)
        // We'll store it in a property or recalculate it there.
        // For now, let's update updateCelestialBodies to use the vector length of the current position as the distance?
        // Or better, standardise on skyDistance representing the center distance?
        // Re-architecture: updateCelestialBodies assumes skyDistance is the location.
        // We should adjust updateCelestialBodies to use a dynamic distance.
        // Hack: We'll overwrite the 'skyDistance' logic in updateCelestialBodies by setting a custom property on the mesh
        // or just by setting the position here and trusting the update loop?
        // The update loop RECALCULATES position. So we must update the loop logic or the constant.

        // Let's store the centerDistance for the update loop to use
        (this.planetMesh as any).orbitDistance = centerDistance;

        this.planetMesh.position = new Vector3(0, centerDistance * 0.8, centerDistance * 0.5);
        this.planetMesh.renderingGroupId = 0; // Background layer

        // Planet material - will be textured with simulation data
        this.planetMaterial = new StandardMaterial("moonFPVPlanetMat", this.scene);
        this.planetMaterial.diffuseColor = new Color3(0.3, 0.5, 0.7); // Earth-like blue (placeholder)
        this.planetMaterial.emissiveColor = new Color3(0.1, 0.15, 0.2); // Slight glow
        this.planetMaterial.specularColor = new Color3(0, 0, 0); // Matte surface - no plastic shine
        this.planetMesh.material = this.planetMaterial;

        console.log(`[MoonFPV] Planet mesh created: enabled=${this.planetMesh.isEnabled()}, pos=${this.planetMesh.position}`);
    }

    /**
     * Create visual meshes for other moons in the system.
     */
    private createSiblingMoons(): void {
        this.siblingMoons.forEach((sibling, index) => {
            // Skip the current moon (we are standing on it)
            if (sibling.id === this.moon.id || sibling.name === this.moon.name) return;

            // Simplified visuals for siblings - just spheres for now
            // In future could use actual textures if available
            const siblingMesh = MeshBuilder.CreateSphere(`sibling_${sibling.name}`, {
                diameter: 1, // Will be scaled by distance
                segments: 32
            }, this.scene);

            // Create material
            const material = new StandardMaterial(`siblingMat_${sibling.name}`, this.scene);

            // Parse color hex
            const colorHex = sibling.color || "#AAAAAA";
            const r = parseInt(colorHex.slice(1, 3), 16) / 255;
            const g = parseInt(colorHex.slice(3, 5), 16) / 255;
            const b = parseInt(colorHex.slice(5, 7), 16) / 255;

            material.diffuseColor = new Color3(r, g, b);
            material.specularColor = new Color3(0.1, 0.1, 0.1);
            material.emissiveColor = new Color3(r * 0.2, g * 0.2, b * 0.2); // Slight glow

            siblingMesh.material = material;
            siblingMesh.renderingGroupId = 0; // Same as planet/sun

            // Store reference with data
            (siblingMesh as any).moonData = sibling;
            this.siblingMeshes.push(siblingMesh);

            console.log(`[MoonFPV] Created sibling moon mesh: ${sibling.name}`);
        });
    }

    /**
     * Update positions of sibling moons in the sky.
     * Calculated relative to planet position and our own orbit.
     */
    private updateSiblingMoons(planetSkyPos: Vector3, skyDistance: number): void {
        // Center position of the system (Planet) in camera space
        const centerPos = planetSkyPos;
        const cameraPos = this.camera.position;

        this.siblingMeshes.forEach(mesh => {
            const sibling = (mesh as any).moonData as MoonData;

            // Calculate orbital phase for sibling (different period)
            // Use same time basis as our own orbit
            const elapsedSeconds = (performance.now() - this.startTime) / 1000;
            const siblingPeriod = sibling.period > 0 ? sibling.period : 100; // fallback
            // Scale period relative to our lunar day for visualization
            // Real physics: T^2 prop a^3. 
            // We use relative speed: rate = (MyPeriod / SiblingPeriod) * MyRate
            // But here we rely on the passed period from backend if available, or just distance scaling?
            // Backend provides Period in seconds. our 'lunarDaySeconds' is a visual acceleration.
            // Let's assume 'lunarDaySeconds' represents 'period' of THIS moon.
            // Sibling Phase = time * (1/SiblingPeriod) * (AccelerationFactor)
            // AccelerationFactor = MyPeriod / lunarDaySeconds

            const myPeriod = this.moon.period || 100;
            const acceleration = myPeriod / this.lunarDaySeconds;
            // Actually, simplified:
            // SiblingAngle = (elapsed / SiblingPeriod) * 2PI * acceleration
            // OR simpler: assume periods are relative.

            // Let's just use simple Kepler scaling if period missing, or relative period
            const periodRatio = this.moon.period / sibling.period;
            const siblingPhase = this.orbitalPhase * periodRatio; // If I'm at phase X, sibling is at phase X * ratio

            // Offset angle (randomize start for variety based on ID/index)
            const seed = sibling.name.charCodeAt(sibling.name.length - 1);
            const offset = (seed % 10) * (Math.PI / 5);

            // Calculate position in orbit relative to planet
            // We are at Angle = 0 relative to planet in Sky View (Planet is fixed)
            // Actually, Planet fixed means WE are fixed. Sibling moves relative to US.

            // Vector to Sibling = VectorToPlanet + VectorPlanetToSibling
            // In Sky view:
            // Planet is at 'planetSkyPos'.
            // Sibling orbits Planet.
            // Visual Scale: We need to scale the physical orbit to the sky view.
            // Sky Distance = d to Planet.
            // Sibling physical distance = sibling.distance.
            // Our physical distance = this.moonDistanceKm * 1000.

            // Angular separation approx = (SiblingOrbitPos - MyOrbitPos) / DistToPlanet
            // Just project Sibling Orbit onto the sky plane at Planet Distance.

            const siblingDistM = sibling.distance;
            const myDistM = this.moonDistanceKm * 1000;

            // Relative radius in Sky Units
            // If planet is at 'skyDistance', then a moon at 'myDistM' would be at observer.
            // Scale factor = skyDistance / myDistM.
            const scale = skyDistance / myDistM;
            const siblingOrbitRadiusSky = siblingDistM * scale;

            // 3D Orbit orientation
            // Simplified: All orbit in X-Z plane (horizontal in sky if z-up?)
            // In Babylon Y is Up. Planet is at some altitude.
            // We'll align orbit plane with Planet's 'equator' roughly.

            // Sibling position relative to Planet Center
            // X = cos(angle), Z = sin(angle) (assuming Y-up is normal to orbit)
            const sx = Math.sin(siblingPhase + offset) * siblingOrbitRadiusSky;
            const sy = Math.sin((siblingPhase + offset) * 3) * (siblingOrbitRadiusSky * 0.1); // Slight inclination wobble
            const sz = Math.cos(siblingPhase + offset) * siblingOrbitRadiusSky;

            // Billboard/Face camera?
            // Position relative to Planet Surface Center?
            // Planet Mesh is at 'planetSkyPos'.
            const siblingPos = planetSkyPos.add(new Vector3(sx, sy, sz));

            // Apply to mesh (absolute world pos = camera + relative)
            mesh.position = cameraPos.add(siblingPos);

            // Scale mesh based on angular size
            // Real radius / Distance from observer approximately
            // Check distance from us to sibling?
            // Approx dist = skyDistance (since they are near planet)
            // Angular size = 2 * atan(R / D)
            // Visual size = 2 * R * (skyDistance / D_physical) ... 
            // Simplified: Mesh Size = (SiblingRadius * 2) * scale
            const siblingDiameterM = sibling.radius * 2;
            const visualDiameter = siblingDiameterM * scale;

            mesh.scaling = new Vector3(visualDiameter, visualDiameter, visualDiameter);

            // Look at camera
            mesh.lookAt(cameraPos);
        });
    }
    updatePlanetTexture(blob: Blob): void {
        if (!this.planetMaterial || !this.scene) {
            console.warn("[MoonFPV] Cannot update planet texture - material not ready");
            return;
        }

        // Revoke previous URL to prevent memory leak
        if (this.planetTextureUrl) {
            URL.revokeObjectURL(this.planetTextureUrl);
        }

        // Create new texture from blob
        this.planetTextureUrl = URL.createObjectURL(blob);
        const texture = new Texture(this.planetTextureUrl, this.scene);
        texture.hasAlpha = false;

        // Apply to planet material
        this.planetMaterial.diffuseTexture = texture;
        this.planetMaterial.diffuseColor = new Color3(1, 1, 1); // Reset to white so texture isn't tinted
        this.planetMaterial.emissiveColor = new Color3(0.1, 0.1, 0.1); // Moderate glow for visibility
        this.planetMaterial.specularColor = new Color3(0, 0, 0); // No plastic shine

        console.log(`[MoonFPV] Planet texture updated from blob, size=${blob.size}`);
    }

    /**
     * Setup lighting for moon surface.
     */
    private setupLighting(): void {
        // Very dim ambient light (space has minimal reflected light)
        this.ambientLight = new HemisphericLight("moonAmbient", new Vector3(0, 1, 0), this.scene);
        this.ambientLight.intensity = 0.03; // Very dim - only slight visibility in shadow
        this.ambientLight.groundColor = new Color3(0.01, 0.01, 0.01);

        // Directional sun light - direction points FROM the sun
        const sunDistance = this.terrainSize * 1.5;
        const sunDir = new Vector3(-sunDistance, -sunDistance * 0.7, 0).normalize();
        this.sunLight = new DirectionalLight("moonSun", sunDir, this.scene);
        this.sunLight.intensity = 1.5;
        this.sunLight.diffuse = new Color3(1.0, 0.98, 0.9); // Slightly warm

        console.log(`[MoonFPV] Lighting setup: ambient=0.03, sun dir=${sunDir.toString()}`);
    }

    /**
     * Get the calculated gravity for this moon.
     */
    getGravity(): number {
        return this.moonParams.gravity;
    }

    /**
     * Get the horizon distance for this moon.
     */
    getHorizonDistance(): number {
        return this.moonParams.horizonDistance;
    }

    /**
     * Get moon info for display.
     */
    getMoonInfo(): { name: string; gravity: number; horizon: number } {
        return {
            name: this.moon.name,
            gravity: this.moonParams.gravity,
            horizon: this.moonParams.horizonDistance
        };
    }

    // IPlayerController implementation

    handleInput(deltaTime: number): void {
        if (this.disposed) return;

        // Apply moon-specific gravity
        if (!this.isGrounded) {
            this.velocity.y -= this.moonParams.gravity * deltaTime;
        }
    }

    getPosition(): Position {
        const pos = this.camera.position;
        const distance = Math.sqrt(pos.x * pos.x + pos.z * pos.z);
        const lat = Math.atan2(pos.y, distance) * (180 / Math.PI);
        const lon = Math.atan2(pos.z, pos.x) * (180 / Math.PI);
        const altitude = pos.y;
        return { lat, lon, altitude };
    }

    getCamera(): Camera {
        return this.camera;
    }

    activate(): void {
        if (this.disposed) return;
        this.scene.activeCamera = this.camera;
        const canvas = this.scene.getEngine().getRenderingCanvas();
        if (canvas) {
            this.camera.attachControl(canvas, true);
        }
    }

    deactivate(): void {
        if (this.disposed) return;
        this.camera.detachControl();
    }

    jump(): void {
        if (this.isGrounded && !this.disposed) {
            // Jump height scales with lower gravity
            this.velocity.y = Math.sqrt(2 * this.moonParams.gravity * this.moonParams.jumpHeight);
            this.isGrounded = false;
        }
    }

    teleport(position: Vector3): void {
        if (this.disposed) return;
        this.camera.position = position.clone();
        this.velocity = Vector3.Zero();
    }

    isDisposed(): boolean {
        return this.disposed;
    }

    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;

        // Remove position wrapping observer
        if (this.wrapObserver) {
            this.scene.onBeforeRenderObservable.remove(this.wrapObserver);
            this.wrapObserver = null;
        }

        this.camera.detachControl();
        this.camera.dispose();
        this.terrain?.dispose();
        this.planetMesh?.dispose();
        this.starfieldMesh?.dispose();
        this.sunMesh?.dispose();
        this.sunLight?.dispose();
        this.ambientLight?.dispose();
        this.terrain = null;
        this.planetMesh = null;
        this.starfieldMesh = null;
        this.sunMesh = null;
    }
}
