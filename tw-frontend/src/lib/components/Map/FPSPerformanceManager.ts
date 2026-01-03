/**
 * FPSPerformanceManager - Performance optimization for FPS mode.
 * Phase 3: Performance Optimization
 * 
 * Handles:
 * - Aggressive LOD based on altitude
 * - Dynamic render resolution scaling
 * - Frustum culling
 * - Chunk pooling and caching
 * - Performance monitoring
 */

import type { Scene } from "@babylonjs/core/scene";
import type { Camera } from "@babylonjs/core/Cameras/camera";
import type { Engine } from "@babylonjs/core/Engines/engine";
import type { Mesh } from "@babylonjs/core/Meshes/mesh";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { FPS_ALTITUDES } from './FPSTransitionController';

// Performance targets
const TARGET_FPS = 60;
const MIN_FPS = 45;
const RESOLUTION_STEP = 0.1;
const MIN_RESOLUTION = 0.5;
const MAX_RESOLUTION = 1.0;

// LOD configuration by altitude stage
interface AltitudeLODConfig {
    meshSimplification: number;   // 0-1, 0 = full detail
    materialComplexity: number;   // 0-1, 0 = simple, 1 = full
    normalMapsEnabled: boolean;
    shadowsEnabled: boolean;
    detailRadius: number;         // Radius for full detail (in world units)
}

const ALTITUDE_LOD_CONFIGS: Record<string, AltitudeLODConfig> = {
    flying: {
        meshSimplification: 0.75,    // 75% simplified
        materialComplexity: 0.3,      // Simple colors only
        normalMapsEnabled: false,
        shadowsEnabled: false,
        detailRadius: 5.0,           // Wide area, low detail
    },
    low: {
        meshSimplification: 0.5,      // 50% simplified
        materialComplexity: 0.6,      // 3-4 materials
        normalMapsEnabled: true,
        shadowsEnabled: false,
        detailRadius: 1.0,           // Medium area
    },
    ground: {
        meshSimplification: 0.0,      // Full detail
        materialComplexity: 1.0,      // Full materials
        normalMapsEnabled: true,
        shadowsEnabled: true,
        detailRadius: 0.2,           // Small area, full detail
    }
};

export interface PerformanceStats {
    fps: number;
    drawCalls: number;
    triangles: number;
    activeChunks: number;
    textureMemory: number;           // MB
    resolution: number;
    altitude: string;
    lastUpdateTime: number;
}

export interface FPSPerformanceManagerOptions {
    targetFps?: number;
    minFps?: number;
    enableAutoResolution?: boolean;
    chunkPoolSize?: number;
}

/**
 * Manages performance optimization for FPS mode.
 */
export class FPSPerformanceManager {
    private scene: Scene;
    private engine: Engine;
    private targetFps: number;
    private minFps: number;
    private enableAutoResolution: boolean;
    private currentResolution: number = 1.0;
    private fpsHistory: number[] = [];
    private lastFrameTime: number = 0;
    private frameCount: number = 0;
    private currentAltitudeStage: string = 'flying';

    // Chunk pool
    private chunkPool: Mesh[] = [];
    private activeChunks: Set<Mesh> = new Set();
    private chunkPoolSize: number;

    // Stats
    private stats: PerformanceStats = {
        fps: 60,
        drawCalls: 0,
        triangles: 0,
        activeChunks: 0,
        textureMemory: 0,
        resolution: 1.0,
        altitude: 'flying',
        lastUpdateTime: 0
    };

    constructor(scene: Scene, engine: Engine, options: FPSPerformanceManagerOptions = {}) {
        this.scene = scene;
        this.engine = engine;
        this.targetFps = options.targetFps ?? TARGET_FPS;
        this.minFps = options.minFps ?? MIN_FPS;
        this.enableAutoResolution = options.enableAutoResolution ?? true;
        this.chunkPoolSize = options.chunkPoolSize ?? 50;
    }

    /**
     * Update performance metrics and adjust settings.
     * Call once per frame.
     */
    update(camera: Camera, deltaTime: number): void {
        // Calculate FPS
        this.frameCount++;
        const currentTime = performance.now();

        if (currentTime - this.lastFrameTime >= 1000) {
            const fps = this.frameCount;
            this.fpsHistory.push(fps);
            if (this.fpsHistory.length > 10) {
                this.fpsHistory.shift();
            }

            this.stats.fps = fps;
            this.frameCount = 0;
            this.lastFrameTime = currentTime;

            // Auto-adjust resolution based on FPS
            if (this.enableAutoResolution) {
                this.adjustResolution(fps);
            }
        }

        // Update altitude stage
        const altitude = camera.position.length() - 1.0;
        this.updateAltitudeStage(altitude);

        // Update stats
        this.updateStats();
    }

    /**
     * Determine altitude stage from camera altitude.
     */
    private updateAltitudeStage(altitude: number): void {
        if (altitude > FPS_ALTITUDES.FLYING * 1.5) {
            this.currentAltitudeStage = 'orbit';
        } else if (altitude > FPS_ALTITUDES.LOW * 1.5) {
            this.currentAltitudeStage = 'flying';
        } else if (altitude > FPS_ALTITUDES.GROUND * 1.5) {
            this.currentAltitudeStage = 'low';
        } else {
            this.currentAltitudeStage = 'ground';
        }
        this.stats.altitude = this.currentAltitudeStage;
    }

