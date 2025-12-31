<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { Engine } from "@babylonjs/core/Engines/engine";
    import { Scene } from "@babylonjs/core/scene";
    import { ArcRotateCamera } from "@babylonjs/core/Cameras/arcRotateCamera";
    import { HemisphericLight } from "@babylonjs/core/Lights/hemisphericLight";
    import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
    import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";
    import { Texture } from "@babylonjs/core/Materials/Textures/texture";
    import { Vector3 } from "@babylonjs/core/Maths/math.vector";
    import { Color3, Color4 } from "@babylonjs/core/Maths/math.color";

    // Props
    export let textureBlob: Blob | null = null;

    // Internal state
    let canvas: HTMLCanvasElement;
    let engine: Engine | null = null;
    let scene: Scene | null = null;
    let camera: ArcRotateCamera | null = null;
    let globeTexture: Texture | null = null;
    let globeMaterial: StandardMaterial | null = null;
    let objectUrl: string | null = null;

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
        scene.clearColor = new Color4(0.02, 0.02, 0.06, 1); // Dark space #050510

        // Hemispheric light for even illumination
        const light = new HemisphericLight(
            "light",
            new Vector3(0, 1, 0),
            scene,
        );
        light.intensity = 1.0;
        light.groundColor = new Color3(0.3, 0.3, 0.4); // Slight ambient from below

        // ArcRotateCamera for orbit controls
        camera = new ArcRotateCamera(
            "camera",
            Math.PI / 2, // alpha (horizontal rotation)
            Math.PI / 2, // beta (vertical rotation)
            3, // radius (distance from target)
            Vector3.Zero(),
            scene,
        );
        camera.attachControl(canvas, true);
        camera.lowerRadiusLimit = 1.5; // Prevent clipping into planet
        camera.upperRadiusLimit = 10; // Don't lose planet
        camera.wheelPrecision = 50; // Scroll zoom sensitivity
        camera.panningSensibility = 0; // Disable panning, only rotate

        // Create globe mesh
        const globe = MeshBuilder.CreateSphere(
            "globe",
            { segments: 48, diameter: 2 },
            scene,
        );

        // Create material for globe
        globeMaterial = new StandardMaterial("globeMaterial", scene);
        globeMaterial.specularColor = new Color3(0.1, 0.1, 0.1); // Low specular
        globeMaterial.backFaceCulling = true;
        globe.material = globeMaterial;

        // Render loop
        engine.runRenderLoop(() => {
            if (scene) {
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

    // React to texture blob changes
    $: if (textureBlob && scene && globeMaterial) {
        updateTexture(textureBlob);
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

        // Create object URL from blob
        objectUrl = URL.createObjectURL(blob);

        // Create new texture
        globeTexture = new Texture(
            objectUrl,
            scene,
            false,
            true,
            Texture.TRILINEAR_SAMPLINGMODE,
            () => {
                console.log("[BabylonGlobe] Texture loaded successfully");
            },
            (message: string, exception: unknown) => {
                console.error(
                    "[BabylonGlobe] Texture load failed:",
                    message,
                    exception,
                );
            },
        );

        // Fix WebGL Y-flip (equirectangular textures are often inverted)
        globeTexture.vScale = -1;

        // Apply to material
        globeMaterial.diffuseTexture = globeTexture;
    }

    onDestroy(() => {
        // Cleanup
        if (objectUrl) {
            URL.revokeObjectURL(objectUrl);
        }
        if (globeTexture) {
            globeTexture.dispose();
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
