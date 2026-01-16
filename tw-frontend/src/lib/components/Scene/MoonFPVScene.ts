/**
 * MoonFPVScene.ts
 * Dedicated scene for first-person view on moon surfaces.
 * Creates an isolated environment with procedural terrain and the parent planet visible in the sky.
 */
import type { Scene } from "@babylonjs/core/scene";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import type { Mesh } from "@babylonjs/core/Meshes/mesh";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Color3, Color4 } from "@babylonjs/core/Maths/math.color";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
import { HemisphericLight } from "@babylonjs/core/Lights/hemisphericLight";
import { PointLight } from "@babylonjs/core/Lights/pointLight";
import { UniversalCamera } from "@babylonjs/core/Cameras/universalCamera";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import type { MoonData } from "$lib/types/moon";
import { getMoonFPVParams, type MoonFPVParams } from "$lib/types/moon";

// Side-effect imports for mesh building and collisions
import "@babylonjs/core/Meshes/meshBuilder";
import "@babylonjs/core/Collisions/collisionCoordinator";
import "@babylonjs/core/Cameras/Inputs/freeCameraKeyboardMoveInput";
import "@babylonjs/core/Cameras/Inputs/freeCameraMouseInput";

export interface MoonFPVSceneData {
    moon: MoonData;
    planetDiameter: number; // km
    moonOrbitalDistance: number; // km
    onExit?: () => void;
}

export interface MoonFPVSceneCallbacks {
    onExit?: (() => void) | undefined;
}

/**
 * Creates and manages the Moon FPV scene.
 * Follows the SceneFactory pattern used by LobbyScene.
 */
export class MoonFPVScene {
    private scene: Scene | null = null;
    private camera: UniversalCamera | null = null;
    private terrain: Mesh | null = null;
    private planetMesh: Mesh | null = null;
    private sunMesh: Mesh | null = null;
    private starfield: Mesh | null = null;
    private callbacks: MoonFPVSceneCallbacks = {};
    private moonParams: MoonFPVParams | null = null;
    private disposed: boolean = false;

    // Scene data passed from WorldController
    private sceneData: MoonFPVSceneData | null = null;

    // Terrain parameters
    private readonly TERRAIN_SIZE = 200;
    private readonly TERRAIN_SUBDIVISIONS = 64;
    private readonly EYE_HEIGHT = 1.7;

    /**
     * Set scene data before creation.
     * Must be called before create() to provide moon context.
     */
    setSceneData(data: MoonFPVSceneData): void {
        this.sceneData = data;
        if (data.onExit) {
            this.callbacks.onExit = data.onExit;
        }
    }

    /**
     * SceneFactory implementation - create the moon FPV scene.
     */
    async create(scene: Scene): Promise<void> {
        if (!this.sceneData) {
            throw new Error("[MoonFPVScene] Scene data not set. Call setSceneData() before create().");
        }

        this.scene = scene;
        this.moonParams = getMoonFPVParams(this.sceneData.moon);

        // Dark space background
        scene.clearColor = new Color4(0.01, 0.01, 0.02, 1);

        // Set up fog based on moon horizon
        this.setupHorizonFog(scene);

        // Create scene elements
        this.createStarfield(scene);
        this.createTerrain(scene);
        this.createPlanetInSky(scene);
        this.createSun(scene);
        this.setupLighting(scene);
        this.createCamera(scene);

        // Register keyboard handler for ESC
        scene.onKeyboardObservable.add((kbInfo) => {
            if (kbInfo.type === 1 && kbInfo.event.key === "Escape") {
                console.log("[MoonFPVScene] ESC pressed, exiting");
                this.callbacks.onExit?.();
            }
        });

        console.log(`[MoonFPVScene] Created for ${this.sceneData.moon.name}`);
    }

    /**
     * Set callbacks for scene events.
     */
    setCallbacks(callbacks: MoonFPVSceneCallbacks): void {
        this.callbacks = { ...this.callbacks, ...callbacks };
    }

    /**
     * Set up fog based on moon horizon distance.
     */
    private setupHorizonFog(scene: Scene): void {
        if (!this.moonParams) return;

        // Scale horizon for rendering (we don't render full scale)
        const fogStart = Math.min(this.moonParams.horizonDistance / 1000, 500);
        const fogEnd = fogStart * 2;

        scene.fogMode = 2; // Exponential
        scene.fogDensity = 0.001;
        scene.fogColor = new Color3(0.02, 0.02, 0.03);
        scene.fogStart = fogStart;
        scene.fogEnd = fogEnd;
    }

