/**
 * Moon types for frontend rendering and FPV physics.
 * Mirrors backend astronomy.Satellite with fields needed for display.
 */

/** Gravitational constant G = 6.674e-11 m³/(kg·s²) */
export const G_CONSTANT = 6.674e-11;

/** Moon data from backend for rendering and physics */
export interface MoonData {
    id: string;
    name: string;
    mass: number;      // kg
    radius: number;    // meters
    distance: number;  // meters from planet center
    period: number;    // seconds
    color: string;     // hex color
    density: number;   // kg/m³
    destroyed: boolean;
    destroyedAt?: number;
    ringFormed?: boolean;
}

/**
 * Calculate surface gravity for a celestial body.
 * Formula: g = G × M / r²
 * 
 * @param mass - Mass in kg
 * @param radius - Radius in meters
 * @returns Surface gravity in m/s²
 */
export function calculateSurfaceGravity(mass: number, radius: number): number {
    if (radius <= 0 || mass <= 0) return 0;
    return (G_CONSTANT * mass) / (radius * radius);
}

/**
 * Calculate horizon distance from eye height.
 * Formula: d = √(2Rh + h²) where R = radius, h = eye height
 * 
 * @param radius - Body radius in meters
 * @param eyeHeight - Observer height in meters
 * @returns Distance to horizon in meters
 */
export function calculateHorizonDistance(radius: number, eyeHeight: number): number {
    if (radius <= 0 || eyeHeight <= 0) return 0;
    return Math.sqrt(2 * radius * eyeHeight + eyeHeight * eyeHeight);
}

/**
 * Convert a MoonData object to FPV physics parameters.
 * Gravity is calculated from mass/radius.
 * Move speed is scaled for moon size.
 */
export function getMoonFPVParams(moon: MoonData): MoonFPVParams {
    const gravity = calculateSurfaceGravity(moon.mass, moon.radius);
    const horizonDistance = calculateHorizonDistance(moon.radius, 1.7); // Eye height

    // Scale move speed - lower gravity means faster movement feels right
    // Earth gravity = 9.8 m/s², so scale factor = sqrt(9.8/g)
    const gravityScaleFactor = gravity > 0 ? Math.sqrt(9.8 / gravity) : 1;
    const moveSpeed = 5.0 * Math.min(gravityScaleFactor, 4.0); // Cap at 4x speed

    // Jump height scales inversely with gravity
    const jumpHeight = 2.0 * Math.min(gravityScaleFactor, 6.0); // Cap at 6x height

    return {
        gravity,
        horizonDistance,
        moveSpeed,
        jumpHeight,
        moonName: moon.name,
        moonRadius: moon.radius,
        moonColor: moon.color,
    };
}

export interface MoonFPVParams {
    gravity: number;         // m/s²
    horizonDistance: number; // meters
    moveSpeed: number;       // m/s
    jumpHeight: number;      // meters
    moonName: string;
    moonRadius: number;      // meters (for terrain scale)
    moonColor: string;       // For terrain tint
}
