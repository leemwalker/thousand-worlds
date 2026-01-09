import { w as writable } from "./index.js";
import { z } from "zod";
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
const initialWorldState = {
  textureBlob: null,
  heightmapBlob: null,
  materialBlob: null,
  iceBlob: null,
  geo: { seaLevel: 0, maxElevation: 0, minElevation: 0 },
  sim: { satellites: [] }
};
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
  error: null,
  // Scene management
  gameLocation: "LOBBY",
  currentWorldId: null,
  world: initialWorldState
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
    // Scene location management
    setGameLocation: (location) => update((s) => ({ ...s, gameLocation: location })),
    setCurrentWorldId: (worldId) => update((s) => ({ ...s, currentWorldId: worldId })),
    enterWorld: (worldId) => update((s) => ({ ...s, gameLocation: "WORLD", currentWorldId: worldId })),
    returnToLobby: () => update((s) => ({ ...s, gameLocation: "LOBBY", currentWorldId: null })),
    // World State
    updateWorld: (worldState) => update((s) => ({
      ...s,
      world: { ...s.world, ...worldState }
    })),
    // Reset world to molten state (clears all textures, triggers molten planet shader)
    resetWorld: () => update((s) => ({
      ...s,
      world: initialWorldState
    })),
    reset: () => set(initialState)
  };
}
const gameStore = createGameStore();
const ServerMessageTypeSchema = z.enum([
  "game_message",
  "state_update",
  "map_update",
  "combat_event",
  "error",
  "world_map_image_response"
]);
const VisibleTileSchema = z.object({
  x: z.number(),
  y: z.number(),
  biome: z.string().optional(),
  biome_type: z.string().optional(),
  elevation: z.number().optional(),
  occluded: z.boolean().optional(),
  is_player: z.boolean().optional(),
  entities: z.array(z.any()).optional(),
  portal: z.any().optional()
});
z.object({
  tiles: z.array(VisibleTileSchema).optional(),
  cells: z.array(z.any()).optional(),
  // Backend may send as 'cells'
  player_x: z.number().optional(),
  player_y: z.number().optional(),
  grid_size: z.number().optional(),
  world_id: z.string().optional(),
  render_quality: z.string().optional(),
  // Simulation stats
  avg_temperature: z.number().optional(),
  max_elevation: z.number().optional(),
  sea_level: z.number().optional(),
  land_coverage: z.number().optional(),
  simulated_years: z.number().optional(),
  seed: z.number().optional()
});
z.object({
  content: z.string().optional(),
  text: z.string().optional(),
  channel: z.string().optional(),
  sender: z.string().optional(),
  type: z.string().optional(),
  metadata: z.any().optional()
});
z.object({
  stats: z.any().optional(),
  inventory: z.array(z.any()).optional()
});
z.object({
  action: z.string(),
  sourceId: z.string(),
  targetId: z.string(),
  result: z.any().optional()
});
z.object({
  code: z.string().optional(),
  message: z.string()
});
z.object({
  type: ServerMessageTypeSchema,
  timestamp: z.number().optional(),
  data: z.any()
  // Specific data validated per type
});
const EnvelopeSchema = z.object({
  type: z.string(),
  data: z.any()
});
function validateServerMessage(message) {
  {
    const envelope = EnvelopeSchema.safeParse(message);
    if (!envelope.success) {
      return { valid: false, errors: ["Invalid message envelope"] };
    }
    return { valid: true, data: message };
  }
}
export {
  gameStore as g,
  mapStore as m,
  validateServerMessage as v
};
