/**
 * FPSAccessibilityOptions - Accessibility settings for FPS mode.
 * Phase 5: Production Readiness
 * 
 * Handles:
 * - FOV slider (motion sickness + preference)
 * - Head bob toggle
 * - Camera shake intensity
 * - Third-person camera option (stub)
 * - Motion blur toggle
 */

import type { UniversalCamera } from "@babylonjs/core/Cameras/universalCamera";

export interface AccessibilitySettings {
    fov: number;              // Field of view in degrees (60-110)
    headBobEnabled: boolean;  // Head bob during walking
    headBobIntensity: number; // 0-1 intensity multiplier
    cameraShake: number;      // 0-1 shake intensity
    motionBlur: boolean;      // Motion blur effect
    invertY: boolean;         // Invert Y axis
    mouseSensitivity: number; // Mouse look sensitivity
    reducedMotion: boolean;   // System preference for reduced motion
}

const DEFAULT_SETTINGS: AccessibilitySettings = {
    fov: 75,
    headBobEnabled: true,
    headBobIntensity: 0.5,
    cameraShake: 0.5,
    motionBlur: false,
    invertY: false,
    mouseSensitivity: 1.0,
    reducedMotion: false
};

const STORAGE_KEY = 'fps_accessibility_settings';

/**
 * Manages accessibility options for FPS mode.
 */
export class FPSAccessibilityOptions {
    private settings: AccessibilitySettings;
    private camera: UniversalCamera | null = null;
    private onSettingsChange?: (settings: AccessibilitySettings) => void;

    constructor() {
        this.settings = this.loadSettings();
        this.detectReducedMotion();
    }

    /**
     * Detect system preference for reduced motion.
     */
    private detectReducedMotion(): void {
        if (typeof window !== 'undefined' && window.matchMedia) {
            const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
            this.settings.reducedMotion = mediaQuery.matches;

            // Listen for changes
            mediaQuery.addEventListener('change', (e) => {
                this.settings.reducedMotion = e.matches;
                if (e.matches) {
                    // Auto-disable motion effects
                    this.settings.headBobEnabled = false;
                    this.settings.cameraShake = 0;
                    this.settings.motionBlur = false;
                    this.applySettings();
                }
            });
        }
    }

    /**
     * Load settings from localStorage.
     */
    private loadSettings(): AccessibilitySettings {
        if (typeof localStorage === 'undefined') {
            return { ...DEFAULT_SETTINGS };
        }

        try {
            const stored = localStorage.getItem(STORAGE_KEY);
            if (stored) {
                return { ...DEFAULT_SETTINGS, ...JSON.parse(stored) };
            }
        } catch (e) {
            console.warn('[Accessibility] Failed to load settings:', e);
        }

        return { ...DEFAULT_SETTINGS };
    }

    /**
     * Save settings to localStorage.
     */
    private saveSettings(): void {
        if (typeof localStorage === 'undefined') return;

        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(this.settings));
        } catch (e) {
            console.warn('[Accessibility] Failed to save settings:', e);
        }
    }

    /**
     * Set the camera to apply settings to.
     */
    setCamera(camera: UniversalCamera): void {
        this.camera = camera;
        this.applySettings();
    }

    /**
     * Set callback for settings changes.
     */
    setOnChange(callback: (settings: AccessibilitySettings) => void): void {
        this.onSettingsChange = callback;
    }

    /**
     * Apply current settings to camera.
     */
    private applySettings(): void {
        if (this.camera) {
            // FOV (convert degrees to radians for Babylon)
            this.camera.fov = this.settings.fov * (Math.PI / 180);

            // Mouse sensitivity
            this.camera.angularSensibility = 2000 / this.settings.mouseSensitivity;

            // Invert Y
            this.camera.invertRotation = this.settings.invertY;
        }

        this.onSettingsChange?.(this.settings);
    }

    /**
     * Get current settings.
     */
    getSettings(): AccessibilitySettings {
        return { ...this.settings };
    }

    /**
     * Set FOV (60-110 degrees).
     */
    setFOV(fov: number): void {
        this.settings.fov = Math.max(60, Math.min(110, fov));
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Get current FOV.
     */
    getFOV(): number {
        return this.settings.fov;
    }

    /**
     * Enable/disable head bob.
     */
    setHeadBob(enabled: boolean): void {
        this.settings.headBobEnabled = enabled;
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Set head bob intensity (0-1).
     */
    setHeadBobIntensity(intensity: number): void {
        this.settings.headBobIntensity = Math.max(0, Math.min(1, intensity));
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Check if head bob is enabled.
     */
    isHeadBobEnabled(): boolean {
        return this.settings.headBobEnabled && !this.settings.reducedMotion;
    }

    /**
     * Get head bob intensity.
     */
    getHeadBobIntensity(): number {
        return this.settings.headBobIntensity;
    }

    /**
     * Set camera shake intensity (0-1).
     */
    setCameraShake(intensity: number): void {
        this.settings.cameraShake = Math.max(0, Math.min(1, intensity));
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Enable/disable motion blur.
     */
    setMotionBlur(enabled: boolean): void {
        this.settings.motionBlur = enabled;
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Set Y-axis inversion.
     */
    setInvertY(inverted: boolean): void {
        this.settings.invertY = inverted;
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Set mouse sensitivity multiplier.
     */
    setMouseSensitivity(sensitivity: number): void {
        this.settings.mouseSensitivity = Math.max(0.1, Math.min(3.0, sensitivity));
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Reset all settings to defaults.
     */
    resetToDefaults(): void {
        this.settings = { ...DEFAULT_SETTINGS };
        this.detectReducedMotion();
        this.applySettings();
        this.saveSettings();
    }

    /**
     * Get head bob offset for current frame.
     * Returns Y offset to apply to camera position.
     */
    calculateHeadBob(walkSpeed: number, time: number): number {
        if (!this.isHeadBobEnabled()) {
            return 0;
        }

        // Bob frequency increases with speed
        const frequency = 2 + walkSpeed * 3;
        const amplitude = 0.02 * this.settings.headBobIntensity * walkSpeed;

        return Math.sin(time * frequency) * amplitude;
    }

    /**
     * Dispose of resources.
     */
    dispose(): void {
        this.camera = null;
    }
}
