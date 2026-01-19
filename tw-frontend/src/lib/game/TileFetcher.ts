import lz4 from 'lz4js';

export interface TileData {
    face: number;
    level: number;
    x: number;
    y: number;
    width: number;
    height: number;
    heightmap: Float32Array;
    biomes: Uint8Array;
    water: Float32Array;
}

export class TileFetcher {
    // Magic '0xWT', Version 0x01
    private static readonly MAGIC = "WT";
    private static readonly VERSION = 1;

    constructor(private baseUrl: string = '/api/v1/world/tiles') { }

    /**
     * Fetches and decodes a tile from the backend.
     * @param face Cube face index (0-5)
     * @param lod Level of detail
     * @param x Tile X coordinate
     * @param y Tile Y coordinate
     * @returns Promise resolving to decoded TileData
     */
    async fetchTile(face: number, lod: number, x: number, y: number): Promise<TileData> {
        const url = `${this.baseUrl}/${face}/${lod}/${x}/${y}`;
        const response = await fetch(url);

        if (!response.ok) {
            throw new Error(`Failed to fetch tile: ${response.status} ${response.statusText}`);
        }

        const buffer = await response.arrayBuffer();
        return this.decodeTile(buffer);
    }

    /**
     * Decodes the binary tile format.
     * Format:
     * - Magic "WT" (2 bytes)
     * - Version (1 byte)
     * - Meta Length (uint32 LE)
     * - Meta (JSON)
     * - Encoded Data Length (uint32 LE)
     * - LZ4 Compressed Payload
     *   - Heightmap (floats)
     *   - Biomes (uint8)
     *   - Water (floats)
     */
    public decodeTile(buffer: ArrayBuffer): TileData {
        const view = new DataView(buffer);
        let offset = 0;

        // 1. Verify Header
        const magic = String.fromCharCode(view.getUint8(offset), view.getUint8(offset + 1));
        offset += 2;
        if (magic !== TileFetcher.MAGIC) {
            throw new Error(`Invalid magic bytes: ${magic}`);
        }

        const version = view.getUint8(offset);
        offset += 1;
        if (version !== TileFetcher.VERSION) {
            throw new Error(`Unsupported version: ${version}`);
        }

        // 2. Read Metadata
        const metaLen = view.getUint32(offset, true); // Little Endian
        offset += 4;

        const metaBytes = new Uint8Array(buffer, offset, metaLen);
        const metaStr = new TextDecoder().decode(metaBytes);
        const meta = JSON.parse(metaStr);
        offset += metaLen;

        // 3. Read Compressed Payload
        const compressedLen = view.getUint32(offset, true);
        offset += 4;

        const compressedData = new Uint8Array(buffer, offset, compressedLen);

        // Decompress
        // Need to know uncompressed size? 
        // We can calculate it from meta.w * meta.h * (4 + 1 + 4 bytes)
        // Heightmap(4) + Biomes(1) + Water(4) = 9 bytes per pixel
        const pixelCount = meta.w * meta.h;
        const uncompressedSize = pixelCount * 9;

        // LZ4 decompression
        const uncompressedBuffer = lz4.decompress(compressedData);
        if (uncompressedBuffer.byteLength !== uncompressedSize) {
            console.warn(`Decompressed size mismatch: expected ${uncompressedSize}, got ${uncompressedBuffer.byteLength}`);
        }

        // 4. Parse Arrays from Uncompressed Buffer
        // Payload Layout: [Heightmap Floats] [Biome Bytes] [Water Floats]
        const payloadView = new DataView(uncompressedBuffer.buffer, uncompressedBuffer.byteOffset, uncompressedBuffer.byteLength);
        let pOffset = 0;

        // Heightmap (Float32)
        const heightmap = new Float32Array(pixelCount);
        for (let i = 0; i < pixelCount; i++) {
            heightmap[i] = payloadView.getFloat32(pOffset, true);
            pOffset += 4;
        }

        // Biomes (Uint8)
        const biomes = new Uint8Array(uncompressedBuffer.buffer, uncompressedBuffer.byteOffset + pOffset, pixelCount);
        // Note: Uint8Array creation shares buffer, but we need to advance offset manually for next logic
        // Or just create a copy/slice if needed. 
        // Since we are reading sequentially, we can just slice or create view.
        // Uint8Array constructor with offset/length works on the buffer.
        // BUT, creating typed arrays directly on the buffer is more efficient.
        // Let's do that properly.

        // Refined approach for efficiency:
        // Create views directly on the uncompressed buffer if alignment allows.
        // Float32 requires 4-byte alignment. lz4js generic array might not be aligned?
        // Usually it returns Uint8Array. We can copy if needed.

        // Let's stick to reading manually to be safe against alignment issues for now, 
        // or copy to aligned buffer.
        // The loop above for heightmap is safe.
        pOffset += pixelCount;

        // Water (Float32)
        const water = new Float32Array(pixelCount);
        for (let i = 0; i < pixelCount; i++) {
            water[i] = payloadView.getFloat32(pOffset, true);
            pOffset += 4;
        }

        return {
            face: meta.f,
            level: meta.l,
            x: meta.x,
            y: meta.y,
            width: meta.w,
            height: meta.h,
            heightmap,
            biomes,
            water
        };
    }
}
