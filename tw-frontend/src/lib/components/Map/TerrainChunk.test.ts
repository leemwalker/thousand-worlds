/**
 * TerrainChunk Tests - TDD Red Phase
 * Tests for terrain chunk management in first-person mode.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TerrainChunk } from './TerrainChunk';
import { ViewModeManager, getModeForDistance } from './ViewModeManager';
import { ViewMode, type ViewModeString } from './interfaces';

// Mock BabylonJS objects
const mockScene = {
    addMesh: vi.fn(),
    removeMesh: vi.fn()
};

const mockPosition = { x: 0, y: 0, z: 0 };

describe('TerrainChunk', () => {
    let chunk: TerrainChunk;

    beforeEach(() => {
        vi.clearAllMocks();
        chunk = new TerrainChunk(mockScene as any, mockPosition, 0);
    });

    describe('constructor', () => {
        it('should create a TerrainChunk with position and LOD level', () => {
            expect(chunk).toBeDefined();
            expect(chunk.getPosition()).toEqual(mockPosition);
            expect(chunk.getLODLevel()).toBe(0);
        });

        it('should accept custom size parameter', () => {
            const customChunk = new TerrainChunk(mockScene as any, mockPosition, 1, { size: 512 });
            expect(customChunk.getSize()).toBe(512);
        });
    });

    describe('ITerrainChunk interface', () => {
        it('should implement getMesh method', () => {
            expect(chunk.getMesh).toBeDefined();
            expect(typeof chunk.getMesh).toBe('function');
        });

        it('should implement getCollider method', () => {
            expect(chunk.getCollider).toBeDefined();
            expect(typeof chunk.getCollider).toBe('function');
        });

        it('should implement dispose method', () => {
            expect(chunk.dispose).toBeDefined();
            expect(typeof chunk.dispose).toBe('function');
        });

        it('should implement getPosition method', () => {
            expect(chunk.getPosition).toBeDefined();
            expect(typeof chunk.getPosition).toBe('function');
        });

        it('should implement getLODLevel method', () => {
            expect(chunk.getLODLevel).toBeDefined();
            expect(typeof chunk.getLODLevel).toBe('function');
        });
    });

    describe('getMesh', () => {
        it('should return a mesh or null before generation', () => {
            const mesh = chunk.getMesh();
            // Initially null until terrain data is provided
            expect(mesh === null || mesh !== undefined).toBe(true);
        });
    });

    describe('getCollider', () => {
        it('should return collider or null if physics not enabled', () => {
            const collider = chunk.getCollider();
            // Null until physics is enabled
            expect(collider).toBeNull();
        });
    });

    describe('dispose', () => {
        it('should clean up resources without error', () => {
            expect(() => chunk.dispose()).not.toThrow();
        });

        it('should mark chunk as disposed', () => {
            chunk.dispose();
            expect(chunk.isDisposed()).toBe(true);
        });
    });

    describe('generateFromHeightData', () => {
        it('should accept height data array', () => {
            const heightData = new Float32Array(256 * 256);
            expect(() => chunk.generateFromHeightData(heightData, 256, 256)).not.toThrow();
        });
    });
});

describe('ViewModeManager', () => {
    describe('getModeForDistance', () => {
        it('should return orbit mode for far distances', () => {
            const mode: ViewModeString = getModeForDistance(5.0);
            expect(mode).toBe('orbit');
        });

        it('should return terrain mode for close distances', () => {
            const mode: ViewModeString = getModeForDistance(1.1);
            expect(mode).toBe('terrain');
        });

        it('should return tile mode for medium distances', () => {
            const mode: ViewModeString = getModeForDistance(1.5);
            expect(mode).toBe('tile');
        });
    });

    describe('ViewModeManager class', () => {
        it('should start in orbit mode', () => {
            const manager = new ViewModeManager();
            expect(manager.getCurrentMode()).toBe(ViewMode.ORBIT);
        });

        it('should calculate mode correctly from manager', () => {
            const manager = new ViewModeManager();
            expect(manager.getModeForDistance(5.0)).toBe('orbit');
            expect(manager.getModeForDistance(1.1)).toBe('terrain');
            expect(manager.getModeForDistance(1.5)).toBe('tile');
        });
    });
});
