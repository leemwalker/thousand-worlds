import { writable, derived } from 'svelte/store';
import type { UIState, GameMessage } from '../types/ui';
import { CircularBuffer } from '../utils/circular-buffer';

// Initial state
const initialUIState: UIState = {
    layoutMode: 'desktop',
    screenWidth: 1024,
    activePanel: 'none',
    isSidebarOpen: false
};

// Stores
export const uiState = writable<UIState>(initialUIState);

// Circular buffer for game output (prevents unbounded memory growth)
const gameOutputBuffer = new CircularBuffer<GameMessage>(1000);
export const gameOutput = writable<GameMessage[]>([]);

// Derived stores
export const isMobile = derived(uiState, ($ui) => $ui.layoutMode === 'mobile');

// Actions
export function setScreenWidth(width: number) {
    uiState.update(state => ({
        ...state,
        screenWidth: width,
        layoutMode: width < 769 ? 'mobile' : 'desktop'
    }));
}

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
