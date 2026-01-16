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
    private planetDiameterKm: number = 12742; // Default Earth diameter
    private moonDistanceKm: number = 384400; // Default Moon distance

    constructor(scene: Scene, moon: MoonData, options: MoonFPVOptions = {}) {
        this.scene = scene;
        this.moon = moon;
        this.moonParams = getMoonFPVParams(moon);
        this.eyeHeight = options.eyeHeight ?? 1.7;
        this.planetDiameterKm = options.planetDiameterKm ?? 12742;
        // Use passed km distance, or convert from meters if not provided
        // moon.distance is in meters, divide by 1e6 to get km
        this.moonDistanceKm = options.moonDistanceKm ?? (moon.distance / 1e6);

        // Start position - higher up to account for crater depressions
        const startPos = options.startPosition ?? new Vector3(0, this.eyeHeight + 5, 0);

        // Create FPS camera
        this.camera = new UniversalCamera("moonFPSCamera", startPos, scene);
        this.camera.setTarget(startPos.add(new Vector3(0, 0, 1)));
        this.camera.ellipsoid = new Vector3(0.5, 1.0, 0.5); // Slightly taller ellipsoid
        this.camera.ellipsoidOffset = new Vector3(0, 1.0, 0); // Offset to keep camera above ground
        this.camera.checkCollisions = true;
        this.camera.applyGravity = true;

        // Set gravity for the scene based on moon gravity
        scene.gravity = new Vector3(0, -this.moonParams.gravity / 50, 0); // Scaled for scene units

        // Set camera speed based on moon gravity
        this.camera.speed = this.moonParams.moveSpeed;

        // Movement keys
        this.camera.keysUp = [87]; // W
        this.camera.keysDown = [83]; // S
        this.camera.keysLeft = [65]; // A
        this.camera.keysRight = [68]; // D
        this.camera.keysRotateLeft = [81]; // Q
        this.camera.keysRotateRight = [69]; // E
        this.camera.angularSensibility = 500;

        // Set fog based on horizon distance (visibility limit)
        this.setupHorizonFog();

        // Create procedural moon terrain
        this.createProceduralTerrain();

        // Create planet in the sky
        this.createPlanetInSky();

        // Setup lighting (sun direction + ambient)
        this.setupLighting();

        console.log(`[MoonFPV] Created for ${moon.name}: gravity=${this.moonParams.gravity.toFixed(3)} m/s², horizon=${this.moonParams.horizonDistance.toFixed(0)}m`);
    }

    /**
     * Set up fog based on moon horizon distance.
     * Smaller moons have closer horizons, so fog starts sooner.
     */
    private setupHorizonFog(): void {
        // Scale horizon for rendering (we don't render full scale)
        // Use log scale to make it usable for rendering
        const fogStart = Math.min(this.moonParams.horizonDistance / 1000, 500);
        const fogEnd = fogStart * 2;

        this.scene.fogMode = 2; // Exponential
        this.scene.fogDensity = 0.001;
        this.scene.fogColor = new Color3(0.02, 0.02, 0.03); // Dark space
        this.scene.fogStart = fogStart;
        this.scene.fogEnd = fogEnd;
    }

    /**
     * Create simple procedural crater terrain.
     * More complex terrain generation can be added later.
     */
    private createProceduralTerrain(): void {
        // Create a ground plane with craters (larger for more exploration)
        const terrainSize = 500;
        const subdivisions = 128; // Higher resolution for larger terrain

        this.terrain = MeshBuilder.CreateGround("moonTerrain", {
            width: terrainSize,
            height: terrainSize,
            subdivisions: subdivisions,
            updatable: true
        }, this.scene);

        // Apply crater deformations to terrain
        this.applyProcedurcraters(subdivisions);

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

        // Create invisible boundary walls to prevent falling off
        this.createBoundaryWalls(terrainSize);
    }

    /**
     * Create invisible walls around terrain perimeter.
     */
    private createBoundaryWalls(terrainSize: number): void {
        const wallHeight = 50;
        const wallThickness = 2;
        const halfSize = terrainSize / 2;

        // Create invisible wall material
        const invisibleMat = new StandardMaterial("invisibleWall", this.scene);
        invisibleMat.alpha = 0; // Fully transparent
        invisibleMat.disableLighting = true;

        // North wall (positive Z)
        const northWall = MeshBuilder.CreateBox("northWall", {
            width: terrainSize,
            height: wallHeight,
            depth: wallThickness
        }, this.scene);
        northWall.position = new Vector3(0, wallHeight / 2, halfSize);
        northWall.material = invisibleMat;
        northWall.checkCollisions = true;
        northWall.isVisible = false;

        // South wall (negative Z)
        const southWall = MeshBuilder.CreateBox("southWall", {
            width: terrainSize,
            height: wallHeight,
            depth: wallThickness
        }, this.scene);
        southWall.position = new Vector3(0, wallHeight / 2, -halfSize);
        southWall.material = invisibleMat;
        southWall.checkCollisions = true;
        southWall.isVisible = false;

        // East wall (positive X)
        const eastWall = MeshBuilder.CreateBox("eastWall", {
            width: wallThickness,
            height: wallHeight,
            depth: terrainSize
        }, this.scene);
        eastWall.position = new Vector3(halfSize, wallHeight / 2, 0);
        eastWall.material = invisibleMat;
        eastWall.checkCollisions = true;
        eastWall.isVisible = false;

        // West wall (negative X)
        const westWall = MeshBuilder.CreateBox("westWall", {
            width: wallThickness,
            height: wallHeight,
            depth: terrainSize
        }, this.scene);
        westWall.position = new Vector3(-halfSize, wallHeight / 2, 0);
        westWall.material = invisibleMat;
        westWall.checkCollisions = true;
        westWall.isVisible = false;

        console.log(`[MoonFPV] Created boundary walls for ${terrainSize}x${terrainSize} terrain`);
    }

    /**
     * Apply procedural crater deformations to terrain.
     */
    private applyProcedurcraters(subdivisions: number): void {
        if (!this.terrain) return;

        const positions = this.terrain.getVerticesData('position');
        if (!positions) return;

        // Create some random craters
        const craters = [];
        const craterCount = 15 + Math.floor(Math.random() * 10);

        for (let i = 0; i < craterCount; i++) {
            craters.push({
                x: (Math.random() - 0.5) * 180,
                z: (Math.random() - 0.5) * 180,
                radius: 5 + Math.random() * 20,
                depth: 0.5 + Math.random() * 2
            });
        }

        // Deform vertices based on craters
        for (let i = 0; i < positions.length; i += 3) {
            const x = positions[i] ?? 0;
            const z = positions[i + 2] ?? 0;
            let y = 0;

            // Apply each crater's depression
            for (const crater of craters) {
                const dx = x - crater.x;
                const dz = z - crater.z;
                const dist = Math.sqrt(dx * dx + dz * dz);

                if (dist < crater.radius) {
                    // Crater shape: deeper in center, rim at edge
                    const normalizedDist = dist / crater.radius;
                    const craterDepth = -crater.depth * Math.cos(normalizedDist * Math.PI * 0.5);
                    y += craterDepth;

                    // Add slight rim
                    if (normalizedDist > 0.7) {
                        y += crater.depth * 0.2 * (normalizedDist - 0.7) / 0.3;
                    }
                }
            }

            // Add some noise for roughness
            y += (Math.random() - 0.5) * 0.3;

            positions[i + 1] = y;
        }

        this.terrain.updateVerticesData('position', positions);
        this.terrain.createNormals(true);
    }

    /**
     * Create the parent planet visible in the sky.
     * Size and position based on orbital distance.
     */
    private createPlanetInSky(): void {
        // Calculate angular size of planet as seen from moon
        // angularSize = 2 * atan(planetRadius / distance)
        const planetRadiusKm = this.planetDiameterKm / 2;
        const distanceKm = this.moonDistanceKm; // Use km value for accurate calculation
        const angularSizeRad = 2 * Math.atan(planetRadiusKm / distanceKm);
        const angularSizeDeg = angularSizeRad * (180 / Math.PI);

        // Place planet at a fixed distance in scene units (beyond terrain edge)
        const skyDistance = 800; // Scene units - beyond 500 terrain radius
        // Scale planet size proportionally based on angular size
        const planetSceneRadius = skyDistance * Math.tan(angularSizeRad / 2);
        const planetSceneDiameter = Math.max(planetSceneRadius * 2, 20); // Minimum 20 units

        console.log(`[MoonFPV] Planet in sky: planetDiam=${this.planetDiameterKm}km, distance=${distanceKm}km, angularSize=${angularSizeDeg.toFixed(1)}°, sceneDiameter=${planetSceneDiameter.toFixed(1)}`);

        // Create planet sphere
        this.planetMesh = MeshBuilder.CreateSphere("moonFPVPlanet", {
            diameter: planetSceneDiameter,
            segments: 32
        }, this.scene);

        // Position planet in sky - tidally locked moons always face the planet
        // Place it at ~30 degrees above horizon
        this.planetMesh.position = new Vector3(0, skyDistance * 0.4, skyDistance);

        // Simple blue-green material for Earth-like planet
        const planetMat = new StandardMaterial("moonFPVPlanetMat", this.scene);
        planetMat.diffuseColor = new Color3(0.2, 0.4, 0.6);
        planetMat.emissiveColor = new Color3(0.05, 0.1, 0.15);
        planetMat.specularColor = new Color3(0.1, 0.1, 0.1);
        this.planetMesh.material = planetMat;
    }

    /**
     * Setup lighting for moon surface.
     */
    private setupLighting(): void {
        // Ambient light (very dim - space)
        this.ambientLight = new HemisphericLight("moonAmbient", new Vector3(0, 1, 0), this.scene);
        this.ambientLight.intensity = 0.1;
        this.ambientLight.groundColor = new Color3(0.02, 0.02, 0.02);

        // Sun light (distant, directional-like)
        this.sunLight = new PointLight("moonSun", new Vector3(1000, 500, 0), this.scene);
        this.sunLight.intensity = 1.5;
        this.sunLight.diffuse = new Color3(1.0, 0.98, 0.9); // Slightly warm
        this.sunLight.range = 5000;
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

        this.camera.detachControl();
        this.camera.dispose();
        this.terrain?.dispose();
        this.planetMesh?.dispose();
        this.sunLight?.dispose();
        this.ambientLight?.dispose();
        this.terrain = null;
        this.planetMesh = null;
    }
}
