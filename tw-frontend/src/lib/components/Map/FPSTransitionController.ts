/**
 * FPSTransitionController - Manages transitions from orbit to ground-level FPS.
 * Phase 4: First-Person View Implementation
 * 
 * Handles altitude-based camera animation with three stages:
 * - 1000ft: Flying view (wide landscape)
 * - 100ft: Low altitude (terrain details)
 * - 10ft: Ground level (full FPS controls)
 */

import type { Scene } from "@babylonjs/core/scene";
import type { ArcRotateCamera } from "@babylonjs/core/Cameras/arcRotateCamera";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Animation } from "@babylonjs/core/Animations/animation";
import { CubicEase, EasingFunction } from "@babylonjs/core/Animations/easing";

// Altitude in planet radius units (planet radius = 1.0)
// Assuming Earth-like planet (~6371km radius)
// 1000ft ≈ 0.305km / 6371km ≈ 0.0000478
// For visual purposes, we use exaggerated values
const ALTITUDE_FLYING = 0.15;     // 1000ft equivalent (visually)
const ALTITUDE_LOW = 0.05;        // 100ft equivalent  
const ALTITUDE_GROUND = 0.015;    // 10ft equivalent

export interface TransitionTarget {
    lat: number;    // Latitude in degrees (-90 to 90)
    lon: number;    // Longitude in degrees (-180 to 180)
    altitude: number; // Altitude in planet radius units
}

export type TransitionState = 'idle' | 'transitioning' | 'flying' | 'ground';

export interface FPSTransitionControllerOptions {
    transitionDuration?: number;  // Duration of each transition stage (seconds)
    onStateChange?: (state: TransitionState) => void;
    onAltitudeChange?: (altitude: number) => void;
}

/**
 * Converts lat/lon to Cartesian position on planet surface.
 */
function latLonToCartesian(lat: number, lon: number, radius: number): Vector3 {
    const latRad = lat * Math.PI / 180;
    const lonRad = lon * Math.PI / 180;

    const x = radius * Math.cos(latRad) * Math.cos(lonRad);
    const y = radius * Math.sin(latRad);
    const z = radius * Math.cos(latRad) * Math.sin(lonRad);

    return new Vector3(x, y, z);
}

/**
 * Converts Cartesian position to lat/lon.
 */
function cartesianToLatLon(position: Vector3): { lat: number; lon: number } {
    const radius = position.length();
    const lat = Math.asin(position.y / radius) * 180 / Math.PI;
    const lon = Math.atan2(position.z, position.x) * 180 / Math.PI;
    return { lat, lon };
}

/**
 * Controller for transitioning from orbital view to ground-level FPS.
 */
export class FPSTransitionController {
    private scene: Scene;
    private camera: ArcRotateCamera;
    private state: TransitionState = 'idle';
    private target: TransitionTarget | null = null;
    private transitionDuration: number;
    private onStateChange?: (state: TransitionState) => void;
    private onAltitudeChange?: (altitude: number) => void;
    private animationGroup: Animation[] = [];

    constructor(
        scene: Scene,
        camera: ArcRotateCamera,
        options: FPSTransitionControllerOptions = {}
    ) {
        this.scene = scene;
        this.camera = camera;
        this.transitionDuration = options.transitionDuration ?? 2.0;
        this.onStateChange = options.onStateChange;
        this.onAltitudeChange = options.onAltitudeChange;
    }

    /**
     * Get current transition state.
     */
    getState(): TransitionState {
        return this.state;
    }

    /**
     * Get current target, if any.
     */
    getTarget(): TransitionTarget | null {
        return this.target;
    }

    /**
     * Get current altitude (camera radius - planet radius).
     */
    getCurrentAltitude(): number {
        const radius = this.camera.radius;
        return radius - 1.0; // Assuming planet radius = 1.0
    }

    /**
     * Start transition to a target lat/lon at flying altitude (1000ft).
     */
    transitionToFlying(lat: number, lon: number): void {
        this.target = { lat, lon, altitude: ALTITUDE_FLYING };
        this.state = 'transitioning';
        this.onStateChange?.(this.state);

        this.animateToTarget(this.target);
    }

    /**
     * Start transition to ground level (10ft) at current/target position.
     */
    transitionToGround(lat?: number, lon?: number): void {
        const targetLat = lat ?? this.target?.lat ?? 0;
        const targetLon = lon ?? this.target?.lon ?? 0;

        this.target = { lat: targetLat, lon: targetLon, altitude: ALTITUDE_GROUND };
        this.state = 'transitioning';
        this.onStateChange?.(this.state);

        this.animateToTarget(this.target);
    }

    /**
     * Return to orbital view.
     */
    returnToOrbit(): void {
        this.target = null;
        this.state = 'transitioning';
        this.onStateChange?.(this.state);

        // Animate camera radius back to orbit distance
        this.animateCameraRadius(5.0, () => {
            this.state = 'idle';
            this.onStateChange?.(this.state);
        });
    }

