/**
 * AltitudeChunkManager Tests
 * TDD tests for altitude-based terrain chunk loading.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
    AltitudeChunkManager,
    getAltitudeStage,
    latLonToTileCoord,
    getTilesAround
} from './AltitudeChunkManager';

// Mock camera
const createMockCamera = (altitude: number) => ({
    position: { x: 1 + altitude, y: 0, z: 0 }
});

// Mock scene
const mockScene = {} as any;

describe('getAltitudeStage', () => {
    it('should return orbit for high altitudes', () => {
        expect(getAltitudeStage(1.0)).toBe('orbit');
        expect(getAltitudeStage(0.5)).toBe('orbit');
    });

    it('should return flying for medium-high altitudes', () => {
        expect(getAltitudeStage(0.15)).toBe('flying');
        expect(getAltitudeStage(0.1)).toBe('flying');
    });

    it('should return low for medium altitudes', () => {
        expect(getAltitudeStage(0.04)).toBe('low');
        expect(getAltitudeStage(0.03)).toBe('low');
    });

    it('should return ground for low altitudes', () => {
        expect(getAltitudeStage(0.01)).toBe('ground');
        expect(getAltitudeStage(0.005)).toBe('ground');
    });
});

describe('latLonToTileCoord', () => {
    it('should convert equator/prime meridian to tile coord', () => {
        const coord = latLonToTileCoord(0, 0, 2);
        expect(coord.level).toBe(2);
        expect(coord.face).toBeGreaterThanOrEqual(0);
        expect(coord.face).toBeLessThanOrEqual(5);
    });

    it('should produce different tiles for different locations', () => {
        const coord1 = latLonToTileCoord(45, 90, 2);
        const coord2 = latLonToTileCoord(-45, -90, 2);

        // Should be different tiles
        const key1 = `${coord1.face}_${coord1.x}_${coord1.y}`;
        const key2 = `${coord2.face}_${coord2.x}_${coord2.y}`;
        expect(key1).not.toBe(key2);
    });

    it('should return higher resolution tiles for higher levels', () => {
        const coordL1 = latLonToTileCoord(45, 90, 1);
        const coordL3 = latLonToTileCoord(45, 90, 3);

        expect(coordL1.level).toBe(1);
        expect(coordL3.level).toBe(3);
    });
});

describe('getTilesAround', () => {
    it('should return at least the center tile', () => {
        const tiles = getTilesAround(0, 0, 2, 0.01);
        expect(tiles.length).toBeGreaterThanOrEqual(1);
    });

    it('should return more tiles for larger radius', () => {
        const smallRadius = getTilesAround(0, 0, 2, 0.01);
        const largeRadius = getTilesAround(0, 0, 2, 0.1);

        expect(largeRadius.length).toBeGreaterThan(smallRadius.length);
    });

    it('should not return duplicate tiles', () => {
        const tiles = getTilesAround(45, 90, 2, 0.05);
        const keys = tiles.map(t => `${t.face}_${t.level}_${t.x}_${t.y}`);
        const uniqueKeys = new Set(keys);

        expect(keys.length).toBe(uniqueKeys.size);
    });
});

describe('AltitudeChunkManager', () => {
    let manager: AltitudeChunkManager;
    let chunkRequests: any[] = [];

    beforeEach(() => {
        chunkRequests = [];
        manager = new AltitudeChunkManager(mockScene, {
            onChunkRequest: (coords) => chunkRequests.push(coords)
        });
    });

    describe('constructor', () => {
        it('should initialize in orbit stage', () => {
            expect(manager.getCurrentStage()).toBe('orbit');
        });
    });

    describe('update', () => {
        it('should not request chunks in orbit stage', () => {
            manager.setTarget(45, 90);
            manager.update(createMockCamera(2.0) as any);

            expect(chunkRequests.length).toBe(0);
        });

        it('should request chunks when in flying stage', () => {
            manager.setTarget(45, 90);
            manager.update(createMockCamera(0.12) as any);

            expect(chunkRequests.length).toBeGreaterThan(0);
        });

        it('should trigger stage change callback', () => {
            const onStageChange = vi.fn();
            manager = new AltitudeChunkManager(mockScene, { onStageChange });

            manager.update(createMockCamera(0.12) as any);

            expect(onStageChange).toHaveBeenCalledWith('flying');
        });
    });

    describe('markLoaded', () => {
        it('should track loaded chunks', () => {
            manager.markLoaded({ face: 0, level: 2, x: 1, y: 1 });

            const stats = manager.getStats();
            expect(stats.loaded).toBe(1);
        });
    });

    describe('clearChunks', () => {
        it('should clear all loaded chunks', () => {
            manager.markLoaded({ face: 0, level: 2, x: 1, y: 1 });
            manager.markLoaded({ face: 0, level: 2, x: 1, y: 2 });

            manager.clearChunks();

            const stats = manager.getStats();
            expect(stats.loaded).toBe(0);
        });
    });
});
