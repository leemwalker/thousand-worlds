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

    // Spherical surface properties
    private sphereRadius: number = 100; // Radius of walkable sphere
    private gravityObserver: any = null; // Observer for radial gravity
    private cameraOrientationObserver: any = null; // Observer for camera orientation

    constructor(scene: Scene, moon: MoonData, options: MoonFPVOptions = {}) {
        this.scene = scene;
        this.moon = moon;
        this.moonParams = getMoonFPVParams(moon);
        this.eyeHeight = options.eyeHeight ?? 1.7;
        this.planetDiameterKm = options.planetDiameterKm ?? 12742;
        // Use passed km distance, or convert from meters if not provided
        // moon.distance is in meters, divide by 1e6 to get km
        this.moonDistanceKm = options.moonDistanceKm ?? (moon.distance / 1e6);

        // Start position INSIDE the sphere near inner surface
        // Player walks on inside of sphere, so position is radius - eyeHeight from center
        const startPos = new Vector3(0, this.sphereRadius - this.eyeHeight, 0);

        // Create FPS camera
        this.camera = new UniversalCamera("moonFPSCamera", startPos, scene);
        // Look toward center initially (that's "up" in this inside-out world)
        this.camera.setTarget(new Vector3(0, 0, 0));
        this.camera.ellipsoid = new Vector3(0.5, 1.0, 0.5);
        this.camera.ellipsoidOffset = new Vector3(0, 1.0, 0);
        this.camera.checkCollisions = true;
        this.camera.applyGravity = false; // We'll apply custom radial gravity

        // Set camera speed based on moon gravity
        this.camera.speed = this.moonParams.moveSpeed;

        // Movement keys
        this.camera.keysUp = [87]; // W
        this.camera.keysDown = [83]; // S
        this.camera.keysLeft = [65]; // A
        this.camera.keysRight = [68]; // D
        this.camera.angularSensibility = 500;

        // Disable fog for spherical view (we want to see the sky)
        this.scene.fogMode = 0;

        // Create the spherical moon surface (inside-out)
        this.createSphericalTerrain();

        // Create starfield (large sphere around everything)
        this.createStarfield();

        // Create sun in the sky
        this.createSun();

        // Create planet in the sky
        this.createPlanetInSky();

        // Setup lighting (sun direction + ambient)
        this.setupLighting();

        // Setup radial gravity and camera orientation
        this.setupSphericalPhysics();

        console.log(`[MoonFPV] Created spherical surface for ${moon.name}: radius=${this.sphereRadius}, gravity=${this.moonParams.gravity.toFixed(3)} m/s²`);
    }

    /**
     * Create the spherical moon surface using an inside-out sphere.
     * Player walks on the inside surface, looking up at the "sky" (center of sphere).
     */
    private createSphericalTerrain(): void {
        // Create sphere with inverted normals (player is inside)
        this.terrain = MeshBuilder.CreateSphere("moonSurface", {
            diameter: this.sphereRadius * 2,
            segments: 64,
            sideOrientation: 1 // BACKSIDE - render inside
        }, this.scene);

        // Apply crater-like height variations via vertex manipulation
        this.applySphericalCraters();

        // Material based on moon color
        const terrainMat = new StandardMaterial("moonTerrainMat", this.scene);
        const moonColorHex = this.moon.color || "#888888";
        const r = parseInt(moonColorHex.slice(1, 3), 16) / 255;
        const g = parseInt(moonColorHex.slice(3, 5), 16) / 255;
        const b = parseInt(moonColorHex.slice(5, 7), 16) / 255;

        terrainMat.diffuseColor = new Color3(r * 0.8, g * 0.8, b * 0.8);
        terrainMat.specularColor = new Color3(0.1, 0.1, 0.1);
        terrainMat.backFaceCulling = false; // Show inside surface
        this.terrain.material = terrainMat;
        this.terrain.checkCollisions = true;

        console.log(`[MoonFPV] Created spherical terrain, radius=${this.sphereRadius}`);
    }

    /**
     * Apply crater-like deformations to spherical terrain.
     */
    private applySphericalCraters(): void {
        if (!this.terrain) return;

        const positions = this.terrain.getVerticesData('position');
        if (!positions) return;

        // Create random craters
        const craters: Array<{ center: Vector3, radius: number, depth: number }> = [];
        const craterCount = 20 + Math.floor(Math.random() * 15);

        for (let i = 0; i < craterCount; i++) {
            // Random point on unit sphere, then scale
            const theta = Math.random() * Math.PI * 2;
            const phi = Math.acos(2 * Math.random() - 1);
            const center = new Vector3(
                Math.sin(phi) * Math.cos(theta),
                Math.sin(phi) * Math.sin(theta),
                Math.cos(phi)
            ).scale(this.sphereRadius);

            craters.push({
                center,
                radius: 5 + Math.random() * 15, // Crater radius
                depth: 0.5 + Math.random() * 2  // Crater depth
            });
        }

        // Apply crater deformations to vertices
        for (let i = 0; i < positions.length; i += 3) {
            const vertex = new Vector3(
                positions[i] ?? 0,
                positions[i + 1] ?? 0,
                positions[i + 2] ?? 0
            );

            let depression = 0;
            for (const crater of craters) {
                const dist = Vector3.Distance(vertex, crater.center);
                if (dist < crater.radius) {
                    // Parabolic crater profile
                    const t = dist / crater.radius;
                    depression += crater.depth * (1 - t * t);
                }
            }

            // Push vertex inward for craters (since normals face inward)
            if (depression > 0) {
                const normal = vertex.normalize();
                positions[i] = (positions[i] ?? 0) - normal.x * depression;
                positions[i + 1] = (positions[i + 1] ?? 0) - normal.y * depression;
                positions[i + 2] = (positions[i + 2] ?? 0) - normal.z * depression;
            }
        }

        this.terrain.updateVerticesData('position', positions);
    }

    /**
     * Create starfield sphere surrounding the moon surface.
     */
    private createStarfield(): void {
        // Create large sphere for starfield
        const starfieldRadius = this.sphereRadius * 5;
        this.starfieldMesh = MeshBuilder.CreateSphere("fpvStarfield", {
            diameter: starfieldRadius * 2,
            segments: 32,
            sideOrientation: 1 // BACKSIDE - we're inside
        }, this.scene);

        // Simple dark material with emissive stars
        const starMat = new StandardMaterial("starfieldMat", this.scene);
        starMat.diffuseColor = new Color3(0, 0, 0);
        starMat.emissiveColor = new Color3(0.02, 0.02, 0.05);
        starMat.specularColor = new Color3(0, 0, 0);
        starMat.backFaceCulling = false;
        this.starfieldMesh.material = starMat;

        // Starfield doesn't need collisions
        this.starfieldMesh.checkCollisions = false;

        console.log(`[MoonFPV] Created starfield, radius=${starfieldRadius}`);
    }

    /**
     * Create sun visual in the sky.
     */
    private createSun(): void {
        // Create sun sphere at distance
        const sunDistance = this.sphereRadius * 4;
        const sunSize = this.sphereRadius * 0.3;

        this.sunMesh = MeshBuilder.CreateSphere("fpvSun", {
            diameter: sunSize,
            segments: 16
        }, this.scene);

        // Position sun (fixed position for now)
        this.sunMesh.position = new Vector3(sunDistance, sunDistance * 0.5, 0);

        // Bright emissive material
        const sunMat = new StandardMaterial("sunMat", this.scene);
        sunMat.diffuseColor = new Color3(1, 1, 0.8);
        sunMat.emissiveColor = new Color3(1, 0.9, 0.7);
        sunMat.specularColor = new Color3(0, 0, 0);
        this.sunMesh.material = sunMat;

        console.log(`[MoonFPV] Created sun at distance ${sunDistance}`);
    }

    /**
     * Setup radial gravity and camera orientation for spherical surface.
     */
    private setupSphericalPhysics(): void {
        const gravityStrength = this.moonParams.gravity / 50; // Scaled for scene

        this.gravityObserver = this.scene.onBeforeRenderObservable.add(() => {
            if (this.disposed || !this.camera) return;

            const pos = this.camera.position;
            const distFromCenter = pos.length();

            // Radial gravity: pull OUTWARD (toward inner surface of sphere)
            const outwardDir = distFromCenter > 0.01 ? pos.normalize() : new Vector3(0, 1, 0);

            // Apply gravity to velocity
            const deltaTime = this.scene.getEngine().getDeltaTime() / 1000;
            this.velocity.addInPlace(outwardDir.scale(gravityStrength * deltaTime * 60));

            // Apply velocity to position
            const newPos = pos.add(this.velocity.scale(deltaTime));

            // Constrain to inner surface (player at sphereRadius - eyeHeight)
            const targetDist = this.sphereRadius - this.eyeHeight;
            const currentDist = newPos.length();

            if (currentDist > targetDist) {
                // Beyond inner surface - snap to surface
                newPos.normalize().scaleInPlace(targetDist);
                this.velocity = Vector3.Zero(); // Reset velocity when grounded
                this.isGrounded = true;
            } else {
                this.isGrounded = false;
            }

            this.camera.position = newPos;

            // Orient camera: "up" points TOWARD center (that's the sky)
            const upVector = distFromCenter > 0.01 ? pos.normalize().scale(-1) : new Vector3(0, -1, 0);
            this.camera.upVector = upVector;
        });

        console.log(`[MoonFPV] Inside-out physics enabled, gravity=${gravityStrength.toFixed(4)}`);
    }

    /**
     * Create the parent planet visible in the sky.
     * Size and position based on orbital distance.
     */
    private createPlanetInSky(): void {
        // Calculate angular size of planet as seen from moon
        // angularSize = 2 * atan(planetRadius / distance)
        const planetRadiusKm = this.planetDiameterKm / 2;
        const distanceKm = this.moonDistanceKm;
        const angularSizeRad = 2 * Math.atan(planetRadiusKm / distanceKm);
        const angularSizeDeg = angularSizeRad * (180 / Math.PI);

        // Place planet between moon sphere and starfield
        const skyDistance = this.sphereRadius * 3; // Between terrain and starfield
        // Scale planet size proportionally based on angular size
        const planetSceneRadius = skyDistance * Math.tan(angularSizeRad / 2);
        const planetSceneDiameter = Math.max(planetSceneRadius * 2, this.sphereRadius * 0.5);

        console.log(`[MoonFPV] Planet in sky: angularSize=${angularSizeDeg.toFixed(1)}°, sceneDiam=${planetSceneDiameter.toFixed(1)}`);

        // Create planet sphere
        this.planetMesh = MeshBuilder.CreateSphere("moonFPVPlanet", {
            diameter: planetSceneDiameter,
            segments: 32
        }, this.scene);

        // Position planet in sky - tidally locked moons always face the planet
        // For inside-out sphere, "up" from center is toward the sky
        this.planetMesh.position = new Vector3(skyDistance * 0.8, skyDistance * 0.6, 0);

        // Earth-like planet material with emissive glow
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
        const sunDistance = this.sphereRadius * 4;
        this.sunLight = new PointLight("moonSun", new Vector3(sunDistance, sunDistance * 0.5, 0), this.scene);
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

        // Remove spherical physics observer
        if (this.gravityObserver) {
            this.scene.onBeforeRenderObservable.remove(this.gravityObserver);
            this.gravityObserver = null;
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
