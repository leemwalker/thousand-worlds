import type { Entity } from './game';
import type { PointOfInterest } from './pois';
import type { VisibleTile, RenderQuality } from '$lib/components/Map/MapRenderer';

// --- Command Messages (Client -> Server) ---

export type CommandType = 'command' | 'interview_response';

export interface BaseCommand {
    type: CommandType;
    data: unknown;
}

export interface GameCommand extends BaseCommand {
    type: 'command';
    data: {
        text: string;
        payload?: any;
    };
}

export interface InterviewCommand extends BaseCommand {
    type: 'interview_response';
    data: {
        text: string;
    };
}

export type ClientMessage = GameCommand | InterviewCommand;


// --- Server Messages (Server -> Client) ---

export type ServerMessageType =
    | 'game_message'
    | 'state_update'
    | 'map_update'
    | 'combat_event'
    | 'error'
    | 'world_map_image_response'
    | 'world_tile_response'
    | 'world_reset'
    | 'points_of_interest'
    | 'satellites_info'
    | 'moon_destroyed'
    | 'asteroid_impact';

export interface SatellitesInfoMessage extends BaseServerMessage {
    type: 'satellites_info';
    data: {
        satellites: any[];
        rings: any;
    };
}

export interface MoonDestroyedMessage extends BaseServerMessage {
    type: 'moon_destroyed';
    data: {
        type: 'moon_destroyed';
        metadata: {
            moon_id: string;
            debris_mass: number;
        };
    };
}

export interface AsteroidImpactMessage extends BaseServerMessage {
    type: 'asteroid_impact';
    data: {
        type: 'asteroid_impact';
        metadata: {
            location: {
                lat: number;
                lon: number;
            };
            mass: number;
            impact_time: number;
            origin?: {
                distance: number;
                phi: number;
                theta: number;
            };
        };
    };
}

export interface BaseServerMessage {
    type: ServerMessageType;
    timestamp?: number;
}

export interface GameOutputMessage extends BaseServerMessage {
    type: 'game_message';
    data: {
        content: string;
        channel?: string;
        sender?: string;
        type?: string; // Sub-type like 'map_update' sometimes embedded
        metadata?: any; // For map updates embedded in game messages
    };
}

export interface StateUpdateMessage extends BaseServerMessage {
    type: 'state_update';
    data: {
        stats?: any;
        inventory?: any[];
        // Add other state fields as needed
    };
}

export interface MapUpdateMessage extends BaseServerMessage {
    type: 'map_update';
    data: {
        tiles: VisibleTile[];
        player_x: number;
        player_y: number;
        grid_size: number;
        world_id: string;
        render_quality?: RenderQuality;
    };
}

export interface CombatEventMessage extends BaseServerMessage {
    type: 'combat_event';
    data: {
        action: string;
        sourceId: string;
        targetId: string;
        result: any;
    };
}

export interface ErrorMessage extends BaseServerMessage {
    type: 'error';
    data: {
        code: string;
        message: string;
    };
}

export interface WorldResetMessage extends BaseServerMessage {
    type: 'world_reset';
    data: {
        text: string;
        message?: string;
    };
}

export interface PointsOfInterestMessage extends BaseServerMessage {
    type: 'points_of_interest';
    data: {
        pois: PointOfInterest[];
    };
}

// Fixed Interface Names
export interface WorldMapImageMessage extends BaseServerMessage {
    type: 'world_map_image_response';
    data: {
        width: number;
        height: number;
        channel?: number;
        imageBlob: Blob;
        gridData?: ArrayBuffer | null;
        heightmapBlob: Blob | null;
        materialBlob: Blob | null;
        iceBlob: Blob | null;
        normalMapBlob: Blob | null;

        // Stats
        sea_level?: number;
        max_elevation?: number;
        min_elevation?: number;

        // Metadata
        centerX?: number;
        centerY?: number;
        gridSize?: number;
        tiles?: any[];
    };
}

export interface WorldTileMessage extends BaseServerMessage {
    type: 'world_tile_response';
    data: {
        face: number;
        level: number;
        x: number;
        y: number;
        width: number;
        height: number;
        imageSize: number;
        heightmapSize: number;
        imageBytes: Uint8Array;
        heightmapBytes: Uint8Array;
    };
}

// Consolidated ServerMessage Type
export type ServerMessage =
    | GameOutputMessage
    | StateUpdateMessage
    | MapUpdateMessage
    | CombatEventMessage
    | ErrorMessage
    | WorldMapImageMessage
    | WorldTileMessage
    | WorldResetMessage
    | PointsOfInterestMessage
    | SatellitesInfoMessage
    | MoonDestroyedMessage
    | AsteroidImpactMessage;