    /**
     * Handle click on planet surface - raycast to find target position.
     * Returns true if a valid target was found.
     */
    handlePlanetClick(pickX: number, pickY: number): boolean {
        const pickResult = this.scene.pick(pickX, pickY);

        if (pickResult?.hit && pickResult.pickedPoint) {
            // Check if we hit the planet (not sun or moons)
            const meshName = pickResult.pickedMesh?.name ?? '';
            if (meshName.includes('globe') || meshName.includes('planet')) {
                const { lat, lon } = cartesianToLatLon(pickResult.pickedPoint);
                this.transitionToFlying(lat, lon);
                return true;
            }
        }

        return false;
    }

    /**
     * Animate camera to target position.
     */
    private animateToTarget(target: TransitionTarget): void {
        // Calculate target position on planet surface
        const surfacePos = latLonToCartesian(target.lat, target.lon, 1.0);

        // Calculate camera alpha/beta to look at this position
        const targetAlpha = Math.atan2(surfacePos.z, surfacePos.x);
        const targetBeta = Math.PI / 2 - Math.asin(surfacePos.y);
        const targetRadius = 1.0 + target.altitude;

        // Create easing function
        const easingFunction = new CubicEase();
        easingFunction.setEasingMode(EasingFunction.EASINGMODE_EASEINOUT);

        // Animate alpha
        const alphaAnim = new Animation(
            "alphaAnim",
            "alpha",
            30,
            Animation.ANIMATIONTYPE_FLOAT,
            Animation.ANIMATIONLOOPMODE_CONSTANT
        );
        alphaAnim.setKeys([
            { frame: 0, value: this.camera.alpha },
            { frame: 60, value: targetAlpha }
        ]);
        alphaAnim.setEasingFunction(easingFunction);

        // Animate beta
        const betaAnim = new Animation(
            "betaAnim",
            "beta",
            30,
            Animation.ANIMATIONTYPE_FLOAT,
            Animation.ANIMATIONLOOPMODE_CONSTANT
        );
        betaAnim.setKeys([
            { frame: 0, value: this.camera.beta },
            { frame: 60, value: targetBeta }
        ]);
        betaAnim.setEasingFunction(easingFunction);

        // Animate radius
        const radiusAnim = new Animation(
            "radiusAnim",
            "radius",
            30,
            Animation.ANIMATIONTYPE_FLOAT,
            Animation.ANIMATIONLOOPMODE_CONSTANT
        );
        radiusAnim.setKeys([
            { frame: 0, value: this.camera.radius },
            { frame: 60, value: targetRadius }
        ]);
        radiusAnim.setEasingFunction(easingFunction);

        // Run animations
        this.scene.beginDirectAnimation(
            this.camera,
            [alphaAnim, betaAnim, radiusAnim],
            0,
            60,
            false,
            1.0,
            () => {
                // Animation complete
                if (target.altitude <= ALTITUDE_GROUND) {
                    this.state = 'ground';
                } else {
                    this.state = 'flying';
                }
                this.onStateChange?.(this.state);
                this.onAltitudeChange?.(target.altitude);
            }
        );
    }

    /**
     * Animate only camera radius.
     */
    private animateCameraRadius(targetRadius: number, onComplete?: () => void): void {
        const easingFunction = new CubicEase();
        easingFunction.setEasingMode(EasingFunction.EASINGMODE_EASEINOUT);

        const radiusAnim = new Animation(
            "radiusAnim",
            "radius",
            30,
            Animation.ANIMATIONTYPE_FLOAT,
            Animation.ANIMATIONLOOPMODE_CONSTANT
        );
        radiusAnim.setKeys([
            { frame: 0, value: this.camera.radius },
            { frame: 60, value: targetRadius }
        ]);
        radiusAnim.setEasingFunction(easingFunction);

        this.scene.beginDirectAnimation(
            this.camera,
            [radiusAnim],
            0,
            60,
            false,
            1.0,
            onComplete
        );
    }

    /**
     * Descend from current altitude to ground level.
     */
    descendToGround(): void {
        if (this.target) {
            this.transitionToGround(this.target.lat, this.target.lon);
        }
    }

    /**
     * Ascend from ground to flying altitude.
     */
    ascendToFlying(): void {
        if (this.target) {
            this.target.altitude = ALTITUDE_FLYING;
            this.state = 'transitioning';
            this.onStateChange?.(this.state);
            this.animateToTarget(this.target);
        }
    }

    /**
     * Dispose of controller resources.
     */
    dispose(): void {
        this.stopAllAnimations();
    }

    /**
     * Stop any running animations.
     */
    private stopAllAnimations(): void {
        this.scene.stopAllAnimations();
    }
}

// Export altitude constants for use in other modules
export const FPS_ALTITUDES = {
    FLYING: ALTITUDE_FLYING,
    LOW: ALTITUDE_LOW,
    GROUND: ALTITUDE_GROUND
};
