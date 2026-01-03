/**
 * FPSMovementController Tests
 * TDD tests for unified movement controller.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock Babylon.js dependencies
vi.mock("@babylonjs/core/Maths/math.vector", () => {
    const Vector3Mock = vi.fn().mockImplementation((x = 0, y = 0, z = 0) => ({
        x, y, z,
        length: () => Math.sqrt(x * x + y * y + z * z),
        normalize: function () { return this; },
        scale: function (s: number) { return { x: this.x * s, y: this.y * s, z: this.z * s }; },
        clone: function () { return { ...this }; },
        add: function (v: any) { return { x: this.x + v.x, y: this.y + v.y, z: this.z + v.z }; },
    }));
    (Vector3Mock as any).Zero = () => ({ x: 0, y: 0, z: 0 });
    (Vector3Mock as any).Forward = () => ({ x: 0, y: 0, z: 1 });
    (Vector3Mock as any).Up = () => ({ x: 0, y: 1, z: 0 });
    return { Vector3: Vector3Mock };
});

// Re-import after mock
import { FPSMovementController } from './FPSMovementController';

// Mock scene and camera
const mockScene = {
    getEngine: () => ({
        getRenderingCanvas: () => ({
            addEventListener: vi.fn(),
            removeEventListener: vi.fn()
        })
    })
};

const mockCamera = {
    position: { x: 1.05, y: 0, z: 0, length: () => 1.05, clone: () => ({ x: 1.05, y: 0, z: 0 }) },
    getDirection: vi.fn().mockReturnValue({ x: 0, y: 0, z: 1, normalize: () => ({ x: 0, y: 0, z: 1 }), scale: () => ({ x: 0, y: 0, z: 0 }) }),
    fov: 1.3,
    angularSensibility: 2000,
    invertRotation: false
};

describe('FPSMovementController', () => {
    let controller: FPSMovementController;

    beforeEach(() => {
        vi.clearAllMocks();
        controller = new FPSMovementController(mockScene as any);
    });

    describe('constructor', () => {
        it('should create with default options', () => {
            expect(controller).toBeDefined();
            expect(controller.getMode()).toBe('flying');
        });

        it('should accept custom options', () => {
            const customController = new FPSMovementController(mockScene as any, {
                flySpeed: 20,
                walkSpeed: 10,
                swimSpeed: 5
            });
            expect(customController).toBeDefined();
        });
    });

    describe('setMode', () => {
        it('should change movement mode', () => {
            controller.setMode('walking');
            expect(controller.getMode()).toBe('walking');
        });

        it('should accept all valid modes', () => {
            const modes = ['flying', 'walking', 'swimming', 'wading'] as const;
            for (const mode of modes) {
                controller.setMode(mode);
                expect(controller.getMode()).toBe(mode);
            }
        });
    });

    describe('getState', () => {
        it('should return current movement state', () => {
            const state = controller.getState();

            expect(state).toHaveProperty('mode');
            expect(state).toHaveProperty('velocity');
            expect(state).toHaveProperty('isGrounded');
            expect(state).toHaveProperty('isInWater');
            expect(state).toHaveProperty('waterDepth');
            expect(state).toHaveProperty('slopeAngle');
        });
    });

    describe('getOxygenLevel', () => {
        it('should return oxygen level', () => {
            const oxygen = controller.getOxygenLevel();
            expect(oxygen).toBe(100); // Starts at 100
        });
    });

    describe('setCamera', () => {
        it('should set camera reference', () => {
            // Should not throw
            expect(() => controller.setCamera(mockCamera as any)).not.toThrow();
        });
    });

    describe('dispose', () => {
        it('should dispose of resources', () => {
            expect(() => controller.dispose()).not.toThrow();
        });
    });
});
