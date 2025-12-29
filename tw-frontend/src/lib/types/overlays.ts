export type OverlayMode = 'none' | 'tectonics' | 'temp' | 'moisture' | 'elevation' | 'biome' | 'resources' | 'features';

export interface ResourceNode {
    type: string; // "gold", "iron", "cave", "peak", "volcano", "trench"
    x: number;
    y: number;
    val?: number; // Normalized value or abs height/depth
    data?: Record<string, any>; // Metadata for tooltips
}

export interface PlateMetadata {
    id: number;
    name: string;
    type: string;
    center_x: number;
    center_y: number;
    area: number;
}

export interface OverlayData {
    // Tectonics
    tectonics?: number[]; // Grid of Plate IDs
    plate_info?: PlateMetadata[];

    // Environment Heatmaps (0.0 - 1.0)
    temp?: number[];
    moisture?: number[];
    elevation?: number[]; // Projected elevation map

    // Biomes (ID grid)
    biome?: number[];

    // Sparse Resources
    resources?: ResourceNode[];
    minerals?: any[]; // Legacy mineral format if needed

    // Terrain Features
    features?: ResourceNode[];

    // Global Water Level (0.0 - 1.0)
    globalWaterLevel?: number;
}
