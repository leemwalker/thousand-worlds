<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { fade, slide } from "svelte/transition";
    import type { WorldCreationParams } from "$lib/types/WorldCreationParams";
    import { DEFAULT_WORLD_PARAMS } from "$lib/types/WorldCreationParams";

    export let isOpen = false;

    const dispatch = createEventDispatcher<{
        close: void;
        complete: WorldCreationParams;
    }>();

    let params: WorldCreationParams = { ...DEFAULT_WORLD_PARAMS };
    let showAdvanced = false;

    // Generate random name
    function randomizeName() {
        const prefixes = [
            "New",
            "Alpha",
            "Terra",
            "Nova",
            "Proxima",
            "Kepler",
            "Gliese",
            "Ross",
        ];
        const suffixes = [
            "Prime",
            "Centauri",
            "Major",
            "Minor",
            "Eridani",
            "B",
            "c",
            "d",
            "e",
        ];
        const randomPrefix =
            prefixes[Math.floor(Math.random() * prefixes.length)];
        const randomSuffix =
            suffixes[Math.floor(Math.random() * suffixes.length)];
        const randomNum = Math.floor(Math.random() * 999) + 1;
        params.name = `${randomPrefix} ${randomSuffix}-${randomNum}`;
    }

    // Initialize with a random name if empty
    $: if (isOpen && !params.name) {
        randomizeName();
    }

    function handleClose() {
        dispatch("close");
    }

    function handleComplete() {
        if (!params.name) return;
        dispatch("complete", params);
    }
</script>

