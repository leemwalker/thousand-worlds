/**
 * FPSMovementController - Unified movement controller for flying/walking/swimming.
 * Phase 2: Core Movement
 * 
 * Handles altitude-aware movement with automatic mode switching:
 * - Flying: Free movement at 1000ft-100ft altitude
 * - Walking: Ground-clamped with kinematic controls at 10ft
 * - Swimming: Buoyancy and 3D movement when in water
 */

import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import type { Scene } from "@babylonjs/core/scene";
import type { UniversalCamera } from "@babylonjs/core/Cameras/universalCamera";
import { GroundRaycastSystem, type GroundRaycastResult } from './GroundRaycastSystem';
import type { IHeightmapProvider } from './interfaces';

export type MovementMode = 'flying' | 'walking' | 'swimming' | 'wading';

export interface FPSMovementControllerOptions {
    flySpeed?: number;          // Flying mode speed
    walkSpeed?: number;         // Walking mode speed
    swimSpeed?: number;         // Swimming mode speed
    jumpHeight?: number;        // Jump height in world units
    gravity?: number;           // Gravity acceleration
    buoyancy?: number;          // Buoyancy force in water
    eyeHeight?: number;         // Camera height above ground
    wadingDepthThreshold?: number;  // Depth at which wading becomes swimming
}

interface InputState {
    forward: boolean;
    backward: boolean;
    left: boolean;
    right: boolean;
    up: boolean;      // Space for jump/swim up
    down: boolean;    // Shift for swim down
    sprint: boolean;  // Shift for sprint when walking
}

export interface MovementState {
    mode: MovementMode;
    velocity: Vector3;
    isGrounded: boolean;
    isInWater: boolean;
    waterDepth: number;
    slopeAngle: number;
    groundHeight: number;
}

/**
 * Unified movement controller for all FPS modes.
 */
export class FPSMovementController {
    private scene: Scene;
    private camera: UniversalCamera | null = null;
    private groundRaycast: GroundRaycastSystem;
    private heightmapProvider: IHeightmapProvider | null = null;

    // Configuration
    private flySpeed: number;
    private walkSpeed: number;
    private swimSpeed: number;
    private jumpHeight: number;
    private gravity: number;
    private buoyancy: number;
    private eyeHeight: number;
    private wadingDepthThreshold: number;

    // State
    private mode: MovementMode = 'flying';
    private velocity: Vector3 = Vector3.Zero();
    private isGrounded: boolean = false;
    private input: InputState = {
        forward: false, backward: false,
        left: false, right: false,
        up: false, down: false,
        sprint: false
    };

    // Oxygen stub (for watcher - not actively used)
    private oxygenLevel: number = 100;
    private maxOxygen: number = 100;
    private oxygenDrainRate: number = 5; // Per second underwater
    private oxygenRefillRate: number = 20; // Per second above water

    constructor(scene: Scene, options: FPSMovementControllerOptions = {}) {
        this.scene = scene;
        this.groundRaycast = new GroundRaycastSystem({
            capsuleRadius: 0.3,
            maxGroundDistance: 0.5,
            stepHeight: 0.3
        });

        // Apply options with defaults
        this.flySpeed = options.flySpeed ?? 10.0;
        this.walkSpeed = options.walkSpeed ?? 5.0;
        this.swimSpeed = options.swimSpeed ?? 3.0;
        this.jumpHeight = options.jumpHeight ?? 2.0;
        this.gravity = options.gravity ?? 9.8;
        this.buoyancy = options.buoyancy ?? 4.0;
        this.eyeHeight = options.eyeHeight ?? 1.7;
        this.wadingDepthThreshold = options.wadingDepthThreshold ?? 1.0;
    }

    /**
     * Set the camera to control.
     */
    setCamera(camera: UniversalCamera): void {
        this.camera = camera;
        this.setupInputHandlers();
    }

    /**
     * Set heightmap provider for ground detection.
     */
    setHeightmapProvider(provider: IHeightmapProvider): void {
        this.heightmapProvider = provider;
        this.groundRaycast.setHeightmapProvider(provider);
    }

    /**
     * Set planet parameters.
     */
    setPlanetParams(radius: number, seaLevel: number): void {
        this.groundRaycast.setPlanetParams(radius, seaLevel);
    }

    /**
     * Force a specific movement mode.
     */
    setMode(mode: MovementMode): void {
        this.mode = mode;
    }

