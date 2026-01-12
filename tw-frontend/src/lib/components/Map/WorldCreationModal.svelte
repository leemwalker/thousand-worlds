<script lang="ts">
    import { createEventDispatcher } from "svelte";

    export let isOpen = false;
    let step = 0;

    const dispatch = createEventDispatcher();

    function handleComplete() {
        // Dispatch completion event
        dispatch("complete", {
            name: "New World",
            seed: Math.random().toString(36).substring(7),
        });
        step = 0;
    }
</script>

{#if isOpen}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-md"
    >
        <div
            class="w-full max-w-lg bg-gray-900 border border-blue-500/30 rounded-xl p-8 shadow-2xl"
        >
            <h2
                class="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-500 mb-6"
            >
                Genesis Protocol
            </h2>

            {#if step === 0}
                <div class="space-y-4">
                    <p class="text-gray-300">
                        Initiating planetary formation sequence...
                    </p>
                    <button
                        class="w-full py-3 bg-blue-600 hover:bg-blue-500 rounded-lg text-white font-semibold transition-all"
                        on:click={() => (step = 1)}
                    >
                        Initialize Core
                    </button>
                </div>
            {:else}
                <div class="space-y-4">
                    <p class="text-gray-300">
                        Constructing terrain and atmosphere...
                    </p>
                    <button
                        class="w-full py-3 bg-green-600 hover:bg-green-500 rounded-lg text-white font-semibold transition-all"
                        on:click={handleComplete}
                    >
                        Finalize Biosphere
                    </button>
                </div>
            {/if}
        </div>
    </div>
{/if}
