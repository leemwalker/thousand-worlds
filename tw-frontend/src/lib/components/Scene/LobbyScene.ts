/**
 * LobbyScene.ts
 * First-person lobby environment - the Grand Lobby with statue and portal.
 */
import type { Scene } from "@babylonjs/core/scene";
import type { Engine } from "@babylonjs/core/Engines/engine";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import type { Mesh } from "@babylonjs/core/Meshes/mesh";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Color3, Color4 } from "@babylonjs/core/Maths/math.color";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
import { HemisphericLight } from "@babylonjs/core/Lights/hemisphericLight";
import { PointLight } from "@babylonjs/core/Lights/pointLight";
import { CubeTexture } from "@babylonjs/core/Materials/Textures/cubeTexture";
import { ParticleSystem } from "@babylonjs/core/Particles/particleSystem";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { Scene as BabylonScene } from "@babylonjs/core/scene";
import { FirstPersonController } from "../Map/FirstPersonController";

// Import side effects for mesh building
import "@babylonjs/core/Meshes/meshBuilder";

export interface LobbySceneCallbacks {
    onPortalEnter?: () => void;
    onEastPortalEnter?: () => void; // Tropical test world portal
    onStatueInteract?: () => void;
}

/**
 * Creates and manages the Grand Lobby scene.
 */
export class LobbyScene {
    private scene: Scene | null = null;
    private fpsController: FirstPersonController | null = null;
    private floor: Mesh | null = null;
    private statue: Mesh | null = null;
    private portal: Mesh | null = null;
    private portalParticles: ParticleSystem | null = null;
    private eastPortal: Mesh | null = null; // Tropical test world portal
    private eastPortalParticles: ParticleSystem | null = null;
    private callbacks: LobbySceneCallbacks = {};
    private portalEntered: boolean = false;
    private eastPortalEntered: boolean = false;

    // Room dimensions
    private readonly ROOM_WIDTH = 30;
    private readonly ROOM_DEPTH = 40;
    private readonly ROOM_HEIGHT = 12;
    private readonly PORTAL_RADIUS = 2;

    /**
     * Create the lobby scene.
     */
    // SceneFactory implementation
    async create(scene: Scene): Promise<void> {
        this.scene = scene;

        // Clear color for a slightly warm ambient
        scene.clearColor = new Color4(0.02, 0.02, 0.03, 1);

        // Ambient lighting
        const ambient = new HemisphericLight("ambient", new Vector3(0, 1, 0), scene);
        ambient.intensity = 0.4;
        ambient.groundColor = new Color3(0.1, 0.08, 0.06);

        this.createFloor(scene);
        this.createWalls(scene);
        this.createStatue(scene);
        this.createPortal(scene);
        this.createEastPortal(scene); // Tropical test world portal
        this.createFPSController(scene);

        // Register update loop
        scene.onBeforeRenderObservable.add(() => {
            const deltaTime = scene.getEngine().getDeltaTime() / 1000;
            this.update(deltaTime);
        });

        console.log("[LobbyScene] Created");
    }

    /**
     * Set callbacks for interaction.
     */
    setCallbacks(callbacks: LobbySceneCallbacks): void {
        this.callbacks = callbacks;
    }

    /**
     * Create marble floor.
     */
    private createFloor(scene: Scene): void {
        this.floor = MeshBuilder.CreateBox("floor", {
            width: this.ROOM_WIDTH,
            depth: this.ROOM_DEPTH,
            height: 0.5
        }, scene);
        this.floor.position.y = -0.25;
        this.floor.checkCollisions = true;

        const floorMat = new StandardMaterial("floorMat", scene);
        floorMat.diffuseColor = new Color3(0.9, 0.85, 0.8); // Warm marble
        floorMat.specularColor = new Color3(0.3, 0.3, 0.3);
        floorMat.specularPower = 32;
        this.floor.material = floorMat;
    }

