import { TileFetcher, type TileData } from '$lib/game/TileFetcher';
import { RawTexture } from "@babylonjs/core/Materials/Textures/rawTexture";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { type TileCoord, type ITileProvider, CubeFace } from './interfaces';
import type { Scene } from "@babylonjs/core/scene";

export class TileProvider implements ITileProvider {
    private scene: Scene;
    private fetcher: TileFetcher;
    private cache: Map<string, any> = new Map();
    private pendingRequests: Map<string, Promise<any>> = new Map();
    private preloadQueue: TileCoord[] = [];
    private isPreloading = false;

    constructor(
        scene: Scene,
        sendCommand?: (action: string, message?: string) => void
    ) {
        this.scene = scene;
        this.fetcher = new TileFetcher();
    }

    private getTileKey(coord: TileCoord): string {
        return `${coord.face}_${coord.level}_${coord.x}_${coord.y}`;
    }

    async getTile(coord: TileCoord): Promise<{ texture: Texture; heightmap: Texture; raw: TileData }> {
        const key = this.getTileKey(coord);

        if (this.cache.has(key)) {
            return this.cache.get(key)!;
        }

        if (this.pendingRequests.has(key)) {
            return this.pendingRequests.get(key)!;
        }

        const promise = (async () => {
            try {
                // Fetch generic tile data
                const tileData = await this.fetcher.fetchTile(coord.face, coord.level, coord.x, coord.y);

                // Create textures for legacy/rendering support
                // Heightmap is Float32, Biomes is Uint8.
                // We need to create RawTextures.

                // Heightmap Texture (Rfloat)
                const heightmapTexture = RawTexture.CreateRTexture(
                    tileData.heightmap,
                    tileData.width,
                    tileData.height,
                    this.scene,
                    false, // generateMipMaps
                    false, // invertY
                    Texture.NEAREST_SAMPLINGMODE
                );

                // Biome/Water Texture (RG - R=Biome, G=nothing? Or maybe pass raw buffers?)
                // The existing shader expects encoded textures.
                // Let's create a "Data" texture.
                // R=Elevation (normalized? No, we have float heightmap now), G=Biome, B=Entity
                // Wait, existing shader uses R=0.5 sea level.
                // Our Compute Shader uses Real World values.

                // For now, let's create a simplified texture for visualization if needed,
                // BUT the goal is WebGPU Compute Shader which uses the buffers directly.

                // We return the raw data so TileMesh (or GPUTileMesh) can use it.
                // We also return a dummy texture if existing code requires it?
                // The interface change allows returning `any`.

                // Let's return the raw package + textures for compatibility if possible.
                // Re-using existing biome coloring shader might require the specific texture format.
                // Existing WebGLMapRenderer expects:
                // R=Elev(normalized), G=Biome, B=Entity, A=Alpha

                // We'll create a single RGBA texture to match existing render logic if we want to fallback.
                // But for WebGPU, we want `raw` buffers.

                // Let's generate a visualization texture from the data for debugging/fallback
                // (Optional: skip if pure WebGPU)
                const vizTexture = this.createVisualizationTexture(tileData);

                const result = {
                    texture: vizTexture,
                    heightmap: heightmapTexture,
                    raw: tileData
                };

                this.cache.set(key, result);
                return result;
            } finally {
                this.pendingRequests.delete(key);
            }
        })();

        this.pendingRequests.set(key, promise);
        return promise;
    }

    private createVisualizationTexture(data: TileData): Texture {
        const count = data.width * data.height;
        const buffer = new Uint8Array(count * 4);

        for (let i = 0; i < count; i++) {
            // R: Normalized elevation (hacky approximation for fallback)
            // 0.5 = 0m. Range -10000 to 10000 -> 0-1?
            const h = data.heightmap[i] ?? 0;
            const norm = (h + 10000) / 20000; // clamp
            buffer[i * 4 + 0] = Math.floor(Math.max(0, Math.min(1, norm)) * 255);

            // G: Biome ID
            buffer[i * 4 + 1] = data.biomes[i] ?? 0;

            // B: Water? or Entity?
            buffer[i * 4 + 2] = 0;

            // A: Alpha
            buffer[i * 4 + 3] = 255;
        }

        return new RawTexture(
            buffer,
            data.width,
            data.height,
            4, // RGBA
            this.scene,
            false,
            false,
            Texture.NEAREST_SAMPLINGMODE
        );
    }

    preloadTiles(coords: TileCoord[]): void {
        // Implementation similar to before, but using new getTile
        coords.forEach(c => {
            if (!this.isCached(c)) this.getTile(c).catch(() => { });
        });
    }

    isCached(coord: TileCoord): boolean {
        return this.cache.has(this.getTileKey(coord));
    }
}
