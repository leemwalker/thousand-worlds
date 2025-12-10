<script lang="ts">
    export let characterName: string = "Adventurer";
    export let level: number = 1;
    export let experience: number = 0;
    export let nextLevelXP: number = 100;

    export let strength: number = 10;
    export let dexterity: number = 10;
    export let constitution: number = 10;
    export let intelligence: number = 10;
    export let wisdom: number = 10;
    export let charisma: number = 10;

    import type { Skill } from "$lib/types/game";
    export let skills: Record<string, Skill> = {};

    $: xpPercent = (experience / nextLevelXP) * 100;

    let activeTab: "stats" | "skills" = "stats";
</script>

<div
    class="character-sheet bg-gray-900 rounded-lg p-4 w-full max-w-md mx-auto max-h-[80vh] overflow-y-auto"
>
    <!-- Character Info -->
    <div class="mb-4">
        <h2 class="text-2xl font-bold text-gray-100">{characterName}</h2>
        <div class="text-gray-400">Level {level}</div>

        <!-- XP Bar -->
        <div class="mt-2">
            <div class="flex justify-between text-xs text-gray-400 mb-1">
                <span>Experience</span>
                <span>{experience}/{nextLevelXP}</span>
            </div>
            <div class="w-full bg-gray-700 rounded-full h-2">
                <div
                    class="bg-purple-500 h-2 rounded-full transition-all"
                    style="width: {xpPercent}%"
                />
            </div>
        </div>
    </div>

    <!-- Tabs -->
    <div class="flex border-b border-gray-700 mb-4">
        <button
            class="flex-1 py-2 text-center text-sm font-bold uppercase transition-colors {activeTab ===
            'stats'
                ? 'text-blue-400 border-b-2 border-blue-400'
                : 'text-gray-500 hover:text-gray-300'}"
            on:click={() => (activeTab = "stats")}
        >
            Stats
        </button>
        <button
            class="flex-1 py-2 text-center text-sm font-bold uppercase transition-colors {activeTab ===
            'skills'
                ? 'text-blue-400 border-b-2 border-blue-400'
                : 'text-gray-500 hover:text-gray-300'}"
            on:click={() => (activeTab = "skills")}
        >
            Skills
        </button>
    </div>

    <!-- Attributes Tab -->
    {#if activeTab === "stats"}
        <div>
            <h3 class="text-lg font-bold text-gray-100 mb-2 sr-only">
                Attributes
            </h3>
            <div class="grid grid-cols-2 gap-2">
                <div class="bg-gray-800 rounded p-2">
                    <div class="text-xs text-gray-400">Strength</div>
                    <div class="text-xl font-bold text-red-400">{strength}</div>
                </div>
                <div class="bg-gray-800 rounded p-2">
                    <div class="text-xs text-gray-400">Dexterity</div>
                    <div class="text-xl font-bold text-green-400">
                        {dexterity}
                    </div>
                </div>
                <div class="bg-gray-800 rounded p-2">
                    <div class="text-xs text-gray-400">Constitution</div>
                    <div class="text-xl font-bold text-orange-400">
                        {constitution}
                    </div>
                </div>
                <div class="bg-gray-800 rounded p-2">
                    <div class="text-xs text-gray-400">Intelligence</div>
                    <div class="text-xl font-bold text-blue-400">
                        {intelligence}
                    </div>
                </div>
                <div class="bg-gray-800 rounded p-2">
                    <div class="text-xs text-gray-400">Wisdom</div>
                    <div class="text-xl font-bold text-cyan-400">{wisdom}</div>
                </div>
                <div class="bg-gray-800 rounded p-2">
                    <div class="text-xs text-gray-400">Charisma</div>
                    <div class="text-xl font-bold text-purple-400">
                        {charisma}
                    </div>
                </div>
            </div>
        </div>
    {/if}

    <!-- Skills Tab -->
    {#if activeTab === "skills"}
        <div class="mt-0">
            <h3 class="text-lg font-bold text-gray-100 mb-2 sr-only">Skills</h3>
            {#if Object.keys(skills).length === 0}
                <div class="text-gray-500 text-sm italic py-4 text-center">
                    No skills learned yet.
                </div>
            {:else}
                <div class="grid grid-cols-1 gap-2">
                    {#each Object.values(skills) as skill}
                        <div
                            class="bg-gray-800 rounded p-2 flex justify-between items-center"
                        >
                            <div>
                                <div class="text-sm font-bold text-gray-200">
                                    {skill.name}
                                </div>
                                <div class="text-xs text-gray-400">
                                    {skill.category}
                                </div>
                            </div>
                            <div class="text-right">
                                <div class="text-sm font-mono text-blue-400">
                                    Level {skill.level}
                                </div>
                                <div class="text-xs text-gray-500">
                                    {skill.xp} XP
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    {/if}
</div>
