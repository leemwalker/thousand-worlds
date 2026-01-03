/**
 * GroundRaycastSystem - Custom 5-ray ground detection for FPS mode.
 * Phase 2: Core Movement - No physics engine dependency
 * 
 * Uses 5-ray pattern (center + 4 corners) to detect ground surface,
 * calculate slope, and determine player standing position.
 */

import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import type { IHeightmapProvider } from './interfaces';

// Ray pattern for capsule approximation
// Center ray + 4 offset rays at capsule radius
const RAY_OFFSETS = [
    { x: 0, z: 0, weight: 0.4 },      // Center (highest weight)
    { x: 0.3, z: 0, weight: 0.15 },   // Right
    { x: -0.3, z: 0, weight: 0.15 },  // Left
    { x: 0, z: 0.3, weight: 0.15 },   // Front
    { x: 0, z: -0.3, weight: 0.15 },  // Back
];

export interface GroundHit {
    position: Vector3;      // Ground hit position in world space
    normal: Vector3;        // Ground surface normal
    height: number;         // Ground height at this point
    slope: number;          // Slope angle in radians
    isWater: boolean;       // Whether hit water surface
    waterDepth: number;     // Depth below water surface (0 if not in water)
}

export interface GroundRaycastResult {
    isGrounded: boolean;    // Whether any ray hit ground
    groundHeight: number;   // Weighted average ground height
    slopeAngle: number;     // Average slope angle in radians
    slopeDirection: Vector3; // Direction of steepest descent
    isInWater: boolean;     // Whether standing in water
    waterDepth: number;     // Depth of water at position
    rayHits: GroundHit[];   // Individual ray results
}

export interface GroundRaycastSystemOptions {
    capsuleRadius?: number;     // Player capsule radius
    maxGroundDistance?: number; // Max distance to consider "grounded"
    stepHeight?: number;        // Max step height player can climb
}

/**
 * Converts world position to lat/lon for heightmap sampling.
 */
function worldToLatLon(position: Vector3): { lat: number; lon: number } {
    const radius = position.length();
    if (radius < 0.001) return { lat: 0, lon: 0 };

    const normalized = position.scale(1 / radius);
    const lat = Math.asin(normalized.y) * 180 / Math.PI;
    const lon = Math.atan2(normalized.z, normalized.x) * 180 / Math.PI;

    return { lat, lon };
}

/**
 * Custom ground detection using heightmap sampling.
 * Works without physics engine - uses HeightmapProvider for terrain data.
 */
export class GroundRaycastSystem {
    private heightmapProvider: IHeightmapProvider | null = null;
    private capsuleRadius: number;
    private maxGroundDistance: number;
    private stepHeight: number;
    private planetRadius: number = 1.0;
    private seaLevel: number = 0;

    constructor(options: GroundRaycastSystemOptions = {}) {
        this.capsuleRadius = options.capsuleRadius ?? 0.3;
        this.maxGroundDistance = options.maxGroundDistance ?? 0.5;
        this.stepHeight = options.stepHeight ?? 0.3;
    }

    /**
     * Set the heightmap provider for terrain sampling.
     */
    setHeightmapProvider(provider: IHeightmapProvider): void {
        this.heightmapProvider = provider;
    }

    /**
     * Set planet parameters for height calculations.
     */
    setPlanetParams(radius: number, seaLevel: number): void {
        this.planetRadius = radius;
        this.seaLevel = seaLevel;
    }

