/**
 * TileManager Tests - TDD Red Phase
 * Tests for cube-face tile visibility culling and priority loading.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TileManager } from './TileManager';
import { CubeFace, TileCoord, ITileProvider } from './interfaces';

// Mock BabylonJS Camera
const mockCamera = {
    position: { x: 0, y: 0, z: 3 },
    getForwardRay: vi.fn().mockReturnValue({
        direction: { x: 0, y: 0, z: -1 }
    }),
    fov: Math.PI / 4,
    minZ: 0.1,
    maxZ: 100
};

// Mock Tile Provider
const mockTileProvider: ITileProvider = {
    getTile: vi.fn().mockResolvedValue({
        texture: {},
        heightmap: {}
    }),
    preloadTiles: vi.fn(),
    isCached: vi.fn().mockReturnValue(false)
};

describe('TileManager', () => {
    let tileManager: TileManager;

    beforeEach(() => {
        vi.clearAllMocks();
        tileManager = new TileManager(mockTileProvider);
    });

    describe('constructor', () => {
        it('should create a TileManager with default settings', () => {
            expect(tileManager).toBeDefined();
            expect(tileManager.getMaxLevel()).toBe(4);
        });

        it('should accept custom max level', () => {
            const customManager = new TileManager(mockTileProvider, { maxLevel: 6 });
            expect(customManager.getMaxLevel()).toBe(6);
        });
    });

    describe('getVisibleTiles', () => {
        it('should return tiles for all 6 faces at level 0', () => {
            const tiles = tileManager.getVisibleTiles(mockCamera as any, 0);

            expect(tiles.length).toBe(6);
            expect(tiles.some(t => t.face === CubeFace.FRONT)).toBe(true);
            expect(tiles.some(t => t.face === CubeFace.BACK)).toBe(true);
            expect(tiles.some(t => t.face === CubeFace.LEFT)).toBe(true);
            expect(tiles.some(t => t.face === CubeFace.RIGHT)).toBe(true);
            expect(tiles.some(t => t.face === CubeFace.TOP)).toBe(true);
            expect(tiles.some(t => t.face === CubeFace.BOTTOM)).toBe(true);
        });

        it('should return 4 tiles per face at level 1', () => {
            const tiles = tileManager.getVisibleTiles(mockCamera as any, 1);

            // 6 faces * 4 tiles = 24 tiles
            expect(tiles.length).toBe(24);
        });

        it('should filter tiles based on camera direction', () => {
            // Camera looking at front face, back face should be culled
            const frontCamera = {
                ...mockCamera,
                position: { x: 0, y: 0, z: 3 },
                getForwardRay: vi.fn().mockReturnValue({
                    direction: { x: 0, y: 0, z: -1 }
                })
            };

            const tiles = tileManager.getVisibleTilesFiltered(frontCamera as any, 0);

            // Should have fewer than 6 tiles (back face culled)
            expect(tiles.length).toBeLessThan(6);
            expect(tiles.some(t => t.face === CubeFace.FRONT)).toBe(true);
        });
    });

    describe('getPriorityQueue', () => {
        it('should order tiles by distance to camera', () => {
            const queue = tileManager.getPriorityQueue(mockCamera as any, 0);

            expect(queue.length).toBeGreaterThan(0);

            // First tile should be closest (likely front face for camera at z=3)
            for (let i = 0; i < queue.length - 1; i++) {
                expect(queue[i].distance).toBeLessThanOrEqual(queue[i + 1].distance);
            }
        });

        it('should include distance information', () => {
            const queue = tileManager.getPriorityQueue(mockCamera as any, 0);

            queue.forEach(item => {
                expect(item.coord).toBeDefined();
                expect(item.distance).toBeDefined();
                expect(typeof item.distance).toBe('number');
            });
        });
    });

    describe('calculateLevel', () => {
        it('should return level 0 for far camera', () => {
            const farCamera = { ...mockCamera, position: { x: 0, y: 0, z: 10 } };
            const level = tileManager.calculateLevel(farCamera as any);

            expect(level).toBe(0);
        });

        it('should return higher level for close camera', () => {
            const closeCamera = { ...mockCamera, position: { x: 0, y: 0, z: 1.2 } };
            const level = tileManager.calculateLevel(closeCamera as any);

            expect(level).toBeGreaterThan(0);
        });

        it('should cap level at maxLevel', () => {
            const veryCloseCamera = { ...mockCamera, position: { x: 0, y: 0, z: 1.01 } };
            const level = tileManager.calculateLevel(veryCloseCamera as any);

            expect(level).toBeLessThanOrEqual(tileManager.getMaxLevel());
        });
    });

    describe('getTileCenter', () => {
        it('should return 3D position for tile center', () => {
            const coord: TileCoord = { face: CubeFace.FRONT, level: 0, x: 0, y: 0 };
            const center = tileManager.getTileCenter(coord);

            expect(center).toHaveProperty('x');
            expect(center).toHaveProperty('y');
            expect(center).toHaveProperty('z');
        });

        it('should return correct center for front face', () => {
            const coord: TileCoord = { face: CubeFace.FRONT, level: 0, x: 0, y: 0 };
            const center = tileManager.getTileCenter(coord);

            // Front face center should be at z = 1 (unit sphere)
            expect(center.z).toBeCloseTo(1, 1);
        });
    });
});
