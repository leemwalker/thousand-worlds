/**
 * WaterEffects Tests
 * TDD tests for water surface and underwater rendering.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock Babylon.js dependencies
vi.mock("@babylonjs/core/Maths/math.vector", () => ({
    Vector3: vi.fn().mockImplementation((x = 0, y = 0, z = 0) => ({
        x, y, z
    }))
}));

vi.mock("@babylonjs/core/Maths/math.color", () => ({
    Color3: vi.fn().mockImplementation((r = 0, g = 0, b = 0) => ({
        r, g, b
    }))
}));

vi.mock("@babylonjs/core/Materials/shaderMaterial", () => ({
    ShaderMaterial: vi.fn().mockImplementation(() => ({
        setFloat: vi.fn(),
        setVector3: vi.fn(),
        setColor3: vi.fn(),
        alpha: 1,
        backFaceCulling: true,
        dispose: vi.fn()
    }))
}));

vi.mock("@babylonjs/core/Materials/effect", () => ({
    Effect: { ShadersStore: {} }
}));

vi.mock("@babylonjs/core/PostProcesses/postProcess", () => ({
    PostProcess: vi.fn().mockImplementation(() => ({
        onApply: null,
        dispose: vi.fn()
    }))
}));

import { WaterEffects } from './WaterEffects';
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Color3 } from "@babylonjs/core/Maths/math.color";

const mockScene = {};

describe('WaterEffects', () => {
    let effects: WaterEffects;

    beforeEach(() => {
        vi.clearAllMocks();
        effects = new WaterEffects(mockScene as any);
    });

    describe('constructor', () => {
        it('should create with default options', () => {
            expect(effects).toBeDefined();
        });

        it('should accept custom options', () => {
            const customEffects = new WaterEffects(mockScene as any, {
                waterColor: new Color3(0.1, 0.2, 0.3) as any,
                transparency: 0.5,
                waveHeight: 0.01
            });
            expect(customEffects).toBeDefined();
        });
    });

    describe('createSurfaceMaterial', () => {
        it('should create water surface material', () => {
            const material = effects.createSurfaceMaterial();
            expect(material).toBeDefined();
        });
    });

    describe('getSurfaceMaterial', () => {
        it('should return null before creation', () => {
            expect(effects.getSurfaceMaterial()).toBeNull();
        });

        it('should return material after creation', () => {
            effects.createSurfaceMaterial();
            expect(effects.getSurfaceMaterial()).not.toBeNull();
        });
    });

    describe('update', () => {
        it('should update time and camera position', () => {
            effects.createSurfaceMaterial();
            const cameraPos = new Vector3(1, 2, 3) as any;

            expect(() => effects.update(0.016, cameraPos)).not.toThrow();
        });
    });

    describe('setUnderwater', () => {
        it('should toggle underwater mode', () => {
            expect(() => effects.setUnderwater(true, 5.0)).not.toThrow();
            expect(effects.getIsUnderwater()).toBe(true);

            effects.setUnderwater(false);
            expect(effects.getIsUnderwater()).toBe(false);
        });
    });

    describe('setLightDirection', () => {
        it('should update light direction', () => {
            effects.createSurfaceMaterial();
            const direction = new Vector3(1, 0.5, 0) as any;

            expect(() => effects.setLightDirection(direction)).not.toThrow();
        });
    });

    describe('dispose', () => {
        it('should dispose resources', () => {
            effects.createSurfaceMaterial();
            expect(() => effects.dispose()).not.toThrow();
            expect(effects.getSurfaceMaterial()).toBeNull();
        });
    });
});
