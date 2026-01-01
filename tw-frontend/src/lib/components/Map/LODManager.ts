/**
 * LODManager - Level of Detail manager for mesh switching based on camera distance.
 * Implements ILODManager interface.
 */

import type { Mesh, Camera, Scene } from "@babylonjs/core";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import type { ILODManager } from "./interfaces";

/**
 * Configuration for a single LOD level.
 */
export interface LODLevelConfig {
    /** Camera distance threshold (switch TO this level when distance exceeds this) */
    distance: number;
    /** Number of sphere segments for this LOD level */
    segments: number;
}

/**
 * Configuration for LODManager.
 */
export interface LODConfig {
    /** LOD levels sorted by distance (ascending). First = highest detail */
    levels: LODLevelConfig[];
    /** Hysteresis factor to prevent rapid switching (default 0.1 = 10%) */
    hysteresis?: number;
}

/**
 * Manages Level of Detail mesh switching for the globe.
 */
export class LODManager implements ILODManager {
    private config: LODConfig;
    private currentLevel: number = 0;
    private meshes: Map<number, Mesh> = new Map();
    private hysteresis: number;

    constructor(config: LODConfig) {
        this.config = config;
        this.hysteresis = config.hysteresis ?? 0.1;

        // Sort levels by distance ascending (closest = highest detail)
        this.config.levels.sort((a, b) => a.distance - b.distance);
    }

    /**
     * Get appropriate LOD level for given camera distance.
     * Returns level index (0 = highest detail, N = lowest detail).
     */
    getLODLevel(distance: number): number {
        // Find the highest quality level where camera is within threshold
        for (let i = 0; i < this.config.levels.length; i++) {
            if (distance < this.config.levels[i].distance) {
                return i;
            }
        }
        // Beyond all thresholds - return lowest detail
        return this.config.levels.length - 1;
    }

    /**
     * Get segment count for a specific LOD level.
     */
    getSegmentsForLevel(level: number): number {
        const clampedLevel = Math.max(0, Math.min(level, this.config.levels.length - 1));
        return this.config.levels[clampedLevel].segments;
    }

    /**
     * Check if LOD should update based on camera distance.
     * Applies hysteresis to prevent rapid switching at boundaries.
     */
    shouldUpdate(distance: number): boolean {
        const newLevel = this.getLODLevel(distance);

        if (newLevel === this.currentLevel) {
            return false;
        }

        // Apply hysteresis - require distance to be significantly past threshold
        const threshold = this.config.levels[Math.min(newLevel, this.config.levels.length - 1)].distance;
        const hysteresisMargin = threshold * this.hysteresis;

        // When moving to higher detail (lower level), check we're significantly under
        if (newLevel < this.currentLevel) {
            const adjustedThreshold = threshold - hysteresisMargin;
            return distance < adjustedThreshold;
        }

        // When moving to lower detail (higher level), check we're significantly over
        if (newLevel > this.currentLevel) {
            const adjustedThreshold = threshold + hysteresisMargin;
            return distance > adjustedThreshold;
        }

        return true;
    }

    /**
     * Set current LOD level (for initialization or forced updates).
     */
    setCurrentLevel(level: number): void {
        this.currentLevel = Math.max(0, Math.min(level, this.config.levels.length - 1));
    }

    /**
     * Get current LOD level.
     */
    getCurrentLevel(): number {
        return this.currentLevel;
    }

    /**
     * Get mesh for specific LOD level. Creates if not exists.
     */
    getMesh(level: number): Mesh | null {
        return this.meshes.get(level) ?? null;
    }

    /**
     * Create or get mesh for a LOD level in the given scene.
     */
    createMesh(scene: Scene, level: number, name: string = "globe"): Mesh {
        const segments = this.getSegmentsForLevel(level);
        const mesh = MeshBuilder.CreateSphere(
            `${name}_lod${level}`,
            { segments, diameter: 2, updatable: true },
            scene
        );
        this.meshes.set(level, mesh);
        return mesh;
    }

    /**
     * Update LOD based on current camera distance.
     */
    update(camera: Camera): void {
        // Calculate distance from camera to origin (where globe is)
        const distance = camera.position.length();

        if (this.shouldUpdate(distance)) {
            const newLevel = this.getLODLevel(distance);
            this.transition(newLevel);
        }
    }

    /**
     * Transition to a new LOD level.
     */
    private transition(newLevel: number): void {
        // Hide current mesh
        const currentMesh = this.meshes.get(this.currentLevel);
        if (currentMesh) {
            currentMesh.setEnabled(false);
        }

        // Show new mesh
        const newMesh = this.meshes.get(newLevel);
        if (newMesh) {
            newMesh.setEnabled(true);
        }

        this.currentLevel = newLevel;
    }

    /**
     * Dispose of all meshes.
     */
    dispose(): void {
        this.meshes.forEach((mesh) => mesh.dispose());
        this.meshes.clear();
    }
}
