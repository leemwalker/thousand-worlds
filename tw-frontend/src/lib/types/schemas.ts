/**
 * Zod schemas for runtime validation of WebSocket messages.
 * 
 * These schemas validate incoming server messages to catch
 * backend/frontend contract mismatches early.
 * 
 * WARNING: Keep in sync with Go struct definitions.
 * If backend changes, update these schemas accordingly.
 */

import { z } from 'zod';

// --- Base Schemas ---

export const ServerMessageTypeSchema = z.enum([
    'game_message',
    'state_update',
    'map_update',
    'combat_event',
    'error'
]);

// --- Map Update Schema ---

export const VisibleTileSchema = z.object({
    x: z.number(),
    y: z.number(),
    biome: z.string().optional(),
    biome_type: z.string().optional(),
    elevation: z.number().optional(),
    occluded: z.boolean().optional(),
    is_player: z.boolean().optional(),
    entities: z.array(z.any()).optional(),
    portal: z.any().optional(),
});

export const MapUpdateDataSchema = z.object({
    tiles: z.array(VisibleTileSchema).optional(),
    cells: z.array(z.any()).optional(), // Backend may send as 'cells'
    player_x: z.number().optional(),
    player_y: z.number().optional(),
    grid_size: z.number().optional(),
    world_id: z.string().optional(),
    render_quality: z.string().optional(),
});

// --- Game Message Schema ---

export const GameMessageDataSchema = z.object({
    content: z.string().optional(),
    text: z.string().optional(),
    channel: z.string().optional(),
    sender: z.string().optional(),
    type: z.string().optional(),
    metadata: z.any().optional(),
});

// --- State Update Schema ---

export const StateUpdateDataSchema = z.object({
    stats: z.any().optional(),
    inventory: z.array(z.any()).optional(),
});

// --- Combat Event Schema ---

export const CombatEventDataSchema = z.object({
    action: z.string(),
    sourceId: z.string(),
    targetId: z.string(),
    result: z.any().optional(),
});

// --- Error Schema ---

export const ErrorDataSchema = z.object({
    code: z.string().optional(),
    message: z.string(),
});

// --- Union Message Schema ---

export const ServerMessageSchema = z.object({
    type: ServerMessageTypeSchema,
    timestamp: z.number().optional(),
    data: z.any(), // Specific data validated per type
});

// --- Validation Helper ---

export type ValidationResult = {
    valid: boolean;
    errors?: string[];
    data?: unknown;
};

/**
 * Validates a server message and returns typed result.
 * Logs warnings for invalid messages but doesn't throw.
 */
export function validateServerMessage(message: unknown): ValidationResult {
    const result = ServerMessageSchema.safeParse(message);

    if (!result.success) {
        const errors = result.error.errors.map((e: { path: (string | number)[]; message: string }) =>
            `${e.path.join('.')}: ${e.message}`
        );
        console.warn('[WS Validation] Invalid message structure:', errors);
        return { valid: false, errors };
    }

    // Type-specific validation
    const msg = result.data;
    let dataResult;

    switch (msg.type) {
        case 'map_update':
            dataResult = MapUpdateDataSchema.safeParse(msg.data);
            break;
        case 'game_message':
            dataResult = GameMessageDataSchema.safeParse(msg.data);
            break;
        case 'state_update':
            dataResult = StateUpdateDataSchema.safeParse(msg.data);
            break;
        case 'combat_event':
            dataResult = CombatEventDataSchema.safeParse(msg.data);
            break;
        case 'error':
            dataResult = ErrorDataSchema.safeParse(msg.data);
            break;
        default:
            // Unknown type - log but don't fail
            console.warn(`[WS Validation] Unknown message type: ${msg.type}`);
            return { valid: true, data: msg };
    }

    if (!dataResult.success) {
        const errors = dataResult.error.errors.map((e: { path: (string | number)[]; message: string }) =>
            `data.${e.path.join('.')}: ${e.message}`
        );
        console.warn(`[WS Validation] Invalid ${msg.type} data:`, errors);
        return { valid: false, errors };
    }

    return { valid: true, data: msg };
}