    /**
     * Create starfield background sphere.
     */
    private createStarfield(scene: Scene): void {
        this.starfield = MeshBuilder.CreateSphere("moonFPVStarfield", {
            diameter: 2000,
            segments: 16
        }, scene);

        const starfieldMat = new StandardMaterial("moonFPVStarfieldMat", scene);
        starfieldMat.emissiveColor = new Color3(0.02, 0.02, 0.03);
        starfieldMat.disableLighting = true;
        starfieldMat.backFaceCulling = false;
        this.starfield.material = starfieldMat;

        // Add some stars as small emissive spots
        // For now, use a simple dark material - could add procedural stars later
    }

    /**
     * Create procedural moon terrain with craters.
     */
    private createTerrain(scene: Scene): void {
        if (!this.sceneData) return;

        this.terrain = MeshBuilder.CreateGround("moonFPVTerrain", {
            width: this.TERRAIN_SIZE,
            height: this.TERRAIN_SIZE,
            subdivisions: this.TERRAIN_SUBDIVISIONS,
            updatable: true
        }, scene);

        // Apply crater deformations
        this.applyCraterDeformations();

        // Material based on moon color
        const terrainMat = new StandardMaterial("moonFPVTerrainMat", scene);
        const moonColorHex = this.sceneData.moon.color || "#888888";
        const r = parseInt(moonColorHex.slice(1, 3), 16) / 255;
        const g = parseInt(moonColorHex.slice(3, 5), 16) / 255;
        const b = parseInt(moonColorHex.slice(5, 7), 16) / 255;

        terrainMat.diffuseColor = new Color3(r * 0.8, g * 0.8, b * 0.8);
        terrainMat.specularColor = new Color3(0.1, 0.1, 0.1);
        this.terrain.material = terrainMat;
        this.terrain.checkCollisions = true;
    }

    /**
     * Apply procedural crater deformations to terrain.
     */
    private applyCraterDeformations(): void {
        if (!this.terrain) return;

        const positions = this.terrain.getVerticesData("position");
        if (!positions) return;

        // Create random craters
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

        // Deform vertices
        for (let i = 0; i < positions.length; i += 3) {
            const x = positions[i] ?? 0;
            const z = positions[i + 2] ?? 0;
            let y = 0;

            for (const crater of craters) {
                const dx = x - crater.x;
                const dz = z - crater.z;
                const dist = Math.sqrt(dx * dx + dz * dz);

                if (dist < crater.radius) {
                    const normalizedDist = dist / crater.radius;
                    const craterDepth = -crater.depth * Math.cos(normalizedDist * Math.PI * 0.5);
                    y += craterDepth;

                    // Add rim
                    if (normalizedDist > 0.7) {
                        y += crater.depth * 0.2 * (normalizedDist - 0.7) / 0.3;
                    }
                }
            }

            // Add noise for roughness
            y += (Math.random() - 0.5) * 0.3;
            positions[i + 1] = y;
        }

        this.terrain.updateVerticesData("position", positions);
        this.terrain.createNormals(true);
    }

    /**
     * Create the parent planet visible in the sky.
     * Size and position based on orbital distance.
     */
    private createPlanetInSky(scene: Scene): void {
        if (!this.sceneData) return;

        // Calculate angular size of planet as seen from moon
        // angularSize = 2 * atan(planetRadius / distance)
        const planetRadiusKm = this.sceneData.planetDiameter / 2;
        const distanceKm = this.sceneData.moonOrbitalDistance;
        const angularSizeRad = 2 * Math.atan(planetRadiusKm / distanceKm);
        const angularSizeDeg = angularSizeRad * (180 / Math.PI);

        // Place planet at a fixed distance in scene units
        const skyDistance = 500; // Scene units
        // Scale planet size proportionally
        const planetSceneRadius = skyDistance * Math.tan(angularSizeRad / 2);
        const planetSceneDiameter = Math.max(planetSceneRadius * 2, 10); // Minimum 10 units

        console.log(`[MoonFPVScene] Planet in sky: angularSize=${angularSizeDeg.toFixed(1)}°, sceneDiameter=${planetSceneDiameter.toFixed(1)}`);

        // Create planet sphere
        this.planetMesh = MeshBuilder.CreateSphere("skyPlanet", {
            diameter: planetSceneDiameter,
            segments: 32
        }, scene);

        // Position planet in sky - for now, put it at a fixed elevation
        // The moon is tidally locked, so planet position is relatively fixed
        this.planetMesh.position = new Vector3(0, skyDistance * 0.4, skyDistance);

        // Simple blue-green material for now (could use actual planet texture later)
        const planetMat = new StandardMaterial("skyPlanetMat", scene);
        planetMat.diffuseColor = new Color3(0.2, 0.4, 0.6);
        planetMat.emissiveColor = new Color3(0.05, 0.1, 0.15);
        planetMat.specularColor = new Color3(0.1, 0.1, 0.1);
        this.planetMesh.material = planetMat;
    }

