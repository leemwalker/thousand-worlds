/**
 * Binary Grid Data Parser (Sprint 2)
 * 
 * Parses the binary grid format sent from the backend for precise map tooltips.
 * 
 * Binary Format:
 *   Header: 9 bytes
 *     - Magic: "WMAP" (4 bytes)
 *     - Version: 1 (uint8)
 *     - Width: uint16 little-endian
 *     - Height: uint16 little-endian
 *   
 *   Elevation Section: width * height * 4 bytes (float32 LE)
 *   Biome Section: width * height * 1 byte (uint8 BiomeID)
 */

// BiomeID maps to backend gamemap.BiomeID constants
export enum BiomeID {
    Unknown = 0,
    Ocean = 1,
    Lowland = 2,
    Highland = 3,
    Mountain = 4,
    HighMountain = 5,
    Rainforest = 6,
    Desert = 7,
    Grassland = 8,
    DeciduousForest = 9,
    Taiga = 10,
    Tundra = 11,
    Alpine = 12,
    Lake = 13,
    Wetland = 14,
}

// BiomeID to display name mapping
const BIOME_NAMES: Record<BiomeID, string> = {
    [BiomeID.Unknown]: 'Unknown',
    [BiomeID.Ocean]: 'Ocean',
    [BiomeID.Lowland]: 'Lowland',
    [BiomeID.Highland]: 'Highland',
    [BiomeID.Mountain]: 'Mountain',
    [BiomeID.HighMountain]: 'High Mountain',
    [BiomeID.Rainforest]: 'Rainforest',
    [BiomeID.Desert]: 'Desert',
    [BiomeID.Grassland]: 'Grassland',
    [BiomeID.DeciduousForest]: 'Deciduous Forest',
    [BiomeID.Taiga]: 'Taiga',
    [BiomeID.Tundra]: 'Tundra',
    [BiomeID.Alpine]: 'Alpine',
    [BiomeID.Lake]: 'Lake',
    [BiomeID.Wetland]: 'Wetland',
};

export function getBiomeName(id: BiomeID): string {
    return BIOME_NAMES[id] ?? 'Unknown';
}

export interface ParsedGridData {
    width: number;
    height: number;
    elevations: Float32Array;
    biomes: Uint8Array;
}

const HEADER_SIZE = 9;
const MAGIC = 'WMAP';

/**
 * Parse binary grid data from backend
 */
export function parseBinaryGridData(buffer: ArrayBuffer): ParsedGridData | null {
    if (buffer.byteLength < HEADER_SIZE) {
        console.warn('[BinaryDataParser] Data too short:', buffer.byteLength);
        return null;
    }

    const view = new DataView(buffer);

    // Verify magic bytes
    const magic = String.fromCharCode(
        view.getUint8(0),
        view.getUint8(1),
        view.getUint8(2),
        view.getUint8(3)
    );

    if (magic !== MAGIC) {
        console.warn('[BinaryDataParser] Invalid magic:', magic);
        return null;
    }

    // Read version
    const version = view.getUint8(4);
    if (version !== 1) {
        console.warn('[BinaryDataParser] Unsupported version:', version);
        return null;
    }

    // Read dimensions (little-endian)
    const width = view.getUint16(5, true);
    const height = view.getUint16(7, true);
    const size = width * height;

    // Validate data size
    const expectedSize = HEADER_SIZE + size * 4 + size;
    if (buffer.byteLength < expectedSize) {
        console.warn('[BinaryDataParser] Truncated data:', buffer.byteLength, 'expected:', expectedSize);
        return null;
    }

    // Extract elevation data (float32 array starting at offset 9)
    const elevationOffset = HEADER_SIZE;
    const elevations = new Float32Array(buffer, elevationOffset, size);

    // Extract biome data (uint8 array after elevations)
    const biomeOffset = HEADER_SIZE + size * 4;
    const biomes = new Uint8Array(buffer, biomeOffset, size);

    console.log(`[BinaryDataParser] Parsed grid: ${width}x${height}, ${size} cells`);

    return {
        width,
        height,
        elevations,
        biomes,
    };
}

/**
 * MapDataLayer provides convenient access to grid data by coordinates
 */
export class MapDataLayer {
    private data: ParsedGridData;

    constructor(data: ParsedGridData) {
        this.data = data;
    }

    get width(): number {
        return this.data.width;
    }

    get height(): number {
        return this.data.height;
    }

    /**
     * Get elevation at grid coordinates
     */
    getElevation(gridX: number, gridY: number): number {
        if (gridX < 0 || gridX >= this.data.width || gridY < 0 || gridY >= this.data.height) {
            return NaN;
        }
        return this.data.elevations[gridY * this.data.width + gridX];
    }

    /**
     * Get biome ID at grid coordinates
     */
    getBiome(gridX: number, gridY: number): BiomeID {
        if (gridX < 0 || gridX >= this.data.width || gridY < 0 || gridY >= this.data.height) {
            return BiomeID.Unknown;
        }
        return this.data.biomes[gridY * this.data.width + gridX] as BiomeID;
    }

    /**
     * Get biome name at grid coordinates
     */
    getBiomeName(gridX: number, gridY: number): string {
        return getBiomeName(this.getBiome(gridX, gridY));
    }

    /**
     * Convert world coordinates to grid coordinates
     */
    worldToGrid(worldX: number, worldY: number, worldWidth: number, worldHeight: number): { x: number; y: number } {
        // Normalize to 0..1
        const normalizedX = worldX / worldWidth;
        const normalizedY = worldY / worldHeight;

        // Scale to grid dimensions
        const gridX = Math.floor(normalizedX * this.data.width);
        const gridY = Math.floor(normalizedY * this.data.height);

        return {
            x: Math.max(0, Math.min(this.data.width - 1, gridX)),
            y: Math.max(0, Math.min(this.data.height - 1, gridY)),
        };
    }

    /**
     * Get data at world coordinates
     */
    getAtWorld(worldX: number, worldY: number, worldWidth: number, worldHeight: number): { elevation: number; biome: string } {
        const { x, y } = this.worldToGrid(worldX, worldY, worldWidth, worldHeight);
        return {
            elevation: this.getElevation(x, y),
            biome: this.getBiomeName(x, y),
        };
    }
}
