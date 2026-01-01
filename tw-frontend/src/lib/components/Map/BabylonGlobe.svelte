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

    // Types for satellite/moon data
    interface Satellite {
        name: string;
        mass: number; // kg
        distance: number; // km from planet
    }

    // Props
    export let textureBlob: Blob | null = null;
    export let heightData: ArrayBuffer | null = null; // Binary elevation data
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

    // Solar system nodes
    let solarSystemRoot: TransformNode | null = null;
    let orbitNode: TransformNode | null = null;
    let planetNode: TransformNode | null = null;
    let sunMesh: Mesh | null = null;
    let sunLight: PointLight | null = null;
    let moonMeshes: Mesh[] = [];
    let moonOrbitNodes: TransformNode[] = [];

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

        // Glow effect for sun
        const glowLayer = new GlowLayer("glow", scene);
        glowLayer.intensity = 0.8;
        glowLayer.addIncludedOnlyMesh(sunMesh);

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
        camera.lowerRadiusLimit = 1.5; // Prevent clipping into planet
        camera.upperRadiusLimit = 50; // Allow zooming out to see sun
        camera.wheelPrecision = 30; // Scroll zoom sensitivity
        camera.panningSensibility = 0; // Disable panning, only rotate

        // Create globe mesh as child of planetNode
        globe = MeshBuilder.CreateSphere(
            "globe",
            { segments: 128, diameter: 2, updatable: true },
            scene,
        );
        globe.parent = planetNode;

        // Create material for globe
        globeMaterial = new StandardMaterial("globeMaterial", scene);
        globeMaterial.diffuseColor = new Color3(0.2, 0.2, 0.25);
        globeMaterial.specularColor = new Color3(0.2, 0.2, 0.25);
        globeMaterial.specularPower = 32;
        globeMaterial.backFaceCulling = true;
        globe.material = globeMaterial;

        // ===========================================
        // Create Moons (if any satellites provided)
        // ===========================================
        createMoons(scene);

        // Create starfield background
        createStarfield(scene);

        console.log("[BabylonGlobe] Solar system initialized");

        // Check if textureBlob was already set before scene was ready
        if (textureBlob && globeMaterial) {
            console.log(
                "[BabylonGlobe] Found existing textureBlob, applying now...",
            );
            updateTexture(textureBlob);
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

    // React to texture blob changes (with guard to prevent re-applying same texture)
    $: if (
        textureBlob &&
        scene &&
        globeMaterial &&
        textureBlob.size !== lastAppliedBlobSize
    ) {
        console.log(
            "[BabylonGlobe] Reactive: textureBlob received, size:",
            textureBlob.size,
        );
        lastAppliedBlobSize = textureBlob.size;
        updateTexture(textureBlob);
    }

    // React to height data changes (with guard)
    $: if (
        heightData &&
        scene &&
        globe &&
        heightData.byteLength !== lastAppliedHeightDataLength
    ) {
        console.log("[BabylonGlobe] Reactive: heightData received");
        lastAppliedHeightDataLength = heightData.byteLength;
        applyHeightDisplacement(heightData);
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
            const canvas = document.createElement("canvas");
            canvas.width = img.width;
            canvas.height = img.height;
            const ctx = canvas.getContext("2d");

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

            // Generate specular map from color analysis
            const specularData = new Uint8ClampedArray(pixels.length);

            for (let i = 0; i < pixels.length; i += 4) {
                const r = pixels[i];
                const g = pixels[i + 1];
                const b = pixels[i + 2];

                // Calculate material type from color
                // Water: high blue relative to green/red, low brightness
                // Snow/Ice: high overall brightness (white)
                // Land: everything else (low specular)

                const brightness = (r + g + b) / 3;
                const isWater = b > r + 20 && b > g + 10 && brightness < 180;
                const isSnow =
                    brightness > 200 && r > 180 && g > 180 && b > 180;
                const isIce =
                    b > 150 && g > 130 && brightness > 140 && brightness < 220;

                let specular = 20; // Default low specular for land

                if (isWater) {
                    specular = 180; // High specular for water (shiny)
                } else if (isSnow) {
                    specular = 100; // Medium specular for snow
                } else if (isIce) {
                    specular = 140; // Higher specular for ice
                }

                // RGB specular color (grayscale)
                specularData[i] = specular; // R
                specularData[i + 1] = specular; // G
                specularData[i + 2] = specular; // B
                specularData[i + 3] = 255; // A
            }

            // Create diffuse texture from pixel data
            globeTexture = RawTexture.CreateRGBATexture(
                imageData.data,
                img.width,
                img.height,
                scene!,
                true, // generateMipMaps
                false, // invertY (already flipped)
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

            // Apply textures to material
            globeMaterial!.diffuseTexture = globeTexture;
            globeMaterial!.specularTexture = specularTexture;
            globeMaterial!.specularColor = new Color3(0.8, 0.8, 0.9); // Slightly blue tint for water reflections
            globeMaterial!.specularPower = 64; // Sharper highlights

            console.log(
                "[BabylonGlobe] Planet texture and specular map applied successfully",
            );
        } catch (err) {
            console.error("[BabylonGlobe] Texture load failed:", err);
        }
    }

    function applyHeightDisplacement(data: ArrayBuffer) {
        if (!globe || !scene) return;

        // Parse height data (assuming Float32Array of normalized elevations)
        // Format: width (u16), height (u16), then Float32 elevation values
        const view = new DataView(data);
        const gridWidth = view.getUint16(0, true);
        const gridHeight = view.getUint16(2, true);

        console.log(
            `[BabylonGlobe] Applying displacement: ${gridWidth}x${gridHeight}`,
        );

        // Get vertex positions from sphere
        const positions = globe.getVerticesData("position");
        const normals = globe.getVerticesData("normal");

        if (!positions || !normals) {
            console.error("[BabylonGlobe] No vertex data available");
            return;
        }

        const newPositions = new Float32Array(positions.length);
        const elevationRange = maxElevation - minElevation;

        // For each vertex, sample height and displace along normal
        for (let i = 0; i < positions.length; i += 3) {
            const x = positions[i];
            const y = positions[i + 1];
            const z = positions[i + 2];

            const nx = normals[i];
            const ny = normals[i + 1];
            const nz = normals[i + 2];

            // Convert vertex position to UV coordinates (equirectangular)
            // Position is on unit sphere, convert to lat/lon
            const len = Math.sqrt(x * x + y * y + z * z);
            const lat = Math.asin(y / len); // -PI/2 to PI/2
            const lon = Math.atan2(z, x); // -PI to PI

            // UV coordinates
            const u = (lon + Math.PI) / (2 * Math.PI); // 0 to 1
            const v = (lat + Math.PI / 2) / Math.PI; // 0 to 1

            // Sample height from grid
            const gridX = Math.floor(u * (gridWidth - 1));
            const gridY = Math.floor((1 - v) * (gridHeight - 1)); // Flip Y
            const idx = 4 + (gridY * gridWidth + gridX) * 4; // Skip header, 4 bytes per float

            let height = 0;
            if (idx + 4 <= data.byteLength) {
                height = view.getFloat32(idx, true);
            }

            // Normalize height to displacement (0 = sea level, varies by elevation)
            const normalizedHeight = (height - seaLevel) / elevationRange;
            const displacement = normalizedHeight * TERRAIN_SCALE;

            // Displace vertex along normal
            newPositions[i] = x + nx * displacement;
            newPositions[i + 1] = y + ny * displacement;
            newPositions[i + 2] = z + nz * displacement;
        }

        // Update mesh with new positions
        globe.updateVerticesData("position", newPositions);

        // Recompute normals for proper lighting
        globe.createNormals(true);

        console.log("[BabylonGlobe] Displacement applied successfully");
    }

    onDestroy(() => {
        // Cleanup moons
        moonMeshes.forEach((m) => m.dispose());
        moonOrbitNodes.forEach((n) => n.dispose());
        moonMeshes = [];
        moonOrbitNodes = [];

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
