<script lang="ts">
    import { createEventDispatcher } from "svelte";

    export let options: any;

    const dispatch = createEventDispatcher();

    function selectOption(type: string, data?: any) {
        dispatch("select", { type, data });
    }
</script>

<div
    class="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4"
>
    <div
        class="bg-gray-900 border border-blue-500/50 p-6 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto"
    >
        <h2
            class="text-2xl font-bold text-blue-400 mb-6 text-center font-mono uppercase tracking-widest"
        >
            Choose Your Path
        </h2>

        <div class="grid gap-6">
            {#if options.can_enter_as_watcher}
                <button
                    class="p-4 border border-gray-700 hover:border-blue-400 hover:bg-blue-900/20 rounded text-left transition-all group"
                    on:click={() => selectOption("watcher")}
                >
                    <div class="flex items-center justify-between mb-2">
                        <h3
                            class="text-lg font-bold text-gray-200 group-hover:text-blue-300"
                        >
                            Watcher
                        </h3>
                        <span
                            class="text-xs bg-blue-900/50 text-blue-300 px-2 py-1 rounded border border-blue-500/30"
                            >Invisible</span
                        >
                    </div>
                    <p class="text-gray-400 text-sm">
                        Enter as an invisible observer. You can explore freely
                        without stamina cost, but cannot interact with the
                        world.
                    </p>
                </button>
            {/if}

            {#if options.available_npcs && options.available_npcs.length > 0}
                <div class="space-y-3">
                    <div class="flex items-center justify-between">
                        <h3 class="text-lg font-bold text-gray-200">
                            Inhabit Existing NPC
                        </h3>
                        <span class="text-xs text-gray-500"
                            >Select a life to take over</span
                        >
                    </div>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {#each options.available_npcs as npc}
                            <button
                                class="p-3 border border-gray-700 hover:border-green-400 hover:bg-green-900/20 rounded text-left transition-all group h-full flex flex-col"
                                on:click={() => selectOption("npc", npc)}
                            >
                                <div
                                    class="font-bold text-green-300 group-hover:text-green-200 mb-1"
                                >
                                    {npc.name}
                                </div>
                                <div
                                    class="text-xs text-gray-400 mb-2 font-mono"
                                >
                                    {npc.species} • {npc.occupation}
                                </div>
                                <div class="text-xs text-gray-500 flex-1">
                                    {npc.description}
                                </div>
                            </button>
                        {/each}
                    </div>
                </div>
            {/if}

            {#if options.can_create_custom}
                <button
                    class="p-4 border border-gray-700 hover:border-purple-400 hover:bg-purple-900/20 rounded text-left transition-all group"
                    on:click={() => selectOption("custom")}
                >
                    <div class="flex items-center justify-between mb-2">
                        <h3
                            class="text-lg font-bold text-gray-200 group-hover:text-purple-300"
                        >
                            Create New Character
                        </h3>
                        <span
                            class="text-xs bg-purple-900/50 text-purple-300 px-2 py-1 rounded border border-purple-500/30"
                            >Custom</span
                        >
                    </div>
                    <p class="text-gray-400 text-sm">
                        Design your own character from scratch. Choose your
                        species, name, and attributes.
                    </p>
                </button>
            {/if}
        </div>

        <div class="mt-6 text-center">
            <button
                class="text-gray-500 hover:text-gray-300 text-sm underline"
                on:click={() => selectOption("cancel")}
            >
                Cancel
            </button>
        </div>
    </div>
</div>
