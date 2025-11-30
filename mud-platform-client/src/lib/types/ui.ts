export type LayoutMode = 'mobile' | 'desktop';

export interface UIState {
    layoutMode: LayoutMode;
    screenWidth: number;
    activePanel: 'map' | 'stats' | 'inventory' | 'none'; // For mobile tabs if needed
    isSidebarOpen: boolean;
}

export interface TextSegment {
    text: string;
    color?: string;
    bold?: boolean;
    italic?: boolean;
    entityID?: string;
    entityType?: 'npc' | 'item' | 'location' | 'resource';
}

export interface GameMessage {
    id: string;
    type: 'movement' | 'area_description' | 'combat' | 'dialogue' | 'item_acquired' | 'crafting_success' | 'error' | 'system';
    text: string;
    timestamp: Date;
    // specialized fields
    direction?: string;
    entities?: string[]; // IDs
    targetName?: string;
    targetID?: string;
    damage?: number;
    speakerName?: string;
    speakerID?: string;
    itemName?: string;
    itemID?: string;
    itemRarity?: string;
    quality?: string;
    quantity?: number;
}

export interface InventoryItem {
    itemID: string;
    name: string;
    icon?: string;
    quality: 'poor' | 'common' | 'good' | 'excellent' | 'masterwork';
    quantity: number;
    weight: number;
    equipable: boolean;
    slot?: EquipmentSlot;
}

export type EquipmentSlot = 'head' | 'chest' | 'legs' | 'feet' | 'mainHand' | 'offHand';

export interface BehavioralBaseline {
    openness: number;
    conscientiousness: number;
    extraversion: number;
    agreeableness: number;
    neuroticism: number;
}

export interface DriftMetrics {
    baseline: BehavioralBaseline;
    current: BehavioralBaseline;
    driftLevel: 'none' | 'subtle' | 'moderate' | 'severe';
}
