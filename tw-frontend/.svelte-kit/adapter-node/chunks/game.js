import { w as writable } from "./index.js";
function createMapStore() {
  const { subscribe, set, update } = writable({
    data: null,
    lastUpdate: 0
  });
  return {
    subscribe,
    setMapData: (data) => {
      update((store) => ({
        ...store,
        data,
        lastUpdate: Date.now()
      }));
    },
    clear: () => set({
      data: null,
      lastUpdate: 0
    })
  };
}
const mapStore = createMapStore();
const initialState = {
  user: null,
  currentCharacter: null,
  messages: [],
  inventory: [],
  stats: {
    hp: 100,
    max_hp: 100,
    stamina: 100,
    max_stamina: 100,
    mana: 100,
    max_mana: 100,
    level: 1,
    xp: 0
  },
  nearbyEntities: [],
  isLoading: false,
  error: null
};
function createGameStore() {
  const { subscribe, update, set } = writable(initialState);
  return {
    subscribe,
    setUser: (user) => update((s) => ({ ...s, user })),
    clearUser: () => update((s) => ({ ...s, user: null, currentCharacter: null })),
    setCharacter: (character) => update((s) => ({ ...s, currentCharacter: character })),
    addMessage: (message) => update((s) => ({
      ...s,
      messages: [...s.messages, message].slice(-100)
      // Keep last 100
    })),
    setInventory: (items) => update((s) => ({ ...s, inventory: items })),
    updateStats: (stats) => update((s) => ({
      ...s,
      stats: { ...s.stats, ...stats }
    })),
    setLoading: (isLoading) => update((s) => ({ ...s, isLoading })),
    setError: (error) => update((s) => ({ ...s, error })),
    reset: () => set(initialState)
  };
}
createGameStore();
export {
  mapStore as m
};