    /**
     * Create walls using simple boxes (ceiling-less for now).
     */
    private createWalls(scene: Scene): void {
        const wallMat = new StandardMaterial("wallMat", scene);
        wallMat.diffuseColor = new Color3(0.85, 0.8, 0.75);
        wallMat.specularColor = new Color3(0.2, 0.2, 0.2);

        const wallThickness = 0.5;
        const wallHeight = this.ROOM_HEIGHT;

        // North wall
        const northWall = MeshBuilder.CreateBox("northWall", {
            width: this.ROOM_WIDTH,
            height: wallHeight,
            depth: wallThickness
        }, scene);
        northWall.position = new Vector3(0, wallHeight / 2, this.ROOM_DEPTH / 2);
        northWall.material = wallMat;
        northWall.checkCollisions = true;

        // South wall
        const southWall = MeshBuilder.CreateBox("southWall", {
            width: this.ROOM_WIDTH,
            height: wallHeight,
            depth: wallThickness
        }, scene);
        southWall.position = new Vector3(0, wallHeight / 2, -this.ROOM_DEPTH / 2);
        southWall.material = wallMat;
        southWall.checkCollisions = true;

        // East wall (with portal hole for tropical world) - split into two parts
        const eastWallLeft = MeshBuilder.CreateBox("eastWallLeft", {
            width: wallThickness,
            height: wallHeight,
            depth: this.ROOM_DEPTH / 2 - this.PORTAL_RADIUS - 1
        }, scene);
        eastWallLeft.position = new Vector3(
            this.ROOM_WIDTH / 2,
            wallHeight / 2,
            this.ROOM_DEPTH / 4 + this.PORTAL_RADIUS / 2
        );
        eastWallLeft.material = wallMat;
        eastWallLeft.checkCollisions = true;

        const eastWallRight = MeshBuilder.CreateBox("eastWallRight", {
            width: wallThickness,
            height: wallHeight,
            depth: this.ROOM_DEPTH / 2 - this.PORTAL_RADIUS - 1
        }, scene);
        eastWallRight.position = new Vector3(
            this.ROOM_WIDTH / 2,
            wallHeight / 2,
            -this.ROOM_DEPTH / 4 - this.PORTAL_RADIUS / 2
        );
        eastWallRight.material = wallMat;
        eastWallRight.checkCollisions = true;

        // West wall (with portal hole) - split into two parts
        const westWallLeft = MeshBuilder.CreateBox("westWallLeft", {
            width: wallThickness,
            height: wallHeight,
            depth: this.ROOM_DEPTH / 2 - this.PORTAL_RADIUS - 1
        }, scene);
        westWallLeft.position = new Vector3(
            -this.ROOM_WIDTH / 2,
            wallHeight / 2,
            this.ROOM_DEPTH / 4 + this.PORTAL_RADIUS / 2
        );
        westWallLeft.material = wallMat;
        westWallLeft.checkCollisions = true;

        const westWallRight = MeshBuilder.CreateBox("westWallRight", {
            width: wallThickness,
            height: wallHeight,
            depth: this.ROOM_DEPTH / 2 - this.PORTAL_RADIUS - 1
        }, scene);
        westWallRight.position = new Vector3(
            -this.ROOM_WIDTH / 2,
            wallHeight / 2,
            -this.ROOM_DEPTH / 4 - this.PORTAL_RADIUS / 2
        );
        westWallRight.material = wallMat;
        westWallRight.checkCollisions = true;
    }

    /**
     * Create the central statue (cylinder + capsule for now).
     */
    private createStatue(scene: Scene): void {
        // Base pedestal
        const pedestal = MeshBuilder.CreateCylinder("pedestal", {
            height: 1.5,
            diameter: 2
        }, scene);
        pedestal.position = new Vector3(0, 0.75, 0);

        const pedestalMat = new StandardMaterial("pedestalMat", scene);
        pedestalMat.diffuseColor = new Color3(0.6, 0.6, 0.65);
        pedestal.material = pedestalMat;
        pedestal.checkCollisions = true;

        // Statue body (simplified humanoid shape)
        this.statue = MeshBuilder.CreateCapsule("statue", {
            height: 4,
            radius: 0.5
        }, scene);
        this.statue.position = new Vector3(0, 3.5, 0);

        const statueMat = new StandardMaterial("statueMat", scene);
        statueMat.diffuseColor = new Color3(0.9, 0.9, 0.95);
        statueMat.emissiveColor = new Color3(0.02, 0.02, 0.05); // Subtle glow
        statueMat.specularColor = new Color3(0.4, 0.4, 0.4);
        statueMat.specularPower = 64;
        this.statue.material = statueMat;

        // Light the statue
        const statueLight = new PointLight("statueLight", new Vector3(0, 6, 0), scene);
        statueLight.intensity = 0.8;
        statueLight.diffuse = new Color3(0.9, 0.85, 0.7);
        statueLight.range = 15;
    }

