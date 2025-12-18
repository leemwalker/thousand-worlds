export interface User {
    user_id: string;
    email: string;
    username: string;
    created_at: string;
    last_login?: string;
    last_active_character_id?: string;
}

export interface CharacterAttributes {
    strength: number;
    dexterity: number;
    constitution: number;
    intelligence: number;
    wisdom: number;
    charisma: number;
    perception: number;
    willpower: number;
}

export interface Character {
    id: string;
    name: string;
    user_id: string;
    world_id: string;
    species: string;
    background: string;
    attributes: CharacterAttributes;
    created_at: string;
}

// Placeholder for full Item interface
export interface Item {
    id: string;
    name: string;
    description: string;
    weight: number;
    quantity: number;
}

export interface CharacterStats {
    hp: number;
    max_hp: number;
    stamina: number;
    max_stamina: number;
    mana: number;
    max_mana: number;
    level: number;
    xp: number;
}

// Placeholder for Entity interface from MapRenderer
export interface Entity {
    id: string;
    name: string;
    type: string;
    position: { x: number; y: number; z: number };
}

export interface GameMessage {
    type: string;
    timestamp: number;
    content: string;
    sender?: string | undefined;
    channel?: string | undefined;
}
