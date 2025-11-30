import { writable, derived } from 'svelte/store';
import type { UIState, GameMessage } from '../types/ui';

// Initial state
const initialUIState: UIState = {
    layoutMode: 'desktop',
    screenWidth: 1024,
    activePanel: 'none',
    isSidebarOpen: false
};

// Stores
export const uiState = writable<UIState>(initialUIState);

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
    gameOutput.update(messages => [...messages, message]);
}

export function clearGameOutput() {
    gameOutput.set([]);
}
