<script lang="ts">
    import { onMount, tick, onDestroy } from "svelte";
    import { gameStore } from "$lib/stores/game";
    import * as BABYLON from "@babylonjs/core";
    import { Engine } from "@babylonjs/core/Engines/engine";
    import { Scene } from "@babylonjs/core/scene";
    import {
        Vector3,
        Matrix,
        Color3,
        Color4,
    } from "@babylonjs/core/Maths/math";
    import { ArcRotateCamera } from "@babylonjs/core/Cameras/arcRotateCamera";
    import { HemisphericLight } from "@babylonjs/core/Lights/hemisphericLight";
    import { PointLight } from "@babylonjs/core/Lights/pointLight";
    import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
    import { Mesh } from "@babylonjs/core/Meshes/mesh";
    import { TransformNode } from "@babylonjs/core/Meshes/transformNode";
    import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
    import { Texture } from "@babylonjs/core/Materials/Textures/texture";
    import { RawTexture } from "@babylonjs/core/Materials/Textures/rawTexture";
    import { WaterEffects } from "./WaterEffects";
    import { PoiManager } from "./PoiManager"; // NEW
    import { LODManager } from "./LODManager";
    import type { PointOfInterest } from "$lib/types/pois";

    import { ParticleSystem } from "@babylonjs/core/Particles/particleSystem";

    // Interface Mode Management
    import { isTextMode } from "$lib/stores/ui";

    // View Mode Management
    import { ViewModeManager } from "./ViewModeManager";
    import { ViewMode } from "./interfaces";
    import { AsteroidManager } from "./AsteroidManager";
    import { MoltenPlanetShader } from "./MoltenPlanetShader";
    import { DisplacementShader } from "./DisplacementShader";
    import { TileGlobeManager } from "./TileGlobeManager";
    import { FPSTransitionController } from "./FPSTransitionController";
    import { FPSMovementController } from "./FPSMovementController";
    import { FPSPerformanceManager } from "./FPSPerformanceManager";
    import { FPSAccessibilityOptions } from "./FPSAccessibilityOptions";
    import { HorizonRenderer } from "./HorizonRenderer";
    import { PerformanceOverlay } from "./PerformanceOverlay";

    // Types for satellite/moon data
    interface Satellite {
        name: string;
        mass: number; // kg
        distance: number; // km from planet
    }

    // Props
    export let scene: Scene; // Injected by SceneManager
    export let globeTextureBlob: Blob | null = null;
    export let globeHeightmapBlob: Blob | null = null;
    export let materialBlob: Blob | null = null;
    export let iceBlob: Blob | null = null;
    export let normalMapBlob: Blob | null = null;
    export let seaLevel: number = 0;
    export let maxElevation: number = 8848;
    export let minElevation: number = -11000;
    export let planetRadius: number = 6.371e6; // Earth radius in meters

    // Visual exaggeration for orbital view (terrain would be invisible at true scale)
    // Scientific scale = elevationRange / planetRadius ≈ 0.003 for Earth
    // We multiply by this factor to make terrain visible from orbit
    // 15x gives ~0.045 displacement which shows terrain without extreme spikes
    const VISUAL_EXAGGERATION = 15;

    // Calculate dynamic displacement scale based on planet data
    $: elevationRange = Math.abs(maxElevation - minElevation) || 19345; // Default ~19km
    $: scientificDisplacementScale = elevationRange / planetRadius;
    $: orbitalDisplacementScale =
        scientificDisplacementScale * VISUAL_EXAGGERATION;
    // Clamp to reasonable range to prevent extreme values
    $: displacementScale = Math.max(
        0.001,
        Math.min(0.15, orbitalDisplacementScale),
    );

    $: console.log("[WorldController:Debug] Reactive scene prop:", scene);
    $: console.log("[WorldController:Debug] Reactive satellites:", satellites);

    export let satellites: Satellite[] = []; // Moon data from world
    export let rings: any = null; // Ring system data
    export let pois: PointOfInterest[] = []; // POI data from world
    export let simulationSpeed: number = 1.0; // Time multiplier for animations
    export let onSendCommand:
        | ((action: string, message?: string) => void)
        | null = null; // For tile requests

    // Set up tile command callback
    $: sendTileCommand = onSendCommand;

    // Internal state
    // canvas and engine owned by SceneManager
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

    // Tile system for high-resolution streaming
    let tileGlobeManager: TileGlobeManager | null = null;
    // sendTileCommand already declared as reactive statement above

    // View Mode Manager
    let viewModeManager: ViewModeManager | null = null;

    // FPS Mode state
    let fpsMode: boolean = false;
    let fpsTransitionController: FPSTransitionController | null = null;
    let fpsMovementController: FPSMovementController | null = null;
    let fpsPerformanceManager: FPSPerformanceManager | null = null;
    let performanceOverlay: PerformanceOverlay | null = null;
    let fpsAccessibility: FPSAccessibilityOptions | null = null;
    let horizonRenderer: HorizonRenderer | null = null;
    let waterEffects: WaterEffects | null = null;
    let poiManager: PoiManager | null = null; // NEW

    // Animation state
    let lastTime = 0;
    const PLANET_DAY_SECONDS = 10; // Planet completes rotation in 10 real seconds
    const PLANET_YEAR_SECONDS = 120; // Planet completes orbit in 120 real seconds
    const SUN_DISTANCE = 20; // Distance from sun to planet

    // Terrain exaggeration factor (makes mountains visible from space)
    const TERRAIN_SCALE = 0.05; // 5% of radius for max height

    // Interface Mode Lifecycle Management
    let isPaused = false;
    let pauseTimeoutId: ReturnType<typeof setTimeout> | null = null;
    let unsubscribeTextMode: (() => void) | null = null;

    // Molten State
    let isMoltenState: boolean = true;
    let moltenShader: MoltenPlanetShader | null = null;

    // Camera Focus State - for click-to-focus on planet/moons
    type FocusTarget = "planet" | "moon" | null;
    let focusTarget: FocusTarget = "planet"; // Default focus on planet
    let focusedMoonIndex: number = -1; // Which moon is focused (-1 = none)
    let lastClickTime: number = 0; // For detecting double-click on focused object

    // Subscribe to text mode changes for pausing/resuming WebGL
    $: if (typeof window !== "undefined") {
        unsubscribeTextMode?.();
        unsubscribeTextMode = isTextMode.subscribe((inTextMode) => {
            if (inTextMode && !isPaused) {
                pauseRendering();
            } else if (!inTextMode && isPaused) {
                resumeRendering();
            }
        });
    }

    function pauseRendering() {
        if (isPaused) return;

        console.log("[WorldController] Pausing updates (TEXT mode)");
        isPaused = true;

        // We don't control the engineLoop, but we can stop updating our animations
        // SceneManager handles the actual engine stop if needed
    }

    function resumeRendering() {
        if (!isPaused) return;

        console.log("[WorldController] Resuming updates (VISUAL mode)");
        isPaused = false;
        lastTime = performance.now();
    }

    function registerRenderLoop() {
        if (!scene) return;

        // Add animation update to before-render
        scene.onBeforeRenderObservable.add(() => {
            if (isPaused || !scene) return;

            const now = performance.now();
            const deltaTime = (now - lastTime) / 1000; // seconds
            lastTime = now;

            updateAnimation(deltaTime);
        });

        // Start the engine render loop
        const engine = scene.getEngine();
        if (engine) {
            engine.runRenderLoop(() => {
                scene.render();
            });
            console.log("[WorldController] Render loop started");
        }
    }

    function updateAnimation(deltaTime: number) {
        if (!orbitNode || !planetNode) return;

        // Planet axial rotation (day cycle)
        const dayRotation =
            (2 * Math.PI * deltaTime * simulationSpeed) / PLANET_DAY_SECONDS;
        planetNode.rotation.y += dayRotation;

        // Orbital rotation around sun (year cycle)
        const yearRotation =
            (2 * Math.PI * deltaTime * simulationSpeed) / PLANET_YEAR_SECONDS;
        orbitNode.rotateAround(Vector3.Zero(), Vector3.Up(), yearRotation);

        // Update camera target to follow planet
        // Update camera target to follow planet
        if (camera) {
            camera.target = orbitNode.position;
            // Update LOD based on camera distance
            lodManager?.update(camera);

            // Map updates handled by ViewModeManager
            if (viewModeManager) {
                viewModeManager.update(camera, deltaTime);
            }

            // Update tile system for high-resolution streaming
            if (tileGlobeManager) {
                // Determine enablement via ViewModeManager callback,
                // but we still need to update it if enabled
                // (Note: tileGlobeManager.update checks for enablement internally too?)
                // Let's call update if it's enabled or just call it and let it decide.
                // But previously we enabled/disabled here.
                // Now ViewModeManager callback handles enable/disable.
                // We just call update.
                tileGlobeManager.update(camera);
            }
        }

        // Moon orbital animation
        moonOrbitNodes.forEach((moonOrbit, i) => {
            const moonPeriod = 5 + i * 2;
            const moonRotation =
                (2 * Math.PI * deltaTime * simulationSpeed) / moonPeriod;
            moonOrbit.rotation.y += moonRotation;
        });

        // Animate water bump texture UV offset
        waterTime += deltaTime * 0.05 * simulationSpeed;
        if (waterBumpTexture) {
            waterBumpTexture.uOffset = Math.sin(waterTime) * 0.02;
            waterBumpTexture.vOffset = Math.cos(waterTime * 0.7) * 0.01;
        }

        // Update sun direction for displacement shader lighting
        if (displacementShader && orbitNode) {
            const planetPos = orbitNode.position;
            const toSun = new Vector3(-planetPos.x, -planetPos.y, -planetPos.z);
            toSun.normalize();

            if (planetNode) {
                const worldMatrix = planetNode.getWorldMatrix();
                const invWorld = worldMatrix.clone().invert();
                const localSunDir = Vector3.TransformNormal(toSun, invWorld);
                localSunDir.normalize();
                displacementShader.setLightDirection(localSunDir);
            }
        }
    }

    onMount(() => {
        if (!scene) {
            console.error("[WorldController] No scene provided!");
            return;
        }

        console.log("[WorldController] Initializing world scene...");

        // Setup molten state if no texture
        if (!globeTextureBlob && !globeHeightmapBlob) {
            isMoltenState = true;
            moltenShader = new MoltenPlanetShader(scene);
        }

        // Initialize POI Manager
        poiManager = new PoiManager(scene, 10, (poi) => {
            console.log(
                "[WorldController] POI clicked:",
                poi.name,
                poi.coordinates,
            );
            if (fpsTransitionController) {
                // Transition to ground at POI location
                // We use the POI's lat/lon directly
                fpsTransitionController.transitionToGround(
                    poi.coordinates.lat,
                    poi.coordinates.lon,
                );

                // Enable FPV mode flag
                fpsMode = true;
            }
        });

        // Request POIs if empty
        if (pois.length === 0 && onSendCommand) {
            console.log("[WorldController] Requesting POIs...");
            onSendCommand("get_pois", JSON.stringify({ limit: 50 }));
        }

        // Request satellites/rings if empty
        if (satellites.length === 0 && onSendCommand) {
            console.log("[WorldController] Requesting Satellites...");
            onSendCommand("get_satellites");
        }

        // ===========================================
        // Solar System Hierarchy
        // ===========================================

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
        ambient.intensity = 0.6;
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

        camera.attachControl(scene.getEngine().getRenderingCanvas(), true);
        camera.lowerRadiusLimit = 1.05; // Allow very close zoom (just above surface)
        camera.upperRadiusLimit = 50; // Allow zooming out to see sun
        camera.wheelPrecision = 30; // Scroll zoom sensitivity
        camera.panningSensibility = 0; // Disable panning, only rotate
        camera.minZ = 0.01; // Near clip plane - prevents clipping at close range

        // Set as active camera (overrides default camera from scene factory)
        scene.activeCamera = camera;

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

        // Update POI manager radius (LOD mesh 0 is the main one, size should be consistent)
        // LODManager creates sphere with diameter 20 (radius 10) by default in its logic?
        // Let's check LODManager or just set it to 10 as per previous logic.
        if (poiManager) {
            poiManager.setRadius(10);
        }
        globe.parent = planetNode;

        const mediumMesh = lodManager.createMesh(scene, 1, "globe");
        mediumMesh.parent = planetNode;
        mediumMesh.setEnabled(false);

        const lowMesh = lodManager.createMesh(scene, 2, "globe");
        lowMesh.parent = planetNode;
        lowMesh.setEnabled(false);

        // Initialize shader material with default parameters (grey, flat)
        // This ensures spherical UV mapping is used even before textures load
        displacementShader.createMaterial(null);
        const defaultMat = displacementShader.getMaterial();

        // Apply initial material to all meshes
        if (isMoltenState && moltenShader) {
            const moltenMat = moltenShader.getMaterial();
            globe.material = moltenMat;
            mediumMesh.material = moltenMat;
            lowMesh.material = moltenMat;
            console.log("[WorldController] Applied Molten Planet Shader");
        } else if (defaultMat) {
            globe.material = defaultMat;
            mediumMesh.material = defaultMat;
            lowMesh.material = defaultMat;
        }

        // ===========================================
        // Create Moons (if any satellites provided)
        // ===========================================
        createMoons(scene);

        // ===========================================
        // Create Rings (if any provided)
        // ===========================================
        createRings(scene);

        // Create starfield background
        createStarfield(scene);

        // Initialize tile streaming system (if command callback is provided)
        tileGlobeManager = new TileGlobeManager(
            scene,
            planetNode,
            sendTileCommand,
            {
                maxLevel: 4,
                maxActiveTiles: 50,
            },
        );
        console.log("[BabylonGlobe] Tile system initialized");

        // ===========================================
        // Initialize View Mode Manager
        // ===========================================
        viewModeManager = new ViewModeManager({
            orbitThreshold: 2.5,
            terrainThreshold: 1.2,
        });

        viewModeManager.onModeChange((mode, prevMode) => {
            console.log(`[ViewMode] Changed from ${prevMode} to ${mode}`);

            // Handle Tile System
            if (tileGlobeManager) {
                if (mode === ViewMode.TILE) {
                    tileGlobeManager.enable();
                } else if (mode === ViewMode.ORBIT) {
                    tileGlobeManager.disable();
                } else if (mode === ViewMode.TERRAIN) {
                    // Keep active during transition, or handle specially
                    // Usually we might want tiles for far terrain in FPS mode
                    // But for now let's disable to save resources for FPS chunks
                    tileGlobeManager.disable();
                }
            }

            // Handle FPS Mode Transition (Scroll-based)
            if (mode === ViewMode.TERRAIN && prevMode !== ViewMode.TERRAIN) {
                if (fpsTransitionController && !fpsMode) {
                    console.log(
                        "[ViewMode] Triggering scroll-based FPS transition",
                    );
                    // Use new method to transition from current camera look-at
                    if (camera) {
                        fpsTransitionController.transitionFromCurrentPosition();
                    }
                }
            }

            // Handle Exit FPS Mode (Scroll out)
            if (mode !== ViewMode.TERRAIN && prevMode === ViewMode.TERRAIN) {
                if (fpsMode) {
                    console.log("[ViewMode] Exiting FPS mode via scroll");
                    // Return to orbit? Or just disable FPS flag?
                    if (fpsTransitionController) {
                        fpsTransitionController.returnToOrbit();
                        fpsMode = false;
                    }
                }
            }
        });

        // ===========================================
        // Initialize FPS Mode Components
        // ===========================================
        if (scene && camera) {
            const engine = scene.getEngine();
            // Transition controller for orbit-to-ground
            fpsTransitionController = new FPSTransitionController(
                scene,
                camera,
                {
                    onStateChange: (state) => {
                        console.log(`[BabylonGlobe] FPS state: ${state}`);
                        fpsMode = state === "flying" || state === "ground";
                    },
                },
            );

            // Performance manager
            fpsPerformanceManager = new FPSPerformanceManager(
                scene,
                engine as any,
                {
                    targetFps: 60,
                    minFps: 45,
                    enableAutoResolution: true,
                },
            );

            // Accessibility options
            fpsAccessibility = new FPSAccessibilityOptions();

            // Horizon renderer (sky dome)
            horizonRenderer = new HorizonRenderer(scene);

            // Water effects
            waterEffects = new WaterEffects(scene);

            // Performance overlay (starts hidden)
            const canvas = engine.getRenderingCanvas();
            if (canvas) {
                performanceOverlay = new PerformanceOverlay(
                    canvas.parentElement || document.body,
                    {
                        position: "top-right",
                        opacity: 0.85,
                    },
                );
            }

            console.log("[BabylonGlobe] FPS mode components initialized");
        }

        // Set up double-click handler for FPS entry
        const handleDoubleClick = (evt: MouseEvent) => {
            if (fpsTransitionController && !fpsMode) {
                const result = fpsTransitionController.handlePlanetClick(
                    evt.clientX,
                    evt.clientY,
                );
                if (result) {
                    console.log("[BabylonGlobe] Starting FPS transition");
                }
            }
        };

        // Single-click handler for focus selection (planet/moon)
        const handleFocusClick = (evt: MouseEvent) => {
            if (!scene || !camera || fpsMode) return;

            const pickResult = scene.pick(evt.clientX, evt.clientY);
            if (!pickResult || !pickResult.hit || !pickResult.pickedMesh)
                return;

            const meshName = pickResult.pickedMesh.name;
            const now = performance.now();

            // Check if clicked on planet (globe)
            if (meshName.startsWith("globe")) {
                if (focusTarget === "planet" && now - lastClickTime < 500) {
                    // Double-click on already focused planet → enter FPS
                    // Handled by handleDoubleClick already
                } else {
                    focusTarget = "planet";
                    focusedMoonIndex = -1;
                    updateCameraForSunAlignment("planet");
                    console.log("[WorldController] Focus: Planet");
                }
                lastClickTime = now;
                return;
            }

            // Check if clicked on a moon
            const moonMatch = meshName.match(/^moon_(.+)$/);
            if (moonMatch) {
                const moonIndex = moonMeshes.findIndex(
                    (m) => m.name === meshName,
                );
                if (moonIndex >= 0) {
                    if (
                        focusTarget === "moon" &&
                        focusedMoonIndex === moonIndex &&
                        now - lastClickTime < 500
                    ) {
                        // Double-click on already focused moon → could enter moon FPV (future)
                        console.log(
                            "[WorldController] Double-click on moon - FPV not yet implemented",
                        );
                    } else {
                        focusTarget = "moon";
                        focusedMoonIndex = moonIndex;
                        updateCameraForSunAlignment("moon", moonIndex);
                        console.log(
                            `[WorldController] Focus: Moon ${moonIndex}`,
                        );
                    }
                    lastClickTime = now;
                }
            }
        };

        // Position camera between target (planet/moon) and sun
        function updateCameraForSunAlignment(
            target: "planet" | "moon",
            moonIndex: number = -1,
        ) {
            if (!camera || !orbitNode) return;

            let targetPosition: Vector3;

            if (target === "planet") {
                targetPosition = orbitNode.position.clone();
            } else if (
                target === "moon" &&
                moonIndex >= 0 &&
                moonMeshes[moonIndex]
            ) {
                // Get moon's world position
                const moon = moonMeshes[moonIndex];
                targetPosition = moon.getAbsolutePosition();
            } else {
                return;
            }

            // Sun is at origin (0, 0, 0)
            const sunPosition = Vector3.Zero();

            // Direction from sun to target
            const sunToTarget = targetPosition
                .subtract(sunPosition)
                .normalize();

            // Camera should be positioned on the opposite side of target from sun
            // At a distance of camera.radius from target
            const cameraDistance = camera.radius;
            const newCameraPosition = targetPosition.add(
                sunToTarget.scale(cameraDistance),
            );

            // Animate camera to new position
            camera.target = targetPosition;

            // Calculate alpha/beta from new position
            const direction = newCameraPosition.subtract(targetPosition);
            const alpha = Math.atan2(direction.z, direction.x);
            const horizontalDist = Math.sqrt(
                direction.x * direction.x + direction.z * direction.z,
            );
            const beta = Math.atan2(horizontalDist, direction.y);

            camera.alpha = alpha;
            camera.beta = beta;

            console.log(
                `[WorldController] Camera aligned: sun at back, target=${target}`,
            );
        }

        const canvas = scene.getEngine().getRenderingCanvas();
        if (canvas) {
            canvas.addEventListener("dblclick", handleDoubleClick);
            canvas.addEventListener("click", handleFocusClick);
        }

        // Escape key to exit FPS mode
        const handleKeyDown = (evt: KeyboardEvent) => {
            if (evt.key === "Escape" && fpsMode && fpsTransitionController) {
                fpsTransitionController.returnToOrbit();
                fpsMode = false;
                console.log("[BabylonGlobe] Exiting FPS mode");
            }
            // Toggle performance overlay with F3
            if (evt.key === "F3" && performanceOverlay) {
                performanceOverlay.toggle();
            }
        };
        window.addEventListener("keydown", handleKeyDown);

        console.log("[BabylonGlobe] Solar system initialized");

        // Check if globeTextureBlob was already set before scene was ready
        // Set lastAppliedBlobSize to prevent reactive block from also firing
        if (globeTextureBlob && globeMaterial) {
            console.log(
                "[BabylonGlobe] Found existing globeTextureBlob, applying now...",
            );
            lastAppliedBlobSize = globeTextureBlob.size;
            updateTexture(globeTextureBlob);
        }

        // Animation and render loop
        // Animation and render loop logic registered to scene
        lastTime = performance.now();
        registerRenderLoop();

        // Window resize handled by SceneManager
    });

    // React to texture blob changes (with guard to avoid re-applying same blob)
    // Also reapply if objectUrl is null (component remounted but blob cached)
    // NOTE: Use displacementShader instead of globeMaterial - we now use shader-based rendering
    $: if (
        globeTextureBlob &&
        scene &&
        displacementShader &&
        (globeTextureBlob.size !== lastAppliedBlobSize || !objectUrl)
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

    // Watch for material blob changes (data-driven terrain coloring)
    // Also ensure shader material is ready
    $: if (
        materialBlob &&
        scene &&
        displacementShader &&
        displacementShader.getMaterial()
    ) {
        applyMaterialTexture(materialBlob);
    }

    // Watch for normal map blob changes
    // Also ensure shader material is ready
    $: if (
        normalMapBlob &&
        scene &&
        displacementShader &&
        displacementShader.getMaterial()
    ) {
        applyNormalMap(normalMapBlob);
    }

    // Watch for ice blob changes (glacier/polar visualization)
    // Also ensure shader material is ready
    $: if (
        iceBlob &&
        scene &&
        displacementShader &&
        displacementShader.getMaterial()
    ) {
        applyIceTexture(iceBlob);
    }

    // Track last created moon count to avoid unnecessary rebuilds
    let lastMoonCount = 0;

    // Reactively update moons when satellites data changes
    $: if (
        scene &&
        planetNode &&
        satellites &&
        satellites.length !== lastMoonCount
    ) {
        lastMoonCount = satellites.length;
        createMoons(scene);
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
            // Create orbit node for this moon
            const moonOrbit = new TransformNode(`moonOrbit_${sat.name}`, s);
            // Parent to planetNode so moons orbit with the planet
            moonOrbit.parent = planetNode;

            // Calculate moon distance based on satellite data
            // Distance is in METERS (e.g., 186 million meters = 186,000 km)
            // Use logarithmic scaling to compress vast distances into visible range (2-5 units)
            // log10(1e8) = 8, log10(1e9) = 9, log10(1e10) = 10
            // Map log range [7, 10] to normalized distance [1.5, 5]
            const distanceMeters = Math.max(sat.distance, 1e7); // Min 10,000 km
            const logDist = Math.log10(distanceMeters);
            const normalizedDistance = 1.5 + (logDist - 7) * 1.0; // 1e7m -> 1.5, 1e10m -> 4.5

            // Clamp to reasonable range
            const clampedDistance = Math.max(
                1.5,
                Math.min(5.0, normalizedDistance),
            );

            // Create moon mesh
            // Size based on mass - use logarithmic scale
            // Moon mass range: ~1e20 to 1e23 kg (Earth's moon is 7.3e22 kg)
            const massClamped = Math.max(sat.mass, 1e18);
            const logMass = Math.log10(massClamped);
            // Map log range [18, 23] to size [0.1, 0.4]
            const moonSize = 0.1 + ((logMass - 18) / 5) * 0.3;
            const clampedSize = Math.max(0.1, Math.min(0.4, moonSize));

            console.log(
                `[BabylonGlobe] Moon ${sat.name}: dist=${(sat.distance / 1e6).toFixed(0)}km, mass=${sat.mass.toExponential(1)}, normalizedDist=${clampedDistance.toFixed(2)}, size=${clampedSize.toFixed(3)}`,
            );

            const moon = MeshBuilder.CreateSphere(
                `moon_${sat.name}`,
                { segments: 16, diameter: clampedSize },
                s,
            );
            moon.parent = moonOrbit;
            moon.position = new Vector3(clampedDistance, 0, 0);

            // Moon material with procedural cratered texture
            const moonMat = new StandardMaterial(`moonMat_${sat.name}`, s);

            // Create procedural moon texture using canvas
            const moonTexSize = 256;
            const moonCanvas = document.createElement("canvas");
            moonCanvas.width = moonTexSize;
            moonCanvas.height = moonTexSize;
            const moonCtx = moonCanvas.getContext("2d");
            if (moonCtx) {
                // Base gray color with slight variation
                const baseGray = 140 + Math.floor(Math.random() * 40);
                moonCtx.fillStyle = `rgb(${baseGray}, ${baseGray}, ${baseGray - 5})`;
                moonCtx.fillRect(0, 0, moonTexSize, moonTexSize);

                // Add craters (darker circles with lighter rims)
                const numCraters = 20 + Math.floor(Math.random() * 30);
                for (let c = 0; c < numCraters; c++) {
                    const cx = Math.random() * moonTexSize;
                    const cy = Math.random() * moonTexSize;
                    const cr = 3 + Math.random() * 15;

                    // Crater shadow (darker)
                    const gradient = moonCtx.createRadialGradient(
                        cx,
                        cy,
                        0,
                        cx,
                        cy,
                        cr,
                    );
                    const shadowGray =
                        baseGray - 30 - Math.floor(Math.random() * 20);
                    const rimGray = baseGray + 10;
                    gradient.addColorStop(
                        0,
                        `rgb(${shadowGray}, ${shadowGray}, ${shadowGray})`,
                    );
                    gradient.addColorStop(
                        0.7,
                        `rgb(${shadowGray + 10}, ${shadowGray + 10}, ${shadowGray + 10})`,
                    );
                    gradient.addColorStop(
                        0.9,
                        `rgb(${rimGray}, ${rimGray}, ${rimGray})`,
                    );
                    gradient.addColorStop(
                        1,
                        `rgb(${baseGray}, ${baseGray}, ${baseGray})`,
                    );

                    moonCtx.beginPath();
                    moonCtx.arc(cx, cy, cr, 0, Math.PI * 2);
                    moonCtx.fillStyle = gradient;
                    moonCtx.fill();
                }

                // Add some noise texture
                const noiseData = moonCtx.getImageData(
                    0,
                    0,
                    moonTexSize,
                    moonTexSize,
                );
                for (let i = 0; i < noiseData.data.length; i += 4) {
                    const noise = (Math.random() - 0.5) * 10;
                    noiseData.data[i] = Math.max(
                        0,
                        Math.min(255, noiseData.data[i] + noise),
                    );
                    noiseData.data[i + 1] = Math.max(
                        0,
                        Math.min(255, noiseData.data[i + 1] + noise),
                    );
                    noiseData.data[i + 2] = Math.max(
                        0,
                        Math.min(255, noiseData.data[i + 2] + noise),
                    );
                }
                moonCtx.putImageData(noiseData, 0, 0);
            }

            // Create texture from canvas
            const moonTex = new Texture(
                moonCanvas.toDataURL(),
                s,
                false,
                false,
            );
            moonMat.diffuseTexture = moonTex;
            moonMat.specularColor = new Color3(0.1, 0.1, 0.1);
            moonMat.emissiveColor = new Color3(0.05, 0.05, 0.05); // Slight glow for visibility
            moon.material = moonMat;

            // Store references
            moonMeshes.push(moon);
            moonOrbitNodes.push(moonOrbit);

            // Start each moon at different orbital position
            moonOrbit.rotation.y = (index * Math.PI * 2) / satellites.length;
        });

        console.log(`[BabylonGlobe] Created ${moonMeshes.length} moon meshes`);
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

                // Star color variation
                const colorVar = Math.random();
                let r = brightness;
                let g = brightness;
                let b = brightness;

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
        if (!scene || !displacementShader) return;

        // Exit molten state if active
        if (isMoltenState) {
            console.log(
                "[WorldController] Transitioning from Molten State to Texture",
            );
            isMoltenState = false;

            // Switch meshes to displacement shader material
            if (displacementShader && displacementShader.getMaterial()) {
                const mat = displacementShader.getMaterial();
                if (globe && mat) globe.material = mat;
                if (lodManager && mat) {
                    for (let i = 0; i <= 2; i++) {
                        const mesh = lodManager.getMesh(i);
                        if (mesh) mesh.material = mat;
                    }
                }
            }

            if (moltenShader) {
                moltenShader.dispose();
                moltenShader = null;
            }
        }

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

            // Perform safe pixel access for specular map generation
            const specularData = new Uint8ClampedArray(pixels.length);
            for (let i = 0; i < pixels.length; i += 4) {
                // Ensure array index access is safe (TS safety)
                if (i + 3 >= pixels.length) break;

                const r = pixels[i];
                const g = pixels[i + 1];
                const b = pixels[i + 2];

                // Fallback for undefined which shouldn't happen with the break above
                if (r === undefined || g === undefined || b === undefined)
                    continue;

                // Water detection (blue dominant)
                const isWater = b > r + 20 && b > g + 10;
                const specular = isWater ? 200 : 20;

                specularData[i] = specular;
                specularData[i + 1] = specular;
                specularData[i + 2] = specular;
                specularData[i + 3] = 255;
            }

            // Create diffuse texture
            globeTexture = RawTexture.CreateRGBATexture(
                imageData.data,
                img.width,
                img.height,
                scene!,
                true, // Mipmaps required for TRILINEAR sampling and textureGrad
                false,
                Texture.TRILINEAR_SAMPLINGMODE,
            );

            // Pass texture to the shader
            if (displacementShader) {
                // Ensure material exists (might not have heightmap yet)
                if (!displacementShader.getMaterial()) {
                    console.log(
                        "[WorldController] Creating default DisplacementShader material",
                    );
                    displacementShader.createMaterial(null);
                }

                displacementShader.setDiffuseTexture(globeTexture);
                console.log(
                    "[WorldController] Texture applied to Displacement Shader",
                );
            } else {
                console.warn(
                    "[WorldController] Displacement Shader not ready for texture",
                );
            }

            // Use getMaterial() instead of accessing private property
            if (displacementShader && displacementShader.getMaterial()) {
                // Double check we are using the right material
            }

            // Standard Material Fallback REMOVED
            // We now rely solely on DisplacementShader to handle the texture rendering
            // to support the Icosphere geometry (preventing pole pinching).

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

            // Pass specular texture to the shader
            if (displacementShader) {
                displacementShader.setSpecularTexture(specularTexture);
            }

            console.log("[BabylonGlobe] Planet texture updated");
        } catch (err) {
            console.error("[BabylonGlobe] Failed to update texture:", err);
        }
    }

    async function applyHeightDisplacement(blob: Blob) {
        if (!scene || !displacementShader) return;

        console.log(
            `[BabylonGlobe] Applying heightmap displacement from blob (${blob.size} bytes)`,
        );

        try {
            const url = URL.createObjectURL(blob);
            // Load texture from blob URL
            const heightmapTexture = new Texture(url, scene);

            // Wait for load
            heightmapTexture.onLoadObservable.addOnce(() => {
                console.log("[BabylonGlobe] Heightmap texture loaded");

                // Check if material already exists - if so, just update the heightmap
                // This preserves existing texture flags (hasIceTex, hasMaterialTex, etc.)
                let material = displacementShader?.getMaterial();
                if (material) {
                    console.log(
                        "[BabylonGlobe] Updating existing shader heightmap (preserving textures)",
                    );
                    displacementShader?.updateHeightmap(heightmapTexture);

                    // FIX: Ensure we switch from molten mode if active
                    if (isMoltenState) {
                        console.log(
                            "[WorldController] Forced transition from Molten to Displacement (Heightmap update)",
                        );
                        isMoltenState = false;
                        if (moltenShader) {
                            moltenShader.dispose();
                            moltenShader = null;
                        }
                        if (lodManager) {
                            for (let i = 0; i <= 2; i++) {
                                const mesh = lodManager.getMesh(i);
                                if (mesh) mesh.material = material;
                            }
                        }
                    }
                } else {
                    // First time - create material with heightmap
                    // Use dynamic displacement scale based on planet radius and elevation range
                    console.log(
                        `[BabylonGlobe] Creating material with dynamic scale: ${displacementScale.toFixed(4)} (elev=${elevationRange.toFixed(0)}m, radius=${(planetRadius / 1e6).toFixed(2)}Mm, exaggeration=${VISUAL_EXAGGERATION}x)`,
                    );
                    material = displacementShader?.createMaterial(
                        heightmapTexture,
                        displacementScale,
                    );

                    // Apply shader material to all LOD meshes (only needed on first creation)
                    if (material && lodManager) {
                        for (let i = 0; i <= 2; i++) {
                            const mesh = lodManager.getMesh(i);
                            if (mesh) {
                                mesh.material = material;
                                console.log(
                                    `[BabylonGlobe] Applied shader to LOD mesh ${i}`,
                                );
                            }
                        }
                    }
                }

                // Set elevation range for proper sea level coloring
                displacementShader?.setElevationRange(
                    minElevation,
                    maxElevation,
                    seaLevel,
                );

                // Cleanup URL
                setTimeout(() => URL.revokeObjectURL(url), 1000);
            });
        } catch (err) {
            console.error("[BabylonGlobe] Failed to apply heightmap:", err);
        }
    }

    async function applyMaterialTexture(blob: Blob) {
        if (!scene || !displacementShader) return;

        // Validate blob has content
        if (!blob || blob.size === 0) {
            console.warn("[BabylonGlobe] Material blob is empty, skipping");
            return;
        }

        console.log(
            `[BabylonGlobe] Applying material texture from blob (${blob.size} bytes)`,
        );

        try {
            const url = URL.createObjectURL(blob);
            // Use constructor callbacks instead of observables
            const texture = new Texture(
                url,
                scene,
                false, // noMipmap
                false, // invertY
                Texture.TRILINEAR_SAMPLINGMODE,
                () => {
                    // onLoad callback
                    console.log("[BabylonGlobe] Material texture loaded");
                    displacementShader?.setMaterialTexture(texture);
                    setTimeout(() => URL.revokeObjectURL(url), 1000);

                    // FIX: Ensure we switch from molten mode if active
                    if (isMoltenState) {
                        console.log(
                            "[WorldController] Forced transition from Molten to Displacement (Material update)",
                        );
                        isMoltenState = false;
                        if (moltenShader) {
                            moltenShader.dispose();
                            moltenShader = null;
                        }
                        if (lodManager && displacementShader) {
                            const mat = displacementShader.getMaterial();
                            if (mat) {
                                for (let i = 0; i <= 2; i++) {
                                    const mesh = lodManager.getMesh(i);
                                    if (mesh) mesh.material = mat;
                                }
                            }
                        }
                    }

                    // FIX: Ensure we switch from molten mode if active
                    if (isMoltenState) {
                        console.log(
                            "[WorldController] Forced transition from Molten to Displacement (Material update)",
                        );
                        isMoltenState = false;
                        if (moltenShader) {
                            moltenShader.dispose();
                            moltenShader = null;
                        }
                        if (lodManager && displacementShader) {
                            const mat = displacementShader.getMaterial();
                            if (mat) {
                                for (let i = 0; i <= 2; i++) {
                                    const mesh = lodManager.getMesh(i);
                                    if (mesh) mesh.material = mat;
                                }
                            }
                        }
                    }
                },
                (message, exception) => {
                    // onError callback
                    console.error(
                        "[BabylonGlobe] Material texture failed to load:",
                        message,
                        exception,
                    );
                    URL.revokeObjectURL(url);
                },
            );
        } catch (err) {
            console.error(
                "[BabylonGlobe] Failed to apply material texture:",
                err,
            );
        }
    }

    async function applyNormalMap(blob: Blob) {
        if (!scene || !displacementShader) return;

        // Validate blob has content
        if (!blob || blob.size === 0) {
            console.warn("[BabylonGlobe] Normal map blob is empty, skipping");
            return;
        }

        console.log(
            `[BabylonGlobe] Applying normal map from blob (${blob.size} bytes)`,
        );

        try {
            const url = URL.createObjectURL(blob);
            // Use constructor callbacks instead of observables
            const texture = new Texture(
                url,
                scene,
                false, // noMipmap
                false, // invertY
                Texture.TRILINEAR_SAMPLINGMODE,
                () => {
                    // onLoad callback
                    console.log("[BabylonGlobe] Normal map loaded");
                    displacementShader?.setNormalMap(texture);
                    setTimeout(() => URL.revokeObjectURL(url), 1000);
                },
                (message, exception) => {
                    // onError callback
                    console.error(
                        "[BabylonGlobe] Normal map failed to load:",
                        message,
                        exception,
                    );
                    URL.revokeObjectURL(url);
                },
            );
        } catch (err) {
            console.error("[BabylonGlobe] Failed to apply normal map:", err);
        }
    }

    async function applyIceTexture(blob: Blob) {
        if (!scene || !displacementShader) return;

        // Validate blob has content
        if (!blob || blob.size === 0) {
            console.warn("[BabylonGlobe] Ice blob is empty, skipping");
            return;
        }

        console.log(
            `[BabylonGlobe] Applying ice texture from blob (${blob.size} bytes)`,
        );

        try {
            const url = URL.createObjectURL(blob);
            // Use constructor callbacks instead of observables
            const texture = new Texture(
                url,
                scene,
                false, // noMipmap
                false, // invertY
                Texture.TRILINEAR_SAMPLINGMODE,
                () => {
                    // onLoad callback
                    console.log("[BabylonGlobe] Ice texture loaded");
                    displacementShader?.setIceTexture(texture);
                    setTimeout(() => URL.revokeObjectURL(url), 1000);
                },
                (message, exception) => {
                    // onError callback
                    console.error(
                        "[BabylonGlobe] Ice texture failed to load:",
                        message,
                        exception,
                    );
                    URL.revokeObjectURL(url);
                },
            );
        } catch (err) {
            console.error("[BabylonGlobe] Failed to apply ice texture:", err);
        }
    }

    // Reactive: If POI update comes in
    $: if (poiManager && pois) {
        poiManager.updatePOIs(pois);
    }

    let ringMeshes: Mesh[] = [];

    function createRings(s: Scene) {
        if (!planetNode || !rings || !rings.rings || rings.rings.length === 0) {
            // No rings to create
            return;
        }

        console.log(
            `[BabylonGlobe] Creating ${rings.rings.length} planetary rings`,
        );

        // Clear existing rings
        ringMeshes.forEach((m) => m.dispose());
        ringMeshes = [];

        // Sort rings by radius to render inner to outer
        const sortedRings = [...rings.rings].sort(
            (a: any, b: any) => a.inner_radius - b.inner_radius,
        );

        sortedRings.forEach((ring: any, index: number) => {
            // Ring data is in METERS, simple scaling for visual representation
            // Planet radius is assumed 5 units in Babylon
            // Earth radius = 6371 km = 6,371,000 m
            // So scale = 5 / 6,371,000
            const scale = 5 / 6371000;
            const innerRadius = ring.inner_radius * scale;
            const outerRadius = ring.outer_radius * scale;

            // Create Torus for ring
            // Flatten it to look like a disc
            const ringMesh = MeshBuilder.CreateTorus(
                `ring_${ring.id}`,
                {
                    diameter: innerRadius + outerRadius, // Average diameter
                    thickness: outerRadius - innerRadius,
                    tessellation: 64,
                },
                s,
            );

            ringMesh.scaling.y = 0.001; // Flatten
            ringMesh.parent = planetNode;

            // Material
            const ringMat = new StandardMaterial(`ringMat_${ring.id}`, s);
            if (ring.color) {
                ringMat.diffuseColor = Color3.FromHexString(ring.color);
            } else {
                ringMat.diffuseColor = new Color3(0.6, 0.5, 0.4); // Default dusty color
            }

            // Make it slightly transparent
            ringMat.alpha = 0.7;
            ringMat.backFaceCulling = false; // Visible from both sides
            ringMat.specularColor = new Color3(0.1, 0.1, 0.1);

            ringMesh.material = ringMat;
            ringMeshes.push(ringMesh);
            ringMesh.material = ringMat;
            ringMeshes.push(ringMesh);
        });
    }

    // Monitor for simulation events (e.g., moon destruction)
    let processedEventIds = new Set<string>();
    // Asteroid Manager
    let asteroidManager: AsteroidManager | null = null;

    // Reactive: Handle Simulation Events (Transient)
    $: if ($gameStore.world?.sim?.events?.length > 0) {
        const events = $gameStore.world.sim.events;
        // Process latest event
        const latestEvent = events[events.length - 1]; // Simple approach, or use a queue

        if (latestEvent.type === "moon_destroyed") {
            console.log(
                "[WorldController] Processing Moon Destroyed event:",
                latestEvent,
            );
            destroyMoon(latestEvent.metadata.moon_id);
        } else if (latestEvent.type === "asteroid_impact") {
            console.log(
                "[WorldController] Processing Asteroid Impact event:",
                latestEvent,
            );
            if (asteroidManager) {
                asteroidManager.handleImpactEvent(latestEvent);
            }
        }
    }

    onMount(() => {
        if (!scene) return;

        // Initialize Asteroid Manager
        // Note: planetNode is set later in createPlanet/createSolarSystem?
        // Actually solarSystemRoot/planetNode logic is complex.
        // Let's defer passing planetNode until it's created, or pass null and set later.
        asteroidManager = new AsteroidManager(scene, globe); // 'globe' is the mesh

        // ... (existing initialization) ...
    });

    // Watch for globe creation to update manager
    $: if (asteroidManager && globe) {
        asteroidManager.setPlanetNode(globe);
    }

    function destroyMoon(moonNameID: string) {
        if (!scene) return;

        console.log(`[BabylonGlobe] Destroying moon: ${moonNameID}`);

        // Find moon mesh
        const moonMesh = moonMeshes.find(
            (m) =>
                m.name === `moon_${moonNameID}` || m.name.includes(moonNameID),
        );
        if (!moonMesh) {
            console.warn(
                `[BabylonGlobe] Moon mesh not found for destruction: ${moonNameID}`,
            );
            return;
        }

        // Create Explosion Effect
        const particleSystem = new ParticleSystem("moonExplosion", 2000, scene);
        // Use a default particle texture (flare or circle)
        // If we don't have one, we can create a dynamic texture or use a cloud URL if available
        // For now, let's try to use a base64 placeholder or assume a standard texture exists?
        // Actually, Babylon often needs a URL. We can use a generated data URI for a white circle.
        const particleTextureUrl =
            "https://raw.githubusercontent.com/BabylonJS/Babylon.js/master/packages/tools/playground/public/textures/flare.png"; // Fallback
        particleSystem.particleTexture = new Texture(particleTextureUrl, scene);

        particleSystem.emitter = moonMesh.position.clone();
        particleSystem.minEmitBox = new Vector3(-0.5, -0.5, -0.5);
        particleSystem.maxEmitBox = new Vector3(0.5, 0.5, 0.5);

        particleSystem.color1 = new Color4(0.7, 0.7, 0.7, 1.0);
        particleSystem.color2 = new Color4(0.2, 0.2, 0.2, 1.0);
        particleSystem.colorDead = new Color4(0, 0, 0, 0.0);

        particleSystem.minSize = 0.1;
        particleSystem.maxSize = 0.5;

        particleSystem.minLifeTime = 0.3;
        particleSystem.maxLifeTime = 1.5;

        particleSystem.emitRate = 1000;
        particleSystem.gravity = new Vector3(0, 0, 0);

        particleSystem.manualEmitCount = 500; // One-time burst
        particleSystem.start();

        // Dispose moon
        moonMesh.dispose();

        // Remove from our tracking array
        moonMeshes = moonMeshes.filter((m) => m !== moonMesh);

        // Optional: Leave behind a debris cloud (mesh)
        // For now, the particle burst is the visual.
    }

    // Reactive: If rings change (e.g. from update), recreate them
    $: if (rings && scene && planetNode) {
        createRings(scene);
    }

    onDestroy(() => {
        // Cleanup rings
        ringMeshes.forEach((m) => m.dispose());
        ringMeshes = [];

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

        // Cleanup FPS mode components
        if (fpsTransitionController) {
            fpsTransitionController = null;
        }
        if (fpsMovementController) {
            fpsMovementController.dispose();
            fpsMovementController = null;
        }
        if (fpsPerformanceManager) {
            fpsPerformanceManager.dispose();
            fpsPerformanceManager = null;
        }
        if (performanceOverlay) {
            performanceOverlay.dispose();
            performanceOverlay = null;
        }
        if (fpsAccessibility) {
            fpsAccessibility.dispose();
            fpsAccessibility = null;
        }
        if (horizonRenderer) {
            horizonRenderer.dispose();
            horizonRenderer = null;
        }
        if (waterEffects) {
            waterEffects.dispose();
            waterEffects = null;
        }

        if (moltenShader) {
            moltenShader.dispose();
            moltenShader = null;
        }

        if (poiManager) {
            poiManager.dispose();
            poiManager = null;
        }

        // Scene is managed by SceneManager, so we don't dispose it here
        // We only dispose nodes we created attached to the scene
        if (solarSystemRoot) {
            solarSystemRoot.dispose();
        }
    });
</script>

<slot></slot>

<style>
    /* No styles needed (canvas removed) */
</style>
