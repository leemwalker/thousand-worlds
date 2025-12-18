
// Logic Helper for Map Rendering
// Extracted for TDD and clean architecture

interface Position3D { x: number; y: number; z: number; }
interface TileData { x: number; y: number; elevation: number; biome: string; }

// Use local copies of colors to avoid circular deps with MapRenderer constants?
// Or we export this from MapRenderer.
// Ideally, we move constants to a shared `constants.ts` or `types.ts` in Map folder.
// For now, I'll duplicate the critical Lobby/Void/Default check logic, or better, 
// create a pure function that returns "Override" or null.

export function getGeologyStyleOverride(biome: string): string | null {
    const b = biome.toLowerCase();
    if (b === 'lobby' || b === 'default' || b === 'void') {
        return '#333333';
    }
    return null;
}

export function getRenderLayer(player: Position3D, tile: TileData, threshold: number = 200): 'near' | 'far' {
    // Distance in Elevation (Z) Only
    const dz = Math.abs(player.z - tile.elevation);
    return dz < threshold ? 'near' : 'far';
}
