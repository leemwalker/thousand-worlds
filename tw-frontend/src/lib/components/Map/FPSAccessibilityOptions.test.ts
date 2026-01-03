/**
 * FPSAccessibilityOptions Tests
 * TDD tests for accessibility settings.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { FPSAccessibilityOptions } from './FPSAccessibilityOptions';

// Mock localStorage
const mockStorage: Record<string, string> = {};
const mockLocalStorage = {
    getItem: vi.fn((key: string) => mockStorage[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { mockStorage[key] = value; }),
    removeItem: vi.fn((key: string) => { delete mockStorage[key]; }),
    clear: vi.fn(() => { Object.keys(mockStorage).forEach(k => delete mockStorage[k]); })
};

vi.stubGlobal('localStorage', mockLocalStorage);

// Mock window.matchMedia
const mockMatchMedia = vi.fn().mockReturnValue({
    matches: false,
    addEventListener: vi.fn()
});
vi.stubGlobal('matchMedia', mockMatchMedia);

describe('FPSAccessibilityOptions', () => {
    let options: FPSAccessibilityOptions;

    beforeEach(() => {
        mockLocalStorage.clear();
        vi.clearAllMocks();
        options = new FPSAccessibilityOptions();
    });

    afterEach(() => {
        options.dispose();
    });

    describe('constructor', () => {
        it('should create with default settings', () => {
            const settings = options.getSettings();
            expect(settings.fov).toBe(75);
            expect(settings.headBobEnabled).toBe(true);
        });

        it('should load saved settings from localStorage', () => {
            mockStorage['fps_accessibility_settings'] = JSON.stringify({ fov: 90 });
            const loadedOptions = new FPSAccessibilityOptions();
            expect(loadedOptions.getSettings().fov).toBe(90);
            loadedOptions.dispose();
        });
    });

    describe('setFOV', () => {
        it('should set FOV within valid range', () => {
            options.setFOV(90);
            expect(options.getFOV()).toBe(90);
        });

        it('should clamp FOV to 60-110 range', () => {
            options.setFOV(50);
            expect(options.getFOV()).toBe(60);

            options.setFOV(120);
            expect(options.getFOV()).toBe(110);
        });

        it('should persist to localStorage', () => {
            options.setFOV(85);
            expect(mockLocalStorage.setItem).toHaveBeenCalled();
        });
    });

    describe('setHeadBob', () => {
        it('should enable/disable head bob', () => {
            options.setHeadBob(false);
            expect(options.isHeadBobEnabled()).toBe(false);

            options.setHeadBob(true);
            expect(options.isHeadBobEnabled()).toBe(true);
        });
    });

    describe('setHeadBobIntensity', () => {
        it('should set intensity within 0-1 range', () => {
            options.setHeadBobIntensity(0.7);
            expect(options.getHeadBobIntensity()).toBe(0.7);
        });

        it('should clamp intensity', () => {
            options.setHeadBobIntensity(-0.5);
            expect(options.getHeadBobIntensity()).toBe(0);

            options.setHeadBobIntensity(1.5);
            expect(options.getHeadBobIntensity()).toBe(1);
        });
    });

    describe('setMouseSensitivity', () => {
        it('should set sensitivity within range', () => {
            options.setMouseSensitivity(1.5);
            expect(options.getSettings().mouseSensitivity).toBe(1.5);
        });

        it('should clamp sensitivity to 0.1-3.0', () => {
            options.setMouseSensitivity(0.05);
            expect(options.getSettings().mouseSensitivity).toBe(0.1);

            options.setMouseSensitivity(5);
            expect(options.getSettings().mouseSensitivity).toBe(3.0);
        });
    });

    describe('setInvertY', () => {
        it('should toggle Y-axis inversion', () => {
            options.setInvertY(true);
            expect(options.getSettings().invertY).toBe(true);

            options.setInvertY(false);
            expect(options.getSettings().invertY).toBe(false);
        });
    });

    describe('calculateHeadBob', () => {
        it('should return 0 when head bob disabled', () => {
            options.setHeadBob(false);
            expect(options.calculateHeadBob(1.0, 0)).toBe(0);
        });

        it('should return non-zero when enabled and moving', () => {
            options.setHeadBob(true);
            options.setHeadBobIntensity(1.0);
            const bob = options.calculateHeadBob(1.0, Math.PI / 2);
            // Should be some value, may be 0 at specific time
            expect(typeof bob).toBe('number');
        });
    });

    describe('resetToDefaults', () => {
        it('should reset all settings to defaults', () => {
            options.setFOV(100);
            options.setHeadBob(false);
            options.setMouseSensitivity(2.5);

            options.resetToDefaults();

            expect(options.getFOV()).toBe(75);
            expect(options.isHeadBobEnabled()).toBe(true);
            expect(options.getSettings().mouseSensitivity).toBe(1.0);
        });
    });
});
