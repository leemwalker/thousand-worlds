import type { Entity } from './game';
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
    | 'error';

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

export type ServerMessage =
    | GameOutputMessage
    | StateUpdateMessage
    | MapUpdateMessage
    | CombatEventMessage
    | ErrorMessage;