{#if isOpen}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
        transition:fade
    >
        <div
            class="relative w-full max-w-2xl bg-slate-900 border border-slate-700 rounded-lg shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
        >
            <!-- Header -->
            <div
                class="flex items-center justify-between p-6 border-b border-slate-700 bg-slate-800/50"
            >
                <div>
                    <h2 class="text-xl font-bold text-white mb-1">
                        Genesis Protocol
                    </h2>
                    <p class="text-sm text-slate-400">
                        Configure simulation parameters for new world
                        generation.
                    </p>
                </div>
                <!-- svelte-ignore a11y-click-events-have-key-events -->
                <button
                    on:click={handleClose}
                    class="text-slate-400 hover:text-white transition-colors p-2"
                    aria-label="Close modal"
                >
                    <svg
                        class="w-6 h-6"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M6 18L18 6M6 6l12 12"
                        />
                    </svg>
                </button>
            </div>

            <!-- Content - Scrollable -->
            <div class="flex-1 overflow-y-auto p-6 space-y-8">
                <!-- 1. Identity -->
                <section class="space-y-4">
                    <label
                        for="world-name"
                        class="block text-sm font-medium text-slate-300"
                        >World Designation</label
                    >
                    <div class="flex gap-2">
                        <input
                            id="world-name"
                            type="text"
                            bind:value={params.name}
                            placeholder="Enter world name..."
                            class="flex-1 bg-slate-950 border border-slate-700 rounded-md px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-600"
                        />
                        <button
                            on:click={randomizeName}
                            class="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 rounded-md text-slate-300 transition-colors"
                            title="Generate Random Name"
                        >
                            🎲
                        </button>
                    </div>
                </section>

                <!-- 2. Physical Parameters -->
                <section class="space-y-4">
                    <h3
                        class="text-sm font-semibold text-blue-400 uppercase tracking-wider"
                    >
                        Physical Parameters
                    </h3>

                    <!-- Size -->
                    <div class="space-y-2">
                        <span class="block text-sm font-medium text-slate-300"
                            >Class (Size)</span
                        >
                        <div class="grid grid-cols-4 gap-2">
                            {#each ["small", "medium", "large", "huge"] as size}
                                <button
                                    class="flex flex-col items-center justify-center p-3 rounded-md border transition-all {params.size ===
                                    size
                                        ? 'bg-blue-600/20 border-blue-500 text-white'
                                        : 'bg-slate-800/50 border-slate-700 text-slate-400 hover:border-slate-600'}"
                                    on:click={() => (params.size = size)}
                                >
                                    <div
                                        class="w-8 h-8 rounded-full bg-current mb-2 opacity-80"
                                        style="transform: scale({size ===
                                        'small'
                                            ? 0.5
                                            : size === 'medium'
                                              ? 0.75
                                              : size === 'large'
                                                ? 1
                                                : 1.25})"
                                    ></div>
                                    <span class="text-xs capitalize"
                                        >{size}</span
                                    >
                                </button>
                            {/each}
                        </div>
                    </div>

                    <!-- Moon Count -->
                    <div class="space-y-2">
                        <div class="flex justify-between items-center">
                            <span class="text-sm font-medium text-slate-300"
                                >Natural Satellites</span
                            >
                            <span
                                class="text-xs font-mono bg-slate-800 px-2 py-1 rounded text-blue-300"
                                >{params.moonCount}</span
                            >
                        </div>
                        <input
                            type="range"
                            min="0"
                            max="4"
                            step="1"
                            bind:value={params.moonCount}
                            class="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-blue-500"
                        />
                        <div
                            class="flex justify-between text-xs text-slate-500"
                        >
                            <span>None</span>
                            <span>Multiple</span>
                        </div>
                    </div>
                </section>

                <!-- 3. Advanced Configuration (Collapsible) -->
                <div class="border-t border-slate-800 pt-4">
                    <button
                        on:click={() => (showAdvanced = !showAdvanced)}
                        class="flex items-center gap-2 text-sm text-slate-400 hover:text-white transition-colors w-full"
                    >
                        <svg
                            class="w-4 h-4 transition-transform {showAdvanced
                                ? 'rotate-90'
                                : ''}"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M9 5l7 7-7 7"
                            />
                        </svg>
                        Advanced Configuration
                    </button>

                    {#if showAdvanced}
                        <div
                            class="mt-4 space-y-6 pl-4 border-l-2 border-slate-800"
                            transition:slide
                        >
                            <!-- Core Type -->
                            <div class="space-y-2">
                                <span
                                    class="block text-sm font-medium text-slate-300"
                                    >Core Composition</span
                                >
                                <select
                                    bind:value={params.coreType}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-md px-3 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="continental"
                                        >Continental (Standard)</option
                                    >
                                    <option value="volcanic"
                                        >Volcanic (Active)</option
                                    >
                                    <option value="oceanic"
                                        >Oceanic (Water World)</option
                                    >
                                    <option value="ancient"
                                        >Ancient (Eroded)</option
                                    >
                                </select>
                            </div>

                            <!-- Resolution -->
                            <div class="space-y-2">
                                <span
                                    class="block text-sm font-medium text-slate-300"
                                    >Resolution (Grid Size)</span
                                >
                                <select
                                    bind:value={params.resolution}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-md px-3 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value={32}>32 (Fastest)</option>
                                    <option value={64}>64 (Fast)</option>
                                    <option value={128}>128 (Balanced)</option>
                                    <option value={256}
                                        >256 (High Detail)</option
                                    >
                                    <option value={512}>512 (Ultra)</option>
                                </select>
                            </div>

                            <!-- Water Level -->
                            <div class="space-y-2">
                                <span
                                    class="block text-sm font-medium text-slate-300"
                                    >Hydrosphere Level</span
                                >
                                <select
                                    bind:value={params.waterLevel}
                                    class="w-full bg-slate-950 border border-slate-700 rounded-md px-3 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="low">Low (Arid)</option>
                                    <option value="medium"
                                        >Medium (Earth-like)</option
                                    >
                                    <option value="high"
                                        >High (Archipelago)</option
                                    >
                                </select>
                            </div>

                            <!-- Manual Seed -->
                            <div class="space-y-2">
                                <span
                                    class="block text-sm font-medium text-slate-300"
                                    >Simulation Seed (Optional)</span
                                >
                                <input
                                    type="text"
                                    bind:value={params.seed}
                                    placeholder="Leave empty for random..."
                                    class="w-full bg-slate-950 border border-slate-700 rounded-md px-3 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                                />
                            </div>
                        </div>
                    {/if}
                </div>

                <!-- 4. Systems Toggles -->
                <section class="space-y-4 pt-4 border-t border-slate-800">
                    <h3
                        class="text-sm font-semibold text-blue-400 uppercase tracking-wider"
                    >
                        Active Systems
                    </h3>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <!-- Geology (Primary) -->
                        <label
                            class="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50 cursor-pointer"
                        >
                            <span class="text-sm text-slate-200">Geology</span>
                            <input
                                type="checkbox"
                                bind:checked={params.sysGeology}
                                class="accent-blue-500 w-5 h-5 rounded"
                            />
                        </label>

                        <!-- Dependent Systems -->
                        <label
                            class="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50 cursor-pointer {params.sysGeology
                                ? ''
                                : 'opacity-50 pointer-events-none'}"
                        >
                            <span class="text-sm text-slate-200"
                                >Weather & Climate</span
                            >
                            <input
                                type="checkbox"
                                bind:checked={params.sysWeather}
                                disabled={!params.sysGeology}
                                class="accent-blue-500 w-5 h-5 rounded"
                            />
                        </label>

                        <label
                            class="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50 cursor-pointer {params.sysGeology
                                ? ''
                                : 'opacity-50 pointer-events-none'}"
                        >
                            <span class="text-sm text-slate-200"
                                >Life & Evolution</span
                            >
                            <input
                                type="checkbox"
                                bind:checked={params.sysLife}
                                disabled={!params.sysGeology}
                                class="accent-blue-500 w-5 h-5 rounded"
                            />
                        </label>

                        <label
                            class="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50 cursor-pointer {params.sysLife
                                ? ''
                                : 'opacity-50 pointer-events-none'}"
                        >
                            <span class="text-sm text-slate-200">Pathogens</span
                            >
                            <input
                                type="checkbox"
                                bind:checked={params.sysDisease}
                                disabled={!params.sysLife}
                                class="accent-blue-500 w-5 h-5 rounded"
                            />
                        </label>

                        <label
                            class="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50 cursor-pointer {params.sysLife
                                ? ''
                                : 'opacity-50 pointer-events-none'}"
                        >
                            <span class="text-sm text-slate-200">Sapience</span>
                            <input
                                type="checkbox"
                                bind:checked={params.sysSapience}
                                disabled={!params.sysLife}
                                class="accent-blue-500 w-5 h-5 rounded"
                            />
                        </label>

                        <label
                            class="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50 cursor-pointer {params.sysLife
                                ? ''
                                : 'opacity-50 pointer-events-none'}"
                        >
                            <span class="text-sm text-slate-200">Migration</span
                            >
                            <input
                                type="checkbox"
                                bind:checked={params.sysMigration}
                                disabled={!params.sysLife}
                                class="accent-blue-500 w-5 h-5 rounded"
                            />
                        </label>
                    </div>
                </section>
            </div>

            <!-- Footer -->
            <div
                class="p-6 border-t border-slate-700 bg-slate-800/50 flex justify-end gap-3"
            >
                <button
                    on:click={handleClose}
                    class="px-4 py-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-md transition-colors"
                >
                    Cancel
                </button>
                <button
                    on:click={handleComplete}
                    disabled={!params.name}
                    class="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white font-medium rounded-md shadow-lg shadow-blue-900/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    Initialize Simulation
                </button>
            </div>
        </div>
    </div>
{/if}