    /**
     * Get current movement mode.
     */
    getMode(): MovementMode {
        return this.mode;
    }

    /**
     * Get full movement state.
     */
    getState(): MovementState {
        const lastGround = this.lastGroundResult;
        return {
            mode: this.mode,
            velocity: this.velocity.clone(),
            isGrounded: this.isGrounded,
            isInWater: lastGround?.isInWater ?? false,
            waterDepth: lastGround?.waterDepth ?? 0,
            slopeAngle: lastGround?.slopeAngle ?? 0,
            groundHeight: lastGround?.groundHeight ?? 0
        };
    }

    private lastGroundResult: GroundRaycastResult | null = null;

    /**
     * Update movement each frame.
     */
    update(deltaTime: number): void {
        if (!this.camera) return;

        // Calculate ground detection
        const forward = this.camera.getDirection(Vector3.Forward());
        const position = this.camera.position.clone();
        this.lastGroundResult = this.groundRaycast.castFromPosition(position, forward);

        // Determine movement mode based on altitude and water
        this.updateMode(this.lastGroundResult);

        // Apply movement based on mode
        switch (this.mode) {
            case 'flying':
                this.updateFlying(deltaTime);
                break;
            case 'walking':
                this.updateWalking(deltaTime, this.lastGroundResult);
                break;
            case 'swimming':
                this.updateSwimming(deltaTime, this.lastGroundResult);
                break;
            case 'wading':
                this.updateWading(deltaTime, this.lastGroundResult);
                break;
        }

        // Update oxygen (stub)
        this.updateOxygen(deltaTime);
    }

    /**
     * Update movement mode based on current conditions.
     */
    private updateMode(ground: GroundRaycastResult): void {
        if (!this.camera) return;

        const altitude = this.camera.position.length() - 1.0;

        // High altitude = flying
        if (altitude > 0.05) {
            this.mode = 'flying';
            return;
        }

        // In water?
        if (ground.isInWater) {
            if (ground.waterDepth > this.wadingDepthThreshold) {
                this.mode = 'swimming';
            } else {
                this.mode = 'wading';
            }
            return;
        }

        // On ground
        this.mode = 'walking';
    }

    /**
     * Flying mode: Free 3D movement.
     */
    private updateFlying(deltaTime: number): void {
        if (!this.camera) return;

        const speed = this.flySpeed * deltaTime;
        const forward = this.camera.getDirection(Vector3.Forward());
        const right = this.camera.getDirection(Vector3.Right());
        const up = Vector3.Up();

        let movement = Vector3.Zero();

        if (this.input.forward) movement = movement.add(forward.scale(speed));
        if (this.input.backward) movement = movement.add(forward.scale(-speed));
        if (this.input.left) movement = movement.add(right.scale(-speed));
        if (this.input.right) movement = movement.add(right.scale(speed));
        if (this.input.up) movement = movement.add(up.scale(speed));
        if (this.input.down) movement = movement.add(up.scale(-speed));

        this.camera.position = this.camera.position.add(movement);
    }

    /**
     * Walking mode: Ground-clamped movement with gravity.
     */
    private updateWalking(deltaTime: number, ground: GroundRaycastResult): void {
        if (!this.camera) return;

        const speed = this.input.sprint ? this.walkSpeed * 1.5 : this.walkSpeed;
        const moveSpeed = speed * deltaTime;

        // Get camera directions (flattened to horizontal plane)
        const cameraForward = this.camera.getDirection(Vector3.Forward());
        const horizontalForward = new Vector3(cameraForward.x, 0, cameraForward.z).normalize();
        const horizontalRight = new Vector3(cameraForward.z, 0, -cameraForward.x).normalize();

        // Calculate horizontal movement
        let horizontalMove = Vector3.Zero();
        if (this.input.forward) horizontalMove = horizontalMove.add(horizontalForward.scale(moveSpeed));
        if (this.input.backward) horizontalMove = horizontalMove.add(horizontalForward.scale(-moveSpeed));
        if (this.input.left) horizontalMove = horizontalMove.add(horizontalRight.scale(-moveSpeed));
        if (this.input.right) horizontalMove = horizontalMove.add(horizontalRight.scale(moveSpeed));

        // Apply horizontal movement
        this.camera.position = this.camera.position.add(horizontalMove);

        // Handle jumping and gravity
        if (ground.isGrounded) {
            this.isGrounded = true;
            this.velocity.y = 0;

            // Jump
            if (this.input.up) {
                this.velocity.y = Math.sqrt(2 * this.gravity * this.jumpHeight);
                this.isGrounded = false;
            }
        } else {
            this.isGrounded = false;
            this.velocity.y -= this.gravity * deltaTime;
        }

        // Apply vertical velocity
        this.camera.position.y += this.velocity.y * deltaTime;

        // Clamp to ground + eye height
        const minHeight = ground.groundHeight + this.eyeHeight;
        if (this.camera.position.length() < minHeight + 1.0) {
            // Clamp to ground
            const dir = this.camera.position.normalize();
            this.camera.position = dir.scale(minHeight + 1.0);
            if (this.velocity.y < 0) {
                this.velocity.y = 0;
                this.isGrounded = true;
            }
        }
    }

