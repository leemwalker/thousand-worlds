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
            expect(lodManager.getSegmentsForLevel(-1)).toBe(128); // Clamp to 0
            expect(lodManager.getSegmentsForLevel(10)).toBe(32);  // Clamp to max
        });
    });

    describe("shouldUpdate", () => {
        it("should return true when LOD level changes", () => {
            lodManager.setCurrentLevel(0);
            expect(lodManager.shouldUpdate(8)).toBe(true); // Distance 8 -> level 2
        });

        it("should return false when LOD level stays the same", () => {
            lodManager.setCurrentLevel(0);
            expect(lodManager.shouldUpdate(1.2)).toBe(false); // Still level 0
        });
    });

    describe("hysteresis", () => {
        it("should apply hysteresis to prevent rapid switching", () => {
            // Start at level 1 (medium distance)
            lodManager.setCurrentLevel(1);

            // Distance just barely crosses threshold - should NOT switch due to hysteresis
            // Threshold between 1->0 is distance 5, with 10% hysteresis = 4.5-5.5 range
            expect(lodManager.shouldUpdate(4.8)).toBe(false);

            // Distance clearly past threshold - should switch
            expect(lodManager.shouldUpdate(4.0)).toBe(true);
        });
    });
});