    /**
     * Perform ground detection from a world position.
     * Uses 5-ray pattern to sample ground around player.
     */
    castFromPosition(position: Vector3, forward: Vector3): GroundRaycastResult {
        if (!this.heightmapProvider) {
            return this.createEmptyResult();
        }

        const rayHits: GroundHit[] = [];
        let totalWeight = 0;
        let weightedHeight = 0;
        let isInWater = false;
        let maxWaterDepth = 0;

        // Calculate right vector for offset positioning
        const up = position.clone().normalize();
        const right = Vector3.Cross(up, forward).normalize();
        const actualForward = Vector3.Cross(right, up).normalize();

        // Cast rays at each offset position
        for (const offset of RAY_OFFSETS) {
            // Calculate ray origin with offset
            const rayOrigin = position.clone()
                .add(right.scale(offset.x * this.capsuleRadius))
                .add(actualForward.scale(offset.z * this.capsuleRadius));

            // Sample heightmap at this position
            const { lat, lon } = worldToLatLon(rayOrigin);
            const terrainHeight = this.heightmapProvider.getHeightAt(lat, lon);

            // Convert terrain height to world distance from planet center
            const elevRange = this.heightmapProvider.getElevationRange();
            const normalizedHeight = (terrainHeight - elevRange.min) / (elevRange.max - elevRange.min);
            const worldGroundHeight = this.planetRadius + normalizedHeight * 0.05; // 5% terrain scale

            // Calculate ground position in world space
            const groundDir = rayOrigin.clone().normalize();
            const groundPos = groundDir.scale(worldGroundHeight);

            // Check for water
            const inWater = terrainHeight < this.seaLevel;
            const waterSurfaceHeight = this.planetRadius + 0.001; // Slightly above planet
            let waterDepth = 0;

            if (inWater) {
                isInWater = true;
                waterDepth = waterSurfaceHeight - worldGroundHeight;
                maxWaterDepth = Math.max(maxWaterDepth, waterDepth);
            }

            // Calculate slope (using terrain gradient)
            const gradient = this.calculateGradient(lat, lon);
            const slopeAngle = Math.atan(gradient.magnitude);

            const hit: GroundHit = {
                position: groundPos,
                normal: groundDir, // Simplified - actual normal would use gradient
                height: worldGroundHeight,
                slope: slopeAngle,
                isWater: inWater,
                waterDepth: waterDepth
            };

            rayHits.push(hit);
            weightedHeight += worldGroundHeight * offset.weight;
            totalWeight += offset.weight;
        }

        // Calculate average values
        const avgHeight = totalWeight > 0 ? weightedHeight / totalWeight : 0;
        const playerDistance = position.length();
        const distanceToGround = playerDistance - avgHeight;
        const isGrounded = distanceToGround < this.maxGroundDistance && distanceToGround > -this.stepHeight;

        // Calculate average slope
        let totalSlope = 0;
        for (const hit of rayHits) {
            totalSlope += hit.slope;
        }
        const avgSlope = rayHits.length > 0 ? totalSlope / rayHits.length : 0;

        // Calculate slope direction (simplified)
        const slopeDir = this.calculateSlopeDirection(rayHits);

        return {
            isGrounded,
            groundHeight: avgHeight,
            slopeAngle: avgSlope,
            slopeDirection: slopeDir,
            isInWater,
            waterDepth: maxWaterDepth,
            rayHits
        };
    }

    /**
     * Calculate terrain gradient at position for slope detection.
     */
    private calculateGradient(lat: number, lon: number): { x: number; z: number; magnitude: number } {
        if (!this.heightmapProvider) {
            return { x: 0, z: 0, magnitude: 0 };
        }

        const delta = 0.01; // Sample offset in degrees
        const h = this.heightmapProvider.getHeightAt(lat, lon);
        const hN = this.heightmapProvider.getHeightAt(lat + delta, lon);
        const hE = this.heightmapProvider.getHeightAt(lat, lon + delta);

        const gradX = (hE - h) / delta;
        const gradZ = (hN - h) / delta;
        const magnitude = Math.sqrt(gradX * gradX + gradZ * gradZ);

        return { x: gradX, z: gradZ, magnitude };
    }

    /**
     * Calculate direction of steepest descent from ray hits.
     */
    private calculateSlopeDirection(hits: GroundHit[]): Vector3 {
        if (hits.length < 3) {
            return Vector3.Zero();
        }

        // Find lowest and highest points
        let minHeight = Infinity;
        let maxHeight = -Infinity;
        let minPos = Vector3.Zero();
        let maxPos = Vector3.Zero();

        for (const hit of hits) {
            if (hit.height < minHeight) {
                minHeight = hit.height;
                minPos = hit.position;
            }
            if (hit.height > maxHeight) {
                maxHeight = hit.height;
                maxPos = hit.position;
            }
        }

        // Direction from high to low
        const dir = minPos.subtract(maxPos);
        dir.normalize();
        return dir;
    }

    /**
     * Create empty result when no heightmap is available.
     */
    private createEmptyResult(): GroundRaycastResult {
        return {
            isGrounded: false,
            groundHeight: 0,
            slopeAngle: 0,
            slopeDirection: Vector3.Zero(),
            isInWater: false,
            waterDepth: 0,
            rayHits: []
        };
    }

    /**
     * Check if a step up is climbable.
     */
    canClimbStep(currentHeight: number, targetHeight: number): boolean {
        const stepDiff = targetHeight - currentHeight;
        return stepDiff > 0 && stepDiff <= this.stepHeight;
    }
}
