import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import { Mesh } from "@babylonjs/core/Meshes/mesh";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
import { Texture } from "@babylonjs/core/Materials/Textures/texture";
import { Color3, Color4 } from "@babylonjs/core/Maths/math.color";
import { ParticleSystem } from "@babylonjs/core/Particles/particleSystem";
import { Animation } from "@babylonjs/core/Animations/animation";
import type { Scene } from "@babylonjs/core/scene";
import type { AsteroidImpactMessage } from "$lib/types/websocket";
import { cartesianToLatLon } from "./FPSTransitionController"; // Re-using if needed, or implement latLonToCartesian locally

/**
 * Manages asteroid visuals: spawning, approach animation, impact explosion, and crater projection.
 */
export class AsteroidManager {
    private scene: Scene;
    private planetNode: Mesh | null = null; // The target globe for decals

    constructor(scene: Scene, planetNode: Mesh | null = null) {
        this.scene = scene;
        this.planetNode = planetNode;
    }

    public setPlanetNode(node: Mesh) {
        this.planetNode = node;
    }

    /**
     * Handles an incoming asteroid impact event.
     */
    public handleImpactEvent(msg: AsteroidImpactMessage['data']) {
        const metadata = msg.metadata;
        // 1. Calculate impact position (Cartesian)
        const impactPos = this.latLonToCartesian(metadata.location.lat, metadata.location.lon, 1.0); // radius 1.0

        // 2. Calculate start position
        // If origin provided, use it, otherwise fake a high orbit approach
        const startDistance = 5.0;
        const startPos = impactPos.scale(startDistance);
        // Add some random offset to start pos to make it look like an angled trajectory? 
        // Actually, let's just come from the 'origin' direction if we had one.
        // For now, straight down is fine for v1, or slightly angled.

        this.spawnAndAnimateAsteroid(startPos, impactPos, metadata.mass);
    }

    private spawnAndAnimateAsteroid(startPos: Vector3, endPos: Vector3, mass: number) {
        // Create Asteroid Mesh
        const diameter = Math.max(0.01, Math.log10(mass) * 0.005); // Rough scaling
        const asteroid = MeshBuilder.CreateSphere("asteroid", { diameter, segments: 4 }, this.scene);
        asteroid.position = startPos;

        const mat = new StandardMaterial("asteroidMat", this.scene);
        mat.diffuseColor = new Color3(0.4, 0.35, 0.3); // Rocky brown
        mat.specularColor = new Color3(0.1, 0.1, 0.1);
        asteroid.material = mat;

        // Animation
        // Distance roughly 4 units
        const frameRate = 30;
        const durationSeconds = 3; // Fast approach
        const totalFrames = frameRate * durationSeconds;

        const anim = new Animation("asteroidApproach", "position", frameRate, Animation.ANIMATIONTYPE_VECTOR3, Animation.ANIMATIONLOOPMODE_CONSTANT);

        const keys = [
            { frame: 0, value: startPos },
            { frame: totalFrames, value: endPos }
        ];
        anim.setKeys(keys);

        asteroid.animations.push(anim);

        this.scene.beginAnimation(asteroid, 0, totalFrames, false, 1.0, () => {
            // On Impact
            this.createExplosion(endPos, diameter);
            this.createCrater(endPos, diameter);
            asteroid.dispose();
        });
    }

    private createExplosion(position: Vector3, scale: number) {
        // Reuse similar logic to Moon Destruction but smaller/faster
        const particleSystem = new ParticleSystem("impactExplosion", 2000, this.scene);
        particleSystem.particleTexture = new Texture("/textures/flare.png", this.scene);
        particleSystem.emitter = position;
        particleSystem.minEmitBox = new Vector3(-scale, 0, -scale);
        particleSystem.maxEmitBox = new Vector3(scale, 0, scale);

        particleSystem.color1 = new Color4(1.0, 0.5, 0.0, 1.0); // Orange
        particleSystem.color2 = new Color4(1.0, 0.0, 0.0, 1.0); // Red
        particleSystem.colorDead = new Color4(0, 0, 0, 0.0);

        particleSystem.minSize = scale * 0.5;
        particleSystem.maxSize = scale * 2.0;

        particleSystem.minLifeTime = 0.5;
        particleSystem.maxLifeTime = 1.5;

        particleSystem.emitRate = 1000;
        particleSystem.targetStopDuration = 0.5; // Short burst

        particleSystem.start();

        // Auto dispose system after use
        setTimeout(() => {
            particleSystem.dispose();
        }, 2000);
    }

    private createCrater(position: Vector3, scale: number) {
        if (!this.planetNode) return;

        // Decal projection
        // We need a decent crater texture. For now, use a dark circle procedural or noise?
        // Or just a standard material decal.
        // Assuming we have a texture or just use a dark color decal.

        const decalSize = new Vector3(scale * 3, scale * 3, scale * 3);
        const decal = MeshBuilder.CreateDecal("crater", this.planetNode, {
            position: position,
            normal: position.normalize(), // Normal is outward form center
            size: decalSize
        });

        const decalMat = new StandardMaterial("craterMat", this.scene);
        decalMat.diffuseColor = new Color3(0.1, 0.1, 0.1); // Dark grey/black
        decalMat.specularColor = Color3.Black();
        decalMat.zOffset = -1; // Prevent z-fighting

        decal.material = decalMat;
        decal.isPickable = false;

        // Parent to planet so it rotates with it
        decal.setParent(this.planetNode);
    }

    private latLonToCartesian(lat: number, lon: number, radius: number): Vector3 {
        const latRad = lat * Math.PI / 180;
        const lonRad = lon * Math.PI / 180;

        // Babylon uses Left-Handed Y-up usually, assuming standard mapping:
        // y is up (north), x/z are equatorial plane
        const x = radius * Math.cos(latRad) * Math.cos(lonRad);
        const y = radius * Math.sin(latRad);
        const z = radius * Math.cos(latRad) * Math.sin(lonRad);

        return new Vector3(x, y, z);
    }
}
