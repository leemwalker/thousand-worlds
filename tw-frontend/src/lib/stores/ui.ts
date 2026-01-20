import { writable, derived, get } from 'svelte/store';
import type { UIState, GameMessage, InterfaceMode } from '../types/ui';
import { CircularBuffer } from '../utils/circular-buffer';

// ============================================================================
// Device Detection
// ============================================================================

/**
 * Detect device type based on screen size, touch capability, and user agent.
 */
export function detectDeviceType(): 'mobile' | 'desktop' | 'tablet' {
    if (typeof window === 'undefined') return 'desktop'; // SSR fallback

    const hasTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    const isSmallScreen = window.innerWidth < 769;
    const isMobileUA = /Android|iPhone|iPad|iPod/i.test(navigator.userAgent);

    if (isSmallScreen && (hasTouch || isMobileUA)) return 'mobile';
    if (/iPad|Tablet/i.test(navigator.userAgent)) return 'tablet';
    return 'desktop';
}

/**
 * Get the default interface mode based on device type.
 * Mobile defaults to TEXT, Desktop to VISUAL.
 */
export function getDefaultInterfaceMode(): InterfaceMode {
    const device = detectDeviceType();
    return device === 'mobile' ? 'TEXT' : 'VISUAL';
}

/**
 * Load user's preferred mode from localStorage.
 */
function loadPreferredMode(): InterfaceMode | 'auto' {
    if (typeof localStorage === 'undefined') return 'auto';
    const saved = localStorage.getItem('tw-interface-mode');
    if (saved === 'TEXT' || saved === 'VISUAL') return saved;
    return 'auto';
}

/**
 * Determine initial interface mode from preference or device detection.
 */
function getInitialInterfaceMode(): InterfaceMode {
    const preferred = loadPreferredMode();
    if (preferred !== 'auto') return preferred;
    return getDefaultInterfaceMode();
}

// ============================================================================
// Initial State
// ============================================================================

const initialUIState: UIState = {
    layoutMode: 'desktop',
    screenWidth: 1024,
    interfaceMode: getInitialInterfaceMode(),
    preferredMode: loadPreferredMode(),
    activePanel: 'none',
    isSidebarOpen: false
};

// ============================================================================
// Stores
// ============================================================================

export const uiState = writable<UIState>(initialUIState);

// Circular buffer for game output (prevents unbounded memory growth)
const gameOutputBuffer = new CircularBuffer<GameMessage>(1000);
export const gameOutput = writable<GameMessage[]>([]);

// ============================================================================
// Derived Stores
// ============================================================================

export const isMobile = derived(uiState, ($ui) => $ui.layoutMode === 'mobile');

/** Current interface mode (TEXT or VISUAL) */
export const interfaceMode = derived(uiState, ($ui) => $ui.interfaceMode);

/** Whether we're in TEXT/TXT mode */
export const isTextMode = derived(uiState, ($ui) => $ui.interfaceMode === 'TEXT');

/** Whether we're in VISUAL/3D mode */
export const isVisualMode = derived(uiState, ($ui) => $ui.interfaceMode === 'VISUAL');

// ============================================================================
// Actions - Screen/Layout
// ============================================================================

export function setScreenWidth(width: number) {
    uiState.update(state => {
        const newLayoutMode = width < 769 ? 'mobile' : 'desktop';

        // If preferred mode is 'auto', update interface mode based on device
        let newInterfaceMode = state.interfaceMode;
        if (state.preferredMode === 'auto') {
            newInterfaceMode = newLayoutMode === 'mobile' ? 'TEXT' : 'VISUAL';
        }

        return {
            ...state,
            screenWidth: width,
            layoutMode: newLayoutMode,
            interfaceMode: newInterfaceMode
        };
    });
}

// ============================================================================
// Actions - Interface Mode
// ============================================================================

/**
 * Set the interface mode (TEXT or VISUAL).
 * Also updates the user's preferred mode preference.
 */
export function setInterfaceMode(mode: InterfaceMode) {
    uiState.update(state => ({
        ...state,
        interfaceMode: mode,
        preferredMode: mode // User explicitly chose, save preference
    }));

    // Persist to localStorage
    if (typeof localStorage !== 'undefined') {
        localStorage.setItem('tw-interface-mode', mode);
    }
}

/**
 * Toggle between TEXT and VISUAL modes.
 */
export function toggleInterfaceMode() {
    const current = get(uiState).interfaceMode;
    const newMode = current === 'TEXT' ? 'VISUAL' : 'TEXT';
    console.log('[Mode Toggle] Switching from', current, 'to', newMode);
    setInterfaceMode(newMode);
}

/**
 * Reset to auto-detection (device default).
 */
export function resetToAutoMode() {
    const defaultMode = getDefaultInterfaceMode();
    uiState.update(state => ({
        ...state,
        interfaceMode: defaultMode,
        preferredMode: 'auto'
    }));

    if (typeof localStorage !== 'undefined') {
        localStorage.removeItem('tw-interface-mode');
    }
}

// ============================================================================
// Actions - Game Output
// ============================================================================

export function addGameMessage(message: GameMessage) {
    gameOutputBuffer.push(message);
    gameOutput.set(gameOutputBuffer.getAll());
}

export function clearGameOutput() {
    gameOutputBuffer.clear();
    gameOutput.set([]);
}

// Get recent N messages for virtual scrolling
export function getRecentMessages(count: number): GameMessage[] {
    return gameOutputBuffer.getRecent(count);
}

// Get buffer stats for debugging/metrics
export function getBufferStats() {
    return {
        size: gameOutputBuffer.getSize(),
        capacity: gameOutputBuffer.getCapacity(),
        isFull: gameOutputBuffer.isFull()
    };
}

