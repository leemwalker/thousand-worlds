/**
 * FirstPersonController - Stub for first-person camera control.
 * Phase 3.3: Terrain Placeholders for FPS Mode
 * 
 * This is a placeholder implementation that will be expanded
 * when full FPS mode is implemented.
 */

import type { Scene } from "@babylonjs/core/scene";
import type { Camera } from "@babylonjs/core/Cameras/camera";
import { UniversalCamera } from "@babylonjs/core/Cameras/universalCamera";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import type { IPlayerController } from './interfaces';

export interface FirstPersonControllerOptions {
    moveSpeed?: number;
    lookSpeed?: number;
    jumpHeight?: number;
    gravity?: number;
}

interface Position {
    lat: number;
    lon: number;
    altitude: number;
}

/**
 * Controls first-person camera movement and input handling.
 */
export class FirstPersonController implements IPlayerController {
    private scene: Scene;
    private camera: UniversalCamera;
    private moveSpeed: number;
    private lookSpeed: number;
    private jumpHeight: number;
    private gravity: number;
    private velocity: Vector3 = Vector3.Zero();
    private isGrounded: boolean = true;
    private disposed: boolean = false;

    constructor(scene: Scene, startPosition: Vector3, options: FirstPersonControllerOptions = {}) {
        this.scene = scene;
        this.moveSpeed = options.moveSpeed ?? 5.0;
        this.lookSpeed = options.lookSpeed ?? 0.005;
        this.jumpHeight = options.jumpHeight ?? 2.0;
        this.gravity = options.gravity ?? 9.8;

        // Create FPS camera
        this.camera = new UniversalCamera("fpsCamera", startPosition, scene);
        this.camera.setTarget(startPosition.add(new Vector3(0, 0, 1)));
        this.camera.ellipsoid = new Vector3(0.5, 0.9, 0.5); // Player collision shape
        this.camera.checkCollisions = true;
        this.camera.applyGravity = true;

        // Enable WASD controls
        this.camera.keysUp = [87]; // W
        this.camera.keysDown = [83]; // S
        this.camera.keysLeft = [65]; // A
        this.camera.keysRight = [68]; // D

        // Mouse look
        this.camera.angularSensibility = 1000 / this.lookSpeed;
    }

    /**
     * Handle input for player movement.
     * Called each frame with delta time.
     */
    handleInput(deltaTime: number): void {
        if (this.disposed) return;

        // Apply gravity
        if (!this.isGrounded) {
            this.velocity.y -= this.gravity * deltaTime;
        }

        // TODO: Check for ground collision
        // TODO: Apply movement from velocity
    }

    /**
     * Get current world position as lat/lon/altitude.
     */
    getPosition(): Position {
        const pos = this.camera.position;
        // Convert world position to spherical coordinates
        // This is a simplified conversion
        const distance = Math.sqrt(pos.x * pos.x + pos.z * pos.z);
        const lat = Math.atan2(pos.y, distance) * (180 / Math.PI);
        const lon = Math.atan2(pos.z, pos.x) * (180 / Math.PI);
        const altitude = pos.length() - 1.0; // Assuming planet radius = 1

        return { lat, lon, altitude };
    }

    /**
     * Get the camera instance.
     */
    getCamera(): Camera {
        return this.camera;
    }

    /**
     * Set camera as active.
     */
    activate(): void {
        if (this.disposed) return;
        this.scene.activeCamera = this.camera;
        this.camera.attachControl(true);
    }

    /**
     * Deactivate camera controls.
     */
    deactivate(): void {
        if (this.disposed) return;
        this.camera.detachControl();
    }

    /**
     * Jump if grounded.
     */
    jump(): void {
        if (this.isGrounded && !this.disposed) {
            this.velocity.y = Math.sqrt(2 * this.gravity * this.jumpHeight);
            this.isGrounded = false;
        }
    }

    /**
     * Teleport to position.
     */
    teleport(position: Vector3): void {
        if (this.disposed) return;
        this.camera.position = position.clone();
        this.velocity = Vector3.Zero();
    }

    /**
     * Check if controller is disposed.
     */
    isDisposed(): boolean {
        return this.disposed;
    }

    /**
     * Dispose of resources.
     */
    dispose(): void {
        if (this.disposed) return;
        this.disposed = true;

        this.camera.detachControl();
        this.camera.dispose();
    }
}
