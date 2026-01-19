/**
 * MoonFPVController - First-person view controller for moon surfaces.
 * 
 * Extends the concept of FirstPersonController with moon-specific features:
 * - Dynamic gravity calculated from moon's mass and radius
 * - Horizon-based fog distance (smaller moons = closer horizon)
 * - Procedural crater terrain generation
 * - Observation mode only (no editing/interaction)
 */

import type { Scene } from "@babylonjs/core/scene";
import type { Camera } from "@babylonjs/core/Cameras/camera";
import type { Mesh } from "@babylonjs/core/Meshes/mesh";
import { UniversalCamera } from "@babylonjs/core/Cameras/universalCamera";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Color3 } from "@babylonjs/core/Maths/math.color";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { DirectionalLight } from "@babylonjs/core/Lights/directionalLight";
import { HemisphericLight } from "@babylonjs/core/Lights/hemisphericLight";
import type { IPlayerController } from './interfaces';
import type { MoonData, MoonFPVParams } from "$lib/types/moon";
import { getMoonFPVParams, calculateHorizonDistance } from "$lib/types/moon";

// Side-effect imports
import "@babylonjs/core/Collisions/collisionCoordinator";
import "@babylonjs/core/Cameras/Inputs/freeCameraKeyboardMoveInput";
import "@babylonjs/core/Cameras/Inputs/freeCameraMouseInput";

export interface MoonFPVOptions {
    /** Starting position on moon surface */
    startPosition?: Vector3;
    /** Eye height above ground */
    eyeHeight?: number;
    /** Planet diameter in km for sky display */
    planetDiameterKm?: number;
    /** Moon orbital distance in km for angular size calculation */
    moonDistanceKm?: number;
}

interface Position {
    lat: number;
    lon: number;
    altitude: number;
}

/**
 * Controller for first-person observation on moon surfaces.
 * Gravity and movement are calculated from moon properties.
 */
export class MoonFPVController implements IPlayerController {
    private scene: Scene;
    private camera: UniversalCamera;
    private moonParams: MoonFPVParams;
    private velocity: Vector3 = Vector3.Zero();
    private isGrounded: boolean = true;
    private disposed: boolean = false;
    private terrain: Mesh | null = null;
    private eyeHeight: number;

    // Moon-specific
    private moon: MoonData;
    private sunLight: DirectionalLight | null = null;
    private ambientLight: HemisphericLight | null = null;
    private planetMesh: Mesh | null = null;
    private starfieldMesh: Mesh | null = null;
    private sunMesh: Mesh | null = null;
    private planetDiameterKm: number = 12742; // Default Earth diameter
    private moonDistanceKm: number = 384400; // Default Moon distance

    // Terrain properties
    private terrainSize: number = 500; // Size of terrain
    private wrapObserver: any = null; // Observer for position wrapping
    private moonRadiusM: number = 1737000; // Moon radius in meters for curvature
    private playerLongitude: number = 0; // Player longitude on moon surface
    private planetMaterial: StandardMaterial | null = null; // For texture updates
    private planetTextureUrl: string | null = null; // Current texture blob URL

    // Time-based orbital mechanics
    private startTime: number = 0; // Time when FPV started (ms)
    private lunarDaySeconds: number = 60; // How long a "lunar day" lasts in real seconds (accelerated)
    private orbitalPhase: number = 0; // Current orbital phase (0 to 2π)

