import { writable, derived, type Writable } from 'svelte/store';
import type { User, Character, Item, CharacterStats, Entity, GameMessage } from '$lib/types/game';

// Scene location types for visual mode
export type GameLocation = 'LOBBY' | 'WORLD' | 'LOADING';

interface GameState {
    user: User | null;
    currentCharacter: Character | null;
    messages: GameMessage[];
    inventory: Item[];
    stats: CharacterStats;
    nearbyEntities: Entity[];
    isLoading: boolean;
    error: string | null;
    // Scene management
    gameLocation: GameLocation;
    currentWorldId: string | null;
    world: WorldState;
}

export interface WorldState {
    textureBlob: Blob | null;
    heightmapBlob: Blob | null;
    materialBlob: Blob | null;
    iceBlob: Blob | null;
    geo: {
        seaLevel: number;
        maxElevation: number;
        minElevation: number;
    };
    sim: {
        satellites: any[];
    };
}

const initialWorldState: WorldState = {
    textureBlob: null,
    heightmapBlob: null,
    materialBlob: null,
    iceBlob: null,
    geo: { seaLevel: 0, maxElevation: 0, minElevation: 0 },
    sim: { satellites: [] }
};

const initialState: GameState = {
    user: null,
    currentCharacter: null,
    messages: [],
    inventory: [],
    stats: {
        hp: 100, max_hp: 100,
        stamina: 100, max_stamina: 100,
        mana: 100, max_mana: 100,
        level: 1, xp: 0
    },
    nearbyEntities: [],
    isLoading: false,
    error: null,
    // Scene management
    gameLocation: 'LOBBY',
    currentWorldId: null,
    world: initialWorldState
};

function createGameStore() {
    const { subscribe, update, set } = writable<GameState>(initialState);

    return {
        subscribe,
        setUser: (user: User) => update(s => ({ ...s, user })),
        clearUser: () => update(s => ({ ...s, user: null, currentCharacter: null })),

        setCharacter: (character: Character) => update(s => ({ ...s, currentCharacter: character })),

        addMessage: (message: GameMessage) => update(s => ({
            ...s,
            messages: [...s.messages, message].slice(-100) // Keep last 100
        })),

        setInventory: (items: Item[]) => update(s => ({ ...s, inventory: items })),
        updateStats: (stats: Partial<CharacterStats>) => update(s => ({
            ...s,
            stats: { ...s.stats, ...stats }
        })),

        setLoading: (isLoading: boolean) => update(s => ({ ...s, isLoading })),
        setError: (error: string | null) => update(s => ({ ...s, error })),

        // Scene location management
        setGameLocation: (location: GameLocation) => update(s => ({ ...s, gameLocation: location })),
        setCurrentWorldId: (worldId: string | null) => update(s => ({ ...s, currentWorldId: worldId })),
        enterWorld: (worldId: string) => update(s => ({ ...s, gameLocation: 'WORLD', currentWorldId: worldId })),
        returnToLobby: () => update(s => ({ ...s, gameLocation: 'LOBBY', currentWorldId: null })),

        // World State
        updateWorld: (worldState: Partial<WorldState>) => update(s => ({
            ...s,
            world: { ...s.world, ...worldState }
        })),

        reset: () => set(initialState)
    };
}

export const gameStore = createGameStore();
