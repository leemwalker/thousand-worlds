/**
 * Tests for LODManager - handles Level of Detail mesh switching.
 * TDD Red Phase - write failing tests first.
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { LODManager, LODConfig } from "./LODManager";

describe("LODManager", () => {
    let lodManager: LODManager;
    const defaultConfig: LODConfig = {
        levels: [
            { distance: 10, segments: 128 },  // High detail when close
            { distance: 5, segments: 64 },    // Medium detail
            { distance: 2, segments: 32 },    // Low detail when far
        ],
    };

    beforeEach(() => {
        lodManager = new LODManager(defaultConfig);
    });

    describe("getLODLevel", () => {
        it("should return high LOD (0) when camera is very close", () => {
            const level = lodManager.getLODLevel(1.5);
            expect(level).toBe(0); // Highest detail
        });

        it("should return medium LOD (1) at medium distance", () => {
            const level = lodManager.getLODLevel(3.5);
            expect(level).toBe(1);
        });

        it("should return low LOD (2) when camera is far", () => {
            const level = lodManager.getLODLevel(8);
            expect(level).toBe(2);
        });

        it("should return lowest LOD when beyond max distance", () => {
            const level = lodManager.getLODLevel(100);
            expect(level).toBe(2); // Lowest detail level
        });
    });

    describe("getSegmentsForLevel", () => {
        it("should return correct segment count for each LOD level", () => {
            expect(lodManager.getSegmentsForLevel(0)).toBe(128);
            expect(lodManager.getSegmentsForLevel(1)).toBe(64);
            expect(lodManager.getSegmentsForLevel(2)).toBe(32);
        });

        it("should clamp to valid level range", () => {
            // After sorting by distance: [{dist:2, seg:32}, {dist:5, seg:64}, {dist:10, seg:128}]
            // So level 0 = 32 segments, level 2 = 128 segments
            expect(lodManager.getSegmentsForLevel(-1)).toBe(32); // Clamp to 0
            expect(lodManager.getSegmentsForLevel(10)).toBe(128);  // Clamp to max
        });
    });

    describe("shouldUpdate", () => {
        it("should return true when LOD level changes past hysteresis margin", () => {
            // Config sorted: [{dist:2}, {dist:5}, {dist:10}]
            // Level 0 if dist <2, level 1 if <5, level 2 if <10
            // Start at level 0, distance 12 -> level 2 (beyond max)
            // Hysteresis for level 2 threshold (10) = 1.0, so need >11 to trigger update
            lodManager.setCurrentLevel(0);
            expect(lodManager.shouldUpdate(12)).toBe(true); // Distance 12 > 11 (threshold + margin)
        });

        it("should return false when LOD level stays the same", () => {
            lodManager.setCurrentLevel(0);
            expect(lodManager.shouldUpdate(1.2)).toBe(false); // Distance 1.2 < 2 still level 0
        });
    });

    describe("hysteresis", () => {
        it("should apply hysteresis to prevent rapid switching", () => {
            // With distances sorted [2, 5, 10], level 0 is distance <2, level 1 is <5, level 2 is <10
            // Start at level 1 (distance 2-5 range)
            lodManager.setCurrentLevel(1);

            // Distance at 1.9 is in level 0 range - hysteresis should prevent switching
            // if distance is very close to boundary (threshold is 2, 10% hysteresis = 1.8)
            expect(lodManager.shouldUpdate(1.9)).toBe(false);

            // Distance clearly below threshold (with hysteresis margin)
            expect(lodManager.shouldUpdate(1.5)).toBe(true);
        });
    });
});
