<script lang="ts">
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher<{
        command: string;
    }>();

    // Simulation parameters
    let years = 100_000_000;
    let seed = "";
    let waterLevel = "medium";
    let customWaterLevel = "";
    let moons = "";

    // Subsystem toggles
    let enableGeology = true;
    let enableWeather = false;
    let enableLife = false;
    let enableDisease = false;
    let enableSapience = false;
    let enableMigration = false;

    // Presets for quick testing
    const presets = [
        { name: "Quick Geology (10M)", years: 10_000_000, geology: true },
        { name: "Medium Geology (100M)", years: 100_000_000, geology: true },
        { name: "Full Geology (1B)", years: 1_000_000_000, geology: true },
        {
            name: "Life Evolution (500M)",
            years: 500_000_000,
            geology: true,
            life: true,
        },
    ];

    function applyPreset(preset: (typeof presets)[0]) {
        years = preset.years;
        enableGeology = preset.geology ?? false;
        enableWeather = false;
        enableLife = preset.life ?? false;
        enableDisease = false;
        enableSapience = false;
        enableMigration = false;
    }

    function buildCommand(): string {
        const parts = ["world simulate", years.toString()];

        // Add subsystem flags
        const flags: string[] = [];
        if (enableGeology) flags.push("--geology");
        if (enableWeather) flags.push("--weather");
        if (enableLife) flags.push("--life");
        if (enableDisease) flags.push("--disease");
        if (enableSapience) flags.push("--sapience");
        if (enableMigration) flags.push("--migration");

        // If no flags, add --geology by default
        if (flags.length === 0) {
            flags.push("--geology");
        }

        parts.push(...flags);

        // Add optional parameters
        if (seed.trim()) {
            parts.push("--seed", seed.trim());
        }
        if (waterLevel === "custom" && customWaterLevel.trim()) {
            parts.push("--water-level", customWaterLevel.trim());
        } else if (waterLevel !== "medium") {
            parts.push("--water-level", waterLevel);
        }
        if (moons.trim()) {
            parts.push("--moons", moons.trim());
        }

        return parts.join(" ");
    }

    function simulate() {
        dispatch("command", buildCommand());
    }

    function reset() {
        dispatch("command", "world reset");
    }

    function randomSeed() {
        seed = Math.floor(Math.random() * 999_999_999_999).toString();
    }

    function formatYears(y: number): string {
        if (y >= 1_000_000_000) return `${y / 1_000_000_000}B`;
        if (y >= 1_000_000) return `${y / 1_000_000}M`;
        if (y >= 1_000) return `${y / 1_000}K`;
        return y.toString();
    }

    $: commandPreview = buildCommand();
</script>

<div class="simulation-panel p-4 space-y-4">
    <h3 class="text-lg font-bold text-blue-400 border-b border-gray-700 pb-2">
        Simulation Control
    </h3>

    <!-- Presets -->
    <div class="space-y-2">
        <label class="text-sm text-gray-400">Quick Presets</label>
        <div class="flex flex-wrap gap-2">
            {#each presets as preset}
                <button
                    on:click={() => applyPreset(preset)}
                    class="px-3 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded transition-colors"
                >
                    {preset.name}
                </button>
            {/each}
        </div>
    </div>

    <!-- Years -->
    <div class="space-y-1">
        <label class="text-sm text-gray-400" for="years">
            Duration: <span class="text-white font-mono"
                >{formatYears(years)}</span
            > years
        </label>
        <input
            type="range"
            id="years"
            bind:value={years}
            min="1000000"
            max="4500000000"
            step="1000000"
            class="w-full accent-blue-500"
        />
        <div class="flex justify-between text-xs text-gray-500">
            <span>1M</span>
            <span>100M</span>
            <span>1B</span>
            <span>4.5B</span>
        </div>
    </div>

    <!-- Seed -->
    <div class="space-y-1">
        <label class="text-sm text-gray-400" for="seed">Seed (optional)</label>
        <div class="flex gap-2">
            <input
                type="text"
                id="seed"
                bind:value={seed}
                placeholder="Random if empty"
                class="flex-1 bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm"
            />
            <button
                on:click={randomSeed}
                class="px-3 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm"
                title="Generate random seed"
            >
                🎲
            </button>
        </div>
    </div>

    <!-- Water Level -->
    <div class="space-y-1">
        <label class="text-sm text-gray-400" for="water-level"
            >Water Level</label
        >
        <select
            id="water-level"
            bind:value={waterLevel}
            class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm"
        >
            <option value="medium">Medium (default)</option>
            <option value="high">High (flood)</option>
            <option value="low">Low (dry)</option>
            <option value="custom">Custom...</option>
        </select>
        {#if waterLevel === "custom"}
            <input
                type="text"
                bind:value={customWaterLevel}
                placeholder="e.g., 50%, 500m, or high"
                class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm mt-1"
            />
        {/if}
    </div>

    <!-- Moons -->
    <div class="space-y-1">
        <label class="text-sm text-gray-400" for="moons">Moons (optional)</label
        >
        <input
            type="text"
            id="moons"
            bind:value={moons}
            placeholder="Random if empty (0-10)"
            class="w-full bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm"
        />
        <p class="text-xs text-gray-500">
            Affects tides, axial stability, impact shielding
        </p>
    </div>

    <!-- Subsystem Toggles -->
    <div class="space-y-2">
        <label class="text-sm text-gray-400">Subsystems</label>
        <div class="grid grid-cols-2 gap-2">
            <label class="flex items-center gap-2 text-sm">
                <input
                    type="checkbox"
                    bind:checked={enableGeology}
                    class="accent-blue-500"
                />
                <span>Geology</span>
            </label>
            <label class="flex items-center gap-2 text-sm">
                <input
                    type="checkbox"
                    bind:checked={enableWeather}
                    class="accent-blue-500"
                />
                <span>Weather</span>
            </label>
            <label class="flex items-center gap-2 text-sm">
                <input
                    type="checkbox"
                    bind:checked={enableLife}
                    class="accent-blue-500"
                />
                <span>Life</span>
            </label>
            <label class="flex items-center gap-2 text-sm">
                <input
                    type="checkbox"
                    bind:checked={enableDisease}
                    class="accent-blue-500"
                />
                <span>Disease</span>
            </label>
            <label class="flex items-center gap-2 text-sm">
                <input
                    type="checkbox"
                    bind:checked={enableSapience}
                    class="accent-blue-500"
                />
                <span>Sapience</span>
            </label>
            <label class="flex items-center gap-2 text-sm">
                <input
                    type="checkbox"
                    bind:checked={enableMigration}
                    class="accent-blue-500"
                />
                <span>Migration</span>
            </label>
        </div>
    </div>

    <!-- Command Preview -->
    <div class="space-y-1">
        <label class="text-sm text-gray-400">Command Preview</label>
        <code
            class="block bg-gray-800 p-2 rounded text-xs font-mono text-green-400 break-all"
        >
            {commandPreview}
        </code>
    </div>

    <!-- Action Buttons -->
    <div class="flex gap-2 pt-2">
        <button
            on:click={simulate}
            class="flex-1 bg-blue-600 hover:bg-blue-500 text-white font-bold py-3 rounded transition-colors"
        >
            ▶ Simulate
        </button>
        <button
            on:click={reset}
            class="px-4 py-3 bg-red-700 hover:bg-red-600 text-white rounded transition-colors"
            title="Reset world state"
        >
            ↺ Reset
        </button>
    </div>
</div>

<style>
    .simulation-panel {
        background: rgba(17, 24, 39, 0.95);
        border-left: 1px solid rgba(75, 85, 99, 0.5);
    }
</style>
