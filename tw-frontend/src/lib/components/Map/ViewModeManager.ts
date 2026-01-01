/**
 * ViewModeManager - Manages transitions between orbit, tile, and terrain modes.
 * Phase 3.3: Terrain Placeholders for FPS Mode
 * 
 * Controls camera and rendering mode based on distance from planet surface.
 */

import type { Camera } from "@babylonjs/core/Cameras/camera";
import { ViewMode, type ViewModeString } from './interfaces';

export interface ViewModeManagerOptions {
    orbitThreshold?: number;   // Distance above which orbit mode is used
    terrainThreshold?: number; // Distance below which terrain mode is used
    transitionDuration?: number; // Duration of mode transitions in seconds
}

interface ViewModeCallback {
    (mode: ViewMode, previousMode: ViewMode): void;
}

/**
 * Manages view mode transitions based on camera distance.
 */
export class ViewModeManager {
    private currentMode: ViewMode = ViewMode.ORBIT;
    private previousMode: ViewMode = ViewMode.ORBIT;
    private transitionProgress: number = 1.0; // 0 = start, 1 = complete
    private orbitThreshold: number;
    private terrainThreshold: number;
    private transitionDuration: number;
    private callbacks: Set<ViewModeCallback> = new Set();

    constructor(options: ViewModeManagerOptions = {}) {
        this.orbitThreshold = options.orbitThreshold ?? 2.5;
        this.terrainThreshold = options.terrainThreshold ?? 1.2;
        this.transitionDuration = options.transitionDuration ?? 1.0;
    }

    /**
     * Get the appropriate mode for a given camera distance from planet center.
     * Planet radius is assumed to be 1.0.
     */
    getModeForDistance(distance: number): ViewModeString {
        if (distance > this.orbitThreshold) {
            return 'orbit';
        } else if (distance < this.terrainThreshold) {
            return 'terrain';
        }
        return 'tile';
    }

    /**
     * Update the view mode based on camera position.
     * Returns true if mode changed.
     */
    update(camera: Camera, deltaTime: number): boolean {
        // Calculate distance from camera to planet center (assumed at origin)
        const pos = camera.position;
        const distance = Math.sqrt(pos.x * pos.x + pos.y * pos.y + pos.z * pos.z);

        // Determine target mode
        const targetModeString = this.getModeForDistance(distance);
        const targetMode = this.stringToMode(targetModeString);

        // Check if mode changed
        if (targetMode !== this.currentMode && this.transitionProgress >= 1.0) {
            this.previousMode = this.currentMode;
            this.currentMode = targetMode;
            this.transitionProgress = 0;

            // Notify callbacks
            this.notifyCallbacks();

            return true;
        }

        // Update transition progress
        if (this.transitionProgress < 1.0) {
            this.transitionProgress += deltaTime / this.transitionDuration;
            if (this.transitionProgress > 1.0) {
                this.transitionProgress = 1.0;
            }
        }

        return false;
    }

    /**
     * Convert string mode to enum.
     */
    private stringToMode(mode: ViewModeString): ViewMode {
        switch (mode) {
            case 'orbit': return ViewMode.ORBIT;
            case 'tile': return ViewMode.TILE;
            case 'terrain': return ViewMode.TERRAIN;
            case 'transition': return ViewMode.TRANSITION;
            default: return ViewMode.ORBIT;
        }
    }

    /**
     * Get current view mode.
     */
    getCurrentMode(): ViewMode {
        return this.currentMode;
    }

    /**
     * Get current mode as string.
     */
    getCurrentModeString(): ViewModeString {
        return this.currentMode as ViewModeString;
    }

    /**
     * Get previous view mode (for transition animations).
     */
    getPreviousMode(): ViewMode {
        return this.previousMode;
    }

    /**
     * Get transition progress (0 = start, 1 = complete).
     */
    getTransitionProgress(): number {
        return this.transitionProgress;
    }

    /**
     * Check if currently in a transition.
     */
    isTransitioning(): boolean {
        return this.transitionProgress < 1.0;
    }

    /**
     * Register callback for mode changes.
     */
    onModeChange(callback: ViewModeCallback): void {
        this.callbacks.add(callback);
    }

    /**
     * Remove mode change callback.
     */
    offModeChange(callback: ViewModeCallback): void {
        this.callbacks.delete(callback);
    }

    /**
     * Notify all callbacks of mode change.
     */
    private notifyCallbacks(): void {
        for (const callback of this.callbacks) {
            callback(this.currentMode, this.previousMode);
        }
    }

    /**
     * Force a specific mode (for debugging or user override).
     */
    forceMode(mode: ViewMode): void {
        if (mode !== this.currentMode) {
            this.previousMode = this.currentMode;
            this.currentMode = mode;
            this.transitionProgress = 1.0;
            this.notifyCallbacks();
        }
    }

    /**
     * Get thresholds for debugging.
     */
    getThresholds(): { orbit: number; terrain: number } {
        return {
            orbit: this.orbitThreshold,
            terrain: this.terrainThreshold
        };
    }
}

/**
 * Standalone function for mode determination (used in tests).
 */
export function getModeForDistance(distance: number): ViewModeString {
    if (distance > 2.0) return 'orbit';
    if (distance < 1.2) return 'terrain';
    return 'tile';
}