    constructor(scene: Scene, moon: MoonData, options: MoonFPVOptions = {}) {
        this.scene = scene;
        this.moon = moon;
        this.moonParams = getMoonFPVParams(moon);
        this.eyeHeight = options.eyeHeight ?? 1.778; // 5'10" in meters
        this.planetDiameterKm = options.planetDiameterKm ?? 12742;
        // Use passed km distance, or convert from meters if not provided
        // moon.distance is in meters, divide by 1e6 to get km
        this.moonDistanceKm = options.moonDistanceKm ?? (moon.distance / 1e6);

        // Calculate moon radius for terrain curvature
        this.moonRadiusM = moon.radius * 1000; // moon.radius is in km

        // Start position: longitude 70 degrees (planet at 20 degrees altitude)
        // Terrain half-size is 250 (500/2). 70/90 * 250 ≈ 194
        const startX = 194;
        const startPos = new Vector3(startX, this.eyeHeight + 2, 0);

        // Create FPS camera
        this.camera = new UniversalCamera("moonFPSCamera", startPos, scene);
        this.camera.setTarget(startPos.add(new Vector3(0, 0, 10))); // Look forward
        this.camera.ellipsoid = new Vector3(0.5, 1.0, 0.5);
        this.camera.ellipsoidOffset = new Vector3(0, 1.0, 0);
        this.camera.checkCollisions = true;
        this.camera.applyGravity = true;

        // Standard Y-down gravity scaled for moon
        scene.gravity = new Vector3(0, -this.moonParams.gravity / 50, 0);

        // Set camera speed based on moon gravity
        this.camera.speed = this.moonParams.moveSpeed;
        this.camera.maxZ = 50000; // Ensure we can see distant sky objects

        // Movement keys
        this.camera.keysUp = [87]; // W
        this.camera.keysDown = [83]; // S
        this.camera.keysLeft = [65]; // A
        this.camera.keysRight = [68]; // D
        this.camera.angularSensibility = 500;

        // Disable fog for clear sky view
        this.scene.fogMode = 0;

        // Create flat terrain with craters
        this.createFlatTerrain();

        // Create starfield sky dome (large sphere around player)
        this.createStarfield();

        // Create sun in the sky
        this.createSun();

        // Create planet in the sky
        this.createPlanetInSky();

        // Setup lighting
        this.setupLighting();

        // Setup position wrapping for circumnavigation
        this.setupPositionWrapping();

        // Initialize time for orbital mechanics
        this.startTime = performance.now();

        console.log(`[MoonFPV] Created flat terrain for ${moon.name}: terrainSize=${this.terrainSize}, gravity=${this.moonParams.gravity.toFixed(3)} m/s²`);

        // Debug logger
        setInterval(() => {
            if (this.disposed || !this.planetMesh) return;
            console.log(`[MoonFPV] Status: Lat=${this.playerLongitude.toFixed(1)}, PlanetEnabled=${this.planetMesh.isEnabled()}, PlanetPos=${this.planetMesh.position.toString()}`);
        }, 2000);
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
     * Create skybox for star background.
     */
    private createStarfield(): void {
        // Create skybox (large box that follows camera)
        // Must be larger than planet/sun to prevent clipping interaction
        // Planet far side can be ~26000 units away. Camera maxZ is 50000.
        const skyboxSize = 45000;
        this.starfieldMesh = MeshBuilder.CreateBox("skybox", {
            size: skyboxSize
        }, this.scene);

        // Skybox material - dark with subtle stars
        const skyboxMat = new StandardMaterial("skyboxMat", this.scene);
        skyboxMat.backFaceCulling = false;
        skyboxMat.diffuseColor = new Color3(0, 0, 0);
        skyboxMat.emissiveColor = new Color3(0.01, 0.01, 0.02);
        skyboxMat.specularColor = new Color3(0, 0, 0);
        skyboxMat.disableLighting = true;
        this.starfieldMesh.material = skyboxMat;
        this.starfieldMesh.renderingGroupId = 0; // Background layer

        // Skybox stays centered on camera at infinite distance
        this.starfieldMesh.infiniteDistance = true;
        this.starfieldMesh.checkCollisions = false;

        console.log(`[MoonFPV] Created skybox, size=${skyboxSize}`);
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
        this.planetMaterial.diffuseColor = new Color3(0.3, 0.5, 0.7); // Earth-like blue
        this.planetMaterial.emissiveColor = new Color3(0.1, 0.15, 0.2); // Slight glow
        this.planetMaterial.specularColor = new Color3(0.1, 0.1, 0.1);
        this.planetMesh.material = this.planetMaterial;

        console.log(`[MoonFPV] Planet mesh created: enabled=${this.planetMesh.isEnabled()}, pos=${this.planetMesh.position}`);
    }

    /**
     * Update planet texture from simulation blob.
     * Call this when new render data arrives via websocket.
     */
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
        this.planetMaterial.emissiveColor = new Color3(0.05, 0.05, 0.05); // Reduce emissive when textured

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
