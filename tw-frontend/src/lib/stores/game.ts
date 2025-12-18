import { writable, type Writable } from 'svelte/store';
import type { User, Character, Item, CharacterStats, Entity, GameMessage } from '$lib/types/game';

interface GameState {
    user: User | null;
    currentCharacter: Character | null;
    messages: GameMessage[];
    inventory: Item[];
    stats: CharacterStats;
    nearbyEntities: Entity[];
    isLoading: boolean;
    error: string | null;
}

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
    error: null
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

        reset: () => set(initialState)
    };
}

export const gameStore = createGameStore();
