/**
 * FPSPerformanceManager Tests
 * TDD tests for performance optimization manager.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock Babylon.js dependencies
vi.mock("@babylonjs/core/Maths/math.vector", () => {
    const Vector3Mock = vi.fn().mockImplementation((x = 0, y = 0, z = 0) => ({
        x, y, z,
        length: () => Math.sqrt(x * x + y * y + z * z),
        normalize: function () { return this; },
        subtract: function (v: any) { return { x: this.x - v.x, y: this.y - v.y, z: this.z - v.z, normalize: () => ({ x: 0, y: 0, z: 1 }) }; }
    }));
    (Vector3Mock as any).Zero = () => ({ x: 0, y: 0, z: 0 });
    (Vector3Mock as any).Forward = () => ({ x: 0, y: 0, z: 1 });
    (Vector3Mock as any).Dot = () => 0.5;
    return { Vector3: Vector3Mock };
});

vi.mock("./FPSTransitionController", () => ({
    FPS_ALTITUDES: {
        FLYING: 0.15,
        LOW: 0.03,
        GROUND: 0.003
    }
}));

import { FPSPerformanceManager } from './FPSPerformanceManager';

// Mock scene and engine
const mockScene = {
    meshes: [],
    activeCamera: null,
    getEngine: () => mockEngine
};

const mockEngine = {
    drawCalls: 42,
    setHardwareScalingLevel: vi.fn()
};

const mockCamera = {
    position: { x: 1.15, y: 0, z: 0, length: () => 1.15 }
};

describe('FPSPerformanceManager', () => {
    let manager: FPSPerformanceManager;

    beforeEach(() => {
        vi.clearAllMocks();
        manager = new FPSPerformanceManager(mockScene as any, mockEngine as any);
    });

    describe('constructor', () => {
        it('should create with default options', () => {
            expect(manager).toBeDefined();
        });

        it('should accept custom options', () => {
            const customManager = new FPSPerformanceManager(mockScene as any, mockEngine as any, {
                targetFps: 30,
                minFps: 20,
                enableAutoResolution: false
            });
            expect(customManager).toBeDefined();
        });
    });

    describe('getStats', () => {
        it('should return performance statistics', () => {
            const stats = manager.getStats();

            expect(stats).toHaveProperty('fps');
            expect(stats).toHaveProperty('drawCalls');
            expect(stats).toHaveProperty('triangles');
            expect(stats).toHaveProperty('activeChunks');
            expect(stats).toHaveProperty('resolution');
            expect(stats).toHaveProperty('altitude');
        });
    });

    describe('getResolution', () => {
        it('should return current resolution (default 1.0)', () => {
            expect(manager.getResolution()).toBe(1.0);
        });
    });

    describe('setResolution', () => {
        it('should clamp resolution between 0.5 and 1.0', () => {
            manager.setResolution(0.3);
            expect(manager.getResolution()).toBe(0.5);

            manager.setResolution(1.5);
            expect(manager.getResolution()).toBe(1.0);

            manager.setResolution(0.75);
            expect(manager.getResolution()).toBe(0.75);
        });

        it('should call engine setHardwareScalingLevel', () => {
            manager.setResolution(0.8);
            expect(mockEngine.setHardwareScalingLevel).toHaveBeenCalled();
        });
    });

    describe('getAltitudeStage', () => {
        it('should return current altitude stage', () => {
            const stage = manager.getAltitudeStage();
            expect(['orbit', 'flying', 'low', 'ground']).toContain(stage);
        });
    });

    describe('getLODConfig', () => {
        it('should return LOD config for current stage', () => {
            // Update to set stage
            manager.update(mockCamera as any, 0.016);

            const config = manager.getLODConfig();
            // May be null for orbit stage
            if (config) {
                expect(config).toHaveProperty('meshSimplification');
                expect(config).toHaveProperty('materialComplexity');
            }
        });
    });

    describe('setAutoResolution', () => {
        it('should enable/disable auto resolution', () => {
            expect(() => manager.setAutoResolution(false)).not.toThrow();
            expect(() => manager.setAutoResolution(true)).not.toThrow();
        });
    });

    describe('dispose', () => {
        it('should dispose of resources', () => {
            expect(() => manager.dispose()).not.toThrow();
        });
    });
});
