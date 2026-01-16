/**
 * SceneManager.ts
 * Central coordinator for Babylon.js engine and scene management.
 * Owns the single Engine instance and handles scene transitions.
 */
import { Engine } from "@babylonjs/core/Engines/engine";
import { Scene } from "@babylonjs/core/scene";
import { Color4 } from "@babylonjs/core/Maths/math.color";

export type GameLocation = 'LOBBY' | 'WORLD' | 'LOADING' | 'PREVIEW' | 'MOON_FPV';

export interface SceneTransitionOptions {
    fadeDuration?: number; // milliseconds
    fadeColor?: Color4;
}

interface SceneFactory {
    create(scene: Scene): Promise<void>;
    dispose(): void;
}

/**
 * Manages the Babylon.js engine and coordinates scene switching.
 */
export class SceneManager {
    private engine: Engine | null = null;
    private canvas: HTMLCanvasElement | null = null;
    private scenes: Map<GameLocation, Scene> = new Map();
    private sceneFactories: Map<GameLocation, SceneFactory> = new Map();
    private currentLocation: GameLocation = 'LOADING';
    private isTransitioning: boolean = false;
    private renderLoopId: number | null = null;

    // Callbacks for state changes
    private onLocationChange: ((location: GameLocation) => void) | null = null;

    /**
     * Initialize the engine with a canvas element.
     */
    initialize(canvas: HTMLCanvasElement): void {
        if (this.engine) {
            console.warn('[SceneManager] Already initialized');
            return;
        }

        this.canvas = canvas;

        // Create engine with performance options
        this.engine = new Engine(canvas, true, {
            stencil: true,
            preserveDrawingBuffer: true,
            antialias: true,
            powerPreference: 'high-performance'
        });

        // Handle window resize
        window.addEventListener('resize', this.handleResize);

        console.log('[SceneManager] Engine initialized');
    }

    /**
     * Register a scene factory for a location.
     */
    registerSceneFactory(location: GameLocation, factory: SceneFactory): void {
        this.sceneFactories.set(location, factory);
    }

    /**
     * Create a new scene for a location.
     */
    async createScene(location: GameLocation): Promise<Scene> {
        if (!this.engine) {
            throw new Error('[SceneManager] Engine not initialized');
        }

        // Dispose existing scene for this location
        const existingScene = this.scenes.get(location);
        if (existingScene) {
            existingScene.dispose();
        }

        // Create new scene
        const scene = new Scene(this.engine);
        scene.clearColor = new Color4(0, 0, 0, 1);

        // Apply factory if registered
        const factory = this.sceneFactories.get(location);
        if (factory) {
            await factory.create(scene);
        }

        this.scenes.set(location, scene);
        return scene;
    }

    /**
     * Transition to a new location with optional fade effect.
     */
    async transitionTo(
        location: GameLocation,
        options: SceneTransitionOptions = {}
    ): Promise<void> {
        if (this.isTransitioning) {
            console.warn('[SceneManager] Transition already in progress');
            return;
        }

        if (!this.engine) {
            throw new Error('[SceneManager] Engine not initialized');
        }

        const { fadeDuration = 500 } = options;

        this.isTransitioning = true;
        console.log(`[SceneManager] Transitioning to ${location}`);

        try {
            // TODO: Fade out effect (can be added later)

            // Create scene if it doesn't exist
            let targetScene = this.scenes.get(location);
            if (!targetScene) {
                targetScene = await this.createScene(location);
            }

            // Switch active rendering
            this.currentLocation = location;

            // Stop previous render loop if running
            if (this.renderLoopId !== null) {
                this.engine.stopRenderLoop();
            }

            // Start render loop for new scene - use getActiveScene() to ensure
            // we render the same scene that WorldController receives
            this.engine.runRenderLoop(() => {
                const activeScene = this.scenes.get(this.currentLocation);
                activeScene?.render();
            });

            // Notify listeners
            this.onLocationChange?.(location);

            // TODO: Fade in effect

        } finally {
            this.isTransitioning = false;
        }
    }

    /**
     * Get the current active scene.
     */
    getActiveScene(): Scene | null {
        return this.scenes.get(this.currentLocation) ?? null;
    }

    /**
     * Get scene for a specific location.
     */
    getScene(location: GameLocation): Scene | null {
        return this.scenes.get(location) ?? null;
    }

    /**
     * Get the Babylon.js engine.
     */
    getEngine(): Engine | null {
        return this.engine;
    }

    /**
     * Get current location.
     */
    getCurrentLocation(): GameLocation {
        return this.currentLocation;
    }

    /**
     * Check if currently transitioning.
     */
    isInTransition(): boolean {
        return this.isTransitioning;
    }

    /**
     * Stop the current render loop.
     * Call this before disposing a scene to prevent "No camera defined" errors.
     */
    stopRenderLoop(): void {
        if (this.engine) {
            this.engine.stopRenderLoop();
            console.log('[SceneManager] Render loop stopped');
        }
    }

    /**
     * Set callback for location changes.
     */
    setOnLocationChange(callback: (location: GameLocation) => void): void {
        this.onLocationChange = callback;
    }

    /**
     * Handle window resize.
     */
    private handleResize = (): void => {
        this.engine?.resize();
    };

    /**
     * Dispose of all resources.
     */
    dispose(): void {
        console.log('[SceneManager] Disposing...');

        window.removeEventListener('resize', this.handleResize);

        // Dispose all scene factories
        for (const factory of this.sceneFactories.values()) {
            factory.dispose();
        }
        this.sceneFactories.clear();

        // Dispose all scenes
        for (const scene of this.scenes.values()) {
            scene.dispose();
        }
        this.scenes.clear();

        // Dispose engine
        this.engine?.dispose();
        this.engine = null;
        this.canvas = null;
    }
}

// Singleton instance
export const sceneManager = new SceneManager();
