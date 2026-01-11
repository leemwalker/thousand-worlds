import { describe, it, expect } from 'vitest';
import {
    calculateSurfaceGravity,
    calculateHorizonDistance,
    getMoonFPVParams,
    G_CONSTANT,
    type MoonData
} from '$lib/types/moon';

/**
 * BDD Tests: Moon FPV Physics
 * 
 * These tests document user-observable behavior for moon surface FPV.
 * When a watcher stands on a moon, physics should feel appropriate
 * for that moon's size and mass.
 */

describe('Moon FPV Physics', () => {
    // Reference values for Earth's Moon
    const EARTH_MOON: MoonData = {
        id: 'moon-earth',
        name: "Earth's Moon",
        mass: 7.342e22,    // kg
        radius: 1.7374e6,  // meters
        distance: 384400000,
        period: 2360591,   // ~27.3 days in seconds
        color: '#AAAAAA',
        density: 3344,
        destroyed: false
    };

    // Small asteroid moon (Phobos-like)
    const PHOBOS_LIKE: MoonData = {
        id: 'moon-phobos',
        name: "Phobos",
        mass: 1.0659e16,   // kg
        radius: 11100,     // meters (~11km)
        distance: 9377000,
        period: 27553,     // ~7.7 hours
        color: '#8B7355',
        density: 1876,
        destroyed: false
    };

    // Large moon (Ganymede-like)
    const GANYMEDE_LIKE: MoonData = {
        id: 'moon-ganymede',
        name: "Ganymede",
        mass: 1.4819e23,   // kg
        radius: 2634100,   // meters
        distance: 1070400000,
        period: 618153,    // ~7.15 days
        color: '#B8A088',
        density: 1936,
        destroyed: false
    };

    // -------------------------------------------------------------------------
    // Scenario: Surface Gravity Feels Right for Moon Size
    // -------------------------------------------------------------------------
    // Given: A watcher enters FPV on a moon
    // When: They jump or move
    // Then: Gravity should match that moon's actual physics
    //   AND: Earth's Moon should feel ~1.62 m/s² (1/6 Earth)
    //   AND: Phobos should feel nearly weightless
    //   AND: Ganymede should feel similar to Earth's Moon

    describe('Surface Gravity Calculation', () => {
        it("Earth's Moon should have ~1.62 m/s² gravity", () => {
            const gravity = calculateSurfaceGravity(EARTH_MOON.mass, EARTH_MOON.radius);
            expect(gravity).toBeGreaterThan(1.5);
            expect(gravity).toBeLessThan(1.75);
        });

        it("Phobos should feel nearly weightless (~0.0057 m/s²)", () => {
            const gravity = calculateSurfaceGravity(PHOBOS_LIKE.mass, PHOBOS_LIKE.radius);
            expect(gravity).toBeGreaterThan(0.003);
            expect(gravity).toBeLessThan(0.01);
        });

        it("Ganymede should have ~1.43 m/s² gravity", () => {
            const gravity = calculateSurfaceGravity(GANYMEDE_LIKE.mass, GANYMEDE_LIKE.radius);
            expect(gravity).toBeGreaterThan(1.3);
            expect(gravity).toBeLessThan(1.6);
        });

        it("Zero mass or radius should return 0 (no crash)", () => {
            expect(calculateSurfaceGravity(0, 1000)).toBe(0);
            expect(calculateSurfaceGravity(1e22, 0)).toBe(0);
            expect(calculateSurfaceGravity(-1e22, 1000)).toBe(0);
        });
    });

    // -------------------------------------------------------------------------
    // Scenario: Horizon Distance Matches Body Size
    // -------------------------------------------------------------------------
    // Given: A watcher stands on a moon surface
    // When: They look toward the horizon
    // Then: Smaller moons should have closer horizons
    //   AND: Horizon should be calculable for fog/visibility

    describe('Horizon Distance Calculation', () => {
        const EYE_HEIGHT = 1.7; // meters

        it("Earth's Moon horizon should be ~2.4km at eye level", () => {
            const horizon = calculateHorizonDistance(EARTH_MOON.radius, EYE_HEIGHT);
            expect(horizon).toBeGreaterThan(2000);
            expect(horizon).toBeLessThan(3000);
        });

        it("Phobos horizon should be ~60 meters (very close)", () => {
            const horizon = calculateHorizonDistance(PHOBOS_LIKE.radius, EYE_HEIGHT);
            expect(horizon).toBeGreaterThan(40);
            expect(horizon).toBeLessThan(200);
        });

        it("Ganymede horizon should be ~3km", () => {
            const horizon = calculateHorizonDistance(GANYMEDE_LIKE.radius, EYE_HEIGHT);
            expect(horizon).toBeGreaterThan(2500);
            expect(horizon).toBeLessThan(4000);
        });

        it("Zero or negative values should return 0", () => {
            expect(calculateHorizonDistance(0, EYE_HEIGHT)).toBe(0);
            expect(calculateHorizonDistance(EARTH_MOON.radius, 0)).toBe(0);
            expect(calculateHorizonDistance(EARTH_MOON.radius, -1)).toBe(0);
        });
    });

    // -------------------------------------------------------------------------
    // Scenario: Movement Speed Scales with Gravity
    // -------------------------------------------------------------------------
    // Given: A watcher enters FPV on different moons
    // When: They move around
    // Then: Low gravity moons should allow faster movement
    //   AND: High gravity moons should feel heavier

    describe('Movement Parameters from Moon Data', () => {
        it("Earth's Moon should have moderate move speed", () => {
            const params = getMoonFPVParams(EARTH_MOON);
            expect(params.moveSpeed).toBeGreaterThan(5); // Boosted from base 5
            expect(params.moveSpeed).toBeLessThan(15);
        });

        it("Phobos should have fast movement (low gravity)", () => {
            const params = getMoonFPVParams(PHOBOS_LIKE);
            // Very low gravity = very fast (capped at 4x)
            expect(params.moveSpeed).toBe(20); // 5 * 4 (cap)
        });

        it("Ganymede should have similar speed to Earth's Moon", () => {
            const params = getMoonFPVParams(GANYMEDE_LIKE);
            expect(params.moveSpeed).toBeGreaterThan(10);
            expect(params.moveSpeed).toBeLessThan(18);
        });

        it("Jump height should scale with lower gravity", () => {
            const phobosParams = getMoonFPVParams(PHOBOS_LIKE);
            const moonParams = getMoonFPVParams(EARTH_MOON);

            // Phobos jump should be much higher than Moon jump
            expect(phobosParams.jumpHeight).toBeGreaterThan(moonParams.jumpHeight);
        });
    });

    // -------------------------------------------------------------------------
    // Scenario: FPV Parameters Include All Needed Info
    // -------------------------------------------------------------------------
    // Given: Moon data from backend
    // When: FPV params are calculated
    // Then: All display and physics info should be present

    describe('Complete FPV Parameter Generation', () => {
        it("should include all required fields", () => {
            const params = getMoonFPVParams(EARTH_MOON);

            expect(params.gravity).toBeDefined();
            expect(params.horizonDistance).toBeDefined();
            expect(params.moveSpeed).toBeDefined();
            expect(params.jumpHeight).toBeDefined();
            expect(params.moonName).toBe("Earth's Moon");
            expect(params.moonRadius).toBe(EARTH_MOON.radius);
            expect(params.moonColor).toBe('#AAAAAA');
        });

        it("should handle destroyed moons", () => {
            const destroyedMoon: MoonData = {
                ...EARTH_MOON,
                destroyed: true,
                destroyedAt: 4500000000
            };

            // Should still calculate physics (for viewing debris field)
            const params = getMoonFPVParams(destroyedMoon);
            expect(params.gravity).toBeGreaterThan(0);
        });
    });
});
