import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { GameWebSocket } from './websocket';
import { gameStore } from '$lib/stores/game';

// Mock the stores
vi.mock('$lib/stores/game', () => ({
    gameStore: {
        setLoading: vi.fn(),
        setError: vi.fn(),
        addMessage: vi.fn(),
        setInventory: vi.fn(),
        updateStats: vi.fn(),
    }
}));

vi.mock('$lib/stores/map', () => ({
    mapStore: {
        setMapData: vi.fn()
    }
}));

// Mock WebSocket class
class MockWebSocket {
    onopen: () => void = () => { };
    onmessage: (event: any) => void = () => { };
    onclose: () => void = () => { };
    onerror: (err: any) => void = () => { };
    send: (data: string) => void = vi.fn();
    close: () => void = vi.fn();
    readyState: number = WebSocket.OPEN;

    static OPEN = 1;
}

describe('GameWebSocket', () => {
    let wsService: GameWebSocket;
    let mockWs: MockWebSocket;

    beforeEach(() => {
        vi.unstubAllGlobals();
        mockWs = new MockWebSocket();
        vi.stubGlobal('WebSocket', vi.fn(() => mockWs));
        // Also stub window location
        vi.stubGlobal('window', {
            location: { protocol: 'http:', hostname: 'localhost', port: '5173' }
        });

        wsService = new GameWebSocket();
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    it('should connect and update store', () => {
        wsService.connect('char-123');

        expect(window.WebSocket).toHaveBeenCalled();

        // Simulate open
        mockWs.onopen();

        expect(gameStore.setLoading).toHaveBeenCalledWith(false);
    });

    it('should handle game messages', () => {
        wsService.connect();
        mockWs.onopen();

        const msg = {
            type: 'game_message',
            data: {
                content: 'Hello World',
                sender: 'System'
            }
        };

        mockWs.onmessage({ data: JSON.stringify(msg) });

        expect(gameStore.addMessage).toHaveBeenCalledWith(expect.objectContaining({
            content: 'Hello World',
            sender: 'System'
        }));
    });

    it('should handle map updates', () => {
        wsService.connect();
        mockWs.onopen();

        const msg = {
            type: 'map_update',
            data: { tiles: [] }
        };

        mockWs.onmessage({ data: JSON.stringify(msg) });

        // Indirectly check map store called (mocked)
        // We'd expect console logs or store updates
    });

    it('should handle errors', () => {
        wsService.connect();
        mockWs.onerror(new Error('Connection failed'));

        expect(gameStore.setError).toHaveBeenCalled();
    });
});