    /**
     * Create distant sun for reference.
     */
    private createSun(scene: Scene): void {
        this.sunMesh = MeshBuilder.CreateSphere("skySun", {
            diameter: 15,
            segments: 16
        }, scene);

        // Position sun opposite to planet (simplified)
        this.sunMesh.position = new Vector3(-400, 200, -400);

        const sunMat = new StandardMaterial("skySunMat", scene);
        sunMat.emissiveColor = new Color3(1, 0.95, 0.8);
        sunMat.disableLighting = true;
        this.sunMesh.material = sunMat;
    }

    /**
     * Set up scene lighting.
     */
    private setupLighting(scene: Scene): void {
        // Ambient light (very dim for moon surface)
        const ambient = new HemisphericLight("moonFPVAmbient", new Vector3(0, 1, 0), scene);
        ambient.intensity = 0.2;
        ambient.groundColor = new Color3(0.05, 0.05, 0.05);

        // Sun directional light
        const sunLight = new PointLight("moonFPVSunLight", new Vector3(-400, 200, -400), scene);
        sunLight.intensity = 1.5;
        sunLight.diffuse = new Color3(1, 0.98, 0.9);
        sunLight.range = 2000;
    }

    /**
     * Create FPS camera with moon gravity.
     */
    private createCamera(scene: Scene): void {
        if (!this.moonParams) return;

        const startPos = new Vector3(0, this.EYE_HEIGHT + 1, 0);

        this.camera = new UniversalCamera("moonFPVCamera", startPos, scene);
        this.camera.setTarget(startPos.add(new Vector3(0, 0, 1)));
        this.camera.ellipsoid = new Vector3(0.5, 0.9, 0.5);
        this.camera.checkCollisions = true;
        this.camera.applyGravity = true;

        // Set speed based on moon gravity
        this.camera.speed = this.moonParams.moveSpeed;

        // Movement keys
        this.camera.keysUp = [87]; // W
        this.camera.keysDown = [83]; // S
        this.camera.keysLeft = [65]; // A
        this.camera.keysRight = [68]; // D
        this.camera.angularSensibility = 500;

        // Attach controls
        this.camera.attachControl(scene.getEngine().getRenderingCanvas(), true);

        // Make this the active camera
        scene.activeCamera = this.camera;
    }

    /**
     * Get the active camera.
     */
    getCamera(): UniversalCamera | null {
        return this.camera;
    }

    /**
     * Clean up all resources.
     */
    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;

        console.log("[MoonFPVScene] Disposing");

        this.camera?.detachControl();
        this.camera?.dispose();
        this.terrain?.dispose();
        this.planetMesh?.dispose();
        this.sunMesh?.dispose();
        this.starfield?.dispose();

        // Dispose all meshes and lights in scene
        if (this.scene) {
            const allMeshes = [...this.scene.meshes];
            console.log(`[MoonFPVScene] Disposing ${allMeshes.length} meshes`);
            allMeshes.forEach(m => m.dispose());

            const allLights = [...this.scene.lights];
            console.log(`[MoonFPVScene] Disposing ${allLights.length} lights`);
            allLights.forEach(l => l.dispose());
        }

        this.scene = null;
        this.camera = null;
        this.terrain = null;
        this.planetMesh = null;
        this.sunMesh = null;
        this.starfield = null;
        this.sceneData = null;
    }
}
