<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Engine } from "@babylonjs/core/Engines/engine";
    import { Scene } from "@babylonjs/core/scene";
    import { ArcRotateCamera } from "@babylonjs/core/Cameras/arcRotateCamera";
    import { PointLight } from "@babylonjs/core/Lights/pointLight";
    import { HemisphericLight } from "@babylonjs/core/Lights/hemisphericLight";
    import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
    import { TransformNode } from "@babylonjs/core/Meshes/transformNode";
    import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
    import { Texture } from "@babylonjs/core/Materials/Textures/texture";
    import { RawTexture } from "@babylonjs/core/Materials/Textures/rawTexture";
    import { Vector3 } from "@babylonjs/core/Maths/math.vector";
    import { Color3, Color4 } from "@babylonjs/core/Maths/math.color";
    import { GlowLayer } from "@babylonjs/core/Layers/glowLayer";
    import type { Mesh } from "@babylonjs/core/Meshes/mesh";
    import type { VertexData } from "@babylonjs/core/Meshes/mesh.vertexData";
    import { LODManager } from "./LODManager";
    import { DisplacementShader } from "./DisplacementShader";

    // Types for satellite/moon data
    interface Satellite {
        name: string;
        mass: number; // kg
        distance: number; // km from planet
    }

    // Props
    export let globeTextureBlob: Blob | null = null;
    export let globeHeightmapBlob: Blob | null = null;
    export let seaLevel: number = 0;
    export let maxElevation: number = 8848;
    export let minElevation: number = -11000;
    export let satellites: Satellite[] = []; // Moon data from world
    export let simulationSpeed: number = 1.0; // Time multiplier for animations

    // Internal state
    let canvas: HTMLCanvasElement;
    let engine: Engine | null = null;
    let scene: Scene | null = null;
    let camera: ArcRotateCamera | null = null;
    let globe: Mesh | null = null;
    let globeTexture: Texture | null = null;
    let globeMaterial: StandardMaterial | null = null;
    let objectUrl: string | null = null;
    let lastAppliedBlobSize: number = 0; // Guard to prevent re-applying same texture
    let lastAppliedHeightDataLength: number = 0; // Guard for height data
    let waterBumpTexture: Texture | null = null; // Animated water normals
    let waterTime = 0; // Time accumulator for water animation

    // Solar system nodes
    let solarSystemRoot: TransformNode | null = null;
    let orbitNode: TransformNode | null = null;
    let planetNode: TransformNode | null = null;
    let sunMesh: Mesh | null = null;
    let sunLight: PointLight | null = null;
    let moonMeshes: Mesh[] = [];
    let moonOrbitNodes: TransformNode[] = [];

    // LOD system for zoom-based detail
    let lodManager: LODManager | null = null;
    let displacementShader: DisplacementShader | null = null;
    let shaderMaterial: any | null = null; // ShaderMaterial type

    // Animation state
    let lastTime = 0;
    const PLANET_DAY_SECONDS = 10; // Planet completes rotation in 10 real seconds
    const PLANET_YEAR_SECONDS = 120; // Planet completes orbit in 120 real seconds
    const SUN_DISTANCE = 20; // Distance from sun to planet

    // Terrain exaggeration factor (makes mountains visible from space)
    const TERRAIN_SCALE = 0.05; // 5% of radius for max height

    onMount(() => {
        if (!canvas) return;

        // Create engine with options for mobile performance
        engine = new Engine(canvas, true, {
            stencil: true,
            preserveDrawingBuffer: true,
            powerPreference: "high-performance",
        });

        // Create scene with dark space background
        scene = new Scene(engine);
        scene.clearColor = new Color4(0.02, 0.02, 0.06, 1); // Dark space

        // ===========================================
        // Solar System Hierarchy
        // ===========================================

        // Root node for entire solar system
        solarSystemRoot = new TransformNode("solarSystemRoot", scene);

        // Create the Sun at center
        sunMesh = MeshBuilder.CreateSphere(
            "sun",
            { segments: 32, diameter: 3 },
            scene,
        );
        sunMesh.parent = solarSystemRoot;
        sunMesh.position = new Vector3(0, 0, 0);

        // Sun material (emissive - glows)
        const sunMaterial = new StandardMaterial("sunMaterial", scene);
        sunMaterial.emissiveColor = new Color3(1.0, 0.95, 0.8); // Warm yellow-white
        sunMaterial.diffuseColor = new Color3(0, 0, 0);
        sunMaterial.specularColor = new Color3(0, 0, 0);
        sunMaterial.disableLighting = true;
        sunMesh.material = sunMaterial;

        // Sun as light source (PointLight)
        sunLight = new PointLight("sunLight", Vector3.Zero(), scene);
        sunLight.intensity = 2.0;
        sunLight.diffuse = new Color3(1.0, 0.98, 0.95); // Warm sunlight
        sunLight.parent = solarSystemRoot;

        // Note: GlowLayer removed - it ignores depth testing and makes sun visible through planet
        // The sun's emissive material still makes it bright without needing glow

        // Orbit node (rotates around sun - handles year)
        orbitNode = new TransformNode("orbitNode", scene);
        orbitNode.parent = solarSystemRoot;
        orbitNode.position = new Vector3(SUN_DISTANCE, 0, 0); // Start at sun distance

        // Planet node (rotates on axis - handles day)
        planetNode = new TransformNode("planetNode", scene);
        planetNode.parent = orbitNode;
        // Add slight axial tilt (like Earth's 23.5 degrees)
        planetNode.rotation = new Vector3(0, 0, (23.5 * Math.PI) / 180);

        // Ambient light for areas not in direct sunlight
        const ambient = new HemisphericLight(
            "ambient",
            new Vector3(0, 1, 0),
            scene,
        );
        ambient.intensity = 0.25;
        ambient.groundColor = new Color3(0.05, 0.05, 0.1);

        // ArcRotateCamera for orbit controls - targets planet
        camera = new ArcRotateCamera(
            "camera",
            Math.PI / 2, // alpha (horizontal rotation)
            Math.PI / 3, // beta (vertical rotation - 60 degrees from pole)
            5, // radius (distance from target) - increased for solar system view
            new Vector3(SUN_DISTANCE, 0, 0), // Target the planet's position
            scene,
        );
        camera.attachControl(canvas, true);
        camera.lowerRadiusLimit = 1.05; // Allow very close zoom (just above surface)
        camera.upperRadiusLimit = 50; // Allow zooming out to see sun
        camera.wheelPrecision = 30; // Scroll zoom sensitivity
        camera.panningSensibility = 0; // Disable panning, only rotate
        camera.minZ = 0.01; // Near clip plane - prevents clipping at close range

        // Initialize LOD manager with distance thresholds
        lodManager = new LODManager({
            levels: [
                { distance: 3, segments: 128 }, // High detail when close
                { distance: 8, segments: 64 }, // Medium detail
                { distance: 20, segments: 32 }, // Low detail when far
            ],
            hysteresis: 0.15,
        });

        // Initialize displacement shader handler
        displacementShader = new DisplacementShader(scene);

        // Create meshes using LODManager
        globe = lodManager.createMesh(scene, 0, "globe");
        globe.parent = planetNode;

        const mediumMesh = lodManager.createMesh(scene, 1, "globe");
        mediumMesh.parent = planetNode;
        mediumMesh.setEnabled(false);

        const lowMesh = lodManager.createMesh(scene, 2, "globe");
        lowMesh.parent = planetNode;
        lowMesh.setEnabled(false);

        // Create placeholder material initially (until heightmap loads)
        // We'll replace this with the shader material once we have the texture
        globeMaterial = new StandardMaterial("globeMaterial", scene);
        globeMaterial.diffuseColor = new Color3(0.2, 0.2, 0.25);
        globeMaterial.specularColor = new Color3(0.2, 0.2, 0.25);
        globeMaterial.specularPower = 32;
        globeMaterial.backFaceCulling = true; // Sphere doesn't need double-sided

        // Apply initial material to all meshes
        globe.material = globeMaterial;
        mediumMesh.material = globeMaterial;
        lowMesh.material = globeMaterial;

        // ===========================================
        // Create Moons (if any satellites provided)
        // ===========================================
        createMoons(scene);

        // Create starfield background
        createStarfield(scene);

        console.log("[BabylonGlobe] Solar system initialized");

        // Check if globeTextureBlob was already set before scene was ready
        if (globeTextureBlob && globeMaterial) {
            console.log(
                "[BabylonGlobe] Found existing globeTextureBlob, applying now...",
            );
            updateTexture(globeTextureBlob);
        }

        // Animation and render loop
        lastTime = performance.now();
        engine.runRenderLoop(() => {
            if (scene && orbitNode && planetNode) {
                const now = performance.now();
                const deltaTime = (now - lastTime) / 1000; // seconds
                lastTime = now;

                // Planet axial rotation (day cycle)
                const dayRotation =
                    (2 * Math.PI * deltaTime * simulationSpeed) /
                    PLANET_DAY_SECONDS;
                planetNode.rotation.y += dayRotation;

                // Orbital rotation around sun (year cycle)
                const yearRotation =
                    (2 * Math.PI * deltaTime * simulationSpeed) /
                    PLANET_YEAR_SECONDS;
                orbitNode.rotateAround(
                    Vector3.Zero(),
                    Vector3.Up(),
                    yearRotation,
                );

                // Update camera target to follow planet
                if (camera) {
                    camera.target = orbitNode.position;
                    // Update LOD based on camera distance
                    lodManager?.update(camera);
                }

                // Moon orbital animation
                moonOrbitNodes.forEach((moonOrbit, i) => {
                    // Moons orbit faster than planet orbits sun
                    const moonPeriod = 5 + i * 2; // Each moon has different period
                    const moonRotation =
                        (2 * Math.PI * deltaTime * simulationSpeed) /
                        moonPeriod;
                    moonOrbit.rotation.y += moonRotation;
                });

                // Animate water bump texture UV offset for wave motion
                waterTime += deltaTime * 0.05 * simulationSpeed;
                if (waterBumpTexture) {
                    waterBumpTexture.uOffset = Math.sin(waterTime) * 0.02;
                    waterBumpTexture.vOffset = Math.cos(waterTime * 0.7) * 0.01;
                }

                // TODO: LOD update disabled until shader displacement is implemented
                // if (lodManager && camera) {
                //     lodManager.update(camera);
                //     const currentLevel = lodManager.getCurrentLevel();
                //     globe = lodManager.getMesh(currentLevel);
                // }

                scene.render();
            }
        });

        // Handle window resize
        const handleResize = () => {
            if (engine) {
                engine.resize();
            }
        };
        window.addEventListener("resize", handleResize);

        return () => {
            window.removeEventListener("resize", handleResize);
        };
    });

    // React to texture blob changes (with guard to avoid re-applying same blob)
    $: if (
        globeTextureBlob &&
        scene &&
        globeMaterial &&
        globeTextureBlob.size !== lastAppliedBlobSize
    ) {
        console.log(
            "[BabylonGlobe] Reactive: globeTextureBlob received, size:",
            globeTextureBlob.size,
        );
        lastAppliedBlobSize = globeTextureBlob.size;
        updateTexture(globeTextureBlob);
    }

    // Watch for heightmap blob changes
    $: if (globeHeightmapBlob && scene && displacementShader) {
        applyHeightDisplacement(globeHeightmapBlob);
    }

    function createMoons(s: Scene) {
        if (!planetNode || satellites.length === 0) {
            console.log("[BabylonGlobe] No moons to create");
            return;
        }

        console.log(`[BabylonGlobe] Creating ${satellites.length} moon(s)`);

        // Clear existing moons
        moonMeshes.forEach((m) => m.dispose());
        moonOrbitNodes.forEach((n) => n.dispose());
        moonMeshes = [];
        moonOrbitNodes = [];

        satellites.forEach((sat, index) => {
            // Create orbit node for this moon (child of planet)
            const moonOrbit = new TransformNode(`moonOrbit_${sat.name}`, s);
            moonOrbit.parent = planetNode;

            // Calculate moon distance based on satellite data
            // Normalize to visible distance (real distances would be way too far)
            const normalizedDistance = 1.5 + (sat.distance / 400000) * 2; // Scale down significantly

            // Create moon mesh
            // Size based on mass (logarithmic scale)
            const moonSize = 0.1 + Math.log10(sat.mass / 1e20) * 0.05;
            const clampedSize = Math.max(0.05, Math.min(0.3, moonSize));

            const moon = MeshBuilder.CreateSphere(
                `moon_${sat.name}`,
                { segments: 16, diameter: clampedSize },
                s,
            );
            moon.parent = moonOrbit;
            moon.position = new Vector3(normalizedDistance, 0, 0);

            // Moon material (grey/white rocky)
            const moonMat = new StandardMaterial(`moonMat_${sat.name}`, s);
            moonMat.diffuseColor = new Color3(0.7, 0.7, 0.7);
            moonMat.specularColor = new Color3(0.1, 0.1, 0.1);
            moon.material = moonMat;

            // Store references
            moonMeshes.push(moon);
            moonOrbitNodes.push(moonOrbit);

            // Start each moon at different orbital position
            moonOrbit.rotation.y = (index * Math.PI * 2) / satellites.length;
        });
    }

    // Placeholder for future water mesh (Option 3)
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    let waterMesh: Mesh | null = null;

    function createWaterBumpTexture(s: Scene): Texture {
        // Generate procedural water normal map using canvas
        const size = 256;
        const bumpCanvas = document.createElement("canvas");
        bumpCanvas.width = size;
        bumpCanvas.height = size;
        const ctx = bumpCanvas.getContext("2d");

        if (ctx) {
            // Create wave-like normal map pattern
            const imageData = ctx.createImageData(size, size);
            const data = imageData.data;

            for (let y = 0; y < size; y++) {
                for (let x = 0; x < size; x++) {
                    const idx = (y * size + x) * 4;

                    // Multiple wave frequencies for realistic ocean look
                    const freq1 = 0.05;
                    const freq2 = 0.12;
                    const freq3 = 0.25;

                    // Height at this point (using sine waves)
                    const h1 = Math.sin(x * freq1) * Math.cos(y * freq1);
                    const h2 =
                        Math.sin(x * freq2 + 0.5) *
                        Math.cos(y * freq2 + 0.3) *
                        0.5;
                    const h3 = Math.sin(x * freq3) * Math.sin(y * freq3) * 0.25;

                    // Calculate normal from height differences
                    const scale = 2.0;
                    const dx =
                        Math.cos(x * freq1) * freq1 * scale +
                        Math.cos(x * freq2 + 0.5) * freq2 * scale * 0.5 +
                        Math.cos(x * freq3) * freq3 * scale * 0.25;
                    const dy =
                        -Math.sin(y * freq1) * freq1 * scale -
                        Math.sin(y * freq2 + 0.3) * freq2 * scale * 0.5 +
                        Math.cos(y * freq3) * freq3 * scale * 0.25;

                    // Convert to normal map format (0-255)
                    // Normal map: R=X, G=Y, B=Z (Z points up)
                    data[idx] = Math.floor((dx + 1) * 0.5 * 255); // R: X
                    data[idx + 1] = Math.floor((dy + 1) * 0.5 * 255); // G: Y
                    data[idx + 2] = 255; // B: Z (up)
                    data[idx + 3] = 255; // A
                }
            }

            ctx.putImageData(imageData, 0, 0);
        }

        // Create texture from canvas
        const texture = new Texture(
            bumpCanvas.toDataURL(),
            s,
            true, // noMipmap
            false, // invertY
            Texture.TRILINEAR_SAMPLINGMODE,
        );
        texture.wrapU = Texture.WRAP_ADDRESSMODE;
        texture.wrapV = Texture.WRAP_ADDRESSMODE;
        texture.uScale = 10; // Tile the wave pattern
        texture.vScale = 5;

        return texture;
    }

    function createWaterLayer(_s: Scene) {
        // TODO: Option 3 - Create transparent water sphere with:
        // - Fresnel reflections
        // - Animated displacement
        // - Refraction effects
        console.log(
            "[BabylonGlobe] Water layer placeholder - not yet implemented",
        );
    }

    function createStarfield(s: Scene) {
        // Create a large sphere for the starfield (inside-out rendering)
        const starSphere = MeshBuilder.CreateSphere(
            "starfield",
            { segments: 32, diameter: 100, sideOrientation: 1 }, // sideOrientation: 1 = BACKSIDE
            s,
        );

        // Create material for stars - emissive only (no lighting needed)
        const starMaterial = new StandardMaterial("starMaterial", s);
        starMaterial.diffuseColor = new Color3(0, 0, 0); // No diffuse
        starMaterial.specularColor = new Color3(0, 0, 0); // No specular
        starMaterial.disableLighting = true;

        // Create a procedural star texture using canvas
        const width = 2048;
        const height = 1024;
        const starCanvas = document.createElement("canvas");
        starCanvas.width = width;
        starCanvas.height = height;
        const ctx = starCanvas.getContext("2d");

        if (ctx) {
            // Dark background
            ctx.fillStyle = "#030308";
            ctx.fillRect(0, 0, width, height);

            // Generate random stars
            const numStars = 3000;
            for (let i = 0; i < numStars; i++) {
                const x = Math.random() * width;
                const y = Math.random() * height;
                const size = Math.random() * 2 + 0.5;
                const brightness = Math.random() * 0.7 + 0.3;

                // Star color variation (mostly white, some blue/yellow)
                const colorVar = Math.random();
                let r = brightness,
                    g = brightness,
                    b = brightness;
                if (colorVar < 0.1) {
                    r *= 0.7;
                    g *= 0.8;
                    b *= 1.0; // Blue star
                } else if (colorVar < 0.2) {
                    r *= 1.0;
                    g *= 0.9;
                    b *= 0.6; // Yellow star
                }

                ctx.fillStyle = `rgb(${Math.floor(r * 255)}, ${Math.floor(g * 255)}, ${Math.floor(b * 255)})`;
                ctx.beginPath();
                ctx.arc(x, y, size, 0, Math.PI * 2);
                ctx.fill();
            }

            // Add a few brighter stars with glow
            for (let i = 0; i < 50; i++) {
                const x = Math.random() * width;
                const y = Math.random() * height;
                const size = Math.random() * 3 + 2;

                const gradient = ctx.createRadialGradient(
                    x,
                    y,
                    0,
                    x,
                    y,
                    size * 2,
                );
                gradient.addColorStop(0, "rgba(255, 255, 255, 1)");
                gradient.addColorStop(0.3, "rgba(200, 220, 255, 0.8)");
                gradient.addColorStop(1, "rgba(100, 150, 255, 0)");

                ctx.fillStyle = gradient;
                ctx.beginPath();
                ctx.arc(x, y, size * 2, 0, Math.PI * 2);
                ctx.fill();
            }

            // Get image data and create RawTexture
            const imageData = ctx.getImageData(0, 0, width, height);
            const starTexture = RawTexture.CreateRGBATexture(
                imageData.data,
                width,
                height,
                s,
                false, // generateMipMaps
                false, // invertY
                Texture.TRILINEAR_SAMPLINGMODE,
            );

            starMaterial.emissiveTexture = starTexture;
        );

        // Create material for stars - emissive only (no lighting needed)
        const starMaterial = new StandardMaterial("starMaterial", s);
        starMaterial.diffuseColor = new Color3(0, 0, 0); // No diffuse
        starMaterial.specularColor = new Color3(0, 0, 0); // No specular
        starMaterial.disableLighting = true;

        // Create a procedural star texture using canvas
        const width = 2048;
        const height = 1024;
        const starCanvas = document.createElement("canvas");
        starCanvas.width = width;
        starCanvas.height = height;
        const ctx = starCanvas.getContext("2d");

        if (ctx) {
            // Dark background
            ctx.fillStyle = "#030308";
            ctx.fillRect(0, 0, width, height);

            // Generate random stars
            const numStars = 3000;
            for (let i = 0; i < numStars; i++) {
                const x = Math.random() * width;
                const y = Math.random() * height;
                const size = Math.random() * 2 + 0.5;
                const brightness = Math.random() * 0.7 + 0.3;
                
                // Simple white/blue stars
                ctx.fillStyle = `rgba(255, 255, 255, ${brightness})`;
                ctx.beginPath();
                ctx.arc(x, y, size, 0, Math.PI * 2);
                ctx.fill();
            }

            // Get image data and create RawTexture
            const imageData = ctx.getImageData(0, 0, width, height);
            const starTexture = RawTexture.CreateRGBATexture(
                imageData.data,
                width,
                height,
                s,
                false, // generateMipMaps
                false, // invertY
                Texture.TRILINEAR_SAMPLINGMODE,
            );

            starMaterial.emissiveTexture = starTexture;
        }

        starSphere.material = starMaterial;
    }



    async function updateTexture(blob: Blob) {
        if (!scene || !globeMaterial) return;

        // Dispose old texture and object URL
        if (globeTexture) {
            globeTexture.dispose();
            globeTexture = null;
        }
        if (objectUrl) {
            URL.revokeObjectURL(objectUrl);
            objectUrl = null;
        }

        try {
            // Create object URL and load as image
            objectUrl = URL.createObjectURL(blob);
            const img = new Image();

            await new Promise<void>((resolve, reject) => {
                img.onload = () => resolve();
                img.onerror = () => reject(new Error("Failed to load image"));
                img.src = objectUrl!;
            });

            console.log(
                `[BabylonGlobe] Image loaded: ${img.width}x${img.height}`,
            );

            // Draw to canvas to get pixel data
            const tempCanvas = document.createElement("canvas");
            tempCanvas.width = img.width;
            tempCanvas.height = img.height;
            const ctx = tempCanvas.getContext("2d");

            if (!ctx) {
                console.error("[BabylonGlobe] Failed to get canvas context");
                return;
            }

            // Flip vertically for WebGL
            ctx.translate(0, img.height);
            ctx.scale(1, -1);
            ctx.drawImage(img, 0, 0);

            const imageData = ctx.getImageData(0, 0, img.width, img.height);
            const pixels = imageData.data;

            // Generate specular map
            const specularData = new Uint8ClampedArray(pixels.length);
            for (let i = 0; i < pixels.length; i += 4) {
                 const r = pixels[i];
                 const g = pixels[i+1];
                 const b = pixels[i+2];
                 // Water detection (blue dominant)
                 const isWater = b > r + 20 && b > g + 10;
                 const specular = isWater ? 200 : 20;
                 
                 specularData[i] = specular;
                 specularData[i+1] = specular;
                 specularData[i+2] = specular;
                 specularData[i+3] = 255;
            }

            // Create diffuse texture
            globeTexture = RawTexture.CreateRGBATexture(
                imageData.data,
                img.width,
                img.height,
                scene!,
                true,
                false,
                Texture.TRILINEAR_SAMPLINGMODE,
            );

            // Create specular texture
            const specularTexture = RawTexture.CreateRGBATexture(
                specularData,
                img.width,
                img.height,
                scene!,
                true,
                false,
                Texture.TRILINEAR_SAMPLINGMODE,
            );

            globeMaterial.diffuseTexture = globeTexture;
            globeMaterial.specularTexture = specularTexture;

            console.log("[BabylonGlobe] Planet texture updated");

        } catch (err) {
             console.error("[BabylonGlobe] Failed to update texture:", err);
        }
    }

    async function applyHeightDisplacement(blob: Blob) {
        if (!scene || !displacementShader) return;

        console.log(`[BabylonGlobe] Applying heightmap displacement from blob (${blob.size} bytes)`);

        try {
            const url = URL.createObjectURL(blob);
            // Load texture from blob URL
            const heightmapTexture = new Texture(url, scene);
            
            // Wait for load
            heightmapTexture.onLoadObservable.addOnce(() => {
                console.log("[BabylonGlobe] Heightmap texture loaded");
                
                // Update shader with texture and scale
                // Scale factor: maxElevation (8848m) / Earth Radius (6371000m) * Globe Radius (1.0)
                // = ~0.0014. But we want exaggerated terrain.
                const material = displacementShader?.createMaterial(heightmapTexture, 0.05);
                
                // Apply shader material to all LOD meshes
                if (material && lodManager) {
                    // Apply to known LOD levels (0, 1, 2)
                    for (let i = 0; i <= 2; i++) {
                        const mesh = lodManager.getMesh(i);
                        if (mesh) {
                            mesh.material = material;
                            console.log(`[BabylonGlobe] Applied shader to LOD mesh ${i}`);
                        }
                    }
                }
                
                // Cleanup URL
                setTimeout(() => URL.revokeObjectURL(url), 1000); 
            });
            
        } catch (err) {
            console.error("[BabylonGlobe] Failed to apply heightmap:", err);
        }
    }

    onDestroy(() => {
        // Cleanup moons
        moonMeshes.forEach((m) => m.dispose());
        moonOrbitNodes.forEach((n) => n.dispose());
        moonMeshes = [];
        moonOrbitNodes = [];

        // Cleanup LOD manager (disposes all LOD meshes)
        if (lodManager) {
            lodManager.dispose();
            lodManager = null;
        }

        // Cleanup resources
        if (objectUrl) {
            URL.revokeObjectURL(objectUrl);
        }
        if (globeTexture) {
            globeTexture.dispose();
        }
        if (sunMesh) {
            sunMesh.dispose();
        }
        if (scene) {
            scene.dispose();
        }
        if (engine) {
            engine.dispose();
        }
    });
</script>

<canvas bind:this={canvas} class="globe-canvas"></canvas>

<style>
    .globe-canvas {
        width: 100%;
        height: 100%;
        display: block;
        touch-action: none; /* Prevent scroll on touch devices */
    }
</style>