    /**
     * Create the western portal with particle effect.
     */
    private createPortal(scene: Scene): void {
        // Portal frame (torus)
        this.portal = MeshBuilder.CreateTorus("portal", {
            diameter: this.PORTAL_RADIUS * 2,
            thickness: 0.3,
            tessellation: 32
        }, scene);
        this.portal.position = new Vector3(-this.ROOM_WIDTH / 2 + 0.5, this.PORTAL_RADIUS + 0.5, 0);
        this.portal.rotation.y = Math.PI / 2;

        const portalMat = new StandardMaterial("portalMat", scene);
        portalMat.diffuseColor = new Color3(0.3, 0.2, 0.5);
        portalMat.emissiveColor = new Color3(0.2, 0.1, 0.4);
        this.portal.material = portalMat;

        // Portal center (plane for particles)
        const portalCenter = MeshBuilder.CreateDisc("portalCenter", {
            radius: this.PORTAL_RADIUS - 0.3,
            tessellation: 32
        }, scene);
        portalCenter.position = this.portal.position.clone();
        portalCenter.rotation.y = Math.PI / 2;

        const centerMat = new StandardMaterial("centerMat", scene);
        centerMat.diffuseColor = new Color3(0.1, 0.05, 0.2);
        centerMat.emissiveColor = new Color3(0.15, 0.1, 0.3);
        centerMat.alpha = 0.7;
        portalCenter.material = centerMat;

        // Portal light
        const portalLight = new PointLight("portalLight", this.portal.position.clone(), scene);
        portalLight.intensity = 1.2;
        portalLight.diffuse = new Color3(0.5, 0.3, 0.8);
        portalLight.range = 10;

        // Create particle system for portal effect
        this.createPortalParticles(scene, this.portal.position);
    }

    /**
     * Create swirling particle effect for portal.
     */
    private createPortalParticles(scene: Scene, position: Vector3): void {
        this.portalParticles = new ParticleSystem("portalParticles", 500, scene);

        // Create a procedural texture for particles using canvas
        // This avoids the base64 PNG which was causing WebGL errors
        const size = 32;
        const canvas = document.createElement("canvas");
        canvas.width = size;
        canvas.height = size;
        const ctx = canvas.getContext("2d");
        if (ctx) {
            // Create a radial gradient for soft particle
            const gradient = ctx.createRadialGradient(
                size / 2, size / 2, 0,
                size / 2, size / 2, size / 2
            );
            gradient.addColorStop(0, "rgba(255, 255, 255, 1)");
            gradient.addColorStop(0.3, "rgba(200, 200, 255, 0.8)");
            gradient.addColorStop(0.7, "rgba(100, 50, 200, 0.3)");
            gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

            ctx.fillStyle = gradient;
            ctx.fillRect(0, 0, size, size);
        }

        // Create texture from canvas with noMipmap=true to avoid WebGL errors
        const particleTexture = new Texture(
            canvas.toDataURL(),
            scene,
            true,  // noMipmap - prevents glGenerateMipmap error
            false  // invertY
        );
        this.portalParticles.particleTexture = particleTexture;

        this.portalParticles.emitter = position;
        this.portalParticles.minEmitBox = new Vector3(-0.5, -1.5, -0.5);
        this.portalParticles.maxEmitBox = new Vector3(0.5, 1.5, 0.5);

        this.portalParticles.color1 = new Color4(0.5, 0.2, 0.8, 1);
        this.portalParticles.color2 = new Color4(0.2, 0.1, 0.5, 1);
        this.portalParticles.colorDead = new Color4(0, 0, 0.2, 0);

        this.portalParticles.minSize = 0.05;
        this.portalParticles.maxSize = 0.15;

        this.portalParticles.minLifeTime = 0.5;
        this.portalParticles.maxLifeTime = 1.5;

        this.portalParticles.emitRate = 100;

        this.portalParticles.gravity = new Vector3(0, 0.5, 0);

        this.portalParticles.direction1 = new Vector3(-0.5, 1, -0.5);
        this.portalParticles.direction2 = new Vector3(0.5, 1, 0.5);

        this.portalParticles.minAngularSpeed = 0;
        this.portalParticles.maxAngularSpeed = Math.PI;

        this.portalParticles.minEmitPower = 0.5;
        this.portalParticles.maxEmitPower = 1.5;
        this.portalParticles.updateSpeed = 0.01;

        this.portalParticles.start();
    }

