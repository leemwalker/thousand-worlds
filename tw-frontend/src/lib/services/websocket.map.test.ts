import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { GameWebSocket } from './websocket';
import { mapStore } from '$lib/stores/map';

// Mock stores
vi.mock('$lib/stores/map', () => ({
    mapStore: {
        setMapData: vi.fn()
    }
}));

vi.mock('$lib/stores/game', () => ({
    gameStore: {
        setLoading: vi.fn(),
        setError: vi.fn(),
        addMessage: vi.fn(),
        setInventory: vi.fn(),
        updateStats: vi.fn(),
    }
}));

// Mock WebSocket
class MockWebSocket {
    onopen: () => void = () => { };
    onmessage: (event: any) => void = () => { };
    onclose: () => void = () => { };
    onerror: (err: any) => void = () => { };
    send: (data: string) => void = vi.fn();
    close: () => void = vi.fn();
    readyState: number = WebSocket.OPEN;
}

describe('GameWebSocket Map Data Transformation', () => {
    let wsService: GameWebSocket;
    let mockWs: MockWebSocket;

    beforeEach(() => {
        vi.unstubAllGlobals();
        mockWs = new MockWebSocket();
        vi.stubGlobal('WebSocket', vi.fn(() => mockWs));
        vi.stubGlobal('window', {
            location: { protocol: 'http:', hostname: 'localhost', port: '5173' }
        });

        wsService = new GameWebSocket();
        wsService.connect();
        mockWs.onopen();
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    it('should correctly transform backend map format to frontend format', () => {
        // Sample data from backend (snake_case, q/r)
        const backendData = {
            world_id: "world-123",
            year: 100,
            cells: [
                {
                    q: 10,
                    r: 20,
                    biome_type: "forest",
                    elevation: 100,
                    biome_emoji: "🌲",
                    elev_name: "lowland",
                    // Other visual fields omitted for brevity
                },
                {
                    q: 11,
                    r: 20,
                    biome_type: "ocean",
                    elevation: 0,
                    catastrophe: "volcano"
                }
            ]
        };

        const msg = {
            type: 'map_update',
            data: backendData
        };

        mockWs.onmessage({ data: JSON.stringify(msg) });

        // Verify mapStore.setMapData was called with TRANSFORMED data
        expect(mapStore.setMapData).toHaveBeenCalledWith(expect.objectContaining({
            world_id: "world-123",
            // Check tiles transformation
            tiles: expect.arrayContaining([
                expect.objectContaining({
                    x: 10,
                    y: 20,
                    biome: "forest",
                    elevation: 100
                }),
                expect.objectContaining({
                    x: 11,
                    y: 20,
                    biome: "ocean",
                    elevation: 0,
                    // Check if extra props are preserved/mapped if needed, 
                    // though frontend doesn't strictly need 'catastrophe' it might need 'occluded'
                })
            ])
        }));
    });

    it('should handle legacy/already transformed data gracefully (backwards compat)', () => {
        const legacyData = {
            world_id: "world-old",
            tiles: [
                { x: 5, y: 5, biome: "desert", elevation: 50 }
            ]
        };

        const msg = {
            type: 'map_update',
            data: legacyData
        };

        mockWs.onmessage({ data: JSON.stringify(msg) });

        expect(mapStore.setMapData).toHaveBeenCalledWith(expect.objectContaining({
            world_id: "world-old",
            tiles: expect.arrayContaining([
                { x: 5, y: 5, biome: "desert", elevation: 50 }
            ])
        }));
    });
});
