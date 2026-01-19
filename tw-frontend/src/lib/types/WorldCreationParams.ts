export interface WorldCreationParams {
    name: string;
    seed: string;
    size: 'small' | 'medium' | 'large' | 'huge';
    diameter: number;     // Planet diameter in km (1737 Moon to 142984 Jupiter)
    gravity: number;      // Surface gravity multiplier (0.1x to 10x Earth)
    resolution: 256 | 512 | 1024 | 2048 | 4096;
    coreType: 'continental' | 'volcanic' | 'oceanic' | 'ancient';
    waterLevel: 'low' | 'medium' | 'high' | string;
    moonCount: number;
    yearsToSimulate: number;  // 0 = open-ended, else 10^exponent (1e3 to 1e10)
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
    diameter: 12742,   // Earth diameter in km
    gravity: 1.0,      // 1x Earth gravity
    resolution: 512,   // Balanced (new default)
    coreType: 'continental',
    waterLevel: 'medium',
    moonCount: 1,
    yearsToSimulate: 1000000000,  // 1 billion years default
    sysGeology: true,
    sysWeather: true,
    sysLife: true,
    sysDisease: true,
    sysSapience: true,
    sysMigration: true
};

// Reference presets for UI
export const PLANET_PRESETS = {
    moon: { diameter: 3475, gravity: 0.16, label: 'Moon' },
    mars: { diameter: 6779, gravity: 0.38, label: 'Mars' },
    earth: { diameter: 12742, gravity: 1.0, label: 'Earth' },
    superEarth: { diameter: 20000, gravity: 2.5, label: 'Super-Earth' },
    neptune: { diameter: 49244, gravity: 1.14, label: 'Neptune' },
    jupiter: { diameter: 142984, gravity: 2.53, label: 'Jupiter' }
} as const;
