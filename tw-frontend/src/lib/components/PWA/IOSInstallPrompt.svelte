<script lang="ts">
    import { onMount } from "svelte";
    import {
        showIOSInstallPrompt,
        dismissIOSInstallPrompt,
    } from "$lib/stores/pwa";
    import { fly, fade } from "svelte/transition";

    let show = false;

    onMount(() => {
        const unsubscribe = showIOSInstallPrompt.subscribe((value) => {
            show = value;
        });

        return unsubscribe;
    });

    function dismiss() {
        dismissIOSInstallPrompt();
    }
</script>

{#if show}
    <div
        class="fixed inset-0 bg-black/60 z-50 flex items-end md:items-center md:justify-center p-4"
        transition:fade={{ duration: 200 }}
        role="button"
        tabindex="0"
        on:click={dismiss}
        on:keydown={(e) => e.key === "Escape" && dismiss()}
    >
        <div
            class="bg-gray-900 rounded-2xl p-6 max-w-md w-full border border-gray-700 shadow-2xl"
            transition:fly={{ y: 100, duration: 300 }}
            role="document"
            on:click|stopPropagation
            on:keydown|stopPropagation
            tabindex="-1"
        >
            <div class="flex items-center gap-3 mb-4">
                <div
                    class="w-12 h-12 bg-blue-600 rounded-xl flex items-center justify-center"
                >
                    <svg
                        class="w-6 h-6 text-white"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M12 4v16m8-8H4"
                        />
                    </svg>
                </div>
                <div>
                    <h3 class="text-lg font-semibold text-white">
                        Install Thousand Worlds
                    </h3>
                    <p class="text-sm text-gray-400">Add to your Home Screen</p>
                </div>
            </div>

            <p class="text-gray-300 mb-6">
                Get the best experience with offline support and quick access:
            </p>

            <ol class="space-y-3 mb-6 text-sm">
                <li class="flex items-start gap-3 text-gray-300">
                    <span
                        class="flex-shrink-0 w-6 h-6 bg-blue-600/20 text-blue-400 rounded-full flex items-center justify-center text-xs font-bold"
                        >1</span
                    >
                    <span
                        >Tap the <strong>Share</strong> button
                        <svg
                            class="inline w-4 h-4"
                            fill="currentColor"
                            viewBox="0 0 20 20"
                        >
                            <path
                                d="M15 8a3 3 0 10-2.977-2.63l-4.94 2.47a3 3 0 100 4.319l4.94 2.47a3 3 0 10.895-1.789l-4.94-2.47a3.027 3.027 0 000-.74l4.94-2.47C13.456 7.68 14.19 8 15 8z"
                            />
                        </svg>
                    </span>
                </li>
                <li class="flex items-start gap-3 text-gray-300">
                    <span
                        class="flex-shrink-0 w-6 h-6 bg-blue-600/20 text-blue-400 rounded-full flex items-center justify-center text-xs font-bold"
                        >2</span
                    >
                    <span
                        >Scroll and tap <strong>"Add to Home Screen"</strong
                        ></span
                    >
                </li>
                <li class="flex items-start gap-3 text-gray-300">
                    <span
                        class="flex-shrink-0 w-6 h-6 bg-blue-600/20 text-blue-400 rounded-full flex items-center justify-center text-xs font-bold"
                        >3</span
                    >
                    <span>Tap <strong>"Add"</strong> in the top right</span>
                </li>
            </ol>

            <button
                on:click={dismiss}
                class="w-full py-3 bg-gray-800 hover:bg-gray-700 text-white rounded-lg font-medium transition-colors"
            >
                Got it
            </button>
        </div>
    </div>
{/if}

<style>
    /* Prevent body scroll when modal is open */
    :global(body:has(.fixed.inset-0)) {
        overflow: hidden;
    }
</style>