    /**
     * Swimming mode: 3D movement with buoyancy.
     */
    private updateSwimming(deltaTime: number, ground: GroundRaycastResult): void {
        if (!this.camera) return;

        const speed = this.swimSpeed * deltaTime;
        const forward = this.camera.getDirection(Vector3.Forward());
        const right = this.camera.getDirection(Vector3.Right());

        let movement = Vector3.Zero();

        if (this.input.forward) movement = movement.add(forward.scale(speed));
        if (this.input.backward) movement = movement.add(forward.scale(-speed));
        if (this.input.left) movement = movement.add(right.scale(-speed));
        if (this.input.right) movement = movement.add(right.scale(speed));
        if (this.input.up) movement.y += speed;
        if (this.input.down) movement.y -= speed;

        // Apply buoyancy (float upward)
        movement.y += this.buoyancy * deltaTime * 0.1;

        this.camera.position = this.camera.position.add(movement);

        // Clamp to water surface (don't go above)
        const waterSurface = 1.0 + 0.001; // Planet radius + small offset
        if (this.camera.position.length() > waterSurface) {
            const dir = this.camera.position.normalize();
            this.camera.position = dir.scale(waterSurface);
        }
    }

    /**
     * Wading mode: Slower walking in shallow water.
     */
    private updateWading(deltaTime: number, ground: GroundRaycastResult): void {
        // Use walking logic but slower
        const originalSpeed = this.walkSpeed;
        this.walkSpeed *= 0.5; // 50% speed in water
        this.updateWalking(deltaTime, ground);
        this.walkSpeed = originalSpeed;
    }

    /**
     * Update oxygen level (stub for watcher character).
     */
    private updateOxygen(deltaTime: number): void {
        if (this.mode === 'swimming' && this.lastGroundResult && this.lastGroundResult.waterDepth > this.eyeHeight) {
            // Underwater - drain oxygen
            this.oxygenLevel = Math.max(0, this.oxygenLevel - this.oxygenDrainRate * deltaTime);
        } else {
            // Above water - refill oxygen
            this.oxygenLevel = Math.min(this.maxOxygen, this.oxygenLevel + this.oxygenRefillRate * deltaTime);
        }
    }

    /**
     * Get current oxygen level (0-100).
     */
    getOxygenLevel(): number {
        return this.oxygenLevel;
    }

    /**
     * Set up keyboard input handlers.
     */
    private setupInputHandlers(): void {
        const canvas = this.scene.getEngine().getRenderingCanvas();
        if (!canvas) return;

        canvas.addEventListener('keydown', (event) => {
            switch (event.code) {
                case 'KeyW': this.input.forward = true; break;
                case 'KeyS': this.input.backward = true; break;
                case 'KeyA': this.input.left = true; break;
                case 'KeyD': this.input.right = true; break;
                case 'Space': this.input.up = true; break;
                case 'ShiftLeft':
                case 'ShiftRight':
                    this.input.down = true;
                    this.input.sprint = true;
                    break;
            }
        });

        canvas.addEventListener('keyup', (event) => {
            switch (event.code) {
                case 'KeyW': this.input.forward = false; break;
                case 'KeyS': this.input.backward = false; break;
                case 'KeyA': this.input.left = false; break;
                case 'KeyD': this.input.right = false; break;
                case 'Space': this.input.up = false; break;
                case 'ShiftLeft':
                case 'ShiftRight':
                    this.input.down = false;
                    this.input.sprint = false;
                    break;
            }
        });
    }

    /**
     * Dispose of controller resources.
     */
    dispose(): void {
        this.camera = null;
    }
}
