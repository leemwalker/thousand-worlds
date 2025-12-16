// Biome and catastrophe visual constants for minimap rendering
// Rendering is perception-based:
// - High perception (7+): Emoji at 50% opacity over biome Tailwind color
// - Low perception (<7): Char colored with biome hex over elevation square

export type BiomeType = 'OCEAN' | 'RAINFOREST' | 'GRASSLAND' | 'DECIDUOUS' | 'ALPINE' | 'TAIGA' | 'DESERT' | 'TUNDRA';
export type CatastropheType = 'VOLCANO' | 'ASTEROID' | 'FLOOD_BASALT' | 'ANOXIA' | 'ICE_AGE';
export type ElevationType = 'DEEP_OCEAN' | 'SHALLOW_WATER' | 'LOWLAND' | 'HIGHLAND' | 'PEAK';

export interface BiomeVisual {
    emoji: string;
    char: string;
    color: string;
    tailwind: string;
}

export interface ElevationVisual {
    name: string;
    color: string;
    tailwind: string;
}

export interface CatastropheVisual {
    emoji: string;
    char: string;
    color: string;
    tailwind: string;
}

export const BIOME_MAP: Record<string, BiomeVisual> = {
    ocean: { emoji: '🌊', char: '~', color: '#1d4ed8', tailwind: 'bg-blue-700 text-blue-200' },
    rainforest: { emoji: '🌴', char: '%', color: '#065f46', tailwind: 'bg-emerald-800 text-emerald-300' },
    grassland: { emoji: '🌾', char: '"', color: '#84cc16', tailwind: 'bg-lime-500 text-lime-900' },
    deciduous: { emoji: '🌳', char: '&', color: '#16a34a', tailwind: 'bg-green-600 text-green-100' },
    alpine: { emoji: '🏔️', char: '^', color: '#a8a29e', tailwind: 'bg-stone-400 text-stone-800' },
    taiga: { emoji: '🌲', char: '*', color: '#134e4a', tailwind: 'bg-teal-900 text-teal-200' },
    desert: { emoji: '🌵', char: '.', color: '#fcd34d', tailwind: 'bg-amber-300 text-amber-800' },
    tundra: { emoji: '❄️', char: '-', color: '#e2e8f0', tailwind: 'bg-slate-200 text-slate-600' }
};

export const ELEVATION_MAP: Record<string, ElevationVisual> = {
    deep_ocean: { name: 'Deep Ocean', color: '#1e3a5f', tailwind: 'bg-blue-900' },
    shallow_water: { name: 'Shallow Water', color: '#3b82f6', tailwind: 'bg-blue-500' },
    lowland: { name: 'Lowland', color: '#22c55e', tailwind: 'bg-green-500' },
    highland: { name: 'Highland', color: '#a16207', tailwind: 'bg-amber-700' },
    peak: { name: 'Peak', color: '#f5f5f4', tailwind: 'bg-stone-100' }
};

export const CATASTROPHE_MAP: Record<string, CatastropheVisual> = {
    volcano: { emoji: '🌋', char: 'A', color: '#dc2626', tailwind: 'bg-red-600 animate-pulse' },
    asteroid: { emoji: '☄️', char: '@', color: '#ea580c', tailwind: 'bg-orange-600 animate-bounce' },
    flood_basalt: { emoji: '♨️', char: '#', color: '#171717', tailwind: 'bg-neutral-900 text-red-500' },
    anoxia: { emoji: '🦠', char: '~', color: '#6b21a8', tailwind: 'bg-purple-800 text-purple-300' },
    ice_age: { emoji: '🧊', char: '=', color: '#cffafe', tailwind: 'bg-cyan-100 text-cyan-800' }
};

// Get biome visual by type
export function getBiomeVisual(biomeType: string): BiomeVisual {
    return BIOME_MAP[biomeType.toLowerCase()] || BIOME_MAP['grassland'];
}

// Get elevation visual by elevation value in meters
export function getElevationVisual(elevation: number): ElevationVisual {
    if (elevation <= -1000) return ELEVATION_MAP['deep_ocean'];
    if (elevation <= 0) return ELEVATION_MAP['shallow_water'];
    if (elevation <= 500) return ELEVATION_MAP['lowland'];
    if (elevation <= 2000) return ELEVATION_MAP['highland'];
    return ELEVATION_MAP['peak'];
}

// Get catastrophe visual by type
export function getCatastropheVisual(catastropheType: string): CatastropheVisual | null {
    return CATASTROPHE_MAP[catastropheType.toLowerCase()] || null;
}
