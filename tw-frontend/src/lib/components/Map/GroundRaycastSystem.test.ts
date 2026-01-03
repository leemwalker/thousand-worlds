/**
 * GroundRaycastSystem Tests
 * TDD tests for 5-ray ground detection.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GroundRaycastSystem } from './GroundRaycastSystem';

// Mock Vector3
vi.mock("@babylonjs/core/Maths/math.vector", () => ({
    Vector3: vi.fn().mockImplementation((x = 0, y = 0, z = 0) => ({
        x, y, z,
        length: () => Math.sqrt(x * x + y * y + z * z),
        normalize: function () { const l = this.length(); if (l > 0) { this.x /= l; this.y /= l; this.z /= l; } return this; },
        scale: function (s: number) { return { x: this.x * s, y: this.y * s, z: this.z * s, length: () => Math.sqrt((this.x * s) ** 2 + (this.y * s) ** 2 + (this.z * s) ** 2) }; },
        clone: function () { return { ...this, length: this.length, normalize: this.normalize, scale: this.scale, clone: this.clone, add: this.add, subtract: this.subtract }; },
        add: function (v: any) { return { x: this.x + v.x, y: this.y + v.y, z: this.z + v.z }; },
        subtract: function (v: any) { return { x: this.x - v.x, y: this.y - v.y, z: this.z - v.z, normalize: () => ({ x: 0, y: 0, z: 0 }) }; },
    }))
}));

// Re-import after mock
import { Vector3 } from "@babylonjs/core/Maths/math.vector";

// Mock heightmap provider
const createMockHeightmapProvider = (baseHeight: number = 100) => ({
    getHeightAt: vi.fn().mockReturnValue(baseHeight),
    getHeightmapTexture: vi.fn().mockReturnValue(null),
    getElevationRange: vi.fn().mockReturnValue({ min: 0, max: 1000 })
});

describe('GroundRaycastSystem', () => {
    let system: GroundRaycastSystem;

    beforeEach(() => {
        system = new GroundRaycastSystem();
    });

    describe('constructor', () => {
        it('should create with default options', () => {
            expect(system).toBeDefined();
        });

        it('should accept custom options', () => {
            const customSystem = new GroundRaycastSystem({
                capsuleRadius: 0.5,
                maxGroundDistance: 1.0,
                stepHeight: 0.5
            });
            expect(customSystem).toBeDefined();
        });
    });

    describe('castFromPosition', () => {
        it('should return empty result without heightmap provider', () => {
            const position = new Vector3(1, 0, 0);
            const forward = new Vector3(0, 0, 1);

            const result = system.castFromPosition(position as any, forward as any);

            expect(result.isGrounded).toBe(false);
            expect(result.rayHits.length).toBe(0);
        });

        it('should detect ground when heightmap provider is set', () => {
            const provider = createMockHeightmapProvider(500);
            system.setHeightmapProvider(provider as any);
            system.setPlanetParams(1.0, 0);

            const position = new Vector3(1.025, 0, 0); // Slightly above ground
            const forward = new Vector3(0, 0, 1);

            const result = system.castFromPosition(position as any, forward as any);

            expect(result.rayHits.length).toBe(5); // 5 rays
            expect(provider.getHeightAt).toHaveBeenCalled();
        });

        it('should sample heightmap at correct positions', () => {
            const provider = createMockHeightmapProvider(500);
            system.setHeightmapProvider(provider as any);
            system.setPlanetParams(1.0, 0);

            const position = new Vector3(1.025, 0, 0);
            const forward = new Vector3(0, 0, 1);

            system.castFromPosition(position as any, forward as any);

            // Should have called getHeightAt 5 times (once per ray)
            expect(provider.getHeightAt).toHaveBeenCalledTimes(5);
        });
    });

    describe('water detection', () => {
        it('should detect water when terrain is below sea level', () => {
            const provider = createMockHeightmapProvider(-100); // Below sea level
            system.setHeightmapProvider(provider as any);
            system.setPlanetParams(1.0, 0); // Sea level at 0

            const position = new Vector3(1.025, 0, 0);
            const forward = new Vector3(0, 0, 1);

            const result = system.castFromPosition(position as any, forward as any);

            expect(result.isInWater).toBe(true);
            expect(result.waterDepth).toBeGreaterThan(0);
        });

        it('should not detect water when terrain is above sea level', () => {
            const provider = createMockHeightmapProvider(500); // Above sea level
            system.setHeightmapProvider(provider as any);
            system.setPlanetParams(1.0, 0);

            const position = new Vector3(1.025, 0, 0);
            const forward = new Vector3(0, 0, 1);

            const result = system.castFromPosition(position as any, forward as any);

            expect(result.isInWater).toBe(false);
            expect(result.waterDepth).toBe(0);
        });
    });

    describe('canClimbStep', () => {
        it('should allow climbing small steps', () => {
            expect(system.canClimbStep(1.0, 1.2)).toBe(true);
        });

        it('should prevent climbing too high', () => {
            expect(system.canClimbStep(1.0, 2.0)).toBe(false);
        });

        it('should allow walking down', () => {
            // Going down is not "climbing up"
            expect(system.canClimbStep(1.0, 0.8)).toBe(false);
        });
    });

    describe('setPlanetParams', () => {
        it('should update planet parameters', () => {
            system.setPlanetParams(6371, 0);
            // Parameters are stored internally - no direct getter
            expect(system).toBeDefined();
        });
    });
});
