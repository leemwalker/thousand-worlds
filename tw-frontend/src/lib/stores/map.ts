import { writable } from 'svelte/store';
import type { VisibleTile, RenderQuality } from '$lib/components/Map/MapRenderer';

export interface MapData {
    tiles: VisibleTile[];
    player_x: number;
    player_y: number;
    render_quality: RenderQuality;
    grid_size: number;
    scale: number;
    world_id: string;
}

interface MapStoreState {
    data: MapData | null;
    lastUpdate: number;
}

function createMapStore() {
    const { subscribe, set, update } = writable<MapStoreState>({
        data: null,
        lastUpdate: 0
    });

    return {
        subscribe,
        setMapData: (data: MapData) => {
            update(store => ({
                ...store,
                data,
                lastUpdate: Date.now()
            }));
        },
        clear: () => set({
            data: null,
            lastUpdate: 0
        })
    };
}

export const mapStore = createMapStore();
