/**
 * GameSystem.ts
 * Central command routing service - the "bridge" between input sources and outputs.
 * 
 * Accepts inputs from:
 * - CommandInput (typed text commands)
 * - Keyboard (WASD/Arrow keys in 3D mode)
 * - DPad (mobile touch controls)
 * 
 * Distributes to:
 * - WebSocket (backend communication)
 * - Text Log (gameOutput store)
 * - FPS Controller (3D movement when in VISUAL mode)
 */

import { get } from 'svelte/store';
import { gameWebSocket } from './websocket';
import { addGameMessage, interfaceMode } from '$lib/stores/ui';
import type { InterfaceMode } from '$lib/types/ui';
import type { FirstPersonController } from '$lib/components/Map/FirstPersonController';

// Movement direction aliases
const MOVEMENT_ALIASES: Record<string, string> = {
    'n': 'north',
    's': 'south',
    'e': 'east',
    'w': 'west',
    'u': 'up',
    'd': 'down',
    'ne': 'northeast',
    'nw': 'northwest',
    'se': 'southeast',
    'sw': 'southwest',
};

// Valid movement directions
const VALID_DIRECTIONS = new Set([
    'north', 'south', 'east', 'west',
    'up', 'down',
    'northeast', 'northwest', 'southeast', 'southwest',
]);

/**
 * Parse a command to check if it's a movement command.
 * Returns the normalized direction or null if not a movement command.
 */
function parseMovementCommand(command: string): string | null {
    const trimmed = command.trim().toLowerCase();

    // Check for aliases (single letter shortcuts)
    const aliasResult = MOVEMENT_ALIASES[trimmed];
    if (aliasResult) {
        return aliasResult;
    }

    // Check for direct direction
    if (VALID_DIRECTIONS.has(trimmed)) {
        return trimmed;
    }

    // Check for "go <direction>" format
    const goMatch = trimmed.match(/^go\s+(.+)$/);
    if (goMatch && goMatch[1]) {
        const dir = goMatch[1];
        if (VALID_DIRECTIONS.has(dir)) return dir;
        const aliasDir = MOVEMENT_ALIASES[dir];
        if (aliasDir) return aliasDir;
    }

    // Check for "move <direction>" format
    const moveMatch = trimmed.match(/^move\s+(.+)$/);
    if (moveMatch && moveMatch[1]) {
        const dir = moveMatch[1];
        if (VALID_DIRECTIONS.has(dir)) return dir;
        const aliasDir = MOVEMENT_ALIASES[dir];
        if (aliasDir) return aliasDir;
    }

    return null;
}

/**
 * GameSystem - Central command routing service.
 */
class GameSystem {
    private fpsController: FirstPersonController | null = null;
    private currentMode: InterfaceMode = 'TEXT';
    private unsubscribe: (() => void) | null = null;

    constructor() {
        // Subscribe to mode changes
        this.unsubscribe = interfaceMode.subscribe((mode) => {
            this.currentMode = mode;
        });
    }

    /**
     * Process a typed command (from CommandInput).
     * Routes to appropriate handlers based on command type.
     */
    processCommand(command: string): void {
        if (!command.trim()) return;

        // Check if it's a movement command
        const direction = parseMovementCommand(command);
        if (direction) {
            this.processMovement(direction);
            return;
        }

        // For all other commands, send to backend
        gameWebSocket.sendRawCommand(command);
    }

    /**
     * Process a movement command (from text, keyboard, or DPad).
     * @param direction - Normalized direction (north, south, etc.)
     */
    processMovement(direction: string): void {
        // 1. Send to backend
        gameWebSocket.sendRawCommand(`go ${direction}`);

        // 2. If in VISUAL mode with active FPS controller, trigger 3D movement
        if (this.currentMode === 'VISUAL' && this.fpsController) {
            this.triggerFPSMovement(direction);
        }

        // 3. Log to text output (useful for both modes)
        this.logMovement(direction);
    }

    /**
     * Trigger FPS controller movement based on direction.
     */
    private triggerFPSMovement(direction: string): void {
        if (!this.fpsController) return;

        // Get position for confirmation (movement is server-authoritative)
        const pos = this.fpsController.getPosition();
        console.log(`[GameSystem] FPS movement ${direction} from (${pos.lat.toFixed(2)}, ${pos.lon.toFixed(2)})`);

        // Note: Actual position update comes from server response
        // This is just for local feedback/animation
    }

    /**
     * Log a movement action to the game output.
     */
    private logMovement(direction: string): void {
        addGameMessage({
            id: crypto.randomUUID(),
            type: 'movement',
            text: `You head ${direction}.`,
            timestamp: new Date(),
            direction: direction,
        });
    }

    /**
     * Handle FPS controller position updates.
     * Called when player moves in 3D space (e.g., from mouse look + WASD).
     * Generates text log entries for position changes.
     */
    onFPSPositionUpdate(newPosition: { lat: number; lon: number; altitude: number }): void {
        // Only generate log if in VISUAL mode
        if (this.currentMode !== 'VISUAL') return;

        addGameMessage({
            id: crypto.randomUUID(),
            type: 'movement',
            text: `You are at coordinates (${newPosition.lat.toFixed(2)}°, ${newPosition.lon.toFixed(2)}°)`,
            timestamp: new Date(),
        });
    }

    /**
     * Process a keyboard key press (from global keyboard handler).
     * Returns true if the key was handled.
     */
    processKeyPress(key: string, shiftKey: boolean = false): boolean {
        // Only handle movement keys in VISUAL mode
        if (this.currentMode !== 'VISUAL') return false;

        let direction: string | null = null;

        switch (key.toLowerCase()) {
            case 'w':
            case 'arrowup':
                direction = 'north';
                break;
            case 's':
            case 'arrowdown':
                direction = 'south';
                break;
            case 'a':
            case 'arrowleft':
                direction = 'west';
                break;
            case 'd':
            case 'arrowright':
                direction = 'east';
                break;
        }

        if (direction) {
            this.processMovement(direction);
            return true;
        }

        return false;
    }

    /**
     * Attach the FPS controller for 3D movement synchronization.
     */
    setFPSController(controller: FirstPersonController | null): void {
        this.fpsController = controller;
    }

    /**
     * Get the current interface mode.
     */
    getMode(): InterfaceMode {
        return this.currentMode;
    }

    /**
     * Clean up subscriptions.
     */
    dispose(): void {
        if (this.unsubscribe) {
            this.unsubscribe();
            this.unsubscribe = null;
        }
        this.fpsController = null;
    }
}

// Singleton instance
export const gameSystem = new GameSystem();

// Export the class for testing
export { GameSystem, parseMovementCommand };
