import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import SimulationModeLayout from './SimulationModeLayout.svelte';
import { gameStore } from '$lib/stores/game';
import { uiState } from '$lib/stores/ui';
import { gameAPI } from '$lib/services/api';
import { writable } from 'svelte/store';

// Hoist component mocks
const hooks = vi.hoisted(() => {
    class MockSvelteComponent {
        $$ = {
            on_mount: [],
            on_destroy: [],
            before_update: [],
            after_update: [],
            callbacks: {},
            ctx: [],
            dirty: [],
            props: {}
        };
        $on() { }
        $set() { }
        $destroy() { }
    }
    return { MockSvelteComponent };
});

// Mock child components
vi.mock('$lib/components/Scene/SceneCanvas.svelte', () => ({
    default: hooks.MockSvelteComponent
}));
vi.mock('$lib/components/Layout/GameMenuModal.svelte', () => ({
    default: hooks.MockSvelteComponent
}));
vi.mock('$lib/components/Map/WorldController.svelte', () => ({
    default: hooks.MockSvelteComponent
}));
vi.mock('$lib/components/HUD/MessageOverlay.svelte', () => ({
    default: hooks.MockSvelteComponent
}));
vi.mock('$lib/components/Map/WorldCreationModal.svelte', () => ({
    default: hooks.MockSvelteComponent
}));

// Mock LobbyScene
vi.mock('$lib/components/Scene/LobbyScene', () => ({
    LobbyScene: class MockLobbyScene {
        setCallbacks = vi.fn();
        dispose = vi.fn();
        create = vi.fn();
    }
}));

// Mock SceneManager
vi.mock('$lib/components/Scene/SceneManager', () => ({
    sceneManager: {
        registerSceneFactory: vi.fn(),
        transitionTo: vi.fn(),
        getActiveScene: vi.fn(),
        getCurrentLocation: vi.fn(() => 'LOBBY')
    }
}));

// Mock services
vi.mock('$lib/services/websocket', async () => {
    const { writable } = await import('svelte/store');
    return {
        gameWebSocket: {
            connected: writable(true),
            sendRawCommand: vi.fn(),
            connect: vi.fn(),
            disconnect: vi.fn()
        }
    };
});

vi.mock('$lib/services/api', () => ({
    gameAPI: {
        logout: vi.fn()
    }
}));

describe('SimulationModeLayout', () => {
    beforeEach(() => {
        // Reset stores
        gameStore.setGameLocation('LOBBY');
        uiState.update(s => ({ ...s, layoutMode: 'desktop', interfaceMode: 'VISUAL' }));
    });

    it('renders the menu button', () => {
        render(SimulationModeLayout);
        expect(screen.getByText('Menu')).toBeTruthy();
    });

    it('toggles menu open state', async () => {
        const { getByText } = render(SimulationModeLayout);
        const menuBtn = getByText('Menu');
        await fireEvent.click(menuBtn);
    });

    it('shows mobile controls when layoutMode is mobile', async () => {
        uiState.update(s => ({ ...s, layoutMode: 'mobile' }));
        const { getByText } = render(SimulationModeLayout);
        expect(getByText('D-Pad')).toBeTruthy();
    });

    it('hides desktop stats when layoutMode is mobile', async () => {
        uiState.update(s => ({ ...s, layoutMode: 'mobile' }));
        const { queryByText } = render(SimulationModeLayout);
        expect(queryByText('Stats Overlay')).toBeNull();
    });

    it('shows desktop stats when layoutMode is desktop', async () => {
        uiState.update(s => ({ ...s, layoutMode: 'desktop' }));
        const { getByText } = render(SimulationModeLayout);
        expect(getByText('Stats Overlay')).toBeTruthy();
    });

    it('registers LobbyScene as factory on mount', async () => {
        const { LobbyScene } = await import('$lib/components/Scene/LobbyScene');
        const { sceneManager } = await import('$lib/components/Scene/SceneManager');

        render(SimulationModeLayout);

        expect(sceneManager.registerSceneFactory).toHaveBeenCalledWith('LOBBY', expect.any(Object));
    });
});

