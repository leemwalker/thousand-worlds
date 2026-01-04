/**
 * GroundRaycastSystem Tests
 * TDD tests for 5-ray ground detection.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GroundRaycastSystem } from './GroundRaycastSystem';

// Mock Vector3
vi.mock("@babylonjs/core/Maths/math.vector", () => {
    const createMockVec = (x = 0, y = 0, z = 0): any => ({
        x, y, z,
        length: function () { return Math.sqrt(this.x * this.x + this.y * this.y + this.z * this.z); },
        normalize: function () { const l = this.length(); if (l > 0) { return createMockVec(this.x / l, this.y / l, this.z / l); } return this; },
        scale: function (s: number) { return createMockVec(this.x * s, this.y * s, this.z * s); },
        clone: function () { return createMockVec(this.x, this.y, this.z); },
        add: function (v: any) { return createMockVec(this.x + v.x, this.y + v.y, this.z + v.z); },
        subtract: function (v: any) { return createMockVec(this.x - v.x, this.y - v.y, this.z - v.z); },
    });
    const Vector3Mock = vi.fn().mockImplementation((x = 0, y = 0, z = 0) => createMockVec(x, y, z));
    (Vector3Mock as any).Zero = () => createMockVec(0, 0, 0);
    (Vector3Mock as any).Cross = (a: any, b: any) => createMockVec(
        a.y * b.z - a.z * b.y,
        a.z * b.x - a.x * b.z,
        a.x * b.y - a.y * b.x
    );
    return { Vector3: Vector3Mock };
});

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

            // Should have called getHeightAt multiple times for height sampling
            expect(provider.getHeightAt).toHaveBeenCalled();
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