    /**
     * Adjust render resolution based on FPS.
     */
    private adjustResolution(currentFps: number): void {
        const avgFps = this.getAverageFps();

        if (avgFps < this.minFps && this.currentResolution > MIN_RESOLUTION) {
            // Reduce resolution to improve FPS
            this.currentResolution = Math.max(MIN_RESOLUTION, this.currentResolution - RESOLUTION_STEP);
            this.applyResolution();
            console.log(`[Performance] Reduced resolution to ${this.currentResolution.toFixed(2)} (FPS: ${avgFps.toFixed(1)})`);
        } else if (avgFps > this.targetFps + 10 && this.currentResolution < MAX_RESOLUTION) {
            // Increase resolution if we have headroom
            this.currentResolution = Math.min(MAX_RESOLUTION, this.currentResolution + RESOLUTION_STEP);
            this.applyResolution();
            console.log(`[Performance] Increased resolution to ${this.currentResolution.toFixed(2)} (FPS: ${avgFps.toFixed(1)})`);
        }
    }

    /**
     * Apply current resolution to engine.
     */
    private applyResolution(): void {
        this.engine.setHardwareScalingLevel(1 / this.currentResolution);
        this.stats.resolution = this.currentResolution;
    }

    /**
     * Get average FPS from history.
     */
    private getAverageFps(): number {
        if (this.fpsHistory.length === 0) return 60;
        return this.fpsHistory.reduce((a, b) => a + b, 0) / this.fpsHistory.length;
    }

    /**
     * Get LOD config for current altitude.
     */
    getLODConfig(): AltitudeLODConfig | null {
        return ALTITUDE_LOD_CONFIGS[this.currentAltitudeStage] ?? null;
    }

    /**
     * Check if a point is in camera frustum.
     */
    isInFrustum(position: Vector3): boolean {
        const camera = this.scene.activeCamera;
        if (!camera) return true;

        // Use Babylon's built-in frustum check via bounding info
        // Simplified: just check if position is in front of camera
        const cameraPos = camera.position;
        const cameraDir = camera.getDirection(Vector3.Forward());
        const toPoint = position.subtract(cameraPos);
        const dot = Vector3.Dot(toPoint.normalize(), cameraDir);

        return dot > -0.2; // Allow some behind for smooth transitions
    }

    /**
     * Acquire a chunk mesh from the pool.
     */
    acquireChunk(): Mesh | null {
        if (this.chunkPool.length > 0) {
            const chunk = this.chunkPool.pop()!;
            this.activeChunks.add(chunk);
            chunk.setEnabled(true);
            return chunk;
        }
        return null;
    }

    /**
     * Release a chunk mesh back to the pool.
     */
    releaseChunk(chunk: Mesh): void {
        this.activeChunks.delete(chunk);
        chunk.setEnabled(false);

        if (this.chunkPool.length < this.chunkPoolSize) {
            this.chunkPool.push(chunk);
        } else {
            chunk.dispose(); // Pool is full, dispose
        }
    }

    /**
     * Pre-allocate chunk pool.
     */
    initializeChunkPool(createChunkFn: () => Mesh): void {
        for (let i = 0; i < this.chunkPoolSize; i++) {
            const chunk = createChunkFn();
            chunk.setEnabled(false);
            this.chunkPool.push(chunk);
        }
        console.log(`[Performance] Initialized chunk pool with ${this.chunkPoolSize} meshes`);
    }

    /**
     * Update performance statistics.
     */
    private updateStats(): void {
        // Get engine stats
        const sceneStats = this.scene.getEngine();

        this.stats.drawCalls = sceneStats.drawCalls;
        this.stats.activeChunks = this.activeChunks.size;
        this.stats.lastUpdateTime = performance.now();

        // Estimate triangle count from meshes
        let triangles = 0;
        this.scene.meshes.forEach(mesh => {
            if (mesh.isEnabled() && mesh.getTotalVertices) {
                triangles += mesh.getTotalIndices() / 3;
            }
        });
        this.stats.triangles = triangles;
    }

    /**
     * Get current performance stats.
     */
    getStats(): PerformanceStats {
        return { ...this.stats };
    }

    /**
     * Get current resolution scale.
     */
    getResolution(): number {
        return this.currentResolution;
    }

    /**
     * Get current altitude stage.
     */
    getAltitudeStage(): string {
        return this.currentAltitudeStage;
    }

    /**
     * Force a specific resolution.
     */
    setResolution(resolution: number): void {
        this.currentResolution = Math.max(MIN_RESOLUTION, Math.min(MAX_RESOLUTION, resolution));
        this.applyResolution();
    }

    /**
     * Enable or disable auto resolution scaling.
     */
    setAutoResolution(enabled: boolean): void {
        this.enableAutoResolution = enabled;
    }

    /**
     * Dispose of all resources.
     */
    dispose(): void {
        // Dispose pooled chunks
        this.chunkPool.forEach(chunk => chunk.dispose());
        this.chunkPool = [];

        // Dispose active chunks
        this.activeChunks.forEach(chunk => chunk.dispose());
        this.activeChunks.clear();
    }
}
