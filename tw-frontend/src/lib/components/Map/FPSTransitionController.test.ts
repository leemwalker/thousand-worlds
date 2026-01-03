/**
 * FPSTransitionController Tests
 * TDD tests for altitude-based camera transitions.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock Babylon.js dependencies
const mockCamera = {
    alpha: 0,
    beta: Math.PI / 2,
    radius: 5.0,
    position: { x: 5, y: 0, z: 0, length: () => 5 }
};

const mockScene = {
    pick: vi.fn(),
    beginDirectAnimation: vi.fn((target, anims, from, to, loop, speed, onComplete) => {
        // Immediately call onComplete for testing
        onComplete?.();
    }),
    stopAllAnimations: vi.fn()
};

vi.mock("@babylonjs/core/Maths/math.vector", () => ({
    Vector3: vi.fn().mockImplementation((x, y, z) => ({ x, y, z, length: () => Math.sqrt(x * x + y * y + z * z) }))
}));

vi.mock("@babylonjs/core/Animations/animation", () => ({
    Animation: vi.fn().mockImplementation(() => ({
        setKeys: vi.fn(),
        setEasingFunction: vi.fn()
    }))
}));

vi.mock("@babylonjs/core/Animations/easing", () => ({
    CubicEase: vi.fn().mockImplementation(() => ({
        setEasingMode: vi.fn()
    })),
    EasingFunction: {
        EASINGMODE_EASEINOUT: 2
    }
}));

import { FPSTransitionController, FPS_ALTITUDES } from './FPSTransitionController';

describe('FPSTransitionController', () => {
    let controller: FPSTransitionController;

    beforeEach(() => {
        vi.clearAllMocks();
        controller = new FPSTransitionController(
            mockScene as any,
            mockCamera as any
        );
    });

    describe('constructor', () => {
        it('should initialize in idle state', () => {
            expect(controller.getState()).toBe('idle');
        });

        it('should have no target initially', () => {
            expect(controller.getTarget()).toBeNull();
        });
    });

    describe('transitionToFlying', () => {
        it('should set state to transitioning then flying', () => {
            const onStateChange = vi.fn();
            controller = new FPSTransitionController(
                mockScene as any,
                mockCamera as any,
                { onStateChange }
            );

            controller.transitionToFlying(45, -122);

            // Should have transitioned to flying (mock completes immediately)
            expect(onStateChange).toHaveBeenCalledWith('transitioning');
            expect(onStateChange).toHaveBeenCalledWith('flying');
        });

        it('should set target lat/lon', () => {
            controller.transitionToFlying(45, -122);

            const target = controller.getTarget();
            expect(target?.lat).toBe(45);
            expect(target?.lon).toBe(-122);
            expect(target?.altitude).toBe(FPS_ALTITUDES.FLYING);
        });
    });

    describe('transitionToGround', () => {
        it('should transition to ground state', () => {
            const onStateChange = vi.fn();
            controller = new FPSTransitionController(
                mockScene as any,
                mockCamera as any,
                { onStateChange }
            );

            controller.transitionToGround(45, -122);

            expect(onStateChange).toHaveBeenCalledWith('ground');
        });

        it('should use provided lat/lon', () => {
            controller.transitionToGround(30, 60);

            const target = controller.getTarget();
            expect(target?.lat).toBe(30);
            expect(target?.lon).toBe(60);
            expect(target?.altitude).toBe(FPS_ALTITUDES.GROUND);
        });
    });

    describe('returnToOrbit', () => {
        it('should clear target and return to idle', () => {
            controller.transitionToFlying(45, -122);
            controller.returnToOrbit();

            expect(controller.getTarget()).toBeNull();
            expect(controller.getState()).toBe('idle');
        });
    });

    describe('handlePlanetClick', () => {
        it('should return false when no mesh is hit', () => {
            mockScene.pick.mockReturnValue({ hit: false });

            const result = controller.handlePlanetClick(100, 100);

            expect(result).toBe(false);
        });

        it('should return true when planet is hit and start transition', () => {
            mockScene.pick.mockReturnValue({
                hit: true,
                pickedPoint: { x: 1, y: 0, z: 0, length: () => 1 },
                pickedMesh: { name: 'globe_lod0' }
            });

            const result = controller.handlePlanetClick(100, 100);

            expect(result).toBe(true);
            expect(controller.getTarget()).not.toBeNull();
        });

        it('should ignore non-planet meshes', () => {
            mockScene.pick.mockReturnValue({
                hit: true,
                pickedPoint: { x: 1, y: 0, z: 0 },
                pickedMesh: { name: 'sun' }
            });

            const result = controller.handlePlanetClick(100, 100);

            expect(result).toBe(false);
        });
    });

    describe('altitude navigation', () => {
        it('should descend from flying to ground', () => {
            controller.transitionToFlying(45, -122);
            controller.descendToGround();

            expect(controller.getTarget()?.altitude).toBe(FPS_ALTITUDES.GROUND);
        });

        it('should ascend from ground to flying', () => {
            controller.transitionToGround(45, -122);
            controller.ascendToFlying();

            expect(controller.getTarget()?.altitude).toBe(FPS_ALTITUDES.FLYING);
        });
    });

    describe('FPS_ALTITUDES', () => {
        it('should export correct altitude constants', () => {
            expect(FPS_ALTITUDES.FLYING).toBeGreaterThan(FPS_ALTITUDES.LOW);
            expect(FPS_ALTITUDES.LOW).toBeGreaterThan(FPS_ALTITUDES.GROUND);
            expect(FPS_ALTITUDES.GROUND).toBeGreaterThan(0);
        });
    });
});
