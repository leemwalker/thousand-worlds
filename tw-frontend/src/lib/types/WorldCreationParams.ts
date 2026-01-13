export interface WorldCreationParams {
    name: string;
    seed: string;
    size: 'small' | 'medium' | 'large' | 'huge';
    resolution: 32 | 64 | 128 | 256 | 512;
    coreType: 'continental' | 'volcanic' | 'oceanic' | 'ancient';
    waterLevel: 'low' | 'medium' | 'high' | string;
    moonCount: number;
    sysGeology: boolean;
    sysWeather: boolean;
    sysLife: boolean;
    sysDisease: boolean;
    sysSapience: boolean;
    sysMigration: boolean;
    epoch?: string;
    goal?: string;
}

export const DEFAULT_WORLD_PARAMS: WorldCreationParams = {
    name: '',
    seed: '',
    size: 'large',     // Earth-like
    resolution: 128,   // Standard Detail
    coreType: 'continental',
    waterLevel: 'medium',
    moonCount: 1,
    sysGeology: true,
    sysWeather: true,
    sysLife: true,
    sysDisease: true,
    sysSapience: true,
    sysMigration: true
};
