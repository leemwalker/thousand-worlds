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
import { PointLight } from "@babylonjs/core/Lights/pointLight";
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
    private sunLight: PointLight | null = null;
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

        // Start position on flat terrain (y = eye height above ground)
        const startPos = new Vector3(0, this.eyeHeight + 2, 0);

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

        console.log(`[MoonFPV] Created flat terrain for ${moon.name}: terrainSize=${this.terrainSize}, gravity=${this.moonParams.gravity.toFixed(3)} m/s²`);
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
        const skyboxSize = this.terrainSize * 4;
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

        // Skybox stays centered on camera at infinite distance
        this.starfieldMesh.infiniteDistance = true;
        this.starfieldMesh.checkCollisions = false;

        console.log(`[MoonFPV] Created skybox, size=${skyboxSize}`);
    }

    /**
     * Create sun visual in the sky.
     */
    private createSun(): void {
        const sunDistance = this.terrainSize * 1.5;
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

            // Safety: prevent falling through terrain
            if (pos.y < -20) {
                pos.y = this.eyeHeight + 2;
                this.camera.position = pos;
            }
        });

        console.log(`[MoonFPV] Position wrapping enabled with celestial tracking`);
    }

    /**
     * Update sun and planet positions based on player's position on moon.
     */
    private updateCelestialBodies(): void {
        if (!this.planetMesh || !this.sunMesh) return;

        const skyDistance = this.terrainSize * 1.5;

        // Planet altitude: 90° when at lon=0 (directly facing planet), 0° at lon=±90° (horizon)
        // Below horizon when abs(lon) > 90°
        const planetAltitude = 90 - Math.abs(this.playerLongitude);

        if (planetAltitude > 0) {
            // Planet is above horizon
            this.planetMesh.setEnabled(true);
            const altRad = (planetAltitude / 180) * Math.PI;
            // Position planet in sky based on altitude
            this.planetMesh.position = new Vector3(
                skyDistance * Math.cos(altRad) * 0.8,
                skyDistance * Math.sin(altRad),
                0
            );
        } else {
            // Planet is below horizon - on the far side
            this.planetMesh.setEnabled(false);
        }

        // Sun position (roughly opposite to planet side, but with orbital mechanics)
        // For simplicity: sun orbits around the sky as player walks
        const sunAngle = (this.playerLongitude / 180) * Math.PI;
        this.sunMesh.position = new Vector3(
            skyDistance * Math.cos(sunAngle),
            skyDistance * 0.6 + skyDistance * 0.3 * Math.sin(sunAngle * 2),
            skyDistance * Math.sin(sunAngle) * 0.5
        );

        // Update sun light position
        if (this.sunLight) {
            this.sunLight.position = this.sunMesh.position.clone();
        }
    }

    /**
     * Create the parent planet visible in the sky.
     * Size and position based on orbital distance.
     */
    private createPlanetInSky(): void {
        // Calculate angular size of planet as seen from moon
        const planetRadiusKm = this.planetDiameterKm / 2;
        const distanceKm = this.moonDistanceKm;
        const angularSizeRad = 2 * Math.atan(planetRadiusKm / distanceKm);
        const angularSizeDeg = angularSizeRad * (180 / Math.PI);

        // Place planet in sky dome
        const skyDistance = this.terrainSize * 1.5;
        const planetSceneRadius = skyDistance * Math.tan(angularSizeRad / 2);
        const planetSceneDiameter = Math.max(planetSceneRadius * 2, 50);

        console.log(`[MoonFPV] Planet in sky: angularSize=${angularSizeDeg.toFixed(1)}°, diam=${planetSceneDiameter.toFixed(1)}`);

        // Create planet sphere
        this.planetMesh = MeshBuilder.CreateSphere("moonFPVPlanet", {
            diameter: planetSceneDiameter,
            segments: 32
        }, this.scene);

        // Position planet above horizon
        this.planetMesh.position = new Vector3(skyDistance * 0.5, skyDistance * 0.7, skyDistance * 0.5);

        // Earth-like planet material
        const planetMat = new StandardMaterial("moonFPVPlanetMat", this.scene);
        planetMat.diffuseColor = new Color3(0.2, 0.4, 0.6);
        planetMat.emissiveColor = new Color3(0.1, 0.15, 0.2);
        planetMat.specularColor = new Color3(0.1, 0.1, 0.1);
        this.planetMesh.material = planetMat;
    }

    /**
     * Setup lighting for moon surface.
     */
    private setupLighting(): void {
        // Ambient light (very dim - space)
        this.ambientLight = new HemisphericLight("moonAmbient", new Vector3(0, 1, 0), this.scene);
        this.ambientLight.intensity = 0.15;
        this.ambientLight.groundColor = new Color3(0.02, 0.02, 0.02);

        // Sun light - positioned at same location as sun mesh
        const sunDistance = this.terrainSize * 1.5;
        this.sunLight = new PointLight("moonSun", new Vector3(sunDistance, sunDistance * 0.7, 0), this.scene);
        this.sunLight.intensity = 2.0;
        this.sunLight.diffuse = new Color3(1.0, 0.98, 0.9); // Slightly warm
        this.sunLight.range = sunDistance * 3;
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