    /**
     * Create the eastern portal with tropical green/gold particle effect.
     * This portal leads to the tropical test world.
     */
    private createEastPortal(scene: Scene): void {
        // Portal frame (torus) - tropical colors
        this.eastPortal = MeshBuilder.CreateTorus("eastPortal", {
            diameter: this.PORTAL_RADIUS * 2,
            thickness: 0.3,
            tessellation: 32
        }, scene);
        // Position on east wall (opposite of west portal)
        this.eastPortal.position = new Vector3(this.ROOM_WIDTH / 2 - 0.5, this.PORTAL_RADIUS + 0.5, 0);
        this.eastPortal.rotation.y = -Math.PI / 2; // Face inward

        const eastPortalMat = new StandardMaterial("eastPortalMat", scene);
        eastPortalMat.diffuseColor = new Color3(0.2, 0.5, 0.3); // Green-gold
        eastPortalMat.emissiveColor = new Color3(0.1, 0.4, 0.2);
        this.eastPortal.material = eastPortalMat;

        // Portal center
        const eastPortalCenter = MeshBuilder.CreateDisc("eastPortalCenter", {
            radius: this.PORTAL_RADIUS - 0.3,
            tessellation: 32
        }, scene);
        eastPortalCenter.position = this.eastPortal.position.clone();
        eastPortalCenter.rotation.y = -Math.PI / 2;

        const eastCenterMat = new StandardMaterial("eastCenterMat", scene);
        eastCenterMat.diffuseColor = new Color3(0.3, 0.2, 0.1); // Warm amber
        eastCenterMat.emissiveColor = new Color3(0.4, 0.3, 0.1);
        eastCenterMat.alpha = 0.7;
        eastPortalCenter.material = eastCenterMat;

        // Portal light - warm tropical
        const eastPortalLight = new PointLight("eastPortalLight", this.eastPortal.position.clone(), scene);
        eastPortalLight.intensity = 1.2;
        eastPortalLight.diffuse = new Color3(0.6, 0.5, 0.2); // Golden
        eastPortalLight.range = 10;

        // Create particle system with tropical colors
        this.createEastPortalParticles(scene, this.eastPortal.position);
    }

