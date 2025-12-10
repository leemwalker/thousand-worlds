import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MapRenderer } from './MapRenderer';
import type { Position, VisibleTile } from './MapRenderer';

describe('MapRenderer', () => {
    let canvas: HTMLCanvasElement;
    let renderer: MapRenderer;

    beforeEach(() => {
        // Create mock canvas
        canvas = document.createElement('canvas');
        canvas.width = 800;
        canvas.height = 600;
        renderer = new MapRenderer(canvas);
    });

    it('should initialize with canvas', () => {
        expect(renderer).toBeDefined();
    });

    it('should render player at center', () => {
        const playerPos: Position = { x: 0, y: 0 };
        const visibleTiles: VisibleTile[] = [];

        // This should not throw
        expect(() => renderer.render(playerPos, visibleTiles)).not.toThrow();
    });

    it('should render visible tiles', () => {
        const playerPos: Position = { x: 10, y: 10 };
        const visibleTiles: VisibleTile[] = [
            { x: 10, y: 10, biome: 'grassland', elevation: 100, entities: [] },
            { x: 11, y: 10, biome: 'forest', elevation: 200, entities: [] },
        ];

        expect(() => renderer.render(playerPos, visibleTiles)).not.toThrow();
    });

    it('should render entities on tiles', () => {
        const playerPos: Position = { x: 5, y: 5 };
        const visibleTiles: VisibleTile[] = [
            {
                x: 5,
                y: 5,
                biome: 'grassland',
                elevation: 100,
                entities: [
                    { id: '1', type: 'npc', x: 5, y: 5 },
                    { id: '2', type: 'resource', x: 5, y: 5 },
                ],
            },
        ];

        expect(() => renderer.render(playerPos, visibleTiles)).not.toThrow();
    });

    it('should apply fog of war to non-visible tiles', () => {
        const playerPos: Position = { x: 0, y: 0 };
        const visibleTiles: VisibleTile[] = [
            { x: 0, y: 0, biome: 'grassland', elevation: 100, entities: [] },
        ];

        // Should render fog around the visible tile
        expect(() => renderer.render(playerPos, visibleTiles)).not.toThrow();
    });
});
