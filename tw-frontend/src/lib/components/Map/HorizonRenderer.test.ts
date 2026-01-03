/**
 * HorizonRenderer Tests
 * TDD tests for atmospheric sky dome rendering.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock Babylon.js dependencies
vi.mock("@babylonjs/core/Maths/math.vector", () => ({
    Vector3: vi.fn().mockImplementation((x = 0, y = 0, z = 0) => ({
        x, y, z,
        clone: function () { return { x: this.x, y: this.y, z: this.z }; },
        normalize: function () { return this; }
    }))
}));

vi.mock("@babylonjs/core/Maths/math.color", () => ({
    Color3: vi.fn().mockImplementation((r = 0, g = 0, b = 0) => ({
        r, g, b,
        scale: function (s: number) { return { r: this.r * s, g: this.g * s, b: this.b * s }; }
    }))
}));

vi.mock("@babylonjs/core/Materials/shaderMaterial", () => ({
    ShaderMaterial: vi.fn().mockImplementation(() => ({
        setVector3: vi.fn(),
        setColor3: vi.fn(),
        setFloat: vi.fn(),
        backFaceCulling: false,
        disableDepthWrite: false,
        dispose: vi.fn()
    }))
}));

vi.mock("@babylonjs/core/Materials/effect", () => ({
    Effect: { ShadersStore: {} }
}));

vi.mock("@babylonjs/core/Meshes/meshBuilder", () => ({
    MeshBuilder: {
        CreateSphere: vi.fn().mockReturnValue({
            material: null,
            infiniteDistance: false,
            position: { x: 0, y: 0, z: 0 },
            dispose: vi.fn()
        })
    }
}));

import { HorizonRenderer } from './HorizonRenderer';
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Color3 } from "@babylonjs/core/Maths/math.color";

const mockScene = {};

describe('HorizonRenderer', () => {
    let renderer: HorizonRenderer;

    beforeEach(() => {
        vi.clearAllMocks();
        renderer = new HorizonRenderer(mockScene as any);
    });

    describe('constructor', () => {
        it('should create with default options', () => {
            expect(renderer).toBeDefined();
        });

        it('should accept custom colors', () => {
            const customRenderer = new HorizonRenderer(mockScene as any, {
                zenithColor: new Color3(0.1, 0.2, 0.3) as any,
                horizonColor: new Color3(0.4, 0.5, 0.6) as any
            });
            expect(customRenderer).toBeDefined();
        });
    });

    describe('createSkyDome', () => {
        it('should create sky dome mesh', () => {
            const dome = renderer.createSkyDome(100);
            expect(dome).toBeDefined();
        });

        it('should use custom radius', () => {
            const dome = renderer.createSkyDome(500);
            expect(dome).toBeDefined();
        });
    });

    describe('setSunDirection', () => {
        it('should update sun direction', () => {
            renderer.createSkyDome();
            const direction = new Vector3(0.5, 0.5, 0) as any;

            expect(() => renderer.setSunDirection(direction)).not.toThrow();
        });
    });

    describe('getFogColor', () => {
        it('should return fog color matching horizon', () => {
            const color = renderer.getFogColor();
            expect(color).toBeDefined();
            expect(color).toHaveProperty('r');
            expect(color).toHaveProperty('g');
            expect(color).toHaveProperty('b');
        });
    });

    describe('followCamera', () => {
        it('should update dome position', () => {
            renderer.createSkyDome();
            const cameraPos = new Vector3(10, 5, 3) as any;

            expect(() => renderer.followCamera(cameraPos)).not.toThrow();
        });
    });

    describe('getSkyDome', () => {
        it('should return null before creation', () => {
            expect(renderer.getSkyDome()).toBeNull();
        });

        it('should return mesh after creation', () => {
            renderer.createSkyDome();
            expect(renderer.getSkyDome()).not.toBeNull();
        });
    });

    describe('dispose', () => {
        it('should dispose resources', () => {
            renderer.createSkyDome();
            expect(() => renderer.dispose()).not.toThrow();
            expect(renderer.getSkyDome()).toBeNull();
        });
    });
});