    /**
     * Create swirling particle effect for east portal with tropical gold/green colors.
     */
    private createEastPortalParticles(scene: Scene, position: Vector3): void {
        this.eastPortalParticles = new ParticleSystem("eastPortalParticles", 500, scene);

        // Create procedural texture
        const size = 32;
        const canvas = document.createElement("canvas");
        canvas.width = size;
        canvas.height = size;
        const ctx = canvas.getContext("2d");
        if (ctx) {
            const gradient = ctx.createRadialGradient(
                size / 2, size / 2, 0,
                size / 2, size / 2, size / 2
            );
            gradient.addColorStop(0, "rgba(255, 215, 0, 1)"); // Gold center
            gradient.addColorStop(0.3, "rgba(200, 180, 50, 0.8)");
            gradient.addColorStop(0.7, "rgba(50, 150, 80, 0.3)"); // Green edge
            gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

            ctx.fillStyle = gradient;
            ctx.fillRect(0, 0, size, size);
        }

        const particleTexture = new Texture(
            canvas.toDataURL(),
            scene,
            true,
            false
        );
        this.eastPortalParticles.particleTexture = particleTexture;

        this.eastPortalParticles.emitter = position;
        this.eastPortalParticles.minEmitBox = new Vector3(-0.5, -1.5, -0.5);
        this.eastPortalParticles.maxEmitBox = new Vector3(0.5, 1.5, 0.5);

        // Tropical colors: gold and green
        this.eastPortalParticles.color1 = new Color4(0.8, 0.7, 0.2, 1);
        this.eastPortalParticles.color2 = new Color4(0.2, 0.6, 0.3, 1);
        this.eastPortalParticles.colorDead = new Color4(0.1, 0.3, 0, 0);

        this.eastPortalParticles.minSize = 0.05;
        this.eastPortalParticles.maxSize = 0.15;

        this.eastPortalParticles.minLifeTime = 0.5;
        this.eastPortalParticles.maxLifeTime = 1.5;

        this.eastPortalParticles.emitRate = 100;

        this.eastPortalParticles.gravity = new Vector3(0, 0.5, 0);

        this.eastPortalParticles.direction1 = new Vector3(-0.5, 1, -0.5);
        this.eastPortalParticles.direction2 = new Vector3(0.5, 1, 0.5);

        this.eastPortalParticles.minAngularSpeed = 0;
        this.eastPortalParticles.maxAngularSpeed = Math.PI;

        this.eastPortalParticles.minEmitPower = 0.5;
        this.eastPortalParticles.maxEmitPower = 1.5;
        this.eastPortalParticles.updateSpeed = 0.01;

        this.eastPortalParticles.start();
    }

    /**
     * Create FPS controller with floor collision.
     */
    private createFPSController(scene: Scene): void {
        // Start player near the entrance (south side)
        const startPos = new Vector3(0, 1.7, -this.ROOM_DEPTH / 2 + 5);

        this.fpsController = new FirstPersonController(scene, startPos, {
            moveSpeed: 5.0,
            collisionTarget: this.floor,
            eyeHeight: 1.7
        });

        this.fpsController.activate();
    }

    /**
     * Check for portal proximity and trigger callback.
     */
    update(deltaTime: number): void {
        if (!this.fpsController) return;

        this.fpsController.handleInput(deltaTime);

        const pos = this.fpsController.getCamera().position;

        // Check west portal (main world) proximity
        if (!this.portalEntered && this.portal) {
            const portalPos = this.portal.position;
            const distance = Vector3.Distance(pos, portalPos);
            if (distance < this.PORTAL_RADIUS + 1) {
                this.portalEntered = true;
                this.callbacks.onPortalEnter?.();
            }
        }

        // Check east portal (tropical test world) proximity
        if (!this.eastPortalEntered && this.eastPortal) {
            const eastPortalPos = this.eastPortal.position;
            const distance = Vector3.Distance(pos, eastPortalPos);
            if (distance < this.PORTAL_RADIUS + 1) {
                this.eastPortalEntered = true;
                this.callbacks.onEastPortalEnter?.();
            }
        }
    }

    /**
     * Clean up resources.
     */
    dispose(): void {
        console.log("[LobbyScene] Disposing");

        // Dispose controllers and particles
        this.fpsController?.dispose();
        this.portalParticles?.dispose();
        this.eastPortalParticles?.dispose();

        // Dispose all meshes created in this scene
        this.floor?.dispose();
        this.statue?.dispose();
        this.portal?.dispose();
        this.eastPortal?.dispose();

        // Dispose all meshes in the scene that we created
        // (walls, ceiling, lights, etc.)
        if (this.scene) {
            // Dispose ALL meshes - copy array since we're modifying it
            const allMeshes = [...this.scene.meshes];
            console.log(`[LobbyScene] Disposing ${allMeshes.length} meshes`);
            allMeshes.forEach(m => m.dispose());

            // Dispose all lights in the scene
            const allLights = [...this.scene.lights];
            console.log(`[LobbyScene] Disposing ${allLights.length} lights`);
            allLights.forEach(l => l.dispose());
        }

        // Clear references
        this.scene = null;
        this.fpsController = null;
        this.floor = null;
        this.statue = null;
        this.portal = null;
        this.eastPortal = null;
    }
}
