<script lang="ts">
    // Colors matching WebGLMapRenderer.ts shaders
    // Converted from normalized vec3 (0-1) to CSS rgb(0-255)

    const terrainColors = [
        { label: "Summit", color: "rgb(250, 250, 250)", desc: "> 3000m" }, // 0.98, 0.98, 0.98
        { label: "Peak", color: "rgb(158, 158, 158)", desc: "2000-3000m" }, // 0.62, 0.62, 0.62
        { label: "High Mtn", color: "rgb(140, 110, 99)", desc: "1000-2000m" }, // 0.55, 0.43, 0.39
        { label: "Mountain", color: "rgb(161, 135, 128)", desc: "500-1000m" }, // 0.63, 0.53, 0.5
        { label: "Foothill", color: "rgb(196, 224, 166)", desc: "200-500m" }, // 0.77, 0.88, 0.65
        { label: "Plain", color: "rgb(102, 186, 107)", desc: "100-200m" }, // 0.4, 0.73, 0.42
        { label: "Lowland", color: "rgb(46, 125, 51)", desc: "0-100m" }, // 0.18, 0.49, 0.2
        { label: "Coast", color: "rgb(79, 195, 247)", desc: "Sea Level" }, // 0.31, 0.76, 0.97
    ];

    // Dynamic bathymetry colors (matching getBathymetricColor in shader)
    const oceanColors = [
        { label: "Shallow", color: "rgb(0, 153, 204)", desc: "0 to -500m" }, // vec3(0.0, 0.6, 0.8) turquoise
        {
            label: "Mid Ocean",
            color: "rgb(0, 77, 128)",
            desc: "-500 to -2000m",
        }, // vec3(0.0, 0.3, 0.5) ocean blue
        { label: "Deep Ocean", color: "rgb(0, 26, 51)", desc: "< -2000m" }, // vec3(0.0, 0.1, 0.2) navy
    ];

    // Water features (new shader biome IDs 11-13)
    const waterColors = [
        { label: "River", color: "rgb(51, 153, 255)", desc: "Freshwater" }, // vec3(0.2, 0.6, 1.0) bright blue
        { label: "Lake", color: "rgb(38, 102, 179)", desc: "Inland water" }, // vec3(0.15, 0.4, 0.7) deep blue
        { label: "Wetland", color: "rgb(77, 140, 128)", desc: "Marsh/swamp" }, // vec3(0.3, 0.55, 0.5) blue-green
    ];

    const biomeColors = [
        { label: "Ice / Tundra", color: "rgb(204, 217, 230)" }, // 0.8, 0.85, 0.9
        { label: "Taiga", color: "rgb(77, 128, 102)" }, // 0.3, 0.5, 0.4
        { label: "Alpine", color: "rgb(153, 140, 128)" }, // 0.6, 0.55, 0.5
        { label: "Deciduous", color: "rgb(128, 153, 102)" }, // 0.5, 0.6, 0.4
        { label: "Grassland", color: "rgb(102, 153, 77)" }, // 0.4, 0.6, 0.3
        { label: "Savanna", color: "rgb(179, 166, 89)" }, // vec3(0.7, 0.65, 0.35) golden
        { label: "Rainforest", color: "rgb(51, 128, 77)" }, // 0.2, 0.5, 0.3
        { label: "Desert", color: "rgb(230, 204, 128)" }, // 0.9, 0.8, 0.5
    ];

    // Lobby mode colors (entity IDs)
    const lobbyColors = [
        { label: "Wall", color: "rgb(89, 89, 102)", desc: "Structure" }, // vec3(0.35, 0.35, 0.4) dark grey
        { label: "Floor", color: "rgb(217, 209, 199)", desc: "Walkable" }, // vec3(0.85, 0.82, 0.78) marble
        { label: "Portal", color: "rgb(204, 51, 204)", desc: "Teleport" }, // vec3(0.8, 0.2, 0.8) magenta
    ];

    export let mode: "terrain" | "biome" = "terrain";
    export let isLobby: boolean = false;
    export let expanded = false;
</script>

<div
    class="bg-gray-900/90 backdrop-blur border border-gray-700 rounded-lg shadow-xl text-xs overflow-hidden transition-all duration-300"
    class:w-48={expanded}
    class:w-10={!expanded}
>
    <button
        class="w-full p-2 flex items-center justify-between hover:bg-gray-800 transition-colors"
        on:click={() => (expanded = !expanded)}
        title="Toggle Legend"
    >
        <div class="flex items-center gap-2">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4 text-blue-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 01-.447-.894L15 7m0 13V7"
                />
            </svg>
            {#if expanded}
                <span class="font-bold text-gray-200">Map Legend</span>
            {/if}
        </div>
        {#if expanded}
            <div class="text-gray-500">
                {#if isLobby}🏛️{:else if mode === "terrain"}⛰️{:else}🌿{/if}
            </div>
        {/if}
    </button>

    {#if expanded}
        <div
            class="p-3 border-t border-gray-800 space-y-4 max-h-[60vh] overflow-y-auto"
        >
            {#if isLobby}
                <!-- Lobby Mode -->
                <div>
                    <h4
                        class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider"
                    >
                        Lobby
                    </h4>
                    <div class="space-y-1">
                        {#each lobbyColors as item}
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-4 h-4 rounded shadow-sm border border-black/20"
                                    style="background-color: {item.color}"
                                ></div>
                                <div class="flex-1">
                                    <div class="text-gray-300">
                                        {item.label}
                                    </div>
                                    <div class="text-gray-600 text-[10px]">
                                        {item.desc}
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
            {:else}
                <!-- World Mode -->
                <div>
                    <h4
                        class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider"
                    >
                        Topography
                    </h4>
                    <div class="space-y-1">
                        {#each terrainColors as item}
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-4 h-4 rounded shadow-sm border border-black/20"
                                    style="background-color: {item.color}"
                                ></div>
                                <div class="flex-1">
                                    <div class="text-gray-300">
                                        {item.label}
                                    </div>
                                    <div class="text-gray-600 text-[10px]">
                                        {item.desc}
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>

                <div>
                    <h4
                        class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider"
                    >
                        Ocean Depth
                    </h4>
                    <div class="space-y-1">
                        {#each oceanColors as item}
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-4 h-4 rounded shadow-sm border border-black/20"
                                    style="background-color: {item.color}"
                                ></div>
                                <div class="flex-1">
                                    <div class="text-gray-300">
                                        {item.label}
                                    </div>
                                    <div class="text-gray-600 text-[10px]">
                                        {item.desc}
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>

                <div>
                    <h4
                        class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider"
                    >
                        Water Features
                    </h4>
                    <div class="space-y-1">
                        {#each waterColors as item}
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-4 h-4 rounded shadow-sm border border-black/20"
                                    style="background-color: {item.color}"
                                ></div>
                                <div class="flex-1">
                                    <div class="text-gray-300">
                                        {item.label}
                                    </div>
                                    <div class="text-gray-600 text-[10px]">
                                        {item.desc}
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>

                <div>
                    <h4
                        class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider"
                    >
                        Biomes
                    </h4>
                    <div class="space-y-1">
                        {#each biomeColors as item}
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-4 h-4 rounded shadow-sm border border-black/20"
                                    style="background-color: {item.color}"
                                ></div>
                                <div class="text-gray-300">{item.label}</div>
                            </div>
                        {/each}
                    </div>
                </div>
            {/if}
        </div>
    {/if}
</div>
